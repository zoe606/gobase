package persistent

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"
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
	db := tx.DBFromContext(ctx, r.db)
	var role entity.Role
	if err := db.Preload("Permissions").First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("RoleRepo - GetByID: %w", err)
	}
	return &role, nil
}

// GetByName retrieves a role by name with permissions.
func (r *RoleRepo) GetByName(ctx context.Context, name string) (*entity.Role, error) {
	db := tx.DBFromContext(ctx, r.db)
	var role entity.Role
	if err := db.Preload("Permissions").Where("name = ?", name).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("RoleRepo - GetByName: %w", err)
	}
	return &role, nil
}

// GetAll retrieves all roles with permissions.
func (r *RoleRepo) GetAll(ctx context.Context) ([]entity.Role, error) {
	db := tx.DBFromContext(ctx, r.db)
	var roles []entity.Role
	if err := db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("RoleRepo - GetAll: %w", err)
	}
	return roles, nil
}
