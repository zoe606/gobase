package media

import (
	"context"

	mediadto "go-boilerplate/internal/dto/media"
	"go-boilerplate/internal/entity"
)

// GetByID retrieves a media by ID.
func (uc *UseCase) GetByID(ctx context.Context, id uint) (*entity.Media, error) {
	media, err := uc.mediaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// GetByAttachable retrieves media for an attachable entity.
func (uc *UseCase) GetByAttachable(ctx context.Context, req mediadto.GetMediaRequest) (*mediadto.MediaListResponse, error) {
	media, err := uc.mediaRepo.GetByAttachable(ctx, req.AttachableType, req.AttachableID, req.Collection)
	if err != nil {
		return nil, err
	}
	return mediadto.FromEntities(media), nil
}
