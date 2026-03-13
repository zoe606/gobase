package article

import (
	"context"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
)

// Create creates a new article.
func (uc *UseCase) Create(ctx context.Context, userID uint, req articledto.CreateRequest) (*articledto.Response, error) {
	// TODO: Add validation logic

	article := &entity.Article{
		UserID: userID,
		// TODO: Map remaining request fields to entity
	}

	if err := uc.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("article - Create - articleRepo.Create: %w", err)
	}

	return articledto.NewResponse(article), nil
}
