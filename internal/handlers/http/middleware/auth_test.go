package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

// mockLogger implements logger.Interface for testing without console output.
type mockLogger struct{}

var _ logger.Interface = (*mockLogger)(nil)

func newMockLogger() *mockLogger                            { return &mockLogger{} }
func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) GetZapLogger() *zap.Logger             { return zap.NewNop() }

func TestJWTAuth(t *testing.T) {
	t.Parallel()

	jwtService := jwt.New("test-secret", 15*time.Minute, 24*time.Hour)
	l := newMockLogger()

	validToken, _, err := jwtService.GenerateAccessToken(1, "test@example.com", "user", []string{"read", "write"})
	require.NoError(t, err)

	// Create a short-lived JWT service for expired token test.
	shortJWT := jwt.New("test-secret", 1*time.Millisecond, 1*time.Millisecond)
	expiredToken, _, err := shortJWT.GenerateAccessToken(1, "test@example.com", "user", nil)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "no auth header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "Missing authorization header",
		},
		{
			name:       "invalid format without Bearer prefix",
			authHeader: "Token abc123",
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "Invalid authorization format. Use: Bearer <token>",
		},
		{
			name:       "invalid format single word",
			authHeader: "abc123",
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "Invalid authorization format. Use: Bearer <token>",
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + expiredToken,
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "Token has expired",
		},
		{
			name:       "invalid token string",
			authHeader: "Bearer not-a-real-jwt-token",
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "Invalid token",
		},
		{
			name:       "valid token",
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Get("/test", middleware.JWTAuth(jwtService, l), func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"user_id": middleware.GetUserID(c)})
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantMsg != "" {
				body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
				var result map[string]interface{}
				require.NoError(t, json.Unmarshal(body, &result))
				errObj, ok := result["error"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, tt.wantMsg, errObj["message"])
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupRole  func(c *fiber.Ctx)
		roles      []string
		wantStatus int
	}{
		{
			name:       "no role in context",
			setupRole:  func(_ *fiber.Ctx) {},
			roles:      []string{"admin"},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "wrong role",
			setupRole: func(c *fiber.Ctx) {
				c.Locals(middleware.RoleKey, "user")
			},
			roles:      []string{"admin"},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "matching role",
			setupRole: func(c *fiber.Ctx) {
				c.Locals(middleware.RoleKey, "admin")
			},
			roles:      []string{"admin"},
			wantStatus: http.StatusOK,
		},
		{
			name: "case insensitive match",
			setupRole: func(c *fiber.Ctx) {
				c.Locals(middleware.RoleKey, "Admin")
			},
			roles:      []string{"admin"},
			wantStatus: http.StatusOK,
		},
		{
			name: "one of multiple roles matches",
			setupRole: func(c *fiber.Ctx) {
				c.Locals(middleware.RoleKey, "editor")
			},
			roles:      []string{"admin", "editor"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setupRole(c)
				return c.Next()
			}, middleware.RequireRole(tt.roles...), func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestRequirePermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupPerms func(c *fiber.Ctx)
		permission string
		wantStatus int
	}{
		{
			name:       "no permissions in context",
			setupPerms: func(_ *fiber.Ctx) {},
			permission: "write",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "missing required permission",
			setupPerms: func(c *fiber.Ctx) {
				c.Locals(middleware.PermissionsKey, []string{"read"})
			},
			permission: "write",
			wantStatus: http.StatusForbidden,
		},
		{
			name: "has required permission",
			setupPerms: func(c *fiber.Ctx) {
				c.Locals(middleware.PermissionsKey, []string{"read", "write"})
			},
			permission: "write",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setupPerms(c)
				return c.Next()
			}, middleware.RequirePermission(tt.permission), func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupPerms func(c *fiber.Ctx)
		required   []string
		wantStatus int
	}{
		{
			name:       "no permissions in context",
			setupPerms: func(_ *fiber.Ctx) {},
			required:   []string{"write", "delete"},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "none match",
			setupPerms: func(c *fiber.Ctx) {
				c.Locals(middleware.PermissionsKey, []string{"read"})
			},
			required:   []string{"write", "delete"},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "one matches",
			setupPerms: func(c *fiber.Ctx) {
				c.Locals(middleware.PermissionsKey, []string{"read", "write"})
			},
			required:   []string{"write", "delete"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setupPerms(c)
				return c.Next()
			}, middleware.RequireAnyPermission(tt.required...), func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestGetUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		setup  func(c *fiber.Ctx)
		wantID uint
	}{
		{
			name: "with value set",
			setup: func(c *fiber.Ctx) {
				c.Locals(middleware.UserIDKey, uint(42))
			},
			wantID: 42,
		},
		{
			name:   "without value",
			setup:  func(_ *fiber.Ctx) {},
			wantID: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			var gotID uint
			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setup(c)
				gotID = middleware.GetUserID(c)
				return c.SendString("ok")
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantID, gotID)
		})
	}
}

func TestGetEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(c *fiber.Ctx)
		wantEmail string
	}{
		{
			name: "with value set",
			setup: func(c *fiber.Ctx) {
				c.Locals(middleware.EmailKey, "test@example.com")
			},
			wantEmail: "test@example.com",
		},
		{
			name:      "without value",
			setup:     func(_ *fiber.Ctx) {},
			wantEmail: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			var gotEmail string
			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setup(c)
				gotEmail = middleware.GetEmail(c)
				return c.SendString("ok")
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantEmail, gotEmail)
		})
	}
}

func TestGetRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(c *fiber.Ctx)
		wantRole string
	}{
		{
			name: "with value set",
			setup: func(c *fiber.Ctx) {
				c.Locals(middleware.RoleKey, "admin")
			},
			wantRole: "admin",
		},
		{
			name:     "without value",
			setup:    func(_ *fiber.Ctx) {},
			wantRole: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			var gotRole string
			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setup(c)
				gotRole = middleware.GetRole(c)
				return c.SendString("ok")
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantRole, gotRole)
		})
	}
}

func TestGetPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(c *fiber.Ctx)
		wantPerms []string
	}{
		{
			name: "with value set",
			setup: func(c *fiber.Ctx) {
				c.Locals(middleware.PermissionsKey, []string{"read", "write"})
			},
			wantPerms: []string{"read", "write"},
		},
		{
			name:      "without value",
			setup:     func(_ *fiber.Ctx) {},
			wantPerms: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			var gotPerms []string
			app.Get("/test", func(c *fiber.Ctx) error {
				tt.setup(c)
				gotPerms = middleware.GetPermissions(c)
				return c.SendString("ok")
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			assert.Equal(t, tt.wantPerms, gotPerms)
		})
	}
}
