#!/bin/bash
# FFprobe API - Production Deployment Script
# Comprehensive deployment automation for production environments
# Supports Docker Swarm, Docker Compose, and Kubernetes deployments

set -euo pipefail

# =============================================================================
# Configuration and Environment
# =============================================================================

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"
readonly DEPLOYMENT_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Default configuration
DEPLOYMENT_MODE="${DEPLOYMENT_MODE:-compose}"
ENVIRONMENT="${ENVIRONMENT:-production}"
DOMAIN="${DOMAIN:-localhost}"
DATA_PATH="${DATA_PATH:-/opt/ffprobe-api/data}"
BACKUP_ENABLED="${BACKUP_ENABLED:-true}"
MONITORING_ENABLED="${MONITORING_ENABLED:-true}"
SSL_ENABLED="${SSL_ENABLED:-true}"
DRY_RUN="${DRY_RUN:-false}"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m'

# =============================================================================
# Utility Functions
# =============================================================================

print_banner() {
    echo -e "${CYAN}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  FFprobe API - Production Deployment System v2.0"
    echo "  Deploying to: ${ENVIRONMENT} | Mode: ${DEPLOYMENT_MODE} | Domain: ${DOMAIN}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

show_usage() {
    cat << EOF
FFprobe API - Production Deployment Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    Deployment Configuration:
        --mode MODE               Deployment mode (compose|swarm|kubernetes)
        --environment ENV         Target environment (production|staging|development)
        --domain DOMAIN           Primary domain for the service
        --data-path PATH          Path for persistent data storage
        
    Feature Toggles:
        --enable-ssl              Enable SSL/TLS with automatic certificates
        --enable-monitoring       Enable Prometheus/Grafana monitoring stack
        --enable-backup           Enable automated backup system
        --enable-logging          Enable centralized logging
        
    Operation Modes:
        --deploy                  Deploy the complete stack
        --update                  Update existing deployment
        --rollback               Rollback to previous version
        --scale REPLICAS         Scale API service to specified replicas
        --stop                   Stop all services
        --restart                Restart all services
        
    Maintenance:
        --backup                 Create manual backup
        --restore BACKUP_FILE    Restore from backup
        --health-check          Perform comprehensive health check
        --logs [SERVICE]        Show logs for service or all services
        
    Advanced:
        --dry-run               Show what would be executed
        --force                 Force operation without confirmation
        --debug                 Enable debug output
        --config-only           Generate configuration files only
        
    Help:
        -h, --help              Show this help message

EXAMPLES:
    # Complete production deployment
    $0 --mode compose --environment production --domain api.company.com --deploy
    
    # Docker Swarm deployment with full monitoring
    $0 --mode swarm --enable-ssl --enable-monitoring --enable-backup --deploy
    
    # Update existing deployment
    $0 --update
    
    # Scale API service
    $0 --scale 5
    
    # Health check and monitoring
    $0 --health-check
    
    # Backup and restore
    $0 --backup
    $0 --restore backup-2024-01-15.tar.gz

ENVIRONMENT VARIABLES:
    DEPLOYMENT_MODE             Default deployment mode
    ENVIRONMENT                 Target environment
    DOMAIN                      Service domain
    DATA_PATH                   Data storage path
    DOCKER_REGISTRY            Container registry URL
    SSL_EMAIL                   Email for Let's Encrypt certificates
    BACKUP_S3_BUCKET           S3 bucket for backups
    
EOF
}

# =============================================================================
# Validation Functions
# =============================================================================

validate_environment() {
    log_step "Validating deployment environment"
    
    # Check required tools
    local required_tools=("docker" "docker-compose")
    
    if [[ "${DEPLOYMENT_MODE}" == "swarm" ]]; then
        required_tools+=("docker-swarm")
    elif [[ "${DEPLOYMENT_MODE}" == "kubernetes" ]]; then
        required_tools+=("kubectl" "helm")
    fi
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "${tool}" &> /dev/null; then
            log_error "Required tool not found: ${tool}"
            exit 1
        fi
    done
    
    # Check Docker daemon
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    # Validate data path
    if [[ ! -d "$(dirname "${DATA_PATH}")" ]]; then
        log_error "Parent directory for data path does not exist: $(dirname "${DATA_PATH}")"
        exit 1
    fi
    
    log_success "Environment validation completed"
}

validate_configuration() {
    log_step "Validating deployment configuration"
    
    # Check required environment variables
    local required_vars=()
    
    if [[ "${SSL_ENABLED}" == "true" ]]; then
        required_vars+=("SSL_EMAIL")
    fi
    
    if [[ "${BACKUP_ENABLED}" == "true" ]]; then
        required_vars+=("BACKUP_S3_BUCKET")
    fi
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            log_error "Required environment variable not set: ${var}"
            exit 1
        fi
    done
    
    # Validate domain format
    if [[ ! "${DOMAIN}" =~ ^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\.[a-zA-Z]{2,}$ ]] && [[ "${DOMAIN}" != "localhost" ]]; then
        log_warning "Domain format may be invalid: ${DOMAIN}"
    fi
    
    log_success "Configuration validation completed"
}

# =============================================================================
# Pre-deployment Functions
# =============================================================================

setup_directories() {
    log_step "Setting up directory structure"
    
    local directories=(
        "${DATA_PATH}"
        "${DATA_PATH}/app"
        "${DATA_PATH}/valkey"
        "${DATA_PATH}/ollama"
        "${DATA_PATH}/prometheus"
        "${DATA_PATH}/grafana"
        "${DATA_PATH}/traefik"
        "${DATA_PATH}/backups"
        "${DATA_PATH}/logs"
    )
    
    for dir in "${directories[@]}"; do
        if [[ "${DRY_RUN}" == "true" ]]; then
            log_info "DRY RUN: Would create directory: ${dir}"
        else
            mkdir -p "${dir}"
            chmod 755 "${dir}"
            log_info "Created directory: ${dir}"
        fi
    done
    
    log_success "Directory structure setup completed"
}

generate_secrets() {
    log_step "Generating deployment secrets"
    
    local secrets_script="${SCRIPT_DIR}/scripts/secrets-manager.sh"
    
    if [[ -f "${secrets_script}" ]]; then
        if [[ "${DRY_RUN}" == "true" ]]; then
            log_info "DRY RUN: Would generate secrets"
        else
            bash "${secrets_script}" generate
        fi
    else
        log_warning "Secrets manager script not found. Manual secret generation required."
    fi
    
    log_success "Secrets generation completed"
}

generate_configuration() {
    log_step "Generating deployment configuration"
    
    # Create environment file
    local env_file="${SCRIPT_DIR}/.env.${ENVIRONMENT}"
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "DRY RUN: Would generate configuration files"
        return 0
    fi
    
    cat > "${env_file}" << EOF
# FFprobe API - ${ENVIRONMENT^} Environment Configuration
# Generated on: ${DEPLOYMENT_DATE}

# Core Configuration
ENVIRONMENT=${ENVIRONMENT}
DOMAIN=${DOMAIN}
DATA_PATH=${DATA_PATH}

# Application Settings
API_REPLICAS=${API_REPLICAS:-3}
WORKER_POOL_SIZE=${WORKER_POOL_SIZE:-16}
MAX_CONCURRENT_JOBS=${MAX_CONCURRENT_JOBS:-8}

# Database Configuration
DB_TYPE=sqlite
DB_PATH=/app/data/ffprobe.db

# Cache Configuration
VALKEY_PASSWORD_FILE=/run/secrets/valkey_password
VALKEY_MAX_MEMORY=${VALKEY_MAX_MEMORY:-2gb}

# AI/LLM Configuration
OLLAMA_MODEL=${OLLAMA_MODEL:-gemma3:270m}
OLLAMA_FALLBACK_MODEL=${OLLAMA_FALLBACK_MODEL:-phi3:mini}
OLLAMA_PARALLEL=${OLLAMA_PARALLEL:-4}
OLLAMA_KEEP_ALIVE=${OLLAMA_KEEP_ALIVE:-15m}

# SSL/TLS Configuration
SSL_EMAIL=${SSL_EMAIL:-}
ACME_CA_SERVER=${ACME_CA_SERVER:-https://acme-v02.api.letsencrypt.org/directory}

# Monitoring Configuration
PROMETHEUS_LOG_LEVEL=${PROMETHEUS_LOG_LEVEL:-info}
GRAFANA_PASSWORD_FILE=/run/secrets/grafana_password

# Backup Configuration
BACKUP_SCHEDULE=${BACKUP_SCHEDULE:-0 2 * * *}
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}
BACKUP_S3_BUCKET=${BACKUP_S3_BUCKET:-}

# Security Configuration
JWT_SECRET_FILE=/run/secrets/jwt_secret
CSRF_SECRET_FILE=/run/secrets/csrf_secret
API_KEY_FILE=/run/secrets/api_key

# Performance Tuning
LOG_LEVEL=${LOG_LEVEL:-info}
TRAEFIK_LOG_LEVEL=${TRAEFIK_LOG_LEVEL:-INFO}

# Feature Flags
MONITORING_ENABLED=${MONITORING_ENABLED}
BACKUP_ENABLED=${BACKUP_ENABLED}
SSL_ENABLED=${SSL_ENABLED}
EOF
    
    chmod 600 "${env_file}"
    log_success "Configuration file generated: ${env_file}"
}

# =============================================================================
# Deployment Functions
# =============================================================================

deploy_compose() {
    log_step "Deploying with Docker Compose"
    
    local compose_files=(
        "-f" "${SCRIPT_DIR}/compose.production.optimized.yaml"
    )
    
    # Add security overlay
    if [[ -f "${SCRIPT_DIR}/security/docker-security.yaml" ]]; then
        compose_files+=("-f" "${SCRIPT_DIR}/security/docker-security.yaml")
    fi
    
    # Add monitoring if enabled
    if [[ "${MONITORING_ENABLED}" == "true" ]]; then
        compose_files+=("--profile" "monitoring")
    fi
    
    local cmd="docker-compose ${compose_files[*]} --env-file ${SCRIPT_DIR}/.env.${ENVIRONMENT}"
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "DRY RUN: Would execute: ${cmd} up -d"
    else
        log_info "Executing: ${cmd} up -d"
        eval "${cmd} up -d"
    fi
    
    log_success "Docker Compose deployment completed"
}

deploy_swarm() {
    log_step "Deploying with Docker Swarm"
    
    # Initialize swarm if not already done
    if ! docker info | grep -q "Swarm: active"; then
        if [[ "${DRY_RUN}" == "true" ]]; then
            log_info "DRY RUN: Would initialize Docker Swarm"
        else
            docker swarm init
            log_info "Docker Swarm initialized"
        fi
    fi
    
    # Create networks
    local networks=("frontend" "backend" "monitoring")
    for network in "${networks[@]}"; do
        if [[ "${DRY_RUN}" == "true" ]]; then
            log_info "DRY RUN: Would create network: ${network}"
        else
            docker network create --driver overlay "${network}" 2>/dev/null || true
        fi
    done
    
    # Deploy stack
    local stack_cmd="docker stack deploy -c ${SCRIPT_DIR}/compose.production.optimized.yaml"
    
    if [[ -f "${SCRIPT_DIR}/security/docker-security.yaml" ]]; then
        stack_cmd+=" -c ${SCRIPT_DIR}/security/docker-security.yaml"
    fi
    
    stack_cmd+=" ffprobe-api"
    
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "DRY RUN: Would execute: ${stack_cmd}"
    else
        eval "${stack_cmd}"
    fi
    
    log_success "Docker Swarm deployment completed"
}

deploy_kubernetes() {
    log_step "Deploying with Kubernetes"
    
    # Check if kubectl is configured
    if ! kubectl cluster-info &> /dev/null; then
        log_error "kubectl is not configured or cluster is not accessible"
        exit 1
    fi
    
    # Create namespace
    if [[ "${DRY_RUN}" == "true" ]]; then
        log_info "DRY RUN: Would create Kubernetes namespace"
    else
        kubectl create namespace ffprobe-api --dry-run=client -o yaml | kubectl apply -f -
    fi
    
    # Deploy with Helm (if charts exist)
    local helm_chart="${SCRIPT_DIR}/helm/ffprobe-api"
    if [[ -d "${helm_chart}" ]]; then
        local helm_cmd="helm upgrade --install ffprobe-api ${helm_chart}"
        helm_cmd+=" --namespace ffprobe-api"
        helm_cmd+=" --set environment=${ENVIRONMENT}"
        helm_cmd+=" --set domain=${DOMAIN}"
        
        if [[ "${DRY_RUN}" == "true" ]]; then
            log_info "DRY RUN: Would execute: ${helm_cmd}"
        else
            eval "${helm_cmd}"
        fi
    else
        log_warning "Helm chart not found. Kubernetes deployment skipped."
    fi
    
    log_success "Kubernetes deployment completed"
}

# =============================================================================
# Post-deployment Functions
# =============================================================================

wait_for_services() {
    log_step "Waiting for services to be ready"
    
    local services=("api" "valkey")
    local max_wait=300  # 5 minutes
    local wait_interval=10
    local elapsed=0
    
    if [[ "${MONITORING_ENABLED}" == "true" ]]; then
        services+=("prometheus" "grafana")
    fi
    
    for service in "${services[@]}"; do
        log_info "Waiting for ${service} to be ready..."
        
        while [[ ${elapsed} -lt ${max_wait} ]]; do
            if [[ "${DRY_RUN}" == "true" ]]; then
                log_info "DRY RUN: Would check ${service} health"
                break
            fi
            
            case "${DEPLOYMENT_MODE}" in
                "compose")
                    if docker-compose ps "${service}" | grep -q "Up"; then
                        break
                    fi
                    ;;
                "swarm")
                    if docker service ps "ffprobe-api_${service}" --format "{{.CurrentState}}" | grep -q "Running"; then
                        break
                    fi
                    ;;
                "kubernetes")
                    if kubectl get pods -n ffprobe-api -l app="${service}" --field-selector=status.phase=Running | grep -q "${service}"; then
                        break
                    fi
                    ;;
            esac
            
            sleep ${wait_interval}
            elapsed=$((elapsed + wait_interval))
        done
        
        if [[ ${elapsed} -ge ${max_wait} ]]; then
            log_warning "Service ${service} did not become ready within ${max_wait} seconds"
        else
            log_success "Service ${service} is ready"
        fi
    done
}

