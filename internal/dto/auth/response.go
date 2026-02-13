package auth

import "go-boilerplate/internal/entity"

// LoginResponse is the API response for login/register.
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    int64        `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// NewLoginResponse creates a LoginResponse from user entity and tokens.
func NewLoginResponse(user *entity.User, accessToken, refreshToken string, expiresAt int64) *LoginResponse {
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         NewUserResponse(user),
	}
}

// TokenResponse is the API response for token refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// NewTokenResponse creates a TokenResponse from tokens.
func NewTokenResponse(accessToken, refreshToken string, expiresAt int64) *TokenResponse {
	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
}

// UserResponse is the API representation of a user.
type UserResponse struct {
	ID     uint         `json:"id"`
	Email  string       `json:"email"`
	Name   string       `json:"name"`
	Active bool         `json:"active"`
	Role   RoleResponse `json:"role"`
}

// RoleResponse is the API representation of a role in auth context.
type RoleResponse struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// NewUserResponse creates a UserResponse from user entity.
func NewUserResponse(u *entity.User) UserResponse {
	return UserResponse{
		ID:     u.ID,
		Email:  u.Email,
		Name:   u.Name,
		Active: u.Active,
		Role: RoleResponse{
			ID:          u.Role.ID,
			Name:        u.Role.Name,
			Permissions: u.Role.GetPermissionNames(),
		},
	}
}
