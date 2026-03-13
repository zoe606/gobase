package article_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/audit"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		userID uint
		req    articledto.CreateRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:   "Test Article",
					Slug:    "test-article",
					Content: "Some content",
					Excerpt: "Some excerpt",
					Status:  "draft",
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:   "Test Article",
					Slug:    "test-article",
					Content: "Some content",
					Excerpt: "Some excerpt",
					Status:  "draft",
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockArticleRepo := NewMockArticleRepo(ctrl)

			tt.setupMock(mockArticleRepo)

			uc := article.New(mockArticleRepo, audit.NewNoop())
			got, err := uc.Create(tt.args.ctx, tt.args.userID, tt.args.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
