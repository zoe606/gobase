package profile_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/profile"
)

func TestGetProfile(t *testing.T) {
	t.Parallel()

	avatarMediaID := uint(100)

	tests := []struct {
		name      string
		userID    uint
		setupMock func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider)
		wantErr   bool
		wantBio   string
		wantPhone string
		hasAvatar bool
	}{
		{
			name:   "success - existing profile with avatar",
			userID: 1,
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(1)).
					Return(&entity.Profile{
						ID:            1,
						UserID:        1,
						Bio:           "Test bio",
						Phone:         "+1234567890",
						AvatarMediaID: &avatarMediaID,
					}, nil)

				mediaRepo.EXPECT().
					GetByID(gomock.Any(), avatarMediaID).
					Return(&entity.Media{
						ID:   avatarMediaID,
						Path: "profiles/avatar/1/2024/01/test.jpg",
					}, nil)

				storage.EXPECT().
					TemporaryURL(gomock.Any(), "profiles/avatar/1/2024/01/test.jpg", gomock.Any()).
					Return("https://s3.example.com/test.jpg", nil)
			},
			wantErr:   false,
			wantBio:   "Test bio",
			wantPhone: "+1234567890",
			hasAvatar: true,
		},
		{
			name:   "success - existing profile without avatar",
			userID: 2,
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(2)).
					Return(&entity.Profile{
						ID:     2,
						UserID: 2,
						Bio:    "No avatar bio",
						Phone:  "",
					}, nil)
			},
			wantErr:   false,
			wantBio:   "No avatar bio",
			wantPhone: "",
			hasAvatar: false,
		},
		{
			name:   "success - new profile created",
			userID: 3,
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(3)).
					Return(nil, repo.ErrNotFound)

				profileRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr:   false,
			wantBio:   "",
			wantPhone: "",
			hasAvatar: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockProfileRepo := NewMockProfileRepo(ctrl)
			mockMediaRepo := NewMockMediaRepo(ctrl)
			mockStorage := NewMockStorageProvider(ctrl)
			mockLogger := NewMockLogger()

			tt.setupMock(mockProfileRepo, mockMediaRepo, mockStorage)

			uc := profile.New(mockProfileRepo, mockMediaRepo, mockStorage, mockLogger)
			result, err := uc.GetProfile(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.userID, result.UserID)
			assert.Equal(t, tt.wantBio, result.Bio)
			assert.Equal(t, tt.wantPhone, result.Phone)

			if tt.hasAvatar {
				require.NotNil(t, result.Avatar)
				assert.NotEmpty(t, result.Avatar.URL)
			} else {
				assert.Nil(t, result.Avatar)
			}
		})
	}
}
