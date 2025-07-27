#!/bin/bash
set -e

# Ollama Model Setup Script for FFprobe API
# This script helps you download and configure optimal LLM models

echo "🦙 FFprobe API - Ollama Model Setup"
echo "=================================="
echo ""

# Configuration
OLLAMA_URL="${OLLAMA_URL:-http://localhost:11434}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Check if Ollama is running
check_ollama() {
    print_info "Checking Ollama service..."
    
    if curl -s "$OLLAMA_URL/api/version" > /dev/null 2>&1; then
        print_success "Ollama is running at $OLLAMA_URL"
        return 0
    else
        print_error "Ollama is not running at $OLLAMA_URL"
        echo ""
        echo "Please ensure Ollama is running:"
        echo "  • Docker: docker compose up ollama"
        echo "  • Native: ollama serve"
        echo ""
        return 1
    fi
}

# Get system info
get_system_info() {
    echo "🖥️  System Information:"
    echo "   OS: $(uname -s)"
    echo "   Arch: $(uname -m)"
    
    # Check available RAM
    if command -v free >/dev/null 2>&1; then
        RAM_GB=$(free -g | awk '/^Mem:/{print $2}')
        echo "   RAM: ${RAM_GB}GB available"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        RAM_BYTES=$(sysctl -n hw.memsize)
        RAM_GB=$((RAM_BYTES / 1024 / 1024 / 1024))
        echo "   RAM: ${RAM_GB}GB available"
    else
        echo "   RAM: Unable to detect"
        RAM_GB=8  # Default assumption
    fi
    
    # Check for GPU
    if command -v nvidia-smi >/dev/null 2>&1; then
        GPU_INFO=$(nvidia-smi --query-gpu=name,memory.total --format=csv,noheader,nounits | head -1)
        echo "   GPU: $GPU_INFO"
        HAS_GPU=true
    else
        echo "   GPU: Not detected"
        HAS_GPU=false
    fi
    
    echo ""
}

# Model recommendations based on system
recommend_models() {
    echo "🎯 Model Recommendations:"
    echo ""
    
    if [ "$RAM_GB" -lt 4 ]; then
        print_warning "Low RAM detected. Recommending ultra-lightweight models:"
        RECOMMENDED_MODELS=("qwen2:0.5b" "tinyllama:1.1b")
    elif [ "$RAM_GB" -lt 8 ]; then
        print_info "Moderate RAM. Recommending efficient models:"
        RECOMMENDED_MODELS=("qwen2:1.5b" "phi3:mini" "gemma2:2b")
    elif [ "$RAM_GB" -lt 16 ]; then
        print_info "Good RAM available. Recommending balanced models:"
        RECOMMENDED_MODELS=("mistral:7b" "qwen2:7b" "llama3.1:8b")
    else
        print_success "High RAM available. All models supported:"
        RECOMMENDED_MODELS=("llama3.1:8b" "mistral:7b" "qwen2:7b" "codellama:13b")
    fi
    
    # Display model table
    echo "┌─────────────────┬──────────┬─────────┬─────────────────────────────┐"
    echo "│ Model           │ Size     │ RAM     │ Best For                    │"
    echo "├─────────────────┼──────────┼─────────┼─────────────────────────────┤"
    
    for model in "${RECOMMENDED_MODELS[@]}"; do
        case $model in
            "qwen2:0.5b")
                echo "│ qwen2:0.5b      │ 352MB    │ 1GB     │ Ultra-light, basic tasks    │"
                ;;
            "tinyllama:1.1b")
                echo "│ tinyllama:1.1b  │ 637MB    │ 2GB     │ Minimal resources           │"
                ;;
            "qwen2:1.5b")
                echo "│ qwen2:1.5b      │ 934MB    │ 2GB     │ Development, fast testing   │"
                ;;
            "phi3:mini")
                echo "│ phi3:mini       │ 2.3GB    │ 4GB     │ Efficient, good accuracy    │"
                ;;
            "gemma2:2b")
                echo "│ gemma2:2b       │ 1.6GB    │ 3GB     │ Compact, well-rounded       │"
                ;;
            "mistral:7b")
                echo "│ mistral:7b      │ 4.1GB    │ 6GB     │ Best overall (RECOMMENDED)  │"
                ;;
            "qwen2:7b")
                echo "│ qwen2:7b        │ 4.4GB    │ 7GB     │ Multilingual support        │"
                ;;
            "llama3.1:8b")
                echo "│ llama3.1:8b     │ 4.7GB    │ 8GB     │ Highest accuracy            │"
                ;;
            "codellama:13b")
                echo "│ codellama:13b   │ 7.3GB    │ 12GB    │ Code analysis               │"
                ;;
        esac
    done
    
    echo "└─────────────────┴──────────┴─────────┴─────────────────────────────┘"
    echo ""
}

# Interactive model selection
select_model() {
    echo "📥 Model Selection:"
    echo ""
    echo "Choose a model to download:"
    
    # Add option for custom model
    local options=("${RECOMMENDED_MODELS[@]}" "custom" "skip")
    
    for i in "${!options[@]}"; do
        case "${options[$i]}" in
            "mistral:7b")
                echo "  $((i+1)). ${options[$i]} (⭐ RECOMMENDED)"
                ;;
            "custom")
                echo "  $((i+1)). Enter custom model name"
                ;;
            "skip")
                echo "  $((i+1)). Skip model download"
                ;;
            *)
                echo "  $((i+1)). ${options[$i]}"
                ;;
        esac
    done
    
    echo ""
    read -p "Enter your choice (1-${#options[@]}): " choice
    
    if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#options[@]}" ]; then
        SELECTED_MODEL="${options[$((choice-1))]}"
        
        if [ "$SELECTED_MODEL" = "custom" ]; then
            echo ""
            read -p "Enter custom model name (e.g., mistral:7b): " SELECTED_MODEL
        elif [ "$SELECTED_MODEL" = "skip" ]; then
            print_info "Skipping model download"
            return 1
        fi
        
        echo ""
        print_info "Selected model: $SELECTED_MODEL"
        return 0
    else
        print_error "Invalid choice. Please try again."
        select_model
    fi
}

