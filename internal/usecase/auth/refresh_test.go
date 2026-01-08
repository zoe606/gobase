package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/auth"
)

func TestRefresh(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		input authdto.RefreshRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				input: authdto.RefreshRequest{
					RefreshToken: "valid-refresh-token",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				refreshRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-refresh-token").
					Return(&entity.RefreshToken{
						ID:        1,
						UserID:    1,
						Token:     "valid-refresh-token",
						ExpiresAt: time.Now().Add(24 * time.Hour),
					}, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.User{
						ID:     1,
						Email:  "test@example.com",
						Active: true,
						Role:   entity.Role{ID: 1, Name: "user"},
					}, nil)
				refreshRepo.EXPECT().
					DeleteByToken(gomock.Any(), "valid-refresh-token").
					Return(nil)
				jwtSvc.EXPECT().
					GenerateAccessToken(uint(1), "test@example.com", "user", gomock.Any()).
					Return("new-access-token", time.Now().Add(15*time.Minute).Unix(), nil)
				jwtSvc.EXPECT().
					GenerateRefreshToken().
					Return("new-refresh-token", time.Now().Add(24*time.Hour), nil)
				refreshRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "token not found",
			args: args{
				ctx: context.Background(),
				input: authdto.RefreshRequest{
					RefreshToken: "invalid-token",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				refreshRepo.EXPECT().
					GetByToken(gomock.Any(), "invalid-token").
					Return(nil, repo.ErrNotFound)
			},
			wantErr: auth.ErrInvalidToken,
		},
		{
			name: "token expired",
			args: args{
				ctx: context.Background(),
				input: authdto.RefreshRequest{
					RefreshToken: "expired-token",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				refreshRepo.EXPECT().
					GetByToken(gomock.Any(), "expired-token").
					Return(&entity.RefreshToken{
						ID:        1,
						UserID:    1,
						Token:     "expired-token",
						ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
					}, nil)
			},
			wantErr: auth.ErrInvalidToken,
		},
		{
			name: "user not found",
			args: args{
				ctx: context.Background(),
				input: authdto.RefreshRequest{
					RefreshToken: "valid-refresh-token",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				refreshRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-refresh-token").
					Return(&entity.RefreshToken{
						ID:        1,
						UserID:    1,
						Token:     "valid-refresh-token",
						ExpiresAt: time.Now().Add(24 * time.Hour),
					}, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: auth.ErrInvalidToken,
		},
		{
			name: "user not active",
			args: args{
				ctx: context.Background(),
				input: authdto.RefreshRequest{
					RefreshToken: "valid-refresh-token",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				refreshRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-refresh-token").
					Return(&entity.RefreshToken{
						ID:        1,
						UserID:    1,
						Token:     "valid-refresh-token",
						ExpiresAt: time.Now().Add(24 * time.Hour),
					}, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.User{
						ID:     1,
						Email:  "test@example.com",
						Active: false,
					}, nil)
			},
			wantErr: auth.ErrInvalidToken,
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

			tt.setupMock(mockUserRepo, mockRefreshRepo, mockJWT)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT)
			got, err := uc.Refresh(tt.args.ctx, tt.args.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotEmpty(t, got.AccessToken)
			require.NotEmpty(t, got.RefreshToken)
		})
	}
}
