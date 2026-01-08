package translation_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
)

func TestTranslateOutput_ToResponse(t *testing.T) {
	t.Parallel()

	output := &translation.TranslateOutput{
		Translation: entity.Translation{
			ID:          1,
			Source:      "en",
			Destination: "es",
			Original:    "Hello",
			Translation: "Hola",
		},
	}

	resp := output.ToResponse()

	require.Equal(t, uint(1), resp.ID)
	require.Equal(t, "en", resp.Source)
	require.Equal(t, "es", resp.Destination)
	require.Equal(t, "Hello", resp.Original)
	require.Equal(t, "Hola", resp.Translation)
}

func TestHistoryOutput_ToResponse(t *testing.T) {
	t.Parallel()

	output := &translation.HistoryOutput{
		History: []entity.Translation{
			{
				ID:          1,
				Source:      "en",
				Destination: "es",
				Original:    "Hello",
				Translation: "Hola",
			},
			{
				ID:          2,
				Source:      "en",
				Destination: "fr",
				Original:    "World",
				Translation: "Monde",
			},
		},
	}

	resp := output.ToResponse()

	require.Len(t, resp.History, 2)

	require.Equal(t, uint(1), resp.History[0].ID)
	require.Equal(t, "en", resp.History[0].Source)
	require.Equal(t, "es", resp.History[0].Destination)
	require.Equal(t, "Hello", resp.History[0].Original)
	require.Equal(t, "Hola", resp.History[0].Translation)

	require.Equal(t, uint(2), resp.History[1].ID)
	require.Equal(t, "en", resp.History[1].Source)
	require.Equal(t, "fr", resp.History[1].Destination)
	require.Equal(t, "World", resp.History[1].Original)
	require.Equal(t, "Monde", resp.History[1].Translation)
}

func TestHistoryOutput_ToResponse_Empty(t *testing.T) {
	t.Parallel()

	output := &translation.HistoryOutput{
		History: []entity.Translation{},
	}

	resp := output.ToResponse()

	require.NotNil(t, resp.History)
	require.Len(t, resp.History, 0)
}
