package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
	"go-boilerplate/pkg/pagination"
)

func TestTranslationRepo_Store(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.New(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "translations"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	tr := &entity.Translation{Source: "en", Destination: "es", Original: "hello", Translation: "hola"}
	err := r.Store(t.Context(), tr)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepo_GetHistory(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.New(gormDB)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "translations"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "source", "destination", "original", "translation"}).
		AddRow(1, "en", "es", "hello", "hola")
	mock.ExpectQuery(`SELECT \* FROM "translations"`).
		WillReturnRows(rows)

	req := translationdto.HistoryRequest{
		Params: pagination.Params{Page: 1, Limit: 10},
	}
	translations, total, err := r.GetHistory(t.Context(), req)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, translations, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepo_GetByID(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.New(gormDB)

	rows := sqlmock.NewRows([]string{"id", "source", "destination"}).
		AddRow(1, "en", "es")
	mock.ExpectQuery(`SELECT \* FROM "translations" WHERE "translations"\."id" = \$1`).
		WithArgs(1, 1).
		WillReturnRows(rows)

	tr, err := r.GetByID(t.Context(), 1)
	require.NoError(t, err)
	require.Equal(t, uint(1), tr.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTranslationRepo_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.New(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "translations" SET "deleted_at"=`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := r.Delete(t.Context(), 1)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.New(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "translations" SET "deleted_at"=`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := r.Delete(t.Context(), 999)
		require.Error(t, err)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
