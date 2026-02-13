package bankstatement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/bankstatement"
)

func TestUseCase_GetByID(t *testing.T) {
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
		wantItems int
	}{
		{
			name: "success with items",
			id:   1,
			setupMock: func() {
				mockStmtRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.BankStatement{
						ID:     1,
						Status: "completed",
						Bank:   &entity.Bank{Name: "BCA", Code: "BCA"},
					}, nil)
				mockLineItemRepo.EXPECT().
					GetBySource(gomock.Any(), "bank_statement", uint(1)).
					Return([]*entity.LineItem{
						{ID: 1, Date: "2025-01-01", Description: "Transaction 1", Debit: 100000},
						{ID: 2, Date: "2025-01-02", Description: "Transaction 2", Credit: 200000},
					}, nil)
			},
			wantErr:   false,
			wantItems: 2,
		},
		{
			name: "success with no items",
			id:   2,
			setupMock: func() {
				mockStmtRepo.EXPECT().
					GetByID(gomock.Any(), uint(2)).
					Return(&entity.BankStatement{
						ID:     2,
						Status: "pending",
						Bank:   &entity.Bank{Name: "BRI", Code: "BRI"},
					}, nil)
				mockLineItemRepo.EXPECT().
					GetBySource(gomock.Any(), "bank_statement", uint(2)).
					Return([]*entity.LineItem{}, nil)
			},
			wantErr:   false,
			wantItems: 0,
		},
		{
			name: "statement not found",
			id:   99,
			setupMock: func() {
				mockStmtRepo.EXPECT().
					GetByID(gomock.Any(), uint(99)).
					Return(nil, errors.New("record not found"))
			},
			wantErr: true,
		},
		{
			name: "line items repo error",
			id:   1,
			setupMock: func() {
				mockStmtRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.BankStatement{
						ID:     1,
						Status: "completed",
						Bank:   &entity.Bank{Name: "BCA", Code: "BCA"},
					}, nil)
				mockLineItemRepo.EXPECT().
					GetBySource(gomock.Any(), "bank_statement", uint(1)).
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := uc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.id, result.ID)
			assert.Len(t, result.Items, tt.wantItems)
		})
	}
}
