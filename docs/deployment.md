# Deployment

## Docker

### Local Development

```bash
# Start infrastructure only (PostgreSQL and Redis)
make docker-services

# Start full stack using existing images (infrastructure + app + worker)
make docker-dev

# Rebuild and start
make docker-dev-build

# View logs
make docker-logs

# Stop all containers
make docker-stop
```

The supported local and CI database version is PostgreSQL 17. Redis uses version 7. New projects use local storage. Docker app and worker containers share the `upload_data` volume.

`.env` supplies Docker Compose variables. Local `make run` and `make run-worker` commands read `config/config.yaml` and exported environment variables. Container connection ports remain 5432 for PostgreSQL and 6379 for Redis even when host ports change.

### Runtime Permissions

Both runtime images use `scratch` and run as UID `65532` and GID `65532`. The image provides `/uploads` for local storage and `/tmp` for temporary files. Both paths are writable by the runtime user. Configuration and migrations copied into the image belong to this user, including generated files with mode `600`. The system CA certificate bundle remains readable.

Compose mounts the host `config` directory read-only. Before using the example configuration with local Compose, allow directory traversal and reading of the example settings:

```bash
chmod 755 config
chmod 644 config/config.yaml
```

These permissions are for the local example configuration. For files containing production secrets or JWT private keys, use owner `65532` with mode `600`, or group `65532` with mode `640`. Their parent directories must allow traversal by this user or group. Bind mounts keep their host permissions and replace the permissions set in the image. Configure these permissions on the Docker host. Changing mounted configuration does not require rebuilding the image, but does require restarting the app and worker.

A fresh `upload_data` named volume receives the image's directory ownership. App and worker share this volume and the same UID and GID. Custom local storage paths, bind mounts, log file paths, and temporary mounts must also be writable by `65532:65532`. Overriding the container user requires matching permissions on all these paths.

Existing upload volumes may still belong to root. Stop the app and worker and back up the volume before changing ownership. Find the exact existing volume name and replace `myapp_upload_data` below:

```bash
docker compose -f deployment/docker/docker-compose.yml -f deployment/docker/docker-compose.app.yml stop app worker
docker volume ls
docker run --rm --user 0:0 --mount source=myapp_upload_data,target=/uploads alpine:3 chown -R 65532:65532 /uploads
make docker-dev-build
```

The ownership command applies only to the selected upload volume. It does not require deleting the volume or changing the database and Redis volumes. S3 configuration remains optional and does not require local upload permissions.

### Optional S3 or MinIO

To use an existing S3 service, set `STORAGE_DRIVER=s3` and the `STORAGE_S3_*` settings in `.env` for Docker. Set `STORAGE_S3_DOCKER_ENDPOINT` to the endpoint reachable from containers. For local Go processes, export the settings or update `config/config.yaml`. Create the configured bucket before uploading files.

To run MinIO locally, supply a usable `MINIO_IMAGE` in `.env`, then run `make docker-services-s3`. Set `STORAGE_DRIVER=s3`. Use `STORAGE_S3_ENDPOINT=localhost:9000` for local Go processes, or `minio:9000` for Docker containers. Start all services with `COMPOSE_PROFILES=s3 make docker-dev-build`.

