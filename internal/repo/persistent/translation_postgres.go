package persistent

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/postgres"
)

// ErrTranslationNotFound is returned when a translation is not found.
var ErrTranslationNotFound = errors.New("translation not found")

// TranslationRepo implements translation repository with GORM + pgx.
type TranslationRepo struct {
	*postgres.Postgres
}

// New creates a new TranslationRepo.
func New(pg *postgres.Postgres) *TranslationRepo {
	return &TranslationRepo{pg}
}

// GetHistory retrieves all translation history.
func (r *TranslationRepo) GetHistory(ctx context.Context) ([]entity.Translation, error) {
	var translations []entity.Translation

	result := r.DB.WithContext(ctx).
		Order("created_at DESC").
		Find(&translations)

	if result.Error != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistory: %w", result.Error)
	}

	return translations, nil
}

// Store saves a new translation record.
func (r *TranslationRepo) Store(ctx context.Context, t entity.Translation) error {
	result := r.DB.WithContext(ctx).Create(&t)

	if result.Error != nil {
		return fmt.Errorf("TranslationRepo - Store: %w", result.Error)
	}

	return nil
}

// GetByID retrieves a translation by ID.
func (r *TranslationRepo) GetByID(ctx context.Context, id uint) (*entity.Translation, error) {
	var translation entity.Translation

	result := r.DB.WithContext(ctx).First(&translation, id)

	if result.Error != nil {
		return nil, fmt.Errorf("TranslationRepo - GetByID: %w", result.Error)
	}

	return &translation, nil
}

// Delete removes a translation by ID.
func (r *TranslationRepo) Delete(ctx context.Context, id uint) error {
	result := r.DB.WithContext(ctx).Delete(&entity.Translation{}, id)

	if result.Error != nil {
		return fmt.Errorf("TranslationRepo - Delete: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("TranslationRepo - Delete: %w", ErrTranslationNotFound)
	}

	return nil
}
