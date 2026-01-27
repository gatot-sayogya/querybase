.PHONY: help build build-all build-api build-worker build-api-multi build-worker-multi run-api run-worker test clean docker-up docker-down migrate-up migrate-down

help: ## Display this help message
	@echo "QueryBase Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: build-api build-worker ## Build all binaries for native architecture

build-all: clean build-api-multi build-worker-multi ## Build all binaries for all architectures

build-api: ## Build API server binary for native architecture
	@echo "Building API server..."
	@go build -o bin/api ./cmd/api

build-worker: ## Build worker binary for native architecture
	@echo "Building worker..."
	@go build -o bin/worker ./cmd/worker

build-api-multi: ## Build API server for multiple architectures (arm64, amd64)
	@echo "Building API server for multiple architectures..."
	@mkdir -p bin
	@echo "  → Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 go build -o bin/api-linux-arm64 ./cmd/api
	@echo "  → Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build -o bin/api-linux-amd64 ./cmd/api
	@echo "  → Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build -o bin/api-darwin-arm64 ./cmd/api
	@echo "  → Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/api-darwin-amd64 ./cmd/api
	@echo "  → Building for windows/amd64..."
	@GOOS=windows GOARCH=amd64 go build -o bin/api-windows-amd64.exe ./cmd/api
	@echo "✅ API server built for all architectures"

build-worker-multi: ## Build worker for multiple architectures (arm64, amd64)
	@echo "Building worker for multiple architectures..."
	@mkdir -p bin
	@echo "  → Building for linux/arm64..."
	@GOOS=linux GOARCH=arm64 go build -o bin/worker-linux-arm64 ./cmd/worker
	@echo "  → Building for linux/amd64..."
	@GOOS=linux GOARCH=amd64 go build -o bin/worker-linux-amd64 ./cmd/worker
	@echo "  → Building for darwin/arm64..."
	@GOOS=darwin GOARCH=arm64 go build -o bin/worker-darwin-arm64 ./cmd/worker
	@echo "  → Building for darwin/amd64..."
	@GOOS=darwin GOARCH=amd64 go build -o bin/worker-darwin-amd64 ./cmd/worker
	@echo "  → Building for windows/amd64..."
	@GOOS=windows GOARCH=amd64 go build -o bin/worker-windows-amd64.exe ./cmd/worker
	@echo "✅ Worker built for all architectures"

run-api: ## Run API server
	@echo "Starting API server..."
	@go run ./cmd/api/main.go

run-worker: ## Run background worker
	@echo "Starting worker..."
	@go run ./cmd/worker/main.go

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-short: ## Run short tests only (skip database-dependent tests)
	@echo "Running short tests..."
	@go test -v -short ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep total
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-coverage-html: ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"
	@echo "   Open coverage.html in your browser to view"

test-bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -v -bench=. -benchmem ./...

test-integration: ## Run integration tests (requires database)
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

test-watch: ## Run tests on file changes (requires entr)
	@echo "Watching for changes..."
	@find . -name '*.go' | entr -c go test -v ./...

test-auth: ## Run auth package tests
	@echo "Running auth tests..."
	@go test -v ./internal/auth/...

test-service: ## Run service package tests
	@echo "Running service tests..."
	@go test -v ./internal/service/...

test-verbose-coverage: ## Run tests with detailed coverage output
	@echo "Running tests with detailed coverage..."
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out
	@echo ""
	@echo "Coverage by package:"
	@go tool cover -func=coverage.out | grep -E "^github.com/yourorg/querybase/internal/" | awk '{print "  " $$1 ": " $$3}'

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

list: ## List available build artifacts
	@echo "Available binaries:"
	@echo "=================="
	@if [ -d bin ]; then \
		for file in bin/*; do \
			if [ -f "$$file" ]; then \
				size=$$(ls -lh "$$file" | awk '{print $$5}'); \
				basename "$$file" | awk -v s="$$size" '{print "  " $$1 " (" s ")"}'; \
			fi; \
		done; \
	else \
		echo "  No binaries found. Run 'make build' first."; \
	fi

docker-up: ## Start Docker services (PostgreSQL, Redis)
	@echo "Starting Docker services..."
	@docker-compose -f docker/docker-compose.yml up -d

docker-down: ## Stop Docker services
	@echo "Stopping Docker services..."
	@docker-compose -f docker/docker-compose.yml down

docker-logs: ## View Docker logs
	@docker-compose -f docker/docker-compose.yml logs -f

migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@echo "  → Applying migration 000001: Initial schema..."
	@psql -h localhost -U querybase -d querybase -f migrations/000001_init_schema.up.sql 2>/dev/null || echo "    (Already applied or error occurred)"
	@echo "  → Applying migration 000002: Update query_results schema..."
	@psql -h localhost -U querybase -d querybase -f migrations/000002_update_query_results_schema.up.sql 2>/dev/null || echo "    (Already applied or error occurred)"
	@echo "  → Applying migration 000003: Remove caching, rename columns..."
	@psql -h localhost -U querybase -d querybase -f migrations/000003_remove_caching_rename_columns.up.sql 2>/dev/null || echo "    (Already applied or error occurred)"
	@echo "  → Applying migration 000004: Add query_transactions table..."
	@psql -h localhost -U querybase -d querybase -f migrations/000004_add_query_transactions.up.sql
	@echo "✅ Migrations applied successfully"

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@echo "  → Rolling back migration 000004..."
	@psql -h localhost -U querybase -d querybase -f migrations/000004_add_query_transactions.down.sql
	@echo "  → Rolling back migration 000003..."
	@psql -h localhost -U querybase -d querybase -f migrations/000003_remove_caching_rename_columns.down.sql
	@echo "  → Rolling back migration 000002..."
	@psql -h localhost -U querybase -d querybase -f migrations/000002_update_query_results_schema.down.sql
	@echo "  → Rolling back migration 000001..."
	@psql -h localhost -U querybase -d querybase -f migrations/000001_init_schema.down.sql
	@echo "✅ Migrations rolled back"

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@echo "  → query_results table schema:"
	@psql -h localhost -U querybase -d querybase -c "\d query_results" 2>/dev/null || echo "    Table does not exist"

db-shell: ## Open PostgreSQL shell
	@psql -h localhost -U querybase -d querybase

deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

.DEFAULT_GOAL := help
