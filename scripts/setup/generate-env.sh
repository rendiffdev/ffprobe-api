#!/bin/bash

# FFprobe API - Environment Generator
# Creates secure .env configuration automatically

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }

# Generate secure random string
generate_secret() {
    local length=${1:-32}
    openssl rand -hex "$length" 2>/dev/null || \
    head -c "$length" /dev/urandom | base64 | tr -d '+/=\n' | head -c "$length"
}

# Generate API key with proper format
generate_api_key() {
    local env=${1:-dev}
    echo "ffprobe_${env}_sk_$(generate_secret 32)"
}

# Detect deployment mode from arguments or environment
detect_mode() {
    local mode=${1:-${FFPROBE_MODE:-development}}
    
    case "$mode" in
        prod|production) echo "production" ;;
        dev|development) echo "development" ;;
        quick|fast) echo "quick" ;;
        minimal|min) echo "minimal" ;;
        *) echo "development" ;;
    esac
}

# Get system resources for optimization
get_system_resources() {
    # Memory detection
    local memory_gb=4
    if [[ -f /proc/meminfo ]]; then
        local memory_kb=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        memory_gb=$((memory_kb / 1024 / 1024))
    elif command -v sysctl >/dev/null 2>&1; then
        local memory_bytes=$(sysctl -n hw.memsize 2>/dev/null || echo "4294967296")
        memory_gb=$((memory_bytes / 1024 / 1024 / 1024))
    fi
    
    # CPU detection
    local cpu_cores=2
    cpu_cores=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "2")
    
    echo "$memory_gb $cpu_cores"
}

# Create environment file
create_env() {
    local mode=$1
    local backup_existing=${2:-true}
    
    print_info "Generating .env for $mode mode..."
    
    # Backup existing .env if it exists
    if [[ -f .env && "$backup_existing" == "true" ]]; then
        cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
        print_info "Backed up existing .env file"
    fi
    
    # Generate secure values
    local valkey_password=$(generate_secret 16)
    local jwt_secret=$(generate_secret 32)
    local api_key=$(generate_api_key "$mode")
    local grafana_password=$(generate_secret 12)
    
    # Get system info
    local system_info=($(get_system_resources))
    local memory_gb=${system_info[0]}
    local cpu_cores=${system_info[1]}
    
    # Calculate optimal settings based on resources
    local max_concurrent_jobs=$((cpu_cores > 4 ? 4 : cpu_cores))
    local worker_pool_size=$((cpu_cores * 4))
    local upload_memory_limit="${memory_gb}GB"
    
    # Create .env file
    cat > .env << EOF
# FFprobe API Configuration
# Generated automatically on $(date)
# Mode: $mode

# =============================================================================
# BASIC CONFIGURATION
# =============================================================================

# Environment
GO_ENV=$mode

# API Server
API_PORT=8080
HOST=0.0.0.0

# =============================================================================
# AUTHENTICATION & SECURITY
# =============================================================================

# Authentication
ENABLE_AUTH=true
API_KEY=$api_key
JWT_SECRET=$jwt_secret
TOKEN_EXPIRY=24
REFRESH_EXPIRY=168

# Rate limiting
RATE_LIMIT_PER_MINUTE=60
RATE_LIMIT_PER_HOUR=1000
RATE_LIMIT_PER_DAY=10000

# =============================================================================
# DATABASE CONFIGURATION
# =============================================================================

# Database (SQLite embedded)
DB_TYPE=sqlite
DB_PATH=/app/data/rendiff-probe.db

# Connection pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10

# =============================================================================
# CACHE CONFIGURATION
# =============================================================================

# Valkey (Redis-compatible)
VALKEY_HOST=valkey
VALKEY_PORT=6379
VALKEY_PASSWORD=$valkey_password
VALKEY_DB=0

# Cache settings
CACHE_TTL=3600
CACHE_NAMESPACE=ffprobe:
MEMORY_CACHE_SIZE=512MB
MEMORY_CACHE_TTL=300

# =============================================================================
# FILE STORAGE & PROCESSING
# =============================================================================

# Storage paths
UPLOAD_DIR=/app/uploads
REPORTS_DIR=/app/reports
TEMP_DIR=/app/temp
CACHE_DIR=/app/cache
BACKUP_DIR=/app/backup

# File limits (optimized for system)
MAX_FILE_SIZE=53687091200
MAX_CONCURRENT_JOBS=$max_concurrent_jobs
UPLOAD_MEMORY_LIMIT=$upload_memory_limit
PROCESSING_MEMORY_LIMIT=${memory_gb}GB

# Processing settings
PROCESSING_TIMEOUT=300
WORKER_POOL_SIZE=$worker_pool_size

# =============================================================================
# FFMPEG CONFIGURATION
# =============================================================================

# FFmpeg paths (set by Docker container)
FFMPEG_PATH=/usr/local/bin/ffmpeg
FFPROBE_PATH=/usr/local/bin/ffprobe
VMAF_MODEL_PATH=/usr/local/share/vmaf

# =============================================================================
# AI/LLM CONFIGURATION
# =============================================================================

# Local LLM via Ollama
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434

# AI Models (optimized for system resources)
EOF

    # Add model configuration - gemma3:270m as primary for all systems
    if [[ $memory_gb -ge 4 ]]; then
        cat >> .env << EOF
OLLAMA_MODEL=gemma3:270m
OLLAMA_FALLBACK_MODEL=phi3:mini
EOF
    else
        cat >> .env << EOF
OLLAMA_MODEL=gemma3:270m
OLLAMA_FALLBACK_MODEL=gemma3:270m
EOF
    fi
    
    # Add mode-specific configuration
    case "$mode" in
        "production")
            cat >> .env << EOF

