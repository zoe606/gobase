package auth

import (
	"context"

	"go-boilerplate/internal/worker/tasks"
)

func (uc *UseCase) RequestPasswordReset(ctx context.Context, email string) error {
	if _, err := uc.emailLink(uc.resetConfig.BaseURL, "reset-password", ""); err != nil || uc.resetConfig.TokenTTL <= 0 {
		return ErrEmailUnavailable
	}
	return uc.enqueueAuthEmail(ctx, tasks.TypePasswordReset, tasks.AuthEmailPayload{Email: email})
}
