#!/bin/bash
# Production-Grade Deployment Script for FFprobe API
# This script handles secure, zero-downtime deployment with comprehensive checks

set -euo pipefail

# Script metadata
readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly TIMESTAMP="$(date +%Y%m%d_%H%M%S)"

# Configuration
readonly LOG_FILE="/tmp/ffprobe-deploy-${TIMESTAMP}.log"
readonly BACKUP_DIR="${BACKUP_DIR:-/opt/ffprobe/backups}"
readonly ENV_FILE="${ENV_FILE:-${PROJECT_ROOT}/.env.production}"
readonly COMPOSE_FILES=("${PROJECT_ROOT}/compose.yml" "${PROJECT_ROOT}/compose.enterprise.yml")
readonly HEALTH_CHECK_TIMEOUT=300
readonly MAX_DEPLOYMENT_TIME=1800

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Logging functions
log() {
    local level="$1"
    local message="$2"
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo -e "${BLUE}[${timestamp}] [${level}] ${message}${NC}" | tee -a "$LOG_FILE"
}

info() { log "INFO" "$1"; }
warn() { log "WARN" "$1"; }
error() { log "ERROR" "$1" >&2; }
success() { log "SUCCESS" "$1"; }
fatal() { error "$1"; exit 1; }

# Cleanup function
cleanup() {
    if [[ -f "$LOG_FILE" ]]; then
        info "Deployment log saved to: $LOG_FILE"
    fi
}
trap cleanup EXIT

# Check prerequisites
check_prerequisites() {
    info "Checking deployment prerequisites..."
    
    local missing_tools=()
    
    # Check required tools
    for tool in docker curl jq git openssl; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        fatal "Missing required tools: ${missing_tools[*]}"
    fi
    
    # Check Docker Compose version
    local compose_version
    if ! compose_version=$(docker compose version --short 2>/dev/null); then
        fatal "Docker Compose v2 is required"
    fi
    
    local major_version="${compose_version%%.*}"
    if [[ "$major_version" -lt 2 ]]; then
        fatal "Docker Compose v2.0+ is required (found: $compose_version)"
    fi
    
    # Check environment file
    if [[ ! -f "$ENV_FILE" ]]; then
        warn "Environment file not found: $ENV_FILE"
        info "Please create the environment file with production settings"
        return 1
    fi
    
    # Check compose files
    for compose_file in "${COMPOSE_FILES[@]}"; do
        if [[ ! -f "$compose_file" ]]; then
            fatal "Compose file not found: $compose_file"
        fi
    done
    
    # Validate compose configuration
    if ! docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" config >/dev/null 2>&1; then
        fatal "Invalid Docker Compose configuration"
    fi
    
    success "Prerequisites check completed"
}

