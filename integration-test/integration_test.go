package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

const (
	attempts       = 20
	requestTimeout = 5 * time.Second
)

var (
	httpURL    string
	healthPath string
	basePathV1 string
)

func setupURLs() {
	host := os.Getenv("APP_HOST")
	if host == "" {
		host = "app" // default for docker-compose
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	httpURL = "http://" + host + ":" + port
	healthPath = httpURL + "/healthz"
	basePathV1 = httpURL + "/v1"
}

var errHealthCheck = fmt.Errorf("url %s is not available", healthPath)

func doWebRequestWithTimeout(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

func getHealthCheck(url string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func healthCheck(attempts int) error {
	for attempts > 0 {
		statusCode, err := getHealthCheck(healthPath)
		if err != nil {
			log.Printf("Integration tests: health check error: %v, attempts left: %d", err, attempts)
			time.Sleep(time.Second)
			attempts--
			continue
		}

		if statusCode == http.StatusOK {
			return nil
		}

		log.Printf("Integration tests: url %s returned %d, attempts left: %d", healthPath, statusCode, attempts)
		time.Sleep(time.Second)
		attempts--
	}

	return errHealthCheck
}

func TestMain(m *testing.M) {
	setupURLs()

	err := healthCheck(attempts)
	if err != nil {
		log.Fatalf("Integration tests: httpURL %s is not available: %s", httpURL, err)
	}

	log.Printf("Integration tests: httpURL %s is available", httpURL)

	code := m.Run()
	os.Exit(code)
}

// Package-level variables to store auth state between tests.
var (
	authAccessToken  string
	authRefreshToken string
)

// HTTP POST: /v1/auth/register.
func TestHTTPRegisterV1(t *testing.T) {
	url := basePathV1 + "/auth/register"
	body := `{"email":"integration-test@example.com","password":"testpassword123","name":"Integration Test"}`

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d, got %d: %s", http.StatusCreated, resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Data.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if result.Data.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}

	// Store for subsequent tests
	authAccessToken = result.Data.AccessToken
	authRefreshToken = result.Data.RefreshToken
}

// HTTP POST: /v1/auth/login.
func TestHTTPLoginV1(t *testing.T) {
	url := basePathV1 + "/auth/login"
	body := `{"email":"integration-test@example.com","password":"testpassword123"}`

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d, got %d: %s", http.StatusOK, resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Data.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	// Update tokens from login
	authAccessToken = result.Data.AccessToken
	authRefreshToken = result.Data.RefreshToken
}

// HTTP GET: /v1/auth/me.
func TestHTTPGetCurrentUserV1(t *testing.T) {
	if authAccessToken == "" {
		t.Skip("No access token available — register/login test must run first")
	}

	url := basePathV1 + "/auth/me"

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+authAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d, got %d: %s", http.StatusOK, resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Data.Email != "integration-test@example.com" {
		t.Errorf("Expected email 'integration-test@example.com', got '%s'", result.Data.Email)
	}
}

// HTTP POST: /v1/auth/refresh.
func TestHTTPRefreshTokenV1(t *testing.T) {
	if authRefreshToken == "" {
		t.Skip("No refresh token available — register/login test must run first")
	}

	url := basePathV1 + "/auth/refresh"
	body := `{"refresh_token":"` + authRefreshToken + `"}`

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d, got %d: %s", http.StatusOK, resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Data.AccessToken == "" {
		t.Error("Expected non-empty access token after refresh")
	}

	// Update tokens
	authAccessToken = result.Data.AccessToken
	if result.Data.RefreshToken != "" {
		authRefreshToken = result.Data.RefreshToken
	}
}

// HTTP POST: /v1/auth/logout.
func TestHTTPLogoutV1(t *testing.T) {
	if authAccessToken == "" || authRefreshToken == "" {
		t.Skip("No auth tokens available — register/login test must run first")
	}

	url := basePathV1 + "/auth/logout"
	body := `{"refresh_token":"` + authRefreshToken + `"}`

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status %d, got %d: %s", http.StatusNoContent, resp.StatusCode, string(respBody))
	}
}

// HTTP GET: /v1/auth/me without token — expect 401.
func TestHTTPUnauthorizedV1(t *testing.T) {
	url := basePathV1 + "/auth/me"

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

// HTTP POST: /v1/translation/do-translate.
func TestHTTPDoTranslateV1(t *testing.T) {
	tests := []struct {
		description string
		body        string
		expected    int
	}{
		{
			description: "DoTranslate Success",
			body: `{
				"destination": "en",
				"original": "текст для перевода",
				"source": "auto"
			}`,
			expected: http.StatusOK,
		},
		{
			description: "DoTranslate Success with source",
			body: `{
				"destination": "en",
				"original": "Текст для перевода",
				"source": "ru"
			}`,
			expected: http.StatusOK,
		},
		{
			description: "DoTranslate Fail - missing source",
			body: `{
				"destination": "en",
				"original": "текст для перевода"
			}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			url := basePathV1 + "/translation/do-translate"
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			defer cancel()

			resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, resp.StatusCode)
			}
		})
	}
}

// HTTP GET: /v1/translation/history.
func TestHTTPHistoryV1(t *testing.T) {
	url := basePathV1 + "/translation/history"
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body struct {
		History []struct {
			Source      string `json:"source"`
			Destination string `json:"destination"`
			Original    string `json:"original"`
			Translation string `json:"translation"`
		} `json:"history"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
}
