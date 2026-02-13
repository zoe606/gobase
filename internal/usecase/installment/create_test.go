package installment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	installmentdto "go-boilerplate/internal/dto/installment"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/installment"
)

//go:generate mockgen -source=../../repo/contracts.go -destination=./mocks_test.go -package=installment_test

func TestUseCase_Create(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstallmentRepo := NewMockInstallmentRepo(ctrl)
	mockLineItemRepo := NewMockLineItemRepo(ctrl)

	uc := installment.New(mockInstallmentRepo, mockLineItemRepo)

	tests := []struct {
		name      string
		req       installmentdto.CreateRequest
		setupMock func()
		wantErr   bool
		wantName  string
	}{
		{
			name: "success",
			req: installmentdto.CreateRequest{
				Name:          "Home Loan",
				TotalAmount:   500000000,
				MonthlyAmount: 5000000,
				TotalTerms:    120,
				UserID:        1,
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, inst *entity.Installment) error {
						inst.ID = 1
						return nil
					})
			},
			wantErr:  false,
			wantName: "Home Loan",
		},
		{
			name: "success with optional fields",
			req: installmentdto.CreateRequest{
				Name:           "Car Loan",
				Merchant:       "BMW Dealer",
				TotalAmount:    300000000,
				MonthlyAmount:  4000000,
				TotalTerms:     60,
				CompletedTerms: 5,
				StartDate:      "2025-01-01",
				EndDate:        "2030-01-01",
				Notes:          "Monthly auto-debit",
				UserID:         1,
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, inst *entity.Installment) error {
						inst.ID = 2
						return nil
					})
			},
			wantErr:  false,
			wantName: "Car Loan",
		},
		{
			name: "repo error",
			req: installmentdto.CreateRequest{
				Name:          "Failed Loan",
				TotalAmount:   100000000,
				MonthlyAmount: 1000000,
				TotalTerms:    12,
				UserID:        1,
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := uc.Create(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, "active", result.Status)
		})
	}
}
