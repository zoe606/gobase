package entity

import (
	"time"
)

// RefreshToken represents a refresh token for JWT authentication.
type RefreshToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	Token     string    `json:"-" gorm:"uniqueIndex;size:255;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the table name used by GORM.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks if the refresh token has expired.
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}
