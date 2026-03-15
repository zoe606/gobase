package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-boilerplate/config"
	"go-boilerplate/internal/handlers/http/middleware"
	"go-boilerplate/pkg/cache"
	"go-boilerplate/pkg/json"
)

// mockIdempotencyCache is an in-memory cache for testing idempotency middleware.
type mockIdempotencyCache struct {
	mu    sync.Mutex
	store map[string][]byte
}

func newMockIdempotencyCache() *mockIdempotencyCache {
	return &mockIdempotencyCache{store: make(map[string][]byte)}
}

func (m *mockIdempotencyCache) Get(_ context.Context, key string, dest interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.store[key]
	if !ok {
		return cache.ErrNotFound
	}

	return json.Unmarshal(data, dest)
}

func (m *mockIdempotencyCache) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.store[key] = data

	return nil
}

func (m *mockIdempotencyCache) Delete(_ context.Context, _ string) error { return nil }

func (m *mockIdempotencyCache) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *mockIdempotencyCache) DeleteByPrefix(_ context.Context, _ string) error { return nil }

func (m *mockIdempotencyCache) Remember(_ context.Context, _ string, _ time.Duration, _ interface{}, fn func() (interface{}, error)) error {
	_, err := fn()
	return err
}

func TestIdempotency_SkipsGET(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: false,
	}

	app := fiber.New()
	app.Use(middleware.Idempotency(mc, cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
	req.Header.Set("Idempotency-Key", "test-key-123")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Cache should be empty — GET requests bypass idempotency.
	mc.mu.Lock()
	assert.Empty(t, mc.store)
	mc.mu.Unlock()
}

func TestIdempotency_CachesAndReplaysPOST(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: false,
	}

	callCount := 0

	app := fiber.New()
	app.Use(middleware.Idempotency(mc, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		callCount++
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 42})
	})

	// First request — should hit the handler.
	req1 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req1.Header.Set("Idempotency-Key", "unique-key-1")

	resp1, err := app.Test(req1)
	require.NoError(t, err)
	defer resp1.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusCreated, resp1.StatusCode)
	assert.Equal(t, 1, callCount)
	assert.Empty(t, resp1.Header.Get("X-Idempotent-Replay"))

	// Second request with same key — should replay from cache.
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req2.Header.Set("Idempotency-Key", "unique-key-1")

	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusCreated, resp2.StatusCode)
	assert.Equal(t, 1, callCount, "handler should not be called again")
	assert.Equal(t, "true", resp2.Header.Get("X-Idempotent-Replay"))
}

func TestIdempotency_RequiredForPOST(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: true,
	}

	app := fiber.New()
	app.Use(middleware.Idempotency(mc, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
	})

	// POST without Idempotency-Key should return 400.
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIdempotency_NotRequiredAllowsThrough(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: false,
	}

	app := fiber.New()
	app.Use(middleware.Idempotency(mc, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
	})

	// POST without Idempotency-Key should succeed when not required.
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestIdempotency_SkipsDELETE(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: true, // Even with this enabled, DELETE should bypass.
	}

	app := fiber.New()
	app.Use(middleware.Idempotency(mc, cfg))
	app.Delete("/item/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/item/1", http.NoBody)
	req.Header.Set("Idempotency-Key", "delete-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close() //nolint:errcheck // test

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Cache should be empty — DELETE requests bypass idempotency.
	mc.mu.Lock()
	assert.Empty(t, mc.store)
	mc.mu.Unlock()
}

func TestIdempotency_ScopedByEndpoint(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: false,
	}

	app := fiber.New()
	app.Use(middleware.Idempotency(mc, cfg))
	app.Post("/articles", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"type": "article"})
	})
	app.Post("/comments", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"type": "comment"})
	})

	sameKey := "shared-idem-key"

	// Request to /articles
	req1 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/articles", http.NoBody)
	req1.Header.Set("Idempotency-Key", sameKey)

	resp1, err := app.Test(req1)
	require.NoError(t, err)
	defer resp1.Body.Close() //nolint:errcheck // test
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	// Request to /comments with SAME key — must NOT replay the /articles response.
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/comments", http.NoBody)
	req2.Header.Set("Idempotency-Key", sameKey)

	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close() //nolint:errcheck // test
	assert.Equal(t, http.StatusCreated, resp2.StatusCode)
	assert.Empty(t, resp2.Header.Get("X-Idempotent-Replay"), "different endpoint must not replay")
}

func TestIdempotency_ScopedByUser(t *testing.T) {
	t.Parallel()

	mc := newMockIdempotencyCache()
	cfg := config.Idempotency{
		Enabled:         true,
		TTL:             24 * time.Hour,
		RequiredForPost: false,
	}

	app := fiber.New()

	// Simulate JWTAuth by reading a test header and setting Locals
	app.Use(func(c *fiber.Ctx) error {
		if uid := c.Get("X-Test-UserID"); uid != "" {
			id, _ := strconv.ParseUint(uid, 10, 32)
			c.Locals(middleware.UserIDKey, uint(id))
		}
		return c.Next()
	})

	app.Use(middleware.Idempotency(mc, cfg))
	app.Post("/create", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": c.Locals(middleware.UserIDKey)})
	})

	sameKey := "user-idem-key"

	// User 1 creates
	req1 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req1.Header.Set("Idempotency-Key", sameKey)
	req1.Header.Set("X-Test-UserID", "1")

	resp1, err := app.Test(req1)
	require.NoError(t, err)
	defer resp1.Body.Close() //nolint:errcheck // test
	assert.Equal(t, http.StatusCreated, resp1.StatusCode)

	// User 2 sends same key — must NOT get User 1's cached response
	req2 := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/create", http.NoBody)
	req2.Header.Set("Idempotency-Key", sameKey)
	req2.Header.Set("X-Test-UserID", "2")

	resp2, err := app.Test(req2)
	require.NoError(t, err)
	defer resp2.Body.Close() //nolint:errcheck // test
	assert.Equal(t, http.StatusCreated, resp2.StatusCode)
	assert.Empty(t, resp2.Header.Get("X-Idempotent-Replay"), "different user must not replay")
}
