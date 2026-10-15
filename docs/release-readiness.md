# Template Release Readiness

Date: October 5, 2026.

Status: Phase 1 is merged. Local checks and GitHub Actions checks passed.

[Pull request #46](https://github.com/zoe606/gobase/pull/46) merged into `master` on October 5, 2026 as `66270fb626b7b6b5df208ce94b0d17f52fca8890`. The original implementation revision verified by GitHub Actions was `b01619495e95bb7109c374cddd445c4ef0e15181` on `feat/selectable-http-engines`, based on `75af22a`.

## Implemented Changes

- PostgreSQL 17 is used by local Compose, the Integration workflow, and the Engines workflow. Redis uses version 7.
- Generated applications receive their own README and contribution guide. These documents identify the selected engine. Template maintenance tools, maintenance documents, and historical plans are omitted.
- Generated projects include `.dockerignore`.
- `make swag` uses the module's pinned Swagger tool. The engine verifier regenerates Swagger after code generation and checks CRUD routes, success and validation response codes, and create/update request schemas.
- New example configurations use local storage. App and worker containers share the `upload_data` volume. MinIO uses the optional `s3` profile.
- Compose uses fixed internal service ports and consistent PostgreSQL credentials. Host ports and the database name can be changed without giving containers the host connection settings.
- `make docker-services` waits for PostgreSQL and Redis health checks.
- The Dockerfile builds for the requested target platform instead of always compiling for amd64.
- Auth integration tests use a unique registration email for each run. Repeated tests can use the same development database.
- The Engines workflow includes app and worker image builds, readiness and Swagger checks, a welcome email task through the worker, and exit code checks after shutdown.

## Local Verification

Environment: Go 1.27.1 and golangci-lint 2.14.0 on macOS arm64. Container runtime checks used Linux arm64, Docker 29.4.0, PostgreSQL 17, and Redis 7.

| Check | Gin | stdlib | Fiber |
|-------|-----|--------|-------|
| Generate project, resolve and verify modules | Passed | Passed | Passed |
| Build and race tests | Passed | Passed | Passed |
| Generate and wire a CRUD feature | Passed | Passed | Passed |
| Regenerate and check Swagger | Passed | Passed | Passed |
| Lint generated output | Passed | Passed | Passed |
| Local README startup commands | Passed | Passed | Passed |
| Native worker welcome email task | Passed | Passed | Passed |
| Native app and worker shutdown | Passed | Passed | Passed |
| Docker app and worker startup | Passed | Passed | Passed |
| Docker health and Swagger responses | Passed | Passed | Passed |
| Docker worker welcome email task | Passed | Passed | Passed |
| Ten HTTP integration tests against Docker app | Passed | Passed | Passed |
| Docker app and worker exit code 0 after shutdown | Passed | Passed | Passed |
| Reachable vulnerabilities reported by govulncheck | 0 | 0 | 0 |

Worker email checks used the existing noop email sender. They verify queue processing, not delivery through an external email provider. HTTP integration tests include the generated CRUD feature, auth, profile, media, articles, translation, and history.

The checkout passed `go test -count=1 ./...` against a running application. `make check-all` passed with 93% coverage under the configured coverage exclusions. The Fiber integration suite also passed twice in one invocation with `-count=2` on the same database.

The local Docker build and runtime results above cover Linux arm64. GitHub Actions also passed Linux amd64 Docker runtime checks for all three engines. The earlier [engine verification report](http-engines-verification.md) records the separate Linux amd64 binary builds.

## GitHub Actions Verification

The following results passed on implementation revision `b01619495e95bb7109c374cddd445c4ef0e15181` on October 5, 2026. No remote-only code fixes were needed.

| Workflow | Result | Evidence |
|----------|--------|----------|
| CI / Quality | Passed | [Workflow run](https://github.com/zoe606/gobase/actions/runs/37278362473) |
| Engines / Gin | Passed | [Engine job](https://github.com/zoe606/gobase/actions/runs/37278362455/job/111660394210) |
| Engines / stdlib | Passed | [Engine job](https://github.com/zoe606/gobase/actions/runs/37278362455/job/111660394534) |
| Engines / Fiber | Passed | [Engine job](https://github.com/zoe606/gobase/actions/runs/37278362455/job/111660394557) |

Each engine job passed project generation, builds, race tests, CRUD generation and wiring, Swagger checks, lint, vulnerability scanning, and eight HTTP integration tests with PostgreSQL 17 and Redis 7. Each job also built and ran Linux amd64 app and worker images, checked readiness and Swagger, processed a welcome email task with the noop sender, and checked exit code 0 after shutdown. The broader ten-test HTTP suite was verified locally for each engine.

### Verification After Merge

GitHub Actions also passed on `master` revision `66270fb626b7b6b5df208ce94b0d17f52fca8890` after PR #46 merged.

| Workflow | Result | Evidence |
|----------|--------|----------|
| CI / Quality | Passed | [Workflow run](https://github.com/zoe606/gobase/actions/runs/37294989410) |
| Engines / Gin, stdlib, and Fiber | All passed | [Workflow run](https://github.com/zoe606/gobase/actions/runs/37294989373) |

## Compatibility Notes

Engine selection still occurs only during project creation. Existing projects without `.gobase.json` retain Fiber code generation. No runtime engine switch or existing-project conversion is added.

Local storage is the default in new example files. The configuration loader's existing fallback defaults remain unchanged. Existing S3 configurations can continue using S3. Set `STORAGE_S3_DOCKER_ENDPOINT` when the Docker endpoint differs from the local endpoint. Existing Compose setups using MinIO must enable `COMPOSE_PROFILES=s3` and supply a usable `MINIO_IMAGE`.

The community MinIO images could not be pulled from Docker Hub or Quay during these checks. The [upstream repository](https://github.com/minio/minio) is archived. The default quick start does not require MinIO. The optional MinIO profile was inspected as configuration but was not run.

## Repeat the Checks

From the template checkout:

```bash
make check-all
go run ./pkg/tools/verifyengines -output /tmp/gobase-release-check -lint
```

The output directory must be new. In each generated project, follow its README to configure services and start the app and worker. For Docker runtime checks:

```bash
make docker-dev-build
curl http://localhost:8080/readyz
curl http://localhost:8080/swagger/doc.json
APP_HOST=localhost APP_PORT=8080 make test-integration
make docker-stop
```

## Release Preparation

Phase 1 verification and merge are complete. All four verification jobs passed on the merged revision above. The v0.1.0 baseline includes the later foundation, minimal output, and application verification described below. Patch release v0.1.1 adds the verified panic recovery fix. Historical results remain tied to their recorded revisions.

Phase 2.1 article filter validation merged through [PR #53](https://github.com/zoe606/gobase/pull/53). Redis cache and rate limiter cleanup merged through [PR #57](https://github.com/zoe606/gobase/pull/57). Non-root Docker execution merged through [PR #58](https://github.com/zoe606/gobase/pull/58). Minimal output merged through [PR #59](https://github.com/zoe606/gobase/pull/59), closing issue #54. See [runtime permissions](deployment.md#runtime-permissions) for bind mounts and existing upload volumes. Three independent Commerce applications now verify the selected engines and their HTTP workflow. Issues #60 and #56 are complete. Email verification, password reset, and OpenTelemetry metrics export remain deferred unless a selected product requires them. See the [roadmap](roadmap.md).

The results in this report verify the template and generated runtime for the recorded checks. Application-level production readiness is still unverified. Its report must identify the business workflow, application revisions, production environment, workload criteria, deployment, and failure and recovery results. It must also identify any substituted providers and untested behavior. Monorepo generation is outside the current plan.

### Phase 2.3 Local Verification

On October 8, 2026, generated Gin, stdlib, and Fiber projects passed module verification, builds, race tests, CRUD generation, Swagger checks, and lint. Both Linux arm64 Docker images for each engine ran as UID and GID `65532`. Checks verified configuration files with mode `600`, migrations, system CA certificates, temporary file writes, and fresh shared upload volumes. Eight HTTP integration tests passed for each engine, including upload and all three image variants written by the worker. Welcome email tasks used the noop sender. All six containers exited with code `0` after shutdown.

The generated Gin quick start also passed with Compose's host configuration mount and a fresh upload volume. A root-owned test volume retained its existing data after the documented ownership update, and upload and image processing passed after restart. The checkout passed `go test -count=1 ./...` against that stack. `make check-all` passed with 93% coverage, no lint issues, and no reachable vulnerabilities. These local Docker checks used Docker 29.4.0 on Linux arm64. The Engines workflow repeats the non-root runtime checks on Linux amd64 for every engine.

### Phase 2.3 GitHub Actions Verification

On October 8, 2026, [Quality](https://github.com/zoe606/gobase/actions/runs/37784776610) and [all three engine jobs](https://github.com/zoe606/gobase/actions/runs/37784776557) passed on implementation revision `8af33a54e20837a734c306fd0f373d14a76d63df` in [PR #58](https://github.com/zoe606/gobase/pull/58). Each engine job passed generation, builds, race tests, code generation, Swagger, lint, vulnerability scanning, and HTTP integration. Linux amd64 Docker checks also passed for actual app and worker identities, filesystem access, shared uploads, image variants, welcome email processing, and graceful shutdown. Final revision `11e6151e066bf38017a78c2ad0ab2ded4933ce24` passed [Quality](https://github.com/zoe606/gobase/actions/runs/37785534960) and [all engine jobs](https://github.com/zoe606/gobase/actions/runs/37785534931) before PR #58 merged as `619162fddc412a32c2b4344171d083f0dac4dbf8`.

### Phase 3 Local Verification

On October 8, 2026, `feat/minimal-service-output`, based on merged revision `619162fddc412a32c2b4344171d083f0dac4dbf8`, passed the following checks for fresh minimal output in all three engines:

| Check | Gin | stdlib | Fiber |
|-------|-----|--------|-------|
| Module resolution, verification, and excluded dependency checks | Passed | Passed | Passed |
| Builds, race tests, and lint | Passed | Passed | Passed |
| Parsing, validation, response envelope, request IDs, 404, and recovery contracts | Passed | Passed | Passed |
| Startup without configuration files or external infrastructure | Passed | Passed | Passed |
| Occupied-port startup failure and SIGTERM shutdown | Passed | Passed | Passed |
| Clear unsupported CRUD and wiring errors without persistence output | Passed | Passed | Passed |
| App-only Compose quick start with host configuration mount | Passed | Passed | Passed |
| Linux arm64 non-root identity, configuration, CA certificates, and temporary writes | Passed | Passed | Passed |
| Docker health, error response, and exit code 0 after shutdown | Passed | Passed | Passed |
| Reachable vulnerabilities reported by govulncheck | 0 | 0 | 0 |

Full-profile output also passed generation, module verification, builds, race tests, CRUD generation and wiring, Swagger checks, and lint for all three engines. Generated Gin integration tests passed with dedicated PostgreSQL 17 and Redis 7 fixtures and a running worker. The checkout passed `go test -count=1 ./...` against that app and `make check-all` with 93% coverage, no lint issues, and no reachable vulnerabilities. The fixture processes stopped cleanly, and their containers and volumes were removed.

Minimal checks use Go 1.27.1 and golangci-lint 2.14.0 on macOS arm64. Compose runtime checks use Docker 29.4.0 and Linux arm64. Vulnerability scans report no reachable vulnerabilities; findings in unused package or module paths remain separate results. The Engines workflow adds three infrastructure-free Minimal jobs with Linux amd64 Docker checks. Remote results must be inspected for the implementing PR before merge.

Repeat minimal verification from the template checkout:

```bash
go run ./pkg/tools/verifyengines -profile=minimal -lint -output=/tmp/gobase-minimal-check
```

Use a new output directory. Follow each generated README for its app-only Compose commands. The [minimal service guide](minimal-services.md) records the file set and omitted dependencies.

### October 9 Security Patch Verification

The October 9 scan reported newly published standard-library and x/net vulnerabilities in the previous baseline. The source, generated documentation, workflows, and Docker builders now use Go 1.27.2. Full and minimal modules pin x/net v0.60.0. These versions contain the fixes recorded in [GO-2026-6617](https://pkg.go.dev/vuln/GO-2026-6617).

Final source revision `631ff36515fce259c0eda237c7716feb359cc96d` passed `make check-all` with no lint issues, no reachable vulnerabilities, and 86.8% selective coverage. The generator workflow test covers layer generation, contracts, preserving user-owned files, and forced regeneration. The CI threshold remains 85%. Fresh full and minimal output passed module verification, builds, race tests, and lint for Gin, stdlib, and Fiber. Full output also passed CRUD generation, wiring, and Swagger checks. Minimal output passed native startup, occupied-port failure, and SIGTERM shutdown without infrastructure. All three minimal vulnerability scans found no reachable vulnerabilities. Findings in unused package or module paths remain separate results.

These local checks used Go 1.27.2 and golangci-lint 2.14.0 on macOS arm64. All seven GitHub jobs passed on the final revision: [Quality](https://github.com/zoe606/gobase/actions/runs/37894333219), and all three full and all three minimal jobs in [Engines](https://github.com/zoe606/gobase/actions/runs/37894333216). Engine jobs include Linux amd64 Docker runtime checks. [PR #59](https://github.com/zoe606/gobase/pull/59) merged on October 9, 2026 as `a611e8406f491bb259d8a120b89f86b1f7ebaab5`. Issue #54 is closed.

The [Commerce verification report](commerce-verification.md) records the application work based on this merged foundation. Application results do not change the scope of the historical template checks above.

## v0.1.0 Baseline

Baseline date: October 11, 2026.

This is the first versioned template baseline. New projects choose Gin v1.12.0, standard `http.ServeMux`, or Fiber v2.52.14 and the full or minimal profile. Gin and full are the creation defaults. The source checkout retains Fiber. Go 1.27.2 and x/net v0.60.0 are the dependency baseline. Full runtime verification uses PostgreSQL 17 and Redis 7.

Full output includes the existing authentication, persistence, local storage, worker, and CRUD generation features. Minimal output starts without required database, cache, storage, email, or worker infrastructure. Handler APIs are native to the selected engine. JSON response envelopes and error contracts remain shared.

The engine is selected during creation and recorded in `.gobase.json`. Editing this file does not convert an existing application. Existing full-profile creation commands remain supported. Generated applications own their code and do not receive source template updates automatically. Apply relevant foundation fixes and verify them in the owning application.

The release tag must identify a master revision whose Quality and six full/minimal engine jobs passed. The 85% coverage gate remains unchanged. The GitHub release notes record the exact tag commit and its CI runs.

Before this documentation update, all seven jobs passed on `620c8614dd4ac72fe54ce024013593d888db28a1`: [Quality](https://github.com/zoe606/gobase/actions/runs/38150670114) and [Engines](https://github.com/zoe606/gobase/actions/runs/38150670107). Engine checks include generated builds, race tests, lint, reachable vulnerability scans, and the relevant code generation, Swagger, HTTP, and non-root Docker runtime checks.

The independent Commerce workflow passed ten scenarios in both modes after Order PR #6 merged. [Commerce E2E](https://github.com/zoe606/commerce-order/actions/runs/38151098408) and [Order Quality](https://github.com/zoe606/commerce-order/actions/runs/38151098378) passed on `d4fb7e78a1e183a11665730aa008b796d29854dd`. The [Commerce report](commerce-verification.md) records all three application revisions.

Email verification, password reset HTTP integration, and OpenTelemetry metrics export remain deferred. External deployment, backup and restore, workload capacity, customer cancellation, refunds, and real charging need separate application evidence. The release records the tested template and workflow baseline. It does not claim complete Commerce production readiness.

## v0.1.1 Patch Release

Release preparation date: October 12, 2026.

This patch includes the panic recovery fix from [PR #64](https://github.com/zoe606/gobase/pull/64). When a full-profile handler panics before writing a response, Gin, stdlib, and Fiber return HTTP 500 with `INTERNAL_ERROR`, the shared JSON envelope, and the request ID. Recovery runs inside compression so gzip does not commit HTTP 200 while the panic is handled. Panic details remain in server logs. Gin and stdlib cannot replace headers that have already been sent.

Engine selection, full and minimal creation commands, native handler APIs, and dependency versions retain the v0.1.0 baseline. Generated applications own their code and need explicit adoption of the relevant middleware and router changes. The three Commerce applications have adopted and verified the fix. See the [Commerce report](commerce-verification.md#panic-recovery-follow-up) for their revisions and CI results.

All seven source checks passed on the merged fix revision `339ad3c777afe3197831b063c57c92be109fbe86`: [Quality](https://github.com/zoe606/gobase/actions/runs/38202661062) and all six [engine jobs](https://github.com/zoe606/gobase/actions/runs/38202661056). These results precede the release documentation update. The release tag must identify a master revision whose Quality and six full/minimal engine jobs passed. Release notes record that exact commit and its CI runs.

## v0.2.0 Capability Release Candidate

Preparation date: October 15, 2026. Tracking: [#85](https://github.com/zoe606/gobase/issues/85).

The candidate baseline is master `b88a84025d0c0ea93e65098a76630b4630a45628`. At preparation, the latest published release was v0.1.1. v0.2.0 adds optional Full-profile capabilities and hardens existing task processing. Public tag and release publication follow candidate verification. The final candidate revision and readiness PR checks are recorded in #85.

### Dependency and Compatibility Baseline

| Component | Recorded pin |
|-----------|--------------|
| Go | 1.27.2; also used by CI and Docker builders |
| Gin / Fiber | v1.12.0 / v2.52.15 |
| stdlib router | `net/http` from Go 1.27.2 |
| Asynq / go-redis | v0.26.0 / v9.21.0 |
| Gorilla / Fiber WebSocket / fasthttp WebSocket | v1.5.3 / v1.3.4 / v1.5.12 |
| grpc-go / protobuf-go | v1.83.2 / v1.36.11 |
| protoc / protoc-gen-go / protoc-gen-go-grpc | v36.2 / v1.36.11 / v1.6.2 |
| Zap / migrate / x/net | v1.28.0 / v4.20.1 / v0.60.0 |

Go 1.27.2 is listed as stable in the [official downloads](https://go.dev/dl/). Runtime pins are taken from the recorded source and generated modules. This readiness slice keeps those pins. The generated migration CLI instructions and manual Integration workflow now use the existing migrate v4.20.1 module version.

Gin and Full remain creation defaults. Each generated project contains its selected native engine. The source checkout retains Fiber. Existing Full and Minimal creation commands and successful HTTP contracts remain supported. WebSocket and application gRPC remain disabled by default. Minimal output omits these authenticated capabilities and starts without infrastructure. Generated applications own their code; template updates require explicit adoption.

### Fresh Local Generation and Commands

Environment: Go 1.27.2 and golangci-lint 2.14.0 on macOS arm64. Full quick starts use PostgreSQL 17 and Redis 7 containers. Fixture port overrides are PostgreSQL 55460, Redis 56390, HTTP 58080, and gRPC 55051 because default ports are already used locally. Go commands receive exported variables, as documented. Each Full fixture has its own Compose project and volumes.

| Check | Full Gin | Full stdlib | Full Fiber | Minimal Gin | Minimal stdlib | Minimal Fiber |
|-------|----------|-------------|------------|-------------|----------------|---------------|
| Fresh generation, module verification, build, race tests, and lint | Passed | Passed | Passed | Passed | Passed | Passed |
| Dependency isolation and reachable vulnerability scan | Passed; 0 | Passed; 0 | Passed; 0 | Passed; 0 | Passed; 0 | Passed; 0 |
| CRUD generation, wiring, and Swagger regeneration | Passed | Passed | Passed | Omitted | Omitted | Omitted |
| `make init` and generated README `make run` | Passed | Passed | Passed | Passed | Passed | Passed |
| Health, readiness, and disabled WebSocket default | Passed | Passed | Passed | Passed | Passed | Passed |
| `make run-worker` and actual welcome-task processing | Passed; noop sender | Passed; noop sender | Passed; noop sender | Omitted | Omitted | Omitted |
| `make proto` reproduces checked-in code | Passed | Passed | Passed | Omitted | Omitted | Omitted |
| Disabled gRPC rejection, explicit `make run-grpc`, health, and HTTP-issued JWT calls | Passed | Passed | Passed | Omitted | Omitted | Omitted |
| Included gRPC and WebSocket client commands | Passed | Passed | Passed | Omitted | Omitted | Omitted |
| Persistence generation and wiring are rejected | Supported | Supported | Supported | Passed | Passed | Passed |

The native quick starts use the documented Go commands and Ctrl+C process-group shutdown. Compiled-process SIGTERM, active WebSocket close 1001, bounded worker shutdown, and non-root Docker checks are verified by the engine and application workflows. Welcome-task checks use the existing noop sender; they establish processing, not delivery to a mail provider.

Repeat generation with a new output directory for each engine/profile pair:

```bash
go run ./pkg/tools/verifyengines -engine=gin -profile=full -lint -output=/tmp/gobase-v020-full
go run ./pkg/tools/verifyengines -engine=gin -profile=minimal -lint -output=/tmp/gobase-v020-minimal
```

Repeat with `stdlib` and `fiber`, then run `go tool govulncheck ./...` in each output. Use the template README for `make init`. Follow each generated README for configuration, infrastructure, app, and worker commands. Full projects also document optional WebSocket and gRPC activation and clients. No external provider is required for these technical checks.

### Existing Exact-Revision CI and Application Evidence

All seven checks passed on baseline `b88a840`: [Quality 38584080948](https://github.com/zoe606/gobase/actions/runs/38584080948) and all six jobs in [Engines 38584081101](https://github.com/zoe606/gobase/actions/runs/38584081101). Full jobs include native and Linux amd64 Docker HTTP, worker, WebSocket, and gRPC checks. Minimal jobs include non-root Docker startup without infrastructure. A Go proxy download error required a failed-job rerun; the final results pass at the same source revision.

| Mini-app | Engine | Verified merged revision |
|----------|--------|--------------------------|
| Order | Gin | `d44d190dd757bf7268dd68a802b92adf8f121031` |
| Inventory | stdlib | `a748ba125155151aef88018059401636320671bc` |
| Billing | Fiber | `264e746944ad8c20828ae77b2c8fb5ee0e39c203` |

The common [post-merge application workflow 38580411740](https://github.com/zoe606/commerce-order/actions/runs/38580411740) passes 73 groups in each native and non-root Docker mode: 30 WebSocket, 15 controlled Asynq, 4 normal-worker/HTTP producer, 10 HTTP, and 14 gRPC. Both artifacts contain exactly 51 log/revision files. Their revision records match the table. Individual application Quality checks passed; detailed links and limits remain in [#83](https://github.com/zoe606/gobase/issues/83) and the [application report](commerce-verification.md#application-websocket-verification).

Application dependency pins remain independent. Billing's recorded revision uses Fiber v2.52.14; generated candidate Fiber output uses v2.52.15. The application results establish foundation adoption on their recorded revisions. Fresh generation and source engine CI establish the candidate pins.

### Prepared Release Notes

v0.2.0 follows v0.1.1 with these changes:

- Add optional JWT-authenticated WebSocket notification/echo support in Full projects, with native upgrades, explicit origins, bounded sessions, ping/pong, token expiry, and active shutdown.
- Add a separate optional application gRPC process, unary Echo contract, generated Go client, pinned protobuf generation, JWT metadata, verified TLS, health checks, deadlines, and shutdown.
- Apply configured Asynq retries and timeouts when enqueueing new tasks, preserve request contexts, archive malformed JSON without retry, and verify real Redis/worker failure recovery. Existing queued tasks retain their recorded options.
- Include the merged dependency pins, current activation instructions, independent mini-app evidence, and the stable TLS identity rejection fixture.

Release notes must identify the final verified candidate commit and its post-merge Quality and six engine/profile runs. After publication, the versioned creation command is:

```bash
git clone --branch v0.2.0 https://github.com/zoe606/gobase.git gobase
cd gobase
make init ENGINE=gin PROFILE=full MODULE=github.com/your-org/myapp APP_NAME=myapp OUTPUT=../myapp
```

The tag is a template baseline. It does not update existing generated applications. It does not establish browser WebSocket authentication, application broadcasts, production deployment, workload capacity, real email delivery, or complete Commerce product readiness. mTLS, streaming RPCs, a broker, and a monorepo remain separate work. Issues #49 and #50 remain deferred. No public tag is created by this readiness report.

## v0.2.0 Capability Release

Release preparation date: October 15, 2026.

Readiness is complete under [#85](https://github.com/zoe606/gobase/issues/85). [PR #86](https://github.com/zoe606/gobase/pull/86) merged as `862c65d2542b93a48c9ae14df13da50363ab120c`. All seven checks passed on that exact candidate: [Quality](https://github.com/zoe606/gobase/actions/runs/38678084378) and all six [engine/profile jobs](https://github.com/zoe606/gobase/actions/runs/38678084374). Quality reported 87% coverage, zero lint issues, and no reachable vulnerabilities. All six generated scans found no reachable vulnerabilities.

The release includes the optional Full-profile Asynq, WebSocket, and gRPC capabilities and dependency pins recorded above. The fresh generation, documented quick starts, and exact-revision mini-app results retain their recorded scope and limits. The publication change updates only README, roadmap, capability-plan, and release-report text. It directs new projects to v0.2.0 and records completed readiness.

The tag must identify a master revision whose Quality and all six Full/Minimal engine jobs passed. [Release notes](https://github.com/zoe606/gobase/releases/tag/v0.2.0) record the exact tag commit and its passing post-merge workflows. The tag follows those checks. The 85% coverage gate remains unchanged. Generated applications require explicit adoption of relevant template updates. Continue maintaining the verified paths and fix reproduced foundation defects. Issues #49 and #50 remain deferred until needed by an application.
