package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// RequestPasswordReset initiates a password reset for the given email.
// Always returns nil to prevent email enumeration attacks.
func (uc *UseCase) RequestPasswordReset(ctx context.Context, email string) error {
	if uc.passwordResetRepo == nil {
		return fmt.Errorf("RequestPasswordReset: password reset repository not configured")
	}

	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			// Don't reveal if email exists - return nil to prevent enumeration
			return nil
		}
		return fmt.Errorf("RequestPasswordReset - GetByEmail: %w", err)
	}

	// Delete any existing reset tokens for this user
	if err := uc.passwordResetRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("RequestPasswordReset - DeleteByUserID: %w", err)
	}

	// Generate reset token
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("RequestPasswordReset - generateSecureToken: %w", err)
	}

	// Create reset record
	reset := &entity.PasswordReset{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(uc.resetConfig.TokenTTL),
	}

	if err := uc.passwordResetRepo.Create(ctx, reset); err != nil {
		return fmt.Errorf("RequestPasswordReset - Create: %w", err)
	}

	// Queue password reset email task if asynq is configured
	if uc.asynqClient != nil {
		// TODO: Enqueue password reset email task
	}

	return nil
}
