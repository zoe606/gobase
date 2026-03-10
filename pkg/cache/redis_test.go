package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/cache"
)

type testData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type mockMetrics struct {
	hits   int
	misses int
}

func (m *mockMetrics) RecordHit()  { m.hits++ }
func (m *mockMetrics) RecordMiss() { m.misses++ }

func newTestRedisCache(t *testing.T) *cache.RedisCache {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return cache.NewRedis(client, "test:")
}

func TestRedisCache_SetAndGet(t *testing.T) {
	t.Parallel()

	c := newTestRedisCache(t)
	ctx := context.Background()

	err := c.Set(ctx, "user:1", testData{Name: "John", Age: 30}, time.Minute)
	require.NoError(t, err)

	var result testData
	err = c.Get(ctx, "user:1", &result)
	require.NoError(t, err)
	require.Equal(t, "John", result.Name)
	require.Equal(t, 30, result.Age)
}

func TestRedisCache_Get_Miss(t *testing.T) {
	t.Parallel()

	c := newTestRedisCache(t)

	var result string
	err := c.Get(context.Background(), "nonexistent", &result)
	require.ErrorIs(t, err, cache.ErrNotFound)
}

func TestRedisCache_Delete(t *testing.T) {
	t.Parallel()

	c := newTestRedisCache(t)
	ctx := context.Background()

	err := c.Set(ctx, "key", "value", time.Minute)
	require.NoError(t, err)

	err = c.Delete(ctx, "key")
	require.NoError(t, err)

	var result string
	err = c.Get(ctx, "key", &result)
	require.ErrorIs(t, err, cache.ErrNotFound)
}

func TestRedisCache_Exists(t *testing.T) {
	t.Parallel()

	c := newTestRedisCache(t)
	ctx := context.Background()

	exists, err := c.Exists(ctx, "key")
	require.NoError(t, err)
	require.False(t, exists)

	err = c.Set(ctx, "key", "value", time.Minute)
	require.NoError(t, err)

	exists, err = c.Exists(ctx, "key")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestRedisCache_Remember_Miss(t *testing.T) {
	t.Parallel()

	c := newTestRedisCache(t)
	ctx := context.Background()

	var result testData
	called := false
	err := c.Remember(ctx, "user:1", time.Minute, &result, func() (interface{}, error) {
		called = true
		return testData{Name: "Jane", Age: 25}, nil
	})

	require.NoError(t, err)
	require.True(t, called)
	require.Equal(t, "Jane", result.Name)

	// Verify it was cached
	var cached testData
	err = c.Get(ctx, "user:1", &cached)
	require.NoError(t, err)
	require.Equal(t, "Jane", cached.Name)
}

func TestRedisCache_Remember_Hit(t *testing.T) {
	t.Parallel()

	c := newTestRedisCache(t)
	ctx := context.Background()

	// Pre-set the value
	err := c.Set(ctx, "user:1", testData{Name: "Cached", Age: 99}, time.Minute)
	require.NoError(t, err)

	var result testData
	called := false
	err = c.Remember(ctx, "user:1", time.Minute, &result, func() (interface{}, error) {
		called = true
		return testData{Name: "Fresh", Age: 1}, nil
	})

	require.NoError(t, err)
	require.False(t, called) // fn should NOT be called
	require.Equal(t, "Cached", result.Name)
}

func TestRedisCache_MetricsHook(t *testing.T) {
	t.Parallel()

	c := newTestRedisCache(t)
	ctx := context.Background()

	metrics := &mockMetrics{}
	c.SetMetricsHook(metrics)

	// Cache miss
	var result string
	_ = c.Get(ctx, "missing", &result)
	require.Equal(t, 1, metrics.misses)
	require.Equal(t, 0, metrics.hits)

	// Cache hit
	err := c.Set(ctx, "key", "value", time.Minute)
	require.NoError(t, err)
	_ = c.Get(ctx, "key", &result)
	require.Equal(t, 1, metrics.hits)
}
