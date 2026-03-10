package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
)

func TestProfileRepo_Create(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewProfileRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "profiles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	profile := &entity.Profile{UserID: 1, Bio: "Hello"}
	err := r.Create(t.Context(), profile)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileRepo_GetByUserID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewProfileRepo(gormDB)

		rows := sqlmock.NewRows([]string{"id", "user_id", "bio"}).
			AddRow(1, 1, "Hello")
		mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE user_id = \$1`).
			WithArgs(1, 1).
			WillReturnRows(rows)

		profile, err := r.GetByUserID(t.Context(), 1)
		require.NoError(t, err)
		require.Equal(t, uint(1), profile.UserID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewProfileRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "profiles" WHERE user_id = \$1`).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByUserID(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestProfileRepo_Update(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewProfileRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "profiles" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	profile := &entity.Profile{ID: 1, UserID: 1, Bio: "Updated"}
	err := r.Update(t.Context(), profile)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestProfileRepo_Upsert(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewProfileRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "profiles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	profile := &entity.Profile{UserID: 1, Bio: "Hello"}
	err := r.Upsert(t.Context(), profile)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
