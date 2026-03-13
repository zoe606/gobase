package article

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/repo"
)

// Delete deletes a article by ID.
func (uc *UseCase) Delete(ctx context.Context, id uint) error {
	if err := uc.articleRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("article - Delete - articleRepo.Delete: %w", err)
	}

	// Audit log (best-effort)
	_ = uc.auditLogger.LogDelete(ctx, "article", id, nil, nil)

	return nil
}
