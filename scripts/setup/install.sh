#!/bin/bash

# FFprobe API Interactive Installer
# This script collects user preferences and sets up the entire environment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Installer configuration
INSTALLER_VERSION="1.0.0"
MIN_DOCKER_VERSION="24.0.0"
MIN_COMPOSE_VERSION="2.20.0"
PROJECT_NAME="ffprobe-api"

# Global variables
INSTALL_DIR=""
DEPLOYMENT_MODE=""
DOMAIN_NAME=""
EMAIL=""
ENABLE_SSL=""
DATA_PATH=""
BACKUP_PATH=""
LOG_LEVEL=""

# Configuration arrays
declare -A CONFIG
declare -A SECRETS
declare -A NETWORK_CONFIG
declare -A RESOURCES

# =============================================================================
# Utility Functions
# =============================================================================

print_banner() {
    clear
    echo -e "${CYAN}"
    cat << "EOF"
    ████████╗████████╗██████╗ ██████╗  ██████╗ ██████╗ ███████╗     █████╗ ██████╗ ██╗
    ██╔══════╝██╔══════╝██╔══██╗██╔══██╗██╔═══██╗██╔══██╗██╔════╝    ██╔══██╗██╔══██╗██║
    ████████╗████████╗██████╔╝██████╔╝██║   ██║██████╔╝█████╗      ███████║██████╔╝██║
    ██╔══════╝██╔══════╝██╔═══╝ ██╔══██╗██║   ██║██╔══██╗██╔══╝      ██╔══██║██╔═══╝ ██║
    ██║       ██║      ██║     ██║  ██║╚██████╔╝██████╔╝███████╗    ██║  ██║██║     ██║
    ╚═╝       ╚═╝      ╚═╝     ╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝    ╚═╝  ╚═╝╚═╝     ╚═╝
    
    🎬 Enterprise-Grade Media Analysis API with Netflix VMAF Integration
    📦 Interactive Installer v${INSTALLER_VERSION}
EOF
    echo -e "${NC}"
}

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

info() {
    echo -e "${BLUE}[INFO] $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

prompt() {
    echo -e "${PURPLE}$1${NC}"
}

# Validate input with pattern
validate_input() {
    local input="$1"
    local pattern="$2"
    local error_msg="$3"
    
    if [[ ! $input =~ $pattern ]]; then
        error "$error_msg"
    fi
}

# Generate secure random string
generate_secret() {
    local length=${1:-32}
    openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Compare version numbers
version_ge() {
    printf '%s\n%s\n' "$2" "$1" | sort -V -C
}

# =============================================================================
# System Requirements Check
# =============================================================================

check_requirements() {
    log "Checking system requirements..."
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        error "Please do not run this installer as root. Use a regular user with sudo privileges."
    fi
    
    # Check sudo access
    if ! sudo -n true 2>/dev/null; then
        warn "This installer requires sudo privileges. You may be prompted for your password."
        if ! sudo true; then
            error "Sudo access is required for installation."
        fi
    fi
    
    # Check Docker
    if ! command_exists docker; then
        error "Docker is not installed. Please install Docker first: https://docs.docker.com/get-docker/"
    fi
    
    local docker_version=$(docker --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    if ! version_ge "$docker_version" "$MIN_DOCKER_VERSION"; then
        error "Docker version $docker_version is too old. Minimum required: $MIN_DOCKER_VERSION"
    fi
    
    # Check Docker Compose
    if ! command_exists "docker compose"; then
        error "Docker Compose v2 is not installed or not available as 'docker compose'"
    fi
    
    local compose_version=$(docker compose version --short 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    if ! version_ge "$compose_version" "$MIN_COMPOSE_VERSION"; then
        error "Docker Compose version $compose_version is too old. Minimum required: $MIN_COMPOSE_VERSION"
    fi
    
    # Check available disk space (minimum 10GB)
    local available_space=$(df . | awk 'NR==2{print $4}')
    local min_space=$((10 * 1024 * 1024)) # 10GB in KB
    
    if [[ $available_space -lt $min_space ]]; then
        error "Insufficient disk space. Required: 10GB, Available: $(($available_space / 1024 / 1024))GB"
    fi
    
    # Check memory (minimum 4GB recommended)
    local total_memory=$(free -m | awk 'NR==2{print $2}')
    if [[ $total_memory -lt 4096 ]]; then
        warn "System has less than 4GB RAM. Performance may be limited."
    fi
    
    success "System requirements check passed"
}

# =============================================================================
# Installation Mode Selection
# =============================================================================

select_deployment_mode() {
    log "Selecting deployment mode..."
    
    echo ""
    prompt "🚀 Select your deployment mode:"
    echo "1. 🔧 Development - Local development with debugging tools"
    echo "2. 🧪 Staging - Pre-production testing environment"
    echo "3. 🏭 Production - Full production deployment with security"
    echo "4. 🐳 Docker Swarm - Multi-node Docker Swarm deployment"
    echo "5. ☸️  Kubernetes - Generate Kubernetes manifests"
    echo ""
    
    while true; do
        read -p "Enter your choice (1-5): " mode_choice
        case $mode_choice in
            1)
                DEPLOYMENT_MODE="development"
                CONFIG[environment]="development"
                CONFIG[log_level]="debug"
                CONFIG[enable_auth]="false"
                CONFIG[enable_rate_limit]="false"
                break
                ;;
            2)
                DEPLOYMENT_MODE="staging"
                CONFIG[environment]="staging"
                CONFIG[log_level]="info"
                CONFIG[enable_auth]="true"
                CONFIG[enable_rate_limit]="true"
                break
                ;;
            3)
                DEPLOYMENT_MODE="production"
                CONFIG[environment]="production"
                CONFIG[log_level]="warn"
                CONFIG[enable_auth]="true"
                CONFIG[enable_rate_limit]="true"
                CONFIG[enable_csrf]="true"
                break
                ;;
            4)
                DEPLOYMENT_MODE="swarm"
                CONFIG[environment]="production"
                CONFIG[log_level]="warn"
                CONFIG[enable_auth]="true"
                CONFIG[enable_rate_limit]="true"
                break
                ;;
            5)
                DEPLOYMENT_MODE="kubernetes"
                CONFIG[environment]="production"
                CONFIG[log_level]="info"
                CONFIG[enable_auth]="true"
                CONFIG[enable_rate_limit]="true"
                break
                ;;
            *)
                error "Invalid choice. Please select 1-5."
                ;;
        esac
    done
    
    success "Selected deployment mode: $DEPLOYMENT_MODE"
}

# =============================================================================
# Directory and Path Configuration
# =============================================================================

