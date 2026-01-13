package entity

import (
	"time"
)

// Article Blog articles/posts with status workflow
type Article struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	UserID       uint       `json:"user_id" gorm:"not null"`
	Title        string     `json:"title" gorm:"not null;size:255"`
	Slug         string     `json:"slug" gorm:"uniqueIndex;not null;size:255"`
	Content      *string    `json:"content,omitempty"`
	Excerpt      *string    `json:"excerpt,omitempty" gorm:"size:500"`
	CoverMediaID *uint      `json:"cover_media_id,omitempty"`
	Status       *string    `json:"status,omitempty" gorm:"size:20;default:'draft'"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	ViewCount    *int       `json:"view_count,omitempty" gorm:"default:0"`
	CreatedAt    time.Time  `json:"created_at,omitempty" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty" gorm:"autoUpdateTime"`
	User         *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CoverMedia   *Media     `json:"coverMedia,omitempty" gorm:"foreignKey:CoverMediaID"`
}

// TableName returns the table name.
func (Article) TableName() string {
	return "articles"
}
