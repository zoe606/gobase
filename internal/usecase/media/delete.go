package media

import (
	"context"
	"fmt"
)

// Delete removes a media record and its file.
func (uc *UseCase) Delete(ctx context.Context, id uint) error {
	media, err := uc.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from storage.
	if err := uc.storage.Delete(ctx, media.Path); err != nil {
		uc.l.Warn("Failed to delete media file from storage",
			"id", id,
			"path", media.Path,
			"error", err,
		)
	}

	// Delete variants.
	for _, variantPath := range media.Variants {
		if path, ok := variantPath.(string); ok {
			if err := uc.storage.Delete(ctx, path); err != nil {
				uc.l.Warn("Failed to delete variant file",
					"id", id,
					"path", path,
					"error", err,
				)
			}
		}
	}

	// Delete from database.
	if err := uc.mediaRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}

	uc.l.Info("Media deleted", "id", id)
	return nil
}
