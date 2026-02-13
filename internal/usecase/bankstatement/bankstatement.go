package bankstatement

import (
	"go-boilerplate/internal/repo"
)

type UseCase struct {
	bankRepo     repo.BankRepo
	stmtRepo     repo.BankStatementRepo
	lineItemRepo repo.LineItemRepo
	parsers      map[string]Parser
}

func New(bankRepo repo.BankRepo, stmtRepo repo.BankStatementRepo, lineItemRepo repo.LineItemRepo) *UseCase {
	uc := &UseCase{
		bankRepo:     bankRepo,
		stmtRepo:     stmtRepo,
		lineItemRepo: lineItemRepo,
		parsers:      make(map[string]Parser),
	}
	uc.parsers["BCA"] = &BCAParser{}
	uc.parsers["BRI"] = &BRIParser{}
	return uc
}
