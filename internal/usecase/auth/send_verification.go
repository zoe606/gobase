package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// SendVerificationEmail sends a verification email to the user.
// If auto_verify is enabled (development mode), it marks the email as verified immediately.
func (uc *UseCase) SendVerificationEmail(ctx context.Context, userID uint) error {
	if !uc.verificationConfig.Enabled {
		return nil // Verification disabled, silently succeed
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return fmt.Errorf("SendVerificationEmail: user not found")
		}
		return fmt.Errorf("SendVerificationEmail - GetByID: %w", err)
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
			return fmt.Errorf("SendVerificationEmail - Update: %w", err)
		}
		return nil
	}

	// Generate verification token
	token, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("SendVerificationEmail - generateSecureToken: %w", err)
	}

	// Create verification record
	verification := &entity.EmailVerification{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(uc.verificationConfig.TokenTTL),
	}

	if uc.emailVerificationRepo == nil {
		return fmt.Errorf("SendVerificationEmail: email verification repository not configured")
	}

	if err := uc.emailVerificationRepo.Create(ctx, verification); err != nil {
		return fmt.Errorf("SendVerificationEmail - Create: %w", err)
	}

	// Queue email sending task if asynq is configured
	if uc.asynqClient != nil {
		// TODO: Enqueue email verification task
		// This would be implemented similar to the welcome email task
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token.
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
