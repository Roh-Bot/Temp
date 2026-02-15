.PHONY: help build run test clean docker-up docker-down swagger migrate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	go build -o bin/task-manager cmd/blog-api/main.go

run: ## Run the application
	go run cmd/blog-api/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-up: ## Start docker containers
	docker-compose up -d

docker-down: ## Stop docker containers
	docker-compose down

docker-logs: ## View docker logs
	docker-compose logs -f api-go

swagger: ## Generate swagger documentation
	swag init -g cmd/blog-api/main.go

migrate: ## Run database migrations
	psql -h localhost -U postgres -d taskmanager -f migrations/001_init.sql

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

.DEFAULT_GOAL := help
