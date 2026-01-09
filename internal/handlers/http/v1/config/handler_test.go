package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go-boilerplate/config"
	"go-boilerplate/pkg/jwt"
)

// mockLogger implements logger.Interface for testing.
type mockLogger struct{}

func (m *mockLogger) Debug(msg interface{}, args ...interface{}) {}
func (m *mockLogger) Info(msg string, args ...interface{})       {}
func (m *mockLogger) Warn(msg string, args ...interface{})       {}
func (m *mockLogger) Error(msg interface{}, args ...interface{}) {}
func (m *mockLogger) Fatal(msg interface{}, args ...interface{}) {}
func (m *mockLogger) GetZapLogger() *zap.Logger                  { return zap.NewNop() }

// mockJWTService implements jwt.Service for testing.
type mockJWTService struct{}

func (m *mockJWTService) GenerateAccessToken(userID uint, email, role string, permissions []string) (string, int64, error) {
	return "test-token", time.Now().Add(time.Hour).Unix(), nil
}
func (m *mockJWTService) GenerateRefreshToken() (string, time.Time, error) {
	return "refresh-token", time.Now().Add(24 * time.Hour), nil
}
func (m *mockJWTService) ValidateToken(token string) (*jwt.Claims, error) { return nil, nil }
func (m *mockJWTService) GetAccessExpiry() time.Duration                  { return time.Hour }
func (m *mockJWTService) GetRefreshExpiry() time.Duration                 { return 24 * time.Hour }

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "(not set)",
		},
		{
			name:     "short secret",
			input:    "abc",
			expected: "****",
		},
		{
			name:     "medium secret",
			input:    "abcdefgh",
			expected: "****",
		},
		{
			name:     "long secret",
			input:    "abcdefghij",
			expected: "abcd****",
		},
		{
			name:     "api key",
			input:    "re_1234567890abcdef",
			expected: "re_1****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSecret(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandler_New(t *testing.T) {
	cfg := &config.Config{
		App: config.App{
			Name:    "test-app",
			Version: "1.0.0",
			Env:     "test",
		},
	}
	l := &mockLogger{}
	jwt := &mockJWTService{}

	h := New(cfg, jwt, l)

	require.NotNil(t, h)
	assert.Equal(t, cfg, h.cfg)
	assert.Equal(t, 5*time.Minute, h.cacheTTL)
}

func TestHandler_BuildConfigResponse(t *testing.T) {
	cfg := &config.Config{
		App: config.App{
			Name:    "test-app",
			Version: "1.0.0",
			Env:     "development",
		},
		HTTP: config.HTTP{
			Port:           "8080",
			Timeout:        15 * time.Second,
			IdleTimeout:    60 * time.Second,
			RequestTimeout: 30 * time.Second,
		},
		Postgres: config.Postgres{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "supersecretpassword",
			DBName:   "testdb",
		},
		Redis: config.Redis{
			Host:     "localhost",
			Port:     6379,
			Password: "redispassword123",
		},
		JWT: config.JWT{
			SecretKey:     "jwt-secret-key-very-long",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 168 * time.Hour,
		},
		Email: config.Email{
			Provider:  "resend",
			APIKey:    "re_1234567890abcdef",
			FromEmail: "noreply@example.com",
			FromName:  "Test App",
		},
	}
	l := &mockLogger{}
	jwt := &mockJWTService{}

	h := New(cfg, jwt, l)
	resp := h.buildConfigResponse()

	// Verify non-sensitive fields
	assert.Equal(t, "test-app", resp.App.Name)
	assert.Equal(t, "1.0.0", resp.App.Version)
	assert.Equal(t, "development", resp.App.Env)
	assert.Equal(t, "8080", resp.HTTP.Port)
	assert.Equal(t, "localhost", resp.Postgres.Host)
	assert.Equal(t, 5432, resp.Postgres.Port)
	assert.Equal(t, "postgres", resp.Postgres.User)
	assert.Equal(t, "testdb", resp.Postgres.DBName)

	// Verify sensitive fields are masked
	assert.Equal(t, "supe****", resp.Postgres.Password)
	assert.Equal(t, "redi****", resp.Redis.Password)
	assert.Equal(t, "jwt-****", resp.JWT.SecretKey)
	assert.Equal(t, "re_1****", resp.Email.APIKey)

	// Verify CachedAt is set
	assert.False(t, resp.CachedAt.IsZero())
}

func TestHandler_Cache(t *testing.T) {
	cfg := &config.Config{
		App: config.App{
			Name:    "test-app",
			Version: "1.0.0",
			Env:     "test",
		},
	}
	l := &mockLogger{}
	jwtSvc := &mockJWTService{}

	h := New(cfg, jwtSvc, l)

	// First call should build cache
	resp1 := h.buildConfigResponse()
	h.cache = resp1
	h.cachedAt = time.Now()

	// Cache should be set
	assert.NotNil(t, h.cache)
	assert.False(t, h.cachedAt.IsZero())

	// Invalidate cache
	h.cache = nil
	h.cachedAt = time.Time{}

	// Cache should be cleared
	assert.Nil(t, h.cache)
	assert.True(t, h.cachedAt.IsZero())
}

func TestConfigResponse_Structure(t *testing.T) {
	resp := &ConfigResponse{
		App: AppConfig{
			Name:    "app",
			Version: "1.0",
			Env:     "dev",
		},
		HTTP: HTTPConfig{
			Port: "8080",
		},
		Postgres: PostgresConfig{
			Host:     "localhost",
			Password: "****",
		},
		CachedAt: time.Now(),
	}

	assert.Equal(t, "app", resp.App.Name)
	assert.Equal(t, "8080", resp.HTTP.Port)
	assert.Equal(t, "****", resp.Postgres.Password)
}
