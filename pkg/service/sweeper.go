package service

import (
	"context"
	"log/slog"
	"time"

	"distributed-ticket-booking/models"
)

// Sweeper is the background worker that releases expired RESERVED seats back to
// AVAILABLE when payment was not completed within the hold window (README
// Phase 5).
type Sweeper struct {
	seats    models.SeatStore
	holds    models.ReservationStore
	interval time.Duration
	log      *slog.Logger
}

// NewSweeper constructs the expiry sweeper.
func NewSweeper(
	seats models.SeatStore,
	holds models.ReservationStore,
	interval time.Duration,
	log *slog.Logger,
) *Sweeper {
	return &Sweeper{seats: seats, holds: holds, interval: interval, log: log}
}

// Run ticks until ctx is cancelled, releasing expired holds on each tick.
func (s *Sweeper) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Info("sweeper started", "interval", s.interval)
	for {
		select {
		case <-ctx.Done():
			s.log.Info("sweeper stopped")
			return
		case <-ticker.C:
			s.sweep(ctx)
		}
	}
}

func (s *Sweeper) sweep(ctx context.Context) {
	// TODO(phase-5): both calls should run inside one tx.
	if s.seats == nil || s.holds == nil {
		return
	}
	freed, err := s.seats.ReleaseExpired(ctx)
	if err != nil {
		s.log.Error("release expired seats failed", "err", err)
		return
	}
	if _, err := s.holds.MarkExpired(ctx); err != nil {
		s.log.Error("mark expired reservations failed", "err", err)
		return
	}
	if freed > 0 {
		s.log.Info("swept expired holds", "seats_freed", freed)
	}
}
