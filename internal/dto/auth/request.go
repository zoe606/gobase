// Package auth provides DTOs for authentication operations.
package authdto

// RegisterRequest represents registration request.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
	Name     string `json:"name" validate:"required,min=2" example:"John Doe"`
}

// LoginRequest represents login request.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// RefreshRequest represents refresh token request.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
