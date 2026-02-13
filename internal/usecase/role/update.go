package role

import (
	"context"
	"errors"
	"fmt"

	roledto "go-boilerplate/internal/dto/role"
	"go-boilerplate/internal/repo"
)

// Update updates an existing role.
func (uc *UseCase) Update(ctx context.Context, id uint, req roledto.UpdateRequest) (*roledto.Response, error) {
	// Get existing role
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("role - Update - GetByID: %w", err)
	}

	// Check name uniqueness if changing
	if req.Name != nil && *req.Name != role.Name {
		_, err := uc.roleRepo.GetByName(ctx, *req.Name)
		if err == nil {
			return nil, ErrRoleNameExists
		}
		if !errors.Is(err, repo.ErrNotFound) {
			return nil, fmt.Errorf("role - Update - GetByName: %w", err)
		}
		role.Name = *req.Name
	}

	// Update description if provided
	if req.Description != nil {
		role.Description = *req.Description
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("role - Update - roleRepo.Update: %w", err)
	}

	return roledto.NewResponse(role), nil
}
