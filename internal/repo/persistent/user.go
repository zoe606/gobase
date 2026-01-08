package persistent

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"

	"gorm.io/gorm"
)

// UserRepo implements user repository using GORM.
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo creates a new user repository.
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create creates a new user.
func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("UserRepo - Create: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a user by ID with role and permissions.
func (r *UserRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepo - GetByID: %w", result.Error)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email with role and permissions.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("email = ?", email).
		First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepo - GetByEmail: %w", result.Error)
	}
	return &user, nil
}

// EmailExists checks if an email already exists.
func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("email = ?", email).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("UserRepo - EmailExists: %w", result.Error)
	}
	return count > 0, nil
}
