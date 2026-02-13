package persistent

import (
	"context"
	"errors"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"

	"gorm.io/gorm"
)

type BankRepo struct {
	db *gorm.DB
}

func NewBankRepo(db *gorm.DB) *BankRepo {
	return &BankRepo{db: db}
}

func (r *BankRepo) List(ctx context.Context) ([]*entity.Bank, error) {
	db := tx.DBFromContext(ctx, r.db)
	var banks []*entity.Bank
	if err := db.Find(&banks).Error; err != nil {
		return nil, err
	}
	return banks, nil
}

func (r *BankRepo) GetByID(ctx context.Context, id uint) (*entity.Bank, error) {
	db := tx.DBFromContext(ctx, r.db)
	var bank entity.Bank
	err := db.First(&bank, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &bank, nil
}

func (r *BankRepo) GetByCode(ctx context.Context, code string) (*entity.Bank, error) {
	db := tx.DBFromContext(ctx, r.db)
	var bank entity.Bank
	err := db.Where("code = ?", code).First(&bank).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &bank, nil
}
