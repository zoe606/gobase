package media_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mediadto "go-boilerplate/internal/dto/media"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/internal/usecase/media"
)

func TestUpload(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		req mediadto.UploadRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider)
		wantErr   error
	}{
		{
			name: "success - image upload",
			args: args{
				ctx: context.Background(),
				req: mediadto.UploadRequest{
					File:           bytes.NewReader([]byte("test image content")),
					Filename:       "test.jpg",
					Size:           18,
					MimeType:       "image/jpeg",
					AttachableType: "users",
					AttachableID:   1,
					Collection:     "avatar",
				},
			},
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), int64(18), "image/jpeg").
					Return(&storage.FileInfo{
						Path:     "users/avatar/1/2024/01/uuid.jpg",
						Size:     18,
						MimeType: "image/jpeg",
						Hash:     "abc123",
					}, nil)
				mediaRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "success - document upload",
			args: args{
				ctx: context.Background(),
				req: mediadto.UploadRequest{
					File:           bytes.NewReader([]byte("test pdf content")),
					Filename:       "document.pdf",
					Size:           16,
					MimeType:       "application/pdf",
					AttachableType: "posts",
					AttachableID:   5,
					Collection:     "attachments",
				},
			},
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), int64(16), "application/pdf").
					Return(&storage.FileInfo{
						Path:     "posts/attachments/5/2024/01/uuid.pdf",
						Size:     16,
						MimeType: "application/pdf",
						Hash:     "def456",
					}, nil)
				mediaRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "error - file too large",
			args: args{
				ctx: context.Background(),
				req: mediadto.UploadRequest{
					File:           bytes.NewReader([]byte("large content")),
					Filename:       "large.jpg",
					Size:           20 * 1024 * 1024, // 20MB
					MimeType:       "image/jpeg",
					AttachableType: "users",
					AttachableID:   1,
					Collection:     "avatar",
				},
			},
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				// No mocks needed - should fail before reaching storage
			},
			wantErr: media.ErrFileTooLarge,
		},
		{
			name: "error - invalid mime type",
			args: args{
				ctx: context.Background(),
				req: mediadto.UploadRequest{
					File:           bytes.NewReader([]byte("exe content")),
					Filename:       "malware.exe",
					Size:           100,
					MimeType:       "application/x-msdownload",
					AttachableType: "users",
					AttachableID:   1,
					Collection:     "documents",
				},
			},
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				// No mocks needed - should fail before reaching storage
			},
			wantErr: media.ErrInvalidMimeType,
		},
		{
			name: "error - storage put failed",
			args: args{
				ctx: context.Background(),
				req: mediadto.UploadRequest{
					File:           bytes.NewReader([]byte("test content")),
					Filename:       "test.jpg",
					Size:           12,
					MimeType:       "image/jpeg",
					AttachableType: "users",
					AttachableID:   1,
					Collection:     "avatar",
				},
			},
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), int64(12), "image/jpeg").
					Return(nil, errors.New("storage error"))
			},
			wantErr: errors.New("storage error"),
		},
		{
			name: "error - db create failed with cleanup",
			args: args{
				ctx: context.Background(),
				req: mediadto.UploadRequest{
					File:           bytes.NewReader([]byte("test content")),
					Filename:       "test.jpg",
					Size:           12,
					MimeType:       "image/jpeg",
					AttachableType: "users",
					AttachableID:   1,
					Collection:     "avatar",
				},
			},
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					Put(gomock.Any(), gomock.Any(), gomock.Any(), int64(12), "image/jpeg").
					Return(&storage.FileInfo{
						Path:     "users/avatar/1/2024/01/uuid.jpg",
						Size:     12,
						MimeType: "image/jpeg",
						Hash:     "abc123",
					}, nil)
				mediaRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
				// Should cleanup the uploaded file
				storageProvider.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMediaRepo := NewMockMediaRepo(ctrl)
			mockStorage := NewMockStorageProvider(ctrl)
			mockLogger := NewMockLogger()

			tt.setupMock(mockMediaRepo, mockStorage)

			uc := media.New(mockMediaRepo, mockStorage, nil, mockLogger, "s3", 10*1024*1024) // 10MB max
			result, err := uc.Upload(tt.args.ctx, tt.args.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, media.ErrFileTooLarge) ||
					errors.Is(tt.wantErr, media.ErrInvalidMimeType) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.Filename)
			assert.Equal(t, tt.args.req.MimeType, result.MimeType)
			assert.Equal(t, tt.args.req.Size, result.Size)
		})
	}
}
