# HTTP Engine Verification

Local verification completed on October 5, 2026, using Go 1.27.1 on macOS arm64. Integration tests used isolated PostgreSQL 17 and Redis 7 instances.

| Check | Gin | Standard net/http | Fiber |
|-------|-----|-------------------|-------|
| Generate project and resolve modules | Passed | Passed | Passed |
| Build all packages | Passed | Passed | Passed |
| Race tests for `internal` and `pkg` | Passed | Passed | Passed |
| Shared HTTP contract tests | Passed | Passed | Passed |
| Generate and wire a new CRUD feature | Passed | Passed | Passed |
| Real HTTP CRUD for the generated feature | Passed | Passed | Passed |
| Generated use case and HTTP tests | Passed | Passed | Passed |
| golangci-lint v2.14.0 after code generation | Passed | Passed | Passed |
| Auth, profile, media, and article integration tests | Passed | Passed | Passed |
| Translation and history integration tests | Passed | Passed | Passed |
| Linux amd64 app and worker builds with CGO disabled | Passed | Passed | Passed |
| Reachable vulnerabilities reported by govulncheck | 0 | 0 | 0 |

The canonical checkout also passed `go test -count=1 ./...` with a running application and a fresh integration database. `make check-all` passed with 93% coverage under the repository's configured coverage exclusions.

Runtime dependency checks confirmed that Gin and standard HTTP outputs do not import Fiber or fasthttp. The standard HTTP output also does not import Gin. Project generation reads independent engine templates. A regression test confirms that it succeeds when checkout handlers contain invalid source. CRUD generation also uses independent templates for each engine.

The shared HTTP suite checks response envelopes, status codes, routing, JSON and form parsing, multipart uploads, JWT authentication, pagination, CORS, compression, body limits, rate limiting, health probes, Swagger, metrics, and idempotent response replay. Replay checks include separate users and expired tokens.

Gin handlers use `*gin.Context`. Standard HTTP handlers use `http.ResponseWriter` and `*http.Request`. Fiber handlers use `*fiber.Ctx`. Gin writer tests also cover middleware response capture, `CloseNotify`, and `Hijack` through standard HTTP middleware. Each engine passed all ten integration tests, including the generated feature, existing CRUD, auth, translation, and history.

## Repeat the Checks

Install golangci-lint v2.14.0 or a newer version that supports Go 1.27. Run these commands from the template checkout:

```bash
make check-all
go run ./pkg/tools/verifyengines -lint
```

The Engines workflow repeats generation, builds, tests, code generation, lint, dependency scans, and integration tests for each engine. Its GitHub Actions result was not part of this local verification.
