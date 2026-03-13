package lock

import (
	"context"
	"time"
)

// NoopLocker always succeeds — for single-instance deployments.
type NoopLocker struct{}

// NewNoop creates a new no-op locker.
func NewNoop() *NoopLocker {
	return &NoopLocker{}
}

type noopUnlocker struct{}

func (n *noopUnlocker) Unlock(_ context.Context) error { return nil }

// Lock always succeeds immediately.
func (n *NoopLocker) Lock(_ context.Context, _ string, _ time.Duration) (Unlocker, error) {
	return &noopUnlocker{}, nil
}

// TryLock always succeeds.
func (n *NoopLocker) TryLock(_ context.Context, _ string, _ time.Duration) (Unlocker, bool, error) {
	return &noopUnlocker{}, true, nil
}
