package user

import (
	"context"
	"errors"
	"fmt"

	userdto "go-boilerplate/internal/dto/user"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// Create creates a new user.
func (uc *UseCase) Create(ctx context.Context, req userdto.CreateRequest) (*userdto.Response, error) {
	// Check if email already exists
	exists, err := uc.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user - Create - EmailExists: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Validate role exists
	role, err := uc.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, fmt.Errorf("user - Create - GetRole: %w", err)
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("user - Create - hashPassword: %w", err)
	}

	// Set active default if not provided
	active := true
	if req.Active != nil {
		active = *req.Active
	}

	user := &entity.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		RoleID:   req.RoleID,
		Active:   active,
		Role:     *role,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("user - Create - userRepo.Create: %w", err)
	}

	return userdto.NewResponse(user), nil
}
