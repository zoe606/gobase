package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/jwt"
)

// generateRSAKeys generates a 2048-bit RSA key pair, writes them to temp files,
// and returns the private and public key file paths.
func generateRSAKeys(t *testing.T) (privPath, pubPath string) {
	t.Helper()

	dir := t.TempDir()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Write private key (PKCS1 PEM)
	privPath = filepath.Join(dir, "private.pem")
	privFile, err := os.Create(privPath)
	require.NoError(t, err)
	defer privFile.Close()

	err = pem.Encode(privFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	})
	require.NoError(t, err)

	// Write public key (PKIX PEM)
	pubPath = filepath.Join(dir, "public.pem")
	pubFile, err := os.Create(pubPath)
	require.NoError(t, err)
	defer pubFile.Close()

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	err = pem.Encode(pubFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	require.NoError(t, err)

	return privPath, pubPath
}

func TestRS256_GenerateAndValidate(t *testing.T) {
	t.Parallel()

	privPath, pubPath := generateRSAKeys(t)
	svc, err := jwt.NewRS256(privPath, pubPath, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, expiresAt, err := svc.GenerateAccessToken(42, "user@example.com", "admin", []string{"read", "write"})
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)
	require.Greater(t, expiresAt, time.Now().Unix())

	claims, err := svc.ValidateToken(tokenStr)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.Equal(t, uint(42), claims.UserID)
	require.Equal(t, "user@example.com", claims.Email)
	require.Equal(t, "admin", claims.Role)
	require.Equal(t, []string{"read", "write"}, claims.Permissions)
}

func TestRS256_WrongKey(t *testing.T) {
	t.Parallel()

	privPath1, pubPath1 := generateRSAKeys(t)
	privPath2, _ := generateRSAKeys(t)

	// Sign with key1 private, but verify with key2 public (different pair)
	svc1, err := jwt.NewRS256(privPath1, pubPath1, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	_, pubPath2 := generateRSAKeys(t)
	svc2, err := jwt.NewRS256(privPath2, pubPath2, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, _, err := svc1.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	// Validate with svc2 (different public key) — should fail
	_, err = svc2.ValidateToken(tokenStr)
	require.ErrorIs(t, err, jwt.ErrInvalidToken)
}

func TestRS256_InvalidKeyPath(t *testing.T) {
	t.Parallel()

	_, err := jwt.NewRS256("/nonexistent/private.pem", "/nonexistent/public.pem", 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
}

func TestRS256_RefreshToken(t *testing.T) {
	t.Parallel()

	privPath, pubPath := generateRSAKeys(t)
	svc, err := jwt.NewRS256(privPath, pubPath, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	token, expiresAt, err := svc.GenerateRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.True(t, expiresAt.After(time.Now()))
}

func TestRS256_ExpiredToken(t *testing.T) {
	t.Parallel()

	privPath, pubPath := generateRSAKeys(t)
	svc, err := jwt.NewRS256(privPath, pubPath, 1*time.Nanosecond, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, _, err := svc.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	time.Sleep(2 * time.Millisecond)

	_, err = svc.ValidateToken(tokenStr)
	require.ErrorIs(t, err, jwt.ErrExpiredToken)
}

func TestRS256_InvalidPEM(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(badFile, []byte("not a pem"), 0o600))

	_, err := jwt.NewRS256(badFile, badFile, 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
}

func TestRS256_PKCS8Key(t *testing.T) {
	t.Parallel()

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	dir := t.TempDir()

	// Write private key in PKCS8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	require.NoError(t, err)
	privPath := filepath.Join(dir, "private.pem")
	require.NoError(t, os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{
		Type: "PRIVATE KEY", Bytes: pkcs8Bytes,
	}), 0o600))

	// Write public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	pubPath := filepath.Join(dir, "public.pem")
	require.NoError(t, os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{
		Type: "PUBLIC KEY", Bytes: pubBytes,
	}), 0o600))

	svc, err := jwt.NewRS256(privPath, pubPath, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, _, err := svc.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(tokenStr)
	require.NoError(t, err)
	require.Equal(t, uint(1), claims.UserID)
}

func TestRS256_WrongPublicKeyType(t *testing.T) {
	t.Parallel()

	// Generate RSA private key and EC public key — type mismatch
	privPath, _ := generateRSAKeys(t)
	_, ecPubPath := generateES256Keys(t)

	_, err := jwt.NewRS256(privPath, ecPubPath, 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not an RSA public key")
}

func TestRS256_InvalidPublicKeyPEM(t *testing.T) {
	t.Parallel()

	privPath, _ := generateRSAKeys(t)
	dir := t.TempDir()
	badPub := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(badPub, []byte("not a pem"), 0o600))

	_, err := jwt.NewRS256(privPath, badPub, 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
}

func TestRS256_GetExpiry(t *testing.T) {
	t.Parallel()

	accessExpiry := 15 * time.Minute
	refreshExpiry := 24 * time.Hour

	privPath, pubPath := generateRSAKeys(t)
	svc, err := jwt.NewRS256(privPath, pubPath, accessExpiry, refreshExpiry)
	require.NoError(t, err)

	require.Equal(t, accessExpiry, svc.GetAccessExpiry())
	require.Equal(t, refreshExpiry, svc.GetRefreshExpiry())
}
