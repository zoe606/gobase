package persistent

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/pagination"
	"go-boilerplate/pkg/tx"
)

// TranslationRepo implements translation repository with GORM.
type TranslationRepo struct {
	db *gorm.DB
}

// New creates a new TranslationRepo.
func New(db *gorm.DB) *TranslationRepo {
	return &TranslationRepo{db: db}
}

// GetHistory retrieves translation history with pagination.
func (r *TranslationRepo) GetHistory(ctx context.Context, params pagination.Params) ([]entity.Translation, int64, error) {
	db := tx.DBFromContext(ctx, r.db)

	// Count total records
	var total int64
	if err := db.Model(&entity.Translation{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("TranslationRepo - GetHistory - Count: %w", err)
	}

	// Get paginated results
	var translations []entity.Translation
	query := params.Apply(db.Model(&entity.Translation{}), []string{"created_at", "id"})
	if err := query.Find(&translations).Error; err != nil {
		return nil, 0, fmt.Errorf("TranslationRepo - GetHistory: %w", err)
	}

	return translations, total, nil
}

// Store saves a new translation record.
func (r *TranslationRepo) Store(ctx context.Context, t *entity.Translation) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(t).Error; err != nil {
		return fmt.Errorf("TranslationRepo - Store: %w", err)
	}
	return nil
}

// GetByID retrieves a translation by ID.
func (r *TranslationRepo) GetByID(ctx context.Context, id uint) (*entity.Translation, error) {
	db := tx.DBFromContext(ctx, r.db)
	var translation entity.Translation
	if err := db.First(&translation, id).Error; err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetByID: %w", err)
	}
	return &translation, nil
}

// Delete removes a translation by ID.
func (r *TranslationRepo) Delete(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Delete(&entity.Translation{}, id)
	if result.Error != nil {
		return fmt.Errorf("TranslationRepo - Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("TranslationRepo - Delete: %w", repo.ErrNotFound)
	}
	return nil
}
