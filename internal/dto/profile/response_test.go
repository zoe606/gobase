package profiledto_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	profiledto "go-boilerplate/internal/dto/profile"
	"go-boilerplate/internal/entity"
)

func TestFromEntity(t *testing.T) {
	t.Parallel()

	t.Run("with avatar", func(t *testing.T) {
		t.Parallel()
		avatarID := uint(100)
		p := &entity.Profile{
			UserID:        1,
			AvatarMediaID: &avatarID,
			Bio:           "Hello",
			Phone:         "+1234567890",
		}

		resp := profiledto.FromEntity(p, "https://example.com/avatar.jpg")
		require.Equal(t, uint(1), resp.UserID)
		require.Equal(t, "Hello", resp.Bio)
		require.Equal(t, "+1234567890", resp.Phone)
		require.NotNil(t, resp.Avatar)
		require.Equal(t, uint(100), resp.Avatar.ID)
		require.Equal(t, "https://example.com/avatar.jpg", resp.Avatar.URL)
	})

	t.Run("without avatar - nil media id", func(t *testing.T) {
		t.Parallel()
		p := &entity.Profile{
			UserID: 1,
			Bio:    "Hello",
		}

		resp := profiledto.FromEntity(p, "https://example.com/avatar.jpg")
		require.Nil(t, resp.Avatar)
	})

	t.Run("without avatar - empty url", func(t *testing.T) {
		t.Parallel()
		avatarID := uint(100)
		p := &entity.Profile{
			UserID:        1,
			AvatarMediaID: &avatarID,
		}

		resp := profiledto.FromEntity(p, "")
		require.Nil(t, resp.Avatar)
	})
}
