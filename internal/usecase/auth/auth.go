// Package auth provides authentication use cases.
package auth

import (
	"context"
	"errors"
	"fmt"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/hasher"
	"go-boilerplate/pkg/jwt"
)

// Common errors.
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailExists         = errors.New("email already exists")
	ErrUserNotActive       = errors.New("user account is not active")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrDefaultRoleNotFound = errors.New("default role not found")
)

// UseCase implements authentication business logic.
type UseCase struct {
	userRepo         repo.UserRepo
	roleRepo         repo.RoleRepo
	refreshTokenRepo repo.RefreshTokenRepo
	jwtService       jwt.Service
}

// New creates a new auth use case.
func New(
	userRepo repo.UserRepo,
	roleRepo repo.RoleRepo,
	refreshTokenRepo repo.RefreshTokenRepo,
	jwtService jwt.Service,
) *UseCase {
	return &UseCase{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

// Register registers a new user.
func (uc *UseCase) Register(ctx context.Context, input authdto.RegisterInput) (*authdto.RegisterOutput, error) {
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
		return nil, fmt.Errorf("Auth - Register - GetRole: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("Auth - Register: %w", ErrDefaultRoleNotFound)
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
	accessToken, expiresAt, err := uc.jwtService.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role.Name,
		user.Role.GetPermissionNames(),
	)
	if err != nil {
		return nil, fmt.Errorf("Auth - Register - GenerateAccessToken: %w", err)
	}

	refreshToken, refreshExpiresAt, err := uc.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("Auth - Register - GenerateRefreshToken: %w", err)
	}

	// Store refresh token
	if err := uc.refreshTokenRepo.Create(ctx, &entity.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return nil, fmt.Errorf("Auth - Register - StoreRefreshToken: %w", err)
	}

	return &authdto.RegisterOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Login authenticates a user.
func (uc *UseCase) Login(ctx context.Context, input authdto.LoginInput) (*authdto.LoginOutput, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("Auth - Login - GetByEmail: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check password
	if !hasher.Check(input.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.Active {
		return nil, ErrUserNotActive
	}

	// Generate tokens
	accessToken, expiresAt, err := uc.jwtService.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role.Name,
		user.Role.GetPermissionNames(),
	)
	if err != nil {
		return nil, fmt.Errorf("Auth - Login - GenerateAccessToken: %w", err)
	}

	refreshToken, refreshExpiresAt, err := uc.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("Auth - Login - GenerateRefreshToken: %w", err)
	}

	// Store refresh token
	if err := uc.refreshTokenRepo.Create(ctx, &entity.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return nil, fmt.Errorf("Auth - Login - StoreRefreshToken: %w", err)
	}

	return &authdto.LoginOutput{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Refresh refreshes an access token using a refresh token.
func (uc *UseCase) Refresh(ctx context.Context, input authdto.RefreshInput) (*authdto.RefreshOutput, error) {
	// Get refresh token
	token, err := uc.refreshTokenRepo.GetByToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("Auth - Refresh - GetByToken: %w", err)
	}
	if token == nil || token.IsExpired() {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, fmt.Errorf("Auth - Refresh - GetByID: %w", err)
	}
	if user == nil || !user.Active {
		return nil, ErrInvalidToken
	}

	// Delete old refresh token
	if err := uc.refreshTokenRepo.DeleteByToken(ctx, input.RefreshToken); err != nil {
		return nil, fmt.Errorf("Auth - Refresh - DeleteByToken: %w", err)
	}

	// Generate new tokens
	accessToken, expiresAt, err := uc.jwtService.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role.Name,
		user.Role.GetPermissionNames(),
	)
	if err != nil {
		return nil, fmt.Errorf("Auth - Refresh - GenerateAccessToken: %w", err)
	}

	newRefreshToken, refreshExpiresAt, err := uc.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("Auth - Refresh - GenerateRefreshToken: %w", err)
	}

	// Store new refresh token
	if err := uc.refreshTokenRepo.Create(ctx, &entity.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return nil, fmt.Errorf("Auth - Refresh - StoreRefreshToken: %w", err)
	}

	return &authdto.RefreshOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout invalidates a refresh token.
func (uc *UseCase) Logout(ctx context.Context, refreshToken string) error {
	if err := uc.refreshTokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("Auth - Logout - DeleteByToken: %w", err)
	}
	return nil
}

// GetCurrentUser retrieves the current user by ID.
func (uc *UseCase) GetCurrentUser(ctx context.Context, userID uint) (*authdto.UserOutput, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("Auth - GetCurrentUser - GetByID: %w", err)
	}
	if user == nil {
		return nil, nil
	}
	return &authdto.UserOutput{User: user}, nil
}
