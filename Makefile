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

##@ Docker

.PHONY: docker-up
docker-up: ## Start Docker containers (PostgreSQL)
	docker compose up -d db

.PHONY: docker-down
docker-down: ## Stop Docker containers
	docker compose down

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t go-boilerplate:latest .

.PHONY: docker-logs
docker-logs: ## View Docker logs
	docker compose logs -f

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
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf ./bin ./tmp coverage.txt coverage.html

##@ Pre-commit

.PHONY: pre-commit
pre-commit: fmt lint test ## Run all checks before commit
