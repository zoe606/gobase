package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestProfile_HasAvatar(t *testing.T) {
	t.Parallel()

	t.Run("no avatar", func(t *testing.T) {
		t.Parallel()
		p := &entity.Profile{AvatarMediaID: nil}
		require.False(t, p.HasAvatar())
	})

	t.Run("has avatar", func(t *testing.T) {
		t.Parallel()
		id := uint(1)
		p := &entity.Profile{AvatarMediaID: &id}
		require.True(t, p.HasAvatar())
	})
}
