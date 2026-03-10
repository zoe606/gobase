package articledto_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/pagination"
)

func TestNewResponse(t *testing.T) {
	t.Parallel()

	t.Run("nil article", func(t *testing.T) {
		t.Parallel()
		require.Nil(t, articledto.NewResponse(nil))
	})

	t.Run("valid article", func(t *testing.T) {
		t.Parallel()
		content := "test content"
		status := "published"
		now := time.Now()

		article := &entity.Article{
			ID:      1,
			UserID:  2,
			Title:   "Test",
			Slug:    "test",
			Content: &content,
			Status:  &status,
		}

		resp := articledto.NewResponse(article)
		require.NotNil(t, resp)
		require.Equal(t, uint(1), resp.ID)
		require.Equal(t, uint(2), resp.UserID)
		require.Equal(t, "Test", resp.Title)
		require.Equal(t, "test", resp.Slug)
		require.Equal(t, &content, resp.Content)
		require.Equal(t, &status, resp.Status)
		_ = now
	})
}

func TestNewListResponse(t *testing.T) {
	t.Parallel()

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()
		resp := articledto.NewListResponse(nil, 0, pagination.Params{Page: 1, Limit: 10})
		require.NotNil(t, resp)
		require.Empty(t, resp.Data)
		require.Equal(t, int64(0), resp.Meta.Total)
	})

	t.Run("with articles", func(t *testing.T) {
		t.Parallel()
		articles := []*entity.Article{
			{ID: 1, Title: "First"},
			{ID: 2, Title: "Second"},
		}
		params := pagination.Params{Page: 1, Limit: 10}
		resp := articledto.NewListResponse(articles, 2, params)
		require.Len(t, resp.Data, 2)
		require.Equal(t, uint(1), resp.Data[0].ID)
		require.Equal(t, int64(2), resp.Meta.Total)
	})
}
