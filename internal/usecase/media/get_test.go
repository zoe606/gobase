package media_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/dto/media"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/media"
)

func TestGetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        uint
		setupMock func(mediaRepo *MockMediaRepo)
		want      *entity.Media
		wantErr   error
	}{
		{
			name: "success",
			id:   1,
			setupMock: func(mediaRepo *MockMediaRepo) {
				mediaRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Media{
						ID:             1,
						AttachableType: "users",
						AttachableID:   1,
						Collection:     "avatar",
						Filename:       "test.jpg",
						OriginalName:   "avatar.jpg",
						MimeType:       "image/jpeg",
						Size:           1024,
						Type:           entity.MediaTypeImage,
					}, nil)
			},
			want: &entity.Media{
				ID:             1,
				AttachableType: "users",
				AttachableID:   1,
				Collection:     "avatar",
				Filename:       "test.jpg",
				OriginalName:   "avatar.jpg",
				MimeType:       "image/jpeg",
				Size:           1024,
				Type:           entity.MediaTypeImage,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   999,
			setupMock: func(mediaRepo *MockMediaRepo) {
				mediaRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			want:    nil,
			wantErr: repo.ErrNotFound,
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

			tt.setupMock(mockMediaRepo)

			uc := media.New(mockMediaRepo, mockStorage, nil, mockLogger, "s3", 10*1024*1024)
			result, err := uc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.want.ID, result.ID)
			assert.Equal(t, tt.want.Filename, result.Filename)
		})
	}
}

func TestGetByAttachable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       mediadto.GetMediaRequest
		setupMock func(mediaRepo *MockMediaRepo)
		wantCount int
		wantErr   error
	}{
		{
			name: "success - with collection",
			req: mediadto.GetMediaRequest{
				AttachableType: "users",
				AttachableID:   1,
				Collection:     "avatar",
			},
			setupMock: func(mediaRepo *MockMediaRepo) {
				mediaRepo.EXPECT().
					GetByAttachable(gomock.Any(), "users", uint(1), "avatar").
					Return([]*entity.Media{
						{ID: 1, Filename: "avatar1.jpg"},
						{ID: 2, Filename: "avatar2.jpg"},
					}, nil)
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name: "success - empty collection (all)",
			req: mediadto.GetMediaRequest{
				AttachableType: "users",
				AttachableID:   1,
				Collection:     "",
			},
			setupMock: func(mediaRepo *MockMediaRepo) {
				mediaRepo.EXPECT().
					GetByAttachable(gomock.Any(), "users", uint(1), "").
					Return([]*entity.Media{
						{ID: 1, Filename: "avatar.jpg", Collection: "avatar"},
						{ID: 2, Filename: "doc.pdf", Collection: "documents"},
					}, nil)
			},
			wantCount: 2,
			wantErr:   nil,
		},
		{
			name: "success - no results",
			req: mediadto.GetMediaRequest{
				AttachableType: "posts",
				AttachableID:   999,
				Collection:     "images",
			},
			setupMock: func(mediaRepo *MockMediaRepo) {
				mediaRepo.EXPECT().
					GetByAttachable(gomock.Any(), "posts", uint(999), "images").
					Return([]*entity.Media{}, nil)
			},
			wantCount: 0,
			wantErr:   nil,
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

			tt.setupMock(mockMediaRepo)

			uc := media.New(mockMediaRepo, mockStorage, nil, mockLogger, "s3", 10*1024*1024)
			result, err := uc.GetByAttachable(context.Background(), tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Items, tt.wantCount)
		})
	}
}