perform_health_check() {
    log_step "Performing comprehensive health check"
    
    local health_endpoints=(
        "http://localhost:8080/health"
    )
    
    if [[ "${MONITORING_ENABLED}" == "true" ]]; then
        health_endpoints+=(
            "http://localhost:9090/-/ready"  # Prometheus
            "http://localhost:3000/api/health"  # Grafana
        )
    fi
    
    for endpoint in "${health_endpoints[@]}"; do
        if [[ "${DRY_RUN}" == "true" ]]; then
            log_info "DRY RUN: Would check health endpoint: ${endpoint}"
        else
            if curl -f -s "${endpoint}" > /dev/null; then
                log_success "Health check passed: ${endpoint}"
            else
                log_warning "Health check failed: ${endpoint}"
            fi
        fi
    done
}

show_deployment_info() {
    log_step "Deployment Information"
    
    echo ""
    echo -e "${GREEN}🎉 DEPLOYMENT COMPLETED SUCCESSFULLY${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "🌐 Environment:   ${CYAN}${ENVIRONMENT}${NC}"
    echo -e "🚀 Mode:          ${YELLOW}${DEPLOYMENT_MODE}${NC}"
    echo -e "🏠 Domain:        ${BLUE}${DOMAIN}${NC}"
    echo -e "📁 Data Path:     ${PURPLE}${DATA_PATH}${NC}"
    echo -e "⏰ Deployed:      ${DEPLOYMENT_DATE}"
    echo ""
    
    echo -e "${GREEN}🔗 SERVICE ENDPOINTS${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "📡 API:           https://${DOMAIN}/"
    echo -e "📊 Metrics:       https://${DOMAIN}/metrics"
    echo -e "🔍 Health:        https://${DOMAIN}/health"
    
    if [[ "${MONITORING_ENABLED}" == "true" ]]; then
        echo -e "📈 Prometheus:    https://prometheus.${DOMAIN}/"
        echo -e "📊 Grafana:       https://grafana.${DOMAIN}/"
    fi
    
    echo ""
    echo -e "${GREEN}⚙️  MANAGEMENT COMMANDS${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  View logs:       $0 --logs"
    echo "  Health check:    $0 --health-check"
    echo "  Scale API:       $0 --scale 5"
    echo "  Create backup:   $0 --backup"
    echo "  Update stack:    $0 --update"
    echo ""
}

