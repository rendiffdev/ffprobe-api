#!/bin/bash

# FFprobe API - Universal Automated Setup Script
# Works on Linux, macOS, and Windows (via WSL/Git Bash)
# One command to deploy everything with latest stable versions

set -e

# Colors for better UX
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/rendiffdev/ffprobe-api.git"
INSTALL_DIR="${HOME}/ffprobe-api"
DATA_DIR="${HOME}/ffprobe-api-data"
COMPOSE_VERSION="latest"
MIN_DOCKER_VERSION="24.0.0"
# Default minimum requirements - will be updated based on deployment mode
MIN_RAM_GB=4
MIN_DISK_GB=10
MIN_CPU_CORES=2

# Deployment modes
DEPLOYMENT_MODES=("quick" "production" "development" "custom")

# ASCII Art Banner
show_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
    ╔══════════════════════════════════════════════════╗
    ║                                                  ║
    ║     FFprobe API - Automated Setup Wizard        ║
    ║                                                  ║
    ║     🚀 One-Command Deployment                   ║
    ║     🐳 Fully Dockerized                         ║
    ║     🤖 AI-Powered Analysis                      ║
    ║     ⚡ Latest Stable Versions                   ║
    ║                                                  ║
    ╚══════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ${NC}  $1"
}

log_success() {
    echo -e "${GREEN}✓${NC}  $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC}  $1"
}

log_error() {
    echo -e "${RED}✗${NC}  $1"
}

log_step() {
    echo -e "\n${MAGENTA}▶${NC}  ${WHITE}$1${NC}"
}

# Progress bar
show_progress() {
    local current=$1
    local total=$2
    local width=50
    local percentage=$((current * 100 / total))
    local filled=$((width * current / total))
    
    printf "\r["
    printf "%${filled}s" | tr ' ' '='
    printf "%$((width - filled))s" | tr ' ' '>'
    printf "] %3d%%" $percentage
    
    if [ $current -eq $total ]; then
        echo
    fi
}

# Spinner for long operations
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        sleep $delay
        printf "\b\b\b\b\b\b"
    done
    printf "    \b\b\b\b"
}

# Set deployment-specific requirements
set_deployment_requirements() {
    local mode=$1
    
    case $mode in
        quick)
            MIN_RAM_GB=3
            MIN_DISK_GB=8
            MIN_CPU_CORES=2
            REQUIRED_PORTS=(8080 5432 6379 11434)
            log_info "Quick Start mode requirements: ${MIN_RAM_GB}GB RAM, ${MIN_DISK_GB}GB disk, ${MIN_CPU_CORES} CPU cores"
            ;;
        production)
            MIN_RAM_GB=8
            MIN_DISK_GB=20
            MIN_CPU_CORES=4
            REQUIRED_PORTS=(8080 5432 6379 11434 80 443 3000 9090)
            log_info "Production mode requirements: ${MIN_RAM_GB}GB RAM, ${MIN_DISK_GB}GB disk, ${MIN_CPU_CORES} CPU cores"
            ;;
        development)
            MIN_RAM_GB=4
            MIN_DISK_GB=15
            MIN_CPU_CORES=2
            REQUIRED_PORTS=(8080 5432 6379 11434 2345)
            log_info "Development mode requirements: ${MIN_RAM_GB}GB RAM, ${MIN_DISK_GB}GB disk, ${MIN_CPU_CORES} CPU cores"
            ;;
        minimal)
            MIN_RAM_GB=2
            MIN_DISK_GB=6
            MIN_CPU_CORES=1
            REQUIRED_PORTS=(8080 5432 6379 11434)
            log_info "Minimal mode requirements: ${MIN_RAM_GB}GB RAM, ${MIN_DISK_GB}GB disk, ${MIN_CPU_CORES} CPU cores"
            ;;
        *)
            # Default/unknown mode
            MIN_RAM_GB=4
            MIN_DISK_GB=10
            MIN_CPU_CORES=2
            REQUIRED_PORTS=(8080 5432 6379 11434)
            ;;
    esac
}

