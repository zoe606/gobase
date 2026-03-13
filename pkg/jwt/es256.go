package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// es256Service implements the Service interface using ES256.
type es256Service struct {
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewES256 creates a new JWT service using ES256 (ECDSA + SHA-256 on P-256 curve) signing.
func NewES256(privateKeyPath, publicKeyPath string, accessExpiry, refreshExpiry time.Duration) (Service, error) {
	privKey, err := loadECPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("jwt.NewES256: load private key: %w", err)
	}

	if privKey.Curve != elliptic.P256() {
		return nil, errors.New("jwt.NewES256: private key must use P-256 curve for ES256")
	}

	pubKey, err := loadECPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("jwt.NewES256: load public key: %w", err)
	}

	return &es256Service{
		privateKey:    privKey,
		publicKey:     pubKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// GenerateAccessToken generates a new access token signed with ES256.
func (s *es256Service) GenerateAccessToken(userID uint, email, role string, permissions []string) (tokenStr string, expiresAtUnix int64, err error) {
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

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodES256, claims)
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAtTime.Unix(), nil
}

// GenerateRefreshToken generates a new opaque refresh token.
func (s *es256Service) GenerateRefreshToken() (string, time.Time, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, err
	}

	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(s.refreshExpiry)

	return token, expiresAt, nil
}

// ValidateToken validates an ES256-signed token and returns the claims.
func (s *es256Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (any, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodECDSA); !ok {
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
func (s *es256Service) GetAccessExpiry() time.Duration {
	return s.accessExpiry
}

// GetRefreshExpiry returns the refresh token expiry duration.
func (s *es256Service) GetRefreshExpiry() time.Duration {
	return s.refreshExpiry
}

// loadECPrivateKey loads an ECDSA private key from a PEM file.
// It tries EC format first, then falls back to PKCS8.
func loadECPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// Try EC private key format first
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Fall back to PKCS8
	keyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key (EC/PKCS8): %w", err)
	}

	ecKey, ok := keyIface.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("PKCS8 key is not an ECDSA private key")
	}

	return ecKey, nil
}

// loadECPublicKey loads an ECDSA public key from a PEM file using PKIX format.
func loadECPublicKey(path string) (*ecdsa.PublicKey, error) {
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

	ecKey, ok := keyIface.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not an ECDSA public key")
	}

	return ecKey, nil
}
