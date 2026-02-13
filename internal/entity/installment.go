package entity

import (
	"time"

	"gorm.io/gorm"
)

type Installment struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         uint           `json:"user_id" gorm:"not null"`
	Name           string         `json:"name" gorm:"size:255;not null"`
	Merchant       *string        `json:"merchant,omitempty" gorm:"size:255"`
	TotalAmount    float64        `json:"total_amount" gorm:"type:numeric(15,2);not null"`
	MonthlyAmount  float64        `json:"monthly_amount" gorm:"type:numeric(15,2);not null"`
	TotalTerms     int            `json:"total_terms" gorm:"not null"`
	CompletedTerms int            `json:"completed_terms" gorm:"default:0"`
	StartDate      *string        `json:"start_date,omitempty" gorm:"type:date"`
	EndDate        *string        `json:"end_date,omitempty" gorm:"type:date"`
	Status         string         `json:"status" gorm:"size:20;default:'active'"`
	Notes          *string        `json:"notes,omitempty"`
	User           *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Installment) TableName() string {
	return "installments"
}
