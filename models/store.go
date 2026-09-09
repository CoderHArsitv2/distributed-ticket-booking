// Package models holds the domain entities and the data-access contracts for
// each of them. The GORM-backed implementations live alongside their entity
// (event.go, seat.go, reservation.go, booking.go); this file is the single
// place to look for the interfaces the service layer depends on.
package models

import "context"

// EventStore reads and writes event records.
type EventStore interface {
	GetByID(ctx context.Context, id int64) (*Event, error)
	List(ctx context.Context) ([]Event, error)
	Create(ctx context.Context, e *Event) error
}

// SeatStore reads and writes seat inventory. The locking strategies build on
// the transactional primitives here.
type SeatStore interface {
	GetByID(ctx context.Context, id int64) (*Seat, error)
	ListByEvent(ctx context.Context, eventID int64) ([]Seat, error)
	// ReserveOptimistic performs the version-guarded CAS update.
	ReserveOptimistic(ctx context.Context, seatID, expectedVersion int64) (bool, error)
	// ReleaseExpired flips RESERVED seats past their reserved_until back to
	// AVAILABLE and returns the number affected. Used by the sweeper.
	ReleaseExpired(ctx context.Context) (int64, error)
}

// ReservationStore manages ephemeral holds.
type ReservationStore interface {
	Create(ctx context.Context, r *Reservation) error
	GetActiveBySeat(ctx context.Context, seatID int64) (*Reservation, error)
	MarkExpired(ctx context.Context) (int64, error)
}

// BookingStore persists finalized orders and their seat snapshots.
type BookingStore interface {
	Create(ctx context.Context, b *Booking, seats []BookingSeat) error
	GetByReference(ctx context.Context, ref string) (*Booking, error)
}