configure_paths() {
    log "Configuring installation paths..."
    
    # Installation directory
    local default_install_dir="$HOME/ffprobe-api"
    read -p "📁 Installation directory [$default_install_dir]: " user_install_dir
    INSTALL_DIR="${user_install_dir:-$default_install_dir}"
    
    # Validate and create directory
    if [[ -d "$INSTALL_DIR" ]]; then
        warn "Directory $INSTALL_DIR already exists."
        read -p "Do you want to continue? (y/N): " continue_install
        if [[ ! "$continue_install" =~ ^[Yy]$ ]]; then
            error "Installation cancelled."
        fi
    fi
    
    mkdir -p "$INSTALL_DIR"
    CONFIG[install_dir]="$INSTALL_DIR"
    
    # Data directory
    local default_data_dir="$INSTALL_DIR/data"
    read -p "💾 Data storage directory [$default_data_dir]: " user_data_dir
    DATA_PATH="${user_data_dir:-$default_data_dir}"
    CONFIG[data_path]="$DATA_PATH"
    
    # Backup directory
    local default_backup_dir="$INSTALL_DIR/backups"
    read -p "🔄 Backup directory [$default_backup_dir]: " user_backup_dir
    BACKUP_PATH="${user_backup_dir:-$default_backup_dir}"
    CONFIG[backup_path]="$BACKUP_PATH"
    
    # Create directories
    mkdir -p "$DATA_PATH"/{postgres,redis,uploads,reports,models,logs,temp,cache,prometheus,grafana}
    mkdir -p "$BACKUP_PATH"
    
    success "Paths configured successfully"
}

# =============================================================================
# Network Configuration
# =============================================================================

configure_network() {
    log "Configuring network settings..."
    
    # Domain name (for production/staging)
    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        while true; do
            read -p "🌐 Domain name (e.g., api.yourcompany.com): " domain_input
            if [[ -n "$domain_input" ]]; then
                validate_input "$domain_input" '^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$' "Invalid domain name format"
                DOMAIN_NAME="$domain_input"
                CONFIG[domain_name]="$DOMAIN_NAME"
                break
            else
                warn "Domain name is required for $DEPLOYMENT_MODE mode"
            fi
        done
    fi
    
    # Email for SSL certificates
    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        while true; do
            read -p "📧 Email address (for SSL certificates): " email_input
            if [[ "$email_input" =~ ^[^@]+@[^@]+\.[^@]+$ ]]; then
                EMAIL="$email_input"
                CONFIG[email]="$EMAIL"
                break
            else
                warn "Please enter a valid email address"
            fi
        done
    fi
    
    # SSL Configuration
    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        echo ""
        prompt "🔒 SSL/TLS Configuration:"
        echo "1. 🔥 Let's Encrypt (Automatic, Free)"
        echo "2. 📜 Custom certificates (Provide your own)"
        echo "3. 🚫 Disable SSL (Not recommended for production)"
        echo ""
        
        while true; do
            read -p "Select SSL option (1-3): " ssl_choice
            case $ssl_choice in
                1)
                    ENABLE_SSL="letsencrypt"
                    CONFIG[ssl_type]="letsencrypt"
                    break
                    ;;
                2)
                    ENABLE_SSL="custom"
                    CONFIG[ssl_type]="custom"
                    read -p "📁 SSL certificate file path: " ssl_cert_path
                    read -p "🔑 SSL private key file path: " ssl_key_path
                    CONFIG[ssl_cert_path]="$ssl_cert_path"
                    CONFIG[ssl_key_path]="$ssl_key_path"
                    break
                    ;;
                3)
                    ENABLE_SSL="disabled"
                    CONFIG[ssl_type]="disabled"
                    if [[ "$DEPLOYMENT_MODE" == "production" ]]; then
                        warn "Disabling SSL in production is not recommended!"
                        read -p "Are you sure? (y/N): " confirm_no_ssl
                        if [[ ! "$confirm_no_ssl" =~ ^[Yy]$ ]]; then
                            continue
                        fi
                    fi
                    break
                    ;;
                *)
                    warn "Invalid choice. Please select 1-3."
                    ;;
            esac
        done
    fi
    
    # API Port
    read -p "🔌 API port [8080]: " api_port
    CONFIG[api_port]="${api_port:-8080}"
    
    # Load balancer ports (if not development)
    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        read -p "🌐 HTTP port [80]: " http_port
        read -p "🔒 HTTPS port [443]: " https_port
        CONFIG[http_port]="${http_port:-80}"
        CONFIG[https_port]="${https_port:-443}"
    fi
    
    success "Network configuration completed"
}

# =============================================================================
# Security Configuration
# =============================================================================

configure_security() {
    log "Configuring security settings..."
    
    echo ""
    prompt "🔐 Security Configuration"
    echo ""
    
    # Generate or collect API Key
    echo "🔑 API Key Configuration:"
    echo "1. 🎲 Generate random API key (Recommended)"
    echo "2. ✏️  Enter custom API key (32+ characters)"
    echo ""
    
    while true; do
        read -p "Select API key option (1-2): " api_key_choice
        case $api_key_choice in
            1)
                SECRETS[api_key]=$(generate_secret 32)
                info "Generated API key: ${SECRETS[api_key]}"
                break
                ;;
            2)
                while true; do
                    read -s -p "Enter API key (32+ characters): " custom_api_key
                    echo ""
                    if [[ ${#custom_api_key} -ge 32 ]]; then
                        SECRETS[api_key]="$custom_api_key"
                        break 2
                    else
                        warn "API key must be at least 32 characters long"
                    fi
                done
                ;;
            *)
                warn "Invalid choice. Please select 1-2."
                ;;
        esac
    done
    
    # JWT Secret
    echo ""
    echo "🎫 JWT Secret Configuration:"
    echo "1. 🎲 Generate random JWT secret (Recommended)"
    echo "2. ✏️  Enter custom JWT secret (32+ characters)"
    echo ""
    
    while true; do
        read -p "Select JWT secret option (1-2): " jwt_choice
        case $jwt_choice in
            1)
                SECRETS[jwt_secret]=$(generate_secret 32)
                info "Generated JWT secret: ${SECRETS[jwt_secret]}"
                break
                ;;
            2)
                while true; do
                    read -s -p "Enter JWT secret (32+ characters): " custom_jwt_secret
                    echo ""
                    if [[ ${#custom_jwt_secret} -ge 32 ]]; then
                        SECRETS[jwt_secret]="$custom_jwt_secret"
                        break 2
                    else
                        warn "JWT secret must be at least 32 characters long"
                    fi
                done
                ;;
            *)
                warn "Invalid choice. Please select 1-2."
                ;;
        esac
    done
    
    # Database Password
    echo ""
    echo "🗄️ Database Password:"
    echo "1. 🎲 Generate random password (Recommended)"
    echo "2. ✏️  Enter custom password"
    echo ""
    
    while true; do
        read -p "Select database password option (1-2): " db_pass_choice
        case $db_pass_choice in
            1)
                SECRETS[postgres_password]=$(generate_secret 24)
                info "Generated database password: ${SECRETS[postgres_password]}"
                break
                ;;
            2)
                read -s -p "Enter database password: " custom_db_pass
                echo ""
                SECRETS[postgres_password]="$custom_db_pass"
                break
                ;;
            *)
                warn "Invalid choice. Please select 1-2."
                ;;
        esac
    done
    
    # Redis Password
    echo ""
    echo "🔴 Redis Password:"
    echo "1. 🎲 Generate random password (Recommended)"
    echo "2. ✏️  Enter custom password"
    echo ""
    
    while true; do
        read -p "Select Redis password option (1-2): " redis_pass_choice
        case $redis_pass_choice in
            1)
                SECRETS[redis_password]=$(generate_secret 24)
                info "Generated Redis password: ${SECRETS[redis_password]}"
                break
                ;;
            2)
                read -s -p "Enter Redis password: " custom_redis_pass
                echo ""
                SECRETS[redis_password]="$custom_redis_pass"
                break
                ;;
            *)
                warn "Invalid choice. Please select 1-2."
                ;;
        esac
    done
    
    # Grafana Admin Password
    echo ""
    echo "📊 Grafana Admin Password:"
    echo "1. 🎲 Generate random password (Recommended)"
    echo "2. ✏️  Enter custom password"
    echo ""
    
    while true; do
        read -p "Select Grafana password option (1-2): " grafana_pass_choice
        case $grafana_pass_choice in
            1)
                SECRETS[grafana_password]=$(generate_secret 16)
                info "Generated Grafana password: ${SECRETS[grafana_password]}"
                break
                ;;
            2)
                read -s -p "Enter Grafana admin password: " custom_grafana_pass
                echo ""
                SECRETS[grafana_password]="$custom_grafana_pass"
                break
                ;;
            *)
                warn "Invalid choice. Please select 1-2."
                ;;
        esac
    done
    
    # Rate Limiting Configuration
    if [[ "${CONFIG[enable_rate_limit]}" == "true" ]]; then
        echo ""
        prompt "⚡ Rate Limiting Configuration:"
        read -p "Requests per minute [60]: " rate_per_minute
        read -p "Requests per hour [1000]: " rate_per_hour
        read -p "Requests per day [10000]: " rate_per_day
        
        CONFIG[rate_limit_per_minute]="${rate_per_minute:-60}"
        CONFIG[rate_limit_per_hour]="${rate_per_hour:-1000}"
        CONFIG[rate_limit_per_day]="${rate_per_day:-10000}"
    fi
    
    success "Security configuration completed"
}

