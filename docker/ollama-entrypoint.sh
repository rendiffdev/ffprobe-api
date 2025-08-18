#!/bin/bash
set -e

echo "🦙 Starting Ollama service..."

# Start ollama service in background
ollama serve &
OLLAMA_PID=$!

echo "⏳ Waiting for Ollama to be ready..."
# Wait for ollama to be ready
sleep 10

# Function to check if ollama is ready
wait_for_ollama() {
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:11434/api/version > /dev/null 2>&1; then
            echo "✅ Ollama is ready!"
            return 0
        fi
        echo "⏳ Attempt $attempt/$max_attempts: Waiting for Ollama..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ Ollama failed to start after $max_attempts attempts"
    return 1
}

# Wait for ollama to be ready
if wait_for_ollama; then
    echo "🎯 Ollama is running, checking for models..."
    
    # Primary model (Gemma 3 270M - fast and efficient)
    MODEL_NAME="${OLLAMA_MODEL:-gemma3:270m}"
    echo "🔍 Checking for primary model: $MODEL_NAME"
    
    if ! ollama list | grep -q "$MODEL_NAME"; then
        echo "📥 Primary model $MODEL_NAME not found. Downloading..."
        echo "⚠️  This is a small model (~200MB) and should download quickly..."
        
        # Download the primary model
        if ollama pull "$MODEL_NAME"; then
            echo "✅ Primary model $MODEL_NAME downloaded successfully!"
        else
            echo "⚠️  Failed to download primary model $MODEL_NAME"
            echo "⚠️  Will continue with fallback model only"
        fi
    else
        echo "✅ Primary model $MODEL_NAME already available!"
    fi
    
    # Fallback model (Phi-3 Mini - better reasoning)
    FALLBACK_MODEL="${OLLAMA_FALLBACK_MODEL:-phi3:mini}"
    echo "🔍 Checking for fallback model: $FALLBACK_MODEL"
    
    if ! ollama list | grep -q "$FALLBACK_MODEL"; then
        echo "📥 Fallback model $FALLBACK_MODEL not found. Downloading..."
        echo "⚠️  This model is larger (~2GB) and may take a few minutes..."
        
        # Download the fallback model
        if ollama pull "$FALLBACK_MODEL"; then
            echo "✅ Fallback model $FALLBACK_MODEL downloaded successfully!"
        else
            echo "⚠️  Failed to download fallback model $FALLBACK_MODEL"
            echo "⚠️  The system will work with available models only"
        fi
    else
        echo "✅ Fallback model $FALLBACK_MODEL already available!"
    fi
    
    echo "📋 Available models:"
    ollama list
    
    echo "🎬 Ollama setup complete! Models are ready for FFprobe API."
else
    echo "❌ Failed to start Ollama service"
    exit 1
fi

# Keep the main process running
echo "🔄 Keeping Ollama service running..."
wait $OLLAMA_PID