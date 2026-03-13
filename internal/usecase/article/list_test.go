package article_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/pagination"
)

func TestList(t *testing.T) {
	t.Parallel()

	now := time.Now()
	content := "Test content"
	status := "published"

	type args struct {
		ctx context.Context
		req articledto.ListRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
		wantLen   int
		wantTotal int64
	}{
		{
			name: "success with results",
			args: args{
				ctx: context.Background(),
				req: articledto.ListRequest{
					Params: pagination.Params{
						Page:  1,
						Limit: 20,
					},
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.Article{
						{
							ID:        1,
							UserID:    1,
							Title:     "Article 1",
							Slug:      "article-1",
							Content:   &content,
							Status:    &status,
							CreatedAt: now,
							UpdatedAt: now,
						},
						{
							ID:        2,
							UserID:    1,
							Title:     "Article 2",
							Slug:      "article-2",
							Content:   &content,
							Status:    &status,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, int64(2), nil)
			},
			wantErr:   nil,
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name: "success with empty results",
			args: args{
				ctx: context.Background(),
				req: articledto.ListRequest{
					Params: pagination.Params{
						Page:  1,
						Limit: 20,
					},
					Status: "draft",
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.Article{}, int64(0), nil)
			},
			wantErr:   nil,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name: "success with filter by user_id",
			args: args{
				ctx: context.Background(),
				req: articledto.ListRequest{
					Params: pagination.Params{
						Page:  1,
						Limit: 10,
					},
					UserID: 1,
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.Article{
						{
							ID:        1,
							UserID:    1,
							Title:     "Article 1",
							Slug:      "article-1",
							Content:   &content,
							Status:    &status,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, int64(1), nil)
			},
			wantErr:   nil,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name: "normalizes zero page and limit",
			args: args{
				ctx: context.Background(),
				req: articledto.ListRequest{
					Params: pagination.Params{
						Page:  0,
						Limit: 0,
					},
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return([]*entity.Article{}, int64(0), nil)
			},
			wantErr:   nil,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name: "repo error",
			args: args{
				ctx: context.Background(),
				req: articledto.ListRequest{
					Params: pagination.Params{
						Page:  1,
						Limit: 20,
					},
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, int64(0), errors.New("database error"))
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
			got, err := uc.List(tt.args.ctx, tt.args.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Len(t, got.Data, tt.wantLen)
			require.Equal(t, tt.wantTotal, got.Meta.Total)
		})
	}
}
