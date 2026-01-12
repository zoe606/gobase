package response

import (
	"net/http"

	"go-boilerplate/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

// getRequestID extracts the request ID from Fiber context.
func getRequestID(c *fiber.Ctx) string {
	if id := c.Locals("requestid"); id != nil {
		if reqID, ok := id.(string); ok {
			return reqID
		}
	}
	return c.Get(fiber.HeaderXRequestID)
}

// OK sends a 200 OK response with data.
func OK[T any](c *fiber.Ctx, data T) error {
	return c.Status(http.StatusOK).JSON(Response[T]{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// OKWithMeta sends a 200 OK response with data and pagination metadata.
func OKWithMeta[T any](c *fiber.Ctx, data T, meta *Meta) error {
	return c.Status(http.StatusOK).JSON(Response[T]{
		Success:   true,
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 Created response with data.
func Created[T any](c *fiber.Ctx, data T) error {
	return c.Status(http.StatusCreated).JSON(Response[T]{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// NoContent sends a 204 No Content response.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

// Error sends an error response with the specified status code.
func Error(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
		RequestID: getRequestID(c),
	})
}

// ErrorWithDetails sends an error response with field-level details.
func ErrorWithDetails(c *fiber.Ctx, status int, code, message string, details map[string]string) error {
	return c.Status(status).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: getRequestID(c),
	})
}

// BadRequest sends a 400 Bad Request response.
func BadRequest(c *fiber.Ctx, code, message string) error {
	return Error(c, http.StatusBadRequest, code, message)
}

// ValidationError sends a 400 Bad Request response with validation details.
func ValidationError(c *fiber.Ctx, details map[string]string) error {
	return ErrorWithDetails(c, http.StatusBadRequest, apperror.CodeValidationError, "Validation failed", details)
}

// Unauthorized sends a 401 Unauthorized response.
func Unauthorized(c *fiber.Ctx, message string) error {
	return Error(c, http.StatusUnauthorized, apperror.CodeUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response.
func Forbidden(c *fiber.Ctx, message string) error {
	return Error(c, http.StatusForbidden, apperror.CodeForbidden, message)
}

// NotFound sends a 404 Not Found response.
func NotFound(c *fiber.Ctx, message string) error {
	return Error(c, http.StatusNotFound, apperror.CodeNotFound, message)
}

// Conflict sends a 409 Conflict response.
func Conflict(c *fiber.Ctx, message string) error {
	return Error(c, http.StatusConflict, apperror.CodeConflict, message)
}

// RateLimited sends a 429 Too Many Requests response.
func RateLimited(c *fiber.Ctx, message string) error {
	return Error(c, http.StatusTooManyRequests, apperror.CodeRateLimited, message)
}

// InternalError sends a 500 Internal Server Error response.
// The actual error is not exposed to clients for security.
func InternalError(c *fiber.Ctx) error {
	return Error(c, http.StatusInternalServerError, apperror.CodeInternalError, "An unexpected error occurred")
}

// FromAppError sends an error response based on AppError.
func FromAppError(c *fiber.Ctx, err *apperror.AppError) error {
	return Error(c, err.Status, err.Code, err.Message)
}
