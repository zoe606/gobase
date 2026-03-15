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
	"go-boilerplate/pkg/cache"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)

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
		wantTitle string
	}{
		{
			name: "success - all fields mapped",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:        "Test Article",
					Slug:         "test-article",
					Content:      "Some content",
					Excerpt:      "Some excerpt",
					CoverMediaID: 42,
					Status:       "published",
					PublishedAt:  &now,
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, a *entity.Article) error {
						// Verify fields are actually mapped
						require.Equal(t, uint(1), a.UserID)
						require.Equal(t, "Test Article", a.Title)
						require.Equal(t, "test-article", a.Slug)
						require.NotNil(t, a.Content)
						require.Equal(t, "Some content", *a.Content)
						require.NotNil(t, a.Excerpt)
						require.Equal(t, "Some excerpt", *a.Excerpt)
						require.NotNil(t, a.CoverMediaID)
						require.Equal(t, uint(42), *a.CoverMediaID)
						require.NotNil(t, a.Status)
						require.Equal(t, "published", *a.Status)
						require.NotNil(t, a.PublishedAt)
						a.ID = 1
						return nil
					})
			},
			wantErr:   nil,
			wantTitle: "Test Article",
		},
		{
			name: "success - empty status defaults to draft",
			args: args{
				ctx:    context.Background(),
				userID: 2,
				req: articledto.CreateRequest{
					Title:        "Draft Article",
					Slug:         "draft-article",
					Content:      "Draft content",
					Excerpt:      "Draft excerpt",
					CoverMediaID: 1,
				},
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, a *entity.Article) error {
						require.NotNil(t, a.Status)
						require.Equal(t, "draft", *a.Status)
						require.Nil(t, a.PublishedAt)
						a.ID = 2
						return nil
					})
			},
			wantErr:   nil,
			wantTitle: "Draft Article",
		},
		{
			name: "repo error",
			args: args{
				ctx:    context.Background(),
				userID: 1,
				req: articledto.CreateRequest{
					Title:        "Test Article",
					Slug:         "test-article",
					Content:      "Some content",
					Excerpt:      "Some excerpt",
					CoverMediaID: 1,
					Status:       "draft",
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

			uc := article.New(mockArticleRepo, audit.NewNoop(), cache.NewNoop())
			got, err := uc.Create(tt.args.ctx, tt.args.userID, tt.args.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			if tt.wantTitle != "" {
				require.Equal(t, tt.wantTitle, got.Title)
			}
		})
	}
}
