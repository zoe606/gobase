# Go Boilerplate Makefile

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: dev
dev: ## Run with Air live reload (development mode)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air is not installed. Installing..."; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi

.PHONY: run
run: ## Run the application
	go run ./cmd/app -config ./config/config.yaml

.PHONY: build
build: ## Build the application
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/app ./cmd/app

##@ Code Quality

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	@if command -v gofumpt > /dev/null; then gofumpt -l -w .; fi

.PHONY: lint
lint: ## Run linter
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet..."; \
		go vet ./...; \
	fi

.PHONY: vuln
vuln: ## Run vulnerability check on dependencies
	@if command -v govulncheck > /dev/null; then \
		govulncheck ./...; \
	else \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

.PHONY: test
test: ## Run tests
	go test -v -race -covermode=atomic -coverprofile=coverage.txt ./internal/... ./pkg/...

.PHONY: test-integration
test-integration: ## Run integration tests
	go clean -testcache && go test -v ./integration-test/...

.PHONY: coverage
coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.txt -o coverage.html

##@ Database

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	@if command -v migrate > /dev/null; then \
		migrate -path migrations -database "$(POSTGRES_URL)" up; \
	else \
		echo "golang-migrate not installed. Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
	fi

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	@if command -v migrate > /dev/null; then \
		migrate -path migrations -database "$(POSTGRES_URL)" down 1; \
	else \
		echo "golang-migrate not installed."; \
	fi

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	@if command -v migrate > /dev/null; then \
		migrate create -ext sql -dir migrations -seq $(name); \
	else \
		echo "golang-migrate not installed."; \
	fi

.PHONY: migrate-status
migrate-status: ## Show current migration version
	@if command -v migrate > /dev/null; then \
		migrate -path migrations -database "$(POSTGRES_URL)" version; \
	else \
		echo "golang-migrate not installed."; \
	fi

.PHONY: migrate-force
migrate-force: ## Force migration version (usage: make migrate-force version=N)
	@if command -v migrate > /dev/null; then \
		migrate -path migrations -database "$(POSTGRES_URL)" force $(version); \
	else \
		echo "golang-migrate not installed."; \
	fi

