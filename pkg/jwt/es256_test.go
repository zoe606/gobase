package jwt_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/jwt"
)

// generateES256Keys generates a P-256 ECDSA key pair, writes them to temp files,
// and returns the private and public key file paths.
func generateES256Keys(t *testing.T) (privPath, pubPath string) {
	t.Helper()

	dir := t.TempDir()

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Write private key (EC PEM)
	privPath = filepath.Join(dir, "ec_private.pem")
	privFile, err := os.Create(privPath)
	require.NoError(t, err)
	defer privFile.Close()

	privBytes, err := x509.MarshalECPrivateKey(privKey)
	require.NoError(t, err)

	err = pem.Encode(privFile, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})
	require.NoError(t, err)

	// Write public key (PKIX PEM)
	pubPath = filepath.Join(dir, "ec_public.pem")
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

func TestES256_GenerateAndValidate(t *testing.T) {
	t.Parallel()

	privPath, pubPath := generateES256Keys(t)
	svc, err := jwt.NewES256(privPath, pubPath, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, expiresAt, err := svc.GenerateAccessToken(7, "ec@example.com", "viewer", []string{"read"})
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)
	require.Greater(t, expiresAt, time.Now().Unix())

	claims, err := svc.ValidateToken(tokenStr)
	require.NoError(t, err)
	require.NotNil(t, claims)
	require.Equal(t, uint(7), claims.UserID)
	require.Equal(t, "ec@example.com", claims.Email)
	require.Equal(t, "viewer", claims.Role)
	require.Equal(t, []string{"read"}, claims.Permissions)
}

func TestES256_WrongKey(t *testing.T) {
	t.Parallel()

	privPath1, pubPath1 := generateES256Keys(t)
	privPath2, pubPath2 := generateES256Keys(t)

	svc1, err := jwt.NewES256(privPath1, pubPath1, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	svc2, err := jwt.NewES256(privPath2, pubPath2, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, _, err := svc1.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	// Validate with svc2 (different public key) — should fail
	_, err = svc2.ValidateToken(tokenStr)
	require.ErrorIs(t, err, jwt.ErrInvalidToken)
}

func TestES256_InvalidKeyPath(t *testing.T) {
	t.Parallel()

	_, err := jwt.NewES256("/nonexistent/ec_private.pem", "/nonexistent/ec_public.pem", 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
}

func TestES256_RefreshToken(t *testing.T) {
	t.Parallel()

	privPath, pubPath := generateES256Keys(t)
	svc, err := jwt.NewES256(privPath, pubPath, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	token, expiresAt, err := svc.GenerateRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.True(t, expiresAt.After(time.Now()))
}

func TestES256_ExpiredToken(t *testing.T) {
	t.Parallel()

	privPath, pubPath := generateES256Keys(t)
	svc, err := jwt.NewES256(privPath, pubPath, 1*time.Nanosecond, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, _, err := svc.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	time.Sleep(2 * time.Millisecond)

	_, err = svc.ValidateToken(tokenStr)
	require.ErrorIs(t, err, jwt.ErrExpiredToken)
}

func TestES256_InvalidPEM(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(badFile, []byte("not a pem"), 0o600))

	_, err := jwt.NewES256(badFile, badFile, 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
}

func TestES256_PKCS8Key(t *testing.T) {
	t.Parallel()

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	dir := t.TempDir()

	// Write private key in PKCS8 format
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	require.NoError(t, err)
	privPathPKCS8 := filepath.Join(dir, "private.pem")
	require.NoError(t, os.WriteFile(privPathPKCS8, pem.EncodeToMemory(&pem.Block{
		Type: "PRIVATE KEY", Bytes: pkcs8Bytes,
	}), 0o600))

	// Write public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	pubPathPKCS8 := filepath.Join(dir, "public.pem")
	require.NoError(t, os.WriteFile(pubPathPKCS8, pem.EncodeToMemory(&pem.Block{
		Type: "PUBLIC KEY", Bytes: pubBytes,
	}), 0o600))

	svc, err := jwt.NewES256(privPathPKCS8, pubPathPKCS8, 15*time.Minute, 24*time.Hour)
	require.NoError(t, err)

	tokenStr, _, err := svc.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(tokenStr)
	require.NoError(t, err)
	require.Equal(t, uint(1), claims.UserID)
}

func TestES256_WrongPublicKeyType(t *testing.T) {
	t.Parallel()

	// Generate EC private key and RSA public key — type mismatch
	ecPrivPath, _ := generateES256Keys(t)
	_, rsaPubPath := generateRSAKeys(t)

	_, err := jwt.NewES256(ecPrivPath, rsaPubPath, 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not an ECDSA public key")
}

func TestES256_InvalidPublicKeyPEM(t *testing.T) {
	t.Parallel()

	ecPrivPath, _ := generateES256Keys(t)
	dir := t.TempDir()
	badPub := filepath.Join(dir, "bad.pem")
	require.NoError(t, os.WriteFile(badPub, []byte("not a pem"), 0o600))

	_, err := jwt.NewES256(ecPrivPath, badPub, 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
}

func TestES256_WrongCurve(t *testing.T) {
	t.Parallel()

	// Generate P-384 key (not P-256)
	privKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	dir := t.TempDir()
	privBytes, err := x509.MarshalECPrivateKey(privKey)
	require.NoError(t, err)
	privPath := filepath.Join(dir, "private.pem")
	require.NoError(t, os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{
		Type: "EC PRIVATE KEY", Bytes: privBytes,
	}), 0o600))

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.NoError(t, err)
	pubPath := filepath.Join(dir, "public.pem")
	require.NoError(t, os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{
		Type: "PUBLIC KEY", Bytes: pubBytes,
	}), 0o600))

	_, err = jwt.NewES256(privPath, pubPath, 15*time.Minute, 24*time.Hour)
	require.Error(t, err)
	require.Contains(t, err.Error(), "P-256 curve")
}

func TestES256_GetExpiry(t *testing.T) {
	t.Parallel()

	accessExpiry := 10 * time.Minute
	refreshExpiry := 48 * time.Hour

	privPath, pubPath := generateES256Keys(t)
	svc, err := jwt.NewES256(privPath, pubPath, accessExpiry, refreshExpiry)
	require.NoError(t, err)

	require.Equal(t, accessExpiry, svc.GetAccessExpiry())
	require.Equal(t, refreshExpiry, svc.GetRefreshExpiry())
}
