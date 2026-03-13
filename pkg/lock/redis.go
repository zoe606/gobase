package lock

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RedisClient defines the Redis operations needed for distributed locking.
type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, exp time.Duration) (bool, error)
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) error
}

const unlockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

const lockRetryInterval = 50 * time.Millisecond

// RedisLocker implements distributed locking using Redis SetNX + Lua unlock.
type RedisLocker struct {
	client RedisClient
}

// NewRedis creates a new Redis-backed locker.
func NewRedis(client RedisClient) *RedisLocker {
	return &RedisLocker{client: client}
}

type redisUnlocker struct {
	client RedisClient
	key    string
	value  string
}

func (u *redisUnlocker) Unlock(ctx context.Context) error {
	return u.client.Eval(ctx, unlockScript, []string{u.key}, u.value)
}

// TryLock attempts to acquire the lock without blocking.
func (l *RedisLocker) TryLock(ctx context.Context, key string, ttl time.Duration) (Unlocker, bool, error) {
	value := uuid.New().String()

	ok, err := l.client.SetNX(ctx, key, value, ttl)
	if err != nil {
		return nil, false, err
	}

	if !ok {
		return nil, false, nil
	}

	return &redisUnlocker{client: l.client, key: key, value: value}, true, nil
}

// Lock blocks until the lock is acquired or the context is canceled.
func (l *RedisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (Unlocker, error) {
	for {
		unlocker, ok, err := l.TryLock(ctx, key, ttl)
		if err != nil {
			return nil, err
		}

		if ok {
			return unlocker, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(lockRetryInterval):
		}
	}
}
