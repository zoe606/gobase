package entity

import (
	"time"
)

// Profile represents personalized user data, separate from auth-related User data.
type Profile struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	UserID        uint   `json:"user_id" gorm:"uniqueIndex;not null"`
	User          *User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	AvatarMediaID *uint  `json:"avatar_media_id,omitempty" gorm:"index"`
	Avatar        *Media `json:"avatar,omitempty" gorm:"foreignKey:AvatarMediaID"`
	Bio           string `json:"bio,omitempty"`
	Phone         string `json:"phone,omitempty" gorm:"size:20"`

	// Timestamps.
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name.
func (Profile) TableName() string {
	return "profiles"
}

// HasAvatar returns true if the profile has an avatar.
func (p *Profile) HasAvatar() bool {
	return p.AvatarMediaID != nil
}
