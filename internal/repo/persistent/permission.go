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

// PermissionRepo implements permission repository using GORM.
type PermissionRepo struct {
	db *gorm.DB
}

// NewPermissionRepo creates a new permission repository.
func NewPermissionRepo(db *gorm.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

// List retrieves all permissions.
func (r *PermissionRepo) List(ctx context.Context) ([]*entity.Permission, error) {
	db := tx.DBFromContext(ctx, r.db)
	var permissions []*entity.Permission
	if err := db.Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("PermissionRepo - List: %w", err)
	}
	return permissions, nil
}

// GetByIDs retrieves permissions by their IDs.
func (r *PermissionRepo) GetByIDs(ctx context.Context, ids []uint) ([]*entity.Permission, error) {
	db := tx.DBFromContext(ctx, r.db)
	var permissions []*entity.Permission
	if err := db.Where("id IN ?", ids).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("PermissionRepo - GetByIDs: %w", err)
	}
	return permissions, nil
}

// Create creates a new permission.
func (r *PermissionRepo) Create(ctx context.Context, permission *entity.Permission) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(permission).Error; err != nil {
		return fmt.Errorf("PermissionRepo - Create: %w", err)
	}
	return nil
}

// Delete deletes a permission by ID.
func (r *PermissionRepo) Delete(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Delete(&entity.Permission{}, id)
	if result.Error != nil {
		return fmt.Errorf("PermissionRepo - Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// GetByName retrieves a permission by name.
func (r *PermissionRepo) GetByName(ctx context.Context, name string) (*entity.Permission, error) {
	db := tx.DBFromContext(ctx, r.db)
	var permission entity.Permission
	if err := db.Where("name = ?", name).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("PermissionRepo - GetByName: %w", err)
	}
	return &permission, nil
}

// IsAssignedToAnyRole checks if a permission is assigned to any role.
func (r *PermissionRepo) IsAssignedToAnyRole(ctx context.Context, permissionID uint) (bool, error) {
	db := tx.DBFromContext(ctx, r.db)
	var count int64
	if err := db.Table("role_permissions").Where("permission_id = ?", permissionID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("PermissionRepo - IsAssignedToAnyRole: %w", err)
	}
	return count > 0, nil
}
