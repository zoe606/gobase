package auth_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	handler "go-boilerplate/internal/handlers/http/v1/auth"
	autherrors "go-boilerplate/internal/usecase/auth"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
)

func TestAuthEmailHTTP(t *testing.T) {
	t.Parallel()
	token := strings.Repeat("a", 64)
	verifyBody := `{"token":"` + token + `"}`
	resetBody := `{"token":"` + token + `","new_password":"strong-password"}`
	type scenario struct {
		name, path, body, code string
		status                 int
		call, authenticated    bool
		err                    error
	}
	cases := []scenario{
		{name: "verify", path: "verify-email", body: verifyBody, status: 200, call: true},
		{name: "forgot known", path: "forgot-password", body: `{"email":"known@example.com"}`, status: 202, call: true},
		{name: "forgot unknown", path: "forgot-password", body: `{"email":"unknown@example.com"}`, status: 202, call: true},
		{name: "reset", path: "reset-password", body: resetBody, status: 200, call: true},
		{name: "resend needs JWT", path: "resend-verification", status: 401, code: "UNAUTHORIZED"},
		{name: "resend owns identity", path: "resend-verification", body: `{"user_id":99}`, status: 202, call: true, authenticated: true},
		{name: "bad JSON", path: "verify-email", body: `{invalid`, status: 400, code: "INVALID_JSON"},
		{name: "missing token", path: "verify-email", body: `{}`, status: 400, code: "VALIDATION_ERROR"},
		{name: "nonhex token", path: "verify-email", body: `{"token":"` + strings.Repeat("z", 64) + `"}`, status: 400, code: "VALIDATION_ERROR"},
		{name: "bad email", path: "forgot-password", body: `{"email":"invalid"}`, status: 400, code: "VALIDATION_ERROR"},
		{name: "bad reset JSON", path: "reset-password", body: `{invalid`, status: 400, code: "INVALID_JSON"},
		{name: "bad forgot JSON", path: "forgot-password", body: `{invalid`, status: 400, code: "INVALID_JSON"},
		{name: "short password", path: "reset-password", body: `{"token":"` + token + `","new_password":"short"}`, status: 400, code: "VALIDATION_ERROR"},
		{name: "multibyte password", path: "reset-password", body: `{"token":"` + token + `","new_password":"` + strings.Repeat("é", 40) + `"}`, status: 400, code: "VALIDATION_ERROR"},
	}
	for _, err := range []error{autherrors.ErrVerificationNotFound, autherrors.ErrVerificationExpired, autherrors.ErrVerificationUsed, autherrors.ErrEmailAlreadyVerified} {
		cases = append(cases, scenario{name: err.Error(), path: "verify-email", body: verifyBody, status: 400, code: "INVALID_VERIFICATION_TOKEN", call: true, err: err})
	}
	for _, err := range []error{autherrors.ErrResetTokenNotFound, autherrors.ErrResetTokenExpired, autherrors.ErrResetTokenUsed} {
		cases = append(cases, scenario{name: err.Error(), path: "reset-password", body: resetBody, status: 400, code: "INVALID_RESET_TOKEN", call: true, err: err})
	}
	for _, unavailable := range []struct {
		err  error
		code string
	}{
		{autherrors.ErrVerificationDisabled, "EMAIL_VERIFICATION_DISABLED"},
		{autherrors.ErrEmailUnavailable, "AUTH_EMAIL_UNAVAILABLE"},
		{errors.New("private database failure"), "INTERNAL_ERROR"},
	} {
		status := 503
		if unavailable.code == "INTERNAL_ERROR" {
			status = 500
		}
		cases = append(cases, scenario{name: unavailable.code, path: "verify-email", body: verifyBody, status: status, code: unavailable.code, call: true, err: unavailable.err})
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := NewMockAuth(gomock.NewController(t))
			service := jwt.New("auth-email-test-secret", time.Minute, time.Hour)
			access, _, err := service.GenerateAccessToken(7, "user@example.com", "user", nil)
			require.NoError(t, err)
			if tt.call {
				switch tt.path {
				case "verify-email":
					uc.EXPECT().VerifyEmail(gomock.Any(), token).Return(tt.err)
				case "forgot-password":
					var request struct {
						Email string `json:"email"`
					}
					require.NoError(t, json.Unmarshal([]byte(tt.body), &request))
					uc.EXPECT().RequestPasswordReset(gomock.Any(), request.Email).Return(tt.err)
				case "reset-password":
					uc.EXPECT().ResetPassword(gomock.Any(), autherrors.ResetPasswordInput{Token: token, NewPassword: "strong-password"}).Return(tt.err)
				case "resend-verification":
					uc.EXPECT().SendVerificationEmail(gomock.Any(), uint(7)).Return(tt.err)
				}
			}
			app := fiber.New()
			app.Use(requestid.New())
			handler.New(uc, service, logger.NewDevelopment()).RegisterRoutes(app.Group("/v1"))
			request := httptest.NewRequestWithContext(t.Context(), "POST", "/v1/auth/"+tt.path, bytes.NewBufferString(tt.body))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-Request-ID", "auth-email-contract")
			if tt.authenticated {
				request.Header.Set("Authorization", "Bearer "+access)
			}
			result, err := app.Test(request)
			require.NoError(t, err)
			defer result.Body.Close() //nolint:errcheck // test
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			require.Equal(t, tt.status, result.StatusCode, string(body))
			require.NotContains(t, string(body), "private database failure")
			var envelope struct {
				Success   bool   `json:"success"`
				RequestID string `json:"request_id"`
				Error     struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			require.NoError(t, json.Unmarshal(body, &envelope))
			require.Equal(t, tt.status < 400, envelope.Success)
			require.Equal(t, tt.code, envelope.Error.Code)
			require.Equal(t, "auth-email-contract", envelope.RequestID)
		})
	}
}
