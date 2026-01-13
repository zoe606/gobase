// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	articledto "go-boilerplate/internal/dto/article"
	authdto "go-boilerplate/internal/dto/auth"
	mediadto "go-boilerplate/internal/dto/media"
	profiledto "go-boilerplate/internal/dto/profile"
	translationdto "go-boilerplate/internal/dto/translation"
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
)
