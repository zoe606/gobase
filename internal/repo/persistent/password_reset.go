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

// PasswordResetRepo implements password reset repository using GORM.
type PasswordResetRepo struct {
	db *gorm.DB
}

// NewPasswordResetRepo creates a new password reset repository.
func NewPasswordResetRepo(db *gorm.DB) *PasswordResetRepo {
	return &PasswordResetRepo{db: db}
}

// Create creates a new password reset token.
func (r *PasswordResetRepo) Create(ctx context.Context, reset *entity.PasswordReset) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(reset).Error; err != nil {
		return fmt.Errorf("PasswordResetRepo - Create: %w", err)
	}
	return nil
}

// GetByToken retrieves a password reset by token.
func (r *PasswordResetRepo) GetByToken(ctx context.Context, token string) (*entity.PasswordReset, error) {
	db := tx.DBFromContext(ctx, r.db)
	var reset entity.PasswordReset
	if err := db.Where("token = ?", token).First(&reset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("PasswordResetRepo - GetByToken: %w", err)
	}
	return &reset, nil
}

// MarkAsUsed marks a password reset token as used.
func (r *PasswordResetRepo) MarkAsUsed(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	now := time.Now()
	if err := db.Model(&entity.PasswordReset{}).Where("id = ?", id).Update("used_at", now).Error; err != nil {
		return fmt.Errorf("PasswordResetRepo - MarkAsUsed: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all password reset tokens for a user.
func (r *PasswordResetRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Where("user_id = ?", userID).Delete(&entity.PasswordReset{}).Error; err != nil {
		return fmt.Errorf("PasswordResetRepo - DeleteByUserID: %w", err)
	}
	return nil
}
