package media_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/media"
)

func TestGetURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		media     *entity.Media
		variant   string
		setupMock func(storageProvider *MockStorageProvider)
		wantURL   string
		wantErr   error
	}{
		{
			name: "success - original",
			media: &entity.Media{
				ID:   1,
				Path: "users/avatar/1/2024/01/test.jpg",
			},
			variant: "",
			setupMock: func(storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					TemporaryURL(gomock.Any(), "users/avatar/1/2024/01/test.jpg", time.Hour).
					Return("https://s3.example.com/signed-url", nil)
			},
			wantURL: "https://s3.example.com/signed-url",
			wantErr: nil,
		},
		{
			name: "success - with variant",
			media: &entity.Media{
				ID:   1,
				Path: "users/avatar/1/2024/01/test.jpg",
				Variants: entity.JSONMap{
					"thumb": "users/avatar/1/2024/01/test_thumb.jpg",
				},
			},
			variant: "thumb",
			setupMock: func(storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					TemporaryURL(gomock.Any(), "users/avatar/1/2024/01/test_thumb.jpg", time.Hour).
					Return("https://s3.example.com/signed-thumb-url", nil)
			},
			wantURL: "https://s3.example.com/signed-thumb-url",
			wantErr: nil,
		},
		{
			name: "success - variant not found, fallback to original",
			media: &entity.Media{
				ID:       1,
				Path:     "users/avatar/1/2024/01/test.jpg",
				Variants: entity.JSONMap{},
			},
			variant: "nonexistent",
			setupMock: func(storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					TemporaryURL(gomock.Any(), "users/avatar/1/2024/01/test.jpg", time.Hour).
					Return("https://s3.example.com/signed-url", nil)
			},
			wantURL: "https://s3.example.com/signed-url",
			wantErr: nil,
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

			tt.setupMock(mockStorage)

			uc := media.New(mockMediaRepo, mockStorage, nil, mockLogger, "s3", 10*1024*1024)
			url, err := uc.GetURL(context.Background(), tt.media, tt.variant)

			if tt.wantErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, url)
		})
	}
}

func TestGetPresignedUploadURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filename  string
		setupMock func(storageProvider *MockStorageProvider)
		wantErr   error
	}{
		{
			name:     "success",
			filename: "document.pdf",
			setupMock: func(storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					PresignedUploadURL(gomock.Any(), gomock.Any(), 15*time.Minute).
					Return("https://s3.example.com/presigned-upload-url", nil)
			},
			wantErr: nil,
		},
		{
			name:     "error - storage fails",
			filename: "file.jpg",
			setupMock: func(storageProvider *MockStorageProvider) {
				storageProvider.EXPECT().
					PresignedUploadURL(gomock.Any(), gomock.Any(), 15*time.Minute).
					Return("", errors.New("presigned url error"))
			},
			wantErr: errors.New("presigned url error"),
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

			tt.setupMock(mockStorage)

			uc := media.New(mockMediaRepo, mockStorage, nil, mockLogger, "s3", 10*1024*1024)
			result, err := uc.GetPresignedUploadURL(context.Background(), tt.filename)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.UploadURL)
			assert.NotEmpty(t, result.Path)
			assert.Equal(t, 900, result.ExpiresIn) // 15 minutes
		})
	}
}