# =============================================================================
# Resource Configuration
# =============================================================================

configure_resources() {
    log "Configuring resource allocation..."
    
    echo ""
    prompt "💻 Resource Configuration"
    echo ""
    
    # API Service Resources
    echo "🎬 FFprobe API Service:"
    read -p "Memory limit (e.g., 8G, 4G) [8G]: " api_memory
    read -p "CPU limit (e.g., 4.0, 2.0) [4.0]: " api_cpu
    read -p "Memory reservation (e.g., 2G, 1G) [2G]: " api_memory_res
    read -p "CPU reservation (e.g., 1.0, 0.5) [1.0]: " api_cpu_res
    
    RESOURCES[api_memory_limit]="${api_memory:-8G}"
    RESOURCES[api_cpu_limit]="${api_cpu:-4.0}"
    RESOURCES[api_memory_reservation]="${api_memory_res:-2G}"
    RESOURCES[api_cpu_reservation]="${api_cpu_res:-1.0}"
    
    # Database Resources
    echo ""
    echo "🗄️ PostgreSQL Database:"
    read -p "Memory limit [2G]: " db_memory
    read -p "CPU limit [2.0]: " db_cpu
    read -p "Memory reservation [1G]: " db_memory_res
    read -p "CPU reservation [1.0]: " db_cpu_res
    
    RESOURCES[db_memory_limit]="${db_memory:-2G}"
    RESOURCES[db_cpu_limit]="${db_cpu:-2.0}"
    RESOURCES[db_memory_reservation]="${db_memory_res:-1G}"
    RESOURCES[db_cpu_reservation]="${db_cpu_res:-1.0}"
    
    # Redis Resources
    echo ""
    echo "🔴 Redis Cache:"
    read -p "Memory limit [512M]: " redis_memory
    read -p "CPU limit [1.0]: " redis_cpu
    
    RESOURCES[redis_memory_limit]="${redis_memory:-512M}"
    RESOURCES[redis_cpu_limit]="${redis_cpu:-1.0}"
    
    # File Upload Limits
    echo ""
    echo "📁 File Upload Configuration:"
    read -p "Max file size (e.g., 50G, 10G) [50G]: " max_file_size
    read -p "Max concurrent jobs [4]: " max_concurrent_jobs
    
    CONFIG[max_file_size]="${max_file_size:-50G}"
    CONFIG[max_concurrent_jobs]="${max_concurrent_jobs:-4}"
    
    success "Resource configuration completed"
}

# =============================================================================
# Advanced Configuration
# =============================================================================

configure_advanced() {
    log "Configuring advanced settings..."
    
    echo ""
    prompt "⚙️ Advanced Configuration (Optional)"
    echo ""
    
    # Backup Configuration
    echo "🔄 Backup Settings:"
    read -p "Enable automatic backups? (y/N): " enable_backups
    if [[ "$enable_backups" =~ ^[Yy]$ ]]; then
        CONFIG[enable_backups]="true"
        read -p "Backup retention days [30]: " backup_retention
        read -p "Backup schedule (cron format) [0 2 * * *]: " backup_schedule
        CONFIG[backup_retention]="${backup_retention:-30}"
        CONFIG[backup_schedule]="${backup_schedule:-0 2 * * *}"
    fi
    
    # Monitoring Configuration
    echo ""
    echo "📊 Monitoring & Alerting:"
    read -p "Enable Prometheus metrics? (Y/n): " enable_prometheus
    CONFIG[enable_prometheus]="${enable_prometheus:-Y}"
    
    read -p "Enable Grafana dashboards? (Y/n): " enable_grafana
    CONFIG[enable_grafana]="${enable_grafana:-Y}"
    
    if [[ "${CONFIG[enable_grafana]}" =~ ^[Yy]$ ]]; then
        read -p "Grafana admin username [admin]: " grafana_user
        CONFIG[grafana_user]="${grafana_user:-admin}"
    fi
    
    # Cloud Storage
    echo ""
    echo "☁️ Cloud Storage Integration (Optional):"
    echo "1. 🚫 Local storage only"
    echo "2. 📦 Amazon S3"
    echo "3. 🌍 Google Cloud Storage" 
    echo "4. 🔷 Azure Blob Storage"
    echo ""
    
    read -p "Select storage option (1-4) [1]: " storage_choice
    case "${storage_choice:-1}" in
        1)
            CONFIG[storage_type]="local"
            ;;
        2)
            CONFIG[storage_type]="s3"
            read -p "AWS Access Key ID: " aws_access_key
            read -s -p "AWS Secret Access Key: " aws_secret_key
            echo ""
            read -p "S3 Bucket Name: " s3_bucket
            read -p "AWS Region [us-east-1]: " aws_region
            
            SECRETS[aws_access_key_id]="$aws_access_key"
            SECRETS[aws_secret_access_key]="$aws_secret_key"
            CONFIG[s3_bucket]="$s3_bucket"
            CONFIG[aws_region]="${aws_region:-us-east-1}"
            ;;
        3)
            CONFIG[storage_type]="gcs"
            read -p "GCS Bucket Name: " gcs_bucket
            read -p "Google Cloud Project ID: " gcp_project_id
            read -p "Service Account Key File Path: " gcp_key_file
            
            CONFIG[gcs_bucket]="$gcs_bucket"
            CONFIG[gcp_project_id]="$gcp_project_id"
            CONFIG[gcp_key_file]="$gcp_key_file"
            ;;
        4)
            CONFIG[storage_type]="azure"
            read -p "Azure Storage Account: " azure_account
            read -s -p "Azure Storage Key: " azure_key
            echo ""
            read -p "Azure Container Name: " azure_container
            
            CONFIG[azure_storage_account]="$azure_account"
            SECRETS[azure_storage_key]="$azure_key"
            CONFIG[azure_container]="$azure_container"
            ;;
    esac
    
    # VMAF Models
    echo ""
    echo "🎯 VMAF Quality Models:"
    read -p "Download additional VMAF models? (Y/n): " download_vmaf
    CONFIG[download_vmaf_models]="${download_vmaf:-Y}"
    
    success "Advanced configuration completed"
}

