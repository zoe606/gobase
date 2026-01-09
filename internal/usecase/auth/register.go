package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/hasher"
)

// Register registers a new user.
func (uc *UseCase) Register(ctx context.Context, input authdto.RegisterRequest) (*authdto.LoginResponse, error) {
	// Check if email exists
	exists, err := uc.userRepo.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("Auth - Register - EmailExists: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Get default role
	role, err := uc.roleRepo.GetByName(ctx, "user")
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, fmt.Errorf("Auth - Register: %w", ErrDefaultRoleNotFound)
		}
		return nil, fmt.Errorf("Auth - Register - GetRole: %w", err)
	}

	// Hash password
	passwordHash, err := hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("Auth - Register - Hash: %w", err)
	}

	// Create user
	user := &entity.User{
		Email:    input.Email,
		Password: passwordHash,
		Name:     input.Name,
		RoleID:   role.ID,
		Active:   true,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("Auth - Register - Create: %w", err)
	}

	// Load role for token generation
	user.Role = *role

	// Generate tokens
	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("Auth - Register - %w", err)
	}

	// Store refresh token
	if err := uc.storeRefreshToken(ctx, user.ID, tokens.RefreshToken, tokens.RefreshExpiresAt); err != nil {
		return nil, fmt.Errorf("Auth - Register - StoreRefreshToken: %w", err)
	}

	// Enqueue welcome email (non-blocking, best effort)
	uc.enqueueWelcomeEmail(user)

	return authdto.NewLoginResponse(user, tokens.AccessToken, tokens.RefreshToken, tokens.AccessExpiresAt), nil
}

// enqueueWelcomeEmail enqueues a welcome email task if Asynq is configured.
func (uc *UseCase) enqueueWelcomeEmail(user *entity.User) {
	if uc.asynqClient == nil {
		return
	}

	task, err := tasks.NewWelcomeEmailTask(user.Email, user.Name, uc.appName)
	if err != nil {
		// Log error but don't fail registration
		return
	}

	_, _ = uc.asynqClient.Enqueue(task, asynq.Queue("default"))
}
