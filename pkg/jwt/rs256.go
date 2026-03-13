package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// rs256Service implements the Service interface using RS256.
type rs256Service struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewRS256 creates a new JWT service using RS256 (RSA + SHA-256) signing.
func NewRS256(privateKeyPath, publicKeyPath string, accessExpiry, refreshExpiry time.Duration) (Service, error) {
	privKey, err := loadRSAPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("jwt.NewRS256: load private key: %w", err)
	}

	pubKey, err := loadRSAPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("jwt.NewRS256: load public key: %w", err)
	}

	return &rs256Service{
		privateKey:    privKey,
		publicKey:     pubKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// GenerateAccessToken generates a new access token signed with RS256.
func (s *rs256Service) GenerateAccessToken(userID uint, email, role string, permissions []string) (tokenStr string, expiresAtUnix int64, err error) {
	expiresAtTime := time.Now().Add(s.accessExpiry)

	claims := &Claims{
		UserID:      userID,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(expiresAtTime),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			NotBefore: jwtlib.NewNumericDate(time.Now()),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAtTime.Unix(), nil
}

// GenerateRefreshToken generates a new opaque refresh token.
func (s *rs256Service) GenerateRefreshToken() (string, time.Time, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}

	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(s.refreshExpiry)

	return token, expiresAt, nil
}

// ValidateToken validates an RS256-signed token and returns the claims.
func (s *rs256Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil
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
func (s *rs256Service) GetAccessExpiry() time.Duration {
	return s.accessExpiry
}

// GetRefreshExpiry returns the refresh token expiry duration.
func (s *rs256Service) GetRefreshExpiry() time.Duration {
	return s.refreshExpiry
}

// loadRSAPrivateKey loads an RSA private key from a PEM file.
// It tries PKCS1 format first, then falls back to PKCS8.
func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Try PKCS1 first
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Fall back to PKCS8
	keyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key (PKCS1/PKCS8): %w", err)
	}

	rsaKey, ok := keyIface.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("PKCS8 key is not an RSA private key")
	}

	return rsaKey, nil
}

// loadRSAPublicKey loads an RSA public key from a PEM file using PKIX format.
func loadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	keyIface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	rsaKey, ok := keyIface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not an RSA public key")
	}

	return rsaKey, nil
}
