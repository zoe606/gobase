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

func TestResetPassword(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		input auth.ResetPasswordInput
	}

	tests := []struct {
		name      string
		args      args
		setupUC   func(uc *auth.UseCase, prRepo *MockPasswordResetRepo)
		setupMock func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "valid-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(reset, nil)

				user := &entity.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "old-hashed-password",
					Active:   true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)

				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)

				prRepo.EXPECT().
					MarkAsUsed(gomock.Any(), uint(1)).
					Return(nil)

				refreshRepo.EXPECT().
					DeleteByUserID(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo not configured",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "any-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, _ *MockPasswordResetRepo) {
				uc.WithPasswordReset(nil, auth.ResetConfig{})
			},
			setupMock: func(_ *MockUserRepo, _ *MockRefreshTokenRepo, _ *MockPasswordResetRepo) {},
			wantErr:   errors.New("password reset repository not configured"),
		},
		{
			name: "token not found",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "nonexistent-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(_ *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "nonexistent-token").
					Return(nil, repo.ErrNotFound)
			},
			wantErr: auth.ErrResetTokenNotFound,
		},
		{
			name: "get by token repo error",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "some-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(_ *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "some-token").
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "token already used",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "used-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(_ *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				usedAt := time.Now().Add(-1 * time.Hour)
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    1,
					Token:     "used-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    &usedAt,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "used-token").
					Return(reset, nil)
			},
			wantErr: auth.ErrResetTokenUsed,
		},
		{
			name: "token expired",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "expired-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(_ *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    1,
					Token:     "expired-token",
					ExpiresAt: time.Now().Add(-1 * time.Hour),
					UsedAt:    nil,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "expired-token").
					Return(reset, nil)
			},
			wantErr: auth.ErrResetTokenExpired,
		},
		{
			name: "user not found",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "valid-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    999,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(reset, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: errors.New("user not found"),
		},
		{
			name: "user repo error",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "valid-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(reset, nil)
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "update user error",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "valid-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(reset, nil)

				user := &entity.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "old-hashed-password",
					Active:   true,
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
			name: "mark as used error",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "valid-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, _ *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(reset, nil)

				user := &entity.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "old-hashed-password",
					Active:   true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
				prRepo.EXPECT().
					MarkAsUsed(gomock.Any(), uint(1)).
					Return(errors.New("mark as used error"))
			},
			wantErr: errors.New("mark as used error"),
		},
		{
			name: "delete refresh tokens error",
			args: args{
				ctx: context.Background(),
				input: auth.ResetPasswordInput{
					Token:       "valid-token",
					NewPassword: "newpassword123",
				},
			},
			setupUC: func(uc *auth.UseCase, prRepo *MockPasswordResetRepo) {
				uc.WithPasswordReset(prRepo, auth.ResetConfig{
					TokenTTL: time.Hour,
				})
			},
			setupMock: func(userRepo *MockUserRepo, refreshRepo *MockRefreshTokenRepo, prRepo *MockPasswordResetRepo) {
				reset := &entity.PasswordReset{
					ID:        1,
					UserID:    1,
					Token:     "valid-token",
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}
				prRepo.EXPECT().
					GetByToken(gomock.Any(), "valid-token").
					Return(reset, nil)

				user := &entity.User{
					ID:       1,
					Email:    "test@example.com",
					Password: "old-hashed-password",
					Active:   true,
				}
				userRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(user, nil)
				userRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
				prRepo.EXPECT().
					MarkAsUsed(gomock.Any(), uint(1)).
					Return(nil)
				refreshRepo.EXPECT().
					DeleteByUserID(gomock.Any(), uint(1)).
					Return(errors.New("delete error"))
			},
			wantErr: errors.New("delete error"),
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

			tt.setupMock(mockUserRepo, mockRefreshRepo, mockPRRepo)

			uc := auth.New(mockUserRepo, mockRoleRepo, mockRefreshRepo, mockJWT)
			tt.setupUC(uc, mockPRRepo)

			err := uc.ResetPassword(tt.args.ctx, tt.args.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, auth.ErrResetTokenNotFound) ||
					errors.Is(tt.wantErr, auth.ErrResetTokenUsed) ||
					errors.Is(tt.wantErr, auth.ErrResetTokenExpired) {
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
