package role

import (
	"context"
	"fmt"

	roledto "go-boilerplate/internal/dto/role"
)

// List retrieves all roles.
func (uc *UseCase) List(ctx context.Context) (*roledto.ListResponse, error) {
	roles, err := uc.roleRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("role - List - roleRepo.List: %w", err)
	}

	return roledto.NewListResponse(roles), nil
}
