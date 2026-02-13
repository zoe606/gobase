package bankstatement

import (
	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/pagination"
)

type BankResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func NewBankResponse(bank *entity.Bank) *BankResponse {
	if bank == nil {
		return nil
	}
	return &BankResponse{
		ID:   bank.ID,
		Name: bank.Name,
		Code: bank.Code,
	}
}

type BankListResponse struct {
	Data []*BankResponse `json:"data"`
}

func NewBankListResponse(banks []*entity.Bank) *BankListResponse {
	data := make([]*BankResponse, len(banks))
	for i, b := range banks {
		data[i] = NewBankResponse(b)
	}
	return &BankListResponse{Data: data}
}

type Response struct {
	ID          uint    `json:"id"`
	BankName    string  `json:"bank_name"`
	BankCode    string  `json:"bank_code"`
	PeriodStart *string `json:"period_start"`
	PeriodEnd   *string `json:"period_end"`
	Status      string  `json:"status"`
	TotalDebit  float64 `json:"total_debit"`
	TotalCredit float64 `json:"total_credit"`
	CreatedAt   string  `json:"created_at"`
}

func NewResponse(stmt *entity.BankStatement) *Response {
	if stmt == nil {
		return nil
	}
	r := &Response{
		ID:          stmt.ID,
		PeriodStart: stmt.PeriodStart,
		PeriodEnd:   stmt.PeriodEnd,
		Status:      stmt.Status,
		TotalDebit:  stmt.TotalDebit,
		TotalCredit: stmt.TotalCredit,
		CreatedAt:   stmt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if stmt.Bank != nil {
		r.BankName = stmt.Bank.Name
		r.BankCode = stmt.Bank.Code
	}
	return r
}

type LineItemResponse struct {
	ID            uint           `json:"id"`
	Date          string         `json:"date"`
	Description   string         `json:"description"`
	Category      *string        `json:"category"`
	Debit         float64        `json:"debit"`
	Credit        float64        `json:"credit"`
	Balance       float64        `json:"balance"`
	IsInstallment bool           `json:"is_installment"`
	InstallmentID *uint          `json:"installment_id"`
	Metadata      entity.JSONMap `json:"metadata,omitempty"`
}

func NewLineItemResponse(item *entity.LineItem) *LineItemResponse {
	if item == nil {
		return nil
	}
	return &LineItemResponse{
		ID:            item.ID,
		Date:          item.Date,
		Description:   item.Description,
		Category:      item.Category,
		Debit:         item.Debit,
		Credit:        item.Credit,
		Balance:       item.Balance,
		IsInstallment: item.IsInstallment,
		InstallmentID: item.InstallmentID,
		Metadata:      item.Metadata,
	}
}

type ResponseWithItems struct {
	Response
	Items []*LineItemResponse `json:"items"`
}

func NewResponseWithItems(stmt *entity.BankStatement, items []*entity.LineItem) *ResponseWithItems {
	r := &ResponseWithItems{
		Response: *NewResponse(stmt),
	}
	r.Items = make([]*LineItemResponse, len(items))
	for i, item := range items {
		r.Items[i] = NewLineItemResponse(item)
	}
	return r
}

type ListResponse struct {
	Data []*Response      `json:"data"`
	Meta *pagination.Meta `json:"meta"`
}

func NewListResponse(stmts []*entity.BankStatement, total int64, params pagination.Params) *ListResponse {
	data := make([]*Response, len(stmts))
	for i, stmt := range stmts {
		data[i] = NewResponse(stmt)
	}
	return &ListResponse{
		Data: data,
		Meta: pagination.NewMeta(params.Page, params.Limit, total),
	}
}
