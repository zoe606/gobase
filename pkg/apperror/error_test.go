package apperror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"go-boilerplate/pkg/apperror"
)

func TestAppError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		appErr   *apperror.AppError
		contains string
	}{
		{
			name:     "without underlying error",
			appErr:   apperror.New("TEST_ERROR", "Test message", http.StatusBadRequest),
			contains: "TEST_ERROR: Test message",
		},
		{
			name:     "with underlying error",
			appErr:   apperror.ErrInternal.WithError(errors.New("db connection failed")),
			contains: "db connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.appErr.Error()
			require.Contains(t, result, tt.contains)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("original error")
	appErr := apperror.ErrInternal.WithError(originalErr)

	unwrapped := appErr.Unwrap()
	require.Equal(t, originalErr, unwrapped)
}

func TestAppError_WithError(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("database error")
	appErr := apperror.ErrInternal.WithError(originalErr)

	require.Equal(t, apperror.ErrInternal.Code, appErr.Code)
	require.Equal(t, apperror.ErrInternal.Message, appErr.Message)
	require.Equal(t, apperror.ErrInternal.Status, appErr.Status)
	require.Equal(t, originalErr, appErr.Err)
}

func TestAppError_WithMessage(t *testing.T) {
	t.Parallel()

	customMsg := "Custom error message"
	appErr := apperror.ErrBadRequest.WithMessage(customMsg)

	require.Equal(t, apperror.ErrBadRequest.Code, appErr.Code)
	require.Equal(t, customMsg, appErr.Message)
	require.Equal(t, apperror.ErrBadRequest.Status, appErr.Status)
}

func TestNew(t *testing.T) {
	t.Parallel()

	appErr := apperror.New("CUSTOM_ERROR", "Custom message", http.StatusTeapot)

	require.Equal(t, "CUSTOM_ERROR", appErr.Code)
	require.Equal(t, "Custom message", appErr.Message)
	require.Equal(t, http.StatusTeapot, appErr.Status)
	require.Nil(t, appErr.Err)
}

func TestIsAppError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		isApp  bool
		code   string
	}{
		{
			name:  "is AppError",
			err:   apperror.ErrNotFound,
			isApp: true,
			code:  apperror.CodeNotFound,
		},
		{
			name:  "wrapped AppError",
			err:   fmt.Errorf("wrapped: %w", apperror.ErrUnauthorized),
			isApp: true,
			code:  apperror.CodeUnauthorized,
		},
		{
			name:  "not AppError",
			err:   errors.New("regular error"),
			isApp: false,
		},
		{
			name:  "nil error",
			err:   nil,
			isApp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			appErr, ok := apperror.IsAppError(tt.err)

			require.Equal(t, tt.isApp, ok)
			if tt.isApp {
				require.NotNil(t, appErr)
				require.Equal(t, tt.code, appErr.Code)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    *apperror.AppError
		status int
		code   string
	}{
		{"BadRequest", apperror.ErrBadRequest, http.StatusBadRequest, apperror.CodeBadRequest},
		{"Validation", apperror.ErrValidation, http.StatusBadRequest, apperror.CodeValidationError},
		{"Unauthorized", apperror.ErrUnauthorized, http.StatusUnauthorized, apperror.CodeUnauthorized},
		{"Forbidden", apperror.ErrForbidden, http.StatusForbidden, apperror.CodeForbidden},
		{"NotFound", apperror.ErrNotFound, http.StatusNotFound, apperror.CodeNotFound},
		{"Conflict", apperror.ErrConflict, http.StatusConflict, apperror.CodeConflict},
		{"RateLimited", apperror.ErrRateLimited, http.StatusTooManyRequests, apperror.CodeRateLimited},
		{"Internal", apperror.ErrInternal, http.StatusInternalServerError, apperror.CodeInternalError},
		{"ServiceUnavailable", apperror.ErrServiceUnavailable, http.StatusServiceUnavailable, apperror.CodeServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.status, tt.err.Status)
			require.Equal(t, tt.code, tt.err.Code)
			require.NotEmpty(t, tt.err.Message)
		})
	}
}
