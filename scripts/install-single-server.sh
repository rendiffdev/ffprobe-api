#!/bin/bash

# FFprobe API Single Server Installation
# Minimal setup for development and small-scale deployments
# Version: 1.0.0

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

INSTALL_DIR="$(pwd)"

# Banner
show_banner() {
    clear
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════════╗"
    echo "║               🚀 Single Server Installation                      ║"
    echo "║                  FFprobe API - Quick Setup                      ║"
    echo "║              Perfect for Dev & Small Deployments               ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo ""
}

# System check
check_system() {
    echo -e "${BLUE}🔍 Checking system requirements...${NC}"
    
    # Check RAM
    local ram_gb=$(free -g | awk '/^Mem:/{print $2}')
    if [ $ram_gb -lt 4 ]; then
        echo -e "${YELLOW}⚠️  Warning: Only ${ram_gb}GB RAM available. Recommended: 4GB+${NC}"
        read -p "Continue anyway? (y/N): " continue_low_ram
        if [[ ! $continue_low_ram =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        echo -e "${GREEN}✓ RAM: ${ram_gb}GB available${NC}"
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker not found. Please install Docker first.${NC}"
        echo "Install: curl -fsSL https://get.docker.com | sh"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker: $(docker --version | cut -d' ' -f3)${NC}"
    
    # Check Docker Compose
    if ! docker compose version &> /dev/null; then
        echo -e "${RED}❌ Docker Compose v2 not found.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Docker Compose: $(docker compose version --short)${NC}"
    
    echo ""
}

# Quick configuration
quick_config() {
    echo -e "${BLUE}⚙️  Generating configuration...${NC}"
    
    # Copy base config
    if [ ! -f .env ]; then
        cp .env.example .env
    fi
    
    # Generate secure credentials
    local api_key="ffprobe_dev_sk_$(openssl rand -hex 32)"
    local jwt_secret="$(openssl rand -hex 32)"
    local db_pass="dev_$(openssl rand -hex 8)"
    local redis_pass="dev_$(openssl rand -hex 8)"
    
    # Apply single-server optimizations
    cat > .env.single << EOF
# Single Server Configuration
GO_ENV=development
API_PORT=8080

# Authentication
ENABLE_AUTH=true
API_KEY=$api_key
JWT_SECRET=$jwt_secret

# Database (lightweight config)
POSTGRES_HOST=postgres
POSTGRES_PASSWORD=$db_pass
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5

# Redis (minimal config)
REDIS_HOST=redis
REDIS_PASSWORD=$redis_pass

# AI Processing (optimized for low resources)
ENABLE_LOCAL_LLM=true
OLLAMA_MODEL=phi3:mini
OLLAMA_KEEP_ALIVE=30m
OLLAMA_MAX_LOADED_MODELS=1

# File limits (reasonable for single server)
MAX_FILE_SIZE=5368709120  # 5GB
MAX_CONCURRENT_JOBS=2

# Monitoring (lightweight)
ENABLE_PROMETHEUS=true
ENABLE_GRAFANA=false  # Disabled to save resources

# Performance (single server optimized)
WORKER_POOL_SIZE=4
PROCESSING_TIMEOUT=300
EOF
    
    # Merge with existing .env
    cat .env.single >> .env
    rm .env.single
    
    echo -e "${GREEN}✓ Configuration generated${NC}"
    echo -e "  API Key: ${YELLOW}${api_key:0:32}...${NC}"
    echo ""
}

# Create single server compose
create_compose() {
    echo -e "${BLUE}🐳 Creating single-server Docker configuration...${NC}"
    
    cat > compose.single.yml << 'EOF'
# Single Server Docker Compose
# Optimized for development and small deployments
# Usage: docker compose -f compose.single.yml up -d

services:
  # Main API with reduced resources
  ffprobe-api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - POSTGRES_HOST=postgres
      - POSTGRES_PORT=5432
      - POSTGRES_DB=ffprobe_api
      - POSTGRES_USER=ffprobe
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - OLLAMA_URL=http://ollama:11434
      - OLLAMA_MODEL=${OLLAMA_MODEL:-phi3:mini}
      - API_KEY=${API_KEY}
      - JWT_SECRET=${JWT_SECRET}
      - ENABLE_LOCAL_LLM=true
    depends_on:
      - postgres
      - redis
      - ollama
    volumes:
      - ./data/uploads:/app/uploads
      - ./data/reports:/app/reports
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.5'
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Lightweight PostgreSQL
  postgres:
    image: postgres:16.1-alpine
    environment:
      - POSTGRES_DB=ffprobe_api
      - POSTGRES_USER=ffprobe
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
      - ./migrations/001_initial_schema.up.sql:/docker-entrypoint-initdb.d/01-schema.sql:ro
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ffprobe -d ffprobe_api"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # Lightweight Redis
  redis:
    image: redis:7.2.4-alpine
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - ./data/redis:/data
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
    healthcheck:
      test: ["CMD", "redis-cli", "--no-auth-warning", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # Lightweight Ollama
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    environment:
      - OLLAMA_ORIGINS=*
      - OLLAMA_HOST=0.0.0.0:11434
      - OLLAMA_KEEP_ALIVE=30m
      - OLLAMA_MAX_LOADED_MODELS=1
      - OLLAMA_NUM_PARALLEL=2
    volumes:
      - ./data/ollama:/root/.ollama
      - ./docker/ollama-entrypoint.sh:/ollama-entrypoint.sh:ro
    deploy:
      resources:
        limits:
          memory: 2.5G
          cpus: '1.0'
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:11434/api/version"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 60s
    entrypoint: ["/ollama-entrypoint.sh"]
    restart: unless-stopped

  # Basic monitoring (Prometheus only)
  prometheus:
    image: prom/prometheus:v2.49.1
    ports:
      - "9090:9090"
    volumes:
      - ./docker/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./data/prometheus:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=7d'
      - '--storage.tsdb.retention.size=5GB'
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
    restart: unless-stopped

networks:
  default:
    name: ffprobe-single
EOF
    
    echo -e "${GREEN}✓ Single-server compose file created${NC}"
    echo ""
}

# Deploy services
deploy() {
    echo -e "${BLUE}🚀 Starting deployment...${NC}"
    
    # Create data directories
    mkdir -p data/{postgres,redis,ollama,uploads,reports,prometheus}
    chmod -R 755 data/
    
    # Start services
    echo -e "${BLUE}Starting services (this may take a few minutes)...${NC}"
    docker compose -f compose.single.yml up -d
    
    echo -e "${BLUE}Waiting for services to initialize...${NC}"
    sleep 45
    
    # Health check
    local retries=20
    while [ $retries -gt 0 ]; do
        if curl -s -f "http://localhost:8080/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ API service is ready!${NC}"
            break
        fi
        echo -e "${YELLOW}Waiting for API... ($retries retries left)${NC}"
        sleep 5
        ((retries--))
    done
    
    if [ $retries -eq 0 ]; then
        echo -e "${RED}❌ Service startup failed. Checking logs...${NC}"
        docker compose -f compose.single.yml logs --tail=20
        exit 1
    fi
    
    echo ""
}

# Show completion summary
show_completion() {
    local api_key=$(grep "^API_KEY=" .env | cut -d'=' -f2)
    
    echo -e "${GREEN}🎉 Single Server Installation Complete!${NC}"
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BLUE}📋 Service URLs:${NC}"
    echo -e "  🎬 FFprobe API: ${YELLOW}http://localhost:8080${NC}"
    echo -e "  📊 Health Check: ${YELLOW}http://localhost:8080/health${NC}"
    echo -e "  📈 Prometheus: ${YELLOW}http://localhost:9090${NC}"
    echo ""
    echo -e "${BLUE}🔑 Your API Key:${NC}"
    echo -e "  ${YELLOW}$api_key${NC}"
    echo ""
    echo -e "${BLUE}🧪 Quick Test:${NC}"
    echo -e "${YELLOW}curl -H \"X-API-Key: $api_key\" http://localhost:8080/health${NC}"
    echo ""
    echo -e "${BLUE}⚙️  Management:${NC}"
    echo -e "  Start: ${YELLOW}docker compose -f compose.single.yml up -d${NC}"
    echo -e "  Stop: ${YELLOW}docker compose -f compose.single.yml down${NC}"
    echo -e "  Logs: ${YELLOW}docker compose -f compose.single.yml logs -f${NC}"
    echo ""
    echo -e "${BLUE}📚 Resources:${NC}"
    echo -e "  Memory Usage: ~3.5GB"
    echo -e "  CPU Usage: ~2 cores"
    echo -e "  Disk Usage: ~3GB + video files"
    echo ""
    echo -e "${BLUE}📖 Documentation:${NC}"
    echo -e "  API Guide: ${YELLOW}docs/api/complete-api-guide.md${NC}"
    echo -e "  Troubleshooting: ${YELLOW}docs/TROUBLESHOOTING.md${NC}"
    echo ""
    echo -e "${GREEN}✨ Your lightweight video analysis platform is ready!${NC}"
    echo ""
}

# Main installation flow
main() {
    show_banner
    
    # Check for existing installation
    if [ -f "compose.single.yml" ]; then
        echo -e "${YELLOW}⚠️  Single server installation detected.${NC}"
        read -p "Reinstall? This will recreate configuration. (y/N): " reinstall
        if [[ ! $reinstall =~ ^[Yy]$ ]]; then
            echo "Installation cancelled."
            exit 0
        fi
    fi
    
    check_system
    quick_config
    create_compose
    deploy
    show_completion
}

# Error handling
trap 'echo -e "\n${RED}❌ Installation failed. Run with -v for details.${NC}"; exit 1' ERR

# Check for verbose mode
if [[ "${1:-}" == "-v" ]]; then
    set -x
fi

# Run installation
main "$@"