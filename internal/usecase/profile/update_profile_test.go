package profile_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	profiledto "go-boilerplate/internal/dto/profile"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/profile"
)

func TestUpdateProfile(t *testing.T) {
	t.Parallel()

	avatarMediaID := uint(100)
	bio := "Updated bio"
	phone := "+1234567890"

	tests := []struct {
		name      string
		userID    uint
		req       profiledto.UpdateProfileRequest
		setupMock func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider)
		wantErr   error
		wantBio   string
		wantPhone string
		hasAvatar bool
	}{
		{
			name:   "success - update bio only",
			userID: 1,
			req: profiledto.UpdateProfileRequest{
				Bio: &bio,
			},
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(1)).
					Return(&entity.Profile{
						ID:     1,
						UserID: 1,
						Bio:    "Old bio",
					}, nil)

				profileRepo.EXPECT().
					Upsert(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr:   nil,
			wantBio:   bio,
			wantPhone: "",
			hasAvatar: false,
		},
		{
			name:   "success - update with avatar",
			userID: 1,
			req: profiledto.UpdateProfileRequest{
				Bio:           &bio,
				Phone:         &phone,
				AvatarMediaID: &avatarMediaID,
			},
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(1)).
					Return(&entity.Profile{
						ID:     1,
						UserID: 1,
					}, nil)

				mediaRepo.EXPECT().
					GetByID(gomock.Any(), avatarMediaID).
					Return(&entity.Media{
						ID:             avatarMediaID,
						AttachableType: "profiles",
						AttachableID:   1,
						Path:           "profiles/avatar/1/2024/01/test.jpg",
					}, nil)

				profileRepo.EXPECT().
					Upsert(gomock.Any(), gomock.Any()).
					Return(nil)

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
			wantErr:   nil,
			wantBio:   bio,
			wantPhone: phone,
			hasAvatar: true,
		},
		{
			name:   "success - create new profile",
			userID: 2,
			req: profiledto.UpdateProfileRequest{
				Bio: &bio,
			},
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(2)).
					Return(nil, repo.ErrNotFound)

				profileRepo.EXPECT().
					Upsert(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr:   nil,
			wantBio:   bio,
			wantPhone: "",
			hasAvatar: false,
		},
		{
			name:   "error - media not found",
			userID: 1,
			req: profiledto.UpdateProfileRequest{
				AvatarMediaID: &avatarMediaID,
			},
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(1)).
					Return(&entity.Profile{
						ID:     1,
						UserID: 1,
					}, nil)

				mediaRepo.EXPECT().
					GetByID(gomock.Any(), avatarMediaID).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: profile.ErrInvalidMedia,
		},
		{
			name:   "error - media belongs to different user",
			userID: 1,
			req: profiledto.UpdateProfileRequest{
				AvatarMediaID: &avatarMediaID,
			},
			setupMock: func(profileRepo *MockProfileRepo, mediaRepo *MockMediaRepo, storage *MockStorageProvider) {
				profileRepo.EXPECT().
					GetByUserID(gomock.Any(), uint(1)).
					Return(&entity.Profile{
						ID:     1,
						UserID: 1,
					}, nil)

				mediaRepo.EXPECT().
					GetByID(gomock.Any(), avatarMediaID).
					Return(&entity.Media{
						ID:             avatarMediaID,
						AttachableType: "profiles",
						AttachableID:   2, // Different user
						Path:           "profiles/avatar/2/2024/01/test.jpg",
					}, nil)
			},
			wantErr: profile.ErrInvalidMedia,
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
			result, err := uc.UpdateProfile(context.Background(), tt.userID, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
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
