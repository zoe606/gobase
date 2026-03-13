package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Common errors.
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// isExpiredError checks if a JWT error is an expiration error.
func isExpiredError(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired)
}

// Service defines the JWT service interface.
type Service interface {
	// GenerateAccessToken generates a new access token.
	GenerateAccessToken(userID uint, email, role string, permissions []string) (string, int64, error)

	// GenerateRefreshToken generates a new refresh token.
	GenerateRefreshToken() (string, time.Time, error)

	// ValidateToken validates a token and returns the claims.
	ValidateToken(tokenString string) (*Claims, error)

	// GetAccessExpiry returns the access token expiry duration.
	GetAccessExpiry() time.Duration

	// GetRefreshExpiry returns the refresh token expiry duration.
	GetRefreshExpiry() time.Duration
}

// service implements the Service interface.
type service struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// New creates a new JWT service.
func New(secretKey string, accessExpiry, refreshExpiry time.Duration) Service {
	return &service{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken generates a new access token.
func (s *service) GenerateAccessToken(userID uint, email, role string, permissions []string) (tokenStr string, expiresAtUnix int64, err error) {
	expiresAtTime := time.Now().Add(s.accessExpiry)

	claims := &Claims{
		UserID:      userID,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAtTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAtTime.Unix(), nil
}

// GenerateRefreshToken generates a new refresh token.
func (s *service) GenerateRefreshToken() (string, time.Time, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}

	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(s.refreshExpiry)

	return token, expiresAt, nil
}

// ValidateToken validates a token and returns the claims.
func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})
	if err != nil {
		if isExpiredError(err) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetAccessExpiry returns the access token expiry duration.
func (s *service) GetAccessExpiry() time.Duration {
	return s.accessExpiry
}

// GetRefreshExpiry returns the refresh token expiry duration.
func (s *service) GetRefreshExpiry() time.Duration {
	return s.refreshExpiry
}
