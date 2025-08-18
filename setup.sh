#!/bin/bash

# FFprobe API - Smart Setup Script
# Automatically detects system and sets up the best deployment configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

# System detection
detect_system() {
    print_info "Detecting system configuration..."
    
    # Get system info
    OS=$(uname -s)
    ARCH=$(uname -m)
    CORES=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "4")
    
    # Get memory info
    if [[ "$OS" == "Linux" ]]; then
        MEMORY_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        MEMORY_GB=$((MEMORY_KB / 1024 / 1024))
    elif [[ "$OS" == "Darwin" ]]; then
        MEMORY_BYTES=$(sysctl -n hw.memsize)
        MEMORY_GB=$((MEMORY_BYTES / 1024 / 1024 / 1024))
    else
        MEMORY_GB=4  # Default assumption
    fi
    
    # Get disk space
    DISK_GB=$(df -BG . | tail -1 | awk '{print $4}' | sed 's/G//')
    
    print_info "System: $OS $ARCH"
    print_info "CPU Cores: $CORES"
    print_info "Memory: ${MEMORY_GB}GB"
    print_info "Available Disk: ${DISK_GB}GB"
}

# Check dependencies
check_dependencies() {
    print_info "Checking dependencies..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker not found. Installing Docker..."
        install_docker
    else
        print_success "Docker found: $(docker --version)"
    fi
    
    # Check Docker Compose
    if ! docker compose version &> /dev/null; then
        print_error "Docker Compose not found or outdated"
        install_docker_compose
    else
        print_success "Docker Compose found: $(docker compose version)"
    fi
    
    # Check if Docker is running
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running. Please start Docker and try again."
        exit 1
    fi
}

# Install Docker (basic)
install_docker() {
    print_info "Installing Docker..."
    
    if [[ "$OS" == "Linux" ]]; then
        # Ubuntu/Debian
        if command -v apt &> /dev/null; then
            sudo apt update
            sudo apt install -y docker.io docker-compose-plugin
            sudo systemctl start docker
            sudo systemctl enable docker
            sudo usermod -aG docker $USER
            print_warning "Please log out and back in for Docker permissions to take effect"
        # CentOS/RHEL
        elif command -v yum &> /dev/null; then
            sudo yum install -y docker docker-compose
            sudo systemctl start docker
            sudo systemctl enable docker
            sudo usermod -aG docker $USER
        else
            print_error "Unsupported Linux distribution. Please install Docker manually."
            exit 1
        fi
    elif [[ "$OS" == "Darwin" ]]; then
        print_error "Please install Docker Desktop for Mac from: https://docs.docker.com/desktop/mac/"
        exit 1
    else
        print_error "Unsupported operating system. Please install Docker manually."
        exit 1
    fi
}

# Install Docker Compose
install_docker_compose() {
    print_info "Docker Compose plugin should be included with Docker. Please update Docker."
}

# Recommend deployment mode based on system resources
recommend_deployment() {
    print_info "Analyzing system for optimal deployment mode..."
    
    if [[ $MEMORY_GB -ge 8 && $DISK_GB -ge 20 && $CORES -ge 4 ]]; then
        MODE="production"
        print_success "🏭 Production mode recommended (Full features + monitoring)"
    elif [[ $MEMORY_GB -ge 4 && $DISK_GB -ge 15 && $CORES -ge 2 ]]; then
        MODE="development"
        print_success "🔧 Development mode recommended (Full features, hot reload)"
    elif [[ $MEMORY_GB -ge 3 && $DISK_GB -ge 8 ]]; then
        MODE="quick"
        print_success "⚡ Quick mode recommended (Core features only)"
    elif [[ $MEMORY_GB -ge 2 && $DISK_GB -ge 6 ]]; then
        MODE="minimal"
        print_success "🥅 Minimal mode recommended (Essential services only)"
    else
        MODE="minimal"
        print_warning "⚠️  System resources are low. Minimal mode recommended."
        print_warning "   Consider upgrading: ${MEMORY_GB}GB RAM, ${DISK_GB}GB disk available"
    fi
}

