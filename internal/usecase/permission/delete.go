package permission

import (
	"context"
	"fmt"
)

// Delete deletes a permission by ID.
func (uc *UseCase) Delete(ctx context.Context, id uint) error {
	// Check if permission exists
	permissions, err := uc.permissionRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("permission - Delete - List: %w", err)
	}

	found := false
	for _, p := range permissions {
		if p.ID == id {
			found = true
			break
		}
	}
	if !found {
		return ErrPermissionNotFound
	}

	// Check if permission is assigned to any role
	inUse, err := uc.permissionRepo.IsAssignedToAnyRole(ctx, id)
	if err != nil {
		return fmt.Errorf("permission - Delete - IsAssignedToAnyRole: %w", err)
	}
	if inUse {
		return ErrPermissionInUse
	}

	if err := uc.permissionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("permission - Delete - permissionRepo.Delete: %w", err)
	}

	return nil
}