# OS Detection
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS="Linux";;
        Darwin*)    OS="macOS";;
        MINGW*|MSYS*|CYGWIN*)     OS="Windows";;
        *)          OS="Unknown";;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64)     ARCH="amd64";;
        aarch64|arm64)    ARCH="arm64";;
        *)          ARCH="Unknown";;
    esac
    
    log_info "Detected OS: ${CYAN}$OS${NC} (${ARCH})"
}

# Check system requirements based on deployment mode
check_requirements() {
    local mode=${1:-"quick"}
    
    # Set requirements based on deployment mode
    set_deployment_requirements "$mode"
    
    log_step "Checking System Requirements for $mode Mode"
    
    local checks_passed=0
    local checks_total=5
    local warnings=0
    
    # Check RAM
    echo -n "  Checking RAM... "
    local ram_gb=0
    if [ "$OS" = "macOS" ]; then
        ram_gb=$(($(sysctl -n hw.memsize) / 1024 / 1024 / 1024))
    else
        ram_gb=$(($(grep MemTotal /proc/meminfo | awk '{print $2}') / 1024 / 1024))
    fi
    
    if [ $ram_gb -ge $MIN_RAM_GB ]; then
        echo -e "${GREEN}✓${NC} ${ram_gb}GB available"
        ((checks_passed++))
    else
        echo -e "${RED}✗${NC} Only ${ram_gb}GB available (need ${MIN_RAM_GB}GB)"
    fi
    show_progress $((checks_passed)) $checks_total
    
    # Check disk space
    echo -n "  Checking disk space... "
    local disk_gb=$(df -BG . | awk 'NR==2 {print int($4)}')
    if [ $disk_gb -ge $MIN_DISK_GB ]; then
        echo -e "${GREEN}✓${NC} ${disk_gb}GB available"
        ((checks_passed++))
    else
        echo -e "${RED}✗${NC} Only ${disk_gb}GB available (need ${MIN_DISK_GB}GB)"
    fi
    show_progress $((checks_passed)) $checks_total
    
    # Check internet connection
    echo -n "  Checking internet connection... "
    if curl -s --head https://github.com > /dev/null; then
        echo -e "${GREEN}✓${NC} Connected"
        ((checks_passed++))
    else
        echo -e "${RED}✗${NC} No connection"
    fi
    show_progress $((checks_passed)) $checks_total
    
    # Check ports
    echo -n "  Checking required ports... "
    local ports_available=true
    local port_conflicts=()
    for port in "${REQUIRED_PORTS[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            port_conflicts+=("$port")
            ports_available=false
        fi
    done
    if $ports_available; then
        echo -e "${GREEN}✓${NC} All required ports available (${REQUIRED_PORTS[*]})"
        ((checks_passed++))
    else
        echo -e "${YELLOW}⚠${NC} Ports in use: ${port_conflicts[*]}"
        ((warnings++))
    fi
    show_progress $((checks_passed)) $checks_total
    
    # Check CPU cores
    echo -n "  Checking CPU cores... "
    local cpu_cores=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 1)
    if [ $cpu_cores -ge $MIN_CPU_CORES ]; then
        echo -e "${GREEN}✓${NC} ${cpu_cores} cores available (need $MIN_CPU_CORES)"
        ((checks_passed++))
    else
        echo -e "${YELLOW}⚠${NC} Only ${cpu_cores} core(s) available (need $MIN_CPU_CORES)"
        ((warnings++))
    fi
    show_progress $checks_total $checks_total
    
    # Enhanced requirements evaluation
    echo
    if [ $checks_passed -eq $checks_total ]; then
        log_success "✅ All system requirements met for $mode mode ($checks_passed/$checks_total)"
    elif [ $checks_passed -ge 3 ]; then
        if [ $warnings -gt 0 ]; then
            log_warning "⚠️  System requirements mostly met with $warnings warning(s) ($checks_passed/$checks_total)"
            echo "   This may impact performance but should still work."
        else
            log_success "System requirements check passed ($checks_passed/$checks_total)"
        fi
    else
        log_error "❌ Critical system requirements not met for $mode mode!"
        echo
        echo "   Requirements Summary:"
        echo "   • RAM: ${MIN_RAM_GB}GB minimum"
        echo "   • Disk: ${MIN_DISK_GB}GB free space"
        echo "   • CPU: ${MIN_CPU_CORES}+ cores"
        echo "   • Ports: ${REQUIRED_PORTS[*]}"
        echo "   • Internet connection"
        echo
        
        if [ "$mode" = "production" ]; then
            echo "   💡 Consider switching to 'quick' or 'development' mode for lower requirements."
            echo
            echo "   Continue with production mode anyway? (y/N)"
        else
            echo "   Continue anyway? (y/N)"
        fi
        
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            echo
            log_info "Installation cancelled. You can try:"
            echo "   • Free up system resources"
            echo "   • Use 'quick' mode: curl -fsSL setup.sh | bash -s -- --quick"
            echo "   • Use minimal setup with lower requirements"
            exit 1
        fi
        
        log_warning "⚠️  Proceeding with insufficient resources - performance may be degraded"
    fi
}

