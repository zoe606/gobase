package installment

import (
	"context"
	"fmt"

	installmentdto "go-boilerplate/internal/dto/installment"
)

func (uc *UseCase) GetByID(ctx context.Context, id uint) (*installmentdto.Response, error) {
	inst, err := uc.installmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("installment - GetByID - installmentRepo.GetByID: %w", err)
	}

	return installmentdto.NewResponse(inst), nil
}
