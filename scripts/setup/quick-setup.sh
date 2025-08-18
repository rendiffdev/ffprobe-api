#!/bin/bash

# FFprobe API - Quick Setup
# Minimal setup for development and testing

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }

main() {
    print_info "Starting FFprobe API Quick Setup..."
    
    # Create basic .env if not exists
    if [[ ! -f .env ]]; then
        print_info "Creating basic .env configuration..."
        cat > .env << 'EOF'
GO_ENV=development
API_PORT=8080
DB_TYPE=sqlite
VALKEY_PASSWORD=quickstart123
ENABLE_AUTH=false
OLLAMA_MODEL=gemma3:270m
ENABLE_LOCAL_LLM=true
EOF
        print_success "Basic configuration created"
    fi
    
    # Quick deployment
    print_info "Starting services..."
    docker compose --profile quick up -d
    
    # Wait for readiness
    print_info "Waiting for API..."
    sleep 10
    
    print_success "🎉 FFprobe API Quick Setup Complete!"
    print_info "API: http://localhost:8080"
    print_info "Health: curl http://localhost:8080/health"
}

main "$@"