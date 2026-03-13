package lock_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/lock"
)

func TestNoopLocker_Lock(t *testing.T) {
	locker := lock.NewNoop()
	ctx := context.Background()

	unlocker, err := locker.Lock(ctx, "test-key", 5*time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlocker)

	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}

func TestNoopLocker_TryLock(t *testing.T) {
	locker := lock.NewNoop()
	ctx := context.Background()

	unlocker, ok, err := locker.TryLock(ctx, "test-key", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)
	require.NotNil(t, unlocker)

	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}

func TestNoopLocker_ImplementsLockerInterface(t *testing.T) {
	var _ lock.Locker = lock.NewNoop()
}