# Validate Docker can run the required containers
validate_docker_capabilities() {
    local mode=$1
    
    log_step "Validating Docker Capabilities for $mode Mode"
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker not installed - cannot validate capabilities"
        return 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker daemon not running - cannot validate capabilities"
        return 1
    fi
    
    # Check Docker memory limit
    echo -n "  Checking Docker memory limit... "
    local docker_memory=$(docker system info --format '{{.MemTotal}}' 2>/dev/null || echo 0)
    local docker_memory_gb=$((docker_memory / 1024 / 1024 / 1024))
    
    if [ $docker_memory_gb -ge $MIN_RAM_GB ]; then
        echo -e "${GREEN}✓${NC} ${docker_memory_gb}GB available to Docker"
    else
        echo -e "${YELLOW}⚠${NC} Only ${docker_memory_gb}GB available to Docker (need ${MIN_RAM_GB}GB)"
        if [ "$mode" = "production" ]; then
            log_warning "Docker memory limit too low for production mode"
            echo "   💡 Increase Docker Desktop memory allocation in settings"
        fi
    fi
    
    # Test basic container functionality
    echo -n "  Testing container functionality... "
    if docker run --rm hello-world &> /dev/null; then
        echo -e "${GREEN}✓${NC} Docker can run containers"
    else
        echo -e "${RED}✗${NC} Docker cannot run containers"
        log_error "Docker installation appears to be broken"
        return 1
    fi
    
    log_success "Docker validation complete"
}

# Install Docker if not present
install_docker() {
    log_step "Checking Docker Installation"
    
    if command -v docker &> /dev/null; then
        local docker_version=$(docker --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        log_success "Docker found: version $docker_version"
        
        # Check Docker daemon
        if ! docker info &> /dev/null; then
            log_warning "Docker daemon not running. Starting..."
            if [ "$OS" = "Linux" ]; then
                sudo systemctl start docker
            elif [ "$OS" = "macOS" ]; then
                open -a Docker
                log_info "Waiting for Docker Desktop to start..."
                sleep 10
            fi
        fi
    else
        log_warning "Docker not found. Installing..."
        
        case "$OS" in
            Linux)
                log_info "Installing Docker on Linux..."
                curl -fsSL https://get.docker.com -o get-docker.sh
                sudo sh get-docker.sh
                sudo usermod -aG docker $USER
                rm get-docker.sh
                log_success "Docker installed. Please log out and back in for group changes."
                ;;
            macOS)
                log_info "Please install Docker Desktop from:"
                echo "  https://www.docker.com/products/docker-desktop"
                exit 1
                ;;
            Windows)
                log_info "Please install Docker Desktop from:"
                echo "  https://www.docker.com/products/docker-desktop"
                exit 1
                ;;
        esac
    fi
    
    # Install Docker Compose if needed
    if ! docker compose version &> /dev/null; then
        log_info "Installing Docker Compose plugin..."
        if [ "$OS" = "Linux" ]; then
            sudo apt-get update && sudo apt-get install -y docker-compose-plugin
        fi
    fi
    
    log_success "Docker is ready"
}

