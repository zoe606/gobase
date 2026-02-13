package installment

import (
	"context"
	"fmt"
)

func (uc *UseCase) Delete(ctx context.Context, id uint) error {
	if err := uc.installmentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("installment - Delete - installmentRepo.Delete: %w", err)
	}

	return nil
}