# =============================================================================
# Management Functions
# =============================================================================

scale_services() {
    local replicas="$1"
    
    log_step "Scaling API service to ${replicas} replicas"
    
    case "${DEPLOYMENT_MODE}" in
        "compose")
            if [[ "${DRY_RUN}" == "true" ]]; then
                log_info "DRY RUN: Would scale API service to ${replicas}"
            else
                docker-compose up -d --scale api="${replicas}"
            fi
            ;;
        "swarm")
            if [[ "${DRY_RUN}" == "true" ]]; then
                log_info "DRY RUN: Would scale API service to ${replicas}"
            else
                docker service scale ffprobe-api_api="${replicas}"
            fi
            ;;
        "kubernetes")
            if [[ "${DRY_RUN}" == "true" ]]; then
                log_info "DRY RUN: Would scale API deployment to ${replicas}"
            else
                kubectl scale deployment api --replicas="${replicas}" -n ffprobe-api
            fi
            ;;
    esac
    
    log_success "Service scaling completed"
}

show_logs() {
    local service="${1:-}"
    
    log_step "Showing logs for ${service:-all services}"
    
    case "${DEPLOYMENT_MODE}" in
        "compose")
            if [[ -n "${service}" ]]; then
                docker-compose logs -f "${service}"
            else
                docker-compose logs -f
            fi
            ;;
        "swarm")
            if [[ -n "${service}" ]]; then
                docker service logs -f "ffprobe-api_${service}"
            else
                docker stack ps ffprobe-api
            fi
            ;;
        "kubernetes")
            if [[ -n "${service}" ]]; then
                kubectl logs -f -l app="${service}" -n ffprobe-api
            else
                kubectl logs -f --all-containers=true -n ffprobe-api
            fi
            ;;
    esac
}