# Quick setup wizard
setup_wizard() {
    log_step "Configuration Wizard"
    
    echo -e "\n${CYAN}Choose deployment mode:${NC}"
    echo "  1) ${GREEN}Quick Start${NC} - Minimal config, get running in 2 minutes (3GB RAM, 8GB disk)"
    echo "  2) ${CYAN}Minimal${NC} - Ultra-lightweight for testing (2GB RAM, 6GB disk)"
    echo "  3) ${YELLOW}Production${NC} - Full features, security, monitoring (8GB RAM, 20GB disk)"
    echo "  4) ${BLUE}Development${NC} - Hot reload, debug tools (4GB RAM, 15GB disk)"
    echo "  5) ${MAGENTA}Custom${NC} - Configure everything"
    
    read -p "Select mode [1-5] (default: 1): " mode_choice
    mode_choice=${mode_choice:-1}
    
    case $mode_choice in
        1)
            DEPLOYMENT_MODE="quick"
            log_info "Quick Start mode selected"
            ;;
        2)
            DEPLOYMENT_MODE="minimal"
            log_info "Minimal mode selected"
            ;;
        3)
            DEPLOYMENT_MODE="production"
            log_info "Production mode selected"
            ;;
        4)
            DEPLOYMENT_MODE="development"
            log_info "Development mode selected"
            ;;
        5)
            DEPLOYMENT_MODE="custom"
            log_info "Custom mode selected"
            ;;
        *)
            DEPLOYMENT_MODE="quick"
            log_info "Invalid choice, using Quick Start mode"
            ;;
    esac
}

