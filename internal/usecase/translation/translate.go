package translation

import (
	"context"
	"fmt"

	"go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
)

// Translate performs a translation.
func (uc *UseCase) Translate(ctx context.Context, input translationdto.TranslateRequest) (*translationdto.TranslationResponse, error) {
	t := &entity.Translation{
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

	return translationdto.NewTranslationResponse(translation), nil
}
