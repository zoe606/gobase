package persistent

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"
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
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(media).Error; err != nil {
		return fmt.Errorf("MediaRepo - Create: %w", err)
	}
	return nil
}

// GetByID retrieves a media record by ID.
func (r *MediaRepo) GetByID(ctx context.Context, id uint) (*entity.Media, error) {
	db := tx.DBFromContext(ctx, r.db)
	var media entity.Media
	if err := db.First(&media, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("MediaRepo - GetByID: %w", err)
	}
	return &media, nil
}

// GetByAttachable retrieves media by attachable type, ID, and optional collection.
func (r *MediaRepo) GetByAttachable(ctx context.Context, attachableType string, attachableID uint, collection string) ([]*entity.Media, error) {
	db := tx.DBFromContext(ctx, r.db)
	var media []*entity.Media
	query := db.Where("attachable_type = ? AND attachable_id = ?", attachableType, attachableID)

	if collection != "" {
		query = query.Where("collection = ?", collection)
	}

	if err := query.Order("created_at DESC").Find(&media).Error; err != nil {
		return nil, fmt.Errorf("MediaRepo - GetByAttachable: %w", err)
	}
	return media, nil
}

// Update updates a media record.
func (r *MediaRepo) Update(ctx context.Context, media *entity.Media) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Save(media).Error; err != nil {
		return fmt.Errorf("MediaRepo - Update: %w", err)
	}
	return nil
}

// Delete soft-deletes a media record by ID.
func (r *MediaRepo) Delete(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Delete(&entity.Media{}, id)
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
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Where("attachable_type = ? AND attachable_id = ?", attachableType, attachableID).Delete(&entity.Media{}).Error; err != nil {
		return fmt.Errorf("MediaRepo - DeleteByAttachable: %w", err)
	}
	return nil
}
