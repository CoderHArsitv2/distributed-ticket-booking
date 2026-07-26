package locking

import "context"

// releaseScript is the atomic compare-and-delete Lua script: only delete the
// lock key if it still holds the token we set (prevents releasing someone
// else's lock after our TTL expired).
const releaseScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`

// DistributedLocker uses Redis SET key token NX PX <ttl> to acquire, and the
// atomic Lua script above to release. Works across multiple service instances;
// best for flash sales. For multi-node Redis, swap in Redsync (Redlock).
//
// TODO(phase-4): inject *redis.Client, generate a unique token per Acquire,
// SET NX PX with the hold TTL, and EVAL releaseScript on Release.
type DistributedLocker struct {
	// rdb *redis.Client
	// ttl time.Duration
}

// NewDistributedLocker constructs the Redis distributed-lock strategy.
func NewDistributedLocker() *DistributedLocker { return &DistributedLocker{} }

func (l *DistributedLocker) Name() string { return "distributed" }

func (l *DistributedLocker) Acquire(ctx context.Context, seatID, userID int64) (string, error) {
	// TODO(phase-4): token := uuid; ok := SET lock:<seatID> token NX PX ttl.
	return "", ErrSeatUnavailable
}

func (l *DistributedLocker) Release(ctx context.Context, seatID int64, token string) error {
	// TODO(phase-4): EVAL releaseScript, KEYS=[lock:<seatID>], ARGV=[token].
	_ = releaseScript
	return nil
}
