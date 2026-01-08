// Package response provides standardized HTTP response structures.
// All API responses follow a consistent format for success and error cases.
package response

// Response is the standard success response structure.
// Uses generics for type-safe data field.
type Response[T any] struct {
	Success   bool   `json:"success"`
	Data      T      `json:"data,omitempty"`
	Meta      *Meta  `json:"meta,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorResponse is the standard error response structure.
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}
