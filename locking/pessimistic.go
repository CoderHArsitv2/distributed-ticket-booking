package locking

import "context"

// PessimisticLocker uses SELECT ... FOR UPDATE inside a DB transaction to lock
// the target seat row until commit/rollback. Highest consistency, lowest
// throughput. Best for low-to-medium contention.
//
// TODO(phase-4): inject *sql.DB (or *sqlx.DB) and implement the FOR UPDATE
// transaction. Acquire should BEGIN, SELECT ... FOR UPDATE the seat, verify it
// is AVAILABLE, UPDATE it to RESERVED, then COMMIT.
type PessimisticLocker struct {
	// db *sqlx.DB
}

// NewPessimisticLocker constructs the pessimistic (row-lock) strategy.
func NewPessimisticLocker() *PessimisticLocker { return &PessimisticLocker{} }

func (l *PessimisticLocker) Name() string { return "pessimistic" }

func (l *PessimisticLocker) Acquire(ctx context.Context, seatID, userID int64) (string, error) {
	// TODO(phase-4): implement SELECT ... FOR UPDATE hold.
	return "", ErrSeatUnavailable
}

func (l *PessimisticLocker) Release(ctx context.Context, seatID int64, token string) error {
	// TODO(phase-4): release is implicit at COMMIT/ROLLBACK for the DB row lock;
	// this may only revert an uncommitted hold if needed.
	return nil
}
