# Consolidated Dependency and Toolchain Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace seven overlapping Dependabot pull requests with one verified toolchain, CI action, Go module, and security refresh.

**Architecture:** Start from `origin/master` on `chore/consolidated-dependency-upgrades`, update only the approved toolchain and dependency allowlist, and let Go's minimal version selection adjust required transitives. Verify the result in two layers: local quality/security checks and a disposable integration environment with isolated containers, ports, and volumes.

**Tech Stack:** Go 1.26.5, Go modules, GitHub Actions, Fiber, OpenTelemetry, Docker Compose, PostgreSQL, Redis

## Global Constraints

- Go directive and CI toolchain must be exactly `1.26.5`.
- `actions/checkout` must be exactly `v7`; `codecov/codecov-action` must be exactly `v7`.
- Direct module targets must be: `go-json v0.10.6`, Fiber `v2.52.14`, gRPC `v1.82.0`, `x/crypto v0.54.0`, and all directly required OpenTelemetry modules `v1.44.0`.
- Security transitive targets must be: `x/net v0.57.0` and `x/image v0.44.0`.
- Do not run `go get -u ./...` or upgrade an unrelated direct dependency.
- Transitive changes are allowed only when selected by the approved versions or required to clear a reachable security finding.
- Do not modify application behavior or public APIs.
- Do not modify the Docker runtime user, MeterProvider setup, Redis client lifecycle, or Dependabot configuration.
- Do not close existing Dependabot pull requests before the consolidated replacement is verified and merged.
- Do not push or create a pull request until its title and description are approved.

**Spec:** `docs/superpowers/specs/2026-07-14-consolidated-dependency-upgrades-design.md`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `go.mod` | Modify | Pin Go 1.26.5 and the approved Go module versions |
| `go.sum` | Modify | Record checksums selected by the approved module graph |
| `.github/workflows/ci.yml` | Modify | Run CI with Go 1.26.5, checkout v7, and Codecov v7 |
| `README.md` | Modify | Keep the advertised Go version aligned with the project toolchain |

No application source or test source files should change.

---

### Task 1: Upgrade the Go Toolchain and GitHub Actions

**Files:**
- Modify: `go.mod:3`
- Modify: `.github/workflows/ci.yml:9-10,18,33,51,69,103,114,138,164,204`
- Modify: `README.md:5`

**Interfaces:**
- Consumes: Current `origin/master` toolchain and workflow configuration.
- Produces: A Go 1.26.5 module and CI workflow using checkout v7 and Codecov v7; Task 2 relies on this toolchain to evaluate the upgraded module graph.

- [ ] **Step 1: Capture the security baseline before changing versions**

Run:

```bash
govulncheck ./...
```

Expected: FAIL with reachable findings that include the pre-upgrade Go toolchain and vulnerable `golang.org/x/net` or `golang.org/x/image` versions. Save the finding IDs and affected versions in the execution notes; do not change application code to suppress them.

- [ ] **Step 2: Update the Go directive**

Change the top of `go.mod` to:

```go
module go-boilerplate

go 1.26.5
```

- [ ] **Step 3: Update the CI toolchain and action majors**

In `.github/workflows/ci.yml`, set:

```yaml
env:
  GO_VERSION: '1.26.5'
```

Replace every checkout invocation:

```yaml
- uses: actions/checkout@v7
```

Replace the Codecov invocation:

```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v7
  with:
    files: ./coverage.txt
    fail_ci_if_error: false
  env:
    CODECOV_TOKEN: ${{ secrets.CODECOV_TOKEN }}
```

Keep `actions/setup-go@v6`, `golangci/golangci-lint-action@v7`, and all job behavior unchanged.

- [ ] **Step 4: Align the README badge**

Change `README.md:5` to:

```markdown
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat&logo=go)](https://go.dev/)
```

- [ ] **Step 5: Verify that stale versions are gone and the requested toolchain resolves**

Run:

```bash
rg -n "1\.26\.1|Go-1\.25|actions/checkout@v6|codecov/codecov-action@v5" go.mod README.md .github/workflows/ci.yml
```

