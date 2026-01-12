package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// ResendVerification resends a verification email to a user.
func (uc *UseCase) ResendVerification(ctx context.Context, email string) error {
	if !uc.verificationConfig.Enabled {
		return nil // Verification disabled, silently succeed
	}

	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			// Don't reveal if email exists - return nil to prevent enumeration
			return nil
		}
		return fmt.Errorf("ResendVerification - GetByEmail: %w", err)
	}

	// Check if already verified
	if user.IsEmailVerified() {
		return ErrEmailAlreadyVerified
	}

	// Auto-verify in development mode
	if uc.verificationConfig.AutoVerify {
		now := time.Now()
		user.EmailVerifiedAt = &now
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return fmt.Errorf("ResendVerification - Update: %w", err)
		}
		return nil
	}

	if uc.emailVerificationRepo == nil {
		return fmt.Errorf("ResendVerification: email verification repository not configured")
	}

	// Delete any existing verification tokens for this user
	if err := uc.emailVerificationRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("ResendVerification - DeleteByUserID: %w", err)
	}

	// Generate new verification token
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("ResendVerification - generateSecureToken: %w", err)
	}

	// Create new verification record
	verification := &entity.EmailVerification{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(uc.verificationConfig.TokenTTL),
	}

	if err := uc.emailVerificationRepo.Create(ctx, verification); err != nil {
		return fmt.Errorf("ResendVerification - Create: %w", err)
	}

	// Queue email sending task if asynq is configured
	// TODO: Enqueue email verification task when asynq is available
	_ = uc.asynqClient // Placeholder for future async email task

	return nil
}
