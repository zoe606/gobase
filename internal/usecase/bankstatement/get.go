package bankstatement

import (
	"context"
	"fmt"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
)

func (uc *UseCase) GetByID(ctx context.Context, id uint) (*bankstatementdto.ResponseWithItems, error) {
	stmt, err := uc.stmtRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - GetByID - stmtRepo.GetByID: %w", err)
	}

	items, err := uc.lineItemRepo.GetBySource(ctx, "bank_statement", id)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - GetByID - lineItemRepo.GetBySource: %w", err)
	}

	return bankstatementdto.NewResponseWithItems(stmt, items), nil
}
