package locking

import "context"

// OptimisticLocker uses a version column and a compare-and-swap UPDATE:
//
//	UPDATE seats SET status='RESERVED', version=version+1
//	WHERE id=? AND version=? AND status='AVAILABLE'
//
// If zero rows are affected, another writer won the race. Non-blocking; best for
// low write contention and high-read workloads.
//
// TODO(phase-4): inject *sql.DB (or *sqlx.DB) and implement the CAS UPDATE.
type OptimisticLocker struct {
	// db *sqlx.DB
}

// NewOptimisticLocker constructs the optimistic (version-CAS) strategy.
func NewOptimisticLocker() *OptimisticLocker { return &OptimisticLocker{} }

func (l *OptimisticLocker) Name() string { return "optimistic" }

func (l *OptimisticLocker) Acquire(ctx context.Context, seatID, userID int64) (string, error) {
	// TODO(phase-4): read current version, then conditional UPDATE; retry on miss.
	return "", ErrSeatUnavailable
}

func (l *OptimisticLocker) Release(ctx context.Context, seatID int64, token string) error {
	// TODO(phase-4): conditional UPDATE back to AVAILABLE guarded by version.
	return nil
}
