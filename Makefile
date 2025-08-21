.PHONY: help dev build test test-unit test-integration lint migrate-create migrate-up migrate-down migrate-reset docker-build docker-up docker-down clean

# Load environment variables from .env
-include .env
export

# Application
APP_NAME=go-gin-auth
APP_VERSION?=latest

# Database URL for migrations
ifeq ($(DB_HOST),)
	DB_HOST=localhost
endif
ifeq ($(DB_USER),)
	DB_USER=root
endif
ifeq ($(DB_PASS),)
	DB_PASS=root
endif
ifeq ($(DB_NAME),)
	DB_NAME=authdb
endif
ifeq ($(DB_PORT),)
	DB_PORT=3306
endif

MIGRATE_DSN=mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?multiStatements=true

# Docker
DOCKER_IMAGE=$(APP_NAME):$(APP_VERSION)

## Show help
help:
	@echo ""
	@echo "Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

## Install dependencies
deps: ## Install Go dependencies
	go mod download
	go mod tidy
	go mod verify

## Run application in development mode
dev: ## Run app in development mode
	go run cmd/api/main.go

## Build application binary
build: ## Build application binary
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bin/$(APP_NAME) cmd/api/main.go

## Run tests with coverage
test: ## Run all tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Run unit tests only
test-unit: ## Run unit tests only
	go test -v -race -short ./...

## Run integration tests
test-integration: ## Run integration tests
	go test -v -race -run Integration ./...

## Run linter
lint: ## Run golangci-lint
	golangci-lint run

## Format code
fmt: ## Format Go code
	go fmt ./...

## Vet code
vet: ## Run go vet
	go vet ./...

## Create new migration
migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Please provide migration name: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	migrate create -ext sql -dir internal/storage/migrations -seq $(name)

## Run migrations up
migrate-up: ## Apply all pending migrations
	migrate -path internal/storage/migrations -database "$(MIGRATE_DSN)" up

## Run migrations down
migrate-down: ## Rollback migrations (usage: make migrate-down n=1)
	migrate -path internal/storage/migrations -database "$(MIGRATE_DSN)" down $(or $(n),1)

## Reset all migrations
migrate-reset: ## Reset all migrations (down then up)
	migrate -path internal/storage/migrations -database "$(MIGRATE_DSN)" down -all
	migrate -path internal/storage/migrations -database "$(MIGRATE_DSN)" up

## Force migration version
migrate-force: ## Force migration version (usage: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "Please provide version: make migrate-force version=1"; \
		exit 1; \
	fi
	migrate -path internal/storage/migrations -database "$(MIGRATE_DSN)" force $(version)

## Show migration version
migrate-version: ## Show current migration version
	migrate -path internal/storage/migrations -database "$(MIGRATE_DSN)" version

## Build Docker image
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

## Start services with Docker Compose
docker-up: ## Start all services with Docker Compose
	docker-compose up -d

## Stop Docker Compose services
docker-down: ## Stop all services
	docker-compose down

## View logs
logs: ## View application logs
	docker-compose logs -f api

## Run migrations in Docker
docker-migrate-up: ## Run migrations in Docker environment
	docker-compose exec api sh -c "MIGRATE_DSN='mysql://root:rootpassword@tcp(db:3306)/authdb?multiStatements=true' migrate -path internal/storage/migrations -database \$$MIGRATE_DSN up"

## Create admin user in Docker
docker-create-admin: ## Create admin user in Docker environment
	docker-compose exec api sh -c "go run scripts/create_admin.go"

## Clean up
clean: ## Clean build artifacts and Docker resources
	go clean
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker-compose down --volumes --remove-orphans
	docker system prune -f

## Setup development environment
setup: deps migrate-up ## Setup development environment
	@echo "Development environment setup complete!"
	@echo "Run 'make dev' to start the application"

## Security check
security: ## Run security checks
	gosec ./...

## Generate code (mocks, etc.)
generate: ## Generate code
	go generate ./...

## Install development tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## Check dependencies for updates
deps-check: ## Check for dependency updates
	go list -u -m all

## Update dependencies
deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy
