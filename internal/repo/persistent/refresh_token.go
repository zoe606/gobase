package persistent

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"

	"gorm.io/gorm"
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
	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		return fmt.Errorf("RefreshTokenRepo - Create: %w", result.Error)
	}
	return nil
}

// GetByToken retrieves a refresh token by token string.
func (r *RefreshTokenRepo) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	var refreshToken entity.RefreshToken
	result := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("RefreshTokenRepo - GetByToken: %w", result.Error)
	}
	return &refreshToken, nil
}

// DeleteByToken deletes a refresh token by token string.
func (r *RefreshTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&entity.RefreshToken{})
	if result.Error != nil {
		return fmt.Errorf("RefreshTokenRepo - DeleteByToken: %w", result.Error)
	}
	return nil
}

// DeleteByUserID deletes all refresh tokens for a user.
func (r *RefreshTokenRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.RefreshToken{})
	if result.Error != nil {
		return fmt.Errorf("RefreshTokenRepo - DeleteByUserID: %w", result.Error)
	}
	return nil
}
