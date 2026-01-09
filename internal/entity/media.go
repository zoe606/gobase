package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MediaType represents the type of media.
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeDocument MediaType = "document"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeOther    MediaType = "other"
)

// Media represents a file attachment (polymorphic - can belong to any entity).
type Media struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Polymorphic relationship (like Laravel's morphTo).
	AttachableType string `json:"attachable_type" gorm:"index;size:100"` // e.g., "users", "posts"
	AttachableID   uint   `json:"attachable_id" gorm:"index"`
	Collection     string `json:"collection" gorm:"index;size:50"` // e.g., "avatar", "gallery", "documents"

	// File metadata.
	Filename     string `json:"filename" gorm:"size:255;not null"`
	OriginalName string `json:"original_name" gorm:"size:255;not null"`
	MimeType     string `json:"mime_type" gorm:"size:100;not null"`
	Size         int64  `json:"size" gorm:"not null"` // bytes

	// Storage info.
	Disk string `json:"disk" gorm:"size:50;not null"` // "local", "s3"
	Path string `json:"path" gorm:"size:500;not null"` // storage path

	// Media specific.
	Type MediaType `json:"type" gorm:"size:20;not null"`
	Hash string    `json:"hash" gorm:"size:64;index"` // SHA256 for deduplication

	// Image specific (nullable for non-images).
	Width  *int `json:"width,omitempty"`
	Height *int `json:"height,omitempty"`

	// Variants/conversions (JSON - thumbnails, etc.).
	Variants JSONMap `json:"variants,omitempty" gorm:"type:jsonb"` // {"thumb": "path", "medium": "path"}

	// Metadata (JSON - EXIF, custom data).
	Metadata JSONMap `json:"metadata,omitempty" gorm:"type:jsonb"`

	// Timestamps.
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name.
func (Media) TableName() string {
	return "media"
}

// IsImage returns true if media is an image.
func (m *Media) IsImage() bool {
	return m.Type == MediaTypeImage
}

// GetVariantPath returns the path for a specific variant.
func (m *Media) GetVariantPath(variant string) string {
	if m.Variants == nil {
		return ""
	}
	if path, ok := m.Variants[variant].(string); ok {
		return path
	}
	return ""
}

// JSONMap is a custom type for GORM jsonb columns.
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface.
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface.
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONMap value: %v", value)
	}

	return json.Unmarshal(bytes, j)
}
