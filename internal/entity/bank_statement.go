package entity

import (
	"time"

	"gorm.io/gorm"
)

const (
	BankStatementStatusPending   = "pending"
	BankStatementStatusCompleted = "completed"
	BankStatementStatusFailed    = "failed"
)

type BankStatement struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	BankID       uint           `json:"bank_id" gorm:"not null"`
	MediaID      *uint          `json:"media_id,omitempty"`
	PeriodStart  *string        `json:"period_start,omitempty" gorm:"type:date"`
	PeriodEnd    *string        `json:"period_end,omitempty" gorm:"type:date"`
	Password     *string        `json:"-" gorm:"size:255"`
	Status       string         `json:"status" gorm:"size:20;default:'pending'"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	TotalDebit   float64        `json:"total_debit" gorm:"type:numeric(15,2);default:0"`
	TotalCredit  float64        `json:"total_credit" gorm:"type:numeric(15,2);default:0"`
	Bank         *Bank          `json:"bank,omitempty" gorm:"foreignKey:BankID"`
	Media        *Media         `json:"media,omitempty" gorm:"foreignKey:MediaID"`
	User         *User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (BankStatement) TableName() string {
	return "bank_statements"
}
