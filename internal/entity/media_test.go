package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestMedia_IsImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typ      entity.MediaType
		expected bool
	}{
		{"image type", entity.MediaTypeImage, true},
		{"document type", entity.MediaTypeDocument, false},
		{"video type", entity.MediaTypeVideo, false},
		{"audio type", entity.MediaTypeAudio, false},
		{"other type", entity.MediaTypeOther, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := &entity.Media{Type: tt.typ}
			require.Equal(t, tt.expected, m.IsImage())
		})
	}
}

func TestMedia_GetVariantPath(t *testing.T) {
	t.Parallel()

	t.Run("nil variants", func(t *testing.T) {
		t.Parallel()
		m := &entity.Media{Variants: nil}
		require.Empty(t, m.GetVariantPath("thumb"))
	})

	t.Run("missing variant", func(t *testing.T) {
		t.Parallel()
		m := &entity.Media{Variants: entity.JSONMap{"medium": "path/medium.jpg"}}
		require.Empty(t, m.GetVariantPath("thumb"))
	})

	t.Run("existing variant", func(t *testing.T) {
		t.Parallel()
		m := &entity.Media{Variants: entity.JSONMap{"thumb": "path/thumb.jpg"}}
		require.Equal(t, "path/thumb.jpg", m.GetVariantPath("thumb"))
	})

	t.Run("non-string variant value", func(t *testing.T) {
		t.Parallel()
		m := &entity.Media{Variants: entity.JSONMap{"thumb": 123}}
		require.Empty(t, m.GetVariantPath("thumb"))
	})
}

func TestJSONMap_Value(t *testing.T) {
	t.Parallel()

	t.Run("nil map", func(t *testing.T) {
		t.Parallel()
		var j entity.JSONMap
		val, err := j.Value()
		require.NoError(t, err)
		require.Nil(t, val)
	})

	t.Run("non-nil map", func(t *testing.T) {
		t.Parallel()
		j := entity.JSONMap{"key": "value"}
		val, err := j.Value()
		require.NoError(t, err)
		require.NotNil(t, val)
	})
}

func TestJSONMap_Scan(t *testing.T) {
	t.Parallel()

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()
		var j entity.JSONMap
		err := j.Scan(nil)
		require.NoError(t, err)
		require.Nil(t, j)
	})

	t.Run("valid json bytes", func(t *testing.T) {
		t.Parallel()
		var j entity.JSONMap
		err := j.Scan([]byte(`{"key":"value"}`))
		require.NoError(t, err)
		require.Equal(t, "value", j["key"])
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		var j entity.JSONMap
		err := j.Scan(123)
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()
		var j entity.JSONMap
		err := j.Scan([]byte(`{invalid}`))
		require.Error(t, err)
	})
}
