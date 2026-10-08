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
