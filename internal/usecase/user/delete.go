package user

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/repo"
)

// Delete deletes a user by ID.
func (uc *UseCase) Delete(ctx context.Context, id, currentUserID uint) error {
	// Prevent deleting self
	if id == currentUserID {
		return ErrCannotDeleteSelf
	}

	// Check if user exists
	_, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("user - Delete - GetByID: %w", err)
	}

	if err := uc.userRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("user - Delete - userRepo.Delete: %w", err)
	}

	return nil
}
