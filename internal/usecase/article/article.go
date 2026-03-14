// Package article provides article management use cases.
package article

//go:generate mockgen -source=../../repo/contracts.go -destination=mocks_repo_test.go -package=article_test

import (
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/cache"
)

// UseCase implements article business logic.
type UseCase struct {
	articleRepo repo.ArticleRepo
	auditLogger audit.Logger
	cache       cache.Cache
	cacheKeys   *cache.KeyBuilder
}

// New creates a new article use case.
func New(articleRepo repo.ArticleRepo, auditLogger audit.Logger, articleCache cache.Cache) *UseCase {
	return &UseCase{
		articleRepo: articleRepo,
		auditLogger: auditLogger,
		cache:       articleCache,
		cacheKeys:   cache.NewKeyBuilder("article"),
	}
}
