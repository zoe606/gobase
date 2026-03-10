package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
)

func TestRefreshTokenRepo_Create(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewRefreshTokenRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "refresh_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	token := &entity.RefreshToken{UserID: 1, Token: "abc"}
	err := r.Create(t.Context(), token)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepo_GetByToken(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewRefreshTokenRepo(gormDB)

		rows := sqlmock.NewRows([]string{"id", "user_id", "token"}).
			AddRow(1, 1, "abc")
		mock.ExpectQuery(`SELECT \* FROM "refresh_tokens" WHERE token = \$1`).
			WithArgs("abc", 1).
			WillReturnRows(rows)

		token, err := r.GetByToken(t.Context(), "abc")
		require.NoError(t, err)
		require.Equal(t, uint(1), token.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewRefreshTokenRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "refresh_tokens" WHERE token = \$1`).
			WithArgs("missing", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByToken(t.Context(), "missing")
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRefreshTokenRepo_DeleteByToken(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewRefreshTokenRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "refresh_tokens" WHERE token = \$1`).
		WithArgs("abc").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.DeleteByToken(t.Context(), "abc")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepo_DeleteByUserID(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewRefreshTokenRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "refresh_tokens" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err := r.DeleteByUserID(t.Context(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
