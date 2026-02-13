package installment

import (
	"context"
	"fmt"

	installmentdto "go-boilerplate/internal/dto/installment"
)

func (uc *UseCase) LinkItems(ctx context.Context, installmentID uint, req installmentdto.LinkItemsRequest) error {
	_, err := uc.installmentRepo.GetByID(ctx, installmentID)
	if err != nil {
		return fmt.Errorf("installment - LinkItems - installmentRepo.GetByID: %w", err)
	}

	if err := uc.lineItemRepo.UpdateInstallmentID(ctx, req.LineItemIDs, &installmentID); err != nil {
		return fmt.Errorf("installment - LinkItems - lineItemRepo.UpdateInstallmentID: %w", err)
	}

	return nil
}
