package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"go-boilerplate/config"
	"go-boilerplate/internal/handlers/http/middleware"
)

// newObservedLogger creates a zap logger that captures log entries for testing.
func newObservedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	return zap.New(core), logs
}

func TestStructuredLogger_BasicFields(t *testing.T) {
	t.Parallel()

	zapLogger, logs := newObservedLogger()
	cfg := config.Log{Level: "debug"}

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(middleware.StructuredLogger(zapLogger, cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set("User-Agent", "test-agent")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, 1, logs.Len(), "expected exactly one log entry")
	entry := logs.All()[0]

	assert.Equal(t, "HTTP request", entry.Message)
	assert.Equal(t, zapcore.InfoLevel, entry.Level)

	fieldMap := fieldsToMap(entry.ContextMap())
	assert.Equal(t, "GET", fieldMap["method"])
	assert.Equal(t, "/test", fieldMap["path"])
	assert.Equal(t, int64(http.StatusOK), fieldMap["status"])
	assert.Contains(t, fieldMap, "latency_ms")
	assert.Contains(t, fieldMap, "request_id")
	assert.NotEmpty(t, fieldMap["request_id"])
	assert.Contains(t, fieldMap, "ip")
	assert.Equal(t, "test-agent", fieldMap["user_agent"])
	assert.Contains(t, fieldMap, "bytes_out")
}

func TestStructuredLogger_LogLevelByStatusCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		status    int
		wantLevel zapcore.Level
	}{
		{"2xx returns info", http.StatusOK, zapcore.InfoLevel},
		{"201 returns info", http.StatusCreated, zapcore.InfoLevel},
		{"4xx returns warn", http.StatusBadRequest, zapcore.WarnLevel},
		{"404 returns warn", http.StatusNotFound, zapcore.WarnLevel},
		{"5xx returns error", http.StatusInternalServerError, zapcore.ErrorLevel},
		{"503 returns error", http.StatusServiceUnavailable, zapcore.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zapLogger, logs := newObservedLogger()
			cfg := config.Log{Level: "debug"}

			app := fiber.New()
			app.Use(requestid.New())
			app.Use(middleware.StructuredLogger(zapLogger, cfg))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(tt.status)
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close() //nolint:errcheck // test

			require.Equal(t, 1, logs.Len())
			assert.Equal(t, tt.wantLevel, logs.All()[0].Level)
		})
	}
}

