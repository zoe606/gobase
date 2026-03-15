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
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/cache"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	content := "Original content"
	status := "draft"
	newTitle := "Updated Title"

	tests := []struct {
		name      string
		userID    uint
		id        uint
		req       articledto.UpdateRequest
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
		wantTitle string
	}{
		{
			name:   "success - title updated",
			userID: 1,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Original Title",
						Slug:      "original-title",
						Content:   &content,
						Status:    &status,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, a *entity.Article) error {
						require.Equal(t, "Updated Title", a.Title)
						return nil
					})
			},
			wantErr:   nil,
			wantTitle: "Updated Title",
		},
		{
			name:   "forbidden - different user",
			userID: 99,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Original Title",
						Slug:      "original-title",
						Content:   &content,
						Status:    &status,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr: article.ErrForbidden,
		},
		{
			name:   "not found",
			userID: 1,
			id:     999,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: article.ErrNotFound,
		},
		{
			name:   "get by id repo error",
			userID: 1,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name:   "update repo error",
			userID: 1,
			id:     1,
			req: articledto.UpdateRequest{
				Title: &newTitle,
			},
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Original Title",
						Slug:      "original-title",
						Content:   &content,
						Status:    &status,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("update failed"))
			},
			wantErr: errors.New("update failed"),
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
			got, err := uc.Update(context.Background(), tt.userID, tt.id, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, article.ErrNotFound) || errors.Is(tt.wantErr, article.ErrForbidden) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.id, got.ID)
			if tt.wantTitle != "" {
				require.Equal(t, tt.wantTitle, got.Title)
			}
		})
	}
}
