package persistent_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
)

func TestMediaRepo_Create(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewMediaRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "media"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	media := &entity.Media{Filename: "test.jpg", OriginalName: "test.jpg", MimeType: "image/jpeg", Size: 1024, Disk: "s3", Path: "test/test.jpg", Type: entity.MediaTypeImage}
	err := r.Create(t.Context(), media)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaRepo_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewMediaRepo(gormDB)

		rows := sqlmock.NewRows([]string{"id", "filename", "mime_type"}).
			AddRow(1, "test.jpg", "image/jpeg")
		mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 AND "media"\."deleted_at" IS NULL`).
			WithArgs(1, 1).
			WillReturnRows(rows)

		media, err := r.GetByID(t.Context(), 1)
		require.NoError(t, err)
		require.Equal(t, uint(1), media.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewMediaRepo(gormDB)

		mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 AND "media"\."deleted_at" IS NULL`).
			WithArgs(999, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := r.GetByID(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMediaRepo_GetByAttachable(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewMediaRepo(gormDB)

	rows := sqlmock.NewRows([]string{"id", "filename"}).
		AddRow(1, "a.jpg").
		AddRow(2, "b.jpg")
	mock.ExpectQuery(`SELECT \* FROM "media" WHERE \(attachable_type = \$1 AND attachable_id = \$2\) AND "media"\."deleted_at" IS NULL`).
		WithArgs("users", 1).
		WillReturnRows(rows)

	media, err := r.GetByAttachable(t.Context(), "users", 1, "")
	require.NoError(t, err)
	require.Len(t, media, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaRepo_Update(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewMediaRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "media" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	media := &entity.Media{ID: 1, Filename: "updated.jpg", OriginalName: "updated.jpg", MimeType: "image/jpeg", Disk: "s3", Path: "test", Type: entity.MediaTypeImage}
	err := r.Update(t.Context(), media)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMediaRepo_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewMediaRepo(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "media" SET "deleted_at"=`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := r.Delete(t.Context(), 1)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		gormDB, mock := newTestDB(t)
		r := persistent.NewMediaRepo(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "media" SET "deleted_at"=`).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()

		err := r.Delete(t.Context(), 999)
		require.ErrorIs(t, err, repo.ErrNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMediaRepo_DeleteByAttachable(t *testing.T) {
	t.Parallel()
	gormDB, mock := newTestDB(t)
	r := persistent.NewMediaRepo(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "media" SET "deleted_at"=`).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	err := r.DeleteByAttachable(t.Context(), "users", 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
