package auth_test

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

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/handlers/http/v1/auth"
	"go-boilerplate/internal/repo"
	autherrors "go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

//go:generate mockgen -source=../../../../usecase/contracts.go -destination=./mocks_test.go -package=auth_test

func TestHandler_Login(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := NewMockAuth(ctrl)
	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := auth.New(mockAuthUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			body: `{"email":"test@example.com","password":"password123"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Login(gomock.Any(), authdto.LoginRequest{
						Email:    "test@example.com",
						Password: "password123",
					}).
					Return(&authdto.LoginResponse{
						AccessToken:  "token",
						RefreshToken: "refresh",
						ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
						User: authdto.UserResponse{
							ID:    1,
							Email: "test@example.com",
							Name:  "Test User",
							Role:  "user",
						},
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
			name:       "missing email",
			body:       `{"password":"password123"}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: `{"email":"test@example.com","password":"wrong"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Login(gomock.Any(), authdto.LoginRequest{
						Email:    "test@example.com",
						Password: "wrong",
					}).
					Return(nil, autherrors.ErrInvalidCredentials)
			},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "user not active",
			body: `{"email":"test@example.com","password":"password123"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Login(gomock.Any(), authdto.LoginRequest{
						Email:    "test@example.com",
						Password: "password123",
					}).
					Return(nil, autherrors.ErrUserNotActive)
			},
			wantStatus: fiber.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), "POST", "/v1/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Register(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := NewMockAuth(ctrl)
	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := auth.New(mockAuthUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			body: `{"email":"new@example.com","password":"password123","name":"New User"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Register(gomock.Any(), authdto.RegisterRequest{
						Email:    "new@example.com",
						Password: "password123",
						Name:     "New User",
					}).
					Return(&authdto.LoginResponse{
						AccessToken:  "token",
						RefreshToken: "refresh",
						ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
						User: authdto.UserResponse{
							ID:    1,
							Email: "new@example.com",
							Name:  "New User",
							Role:  "user",
						},
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
			name:       "validation error - missing name",
			body:       `{"email":"new@example.com","password":"password123"}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "email already exists",
			body: `{"email":"existing@example.com","password":"password123","name":"Existing User"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Register(gomock.Any(), authdto.RegisterRequest{
						Email:    "existing@example.com",
						Password: "password123",
						Name:     "Existing User",
					}).
					Return(nil, autherrors.ErrEmailExists)
			},
			wantStatus: fiber.StatusConflict,
		},
		{
			name: "internal error",
			body: `{"email":"new@example.com","password":"password123","name":"New User"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Register(gomock.Any(), authdto.RegisterRequest{
						Email:    "new@example.com",
						Password: "password123",
						Name:     "New User",
					}).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := NewMockAuth(ctrl)
	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := auth.New(mockAuthUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		wantStatus int
	}{
		{
			name: "success",
			body: `{"refresh_token":"valid-refresh-token"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Refresh(gomock.Any(), authdto.RefreshRequest{
						RefreshToken: "valid-refresh-token",
					}).
					Return(&authdto.TokenResponse{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
						ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
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
			name:       "missing refresh_token",
			body:       `{}`,
			setupMock:  func() {},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "invalid token",
			body: `{"refresh_token":"expired-token"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Refresh(gomock.Any(), authdto.RefreshRequest{
						RefreshToken: "expired-token",
					}).
					Return(nil, autherrors.ErrInvalidToken)
			},
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "internal error",
			body: `{"refresh_token":"some-token"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Refresh(gomock.Any(), authdto.RefreshRequest{
						RefreshToken: "some-token",
					}).
					Return(nil, errors.New("database error"))
			},
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/auth/refresh", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := NewMockAuth(ctrl)
	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := auth.New(mockAuthUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token, _, err := jwtService.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		body       string
		setupMock  func()
		withAuth   bool
		wantStatus int
	}{
		{
			name: "success",
			body: `{"refresh_token":"valid-refresh-token"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Logout(gomock.Any(), "valid-refresh-token").
					Return(nil)
			},
			withAuth:   true,
			wantStatus: fiber.StatusNoContent,
		},
		{
			name:       "no auth header",
			body:       `{"refresh_token":"valid-refresh-token"}`,
			setupMock:  func() {},
			withAuth:   false,
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			setupMock:  func() {},
			withAuth:   true,
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "missing refresh_token",
			body:       `{}`,
			setupMock:  func() {},
			withAuth:   true,
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name: "usecase error",
			body: `{"refresh_token":"some-token"}`,
			setupMock: func() {
				mockAuthUC.EXPECT().
					Logout(gomock.Any(), "some-token").
					Return(errors.New("database error"))
			},
			withAuth:   true,
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/auth/logout", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.withAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandler_Me(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := NewMockAuth(ctrl)
	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := logger.NewDevelopment()

	handler := auth.New(mockAuthUC, jwtService, l)

	app := fiber.New()
	handler.RegisterRoutes(app.Group("/v1"))

	token, _, err := jwtService.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		setupMock  func()
		withAuth   bool
		wantStatus int
	}{
		{
			name: "success",
			setupMock: func() {
				mockAuthUC.EXPECT().
					GetCurrentUser(gomock.Any(), uint(1)).
					Return(&authdto.UserResponse{
						ID:    1,
						Email: "test@example.com",
						Name:  "Test User",
						Role:  "user",
					}, nil)
			},
			withAuth:   true,
			wantStatus: fiber.StatusOK,
		},
		{
			name:       "no auth header",
			setupMock:  func() {},
			withAuth:   false,
			wantStatus: fiber.StatusUnauthorized,
		},
		{
			name: "user not found",
			setupMock: func() {
				mockAuthUC.EXPECT().
					GetCurrentUser(gomock.Any(), uint(1)).
					Return(nil, repo.ErrNotFound)
			},
			withAuth:   true,
			wantStatus: fiber.StatusNotFound,
		},
		{
			name: "internal error",
			setupMock: func() {
				mockAuthUC.EXPECT().
					GetCurrentUser(gomock.Any(), uint(1)).
					Return(nil, errors.New("database error"))
			},
			withAuth:   true,
			wantStatus: fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/v1/auth/me", http.NoBody)
			req.Header.Set("Content-Type", "application/json")
			if tt.withAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
