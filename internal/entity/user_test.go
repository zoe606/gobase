package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestUser_IsEmailVerified(t *testing.T) {
	t.Parallel()

	t.Run("not verified", func(t *testing.T) {
		t.Parallel()
		u := &entity.User{EmailVerifiedAt: nil}
		require.False(t, u.IsEmailVerified())
	})

	t.Run("verified", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		u := &entity.User{EmailVerifiedAt: &now}
		require.True(t, u.IsEmailVerified())
	})
}
