package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-boilerplate/pkg/cache"
)

func TestCacheKeyBuilder(t *testing.T) {
	t.Parallel()

	kb := cache.NewKeyBuilder("article")

	assert.Equal(t, "article:1", kb.ID(1))
	assert.Equal(t, "article:42", kb.ID(42))
	assert.Equal(t, "article:list:", kb.ListPrefix())
	assert.Equal(t, "article:list:page=1&size=10", kb.List("page=1&size=10"))
	assert.Equal(t, "article:", kb.Prefix())
}

func TestCacheKeyBuilder_DifferentEntities(t *testing.T) {
	t.Parallel()

	userKB := cache.NewKeyBuilder("user")
	assert.Equal(t, "user:1", userKB.ID(1))
	assert.Equal(t, "user:list:", userKB.ListPrefix())
}
