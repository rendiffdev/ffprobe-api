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
    
    # Check if the model exists
    MODEL_NAME="${OLLAMA_MODEL:-phi3:mini}"
    echo "🔍 Checking for model: $MODEL_NAME"
    
    if ! ollama list | grep -q "$MODEL_NAME"; then
        echo "📥 Model $MODEL_NAME not found. Downloading..."
        echo "⚠️  This may take several minutes depending on your internet connection..."
        
        # Download the model
        if ollama pull "$MODEL_NAME"; then
            echo "✅ Model $MODEL_NAME downloaded successfully!"
        else
            echo "❌ Failed to download model $MODEL_NAME"
            echo "🔄 Trying to download a smaller fallback model..."
            
            # Try phi3:mini as fallback
            echo "📥 Trying fallback model: phi3:mini"
            if ollama pull "phi3:mini"; then
                echo "✅ Fallback model phi3:mini downloaded successfully!"
                echo "⚙️  Using phi3:mini as the default model"
            else
                echo "❌ Failed to download fallback model phi3:mini"
                exit 1
            fi
        fi
    else
        echo "✅ Model $MODEL_NAME already available!"
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