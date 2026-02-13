// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	articledto "go-boilerplate/internal/dto/article"
	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	installmentdto "go-boilerplate/internal/dto/installment"
	translationdto "go-boilerplate/internal/dto/translation"
	userdto "go-boilerplate/internal/dto/user"
	"go-boilerplate/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo defines the translation repository interface.
	TranslationRepo interface {
		Store(context.Context, *entity.Translation) error
		GetHistory(ctx context.Context, req translationdto.HistoryRequest) ([]entity.Translation, int64, error)
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
		Delete(ctx context.Context, id uint) error
		List(ctx context.Context, req userdto.ListRequest) ([]*entity.User, int64, error)
	}

	// RoleRepo defines the role repository interface.
	RoleRepo interface {
		Create(ctx context.Context, role *entity.Role) error
		GetByID(ctx context.Context, id uint) (*entity.Role, error)
		GetByName(ctx context.Context, name string) (*entity.Role, error)
		List(ctx context.Context) ([]*entity.Role, error)
		Update(ctx context.Context, role *entity.Role) error
		Delete(ctx context.Context, id uint) error
		UpdatePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error
	}

	// PermissionRepo defines the permission repository interface.
	PermissionRepo interface {
		List(ctx context.Context) ([]*entity.Permission, error)
		GetByIDs(ctx context.Context, ids []uint) ([]*entity.Permission, error)
		Create(ctx context.Context, permission *entity.Permission) error
		Delete(ctx context.Context, id uint) error
		GetByName(ctx context.Context, name string) (*entity.Permission, error)
		IsAssignedToAnyRole(ctx context.Context, permissionID uint) (bool, error)
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

	// ProfileRepo defines profile repository operations.
	ProfileRepo interface {
		Create(ctx context.Context, profile *entity.Profile) error
		GetByUserID(ctx context.Context, userID uint) (*entity.Profile, error)
		Update(ctx context.Context, profile *entity.Profile) error
		Upsert(ctx context.Context, profile *entity.Profile) error
	}

	// ArticleRepo defines Article repository operations.
	ArticleRepo interface {
		Create(ctx context.Context, article *entity.Article) error
		GetByID(ctx context.Context, id uint) (*entity.Article, error)
		List(ctx context.Context, req articledto.ListRequest) ([]*entity.Article, int64, error)
		Update(ctx context.Context, article *entity.Article) error
		Delete(ctx context.Context, id uint) error
	}

	// BankRepo defines bank repository operations.
	BankRepo interface {
		List(ctx context.Context) ([]*entity.Bank, error)
		GetByID(ctx context.Context, id uint) (*entity.Bank, error)
		GetByCode(ctx context.Context, code string) (*entity.Bank, error)
	}

	// BankStatementRepo defines bank statement repository operations.
	BankStatementRepo interface {
		Create(ctx context.Context, stmt *entity.BankStatement) error
		GetByID(ctx context.Context, id uint) (*entity.BankStatement, error)
		List(ctx context.Context, req bankstatementdto.ListRequest) ([]*entity.BankStatement, int64, error)
		Update(ctx context.Context, stmt *entity.BankStatement) error
		Delete(ctx context.Context, id uint) error
	}

	// LineItemRepo defines line item repository operations.
	LineItemRepo interface {
		BulkCreate(ctx context.Context, items []*entity.LineItem) error
		GetByID(ctx context.Context, id uint) (*entity.LineItem, error)
		GetBySource(ctx context.Context, sourceType string, sourceID uint) ([]*entity.LineItem, error)
		Update(ctx context.Context, item *entity.LineItem) error
		DeleteBySource(ctx context.Context, sourceType string, sourceID uint) error
		UpdateInstallmentID(ctx context.Context, itemIDs []uint, installmentID *uint) error
	}

	// InstallmentRepo defines installment repository operations.
	InstallmentRepo interface {
		Create(ctx context.Context, inst *entity.Installment) error
		GetByID(ctx context.Context, id uint) (*entity.Installment, error)
		List(ctx context.Context, req installmentdto.ListRequest) ([]*entity.Installment, int64, error)
		Update(ctx context.Context, inst *entity.Installment) error
		Delete(ctx context.Context, id uint) error
	}
)
