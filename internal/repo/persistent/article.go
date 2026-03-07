package persistent

import (
	"context"
	"errors"

	"go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"

	"gorm.io/gorm"
)

// ArticleRepo implements repo.ArticleRepo using PostgreSQL.
type ArticleRepo struct {
	db *gorm.DB
}

// NewArticleRepo creates a new Article repository.
func NewArticleRepo(db *gorm.DB) *ArticleRepo {
	return &ArticleRepo{db: db}
}

// Create creates a new article.
func (r *ArticleRepo) Create(ctx context.Context, article *entity.Article) error {
	db := tx.DBFromContext(ctx, r.db)
	return db.Create(article).Error
}

// GetByID retrieves an article by ID.
func (r *ArticleRepo) GetByID(ctx context.Context, id uint) (*entity.Article, error) {
	db := tx.DBFromContext(ctx, r.db)
	var article entity.Article
	err := db.First(&article, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &article, nil
}

// List retrieves a paginated list of articles with filters.
func (r *ArticleRepo) List(ctx context.Context, req articledto.ListRequest) ([]*entity.Article, int64, error) {
	db := tx.DBFromContext(ctx, r.db)
	query := db.Model(&entity.Article{})

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ?", searchPattern, searchPattern)
	}

	// Count total (with filters applied)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	var articles []*entity.Article
	query = req.Apply(query, []string{"id", "created_at", "published_at", "title"})
	if err := query.Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

// Update updates an article.
func (r *ArticleRepo) Update(ctx context.Context, article *entity.Article) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Save(article)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

// Delete deletes an article by ID.
func (r *ArticleRepo) Delete(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Delete(&entity.Article{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}
