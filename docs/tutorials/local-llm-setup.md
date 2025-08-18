# 🤖 Local AI Setup Guide

**Zero-configuration AI setup** - Everything runs automatically in Docker containers.

## 🎯 Dual-Model Architecture

The FFprobe API uses an optimized dual-model AI approach:
- **Primary**: **Gemma 3 270M** (ultra-fast, 270M params, ~200MB RAM) via Ollama
- **Local Fallback**: **Phi-3 Mini** (better reasoning, 3.8B params, 2GB RAM) via Ollama
- **Cloud Fallback**: **OpenRouter API** (optional) for enhanced analysis
- **Zero Setup**: All services containerized and auto-configured

## ✨ What You Get Out of the Box

### 🐳 **Fully Containerized Services**
```bash
docker compose -f docker-image/compose.yaml up -d
# That's it! Everything is configured automatically:
# ✅ Ollama with Gemma 3 270M + Phi-3 Mini models
# ✅ SQLite/PostgreSQL database  
# ✅ Valkey caching
# ✅ FFmpeg/FFprobe integration
# ✅ API server with authentication
```

### 🧠 **AI Processing**
- **Gemma 3 270M**: Google's ultra-efficient model (10x faster, 90% less RAM)
- **Phi-3 Mini**: Microsoft's 3.8B parameter model for complex analysis
- **Professional Analysis**: 8-section video engineering reports
- **Private Processing**: No data leaves your infrastructure
- **Intelligent Fallback**: Automatic model selection based on task complexity

## 🚀 Quick Start (Zero Configuration)

### 1. **Start Everything** 
```bash
# Clone and start - that's all!
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# Copy environment (optional customization)
cp .env.example .env

# Start all services (downloads models automatically)
docker compose up -d

# Verify everything is running
curl http://localhost:8080/health
```

### 2. **First Analysis with AI**
```bash
# Generate API key
export API_KEY="ffprobe_test_sk_$(openssl rand -hex 32)"
echo "API_KEY=$API_KEY" >> .env
docker compose restart

# Analyze video with AI insights
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@your-video.mp4" \
  -F "include_llm=true"
```

**That's it!** No model downloads, no configuration, no setup required.

## ⚙️ Architecture Details

### **Container Services**
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   FFprobe API   │───▶│   PostgreSQL    │    │     Redis       │
│   (Port 8080)   │    │   (Database)    │    │   (Caching)     │
└─────────┬───────┘    └─────────────────┘    └─────────────────┘
          │
          ▼
┌─────────────────┐    ┌─────────────────┐
│     Ollama      │    │   FFmpeg        │
│  (Phi-3 Mini)   │    │  (Processing)   │
│  (Port 11434)   │    │   (Built-in)    │
└─────────────────┘    └─────────────────┘
```

### **AI Processing Flow**
1. **Video Upload** → API receives file
2. **FFprobe Analysis** → Extract technical metadata  
3. **Local AI Analysis** → Phi-3 Mini generates insights
4. **Fallback (Optional)** → OpenRouter if local fails
5. **Return Results** → Complete analysis with AI report

## 🔧 Configuration Options

### **Default Configuration (Recommended)**
```bash
# .env file - default works for most users
ENABLE_LOCAL_LLM=true
OLLAMA_MODEL=gemma3:270m            # Primary: Ultra-fast for structured reports
OLLAMA_FALLBACK_MODEL=phi3:mini     # Fallback: Better for complex reasoning
OLLAMA_URL=http://ollama:11434      # Container networking
ENABLE_OPENROUTER_FALLBACK=false   # Optional cloud fallback
```

### **Model Selection Strategy**
```bash
# Primary Model (Gemma 3 270M) - Used for:
# - Standard FFprobe report generation
# - Structured data conversion
# - Quick quality assessments
# Response time: 0.5-2 seconds

# Fallback Model (Phi-3 Mini) - Automatically used for:
# - Complex FFmpeg command generation
# - Detailed technical explanations
# - Edge cases and errors in primary model
# Response time: 5-10 seconds

# Cloud Fallback (Optional)
ENABLE_OPENROUTER_FALLBACK=true
OPENROUTER_API_KEY=sk-or-your-key
```

### **Resource Limits (Customize if needed)**
```yaml
# docker compose.yml - adjust if you have different hardware
ollama:
  deploy:
    resources:
      limits:
        memory: 3G        # Gemma3 (200MB) + Phi-3 (2GB) + overhead
        cpus: '2.0'       # 2 cores recommended
```

## 📊 System Requirements

### **Minimum Requirements**
- **RAM**: 4GB total (200MB Gemma3 + 2GB Phi-3 + 1.8GB services)
- **Storage**: 3GB (200MB Gemma3 + 2GB Phi-3 + containers)
- **CPU**: 2 cores
- **Docker**: 24.0+ with Compose v2

### **Recommended Requirements**
- **RAM**: 6GB+ (smoother performance)
- **Storage**: 5GB+ (room for analysis data)
- **CPU**: 4+ cores (faster processing)

### **Optional GPU Support**
```yaml
# Uncomment in docker compose.yml for GPU acceleration
ollama:
  runtime: nvidia
  environment:
    - NVIDIA_VISIBLE_DEVICES=all
