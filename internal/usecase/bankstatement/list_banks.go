package bankstatement

import (
	"context"
	"fmt"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
)

func (uc *UseCase) ListBanks(ctx context.Context) (*bankstatementdto.BankListResponse, error) {
	banks, err := uc.bankRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - ListBanks - bankRepo.List: %w", err)
	}

	return bankstatementdto.NewBankListResponse(banks), nil
}