# Generate secure configuration
generate_config() {
    log_step "Generating Configuration"
    
    # Generate secure keys
    API_KEY="ffprobe_$(openssl rand -hex 32)"
    JWT_SECRET="$(openssl rand -hex 32)"
    POSTGRES_PASSWORD="$(openssl rand -hex 16)"
    REDIS_PASSWORD="$(openssl rand -hex 16)"
    GRAFANA_PASSWORD="$(openssl rand -hex 12)"
    
    # Create .env file based on deployment mode
    cat > "$INSTALL_DIR/.env" << EOF
# FFprobe API Configuration
# Generated: $(date)
# Mode: $DEPLOYMENT_MODE

# === CORE SETTINGS ===
GO_ENV=$([ "$DEPLOYMENT_MODE" = "production" ] && echo "production" || echo "development")
API_PORT=8080
HOST=0.0.0.0

# === AUTHENTICATION ===
ENABLE_AUTH=$([ "$DEPLOYMENT_MODE" = "quick" ] && echo "false" || echo "true")
API_KEY=$API_KEY
JWT_SECRET=$JWT_SECRET

# === DATABASE ===
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=ffprobe_api
POSTGRES_USER=ffprobe
POSTGRES_PASSWORD=$POSTGRES_PASSWORD

# === REDIS CACHE ===
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=$REDIS_PASSWORD

# === AI/LLM CONFIGURATION ===
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=gemma3:270m
OLLAMA_FALLBACK_MODEL=phi3:mini

# === FFMPEG (Latest from BtbN) ===
FFMPEG_AUTO_UPDATE=$([ "$DEPLOYMENT_MODE" = "production" ] && echo "false" || echo "true")
FFMPEG_ALLOW_MAJOR_UPDATES=false

# === STORAGE ===
UPLOAD_DIR=/app/uploads
REPORTS_DIR=/app/reports
MAX_FILE_SIZE=53687091200  # 50GB

# === MONITORING ===
ENABLE_PROMETHEUS=$([ "$DEPLOYMENT_MODE" = "production" ] && echo "true" || echo "false")
ENABLE_GRAFANA=$([ "$DEPLOYMENT_MODE" = "production" ] && echo "true" || echo "false")
GRAFANA_PASSWORD=$GRAFANA_PASSWORD

# === RATE LIMITING ===
ENABLE_RATE_LIMIT=$([ "$DEPLOYMENT_MODE" = "production" ] && echo "true" || echo "false")
RATE_LIMIT_PER_MINUTE=60
RATE_LIMIT_PER_HOUR=1000

# === PERFORMANCE ===
WORKER_POOL_SIZE=16
PROCESSING_TIMEOUT=300
EOF
    
    log_success "Configuration generated"
    
    # Save credentials for user
    cat > "$INSTALL_DIR/credentials.txt" << EOF
════════════════════════════════════════════════════════════════
FFprobe API Credentials - SAVE THIS FILE!
════════════════════════════════════════════════════════════════

API Endpoint: http://localhost:8080

API Key: $API_KEY

$([ "$DEPLOYMENT_MODE" = "production" ] && echo "
Grafana Dashboard: http://localhost:3000
  Username: admin
  Password: $GRAFANA_PASSWORD
")

PostgreSQL:
  Host: localhost:5432
  Database: ffprobe_api
  Username: ffprobe
  Password: $POSTGRES_PASSWORD

════════════════════════════════════════════════════════════════
EOF
    
    chmod 600 "$INSTALL_DIR/credentials.txt"
}

# Select and prepare Docker Compose file
prepare_compose() {
    log_step "Preparing Docker Compose Configuration"
    
    case $DEPLOYMENT_MODE in
        quick|minimal)
            # Lightweight setup for quick start and minimal deployments
            cat > "$INSTALL_DIR/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.btbn
    ports:
      - "8080:8080"
    environment:
      - GO_ENV=development
    env_file:
      - .env
    volumes:
      - ./uploads:/app/uploads
      - ./reports:/app/reports
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      ollama:
        condition: service_started
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    ports:
      - "11434:11434"
    environment:
      - OLLAMA_MODELS=${OLLAMA_MODEL},${OLLAMA_FALLBACK_MODEL}
    deploy:
      resources:
        limits:
          memory: 4G
    entrypoint: ["/bin/sh", "-c"]
    command: |
      "ollama serve & 
       sleep 10 && 
       ollama pull ${OLLAMA_MODEL} && 
       ollama pull ${OLLAMA_FALLBACK_MODEL} && 
       tail -f /dev/null"

volumes:
  postgres_data:
  redis_data:
  ollama_data:
EOF
            ;;
            
        production)
            # Full production setup with monitoring
            cp "$INSTALL_DIR/compose.production.yml" "$INSTALL_DIR/docker-compose.yml" 2>/dev/null || \
            cat > "$INSTALL_DIR/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.btbn
    ports:
      - "8080:8080"
    environment:
      - GO_ENV=production
    env_file:
      - .env
    volumes:
      - ./uploads:/app/uploads
      - ./reports:/app/reports
      - ./backup:/app/backup
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      ollama:
        condition: service_started
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_INITDB_ARGS: "--encoding=UTF8 --locale=en_US.UTF-8"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backup/postgres:/backup
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    environment:
      - OLLAMA_MODELS=${OLLAMA_MODEL},${OLLAMA_FALLBACK_MODEL}
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '2'
    restart: always
    entrypoint: ["/bin/sh", "-c"]
    command: |
      "ollama serve & 
       sleep 10 && 
       ollama pull ${OLLAMA_MODEL} && 
       ollama pull ${OLLAMA_FALLBACK_MODEL} && 
       tail -f /dev/null"

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - api
    restart: always

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    ports:
      - "9090:9090"
    restart: always

  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
      - GF_INSTALL_PLUGINS=redis-datasource
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
    ports:
      - "3000:3000"
    restart: always

volumes:
  postgres_data:
  redis_data:
  ollama_data:
  prometheus_data:
  grafana_data:
