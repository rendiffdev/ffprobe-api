#!/bin/bash
# Production-Grade Entrypoint Script for FFprobe API
# Enterprise deployment with comprehensive health checks and monitoring
# Security hardened with proper signal handling and graceful shutdown

set -euo pipefail

# =============================================================================
# Configuration and Constants
# =============================================================================

readonly SCRIPT_NAME="$(basename "$0")"
readonly PID_FILE="/tmp/ffprobe-api.pid"
readonly LOG_FILE="${LOG_FILE:-/app/logs/entrypoint.log}"
readonly HEALTH_CHECK_URL="http://localhost:${API_PORT:-8080}/health"
readonly STARTUP_TIMEOUT="${STARTUP_TIMEOUT:-60}"
readonly SHUTDOWN_TIMEOUT="${SHUTDOWN_TIMEOUT:-30}"

# Colors for output (if terminal supports it)
if [[ -t 1 ]]; then
    readonly RED='\033[0;31m'
    readonly GREEN='\033[0;32m'
    readonly YELLOW='\033[1;33m'
    readonly BLUE='\033[0;34m'
    readonly NC='\033[0m' # No Color
else
    readonly RED=''
    readonly GREEN=''
    readonly YELLOW=''
    readonly BLUE=''
    readonly NC=''
fi

# =============================================================================
# Logging Functions
# =============================================================================

log() {
    local level="$1"
    shift
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    local message="$*"
    
    # Structured logging for production
    if [[ "${LOG_FORMAT:-json}" == "json" ]]; then
        echo "{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"component\":\"entrypoint\",\"message\":\"$message\"}" | tee -a "${LOG_FILE}" 2>/dev/null || echo "{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"component\":\"entrypoint\",\"message\":\"$message\"}"
    else
        echo -e "${timestamp} [${level}] ${SCRIPT_NAME}: $message" | tee -a "${LOG_FILE}" 2>/dev/null || echo -e "${timestamp} [${level}] ${SCRIPT_NAME}: $message"
    fi
}

log_info() {
    log "INFO" "$@"
}

log_warn() {
    log "WARN" "$@"
}

log_error() {
    log "ERROR" "$@"
}

log_debug() {
    if [[ "${DEBUG:-false}" == "true" ]] || [[ "${LOG_LEVEL:-info}" == "debug" ]]; then
        log "DEBUG" "$@"
    fi
}

# =============================================================================
# Signal Handling and Graceful Shutdown
# =============================================================================

cleanup() {
    local exit_code=$?
    log_info "Cleanup initiated with exit code: $exit_code"
    
    if [[ -f "$PID_FILE" ]] && [[ -s "$PID_FILE" ]]; then
        local pid=$(cat "$PID_FILE")
        if kill -0 "$pid" 2>/dev/null; then
            log_info "Sending SIGTERM to application (PID: $pid)"
            kill -TERM "$pid" 2>/dev/null || true
            
            # Wait for graceful shutdown
            local count=0
            while kill -0 "$pid" 2>/dev/null && [[ $count -lt $SHUTDOWN_TIMEOUT ]]; do
                sleep 1
                ((count++))
            done
            
            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                log_warn "Application did not shutdown gracefully, sending SIGKILL"
                kill -KILL "$pid" 2>/dev/null || true
            fi
        fi
        rm -f "$PID_FILE"
    fi
    
    log_info "Cleanup completed"
    exit $exit_code
}

# Setup signal handlers
trap cleanup EXIT INT TERM

# =============================================================================
# Environment Validation
# =============================================================================

