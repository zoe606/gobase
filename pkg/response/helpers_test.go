package response_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/apperror"
	"go-boilerplate/pkg/response"
)

func TestOKWithMeta(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		data := []string{"a", "b"}
		meta := &response.Meta{Page: 1, Limit: 10, Total: 2, TotalPages: 1}
		return response.OKWithMeta(c, data, meta)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.Response[[]string]
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.NotNil(t, result.Meta)
	require.Equal(t, 1, result.Meta.Page)
	require.Equal(t, int64(2), result.Meta.Total)
}

func TestRateLimited(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.RateLimited(c, "Too many requests")
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test
	require.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
}

func TestErrorWithDetails(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return response.ErrorWithDetails(c, http.StatusBadRequest, "INVALID", "Bad data", map[string]string{
			"field": "is required",
		})
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.ErrorResponse
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.Equal(t, "INVALID", result.Error.Code)
	require.NotEmpty(t, result.Error.Details)
}

func TestFromAppError(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		appErr := apperror.New("CUSTOM_ERROR", "Something went wrong", http.StatusUnprocessableEntity)
		return response.FromAppError(c, appErr)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test
	require.Equal(t, 422, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.ErrorResponse
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.Equal(t, "CUSTOM_ERROR", result.Error.Code)
	require.Equal(t, "Something went wrong", result.Error.Message)
}

func TestGetRequestID_WithLocal(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals("requestid", "req-123")
		return response.OK(c, "data")
	})
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.Response[string]
	require.NoError(t, json.Unmarshal(body, &result))
	require.Equal(t, "req-123", result.RequestID)
}

func TestGetRequestID_WithHeader(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.OK(c, "data")
	})
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set(fiber.HeaderXRequestID, "hdr-456")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.Response[string]
	require.NoError(t, json.Unmarshal(body, &result))
	require.Equal(t, "hdr-456", result.RequestID)
}
