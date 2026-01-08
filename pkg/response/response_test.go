package response_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/response"
)

func TestOK(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.OK(c, map[string]string{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.Response[map[string]string]
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, "success", result.Data["message"])
}

func TestCreated(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return response.Created(c, map[string]int{"id": 1})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestNoContent(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Delete("/test", func(c *fiber.Ctx) error {
		return response.NoContent(c)
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestBadRequest(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.BadRequest(c, "INVALID_INPUT", "Invalid input provided")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.ErrorResponse
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.False(t, result.Success)
	require.Equal(t, "INVALID_INPUT", result.Error.Code)
	require.Equal(t, "Invalid input provided", result.Error.Message)
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.Unauthorized(c, "Invalid credentials")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestForbidden(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.Forbidden(c, "Access denied")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestNotFound(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.NotFound(c, "Resource not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestConflict(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		return response.Conflict(c, "Resource already exists")
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusConflict, resp.StatusCode)
}

func TestValidationError(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Post("/test", func(c *fiber.Ctx) error {
		errors := map[string]string{
			"email":    "invalid email format",
			"password": "password too short",
		}
		return response.ValidationError(c, errors)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode) // ValidationError returns 400

	body, _ := io.ReadAll(resp.Body) //nolint:errcheck // test
	var result response.ErrorResponse
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	require.False(t, result.Success)
	require.Equal(t, "VALIDATION_ERROR", result.Error.Code)
	require.NotEmpty(t, result.Error.Details)
}

func TestInternalError(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return response.InternalError(c)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
