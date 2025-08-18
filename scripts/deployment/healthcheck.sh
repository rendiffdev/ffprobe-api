#!/bin/bash

# FFprobe API - Health Check Script
# Comprehensive health monitoring for all services

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

check_api() {
    print_info "Checking API service..."
    
    if curl -sf http://localhost:8080/health >/dev/null 2>&1; then
        local response=$(curl -s http://localhost:8080/health)
        if echo "$response" | grep -q '"status":"healthy"'; then
            print_success "API service healthy"
            return 0
        else
            print_warning "API service running but not healthy"
            echo "$response"
            return 1
        fi
    else
        print_error "API service not responding"
        return 1
    fi
}

check_database() {
    print_info "Checking database..."
    
    # For SQLite, check if file exists and is accessible
    if docker compose exec -T api test -f /app/data/ffprobe.db 2>/dev/null; then
        print_success "Database accessible"
        return 0
    else
        print_error "Database not accessible"
        return 1
    fi
}

check_valkey() {
    print_info "Checking Valkey cache..."
    
    if docker compose exec -T valkey valkey-cli ping 2>/dev/null | grep -q "PONG"; then
        print_success "Valkey cache healthy"
        return 0
    else
        print_error "Valkey cache not responding"
        return 1
    fi
}

check_ollama() {
    print_info "Checking Ollama AI service..."
    
    if curl -sf http://localhost:11434/api/version >/dev/null 2>&1; then
        print_success "Ollama AI service healthy"
        
        # Check if models are loaded
        local models=$(curl -s http://localhost:11434/api/tags | jq -r '.models[]?.name' 2>/dev/null | wc -l)
        if [[ $models -gt 0 ]]; then
            print_success "AI models available: $models"
        else
            print_warning "No AI models found - run: make setup-ollama"
        fi
        return 0
    else
        print_error "Ollama AI service not responding"
        return 1
    fi
}

check_docker_services() {
    print_info "Checking Docker services status..."
    
    local services=$(docker compose ps --format "table {{.Service}}\t{{.Status}}" | grep -v "SERVICE")
    
    if [[ -z "$services" ]]; then
        print_error "No services found"
        return 1
    fi
    
    echo "$services" | while IFS=$'\t' read -r service status; do
        if echo "$status" | grep -q "Up"; then
            print_success "Service $service: $status"
        else
            print_error "Service $service: $status"
        fi
    done
}

check_disk_space() {
    print_info "Checking disk space..."
    
    local usage=$(df -h . | tail -1)
    print_info "Disk usage: $usage"
    
    local percent=$(echo "$usage" | awk '{print $5}' | sed 's/%//')
    if [[ $percent -gt 90 ]]; then
        print_error "Disk usage critical: ${percent}%"
        return 1
    elif [[ $percent -gt 80 ]]; then
        print_warning "Disk usage high: ${percent}%"
    else
        print_success "Disk usage normal: ${percent}%"
    fi
}

check_memory_usage() {
    print_info "Checking memory usage..."
    
    if command -v free >/dev/null 2>&1; then
        local mem_info=$(free -h | head -2 | tail -1)
        print_info "Memory: $mem_info"
    fi
    
    # Check Docker containers memory
    if command -v docker >/dev/null 2>&1; then
        local container_stats=$(docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep -v "CONTAINER")
        if [[ -n "$container_stats" ]]; then
            print_info "Container resource usage:"
            echo "$container_stats"
        fi
    fi
}

run_comprehensive_test() {
    print_info "Running comprehensive API test..."
    
    # Test basic endpoint
    if curl -sf http://localhost:8080/health >/dev/null; then
        print_success "Health endpoint responding"
    else
        print_error "Health endpoint failed"
        return 1
    fi
    
    # Test with a small file (README)
    if [[ -f README.md ]]; then
        print_info "Testing file upload with README.md..."
        local response=$(curl -s -X POST -F "file=@README.md" http://localhost:8080/api/v1/probe/file 2>/dev/null)
        
        if [[ $? -eq 0 && -n "$response" ]]; then
            print_success "File upload test passed"
        else
            print_warning "File upload test failed (this may be normal if auth is enabled)"
        fi
    fi
}

main() {
    echo "🏥 FFprobe API - Health Check"
    echo "============================"
    
    local failed=0
    
    # Core service checks
    check_api || failed=$((failed + 1))
    check_database || failed=$((failed + 1))
    check_valkey || failed=$((failed + 1))
    check_ollama || failed=$((failed + 1))
    
    # System checks
    check_docker_services
    check_disk_space || failed=$((failed + 1))
    check_memory_usage
    
    # Comprehensive test
    run_comprehensive_test || failed=$((failed + 1))
    
    echo
    echo "Health Check Summary:"
    echo "===================="
    
    if [[ $failed -eq 0 ]]; then
        print_success "🎉 All systems healthy!"
        exit 0
    else
        print_error "$failed service(s) have issues"
        print_info "Check logs with: make logs"
        exit 1
    fi
}

main "$@"