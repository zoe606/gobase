package permission_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	permissiondto "go-boilerplate/internal/dto/permission"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/handlers/http/v1/permission"
	permissionusecase "go-boilerplate/internal/usecase/permission"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

//go:generate mockgen -source=../../../../usecase/contracts.go -destination=./mocks_test.go -package=permission_test

const testSecret = "test-secret"

// generateTestToken creates a valid JWT token with permissions:read for tests.
func generateTestToken(t *testing.T, jwtService jwt.Service) string {
	t.Helper()
	token, _, err := jwtService.GenerateAccessToken(1, "admin@test.com", "admin", []string{"permissions:read", "permissions:write", "permissions:delete"})
	require.NoError(t, err)
	return token
}

// authRequest creates an http.Request with Authorization header.
func authRequest(t *testing.T, method, url string, body *bytes.Buffer, token string) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, body)
	} else {
		req = httptest.NewRequest(method, url, http.NoBody)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandler_Create(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermUC := NewMockPermission(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := permission.New(mockPermUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			body: `{"resource":"users","action":"write"}`,
			setupMock: func() {
				mockPermUC.EXPECT().
					Create(gomock.Any(), permissiondto.CreateRequest{
						Resource: "users",
						Action:   "write",
					}).
					Return(&permissiondto.Response{
						ID:       1,
						Name:     "users:write",
						Resource: "users",
						Action:   "write",
					}, nil)
			},
			wantStatus: fiber.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "missing resource",
			body:       `{"action":"read"}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "missing action",
			body:       `{"resource":"users"}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "resource too short",
			body:       `{"resource":"a","action":"read"}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "permission already exists",
			body: `{"resource":"users","action":"read"}`,
			setupMock: func() {
				mockPermUC.EXPECT().
					Create(gomock.Any(), permissiondto.CreateRequest{
						Resource: "users",
						Action:   "read",
					}).
					Return(nil, permissionusecase.ErrPermissionExists)
			},
			wantStatus: fiber.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "POST", "/v1/permissions/", bytes.NewBufferString(tt.body), token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermUC := NewMockPermission(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := permission.New(mockPermUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	req := httptest.NewRequest("POST", "/v1/permissions/", bytes.NewBufferString(`{"resource":"users","action":"read"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_Delete(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermUC := NewMockPermission(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := permission.New(mockPermUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

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
				mockPermUC.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(nil)
			},
			wantStatus: fiber.StatusNoContent,
		},
		{
			name:       "invalid id",
			id:         "abc",
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "99",
			setupMock: func() {
				mockPermUC.EXPECT().
					Delete(gomock.Any(), uint(99)).
					Return(permissionusecase.ErrPermissionNotFound)
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name: "permission in use",
			id:   "1",
			setupMock: func() {
				mockPermUC.EXPECT().
					Delete(gomock.Any(), uint(1)).
					Return(permissionusecase.ErrPermissionInUse)
			},
			wantStatus: fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "DELETE", "/v1/permissions/"+tt.id, nil, token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermUC := NewMockPermission(ctrl)
	jwtService := jwt.New(testSecret, 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := permission.New(mockPermUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token := generateTestToken(t, jwtService)

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			setupMock: func() {
				mockPermUC.EXPECT().
					List(gomock.Any()).
					Return([]*entity.Permission{
						{ID: 1, Name: "users:read", Resource: "users", Action: "read"},
						{ID: 2, Name: "users:write", Resource: "users", Action: "write"},
					}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
		{
			name: "empty list",
			setupMock: func() {
				mockPermUC.EXPECT().
					List(gomock.Any()).
					Return([]*entity.Permission{}, nil)
			},
			wantStatus: fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := authRequest(t, "GET", "/v1/permissions/", nil, token)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
