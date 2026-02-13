package bankstatement

import (
	"io"

	"go-boilerplate/pkg/pagination"
)

type UploadRequest struct {
	File     io.Reader `json:"-"`
	Filename string    `json:"-"`
	Size     int64     `json:"-"`
	BankID   uint      `json:"bank_id" validate:"required"`
	Password string    `json:"password"`
	UserID   uint      `json:"-"`
}

type ListRequest struct {
	pagination.Params
	Status string `query:"status" validate:"omitempty,oneof=pending completed failed"`
	UserID uint   `json:"-"`
}

type UpdateLineItemRequest struct {
	Category      *string `json:"category"`
	IsInstallment *bool   `json:"is_installment"`
	InstallmentID *uint   `json:"installment_id"`
}
