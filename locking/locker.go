// Package locking defines the concurrency-control strategies used to guarantee
// zero double-booking under write contention. Each strategy implements the
// SeatLocker interface so the reservation service can be configured at runtime.
package locking

import (
	"context"
	"errors"
)

// ErrSeatUnavailable is returned when a seat cannot be locked/held because it is
// already reserved or booked by someone else.
var ErrSeatUnavailable = errors.New("seat unavailable")

// ErrLockLost is returned when a previously held lock could no longer be
// confirmed at release time (e.g. optimistic version mismatch or Redis token
// mismatch). Callers should treat this as a failed operation, not success.
var ErrLockLost = errors.New("lock lost")

// SeatLocker acquires exclusive control of a seat long enough to transition it
// to RESERVED, then releases. Implementations must be safe for concurrent use.
type SeatLocker interface {
	// Acquire attempts to hold the given seat for the user. On success it
	// returns an opaque token that must be passed back to Release.
	Acquire(ctx context.Context, seatID, userID int64) (token string, err error)

	// Release relinquishes a previously acquired hold identified by token.
	Release(ctx context.Context, seatID int64, token string) error

	// Name identifies the strategy for logging and benchmarking.
	Name() string
}
