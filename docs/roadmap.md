# Gobase Roadmap

Updated: October 5, 2026.

This document plans work on the gobase template. Phase 1 is merged. Phase 2 and later work remain proposals. Generated applications can adopt later changes according to their needs.

## Current Baseline

The HTTP engine refactor is implemented and verified locally and in GitHub Actions. Project creation supports Gin, standard `net/http`, and Fiber. Each engine has independent endpoint and CRUD generator templates. Use cases, DTOs, entities, repositories, migrations, and workers remain shared.

The engine is selected during project creation. `.gobase.json` records that choice for code generation. Changing this file does not migrate an existing application. The runnable template checkout retains Fiber for compatibility.

All three generated projects passed local builds, race tests, lint, shared HTTP contracts, and integration tests with PostgreSQL and Redis. [PR #46](https://github.com/zoe606/gobase/pull/46) merged on October 5, 2026 as `66270fb626b7b6b5df208ce94b0d17f52fca8890`. GitHub Actions Quality and all three engine jobs passed on that `master` revision, including Linux amd64 Docker app and worker runtime checks. See the [release readiness report](release-readiness.md) for workflow links and the [local verification results](http-engines-verification.md) for environment details.

## Phase Order

| Phase | Outcome | Status |
|-------|---------|--------|
| HTTP engine refactor | Three engines with native handlers and shared core | Implemented; local and GitHub verification passed |
| 1. Template release readiness | Reproducible setup and CI evidence for all engines | Complete and merged; master CI passed |
| 2. Application hardening | Complete auth flows, metrics export, and resource lifecycle | Proposed; start with 2.1 after agreeing its response contract |
| Later: starter profiles | Optional smaller outputs for different project needs | Deferred; design not selected |

## Phase 1: Template Release Readiness

Release readiness and maintenance checks are complete and merged. Release publication remains a separate step.

See the [release readiness report](release-readiness.md) for implemented changes, completed local and GitHub checks, and compatibility notes.

### Work

- Run the Engines workflow for Gin, stdlib, and Fiber. Inspect build, test, code generation, lint, vulnerability scan, and integration results. Record the workflow run and commit used.
- Keep local Compose and integration workflows on PostgreSQL 17. Update the documented support and verification together when changing this version.
- Follow the README from a new output directory for each engine. Check configuration, migrations, HTTP startup, worker startup, health probes, and shutdown. Record the commands that actually work.
- Build and run the Docker app and worker images for each output. The existing Linux binary builds do not verify the runtime images.
- Regenerate Swagger after generating and wiring a feature in each engine. Check that documented routes, request schemas, and response codes match the running application.
- Update the contribution guide with the source locations for endpoint templates, CRUD templates, shared HTTP helpers, and shared contract tests. Endpoint changes must update all affected engine templates. Changes to the runnable checkout may also require its Fiber handler to be updated.
- Separate template maintenance instructions from generated application instructions. Generated documentation should identify the selected engine and avoid commands for tools omitted from the output.
- Record the verified revision and compatibility notes in release documentation.

### Completion Criteria

- All three engine jobs pass on the revision being released.
- A fresh generated project follows the documented setup successfully for every engine.
- Docker app and worker images start, perform their expected work, and stop cleanly for every engine.
- Generated CRUD routes and Swagger documentation agree with the tested HTTP contracts.
- The supported PostgreSQL version or versions are explicit and tested.
- Contributors can identify the files and verification required for a shared endpoint change.

### Verification

Reuse `make check-all`, `go run ./pkg/tools/verifyengines -lint`, and the existing integration suite. Add focused coverage only for behavior or generation paths that these checks do not cover. Record Docker runtime and Swagger checks separately from the existing binary build results.

## Phase 2: Application Hardening

Complete the remaining application work in small changes. This continues the backlog listed in the [Phase 4a design](superpowers/specs/2026-03-15-phase4a-critical-fixes-design.md#out-of-scope). Recheck each item against the current implementation before starting it.

### Proposed Execution Order

Implement one slice at a time on a new branch from `master`. Keep the existing order below. Each slice needs its own scope, verification, and PR before the next slice starts. Phase 2 implementation has not started.

Update the template README and `pkg/scaffold/application/README.md.tmpl` when a slice changes feature availability or setup instructions.

| Order | Proposed issue | Completion evidence |
|-------|----------------|---------------------|
| 2.1 | Validate article list filters across HTTP engines | Agreed invalid-filter response, preserved valid requests, and passing shared HTTP contracts |
| 2.2 | Connect email verification and password reset through HTTP and workers | HTTP, queue, worker, persistence, and token consumption checks for every engine |
| 2.3 | Configure OpenTelemetry metrics export and shutdown | Exported measurements when enabled, working disabled mode, and provider shutdown |
| 2.4 | Define Redis client ownership and shutdown | Cache and rate limiter lifecycle checks, including disabled and memory configurations |
| 2.5 | Run Docker app and worker as a non-root user | Startup, upload, worker processing, and shutdown under the configured UID and GID |

The recommended first slice is 2.1. Before implementing it, agree which invalid filter values return HTTP 400 and the error code returned in the existing response envelope. Pagination normalization must retain its current behavior unless a separate change is agreed.

### 2.1 Article List Validation

`ListRequest.Status` declares allowed values, but the current article list handlers do not call the validator. The use case normalizes pagination.

- Define the response contract for invalid filter values before changing it. Rejecting a value that was previously accepted is an observable behavior change.
- Apply the agreed validation consistently in the runnable checkout and all engine templates.
- Extend the shared HTTP contract suite for the changed cases. Existing valid requests must retain their behavior.

### 2.2 Email Verification and Password Reset

The auth use cases and tests exist. HTTP auth routes currently cover register, login, refresh, logout, and the current user. Verification and reset use cases still contain email enqueue placeholders.

- Register the missing HTTP flows using each engine's native handler API.
- Connect verification, resend, and password reset requests to the existing Asynq email worker.
- Test request, queued email, worker processing, and token consumption together. Include expired tokens, repeated consumption, delivery failure, and the existing password reset response for an unknown email.
- Update DTOs, Swagger, configuration instructions, and shared HTTP contracts.

### 2.3 OpenTelemetry Metrics

`pkg/telemetry/telemetry.go` configures a trace provider. `pkg/telemetry/metrics.go` declares metric instruments, but telemetry initialization does not configure a MeterProvider or a metrics exporter.

- Select and document the metrics export configuration before implementation.
- Initialize and shut down the metrics provider with the existing telemetry lifecycle.
- Verify that database, cache, and worker instruments produce exported measurements when enabled. Keep disabled telemetry working.
- Preserve the existing HTTP `/metrics` behavior. Document how it relates to OpenTelemetry export.

### 2.4 Redis Ownership and Shutdown

`initAppCache` and `initRateLimitStorage` currently create separate Redis clients.

- Make connection ownership and shutdown explicit for these clients. Share a client only where configuration and lifecycle are compatible.
- Preserve memory rate limiting and disabled-cache configurations.
- Check readiness, shutdown, and Redis failure behavior with both cache and Redis rate limiting enabled.

### 2.5 Docker Non-Root Runtime

The final Docker image uses `scratch` without a `USER` instruction.

- Run app and worker images with an explicit non-root UID and GID.
- Set permissions for local uploads and any required temporary files. Keep configuration, migrations, and certificates readable.
- Verify startup, an upload, a worker task, and graceful shutdown in the runtime images.

### Completion Criteria

- Each completed slice has focused verification and updated documentation.
- Shared HTTP changes pass the contracts and integration tests for all three engines.
- Email verification and password reset work through HTTP, queue, worker, and persistence.
- Enabled telemetry exports measurements, and owned Redis clients close during shutdown.
- Docker app and worker images operate as the configured non-root user.
- Existing valid API requests, engine selection, code generation, and worker behavior remain covered by regression checks.

### GitHub Issue Tracking

No GitHub issues are open as of October 5, 2026. The proposed issues above have not been created.

Create one issue for each Phase 2 slice. Each issue should link to its section in this document and state the current problem, implementation scope, unresolved decisions, completion criteria, and required verification. This document defines the phase order and scope. GitHub issues track execution status and links to implementing PRs.

Link each implementation PR to its issue. Use `Closes #<issue-number>` only when the PR completes all acceptance criteria. Close the issue after the PR is merged and the relevant checks pass. If a PR completes only part of an issue, keep it open and list the remaining work. Record deferred work explicitly instead of marking it complete.

## Deferred Decisions

After Phases 1 and 2, consider a smaller starter profile that omits example features such as articles and translation. The current full output must remain the default until a separate design defines feature dependencies, migrations, wiring, and tests. No preset system is included in this plan.

Additional engines, runtime engine switching, conversion of existing applications, and a rewrite of the shared business layers are outside these phases.

## References

- [HTTP engine architecture](http-engines.md)
- [Local engine verification](http-engines-verification.md)
- [Code patterns](code-patterns.md)
- [Deployment](deployment.md)
