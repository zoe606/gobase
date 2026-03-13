package lock_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/lock"
)

type mockRedisClient struct {
	mu    sync.Mutex
	store map[string]string
}

func newMockRedis() *mockRedisClient {
	return &mockRedisClient{store: make(map[string]string)}
}

func (m *mockRedisClient) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.store[key]; exists {
		return false, nil
	}

	m.store[key] = value.(string)

	return true, nil
}

func (m *mockRedisClient) Eval(_ context.Context, _ string, keys []string, args ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(keys) == 0 || len(args) == 0 {
		return lock.ErrLockNotAcquired
	}

	key := keys[0]

	val, ok := args[0].(string)
	if !ok {
		return lock.ErrLockNotAcquired
	}

	stored, exists := m.store[key]
	if !exists || stored != val {
		return lock.ErrLockNotAcquired
	}

	delete(m.store, key)

	return nil
}

func TestRedisLocker_TryLock_Success(t *testing.T) {
	client := newMockRedis()
	locker := lock.NewRedis(client)
	ctx := context.Background()

	unlocker, ok, err := locker.TryLock(ctx, "my-key", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)
	require.NotNil(t, unlocker)

	// Unlock should succeed.
	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}

func TestRedisLocker_TryLock_AlreadyLocked(t *testing.T) {
	client := newMockRedis()
	locker := lock.NewRedis(client)
	ctx := context.Background()

	// Acquire the lock.
	unlocker, ok, err := locker.TryLock(ctx, "my-key", 10*time.Second)
	require.NoError(t, err)
	require.True(t, ok)

	// Second attempt should fail.
	unlocker2, ok2, err2 := locker.TryLock(ctx, "my-key", 10*time.Second)
	require.NoError(t, err2)
	assert.False(t, ok2)
	assert.Nil(t, unlocker2)

	// Cleanup.
	require.NoError(t, unlocker.Unlock(ctx))
}

func TestRedisLocker_Lock_Success(t *testing.T) {
	client := newMockRedis()
	locker := lock.NewRedis(client)
	ctx := context.Background()

	unlocker, err := locker.Lock(ctx, "my-key", 10*time.Second)
	require.NoError(t, err)
	require.NotNil(t, unlocker)

	err = unlocker.Unlock(ctx)
	assert.NoError(t, err)
}

func TestRedisLocker_Lock_ContextCancelled(t *testing.T) {
	client := newMockRedis()
	locker := lock.NewRedis(client)

	// Pre-acquire the lock so Lock() will block.
	bgCtx := context.Background()

	_, ok, err := locker.TryLock(bgCtx, "held-key", 10*time.Second)
	require.NoError(t, err)
	require.True(t, ok)

	// Create a context with a very short timeout.
	ctx, cancel := context.WithTimeout(bgCtx, 100*time.Millisecond)
	defer cancel()

	_, err = locker.Lock(ctx, "held-key", 10*time.Second)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestRedisLocker_ImplementsLockerInterface(t *testing.T) {
	var _ lock.Locker = lock.NewRedis(newMockRedis())
}

func TestRedisLocker_Unlock_WrongOwner(t *testing.T) {
	client := newMockRedis()
	locker := lock.NewRedis(client)
	ctx := context.Background()

	// Acquire the lock.
	unlocker, ok, err := locker.TryLock(ctx, "owner-key", 10*time.Second)
	require.NoError(t, err)
	require.True(t, ok)

	// Manually tamper with the stored value to simulate a different owner.
	client.mu.Lock()
	client.store["owner-key"] = "different-owner-value"
	client.mu.Unlock()

	// Unlock should fail because the owner value doesn't match.
	err = unlocker.Unlock(ctx)
	assert.ErrorIs(t, err, lock.ErrLockNotAcquired)
}
