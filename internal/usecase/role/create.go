package role

import (
	"context"
	"errors"
	"fmt"

	roledto "go-boilerplate/internal/dto/role"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// Create creates a new role.
func (uc *UseCase) Create(ctx context.Context, req roledto.CreateRequest) (*roledto.Response, error) {
	// Check if role name already exists
	_, err := uc.roleRepo.GetByName(ctx, req.Name)
	if err == nil {
		return nil, ErrRoleNameExists
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return nil, fmt.Errorf("role - Create - GetByName: %w", err)
	}

	role := &entity.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("role - Create - roleRepo.Create: %w", err)
	}

	return roledto.NewResponse(role), nil
}
