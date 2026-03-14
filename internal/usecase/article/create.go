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

	// Audit log (best-effort — don't fail the operation)
	_ = uc.auditLogger.LogCreate(ctx, "article", article.ID, &userID, map[string]any{
		"title": req.Title,
		"slug":  req.Slug,
	})

	// Invalidate list cache (new article changes any list)
	_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())

	return articledto.NewResponse(article), nil
}
