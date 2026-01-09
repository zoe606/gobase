package media_test

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mediadto "go-boilerplate/internal/dto/media"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/handlers/http/v1/media"
	"go-boilerplate/internal/repo"
	"go-boilerplate/pkg/jwt"
)

func setupTestApp(t *testing.T, mockMediaUC *MockMedia) (*fiber.App, string) {
	t.Helper()

	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := NewMockLogger()

	handler := media.New(mockMediaUC, jwtService, l, 10*1024*1024) // 10MB max

	app := fiber.New()

	handler.RegisterRoutes(app.Group("/v1"))

	// Generate a valid test token
	token, _, err := jwtService.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	return app, token
}

func TestHandler_GetByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMediaUC := NewMockMedia(ctrl)
	app, token := setupTestApp(t, mockMediaUC)

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
				mockMediaUC.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(&entity.Media{
						ID:             1,
						AttachableType: "users",
						AttachableID:   1,
						Collection:     "avatar",
						Filename:       "test.jpg",
						OriginalName:   "avatar.jpg",
						MimeType:       "image/jpeg",
						Size:           1024,
						Type:           entity.MediaTypeImage,
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
				mockMediaUC.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantStatus: fiber.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/v1/media/"+tt.id, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_GetURL(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMediaUC := NewMockMedia(ctrl)
	app, token := setupTestApp(t, mockMediaUC)

	tests := []struct {
		name       string
		id         string
		variant    string
		setupMock  func()
		wantStatus int
	}{
		{
			name:    "success - original",
			id:      "1",
			variant: "",
			setupMock: func() {
				mockMedia := &entity.Media{
					ID:   1,
					Path: "users/avatar/1/test.jpg",
				}
				mockMediaUC.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(mockMedia, nil)
				mockMediaUC.EXPECT().
					GetURL(gomock.Any(), mockMedia, "").
					Return("https://s3.example.com/signed-url", nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:    "success - with variant",
			id:      "1",
			variant: "thumb",
			setupMock: func() {
				mockMedia := &entity.Media{
					ID:   1,
					Path: "users/avatar/1/test.jpg",
				}
				mockMediaUC.EXPECT().
					GetByID(gomock.Any(), uint(1)).
					Return(mockMedia, nil)
				mockMediaUC.EXPECT().
					GetURL(gomock.Any(), mockMedia, "thumb").
					Return("https://s3.example.com/signed-thumb-url", nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "invalid id",
			id:         "invalid",
			variant:    "",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:    "not found",
			id:      "999",
			variant: "",
			setupMock: func() {
				mockMediaUC.EXPECT().
					GetByID(gomock.Any(), uint(999)).
					Return(nil, repo.ErrNotFound)
			},
			wantStatus: fiber.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			url := "/v1/media/" + tt.id + "/url"
			if tt.variant != "" {
				url += "?variant=" + tt.variant
			}

			req := httptest.NewRequest("GET", url, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMediaUC := NewMockMedia(ctrl)
	app, token := setupTestApp(t, mockMediaUC)

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
				mockMediaUC.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
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
				mockMediaUC.EXPECT().
					Delete(gomock.Any(), uint(999)).
					Return(repo.ErrNotFound)
			},
			wantStatus: fiber.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("DELETE", "/v1/media/"+tt.id, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_GetPresignedURL(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMediaUC := NewMockMedia(ctrl)
	app, token := setupTestApp(t, mockMediaUC)

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			body: `{"filename":"document.pdf"}`,
			setupMock: func() {
				mockMediaUC.EXPECT().
					GetPresignedUploadURL(gomock.Any(), "document.pdf").
					Return(&mediadto.PresignedURLResponse{
						UploadURL: "https://s3.example.com/presigned-upload",
						Path:      "temp/2024/01/uuid.pdf",
						ExpiresIn: 900,
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
			name:       "missing filename",
			body:       `{}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "usecase error",
			body: `{"filename":"test.jpg"}`,
			setupMock: func() {
				mockMediaUC.EXPECT().
					GetPresignedUploadURL(gomock.Any(), "test.jpg").
					Return(nil, errors.New("storage error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("POST", "/v1/media/presigned-url", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_GetByAttachable(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMediaUC := NewMockMedia(ctrl)
	app, token := setupTestApp(t, mockMediaUC)

	tests := []struct {
		name       string
		query      string
		setupMock  func()
		wantStatus int
	}{
		{
			name:  "success - with collection",
			query: "?attachable_type=users&attachable_id=1&collection=avatar",
			setupMock: func() {
				mockMediaUC.EXPECT().
					GetByAttachable(gomock.Any(), mediadto.GetMediaRequest{
						AttachableType: "users",
						AttachableID:   1,
						Collection:     "avatar",
					}).
					Return(&mediadto.MediaListResponse{
						Items: []*mediadto.MediaResponse{
							{ID: 1, Filename: "avatar.jpg"},
						},
						Total: 1,
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:  "success - without collection",
			query: "?attachable_type=users&attachable_id=1",
			setupMock: func() {
				mockMediaUC.EXPECT().
					GetByAttachable(gomock.Any(), mediadto.GetMediaRequest{
						AttachableType: "users",
						AttachableID:   1,
						Collection:     "",
					}).
					Return(&mediadto.MediaListResponse{
						Items: []*mediadto.MediaResponse{},
						Total: 0,
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "missing attachable_type",
			query:      "?attachable_id=1",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "invalid attachable_id",
			query:      "?attachable_type=users&attachable_id=invalid",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:  "usecase error",
			query: "?attachable_type=users&attachable_id=1",
			setupMock: func() {
				mockMediaUC.EXPECT().
					GetByAttachable(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/v1/media"+tt.query, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
