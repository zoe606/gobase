package auth

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/repo"
)

func (uc *UseCase) ResendVerification(ctx context.Context, email string) error {
	if !uc.verificationConfig.Enabled {
		return ErrVerificationDisabled
	}
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, repo.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("resend verification user: %w", err)
	}
	return uc.SendVerificationEmail(ctx, user.ID)
}
