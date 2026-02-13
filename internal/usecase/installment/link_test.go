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

func TestUseCase_LinkItems(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstallmentRepo := NewMockInstallmentRepo(ctrl)
	mockLineItemRepo := NewMockLineItemRepo(ctrl)

	uc := installment.New(mockInstallmentRepo, mockLineItemRepo)

	tests := []struct {
		name          string
		installmentID uint
		req           installmentdto.LinkItemsRequest
		setupMock     func()
		wantErr       bool
	}{
		{
			name:          "success",
			installmentID: 1,
			req: installmentdto.LinkItemsRequest{
				LineItemIDs: []uint{10, 20, 30},
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Installment{ID: 1, Name: "Home Loan"}, nil)
				mockLineItemRepo.EXPECT().
					UpdateInstallmentID(gomock.Any(), []uint{10, 20, 30}, gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "installment not found",
			installmentID: 99,
			req: installmentdto.LinkItemsRequest{
				LineItemIDs: []uint{10},
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					GetByID(gomock.Any(), uint(99)).
					Return(nil, errors.New("record not found"))
			},
			wantErr: true,
		},
		{
			name:          "update installment id error",
			installmentID: 1,
			req: installmentdto.LinkItemsRequest{
				LineItemIDs: []uint{10, 20},
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Installment{ID: 1, Name: "Home Loan"}, nil)
				mockLineItemRepo.EXPECT().
					UpdateInstallmentID(gomock.Any(), []uint{10, 20}, gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := uc.LinkItems(context.Background(), tt.installmentID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
