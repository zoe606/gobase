package persistent

import (
	"context"
	"errors"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"

	"gorm.io/gorm"
)

type BankStatementRepo struct {
	db *gorm.DB
}

func NewBankStatementRepo(db *gorm.DB) *BankStatementRepo {
	return &BankStatementRepo{db: db}
}

func (r *BankStatementRepo) Create(ctx context.Context, stmt *entity.BankStatement) error {
	db := tx.DBFromContext(ctx, r.db)
	return db.Create(stmt).Error
}

func (r *BankStatementRepo) GetByID(ctx context.Context, id uint) (*entity.BankStatement, error) {
	db := tx.DBFromContext(ctx, r.db)
	var stmt entity.BankStatement
	err := db.Preload("Bank").First(&stmt, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &stmt, nil
}

func (r *BankStatementRepo) List(ctx context.Context, req bankstatementdto.ListRequest) ([]*entity.BankStatement, int64, error) {
	db := tx.DBFromContext(ctx, r.db)
	query := db.Model(&entity.BankStatement{}).Preload("Bank")

	if req.UserID > 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var stmts []*entity.BankStatement
	query = req.Apply(query, []string{"id", "created_at", "total_debit", "total_credit"})
	if err := query.Find(&stmts).Error; err != nil {
		return nil, 0, err
	}

	return stmts, total, nil
}

func (r *BankStatementRepo) Update(ctx context.Context, stmt *entity.BankStatement) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Save(stmt)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *BankStatementRepo) Delete(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Delete(&entity.BankStatement{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}
