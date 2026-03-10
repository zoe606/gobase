package translation_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/handlers/http/v1/translation"
	"go-boilerplate/pkg/pagination"
)

func setupTestApp(t *testing.T, mockTranslationUC *MockTranslation) *fiber.App {
	t.Helper()

	l := NewMockLogger()
	handler := translation.New(mockTranslationUC, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	return app
}

func TestHandler_Translate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationUC := NewMockTranslation(ctrl)
	app := setupTestApp(t, mockTranslationUC)

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			body: `{"source":"en","destination":"sv","original":"hello world"}`,
			setupMock: func() {
				mockTranslationUC.EXPECT().
					Translate(gomock.Any(), translationdto.TranslateRequest{
						Source:      "en",
						Destination: "sv",
						Original:    "hello world",
					}).
					Return(&translationdto.TranslationResponse{
						ID:          1,
						Source:      "en",
						Destination: "sv",
						Original:    "hello world",
						Translation: "hej varlden",
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "missing required fields",
			body:       `{"source":"en"}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "usecase error",
			body: `{"source":"en","destination":"sv","original":"hello"}`,
			setupMock: func() {
				mockTranslationUC.EXPECT().
					Translate(gomock.Any(), translationdto.TranslateRequest{
						Source:      "en",
						Destination: "sv",
						Original:    "hello",
					}).
					Return(nil, errors.New("translation service error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/translation/do-translate", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_History(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTranslationUC := NewMockTranslation(ctrl)
	app := setupTestApp(t, mockTranslationUC)

	tests := []struct {
		name       string
		query      string
		setupMock  func()
		wantStatus int
	}{
		{
			name:  "success",
			query: "?page=1&limit=20",
			setupMock: func() {
				mockTranslationUC.EXPECT().
					History(gomock.Any(), translationdto.HistoryRequest{
						Params: pagination.Params{
							Page:  1,
							Limit: 20,
						},
					}).
					Return(&translationdto.HistoryResponse{
						Items: []translationdto.TranslationResponse{
							{
								ID:          1,
								Source:      "en",
								Destination: "sv",
								Original:    "hello",
								Translation: "hej",
							},
						},
						Meta: pagination.NewMeta(1, 20, 1),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:  "success with search",
			query: "?page=1&limit=10&search=hello",
			setupMock: func() {
				mockTranslationUC.EXPECT().
					History(gomock.Any(), translationdto.HistoryRequest{
						Params: pagination.Params{
							Page:  1,
							Limit: 10,
						},
						Search: "hello",
					}).
					Return(&translationdto.HistoryResponse{
						Items: []translationdto.TranslationResponse{
							{
								ID:          1,
								Source:      "en",
								Destination: "sv",
								Original:    "hello",
								Translation: "hej",
							},
						},
						Meta: pagination.NewMeta(1, 10, 1),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:  "usecase error",
			query: "?page=1&limit=20",
			setupMock: func() {
				mockTranslationUC.EXPECT().
					History(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/translation/history"+tt.query, http.NoBody)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
