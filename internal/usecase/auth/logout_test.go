package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/usecase/auth"
)

func TestLogout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		refreshToken string
		setupMock    func(refreshRepo *MockRefreshTokenRepo)
		wantErr      error
	}{
		{
			name:         "success",
			refreshToken: "valid-refresh-token",
			setupMock: func(refreshRepo *MockRefreshTokenRepo) {
				refreshRepo.EXPECT().
					DeleteByToken(gomock.Any(), "valid-refresh-token").
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:         "repo error",
			refreshToken: "valid-refresh-token",
			setupMock: func(refreshRepo *MockRefreshTokenRepo) {
				refreshRepo.EXPECT().
					DeleteByToken(gomock.Any(), "valid-refresh-token").
					Return(errors.New("database error"))
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

			tt.setupMock(mockRefreshRepo)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT)
			err := uc.Logout(context.Background(), tt.refreshToken)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
