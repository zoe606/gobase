package entity

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;size:255;not null"`
	Password  string         `json:"-" gorm:"size:255;not null"`
	Name      string         `json:"name" gorm:"size:255"`
	RoleID    uint           `json:"role_id" gorm:"not null"`
	Role      Role           `json:"role" gorm:"foreignKey:RoleID"`
	Active    bool           `json:"active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName overrides the table name used by GORM.
func (User) TableName() string {
	return "users"
}
