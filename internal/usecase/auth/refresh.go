package auth

import (
	"context"
	"errors"
	"fmt"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/repo"
)

// Refresh refreshes an access token using a refresh token.
func (uc *UseCase) Refresh(ctx context.Context, input authdto.RefreshRequest) (*authdto.TokenResponse, error) {
	// Get refresh token
	token, err := uc.refreshTokenRepo.GetByToken(ctx, input.RefreshToken)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("Auth - Refresh - GetByToken: %w", err)
	}
	if token.IsExpired() {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("Auth - Refresh - GetByID: %w", err)
	}
	if !user.Active {
		return nil, ErrInvalidToken
	}

	// Delete old refresh token
	if err := uc.refreshTokenRepo.DeleteByToken(ctx, input.RefreshToken); err != nil {
		return nil, fmt.Errorf("Auth - Refresh - DeleteByToken: %w", err)
	}

	// Generate new tokens
	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("Auth - Refresh - %w", err)
	}

	// Store new refresh token
	if err := uc.storeRefreshToken(ctx, user.ID, tokens.RefreshToken, tokens.RefreshExpiresAt); err != nil {
		return nil, fmt.Errorf("Auth - Refresh - StoreRefreshToken: %w", err)
	}

	return authdto.NewTokenResponse(tokens.AccessToken, tokens.RefreshToken, tokens.AccessExpiresAt), nil
}
