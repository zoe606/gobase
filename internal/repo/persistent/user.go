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
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("UserRepo - Create: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID with role and permissions.
func (r *UserRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	db := tx.DBFromContext(ctx, r.db)
	var user entity.User
	if err := db.Preload("Role.Permissions").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepo - GetByID: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email with role and permissions.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	db := tx.DBFromContext(ctx, r.db)
	var user entity.User
	if err := db.Preload("Role.Permissions").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, fmt.Errorf("UserRepo - GetByEmail: %w", err)
	}
	return &user, nil
}

// EmailExists checks if an email already exists.
func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	db := tx.DBFromContext(ctx, r.db)
	var count int64
	if err := db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("UserRepo - EmailExists: %w", err)
	}
	return count > 0, nil
}

// Update updates a user record.
func (r *UserRepo) Update(ctx context.Context, user *entity.User) error {
	db := tx.DBFromContext(ctx, r.db)
	if err := db.Save(user).Error; err != nil {
		return fmt.Errorf("UserRepo - Update: %w", err)
	}
	return nil
}