# Validate environment configuration
validate_environment() {
    info "Validating environment configuration..."
    
    # Source environment file
    set -a
    # shellcheck source=/dev/null
    source "$ENV_FILE"
    set +a
    
    local validation_errors=()
    
    # Required variables
    local required_vars=(
        "POSTGRES_DB"
        "POSTGRES_USER" 
        "POSTGRES_PASSWORD"
        "REDIS_PASSWORD"
        "API_KEY"
        "JWT_SECRET"
        "ENCRYPTION_KEY"
        "DOMAIN"
        "DATA_PATH"
        "SSL_PATH"
        "GRAFANA_PASSWORD"
    )
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            validation_errors+=("Missing required variable: $var")
        fi
    done
    
    # Validate secret lengths
    if [[ -n "${API_KEY:-}" ]] && [[ ${#API_KEY} -lt 32 ]]; then
        validation_errors+=("API_KEY must be at least 32 characters")
    fi
    
    if [[ -n "${JWT_SECRET:-}" ]] && [[ ${#JWT_SECRET} -lt 32 ]]; then
        validation_errors+=("JWT_SECRET must be at least 32 characters")
    fi
    
    if [[ -n "${ENCRYPTION_KEY:-}" ]] && [[ ${#ENCRYPTION_KEY} -lt 32 ]]; then
        validation_errors+=("ENCRYPTION_KEY must be at least 32 characters")
    fi
    
    # Validate paths
    for path_var in "DATA_PATH" "SSL_PATH"; do
        local path_value="${!path_var:-}"
        if [[ -n "$path_value" ]] && [[ ! -d "$path_value" ]]; then
            validation_errors+=("Directory does not exist: $path_var=$path_value")
        fi
    done
    
    if [[ ${#validation_errors[@]} -gt 0 ]]; then
        error "Environment validation failed:"
        for err in "${validation_errors[@]}"; do
            error "  - $err"
        done
        return 1
    fi
    
    success "Environment validation completed"
}

# Create backup of current deployment
create_backup() {
    info "Creating backup of current deployment..."
    
    local backup_name="ffprobe-backup-${TIMESTAMP}"
    local backup_path="${BACKUP_DIR}/${backup_name}"
    
    mkdir -p "$backup_path"
    
    # Backup environment file
    if [[ -f "$ENV_FILE" ]]; then
        cp "$ENV_FILE" "${backup_path}/env.backup"
    fi
    
    # Backup compose files
    for compose_file in "${COMPOSE_FILES[@]}"; do
        if [[ -f "$compose_file" ]]; then
            cp "$compose_file" "${backup_path}/$(basename "$compose_file").backup"
        fi
    done
    
    # Export current container configurations
    if docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" ps --format json > "${backup_path}/containers.json" 2>/dev/null; then
        info "Container state backed up"
    fi
    
    # Create database backup if PostgreSQL is running
    if docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" exec -T postgres pg_isready >/dev/null 2>&1; then
        info "Creating database backup..."
        if docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" exec -T postgres pg_dump -U "${POSTGRES_USER}" "${POSTGRES_DB}" > "${backup_path}/database.sql"; then
            success "Database backup created"
        else
            warn "Database backup failed"
        fi
    fi
    
    echo "$backup_name" > "${BACKUP_DIR}/latest"
    success "Backup created: $backup_path"
}

# Build application images
build_images() {
    info "Building application images..."
    
    local build_args=(
        "--build-arg" "BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
        "--build-arg" "VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
        "--build-arg" "VERSION=${VERSION:-$(git describe --tags --always 2>/dev/null || echo 'latest')}"
    )
    
    # Build with production Dockerfile
    info "Building production image..."
    if docker build \
        -f "${PROJECT_ROOT}/Dockerfile.production" \
        --target production-runtime \
        "${build_args[@]}" \
        -t "ffprobe-api:${VERSION:-latest}" \
        -t "ffprobe-api:latest" \
        "$PROJECT_ROOT"; then
        success "Production image built successfully"
    else
        fatal "Failed to build production image"
    fi
    
    # Security scan
    if command -v trivy >/dev/null 2>&1; then
        info "Running security scan on built image..."
        if trivy image --exit-code 1 --severity HIGH,CRITICAL "ffprobe-api:${VERSION:-latest}"; then
            success "Security scan passed"
        else
            warn "Security scan found vulnerabilities - review before deploying"
        fi
    fi
}

# Deploy services with zero-downtime strategy
deploy_services() {
    info "Starting zero-downtime deployment..."
    
    local compose_cmd=(
        "docker" "compose"
        "-f" "${COMPOSE_FILES[0]}"
        "-f" "${COMPOSE_FILES[1]}"
        "--env-file" "$ENV_FILE"
    )
    
    # Pull latest images for dependencies
    info "Pulling latest dependency images..."
    "${compose_cmd[@]}" pull postgres redis nginx prometheus grafana
    
    # Start/update infrastructure services first
    info "Deploying infrastructure services..."
    "${compose_cmd[@]}" up -d postgres redis
    
    # Wait for infrastructure to be ready
    wait_for_service "postgres" "5432" "PostgreSQL"
    wait_for_service "redis" "6379" "Redis"
    
    # Deploy application with rolling update
    info "Deploying application services..."
    "${compose_cmd[@]}" up -d --scale ffprobe-api=0  # Scale down first
    sleep 10
    
    # Scale up gradually
    for scale in 1 2 3; do
        info "Scaling ffprobe-api to $scale instances..."
        "${compose_cmd[@]}" up -d --scale "ffprobe-api=$scale"
        
        # Wait for health check
        sleep 30
        if ! check_application_health; then
            error "Health check failed during scaling to $scale instances"
            rollback_deployment
            return 1
        fi
    done
    
    # Deploy reverse proxy and monitoring
    info "Deploying reverse proxy and monitoring..."
    "${compose_cmd[@]}" up -d nginx prometheus grafana
    
    success "Deployment completed successfully"
}

# Wait for service to be ready
wait_for_service() {
    local service="$1"
    local port="$2"
    local name="$3"
    local timeout="${4:-60}"
    
    info "Waiting for $name to be ready..."
    
    local elapsed=0
    while [[ $elapsed -lt $timeout ]]; do
        if docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" exec -T "$service" true >/dev/null 2>&1; then
            success "$name is ready"
            return 0
        fi
        
        sleep 5
        ((elapsed+=5))
        
        if [[ $((elapsed % 20)) -eq 0 ]]; then
            info "Still waiting for $name... (${elapsed}/${timeout}s)"
        fi
    done
    
    error "$name did not become ready within ${timeout}s"
    return 1
}

# Check application health
check_application_health() {
    info "Checking application health..."
    
    local health_endpoint="http://localhost:8080/health"
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -sf "$health_endpoint" >/dev/null 2>&1; then
            success "Application health check passed"
            return 0
        fi
        
        if [[ $((attempt % 5)) -eq 0 ]]; then
            info "Health check attempt $attempt/$max_attempts..."
        fi
        
        sleep 10
        ((attempt++))
    done
    
    error "Application health check failed after $max_attempts attempts"
    return 1
}

# Run deployment tests
run_deployment_tests() {
    info "Running deployment tests..."
    
    local test_results=()
    
    # Test API endpoints
    local api_tests=(
        "http://localhost:8080/health:Health endpoint"
        "http://localhost:8080/metrics:Metrics endpoint"
        "http://localhost:8080/api/v1/status:Status endpoint"
    )
    
    for test in "${api_tests[@]}"; do
        local url="${test%:*}"
        local description="${test#*:}"
        
        if curl -sf "$url" >/dev/null 2>&1; then
            test_results+=("✅ $description")
        else
            test_results+=("❌ $description")
        fi
    done
    
    # Test database connectivity
    if docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" exec -T postgres pg_isready >/dev/null 2>&1; then
        test_results+=("✅ PostgreSQL connectivity")
    else
        test_results+=("❌ PostgreSQL connectivity")
    fi
    
    # Test Redis connectivity
    if docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" exec -T redis redis-cli ping >/dev/null 2>&1; then
        test_results+=("✅ Redis connectivity")
    else  
        test_results+=("❌ Redis connectivity")
    fi
    
    # Display test results
    info "Deployment test results:"
    for result in "${test_results[@]}"; do
        echo "  $result"
    done
    
    # Check if any tests failed
    local failed_tests=0
    for result in "${test_results[@]}"; do
        if [[ "$result" == *"❌"* ]]; then
            ((failed_tests++))
        fi
    done
    
    if [[ $failed_tests -gt 0 ]]; then
        error "$failed_tests deployment test(s) failed"
        return 1
    fi
    
    success "All deployment tests passed"
}

# Rollback to previous deployment
rollback_deployment() {
    error "Initiating deployment rollback..."
    
    if [[ ! -f "${BACKUP_DIR}/latest" ]]; then
        fatal "No backup available for rollback"
    fi
    
    local latest_backup
    latest_backup="$(cat "${BACKUP_DIR}/latest")"
    local backup_path="${BACKUP_DIR}/${latest_backup}"
    
    if [[ ! -d "$backup_path" ]]; then
        fatal "Backup directory not found: $backup_path"
    fi
    
    info "Rolling back to backup: $latest_backup"
    
    # Restore environment file
    if [[ -f "${backup_path}/env.backup" ]]; then
        cp "${backup_path}/env.backup" "$ENV_FILE"
    fi
    
    # Restore database if available
    if [[ -f "${backup_path}/database.sql" ]]; then
        info "Restoring database backup..."
        docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" exec -T postgres psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" < "${backup_path}/database.sql"
    fi
    
    # Restart services
    docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" restart
    
    warn "Rollback completed"
}

# Show deployment status
show_status() {
    info "Current deployment status:"
    
    # Service status
    docker compose -f "${COMPOSE_FILES[0]}" -f "${COMPOSE_FILES[1]}" ps
    
    # Resource usage
    echo
    info "Resource usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}"
    
    # Application health
    echo
    info "Application health status:"
    if curl -sf "http://localhost:8080/health" | jq . 2>/dev/null; then
        success "Application is healthy"
    else
        error "Application health check failed"
    fi
}

# Clean up old images and containers
cleanup_resources() {
    info "Cleaning up unused Docker resources..."
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes (be careful in production)
    if [[ "${CLEANUP_VOLUMES:-false}" == "true" ]]; then
        warn "Cleaning up unused volumes (this may remove data!)"
        docker volume prune -f
    fi
    
    # Remove old backups (keep last 10)
    if [[ -d "$BACKUP_DIR" ]]; then
        find "$BACKUP_DIR" -maxdepth 1 -name "ffprobe-backup-*" -type d | sort -r | tail -n +11 | xargs rm -rf
    fi
    
    success "Resource cleanup completed"
}

# Main function
main() {
    local action="${1:-deploy}"
    
    case "$action" in
        "deploy")
            info "Starting production deployment..."
            check_prerequisites
            validate_environment
            create_backup
            build_images
            deploy_services
            run_deployment_tests
            success "Production deployment completed successfully!"
            ;;
        "rollback")
            rollback_deployment
            ;;
        "status")
            show_status
            ;;
        "cleanup")
            cleanup_resources
            ;;
        "test")
            run_deployment_tests
            ;;
        *)
            echo "Usage: $0 {deploy|rollback|status|cleanup|test}"
            echo
            echo "Commands:"
            echo "  deploy   - Deploy application to production"
            echo "  rollback - Rollback to previous backup"
            echo "  status   - Show current deployment status"
            echo "  cleanup  - Clean up unused resources"
            echo "  test     - Run deployment tests"
            exit 1
            ;;
    esac
}

# Handle interruption gracefully
trap 'error "Deployment interrupted by user"; exit 130' INT

# Execute main function
main "$@"