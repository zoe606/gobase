package installment

import (
	"context"
	"fmt"

	installmentdto "go-boilerplate/internal/dto/installment"
)

func (uc *UseCase) Update(ctx context.Context, id uint, req installmentdto.UpdateRequest) (*installmentdto.Response, error) {
	inst, err := uc.installmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("installment - Update - installmentRepo.GetByID: %w", err)
	}

	if req.Name != nil {
		inst.Name = *req.Name
	}
	if req.Merchant != nil {
		inst.Merchant = req.Merchant
	}
	if req.TotalAmount != nil {
		inst.TotalAmount = *req.TotalAmount
	}
	if req.MonthlyAmount != nil {
		inst.MonthlyAmount = *req.MonthlyAmount
	}
	if req.TotalTerms != nil {
		inst.TotalTerms = *req.TotalTerms
	}
	if req.CompletedTerms != nil {
		inst.CompletedTerms = *req.CompletedTerms
	}
	if req.StartDate != nil {
		inst.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		inst.EndDate = req.EndDate
	}
	if req.Status != nil {
		inst.Status = *req.Status
	}
	if req.Notes != nil {
		inst.Notes = req.Notes
	}

	if err := uc.installmentRepo.Update(ctx, inst); err != nil {
		return nil, fmt.Errorf("installment - Update - installmentRepo.Update: %w", err)
	}

	return installmentdto.NewResponse(inst), nil
}
