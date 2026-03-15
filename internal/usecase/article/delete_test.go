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

func TestDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	status := "draft"

	tests := []struct {
		name      string
		userID    uint
		id        uint
		setupMock func(articleRepo *MockArticleRepo)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			id:     1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Test Article",
						Status:    &status,
						CreatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "forbidden - different user",
			userID: 99,
			id:     1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Test Article",
						Status:    &status,
						CreatedAt: now,
					}, nil)
			},
			wantErr: article.ErrForbidden,
		},
		{
			name:   "not found",
			userID: 1,
			id:     999,
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
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name:   "delete repo error",
			userID: 1,
			id:     1,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Article{
						ID:        1,
						UserID:    1,
						Title:     "Test Article",
						Status:    &status,
						CreatedAt: now,
					}, nil)
				articleRepo.EXPECT().
					Delete(gomock.Any(), uint(1)).
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
			err := uc.Delete(context.Background(), tt.userID, tt.id)

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
		})
	}
}
