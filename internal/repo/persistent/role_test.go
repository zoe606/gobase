package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
)

func TestRoleRepo_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewRoleRepo(gormDB)

		mock.MatchExpectationsInOrder(false)

		roleRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "admin")
		mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
			WithArgs(1, 1).
			WillReturnRows(roleRows)

		permRows := sqlmock.NewRows([]string{"id", "name"})
		mock.ExpectQuery(`SELECT .* FROM "permissions"`).
			WillReturnRows(permRows)

		mock.ExpectQuery(`SELECT .* FROM "role_permissions"`).
			WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}))

		role, err := r.GetByID(t.Context(), 1)
		require.NoError(t, err)
		require.Equal(t, "admin", role.Name)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewRoleRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "roles" WHERE "roles"\."id" = \$1`).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByID(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRoleRepo_GetByName(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewRoleRepo(gormDB)

		mock.MatchExpectationsInOrder(false)

		roleRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "user")
		mock.ExpectQuery(`SELECT \* FROM "roles" WHERE name = \$1`).
			WithArgs("user", 1).
			WillReturnRows(roleRows)

		permRows := sqlmock.NewRows([]string{"id", "name"})
		mock.ExpectQuery(`SELECT .* FROM "permissions"`).
			WillReturnRows(permRows)

		mock.ExpectQuery(`SELECT .* FROM "role_permissions"`).
			WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}))

		role, err := r.GetByName(t.Context(), "user")
		require.NoError(t, err)
		require.Equal(t, "user", role.Name)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewRoleRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "roles" WHERE name = \$1`).
			WithArgs("missing", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByName(t.Context(), "missing")
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRoleRepo_GetAll(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewRoleRepo(gormDB)

	mock.MatchExpectationsInOrder(false)

	roleRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "admin").
		AddRow(2, "user")
	mock.ExpectQuery(`SELECT \* FROM "roles"`).
		WillReturnRows(roleRows)

	permRows := sqlmock.NewRows([]string{"id", "name"})
	mock.ExpectQuery(`SELECT .* FROM "permissions"`).
		WillReturnRows(permRows)

	mock.ExpectQuery(`SELECT .* FROM "role_permissions"`).
		WillReturnRows(sqlmock.NewRows([]string{"role_id", "permission_id"}))

	roles, err := r.GetAll(t.Context())
	require.NoError(t, err)
	require.Len(t, roles, 2)
}