# Download model
download_model() {
    local model="$1"
    
    print_info "Downloading model: $model"
    echo "This may take several minutes depending on your internet connection..."
    echo ""
    
    # Use curl to pull the model
    if curl -X POST "$OLLAMA_URL/api/pull" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"$model\"}" \
        --fail-with-body; then
        print_success "Model $model downloaded successfully!"
        return 0
    else
        print_error "Failed to download model $model"
        return 1
    fi
}

# List available models
list_models() {
    print_info "Checking available models..."
    echo ""
    
    if MODELS=$(curl -s "$OLLAMA_URL/api/tags" | grep -o '"name":"[^"]*"' | cut -d'"' -f4); then
        if [ -n "$MODELS" ]; then
            echo "📋 Available models:"
            echo "$MODELS" | while read -r model; do
                echo "  • $model"
            done
        else
            print_warning "No models currently available"
        fi
    else
        print_error "Failed to retrieve model list"
    fi
    echo ""
}

# Update environment configuration
update_env_config() {
    local model="$1"
    local env_file="$PROJECT_ROOT/.env"
    
    if [ ! -f "$env_file" ]; then
        print_warning ".env file not found. Creating from template..."
        if [ -f "$PROJECT_ROOT/.env.example" ]; then
            cp "$PROJECT_ROOT/.env.example" "$env_file"
            print_success "Created .env from template"
        else
            print_error ".env.example not found. Cannot create .env file."
            return 1
        fi
    fi
    
    # Update OLLAMA_MODEL in .env file
    if grep -q "^OLLAMA_MODEL=" "$env_file"; then
        # Update existing line
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/^OLLAMA_MODEL=.*/OLLAMA_MODEL=$model/" "$env_file"
        else
            sed -i "s/^OLLAMA_MODEL=.*/OLLAMA_MODEL=$model/" "$env_file"
        fi
        print_success "Updated OLLAMA_MODEL in .env file"
    else
        # Add new line
        echo "OLLAMA_MODEL=$model" >> "$env_file"
        print_success "Added OLLAMA_MODEL to .env file"
    fi
    
    # Ensure local LLM is enabled
    if grep -q "^ENABLE_LOCAL_LLM=" "$env_file"; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/^ENABLE_LOCAL_LLM=.*/ENABLE_LOCAL_LLM=true/" "$env_file"
        else
            sed -i "s/^ENABLE_LOCAL_LLM=.*/ENABLE_LOCAL_LLM=true/" "$env_file"
        fi
    else
        echo "ENABLE_LOCAL_LLM=true" >> "$env_file"
    fi
}

# Test model
test_model() {
    local model="$1"
    
    print_info "Testing model: $model"
    echo ""
    
    # Test with a simple prompt
    local test_prompt="Explain what FFprobe is in one sentence."
    
    if response=$(curl -s -X POST "$OLLAMA_URL/api/generate" \
        -H "Content-Type: application/json" \
        -d "{\"model\": \"$model\", \"prompt\": \"$test_prompt\", \"stream\": false}" \
        | grep -o '"response":"[^"]*"' | cut -d'"' -f4); then
        
        if [ -n "$response" ]; then
            print_success "Model test successful!"
            echo "Response: $response"
        else
            print_warning "Model responded but with empty response"
        fi
    else
        print_error "Model test failed"
        return 1
    fi
    echo ""
}

# Main execution
main() {
    echo "Starting Ollama setup for FFprobe API..."
    echo ""
    
    # Step 1: Check Ollama
    if ! check_ollama; then
        exit 1
    fi
    
    # Step 2: System info
    get_system_info
    
    # Step 3: Show current models
    list_models
    
    # Step 4: Recommend models
    recommend_models
    
    # Step 5: Model selection
    if select_model; then
        echo ""
        print_info "Proceeding with model: $SELECTED_MODEL"
        echo ""
        
        # Step 6: Download model
        if download_model "$SELECTED_MODEL"; then
            # Step 7: Update configuration
            update_env_config "$SELECTED_MODEL"
            
            # Step 8: Test model
            test_model "$SELECTED_MODEL"
            
            echo ""
            print_success "Ollama setup completed successfully!"
            echo ""
            echo "🚀 Next steps:"
            echo "  1. Restart FFprobe API: docker compose restart ffprobe-api"
            echo "  2. Test AI features: curl http://localhost:8080/api/v1/genai/health"
            echo "  3. Read the Local LLM guide: docs/tutorials/local-llm-setup.md"
            echo ""
            print_info "Your local AI-powered media analysis is ready!"
            
        else
            print_error "Model download failed. Please check your internet connection and try again."
            exit 1
        fi
    else
        print_info "Setup completed without downloading a model."
        echo ""
        echo "You can run this script again later to download models:"
        echo "  ./scripts/setup/setup-ollama.sh"
    fi
}

# Run main function
main "$@"