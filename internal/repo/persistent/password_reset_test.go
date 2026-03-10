package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
)

func TestPasswordResetRepo_Create(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewPasswordResetRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "password_resets"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	pr := &entity.PasswordReset{UserID: 1, Token: "reset-token"}
	err := r.Create(t.Context(), pr)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetRepo_GetByToken(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewPasswordResetRepo(gormDB)

		rows := sqlmock.NewRows([]string{"id", "user_id", "token"}).
			AddRow(1, 1, "reset-token")
		mock.ExpectQuery(`SELECT \* FROM "password_resets" WHERE token = \$1`).
			WithArgs("reset-token", 1).
			WillReturnRows(rows)

		pr, err := r.GetByToken(t.Context(), "reset-token")
		require.NoError(t, err)
		require.Equal(t, uint(1), pr.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewPasswordResetRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "password_resets" WHERE token = \$1`).
			WithArgs("missing", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByToken(t.Context(), "missing")
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestPasswordResetRepo_MarkAsUsed(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewPasswordResetRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "password_resets" SET "used_at"=\$1 WHERE id = \$2`).
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.MarkAsUsed(t.Context(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPasswordResetRepo_DeleteByUserID(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewPasswordResetRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "password_resets" WHERE user_id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := r.DeleteByUserID(t.Context(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
