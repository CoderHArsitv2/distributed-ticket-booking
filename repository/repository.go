// Package repository defines the data-access contracts (repository pattern).
// GORM-backed implementations live in gorm_repository.go.
package repository

import (
	"context"

	"distributed-ticket-booking/models"
)

// EventRepository reads and writes event records.
type EventRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	List(ctx context.Context) ([]models.Event, error)
	Create(ctx context.Context, e *models.Event) error
}

// SeatRepository reads and writes seat inventory. The locking strategies build
// on the transactional primitives here.
type SeatRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Seat, error)
	ListByEvent(ctx context.Context, eventID int64) ([]models.Seat, error)
	// ReserveOptimistic performs the version-guarded CAS update.
	ReserveOptimistic(ctx context.Context, seatID, expectedVersion int64) (bool, error)
	// ReleaseExpired flips RESERVED seats past their reserved_until back to
	// AVAILABLE and returns the number affected. Used by the sweeper.
	ReleaseExpired(ctx context.Context) (int64, error)
}

// ReservationRepository manages ephemeral holds.
type ReservationRepository interface {
	Create(ctx context.Context, r *models.Reservation) error
	GetActiveBySeat(ctx context.Context, seatID int64) (*models.Reservation, error)
	MarkExpired(ctx context.Context) (int64, error)
}

// BookingRepository persists finalized orders and their seat snapshots.
type BookingRepository interface {
	Create(ctx context.Context, b *models.Booking, seats []models.BookingSeat) error
	GetByReference(ctx context.Context, ref string) (*models.Booking, error)
}
