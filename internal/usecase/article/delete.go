package article

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/repo"
)

// Delete deletes a article by ID after verifying ownership.
func (uc *UseCase) Delete(ctx context.Context, userID uint, id uint) error {
	// Fetch article to verify ownership
	article, err := uc.articleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("article - Delete - articleRepo.GetByID: %w", err)
	}

	if article.UserID != userID {
		return ErrForbidden
	}

	if err := uc.articleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("article - Delete - articleRepo.Delete: %w", err)
	}

	// Audit log (best-effort)
	_ = uc.auditLogger.LogDelete(ctx, "article", id, &userID, nil)

	// Invalidate caches
	_ = uc.cache.Delete(ctx, uc.cacheKeys.ID(id))
	_ = uc.cache.DeleteByPrefix(ctx, uc.cacheKeys.ListPrefix())

	return nil
}
