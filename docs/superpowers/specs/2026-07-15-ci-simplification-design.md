# CI Simplification Design

## Context

The repository currently runs nine GitHub Actions jobs. Most jobs repeat checkout, Go setup, module downloads, and cache initialization. Several checks are non-blocking, the unit-test package selection is duplicated from the Makefile, and the `CI Success` aggregator can pass before integration testing finishes.

The repository is a reusable Go boilerplate. Its automatic CI should provide one clear quality result without maintaining a large orchestration graph. End-to-end integration coverage remains useful, but it does not need to run on every pull request.

## Goals

- Reduce automatic CI to one required-quality result.
- Reuse Makefile test and coverage behavior instead of duplicating it in YAML.
- Retain formatting, linting, building, race-enabled tests, coverage enforcement, vulnerability scanning, and Codecov reporting.
- Preserve end-to-end PostgreSQL and Redis verification as a manually dispatched workflow.
- Pin tool versions where the repository already defines an approved version.
- Make readiness failures explicit and always clean up the background application process.

## Non-goals

- Change application code, tests, Docker images, migrations, or dependency versions.
- Add deployment, release, artifact-publishing, matrix, or scheduled workflows.
- Configure branch protection.
- Make integration testing an automatic pull-request gate.

## Automatic Quality Workflow

`.github/workflows/ci.yml` will keep the existing pull-request and `master` push triggers. It will contain one `Quality` job on `ubuntu-latest` with these ordered steps:

1. Check out the repository with `actions/checkout@v7`.
2. Install Go 1.26.5 with `actions/setup-go@v6` and its module/build cache.
3. Check formatting with `gofmt -l` without modifying files.
4. Run `golangci/golangci-lint-action@v7` with golangci-lint v2.11.2.
5. Compile all packages and commands with `go build ./...`.
6. Run `make test` to execute the existing race-enabled unit-test and coverage path.
7. Enforce `COVERAGE_THRESHOLD=85 make coverage-check`.
8. Run the pinned vulnerability scanner with `go tool govulncheck ./...`.
9. Upload `coverage.txt` with `codecov/codecov-action@v7`.

Formatting, linting, building, tests, coverage, and vulnerability scanning are blocking. Codecov remains advisory through `fail_ci_if_error: false`.

The workflow will remove the separate Format Check, Build, Test, Security Scan, YAML Lint, Dockerfile Lint, Integration Tests, and CI Success jobs. Nancy will also be removed because it is downloaded dynamically and its result is explicitly ignored.

## Manual Integration Workflow

`.github/workflows/integration.yml` will be a separate workflow triggered only by `workflow_dispatch`.

It will:

1. Start PostgreSQL 18 and Redis 7 service containers using the existing health checks.
2. Check out the selected ref and install Go 1.26.5 with caching.
3. Apply migrations with the pinned `github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1` command.
4. Build and start the application with the existing integration environment variables.
5. Poll `/healthz` for up to 30 seconds and fail the step if the application never becomes ready.
6. Run `go test -count=1 -v ./integration-test/...`.
7. Stop the application in an `if: always()` cleanup step.

The workflow will not run automatically for pushes or pull requests. Once merged to the default branch, maintainers can start it from GitHub Actions and choose the ref to test.

## Failure Behavior

- Any blocking Quality step failure makes the single automatic job fail.
- Module download or external service failures remain visible at the step that experienced them rather than producing a second aggregator failure.
- A readiness timeout fails before integration tests run, avoiding misleading HTTP failures against an unavailable app.
- Cleanup runs whether integration setup or tests pass or fail.

## Verification

Before updating pull request #38:

- Validate both workflow files as YAML.
- Run `gofmt -l .` and confirm no files are reported.
- Run `golangci-lint run`.
- Run `go build ./...`.
- Run `make test`.
- Run `COVERAGE_THRESHOLD=85 make coverage-check`.
- Run `go tool govulncheck ./...`.
- Confirm the diff contains only the approved design/plan documentation and the two workflow files.

After pushing, confirm the pull request reports a single automatic `Quality` job. The manual Integration workflow is verified from GitHub Actions after the workflow exists on the default branch.

## Trade-offs

The automatic steps will run sequentially, so the Quality job can take longer than the current parallel required jobs. In exchange, it uses one runner, performs one Go setup, shares one module cache, removes duplicated test logic, and produces one clear result. Manual integration reduces pull-request time and runner usage while preserving an on-demand end-to-end check.
