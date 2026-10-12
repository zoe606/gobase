package app

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"go-boilerplate/config"
	"go-boilerplate/pkg/cache"
	"go-boilerplate/pkg/logger"
)

func redisTestConfig(t *testing.T) (*config.Config, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	host, port, err := net.SplitHostPort(mr.Addr())
	require.NoError(t, err)
	portNumber, err := strconv.Atoi(port)
	require.NoError(t, err)
	mr.RequireAuth("test-password")
	mr.Select(2)
	return &config.Config{
		Redis: config.Redis{Host: host, Port: portNumber, Password: "test-password", DB: 2},
		Cache: config.Cache{Prefix: "app:test:"},
	}, mr
}

func TestRedisStoresLifecycle(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name         string
		cacheEnabled bool
		store        string
	}{
		{"disabled cache and memory limiter", false, "memory"},
		{"cache and memory limiter", true, "memory"},
		{"disabled cache and Redis limiter", false, "redis"},
		{"cache and Redis limiter", true, "redis"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg, mr := redisTestConfig(t)
			cfg.Cache.Enabled = test.cacheEnabled
			cfg.RateLimit.Store = test.store
			if !test.cacheEnabled && test.store == "memory" {
				mr.Close()
			}
			appCache, store, closeRedis := initRedisStores(cfg, logger.New("error"))
			t.Cleanup(closeRedis)

			require.NoError(t, appCache.Set(t.Context(), "value", "cached", time.Minute))
			var value string
			if test.cacheEnabled {
				require.NoError(t, appCache.Get(t.Context(), "value", &value))
				require.Equal(t, "cached", value)
				stored, err := mr.Get("app:test:value")
				require.NoError(t, err)
				require.Equal(t, `"cached"`, stored)
			} else {
				require.ErrorIs(t, appCache.Get(t.Context(), "value", &value), cache.ErrNotFound)
			}
			if test.store == "redis" {
				require.NotNil(t, store)
				require.NoError(t, store.Set("ratelimit:test", []byte("limit-state"), time.Minute))
				stored, err := store.Get("ratelimit:test")
				require.NoError(t, err)
				require.Equal(t, []byte("limit-state"), stored)
			} else {
				require.Nil(t, store)
			}

			closeRedis()
			closeRedis()
			if test.cacheEnabled {
				require.ErrorIs(t, appCache.Set(t.Context(), "after-close", "value", time.Minute), goredis.ErrClosed)
			} else {
				require.NoError(t, appCache.Set(t.Context(), "after-close", "value", time.Minute))
			}
			if store != nil {
				_, err := store.Get("ratelimit:test")
				require.ErrorIs(t, err, goredis.ErrClosed)
			}
		})
	}
}

func TestRedisStoresUseSeparateClients(t *testing.T) {
	t.Parallel()
	cfg, _ := redisTestConfig(t)
	cfg.Cache.Enabled = true
	cfg.RateLimit.Store = "redis"
	appCache, store, closeRedis := initRedisStores(cfg, logger.New("error"))
	t.Cleanup(closeRedis)

	require.NoError(t, store.Close())
	require.NoError(t, appCache.Set(t.Context(), "value", "cached", time.Minute))
	closeRedis()
	require.ErrorIs(t, appCache.Set(t.Context(), "after-close", "value", time.Minute), goredis.ErrClosed)
}

func TestRedisStoresCleanupAfterRedisFailure(t *testing.T) {
	t.Parallel()
	cfg, mr := redisTestConfig(t)
	cfg.Cache.Enabled = true
	cfg.RateLimit.Store = "redis"
	appCache, store, closeRedis := initRedisStores(cfg, logger.New("error"))
	t.Cleanup(closeRedis)

	require.NoError(t, appCache.Set(t.Context(), "value", "cached", time.Minute))
	require.NoError(t, store.Set("ratelimit:test", []byte("limit-state"), time.Minute))
	mr.Close()
	require.Error(t, appCache.Set(t.Context(), "during-outage", "value", time.Minute))
	_, err := store.Get("ratelimit:test")
	require.Error(t, err)
	closeRedis()
	require.ErrorIs(t, appCache.Set(t.Context(), "after-close", "value", time.Minute), goredis.ErrClosed)
	_, err = store.Get("ratelimit:test")
	require.ErrorIs(t, err, goredis.ErrClosed)
}
