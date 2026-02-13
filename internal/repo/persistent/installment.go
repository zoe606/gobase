package persistent

import (
	"context"
	"errors"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/tx"

	"gorm.io/gorm"
)

type InstallmentRepo struct {
	db *gorm.DB
}

func NewInstallmentRepo(db *gorm.DB) *InstallmentRepo {
	return &InstallmentRepo{db: db}
}

func (r *InstallmentRepo) Create(ctx context.Context, inst *entity.Installment) error {
	db := tx.DBFromContext(ctx, r.db)
	return db.Create(inst).Error
}

func (r *InstallmentRepo) GetByID(ctx context.Context, id uint) (*entity.Installment, error) {
	db := tx.DBFromContext(ctx, r.db)
	var inst entity.Installment
	err := db.First(&inst, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &inst, nil
}

func (r *InstallmentRepo) List(ctx context.Context, req installmentdto.ListRequest) ([]*entity.Installment, int64, error) {
	db := tx.DBFromContext(ctx, r.db)
	query := db.Model(&entity.Installment{})

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

	var installments []*entity.Installment
	query = req.Apply(query, []string{"id", "created_at", "name", "total_amount"})
	if err := query.Find(&installments).Error; err != nil {
		return nil, 0, err
	}

	return installments, total, nil
}

func (r *InstallmentRepo) Update(ctx context.Context, inst *entity.Installment) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Save(inst)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}

func (r *InstallmentRepo) Delete(ctx context.Context, id uint) error {
	db := tx.DBFromContext(ctx, r.db)
	result := db.Delete(&entity.Installment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repo.ErrNotFound
	}
	return nil
}
