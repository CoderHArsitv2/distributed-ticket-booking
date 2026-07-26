// Package db opens the GORM/PostgreSQL connection and runs schema migration.
package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"distributed-ticket-booking/config"
	"distributed-ticket-booking/models"
)

// Open connects to PostgreSQL via GORM and configures the connection pool.
func Open(cfg *config.Config) (*gorm.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		SkipDefaultTransaction: true, // we manage transactions explicitly in the lockers
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return gormDB, nil
}

// AutoMigrate creates/updates tables to match the model definitions. The raw
// SQL migrations under migrations/ remain the source of truth for production;
// AutoMigrate is convenient for local development.
func AutoMigrate(gormDB *gorm.DB) error {
	return gormDB.AutoMigrate(
		&models.Event{},
		&models.Seat{},
		&models.Reservation{},
		&models.Booking{},
		&models.BookingSeat{},
	)
}

// Ping verifies connectivity within a short timeout.
func Ping(gormDB *gorm.DB) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
