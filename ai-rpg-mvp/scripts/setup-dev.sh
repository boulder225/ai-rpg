#!/bin/bash

# AI RPG Context Tracker - Development Setup Script
# This script helps set up the development environment

set -e  # Exit on any error

echo "🎮 AI RPG Context Tracker - Development Setup"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    
    if ! command_exists go; then
        missing_tools+=("go")
    else
        GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
        log_success "Go $GO_VERSION installed"
    fi
    
    if ! command_exists docker; then
        missing_tools+=("docker")
    else
        log_success "Docker installed"
    fi
    
    if ! command_exists docker-compose; then
        if ! command_exists docker && docker compose version >/dev/null 2>&1; then
            missing_tools+=("docker-compose")
        fi
    else
        log_success "Docker Compose installed"
    fi
    
    if ! command_exists make; then
        missing_tools+=("make")
    else
        log_success "Make installed"
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        echo "Please install the missing tools and run this script again."
        exit 1
    fi
}

# Setup environment file
setup_env() {
    log_info "Setting up environment configuration..."
    
    if [ ! -f .env ]; then
        cp .env.example .env
        log_success "Created .env file from .env.example"
        log_warning "Please edit .env file with your configuration (especially AI_API_KEY)"
    else
        log_info ".env file already exists"
    fi
}

# Install Go dependencies
install_dependencies() {
    log_info "Installing Go dependencies..."
    
    if ! go mod download; then
        log_error "Failed to download Go dependencies"
        exit 1
    fi
    
    log_success "Go dependencies installed"
    
    # Install development tools
    log_info "Installing development tools..."
    
    go install golang.org/x/lint/golint@latest 2>/dev/null || log_warning "Failed to install golint"
    go install golang.org/x/tools/cmd/goimports@latest 2>/dev/null || log_warning "Failed to install goimports"
    
    log_success "Development tools installed"
}

# Setup database
setup_database() {
    log_info "Setting up database..."
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    
    # Start PostgreSQL container
    if docker-compose up -d postgres; then
        log_success "PostgreSQL container started"
        
        # Wait for database to be ready
        log_info "Waiting for database to be ready..."
        sleep 5
        
        # Test database connection
        local max_attempts=30
        local attempt=1
        
        while [ $attempt -le $max_attempts ]; do
            if docker-compose exec -T postgres pg_isready -U rpguser -d rpgdb >/dev/null 2>&1; then
                log_success "Database is ready"
                break
            fi
            
            if [ $attempt -eq $max_attempts ]; then
                log_error "Database failed to start after $max_attempts attempts"
                exit 1
            fi
            
            echo -n "."
            sleep 2
            ((attempt++))
        done
    else
        log_error "Failed to start PostgreSQL container"
        exit 1
    fi
}

# Run tests
run_tests() {
    log_info "Running tests..."
    
    if make test; then
        log_success "All tests passed"
    else
        log_warning "Some tests failed, but setup can continue"
    fi
}

# Start development server
start_server() {
    local start_server_choice
    echo
    read -p "Would you like to start the development server now? (y/n): " start_server_choice
    
    if [[ $start_server_choice =~ ^[Yy]$ ]]; then
        log_info "Starting development server..."
        log_info "Server will be available at http://localhost:8080"
        log_info "Press Ctrl+C to stop the server"
        echo
        make run
    else
        log_info "You can start the server later with: make run"
    fi
}

# Main setup function
main() {
    echo
    log_info "Starting development environment setup..."
    echo
    
    check_prerequisites
    echo
    
    setup_env
    echo
    
    install_dependencies
    echo
    
    setup_database
    echo
    
    run_tests
    echo
    
    log_success "Development environment setup complete!"
    log_info "Quick commands:"
    echo "  make run      - Start the web server"
    echo "  make test     - Run tests"
    echo "  make examples - Run basic usage example"
    echo "  make check    - Run all quality checks"
    echo
    
    start_server
}

# Handle script interruption
trap 'echo -e "\n${YELLOW}Setup interrupted${NC}"; exit 1' INT

# Run main function
main "$@"
