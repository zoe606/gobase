package authdto_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
)

func TestNewLoginResponse(t *testing.T) {
	t.Parallel()

	user := &entity.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  entity.Role{Name: "admin"},
	}

	resp := authdto.NewLoginResponse(user, "access-token", "refresh-token", 1234567890)
	require.Equal(t, "access-token", resp.AccessToken)
	require.Equal(t, "refresh-token", resp.RefreshToken)
	require.Equal(t, int64(1234567890), resp.ExpiresAt)
	require.Equal(t, uint(1), resp.User.ID)
	require.Equal(t, "test@example.com", resp.User.Email)
	require.Equal(t, "admin", resp.User.Role)
}

func TestNewTokenResponse(t *testing.T) {
	t.Parallel()

	resp := authdto.NewTokenResponse("access", "refresh", 9999)
	require.Equal(t, "access", resp.AccessToken)
	require.Equal(t, "refresh", resp.RefreshToken)
	require.Equal(t, int64(9999), resp.ExpiresAt)
}

func TestNewUserResponse(t *testing.T) {
	t.Parallel()

	user := &entity.User{
		ID:    42,
		Email: "user@test.com",
		Name:  "John",
		Role:  entity.Role{Name: "user"},
	}

	resp := authdto.NewUserResponse(user)
	require.Equal(t, uint(42), resp.ID)
	require.Equal(t, "user@test.com", resp.Email)
	require.Equal(t, "John", resp.Name)
	require.Equal(t, "user", resp.Role)
}
