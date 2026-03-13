package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/audit"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		input authdto.RegisterRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(userRepo *MockUserRepo, roleRepo *MockRoleRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				input: authdto.RegisterRequest{
					Email:    "new@example.com",
					Password: "password123",
					Name:     "New User",
				},
			},
			setupMock: func(userRepo *MockUserRepo, roleRepo *MockRoleRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				userRepo.EXPECT().
					EmailExists(gomock.Any(), "new@example.com").
					Return(false, nil)
				roleRepo.EXPECT().
					GetByName(gomock.Any(), "user").
					Return(&entity.Role{ID: 1, Name: "user"}, nil)
				userRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				jwtSvc.EXPECT().
					GenerateAccessToken(gomock.Any(), "new@example.com", "user", gomock.Any()).
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
			name: "email already exists",
			args: args{
				ctx: context.Background(),
				input: authdto.RegisterRequest{
					Email:    "existing@example.com",
					Password: "password123",
					Name:     "Existing User",
				},
			},
			setupMock: func(userRepo *MockUserRepo, roleRepo *MockRoleRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				userRepo.EXPECT().
					EmailExists(gomock.Any(), "existing@example.com").
					Return(true, nil)
			},
			wantErr: auth.ErrEmailExists,
		},
		{
			name: "default role not found",
			args: args{
				ctx: context.Background(),
				input: authdto.RegisterRequest{
					Email:    "new@example.com",
					Password: "password123",
					Name:     "New User",
				},
			},
			setupMock: func(userRepo *MockUserRepo, roleRepo *MockRoleRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				userRepo.EXPECT().
					EmailExists(gomock.Any(), "new@example.com").
					Return(false, nil)
				roleRepo.EXPECT().
					GetByName(gomock.Any(), "user").
					Return(nil, repo.ErrNotFound)
			},
			wantErr: auth.ErrDefaultRoleNotFound,
		},
		{
			name: "email exists check error",
			args: args{
				ctx: context.Background(),
				input: authdto.RegisterRequest{
					Email:    "new@example.com",
					Password: "password123",
					Name:     "New User",
				},
			},
			setupMock: func(userRepo *MockUserRepo, roleRepo *MockRoleRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				userRepo.EXPECT().
					EmailExists(gomock.Any(), "new@example.com").
					Return(false, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "create user error",
			args: args{
				ctx: context.Background(),
				input: authdto.RegisterRequest{
					Email:    "new@example.com",
					Password: "password123",
					Name:     "New User",
				},
			},
			setupMock: func(userRepo *MockUserRepo, roleRepo *MockRoleRepo, refreshRepo *MockRefreshTokenRepo, jwtSvc *MockService) {
				userRepo.EXPECT().
					EmailExists(gomock.Any(), "new@example.com").
					Return(false, nil)
				roleRepo.EXPECT().
					GetByName(gomock.Any(), "user").
					Return(&entity.Role{ID: 1, Name: "user"}, nil)
				userRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("create error"))
			},
			wantErr: errors.New("create error"),
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

			tt.setupMock(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT, audit.NewNoop())
			got, err := uc.Register(tt.args.ctx, tt.args.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, auth.ErrEmailExists) ||
					errors.Is(tt.wantErr, auth.ErrDefaultRoleNotFound) {
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
