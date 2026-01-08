package persistent

import (
	"context"
	"errors"
	"fmt"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/postgres"
	"gorm.io/gorm"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.New("user not found")

// UserRepo implements user repository using PostgreSQL.
type UserRepo struct {
	*postgres.Postgres
}

// NewUserRepo creates a new user repository.
func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

// Create creates a new user.
func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	result := r.DB.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("UserRepo - Create: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a user by ID with role and permissions.
func (r *UserRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	result := r.DB.WithContext(ctx).
		Preload("Role.Permissions").
		First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepo - GetByID: %w", result.Error)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email with role and permissions.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	result := r.DB.WithContext(ctx).
		Preload("Role.Permissions").
		Where("email = ?", email).
		First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepo - GetByEmail: %w", result.Error)
	}
	return &user, nil
}

// Update updates a user.
func (r *UserRepo) Update(ctx context.Context, user *entity.User) error {
	result := r.DB.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("UserRepo - Update: %w", result.Error)
	}
	return nil
}

// Delete soft-deletes a user.
func (r *UserRepo) Delete(ctx context.Context, id uint) error {
	result := r.DB.WithContext(ctx).Delete(&entity.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("UserRepo - Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("UserRepo - Delete: %w", ErrUserNotFound)
	}
	return nil
}

// EmailExists checks if an email already exists.
func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.DB.WithContext(ctx).
		Model(&entity.User{}).
		Where("email = ?", email).
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("UserRepo - EmailExists: %w", result.Error)
	}
	return count > 0, nil
}
