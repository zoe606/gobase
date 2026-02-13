package role

import (
	"context"
	"errors"
	"fmt"

	roledto "go-boilerplate/internal/dto/role"
	"go-boilerplate/internal/repo"
)

// AssignPermissions assigns permissions to a role.
func (uc *UseCase) AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) (*roledto.Response, error) {
	// Check if role exists
	_, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("role - AssignPermissions - GetByID: %w", err)
	}

	// Update permissions
	if err := uc.roleRepo.UpdatePermissions(ctx, roleID, permissionIDs); err != nil {
		return nil, fmt.Errorf("role - AssignPermissions - UpdatePermissions: %w", err)
	}

	// Fetch updated role
	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("role - AssignPermissions - GetByID after update: %w", err)
	}

	return roledto.NewResponse(role), nil
}
