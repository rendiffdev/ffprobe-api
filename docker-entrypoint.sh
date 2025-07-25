#!/bin/sh
# Docker entrypoint script for FFprobe API

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}" >&2
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

# Function to wait for a service
wait_for_service() {
    local host=$1
    local port=$2
    local service=$3
    local timeout=${4:-30}
    
    log "Waiting for $service at $host:$port..."
    
    for i in $(seq 1 $timeout); do
        if nc -z "$host" "$port" 2>/dev/null; then
            success "$service is available!"
            return 0
        fi
        
        log "Waiting for $service... ($i/$timeout)"
        sleep 1
    done
    
    error "$service is not available after ${timeout}s"
    return 1
}

# Function to check FFmpeg installation
check_ffmpeg() {
    log "Checking FFmpeg installation..."
    
    if ! command -v ffmpeg >/dev/null 2>&1; then
        error "FFmpeg not found"
        exit 1
    fi
    
    if ! command -v ffprobe >/dev/null 2>&1; then
        error "FFprobe not found"
        exit 1
    fi
    
    # Check VMAF models
    if [ ! -d "/usr/local/share/vmaf" ]; then
        warn "VMAF models directory not found"
    else
        success "VMAF models found"
    fi
    
    success "FFmpeg/FFprobe installation verified"
}

# Function to validate environment
validate_environment() {
    log "Validating environment variables..."
    
    # Check required variables
    required_vars="POSTGRES_HOST POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD"
    for var in $required_vars; do
        if [ -z "$(eval echo \$$var)" ]; then
            error "Required environment variable $var is not set"
            exit 1
        fi
    done
    
    # Warn about default secrets
    if [ "$API_KEY" = "your-super-secret-api-key-change-in-production" ]; then
        warn "Using default API_KEY - CHANGE IN PRODUCTION!"
    fi
    
    if [ "$JWT_SECRET" = "your-super-secret-jwt-key-change-in-production" ]; then
        warn "Using default JWT_SECRET - CHANGE IN PRODUCTION!"
    fi
    
    success "Environment validation completed"
}

# Function to create directories
create_directories() {
    log "Creating application directories..."
    
    # Create required directories
    mkdir -p "$UPLOAD_DIR" "$REPORTS_DIR" "$TEMP_DIR" "$CACHE_DIR" "$BACKUP_DIR"
    
    # Set permissions
    chown -R ffprobe:ffprobe "$UPLOAD_DIR" "$REPORTS_DIR" "$TEMP_DIR" "$CACHE_DIR" "$BACKUP_DIR" 2>/dev/null || true
    
    success "Directories created successfully"
}

# Function to run health checks
run_health_checks() {
    log "Running health checks..."
    
    # Check if health check script exists
    if [ -f "/app/scripts/healthcheck.sh" ]; then
        chmod +x /app/scripts/healthcheck.sh
        /app/scripts/healthcheck.sh || warn "Health check script failed"
    else
        warn "Health check script not found"
    fi
}

# Main execution
main() {
    log "Starting FFprobe API container..."
    log "Container user: $(whoami)"
    log "Working directory: $(pwd)"
    
    # Run pre-startup checks
    validate_environment
    check_ffmpeg
    create_directories
    
    # Wait for dependencies if configured
    if [ -n "$POSTGRES_HOST" ] && [ -n "$POSTGRES_PORT" ]; then
        wait_for_service "$POSTGRES_HOST" "${POSTGRES_PORT:-5432}" "PostgreSQL" 60
    fi
    
    if [ -n "$REDIS_HOST" ] && [ -n "$REDIS_PORT" ]; then
        wait_for_service "$REDIS_HOST" "${REDIS_PORT:-6379}" "Redis" 30
    fi
    
    # Log startup information
    log "Configuration:"
    log "  - API Port: ${API_PORT:-8080}"
    log "  - Log Level: ${LOG_LEVEL:-info}"
    log "  - Upload Dir: ${UPLOAD_DIR:-/app/uploads}"
    log "  - Reports Dir: ${REPORTS_DIR:-/app/reports}"
    log "  - Max File Size: ${MAX_FILE_SIZE:-53687091200} bytes"
    log "  - Auth Enabled: ${ENABLE_AUTH:-true}"
    log "  - Rate Limit Enabled: ${ENABLE_RATE_LIMIT:-true}"
    
    # Start the application
    success "Starting FFprobe API..."
    
    # Execute the main command
    exec "$@"
}

# Handle signals for graceful shutdown
trap 'log "Received shutdown signal, stopping..."; exit 0' TERM INT

# Run main function
main "$@"