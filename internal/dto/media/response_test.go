package mediadto_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	mediadto "go-boilerplate/internal/dto/media"
	"go-boilerplate/internal/entity"
)

func TestFromEntity(t *testing.T) {
	t.Parallel()

	width := 800
	height := 600
	m := &entity.Media{
		ID:             1,
		AttachableType: "users",
		AttachableID:   2,
		Collection:     "avatar",
		Filename:       "test.jpg",
		OriginalName:   "photo.jpg",
		MimeType:       "image/jpeg",
		Size:           1024,
		Type:           entity.MediaTypeImage,
		Width:          &width,
		Height:         &height,
		Variants:       entity.JSONMap{"thumb": "path/thumb.jpg"},
	}

	resp := mediadto.FromEntity(m)
	require.Equal(t, uint(1), resp.ID)
	require.Equal(t, "users", resp.AttachableType)
	require.Equal(t, uint(2), resp.AttachableID)
	require.Equal(t, "avatar", resp.Collection)
	require.Equal(t, "test.jpg", resp.Filename)
	require.Equal(t, "photo.jpg", resp.OriginalName)
	require.Equal(t, "image/jpeg", resp.MimeType)
	require.Equal(t, int64(1024), resp.Size)
	require.Equal(t, entity.MediaTypeImage, resp.Type)
	require.Equal(t, &width, resp.Width)
	require.Equal(t, &height, resp.Height)
	require.NotNil(t, resp.Variants)
}

func TestFromEntities(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()
		resp := mediadto.FromEntities([]*entity.Media{})
		require.NotNil(t, resp)
		require.Empty(t, resp.Items)
		require.Equal(t, 0, resp.Total)
	})

	t.Run("multiple items", func(t *testing.T) {
		t.Parallel()
		media := []*entity.Media{
			{ID: 1, Filename: "a.jpg", Type: entity.MediaTypeImage},
			{ID: 2, Filename: "b.pdf", Type: entity.MediaTypeDocument},
		}
		resp := mediadto.FromEntities(media)
		require.Len(t, resp.Items, 2)
		require.Equal(t, 2, resp.Total)
		require.Equal(t, uint(1), resp.Items[0].ID)
		require.Equal(t, uint(2), resp.Items[1].ID)
	})
}