func TestStructuredLogger_NoBodyLoggedByDefault(t *testing.T) {
	t.Parallel()

	zapLogger, logs := newObservedLogger()
	cfg := config.Log{Level: "debug"}

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(middleware.StructuredLogger(zapLogger, cfg))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("response-body")
	})

	body := bytes.NewBufferString(`{"password":"secret123","username":"john"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	require.Equal(t, 1, logs.Len())
	fieldMap := fieldsToMap(logs.All()[0].ContextMap())
	assert.NotContains(t, fieldMap, "request_body")
	assert.NotContains(t, fieldMap, "response_body")
}

func TestStructuredLogger_RequestBodyRedaction(t *testing.T) {
	t.Parallel()

	zapLogger, logs := newObservedLogger()
	cfg := config.Log{
		Level:          "debug",
		LogRequestBody: true,
		RedactFields:   "password,token,secret",
	}

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(middleware.StructuredLogger(zapLogger, cfg))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	body := bytes.NewBufferString(`{"password":"secret123","username":"john","token":"abc"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	require.Equal(t, 1, logs.Len())
	fieldMap := fieldsToMap(logs.All()[0].ContextMap())
	assert.Contains(t, fieldMap, "request_body")

	reqBody, ok := fieldMap["request_body"].(string)
	require.True(t, ok)
	assert.NotContains(t, reqBody, "secret123")
	assert.NotContains(t, reqBody, "abc")
	assert.Contains(t, reqBody, "[REDACTED]")
	assert.Contains(t, reqBody, "john")
}

func TestStructuredLogger_ResponseBodyLogged(t *testing.T) {
	t.Parallel()

	zapLogger, logs := newObservedLogger()
	cfg := config.Log{
		Level:           "debug",
		LogResponseBody: true,
	}

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(middleware.StructuredLogger(zapLogger, cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("hello-response")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	require.Equal(t, 1, logs.Len())
	fieldMap := fieldsToMap(logs.All()[0].ContextMap())
	assert.Contains(t, fieldMap, "response_body")
	assert.Equal(t, "hello-response", fieldMap["response_body"])
}

func TestStructuredLogger_UserIDFromContext(t *testing.T) {
	t.Parallel()

	zapLogger, logs := newObservedLogger()
	cfg := config.Log{Level: "debug"}

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(middleware.StructuredLogger(zapLogger, cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("user_id", uint(42))
		return c.SendString("ok")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	require.Equal(t, 1, logs.Len())
	fieldMap := fieldsToMap(logs.All()[0].ContextMap())
	assert.Equal(t, uint64(42), fieldMap["user_id"])
}

func TestParseRedactFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect []string
	}{
		{"empty string", "", nil},
		{"single field", "password", []string{"password"}},
		{"multiple fields", "password,token,secret", []string{"password", "token", "secret"}},
		{"fields with spaces", " password , token , secret ", []string{"password", "token", "secret"}},
		{"trailing comma", "password,token,", []string{"password", "token"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := middleware.ParseRedactFields(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestRedactJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		fields []string
		check  func(t *testing.T, result string)
	}{
		{
			name:   "redacts password field",
			body:   `{"password":"secret123","username":"john"}`,
			fields: []string{"password"},
			check: func(t *testing.T, result string) {
				t.Helper()
				assert.NotContains(t, result, "secret123")
				assert.Contains(t, result, "[REDACTED]")
				assert.Contains(t, result, "john")
			},
		},
		{
			name:   "redacts multiple fields",
			body:   `{"password":"mysecretpass","token":"mytoken123","name":"john"}`,
			fields: []string{"password", "token"},
			check: func(t *testing.T, result string) {
				t.Helper()
				assert.NotContains(t, result, "mysecretpass")
				assert.NotContains(t, result, "mytoken123")
				assert.Contains(t, result, "john")
			},
		},
		{
			name:   "returns input for invalid JSON",
			body:   `not-json`,
			fields: []string{"password"},
			check: func(t *testing.T, result string) {
				t.Helper()
				assert.Equal(t, "not-json", result)
			},
		},
		{
			name:   "returns input for empty fields",
			body:   `{"password":"secret"}`,
			fields: nil,
			check: func(t *testing.T, result string) {
				t.Helper()
				assert.Equal(t, `{"password":"secret"}`, result)
			},
		},
		{
			name:   "handles nested objects without redacting nested keys",
			body:   `{"user":{"password":"secret"},"name":"john"}`,
			fields: []string{"password"},
			check: func(t *testing.T, result string) {
				t.Helper()
				// Top-level password should be redacted if present, nested should stay
				assert.Contains(t, result, "john")
			},
		},
		{
			name:   "empty body",
			body:   "",
			fields: []string{"password"},
			check: func(t *testing.T, result string) {
				t.Helper()
				assert.Equal(t, "", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := middleware.RedactJSON(tt.body, tt.fields)
			tt.check(t, result)
		})
	}
}

func TestStructuredLogger_3xxStatus(t *testing.T) {
	t.Parallel()

	zapLogger, logs := newObservedLogger()
	cfg := config.Log{Level: "debug"}

	app := fiber.New()
	app.Use(requestid.New())
	app.Use(middleware.StructuredLogger(zapLogger, cfg))
	app.Get("/redirect", func(c *fiber.Ctx) error {
		return c.Redirect("/other", http.StatusMovedPermanently)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/redirect", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	require.Equal(t, 1, logs.Len())
	// 3xx should log at info level (default)
	assert.Equal(t, zapcore.InfoLevel, logs.All()[0].Level)
}

// fieldsToMap converts a zap observed context map for easier assertions.
func fieldsToMap(m map[string]interface{}) map[string]interface{} {
	return m
}
