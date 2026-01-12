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

// RefreshTokenRepo implements refresh token repository using GORM.
type RefreshTokenRepo struct {
	db *gorm.DB
}

// NewRefreshTokenRepo creates a new refresh token repository.
func NewRefreshTokenRepo(db *gorm.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepo) Create(ctx context.Context, token *entity.RefreshToken) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(token).Error; err != nil {
		return fmt.Errorf("RefreshTokenRepo - Create: %w", err)
	}
	return nil
}

// GetByToken retrieves a refresh token by token string.
func (r *RefreshTokenRepo) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	db := tx.DBFromContext(ctx, r.db)
	var refreshToken entity.RefreshToken
	if err := db.Where("token = ?", token).First(&refreshToken).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("RefreshTokenRepo - GetByToken: %w", err)
	}
	return &refreshToken, nil
}

// DeleteByToken deletes a refresh token by token string.
func (r *RefreshTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Where("token = ?", token).Delete(&entity.RefreshToken{}).Error; err != nil {
		return fmt.Errorf("RefreshTokenRepo - DeleteByToken: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all refresh tokens for a user.
func (r *RefreshTokenRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Where("user_id = ?", userID).Delete(&entity.RefreshToken{}).Error; err != nil {
		return fmt.Errorf("RefreshTokenRepo - DeleteByUserID: %w", err)
	}
	return nil
}