# Interactive mode selection
select_mode() {
    echo
    print_info "Available deployment modes:"
    echo "  1) 🥅 minimal    - API + SQLite + Valkey + Ollama (2GB RAM, 6GB disk)"
    echo "  2) ⚡ quick      - Same as minimal + dev settings (3GB RAM, 8GB disk)"  
    echo "  3) 🔧 development - Full features + hot reload (4GB RAM, 15GB disk)"
    echo "  4) 🏭 production - Full features + monitoring (8GB RAM, 20GB disk)"
    echo "  5) 🎯 auto       - Let the system choose (recommended: $MODE)"
    echo
    
    read -p "Choose deployment mode [1-5] (default: 5): " choice
    
    case $choice in
        1) MODE="minimal" ;;
        2) MODE="quick" ;;
        3) MODE="development" ;;
        4) MODE="production" ;;
        5|"") ;; # Keep recommended mode
        *) print_warning "Invalid choice. Using recommended: $MODE" ;;
    esac
    
    print_success "Selected mode: $MODE"
}

# Generate environment file
generate_env() {
    print_info "Generating .env configuration..."
    
    # Copy template
    if [[ -f ".env.example" ]]; then
        cp .env.example .env
        print_success "Generated .env from template"
    else
        create_basic_env
    fi
    
    # Generate secure passwords
    VALKEY_PASSWORD=$(openssl rand -hex 16)
    JWT_SECRET=$(openssl rand -hex 32)
    API_KEY="ffprobe_$(echo $MODE)_sk_$(openssl rand -hex 32)"
    
    # Update .env file
    if [[ "$OS" == "Darwin" ]]; then
        # macOS sed
        sed -i '' "s/change_this_secure_valkey_password/$VALKEY_PASSWORD/g" .env
        sed -i '' "s/your-jwt-secret-key-minimum-32-characters-long-for-security/$JWT_SECRET/g" .env
        sed -i '' "s/ffprobe_test_sk_.*/$API_KEY/g" .env
    else
        # Linux sed
        sed -i "s/change_this_secure_valkey_password/$VALKEY_PASSWORD/g" .env
        sed -i "s/your-jwt-secret-key-minimum-32-characters-long-for-security/$JWT_SECRET/g" .env
        sed -i "s/ffprobe_test_sk_.*/$API_KEY/g" .env
    fi
    
    print_success "Generated secure credentials"
    print_info "API Key: $API_KEY"
    print_warning "Save this API key - you'll need it for requests"
}

# Create basic .env if template doesn't exist
create_basic_env() {
    cat > .env << 'EOF'
# FFprobe API Configuration
GO_ENV=development
API_PORT=8080
HOST=0.0.0.0

# Database
DB_TYPE=sqlite
DB_PATH=/app/data/ffprobe.db

# Cache
VALKEY_HOST=valkey
VALKEY_PORT=6379
VALKEY_PASSWORD=change_this_secure_valkey_password

# Security
ENABLE_AUTH=true
API_KEY=ffprobe_dev_sk_0123456789abcdef0123456789abcdef01234567890abcdef0123456789abcdef
JWT_SECRET=your-jwt-secret-key-minimum-32-characters-long-for-security

# LLM Configuration
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=gemma3:270m
OLLAMA_FALLBACK_MODEL=phi3:mini

# Storage
UPLOAD_DIR=/app/uploads
TEMP_DIR=/app/temp
MAX_FILE_SIZE=53687091200

# Performance
MAX_CONCURRENT_JOBS=2
PROCESSING_TIMEOUT=300
EOF
}

