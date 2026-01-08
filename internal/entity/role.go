package entity

import (
	"time"
)

// Role represents a user role with associated permissions.
type Role struct {
	ID          uint         `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"uniqueIndex;size:50;not null"`
	Description string       `json:"description" gorm:"size:255"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the table name used by GORM.
func (Role) TableName() string {
	return "roles"
}

// HasPermission checks if the role has a specific permission.
func (r *Role) HasPermission(permissionName string) bool {
	for _, p := range r.Permissions {
		if p.Name == permissionName {
			return true
		}
	}
	return false
}

// GetPermissionNames returns a slice of permission names.
func (r *Role) GetPermissionNames() []string {
	names := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		names[i] = p.Name
	}
	return names
}
