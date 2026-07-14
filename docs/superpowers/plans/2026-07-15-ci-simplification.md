# CI Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the nine-job automatic CI graph with one Quality job and preserve PostgreSQL/Redis integration testing as a manually dispatched workflow.

**Architecture:** `.github/workflows/ci.yml` becomes the single automatic quality gate and reuses Makefile targets for unit tests and coverage. `.github/workflows/integration.yml` owns the manual end-to-end environment, including explicit readiness failure and unconditional app cleanup.

**Tech Stack:** GitHub Actions, Go 1.26.5, GNU Make, golangci-lint v2.11.2, govulncheck v1.1.4, PostgreSQL 18, Redis 7, Codecov v7.

## Global Constraints

- Automatic CI runs on pull requests to `main` or `master` and pushes to `main` or `master`.
- Automatic CI exposes one job named `Quality`.
- The coverage threshold remains exactly 85%.
- Codecov remains non-blocking with `fail_ci_if_error: false`.
- Integration runs only through `workflow_dispatch`.
- Integration uses PostgreSQL 18 Alpine, Redis 7 Alpine, and migrate v4.19.1.
- Do not change application code, tests, Docker images, migrations, dependencies, or branch protection.
- Do not push the branch or mutate pull request #38 until the user approves the verified final diff.
- Keep Dependabot pull requests #26 and #32 open until pull request #38 is merged.

---

### Task 1: Replace the Automatic CI Graph with One Quality Job

**Files:**
- Modify: `.github/workflows/ci.yml`
- Verify: `Makefile`

**Interfaces:**
- Consumes: `make test`, `make coverage-check`, the Go tool declaration for `govulncheck`, and `coverage.txt` produced by the test target.
- Produces: One automatic GitHub Actions job with the stable display name `Quality`.

- [ ] **Step 1: Record the current structural failure**

Run:

```bash
job_count=$(awk '/^jobs:/{in_jobs=1; next} in_jobs && /^  [a-z0-9-]+:$/ {count++} END {print count+0}' .github/workflows/ci.yml)
test "$job_count" -eq 1
```

Expected: FAIL because the current workflow defines nine jobs.

- [ ] **Step 2: Replace `.github/workflows/ci.yml` with the single-job workflow**

Write this complete file:

```yaml
name: CI

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

permissions:
  contents: read

env:
  GO_VERSION: '1.26.5'
  COVERAGE_THRESHOLD: '85'

jobs:
  quality:
    name: Quality
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7

      - uses: actions/setup-go@v6
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Check formatting
        run: |
          files="$(gofmt -l .)"
          if [ -n "$files" ]; then
            echo "The following files need formatting:"
            printf '%s\n' "$files"
            exit 1
          fi

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v7
        with:
          version: v2.11.2

      - name: Build
        run: go build ./...

      - name: Run tests
        run: make test

      - name: Check coverage threshold
        run: make coverage-check

      - name: Run vulnerability scan
        run: go tool govulncheck ./...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v7
        with:
          files: ./coverage.txt
          fail_ci_if_error: false
        env:
          CODECOV_TOKEN: ${{ secrets.CODECOV_TOKEN }}
```

- [ ] **Step 3: Verify the new structure and removed jobs**

Run:

```bash
job_count=$(awk '/^jobs:/{in_jobs=1; next} in_jobs && /^  [a-z0-9-]+:$/ {count++} END {print count+0}' .github/workflows/ci.yml)
test "$job_count" -eq 1
rg -n '^    name: Quality$' .github/workflows/ci.yml
! rg -n 'CI Success|Nancy|YAML Lint|Dockerfile Lint|Integration Tests' .github/workflows/ci.yml
```

Expected: exit 0, one `Quality` match, and no removed-job matches.

- [ ] **Step 4: Validate YAML syntax**

Run:

```bash
ruby -e 'require "yaml"; YAML.parse_file(ARGV.fetch(0)); puts "valid YAML"' .github/workflows/ci.yml
```

