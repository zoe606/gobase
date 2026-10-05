package httphandler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go-boilerplate/config"
	articledto "go-boilerplate/internal/dto/article"
	authdto "go-boilerplate/internal/dto/auth"
	mediadto "go-boilerplate/internal/dto/media"
	profiledto "go-boilerplate/internal/dto/profile"
	translationdto "go-boilerplate/internal/dto/translation"
	"go-boilerplate/internal/usecase"
	articleuc "go-boilerplate/internal/usecase/article"
	"go-boilerplate/pkg/cache"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/pagination"
)

type contractHealth struct{ err error }

func (h contractHealth) Ping() error { return h.err }

type contractAuth struct {
	usecase.Auth
	t *testing.T
}

func (u contractAuth) Login(ctx context.Context, req authdto.LoginRequest) (*authdto.LoginResponse, error) {
	require.Equal(u.t, "user@example.com", req.Email)
	require.Equal(u.t, "password123", req.Password)
	_, ok := ctx.Deadline()
	require.True(u.t, ok)
	return &authdto.LoginResponse{AccessToken: "access", RefreshToken: "refresh", User: authdto.UserResponse{ID: 7, Email: req.Email}}, nil
}
func (u contractAuth) GetCurrentUser(_ context.Context, id uint) (*authdto.UserResponse, error) {
	require.Equal(u.t, uint(7), id)
	return &authdto.UserResponse{ID: id, Email: "user@example.com"}, nil
}

type contractArticle struct {
	usecase.Article
	t *testing.T
}

func (u contractArticle) GetByID(_ context.Context, id uint) (*articledto.Response, error) {
	if id == 404 {
		return nil, articleuc.ErrNotFound
	}
	if id == 500 {
		return nil, errors.New("database unavailable")
	}
	return &articledto.Response{ID: id, Title: "Article"}, nil
}
func (u contractArticle) List(_ context.Context, req articledto.ListRequest) (*articledto.ListResponse, error) {
	require.Equal(u.t, 3, req.Page)
	require.Equal(u.t, 5, req.Limit)
	require.Equal(u.t, uint(7), req.UserID)
	require.Equal(u.t, "draft", req.Status)
	return &articledto.ListResponse{Data: []*articledto.Response{{ID: 9, Title: "Article"}}, Meta: pagination.NewMeta(req.Page, req.Limit, 21)}, nil
}
func (u contractArticle) Create(_ context.Context, id uint, req articledto.CreateRequest) (*articledto.Response, error) {
	require.Equal(u.t, uint(7), id)
	require.Equal(u.t, "Article", req.Title)
	require.Equal(u.t, uint(1), req.CoverMediaID)
	return &articledto.Response{ID: 9, UserID: id, Title: req.Title}, nil
}
func (u contractArticle) Update(_ context.Context, userID, id uint, req articledto.UpdateRequest) (*articledto.Response, error) {
	require.Equal(u.t, uint(7), userID)
	if id == 403 {
		return nil, articleuc.ErrForbidden
	}
	require.NotNil(u.t, req.Title)
	require.Equal(u.t, "Updated", *req.Title)
	return &articledto.Response{ID: id, UserID: userID, Title: *req.Title}, nil
}
func (u contractArticle) Delete(_ context.Context, userID, id uint) error {
	require.Equal(u.t, uint(7), userID)
	require.Equal(u.t, uint(9), id)
	return nil
}

type contractProfile struct {
	usecase.Profile
	t *testing.T
}

func (u contractProfile) GetProfile(_ context.Context, id uint) (*profiledto.ProfileResponse, error) {
	require.Equal(u.t, uint(7), id)
	return &profiledto.ProfileResponse{UserID: id}, nil
}
func (u contractProfile) UpdateProfile(_ context.Context, id uint, req profiledto.UpdateProfileRequest) (*profiledto.ProfileResponse, error) {
	require.Equal(u.t, uint(7), id)
	require.NotNil(u.t, req.Bio)
	require.Equal(u.t, "hello", *req.Bio)
	return &profiledto.ProfileResponse{UserID: id, Bio: *req.Bio}, nil
}

type contractTranslation struct{ usecase.Translation }