# =============================================================================
# LLM CONFIGURATION
# =============================================================================

configure_llm() {
    log "Configuring AI/LLM settings..."
    
    echo "🤖 AI/LLM Configuration:"
    echo "The FFprobe API can use AI models to generate intelligent analysis reports."
    echo "Choose your preferred configuration:"
    echo ""
    echo "1. 🦙 Local LLM only (Ollama) - Private, no API costs"
    echo "2. ☁️  Cloud LLM only (OpenRouter) - Requires API key, always available"
    echo "3. 🔄 Local + Cloud fallback (Recommended) - Best of both worlds"
    echo "4. ❌ Disable AI features - Basic analysis only"
    echo ""
    
    while true; do
        read -p "Choose LLM configuration (1-4): " llm_choice
        case "$llm_choice" in
            1)
                CONFIG[llm_mode]="local"
                CONFIG[enable_local_llm]="true"
                CONFIG[enable_openrouter]="false"
                configure_local_llm
                break
                ;;
            2)
                CONFIG[llm_mode]="cloud"
                CONFIG[enable_local_llm]="false"
                CONFIG[enable_openrouter]="true"
                configure_openrouter
                break
                ;;
            3)
                CONFIG[llm_mode]="hybrid"
                CONFIG[enable_local_llm]="true"
                CONFIG[enable_openrouter]="true"
                configure_local_llm
                configure_openrouter
                break
                ;;
            4)
                CONFIG[llm_mode]="disabled"
                CONFIG[enable_local_llm]="false"
                CONFIG[enable_openrouter]="false"
                warn "AI features will be disabled. Analysis reports will be basic technical data only."
                break
                ;;
            *)
                warn "Please enter 1, 2, 3, or 4"
                ;;
        esac
    done
    
    success "LLM configuration completed"
}

configure_local_llm() {
    echo ""
    echo "🦙 Local LLM (Ollama) Configuration:"
    echo "Ollama will run locally in a Docker container with your chosen model."
    echo ""
    
    # Check system resources
    local ram_gb=8
    if command -v free >/dev/null 2>&1; then
        ram_gb=$(free -g | awk '/^Mem:/{print $2}')
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        local ram_bytes=$(sysctl -n hw.memsize 2>/dev/null || echo "8589934592")
        ram_gb=$((ram_bytes / 1024 / 1024 / 1024))
    fi
    
    echo "📊 Detected System RAM: ${ram_gb}GB"
    echo ""
    
    # Model recommendations based on RAM
    echo "🎯 Recommended Models:"
    if [ "$ram_gb" -lt 4 ]; then
        echo "  1. qwen2:0.5b (352MB) - Ultra-lightweight"
        echo "  2. tinyllama:1.1b (637MB) - Minimal resources"
        local default_model="qwen2:0.5b"
    elif [ "$ram_gb" -lt 8 ]; then
        echo "  1. qwen2:1.5b (934MB) - Fast and efficient ⭐"
        echo "  2. phi3:mini (2.3GB) - Good accuracy"
        echo "  3. gemma2:2b (1.6GB) - Balanced"
        local default_model="qwen2:1.5b"
    else
        echo "  1. mistral:7b (4.1GB) - Best overall ⭐ RECOMMENDED"
        echo "  2. qwen2:7b (4.4GB) - Multilingual support"
        echo "  3. llama3.1:8b (4.7GB) - Highest accuracy"
        echo "  4. qwen2:1.5b (934MB) - Fast development"
        local default_model="mistral:7b"
    fi
    echo ""
    
    read -p "Select Ollama model [$default_model]: " ollama_model
    CONFIG[ollama_model]="${ollama_model:-$default_model}"
    
    # Resource allocation
    echo ""
    echo "💾 Resource Allocation:"
    local default_memory="6G"
    if [ "$ram_gb" -lt 8 ]; then
        default_memory="3G"
    fi
    
    read -p "Ollama memory limit [$default_memory]: " ollama_memory
    CONFIG[ollama_memory]="${ollama_memory:-$default_memory}"
    
    read -p "Max parallel requests [4]: " ollama_parallel
    CONFIG[ollama_parallel]="${ollama_parallel:-4}"
    
    # GPU support
    if command -v nvidia-smi >/dev/null 2>&1; then
        echo ""
        echo "🎮 GPU Support:"
        echo "NVIDIA GPU detected. Enable GPU acceleration?"
        read -p "Enable GPU support? (Y/n): " enable_gpu
        CONFIG[ollama_gpu]="${enable_gpu:-Y}"
    else
        CONFIG[ollama_gpu]="false"
    fi
}