Expected: `valid YAML`.

- [ ] **Step 5: Run the automatic Quality commands locally**

Run:

```bash
test -z "$(gofmt -l .)"
golangci-lint run
go build ./...
make test
COVERAGE_THRESHOLD=85 make coverage-check
go tool govulncheck ./...
```

Expected: every command exits 0, coverage is at least 85%, and govulncheck reports zero reachable vulnerabilities.

- [ ] **Step 6: Commit the automatic workflow**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: consolidate automatic quality checks"
```

---

### Task 2: Add the Manual Integration Workflow

**Files:**
- Create: `.github/workflows/integration.yml`
- Reference: `integration-test/integration_test.go`
- Reference: `migrations/`

**Interfaces:**
- Consumes: the app at `./cmd/app`, migrations under `migrations/`, `/healthz`, and integration tests configured through `APP_HOST` and `APP_PORT`.
- Produces: A GitHub Actions workflow named `Integration` with one manually dispatched job named `Integration Tests`.

- [ ] **Step 1: Record that the manual workflow is absent**

Run:

```bash
test -f .github/workflows/integration.yml
```

Expected: FAIL because the file does not exist yet.

- [ ] **Step 2: Create `.github/workflows/integration.yml`**

Write this complete file:

```yaml
name: Integration

on:
  workflow_dispatch:

permissions:
  contents: read

env:
  GO_VERSION: '1.26.5'

jobs:
  integration:
    name: Integration Tests
    runs-on: ubuntu-latest
    timeout-minutes: 10
    services:
      postgres:
        image: postgres:18-alpine
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: test_db
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v7

      - uses: actions/setup-go@v6
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run migrations
        env:
          POSTGRES_URL: postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable
        run: |
          go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1 \
            -path migrations \
            -database "$POSTGRES_URL" \
            up

      - name: Build and start app
        env:
          APP_ENV: development
          HTTP_PORT: '8080'
          POSTGRES_HOST: localhost
          POSTGRES_PORT: '5432'
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DBNAME: test_db
          POSTGRES_SSLMODE: disable
          REDIS_HOST: localhost
          REDIS_PORT: '6379'
          JWT_SECRET_KEY: integration-test-secret-key
          STORAGE_DRIVER: local
          STORAGE_LOCAL_PATH: /tmp/uploads
          STORAGE_LOCAL_URL: http://localhost:8080/uploads
        run: |
          CGO_ENABLED=0 go build -o ./bin/app ./cmd/app
          mkdir -p /tmp/uploads
          ./bin/app > /tmp/app.log 2>&1 &
          echo $! > /tmp/app.pid

          ready=false
          for attempt in $(seq 1 30); do
            if curl -fsS http://localhost:8080/healthz > /dev/null; then
              ready=true
              break
            fi
            if ! kill -0 "$(cat /tmp/app.pid)" 2>/dev/null; then
              echo "Application exited before becoming ready"
              cat /tmp/app.log
              exit 1
            fi
            echo "Waiting for app... ($attempt/30)"
            sleep 1
          done

          if [ "$ready" != true ]; then
            echo "Application did not become ready within 30 seconds"
            cat /tmp/app.log
            exit 1
          fi

      - name: Run integration tests
        env:
          APP_HOST: localhost
          APP_PORT: '8080'
        run: go test -count=1 -v ./integration-test/...

      - name: Stop app
        if: always()
        run: |
          if [ -f /tmp/app.pid ]; then
            kill "$(cat /tmp/app.pid)" 2>/dev/null || true
          fi
          if [ -f /tmp/app.log ]; then
            cat /tmp/app.log
          fi
