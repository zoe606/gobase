package user

import (
	"context"
	"errors"
	"fmt"

	userdto "go-boilerplate/internal/dto/user"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// Update updates an existing user.
func (uc *UseCase) Update(ctx context.Context, id uint, req userdto.UpdateRequest) (*userdto.Response, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user - Update - GetByID: %w", err)
	}

	if err := uc.applyUpdates(ctx, user, req); err != nil {
		return nil, err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("user - Update - userRepo.Update: %w", err)
	}

	return userdto.NewResponse(user), nil
}

// applyUpdates applies the update request fields to the user.
func (uc *UseCase) applyUpdates(ctx context.Context, user *entity.User, req userdto.UpdateRequest) error {
	if err := uc.updateEmail(ctx, user, req.Email); err != nil {
		return err
	}

	if err := uc.updateRole(ctx, user, req.RoleID); err != nil {
		return err
	}

	if err := uc.updatePassword(user, req.Password); err != nil {
		return err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Active != nil {
		user.Active = *req.Active
	}

	return nil
}

// updateEmail validates and updates the email if changed.
func (uc *UseCase) updateEmail(ctx context.Context, user *entity.User, email *string) error {
	if email == nil || *email == user.Email {
		return nil
	}

	exists, err := uc.userRepo.EmailExists(ctx, *email)
	if err != nil {
		return fmt.Errorf("user - Update - EmailExists: %w", err)
	}

	if exists {
		return ErrEmailExists
	}

	user.Email = *email

	return nil
}

// updateRole validates and updates the role if changed.
func (uc *UseCase) updateRole(ctx context.Context, user *entity.User, roleID *uint) error {
	if roleID == nil || *roleID == user.RoleID {
		return nil
	}

	role, err := uc.roleRepo.GetByID(ctx, *roleID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("user - Update - GetRole: %w", err)
	}

	user.RoleID = *roleID
	user.Role = *role

	return nil
}

// updatePassword hashes and updates the password if provided.
func (uc *UseCase) updatePassword(user *entity.User, password *string) error {
	if password == nil || *password == "" {
		return nil
	}

	hashedPassword, err := hashPassword(*password)
	if err != nil {
		return fmt.Errorf("user - Update - hashPassword: %w", err)
	}

	user.Password = hashedPassword

	return nil
}
