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

Phase 1 verification and merge are complete. All four verification jobs passed on the merged revision above. Release publication and release notes remain separate work for the revision selected for that release.

Phase 2.1 article filter validation merged through [PR #53](https://github.com/zoe606/gobase/pull/53). Redis cache and rate limiter cleanup merged through [PR #57](https://github.com/zoe606/gobase/pull/57). Non-root Docker execution merged through [PR #58](https://github.com/zoe606/gobase/pull/58). See [runtime permissions](deployment.md#runtime-permissions) for bind mounts and existing upload volumes. The Phase 1 results above remain evidence for their recorded revisions. This source adds minimal service output under issue #54. Monorepo generation and a two-service HTTP communication example remain planned. Email verification, password reset, and OpenTelemetry metrics export remain deferred. See the [roadmap](roadmap.md).

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
