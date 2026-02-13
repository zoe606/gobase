package installment

import (
	"go-boilerplate/pkg/pagination"
)

type CreateRequest struct {
	Name           string  `json:"name" validate:"required"`
	Merchant       string  `json:"merchant"`
	TotalAmount    float64 `json:"total_amount" validate:"required,gt=0"`
	MonthlyAmount  float64 `json:"monthly_amount" validate:"required,gt=0"`
	TotalTerms     int     `json:"total_terms" validate:"required,gt=0"`
	CompletedTerms int     `json:"completed_terms" validate:"min=0"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
	Notes          string  `json:"notes"`
	UserID         uint    `json:"-"`
}

type UpdateRequest struct {
	Name           *string  `json:"name,omitempty"`
	Merchant       *string  `json:"merchant,omitempty"`
	TotalAmount    *float64 `json:"total_amount,omitempty"`
	MonthlyAmount  *float64 `json:"monthly_amount,omitempty"`
	TotalTerms     *int     `json:"total_terms,omitempty"`
	CompletedTerms *int     `json:"completed_terms,omitempty"`
	StartDate      *string  `json:"start_date,omitempty"`
	EndDate        *string  `json:"end_date,omitempty"`
	Status         *string  `json:"status,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}

type ListRequest struct {
	pagination.Params
	Status string `query:"status" validate:"omitempty,oneof=active completed canceled"`
	UserID uint   `json:"-"`
}

type LinkItemsRequest struct {
	LineItemIDs []uint `json:"line_item_ids" validate:"required,min=1"`
}
