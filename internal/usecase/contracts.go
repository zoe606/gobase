// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	authdto "go-boilerplate/internal/dto/auth"
	translationdto "go-boilerplate/internal/dto/translation"
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
)
