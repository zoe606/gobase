// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	authdto "go-boilerplate/internal/dto/auth"
	mediadto "go-boilerplate/internal/dto/media"
	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation defines the translation use case interface.
	Translation interface {
		Translate(context.Context, translationdto.TranslateRequest) (*translationdto.TranslationResponse, error)
		History(context.Context) (*translationdto.HistoryResponse, error)
	}

	// Auth defines the authentication use case interface.
	Auth interface {
		Register(ctx context.Context, input authdto.RegisterRequest) (*authdto.LoginResponse, error)
		Login(ctx context.Context, input authdto.LoginRequest) (*authdto.LoginResponse, error)
		Refresh(ctx context.Context, input authdto.RefreshRequest) (*authdto.TokenResponse, error)
		Logout(ctx context.Context, refreshToken string) error
		GetCurrentUser(ctx context.Context, userID uint) (*authdto.UserResponse, error)
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
)