configure_openrouter() {
    echo ""
    echo "☁️  OpenRouter Cloud LLM Configuration:"
    echo "OpenRouter provides access to multiple AI models via API."
    echo "Get your API key from: https://openrouter.ai/keys"
    echo ""
    
    while true; do
        read -s -p "🔑 OpenRouter API Key: " openrouter_key
        echo
        if [[ -n "$openrouter_key" && ${#openrouter_key} -ge 10 ]]; then
            SECRETS[openrouter_api_key]="$openrouter_key"
            break
        else
            warn "Please enter a valid OpenRouter API key (minimum 10 characters)"
        fi
    done
    
    # Model selection
    echo ""
    echo "🎯 Cloud Model Selection:"
    echo "  1. anthropic/claude-3-haiku (Fast, cost-effective) ⭐"
    echo "  2. anthropic/claude-3-sonnet (Balanced performance)"
    echo "  3. openai/gpt-4o-mini (OpenAI, efficient)"
    echo "  4. meta-llama/llama-3.1-8b-instruct (Open source)"
    echo ""
    
    while true; do
        read -p "Select cloud model (1-4) [1]: " cloud_model_choice
        case "${cloud_model_choice:-1}" in
            1)
                CONFIG[openrouter_model]="anthropic/claude-3-haiku"
                break
                ;;
            2)
                CONFIG[openrouter_model]="anthropic/claude-3-sonnet"
                break
                ;;
            3)
                CONFIG[openrouter_model]="openai/gpt-4o-mini"
                break
                ;;
            4)
                CONFIG[openrouter_model]="meta-llama/llama-3.1-8b-instruct"
                break
                ;;
            *)
                warn "Please enter 1, 2, 3, or 4"
                ;;
        esac
    done
    
    # Usage preferences
    echo ""
    echo "📊 Usage Configuration:"
    read -p "Request timeout (seconds) [120]: " openrouter_timeout
    CONFIG[openrouter_timeout]="${openrouter_timeout:-120}"
    
    read -p "Max tokens per request [2000]: " openrouter_max_tokens
    CONFIG[openrouter_max_tokens]="${openrouter_max_tokens:-2000}"
}

# =============================================================================
# Configuration Summary and Confirmation
# =============================================================================

show_configuration_summary() {
    clear
    log "Configuration Summary"
    echo ""
    
    echo -e "${CYAN}🚀 Deployment Configuration${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "%-25s: %s\n" "Deployment Mode" "$DEPLOYMENT_MODE"
    printf "%-25s: %s\n" "Environment" "${CONFIG[environment]}"
    printf "%-25s: %s\n" "Installation Directory" "${CONFIG[install_dir]}"
    printf "%-25s: %s\n" "Data Directory" "${CONFIG[data_path]}"
    printf "%-25s: %s\n" "Log Level" "${CONFIG[log_level]}"
    echo ""
    
    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        echo -e "${CYAN}🌐 Network Configuration${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        printf "%-25s: %s\n" "Domain Name" "${CONFIG[domain_name]}"
        printf "%-25s: %s\n" "Email" "${CONFIG[email]}"
        printf "%-25s: %s\n" "SSL Type" "${CONFIG[ssl_type]}"
        printf "%-25s: %s\n" "HTTP Port" "${CONFIG[http_port]}"
        printf "%-25s: %s\n" "HTTPS Port" "${CONFIG[https_port]}"
        echo ""
    fi
    
    echo -e "${CYAN}🔐 Security Configuration${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "%-25s: %s\n" "API Key" "${SECRETS[api_key]:0:8}...***"
    printf "%-25s: %s\n" "JWT Secret" "${SECRETS[jwt_secret]:0:8}...***"
    printf "%-25s: %s\n" "Database Password" "${SECRETS[postgres_password]:0:4}...***"
    printf "%-25s: %s\n" "Redis Password" "${SECRETS[redis_password]:0:4}...***"
    printf "%-25s: %s\n" "Enable Authentication" "${CONFIG[enable_auth]}"
    printf "%-25s: %s\n" "Enable Rate Limiting" "${CONFIG[enable_rate_limit]}"
    echo ""
    
    echo -e "${CYAN}💻 Resource Configuration${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "%-25s: %s\n" "API Memory Limit" "${RESOURCES[api_memory_limit]}"
    printf "%-25s: %s\n" "API CPU Limit" "${RESOURCES[api_cpu_limit]}"
    printf "%-25s: %s\n" "DB Memory Limit" "${RESOURCES[db_memory_limit]}"
    printf "%-25s: %s\n" "Max File Size" "${CONFIG[max_file_size]}"
    printf "%-25s: %s\n" "Max Concurrent Jobs" "${CONFIG[max_concurrent_jobs]}"
    echo ""
    
    echo -e "${CYAN}⚙️ Features${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "%-25s: %s\n" "Storage Type" "${CONFIG[storage_type]}"
    printf "%-25s: %s\n" "Enable Backups" "${CONFIG[enable_backups]:-false}"
    printf "%-25s: %s\n" "Enable Prometheus" "${CONFIG[enable_prometheus]}"
    printf "%-25s: %s\n" "Enable Grafana" "${CONFIG[enable_grafana]}"
    printf "%-25s: %s\n" "Download VMAF Models" "${CONFIG[download_vmaf_models]}"
    echo ""
    
    echo -e "${YELLOW}⚠️ Important Notes:${NC}"
    echo "• All secrets will be stored in environment files"
    echo "• Database and application data will be persistent"
    echo "• Configuration can be modified after installation"
    if [[ "$DEPLOYMENT_MODE" == "production" ]]; then
        echo "• Production mode includes security hardening"
        echo "• SSL certificates will be configured"
    fi
    echo ""
    
    prompt "📋 Please review the configuration above."
    read -p "Proceed with installation? (Y/n): " confirm_install
    
    if [[ ! "${confirm_install:-Y}" =~ ^[Yy]$ ]]; then
        warn "Installation cancelled by user."
        exit 0
    fi
}

# =============================================================================
# Environment File Generation
# =============================================================================

