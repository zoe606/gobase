package persistent

import (
	"context"
	"fmt"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"

	"gorm.io/gorm"
)

// TranslationRepo implements translation repository with GORM.
type TranslationRepo struct {
	db *gorm.DB
}

// New creates a new TranslationRepo.
func New(db *gorm.DB) *TranslationRepo {
	return &TranslationRepo{db: db}
}

// GetHistory retrieves all translation history.
func (r *TranslationRepo) GetHistory(ctx context.Context) ([]entity.Translation, error) {
	var translations []entity.Translation

	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&translations)

	if result.Error != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistory: %w", result.Error)
	}

	return translations, nil
}

// Store saves a new translation record.
func (r *TranslationRepo) Store(ctx context.Context, t *entity.Translation) error {
	result := r.db.WithContext(ctx).Create(t)

	if result.Error != nil {
		return fmt.Errorf("TranslationRepo - Store: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a translation by ID.
func (r *TranslationRepo) GetByID(ctx context.Context, id uint) (*entity.Translation, error) {
	var translation entity.Translation

	result := r.db.WithContext(ctx).First(&translation, id)

	if result.Error != nil {
		return nil, fmt.Errorf("TranslationRepo - GetByID: %w", result.Error)
	}

	return &translation, nil
}

// Delete removes a translation by ID.
func (r *TranslationRepo) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Translation{}, id)

	if result.Error != nil {
		return fmt.Errorf("TranslationRepo - Delete: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("TranslationRepo - Delete: %w", repo.ErrNotFound)
	}

	return nil
}
