// Package entity defines main entities for business logic (services), database mapping and
// HTTP response objects if suitable. Each logic group entities in own file.
package entity

import (
	"time"

	"gorm.io/gorm"
)

// Translation represents a translation record.
type Translation struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Source      string         `json:"source" gorm:"size:255" example:"auto"`
	Destination string         `json:"destination" gorm:"size:255" example:"en"`
	Original    string         `json:"original" gorm:"size:1000" example:"текст для перевода"`
	Translation string         `json:"translation" gorm:"size:1000" example:"text for translation"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName overrides the table name used by GORM.
func (Translation) TableName() string {
	return "history"
}
