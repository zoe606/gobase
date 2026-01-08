package auth

import (
	"context"
	"errors"
	"fmt"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/repo"
)

// GetCurrentUser retrieves the current user by ID.
func (uc *UseCase) GetCurrentUser(ctx context.Context, userID uint) (*authdto.UserResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("Auth - GetCurrentUser - GetByID: %w", err)
	}
	resp := authdto.NewUserResponse(user)
	return &resp, nil
}
