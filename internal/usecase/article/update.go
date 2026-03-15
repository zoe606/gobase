package article

import (
	"context"
	"errors"
	"fmt"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/repo"
)

// Update updates a article.
func (uc *UseCase) Update(ctx context.Context, userID, id uint, req articledto.UpdateRequest) (*articledto.Response, error) {
	// Get existing article
	article, err := uc.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("article - Update - articleRepo.GetByID: %w", err)
	}

	// Ownership check
	if article.UserID != userID {
		return nil, ErrForbidden
	}

	// Capture old values for audit
	oldValues := map[string]any{"title": article.Title, "slug": article.Slug}

	// Apply non-nil fields from request
	if req.Title != nil {
		article.Title = *req.Title
	}
	if req.Slug != nil {
		article.Slug = *req.Slug
	}
	if req.Content != nil {
		article.Content = req.Content
	}
	if req.Excerpt != nil {
		article.Excerpt = req.Excerpt
	}
	if req.CoverMediaID != nil {
		article.CoverMediaID = req.CoverMediaID
	}
	if req.Status != nil {
		article.Status = req.Status
	}
	if req.PublishedAt != nil {
		article.PublishedAt = req.PublishedAt
	}

	if err := uc.articleRepo.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("article - Update - articleRepo.Update: %w", err)
	}

	// Audit log (best-effort)
	_ = uc.auditLogger.LogUpdate(ctx, "article", article.ID, &userID, oldValues, map[string]any{
		"title": article.Title,
	})

	// Invalidate caches
	_ = uc.cache.Delete(ctx, uc.cacheKeys.ID(id))
	_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())

	return articledto.NewResponse(article), nil
}
