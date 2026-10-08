# Minimal Service Output

The `minimal` profile creates an HTTP service without the bundled application features. The `full` profile remains the default. Engine selection remains independent: each profile supports Gin, stdlib, and Fiber.

## File and Dependency Set

| Files | Responsibility |
|-------|----------------|
| `cmd/app`, `internal/app` | Configuration loading, HTTP startup, signals, and graceful shutdown |
| `config` | Application name, environment, HTTP settings, and logging settings |
| `internal/handlers/http` | Native engine routes, request IDs, recovery, and health checks |
| `pkg/httpserver` | Selected native router and HTTP server lifecycle |
| `pkg/response`, `pkg/apperror`, `pkg/json` | Existing response envelope, HTTP errors, and JSON encoding |
| `internal/handlers/http/v1/helper.go` | Existing field validation messages |
| `pkg/request` in Gin and stdlib | Existing request parsing helpers; Fiber uses its native context |
| `pkg/logger` | Existing Zap logger |
| `pkg/project`, `.gobase.json` | Creation profile and engine metadata |
| `pkg/codegen/cmd/codegen`, `pkg/codegen/cmd/wire` | Explicit errors for unsupported persistence generation |
| `Makefile`, Docker files, README, contribution guide, CI | Commands and documentation for this output |

Runtime dependencies are the selected engine, Viper, Zap, the existing JSON codec, and validation helpers. Test dependencies use the existing Testify version. Vulnerability scanning uses the pinned Go tool. Module versions come from the template's existing dependency pins. Gin retains its existing Gin and QUIC pins.

Minimal output omits auth, profiles, articles, media, translation, repositories, entities, migrations, Swagger, telemetry exporters, storage, queue clients, email delivery, and workers. It does not contain PostgreSQL, Redis, Asynq, GORM, MinIO, or email provider dependencies. Its Compose setup starts only the app.

## Creation

```bash
make init PROFILE=minimal ENGINE=gin MODULE=example.com/catalog APP_NAME=catalog OUTPUT=../catalog
```

Use `ENGINE=stdlib` or `ENGINE=fiber` for the other engines. Omitting `PROFILE` preserves the full application output. Existing `.gobase.json` files without a profile are treated as full applications. Projects without settings retain the existing Fiber and full-profile defaults.

The profile is selected during creation. Editing `.gobase.json` does not convert an existing application.

## HTTP and Startup Contract

- Configuration files are optional for direct startup. Environment variables and defaults work without external infrastructure.
- `/healthz` returns HTTP 200 and `OK` while the HTTP server is running.
- `/readyz` returns HTTP 200 after configuration and router initialization. There are no database or Redis probes because these dependencies are omitted.
- Request and response helpers keep the selected engine's API and the existing generic response format.
- Request IDs and recovery use the selected native integration.
- SIGINT and SIGTERM trigger bounded HTTP shutdown. Startup and shutdown errors return a nonzero process exit code.
- Docker uses the existing non-root UID and GID `65532`, readable configuration and CA certificates, and a writable temporary directory.

## Code Generation

Minimal output supports handwritten native HTTP endpoints and the included parsing, validation, and response helpers. SQL migration-based CRUD generation and automatic persistence wiring are not supported.

`make gen`, `make gen-entity`, `make gen-full`, and `make wire` fail with a message that persistence generation requires the full profile. The corresponding Go command paths return the same explanation. They do not write partial application files.

Adding persistence, conversion between profiles, feature composition, workspace generation, and service communication are separate work. See the [roadmap](roadmap.md).

## Verification

Verify fresh output for each engine without running PostgreSQL, Redis, storage, email services, or a worker. Check module contents, builds, race tests, valid and invalid HTTP requests, health probes, process shutdown, unsupported generation errors, and the non-root Docker image. The existing full-profile verification remains required.
