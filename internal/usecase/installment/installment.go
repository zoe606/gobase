package installment

import (
	"go-boilerplate/internal/repo"
)

type UseCase struct {
	installmentRepo repo.InstallmentRepo
	lineItemRepo    repo.LineItemRepo
}

func New(installmentRepo repo.InstallmentRepo, lineItemRepo repo.LineItemRepo) *UseCase {
	return &UseCase{
		installmentRepo: installmentRepo,
		lineItemRepo:    lineItemRepo,
	}
}
