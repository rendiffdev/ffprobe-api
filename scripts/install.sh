#!/bin/bash

# FFprobe API Professional Installer
# Interactive configuration and deployment script
# Version: 1.0.0

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration variables
INSTALL_DIR="$(pwd)"
CONFIG_FILE=".env"
COMPOSE_FILE="compose.yml"
DEPLOYMENT_TYPE=""
MONITORING_TYPE=""
AI_CONFIG=""
STORAGE_TYPE=""
SCALE_PROFILE=""

# Banner
show_banner() {
    clear
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════════╗"
    echo "║                    🎬 FFprobe API Installer                      ║"
    echo "║              Professional Video Analysis Platform                 ║"
    echo "║                        by Rendiff                               ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo ""
}

# Progress indicator
show_progress() {
    local step=$1
    local total=$2
    local description=$3
    local progress=$((step * 100 / total))
    
    echo -e "${CYAN}[${step}/${total}] ${description}${NC}"
    echo -ne "${BLUE}Progress: ["
    for ((i=0; i<progress/5; i++)); do echo -ne "█"; done
    for ((i=progress/5; i<20; i++)); do echo -ne "░"; done
    echo -e "] ${progress}%${NC}"
    echo ""
}

# System requirements check
check_requirements() {
    show_progress 1 8 "Checking system requirements..."
    
    local missing_deps=()
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    else
        echo -e "${GREEN}✓ Docker found: $(docker --version | cut -d' ' -f3)${NC}"
    fi
    
    # Check Docker Compose
    if ! docker compose version &> /dev/null; then
        missing_deps+=("docker-compose-v2")
    else
        echo -e "${GREEN}✓ Docker Compose found: $(docker compose version --short)${NC}"
    fi
    
    # Check other utilities
    for cmd in curl openssl jq; do
        if ! command -v $cmd &> /dev/null; then
            missing_deps+=($cmd)
        else
            echo -e "${GREEN}✓ $cmd found${NC}"
        fi
    done
    
    # Check system resources
    local ram_gb=$(free -g | awk '/^Mem:/{print $2}')
    local disk_gb=$(df -BG . | awk 'NR==2{print int($4)}')
    
    echo -e "${BLUE}System Resources:${NC}"
    echo -e "  RAM: ${ram_gb}GB available"
    echo -e "  Disk: ${disk_gb}GB free"
    
    if [ $ram_gb -lt 4 ]; then
        echo -e "${YELLOW}⚠️  Warning: Less than 4GB RAM available. AI features may be limited.${NC}"
    fi
    
    if [ $disk_gb -lt 5 ]; then
        echo -e "${YELLOW}⚠️  Warning: Less than 5GB disk space. Model downloads may fail.${NC}"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo -e "${RED}❌ Missing dependencies: ${missing_deps[*]}${NC}"
        echo ""
        echo "Please install missing dependencies and run the installer again."
        echo ""
        echo "Installation commands:"
        echo "  Ubuntu/Debian: sudo apt update && sudo apt install docker.io docker-compose-v2 curl openssl jq"
        echo "  CentOS/RHEL: sudo yum install docker docker-compose curl openssl jq"
        echo "  macOS: brew install docker docker-compose curl openssl jq"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All requirements satisfied${NC}"
    sleep 1
}

# Deployment type selection
select_deployment_type() {
    show_progress 2 8 "Selecting deployment configuration..."
    
    echo -e "${WHITE}Choose your deployment type:${NC}"
    echo ""
    echo "1) 🚀 Development (Single server, minimal resources)"
    echo "   - RAM: 4GB, CPU: 2 cores"
    echo "   - Best for: Local development, testing, small projects"
    echo ""
    echo "2) 🏢 Production (Standard deployment)"
    echo "   - RAM: 8GB, CPU: 4 cores"
    echo "   - Best for: Small to medium production workloads"
    echo ""
    echo "3) 🌐 Enterprise (Scalable architecture)"
    echo "   - RAM: 16GB+, CPU: 8+ cores"
    echo "   - Best for: High availability, large scale, multiple instances"
    echo ""
    
    while true; do
        read -p "Enter your choice (1-3): " choice
        case $choice in
            1)
                DEPLOYMENT_TYPE="development"
                SCALE_PROFILE="dev"
                COMPOSE_FILE="compose.yml"
                echo -e "${GREEN}✓ Development deployment selected${NC}"
                break
                ;;
            2)
                DEPLOYMENT_TYPE="production"
                SCALE_PROFILE="standard"
                COMPOSE_FILE="compose.yml"
                echo -e "${GREEN}✓ Production deployment selected${NC}"
                break
                ;;
            3)
                DEPLOYMENT_TYPE="enterprise"
                SCALE_PROFILE="enterprise"
                COMPOSE_FILE="compose.enterprise.yml"
                echo -e "${GREEN}✓ Enterprise deployment selected${NC}"
                break
                ;;
            *)
                echo -e "${RED}Invalid choice. Please enter 1, 2, or 3.${NC}"
                ;;
        esac
    done
    sleep 1
}

