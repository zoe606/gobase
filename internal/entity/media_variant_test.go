package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestDefaultImageVariants(t *testing.T) {
	t.Parallel()

	variants := entity.DefaultImageVariants()
	require.Len(t, variants, 3)
	require.Equal(t, "thumb", variants[0].Name)
	require.Equal(t, "medium", variants[1].Name)
	require.Equal(t, "large", variants[2].Name)
}

func TestIsAllowedMimeType(t *testing.T) {
	t.Parallel()

	imageTypes := entity.AllowedImageMimeTypes()

	t.Run("allowed", func(t *testing.T) {
		t.Parallel()
		require.True(t, entity.IsAllowedMimeType("image/jpeg", imageTypes))
	})

	t.Run("not allowed", func(t *testing.T) {
		t.Parallel()
		require.False(t, entity.IsAllowedMimeType("text/html", imageTypes))
	})

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		require.False(t, entity.IsAllowedMimeType("image/jpeg", nil))
	})
}

func TestAllowedMimeTypes(t *testing.T) {
	t.Parallel()

	t.Run("image types", func(t *testing.T) {
		t.Parallel()
		types := entity.AllowedImageMimeTypes()
		require.Contains(t, types, "image/jpeg")
		require.Contains(t, types, "image/png")
	})

	t.Run("document types", func(t *testing.T) {
		t.Parallel()
		types := entity.AllowedDocumentMimeTypes()
		require.Contains(t, types, "application/pdf")
	})

	t.Run("video types", func(t *testing.T) {
		t.Parallel()
		types := entity.AllowedVideoMimeTypes()
		require.Contains(t, types, "video/mp4")
	})

	t.Run("audio types", func(t *testing.T) {
		t.Parallel()
		types := entity.AllowedAudioMimeTypes()
		require.Contains(t, types, "audio/mpeg")
	})

	t.Run("default all types", func(t *testing.T) {
		t.Parallel()
		all := entity.DefaultAllowedMimeTypes()
		require.True(t, len(all) > 10)
		require.Contains(t, all, "image/jpeg")
		require.Contains(t, all, "application/pdf")
		require.Contains(t, all, "video/mp4")
		require.Contains(t, all, "audio/mpeg")
	})
}
