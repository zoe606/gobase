package persistent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/postgres"
	"gorm.io/gorm"
)

// RefreshTokenRepo implements refresh token repository using PostgreSQL.
type RefreshTokenRepo struct {
	*postgres.Postgres
}

// NewRefreshTokenRepo creates a new refresh token repository.
func NewRefreshTokenRepo(pg *postgres.Postgres) *RefreshTokenRepo {
	return &RefreshTokenRepo{pg}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepo) Create(ctx context.Context, token *entity.RefreshToken) error {
	result := r.DB.WithContext(ctx).Create(token)
	if result.Error != nil {
		return fmt.Errorf("RefreshTokenRepo - Create: %w", result.Error)
	}
	return nil
}

// GetByToken retrieves a refresh token by token string.
func (r *RefreshTokenRepo) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	var refreshToken entity.RefreshToken
	result := r.DB.WithContext(ctx).
		Where("token = ?", token).
		First(&refreshToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("RefreshTokenRepo - GetByToken: %w", result.Error)
	}
	return &refreshToken, nil
}

// DeleteByToken deletes a refresh token by token string.
func (r *RefreshTokenRepo) DeleteByToken(ctx context.Context, token string) error {
	result := r.DB.WithContext(ctx).
		Where("token = ?", token).
		Delete(&entity.RefreshToken{})
	if result.Error != nil {
		return fmt.Errorf("RefreshTokenRepo - DeleteByToken: %w", result.Error)
	}
	return nil
}

// DeleteByUserID deletes all refresh tokens for a user.
func (r *RefreshTokenRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	result := r.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&entity.RefreshToken{})
	if result.Error != nil {
		return fmt.Errorf("RefreshTokenRepo - DeleteByUserID: %w", result.Error)
	}
	return nil
}

// DeleteExpired deletes all expired refresh tokens.
func (r *RefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	result := r.DB.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.RefreshToken{})
	if result.Error != nil {
		return fmt.Errorf("RefreshTokenRepo - DeleteExpired: %w", result.Error)
	}
	return nil
}
