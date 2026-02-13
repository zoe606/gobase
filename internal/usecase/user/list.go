package user

import (
	"context"
	"fmt"

	userdto "go-boilerplate/internal/dto/user"
)

// List retrieves a paginated list of users with filters.
func (uc *UseCase) List(ctx context.Context, req userdto.ListRequest) (*userdto.ListResponse, error) {
	req.Normalize()

	users, total, err := uc.userRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("user - List - userRepo.List: %w", err)
	}

	return userdto.NewListResponse(users, total, req.Params), nil
}
