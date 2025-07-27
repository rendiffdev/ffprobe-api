# 🤖 Local LLM Setup Guide

Complete guide for setting up local Large Language Models (LLMs) with FFprobe API using Ollama.

## 🎯 Overview

The FFprobe API integrates with **Ollama** to provide local AI-powered media analysis, enabling:
- **Private AI processing** - No data sent to external services
- **Cost-effective analysis** - No API costs for LLM usage
- **Fast responses** - Local processing eliminates network latency
- **Offline capability** - Works without internet connection

## 🦙 What is Ollama?

[Ollama](https://ollama.ai) is a powerful tool that makes it easy to run large language models locally. It handles model management, provides a simple API, and optimizes performance for local hardware.

## 📋 Prerequisites

### System Requirements
- **RAM**: 4GB minimum (8GB+ recommended)
- **Storage**: 10GB+ free space for models
- **CPU**: 4+ cores recommended
- **Docker**: Version 24.0+ with Compose v2

### Optional (Recommended)
- **GPU**: NVIDIA GPU with 4GB+ VRAM for faster inference
- **Docker with GPU support** for accelerated processing

## 🚀 Quick Setup

### 1. **Interactive Installation** (Recommended)
```bash
# Run the interactive installer
make install

# Follow the prompts and select:
# - Enable Local LLM: Yes
# - LLM Model: mistral:7b (or smaller for development)
# - GPU Support: Yes/No based on your setup
```

### 2. **Manual Configuration**
```bash
# Copy environment template
cp .env.example .env

# Edit configuration
nano .env

# Enable local LLM
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=mistral:7b

# Start services
docker compose up -d
```

## 🎯 Model Selection Guide

### **Recommended Models by Use Case**

| Model | Size | RAM | Use Case | Performance |
|-------|------|-----|----------|-------------|
| **mistral:7b** | 4.1GB | 6GB | **Best overall** - Balanced performance | ⭐⭐⭐⭐⭐ |
| **qwen2:1.5b** | 934MB | 2GB | **Development** - Fast, lightweight | ⭐⭐⭐⭐ |
| **phi3:mini** | 2.3GB | 4GB | **Production lite** - Efficient, accurate | ⭐⭐⭐⭐ |
| **gemma2:2b** | 1.6GB | 3GB | **Compact** - Good balance | ⭐⭐⭐⭐ |

### **Advanced Models** (More powerful, higher resource requirements)
| Model | Size | RAM | Use Case | Performance |
|-------|------|-----|----------|-------------|
| **llama3.1:8b** | 4.7GB | 8GB | **High accuracy** - Best quality | ⭐⭐⭐⭐⭐ |
| **qwen2:7b** | 4.4GB | 7GB | **Multilingual** - International support | ⭐⭐⭐⭐⭐ |

### **Model Selection Criteria**

**For Development:**
```bash
OLLAMA_MODEL=qwen2:1.5b  # Fast downloads, quick testing
```

**For Production:**
```bash
OLLAMA_MODEL=mistral:7b  # Best balance of quality and speed
```

**For High-End Servers:**
```bash
OLLAMA_MODEL=llama3.1:8b  # Maximum accuracy
```

## ⚙️ Configuration Options

### Environment Variables

```bash
# Core LLM Settings
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=mistral:7b

# Performance Tuning
OLLAMA_MAX_LOADED_MODELS=2
OLLAMA_NUM_PARALLEL=4
OLLAMA_MAX_QUEUE=128
OLLAMA_KEEP_ALIVE=24h

# Fallback Configuration
OPENROUTER_API_KEY=sk-or-xxx  # Optional cloud fallback
```

### Docker Resource Limits

**For Development:**
```yaml
ollama:
  deploy:
    resources:
      limits:
        memory: 3G
        cpus: '2.0'
```

**For Production:**
```yaml
ollama:
  deploy:
    resources:
      limits:
        memory: 8G
        cpus: '4.0'
```

## 🖥️ GPU Support (Optional)

### NVIDIA GPU Setup

1. **Install NVIDIA Docker Support:**
```bash
# Install nvidia-docker2
sudo apt update
sudo apt install nvidia-docker2
sudo systemctl restart docker
```

2. **Enable GPU in Compose:**
```yaml
ollama:
  runtime: nvidia
  environment:
    - NVIDIA_VISIBLE_DEVICES=all
  deploy:
    resources:
      reservations:
        devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
```

3. **Verify GPU Usage:**
```bash
# Check GPU usage during inference
docker exec ffprobe-ollama nvidia-smi
```

## 🔧 Model Management

### Download Models

**Automatic Download:**
Models are downloaded automatically on first startup via the entrypoint script.

**Manual Download:**
```bash
# Download via API
curl -X POST http://localhost:8080/api/v1/genai/pull-model \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"model": "mistral:7b"}'

# Download via Ollama CLI
docker exec ffprobe-ollama ollama pull mistral:7b
```

**List Available Models:**
```bash
# Via API
curl http://localhost:8080/api/v1/genai/health

# Via Ollama CLI
docker exec ffprobe-ollama ollama list
```

### Model Storage

Models are stored in the persistent volume:
```bash
# Check model storage
ls -la ./data/ollama/models/
```

## 🧪 Testing Local LLM

### Health Check
```bash
# Check LLM service health
curl http://localhost:8080/api/v1/genai/health

# Expected response:
{
  \"overall_status\": \"healthy\",
  \"services\": {
    \"ollama\": {
      \"healthy\": true,
      \"models\": [\"mistral:7b\"],
      \"configured_model\": \"mistral:7b\",
      \"model_available\": true
    }
  }
}
```

### Test AI Analysis
```bash
# Analyze a video file
curl -X POST http://localhost:8080/api/v1/ask \
  -H \"Content-Type: application/json\" \
  -H \"X-API-Key: your-api-key\" \
  -d '{
    \"source\": \"https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4\",
    \"question\": \"What are the key characteristics of this video?\"
  }'
```

### Performance Testing
```bash
# Time a request to measure local LLM performance
time curl -X POST http://localhost:8080/api/v1/genai/analysis \
  -H \"Content-Type: application/json\" \
  -H \"X-API-Key: your-api-key\" \
  -d '{\"analysis_id\": \"your-analysis-id\"}'
```

## 📊 Performance Optimization

### Memory Optimization
```bash
# Adjust model keep-alive time
OLLAMA_KEEP_ALIVE=5m  # Unload models after 5 minutes of inactivity

# Limit concurrent models
OLLAMA_MAX_LOADED_MODELS=1
```

### CPU Optimization
```bash
# Adjust parallel processing
OLLAMA_NUM_PARALLEL=2  # Reduce for lower-end CPUs
```

### Request Batching
```bash
# Process multiple requests efficiently
OLLAMA_MAX_QUEUE=64  # Queue more requests
```

## 🔄 Fallback Configuration

Configure cloud LLM fallback for when local models are unavailable:

```bash
# Environment configuration
ENABLE_LOCAL_LLM=true
OPENROUTER_API_KEY=sk-or-your-key-here

# Service behavior:
# 1. Try local Ollama first
# 2. Fall back to OpenRouter if local fails
# 3. Return error if both fail
```

## 🐛 Troubleshooting

### Common Issues

**1. Model Download Fails**
```bash
# Check disk space
df -h ./data/ollama/

# Manual download
docker exec ffprobe-ollama ollama pull mistral:7b

# Try smaller model
OLLAMA_MODEL=qwen2:1.5b
```

**2. Out of Memory**
```bash
# Use smaller model
OLLAMA_MODEL=phi3:mini

# Increase swap space
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

**3. Slow Performance**
```bash
# Check resource usage
docker stats ffprobe-ollama

# Optimize settings
OLLAMA_NUM_PARALLEL=2
OLLAMA_MAX_LOADED_MODELS=1
```

**4. Connection Issues**
```bash
# Check Ollama service health
curl http://localhost:11434/api/version

# Check container logs
docker logs ffprobe-ollama

# Restart service
docker restart ffprobe-ollama
```

### Debugging Commands

```bash
# Container logs
docker logs ffprobe-ollama -f

# Resource usage
docker exec ffprobe-ollama ps aux
docker exec ffprobe-ollama free -h

# Test Ollama directly
docker exec ffprobe-ollama ollama run mistral:7b \"Hello, how are you?\"
```

## 📈 Monitoring

### Resource Monitoring
```bash
# Monitor GPU usage (if available)
watch -n 1 nvidia-smi

# Monitor CPU and memory
docker stats ffprobe-ollama

# Check model performance
curl http://localhost:8080/api/v1/genai/health | jq
```

### Performance Metrics
```bash
# Prometheus metrics endpoint
curl http://localhost:9090/metrics | grep ollama

# Grafana dashboard
# - Import provided dashboard: grafana/ollama-dashboard.json
# - View at: http://localhost:3000
```

## 🎬 Use Cases

### **1. Media Analysis**
```bash
curl -X POST http://localhost:8080/api/v1/genai/analysis \
  -H \"Content-Type: application/json\" \
  -H \"X-API-Key: your-api-key\" \
  -d '{\"analysis_id\": \"uuid-here\"}'
```

### **2. Interactive Q&A**
```bash
curl -X POST http://localhost:8080/api/v1/ask \
  -H \"Content-Type: application/json\" \
  -H \"X-API-Key: your-api-key\" \
  -d '{
    \"analysis_id\": \"uuid-here\",
    \"question\": \"What is the bitrate of this video?\"
  }'
```

### **3. Quality Insights**
```bash
curl http://localhost:8080/api/v1/genai/quality-insights/uuid-here \
  -H \"X-API-Key: your-api-key\"
```

## 🔒 Security Considerations

### **Data Privacy**
- ✅ All processing happens locally
- ✅ No data sent to external services
- ✅ Complete control over AI models

### **Network Security**
- ✅ Ollama runs in isolated Docker network
- ✅ No external network access required
- ✅ API endpoints protected by authentication

### **Resource Isolation**
- ✅ Resource limits prevent system overload
- ✅ Container security with non-root users
- ✅ Read-only filesystem where possible

## 📚 Additional Resources

- **Ollama Documentation**: https://ollama.ai/docs
- **Model Library**: https://ollama.ai/library
- **Hugging Face Models**: https://huggingface.co/models
- **FFprobe API Docs**: [../api/](../api/)

## 🆘 Support

For issues or questions:
- **GitHub Issues**: [Report Issues](https://github.com/your-org/ffprobe-api/issues)
- **Health Check**: `make health-check`
- **Logs**: `docker logs ffprobe-ollama`

---

**🎬 Ready to analyze media with local AI!** Your private, fast, and cost-effective LLM setup is complete! 🚀