EOF
            ;;
            
        development)
            # Development setup with hot reload
            cat > "$INSTALL_DIR/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.dev
      target: development
    ports:
      - "8080:8080"
      - "2345:2345"  # Delve debugger
    environment:
      - GO_ENV=development
      - DEV_ENABLE_DEBUG=true
      - DEV_DISABLE_AUTH=true
      - DEV_DISABLE_RATE_LIMIT=true
    env_file:
      - .env
    volumes:
      - .:/app
      - /app/vendor
    depends_on:
      - postgres
      - redis
      - ollama
    command: air -c .air.toml

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    ports:
      - "11434:11434"
    environment:
      - OLLAMA_MODELS=${OLLAMA_MODEL}
    entrypoint: ["/bin/sh", "-c"]
    command: |
      "ollama serve & 
       sleep 10 && 
       ollama pull ${OLLAMA_MODEL} && 
       tail -f /dev/null"

volumes:
  postgres_data:
  redis_data:
  ollama_data:
EOF
            ;;
    esac
    
    log_success "Docker Compose configuration ready"
}

# Deploy the stack
deploy_stack() {
    log_step "Deploying FFprobe API Stack"
    
    cd "$INSTALL_DIR"
    
    # Pull latest images
    log_info "Pulling latest Docker images..."
    docker compose pull 2>/dev/null &
    spinner $!
    
    # Build the API image
    log_info "Building FFprobe API image with latest FFmpeg..."
    docker compose build --no-cache api 2>/dev/null &
    spinner $!
    
    # Start services
    log_info "Starting services..."
    docker compose up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to be ready..."
    local max_attempts=60
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            log_success "API is ready!"
            break
        fi
        sleep 2
        attempt=$((attempt + 1))
        show_progress $attempt $max_attempts
    done
    
    if [ $attempt -eq $max_attempts ]; then
        log_error "Services failed to start. Check logs with: docker compose logs"
        exit 1
    fi
}

# Run post-deployment tests
run_tests() {
    log_step "Running Health Checks"
    
    # Test API
    echo -n "  Testing API endpoint... "
    if curl -s http://localhost:8080/health | grep -q "ok"; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
    fi
    
    # Test FFmpeg
    echo -n "  Testing FFmpeg... "
    if docker compose exec -T api ffmpeg -version > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
    fi
    
    # Test Ollama
    echo -n "  Testing AI models... "
    if curl -s http://localhost:11434/api/tags | grep -q "$OLLAMA_MODEL"; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${YELLOW}⚠${NC} Models still downloading"
    fi
    
    # Test database
    echo -n "  Testing database... "
    if docker compose exec -T postgres pg_isready > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${RED}✗${NC}"
    fi
    
    log_success "Health checks complete"
}

# Show completion message
show_completion() {
    echo
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}║                                                              ║${NC}"
    echo -e "${GREEN}║         🎉 FFprobe API Successfully Deployed! 🎉            ║${NC}"
    echo -e "${GREEN}║                                                              ║${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════${NC}"
    echo
    echo -e "${CYAN}Access your API at:${NC} ${WHITE}http://localhost:8080${NC}"
    echo
    echo -e "${CYAN}Quick Test:${NC}"
    echo -e "  ${WHITE}curl http://localhost:8080/health${NC}"
    echo
    
    if [ "$DEPLOYMENT_MODE" != "quick" ]; then
        echo -e "${CYAN}Your credentials have been saved to:${NC}"
        echo -e "  ${WHITE}$INSTALL_DIR/credentials.txt${NC}"
        echo
    fi
    
    echo -e "${CYAN}Useful Commands:${NC}"
    echo -e "  ${WHITE}cd $INSTALL_DIR${NC}"
    echo -e "  ${WHITE}docker compose logs -f${NC}       # View logs"
    echo -e "  ${WHITE}docker compose stop${NC}          # Stop services"
    echo -e "  ${WHITE}docker compose start${NC}         # Start services"
    echo -e "  ${WHITE}docker compose down${NC}          # Remove services"
    echo
    
    if [ "$DEPLOYMENT_MODE" = "production" ]; then
        echo -e "${CYAN}Monitoring:${NC}"
        echo -e "  Grafana: ${WHITE}http://localhost:3000${NC}"
        echo -e "  Prometheus: ${WHITE}http://localhost:9090${NC}"
        echo
    fi
    
    echo -e "${YELLOW}Documentation:${NC} ${WHITE}https://github.com/rendiffdev/ffprobe-api${NC}"
    echo
}

