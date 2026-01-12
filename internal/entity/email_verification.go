package entity

import (
	"time"
)

// EmailVerification represents an email verification token.
type EmailVerification struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	UserID    uint       `json:"user_id" gorm:"not null;index"`
	User      *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Token     string     `json:"-" gorm:"uniqueIndex;size:64;not null"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName overrides the table name used by GORM.
func (EmailVerification) TableName() string {
	return "email_verifications"
}

// IsExpired checks if the verification token has expired.
func (e *EmailVerification) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// IsUsed checks if the verification token has been used.
func (e *EmailVerification) IsUsed() bool {
	return e.UsedAt != nil
}