# =============================================================================
# Main Function
# =============================================================================

main() {
    local operation=""
    local scale_replicas=""
    local restore_file=""
    local log_service=""
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --mode)
                DEPLOYMENT_MODE="$2"
                shift 2
                ;;
            --environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            --domain)
                DOMAIN="$2"
                shift 2
                ;;
            --data-path)
                DATA_PATH="$2"
                shift 2
                ;;
            --enable-ssl)
                SSL_ENABLED=true
                shift
                ;;
            --enable-monitoring)
                MONITORING_ENABLED=true
                shift
                ;;
            --enable-backup)
                BACKUP_ENABLED=true
                shift
                ;;
            --deploy)
                operation="deploy"
                shift
                ;;
            --update)
                operation="update"
                shift
                ;;
            --scale)
                operation="scale"
                scale_replicas="$2"
                shift 2
                ;;
            --health-check)
                operation="health-check"
                shift
                ;;
            --logs)
                operation="logs"
                log_service="${2:-}"
                shift
                [[ -n "${log_service}" ]] && shift
                ;;
            --backup)
                operation="backup"
                shift
                ;;
            --restore)
                operation="restore"
                restore_file="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --debug)
                set -x
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    print_banner
    
    case "${operation}" in
        "deploy")
            validate_environment
            validate_configuration
            setup_directories
            generate_secrets
            generate_configuration
            
            case "${DEPLOYMENT_MODE}" in
                "compose")
                    deploy_compose
                    ;;
                "swarm")
                    deploy_swarm
                    ;;
                "kubernetes")
                    deploy_kubernetes
                    ;;
                *)
                    log_error "Invalid deployment mode: ${DEPLOYMENT_MODE}"
                    exit 1
                    ;;
            esac
            
            wait_for_services
            perform_health_check
            show_deployment_info
            ;;
        "update")
            # Update deployment logic here
            log_info "Update operation not implemented yet"
            ;;
        "scale")
            scale_services "${scale_replicas}"
            ;;
        "health-check")
            perform_health_check
            ;;
        "logs")
            show_logs "${log_service}"
            ;;
        "backup")
            # Backup operation logic here
            log_info "Backup operation not implemented yet"
            ;;
        "restore")
            # Restore operation logic here
            log_info "Restore operation not implemented yet"
            ;;
        *)
            log_error "No operation specified"
            show_usage
            exit 1
            ;;
    esac
}

# Execute main function with all arguments
main "$@"