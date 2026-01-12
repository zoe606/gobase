package persistent

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"
)

// ProfileRepo implements profile repository using GORM.
type ProfileRepo struct {
	db *gorm.DB
}

// NewProfileRepo creates a new profile repository.
func NewProfileRepo(db *gorm.DB) *ProfileRepo {
	return &ProfileRepo{db: db}
}

// Create creates a new profile record.
func (r *ProfileRepo) Create(ctx context.Context, profile *entity.Profile) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(profile).Error; err != nil {
		return fmt.Errorf("ProfileRepo - Create: %w", err)
	}
	return nil
}

// GetByUserID retrieves a profile by user ID.
func (r *ProfileRepo) GetByUserID(ctx context.Context, userID uint) (*entity.Profile, error) {
	db := tx.DBFromContext(ctx, r.db)
	var profile entity.Profile
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("ProfileRepo - GetByUserID: %w", err)
	}
	return &profile, nil
}

// Update updates a profile record.
func (r *ProfileRepo) Update(ctx context.Context, profile *entity.Profile) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Save(profile).Error; err != nil {
		return fmt.Errorf("ProfileRepo - Update: %w", err)
	}
	return nil
}

// Upsert creates or updates a profile based on user_id.
func (r *ProfileRepo) Upsert(ctx context.Context, profile *entity.Profile) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"bio", "phone", "avatar_media_id", "updated_at"}),
	}).Create(profile).Error; err != nil {
		return fmt.Errorf("ProfileRepo - Upsert: %w", err)
	}
	return nil
}