```

- [ ] **Step 3: Validate the manual-only trigger and runtime contract**

Run:

```bash
rg -n '^  workflow_dispatch:$' .github/workflows/integration.yml
! rg -n '^  (push|pull_request):$' .github/workflows/integration.yml
rg -n 'migrate@v4\.19\.1|go test -count=1 -v ./integration-test/\.\.\.|if: always\(\)' .github/workflows/integration.yml
```

Expected: one `workflow_dispatch` match, no automatic-trigger matches, and matches for the pinned migration, uncached integration run, and unconditional cleanup.

- [ ] **Step 4: Validate YAML syntax**

Run:

```bash
ruby -e 'require "yaml"; YAML.parse_file(ARGV.fetch(0)); puts "valid YAML"' .github/workflows/integration.yml
```

Expected: `valid YAML`.

- [ ] **Step 5: Commit the manual workflow**

```bash
git add .github/workflows/integration.yml
git commit -m "ci: make integration verification manual"
```

---

### Task 3: Verify the Complete CI Simplification and Prepare PR Update

**Files:**
- Verify: `.github/workflows/ci.yml`
- Verify: `.github/workflows/integration.yml`
- Verify: `docs/superpowers/specs/2026-07-15-ci-simplification-design.md`
- Verify: `docs/superpowers/plans/2026-07-15-ci-simplification.md`

**Interfaces:**
- Consumes: The automatic and manual workflows from Tasks 1 and 2.
- Produces: Verified commits ready for explicit approval before updating pull request #38.

- [ ] **Step 1: Validate both workflow files together**

Run:

```bash
ruby -e 'require "yaml"; ARGV.each { |path| YAML.parse_file(path); puts "#{path}: valid YAML" }' \
  .github/workflows/ci.yml \
  .github/workflows/integration.yml
```

Expected: both paths report `valid YAML`.

- [ ] **Step 2: Run the full local quality gate**

Run:

```bash
test -z "$(gofmt -l .)"
golangci-lint run
go build ./...
make test
COVERAGE_THRESHOLD=85 make coverage-check
go tool govulncheck ./...
```

Expected: every command exits 0, coverage is at least 85%, and govulncheck reports zero reachable vulnerabilities.

- [ ] **Step 3: Audit the final branch scope**

Run:

```bash
git diff --check
git status --short --branch
git diff --stat origin/chore/consolidated-dependency-upgrades...HEAD
git diff --name-status origin/chore/consolidated-dependency-upgrades...HEAD
git log --oneline origin/chore/consolidated-dependency-upgrades..HEAD
```

Expected new tracked paths relative to the published PR branch:

```text
M .github/workflows/ci.yml
A .github/workflows/integration.yml
A docs/superpowers/plans/2026-07-15-ci-simplification.md
A docs/superpowers/specs/2026-07-15-ci-simplification-design.md
```

- [ ] **Step 4: Request final approval before pushing**

Present the commits, exact changed paths, local verification results, and the fact that the next push will replace the current nine-job PR run with one automatic `Quality` job. Do not push until the user explicitly approves.

- [ ] **Step 5: Push only after approval**

Run:

```bash
git push origin chore/consolidated-dependency-upgrades
```

Expected: the push succeeds and updates pull request #38.

- [ ] **Step 6: Verify the live pull request checks**

Run:

```bash
gh pr checks 38 --repo zoe606/gobase
```

Expected: the new workflow exposes one automatic `Quality` job. Do not dispatch the Integration workflow before it exists on the default branch.

- [ ] **Step 7: Close superseded Dependabot pull requests only after merge**

First verify the replacement state:

```bash
gh pr view 38 --repo zoe606/gobase --json state,mergedAt,url
```

Expected before cleanup: `state` is `MERGED` and `mergedAt` is non-null. If either condition is false, stop and leave #26 and #32 open.

After the merge condition is satisfied, run:

```bash
gh pr close 26 --repo zoe606/gobase --comment "Superseded by #38, which includes github.com/goccy/go-json v0.10.6."
gh pr close 32 --repo zoe606/gobase --comment "Superseded by #38, which includes codecov/codecov-action v7."
```

Expected: #26 and #32 close with links back to the merged consolidated pull request.
