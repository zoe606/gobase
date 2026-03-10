package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestEmailVerification_IsExpired(t *testing.T) {
	t.Parallel()

	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		ev := &entity.EmailVerification{ExpiresAt: time.Now().Add(-time.Hour)}
		require.True(t, ev.IsExpired())
	})

	t.Run("not expired", func(t *testing.T) {
		t.Parallel()
		ev := &entity.EmailVerification{ExpiresAt: time.Now().Add(time.Hour)}
		require.False(t, ev.IsExpired())
	})
}

func TestEmailVerification_IsUsed(t *testing.T) {
	t.Parallel()

	t.Run("not used", func(t *testing.T) {
		t.Parallel()
		ev := &entity.EmailVerification{UsedAt: nil}
		require.False(t, ev.IsUsed())
	})

	t.Run("used", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		ev := &entity.EmailVerification{UsedAt: &now}
		require.True(t, ev.IsUsed())
	})
}
