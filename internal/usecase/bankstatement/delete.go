package bankstatement

import (
	"context"
	"fmt"
)

func (uc *UseCase) Delete(ctx context.Context, id uint) error {
	if err := uc.lineItemRepo.DeleteBySource(ctx, "bank_statement", id); err != nil {
		return fmt.Errorf("bankstatement - Delete - lineItemRepo.DeleteBySource: %w", err)
	}

	if err := uc.stmtRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("bankstatement - Delete - stmtRepo.Delete: %w", err)
	}

	return nil
}
