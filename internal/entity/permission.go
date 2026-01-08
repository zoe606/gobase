package entity

import (
	"time"
)

// Permission represents a permission that can be assigned to roles.
// Permissions follow the format "resource:action" (e.g., "user:read", "translation:write").
type Permission struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Resource  string    `json:"resource" gorm:"size:50;not null"`
	Action    string    `json:"action" gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName overrides the table name used by GORM.
func (Permission) TableName() string {
	return "permissions"
}
