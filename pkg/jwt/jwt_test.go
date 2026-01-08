package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/jwt"
)

func TestGenerateAccessToken(t *testing.T) {
	t.Parallel()

	svc := jwt.New("test-secret-key", 15*time.Minute, 24*time.Hour)

	tests := []struct {
		name        string
		userID      uint
		email       string
		role        string
		permissions []string
	}{
		{
			name:        "success",
			userID:      1,
			email:       "test@example.com",
			role:        "user",
			permissions: []string{"read", "write"},
		},
		{
			name:        "empty permissions",
			userID:      2,
			email:       "admin@example.com",
			role:        "admin",
			permissions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, expiresAt, err := svc.GenerateAccessToken(tt.userID, tt.email, tt.role, tt.permissions)

			require.NoError(t, err)
			require.NotEmpty(t, token)
			require.Greater(t, expiresAt, time.Now().Unix())
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	t.Parallel()

	svc := jwt.New("test-secret-key", 15*time.Minute, 24*time.Hour)

	token1, expiresAt1, err1 := svc.GenerateRefreshToken()
	require.NoError(t, err1)
	require.NotEmpty(t, token1)
	require.True(t, expiresAt1.After(time.Now()))

	// Generate another token to ensure uniqueness
	token2, _, err2 := svc.GenerateRefreshToken()
	require.NoError(t, err2)
	require.NotEqual(t, token1, token2)
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	svc := jwt.New("test-secret-key", 15*time.Minute, 24*time.Hour)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "valid token",
			setup: func() string {
				token, _, _ := svc.GenerateAccessToken(1, "test@example.com", "user", []string{"read"}) //nolint:errcheck // test setup
				return token
			},
			wantErr: nil,
		},
		{
			name: "malformed token",
			setup: func() string {
				return "invalid-token"
			},
			wantErr: jwt.ErrInvalidToken,
		},
		{
			name: "wrong signature",
			setup: func() string {
				otherSvc := jwt.New("different-secret", 15*time.Minute, 24*time.Hour)
				token, _, _ := otherSvc.GenerateAccessToken(1, "test@example.com", "user", nil) //nolint:errcheck // test setup
				return token
			},
			wantErr: jwt.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token := tt.setup()
			claims, err := svc.ValidateToken(token)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)
			require.Equal(t, uint(1), claims.UserID)
			require.Equal(t, "test@example.com", claims.Email)
		})
	}
}

func TestGetExpiry(t *testing.T) {
	t.Parallel()

	accessExpiry := 15 * time.Minute
	refreshExpiry := 24 * time.Hour

	svc := jwt.New("test-secret", accessExpiry, refreshExpiry)

	require.Equal(t, accessExpiry, svc.GetAccessExpiry())
	require.Equal(t, refreshExpiry, svc.GetRefreshExpiry())
}
