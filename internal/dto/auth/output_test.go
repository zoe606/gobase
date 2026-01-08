//nolint:dupl // Test functions intentionally have similar structure.
package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
)

func TestRegisterOutput_ToResponse(t *testing.T) {
	t.Parallel()

	output := &auth.RegisterOutput{
		User: &entity.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
			Role:  entity.Role{Name: "user"},
		},
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		ExpiresAt:    1234567890,
	}

	resp := output.ToResponse()

	require.Equal(t, "access_token_123", resp.AccessToken)
	require.Equal(t, "refresh_token_456", resp.RefreshToken)
	require.Equal(t, int64(1234567890), resp.ExpiresAt)
	require.Equal(t, uint(1), resp.User.ID)
	require.Equal(t, "test@example.com", resp.User.Email)
	require.Equal(t, "Test User", resp.User.Name)
	require.Equal(t, "user", resp.User.Role)
}

func TestLoginOutput_ToResponse(t *testing.T) {
	t.Parallel()

	output := &auth.LoginOutput{
		User: &entity.User{
			ID:    2,
			Email: "login@example.com",
			Name:  "Login User",
			Role:  entity.Role{Name: "admin"},
		},
		AccessToken:  "login_access_token",
		RefreshToken: "login_refresh_token",
		ExpiresAt:    9876543210,
	}

	resp := output.ToResponse()

	require.Equal(t, "login_access_token", resp.AccessToken)
	require.Equal(t, "login_refresh_token", resp.RefreshToken)
	require.Equal(t, int64(9876543210), resp.ExpiresAt)
	require.Equal(t, uint(2), resp.User.ID)
	require.Equal(t, "login@example.com", resp.User.Email)
	require.Equal(t, "Login User", resp.User.Name)
	require.Equal(t, "admin", resp.User.Role)
}

func TestRefreshOutput_ToResponse(t *testing.T) {
	t.Parallel()

	output := &auth.RefreshOutput{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresAt:    1111111111,
	}

	resp := output.ToResponse()

	require.Equal(t, "new_access_token", resp.AccessToken)
	require.Equal(t, "new_refresh_token", resp.RefreshToken)
	require.Equal(t, int64(1111111111), resp.ExpiresAt)
}

func TestUserOutput_ToResponse(t *testing.T) {
	t.Parallel()

	output := &auth.UserOutput{
		User: &entity.User{
			ID:    5,
			Email: "user@example.com",
			Name:  "Regular User",
			Role:  entity.Role{Name: "moderator"},
		},
	}

	resp := output.ToResponse()

	require.Equal(t, uint(5), resp.ID)
	require.Equal(t, "user@example.com", resp.Email)
	require.Equal(t, "Regular User", resp.Name)
	require.Equal(t, "moderator", resp.Role)
}
