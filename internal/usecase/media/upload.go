package media

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"

	mediadto "go-boilerplate/internal/dto/media"
	"go-boilerplate/internal/entity"
)

// Upload handles file upload.
func (uc *UseCase) Upload(ctx context.Context, req mediadto.UploadRequest) (*mediadto.MediaResponse, error) {
	// Validate file size.
	if req.Size > uc.maxSize {
		return nil, ErrFileTooLarge
	}

	// Validate MIME type.
	if !entity.IsAllowedMimeType(req.MimeType, entity.DefaultAllowedMimeTypes()) {
		return nil, ErrInvalidMimeType
	}

	// Generate unique filename.
	ext := filepath.Ext(req.Filename)
	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	storagePath := buildStoragePath(req.AttachableType, req.Collection, req.AttachableID, uniqueName)

	// Store file.
	fileInfo, err := uc.storage.Put(ctx, storagePath, req.File, req.Size, req.MimeType)
	if err != nil {
		return nil, fmt.Errorf("store file: %w", err)
	}

	// Determine media type.
	mediaType := detectMediaType(req.MimeType)

	// Create media record.
	media := &entity.Media{
		AttachableType: req.AttachableType,
		AttachableID:   req.AttachableID,
		Collection:     req.Collection,
		Filename:       uniqueName,
		OriginalName:   req.Filename,
		MimeType:       req.MimeType,
		Size:           fileInfo.Size,
		Disk:           uc.disk,
		Path:           fileInfo.Path,
		Type:           mediaType,
		Hash:           fileInfo.Hash,
	}

	if err := uc.mediaRepo.Create(ctx, media); err != nil {
		// Cleanup uploaded file on DB error.
		_ = uc.storage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("create media: %w", err)
	}

	// Queue image processing if it's an image.
	if mediaType == entity.MediaTypeImage {
		uc.queueImageProcessing(media.ID)
	}

	uc.l.Info("Media uploaded",
		"id", media.ID,
		"type", mediaType,
		"size", media.Size,
		"attachable", fmt.Sprintf("%s:%d", req.AttachableType, req.AttachableID),
	)

	return mediadto.FromEntity(media), nil
}
