package article

import (
	"context"
	"errors"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/repo"
)

// Update updates a article.
func (uc *UseCase) Update(ctx context.Context, id uint, req articledto.UpdateRequest) (*articledto.Response, error) {
	// Get existing article
	article, err := uc.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("article - Update - articleRepo.GetByID: %w", err)
	}

	// TODO: Update fields from request (check for non-nil pointers)
	_ = article // Remove this line after implementing

	if err := uc.articleRepo.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("article - Update - articleRepo.Update: %w", err)
	}

	// Audit log (best-effort)
	_ = uc.auditLogger.LogUpdate(ctx, "article", article.ID, nil, nil, map[string]any{
		"title": article.Title,
	})

	return articledto.NewResponse(article), nil
}
