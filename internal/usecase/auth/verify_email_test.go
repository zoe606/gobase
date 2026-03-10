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

func TestVerifyEmail(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		token string
	}

	tests := []struct {
		name      string
		args      args
		setupUC   func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo)
		setupMock func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx:   context.Background(),
				token: "valid-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled:    true,
					AutoVerify: false,
					TokenTTL:   24 * time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(verification, nil)

				user := &entity.User{
					ID:    1,
					Email: "test@example.com",
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)

				evRepo.EXPECT().
					MarkAsUsed(gomock.Any(), uint(1)).
					Return(nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "verification disabled - returns nil",
			args: args{
				ctx:   context.Background(),
				token: "any-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: false,
				})
			},
			setupMock: func(_ *MockUserRepo, _ *MockEmailVerificationRepo) {
				// No mock calls expected
			},
			wantErr: nil,
		},
		{
			name: "repo not configured",
			args: args{
				ctx:   context.Background(),
				token: "any-token",
			},
			setupUC: func(uc *auth.UseCase, _ *MockEmailVerificationRepo) {
				uc.WithEmailVerification(nil, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(_ *MockUserRepo, _ *MockEmailVerificationRepo) {},
			wantErr:   errors.New("email verification repository not configured"),
		},
		{
			name: "token not found",
			args: args{
				ctx:   context.Background(),
				token: "nonexistent-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(_ *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "nonexistent-token").
					Return(nil, repo.ErrNotFound)
			},
			wantErr: auth.ErrVerificationNotFound,
		},
		{
			name: "get by token repo error",
			args: args{
				ctx:   context.Background(),
				token: "some-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(_ *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "some-token").
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "token already used",
			args: args{
				ctx:   context.Background(),
				token: "used-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(_ *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				usedAt := time.Now().Add(-1 * time.Hour)
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    1,
					Token:     "used-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    &usedAt,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "used-token").
					Return(verification, nil)
			},
			wantErr: auth.ErrVerificationUsed,
		},
		{
			name: "token expired",
			args: args{
				ctx:   context.Background(),
				token: "expired-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(_ *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    1,
					Token:     "expired-token",
					ExpiresAt: time.Now().Add(-1 * time.Hour),
					UsedAt:    nil,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "expired-token").
					Return(verification, nil)
			},
			wantErr: auth.ErrVerificationExpired,
		},
		{
			name: "user not found",
			args: args{
				ctx:   context.Background(),
				token: "valid-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    999,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(verification, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: errors.New("user not found"),
		},
		{
			name: "user repo error",
			args: args{
				ctx:   context.Background(),
				token: "valid-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(verification, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "email already verified",
			args: args{
				ctx:   context.Background(),
				token: "valid-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(verification, nil)

				now := time.Now()
				user := &entity.User{
					ID:              1,
					Email:           "test@example.com",
					EmailVerifiedAt: &now,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
			},
			wantErr: auth.ErrEmailAlreadyVerified,
		},
		{
			name: "mark as used error",
			args: args{
				ctx:   context.Background(),
				token: "valid-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(verification, nil)

				user := &entity.User{
					ID:    1,
					Email: "test@example.com",
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				evRepo.EXPECT().
					MarkAsUsed(gomock.Any(), uint(1)).
					Return(errors.New("mark as used error"))
			},
			wantErr: errors.New("mark as used error"),
		},
		{
			name: "update user error",
			args: args{
				ctx:   context.Background(),
				token: "valid-token",
			},
			setupUC: func(uc *auth.UseCase, evRepo *MockEmailVerificationRepo) {
				uc.WithEmailVerification(evRepo, auth.VerificationConfig{
					Enabled: true,
				})
			},
			setupMock: func(userRepo *MockUserRepo, evRepo *MockEmailVerificationRepo) {
				verification := &entity.EmailVerification{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				evRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(verification, nil)

				user := &entity.User{
					ID:    1,
					Email: "test@example.com",
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				evRepo.EXPECT().
					MarkAsUsed(gomock.Any(), uint(1)).
					Return(nil)
				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("update error"))
			},
			wantErr: errors.New("update error"),
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

			err := uc.VerifyEmail(tt.args.ctx, tt.args.token)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, auth.ErrVerificationNotFound) ||
					errors.Is(tt.wantErr, auth.ErrVerificationUsed) ||
					errors.Is(tt.wantErr, auth.ErrVerificationExpired) ||
					errors.Is(tt.wantErr, auth.ErrEmailAlreadyVerified) {
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
