package bankstatement

import (
	"context"
	"fmt"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
)

func (uc *UseCase) UpdateLineItem(ctx context.Context, itemID uint, req bankstatementdto.UpdateLineItemRequest) (*bankstatementdto.LineItemResponse, error) {
	item, err := uc.lineItemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - UpdateLineItem - lineItemRepo.GetByID: %w", err)
	}

	if req.Category != nil {
		item.Category = req.Category
	}
	if req.IsInstallment != nil {
		item.IsInstallment = *req.IsInstallment
	}
	if req.InstallmentID != nil {
		item.InstallmentID = req.InstallmentID
		item.IsInstallment = true
	}

	if err := uc.lineItemRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("bankstatement - UpdateLineItem - lineItemRepo.Update: %w", err)
	}

	return bankstatementdto.NewLineItemResponse(item), nil
}
