package bankstatement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/usecase/bankstatement"
)

func TestUseCase_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBankRepo := NewMockBankRepo(ctrl)
	mockStmtRepo := NewMockBankStatementRepo(ctrl)
	mockLineItemRepo := NewMockLineItemRepo(ctrl)

	uc := bankstatement.New(mockBankRepo, mockStmtRepo, mockLineItemRepo)

	tests := []struct {
		name      string
		id        uint
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			id:   1,
			setupMock: func() {
				mockLineItemRepo.EXPECT().
					DeleteBySource(gomock.Any(), "bank_statement", uint(1)).
					Return(nil)
				mockStmtRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "line items delete error",
			id:   2,
			setupMock: func() {
				mockLineItemRepo.EXPECT().
					DeleteBySource(gomock.Any(), "bank_statement", uint(2)).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "statement delete error",
			id:   3,
			setupMock: func() {
				mockLineItemRepo.EXPECT().
					DeleteBySource(gomock.Any(), "bank_statement", uint(3)).
					Return(nil)
				mockStmtRepo.EXPECT().
					Delete(gomock.Any(), uint(3)).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := uc.Delete(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
