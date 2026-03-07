package auth_test

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/handlers/http/v1/auth"
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

			req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
