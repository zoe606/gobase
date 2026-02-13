// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	articledto "go-boilerplate/internal/dto/article"
	authdto "go-boilerplate/internal/dto/auth"
	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	installmentdto "go-boilerplate/internal/dto/installment"
	mediadto "go-boilerplate/internal/dto/media"
	permissiondto "go-boilerplate/internal/dto/permission"
	profiledto "go-boilerplate/internal/dto/profile"
	roledto "go-boilerplate/internal/dto/role"
	translationdto "go-boilerplate/internal/dto/translation"
	userdto "go-boilerplate/internal/dto/user"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/auth"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation defines the translation use case interface.
	Translation interface {
		Translate(context.Context, translationdto.TranslateRequest) (*translationdto.TranslationResponse, error)
		History(context.Context, translationdto.HistoryRequest) (*translationdto.HistoryResponse, error)
	}

	// Auth defines the authentication use case interface.
	Auth interface {
		Register(ctx context.Context, input authdto.RegisterRequest) (*authdto.LoginResponse, error)
		Login(ctx context.Context, input authdto.LoginRequest) (*authdto.LoginResponse, error)
		Refresh(ctx context.Context, input authdto.RefreshRequest) (*authdto.TokenResponse, error)
		Logout(ctx context.Context, refreshToken string) error
		GetCurrentUser(ctx context.Context, userID uint) (*authdto.UserResponse, error)

		// Email verification
		SendVerificationEmail(ctx context.Context, userID uint) error
		VerifyEmail(ctx context.Context, token string) error
		ResendVerification(ctx context.Context, email string) error

		// Password reset
		RequestPasswordReset(ctx context.Context, email string) error
		ResetPassword(ctx context.Context, input auth.ResetPasswordInput) error
	}

	// Media defines the media use case interface.
	Media interface {
		Upload(ctx context.Context, req mediadto.UploadRequest) (*mediadto.MediaResponse, error)
		GetByID(ctx context.Context, id uint) (*entity.Media, error)
		GetByAttachable(ctx context.Context, req mediadto.GetMediaRequest) (*mediadto.MediaListResponse, error)
		GetURL(ctx context.Context, media *entity.Media, variant string) (string, error)
		GetPresignedUploadURL(ctx context.Context, filename string) (*mediadto.PresignedURLResponse, error)
		Delete(ctx context.Context, id uint) error
	}

	// Profile defines the profile use case interface.
	Profile interface {
		GetProfile(ctx context.Context, userID uint) (*profiledto.ProfileResponse, error)
		UpdateProfile(ctx context.Context, userID uint, req profiledto.UpdateProfileRequest) (*profiledto.ProfileResponse, error)
	}

	// Article defines Article use case operations.
	Article interface {
		Create(ctx context.Context, req articledto.CreateRequest) (*articledto.Response, error)
		GetByID(ctx context.Context, id uint) (*articledto.Response, error)
		List(ctx context.Context, req articledto.ListRequest) (*articledto.ListResponse, error)
		Update(ctx context.Context, id uint, req articledto.UpdateRequest) (*articledto.Response, error)
		Delete(ctx context.Context, id uint) error
	}

	// User defines user management use case operations.
	User interface {
		Create(ctx context.Context, req userdto.CreateRequest) (*userdto.Response, error)
		GetByID(ctx context.Context, id uint) (*userdto.Response, error)
		List(ctx context.Context, req userdto.ListRequest) (*userdto.ListResponse, error)
		Update(ctx context.Context, id uint, req userdto.UpdateRequest) (*userdto.Response, error)
		Delete(ctx context.Context, id uint, currentUserID uint) error
	}

	// Role defines role management use case operations.
	Role interface {
		Create(ctx context.Context, req roledto.CreateRequest) (*roledto.Response, error)
		GetByID(ctx context.Context, id uint) (*roledto.Response, error)
		List(ctx context.Context) (*roledto.ListResponse, error)
		Update(ctx context.Context, id uint, req roledto.UpdateRequest) (*roledto.Response, error)
		Delete(ctx context.Context, id uint) error
		AssignPermissions(ctx context.Context, roleID uint, permissionIDs []uint) (*roledto.Response, error)
	}

	// Permission defines permission use case operations.
	Permission interface {
		List(ctx context.Context) ([]*entity.Permission, error)
		Create(ctx context.Context, req permissiondto.CreateRequest) (*permissiondto.Response, error)
		Delete(ctx context.Context, id uint) error
	}

	// BankStatement defines bank statement use case operations.
	BankStatement interface {
		Upload(ctx context.Context, req bankstatementdto.UploadRequest) (*bankstatementdto.ResponseWithItems, error)
		GetByID(ctx context.Context, id uint) (*bankstatementdto.ResponseWithItems, error)
		List(ctx context.Context, req bankstatementdto.ListRequest) (*bankstatementdto.ListResponse, error)
		UpdateLineItem(ctx context.Context, itemID uint, req bankstatementdto.UpdateLineItemRequest) (*bankstatementdto.LineItemResponse, error)
		Delete(ctx context.Context, id uint) error
		ListBanks(ctx context.Context) (*bankstatementdto.BankListResponse, error)
	}

	// Installment defines installment use case operations.
	Installment interface {
		Create(ctx context.Context, req installmentdto.CreateRequest) (*installmentdto.Response, error)
		GetByID(ctx context.Context, id uint) (*installmentdto.Response, error)
		List(ctx context.Context, req installmentdto.ListRequest) (*installmentdto.ListResponse, error)
		Update(ctx context.Context, id uint, req installmentdto.UpdateRequest) (*installmentdto.Response, error)
		Delete(ctx context.Context, id uint) error
		LinkItems(ctx context.Context, installmentID uint, req installmentdto.LinkItemsRequest) error
	}
)
