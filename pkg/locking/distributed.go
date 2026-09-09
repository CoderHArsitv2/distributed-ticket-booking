package locking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"distributed-ticket-booking/models"
)

// releaseScript is the atomic compare-and-delete Lua script: only delete the
// lock key if it still holds the token we set (prevents releasing someone
// else's lock after our TTL expired).
const releaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`

// distributedLocker uses Redis SET key token NX PX <ttl> to acquire, and the
// atomic Lua script above to release. Works across multiple service instances;
// best for flash sales. For multi-node Redis, swap in Redsync (Redlock).
//
// Acquiring the Redis lock grants the right to flip the seat to RESERVED in the
// database, keeping the durable seat state consistent across the three
// strategies.
type distributedLocker struct {
	rdb *redis.Client
	db  *gorm.DB
	ttl time.Duration
}

// newDistributedLocker dials Redis from redisURL and constructs the
// distributed-lock strategy. The client it opens is owned by the locker and
// shut down by Close.
func newDistributedLocker(redisURL string, db *gorm.DB, ttl time.Duration) (*distributedLocker, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return &distributedLocker{rdb: redis.NewClient(opts), db: db, ttl: ttl}, nil
}

func (l *distributedLocker) Name() string { return string(StrategyDistributed) }

// Close shuts down the Redis client this locker opened.
func (l *distributedLocker) Close() error { return l.rdb.Close() }

func lockKey(seatID int64) string { return fmt.Sprintf("seatlock:%d", seatID) }

func (l *distributedLocker) Acquire(ctx context.Context, seatID, userID int64) (string, error) {
	token := uuid.NewString()

	ok, err := l.rdb.SetNX(ctx, lockKey(seatID), token, l.ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrSeatUnavailable
	}

	// Lock held: durably flip the seat to RESERVED.
	until := time.Now().Add(l.ttl)
	res := l.db.WithContext(ctx).Model(&models.Seat{}).
		Where("id = ? AND status = ?", seatID, models.SeatAvailable).
		Updates(map[string]any{
			"status":         models.SeatReserved,
			"reserved_until": until,
			"version":        gorm.Expr("version + 1"),
		})
	if res.Error != nil || res.RowsAffected == 0 {
		// Roll back the Redis lock we just took so we don't strand the seat.
		_ = l.Release(ctx, seatID, token)
		if res.Error != nil {
			return "", res.Error
		}
		return "", ErrSeatUnavailable
	}
	return token, nil
}

func (l *distributedLocker) Release(ctx context.Context, seatID int64, token string) error {
	// Atomically drop the lock only if we still own it.
	if err := l.rdb.Eval(ctx, releaseScript, []string{lockKey(seatID)}, token).Err(); err != nil && err != redis.Nil {
		return err
	}
	// Revert the seat if it is still held.
	return l.db.WithContext(ctx).Model(&models.Seat{}).
		Where("id = ? AND status = ?", seatID, models.SeatReserved).
		Updates(map[string]any{
			"status":         models.SeatAvailable,
			"reserved_until": nil,
			"version":        gorm.Expr("version + 1"),
		}).Error
}
