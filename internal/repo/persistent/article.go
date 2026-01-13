package persistent

import (
	"context"
	"errors"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"

	"gorm.io/gorm"
)

// ArticlePostgres implements repo.ArticleRepo using PostgreSQL.
type ArticlePostgres struct {
	db *gorm.DB
}

// NewArticlePostgres creates a new Article repository.
func NewArticlePostgres(db *gorm.DB) *ArticlePostgres {
	return &ArticlePostgres{db: db}
}

// Create creates a new article.
func (r *ArticlePostgres) Create(ctx context.Context, article *entity.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

// GetByID retrieves a article by ID.
func (r *ArticlePostgres) GetByID(ctx context.Context, id uint) (*entity.Article, error) {
	var article entity.Article
	err := r.db.WithContext(ctx).First(&article, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &article, nil
}

// List retrieves a paginated list of articles.
func (r *ArticlePostgres) List(ctx context.Context, limit, offset int) ([]*entity.Article, int64, error) {
	var articles []*entity.Article
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&entity.Article{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("id DESC").
		Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// Update updates a article.
func (r *ArticlePostgres) Update(ctx context.Context, article *entity.Article) error {
	result := r.db.WithContext(ctx).Save(article)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// Delete deletes a article by ID.
func (r *ArticlePostgres) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entity.Article{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}
