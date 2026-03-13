package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/audit"
)

func TestGetCurrentUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		userID    uint
		setupMock func(userRepo *MockUserRepo)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			setupMock: func(userRepo *MockUserRepo) {
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.User{
						ID:     1,
						Email:  "test@example.com",
						Name:   "Test User",
						Active: true,
						Role:   entity.Role{ID: 1, Name: "user"},
					}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(userRepo *MockUserRepo) {
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: repo.ErrNotFound,
		},
		{
			name:   "repo error",
			userID: 1,
			setupMock: func(userRepo *MockUserRepo) {
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := NewMockUserRepo(ctrl)
			mockRoleRepo := NewMockRoleRepo(ctrl)
			mockRefreshRepo := NewMockRefreshTokenRepo(ctrl)
			mockJWT := NewMockService(ctrl)

			tt.setupMock(mockUserRepo)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT, audit.NewNoop())
			got, err := uc.GetCurrentUser(context.Background(), tt.userID)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, repo.ErrNotFound) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.userID, got.ID)
		})
	}
}
