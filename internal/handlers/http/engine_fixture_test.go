package httphandler

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"go-boilerplate/config"
	articlehandler "go-boilerplate/internal/handlers/http/v1/article"
	"go-boilerplate/internal/handlers/http/v1/auth"
	confighandler "go-boilerplate/internal/handlers/http/v1/config"
	mediahandler "go-boilerplate/internal/handlers/http/v1/media"
	profilehandler "go-boilerplate/internal/handlers/http/v1/profile"
	"go-boilerplate/internal/handlers/http/v1/translation"
	"go-boilerplate/internal/usecase"
	"go-boilerplate/pkg/cache"
	"go-boilerplate/pkg/jwt"
	"go-boilerplate/pkg/logger"
	"go-boilerplate/pkg/ratelimiter"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func contractHandler(t *testing.T, cfg *config.Config, translationUC usecase.Translation, authUC usecase.Auth, mediaUC usecase.Media, profileUC usecase.Profile, articleUC usecase.Article, service jwt.Service, checker HealthChecker, appCache cache.Cache, stores ...ratelimiter.Storage) http.Handler {
	t.Helper()
	app := fiber.New(fiber.Config{BodyLimit: cfg.HTTP.BodyLimit})
	l := logger.New("error")
	var store ratelimiter.Storage
	if len(stores) > 0 {
		store = stores[0]
	}
	setupMiddleware(app, cfg, l, store, appCache, service)
	setupOptionalFeatures(app, cfg)
	setupHealthEndpoints(app, checker)
	group := app.Group("/v1")
	translation.New(translationUC, l).RegisterRoutes(group)
	auth.New(authUC, service, l).RegisterRoutes(group)
	confighandler.New(cfg, service, l).RegisterRoutes(group)
	mediahandler.New(mediaUC, service, l, cfg.Storage.MaxSize).RegisterRoutes(group)
	profilehandler.New(profileUC, service, l).RegisterRoutes(group)
	articlehandler.New(articleUC, service, l).RegisterRoutes(group)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = app.Listener(listener) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
		_ = listener.Close()
	})
	return contractProxy(t, "http://"+listener.Addr().String())
}