.PHONY: migrate-validate
migrate-validate: ## Validate migration files have matching up/down pairs
	@echo "Checking migration file pairs..."
	@for f in migrations/*.up.sql; do \
		down="$${f%.up.sql}.down.sql"; \
		if [ ! -f "$$down" ]; then \
			echo "Missing down migration: $$down"; \
			exit 1; \
		fi; \
	done
	@echo "All migrations have matching up/down pairs."

##@ Environment Migrations

.PHONY: migrate-staging
migrate-staging: ## Run migrations on staging
	@if [ -z "$$STAGING_DATABASE_URL" ]; then \
		echo "Error: STAGING_DATABASE_URL not set"; \
		echo "Export: export STAGING_DATABASE_URL='postgres://...'"; \
		exit 1; \
	fi
	@echo "=== Running migrations on STAGING ==="
	migrate -path migrations -database "$$STAGING_DATABASE_URL" up
	migrate -path migrations -database "$$STAGING_DATABASE_URL" version

.PHONY: migrate-staging-down
migrate-staging-down: ## Rollback 1 migration on staging
	@if [ -z "$$STAGING_DATABASE_URL" ]; then echo "Error: STAGING_DATABASE_URL not set"; exit 1; fi
	migrate -path migrations -database "$$STAGING_DATABASE_URL" down 1

.PHONY: migrate-prod
migrate-prod: ## Run migrations on production (requires confirmation)
	@if [ -z "$$PROD_DATABASE_URL" ]; then \
		echo "Error: PROD_DATABASE_URL not set"; exit 1; \
	fi
	@echo "WARNING: About to run migrations on PRODUCTION"
	@echo "Database: $$PROD_DATABASE_URL" | sed 's/:.*@/:***@/'
	@read -p "Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || exit 1
	migrate -path migrations -database "$$PROD_DATABASE_URL" up
	migrate -path migrations -database "$$PROD_DATABASE_URL" version

.PHONY: migrate-prod-down
migrate-prod-down: ## Rollback 1 migration on production (requires confirmation)
	@if [ -z "$$PROD_DATABASE_URL" ]; then echo "Error: PROD_DATABASE_URL not set"; exit 1; fi
	@echo "WARNING: About to ROLLBACK on PRODUCTION"
	@read -p "Type 'yes' to confirm: " confirm && [ "$$confirm" = "yes" ] || exit 1
	migrate -path migrations -database "$$PROD_DATABASE_URL" down 1

##@ Docker (Local Development)

.PHONY: docker-services
docker-services: ## Start DB and Redis (for air users)
	docker compose -f deployment/docker/docker-compose.yml --env-file .env up -d

.PHONY: docker-dev
docker-dev: ## Start full stack in Docker (DB + Redis + App + Worker)
	docker compose -f deployment/docker/docker-compose.yml -f deployment/docker/docker-compose.app.yml --env-file .env up -d

.PHONY: docker-dev-build
docker-dev-build: ## Rebuild and start full stack
	docker compose -f deployment/docker/docker-compose.yml -f deployment/docker/docker-compose.app.yml --env-file .env up -d --build

.PHONY: docker-stop
docker-stop: ## Stop all containers
	docker compose -f deployment/docker/docker-compose.yml -f deployment/docker/docker-compose.app.yml down 2>/dev/null || docker compose -f deployment/docker/docker-compose.yml down

.PHONY: docker-logs
docker-logs: ## View Docker logs
	docker compose -f deployment/docker/docker-compose.yml -f deployment/docker/docker-compose.app.yml logs -f

.PHONY: docker-monitoring
docker-monitoring: ## Start services with Asynqmon dashboard
	docker compose -f deployment/docker/docker-compose.yml --profile monitoring --env-file .env up -d

##@ Worker

.PHONY: run-worker
run-worker: ## Run Asynq worker locally
	go run ./cmd/worker -config ./config/config.yaml

.PHONY: build-worker
build-worker: ## Build worker binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/worker ./cmd/worker

##@ Documentation

.PHONY: swag
swag: ## Generate Swagger documentation
	@if command -v swag > /dev/null; then \
		swag init -g internal/handlers/http/router.go; \
	else \
		echo "swag not installed. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

##@ Dependencies

.PHONY: deps
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy
	go mod verify

.PHONY: deps-update
deps-update: ## Update all dependencies
	go get -u ./...
	go mod tidy

.PHONY: tools
tools: ## Install development tools
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install go.uber.org/mock/mockgen@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf ./bin ./tmp coverage.txt coverage.html

##@ Code Generation

.PHONY: generate
generate: ## Generate mocks and other code
	go generate ./...

##@ Git Hooks

.PHONY: install-hooks
install-hooks: ## Install git hooks for pre-commit and pre-push checks
	@echo "Installing git hooks..."
	@cp .githooks/pre-commit .git/hooks/pre-commit 2>/dev/null || \
		(mkdir -p .git/hooks && cp .githooks/pre-commit .git/hooks/pre-commit)
	@cp .githooks/pre-push .git/hooks/pre-push 2>/dev/null || \
		(mkdir -p .git/hooks && cp .githooks/pre-push .git/hooks/pre-push)
	@chmod +x .git/hooks/pre-commit .git/hooks/pre-push 2>/dev/null || true
	@echo "Git hooks installed successfully!"
	@echo "  - pre-commit: Fast checks (format, vet, basic lint)"
	@echo "  - pre-push: Full checks (build, test, lint, vuln)"

.PHONY: uninstall-hooks
uninstall-hooks: ## Remove git hooks
	@rm -f .git/hooks/pre-commit .git/hooks/pre-push
	@echo "Git hooks removed"

##@ Pre-commit

.PHONY: pre-commit
pre-commit: fmt lint test ## Run all checks before commit

.PHONY: check-all
check-all: fmt lint vuln test ## Run all quality checks including vulnerability scan

.PHONY: ci
ci: fmt lint vuln test build ## Run full CI pipeline locally
