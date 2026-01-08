package translation

import (
	"context"
	"fmt"

	translationdto "go-boilerplate/internal/dto/translation"
)

// History retrieves translation history from store.
func (uc *UseCase) History(ctx context.Context) (*translationdto.HistoryResponse, error) {
	translations, err := uc.repo.GetHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return translationdto.NewHistoryResponse(translations), nil
}
