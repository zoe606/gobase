package article

import (
	"context"
	"fmt"

	"go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
)

// Create creates a new article.
func (uc *UseCase) Create(ctx context.Context, req articledto.CreateRequest) (*articledto.Response, error) {
	// TODO: Add validation logic

	article := &entity.Article{
		// TODO: Map request fields to entity
	}

	if err := uc.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("article - Create - articleRepo.Create: %w", err)
	}

	return articledto.NewResponse(article), nil
}
