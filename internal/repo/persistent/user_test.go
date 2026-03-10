package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
)

func TestUserRepo_Create(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewUserRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	user := &entity.User{Email: "test@example.com", Password: "hashed", Name: "Test", RoleID: 1}
	err := r.Create(t.Context(), user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewUserRepo(gormDB)

		mock.MatchExpectationsInOrder(false)

		userRows := sqlmock.NewRows([]string{"id", "email", "name", "role_id"}).
			AddRow(1, "test@example.com", "Test", 1)
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1`).
			WithArgs(1, 1).
			WillReturnRows(userRows)

		roleRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "user")
		mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
			WithArgs(1).
			WillReturnRows(roleRows)

		mock.ExpectQuery(`SELECT .* FROM "role_permissions"`).
			WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}))

		user, err := r.GetByID(t.Context(), 1)
		require.NoError(t, err)
		require.Equal(t, uint(1), user.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewUserRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = \$1`).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByID(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_GetByEmail(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewUserRepo(gormDB)

		mock.MatchExpectationsInOrder(false)

		userRows := sqlmock.NewRows([]string{"id", "email", "name", "role_id"}).
			AddRow(1, "test@example.com", "Test", 1)
		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1`).
			WithArgs("test@example.com", 1).
			WillReturnRows(userRows)

		roleRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "user")
		mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
			WithArgs(1).
			WillReturnRows(roleRows)

		mock.ExpectQuery(`SELECT .* FROM "role_permissions"`).
			WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}))

		user, err := r.GetByEmail(t.Context(), "test@example.com")
		require.NoError(t, err)
		require.Equal(t, "test@example.com", user.Email)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewUserRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "users" WHERE email = \$1`).
			WithArgs("missing@test.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByEmail(t.Context(), "missing@test.com")
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_EmailExists(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewUserRepo(gormDB)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE email = \$1`).
			WithArgs("test@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		exists, err := r.EmailExists(t.Context(), "test@example.com")
		require.NoError(t, err)
		require.True(t, exists)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("does not exist", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewUserRepo(gormDB)

		mock.ExpectQuery(`SELECT count\(\*\) FROM "users" WHERE email = \$1`).
			WithArgs("new@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		exists, err := r.EmailExists(t.Context(), "new@example.com")
		require.NoError(t, err)
		require.False(t, exists)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUserRepo_Update(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewUserRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "users" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	user := &entity.User{ID: 1, Email: "updated@example.com", Password: "hash", Name: "Updated", RoleID: 1}
	err := r.Update(t.Context(), user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
