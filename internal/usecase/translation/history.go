package translation

import (
	"context"
	"fmt"

	"go-boilerplate/internal/dto/translation"
)

// History retrieves translation history from store with pagination.
func (uc *UseCase) History(ctx context.Context, req translationdto.HistoryRequest) (*translationdto.HistoryResponse, error) {
	req.Normalize()

	translations, total, err := uc.repo.GetHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return translationdto.NewHistoryResponse(translations, req.Params, total), nil
}
