package authdto

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
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// NewUserResponse creates a UserResponse from user entity.
func NewUserResponse(u *entity.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Role:  u.Role.Name,
	}
}