# AI configuration
configure_ai() {
    show_progress 3 8 "Configuring AI analysis..."
    
    echo -e "${WHITE}AI Analysis Configuration:${NC}"
    echo ""
    echo "1) 🤖 Local AI Only (Phi-3 Mini, 2GB RAM, Private)"
    echo "   - No external API calls"
    echo "   - Complete data privacy"
    echo "   - No additional costs"
    echo ""
    echo "2) ☁️  Local + Cloud Fallback (Local primary, OpenRouter fallback)"
    echo "   - Best reliability"
    echo "   - Enhanced analysis quality"
    echo "   - Requires OpenRouter API key"
    echo ""
    echo "3) 🚫 Disable AI Analysis (FFprobe only)"
    echo "   - Minimal resource usage"
    echo "   - Technical analysis only"
    echo ""
    
    while true; do
        read -p "Enter your choice (1-3): " choice
        case $choice in
            1)
                AI_CONFIG="local_only"
                echo -e "${GREEN}✓ Local AI only configured${NC}"
                break
                ;;
            2)
                AI_CONFIG="local_with_fallback"
                echo -e "${GREEN}✓ Local + Cloud fallback configured${NC}"
                echo ""
                echo -e "${YELLOW}You'll need an OpenRouter API key. Get one at: https://openrouter.ai/keys${NC}"
                read -p "Enter your OpenRouter API key (or press Enter to configure later): " openrouter_key
                if [ -n "$openrouter_key" ]; then
                    echo "OPENROUTER_API_KEY=$openrouter_key" >> .env.tmp
                fi
                break
                ;;
            3)
                AI_CONFIG="disabled"
                echo -e "${GREEN}✓ AI analysis disabled${NC}"
                break
                ;;
            *)
                echo -e "${RED}Invalid choice. Please enter 1, 2, or 3.${NC}"
                ;;
        esac
    done
    sleep 1
}

# Monitoring configuration
configure_monitoring() {
    show_progress 4 8 "Configuring monitoring..."
    
    echo -e "${WHITE}Monitoring Configuration:${NC}"
    echo ""
    echo "1) 📊 Local Monitoring (Prometheus + Grafana)"
    echo "   - Self-hosted dashboards"
    echo "   - Complete control"
    echo "   - Additional ~1GB RAM usage"
    echo ""
    echo "2) ☁️  Grafana Cloud (Recommended for production)"
    echo "   - Managed service"
    echo "   - No local resources"
    echo "   - Requires Grafana Cloud account"
    echo ""
    echo "3) 📈 Basic Monitoring (Health checks only)"
    echo "   - Minimal resource usage"
    echo "   - Basic health endpoints"
    echo ""
    
    while true; do
        read -p "Enter your choice (1-3): " choice
        case $choice in
            1)
                MONITORING_TYPE="local"
                echo -e "${GREEN}✓ Local monitoring configured${NC}"
                break
                ;;
            2)
                MONITORING_TYPE="grafana_cloud"
                echo -e "${GREEN}✓ Grafana Cloud monitoring configured${NC}"
                echo ""
                echo -e "${YELLOW}You'll need Grafana Cloud credentials. Sign up at: https://grafana.com/products/cloud/${NC}"
                read -p "Enter your Grafana Cloud API key (or press Enter to configure later): " grafana_key
                if [ -n "$grafana_key" ]; then
                    echo "GRAFANA_CLOUD_API_KEY=$grafana_key" >> .env.tmp
                fi
                break
                ;;
            3)
                MONITORING_TYPE="basic"
                echo -e "${GREEN}✓ Basic monitoring configured${NC}"
                break
                ;;
            *)
                echo -e "${RED}Invalid choice. Please enter 1, 2, or 3.${NC}"
                ;;
        esac
    done
    sleep 1
}

