package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/ratelimiter"
)

func newTestAdapter(t *testing.T) (*ratelimiter.RedisAdapter, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	return ratelimiter.NewRedisAdapter(client), mr
}

func TestRedisAdapter_GetSetDelete(t *testing.T) {
	t.Parallel()

	adapter, _ := newTestAdapter(t)

	// Get non-existent key returns nil (not an error).
	val, err := adapter.Get("missing")
	require.NoError(t, err)
	assert.Nil(t, val)

	// Set stores a value with expiration.
	err = adapter.Set("key1", []byte("value1"), time.Minute)
	require.NoError(t, err)

	// Get existing key returns the value.
	val, err = adapter.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, []byte("value1"), val)

	// Delete removes the key.
	err = adapter.Delete("key1")
	require.NoError(t, err)

	val, err = adapter.Get("key1")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestRedisAdapter_Expiration(t *testing.T) {
	t.Parallel()

	adapter, mr := newTestAdapter(t)

	err := adapter.Set("expiring", []byte("data"), 2*time.Second)
	require.NoError(t, err)

	val, err := adapter.Get("expiring")
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), val)

	// Fast-forward time in miniredis to expire the key.
	mr.FastForward(3 * time.Second)

	val, err = adapter.Get("expiring")
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestRedisAdapter_Reset(t *testing.T) {
	t.Parallel()

	adapter, _ := newTestAdapter(t)

	_ = adapter.Set("a", []byte("1"), time.Minute)
	_ = adapter.Set("b", []byte("2"), time.Minute)

	err := adapter.Reset()
	require.NoError(t, err)

	val, _ := adapter.Get("a")
	assert.Nil(t, val)

	val, _ = adapter.Get("b")
	assert.Nil(t, val)
}

func TestRedisAdapter_Close(t *testing.T) {
	t.Parallel()

	adapter, _ := newTestAdapter(t)

	err := adapter.Close()
	require.NoError(t, err)
}
