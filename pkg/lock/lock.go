// Package lock provides a distributed lock abstraction.
package lock

import (
	"context"
	"errors"
	"time"
)

// ErrLockNotAcquired is returned when a lock cannot be obtained.
var ErrLockNotAcquired = errors.New("lock: not acquired")

// Unlocker releases a held lock.
type Unlocker interface {
	Unlock(ctx context.Context) error
}

// Locker provides distributed locking.
type Locker interface {
	// Lock blocks until the lock is acquired or ctx is canceled.
	Lock(ctx context.Context, key string, ttl time.Duration) (Unlocker, error)

	// TryLock attempts to acquire the lock without blocking.
	// Returns (unlocker, true, nil) on success, (nil, false, nil) if lock is held.
	TryLock(ctx context.Context, key string, ttl time.Duration) (Unlocker, bool, error)
}
