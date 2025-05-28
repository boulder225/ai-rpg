
# AI RPG MCP Server Makefile

.PHONY: build run test clean deps

# Build the MCP server
build:
	go build -o ai-rpg-mcp-server main.go

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build and run server (for testing)
run: build
	./ai-rpg-mcp-server

# Test MCP protocol compliance
test: build
	@echo "Testing MCP server protocol..."
	@echo '{"method":"initialize","id":"test","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./ai-rpg-mcp-server
	@echo
	@echo "Testing tools list..."
	@echo '{"method":"tools/list","id":"test"}' | ./ai-rpg-mcp-server

# Test specific tool
test-create-session: build
	@echo "Testing session creation..."
	@echo '{"method":"tools/call","id":"test","params":{"name":"create_session","arguments":{"playerID":"test123","playerName":"TestHero"}}}' | ./ai-rpg-mcp-server

# Clean build artifacts
clean:
	rm -f ai-rpg-mcp-server

# Development setup
setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp ../.env.example .env; \
		echo "Created .env file - please edit with your API keys"; \
	fi
	make deps
	make build
	@echo "Setup complete! Run 'make test' to verify installation."

# Build for different platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o ai-rpg-mcp-server-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o ai-rpg-mcp-server-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o ai-rpg-mcp-server-windows-amd64.exe main.go

# Development mode with file watching (requires entr)
dev:
	find . -name "*.go" | entr -r make run

# Lint and format
lint:
	go fmt ./...
	go vet ./...

# Generate documentation
docs:
	@echo "Available MCP Tools:"
	@echo '{"method":"tools/list","id":"doc"}' | ./ai-rpg-mcp-server | jq '.result.tools[] | {name, description}'

help:
	@echo "AI RPG MCP Server - Available Commands:"
	@echo "  make build          - Build the MCP server binary"
	@echo "  make run            - Build and run the server"
	@echo "  make test           - Test MCP protocol compliance"
	@echo "  make test-create-session - Test session creation tool"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make setup          - Set up development environment"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make dev            - Development mode with file watching"
	@echo "  make lint           - Format and lint code"
	@echo "  make docs           - Generate tool documentation"
	@echo "  make help           - Show this help message"
