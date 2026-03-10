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
)

func TestSendVerificationEmail(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uint
	}

	tests := []struct {
		name      string
		args      args
		setupUC   func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo)
		setupMock func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo)
		wantErr   error
	}{
		{
			name: "success - verification disabled",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, _ *MockEmailVerificationRepo) {
				uc.WithEmailVerification(nil, auth.VerificationConfig{
					Enabled: false,
				})
			},
			setupMock: func(_ *MockUserRepo, _ *MockEmailVerificationRepo) {
				// No mock calls expected - verification is disabled
			},
			wantErr: nil,
		},
		{
			name: "success - auto verify",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled:    true,
					AutoVerify: true,
					TokenTTL:   24 * time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockEmailVerificationRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "success - creates verification token",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled:    true,
					AutoVerify: false,
					TokenTTL:   24 * time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				evRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			args: args{
				ctx:    context.Background(),
				userID: 999,
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockEmailVerificationRepo) {
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: errors.New("user not found"),
		},
		{
			name: "user repo error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockEmailVerificationRepo) {
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "email already verified",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockEmailVerificationRepo) {
				now := time.Now()
				user := &entity.User{
					ID:              1,
					Email:           "test@example.com",
					Active:          true,
					EmailVerifiedAt: &now,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
			},
			wantErr: auth.ErrEmailAlreadyVerified,
		},
		{
			name: "auto verify - update error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled:    true,
					AutoVerify: true,
					TokenTTL:   24 * time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockEmailVerificationRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("update error"))
			},
			wantErr: errors.New("update error"),
		},
		{
			name: "repo not configured",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, _ *MockEmailVerificationRepo) {
				uc.WithEmailVerification(nil, auth.VerificationConfig{
					Enabled:    true,
					AutoVerify: false,
					TokenTTL:   24 * time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockEmailVerificationRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
			},
			wantErr: errors.New("email verification repository not configured"),
		},
		{
			name: "create verification error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled:    true,
					AutoVerify: false,
					TokenTTL:   24 * time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				user := &entity.User{
					ID:     1,
					Email:  "test@example.com",
					Active: true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				evRepo.EXPECT().
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
			mockEVRepo := NewMockEmailVerificationRepo(ctrl)

			tt.setupMock(mockUserRepo, mockEVRepo)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT)
			tt.setupUC(uc, mockEVRepo)

			err := uc.SendVerificationEmail(tt.args.ctx, tt.args.userID)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, auth.ErrEmailAlreadyVerified) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
