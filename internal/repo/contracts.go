// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/pagination"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo defines the translation repository interface.
	TranslationRepo interface {
		Store(context.Context, *entity.Translation) error
		GetHistory(ctx context.Context, params pagination.Params) ([]entity.Translation, int64, error)
	}

	// TranslationWebAPI defines the translation web API interface.
	TranslationWebAPI interface {
		Translate(*entity.Translation) (*entity.Translation, error)
	}

	// UserRepo defines the user repository interface.
	UserRepo interface {
		Create(ctx context.Context, user *entity.User) error
		GetByID(ctx context.Context, id uint) (*entity.User, error)
		GetByEmail(ctx context.Context, email string) (*entity.User, error)
		EmailExists(ctx context.Context, email string) (bool, error)
		Update(ctx context.Context, user *entity.User) error
	}

	// RoleRepo defines the role repository interface.
	RoleRepo interface {
		GetByName(ctx context.Context, name string) (*entity.Role, error)
	}

	// RefreshTokenRepo defines the refresh token repository interface.
	RefreshTokenRepo interface {
		Create(ctx context.Context, token *entity.RefreshToken) error
		GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
		DeleteByToken(ctx context.Context, token string) error
		DeleteByUserID(ctx context.Context, userID uint) error
	}

	// MediaRepo defines media storage operations.
	MediaRepo interface {
		Create(ctx context.Context, media *entity.Media) error
		GetByID(ctx context.Context, id uint) (*entity.Media, error)
		GetByAttachable(ctx context.Context, attachableType string, attachableID uint, collection string) ([]*entity.Media, error)
		Update(ctx context.Context, media *entity.Media) error
		Delete(ctx context.Context, id uint) error
		DeleteByAttachable(ctx context.Context, attachableType string, attachableID uint) error
	}

	// EmailVerificationRepo defines email verification token operations.
	EmailVerificationRepo interface {
		Create(ctx context.Context, verification *entity.EmailVerification) error
		GetByToken(ctx context.Context, token string) (*entity.EmailVerification, error)
		GetLatestByUserID(ctx context.Context, userID uint) (*entity.EmailVerification, error)
		MarkAsUsed(ctx context.Context, id uint) error
		DeleteByUserID(ctx context.Context, userID uint) error
	}

	// PasswordResetRepo defines password reset token operations.
	PasswordResetRepo interface {
		Create(ctx context.Context, reset *entity.PasswordReset) error
		GetByToken(ctx context.Context, token string) (*entity.PasswordReset, error)
		MarkAsUsed(ctx context.Context, id uint) error
		DeleteByUserID(ctx context.Context, userID uint) error
	}
)
