package entity_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestRefreshToken_IsExpired(t *testing.T) {
	t.Parallel()

	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		rt := &entity.RefreshToken{ExpiresAt: time.Now().Add(-time.Hour)}
		require.True(t, rt.IsExpired())
	})

	t.Run("not expired", func(t *testing.T) {
		t.Parallel()
		rt := &entity.RefreshToken{ExpiresAt: time.Now().Add(time.Hour)}
		require.False(t, rt.IsExpired())
	})
}
