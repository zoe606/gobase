package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"
)

// EmailVerificationRepo implements email verification repository using GORM.
type EmailVerificationRepo struct {
	db *gorm.DB
}

// NewEmailVerificationRepo creates a new email verification repository.
func NewEmailVerificationRepo(db *gorm.DB) *EmailVerificationRepo {
	return &EmailVerificationRepo{db: db}
}

// Create creates a new email verification token.
func (r *EmailVerificationRepo) Create(ctx context.Context, verification *entity.EmailVerification) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(verification).Error; err != nil {
		return fmt.Errorf("EmailVerificationRepo - Create: %w", err)
	}
	return nil
}

// GetByToken retrieves an email verification by token.
func (r *EmailVerificationRepo) GetByToken(ctx context.Context, token string) (*entity.EmailVerification, error) {
	db := tx.DBFromContext(ctx, r.db)
	var verification entity.EmailVerification
	if err := db.Where("token = ?", token).First(&verification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("EmailVerificationRepo - GetByToken: %w", err)
	}
	return &verification, nil
}

// GetLatestByUserID retrieves the latest email verification for a user.
func (r *EmailVerificationRepo) GetLatestByUserID(ctx context.Context, userID uint) (*entity.EmailVerification, error) {
	db := tx.DBFromContext(ctx, r.db)
	var verification entity.EmailVerification
	if err := db.Where("user_id = ?", userID).Order("created_at DESC").First(&verification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("EmailVerificationRepo - GetLatestByUserID: %w", err)
	}
	return &verification, nil
}

// MarkAsUsed marks an email verification token as used.
func (r *EmailVerificationRepo) MarkAsUsed(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	now := time.Now()
	if err := db.Model(&entity.EmailVerification{}).Where("id = ?", id).Update("used_at", now).Error; err != nil {
		return fmt.Errorf("EmailVerificationRepo - MarkAsUsed: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all email verification tokens for a user.
func (r *EmailVerificationRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Where("user_id = ?", userID).Delete(&entity.EmailVerification{}).Error; err != nil {
		return fmt.Errorf("EmailVerificationRepo - DeleteByUserID: %w", err)
	}
	return nil
}