```

## 🧪 Verify Setup

### **Health Checks**
```bash
# Check all services
docker compose ps

# Expected output:
# ffprobe-api     Up (healthy)
# postgres        Up (healthy)  
# redis           Up (healthy)
# ollama          Up (healthy)

# Check AI service specifically
curl http://localhost:8080/api/v1/genai/health
```

### **Test AI Analysis**
```bash
# Create test video
docker run --rm -v $(pwd):/work jrottenberg/ffmpeg:4.4-alpine \
  -f lavfi -i testsrc=duration=10:size=320x240:rate=30 \
  -c:v libx264 /work/test.mp4

# Analyze with AI
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@test.mp4" \
  -F "include_llm=true" | jq '.llm_report'
```

## 🔍 What Happens Automatically

### **On First Startup**
1. **Container Build**: All services start automatically
2. **Model Downloads**: 
   - Gemma 3 270M (~200MB, downloads in ~1 minute)
   - Phi-3 Mini (~2GB, downloads in ~5 minutes)
3. **Database Setup**: PostgreSQL initializes with schemas
4. **Service Health**: All services wait for dependencies
5. **Ready State**: API becomes available

### **Model Management**
- **Auto-Download**: Both models download on first run
- **Persistent Storage**: Models cached in `./data/ollama/`
- **Intelligent Selection**: Automatic model choice based on task
- **Health Monitoring**: Both models' availability checked continuously
- **Graceful Degradation**: System works even if one model fails

## 🚨 Troubleshooting

### **Service Won't Start**
```bash
# Check logs
docker compose logs ollama
docker compose logs ffprobe-api

# Common issues:
# - Insufficient RAM (need 4GB minimum)
# - Port conflicts (change ports in .env)
# - Docker daemon not running
```

### **Model Download Issues**
```bash
# Check disk space
df -h

# Manual model downloads
docker compose exec ollama ollama pull gemma3:270m  # Primary (fast)
docker compose exec ollama ollama pull phi3:mini    # Fallback (smart)

# Reset everything if needed
docker compose down -v
docker compose up -d
```

### **AI Analysis Not Working**
```bash
# Check model status
curl http://localhost:11434/api/tags

# Verify API can reach Ollama
docker compose exec ffprobe-api curl http://ollama:11434/api/version

# Check environment variables
docker compose exec ffprobe-api env | grep OLLAMA
```

## 🎯 Production Considerations

### **For Production Deployment**
```bash
# Use production Dockerfile
docker build -f Dockerfile.production -t ffprobe-api:prod .

# Production environment
cp .env.example .env.production
# Configure with production values

# Deploy with production script
./scripts/deployment/production-deploy.sh
```

### **Resource Planning**
- **Development**: 4GB RAM, 2 cores (both models)
- **Light Production**: 4GB RAM, 4 cores (Gemma3 only)
- **Heavy Production**: 6GB+ RAM, 4+ cores (both models)
- **Scale Horizontally**: Multiple API containers, shared Ollama

### **Performance Comparison**
| Workload | Gemma 3 270M | Phi-3 Mini | Recommendation |
|----------|--------------|------------|----------------|
| Simple Reports | 0.5-2s | 5-10s | Use Gemma3 |
| Complex Analysis | 2-3s (limited) | 8-12s | Use Phi-3 |
| Concurrent Requests | 20-30 | 3-5 | Use Gemma3 |
| Memory per Request | ~50MB | ~500MB | Use Gemma3 |

## ✅ Benefits of This Approach

### **Developer Experience**
- **Zero Setup**: `docker compose up -d` and you're ready
- **No Model Hunting**: Single proven model (Phi-3 Mini)
- **No Configuration**: Sensible defaults work out of the box
- **Consistent Environment**: Same setup for dev/staging/prod

### **Operational Benefits**
- **Private AI**: No data sent to external services
- **Cost Effective**: No LLM API costs
- **Reliable**: No external API dependencies
- **Fast**: Local processing, no network latency

### **Production Ready**
- **Containerized**: Easy deployment and scaling
- **Monitored**: Health checks and metrics included
- **Fallback**: Optional cloud API for enhanced features
- **Secure**: Complete data privacy and control

---

## 🎬 You're Ready!

Your zero-configuration AI-powered video analysis system is ready:

1. **Start**: `docker compose up -d`
2. **Wait**: ~2 minutes for model download
3. **Analyze**: Upload videos and get AI insights
4. **Scale**: Add more containers as needed

**Need help?** Check [Troubleshooting Guide](../TROUBLESHOOTING.md) or create an issue.