package persistent

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"

	"gorm.io/gorm"
)

// RoleRepo implements role repository using GORM.
type RoleRepo struct {
	db *gorm.DB
}

// NewRoleRepo creates a new role repository.
func NewRoleRepo(db *gorm.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

// GetByID retrieves a role by ID with permissions.
func (r *RoleRepo) GetByID(ctx context.Context, id uint) (*entity.Role, error) {
	var role entity.Role
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("RoleRepo - GetByID: %w", result.Error)
	}
	return &role, nil
}

// GetByName retrieves a role by name with permissions.
func (r *RoleRepo) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	var role entity.Role
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("name = ?", name).
		First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("RoleRepo - GetByName: %w", result.Error)
	}
	return &role, nil
}

// GetAll retrieves all roles with permissions.
func (r *RoleRepo) GetAll(ctx context.Context) ([]entity.Role, error) {
	var roles []entity.Role
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		Find(&roles)
	if result.Error != nil {
		return nil, fmt.Errorf("RoleRepo - GetAll: %w", result.Error)
	}
	return roles, nil
}
