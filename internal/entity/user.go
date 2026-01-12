package entity

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system.
type User struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Email           string         `json:"email" gorm:"uniqueIndex;size:255;not null"`
	Password        string         `json:"-" gorm:"size:255;not null"`
	Name            string         `json:"name" gorm:"size:255"`
	RoleID          uint           `json:"role_id" gorm:"not null"`
	Role            Role           `json:"role" gorm:"foreignKey:RoleID"`
	Active          bool           `json:"active" gorm:"default:true"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// IsEmailVerified returns true if the user's email has been verified.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// TableName overrides the table name used by GORM.
func (User) TableName() string {
	return "users"
}