The legacy community image could not be pulled during Phase 1 verification. The [MinIO repository](https://github.com/minio/minio) is archived. MinIO is omitted from the default quick start, and its image must be provided separately when enabling its profile.

### Production Build

```bash
# Build production binary
make build

# Run migrations on production
export PROD_DATABASE_URL='postgres://user:pass@host:5432/db?sslmode=require'
make migrate-prod
```

Build app and worker images with the same Dockerfile:

```bash
docker build -f deployment/docker/Dockerfile --build-arg TARGET=app -t myapp:app .
docker build -f deployment/docker/Dockerfile --build-arg TARGET=worker -t myapp:worker .
```

The Dockerfile uses Docker's target platform arguments for cross compilation. Use `--platform=linux/amd64` or `--platform=linux/arm64` when selecting a platform explicitly. See [Docker build variables](https://docs.docker.com/build/building/variables/#multi-platform-build-arguments).

### Runtime Verification

The Engines workflow builds and runs both images for Gin, stdlib, and Fiber. It checks the image user and actual process UID and GID, readable configuration, migrations and system certificates, writable temporary files, and a fresh shared upload volume. HTTP integration tests with `TEST_WORKER_ENABLED=true` wait for all three image variants before deleting the uploaded media. A welcome email task uses the noop sender. Both containers must exit with code `0` after graceful shutdown.

To include the image worker check when testing against a running app and worker:

```bash
APP_HOST=localhost APP_PORT=8080 TEST_WORKER_ENABLED=true make test-integration
```

## Background Tasks

The full profile includes `cmd/worker` and email and image tasks. The selected HTTP engine does not change worker behavior. Minimal projects do not include Asynq or require Redis.

| Setting | Default | Behavior |
|---------|---------|----------|
| `ASYNQ_CONCURRENCY` | `10` | Maximum concurrent worker tasks |
| `ASYNQ_MAX_RETRY` | `3` | Retries after the initial attempt; `0` disables retries |
| `ASYNQ_JOB_TIMEOUT` | `5m` | Maximum duration per task execution |

Export these values for local Go processes or set `asynq.concurrency`, `asynq.max_retry`, and `asynq.job_timeout` in `config/config.yaml`. Retry and timeout settings belong to the producer app. Restart that app to change options for new tasks. Existing queued tasks retain their recorded options. The worker error log reports each task's actual retry limit. Custom producers must pass their own options when enqueuing tasks.

Retry limits cannot be negative. Positive timeouts must use whole seconds because Asynq stores task timeouts in seconds. For example, `2s` and `5m` are valid; `200ms` and `1.5s` fail configuration validation. With the pinned Asynq v0.26.0, `0` uses its 30-minute fallback when no deadline is supplied. Use a positive value to set the intended limit explicitly.

Email and image handlers archive malformed JSON immediately using `asynq.SkipRetry`. Other handler errors follow the task's retry limit and Asynq's default retry delay. Exhausted tasks are archived. See [Asynq retry configuration](https://github.com/hibiken/asynq/wiki/Task-Retry).

Task execution can repeat after a crash or incomplete shutdown. The template does not guarantee exactly-once execution or email delivery. Handlers with external effects need an idempotency policy for their operation. On SIGTERM, the worker waits up to Asynq's default eight-second shutdown timeout. Unfinished work is returned to the queue. After a forced kill, recovery waits for the active task lease to expire.

Welcome-email submission remains best effort and does not fail registration. Image submission errors are logged without failing the completed upload. Both enqueue calls use the request context. Explicit callers should use `EnqueueContext` with a deadline and handle its error when Redis is unavailable. Readiness continues to check PostgreSQL; it does not assert Redis availability.

### Failure and Recovery Verification

Run against a dedicated Redis 7 instance:

```bash
ASYNQ_TEST_REDIS_ADDR=localhost:6379 go test -count=1 -race -timeout=4m -v ./integration-test/asynq
```

This suite starts separate worker processes with the real email handler and a controlled sender. It verifies temporary failures, retry exhaustion, malformed JSON, pending and active restart, repeated execution, task timeout, Redis connection interruption and recovery, and SIGTERM during processing. The controlled sender stores one logical effect atomically in Redis. This checks fixture idempotency, not actual email delivery. Tests use database 15 with unique queues and result keys, and delete only their own fixture data. They skip when `ASYNQ_TEST_REDIS_ADDR` is unset. Forced-kill recovery may take over one minute while the lease expires.

To check actual HTTP producer options, run the app with `ASYNQ_MAX_RETRY=1` and `ASYNQ_JOB_TIMEOUT=2s`, with no worker consuming its queue. Point the test at the same Redis database 0:

```bash
APP_HOST=localhost APP_PORT=8080 TEST_ASYNQ_REDIS_ADDR=localhost:6379 \
  TEST_ASYNQ_MAX_RETRY=1 TEST_ASYNQ_JOB_TIMEOUT=2s \
  go test -count=1 -v ./integration-test -run '^TestHTTPCRUDV1$'
```

This checks welcome-email and uploaded-image task metadata without changing successful HTTP responses. Existing Docker checks with `TEST_WORKER_ENABLED=true` verify successful processing by the actual worker. The Engines workflow runs these checks for each generated Gin, stdlib, and Fiber project. Exact proposed revisions and CI results are recorded in [#66](https://github.com/zoe606/gobase/issues/66).

## Production Checklist

- [ ] Set `APP_ENV=production`
- [ ] Set a secure `JWT_SECRET_KEY` (min 32 characters)
- [ ] Configure CORS origins (`CORS_ALLOW_ORIGINS`)
- [ ] Set up S3 storage (`STORAGE_DRIVER=s3`)
- [ ] Configure email provider (`EMAIL_PROVIDER=resend`)
- [ ] Disable Swagger (`SWAGGER_ENABLED=false`)
- [ ] Set proper rate limits

## Environment Variables

| Category | Key Variables |
|----------|--------------|
| **App** | `APP_ENV`, `HTTP_PORT` |
| **Database** | `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DBNAME` |
| **Redis** | `REDIS_HOST`, `REDIS_PORT` |
| **JWT** | `JWT_SECRET_KEY`, `JWT_ACCESS_EXPIRY`, `JWT_REFRESH_EXPIRY` |
| **Storage** | `STORAGE_DRIVER` (local/s3), `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` |
| **Email** | `EMAIL_PROVIDER` (resend/noop), `EMAIL_API_KEY`, `EMAIL_FROM` |

See `.env.example` for the complete list with defaults.

## Health Endpoints

### Redis Ownership and Shutdown

The application bootstrap creates separate Redis clients for the enabled application cache and Redis rate limiter. It closes both after HTTP shutdown. Cleanup is safe to call more than once and still closes the remaining client if one client is already closed.

The cache uses its client without owning its lifecycle. The rate limiter storage closes its own client when the bootstrap calls `Close`. Separate clients preserve that ownership and keep closing one store from closing the other. Both use the existing Redis address, password, and database settings.

Disabled cache creates no Redis cache client. Memory rate limiting uses the selected engine's memory store and creates no rate limiter Redis client. The Asynq queue client has its own existing shutdown path.

Readiness checks PostgreSQL. Cache and rate limiting retain their existing Redis error handling.

### Probes

| Endpoint | Purpose |
|----------|---------|
| `GET /healthz` | Liveness probe; returns `OK` with status 200 |
| `GET /readyz` | Readiness probe; checks PostgreSQL connectivity |
| `GET /metrics` | Prometheus metrics (when enabled) |
