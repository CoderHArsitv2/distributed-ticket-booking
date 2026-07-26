package locking

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"distributed-ticket-booking/models"
)

// PessimisticLocker uses SELECT ... FOR UPDATE inside a GORM transaction to lock
// the target seat row until commit/rollback. Highest consistency, lowest
// throughput. Best for low-to-medium contention.
type PessimisticLocker struct {
	db  *gorm.DB
	ttl time.Duration
}

// NewPessimisticLocker constructs the pessimistic (row-lock) strategy.
func NewPessimisticLocker(db *gorm.DB, ttl time.Duration) *PessimisticLocker {
	return &PessimisticLocker{db: db, ttl: ttl}
}

func (l *PessimisticLocker) Name() string { return "pessimistic" }

func (l *PessimisticLocker) Acquire(ctx context.Context, seatID, userID int64) (string, error) {
	err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var seat models.Seat
		// Row-level lock held until this transaction commits/rolls back.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&seat, seatID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSeatUnavailable
			}
			return err
		}
		if seat.Status != models.SeatAvailable {
			return ErrSeatUnavailable
		}
		until := time.Now().Add(l.ttl)
		return tx.Model(&seat).Updates(map[string]any{
			"status":         models.SeatReserved,
			"reserved_until": until,
			"version":        gorm.Expr("version + 1"),
		}).Error
	})
	if err != nil {
		return "", err
	}
	// DB strategies do not use an opaque token; the seat row IS the lock.
	return "", nil
}

func (l *PessimisticLocker) Release(ctx context.Context, seatID int64, token string) error {
	return l.db.WithContext(ctx).Model(&models.Seat{}).
		Where("id = ? AND status = ?", seatID, models.SeatReserved).
		Updates(map[string]any{
			"status":         models.SeatAvailable,
			"reserved_until": nil,
			"version":        gorm.Expr("version + 1"),
		}).Error
}