# Storage configuration
configure_storage() {
    show_progress 5 8 "Configuring storage..."
    
    echo -e "${WHITE}Storage Configuration:${NC}"
    echo ""
    echo "1) 💾 Local Storage (Default)"
    echo "   - Files stored on local disk"
    echo "   - No additional setup required"
    echo ""
    echo "2) ☁️  AWS S3"
    echo "   - Cloud storage"
    echo "   - Scalable and durable"
    echo ""
    echo "3) 📦 Google Cloud Storage"
    echo "   - Google Cloud integration"
    echo "   - Global content delivery"
    echo ""
    
    while true; do
        read -p "Enter your choice (1-3): " choice
        case $choice in
            1)
                STORAGE_TYPE="local"
                echo -e "${GREEN}✓ Local storage configured${NC}"
                break
                ;;
            2)
                STORAGE_TYPE="s3"
                echo -e "${GREEN}✓ AWS S3 storage configured${NC}"
                echo ""
                read -p "AWS Access Key ID: " aws_key
                read -s -p "AWS Secret Access Key: " aws_secret
                echo ""
                read -p "S3 Bucket Name: " s3_bucket
                read -p "AWS Region (default: us-east-1): " aws_region
                aws_region=${aws_region:-us-east-1}
                
                echo "AWS_ACCESS_KEY_ID=$aws_key" >> .env.tmp
                echo "AWS_SECRET_ACCESS_KEY=$aws_secret" >> .env.tmp
                echo "S3_BUCKET=$s3_bucket" >> .env.tmp
                echo "AWS_REGION=$aws_region" >> .env.tmp
                echo "STORAGE_TYPE=s3" >> .env.tmp
                break
                ;;
            3)
                STORAGE_TYPE="gcs"
                echo -e "${GREEN}✓ Google Cloud Storage configured${NC}"
                echo ""
                read -p "GCS Bucket Name: " gcs_bucket
                read -p "GCP Project ID: " gcp_project
                
                echo "GCS_BUCKET=$gcs_bucket" >> .env.tmp
                echo "GCP_PROJECT_ID=$gcp_project" >> .env.tmp
                echo "STORAGE_TYPE=gcs" >> .env.tmp
                break
                ;;
            *)
                echo -e "${RED}Invalid choice. Please enter 1, 2, or 3.${NC}"
                ;;
        esac
    done
    sleep 1
}

# Generate configuration
generate_config() {
    show_progress 6 8 "Generating configuration..."
    
    # Start with base configuration
    cp .env.example .env
    
    # Generate secure credentials
    local api_key="ffprobe_live_sk_$(openssl rand -hex 32)"
    local jwt_secret="$(openssl rand -hex 32)"
    local db_password="$(openssl rand -hex 16)"
    local redis_password="$(openssl rand -hex 16)"
    local grafana_password="$(openssl rand -hex 12)"
    
    # Apply configuration based on selections
    cat >> .env << EOF

# Generated Configuration
GO_ENV=$DEPLOYMENT_TYPE
API_KEY=$api_key
JWT_SECRET=$jwt_secret
POSTGRES_PASSWORD=$db_password
REDIS_PASSWORD=$redis_password
GRAFANA_PASSWORD=$grafana_password

# Deployment Configuration
SCALE_PROFILE=$SCALE_PROFILE
DEPLOYMENT_TYPE=$DEPLOYMENT_TYPE

# AI Configuration
EOF

    case $AI_CONFIG in
        "local_only")
            cat >> .env << EOF
ENABLE_LOCAL_LLM=true
OLLAMA_MODEL=phi3:mini
EOF
            ;;
        "local_with_fallback")
            cat >> .env << EOF
ENABLE_LOCAL_LLM=true
OLLAMA_MODEL=phi3:mini
ENABLE_OPENROUTER_FALLBACK=true
EOF
            ;;
        "disabled")
            cat >> .env << EOF
ENABLE_LOCAL_LLM=false
EOF
            ;;
    esac

    # Add monitoring configuration
    case $MONITORING_TYPE in
        "local")
            cat >> .env << EOF
ENABLE_PROMETHEUS=true
ENABLE_GRAFANA=true
EOF
            ;;
        "grafana_cloud")
            cat >> .env << EOF
ENABLE_PROMETHEUS=true
ENABLE_GRAFANA=false
MONITORING_TYPE=grafana_cloud
EOF
            ;;
        "basic")
            cat >> .env << EOF
ENABLE_PROMETHEUS=false
ENABLE_GRAFANA=false
EOF
            ;;
    esac
    
    # Add any additional configuration from temporary file
    if [ -f .env.tmp ]; then
        cat .env.tmp >> .env
        rm .env.tmp
    fi
    
    echo -e "${GREEN}✓ Configuration file generated${NC}"
    sleep 1
}

