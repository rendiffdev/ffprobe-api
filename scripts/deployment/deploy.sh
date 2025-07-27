#!/bin/bash

# =============================================================================
# FFprobe API Production Deployment Script
# =============================================================================
# This script deploys the FFprobe API to production environment
# with proper security checks and rollback capabilities.
#
# Usage:
#   ./scripts/deploy.sh [environment] [version]
#   ./scripts/deploy.sh production v1.2.3
#   ./scripts/deploy.sh staging latest
#
# Prerequisites:
#   - Docker and Docker Compose installed
#   - Production environment variables configured
#   - SSL certificates in place
#   - Database migrations ready
# =============================================================================

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
LOG_FILE="/var/log/ffprobe-api-deploy.log"
BAKCUP_DIR="/var/backups/ffprobe-api"
MAX_ROLLBACK_VERSIONS=5

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="${1:-production}"
VERSION="${2:-latest}"
FORCE_DEPLOY="${3:-false}"

# =============================================================================
# Utility Functions
# =============================================================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
    
    case "$level" in
        "INFO")
            echo -e "${BLUE}[INFO]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "WARN")
            echo -e "${YELLOW}[WARN]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "ERROR")
            echo -e "${RED}[ERROR]${NC} $message" | tee -a "$LOG_FILE"
            ;;
        "SUCCESS")
            echo -e "${GREEN}[SUCCESS]${NC} $message" | tee -a "$LOG_FILE"
            ;;
    esac
    
    echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
}

check_prerequisites() {
    log "INFO" "Checking prerequisites..."
    
    # Check if running as root or with sudo for production
    if [[ "$ENVIRONMENT" == "production" && $EUID -ne 0 ]]; then
        log "ERROR" "Production deployment requires root privileges"
        exit 1
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log "ERROR" "Docker is not installed"
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker &> /dev/null || ! docker compose version &> /dev/null; then
        log "ERROR" "Docker Compose v2+ is not installed"
        exit 1
    fi
    
    # Check if environment file exists
    local env_file="$PROJECT_ROOT/.env.$ENVIRONMENT"
    if [[ ! -f "$env_file" ]]; then
        log "ERROR" "Environment file not found: $env_file"
        exit 1
    fi
    
    # Check if Docker Compose files exist
    local compose_file="$PROJECT_ROOT/compose.yml"
    local env_compose_file="$PROJECT_ROOT/compose.$ENVIRONMENT.yml"
    
    if [[ ! -f "$compose_file" ]]; then
        log "ERROR" "Docker Compose file not found: $compose_file"
        exit 1
    fi
    
    if [[ ! -f "$env_compose_file" ]]; then
        log "ERROR" "Environment-specific Docker Compose file not found: $env_compose_file"
        exit 1
    fi
    
    log "SUCCESS" "Prerequisites check passed"
}

