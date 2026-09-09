// Package locking defines the concurrency-control strategies used to guarantee
// zero double-booking under write contention.
//
// This file is the package's entire public surface: callers depend on the
// SeatLocker interface and build one through New. The concrete strategies
// (pessimistic row locks, optimistic version CAS, Redis distributed locks) are
// unexported on purpose — swapping a strategy is a configuration change, never
// a code change in the caller.
package locking

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Strategy names a concurrency-control implementation. It is the value callers
// pass to New to pick one.
type Strategy string

const (
	StrategyPessimistic Strategy = "pessimistic" // SELECT ... FOR UPDATE
	StrategyOptimistic  Strategy = "optimistic"  // version column CAS
	StrategyDistributed Strategy = "distributed" // Redis SET NX PX + Lua release
)

// ErrSeatUnavailable is returned when a seat cannot be locked/held because it is
// already reserved or booked by someone else.
var ErrSeatUnavailable = errors.New("seat unavailable")

// ErrLockLost is returned when a previously held lock could no longer be
// confirmed at release time (e.g. optimistic version mismatch or Redis token
// mismatch). Callers should treat this as a failed operation, not success.
var ErrLockLost = errors.New("lock lost")

// ErrUnknownStrategy is returned by New for a Strategy it has no implementation
// for.
var ErrUnknownStrategy = errors.New("unknown locking strategy")

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

	// Close releases any resources the strategy owns (e.g. the Redis client).
	// It is a no-op for the database-only strategies, so callers can always
	// defer it without knowing which strategy is active.
	Close() error
}

// Config is everything the strategies need to be built. RedisURL is only read
// by StrategyDistributed.
type Config struct {
	Strategy Strategy
	DB       *gorm.DB
	RedisURL string
	HoldTTL  time.Duration
}

// New builds the configured strategy and returns it as a SeatLocker. This is
// the only constructor the package exports, so callers cannot bind themselves
// to a concrete implementation.
func New(cfg Config) (SeatLocker, error) {
	switch cfg.Strategy {
	case StrategyPessimistic:
		return newPessimisticLocker(cfg.DB, cfg.HoldTTL), nil
	case StrategyOptimistic:
		return newOptimisticLocker(cfg.DB, cfg.HoldTTL), nil
	case StrategyDistributed:
		// Assigned first so a failed dial returns a nil interface, not a
		// non-nil interface holding a nil pointer.
		locker, err := newDistributedLocker(cfg.RedisURL, cfg.DB, cfg.HoldTTL)
		if err != nil {
			return nil, err
		}
		return locker, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownStrategy, cfg.Strategy)
	}
}

// Compile-time proof that every strategy satisfies the contract above.
var (
	_ SeatLocker = (*pessimisticLocker)(nil)
	_ SeatLocker = (*optimisticLocker)(nil)
	_ SeatLocker = (*distributedLocker)(nil)
)
