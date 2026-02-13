package permission

import (
	"context"
	"errors"
	"fmt"

	permissiondto "go-boilerplate/internal/dto/permission"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// Create creates a new permission.
func (uc *UseCase) Create(ctx context.Context, req permissiondto.CreateRequest) (*permissiondto.Response, error) {
	name := req.Resource + ":" + req.Action

	// Check if permission name already exists
	_, err := uc.permissionRepo.GetByName(ctx, name)
	if err == nil {
		return nil, ErrPermissionExists
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, fmt.Errorf("permission - Create - GetByName: %w", err)
	}

	p := &entity.Permission{
		Name:     name,
		Resource: req.Resource,
		Action:   req.Action,
	}

	if err := uc.permissionRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("permission - Create - permissionRepo.Create: %w", err)
	}

	return permissiondto.NewResponse(p), nil
}
