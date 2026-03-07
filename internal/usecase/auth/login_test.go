package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/hasher"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	// Pre-hash a password for testing
	hashedPassword, err := hasher.Hash("password123")
	require.NoError(t, err)

	type args struct {
		ctx   context.Context
		input authdto.LoginRequest
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
				input: authdto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				user := &entity.User{
					ID:       1,
					Email:    "test@example.com",
					Password: hashedPassword,
					Name:     "Test User",
					Active:   true,
					Role: entity.Role{
						ID:   1,
						Name: "user",
					},
				}
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)
				jwtSvc.EXPECT().
					GenerateAccessToken(uint(1), "test@example.com", "user", gomock.Any()).
					Return("access-token", time.Now().Add(15*time.Minute).Unix(), nil)
				jwtSvc.EXPECT().
					GenerateRefreshToken().
					Return("refresh-token", time.Now().Add(24*time.Hour), nil)
				refreshRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			args: args{
				ctx: context.Background(),
				input: authdto.LoginRequest{
					Email:    "notfound@example.com",
					Password: "password123",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "notfound@example.com").
					Return(nil, repo.ErrNotFound)
			},
			wantErr: auth.ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			args: args{
				ctx: context.Background(),
				input: authdto.LoginRequest{
					Email:    "test@example.com",
					Password: "wrongpassword",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				user := &entity.User{
					ID:       1,
					Email:    "test@example.com",
					Password: hashedPassword,
					Active:   true,
				}
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)
			},
			wantErr: auth.ErrInvalidCredentials,
		},
		{
			name: "user not active",
			args: args{
				ctx: context.Background(),
				input: authdto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				user := &entity.User{
					ID:       1,
					Email:    "test@example.com",
					Password: hashedPassword,
					Active:   false,
				}
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)
			},
			wantErr: auth.ErrUserNotActive,
		},
		{
			name: "repo error",
			args: args{
				ctx: context.Background(),
				input: authdto.LoginRequest{
					Email:    "test@example.com",
					Password: "password123",
				},
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
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

			tt.setupMock(mockUserRepo, mockRefreshRepo, mockJWT)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT)
			got, err := uc.Login(tt.args.ctx, tt.args.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, auth.ErrInvalidCredentials) ||
					errors.Is(tt.wantErr, auth.ErrUserNotActive) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.NotEmpty(t, got.AccessToken)
			require.NotEmpty(t, got.RefreshToken)
		})
	}
}
