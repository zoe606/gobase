package installment

import (
	"context"
	"fmt"

	installmentdto "go-boilerplate/internal/dto/installment"
)

func (uc *UseCase) List(ctx context.Context, req installmentdto.ListRequest) (*installmentdto.ListResponse, error) {
	req.Normalize()

	installments, total, err := uc.installmentRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("installment - List - installmentRepo.List: %w", err)
	}

	return installmentdto.NewListResponse(installments, total, req.Params), nil
}
