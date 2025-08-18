#!/bin/bash

# FFprobe API - Configuration Validation
# Validates system configuration and requirements

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

ERRORS=0
WARNINGS=0

check_docker() {
    print_info "Checking Docker..."
    
    if ! command -v docker >/dev/null 2>&1; then
        print_error "Docker not found"
        ERRORS=$((ERRORS + 1))
        return
    fi
    
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker daemon not running"
        ERRORS=$((ERRORS + 1))
        return
    fi
    
    if ! docker compose version >/dev/null 2>&1; then
        print_error "Docker Compose not available"
        ERRORS=$((ERRORS + 1))
        return
    fi
    
    print_success "Docker configuration valid"
}

check_compose_files() {
    print_info "Checking Docker Compose files..."
    
    for file in compose.yaml compose.development.yaml compose.production.yaml compose.sqlite.yaml; do
        if [[ ! -f "$file" ]]; then
            print_error "Missing compose file: $file"
            ERRORS=$((ERRORS + 1))
            continue
        fi
        
        if ! docker compose -f "$file" config >/dev/null 2>&1; then
            print_error "Invalid compose file: $file"
            ERRORS=$((ERRORS + 1))
        else
            print_success "Valid compose file: $file"
        fi
    done
}

check_env_file() {
    print_info "Checking environment configuration..."
    
    if [[ ! -f .env ]]; then
        print_warning ".env file not found - will use defaults"
        WARNINGS=$((WARNINGS + 1))
        return
    fi
    
    # Check required variables
    local required_vars=("API_PORT" "DB_TYPE" "VALKEY_HOST")
    
    for var in "${required_vars[@]}"; do
        if ! grep -q "^${var}=" .env; then
            print_warning "Missing environment variable: $var"
            WARNINGS=$((WARNINGS + 1))
        fi
    done
    
    # Check for default passwords
    if grep -q "change_this" .env; then
        print_warning "Default passwords detected - consider changing them"
        WARNINGS=$((WARNINGS + 1))
    fi
    
    print_success "Environment configuration checked"
}

check_ports() {
    print_info "Checking port availability..."
    
    local ports=(8080 6379 11434 5432)
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            print_warning "Port $port is already in use"
            WARNINGS=$((WARNINGS + 1))
        fi
    done
    
    print_success "Port check complete"
}

check_disk_space() {
    print_info "Checking disk space..."
    
    local available
    if command -v df >/dev/null 2>&1; then
        available=$(df -BG . | tail -1 | awk '{print $4}' | sed 's/G//')
        
        if [[ $available -lt 5 ]]; then
            print_error "Insufficient disk space: ${available}GB available (minimum 5GB required)"
            ERRORS=$((ERRORS + 1))
        elif [[ $available -lt 10 ]]; then
            print_warning "Low disk space: ${available}GB available (10GB+ recommended)"
            WARNINGS=$((WARNINGS + 1))
        else
            print_success "Sufficient disk space: ${available}GB"
        fi
    fi
}

check_memory() {
    print_info "Checking memory..."
    
    local memory_gb
    if [[ -f /proc/meminfo ]]; then
        local memory_kb=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        memory_gb=$((memory_kb / 1024 / 1024))
    elif command -v sysctl >/dev/null 2>&1; then
        local memory_bytes=$(sysctl -n hw.memsize 2>/dev/null || echo "4294967296")
        memory_gb=$((memory_bytes / 1024 / 1024 / 1024))
    else
        print_warning "Cannot detect memory - assuming sufficient"
        return
    fi
    
    if [[ $memory_gb -lt 2 ]]; then
        print_error "Insufficient memory: ${memory_gb}GB (minimum 2GB required)"
        ERRORS=$((ERRORS + 1))
    elif [[ $memory_gb -lt 4 ]]; then
        print_warning "Low memory: ${memory_gb}GB (4GB+ recommended)"
        WARNINGS=$((WARNINGS + 1))
    else
        print_success "Sufficient memory: ${memory_gb}GB"
    fi
}

main() {
    echo "🔍 FFprobe API - Configuration Validation"
    echo "========================================"
    
    check_docker
    check_compose_files
    check_env_file
    check_ports
    check_disk_space
    check_memory
    
    echo
    echo "Validation Results:"
    echo "=================="
    
    if [[ $ERRORS -gt 0 ]]; then
        print_error "$ERRORS error(s) found - setup may fail"
        exit 1
    fi
    
    if [[ $WARNINGS -gt 0 ]]; then
        print_warning "$WARNINGS warning(s) found - setup may have issues"
    fi
    
    print_success "✅ Configuration validation passed!"
    echo
    print_info "System is ready for FFprobe API deployment"
}

main "$@"