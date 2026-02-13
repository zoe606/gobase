package permission_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	permissiondto "go-boilerplate/internal/dto/permission"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/permission"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		req permissiondto.CreateRequest
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
				req: permissiondto.CreateRequest{
					Resource: "users",
					Action:   "read",
				},
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					GetByName(gomock.Any(), "users:read").
					Return(nil, repo.ErrNotFound)
				permRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "permission already exists",
			args: args{
				ctx: context.Background(),
				req: permissiondto.CreateRequest{
					Resource: "users",
					Action:   "read",
				},
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					GetByName(gomock.Any(), "users:read").
					Return(&entity.Permission{ID: 1, Name: "users:read"}, nil)
			},
			wantErr: permission.ErrPermissionExists,
		},
		{
			name: "GetByName repo error",
			args: args{
				ctx: context.Background(),
				req: permissiondto.CreateRequest{
					Resource: "users",
					Action:   "write",
				},
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					GetByName(gomock.Any(), "users:write").
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "Create repo error",
			args: args{
				ctx: context.Background(),
				req: permissiondto.CreateRequest{
					Resource: "articles",
					Action:   "delete",
				},
			},
			setupMock: func(permRepo *MockPermissionRepo) {
				permRepo.EXPECT().
					GetByName(gomock.Any(), "articles:delete").
					Return(nil, repo.ErrNotFound)
				permRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("create error"))
			},
			wantErr: errors.New("create error"),
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
			got, err := uc.Create(tt.args.ctx, tt.args.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, permission.ErrPermissionExists) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, "users:read", got.Name)
			require.Equal(t, "users", got.Resource)
			require.Equal(t, "read", got.Action)
		})
	}
}
