package translation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/usecase/translation"
)

func TestTranslate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx   context.Context
		input translationdto.TranslateRequest
	}

	tests := []struct {
		name      string
		args      args
		setupMock func(repo *MockTranslationRepo, webAPI *MockTranslationWebAPI)
		want      *translationdto.TranslationResponse
		wantErr   error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				input: translationdto.TranslateRequest{
					Source:      "en",
					Destination: "es",
					Original:    "hello",
				},
			},
			setupMock: func(repo *MockTranslationRepo, webAPI *MockTranslationWebAPI) {
				translated := &entity.Translation{
					Source:      "en",
					Destination: "es",
					Original:    "hello",
					Translation: "hola",
				}
				webAPI.EXPECT().
					Translate(gomock.Any()).
					Return(translated, nil)
				repo.EXPECT().
					Store(gomock.Any(), translated).
					Return(nil)
			},
			want: translationdto.NewTranslationResponse(&entity.Translation{
				Source:      "en",
				Destination: "es",
				Original:    "hello",
				Translation: "hola",
			}),
			wantErr: nil,
		},
		{
			name: "webAPI error",
			args: args{
				ctx: context.Background(),
				input: translationdto.TranslateRequest{
					Source:      "en",
					Destination: "es",
					Original:    "hello",
				},
			},
			setupMock: func(repo *MockTranslationRepo, webAPI *MockTranslationWebAPI) {
				webAPI.EXPECT().
					Translate(gomock.Any()).
					Return(nil, errors.New("api error"))
			},
			want:    nil,
			wantErr: errors.New("api error"),
		},
		{
			name: "repo store error",
			args: args{
				ctx: context.Background(),
				input: translationdto.TranslateRequest{
					Source:      "en",
					Destination: "es",
					Original:    "hello",
				},
			},
			setupMock: func(repo *MockTranslationRepo, webAPI *MockTranslationWebAPI) {
				translated := &entity.Translation{
					Source:      "en",
					Destination: "es",
					Original:    "hello",
					Translation: "hola",
				}
				webAPI.EXPECT().
					Translate(gomock.Any()).
					Return(translated, nil)
				repo.EXPECT().
					Store(gomock.Any(), translated).
					Return(errors.New("store error"))
			},
			want:    nil,
			wantErr: errors.New("store error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := NewMockTranslationRepo(ctrl)
			mockWebAPI := NewMockTranslationWebAPI(ctrl)

			tt.setupMock(mockRepo, mockWebAPI)

			uc := translation.New(mockRepo, mockWebAPI)
			got, err := uc.Translate(tt.args.ctx, tt.args.input)

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
