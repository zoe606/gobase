package response_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/response"
)

func TestNewMeta(t *testing.T) {
	t.Parallel()

	t.Run("normal pagination", func(t *testing.T) {
		t.Parallel()
		meta := response.NewMeta(2, 10, 55)
		require.Equal(t, 2, meta.Page)
		require.Equal(t, 10, meta.Limit)
		require.Equal(t, int64(55), meta.Total)
		require.Equal(t, 6, meta.TotalPages)
	})

	t.Run("exact division", func(t *testing.T) {
		t.Parallel()
		meta := response.NewMeta(1, 10, 30)
		require.Equal(t, 3, meta.TotalPages)
	})

	t.Run("zero limit", func(t *testing.T) {
		t.Parallel()
		meta := response.NewMeta(1, 0, 100)
		require.Equal(t, 0, meta.TotalPages)
	})

	t.Run("zero total", func(t *testing.T) {
		t.Parallel()
		meta := response.NewMeta(1, 10, 0)
		require.Equal(t, 0, meta.TotalPages)
	})
}
