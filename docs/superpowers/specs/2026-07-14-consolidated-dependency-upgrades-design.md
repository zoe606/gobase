# Consolidated Dependency and Toolchain Upgrade — Design Spec

**Date:** 2026-07-14
**Status:** Approved
**Scope:** Replace seven overlapping Dependabot pull requests with one targeted dependency and security refresh

## Goal

Bring the Go toolchain, the dependencies represented by the open Dependabot pull requests, and the vulnerable transitive modules to their current secure versions without performing a broad update of unrelated direct dependencies.

## Current State

The repository has seven open Dependabot pull requests:

| PR | Dependency | Proposed version | Current latest |
|----|------------|------------------|----------------|
| #26 | `github.com/goccy/go-json` | `v0.10.6` | `v0.10.6` |
| #32 | `codecov/codecov-action` | `v7` | `v7` |
| #33 | `github.com/gofiber/fiber/v2` | `v2.52.13` | `v2.52.14` |
| #34 | `google.golang.org/grpc` | `v1.81.1` | `v1.82.0` |
| #35 | `golang.org/x/crypto` | `v0.53.0` | `v0.54.0` |
| #36 | `go.opentelemetry.io/otel/trace` | `v1.44.0` | `v1.44.0` |
| #37 | `actions/checkout` | `v7` | `v7` |

PR #26 is green. PRs #32–#37 pass build, lint, unit tests, and integration tests but fail the shared Security Scan. The failure is caused primarily by the repository's Go 1.26.1 toolchain and vulnerable transitive modules, rather than by the individual Dependabot changes.

The current secure Go release is 1.26.5. The security scan also identifies reachable vulnerabilities in `golang.org/x/net` and `golang.org/x/image`.

## Decision

Create one consolidated upgrade from `origin/master` and supersede the seven Dependabot pull requests after the replacement is verified and merged.

The upgrade is intentionally targeted:

- Update the Go patch release and every dependency represented by the seven open pull requests.
- Use the current available version instead of copying a stale Dependabot version.
- Align all directly required OpenTelemetry modules at the same release.
- Update transitive modules required to remove reachable security findings.
- Allow `go mod tidy` to adjust transitives required by the selected direct versions.
- Do not run a blanket `go get -u ./...` or upgrade unrelated direct dependencies.

## Version Changes

### Toolchain and CI

- Go directive and CI toolchain: `1.26.1` → `1.26.5`
- `actions/checkout`: `v6` → `v7`
- `codecov/codecov-action`: `v5` → `v7`
- README Go badge: align with Go 1.26

The Docker builder already uses the Go 1.26 Alpine tag, which resolves to the current patch release. Changing its runtime user remains part of Phase 4b and is not included here.

### Go Modules

- `github.com/goccy/go-json`: `v0.10.5` → `v0.10.6`
- `github.com/gofiber/fiber/v2`: `v2.52.12` → `v2.52.14`
- `google.golang.org/grpc`: `v1.79.2` → `v1.82.0`
- `golang.org/x/crypto`: `v0.48.0` → `v0.54.0`
- OpenTelemetry API, metric, trace, SDK, and OTLP trace exporter modules: `v1.42.0` → `v1.44.0`
- `golang.org/x/net`: `v0.51.0` → `v0.57.0`
- `golang.org/x/image`: `v0.36.0` → `v0.44.0`

The implementation will record the exact final transitive versions produced by `go mod tidy`. Transitive changes are acceptable only when required by the selected direct dependencies or by the security fixes above.

## Files in Scope

- `go.mod`
- `go.sum`
- `.github/workflows/ci.yml`
- `README.md`

No application behavior or public API changes are expected.

## Verification

The consolidated result must pass:

1. `go mod tidy`
2. `go mod verify`
3. `go build ./...`
4. `go test ./internal/... ./pkg/...`
5. The repository lint and coverage checks
6. `govulncheck ./...`
7. Integration tests with the application, PostgreSQL, and Redis running
8. A final diff review confirming that no unrelated direct dependency was upgraded

If a selected current release introduces an incompatibility, keep the highest secure compatible release and document the downgrade with evidence from the failing build or test.

## Pull Request Handling

- Prepare the consolidated change locally on `chore/consolidated-dependency-upgrades`.
- Do not close the existing Dependabot pull requests before the replacement is verified and merged.
- Do not push or create the replacement pull request until its title and description are approved.
- After merge, confirm which Dependabot pull requests GitHub closes automatically before manually closing any remainder.

## Out of Scope

- Broad upgrades of dependencies not represented by the open pull requests
- Application refactoring or feature work
- OpenTelemetry MeterProvider implementation
- Redis client consolidation
- Docker non-root runtime configuration
- Dependabot configuration changes
