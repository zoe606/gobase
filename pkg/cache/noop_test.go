package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/cache"
)

func TestNoopCache_Get(t *testing.T) {
	t.Parallel()

	c := cache.NewNoop()
	var result string
	err := c.Get(context.Background(), "key", &result)
	require.ErrorIs(t, err, cache.ErrNotFound)
}

func TestNoopCache_Set(t *testing.T) {
	t.Parallel()

	c := cache.NewNoop()
	err := c.Set(context.Background(), "key", "value", time.Minute)
	require.NoError(t, err)
}

func TestNoopCache_Delete(t *testing.T) {
	t.Parallel()

	c := cache.NewNoop()
	err := c.Delete(context.Background(), "key")
	require.NoError(t, err)
}

func TestNoopCache_Exists(t *testing.T) {
	t.Parallel()

	c := cache.NewNoop()
	exists, err := c.Exists(context.Background(), "key")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestNoopCache_Remember(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		c := cache.NewNoop()
		var result map[string]string

		called := false
		err := c.Remember(context.Background(), "key", time.Minute, &result, func() (interface{}, error) {
			called = true
			return map[string]string{"hello": "world"}, nil
		})

		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, "world", result["hello"])
	})

	t.Run("fn error", func(t *testing.T) {
		t.Parallel()
		c := cache.NewNoop()
		var result string

		err := c.Remember(context.Background(), "key", time.Minute, &result, func() (interface{}, error) {
			return nil, context.Canceled
		})

		require.ErrorIs(t, err, context.Canceled)
	})
}
