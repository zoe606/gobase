package permission

import (
	"context"
	"fmt"

	"go-boilerplate/internal/entity"
)

// List retrieves all permissions.
func (uc *UseCase) List(ctx context.Context) ([]*entity.Permission, error) {
	permissions, err := uc.permissionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("permission - List - permissionRepo.List: %w", err)
	}

	return permissions, nil
}
