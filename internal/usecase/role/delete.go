package role

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/repo"
)

// Delete deletes a role by ID.
func (uc *UseCase) Delete(ctx context.Context, id uint) error {
	// Check if role exists
	_, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("role - Delete - GetByID: %w", err)
	}

	if err := uc.roleRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("role - Delete - roleRepo.Delete: %w", err)
	}

	return nil
}
