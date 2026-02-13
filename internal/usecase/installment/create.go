package installment

import (
	"context"
	"fmt"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/internal/entity"
)

func (uc *UseCase) Create(ctx context.Context, req installmentdto.CreateRequest) (*installmentdto.Response, error) {
	inst := &entity.Installment{
		UserID:         req.UserID,
		Name:           req.Name,
		TotalAmount:    req.TotalAmount,
		MonthlyAmount:  req.MonthlyAmount,
		TotalTerms:     req.TotalTerms,
		CompletedTerms: req.CompletedTerms,
		Status:         "active",
	}
	if req.Merchant != "" {
		inst.Merchant = &req.Merchant
	}
	if req.StartDate != "" {
		inst.StartDate = &req.StartDate
	}
	if req.EndDate != "" {
		inst.EndDate = &req.EndDate
	}
	if req.Notes != "" {
		inst.Notes = &req.Notes
	}

	if err := uc.installmentRepo.Create(ctx, inst); err != nil {
		return nil, fmt.Errorf("installment - Create - installmentRepo.Create: %w", err)
	}

	return installmentdto.NewResponse(inst), nil
}