Expected: no matches.

Run:

```bash
rg -n "go 1\.26\.5|GO_VERSION: '1\.26\.5'|Go-1\.26|actions/checkout@v7|codecov/codecov-action@v7" go.mod README.md .github/workflows/ci.yml
```

Expected: the Go directive, CI toolchain, README badge, eight checkout v7 usages, and one Codecov v7 usage are listed.

Run:

```bash
go env GOVERSION
```

Expected: `go1.26.5`. With `GOTOOLCHAIN=auto`, Go may download the requested patch toolchain on the first invocation.

Run:

```bash
git diff --check
```

Expected: exit 0 with no whitespace errors.

- [ ] **Step 6: Commit the toolchain and CI action upgrade**

```bash
git add go.mod .github/workflows/ci.yml README.md
git commit -m "chore: update Go toolchain and CI actions"
```

---

### Task 2: Upgrade the Approved Go Module Set

**Files:**
- Modify: `go.mod:13-46,48-end`
- Modify: `go.sum`

**Interfaces:**
- Consumes: Go 1.26.5 from Task 1 and the exact approved module allowlist.
- Produces: A tidy, verified module graph with all approved direct modules aligned and the reachable `x/net` and `x/image` findings removed.

- [ ] **Step 1: Confirm that module vulnerabilities remain after only the toolchain upgrade**

Run:

```bash
govulncheck ./...
```

Expected: FAIL only for remaining reachable dependency findings, including the pre-upgrade `golang.org/x/net v0.51.0` or `golang.org/x/image v0.36.0`. Standard-library findings fixed by Go 1.26.5 must no longer appear.

- [ ] **Step 2: Request only the approved module versions**

Run:

```bash
go get \
  github.com/goccy/go-json@v0.10.6 \
  github.com/gofiber/fiber/v2@v2.52.14 \
  google.golang.org/grpc@v1.82.0 \
  golang.org/x/crypto@v0.54.0 \
  golang.org/x/net@v0.57.0 \
  golang.org/x/image@v0.44.0 \
  go.opentelemetry.io/otel@v1.44.0 \
  go.opentelemetry.io/otel/metric@v1.44.0 \
  go.opentelemetry.io/otel/trace@v1.44.0 \
  go.opentelemetry.io/otel/sdk@v1.44.0 \
  go.opentelemetry.io/otel/sdk/metric@v1.44.0 \
  go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.44.0
```

Expected: exit 0. Go may update indirect dependencies required by these exact versions. Do not add `-u`.

- [ ] **Step 3: Normalize the module graph**

Run:

```bash
go mod tidy
go mod verify
```

Expected: both commands exit 0; `go mod verify` prints `all modules verified`.

- [ ] **Step 4: Assert the selected versions**

Run:

```bash
go list -m -f '{{.Path}} {{.Version}}' \
  github.com/goccy/go-json \
  github.com/gofiber/fiber/v2 \
  google.golang.org/grpc \
  golang.org/x/crypto \
  golang.org/x/net \
  golang.org/x/image \
  go.opentelemetry.io/otel \
  go.opentelemetry.io/otel/metric \
  go.opentelemetry.io/otel/trace \
  go.opentelemetry.io/otel/sdk \
  go.opentelemetry.io/otel/sdk/metric \
  go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
```

Expected:

```text
github.com/goccy/go-json v0.10.6
github.com/gofiber/fiber/v2 v2.52.14
google.golang.org/grpc v1.82.0
golang.org/x/crypto v0.54.0
golang.org/x/net v0.57.0
golang.org/x/image v0.44.0
go.opentelemetry.io/otel v1.44.0
go.opentelemetry.io/otel/metric v1.44.0
go.opentelemetry.io/otel/trace v1.44.0
go.opentelemetry.io/otel/sdk v1.44.0
go.opentelemetry.io/otel/sdk/metric v1.44.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0
```

- [ ] **Step 5: Review the module diff against the allowlist**

Run:

```bash
git diff -- go.mod go.sum
```

Expected direct changes:

