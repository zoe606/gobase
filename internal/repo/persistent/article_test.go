package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
	"go-boilerplate/pkg/pagination"
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

func TestArticleRepo_Create(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewArticleRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "articles"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	article := &entity.Article{Title: "Test", Slug: "test", UserID: 1}
	err := r.Create(t.Context(), article)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleRepo_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewArticleRepo(gormDB)

		rows := sqlmock.NewRows([]string{"id", "title", "slug", "user_id"}).
			AddRow(1, "Test", "test", 1)
		mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"\."id" = \$1`).
			WithArgs(1, 1). // id + LIMIT
			WillReturnRows(rows)

		article, err := r.GetByID(t.Context(), 1)
		require.NoError(t, err)
		require.Equal(t, uint(1), article.ID)
		require.Equal(t, "Test", article.Title)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewArticleRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "articles" WHERE "articles"\."id" = \$1`).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByID(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestArticleRepo_List(t *testing.T) {
	t.Parallel()

	gormDB, mock := newTestDB(t)
	r := persistent.NewArticleRepo(gormDB)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "articles"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "title", "slug", "user_id"}).
		AddRow(1, "Test", "test", 1)
	mock.ExpectQuery(`SELECT \* FROM "articles"`).
		WillReturnRows(rows)

	req := articledto.ListRequest{
		Params: pagination.Params{Page: 1, Limit: 10},
	}
	articles, total, err := r.List(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, articles, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleRepo_Update(t *testing.T) {
	t.Parallel()

	gormDB, mock := newTestDB(t)
	r := persistent.NewArticleRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "articles" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	article := &entity.Article{ID: 1, Title: "Updated", Slug: "updated", UserID: 1}
	err := r.Update(t.Context(), article)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestArticleRepo_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewArticleRepo(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "articles" WHERE "articles"\."id" = \$1`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := r.Delete(t.Context(), 1)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewArticleRepo(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "articles" WHERE "articles"\."id" = \$1`).
			WithArgs(999).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := r.Delete(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