# Cleanup function
cleanup() {
    if [ $? -ne 0 ]; then
        log_error "Setup failed. Cleaning up..."
        cd "$INSTALL_DIR" 2>/dev/null && docker compose down 2>/dev/null
    fi
}

# Main installation flow
main() {
    trap cleanup EXIT
    
    # Show banner
    clear
    show_banner
    
    # Detect OS
    detect_os
    
    # Run setup wizard (unless non-interactive mode)
    if [ "$NON_INTERACTIVE" != "true" ]; then
        setup_wizard
    fi
    
    # Check system requirements based on selected deployment mode
    check_requirements "$DEPLOYMENT_MODE"
    
    # Install Docker if needed
    install_docker
    
    # Validate Docker can handle the deployment mode
    validate_docker_capabilities "$DEPLOYMENT_MODE"
    
    # Clone or update repository
    log_step "Setting up FFprobe API"
    if [ -d "$INSTALL_DIR" ]; then
        log_info "Updating existing installation..."
        cd "$INSTALL_DIR"
        git pull origin main 2>/dev/null || true
    else
        log_info "Cloning repository..."
        git clone "$REPO_URL" "$INSTALL_DIR" 2>/dev/null || {
            # If repo doesn't exist, create directory structure
            mkdir -p "$INSTALL_DIR"
            cd "$INSTALL_DIR"
            
            # Download essential files
            curl -sL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/Dockerfile.btbn -o Dockerfile.btbn
            curl -sL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/go.mod -o go.mod
            curl -sL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/go.sum -o go.sum
            
            # Create minimal structure
            mkdir -p cmd/ffprobe-api internal migrations
        }
    fi
    
    # Generate configuration
    generate_config
    
    # Prepare Docker Compose
    prepare_compose
    
    # Deploy the stack
    deploy_stack
    
    # Run tests
    run_tests
    
    # Show completion message
    show_completion
}

# Handle arguments for non-interactive mode
case "$1" in
    "--quick" | "-q")
        DEPLOYMENT_MODE="quick"
        NON_INTERACTIVE=true
        log_info "Non-interactive quick mode selected"
        ;;
    "--minimal" | "-m")
        DEPLOYMENT_MODE="minimal"
        NON_INTERACTIVE=true
        log_info "Non-interactive minimal mode selected"
        ;;
    "--production" | "-p")
        DEPLOYMENT_MODE="production"
        NON_INTERACTIVE=true
        log_info "Non-interactive production mode selected"
        ;;
    "--development" | "-d")
        DEPLOYMENT_MODE="development"
        NON_INTERACTIVE=true
        log_info "Non-interactive development mode selected"
        ;;
    "--help" | "-h")
        echo "FFprobe API Automated Setup"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Deployment Options:"
        echo "  -q, --quick       Quick installation (3GB RAM, 8GB disk)"
        echo "  -m, --minimal     Minimal installation (2GB RAM, 6GB disk)"
        echo "  -p, --production  Production installation (8GB RAM, 20GB disk)"
        echo "  -d, --development Development installation (4GB RAM, 15GB disk)"
        echo ""
        echo "Other Options:"
        echo "  -h, --help        Show this help message"
        echo ""
        echo "Without options, runs interactive setup wizard"
        echo ""
        echo "System Requirements Check:"
        echo "The script will automatically check if your system meets the"
        echo "minimum requirements for the selected deployment mode."
        exit 0
        ;;
esac

# Run main installation
main