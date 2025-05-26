# AI RPG MVP - Context Tracking System
# Development Makefile

.PHONY: help build test run clean docker install deps fmt lint vet coverage examples

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the project"
	@echo "  test      - Run tests"
	@echo "  run       - Run the web server example"
	@echo "  examples  - Run basic usage example"
	@echo "  clean     - Clean build artifacts"
	@echo "  docker    - Build Docker image"
	@echo "  install   - Install dependencies"
	@echo "  deps      - Download dependencies"
	@echo "  fmt       - Format code"
	@echo "  lint      - Run linter"
	@echo "  vet       - Run go vet"
	@echo "  coverage  - Run tests with coverage"

# Build targets
build:
	@echo "Building project..."
	go build -v ./...

test:
	@echo "Running tests..."
	go test -v ./...

test-race:
	@echo "Running tests with race detection..."
	go test -race -v ./...

coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Development targets
run:
	@echo "Starting web server..."
	go run examples/web_server.go

examples:
	@echo "Running basic usage example..."
	go run examples/basic_usage.go

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Running go vet..."
	go vet ./...

lint:
	@echo "Running golint..."
	@command -v golint >/dev/null 2>&1 || { echo "Installing golint..."; go install golang.org/x/lint/golint@latest; }
	golint ./...

# Dependency management
deps:
	@echo "Downloading dependencies..."
	go mod download

install: deps
	@echo "Installing tools..."
	go install golang.org/x/lint/golint@latest
	go install golang.org/x/tools/cmd/goimports@latest

tidy:
	@echo "Tidying up go.mod..."
	go mod tidy

# Docker targets
docker:
	@echo "Building Docker image..."
	docker build -t ai-rpg-context-tracker .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 ai-rpg-context-tracker

# Benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./context

# Database targets
db-up:
	@echo "Starting PostgreSQL with Docker..."
	docker run --name postgres-rpg \
		-e POSTGRES_USER=rpguser \
		-e POSTGRES_PASSWORD=rpgpass \
		-e POSTGRES_DB=rpgdb \
		-p 5432:5432 \
		-d postgres:15

db-down:
	@echo "Stopping PostgreSQL..."
	docker stop postgres-rpg || true
	docker rm postgres-rpg || true

# Integration tests (requires database)
test-integration: db-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	@echo "Running integration tests..."
	POSTGRES_URL="postgres://rpguser:rpgpass@localhost:5432/rpgdb?sslmode=disable" \
		go test -tags=integration -v ./...
	@make db-down

# Performance testing
load-test:
	@echo "Running load test (requires web server to be running)..."
	@command -v hey >/dev/null 2>&1 || { echo "Please install 'hey' load testing tool"; exit 1; }
	hey -n 1000 -c 10 -m POST -H "Content-Type: application/json" \
		-d '{"player_id":"load_test","player_name":"LoadTest"}' \
		http://localhost:8080/api/session/create

# Clean up
clean:
	@echo "Cleaning up..."
	go clean
	rm -f coverage.out coverage.html
	rm -rf bin/

# All quality checks
check: fmt vet lint test
	@echo "All checks passed!"

# Full CI pipeline
ci: deps check test-race coverage
	@echo "CI pipeline completed!"

# Development setup
dev-setup: install
	@echo "Development environment setup complete!"
	@echo "Run 'make run' to start the web server"
	@echo "Run 'make examples' to see basic usage"
	@echo "Run 'make test' to run tests"
