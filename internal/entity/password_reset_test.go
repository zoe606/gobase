package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestPasswordReset_IsExpired(t *testing.T) {
	t.Parallel()

	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		pr := &entity.PasswordReset{ExpiresAt: time.Now().Add(-time.Hour)}
		require.True(t, pr.IsExpired())
	})

	t.Run("not expired", func(t *testing.T) {
		t.Parallel()
		pr := &entity.PasswordReset{ExpiresAt: time.Now().Add(time.Hour)}
		require.False(t, pr.IsExpired())
	})
}

func TestPasswordReset_IsUsed(t *testing.T) {
	t.Parallel()

	t.Run("not used", func(t *testing.T) {
		t.Parallel()
		pr := &entity.PasswordReset{UsedAt: nil}
		require.False(t, pr.IsUsed())
	})

	t.Run("used", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		pr := &entity.PasswordReset{UsedAt: &now}
		require.True(t, pr.IsUsed())
	})
}
