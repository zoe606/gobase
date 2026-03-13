package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/audit"
)

func TestRequestPasswordReset(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		email string
	}

	tests := []struct {
		name      string
		args      args
		setupUC   func(uc *auth.UseCase, prRepo *MockPasswordResetRepo)
		setupMock func(userRepo *MockUserRepo, prRepo *MockPasswordResetRepo)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, prRepo *MockPasswordResetRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)
				prRepo.EXPECT().
					DeleteByUserID(gomock.Any(), uint(1)).
					Return(nil)
				prRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo not configured",
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			setupUC: func(uc *auth.UseCase, _ *MockPasswordResetRepo) {
				uc.WithPasswordReset(nil, auth.ResetConfig{})
			},
			setupMock: func(_ *MockUserRepo, _ *MockPasswordResetRepo) {},
			wantErr:   errors.New("password reset repository not configured"),
		},
		{
			name: "user not found - returns nil to prevent enumeration",
			args: args{
				ctx:   context.Background(),
				email: "notfound@example.com",
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockPasswordResetRepo) {
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "notfound@example.com").
					Return(nil, repo.ErrNotFound)
			},
			wantErr: nil,
		},
		{
			name: "user repo error",
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockPasswordResetRepo) {
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "delete existing tokens error",
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, prRepo *MockPasswordResetRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)
				prRepo.EXPECT().
					DeleteByUserID(gomock.Any(), uint(1)).
					Return(errors.New("delete error"))
			},
			wantErr: errors.New("delete error"),
		},
		{
			name: "create reset token error",
			args: args{
				ctx:   context.Background(),
				email: "test@example.com",
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, prRepo *MockPasswordResetRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(user, nil)
				prRepo.EXPECT().
					DeleteByUserID(gomock.Any(), uint(1)).
					Return(nil)
				prRepo.EXPECT().
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
			mockPRRepo := NewMockPasswordResetRepo(ctrl)

			tt.setupMock(mockUserRepo, mockPRRepo)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT, audit.NewNoop())
			tt.setupUC(uc, mockPRRepo)

			err := uc.RequestPasswordReset(tt.args.ctx, tt.args.email)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
