package asynctx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/asynctx"
)

func TestNewJobContext(t *testing.T) {
	t.Parallel()

	t.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := asynctx.NewJobContext(10 * time.Second)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(10*time.Second), deadline, time.Second)
	})

	t.Run("zero timeout uses default", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := asynctx.NewJobContext(0)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(asynctx.DefaultJobTimeout), deadline, time.Second)
	})
}

func TestFromAsynqContext(t *testing.T) {
	t.Parallel()

	t.Run("adds timeout to existing context", func(t *testing.T) {
		t.Parallel()
		parent := context.Background()
		ctx, cancel := asynctx.FromAsynqContext(parent, 5*time.Second)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(5*time.Second), deadline, time.Second)
	})

	t.Run("zero timeout uses default", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := asynctx.FromAsynqContext(context.Background(), 0)
		defer cancel()

		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(asynctx.DefaultJobTimeout), deadline, time.Second)
	})
}

func TestNewBackgroundContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := asynctx.NewBackgroundContext()
	defer cancel()

	_, ok := ctx.Deadline()
	require.False(t, ok) // no deadline
	require.NoError(t, ctx.Err())

	cancel()
	require.Error(t, ctx.Err())
}
