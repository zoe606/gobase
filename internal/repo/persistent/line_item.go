package persistent

import (
	"context"
	"errors"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"

	"gorm.io/gorm"
)

type LineItemRepo struct {
	db *gorm.DB
}

func NewLineItemRepo(db *gorm.DB) *LineItemRepo {
	return &LineItemRepo{db: db}
}

func (r *LineItemRepo) BulkCreate(ctx context.Context, items []*entity.LineItem) error {
	if len(items) == 0 {
		return nil
	}
	db := tx.DBFromContext(ctx, r.db)
	return db.Create(&items).Error
}

func (r *LineItemRepo) GetBySource(ctx context.Context, sourceType string, sourceID uint) ([]*entity.LineItem, error) {
	db := tx.DBFromContext(ctx, r.db)
	var items []*entity.LineItem
	err := db.Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Order("date ASC, id ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *LineItemRepo) Update(ctx context.Context, item *entity.LineItem) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Save(item)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *LineItemRepo) GetByID(ctx context.Context, id uint) (*entity.LineItem, error) {
	db := tx.DBFromContext(ctx, r.db)
	var item entity.LineItem
	err := db.First(&item, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *LineItemRepo) DeleteBySource(ctx context.Context, sourceType string, sourceID uint) error {
	db := tx.DBFromContext(ctx, r.db)
	return db.Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Delete(&entity.LineItem{}).Error
}

func (r *LineItemRepo) UpdateInstallmentID(ctx context.Context, itemIDs []uint, installmentID *uint) error {
	db := tx.DBFromContext(ctx, r.db)
	updates := map[string]interface{}{
		"installment_id": installmentID,
		"is_installment": installmentID != nil,
	}
	return db.Model(&entity.LineItem{}).Where("id IN ?", itemIDs).Updates(updates).Error
}
