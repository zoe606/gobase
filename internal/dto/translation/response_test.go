package translationdto_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/pagination"
)

func TestNewTranslationResponse(t *testing.T) {
	t.Parallel()

	tr := &entity.Translation{
		ID:          1,
		Source:      "en",
		Destination: "es",
		Original:    "hello",
		Translation: "hola",
	}

	resp := translationdto.NewTranslationResponse(tr)
	require.Equal(t, uint(1), resp.ID)
	require.Equal(t, "en", resp.Source)
	require.Equal(t, "es", resp.Destination)
	require.Equal(t, "hello", resp.Original)
	require.Equal(t, "hola", resp.Translation)
}

func TestNewHistoryResponse(t *testing.T) {
	t.Parallel()

	t.Run("with items", func(t *testing.T) {
		t.Parallel()
		translations := []entity.Translation{
			{ID: 1, Source: "en", Destination: "es", Original: "hello", Translation: "hola"},
			{ID: 2, Source: "en", Destination: "fr", Original: "hello", Translation: "bonjour"},
		}
		params := pagination.Params{Page: 1, Limit: 10}

		resp := translationdto.NewHistoryResponse(translations, params, 2)
		require.Len(t, resp.Items, 2)
		require.Equal(t, uint(1), resp.Items[0].ID)
		require.Equal(t, uint(2), resp.Items[1].ID)
		require.NotNil(t, resp.Meta)
		require.Equal(t, int64(2), resp.Meta.Total)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		params := pagination.Params{Page: 1, Limit: 10}

		resp := translationdto.NewHistoryResponse(nil, params, 0)
		require.Empty(t, resp.Items)
		require.Equal(t, int64(0), resp.Meta.Total)
	})

	t.Run("normalizes params", func(t *testing.T) {
		t.Parallel()
		params := pagination.Params{Page: 0, Limit: 0}

		resp := translationdto.NewHistoryResponse(nil, params, 50)
		require.Equal(t, pagination.DefaultPage, resp.Meta.Page)
		require.Equal(t, pagination.DefaultLimit, resp.Meta.Limit)
	})
}
