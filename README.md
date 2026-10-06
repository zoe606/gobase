# Go Clean Architecture Boilerplate

Production-ready REST API boilerplate with Clean Architecture, JWT authentication, media uploads, and background workers.

[![Go Version](https://img.shields.io/badge/Go-1.27.1-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![HTTP Engines](https://img.shields.io/badge/HTTP-Gin%20%7C%20net%2Fhttp%20%7C%20Fiber-00ACD7?logo=go)](docs/http-engines.md)
[![GORM](https://img.shields.io/badge/GORM-ORM-00ADD8)](https://gorm.io/)
[![Swagger](https://img.shields.io/badge/Swagger-API%20Docs-85EA2D?logo=swagger)](https://swagger.io/)

## Features

- **Authentication** — JWT access/refresh tokens, email verification, password reset, roles & permissions
- **Media Management** — S3/Local storage, polymorphic attachments, image processing, presigned URLs
- **Articles** — Full CRUD with draft/publish workflow, cover images, SEO slugs
- **Translation** — Google Translate API integration with history
- **Background Jobs** — Asynq workers for emails, image processing
- **Code Generation** — Scaffold all layers from a single migration file
- **HTTP Engine Selection** — Create a project with Gin, standard `net/http`, or Fiber

## Create a Project

Install Go 1.27.1 or newer. Select the engine when creating the project:

```bash
git clone https://github.com/zoe606/gobase.git gobase
cd gobase
make init ENGINE=gin MODULE=github.com/your-org/myapp APP_NAME=myapp OUTPUT=../myapp
```

Gin v1.12.0 is the default. Use `ENGINE=stdlib` for `http.ServeMux`, or `ENGINE=fiber` for Fiber v2. Each output contains one engine. See [HTTP Engines](docs/http-engines.md) for the architecture and compatibility rules.

## Run a Project

```bash
# Configure the generated project
cd ../myapp
cp .env.example .env
cp config/config.example.yaml config/config.yaml

# Start infrastructure
make docker-services

# Run the application
make run
```

The development app applies database migrations on startup. Start the worker in another terminal with `make run-worker`. The `.env` file supplies Docker Compose variables. Local Go commands read `config/config.yaml` and exported environment variables; update both configurations when changing connection ports.

New projects use local storage. MinIO is optional; see [Deployment](docs/deployment.md) for S3 configuration.

Verify it's running:

```bash
curl http://localhost:8080/healthz    # Liveness probe
curl http://localhost:8080/readyz     # Readiness probe (checks PostgreSQL)
```

## API Documentation

Once the app is running, Swagger UI is available at:

```
http://localhost:8080/swagger/
```

All endpoints, request/response schemas, and authentication requirements are documented there.

`GET /v1/articles` accepts an omitted or empty `status`, or the exact values `draft` and `published`. Other values return HTTP 400 with `VALIDATION_ERROR` and field details. Malformed queries retain `INVALID_QUERY`. Pagination continues to use `page` and `limit` with the existing defaults and normalization.

## Architecture

```mermaid
flowchart TB
    subgraph Client
        Browser[Browser/Mobile]
    end

    subgraph "Go Application"
        subgraph "HTTP Layer"
            Server[Selected HTTP Engine]
            MW[Middleware<br/>JWT, CORS, RateLimit]
            Handlers[Handlers]
        end

        subgraph "Business Layer"
            UseCases[Use Cases]
        end

        subgraph "Data Layer"
            Repos[Repositories]
        end
    end

    subgraph "Infrastructure"
        PG[(PostgreSQL)]
        Redis[(Redis)]
        S3[(MinIO/S3)]
    end

    subgraph "Background"
        Worker[Asynq Worker]
        Queue[Task Queue]
    end

    Browser --> Server
    Server --> MW --> Handlers
    Handlers --> UseCases
    UseCases --> Repos
    Repos --> PG
    UseCases --> Redis
    UseCases --> S3
    UseCases --> Queue
    Queue --> Worker
    Worker --> PG
    Worker --> S3
```

### Layer Flow

```
Handler → UseCase → Repository → Entity/External API
```

- **Handlers** — Parse requests, validate input, return JSON responses
- **Use Cases** — Business logic, orchestrate repositories, return domain errors
- **Repositories** — Data access (PostgreSQL, S3, external APIs)
- **Entities** — GORM domain models

## Project Structure

```
├── cmd/
│   ├── app/                    # HTTP server entrypoint
│   └── worker/                 # Background worker entrypoint
├── config/                     # Configuration (Viper)
├── internal/
│   ├── app/                    # DI container & bootstrap
│   ├── dto/                    # Request/Response DTOs
│   ├── entity/                 # GORM domain models
│   ├── handlers/http/          # HTTP handlers for the selected engine
│   │   ├── middleware/
│   │   └── v1/
│   ├── repo/                   # Repository implementations
│   │   ├── persistent/         # PostgreSQL repos
│   │   ├── storage/            # S3/Local file storage
│   │   └── webapi/             # External APIs
│   ├── usecase/                # Business logic
│   └── worker/                 # Asynq task handlers
├── pkg/                        # Reusable packages
├── migrations/                 # SQL migration files
├── docs/                       # Swagger documentation
└── deployment/docker/          # Docker configuration
```

Each usecase method gets its own file with a corresponding test file (SOLID principle):

```
internal/usecase/auth/
├── auth.go            # Struct + constructor
├── errors.go          # Domain errors
├── login.go           # Login method
├── login_test.go      # Login tests
├── register.go
├── register_test.go
└── ...
```

## Configuration

Copy `.env.example` and configure the required variables. See [Deployment](docs/deployment.md) for the full environment variable reference and production checklist.

## Commands

### Development

| Command | Description |
|---------|-------------|
| `make run` | Run application |
| `make dev` | Run with Air hot reload |
| `make build` | Build binary |
| `make run-worker` | Run background worker |

### Code Quality

| Command | Description |
|---------|-------------|
| `make check-all` | Run all checks (format, lint, vuln, test) |
| `make fmt` | Format code |
| `make lint` | Run linter |
| `make vuln` | Check vulnerabilities |
| `make test` | Run unit tests |

### Database

| Command | Description |
|---------|-------------|
| `make migrate-up` | Apply migrations |
| `make migrate-down` | Rollback 1 migration |
| `make migrate-create name=X` | Create new migration |
| `make migrate-status` | Show current version |

### Code Generation

| Command | Description |
|---------|-------------|
| `make gen-full MIGRATION=X` | Generate all layers from migration |
| `make wire` | Auto-wire DI, routes, and contracts |
| `make generate` | Regenerate mocks |
| `make swag` | Regenerate Swagger docs |

### Docker

| Command | Description |
|---------|-------------|
| `make docker-services` | Start PostgreSQL and Redis |
| `make docker-services-s3` | Start services with optional MinIO |
| `make docker-dev` | Start full stack |
| `make docker-stop` | Stop all containers |
| `make docker-logs` | View container logs |

## Adding a New Feature

```bash
make migrate-create name=create_orders       # 1. Create migration
# Edit the .up.sql and .down.sql files        # 2. Write SQL
make migrate-up                               # 3. Run migration
make gen-full MIGRATION=000012               # 4. Generate all layers
make wire                                     # 5. Auto-wire DI, routes, contracts
make check-all                                # 6. Verify everything
make swag                                     # 7. Regenerate Swagger docs
```

## Testing

```bash
make test                # Unit tests with coverage
make test-integration    # Integration tests (requires a running app, PostgreSQL, and Redis)
make coverage            # Generate HTML coverage report
make generate            # Regenerate mocks after interface changes
```

## Further Reading

| Document | Description |
|----------|-------------|
| [HTTP Engines](docs/http-engines.md) | Engine selection, independent templates, shared core, and compatibility |
| [Roadmap](docs/roadmap.md) | Current baseline, release readiness, and proposed application hardening |
| [Release Readiness](docs/release-readiness.md) | Phase 1 implementation, local runtime results, and passing CI evidence |
| [Code Patterns](docs/code-patterns.md) | Error handling, validation, transactions, response format, SOLID file organization, reusable packages |
| [Deployment](docs/deployment.md) | Docker setup, production build, environment variables, production checklist |

## License

MIT License — see [LICENSE](LICENSE) for details.
