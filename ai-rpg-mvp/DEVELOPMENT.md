# Development Guide

This guide will help you set up and contribute to the AI RPG Context Tracking System.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Software

- **Go 1.21+**: [Download and install Go](https://golang.org/dl/)
- **PostgreSQL 15+**: [Install PostgreSQL](https://www.postgresql.org/download/)
- **Docker**: [Install Docker](https://docs.docker.com/get-docker/) (for containerized development)
- **Git**: Version control
- **Make**: Build automation (usually pre-installed on Unix systems)

### Optional Tools

- **Redis**: For caching (can be disabled in development)
- **hey**: For load testing (`go install github.com/rakyll/hey@latest`)
- **golint**: For code linting (`go install golang.org/x/lint/golint@latest`)
- **Air**: For live reloading (`go install github.com/cosmtrek/air@latest`)

## Getting Started

### 1. Clone the Repository

```bash
git clone <repository-url>
cd ai-rpg-mvp
```

### 2. Environment Setup

Copy the example environment file and customize it:

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Database Setup

#### Option A: Using Docker (Recommended)

```bash
# Start PostgreSQL with Docker Compose
docker-compose up -d postgres

# Or start everything
docker-compose up -d
```

#### Option B: Local PostgreSQL

```bash
# Create database
createdb rpgdb

# Create user (optional)
psql -c "CREATE USER rpguser WITH PASSWORD 'rpgpass';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE rpgdb TO rpguser;"

# Run initialization script
psql -d rpgdb -f scripts/init.sql
```

### 4. Install Dependencies

```bash
# Download Go modules
make deps

# Install development tools
make install
```

### 5. Run the Application

```bash
# Development server
make run

# Or run basic example
make examples
```

The web server will be available at `http://localhost:8080`.

## Development Workflow

### Daily Development

1. **Start the development environment:**
   ```bash
   docker-compose up -d postgres  # Start database
   make run                       # Start application
   ```

2. **Make changes to the code**

3. **Run tests:**
   ```bash
   make test
   ```

4. **Check code quality:**
   ```bash
   make check  # Runs fmt, vet, lint, and test
   ```

### Live Reloading

For faster development, you can use Air for live reloading:

```bash
# Install air if not already installed
go install github.com/cosmtrek/air@latest

# Create air configuration
cat > .air.toml << EOF
root = "."
cmd = "go run examples/web_server.go"
bin = "tmp/main"
include_ext = ["go", "tpl", "tmpl", "html"]
exclude_dir = ["tmp", "vendor"]
EOF

# Start with live reloading
air
```

### Project Structure

```
ai-rpg-mvp/
├── context/                 # Core context tracking package
│   ├── types.go            # Data structures
│   ├── manager.go          # Context manager
│   ├── events.go           # Event processing
│   ├── storage.go          # Storage implementations
│   ├── ai_integration.go   # AI prompt generation
│   └── manager_test.go     # Tests
├── config/                 # Configuration management
│   └── config.go          # Configuration structures
├── examples/               # Usage examples
│   ├── basic_usage.go     # Command-line example
│   └── web_server.go      # Web API server
├── scripts/                # Database and deployment scripts
│   └── init.sql           # Database initialization
├── .github/workflows/      # CI/CD pipelines
├── docker-compose.yml      # Development environment
├── Dockerfile             # Container image
├── Makefile              # Build automation
└── README.md             # Project documentation
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Run tests with coverage
make coverage

# Run benchmarks
make bench

# Run integration tests (requires database)
make test-integration
```

### Writing Tests

- Place test files in the same package as the code they test
- Use `_test.go` suffix for test files
- Follow Go testing conventions
- Include both unit tests and integration tests
- Aim for >80% test coverage

Example test structure:

```go
func TestContextManager_CreateSession(t *testing.T) {
    // Setup
    storage := NewMemoryStorage()
    cm := NewContextManager(storage)
    defer cm.Shutdown()

    // Test
    sessionID, err := cm.CreateSession("player123", "TestPlayer")
    
    // Assertions
    if err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }
    // ... more assertions
}
```

### Test Categories

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **Benchmark Tests**: Performance testing
- **End-to-End Tests**: Full system testing

## Deployment

### Docker Deployment

```bash
# Build image
make docker

# Run with docker-compose
docker-compose up -d

# Or deploy to production
docker-compose -f docker-compose.prod.yml up -d
```

### Manual Deployment

```bash
# Build binary
go build -o ai-rpg-server examples/web_server.go

# Set environment variables
export POSTGRES_URL="postgres://user:pass@host:5432/db"
export AI_API_KEY="your_api_key"

# Run
./ai-rpg-server
```

### Kubernetes Deployment

See `k8s/` directory for Kubernetes manifests (if available).

## Contributing

### Code Style

- Follow Go conventions and idioms
- Use `gofmt` for formatting
- Run `go vet` and `golint`
- Write clear, descriptive comments
- Keep functions small and focused

### Git Workflow

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and commit:**
   ```bash
   git add .
   git commit -m "Add feature: your feature description"
   ```

3. **Push and create pull request:**
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Ensure CI passes** before requesting review

### Pull Request Guidelines

- Write clear PR descriptions
- Include tests for new features
- Update documentation as needed
- Keep PRs focused and small
- Ensure CI/CD pipeline passes

### Commit Message Format

Use conventional commits:

```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
- `feat(context): add NPC relationship tracking`
- `fix(storage): handle database connection errors`
- `docs(readme): update installation instructions`

## Troubleshooting

### Common Issues

#### Database Connection Errors

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

#### Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy up go.mod
go mod tidy
```

#### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

#### Test Failures

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestContextManager_CreateSession ./context

# Check test coverage
go test -cover ./...
```

### Performance Issues

#### Memory Usage

```bash
# Check memory usage
go tool pprof http://localhost:8080/debug/pprof/heap

# Check goroutines
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

#### Database Performance

```bash
# Check slow queries
docker-compose exec postgres psql -U rpguser -d rpgdb -c "
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"
```

### Debugging

#### Enable Debug Logging

```bash
export LOG_LEVEL=debug
make run
```

#### Database Debugging

```bash
# Connect to database
docker-compose exec postgres psql -U rpguser -d rpgdb

# View active sessions
SELECT * FROM active_sessions;

# Check context statistics
SELECT * FROM context_statistics;
```

### Getting Help

- Check the [GitHub Issues](../../issues) for known problems
- Review the [README](README.md) for basic setup
- Look at the [examples](examples/) for usage patterns
- Ask questions in discussions or create an issue

## Development Best Practices

### Code Organization

- Keep packages focused on single responsibilities
- Use interfaces for testability
- Minimize external dependencies
- Handle errors gracefully
- Use context for cancellation and timeouts

### Performance

- Use connection pooling for databases
- Implement caching where appropriate
- Monitor memory usage and goroutine leaks
- Use benchmarks to measure performance
- Profile code in production-like environments

### Security

- Validate all inputs
- Use parameterized queries to prevent SQL injection
- Implement rate limiting
- Keep dependencies updated
- Use HTTPS in production
- Store secrets in environment variables

### Monitoring

- Log important events and errors
- Use structured logging (JSON format)
- Implement health checks
- Monitor key metrics (response time, error rate, etc.)
- Set up alerts for critical issues

This development guide should get you started with contributing to the AI RPG Context Tracking System. For more specific questions, please check the existing documentation or create an issue.