func (contractTranslation) Translate(_ context.Context, req translationdto.TranslateRequest) (*translationdto.TranslationResponse, error) {
	return &translationdto.TranslationResponse{Source: req.Source, Destination: req.Destination, Original: req.Original, Translation: "hello"}, nil
}
func (contractTranslation) History(_ context.Context, _ translationdto.HistoryRequest) (*translationdto.HistoryResponse, error) {
	return &translationdto.HistoryResponse{Items: []translationdto.TranslationResponse{}, Meta: pagination.NewMeta(1, 20, 0)}, nil
}

type contractMedia struct {
	usecase.Media
	t *testing.T
}

func (u contractMedia) Upload(_ context.Context, req mediadto.UploadRequest) (*mediadto.MediaResponse, error) {
	require.Equal(u.t, "users", req.AttachableType)
	require.Equal(u.t, uint(7), req.AttachableID)
	require.Equal(u.t, "avatar", req.Collection)
	require.Equal(u.t, "hello.txt", req.Filename)
	data, err := io.ReadAll(req.File)
	require.NoError(u.t, err)
	require.Equal(u.t, "upload content", string(data))
	return &mediadto.MediaResponse{ID: 4, Filename: req.Filename}, nil
}

func contractConfig() *config.Config {
	return &config.Config{
		App:       config.App{Name: "contract"},
		HTTP:      config.HTTP{BodyLimit: 4 << 20, RequestTimeout: time.Second},
		CORS:      config.CORS{AllowOrigins: "https://example.com", AllowMethods: "GET,POST,PUT,DELETE,OPTIONS", AllowHeaders: "Authorization,Content-Type", MaxAge: 600},
		RateLimit: config.RateLimit{Max: 1000, Expiration: time.Minute},
		Storage:   config.Storage{MaxSize: 1 << 20},
	}
}

func contractApp(t *testing.T, cfg *config.Config, checker HealthChecker) (httpHandler http.Handler, accessToken string) {
	t.Helper()
	service := jwt.New("contract-secret", time.Hour, time.Hour)
	token, _, err := service.GenerateAccessToken(7, "user@example.com", "user", nil)
	require.NoError(t, err)
	handler := contractHandler(t, cfg, contractTranslation{}, contractAuth{t: t}, contractMedia{t: t}, contractProfile{t: t}, contractArticle{t: t}, service, checker, nil)
	return handler, token
}

func contractProxy(t *testing.T, baseURL string) http.Handler {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequestWithContext(r.Context(), r.Method, baseURL+r.URL.RequestURI(), r.Body)
		require.NoError(t, err)
		req.Header = r.Header.Clone()
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		require.NoError(t, err)
	})
}

