package entity

import (
	"time"
)

// PasswordReset represents a password reset token.
type PasswordReset struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id" gorm:"not null;index"`
	User      *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Token     string     `json:"-" gorm:"uniqueIndex;size:64;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the table name used by GORM.
func (PasswordReset) TableName() string {
	return "password_resets"
}

// IsExpired checks if the reset token has expired.
func (p *PasswordReset) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// IsUsed checks if the reset token has been used.
func (p *PasswordReset) IsUsed() bool {
	return p.UsedAt != nil
}
