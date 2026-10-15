package auth

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/repo"

	"golang.org/x/crypto/bcrypt"
)

// ResetPasswordInput holds the input for resetting a password.
type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

// ResetPassword resets a user's password using the provided token.
func (uc *UseCase) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	return uc.inTransaction(ctx, func(ctx context.Context) error { return uc.resetPassword(ctx, input) })
}

func (uc *UseCase) resetPassword(ctx context.Context, input ResetPasswordInput) error {
	if uc.passwordResetRepo == nil {
		return fmt.Errorf("ResetPassword: password reset repository not configured")
	}

	// Get reset record
	reset, err := uc.passwordResetRepo.GetByToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrResetTokenNotFound
		}
		return fmt.Errorf("ResetPassword - GetByToken: %w", err)
	}

	// Check if already used
	if reset.IsUsed() {
		return ErrResetTokenUsed
	}

	// Check if expired
	if reset.IsExpired() {
		return ErrResetTokenExpired
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, reset.UserID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return fmt.Errorf("ResetPassword: user not found")
		}
		return fmt.Errorf("ResetPassword - GetByID: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("ResetPassword - GenerateFromPassword: %w", err)
	}

	// Mark token as used
	if err := uc.passwordResetRepo.MarkAsUsed(ctx, reset.ID); err != nil {
		if errors.Is(err, repo.ErrTokenConsumed) {
			return ErrResetTokenUsed
		}
		return fmt.Errorf("ResetPassword - MarkAsUsed: %w", err)
	}
	if err := uc.userRepo.SetPassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return fmt.Errorf("ResetPassword - Update: %w", err)
	}

	// Invalidate all refresh tokens for security
	if err := uc.refreshTokenRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("ResetPassword - DeleteByUserID: %w", err)
	}
	if err := uc.passwordResetRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("ResetPassword - Delete reset tokens: %w", err)
	}

	return nil
}
