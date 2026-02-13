package role

import (
	"context"
	"errors"
	"fmt"

	roledto "go-boilerplate/internal/dto/role"
	"go-boilerplate/internal/repo"
)

// GetByID retrieves a role by ID.
func (uc *UseCase) GetByID(ctx context.Context, id uint) (*roledto.Response, error) {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("role - GetByID - roleRepo.GetByID: %w", err)
	}

	return roledto.NewResponse(role), nil
}
