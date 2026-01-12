// Package profile implements the profile use case.
package profile

import (
	"context"
	"time"

	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/storage"
	"go-boilerplate/pkg/logger"
)

// UseCase implements the profile use case.
type UseCase struct {
	profileRepo repo.ProfileRepo
	mediaRepo   repo.MediaRepo
	storage     storage.Provider
	l           logger.Interface
}

// New creates a new profile use case.
func New(
	profileRepo repo.ProfileRepo,
	mediaRepo repo.MediaRepo,
	storageProvider storage.Provider,
	l logger.Interface,
) *UseCase {
	return &UseCase{
		profileRepo: profileRepo,
		mediaRepo:   mediaRepo,
		storage:     storageProvider,
		l:           l,
	}
}

// getAvatarURL generates a URL for the avatar media.
func (uc *UseCase) getAvatarURL(ctx context.Context, mediaID uint) (string, error) {
	media, err := uc.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return "", err
	}

	url, err := uc.storage.TemporaryURL(ctx, media.Path, 1*time.Hour)
	if err != nil {
		return "", err
	}

	return url, nil
}
