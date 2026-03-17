.PHONY: help build run test clean docker-up docker-down migrate-up migrate-down migrate-create sync

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build all binaries
	@echo "Building HTTP server..."
	go build -o bin/app ./cmd/app
	@echo "Building sync CLI..."
	go build -o bin/sync ./cmd/sync
	@echo "✅ Build complete"

run: ## Run HTTP server
	go run ./cmd/app/main.go

run-sync: ## Run sync service with example codes
	go run ./cmd/sync/main.go --supplier=4tochki --codes=2329500,WHS063930

sync: ## Run sync service (alias for run-sync)
	$(MAKE) run-sync

test: ## Run all tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

test-unit: ## Run only unit tests
	go test -v -short ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter (requires golangci-lint)
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run

vet: ## Run go vet
	go vet ./...

docker-build: ## Build Docker images
	docker-compose build

docker-up: ## Start all Docker containers
	docker-compose up -d
	@echo "✅ Services started"
	@echo "API: http://localhost:8080"
	@echo "Database: localhost:5432"

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-restart: ## Restart Docker containers
	docker-compose restart

docker-clean: ## Remove all containers, volumes, and images
	docker-compose down -v --rmi all

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	@if [ -z "$(DB_DSN)" ]; then \
		export DB_DSN="postgres://etalon:etalon_pass@localhost:5432/etalon_price?sslmode=disable"; \
	fi; \
	migrate -path internal/migrations -database "$$DB_DSN" up

migrate-down: ## Run database migrations down
	@echo "Rolling back migrations..."
	@if [ -z "$(DB_DSN)" ]; then \
		export DB_DSN="postgres://etalon:etalon_pass@localhost:5432/etalon_price?sslmode=disable"; \
	fi; \
	migrate -path internal/migrations -database "$$DB_DSN" down

migrate-create: ## Create new migration (use name=migration_name)
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=migration_name"; exit 1; fi
	migrate create -ext sql -dir internal/migrations -seq $(name)

migrate-force: ## Force migration version (use version=N)
	@if [ -z "$(version)" ]; then echo "Usage: make migrate-force version=N"; exit 1; fi
	@if [ -z "$(DB_DSN)" ]; then \
		export DB_DSN="postgres://etalon:etalon_pass@localhost:5432/etalon_price?sslmode=disable"; \
	fi; \
	migrate -path internal/migrations -database "$$DB_DSN" force $(version)

db-setup: ## Setup local database
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to start..."
	@sleep 5
	@echo "Running migrations..."
	$(MAKE) migrate-up
	@echo "✅ Database ready"

db-reset: ## Reset database (drop and recreate)
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to start..."
	@sleep 5
	$(MAKE) migrate-up

curl-health: ## Test health endpoint
	curl http://localhost:8080/healthz

curl-ready: ## Test readiness endpoint
	curl http://localhost:8080/readyz

curl-sync: ## Test sync endpoint with example codes
	curl -X POST http://localhost:8080/sync/4tochki \
		-H "Content-Type: application/json" \
		-d '{"codes":["2329500","WHS063930"]}'

dev: ## Start development environment
	$(MAKE) db-setup
	@echo "Starting application..."
	$(MAKE) run

.DEFAULT_GOAL := help
