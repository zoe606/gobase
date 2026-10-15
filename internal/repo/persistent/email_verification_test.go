package persistent_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
)

func TestEmailVerificationRepo_Create(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewEmailVerificationRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "email_verifications"`).
		WithArgs(1, fmt.Sprintf("%x", sha256.Sum256([]byte("abc"))), sqlmock.AnyArg(), nil, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	ev := &entity.EmailVerification{UserID: 1, Token: "abc"}
	err := r.Create(t.Context(), ev)
	require.NoError(t, err)
	require.Equal(t, "abc", ev.Token)
	require.Equal(t, uint(1), ev.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEmailVerificationRepo_GetByToken(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewEmailVerificationRepo(gormDB)

		rows := sqlmock.NewRows([]string{"id", "user_id", "token"}).
			AddRow(1, 1, "abc")
		mock.ExpectQuery(`SELECT \* FROM "email_verifications" WHERE token = \$1`).
			WithArgs(fmt.Sprintf("%x", sha256.Sum256([]byte("abc"))), 1).
			WillReturnRows(rows)

		ev, err := r.GetByToken(t.Context(), "abc")
		require.NoError(t, err)
		require.Equal(t, uint(1), ev.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewEmailVerificationRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "email_verifications" WHERE token = \$1`).
			WithArgs(fmt.Sprintf("%x", sha256.Sum256([]byte("missing"))), 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByToken(t.Context(), "missing")
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEmailVerificationRepo_GetLatestByUserID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewEmailVerificationRepo(gormDB)

		rows := sqlmock.NewRows([]string{"id", "user_id", "token"}).
			AddRow(2, 1, "latest")
		mock.ExpectQuery(`SELECT \* FROM "email_verifications" WHERE user_id = \$1 ORDER BY created_at DESC`).
			WithArgs(1, 1).
			WillReturnRows(rows)

		ev, err := r.GetLatestByUserID(t.Context(), 1)
		require.NoError(t, err)
		require.Equal(t, uint(2), ev.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewEmailVerificationRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "email_verifications" WHERE user_id = \$1 ORDER BY created_at DESC`).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetLatestByUserID(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestEmailVerificationRepo_MarkAsUsed(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewEmailVerificationRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "email_verifications" SET "used_at"=\$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.MarkAsUsed(t.Context(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestEmailVerificationRepo_DeleteByUserID(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewEmailVerificationRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "email_verifications" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err := r.DeleteByUserID(t.Context(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
