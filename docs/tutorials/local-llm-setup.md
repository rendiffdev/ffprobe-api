# 🤖 Local AI Setup Guide

**Zero-configuration AI setup** - Everything runs automatically in Docker containers.

## 🎯 Simple Architecture

The FFprobe API uses a streamlined AI approach:
- **Primary**: **Phi-3 Mini** (local, private, 2GB RAM) via Ollama
- **Fallback**: **OpenRouter API** (cloud, optional) for enhanced analysis
- **Zero Setup**: All services containerized and auto-configured

## ✨ What You Get Out of the Box

### 🐳 **Fully Containerized Services**
```bash
docker-compose up -d
# That's it! Everything is configured automatically:
# ✅ Ollama with Phi-3 Mini model
# ✅ PostgreSQL database  
# ✅ Redis caching
# ✅ FFmpeg/FFprobe workers
# ✅ API server with authentication
```

### 🧠 **AI Processing**
- **Phi-3 Mini**: Microsoft's efficient 3.8B parameter model (2GB RAM)
- **Professional Analysis**: 8-section video engineering reports
- **Private Processing**: No data leaves your infrastructure
- **Smart Fallback**: OpenRouter API when local LLM is unavailable

## 🚀 Quick Start (Zero Configuration)

### 1. **Start Everything** 
```bash
# Clone and start - that's all!
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# Copy environment (optional customization)
cp .env.example .env

# Start all services (downloads models automatically)
docker-compose up -d

# Verify everything is running
curl http://localhost:8080/health
```

### 2. **First Analysis with AI**
```bash
# Generate API key
export API_KEY="ffprobe_test_sk_$(openssl rand -hex 32)"
echo "API_KEY=$API_KEY" >> .env
docker-compose restart

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
OLLAMA_MODEL=phi3:mini              # Only supported model
OLLAMA_URL=http://ollama:11434      # Container networking
ENABLE_OPENROUTER_FALLBACK=false   # Optional cloud fallback
```

### **Cloud Fallback (Optional)**
```bash
# Add OpenRouter API key for enhanced analysis
ENABLE_OPENROUTER_FALLBACK=true
OPENROUTER_API_KEY=sk-or-your-key
OPENROUTER_MODEL=microsoft/phi-3-mini-128k-instruct
```

### **Resource Limits (Customize if needed)**
```yaml
# docker-compose.yml - adjust if you have different hardware
ollama:
  deploy:
    resources:
      limits:
        memory: 3G        # Phi-3 Mini needs ~2GB
        cpus: '2.0'       # 2 cores recommended
```

## 📊 System Requirements

### **Minimum Requirements**
- **RAM**: 4GB total (2GB for Phi-3 Mini + 2GB for other services)
- **Storage**: 3GB (2GB for model + 1GB for containers)
- **CPU**: 2 cores
- **Docker**: 24.0+ with Compose v2

### **Recommended Requirements**
- **RAM**: 6GB+ (smoother performance)
- **Storage**: 5GB+ (room for analysis data)
- **CPU**: 4+ cores (faster processing)

### **Optional GPU Support**
```yaml
# Uncomment in docker-compose.yml for GPU acceleration
ollama:
  runtime: nvidia
  environment:
    - NVIDIA_VISIBLE_DEVICES=all
```

## 🧪 Verify Setup

### **Health Checks**
```bash
# Check all services
docker-compose ps

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
2. **Model Download**: Phi-3 Mini downloads (~2GB, one-time)
3. **Database Setup**: PostgreSQL initializes with schemas
4. **Service Health**: All services wait for dependencies
5. **Ready State**: API becomes available

### **Model Management**
- **Auto-Download**: Phi-3 Mini downloads on first run
- **Persistent Storage**: Model cached in `./data/ollama/`
- **Version Management**: Model updates handled automatically
- **Health Monitoring**: Model availability checked continuously

## 🚨 Troubleshooting

### **Service Won't Start**
```bash
# Check logs
docker-compose logs ollama
docker-compose logs ffprobe-api

# Common issues:
# - Insufficient RAM (need 4GB minimum)
# - Port conflicts (change ports in .env)
# - Docker daemon not running
```

### **Model Download Issues**
```bash
# Check disk space
df -h

# Manual model download
docker-compose exec ollama ollama pull phi3:mini

# Reset everything if needed
docker-compose down -v
docker-compose up -d
```

### **AI Analysis Not Working**
```bash
# Check model status
curl http://localhost:11434/api/tags

# Verify API can reach Ollama
docker-compose exec ffprobe-api curl http://ollama:11434/api/version

# Check environment variables
docker-compose exec ffprobe-api env | grep OLLAMA
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
- **Development**: 4GB RAM, 2 cores
- **Light Production**: 6GB RAM, 4 cores  
- **Heavy Production**: 8GB+ RAM, 4+ cores
- **Scale Horizontally**: Multiple API containers, shared Ollama

## ✅ Benefits of This Approach

### **Developer Experience**
- **Zero Setup**: `docker-compose up -d` and you're ready
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

1. **Start**: `docker-compose up -d`
2. **Wait**: ~2 minutes for model download
3. **Analyze**: Upload videos and get AI insights
4. **Scale**: Add more containers as needed

**Need help?** Check [Troubleshooting Guide](../TROUBLESHOOTING.md) or create an issue.