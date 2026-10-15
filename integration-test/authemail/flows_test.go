package authemail_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"

	authdto "go-boilerplate/internal/dto/auth"
	"go-boilerplate/internal/entity"
	"go-boilerplate/internal/repo"
	"go-boilerplate/internal/repo/persistent"
	"go-boilerplate/internal/usecase/auth"
	"go-boilerplate/internal/worker/tasks"
	"go-boilerplate/pkg/audit"
	"go-boilerplate/pkg/response"
	"go-boilerplate/pkg/tx"
)

type failingRefreshRepo struct{ repo.RefreshTokenRepo }

func (f failingRefreshRepo) DeleteByUserID(context.Context, uint) error {
	return errors.New("controlled refresh-token deletion failure")
}

func authRequest(t *testing.T, base, path, jwt string, payload any, status int) []byte {
	t.Helper()
	var body []byte
	var err error
	if raw, ok := payload.([]byte); ok {
		body = raw
	} else if payload != nil {
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}
	request, err := http.NewRequestWithContext(t.Context(), "POST", base+"/v1/auth/"+path, bytes.NewReader(body))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "auth-email-e2e")
	if jwt != "" {
		request.Header.Set("Authorization", "Bearer "+jwt)
	}
	result, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	require.NoError(t, err)
	defer result.Body.Close() //nolint:errcheck // test
	body, err = io.ReadAll(result.Body)
	require.NoError(t, err)
	require.Equal(t, status, result.StatusCode)
	var envelope struct {
		Success   bool   `json:"success"`
		RequestID string `json:"request_id"`
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	require.Equal(t, status < 400, envelope.Success)
	require.Equal(t, "auth-email-e2e", envelope.RequestID)
	return body
}

func (f *authFixture) register(t *testing.T, prefix string) authdto.LoginResponse {
	t.Helper()
	address := prefix + uuid.NewString() + "@example.com"
	body := authRequest(t, f.appURL, "register", "", map[string]string{
		"email": address, "password": "initial-password", "name": "Auth Fixture",
	}, 201)
	var result response.Response[authdto.LoginResponse]
	require.NoError(t, json.Unmarshal(body, &result))
	require.NotEmpty(t, result.Data.AccessToken)
	return result.Data
}

func emailToken(t *testing.T, text, path string) string {
	t.Helper()
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "https://") {
			continue
		}
		link, err := url.Parse(line)
		require.NoError(t, err)
		require.Equal(t, "app.example", link.Host)
		require.Equal(t, "/base/"+path, link.Path)
		require.Empty(t, link.RawQuery)
		fragment, err := url.ParseQuery(link.Fragment)
		require.NoError(t, err)
		token := fragment.Get("token")
		require.Len(t, token, 64)
		return token
	}
	t.Fatal("email does not contain the configured frontend link")
	return ""
}

