package tasks_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/json"
)

type mockMediaRepo struct {
	getByIDFn func(ctx context.Context, id uint) (*entity.Media, error)
	updateFn  func(ctx context.Context, media *entity.Media) error
}

func (m *mockMediaRepo) Create(_ context.Context, _ *entity.Media) error { return nil }
func (m *mockMediaRepo) GetByID(ctx context.Context, id uint) (*entity.Media, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, errors.New("not implemented")
}
func (m *mockMediaRepo) GetByAttachable(_ context.Context, _ string, _ uint, _ string) ([]*entity.Media, error) {
	return nil, nil
}
func (m *mockMediaRepo) Update(ctx context.Context, media *entity.Media) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, media)
	}
	return nil
}
func (m *mockMediaRepo) Delete(_ context.Context, _ uint) error                       { return nil }
func (m *mockMediaRepo) DeleteByAttachable(_ context.Context, _ string, _ uint) error { return nil }

type mockStorage struct {
	getFn func(ctx context.Context, path string) (io.ReadCloser, error)
	putFn func(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (*storage.FileInfo, error)
}

func (m *mockStorage) Put(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (*storage.FileInfo, error) {
	if m.putFn != nil {
		return m.putFn(ctx, path, reader, size, mimeType)
	}
	return &storage.FileInfo{Path: path}, nil
}
func (m *mockStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	if m.getFn != nil {
		return m.getFn(ctx, path)
	}
	return nil, errors.New("not implemented")
}
func (m *mockStorage) Delete(_ context.Context, _ string) error         { return nil }
func (m *mockStorage) Exists(_ context.Context, _ string) (bool, error) { return false, nil }
func (m *mockStorage) URL(_ context.Context, _ string) (string, error)  { return "", nil }
func (m *mockStorage) TemporaryURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}
func (m *mockStorage) PresignedUploadURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "", nil
}

func TestImageProcessingHandler_ProcessTask_BadPayload(t *testing.T) {
	t.Parallel()

	handler := tasks.NewImageProcessingHandler(&mockLogger{}, &mockMediaRepo{}, &mockStorage{})

	task := asynq.NewTask(tasks.TypeImageProcessing, []byte(`{invalid`))
	err := handler.ProcessTask(context.Background(), task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unmarshal payload")
}

func TestImageProcessingHandler_ProcessTask_MediaNotFound(t *testing.T) {
	t.Parallel()

	repo := &mockMediaRepo{
		getByIDFn: func(_ context.Context, _ uint) (*entity.Media, error) {
			return nil, errors.New("media not found")
		},
	}

	handler := tasks.NewImageProcessingHandler(&mockLogger{}, repo, &mockStorage{})

	payload := tasks.ImageProcessingPayload{MediaID: 999}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(tasks.TypeImageProcessing, data)
	err = handler.ProcessTask(context.Background(), task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "get media")
}

func TestImageProcessingHandler_ProcessTask_NonImage(t *testing.T) {
	t.Parallel()

	repo := &mockMediaRepo{
		getByIDFn: func(_ context.Context, _ uint) (*entity.Media, error) {
			return &entity.Media{ID: 1, Type: entity.MediaTypeDocument}, nil
		},
	}

	handler := tasks.NewImageProcessingHandler(&mockLogger{}, repo, &mockStorage{})

	payload := tasks.ImageProcessingPayload{MediaID: 1}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(tasks.TypeImageProcessing, data)
	err = handler.ProcessTask(context.Background(), task)
	require.NoError(t, err) // Non-image is skipped, not an error
}

func createTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestImageProcessingHandler_ProcessTask_Success(t *testing.T) {
	t.Parallel()

	pngData := createTestPNG(t)
	updated := false

	repo := &mockMediaRepo{
		getByIDFn: func(_ context.Context, _ uint) (*entity.Media, error) {
			return &entity.Media{
				ID:       1,
				Type:     entity.MediaTypeImage,
				MimeType: "image/png",
				Path:     "uploads/test.png",
			}, nil
		},
		updateFn: func(_ context.Context, media *entity.Media) error {
			updated = true
			require.NotNil(t, media.Width)
			require.NotNil(t, media.Height)
			require.Equal(t, 100, *media.Width)
			require.Equal(t, 80, *media.Height)
			return nil
		},
	}

	stor := &mockStorage{
		getFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(pngData)), nil
		},
		putFn: func(_ context.Context, path string, _ io.Reader, _ int64, _ string) (*storage.FileInfo, error) {
			return &storage.FileInfo{Path: path}, nil
		},
	}

	handler := tasks.NewImageProcessingHandler(&mockLogger{}, repo, stor)

	payload := tasks.ImageProcessingPayload{MediaID: 1}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(tasks.TypeImageProcessing, data)
	err = handler.ProcessTask(context.Background(), task)
	require.NoError(t, err)
	require.True(t, updated)
}

func TestImageProcessingHandler_ProcessTask_StorageGetError(t *testing.T) {
	t.Parallel()

	repo := &mockMediaRepo{
		getByIDFn: func(_ context.Context, _ uint) (*entity.Media, error) {
			return &entity.Media{
				ID:       1,
				Type:     entity.MediaTypeImage,
				MimeType: "image/png",
				Path:     "uploads/test.png",
			}, nil
		},
	}

	stor := &mockStorage{
		getFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return nil, errors.New("storage error")
		},
	}

	handler := tasks.NewImageProcessingHandler(&mockLogger{}, repo, stor)

	payload := tasks.ImageProcessingPayload{MediaID: 1}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	task := asynq.NewTask(tasks.TypeImageProcessing, data)
	err = handler.ProcessTask(context.Background(), task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "get original")
}

func TestNewImageProcessingTask(t *testing.T) {
	t.Parallel()

	task, err := tasks.NewImageProcessingTask(42)
	require.NoError(t, err)
	require.Equal(t, tasks.TypeImageProcessing, task.Type())

	var payload tasks.ImageProcessingPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	require.Equal(t, uint(42), payload.MediaID)
}
