// Command server is the HTTP entrypoint for the distributed ticket-booking
// engine. It loads config, connects to PostgreSQL (GORM) and Redis, selects a
// locking strategy, wires the layers, starts the expiry sweeper, and serves the
// Gin API with graceful shutdown.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"distributed-ticket-booking/config"
	"distributed-ticket-booking/controllers"
	"distributed-ticket-booking/models"
	"distributed-ticket-booking/pkg/locking"
	"distributed-ticket-booking/pkg/service"
	"distributed-ticket-booking/router"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		log.Error("config load failed", "err", err)
		os.Exit(1)
	}

	// PostgreSQL via GORM.
	gormDB, err := config.OpenDB(cfg)
	if err != nil {
		log.Error("database connect failed", "err", err)
		os.Exit(1)
	}
	if err := models.AutoMigrate(gormDB); err != nil {
		log.Error("auto-migrate failed", "err", err)
		os.Exit(1)
	}
	log.Info("database connected")

	// Model stores (GORM-backed).
	seatStore := models.NewSeatStore(gormDB)
	holdStore := models.NewReservationStore(gormDB)

	// Select the concurrency-control strategy from config. main only ever sees
	// the locking.SeatLocker interface, never a concrete strategy.
	locker, err := locking.New(locking.Config{
		Strategy: cfg.LockStrategy,
		DB:       gormDB,
		RedisURL: cfg.RedisURL,
		HoldTTL:  cfg.HoldTTL,
	})
	if err != nil {
		log.Error("locker init failed", "err", err)
		os.Exit(1)
	}
	defer locker.Close()
	log.Info("locking strategy selected", "strategy", locker.Name())

	reservationSvc := service.NewReservationService(locker, seatStore, holdStore, cfg.HoldTTL)

	engine := router.New(router.Deps{
		Health:      controllers.NewHealthController(),
		Reservation: controllers.NewReservationController(reservationSvc),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Root context cancelled on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Background expiry sweeper (Phase 5).
	sweeper := service.NewSweeper(seatStore, holdStore, cfg.SweepEvery, log)
	go sweeper.Run(ctx)

	go func() {
		log.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "err", err)
	}
}
