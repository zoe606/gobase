package article_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	articledto "go-boilerplate/internal/dto/article"
	"go-boilerplate/internal/handlers/http/v1/article"
	articleuc "go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/pagination"
)

func setupTestApp(t *testing.T, mockArticleUC *MockArticle) (app *fiber.App, token string) {
	t.Helper()

	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := NewMockLogger()
	handler := article.New(mockArticleUC, jwtService, l)

	app = fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	// Generate a valid test token
	var err error
	token, _, err = jwtService.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	return app, token
}

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleUC := NewMockArticle(ctrl)
	app, token := setupTestApp(t, mockArticleUC)

	now := time.Now().UTC().Truncate(time.Second)
	validBody := `{
		"title": "Test Article",
		"slug": "test-article",
		"content": "Article content here",
		"excerpt": "Short excerpt",
		"cover_media_id": 1,
		"status": "draft",
		"published_at": "` + now.Format(time.RFC3339) + `",
		"view_count": 1
	}`

	tests := []struct {
		name       string
		body       string
		addAuth    bool
		setupMock  func()
		wantStatus int
	}{
		{
			name:    "success",
			body:    validBody,
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&articledto.Response{
						ID:     1,
						UserID: 1,
						Title:  "Test Article",
					}, nil)
			},
			wantStatus: fiber.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			addAuth:    true,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "validation error - missing required fields",
			body:       `{"title":"only title"}`,
			addAuth:    true,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:    "usecase error",
			body:    validBody,
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
		{
			name:       "unauthorized - no token",
			body:       validBody,
			addAuth:    false,
			setupMock:  func() {},
			wantStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/articles/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			if tt.addAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_GetByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleUC := NewMockArticle(ctrl)
	app, _ := setupTestApp(t, mockArticleUC)

	tests := []struct {
		name       string
		id         string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			setupMock: func() {
				mockArticleUC.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&articledto.Response{
						ID:     1,
						UserID: 1,
						Title:  "Test Article",
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid id",
			id:         "invalid",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "999",
			setupMock: func() {
				mockArticleUC.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, articleuc.ErrNotFound)
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name: "internal error",
			id:   "1",
			setupMock: func() {
				mockArticleUC.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/articles/"+tt.id, http.NoBody)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleUC := NewMockArticle(ctrl)
	app, _ := setupTestApp(t, mockArticleUC)

	tests := []struct {
		name       string
		query      string
		setupMock  func()
		wantStatus int
	}{
		{
			name:  "success - default params",
			query: "",
			setupMock: func() {
				mockArticleUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(&articledto.ListResponse{
						Data: []*articledto.Response{
							{ID: 1, Title: "Article 1"},
							{ID: 2, Title: "Article 2"},
						},
						Meta: pagination.NewMeta(1, 20, 2),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid query params",
			query:      "?user_id=abc",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:  "success - with filters",
			query: "?status=draft&page=1&limit=10",
			setupMock: func() {
				mockArticleUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(&articledto.ListResponse{
						Data: []*articledto.Response{},
						Meta: pagination.NewMeta(1, 10, 0),
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:  "usecase error",
			query: "",
			setupMock: func() {
				mockArticleUC.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/articles/"+tt.query, http.NoBody)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Update(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleUC := NewMockArticle(ctrl)
	app, token := setupTestApp(t, mockArticleUC)

	title := "Updated Title"
	validBody := `{"title":"Updated Title"}`

	tests := []struct {
		name       string
		id         string
		body       string
		addAuth    bool
		setupMock  func()
		wantStatus int
	}{
		{
			name:    "success",
			id:      "1",
			body:    validBody,
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Update(gomock.Any(), uint(1), articledto.UpdateRequest{
						Title: &title,
					}).
					Return(&articledto.Response{
						ID:    1,
						Title: "Updated Title",
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid id",
			id:         "invalid",
			body:       validBody,
			addAuth:    true,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "invalid json",
			id:         "1",
			body:       `{invalid}`,
			addAuth:    true,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:    "not found",
			id:      "999",
			body:    validBody,
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Update(gomock.Any(), uint(999), articledto.UpdateRequest{
						Title: &title,
					}).
					Return(nil, articleuc.ErrNotFound)
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:    "internal error",
			id:      "1",
			body:    validBody,
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Update(gomock.Any(), uint(1), articledto.UpdateRequest{
						Title: &title,
					}).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
		{
			name:       "unauthorized - no token",
			id:         "1",
			body:       validBody,
			addAuth:    false,
			setupMock:  func() {},
			wantStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/v1/articles/"+tt.id, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			if tt.addAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockArticleUC := NewMockArticle(ctrl)
	app, token := setupTestApp(t, mockArticleUC)

	tests := []struct {
		name       string
		id         string
		addAuth    bool
		setupMock  func()
		wantStatus int
	}{
		{
			name:    "success",
			id:      "1",
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantStatus: fiber.StatusNoContent,
		},
		{
			name:       "invalid id",
			id:         "invalid",
			addAuth:    true,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:    "not found",
			id:      "999",
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Delete(gomock.Any(), uint(999)).
					Return(articleuc.ErrNotFound)
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:    "internal error",
			id:      "1",
			addAuth: true,
			setupMock: func() {
				mockArticleUC.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
		{
			name:       "unauthorized - no token",
			id:         "1",
			addAuth:    false,
			setupMock:  func() {},
			wantStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/v1/articles/"+tt.id, http.NoBody)
			req.Header.Set("Content-Type", "application/json")

			if tt.addAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
