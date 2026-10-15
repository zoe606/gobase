package auth

import (
	"context"
	"fmt"
	"time"

	"go-boilerplate/internal/worker/tasks"
)

func (uc *UseCase) SendVerificationEmail(ctx context.Context, userID uint) error {
	if !uc.verificationConfig.Enabled {
		return ErrVerificationDisabled
	}
	if uc.verificationConfig.AutoVerify {
		if uc.production {
			return ErrEmailUnavailable
		}
		return uc.inTransaction(ctx, func(ctx context.Context) error {
			user, err := uc.userRepo.GetByID(ctx, userID)
			if err != nil {
				return fmt.Errorf("auto verification user: %w", err)
			}
			if user.IsEmailVerified() {
				return nil
			}
			now := time.Now()
			return uc.userRepo.SetEmailVerified(ctx, user.ID, now)
		})
	}
	if _, err := uc.emailLink(uc.verificationConfig.BaseURL, "verify-email", ""); err != nil || uc.verificationConfig.TokenTTL <= 0 {
		return ErrEmailUnavailable
	}
	return uc.enqueueAuthEmail(ctx, tasks.TypeEmailVerification, tasks.AuthEmailPayload{UserID: userID})
}
