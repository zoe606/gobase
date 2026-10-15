package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-boilerplate/internal/repo"
)

// VerifyEmail verifies a user's email using the provided token.
func (uc *UseCase) VerifyEmail(ctx context.Context, token string) error {
	return uc.inTransaction(ctx, func(ctx context.Context) error { return uc.verifyEmail(ctx, token) })
}

func (uc *UseCase) verifyEmail(ctx context.Context, token string) error {
	if !uc.verificationConfig.Enabled {
		return ErrVerificationDisabled
	}

	if uc.emailVerificationRepo == nil {
		return fmt.Errorf("VerifyEmail: email verification repository not configured")
	}

	// Get verification record
	verification, err := uc.emailVerificationRepo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrVerificationNotFound
		}
		return fmt.Errorf("VerifyEmail - GetByToken: %w", err)
	}

	// Check if already used
	if verification.IsUsed() {
		return ErrVerificationUsed
	}

	// Check if expired
	if verification.IsExpired() {
		return ErrVerificationExpired
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, verification.UserID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return fmt.Errorf("VerifyEmail: user not found")
		}
		return fmt.Errorf("VerifyEmail - GetByID: %w", err)
	}

	// Check if already verified
	if user.IsEmailVerified() {
		return ErrEmailAlreadyVerified
	}

	// Mark token as used
	if err := uc.emailVerificationRepo.MarkAsUsed(ctx, verification.ID); err != nil {
		if errors.Is(err, repo.ErrTokenConsumed) {
			return ErrVerificationUsed
		}
		return fmt.Errorf("VerifyEmail - MarkAsUsed: %w", err)
	}

	// Mark user email as verified
	now := time.Now()
	if err := uc.userRepo.SetEmailVerified(ctx, user.ID, now); err != nil {
		return fmt.Errorf("VerifyEmail - Update: %w", err)
	}

	return nil
}
