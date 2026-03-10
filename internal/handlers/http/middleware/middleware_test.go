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

	"go-boilerplate/internal/handlers/http/middleware"
)

func TestTimeout(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", middleware.Timeout(2*time.Second), func(c *fiber.Ctx) error {
		deadline, ok := c.UserContext().Deadline()
		return c.JSON(fiber.Map{
			"has_deadline": ok,
			"deadline":     deadline.Format(time.RFC3339Nano),
		})
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))
	assert.True(t, result["has_deadline"].(bool))
	assert.NotEmpty(t, result["deadline"])
}

func TestLogger(t *testing.T) {
	t.Parallel()

	l := newMockLogger()
	app := fiber.New()
	app.Get("/test", middleware.Logger(l), func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	assert.Equal(t, "ok", string(body))
}

func TestRecovery(t *testing.T) {
	t.Parallel()

	l := newMockLogger()
	app := fiber.New()
	app.Use(middleware.Recovery(l))
	app.Get("/test", func(_ *fiber.Ctx) error {
		panic("test panic")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
