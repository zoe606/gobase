// Package auth provides authentication use cases.
package auth

//go:generate mockgen -source=../../repo/contracts.go -destination=mocks_repo_test.go -package=auth_test
//go:generate mockgen -source=../../../pkg/jwt/jwt.go -destination=mocks_jwt_test.go -package=auth_test

import (
	"context"
	"fmt"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/asynq"
	"go-boilerplate/pkg/jwt"
)

// UseCase implements authentication business logic.
type UseCase struct {
	userRepo              repo.UserRepo
	roleRepo              repo.RoleRepo
	refreshTokenRepo      repo.RefreshTokenRepo
	emailVerificationRepo repo.EmailVerificationRepo
	passwordResetRepo     repo.PasswordResetRepo
	jwtService            jwt.Service
	asynqClient           *asynq.Client
	appName               string
	verificationConfig    VerificationConfig
	resetConfig           ResetConfig
}

// VerificationConfig holds email verification configuration.
type VerificationConfig struct {
	Enabled    bool
	AutoVerify bool
	TokenTTL   time.Duration
	BaseURL    string
}

// ResetConfig holds password reset configuration.
type ResetConfig struct {
	TokenTTL time.Duration
	BaseURL  string
}

// New creates a new auth use case.
func New(
	userRepo repo.UserRepo,
	roleRepo repo.RoleRepo,
	refreshTokenRepo repo.RefreshTokenRepo,
	jwtService jwt.Service,
) *UseCase {
	return &UseCase{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		verificationConfig: VerificationConfig{
			Enabled:    true,
			AutoVerify: true,
			TokenTTL:   24 * time.Hour,
			BaseURL:    "http://localhost:3000",
		},
		resetConfig: ResetConfig{
			TokenTTL: time.Hour,
			BaseURL:  "http://localhost:3000",
		},
	}
}

// WithEmailVerification sets the email verification repository and config.
func (uc *UseCase) WithEmailVerification(evRepo repo.EmailVerificationRepo, cfg VerificationConfig) *UseCase {
	uc.emailVerificationRepo = evRepo
	uc.verificationConfig = cfg
	return uc
}

// WithPasswordReset sets the password reset repository and config.
func (uc *UseCase) WithPasswordReset(prRepo repo.PasswordResetRepo, cfg ResetConfig) *UseCase {
	uc.passwordResetRepo = prRepo
	uc.resetConfig = cfg
	return uc
}

// WithAsynq sets the Asynq client and app name for background jobs.
func (uc *UseCase) WithAsynq(client *asynq.Client, appName string) *UseCase {
	uc.asynqClient = client
	uc.appName = appName

	return uc
}

// tokenPair holds generated access and refresh tokens.
type tokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  int64
	RefreshExpiresAt time.Time
}

// generateTokens creates access and refresh tokens for a user.
// This is a shared helper used by Register, Login, and Refresh.
func (uc *UseCase) generateTokens(user *entity.User) (*tokenPair, error) {
	accessToken, expiresAt, err := uc.jwtService.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role.Name,
		user.Role.GetPermissionNames(),
	)
	if err != nil {
		return nil, fmt.Errorf("GenerateAccessToken: %w", err)
	}

	refreshToken, refreshExpiresAt, err := uc.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("GenerateRefreshToken: %w", err)
	}

	return &tokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  expiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// storeRefreshToken stores a refresh token in the repository.
func (uc *UseCase) storeRefreshToken(ctx context.Context, userID uint, token string, expiresAt time.Time) error {
	return uc.refreshTokenRepo.Create(ctx, &entity.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
}