# =============================================================================
# PRODUCTION SETTINGS
# =============================================================================

# Monitoring
ENABLE_PROMETHEUS=true
ENABLE_GRAFANA=true
GRAFANA_USER=admin
GRAFANA_PASSWORD=$grafana_password

# Performance
ENABLE_DETAILED_METRICS=true
LOG_FORMAT=json
LOG_OUTPUT=stdout

# Security
ENABLE_AUTH=true
HSTS_MAX_AGE=31536000
HSTS_INCLUDE_SUBDOMAINS=true

# Backup
ENABLE_BACKUPS=true
BACKUP_RETENTION_DAYS=30
BACKUP_SCHEDULE="0 2 * * *"
EOF
            ;;
            
        "development")
            cat >> .env << EOF

# =============================================================================
# DEVELOPMENT SETTINGS
# =============================================================================

# Development overrides
DEV_ENABLE_DEBUG=true
DEV_DISABLE_AUTH=false
DEV_DISABLE_RATE_LIMIT=false
DEV_VERBOSE_LOGGING=true

# Monitoring (optional)
ENABLE_PROMETHEUS=false
ENABLE_GRAFANA=false

# Relaxed settings
LOG_FORMAT=text
LOG_REQUESTS=true
EOF
            ;;
            
        "quick"|"minimal")
            cat >> .env << EOF

# =============================================================================
# QUICK/MINIMAL SETTINGS
# =============================================================================

# Minimal features only
ENABLE_AUTH=false
ENABLE_PROMETHEUS=false
ENABLE_GRAFANA=false
ENABLE_BACKUPS=false

# Basic logging
LOG_FORMAT=text
LOG_OUTPUT=stdout
LOG_REQUESTS=false

# Reduced limits
MAX_CONCURRENT_JOBS=2
WORKER_POOL_SIZE=4
MEMORY_CACHE_SIZE=256MB
EOF
            ;;
    esac
    
    # Add common footer
    cat >> .env << EOF

# =============================================================================
# NETWORK & CORS
# =============================================================================

# CORS
ALLOWED_ORIGINS=*
TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16

# =============================================================================
# LOGGING
# =============================================================================

# Log rotation
LOG_ROTATION=true
LOG_MAX_SIZE=100MB
LOG_MAX_AGE=30
LOG_MAX_BACKUPS=10

# =============================================================================
# PERFORMANCE TUNING
# =============================================================================

# System optimizations (auto-detected)
# Memory: ${memory_gb}GB, CPU: ${cpu_cores} cores
# Configured for optimal performance on this system

EOF
    
    print_success "Generated .env file for $mode mode"
    print_info "API Key: $api_key"
    print_warning "⚠️  Save your API key - you'll need it for requests!"
    
    if [[ "$mode" == "production" ]]; then
        print_info "Grafana Password: $grafana_password"
        print_warning "⚠️  Save your Grafana password for monitoring access!"
    fi
}

# Validate generated environment
validate_env() {
    print_info "Validating generated environment..."
    
    if [[ ! -f .env ]]; then
        print_error "❌ .env file not found"
        return 1
    fi
    
    # Check required variables
    local required_vars=("API_PORT" "DB_TYPE" "VALKEY_PASSWORD" "API_KEY" "JWT_SECRET")
    local missing=0
    
    for var in "${required_vars[@]}"; do
        if ! grep -q "^${var}=" .env; then
            print_error "❌ Missing required variable: $var"
            missing=$((missing + 1))
        fi
    done
    
    if [[ $missing -eq 0 ]]; then
        print_success "Environment validation passed"
        return 0
    else
        print_error "❌ Environment validation failed: $missing missing variables"
        return 1
    fi
}

# Show usage
show_usage() {
    echo "Usage: $0 [mode] [options]"
    echo
    echo "Modes:"
    echo "  production  - Full production setup with monitoring"
    echo "  development - Development setup with debugging"
    echo "  quick       - Quick start with minimal features"
    echo "  minimal     - Minimal setup for testing"
    echo
    echo "Options:"
    echo "  --no-backup - Don't backup existing .env file"
    echo "  --validate  - Validate generated .env only"
    echo "  --help      - Show this help"
}

# Main function
main() {
    local mode="development"
    local backup_existing=true
    local validate_only=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-backup)
                backup_existing=false
                shift
                ;;
            --validate)
                validate_only=true
                shift
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            -*)
                print_warning "Unknown option: $1"
                shift
                ;;
            *)
                mode=$(detect_mode "$1")
                shift
                ;;
        esac
    done
    
    if [[ "$validate_only" == "true" ]]; then
        validate_env
        exit $?
    fi
    
    print_info "🔐 FFprobe API - Environment Generator"
    print_info "Detected mode: $mode"
    
    create_env "$mode" "$backup_existing"
    validate_env
    
    echo
    print_success "🎉 Environment configuration complete!"
    print_info "Your .env file has been generated with secure defaults"
    print_info "Run 'make quick' to start the API with these settings"
}

# Run main function
main "$@"