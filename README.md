# gobase

Go HTTP template with full application and minimal service profiles. Select Gin, standard `net/http`, or Fiber when creating a project. The default full profile includes Clean Architecture, JWT authentication, media uploads, and background workers.

[![Go Version](https://img.shields.io/badge/Go-1.27.2-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![HTTP Engines](https://img.shields.io/badge/HTTP-Gin%20%7C%20net%2Fhttp%20%7C%20Fiber-00ACD7?logo=go)](docs/http-engines.md)
[![GORM](https://img.shields.io/badge/GORM-ORM-00ADD8)](https://gorm.io/)
[![Swagger](https://img.shields.io/badge/Swagger-API%20Docs-85EA2D?logo=swagger)](https://swagger.io/)

## Features

The following application features belong to the default `full` profile. The `minimal` profile includes configuration, logging, native HTTP routing, request and response helpers, health probes, and graceful shutdown without required infrastructure.

- **Authentication** — Register, login, JWT access/refresh tokens, logout, and current user endpoints
- **Authorization** — JWT role and permission middleware
- **Media Management** — S3/Local storage, polymorphic attachments, image processing, presigned URLs
- **Articles** — Full CRUD with draft/publish workflow, cover images, SEO slugs
- **Translation** — Google Translate API integration with history
- **Background Jobs** — Asynq workers for emails, image processing
- **Optional WebSocket** — JWT handshake, native upgrades, bounded sessions, and active shutdown in Full projects
- **Optional gRPC** — A separate JWT-authenticated Echo server, generated Go client, health checks, and TLS in Full projects
- **Code Generation** — Scaffold all layers from a single migration file
- **HTTP Engine Selection** — Create a project with Gin, standard `net/http`, or Fiber

## Project Status

`v0.2.0` is the current template release. It includes Full and Minimal output for Gin, standard `net/http`, and Fiber. Full projects add the verified Asynq task policy, optional JWT-authenticated WebSocket, and an optional application gRPC server and client. It retains the HTTP panic recovery fix from v0.1.1. See the [release report](docs/release-readiness.md#v020-capability-release) and [release notes](https://github.com/zoe606/gobase/releases/tag/v0.2.0) for compatibility, dependency pins, and verification.

Release readiness is complete under [#85](https://github.com/zoe606/gobase/issues/85). All six fresh engine/profile outputs and documented quick starts passed. The three independent mini-apps also passed their native and Docker capability checks. The release notes identify the exact tag commit and its passing CI runs. Generated applications own their code and require explicit adoption of template updates.

The HTTP engine refactor and Phase 1 release readiness work are merged into `master` through [PR #46](https://github.com/zoe606/gobase/pull/46). Quality and all three engine jobs passed after the merge. Verification covers project generation, native handlers, CRUD generation, Swagger, HTTP integration, and Docker app and worker execution. See the [release readiness report](docs/release-readiness.md) for the tested revisions and workflow links.

Engine foundation is merged through [PR #53](https://github.com/zoe606/gobase/pull/53), [PR #57](https://github.com/zoe606/gobase/pull/57), and [PR #58](https://github.com/zoe606/gobase/pull/58). Minimal service generation merged in [PR #59](https://github.com/zoe606/gobase/pull/59). Go 1.27.2 and x/net v0.60.0 are the current dependency baseline. Docker images run as UID and GID `65532`.

Three standalone Commerce applications test the foundation: Order with Gin, Inventory with stdlib, and Billing with Fiber. Each has an independent repository, module, database, and native handlers. All ten HTTP workflow scenarios passed locally and in Ubuntu CI with native processes and Docker containers. The Linux workflow merged through [Order PR #6](https://github.com/zoe606/commerce-order/pull/6), and both modes passed after merge. Issues #60 and #56 are closed. See the [Commerce verification report](docs/commerce-verification.md) for revisions, commands, results, and limits. Further Commerce business features and monorepo generation are outside the active plan.

The [capability plan](docs/plans/foundation-capabilities.md) tracks Asynq failure and recovery, optional WebSocket support, and an optional application gRPC server. Asynq work merged through [PR #70](https://github.com/zoe606/gobase/pull/70), with passing source and six engine/profile jobs. See [background task behavior](docs/deployment.md#background-tasks) and [#66](https://github.com/zoe606/gobase/issues/66) for commands and verification.

WebSocket work merged through [PR #77](https://github.com/zoe606/gobase/pull/77), with all seven checks passing before and after merge. Full projects include optional `GET /v1/ws` with existing JWT access tokens, explicit allowed origins, echo and notification frames, connection limits, and graceful shutdown. See [WebSocket setup](docs/websocket.md) and [#67](https://github.com/zoe606/gobase/issues/67) for scope and verification.

Full projects also include optional `cmd/grpc` and `cmd/grpc-client`, one unary Echo RPC, shared JWT configuration, verified TLS, standard health checks, and bounded shutdown. The gRPC process and port are separate from the chosen HTTP engine. Both optional capabilities are disabled by default. See [gRPC setup](docs/grpc.md) and [#68](https://github.com/zoe606/gobase/issues/68) for configuration and verification. These capabilities are included in v0.2.0. Minimal startup remains free of required infrastructure.

The independent Commerce mini-apps also use a technical Probe RPC to verify calls between Order (Gin), Inventory (stdlib), and Billing (Fiber). Their native and Docker checks exercise six directed pairs, three service chains, HTTP-issued JWTs, verified TLS, bounded peer errors, and recovery. Each repository generates its own protobuf code. See the [application verification report](docs/commerce-verification.md#application-grpc-interoperability) and [#79](https://github.com/zoe606/gobase/issues/79) for revisions and CI evidence.

The [Asynq application follow-up](https://github.com/zoe606/gobase/issues/81) verifies normal workers and controlled task failures in those same independent repositories. The [WebSocket application follow-up](https://github.com/zoe606/gobase/issues/83) verifies each actual HTTP process and client example through thirty native/Docker connection scenarios. See the [application report](docs/commerce-verification.md#application-websocket-verification) for revisions, CI evidence, and limits. The [capability plan](docs/plans/foundation-capabilities.md#after-phase-6) records completed readiness and continued maintenance of these verified paths.

Email verification and password reset have use cases and token persistence, but their HTTP routes and queued email tasks still need to be connected. These flows and OpenTelemetry metrics export remain deferred application features.

## Create a Project

Install Go 1.27.2 or newer. Select the engine when creating the project:

```bash
git clone --branch v0.2.0 https://github.com/zoe606/gobase.git gobase
cd gobase
make init ENGINE=gin MODULE=github.com/your-org/myapp APP_NAME=myapp OUTPUT=../myapp
```

Gin v1.12.0 is the default. Use `ENGINE=stdlib` for `http.ServeMux`, or `ENGINE=fiber` for Fiber v2. Each output contains one engine. See [HTTP Engines](docs/http-engines.md) for the architecture and compatibility rules.

`PROFILE=full` is the default and preserves existing creation commands. To create a service without bundled business features or infrastructure:

```bash
make init PROFILE=minimal ENGINE=stdlib MODULE=github.com/your-org/service APP_NAME=service OUTPUT=../service
cd ../service
make run
```

Both `/healthz` and `/readyz` return `OK`. No database, Redis, storage, email sender, or worker is required. The generated README describes its native handler API and app-only Docker setup. Configuration files are optional. Export environment variables to override defaults. SQL CRUD and persistence wiring commands return a clear error because persistence is omitted. See [minimal services](docs/minimal-services.md) for the file and dependency set.

## Run a Project

These instructions apply to the `full` profile. Minimal services use the commands above and their generated README.

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

HTTP endpoints, request/response schemas, and authentication requirements are documented there. The optional WebSocket handshake and frame contract are documented in [WebSocket setup](docs/websocket.md).

`GET /v1/articles` accepts an omitted or empty `status`, or the exact values `draft` and `published`. Other values return HTTP 400 with `VALIDATION_ERROR` and field details. Malformed queries retain `INVALID_QUERY`. Pagination continues to use `page` and `limit` with the existing defaults and normalization.

## Architecture

The diagram below describes the full application. Minimal services keep only configuration, logging, the selected native HTTP engine, HTTP helpers, and process lifecycle. Both profiles use the generic `response.Response[T]` envelope.

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
        Storage[(Local files / S3)]
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
    UseCases --> Storage
    UseCases --> Queue
    Queue --> Worker
    Worker --> PG
    Worker --> Storage
```

### Layer Flow

```
Handler → UseCase → Repository → Entity/External API
```

- **Handlers** — Parse requests, validate input, return JSON responses
- **Use Cases** — Business logic, orchestrate repositories, return domain errors
- **Repositories** — Data access (PostgreSQL, file storage, external APIs)
- **Entities** — GORM domain models

## Project Structure

The directories below describe the full profile. Generated HTTP handlers use the selected engine. The template checkout runs Fiber; full engine sources are maintained under `pkg/scaffold/templates/{gin,stdlib,fiber}`, minimal sources under `pkg/scaffold/templates/minimal`, and CRUD templates under `pkg/codegen/generator/templates/handler`.

```
├── cmd/
│   ├── app/                    # HTTP server entrypoint
│   └── worker/                 # Background worker entrypoint
├── config/                     # Configuration (Viper)
├── internal/
│   ├── app/                    # DI container & bootstrap
│   ├── dto/                    # Request/Response DTOs
│   ├── entity/                 # GORM domain models
│   ├── handlers/http/          # Native HTTP handlers
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
├── docs/                       # Swagger and project documentation
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

Copy both example files as shown in [Run a Project](#run-a-project). Docker Compose reads `.env`. Local Go commands read `config/config.yaml` and exported environment variables. See [Deployment](docs/deployment.md) for storage options, environment variables, and production configuration.

## Commands

The tables below describe full-profile commands. Run application commands from the generated project's directory. Project creation, `make test-engines`, and `make test-minimal` run from the template checkout. Minimal services provide their own smaller Makefile.

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
| `make test-engines` | Verify generated Gin, stdlib, and Fiber projects from the template checkout |
| `make test-minimal` | Verify minimal output, HTTP contracts, startup, and shutdown for all three engines without infrastructure |

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
| `make docker-dev-build` | Rebuild and start full stack |
| `make docker-stop` | Stop all containers |
| `make docker-logs` | View container logs |

The app and worker run as UID and GID `65532`. Follow [runtime permissions](docs/deployment.md#runtime-permissions) for the local configuration bind mount and any existing upload volume before starting the full stack.

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
| [Roadmap](docs/roadmap.md) | Engine foundation, standalone application validation, and HTTP workflow plans |
| [Release Readiness](docs/release-readiness.md) | Release baselines, v0.2.0 verification, runtime results, and CI evidence |
| [Code Patterns](docs/code-patterns.md) | Error handling, validation, transactions, response format, SOLID file organization, reusable packages |
| [Deployment](docs/deployment.md) | Docker setup, production build, environment variables, production checklist |

## License

MIT License — see [LICENSE](LICENSE) for details.
