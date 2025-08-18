#!/bin/bash

# FFprobe API - Ollama Model Setup
# Downloads and configures AI models for analysis

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }

download_model() {
    local model=$1
    print_info "Downloading model: $model..."
    
    if docker compose exec ollama ollama show $model >/dev/null 2>&1; then
        print_success "Model $model already available"
        return 0
    fi
    
    print_info "Pulling $model (this may take a few minutes)..."
    if docker compose exec ollama ollama pull $model; then
        print_success "Successfully downloaded $model"
    else
        print_warning "Failed to download $model"
        return 1
    fi
}

wait_for_ollama() {
    print_info "Waiting for Ollama service to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost:11434/api/version >/dev/null 2>&1; then
            print_success "Ollama service is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_warning "Ollama service not ready after 60 seconds"
    return 1
}

detect_system_resources() {
    # Detect memory for model selection
    local memory_gb=4
    
    if [[ -f /proc/meminfo ]]; then
        local memory_kb=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        memory_gb=$((memory_kb / 1024 / 1024))
    elif command -v sysctl >/dev/null 2>&1; then
        local memory_bytes=$(sysctl -n hw.memsize 2>/dev/null || echo "4294967296")
        memory_gb=$((memory_bytes / 1024 / 1024 / 1024))
    fi
    
    print_info "Detected system memory: ${memory_gb}GB"
    echo "$memory_gb"
}

select_models_for_system() {
    local memory_gb=$1
    local models=()
    
    # gemma3:270m is ALWAYS the primary model, downloaded first
    if [[ $memory_gb -ge 16 ]]; then
        print_info "High-end system detected - gemma3:270m as primary + additional models"
        models=("gemma3:270m" "phi3:mini" "llama3:8b")
    elif [[ $memory_gb -ge 4 ]]; then
        print_info "Standard system detected - gemma3:270m as primary + phi3:mini fallback"
        models=("gemma3:270m" "phi3:mini")
    else
        print_info "Limited system detected - gemma3:270m only (primary)"
        models=("gemma3:270m")
    fi
    
    printf '%s\n' "${models[@]}"
}

main() {
    print_info "🤖 Setting up Ollama AI models..."
    
    # Check if Ollama service exists in compose
    if ! docker compose config 2>/dev/null | grep -q "ollama:"; then
        print_warning "Ollama service not found in docker compose configuration"
        print_info "Make sure you have ollama service defined in your compose files"
        exit 1
    fi
    
    # Start Ollama if not running
    if ! docker compose ps ollama 2>/dev/null | grep -q "Up"; then
        print_info "Starting Ollama service..."
        docker compose up -d ollama
    fi
    
    # Wait for service to be ready
    if ! wait_for_ollama; then
        print_error "Failed to start Ollama service"
        print_info "Check logs with: docker compose logs ollama"
        exit 1
    fi
    
    # Detect system and select appropriate models
    local memory_gb=$(detect_system_resources)
    local models=($(select_models_for_system "$memory_gb"))
    
    print_info "Selected models for download: ${models[*]}"
    
    # Download models
    local failed=0
    for model in "${models[@]}"; do
        if ! download_model "$model"; then
            failed=$((failed + 1))
        fi
        # Small delay between downloads
        sleep 1
    done
    
    # Summary
    echo
    if [[ $failed -eq 0 ]]; then
        print_success "🎉 All AI models downloaded successfully!"
    else
        print_warning "⚠️  $failed model(s) failed to download"
    fi
    
    print_info "Available models:"
    if docker compose exec ollama ollama list 2>/dev/null; then
        print_success "Ollama setup complete!"
    else
        print_warning "Could not list models - Ollama may still be initializing"
    fi
    
    echo
    print_info "💡 Model Usage Tips:"
    print_info "• gemma3:270m - PRIMARY MODEL ⭐ Ultra-fast, 270M params (~200MB RAM)"
    print_info "• phi3:mini - Fallback model, better reasoning, 3.8B params (~2GB RAM)"
    print_info "• llama3:8b - Optional production model, 8B params (~8GB RAM)"
    
    echo
    print_info "🔧 To test AI analysis:"
    print_info "curl -X POST -F 'file=@video.mp4' -F 'include_llm=true' \\"
    print_info "     http://localhost:8080/api/v1/probe/file"
}

main "$@"