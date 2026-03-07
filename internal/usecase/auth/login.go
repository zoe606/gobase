package auth

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/hasher"
)

// Login authenticates a user.
func (uc *UseCase) Login(ctx context.Context, input authdto.LoginRequest) (*authdto.LoginResponse, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth - Login - GetByEmail: %w", err)
	}

	// Check password
	if !hasher.Check(input.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.Active {
		return nil, ErrUserNotActive
	}

	// Generate tokens
	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("auth - Login - %w", err)
	}

	// Store refresh token
	if err := uc.storeRefreshToken(ctx, user.ID, tokens.RefreshToken, tokens.RefreshExpiresAt); err != nil {
		return nil, fmt.Errorf("auth - Login - storeRefreshToken: %w", err)
	}

	return authdto.NewLoginResponse(user, tokens.AccessToken, tokens.RefreshToken, tokens.AccessExpiresAt), nil
}
