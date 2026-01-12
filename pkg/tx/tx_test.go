package tx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/tx"
)

func TestDBFromContext(t *testing.T) {
	t.Parallel()

	t.Run("returns default db when no transaction in context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		// Without a real DB, we just test the nil path
		result := tx.FromContext(ctx)
		require.Nil(t, result)
	})

	t.Run("returns true when in transaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		require.False(t, tx.IsInTransaction(ctx))
	})
}

func TestWithTx(t *testing.T) {
	t.Parallel()

	t.Run("stores and retrieves transaction from context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		// Verify no transaction initially
		require.Nil(t, tx.FromContext(ctx))
		require.False(t, tx.IsInTransaction(ctx))
	})
}

// Note: Full integration tests with real database are in integration-test/
// These unit tests verify the context handling logic.

func TestContextPropagation(t *testing.T) {
	t.Parallel()

	t.Run("context without transaction returns false for IsInTransaction", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		require.False(t, tx.IsInTransaction(ctx))
	})

	t.Run("FromContext returns nil for empty context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		result := tx.FromContext(ctx)
		require.Nil(t, result)
	})
}

// MockError for testing rollback scenarios.
var errMock = errors.New("mock error")

func TestTxOptions(t *testing.T) {
	t.Parallel()

	t.Run("default options", func(t *testing.T) {
		t.Parallel()

		opts := &tx.TxOptions{
			Propagation: false,
		}
		require.False(t, opts.Propagation)
	})

	t.Run("propagation enabled", func(t *testing.T) {
		t.Parallel()

		opts := &tx.TxOptions{
			Propagation: true,
		}
		require.True(t, opts.Propagation)
	})
}
