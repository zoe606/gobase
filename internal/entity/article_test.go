package entity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/internal/entity"
)

func TestArticle_TableName(t *testing.T) {
	t.Parallel()
	require.Equal(t, "articles", entity.Article{}.TableName())
}