func TestAuthEmailFlows(t *testing.T) {
	f := newAuthFixture(t)
	t.Run("verification transport persistence and replay", func(t *testing.T) {
		user := f.register(t, "verify-")
		authRequest(t, f.appURL, "resend-verification", "", nil, 401)
		authRequest(t, f.appURL, "resend-verification", user.AccessToken, map[string]uint{"user_id": user.User.ID + 1}, 202)
		message := f.mail(t, "Verify your email", user.User.Email)
		token := emailToken(t, message.Text, "verify-email")
		var stored entity.EmailVerification
		require.NoError(t, f.db.Where("user_id = ?", user.User.ID).First(&stored).Error)
		require.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte(token))), stored.Token)
		job := f.task(t, tasks.TypeEmailVerification, "", user.User.ID, asynq.TaskStateCompleted)
		require.NotContains(t, string(job.Payload), token)
		require.Equal(t, 2, job.MaxRetry)
		require.Equal(t, 2*time.Second, job.Timeout)
		authRequest(t, f.appURL, "verify-email", "", map[string]string{"token": token}, 200)
		authRequest(t, f.appURL, "verify-email", "", map[string]string{"token": token}, 400)
		var updated entity.User
		require.NoError(t, f.db.First(&updated, user.User.ID).Error)
		require.NotNil(t, updated.EmailVerifiedAt)
	})
	t.Run("unknown address has identical acceptance and completes without email", func(t *testing.T) {
		user := f.register(t, "known-")
		unknown := "unknown-" + uuid.NewString() + "@example.com"
		knownResult := authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": user.User.Email}, 202)
		unknownResult := authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": unknown}, 202)
		require.JSONEq(t, string(knownResult), string(unknownResult))
		f.task(t, tasks.TypePasswordReset, unknown, 0, asynq.TaskStateCompleted)
		count, err := f.redis.LLen(t.Context(), f.prefix+":Reset your password:"+unknown+":messages").Result()
		require.NoError(t, err)
		require.Zero(t, count)
		f.mail(t, "Reset your password", user.User.Email)
	})
	t.Run("reset password and revoke refresh tokens", func(t *testing.T) {
		user := f.register(t, "reset-")
		authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": user.User.Email}, 202)
		token := emailToken(t, f.mail(t, "Reset your password", user.User.Email).Text, "reset-password")
		var stored entity.PasswordReset
		require.NoError(t, f.db.Where("user_id = ?", user.User.ID).First(&stored).Error)
		require.Equal(t, fmt.Sprintf("%x", sha256.Sum256([]byte(token))), stored.Token)
		payload := map[string]string{"token": token, "new_password": "replacement-password"}
		authRequest(t, f.appURL, "reset-password", "", payload, 200)
		authRequest(t, f.appURL, "reset-password", "", payload, 400)
		authRequest(t, f.appURL, "login", "", map[string]string{"email": user.User.Email, "password": "initial-password"}, 401)
		authRequest(t, f.appURL, "login", "", map[string]string{"email": user.User.Email, "password": "replacement-password"}, 200)
		authRequest(t, f.appURL, "refresh", "", map[string]string{"refresh_token": user.RefreshToken}, 401)
	})
	t.Run("password reset rolls back when refresh-token deletion fails", func(t *testing.T) {
		user := f.register(t, "rollback-")
		authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": user.User.Email}, 202)
		token := emailToken(t, f.mail(t, "Reset your password", user.User.Email).Text, "reset-password")
		uc := auth.New(persistent.NewUserRepo(f.db), nil, failingRefreshRepo{}, nil, audit.NewNoop()).
			WithPasswordReset(persistent.NewPasswordResetRepo(f.db), auth.ResetConfig{}).WithTransactions(tx.New(f.db))
		err := uc.ResetPassword(t.Context(), auth.ResetPasswordInput{Token: token, NewPassword: "replacement-password"})
		require.ErrorContains(t, err, "controlled refresh-token deletion failure")
		var stored entity.PasswordReset
		require.NoError(t, f.db.Where("user_id = ?", user.User.ID).First(&stored).Error)
		require.Nil(t, stored.UsedAt)
		authRequest(t, f.appURL, "login", "", map[string]string{"email": user.User.Email, "password": "initial-password"}, 200)
		authRequest(t, f.appURL, "refresh", "", map[string]string{"refresh_token": user.RefreshToken}, 200)
		authRequest(t, f.appURL, "reset-password", "", map[string]string{"token": token, "new_password": "replacement-password"}, 200)
	})
	t.Run("native validation rejects invalid token and oversized password", func(t *testing.T) {
		authRequest(t, f.appURL, "verify-email", "", map[string]string{"token": "invalid"}, 400)
		authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": "invalid"}, 400)
		authRequest(t, f.appURL, "reset-password", "", map[string]string{"token": strings.Repeat("a", 64), "new_password": strings.Repeat("界", 25)}, 400)
		for _, path := range []string{"verify-email", "reset-password", "VeRiFy-EmAiL/", "ReSeT-PaSsWoRd/"} {
			authRequest(t, f.appURL, path, "", []byte(`{"token":"auth-email-malformed-sensitive","new_password":`), 400)
		}
	})
	t.Run("expired verification and reset tokens", func(t *testing.T) {
		for _, purpose := range []string{"verification", "reset"} {
			user := f.register(t, "expired-")
			if purpose == "verification" {
				authRequest(t, f.appURL, "resend-verification", user.AccessToken, nil, 202)
				token := emailToken(t, f.mail(t, "Verify your email", user.User.Email).Text, "verify-email")
				require.NoError(t, f.db.Model(&entity.EmailVerification{}).Where("user_id = ?", user.User.ID).Update("expires_at", time.Now().Add(-time.Minute)).Error)
				authRequest(t, f.appURL, "verify-email", "", map[string]string{"token": token}, 400)
			} else {
				authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": user.User.Email}, 202)
				token := emailToken(t, f.mail(t, "Reset your password", user.User.Email).Text, "reset-password")
				require.NoError(t, f.db.Model(&entity.PasswordReset{}).Where("user_id = ?", user.User.ID).Update("expires_at", time.Now().Add(-time.Minute)).Error)
				authRequest(t, f.appURL, "reset-password", "", map[string]string{"token": token, "new_password": "replacement-password"}, 400)
			}
		}
	})
	t.Run("concurrent reset consumes a token once", func(t *testing.T) {
		user := f.register(t, "race-")
		authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": user.User.Email}, 202)
		token := emailToken(t, f.mail(t, "Reset your password", user.User.Email).Text, "reset-password")
		payload, err := json.Marshal(map[string]string{"token": token, "new_password": "replacement-password"})
		require.NoError(t, err)
		statuses := make(chan int, 4)
		for range 4 {
			go func() {
				request, requestErr := http.NewRequestWithContext(t.Context(), "POST", f.appURL+"/v1/auth/reset-password", bytes.NewReader(payload))
				if requestErr != nil {
					statuses <- 0
					return
				}
				request.Header.Set("Content-Type", "application/json")
				result, requestErr := (&http.Client{Timeout: 5 * time.Second}).Do(request)
				if requestErr != nil {
					statuses <- 0
					return
				}
				_ = result.Body.Close()
				statuses <- result.StatusCode
			}()
		}
		successes := 0
		for range 4 {
			status := <-statuses
			if status == 200 {
				successes++
			} else {
				require.Equal(t, 400, status)
			}
		}
		require.Equal(t, 1, successes)
	})
	t.Run("delivery retries and exhaustion", func(t *testing.T) {
		for _, prefix := range []string{"retry-", "fail-"} {
			user := f.register(t, prefix)
			authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": user.User.Email}, 202)
			state := asynq.TaskStateCompleted
			if prefix == "fail-" {
				state = asynq.TaskStateArchived
			}
			job := f.task(t, tasks.TypePasswordReset, user.User.Email, 0, state)
			if prefix == "retry-" {
				require.Equal(t, 1, job.Retried)
				token := emailToken(t, f.mail(t, "Reset your password", user.User.Email).Text, "reset-password")
				authRequest(t, f.appURL, "reset-password", "", map[string]string{"token": token, "new_password": "replacement-password"}, 200)
			} else {
				require.Equal(t, 2, job.Retried)
			}
		}
	})
	t.Run("malformed tasks do not retry", func(t *testing.T) {
		job, err := f.client.EnqueueContext(t.Context(), asynq.NewTask(tasks.TypeEmailVerification, []byte("{invalid")), asynq.MaxRetry(2))
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			info, lookupErr := f.inspector.GetTaskInfo("default", job.ID)
			return lookupErr == nil && info.State == asynq.TaskStateArchived && info.Retried == 0
		}, 10*time.Second, 50*time.Millisecond)
	})
	t.Run("unavailable sender and provider-independent token consumption", func(t *testing.T) {
		user := f.register(t, "outage-")
		authRequest(t, f.appURL, "forgot-password", "", map[string]string{"email": user.User.Email}, 202)
		token := emailToken(t, f.mail(t, "Reset your password", user.User.Email).Text, "reset-password")
		unavailable := f.startApp(t, map[string]string{"EMAIL_API_KEY": ""})
		known := authRequest(t, unavailable, "forgot-password", "", map[string]string{"email": user.User.Email}, 503)
		unknown := authRequest(t, unavailable, "forgot-password", "", map[string]string{"email": "unknown@example.com"}, 503)
		require.JSONEq(t, string(known), string(unknown))
		authRequest(t, unavailable, "resend-verification", user.AccessToken, nil, 503)
		authRequest(t, unavailable, "reset-password", "", map[string]string{"token": token, "new_password": "replacement-password"}, 200)
	})
	t.Run("queue failure has identical unavailable responses", func(t *testing.T) {
		user := f.register(t, "queue-outage-")
		unavailable := f.startApp(t, map[string]string{"REDIS_PORT": "1"})
		known := authRequest(t, unavailable, "forgot-password", "", map[string]string{"email": user.User.Email}, 503)
		unknown := authRequest(t, unavailable, "forgot-password", "", map[string]string{"email": "unknown@example.com"}, 503)
		require.JSONEq(t, string(known), string(unknown))
		authRequest(t, unavailable, "resend-verification", user.AccessToken, nil, 503)
	})
	t.Run("disabled verification and development auto verification", func(t *testing.T) {
		user := f.register(t, "config-")
		disabled := f.startApp(t, map[string]string{"EMAIL_VERIFICATION_ENABLED": "false"})
		authRequest(t, disabled, "verify-email", "", map[string]string{"token": strings.Repeat("a", 64)}, 503)
		authRequest(t, disabled, "resend-verification", user.AccessToken, nil, 503)
		auto := f.startApp(t, map[string]string{"EMAIL_VERIFICATION_AUTO_VERIFY": "true", "EMAIL_API_KEY": ""})
		authRequest(t, auto, "resend-verification", user.AccessToken, nil, 202)
		var updated entity.User
		require.NoError(t, f.db.First(&updated, user.User.ID).Error)
		require.NotNil(t, updated.EmailVerifiedAt)
	})
}
