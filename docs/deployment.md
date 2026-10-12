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
