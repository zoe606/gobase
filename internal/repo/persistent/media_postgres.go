package persistent

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"

	"gorm.io/gorm"
)

// MediaRepo implements media repository using GORM.
type MediaRepo struct {
	db *gorm.DB
}

// NewMediaRepo creates a new media repository.
func NewMediaRepo(db *gorm.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

// Create creates a new media record.
func (r *MediaRepo) Create(ctx context.Context, media *entity.Media) error {
	result := r.db.WithContext(ctx).Create(media)
	if result.Error != nil {
		return fmt.Errorf("MediaRepo - Create: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a media record by ID.
func (r *MediaRepo) GetByID(ctx context.Context, id uint) (*entity.Media, error) {
	var media entity.Media
	result := r.db.WithContext(ctx).First(&media, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("MediaRepo - GetByID: %w", result.Error)
	}
	return &media, nil
}

// GetByAttachable retrieves media by attachable type, ID, and optional collection.
func (r *MediaRepo) GetByAttachable(ctx context.Context, attachableType string, attachableID uint, collection string) ([]*entity.Media, error) {
	var media []*entity.Media
	query := r.db.WithContext(ctx).
		Where("attachable_type = ? AND attachable_id = ?", attachableType, attachableID)

	if collection != "" {
		query = query.Where("collection = ?", collection)
	}

	result := query.Order("created_at DESC").Find(&media)
	if result.Error != nil {
		return nil, fmt.Errorf("MediaRepo - GetByAttachable: %w", result.Error)
	}
	return media, nil
}

// Update updates a media record.
func (r *MediaRepo) Update(ctx context.Context, media *entity.Media) error {
	result := r.db.WithContext(ctx).Save(media)
	if result.Error != nil {
		return fmt.Errorf("MediaRepo - Update: %w", result.Error)
	}
	return nil
}

// Delete soft-deletes a media record by ID.
func (r *MediaRepo) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Media{}, id)
	if result.Error != nil {
		return fmt.Errorf("MediaRepo - Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// DeleteByAttachable soft-deletes all media for an attachable entity.
func (r *MediaRepo) DeleteByAttachable(ctx context.Context, attachableType string, attachableID uint) error {
	result := r.db.WithContext(ctx).
		Where("attachable_type = ? AND attachable_id = ?", attachableType, attachableID).
		Delete(&entity.Media{})
	if result.Error != nil {
		return fmt.Errorf("MediaRepo - DeleteByAttachable: %w", result.Error)
	}
	return nil
}
