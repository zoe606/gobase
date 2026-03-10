package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestRole_HasPermission(t *testing.T) {
	t.Parallel()

	role := &entity.Role{
		Permissions: []entity.Permission{
			{Name: "user:read"},
			{Name: "user:write"},
		},
	}

	t.Run("has permission", func(t *testing.T) {
		t.Parallel()
		require.True(t, role.HasPermission("user:read"))
	})

	t.Run("missing permission", func(t *testing.T) {
		t.Parallel()
		require.False(t, role.HasPermission("admin:delete"))
	})

	t.Run("empty permissions", func(t *testing.T) {
		t.Parallel()
		emptyRole := &entity.Role{}
		require.False(t, emptyRole.HasPermission("user:read"))
	})
}

func TestRole_GetPermissionNames(t *testing.T) {
	t.Parallel()

	t.Run("with permissions", func(t *testing.T) {
		t.Parallel()
		role := &entity.Role{
			Permissions: []entity.Permission{
				{Name: "user:read"},
				{Name: "user:write"},
			},
		}
		names := role.GetPermissionNames()
		require.Equal(t, []string{"user:read", "user:write"}, names)
	})

	t.Run("empty permissions", func(t *testing.T) {
		t.Parallel()
		role := &entity.Role{}
		names := role.GetPermissionNames()
		require.Empty(t, names)
	})
}
