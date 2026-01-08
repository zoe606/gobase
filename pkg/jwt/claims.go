// Package jwt provides JWT token generation and validation.
package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure.
type Claims struct {
	UserID      uint     `json:"user_id"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// TokenPair represents an access and refresh token pair.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}
