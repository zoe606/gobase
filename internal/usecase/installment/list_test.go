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
	"go-boilerplate/pkg/pagination"
)

func TestUseCase_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInstallmentRepo := NewMockInstallmentRepo(ctrl)
	mockLineItemRepo := NewMockLineItemRepo(ctrl)

	uc := installment.New(mockInstallmentRepo, mockLineItemRepo)

	tests := []struct {
		name      string
		req       installmentdto.ListRequest
		setupMock func()
		wantErr   bool
		wantTotal int64
		wantLen   int
	}{
		{
			name: "success with results",
			req: installmentdto.ListRequest{
				Params: pagination.Params{Page: 1, Limit: 10},
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.Installment{
						{ID: 1, Name: "Home Loan", Status: "active"},
						{ID: 2, Name: "Car Loan", Status: "active"},
					}, int64(2), nil)
			},
			wantErr:   false,
			wantTotal: 2,
			wantLen:   2,
		},
		{
			name: "success empty result",
			req: installmentdto.ListRequest{
				Params: pagination.Params{Page: 1, Limit: 10},
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.Installment{}, int64(0), nil)
			},
			wantErr:   false,
			wantTotal: 0,
			wantLen:   0,
		},
		{
			name: "repo error",
			req: installmentdto.ListRequest{
				Params: pagination.Params{Page: 1, Limit: 10},
			},
			setupMock: func() {
				mockInstallmentRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := uc.List(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Data, tt.wantLen)
			assert.Equal(t, tt.wantTotal, result.Meta.Total)
		})
	}
}
