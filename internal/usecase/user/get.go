package user

import (
	"context"
	"errors"
	"fmt"

	userdto "go-boilerplate/internal/dto/user"
	"go-boilerplate/internal/repo"
)

// GetByID retrieves a user by ID.
func (uc *UseCase) GetByID(ctx context.Context, id uint) (*userdto.Response, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user - GetByID - userRepo.GetByID: %w", err)
	}

	return userdto.NewResponse(user), nil
}
