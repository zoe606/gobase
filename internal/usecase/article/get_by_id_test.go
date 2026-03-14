package article_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/cache"
)

func TestGetByID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	content := "Test content"
	status := "published"

	tests := []struct {
		name      string
		id        uint
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
	}{
		{
			name: "success",
			id:   1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Test Article",
						Slug:      "test-article",
						Content:   &content,
						Status:    &status,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   999,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantErr: article.ErrNotFound,
		},
		{
			name: "repo error",
			id:   1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
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
			got, err := uc.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, article.ErrNotFound) {
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, tt.id, got.ID)
		})
	}
}
