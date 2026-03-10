package article_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/usecase/article"
)

func TestDelete(t *testing.T) {
	t.Parallel()

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
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			id:   999,
			setupMock: func(articleRepo *MockArticleRepo) {
				articleRepo.EXPECT().
					Delete(gomock.Any(), uint(999)).
					Return(repo.ErrNotFound)
			},
			wantErr: article.ErrNotFound,
		},
		{
			name: "repo error",
			id:   1,
			setupMock: func(articleRepo *MockArticleRepo) {
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

			uc := article.New(mockArticleRepo)
			err := uc.Delete(context.Background(), tt.id)

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
		})
	}
}