validate_environment() {
    log_info "Validating environment configuration..."
    
    local errors=0
    
    # Check required directories
    local required_dirs=(
        "${UPLOAD_DIR:-/app/uploads}"
        "${REPORTS_DIR:-/app/reports}"
        "${TEMP_DIR:-/app/temp}"
        "${CACHE_DIR:-/app/cache}"
        "/app/logs"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            log_error "Required directory does not exist: $dir"
            ((errors++))
        elif [[ ! -w "$dir" ]]; then
            log_error "Directory is not writable: $dir"
            ((errors++))
        fi
    done
    
    # Check required binaries
    local required_binaries=(
        "${FFMPEG_PATH:-/usr/local/bin/ffmpeg}"
        "${FFPROBE_PATH:-/usr/local/bin/ffprobe}"
    )
    
    for binary in "${required_binaries[@]}"; do
        if [[ ! -x "$binary" ]]; then
            log_error "Required binary is not executable: $binary"
            ((errors++))
        fi
    done
    
    # Check VMAF models
    local vmaf_dir="${VMAF_MODEL_PATH:-/usr/local/share/vmaf}"
    if [[ ! -d "$vmaf_dir" ]]; then
        log_error "VMAF models directory does not exist: $vmaf_dir"
        ((errors++))
    else
        local model_count=$(find "$vmaf_dir" -name "*.json" | wc -l)
        if [[ $model_count -eq 0 ]]; then
            log_error "No VMAF models found in: $vmaf_dir"
            ((errors++))
        else
            log_info "Found $model_count VMAF models in $vmaf_dir"
        fi
    fi
    
    # Validate environment variables
    local required_vars=()
    
    # Add required vars based on configuration
    if [[ "${ENABLE_AUTH:-true}" == "true" ]]; then
        if [[ -z "${API_KEY:-}" ]] && [[ -z "${JWT_SECRET:-}" ]]; then
            log_error "Authentication is enabled but neither API_KEY nor JWT_SECRET is set"
            ((errors++))
        fi
        
        if [[ -n "${API_KEY:-}" ]] && [[ ${#API_KEY} -lt 32 ]]; then
            log_error "API_KEY must be at least 32 characters long"
            ((errors++))
        fi
        
        if [[ -n "${JWT_SECRET:-}" ]] && [[ ${#JWT_SECRET} -lt 32 ]]; then
            log_error "JWT_SECRET must be at least 32 characters long"
            ((errors++))
        fi
    fi
    
    # Database connection validation
    if [[ -n "${POSTGRES_HOST:-}" ]]; then
        if [[ -z "${POSTGRES_PASSWORD:-}" ]]; then
            log_error "POSTGRES_PASSWORD is required when POSTGRES_HOST is set"
            ((errors++))
        fi
    fi
    
    if [[ $errors -gt 0 ]]; then
        log_error "Environment validation failed with $errors errors"
        return 1
    fi
    
    log_info "Environment validation completed successfully"
    return 0
}

# =============================================================================
# System Health Checks
# =============================================================================

check_system_health() {
    log_info "Performing system health checks..."
    
    # Check available disk space
    local temp_dir="${TEMP_DIR:-/app/temp}"
    local available_space=$(df "$temp_dir" | awk 'NR==2 {print $4}')
    local min_space_kb=$((1024 * 1024)) # 1GB in KB
    
    if [[ $available_space -lt $min_space_kb ]]; then
        log_warn "Low disk space available: ${available_space}KB (minimum recommended: ${min_space_kb}KB)"
    fi
    
    # Check memory availability
    if command -v free >/dev/null 2>&1; then
        local available_mem=$(free -m | awk 'NR==2{printf "%.0f", $7}')
        local min_mem_mb=512
        
        if [[ $available_mem -lt $min_mem_mb ]]; then
            log_warn "Low memory available: ${available_mem}MB (minimum recommended: ${min_mem_mb}MB)"
        else
            log_info "Available memory: ${available_mem}MB"
        fi
    fi
    
    # Test FFmpeg/FFprobe functionality
    log_debug "Testing FFmpeg functionality..."
    if timeout 10s "${FFMPEG_PATH:-/usr/local/bin/ffmpeg}" -version >/dev/null 2>&1; then
        log_debug "FFmpeg is working correctly"
    else
        log_error "FFmpeg test failed"
        return 1
    fi
    
    if timeout 10s "${FFPROBE_PATH:-/usr/local/bin/ffprobe}" -version >/dev/null 2>&1; then
        log_debug "FFprobe is working correctly"
    else
        log_error "FFprobe test failed"
        return 1
    fi
    
    log_info "System health checks completed successfully"
    return 0
}

# =============================================================================
# Dependency Health Checks
# =============================================================================

check_dependencies() {
    log_info "Checking external dependencies..."
    
    # Check database connectivity
    if [[ -n "${POSTGRES_HOST:-}" ]]; then
        log_debug "Checking database connectivity..."
        local max_attempts=30
        local attempt=1
        
        while [[ $attempt -le $max_attempts ]]; do
            if timeout 5s pg_isready -h "${POSTGRES_HOST}" -p "${POSTGRES_PORT:-5432}" -U "${POSTGRES_USER:-ffprobe}" >/dev/null 2>&1; then
                log_info "Database connection established"
                break
            fi
            
            if [[ $attempt -eq $max_attempts ]]; then
                log_error "Failed to connect to database after $max_attempts attempts"
                return 1
            fi
            
            log_debug "Database connection attempt $attempt/$max_attempts failed, retrying..."
            sleep 2
            ((attempt++))
        done
    fi
    
    # Check Redis connectivity
    if [[ -n "${REDIS_HOST:-}" ]]; then
        log_debug "Checking Redis connectivity..."
        local redis_cmd="redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT:-6379}"
        
        if [[ -n "${REDIS_PASSWORD:-}" ]]; then
            redis_cmd="$redis_cmd -a ${REDIS_PASSWORD}"
        fi
        
        if timeout 5s $redis_cmd ping >/dev/null 2>&1; then
            log_info "Redis connection established"
        else
            log_warn "Redis connection failed - continuing without cache"
        fi
    fi
    
    # Check Ollama connectivity (if enabled)
    if [[ "${ENABLE_LOCAL_LLM:-true}" == "true" ]] && [[ -n "${OLLAMA_URL:-}" ]]; then
        log_debug "Checking Ollama connectivity..."
        local ollama_health_url="${OLLAMA_URL}/api/version"
        
        if timeout 10s curl -sf "$ollama_health_url" >/dev/null 2>&1; then
            log_info "Ollama service is available"
        else
            log_warn "Ollama service is not available - AI features may be limited"
        fi
    fi
    
    log_info "Dependency checks completed"
    return 0
}

# =============================================================================
# Application Startup
# =============================================================================

start_application() {
    log_info "Starting FFprobe API application..."
    
    # Change to application directory
    cd /app
    
    # Start the application in background
    local app_cmd="${1:-/usr/local/bin/ffprobe-api}"
    
    # Add any additional arguments
    shift || true
    
    log_info "Executing: $app_cmd $*"
    
    # Start application and capture PID
    exec "$app_cmd" "$@" &
    local app_pid=$!
    
    # Save PID for cleanup
    echo "$app_pid" > "$PID_FILE"
    
    log_info "Application started with PID: $app_pid"
    
    # Wait for application to be ready
    local attempt=1
    local max_attempts=$((STARTUP_TIMEOUT))
    
    log_info "Waiting for application to be ready..."
    
    while [[ $attempt -le $max_attempts ]]; do
        if kill -0 "$app_pid" 2>/dev/null; then
            if timeout 5s curl -sf "$HEALTH_CHECK_URL" >/dev/null 2>&1; then
                log_info "Application is ready and healthy (attempt $attempt/$max_attempts)"
                break
            fi
        else
            log_error "Application process died during startup"
            return 1
        fi
        
        if [[ $attempt -eq $max_attempts ]]; then
            log_error "Application failed to become ready within ${STARTUP_TIMEOUT} seconds"
            return 1
        fi
        
        log_debug "Health check attempt $attempt/$max_attempts failed, retrying..."
        sleep 1
        ((attempt++))
    done
    
    log_info "Application startup completed successfully"
    
    # Wait for the application to exit
    wait "$app_pid"
    local exit_code=$?
    
    log_info "Application exited with code: $exit_code"
    return $exit_code
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    log_info "FFprobe API Production Entrypoint Starting..."
    log_info "Version: Production-Grade v1.0.0"
    log_info "Environment: ${GO_ENV:-production}"
    log_info "Security Hardened: ${SECURITY_HARDENED:-true}"
    log_info "Monitoring Enabled: ${MONITORING_ENABLED:-true}"
    
    # Create log directory if it doesn't exist
    mkdir -p "$(dirname "$LOG_FILE")" 2>/dev/null || true
    
    # Validate environment
    if ! validate_environment; then
        log_error "Environment validation failed"
        exit 1
    fi
    
    # Perform system health checks
    if ! check_system_health; then
        log_error "System health checks failed"
        exit 1
    fi
    
    # Check dependencies
    if ! check_dependencies; then
        log_error "Dependency checks failed"
        exit 1
    fi
    
    # Export additional environment variables for the application
    export DOCKER_ENTRYPOINT_VERSION="production-1.0.0"
    export STARTUP_TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    
    log_info "All pre-startup checks passed, starting application..."
    
    # Start the application
    start_application "$@"
}

# Execute main function with all arguments
main "$@"