validate_environment() {
    log "INFO" "Validating $ENVIRONMENT environment configuration..."
    
    local env_file="$PROJECT_ROOT/.env.$ENVIRONMENT"
    
    # Source environment file
    set -a  # Automatically export all variables
    source "$env_file"
    set +a
    
    # Check critical environment variables
    local required_vars=(
        "API_KEY"
        "JWT_SECRET"
        "POSTGRES_PASSWORD"
        "POSTGRES_HOST"
        "POSTGRES_DB"
    )
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            log "ERROR" "Required environment variable $var is not set"
            exit 1
        fi
    done
    
    # Check for default/insecure values in production
    if [[ "$ENVIRONMENT" == "production" ]]; then
        if [[ "$API_KEY" == *"CHANGE_THIS"* ]] || [[ ${#API_KEY} -lt 32 ]]; then
            log "ERROR" "API_KEY must be changed from default and be at least 32 characters"
            exit 1
        fi
        
        if [[ "$JWT_SECRET" == *"CHANGE_THIS"* ]] || [[ ${#JWT_SECRET} -lt 32 ]]; then
            log "ERROR" "JWT_SECRET must be changed from default and be at least 32 characters"
            exit 1
        fi
        
        if [[ "$POSTGRES_PASSWORD" == *"CHANGE_THIS"* ]] || [[ ${#POSTGRES_PASSWORD} -lt 8 ]]; then
            log "ERROR" "POSTGRES_PASSWORD must be changed from default and be at least 8 characters"
            exit 1
        fi
    fi
    
    log "SUCCESS" "Environment validation passed"
}

create_backup() {
    log "INFO" "Creating backup before deployment..."
    
    local backup_timestamp="$(date '+%Y%m%d_%H%M%S')"
    local backup_path="$BAKUP_DIR/$ENVIRONMENT/$backup_timestamp"
    
    # Create backup directory
    mkdir -p "$backup_path"
    
    # Backup database
    log "INFO" "Backing up database..."
    docker compose -f "$PROJECT_ROOT/compose.yml" -f "$PROJECT_ROOT/compose.$ENVIRONMENT.yml" exec -T postgres pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > "$backup_path/database.sql" || {
        log "WARN" "Database backup failed - continuing anyway"
    }
    
    # Backup uploaded files
    if [[ -d "$PROJECT_ROOT/uploads" ]]; then
        log "INFO" "Backing up uploaded files..."
        cp -r "$PROJECT_ROOT/uploads" "$backup_path/" || {
            log "WARN" "File backup failed - continuing anyway"
        }
    fi
    
    # Backup configuration
    log "INFO" "Backing up configuration..."
    cp "$PROJECT_ROOT/.env.$ENVIRONMENT" "$backup_path/env_backup" || {
        log "WARN" "Configuration backup failed - continuing anyway"
    }
    
    # Store current version info
    echo "$VERSION" > "$backup_path/version.txt"
    
    # Clean old backups (keep only last 5)
    find "$BAKUP_DIR/$ENVIRONMENT" -maxdepth 1 -type d | sort -r | tail -n +$((MAX_ROLLBACK_VERSIONS + 1)) | xargs rm -rf 2>/dev/null || true
    
    log "SUCCESS" "Backup created at $backup_path"
    echo "$backup_path" > /tmp/ffprobe_last_backup
}

run_health_check() {
    log "INFO" "Running health checks..."
    
    local max_attempts=30
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        if curl -f "http://localhost:8080/health" &>/dev/null; then
            log "SUCCESS" "Health check passed"
            return 0
        fi
        
        log "INFO" "Health check attempt $((attempt + 1))/$max_attempts failed, retrying in 10 seconds..."
        sleep 10
        ((attempt++))
    done
    
    log "ERROR" "Health checks failed after $max_attempts attempts"
    return 1
}

run_smoke_tests() {
    log "INFO" "Running smoke tests..."
    
    # Test API endpoints
    local endpoints=(
        "/health"
        "/metrics"
        "/api/v1/system/version"
    )
    
    for endpoint in "${endpoints[@]}"; do
        log "INFO" "Testing endpoint: $endpoint"
        if ! curl -f "http://localhost:8080$endpoint" &>/dev/null; then
            log "ERROR" "Smoke test failed for endpoint: $endpoint"
            return 1
        fi
    done
    
    log "SUCCESS" "Smoke tests passed"
}

deploy() {
    log "INFO" "Starting deployment of version $VERSION to $ENVIRONMENT..."
    
    cd "$PROJECT_ROOT"
    
    # Copy environment file
    cp ".env.$ENVIRONMENT" ".env"
    
    # Pull latest images if version is 'latest'
    if [[ "$VERSION" == "latest" ]]; then
        log "INFO" "Pulling latest Docker images..."
        docker compose -f "docker compose.yml" -f "docker compose.$ENVIRONMENT.yml" pull
    fi
    
    # Stop existing services
    log "INFO" "Stopping existing services..."
    docker compose -f "docker compose.yml" -f "docker compose.$ENVIRONMENT.yml" down --remove-orphans
    
    # Start services
    log "INFO" "Starting services..."
    docker compose -f "docker compose.yml" -f "docker compose.$ENVIRONMENT.yml" up -d
    
    # Wait for services to be ready
    log "INFO" "Waiting for services to be ready..."
    sleep 30
    
    # Run health checks
    if ! run_health_check; then
        log "ERROR" "Deployment failed health checks"
        return 1
    fi
    
    # Run smoke tests
    if ! run_smoke_tests; then
        log "ERROR" "Deployment failed smoke tests"
        return 1
    fi
    
    log "SUCCESS" "Deployment completed successfully"
}

rollback() {
    log "WARN" "Starting rollback process..."
    
    if [[ ! -f "/tmp/ffprobe_last_backup" ]]; then
        log "ERROR" "No backup information found for rollback"
        exit 1
    fi
    
    local backup_path="$(cat /tmp/ffprobe_last_backup)"
    
    if [[ ! -d "$backup_path" ]]; then
        log "ERROR" "Backup directory not found: $backup_path"
        exit 1
    fi
    
    log "INFO" "Rolling back using backup: $backup_path"
    
    # Stop current services
    docker compose -f "$PROJECT_ROOT/docker compose.yml" -f "$PROJECT_ROOT/docker compose.$ENVIRONMENT.yml" down
    
    # Restore configuration
    if [[ -f "$backup_path/env_backup" ]]; then
        cp "$backup_path/env_backup" "$PROJECT_ROOT/.env.$ENVIRONMENT"
        cp "$backup_path/env_backup" "$PROJECT_ROOT/.env"
    fi
    
    # Restore database
    if [[ -f "$backup_path/database.sql" ]]; then
        log "INFO" "Restoring database..."
        # Start database first
        docker compose -f "$PROJECT_ROOT/docker compose.yml" -f "$PROJECT_ROOT/docker compose.$ENVIRONMENT.yml" up -d postgres
        sleep 10
        docker compose -f "$PROJECT_ROOT/docker compose.yml" -f "$PROJECT_ROOT/docker compose.$ENVIRONMENT.yml" exec -T postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" < "$backup_path/database.sql"
    fi
    
    # Start all services
    docker compose -f "$PROJECT_ROOT/docker compose.yml" -f "$PROJECT_ROOT/docker compose.$ENVIRONMENT.yml" up -d
    
    # Wait and check health
    sleep 30
    if run_health_check; then
        log "SUCCESS" "Rollback completed successfully"
    else
        log "ERROR" "Rollback failed health checks"
        exit 1
    fi
}

cleanup() {
    log "INFO" "Performing cleanup..."
    
    # Remove unused Docker images
    docker image prune -f
    
    # Remove unused volumes (be careful!)
    # docker volume prune -f
    
    log "SUCCESS" "Cleanup completed"
}

show_status() {
    log "INFO" "Current deployment status:"
    
    echo "Environment: $ENVIRONMENT"
    echo "Version: $VERSION"
    echo "Services:"
    docker compose -f "$PROJECT_ROOT/docker compose.yml" -f "$PROJECT_ROOT/docker compose.$ENVIRONMENT.yml" ps
    
    echo ""
    echo "Resource usage:"
    docker stats --no-stream
    
    echo ""
    echo "Recent logs:"
    docker compose -f "$PROJECT_ROOT/docker compose.yml" -f "$PROJECT_ROOT/docker compose.$ENVIRONMENT.yml" logs --tail=20
}

show_help() {
    cat << EOF
FFprobe API Deployment Script

Usage:
    $0 [COMMAND] [ENVIRONMENT] [VERSION]

Commands:
    deploy       Deploy the application (default)
    rollback     Rollback to previous version
    status       Show current deployment status
    cleanup      Clean up unused resources
    help         Show this help message

Environments:
    production   Production environment
    staging      Staging environment
    development  Development environment

Examples:
    $0 deploy production v1.2.3
    $0 rollback production
    $0 status production
    $0 cleanup

EOF
}

# =============================================================================
# Main Script Logic
# =============================================================================

# Parse command
COMMAND="${1:-deploy}"

case "$COMMAND" in
    "deploy")
        shift
        ENVIRONMENT="${1:-production}"
        VERSION="${2:-latest}"
        
        log "INFO" "Starting deployment process..."
        check_prerequisites
        validate_environment
        create_backup
        
        if deploy; then
            cleanup
            log "SUCCESS" "🚀 Deployment successful!"
            show_status
        else
            log "ERROR" "Deployment failed, initiating rollback..."
            rollback
            exit 1
        fi
        ;;
    
    "rollback")
        shift
        ENVIRONMENT="${1:-production}"
        
        log "WARN" "Starting rollback process..."
        check_prerequisites
        rollback
        ;;
    
    "status")
        shift
        ENVIRONMENT="${1:-production}"
        show_status
        ;;
    
    "cleanup")
        cleanup
        ;;
    
    "help" | "-h" | "--help")
        show_help
        ;;
    
    *)
        log "ERROR" "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac

log "INFO" "Script completed"
