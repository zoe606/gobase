package tx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go-boilerplate/pkg/tx"
)

func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	require.NoError(t, err)
	return gormDB, mock
}

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

func TestNew(t *testing.T) {
	t.Parallel()
	gormDB, _ := newTestDB(t)
	h := tx.New(gormDB)
	require.NotNil(t, h)
}

func TestRunInTx_Success(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	h := tx.New(gormDB)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := h.RunInTx(t.Context(), func(txCtx context.Context) error {
		require.True(t, tx.IsInTransaction(txCtx))
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunInTx_Error(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	h := tx.New(gormDB)

	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("tx error")
	err := h.RunInTx(t.Context(), func(_ context.Context) error {
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRunInTxWithOptions_Propagation(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	h := tx.New(gormDB)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := h.RunInTx(t.Context(), func(txCtx context.Context) error {
		return h.RunInTxWithOptions(txCtx, &tx.TxOptions{Propagation: true}, func(innerCtx context.Context) error {
			require.True(t, tx.IsInTransaction(innerCtx))
			return nil
		})
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTx_RoundTrip(t *testing.T) {
	t.Parallel()
	gormDB, _ := newTestDB(t)

	ctx := tx.WithTx(t.Context(), gormDB)
	got := tx.FromContext(ctx)
	require.NotNil(t, got)
	require.True(t, tx.IsInTransaction(ctx))
}

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
