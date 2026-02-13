package entity

import (
	"time"

	"gorm.io/gorm"
)

type Bank struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"size:100;not null"`
	Code            string         `json:"code" gorm:"uniqueIndex;size:20;not null"`
	DefaultPassword *string        `json:"default_password,omitempty" gorm:"size:255"`
	CreatedAt       time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Bank) TableName() string {
	return "banks"
}
