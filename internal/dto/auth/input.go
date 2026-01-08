// Package auth provides DTOs for authentication operations.
package auth

// RegisterInput represents registration input.
type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

// LoginInput represents login input.
type LoginInput struct {
	Email    string
	Password string
}

// RefreshInput represents refresh token input.
type RefreshInput struct {
	RefreshToken string
}

// LogoutInput represents logout input.
type LogoutInput struct {
	RefreshToken string
}
