package article

import (
	"context"
	"errors"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/repo"
)

// GetByID retrieves a article by ID.
func (uc *UseCase) GetByID(ctx context.Context, id uint) (*articledto.Response, error) {
	article, err := uc.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("article - GetByID - articleRepo.GetByID: %w", err)
	}

	return articledto.NewResponse(article), nil
}
