package translation

import (
	"context"
	"fmt"

	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/repo"
)

// History retrieves translation history from store with pagination.
func (uc *UseCase) History(ctx context.Context, req translationdto.HistoryRequest) (*translationdto.HistoryResponse, error) {
	req.Normalize()

	params := repo.TranslationHistoryParams{
		Params: req.Params,
		Search: req.Search,
	}

	translations, total, err := uc.repo.GetHistory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return translationdto.NewHistoryResponse(translations, req.Params, total), nil
}
