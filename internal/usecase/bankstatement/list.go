package bankstatement

import (
	"context"
	"fmt"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
)

func (uc *UseCase) List(ctx context.Context, req bankstatementdto.ListRequest) (*bankstatementdto.ListResponse, error) {
	req.Normalize()

	stmts, total, err := uc.stmtRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("bankstatement - List - stmtRepo.List: %w", err)
	}

	return bankstatementdto.NewListResponse(stmts, total, req.Params), nil
}
