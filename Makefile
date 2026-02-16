.PHONY: help build run test clean docker-up docker-down swagger migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building the application"
	@echo "Current directory: $(CURDIR)"
	@echo $(wildcard ./internal/config/*)
	go build -ldflags="-s -w" -tags 'no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb netgo' -o bin/task-manager cmd/blog-api/main.go
	@echo "Make sure to change this command for LINUX/OSX"
	copy .\internal\config\config.yaml .\bin
	@echo "Build successful"

run: ## Run the application
	go run cmd/blog-api/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-unit: ## Run unit tests only
	go test -v ./cmd/api ./internal/application ./internal/worker

clean: ## Clean build artifacts
	rm -rf bin/

docker-up: ## Start docker containers
	docker-compose up -d

docker-down: ## Stop docker containers
	docker-compose down

docker-down-volume: ## Stop docker containers
	docker-compose down -v

docker-logs: ## View docker logs
	docker-compose logs -f api-go

swagger: ## Generate swagger documentation
	swag init -g cmd/blog-api/main.go

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

migrate-up:
	goose -dir ./migrations postgres "host=localhost port=5432 database=taskmanager user=postgres password=admin" up

migrate-down:
	goose -dir ./migrations postgres "host=localhost port=5432 database=taskmanager user=postgres password=admin" down


.DEFAULT_GOAL := help