```text
github.com/goccy/go-json v0.10.6
github.com/gofiber/fiber/v2 v2.52.14
google.golang.org/grpc v1.82.0
golang.org/x/crypto v0.54.0
go.opentelemetry.io/otel v1.44.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0
go.opentelemetry.io/otel/metric v1.44.0
go.opentelemetry.io/otel/sdk v1.44.0
go.opentelemetry.io/otel/trace v1.44.0
```

`golang.org/x/net v0.57.0`, `golang.org/x/image v0.44.0`, `go.opentelemetry.io/otel/sdk/metric v1.44.0`, and dependencies required by the approved graph may change in the indirect block. If any other direct requirement changes, restore its original version unless `go mod graph` proves that an approved target requires the higher version; record that proof before continuing.

- [ ] **Step 6: Verify build, unit packages, and vulnerability scan**

Run:

```bash
go build ./...
go test ./internal/... ./pkg/...
govulncheck ./...
```

Expected: build and tests exit 0; `govulncheck` reports `No vulnerabilities found.`

- [ ] **Step 7: Commit the Go module upgrade**

```bash
git add go.mod go.sum
git commit -m "deps: consolidate Go module upgrades"
```

---

### Task 3: Run the Full Local Quality Gate

**Files:**
- Verify only: all tracked project files
- Generated and ignored: `coverage.txt`

**Interfaces:**
- Consumes: The committed toolchain, CI action, and Go module upgrades from Tasks 1 and 2.
- Produces: Fresh evidence that formatting, lint, vulnerability scanning, race-enabled tests, coverage, and builds pass together.

- [ ] **Step 1: Run the repository's complete local check**

Run:

```bash
make check-all
```

Expected: formatting, lint, vulnerability scan, race-enabled unit tests, and the repository coverage threshold all pass.

- [ ] **Step 2: Enforce the CI coverage threshold**

Run:

```bash
COVERAGE_THRESHOLD=85 make coverage-check
```

Expected: output includes `PASS: Coverage meets threshold` with coverage at or above 85%.

- [ ] **Step 3: Build every package and both production binaries**

Run:

```bash
go build ./...
CGO_ENABLED=0 go build -ldflags="-s -w" -o /private/tmp/gobase-dependency-upgrade-app ./cmd/app
CGO_ENABLED=0 go build -ldflags="-s -w" -o /private/tmp/gobase-dependency-upgrade-worker ./cmd/worker
```

Expected: all commands exit 0 and both binaries exist under `/private/tmp`.

- [ ] **Step 4: Confirm verification did not create tracked changes**

Run:

```bash
git diff --check
git status --short
```

Expected: `git diff --check` exits 0. `git status --short` shows only the pre-existing untracked local agent configuration files; no tracked file is modified.

---

### Task 4: Run Disposable Integration Verification

**Files:**
- Verify: `integration-test/integration_test.go`
- Use: `deployment/docker/docker-compose.yml`
- Disposable artifacts: `/private/tmp/gobase-dependency-upgrade-app`, `/private/tmp/gobase-dependency-upgrade-uploads`

**Interfaces:**
- Consumes: The verified app binary from Task 3 and migrations under `migrations/`.
- Produces: Fresh end-to-end evidence against isolated PostgreSQL and Redis services. The environment is named `gobase-dependency-upgrade` and must not share normal project volumes or container names.

- [ ] **Step 1: Start isolated PostgreSQL and Redis services**

Run:

```bash
COMPOSE_PROJECT_NAME=gobase-dependency-upgrade \
APP_NAME=gobase-dependency-upgrade \
POSTGRES_PORT=55432 \
REDIS_PORT=56379 \
POSTGRES_DBNAME=test_db \
docker compose \
  -f deployment/docker/docker-compose.yml \
  --env-file .env.example \
  up -d --wait db redis
```

Expected: the uniquely named database and Redis containers become healthy on ports 55432 and 56379.

- [ ] **Step 2: Apply migrations to the disposable database**

Run:

```bash
go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 \
  -path migrations \
  -database 'postgres://postgres:postgres@localhost:55432/test_db?sslmode=disable' \
  up
```

Expected: exit 0 with all migrations applied.

- [ ] **Step 3: Start the application in a long-running terminal session**

Run:

```bash
env \
  APP_ENV=development \
  HTTP_PORT=18080 \
  POSTGRES_HOST=localhost \
  POSTGRES_PORT=55432 \
  POSTGRES_USER=postgres \
  POSTGRES_PASSWORD=postgres \
  POSTGRES_DBNAME=test_db \
  POSTGRES_SSLMODE=disable \
  REDIS_HOST=localhost \
  REDIS_PORT=56379 \
  JWT_SECRET_KEY=integration-test-secret-key \
  STORAGE_DRIVER=local \
  STORAGE_LOCAL_PATH=/private/tmp/gobase-dependency-upgrade-uploads \
  STORAGE_LOCAL_URL=http://localhost:18080/uploads \
  /private/tmp/gobase-dependency-upgrade-app
```

Expected: the process remains running and logs that the HTTP server is listening on port 18080.

- [ ] **Step 4: Verify health and run integration tests from another terminal**

Run:

```bash
curl -fsS http://localhost:18080/healthz
APP_HOST=localhost APP_PORT=18080 go test -count=1 -v ./integration-test/...
```

Expected: the health endpoint succeeds and all integration tests pass.

- [ ] **Step 5: Stop the app and remove only the disposable Compose project**

Stop the long-running application session with `Ctrl-C`, then run:

```bash
COMPOSE_PROJECT_NAME=gobase-dependency-upgrade \
APP_NAME=gobase-dependency-upgrade \
POSTGRES_PORT=55432 \
REDIS_PORT=56379 \
POSTGRES_DBNAME=test_db \
docker compose \
  -f deployment/docker/docker-compose.yml \
  --env-file .env.example \
  down -v
```

Expected: only containers, the network, and volumes belonging to `gobase-dependency-upgrade` are removed.

---

### Task 5: Audit the Final Diff and Prepare the Pull Request Handoff

**Files:**
- Verify: `go.mod`, `go.sum`, `.github/workflows/ci.yml`, `README.md`
- Do not create or modify a pull request yet

**Interfaces:**
- Consumes: All commits and verification evidence from Tasks 1–4.
- Produces: An approval-ready pull request title and description while leaving Dependabot PRs #26 and #32–#37 open.

- [ ] **Step 1: Confirm branch scope and cleanliness**

Run:

```bash
git status --short --branch
git diff --stat origin/master...HEAD
git diff --name-status origin/master...HEAD
git log --oneline origin/master..HEAD
```

Expected tracked paths:

```text
.github/workflows/ci.yml
README.md
docs/superpowers/plans/2026-07-14-consolidated-dependency-upgrades.md
docs/superpowers/specs/2026-07-14-consolidated-dependency-upgrades-design.md
go.mod
go.sum
```

Pre-existing untracked local agent configuration files may remain. No application source file should appear.

- [ ] **Step 2: Draft the pull request title**

Use:

```text
deps: consolidate dependency and security upgrades
```

- [ ] **Step 3: Draft the pull request description**

Use this structure only after every listed verification command has passed:

```markdown
## Summary

- update Go from 1.26.1 to 1.26.5 and refresh CI actions
- consolidate Dependabot PRs #26 and #32–#37 using current dependency releases
- align OpenTelemetry modules at 1.44.0 and update vulnerable x/net and x/image versions

## Verification

- [x] `make check-all`
- [x] `COVERAGE_THRESHOLD=85 make coverage-check`
- [x] `go build ./...`
- [x] `govulncheck ./...`
- [x] `APP_HOST=localhost APP_PORT=18080 go test -count=1 -v ./integration-test/...`

## Supersedes

#26, #32, #33, #34, #35, #36, #37
```

- [ ] **Step 4: Request approval before external changes**

Present the final diff summary, verification evidence, proposed title, and proposed description to the user. Do not push the branch, create the pull request, or close any Dependabot pull request until the user explicitly approves those external actions.
