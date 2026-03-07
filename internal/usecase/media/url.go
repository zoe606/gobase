package media

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"go-boilerplate/internal/dto/media"
	"go-boilerplate/internal/entity"
)

// GetURL returns a URL for accessing the media.
func (uc *UseCase) GetURL(ctx context.Context, media *entity.Media, variant string) (string, error) {
	path := media.Path
	if variant != "" {
		if variantPath := media.GetVariantPath(variant); variantPath != "" {
			path = variantPath
		}
	}

	return uc.storage.TemporaryURL(ctx, path, 1*time.Hour)
}

// GetPresignedUploadURL returns a URL for direct client upload.
func (uc *UseCase) GetPresignedUploadURL(ctx context.Context, filename string) (*mediadto.PresignedURLResponse, error) {
	ext := filepath.Ext(filename)
	uniqueName := fmt.Sprintf("temp/%s/%s%s", time.Now().Format("2006/01/02"), uuid.New().String(), ext)
	expiry := 15 * time.Minute

	url, err := uc.storage.PresignedUploadURL(ctx, uniqueName, expiry)
	if err != nil {
		return nil, fmt.Errorf("presigned url: %w", err)
	}

	return &mediadto.PresignedURLResponse{
		UploadURL: url,
		Path:      uniqueName,
		ExpiresIn: int(expiry.Seconds()),
	}, nil
}