generate_environment_files() {
    log "Generating environment configuration files..."
    
    local env_file="$INSTALL_DIR/.env"
    local env_prod_file="$INSTALL_DIR/.env.production"
    local env_dev_file="$INSTALL_DIR/.env.development"
    
    # Base environment file
    cat > "$env_file" << EOF
# FFprobe API Configuration
# Generated by installer on $(date)

# Deployment Configuration
ENVIRONMENT=${CONFIG[environment]}
DEPLOYMENT_MODE=$DEPLOYMENT_MODE
LOG_LEVEL=${CONFIG[log_level]}

# Network Configuration
API_PORT=${CONFIG[api_port]}
EOF

    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        cat >> "$env_file" << EOF
DOMAIN_NAME=${CONFIG[domain_name]}
HTTP_PORT=${CONFIG[http_port]}
HTTPS_PORT=${CONFIG[https_port]}
EOF
    fi

    # Security Configuration
    cat >> "$env_file" << EOF

# Security Configuration
API_KEY=${SECRETS[api_key]}
JWT_SECRET=${SECRETS[jwt_secret]}
ENABLE_AUTH=${CONFIG[enable_auth]}
ENABLE_RATE_LIMIT=${CONFIG[enable_rate_limit]}
ENABLE_CSRF=${CONFIG[enable_csrf]:-false}

# Database Configuration
POSTGRES_PASSWORD=${SECRETS[postgres_password]}
POSTGRES_USER=ffprobe
POSTGRES_DB=ffprobe_api
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Redis Configuration
REDIS_PASSWORD=${SECRETS[redis_password]}
REDIS_HOST=redis
REDIS_PORT=6379

# File Storage Configuration
DATA_PATH=${CONFIG[data_path]}
BACKUP_PATH=${CONFIG[backup_path]}
MAX_FILE_SIZE=${CONFIG[max_file_size]}
MAX_CONCURRENT_JOBS=${CONFIG[max_concurrent_jobs]}

# AI/LLM Configuration
ENABLE_LOCAL_LLM=${CONFIG[enable_local_llm]:-false}
OLLAMA_MODEL=${CONFIG[ollama_model]:-mistral:7b}
OLLAMA_URL=http://ollama:11434
OLLAMA_MAX_LOADED_MODELS=${CONFIG[ollama_parallel]:-4}
OLLAMA_KEEP_ALIVE=24h

# Monitoring Configuration
GRAFANA_PASSWORD=${SECRETS[grafana_password]}
GRAFANA_USER=${CONFIG[grafana_user]:-admin}
ENABLE_PROMETHEUS=${CONFIG[enable_prometheus]:-true}
ENABLE_GRAFANA=${CONFIG[enable_grafana]:-true}
EOF

    # OpenRouter configuration (if enabled)
    if [[ "${CONFIG[enable_openrouter]}" == "true" ]]; then
        cat >> "$env_file" << EOF

# OpenRouter Cloud LLM Configuration
OPENROUTER_API_KEY=${SECRETS[openrouter_api_key]}
OPENROUTER_MODEL=${CONFIG[openrouter_model]:-anthropic/claude-3-haiku}
OPENROUTER_TIMEOUT=${CONFIG[openrouter_timeout]:-120}
OPENROUTER_MAX_TOKENS=${CONFIG[openrouter_max_tokens]:-2000}
EOF
    fi

    # Rate limiting (if enabled)
    if [[ "${CONFIG[enable_rate_limit]}" == "true" ]]; then
        cat >> "$env_file" << EOF

# Rate Limiting Configuration
RATE_LIMIT_PER_MINUTE=${CONFIG[rate_limit_per_minute]}
RATE_LIMIT_PER_HOUR=${CONFIG[rate_limit_per_hour]}
RATE_LIMIT_PER_DAY=${CONFIG[rate_limit_per_day]}
EOF
    fi

    # Cloud storage configuration
    if [[ "${CONFIG[storage_type]}" != "local" ]]; then
        case "${CONFIG[storage_type]}" in
            "s3")
                cat >> "$env_file" << EOF

# AWS S3 Configuration
AWS_ACCESS_KEY_ID=${SECRETS[aws_access_key_id]}
AWS_SECRET_ACCESS_KEY=${SECRETS[aws_secret_access_key]}
S3_BUCKET=${CONFIG[s3_bucket]}
AWS_REGION=${CONFIG[aws_region]}
EOF
                ;;
            "gcs")
                cat >> "$env_file" << EOF

# Google Cloud Storage Configuration
GCS_BUCKET=${CONFIG[gcs_bucket]}
GCP_PROJECT_ID=${CONFIG[gcp_project_id]}
GOOGLE_APPLICATION_CREDENTIALS=${CONFIG[gcp_key_file]}
EOF
                ;;
            "azure")
                cat >> "$env_file" << EOF

# Azure Blob Storage Configuration
AZURE_STORAGE_ACCOUNT=${CONFIG[azure_storage_account]}
AZURE_STORAGE_KEY=${SECRETS[azure_storage_key]}
AZURE_CONTAINER=${CONFIG[azure_container]}
EOF
                ;;
        esac
    fi

    # Backup configuration (if enabled)
    if [[ "${CONFIG[enable_backups]}" == "true" ]]; then
        cat >> "$env_file" << EOF

# Backup Configuration
ENABLE_BACKUPS=true
BACKUP_RETENTION_DAYS=${CONFIG[backup_retention]}
BACKUP_SCHEDULE="${CONFIG[backup_schedule]}"
EOF
    fi

    # Generate production-specific overrides
    if [[ "$DEPLOYMENT_MODE" == "production" ]]; then
        cat > "$env_prod_file" << EOF
# Production Environment Overrides
LOG_LEVEL=warn
ENABLE_AUTH=true
ENABLE_RATE_LIMIT=true
ENABLE_CSRF=true
ENABLE_DEBUG=false
EOF
    fi

    # Generate development-specific overrides
    cat > "$env_dev_file" << EOF
# Development Environment Overrides
LOG_LEVEL=debug
ENABLE_AUTH=false
ENABLE_RATE_LIMIT=false
ENABLE_CSRF=false
ENABLE_DEBUG=true

# Development passwords (override in production)
POSTGRES_PASSWORD=dev_password_change_this
REDIS_PASSWORD=dev_redis_pass
API_KEY=dev_api_key_change_this_minimum_32_chars
JWT_SECRET=dev_jwt_secret_change_this_minimum_32_chars
GRAFANA_PASSWORD=admin_change_this
EOF

    # Set secure permissions
    chmod 600 "$env_file" "$env_prod_file" "$env_dev_file"
    
    success "Environment files generated"
    info "Main config: $env_file"
    info "Production overrides: $env_prod_file"
    info "Development overrides: $env_dev_file"
}

# =============================================================================
# Docker Compose Configuration
# =============================================================================

setup_docker_compose() {
    log "Setting up Docker Compose configuration..."
    
    # Copy project files to installation directory
    if [[ "$PWD" != "$INSTALL_DIR" ]]; then
        info "Copying project files to $INSTALL_DIR..."
        cp -r . "$INSTALL_DIR/"
    fi
    
    cd "$INSTALL_DIR"
    
    # Update resource limits in compose files
    if command_exists yq; then
        info "Updating resource limits in Docker Compose files..."
        
        # Update API service resources
        yq eval ".services.ffprobe-api.deploy.resources.limits.memory = \"${RESOURCES[api_memory_limit]}\"" -i compose.yml
        yq eval ".services.ffprobe-api.deploy.resources.limits.cpus = \"${RESOURCES[api_cpu_limit]}\"" -i compose.yml
        yq eval ".services.ffprobe-api.deploy.resources.reservations.memory = \"${RESOURCES[api_memory_reservation]}\"" -i compose.yml
        yq eval ".services.ffprobe-api.deploy.resources.reservations.cpus = \"${RESOURCES[api_cpu_reservation]}\"" -i compose.yml
        
        # Update database resources
        yq eval ".services.postgres.deploy.resources.limits.memory = \"${RESOURCES[db_memory_limit]}\"" -i compose.yml
        yq eval ".services.postgres.deploy.resources.limits.cpus = \"${RESOURCES[db_cpu_limit]}\"" -i compose.yml
        yq eval ".services.postgres.deploy.resources.reservations.memory = \"${RESOURCES[db_memory_reservation]}\"" -i compose.yml
        yq eval ".services.postgres.deploy.resources.reservations.cpus = \"${RESOURCES[db_cpu_reservation]}\"" -i compose.yml
        
        # Update Redis resources
        yq eval ".services.redis.deploy.resources.limits.memory = \"${RESOURCES[redis_memory_limit]}\"" -i compose.yml
        yq eval ".services.redis.deploy.resources.limits.cpus = \"${RESOURCES[redis_cpu_limit]}\"" -i compose.yml
    else
        warn "yq not found - Docker Compose resource limits will use defaults"
    fi
    
    success "Docker Compose configuration ready"
}

