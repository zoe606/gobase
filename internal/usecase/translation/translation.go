package translation

import (
	"context"
	"fmt"

	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
)

// UseCase handles translation business logic.
type UseCase struct {
	repo   repo.TranslationRepo
	webAPI repo.TranslationWebAPI
}

// New creates a new translation use case.
func New(r repo.TranslationRepo, w repo.TranslationWebAPI) *UseCase {
	return &UseCase{
		repo:   r,
		webAPI: w,
	}
}

// History retrieves translation history from store.
func (uc *UseCase) History(ctx context.Context) (*translationdto.HistoryOutput, error) {
	translations, err := uc.repo.GetHistory(ctx)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return &translationdto.HistoryOutput{History: translations}, nil
}

// Translate performs a translation.
func (uc *UseCase) Translate(ctx context.Context, input translationdto.TranslateInput) (*translationdto.TranslateOutput, error) {
	t := entity.Translation{
		Source:      input.Source,
		Destination: input.Destination,
		Original:    input.Original,
	}

	translation, err := uc.webAPI.Translate(t)
	if err != nil {
		return nil, fmt.Errorf("TranslationUseCase - Translate - s.webAPI.Translate: %w", err)
	}

	if err = uc.repo.Store(ctx, translation); err != nil {
		return nil, fmt.Errorf("TranslationUseCase - Translate - s.repo.Store: %w", err)
	}

	return &translationdto.TranslateOutput{Translation: translation}, nil
}
