# Deployment

## Docker

### Local Development

```bash
# Start infrastructure only (PostgreSQL, Redis, MinIO) — use with Air for hot reload
make docker-services

# Start full stack (infrastructure + app + worker)
make docker-dev

# Rebuild and start
make docker-dev-build

# View logs
make docker-logs

# Stop all containers
make docker-stop
```

### Production Build

```bash
# Build production binary
make build

# Run migrations on production
export PROD_DATABASE_URL='postgres://user:pass@host:5432/db?sslmode=require'
make migrate-prod
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
| **Database** | `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` |
| **Redis** | `REDIS_HOST`, `REDIS_PORT` |
| **JWT** | `JWT_SECRET_KEY`, `JWT_ACCESS_EXPIRY`, `JWT_REFRESH_EXPIRY` |
| **Storage** | `STORAGE_DRIVER` (local/s3), `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` |
| **Email** | `EMAIL_PROVIDER` (resend/noop), `EMAIL_API_KEY`, `EMAIL_FROM` |

See `.env.example` for the complete list with defaults.

## Health Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /healthz` | Liveness probe — returns `{"status":"ok"}` |
| `GET /readyz` | Readiness probe — checks DB and Redis connectivity |
| `GET /metrics` | Prometheus metrics (when enabled) |