# Deploy based on selected mode
deploy() {
    print_info "Deploying FFprobe API in $MODE mode..."
    
    case $MODE in
        "minimal")
            docker compose -f docker-image/compose.yaml --profile minimal up -d
            ;;
        "quick")
            docker compose -f docker-image/compose.yaml --profile quick up -d
            ;;
        "development")
            docker compose -f docker-image/compose.yaml -f docker-image/compose.development.yaml --profile development up -d
            ;;
        "production")
            docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml --profile production up -d
            ;;
    esac
    
    print_success "Deployment started!"
}

# Wait for services to be ready
wait_for_services() {
    print_info "Waiting for services to start..."
    
    local max_attempts=60
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost:8080/health >/dev/null 2>&1; then
            print_success "API is ready!"
            break
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -eq $max_attempts ]; then
        print_warning "Services may still be starting. Check with: docker compose logs"
    fi
}

# Pull Ollama models in background
setup_ollama_models() {
    print_info "Setting up AI models (this may take a few minutes)..."
    
    # Start model downloads in background
    docker compose exec -d ollama ollama pull gemma3:270m || true
    docker compose exec -d ollama ollama pull phi3:mini || true
    
    print_info "AI models are downloading in the background..."
    print_info "Check progress with: docker compose logs ollama"
}

# Print final instructions
print_success_message() {
    echo
    print_success "🎉 FFprobe API is ready!"
    echo
    print_info "📍 API URL: http://localhost:8080"
    print_info "🔑 API Key: $(grep API_KEY .env | cut -d'=' -f2)"
    echo
    print_info "📚 Quick Start:"
    echo "   # Health check"
    echo "   curl http://localhost:8080/health"
    echo
    echo "   # Analyze a video file"
    echo "   curl -X POST -F 'file=@video.mp4' \\"
    echo "        -H 'X-API-Key: $(grep API_KEY .env | cut -d'=' -f2)' \\"
    echo "        http://localhost:8080/api/v1/probe/file"
    echo
    echo "   # AI-powered analysis"
    echo "   curl -X POST -F 'file=@video.mp4' -F 'include_llm=true' \\"
    echo "        -H 'X-API-Key: $(grep API_KEY .env | cut -d'=' -f2)' \\"
    echo "        http://localhost:8080/api/v1/probe/file"
    echo
    print_info "📖 Management Commands:"
    echo "   make logs       # View logs"
    echo "   make stop       # Stop services"
    echo "   make restart    # Restart services"
    echo "   make health     # Check status"
    echo
    if [[ "$MODE" == "production" ]]; then
        print_info "📊 Monitoring: http://localhost:3000 (admin/$(grep GRAFANA_PASSWORD .env | cut -d'=' -f2))"
    fi
}

# Handle command line arguments
handle_args() {
    case "${1:-}" in
        "--quick"|"quick")
            MODE="quick"
            AUTO_MODE=true
            ;;
        "--minimal"|"minimal")
            MODE="minimal"
            AUTO_MODE=true
            ;;
        "--development"|"dev")
            MODE="development"
            AUTO_MODE=true
            ;;
        "--production"|"prod")
            MODE="production"
            AUTO_MODE=true
            ;;
        "--help"|"-h")
            print_info "FFprobe API Setup Script"
            echo "Usage: $0 [mode]"
            echo "Modes: quick, minimal, development, production"
            echo "Or run without arguments for interactive setup"
            exit 0
            ;;
    esac
}

# Main execution
main() {
    echo "🎬 FFprobe API - Smart Setup Script"
    echo "===================================="
    
    handle_args "$@"
    
    # System checks
    detect_system
    check_dependencies
    
    # Mode selection
    if [[ -z "${AUTO_MODE:-}" ]]; then
        recommend_deployment
        select_mode
    else
        print_success "Auto mode: $MODE"
    fi
    
    # Setup
    generate_env
    deploy
    wait_for_services
    setup_ollama_models
    print_success_message
    
    echo
    print_success "Setup complete! 🚀"
}

# Run main function with all arguments
main "$@"