package entity

import (
	"time"
)

type LineItem struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	SourceType    string    `json:"source_type" gorm:"size:100;not null;index:idx_line_items_source"`
	SourceID      uint      `json:"source_id" gorm:"not null;index:idx_line_items_source"`
	Date          string    `json:"date" gorm:"type:date;not null"`
	Description   string    `json:"description" gorm:"not null"`
	Category      *string   `json:"category,omitempty" gorm:"size:100"`
	Debit         float64   `json:"debit" gorm:"type:numeric(15,2);default:0"`
	Credit        float64   `json:"credit" gorm:"type:numeric(15,2);default:0"`
	Balance       float64   `json:"balance" gorm:"type:numeric(15,2);default:0"`
	IsInstallment bool      `json:"is_installment" gorm:"default:false"`
	InstallmentID *uint     `json:"installment_id,omitempty" gorm:"index"`
	Metadata      JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (LineItem) TableName() string {
	return "line_items"
}