# =============================================================================
# SSL Certificate Setup
# =============================================================================

setup_ssl_certificates() {
    if [[ "$ENABLE_SSL" == "disabled" || "$DEPLOYMENT_MODE" == "development" ]]; then
        return 0
    fi
    
    log "Setting up SSL certificates..."
    
    local ssl_dir="$INSTALL_DIR/docker/ssl"
    mkdir -p "$ssl_dir"
    
    case "$ENABLE_SSL" in
        "letsencrypt")
            info "Setting up Let's Encrypt certificates..."
            
            # Check if certbot is installed
            if ! command_exists certbot; then
                info "Installing certbot..."
                sudo apt-get update && sudo apt-get install -y certbot
            fi
            
            # Generate certificates
            info "Generating Let's Encrypt certificates for $DOMAIN_NAME..."
            sudo certbot certonly --standalone \
                --email "$EMAIL" \
                --agree-tos \
                --no-eff-email \
                -d "$DOMAIN_NAME"
            
            # Copy certificates to SSL directory
            sudo cp "/etc/letsencrypt/live/$DOMAIN_NAME/fullchain.pem" "$ssl_dir/cert.crt"
            sudo cp "/etc/letsencrypt/live/$DOMAIN_NAME/privkey.pem" "$ssl_dir/key.key"
            sudo chown $USER:$USER "$ssl_dir"/*
            ;;
            
        "custom")
            info "Setting up custom SSL certificates..."
            
            if [[ -f "${CONFIG[ssl_cert_path]}" && -f "${CONFIG[ssl_key_path]}" ]]; then
                cp "${CONFIG[ssl_cert_path]}" "$ssl_dir/cert.crt"
                cp "${CONFIG[ssl_key_path]}" "$ssl_dir/key.key"
            else
                error "Custom SSL certificate files not found"
            fi
            ;;
    esac
    
    # Set proper permissions
    chmod 644 "$ssl_dir/cert.crt"
    chmod 600 "$ssl_dir/key.key"
    
    success "SSL certificates configured"
}

# =============================================================================
# Application Deployment
# =============================================================================

deploy_application() {
    log "Deploying FFprobe API..."
    
    cd "$INSTALL_DIR"
    
    # Create data directories
    info "Creating data directories..."
    mkdir -p "${CONFIG[data_path]}"/{postgres,redis,uploads,reports,models,logs,temp,cache,prometheus,grafana,backup}
    
    # Set proper permissions
    sudo chown -R $USER:$USER "${CONFIG[data_path]}"
    chmod -R 755 "${CONFIG[data_path]}"
    
    # Download VMAF models if requested
    if [[ "${CONFIG[download_vmaf_models]}" =~ ^[Yy]$ ]]; then
        info "Downloading additional VMAF models..."
        mkdir -p "${CONFIG[data_path]}/models"
        
        local models_dir="${CONFIG[data_path]}/models"
        cd "$models_dir"
        
        # Download additional VMAF models
        wget -q -O vmaf_v0.6.1.json "https://github.com/Netflix/vmaf/raw/master/model/vmaf_v0.6.1.json"
        wget -q -O vmaf_4k_v0.6.1.json "https://github.com/Netflix/vmaf/raw/master/model/vmaf_4k_v0.6.1.json"
        wget -q -O vmaf_b_v0.6.3.json "https://github.com/Netflix/vmaf/raw/master/model/vmaf_b_v0.6.3.json"
        
        cd "$INSTALL_DIR"
        success "VMAF models downloaded"
    fi
    
    # Build and start services based on deployment mode
    case "$DEPLOYMENT_MODE" in
        "development")
            info "Starting development environment..."
            docker compose -f compose.yml -f compose.dev.yml build
            docker compose -f compose.yml -f compose.dev.yml up -d
            ;;
        "staging"|"production")
            info "Starting $DEPLOYMENT_MODE environment..."
            docker compose -f compose.yml -f compose.prod.yml build
            docker compose -f compose.yml -f compose.prod.yml up -d
            ;;
        "swarm")
            info "Initializing Docker Swarm..."
            docker swarm init 2>/dev/null || true
            docker stack deploy -c compose.yml -c compose.prod.yml ffprobe-api
            ;;
        "kubernetes")
            info "Generating Kubernetes manifests..."
            # This would generate k8s manifests - not implemented in this script
            warn "Kubernetes deployment requires manual manifest generation"
            ;;
    esac
    
    success "Application deployment completed"
}

# =============================================================================
# Post-Installation Verification
# =============================================================================

verify_installation() {
    log "Verifying installation..."
    
    # Wait for services to start
    info "Waiting for services to start..."
    sleep 30
    
    # Check service health
    local api_port="${CONFIG[api_port]}"
    local max_attempts=10
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -f -s "http://localhost:$api_port/health" >/dev/null 2>&1; then
            success "✅ FFprobe API is responding"
            break
        else
            if [[ $attempt -eq $max_attempts ]]; then
                error "❌ FFprobe API is not responding after $max_attempts attempts"
            fi
            info "Attempt $attempt/$max_attempts: Waiting for API to respond..."
            sleep 10
            ((attempt++))
        fi
    done
    
    # Check database connection
    if docker compose exec postgres pg_isready -U ffprobe -d ffprobe_api >/dev/null 2>&1; then
        success "✅ Database is ready"
    else
        warn "⚠️  Database connection issues detected"
    fi
    
    # Check Redis connection
    if docker compose exec redis redis-cli ping >/dev/null 2>&1; then
        success "✅ Redis is ready"
    else
        warn "⚠️  Redis connection issues detected"
    fi
    
    # Test API functionality
    info "Testing API functionality..."
    local api_key="${SECRETS[api_key]}"
    
    if curl -f -s -H "X-API-Key: $api_key" "http://localhost:$api_port/api/v1/system/version" >/dev/null 2>&1; then
        success "✅ API authentication working"
    else
        warn "⚠️  API authentication test failed"
    fi
    
    success "Installation verification completed"
}

# =============================================================================
# Backup Setup
# =============================================================================

setup_backup_system() {
    if [[ "${CONFIG[enable_backups]}" != "true" ]]; then
        return 0
    fi
    
    log "Setting up backup system..."
    
    local backup_script="$INSTALL_DIR/scripts/backup-automated.sh"
    
    # Create backup script
    cat > "$backup_script" << 'EOF'
#!/bin/bash

# Automated backup script for FFprobe API
# Generated by installer

set -euo pipefail

BACKUP_DIR="${BACKUP_PATH}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup database
echo "Backing up database..."
docker compose exec -T postgres pg_dump -U ffprobe ffprobe_api > "$BACKUP_DIR/postgres_$TIMESTAMP.sql"

# Backup Redis data
echo "Backing up Redis..."
docker compose exec -T redis redis-cli BGSAVE
sleep 5
docker compose cp redis:/data/dump.rdb "$BACKUP_DIR/redis_$TIMESTAMP.rdb"

# Backup uploaded files
echo "Backing up uploads..."
tar -czf "$BACKUP_DIR/uploads_$TIMESTAMP.tar.gz" -C "${DATA_PATH}" uploads/

# Backup configuration
echo "Backing up configuration..."
tar -czf "$BACKUP_DIR/config_$TIMESTAMP.tar.gz" .env* compose*.yml docker/

# Cleanup old backups
echo "Cleaning up backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -type f -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $TIMESTAMP"
EOF

    chmod +x "$backup_script"
    
    # Add to crontab
    local cron_schedule="${CONFIG[backup_schedule]}"
    (crontab -l 2>/dev/null; echo "$cron_schedule cd $INSTALL_DIR && $backup_script") | crontab -
    
    success "Backup system configured"
    info "Backup schedule: $cron_schedule"
    info "Backup retention: ${CONFIG[backup_retention]} days"
}

# =============================================================================
# Installation Summary
# =============================================================================

show_installation_summary() {
    clear
    log "Installation Complete! 🎉"
    echo ""
    
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}                    🎬 FFprobe API Successfully Deployed!                     ${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    # Access Information
    echo -e "${CYAN}🌐 Access Information${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [[ "$DEPLOYMENT_MODE" == "development" ]]; then
        printf "%-25s: %s\n" "API Endpoint" "http://localhost:${CONFIG[api_port]}"
        printf "%-25s: %s\n" "Health Check" "http://localhost:${CONFIG[api_port]}/health"
        printf "%-25s: %s\n" "Adminer (DB)" "http://localhost:8090"
        printf "%-25s: %s\n" "Redis Commander" "http://localhost:8091"
    else
        if [[ "$ENABLE_SSL" != "disabled" ]]; then
            printf "%-25s: %s\n" "API Endpoint" "https://${CONFIG[domain_name]}"
            printf "%-25s: %s\n" "Health Check" "https://${CONFIG[domain_name]}/health"
        else
            printf "%-25s: %s\n" "API Endpoint" "http://${CONFIG[domain_name]}"
            printf "%-25s: %s\n" "Health Check" "http://${CONFIG[domain_name]}/health"
        fi
    fi
    
    printf "%-25s: %s\n" "Grafana Dashboard" "http://localhost:3000"
    printf "%-25s: %s\n" "Prometheus Metrics" "http://localhost:9090"
    echo ""
    
    # Authentication Information
    echo -e "${CYAN}🔐 Authentication${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "%-25s: %s\n" "API Key" "${SECRETS[api_key]}"
    if [[ "${CONFIG[enable_grafana]}" =~ ^[Yy]$ ]]; then
        printf "%-25s: %s\n" "Grafana Username" "${CONFIG[grafana_user]}"
        printf "%-25s: %s\n" "Grafana Password" "${SECRETS[grafana_password]}"
    fi
    echo ""
    
    # File Locations
    echo -e "${CYAN}📁 Important Files${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    printf "%-25s: %s\n" "Installation Directory" "${CONFIG[install_dir]}"
    printf "%-25s: %s\n" "Configuration File" "${CONFIG[install_dir]}/.env"
    printf "%-25s: %s\n" "Data Directory" "${CONFIG[data_path]}"
    printf "%-25s: %s\n" "Backup Directory" "${CONFIG[backup_path]}"
    printf "%-25s: %s\n" "Log Files" "${CONFIG[data_path]}/logs"
    echo ""
    
    # Quick Start Commands
    echo -e "${CYAN}🚀 Quick Start Commands${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "cd ${CONFIG[install_dir]}"
    echo ""
    echo "# Check service status"
    echo "docker compose ps"
    echo ""
    echo "# View logs"
    echo "docker compose logs -f ffprobe-api"
    echo ""
    echo "# Test API"
    echo "curl -H \"X-API-Key: ${SECRETS[api_key]}\" http://localhost:${CONFIG[api_port]}/health"
    echo ""
    echo "# Stop services"
    echo "docker compose down"
    echo ""
    echo "# Start services"
    case "$DEPLOYMENT_MODE" in
        "development")
            echo "docker compose -f compose.yml -f compose.dev.yml up -d"
            ;;
        "staging"|"production")
            echo "docker compose -f compose.yml -f compose.prod.yml up -d"
            ;;
    esac
    echo ""
    
    # Management Commands
    echo -e "${CYAN}🛠️ Management Commands${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "# Backup data"
    echo "./scripts/backup.sh"
    echo ""
    echo "# Deploy updates"
    echo "./scripts/deploy.sh deploy $DEPLOYMENT_MODE latest"
    echo ""
    echo "# Scale API instances (production)"
    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        echo "docker compose -f compose.yml -f compose.prod.yml up -d --scale ffprobe-api=3"
    else
        echo "# (Not available in development mode)"
    fi
    echo ""
    
    # Security Reminders
    if [[ "$DEPLOYMENT_MODE" != "development" ]]; then
        echo -e "${YELLOW}⚠️ Security Reminders${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "• Keep your API key and other secrets secure"
        echo "• Regularly update the application and dependencies"
        echo "• Monitor logs for security events"
        echo "• Set up automated backups"
        if [[ "$ENABLE_SSL" == "letsencrypt" ]]; then
            echo "• Let's Encrypt certificates will auto-renew"
        fi
        echo "• Review firewall and network security settings"
        echo ""
    fi
    
    # Support Information
    echo -e "${CYAN}🆘 Support & Documentation${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📖 Documentation: ${CONFIG[install_dir]}/docs/"
    echo "🔧 Configuration: ${CONFIG[install_dir]}/docs/deployment/configuration.md"
    echo "🎯 API Examples: ${CONFIG[install_dir]}/docs/tutorials/api_usage.md"
    echo "🔒 Security Guide: ${CONFIG[install_dir]}/SECURITY_AUDIT_REPORT.md"
    echo "🤝 Contributing: ${CONFIG[install_dir]}/CONTRIBUTING.md"
    echo ""
    
    success "🎉 Enjoy your new FFprobe API installation!"
    echo ""
}

# =============================================================================
# Main Installation Flow
# =============================================================================

main() {
    # Trap to cleanup on error
    trap 'error "Installation failed. Check the logs above for details."' ERR
    
    print_banner
    
    log "Starting FFprobe API Interactive Installer v$INSTALLER_VERSION"
    
    # Installation steps
    check_requirements
    select_deployment_mode
    configure_paths
    configure_network
    configure_security
    configure_resources
    configure_llm
    configure_storage_enhanced
    configure_advanced
    
    # Show summary and confirm
    show_configuration_summary
    
    # Generate configuration and deploy
    generate_environment_files
    setup_docker_compose
    setup_ssl_certificates
    deploy_application
    setup_backup_system
    
    # Verify and complete
    verify_installation
    show_installation_summary
    
    log "Installation completed successfully! 🎉"
}

# =============================================================================
# Script Execution
# =============================================================================

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi