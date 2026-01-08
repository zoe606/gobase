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
)

func TestHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(repo *MockTranslationRepo)
		want      *translationdto.HistoryResponse
		wantErr   error
	}{
		{
			name: "success with results",
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
					GetHistory(gomock.Any()).
					Return(translations, nil)
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
			}),
			wantErr: nil,
		},
		{
			name: "success with empty result",
			setupMock: func(repo *MockTranslationRepo) {
				repo.EXPECT().
					GetHistory(gomock.Any()).
					Return(nil, nil)
			},
			want:    translationdto.NewHistoryResponse(nil),
			wantErr: nil,
		},
		{
			name: "repo error",
			setupMock: func(repo *MockTranslationRepo) {
				repo.EXPECT().
					GetHistory(gomock.Any()).
					Return(nil, errors.New("database error"))
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
			got, err := uc.History(context.Background())

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
