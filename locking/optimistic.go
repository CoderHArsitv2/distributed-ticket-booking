package locking

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"distributed-ticket-booking/models"
)

// OptimisticLocker uses a version column and a compare-and-swap UPDATE:
//
//	UPDATE seats SET status='RESERVED', version=version+1
//	WHERE id=? AND version=? AND status='AVAILABLE'
//
// If zero rows are affected, another writer won the race. Non-blocking; best for
// low write contention and high-read workloads.
type OptimisticLocker struct {
	db  *gorm.DB
	ttl time.Duration
}

// NewOptimisticLocker constructs the optimistic (version-CAS) strategy.
func NewOptimisticLocker(db *gorm.DB, ttl time.Duration) *OptimisticLocker {
	return &OptimisticLocker{db: db, ttl: ttl}
}

func (l *OptimisticLocker) Name() string { return "optimistic" }

func (l *OptimisticLocker) Acquire(ctx context.Context, seatID, userID int64) (string, error) {
	// Read the current version, then attempt a version-guarded CAS update.
	var seat models.Seat
	if err := l.db.WithContext(ctx).First(&seat, seatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSeatUnavailable
		}
		return "", err
	}
	if seat.Status != models.SeatAvailable {
		return "", ErrSeatUnavailable
	}

	until := time.Now().Add(l.ttl)
	res := l.db.WithContext(ctx).Model(&models.Seat{}).
		Where("id = ? AND version = ? AND status = ?", seatID, seat.Version, models.SeatAvailable).
		Updates(map[string]any{
			"status":         models.SeatReserved,
			"reserved_until": until,
			"version":        gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return "", res.Error
	}
	if res.RowsAffected == 0 {
		// Another writer bumped the version between our read and write.
		return "", ErrSeatUnavailable
	}
	return "", nil
}

func (l *OptimisticLocker) Release(ctx context.Context, seatID int64, token string) error {
	return l.db.WithContext(ctx).Model(&models.Seat{}).
		Where("id = ? AND status = ?", seatID, models.SeatReserved).
		Updates(map[string]any{
			"status":         models.SeatAvailable,
			"reserved_until": nil,
			"version":        gorm.Expr("version + 1"),
		}).Error
}