func TestEngineAPIContract(t *testing.T) {
	handler, token := contractApp(t, contractConfig(), contractHealth{})
	tests := []struct {
		method, path, body string
		auth               bool
		status             int
		contains           string
	}{
		{"GET", "/healthz", "", false, 200, "OK"},
		{"GET", "/readyz", "", false, 200, "OK"},
		{"GET", "/v1/articles/9", "", false, 200, "\"id\":9"},
		{"GET", "/V1/ARTICLES/9/", "", false, 200, "\"id\":9"},
		{"GET", "/v1/articles/bad", "", false, 400, "INVALID_ID"},
		{"GET", "/v1/articles/404", "", false, 404, "NOT_FOUND"},
		{"GET", "/v1/articles/500", "", false, 500, "INTERNAL_ERROR"},
		{"GET", "/v1/articles?page=3&limit=5&user_id=7&status=draft", "", false, 200, "\"page\":3"},
		{"GET", "/v1/articles?user_id=bad", "", false, 400, "INVALID_QUERY"},
		{"POST", "/v1/articles", "{}", false, 401, "UNAUTHORIZED"},
		{"POST", "/v1/articles", "{", true, 400, "INVALID_JSON"},
		{"POST", "/v1/articles", "{}", true, 400, "VALIDATION_ERROR"},
		{"POST", "/v1/articles", "{\"title\":\"Article\",\"slug\":\"article\",\"content\":\"content\",\"excerpt\":\"excerpt\",\"cover_media_id\":1}", true, 201, "\"user_id\":7"},
		{"PUT", "/v1/articles/9", "{\"title\":\"Updated\"}", true, 200, "\"title\":\"Updated\""},
		{"PUT", "/v1/articles/403", "{\"title\":\"Updated\"}", true, 403, "FORBIDDEN"},
		{"DELETE", "/v1/articles/9", "", true, 204, ""},
		{"POST", "/v1/auth/login", "{\"email\":\"user@example.com\",\"password\":\"password123\"}", false, 200, "\"access_token\":\"access\""},
		{"POST", "/v1/auth/login", "{}", false, 400, "VALIDATION_ERROR"},
		{"GET", "/v1/auth/me", "", true, 200, "\"id\":7"},
		{"GET", "/v1/auth/me", "", false, 401, "Missing authorization header"},
		{"GET", "/v1/profile", "", true, 200, "\"user_id\":7"},
		{"PATCH", "/v1/profile", "{\"bio\":\"hello\"}", true, 200, "\"bio\":\"hello\""},
		{"POST", "/v1/translation/do-translate", "{\"source\":\"en\",\"destination\":\"id\",\"original\":\"hello\"}", false, 200, "hello"},
		{"GET", "/v1/translation/history", "", false, 200, "meta"},
		{"GET", "/unknown", "", false, 404, "Cannot GET /unknown"},
		{"PATCH", "/healthz", "", false, 405, "Method Not Allowed"},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path+" "+tt.body, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Request-ID", "contract-request")
			if tt.auth {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			require.Equal(t, tt.status, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), tt.contains)
			require.Equal(t, "contract-request", recorder.Header().Get("X-Request-ID"))
			if strings.HasPrefix(recorder.Header().Get("Content-Type"), "application/json") {
				require.True(t, json.Valid(recorder.Body.Bytes()), recorder.Body.String())
			}
			if tt.status == 204 {
				require.Empty(t, recorder.Body.String())
			}
		})
	}
	t.Run("HEAD", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodHead, "/v1/articles/9", http.NoBody)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, 200, recorder.Code)
		require.Empty(t, recorder.Body.String())
	})
	t.Run("CORS preflight", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodOptions, "/v1/articles", http.NoBody)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, 204, recorder.Code)
		require.Equal(t, "https://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
	})
	t.Run("form login", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), "POST", "/v1/auth/login", strings.NewReader("email=user%40example.com&password=password123"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, 200, recorder.Code, recorder.Body.String())
		require.Contains(t, recorder.Body.String(), "access_token")
	})
	t.Run("form profile update", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), "PATCH", "/v1/profile", strings.NewReader("bio=hello"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, 200, recorder.Code, recorder.Body.String())
		require.Contains(t, recorder.Body.String(), "hello")
	})
	t.Run("gzip response", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), "GET", "/v1/articles/9", http.NoBody)
		req.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, 200, recorder.Code)
		if recorder.Header().Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(recorder.Body)
			require.NoError(t, err)
			defer reader.Close()
			body, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Contains(t, string(body), "Article")
		} else {
			require.Contains(t, recorder.Body.String(), "Article")
		}
	})
	t.Run("multipart upload", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		file, err := writer.CreateFormFile("file", "hello.txt")
		require.NoError(t, err)
		_, err = file.Write([]byte("upload content"))
		require.NoError(t, err)
		require.NoError(t, writer.WriteField("attachable_type", "users"))
		require.NoError(t, writer.WriteField("attachable_id", "7"))
		require.NoError(t, writer.WriteField("collection", "avatar"))
		require.NoError(t, writer.Close())
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/media/upload", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, 201, recorder.Code, recorder.Body.String())
		require.Contains(t, recorder.Body.String(), "hello.txt")
	})
}

type contractCache struct {
	cache.Cache
	mu     sync.Mutex
	values map[string][]byte
}

func (c *contractCache) Get(_ context.Context, key string, dest any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, ok := c.values[key]
	if !ok {
		return cache.ErrNotFound
	}
	return json.Unmarshal(data, dest)
}
func (c *contractCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = data
	return nil
}

type contractTokenService struct {
	jwt.Service
	expired atomic.Bool
}

func (s *contractTokenService) ValidateToken(token string) (*jwt.Claims, error) {
	if s.expired.Load() {
		return nil, jwt.ErrExpiredToken
	}
	return s.Service.ValidateToken(token)
}

type contractIdempotentArticle struct {
	usecase.Article
	calls atomic.Int32
}

