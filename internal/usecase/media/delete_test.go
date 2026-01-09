package media_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/media"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        uint
		setupMock func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider)
		wantErr   error
	}{
		{
			name: "success - with variants",
			id:   1,
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				mediaRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Media{
						ID:   1,
						Path: "users/avatar/1/test.jpg",
						Variants: entity.JSONMap{
							"thumb":  "users/avatar/1/test_thumb.jpg",
							"medium": "users/avatar/1/test_medium.jpg",
						},
					}, nil)
				storageProvider.EXPECT().
					Delete(gomock.Any(), "users/avatar/1/test.jpg").
					Return(nil)
				storageProvider.EXPECT().
					Delete(gomock.Any(), gomock.Any()).
					Return(nil).Times(2) // for variants
				mediaRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "success - no variants",
			id:   2,
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				mediaRepo.EXPECT().
					GetByID(gomock.Any(), uint(2)).
					Return(&entity.Media{
						ID:       2,
						Path:     "posts/images/5/doc.pdf",
						Variants: nil,
					}, nil)
				storageProvider.EXPECT().
					Delete(gomock.Any(), "posts/images/5/doc.pdf").
					Return(nil)
				mediaRepo.EXPECT().
					Delete(gomock.Any(), uint(2)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "error - not found",
			id:   999,
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				mediaRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: repo.ErrNotFound,
		},
		{
			name: "success - storage delete fails but continues",
			id:   3,
			setupMock: func(mediaRepo *MockMediaRepo, storageProvider *MockStorageProvider) {
				mediaRepo.EXPECT().
					GetByID(gomock.Any(), uint(3)).
					Return(&entity.Media{
						ID:       3,
						Path:     "users/avatar/1/test.jpg",
						Variants: nil,
					}, nil)
				storageProvider.EXPECT().
					Delete(gomock.Any(), "users/avatar/1/test.jpg").
					Return(errors.New("storage error")) // Should continue despite error
				mediaRepo.EXPECT().
					Delete(gomock.Any(), uint(3)).
					Return(nil)
			},
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

			tt.setupMock(mockMediaRepo, mockStorage)

			uc := media.New(mockMediaRepo, mockStorage, nil, mockLogger, "s3", 10*1024*1024)
			err := uc.Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
