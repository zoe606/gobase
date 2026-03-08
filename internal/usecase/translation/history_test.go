package translation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/translation"
	"go-boilerplate/pkg/pagination"
)

func TestHistory(t *testing.T) {
	t.Parallel()

	defaultParams := pagination.NewParams()

	tests := []struct {
		name      string
		request   translationdto.HistoryRequest
		setupMock func(repo *MockTranslationRepo)
		want      *translationdto.HistoryResponse
		wantErr   error
	}{
		{
			name:    "success with results",
			request: translationdto.HistoryRequest{Params: defaultParams},
			setupMock: func(repo *MockTranslationRepo) {
				translations := []entity.Translation{
					{
						Source:      "en",
						Destination: "es",
						Original:    "hello",
						Translation: "hola",
					},
					{
						Source:      "en",
						Destination: "fr",
						Original:    "goodbye",
						Translation: "au revoir",
					},
				}
				repo.EXPECT().
					GetHistory(gomock.Any(), gomock.Any()).
					Return(translations, int64(2), nil)
			},
			want: translationdto.NewHistoryResponse([]entity.Translation{
				{
					Source:      "en",
					Destination: "es",
					Original:    "hello",
					Translation: "hola",
				},
				{
					Source:      "en",
					Destination: "fr",
					Original:    "goodbye",
					Translation: "au revoir",
				},
			}, defaultParams, 2),
			wantErr: nil,
		},
		{
			name:    "success with empty result",
			request: translationdto.HistoryRequest{Params: defaultParams},
			setupMock: func(repo *MockTranslationRepo) {
				repo.EXPECT().
					GetHistory(gomock.Any(), gomock.Any()).
					Return(nil, int64(0), nil)
			},
			want:    translationdto.NewHistoryResponse(nil, defaultParams, 0),
			wantErr: nil,
		},
		{
			name:    "repo error",
			request: translationdto.HistoryRequest{Params: defaultParams},
			setupMock: func(repo *MockTranslationRepo) {
				repo.EXPECT().
					GetHistory(gomock.Any(), gomock.Any()).
					Return(nil, int64(0), errors.New("database error"))
			},
			want:    nil,
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockTranslationRepo(ctrl)
			mockWebAPI := NewMockTranslationWebAPI(ctrl)

			tt.setupMock(mockRepo)

			uc := translation.New(mockRepo, mockWebAPI)
			got, err := uc.History(context.Background(), tt.request)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