func (u *contractIdempotentArticle) Create(_ context.Context, id uint, req articledto.CreateRequest) (*articledto.Response, error) {
	u.calls.Add(1)
	return &articledto.Response{ID: 9, UserID: id, Title: req.Title}, nil
}

func TestEngineIdempotencyContract(t *testing.T) {
	cfg := contractConfig()
	cfg.Idempotency = config.Idempotency{Enabled: true, TTL: time.Hour, RequiredForPost: true}
	service := &contractTokenService{Service: jwt.New("contract-secret", time.Hour, time.Hour)}
	appCache := &contractCache{Cache: cache.NewNoop(), values: map[string][]byte{}}
	article := &contractIdempotentArticle{}
	handler := contractHandler(t, cfg, contractTranslation{}, contractAuth{t: t}, contractMedia{t: t}, contractProfile{t: t}, article, service, contractHealth{}, appCache)
	token, _, err := service.GenerateAccessToken(7, "user@example.com", "user", nil)
	require.NoError(t, err)
	body := "{\"title\":\"Article\",\"slug\":\"article\",\"content\":\"content\",\"excerpt\":\"excerpt\",\"cover_media_id\":1}"
	request := func(token, key string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), "POST", "/v1/articles", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Idempotency-Key", key)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return recorder
	}
	require.Equal(t, 400, request(token, "").Code)
	first := request(token, "create-article")
	require.Equal(t, 201, first.Code, first.Body.String())
	replay := request(token, "create-article")
	require.Equal(t, 201, replay.Code)
	require.Equal(t, "true", replay.Header().Get("X-Idempotent-Replay"))
	require.Equal(t, first.Body.String(), replay.Body.String())
	require.Equal(t, int32(1), article.calls.Load())
	other, _, err := service.GenerateAccessToken(8, "other@example.com", "user", nil)
	require.NoError(t, err)
	second := request(other, "create-article")
	require.Equal(t, 201, second.Code)
	require.Contains(t, second.Body.String(), "\"user_id\":8")
	require.Empty(t, second.Header().Get("X-Idempotent-Replay"))
	require.Equal(t, int32(2), article.calls.Load())
	service.expired.Store(true)
	expired := request(token, "create-article")
	require.Equal(t, 401, expired.Code)
	require.Empty(t, expired.Header().Get("X-Idempotent-Replay"))
}

func TestEngineMiddlewareContract(t *testing.T) {
	t.Run("unhealthy dependency", func(t *testing.T) {
		handler, _ := contractApp(t, contractConfig(), contractHealth{err: errors.New("database unavailable")})
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), "GET", "/readyz", http.NoBody))
		require.Equal(t, 503, recorder.Code)
		require.JSONEq(t, "{\"status\":\"unhealthy\",\"error\":\"database unavailable\"}", recorder.Body.String())
	})
	t.Run("rate limit", func(t *testing.T) {
		cfg := contractConfig()
		cfg.RateLimit.Max = 1
		handler, _ := contractApp(t, cfg, contractHealth{})
		for _, status := range []int{200, 429} {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), "GET", "/healthz", http.NoBody))
			require.Equal(t, status, recorder.Code, recorder.Body.String())
			if status == 429 {
				require.Contains(t, recorder.Body.String(), "RATE_LIMITED")
				require.NotEmpty(t, recorder.Header().Get("Retry-After"))
			}
		}
	})
	t.Run("body limit", func(t *testing.T) {
		cfg := contractConfig()
		cfg.HTTP.BodyLimit = 32
		handler, _ := contractApp(t, cfg, contractHealth{})
		req := httptest.NewRequestWithContext(t.Context(), "POST", "/v1/auth/login", strings.NewReader(strings.Repeat("x", 64)))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, 413, recorder.Code)
	})
	t.Run("swagger and metrics", func(t *testing.T) {
		cfg := contractConfig()
		cfg.Swagger.Enabled = true
		cfg.Metrics.Enabled = true
		handler, _ := contractApp(t, cfg, contractHealth{})
		for _, path := range []string{"/swagger/index.html", "/swagger/doc.json", "/swagger/swagger-ui.css", "/metrics"} {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), "GET", path, http.NoBody))
			require.Equal(t, 200, recorder.Code, recorder.Body.String())
			require.NotEmpty(t, recorder.Body.String())
		}
	})
}
