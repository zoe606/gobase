package bankstatement_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	bankstatementdto "go-boilerplate/internal/dto/bankstatement"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/bankstatement"
	"go-boilerplate/pkg/pagination"
)

//go:generate mockgen -source=../../repo/contracts.go -destination=./mocks_test.go -package=bankstatement_test

func TestUseCase_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBankRepo := NewMockBankRepo(ctrl)
	mockStmtRepo := NewMockBankStatementRepo(ctrl)
	mockLineItemRepo := NewMockLineItemRepo(ctrl)

	uc := bankstatement.New(mockBankRepo, mockStmtRepo, mockLineItemRepo)

	tests := []struct {
		name      string
		req       bankstatementdto.ListRequest
		setupMock func()
		wantErr   bool
		wantTotal int64
		wantLen   int
	}{
		{
			name: "success with results",
			req: bankstatementdto.ListRequest{
				Params: pagination.Params{Page: 1, Limit: 10},
			},
			setupMock: func() {
				mockStmtRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.BankStatement{
						{ID: 1, Status: "completed", Bank: &entity.Bank{Name: "BCA", Code: "BCA"}},
						{ID: 2, Status: "pending", Bank: &entity.Bank{Name: "BRI", Code: "BRI"}},
					}, int64(2), nil)
			},
			wantErr:   false,
			wantTotal: 2,
			wantLen:   2,
		},
		{
			name: "success empty result",
			req: bankstatementdto.ListRequest{
				Params: pagination.Params{Page: 1, Limit: 10},
			},
			setupMock: func() {
				mockStmtRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.BankStatement{}, int64(0), nil)
			},
			wantErr:   false,
			wantTotal: 0,
			wantLen:   0,
		},
		{
			name: "repo error",
			req: bankstatementdto.ListRequest{
				Params: pagination.Params{Page: 1, Limit: 10},
			},
			setupMock: func() {
				mockStmtRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "pagination defaults applied",
			req:  bankstatementdto.ListRequest{},
			setupMock: func() {
				mockStmtRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.BankStatement{}, int64(0), nil)
			},
			wantErr:   false,
			wantTotal: 0,
			wantLen:   0,
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