# Deploy services
deploy_services() {
    show_progress 7 8 "Deploying services..."
    
    echo -e "${BLUE}Starting deployment...${NC}"
    
    # Create necessary directories
    mkdir -p data/{postgres,redis,ollama,uploads,reports,logs,prometheus,grafana}
    
    # Set proper permissions
    chmod 755 data/
    chmod -R 777 data/uploads data/reports data/logs
    
    # Start services based on deployment type
    case $DEPLOYMENT_TYPE in
        "development")
            echo -e "${BLUE}Starting development environment...${NC}"
            docker compose up -d
            ;;
        "production")
            echo -e "${BLUE}Starting production environment...${NC}"
            docker compose -f compose.yml -f compose.production.yml up -d
            ;;
        "enterprise")
            echo -e "${BLUE}Starting enterprise environment...${NC}"
            docker compose -f compose.yml -f compose.enterprise.yml up -d
            ;;
    esac
    
    echo -e "${BLUE}Waiting for services to start...${NC}"
    sleep 30
    
    # Health check
    local retries=30
    while [ $retries -gt 0 ]; do
        if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "${GREEN}✓ API service is healthy${NC}"
            break
        fi
        echo -e "${YELLOW}Waiting for API service... ($retries retries left)${NC}"
        sleep 5
        ((retries--))
    done
    
    if [ $retries -eq 0 ]; then
        echo -e "${RED}❌ API service failed to start. Check logs with: docker compose logs${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ All services deployed successfully${NC}"
    sleep 1
}

# Show installation summary
show_summary() {
    show_progress 8 8 "Installation complete!"
    
    echo ""
    echo -e "${GREEN}🎉 FFprobe API Installation Complete!${NC}"
    echo ""
    echo -e "${WHITE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BLUE}📋 Installation Summary:${NC}"
    echo -e "  🔧 Deployment Type: ${YELLOW}$DEPLOYMENT_TYPE${NC}"
    echo -e "  🤖 AI Configuration: ${YELLOW}$AI_CONFIG${NC}"
    echo -e "  📊 Monitoring: ${YELLOW}$MONITORING_TYPE${NC}"
    echo -e "  💾 Storage: ${YELLOW}$STORAGE_TYPE${NC}"
    echo ""
    echo -e "${BLUE}🌐 Service URLs:${NC}"
    echo -e "  🎬 FFprobe API: ${CYAN}http://localhost:8080${NC}"
    echo -e "  📊 Health Check: ${CYAN}http://localhost:8080/health${NC}"
    
    if [[ "$MONITORING_TYPE" == "local" ]]; then
        echo -e "  📈 Grafana: ${CYAN}http://localhost:3000${NC} (admin/$(grep GRAFANA_PASSWORD .env | cut -d'=' -f2))"
        echo -e "  🔍 Prometheus: ${CYAN}http://localhost:9090${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}🔑 Authentication:${NC}"
    echo -e "  API Key: ${YELLOW}$(grep API_KEY .env | cut -d'=' -f2 | cut -c1-32)...${NC}"
    echo -e "  Full key saved in: ${CYAN}.env${NC}"
    echo ""
    echo -e "${BLUE}🚀 Quick Test:${NC}"
    echo -e "${CYAN}curl -H \"X-API-Key: \$(grep API_KEY .env | cut -d'=' -f2)\" http://localhost:8080/health${NC}"
    echo ""
    echo -e "${BLUE}📚 Documentation:${NC}"
    echo -e "  📖 API Docs: ${CYAN}docs/api/${NC}"
    echo -e "  🔧 Configuration: ${CYAN}docs/deployment/configuration.md${NC}"
    echo -e "  🆘 Troubleshooting: ${CYAN}docs/TROUBLESHOOTING.md${NC}"
    echo ""
    echo -e "${BLUE}⚙️  Management Commands:${NC}"
    echo -e "  Start: ${CYAN}docker compose up -d${NC}"
    echo -e "  Stop: ${CYAN}docker compose down${NC}"
    echo -e "  Logs: ${CYAN}docker compose logs -f${NC}"
    echo -e "  Status: ${CYAN}docker compose ps${NC}"
    echo ""
    echo -e "${WHITE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${GREEN}✨ Your professional video analysis platform is ready to use!${NC}"
    echo ""
}

# Main installation flow
main() {
    show_banner
    
    # Check if already installed
    if [ -f ".env" ] && [ -f "docker-compose.yml" ]; then
        echo -e "${YELLOW}⚠️  Existing installation detected.${NC}"
        read -p "Do you want to reinstall? This will overwrite your current configuration. (y/N): " reinstall
        if [[ ! $reinstall =~ ^[Yy]$ ]]; then
            echo "Installation cancelled."
            exit 0
        fi
        echo ""
    fi
    
    check_requirements
    select_deployment_type
    configure_ai
    configure_monitoring
    configure_storage
    generate_config
    deploy_services
    show_summary
}

# Error handling
trap 'echo -e "\n${RED}❌ Installation failed. Check the logs above for details.${NC}"; exit 1' ERR

# Run main installation
main "$@"