# Template Release Readiness

Date: October 5, 2026.

Status: Local Phase 1 checks passed. GitHub Actions verification is pending.

The working branch is `feat/selectable-http-engines`, based on `75af22a`. Remote workflow results and the tested revision will be recorded after the draft pull request runs.

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

The Docker build and runtime results above cover Linux arm64. Linux amd64 Docker runtime verification remains part of the pending GitHub Actions run. The earlier [engine verification report](http-engines-verification.md) records the separate Linux amd64 binary builds.

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

## Remaining Before Phase 1 Closes

- Publish a tested revision to a branch and run the Engines workflow through a pull request.
- Inspect the Quality result and all three Engine results on the exact revision.
- Record the revision, workflow links, and any remote-only fixes here.
- Use the verified revision for release notes when preparing a release.

Phase 2 application hardening has not started. See the [roadmap](roadmap.md).
