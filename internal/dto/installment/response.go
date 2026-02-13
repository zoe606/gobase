package installment

import (
	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/pagination"
)

type Response struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Merchant       *string `json:"merchant"`
	TotalAmount    float64 `json:"total_amount"`
	MonthlyAmount  float64 `json:"monthly_amount"`
	TotalTerms     int     `json:"total_terms"`
	CompletedTerms int     `json:"completed_terms"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	Status         string  `json:"status"`
	Notes          *string `json:"notes"`
	CreatedAt      string  `json:"created_at"`
}

func NewResponse(inst *entity.Installment) *Response {
	if inst == nil {
		return nil
	}
	return &Response{
		ID:             inst.ID,
		Name:           inst.Name,
		Merchant:       inst.Merchant,
		TotalAmount:    inst.TotalAmount,
		MonthlyAmount:  inst.MonthlyAmount,
		TotalTerms:     inst.TotalTerms,
		CompletedTerms: inst.CompletedTerms,
		StartDate:      inst.StartDate,
		EndDate:        inst.EndDate,
		Status:         inst.Status,
		Notes:          inst.Notes,
		CreatedAt:      inst.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type ResponseWithItems struct {
	Response
	LinkedItems int `json:"linked_items"`
}

type ListResponse struct {
	Data []*Response      `json:"data"`
	Meta *pagination.Meta `json:"meta"`
}

func NewListResponse(installments []*entity.Installment, total int64, params pagination.Params) *ListResponse {
	data := make([]*Response, len(installments))
	for i, inst := range installments {
		data[i] = NewResponse(inst)
	}
	return &ListResponse{
		Data: data,
		Meta: pagination.NewMeta(params.Page, params.Limit, total),
	}
}
