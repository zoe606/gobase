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

// List retrieves all roles with permissions.
func (r *RoleRepo) List(ctx context.Context) ([]*entity.Role, error) {
	db := tx.DBFromContext(ctx, r.db)
	var roles []*entity.Role
	if err := db.Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("RoleRepo - List: %w", err)
	}
	return roles, nil
}

// Create creates a new role.
func (r *RoleRepo) Create(ctx context.Context, role *entity.Role) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(role).Error; err != nil {
		return fmt.Errorf("RoleRepo - Create: %w", err)
	}
	return nil
}

// Update updates a role.
func (r *RoleRepo) Update(ctx context.Context, role *entity.Role) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Save(role).Error; err != nil {
		return fmt.Errorf("RoleRepo - Update: %w", err)
	}
	return nil
}

// Delete deletes a role by ID.
func (r *RoleRepo) Delete(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)

	// First clear the permissions association
	role := &entity.Role{ID: id}
	if err := db.Model(role).Association("Permissions").Clear(); err != nil {
		return fmt.Errorf("RoleRepo - Delete - clear permissions: %w", err)
	}

	result := db.Delete(&entity.Role{}, id)
	if result.Error != nil {
		return fmt.Errorf("RoleRepo - Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// UpdatePermissions updates the permissions assigned to a role.
func (r *RoleRepo) UpdatePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	db := tx.DBFromContext(ctx, r.db)

	var permissions []entity.Permission
	if len(permissionIDs) > 0 {
		if err := db.Where("id IN ?", permissionIDs).Find(&permissions).Error; err != nil {
			return fmt.Errorf("RoleRepo - UpdatePermissions - find permissions: %w", err)
		}
	}

	role := &entity.Role{ID: roleID}
	if err := db.Model(role).Association("Permissions").Replace(permissions); err != nil {
		return fmt.Errorf("RoleRepo - UpdatePermissions: %w", err)
	}
	return nil
}
