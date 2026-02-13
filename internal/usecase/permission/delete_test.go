package permission_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/permission"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		id  uint
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(permRepo *MockPermissionRepo)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					List(gomock.Any()).
					Return([]*entity.Permission{
						{ID: 1, Name: "users:read"},
						{ID: 2, Name: "users:write"},
					}, nil)
				permRepo.EXPECT().
					IsAssignedToAnyRole(gomock.Any(), uint(1)).
					Return(false, nil)
				permRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "permission not found",
			args: args{
				ctx: context.Background(),
				id:  99,
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					List(gomock.Any()).
					Return([]*entity.Permission{
						{ID: 1, Name: "users:read"},
					}, nil)
			},
			wantErr: permission.ErrPermissionNotFound,
		},
		{
			name: "permission in use",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					List(gomock.Any()).
					Return([]*entity.Permission{
						{ID: 1, Name: "users:read"},
					}, nil)
				permRepo.EXPECT().
					IsAssignedToAnyRole(gomock.Any(), uint(1)).
					Return(true, nil)
			},
			wantErr: permission.ErrPermissionInUse,
		},
		{
			name: "List repo error",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "IsAssignedToAnyRole repo error",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					List(gomock.Any()).
					Return([]*entity.Permission{
						{ID: 1, Name: "users:read"},
					}, nil)
				permRepo.EXPECT().
					IsAssignedToAnyRole(gomock.Any(), uint(1)).
					Return(false, errors.New("role check error"))
			},
			wantErr: errors.New("role check error"),
		},
		{
			name: "Delete repo error",
			args: args{
				ctx: context.Background(),
				id:  1,
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					List(gomock.Any()).
					Return([]*entity.Permission{
						{ID: 1, Name: "users:read"},
					}, nil)
				permRepo.EXPECT().
					IsAssignedToAnyRole(gomock.Any(), uint(1)).
					Return(false, nil)
				permRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(errors.New("delete error"))
			},
			wantErr: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPermRepo := NewMockPermissionRepo(ctrl)
			tt.setupMock(mockPermRepo)

			uc := permission.New(mockPermRepo)
			err := uc.Delete(tt.args.ctx, tt.args.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, permission.ErrPermissionNotFound) ||
					errors.Is(tt.wantErr, permission.ErrPermissionInUse) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
