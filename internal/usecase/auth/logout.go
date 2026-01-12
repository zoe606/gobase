package auth

import (
	"context"
	"fmt"
)

// Logout invalidates a refresh token.
func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	if err := uc.refreshTokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("auth - Logout - DeleteByToken: %w", err)
	}
	return nil
}
