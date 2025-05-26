#!/bin/bash

# Claude AI Integration Setup Script
# This script helps set up Claude API integration for the AI RPG system

set -e

echo "🤖 Claude AI Integration Setup"
echo "=============================="

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if .env file exists
if [ ! -f .env ]; then
    log_info "Creating .env file from template..."
    cp .env.example .env
    log_success "Created .env file"
else
    log_info ".env file already exists"
fi

# Check if Claude API key is set
if grep -q "AI_API_KEY=your_claude_api_key_here" .env; then
    echo
    log_warning "Claude API key not configured!"
    echo
    echo "To use Claude AI integration, you need to:"
    echo "1. Get a Claude API key from https://console.anthropic.com/"
    echo "2. Edit the .env file and replace 'your_claude_api_key_here' with your actual API key"
    echo
    read -p "Do you want to open the .env file now? (y/n): " open_env
    
    if [[ $open_env =~ ^[Yy]$ ]]; then
        if command -v code >/dev/null 2>&1; then
            code .env
        elif command -v vim >/dev/null 2>&1; then
            vim .env
        elif command -v nano >/dev/null 2>&1; then
            nano .env
        else
            log_info "Please edit .env manually with your preferred editor"
        fi
    fi
else
    log_success "Claude API key appears to be configured"
fi

echo
log_info "Checking Claude API configuration..."

# Show current AI configuration
echo "Current AI Configuration:"
echo "  Provider: $(grep AI_PROVIDER .env | cut -d'=' -f2)"
echo "  Model: $(grep AI_MODEL .env | cut -d'=' -f2)"
echo "  Max Tokens: $(grep AI_MAX_TOKENS .env | cut -d'=' -f2)"
echo "  Temperature: $(grep AI_TEMPERATURE .env | cut -d'=' -f2)"

echo
log_info "Available Claude models:"
echo "  • claude-3-opus-20240229    - Most capable, best for complex tasks"
echo "  • claude-3-sonnet-20240229  - Balanced performance and speed (default)"
echo "  • claude-3-haiku-20240307   - Fastest, best for simple tasks"

echo
log_info "Testing setup..."

# Test if Go dependencies are installed
if go mod verify >/dev/null 2>&1; then
    log_success "Go dependencies verified"
else
    log_info "Installing Go dependencies..."
    go mod download
    log_success "Go dependencies installed"
fi

echo
log_success "Claude AI integration setup complete!"
echo
echo "Next steps:"
echo "1. Ensure your Claude API key is set in .env"
echo "2. Run the examples:"
echo "   • make claude-example  - Claude AI integration demo"
echo "   • make run            - Web server with Claude GM"
echo "   • make test           - Run tests"
echo
echo "Example commands to try:"
echo "   /look around"
echo "   /talk to tavern keeper"
echo "   /attack goblin"
echo "   /examine mysterious chest"
echo
echo "The AI will generate contextual responses based on:"
echo "   • Your character's reputation and health"
echo "   • Previous actions and relationships"
echo "   • Current location and world state"
echo "   • NPC personalities and dispositions"
