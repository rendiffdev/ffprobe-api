# 🚀 FFprobe API - One-Step Deployment

**Ready-to-use Docker image with zero configuration required!**

[![Docker Hub](https://img.shields.io/docker/v/rendiffdev/ffprobe-api?label=Docker%20Hub)](https://hub.docker.com/r/rendiffdev/ffprobe-api)
[![Image Size](https://img.shields.io/docker/image-size/rendiffdev/ffprobe-api/latest)](https://hub.docker.com/r/rendiffdev/ffprobe-api)

## 🎯 Instant Deployment (30 seconds)

### Option 1: Single Command (Recommended)
```bash
# One command - starts everything automatically
docker run -d \
  --name ffprobe-api \
  -p 8080:8080 \
  -v ffprobe_data:/app/data \
  -v ffprobe_uploads:/app/uploads \
  rendiffdev/ffprobe-api:latest

# Test immediately
curl http://localhost:8080/health
```

### Option 2: Full Stack with AI (Docker Compose)
```bash
# Download ready-to-use compose file
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/compose.yml

# Start everything (auto-downloads AI models)
docker compose up -d

# API ready at http://localhost:8080
curl http://localhost:8080/health
```

## ✅ What You Get Out of the Box

- **🎯 Zero Configuration**: Works immediately without setup
- **📊 SQLite Database**: Embedded, no external DB needed
- **🚀 Valkey Cache**: Redis-compatible caching (optional)
- **🧠 AI Analysis**: Auto-downloads Gemma3 & Phi3 models
- **🔍 20+ QC Categories**: Professional quality control
- **📈 Performance Optimized**: 8 workers, 20GB file support
- **🛡️ Rate Limited**: 100 req/min protection
- **💾 Persistent Storage**: Data survives restarts

## 📋 System Requirements

- **RAM**: 4GB minimum, 8GB recommended (with AI)
- **Storage**: 10GB free space (for AI models)
- **Docker**: 24.0+ with Compose v2.20+
- **Ports**: 8080 (API), optional 6379 (cache), 11434 (AI)

## 🔧 Advanced Usage

### Production Deployment
```bash
# Production-ready with SSL and monitoring
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/compose.prod.yml

# Configure your domain
echo "DOMAIN=api.yourdomain.com" > .env
echo "ACME_EMAIL=admin@yourdomain.com" >> .env

# Deploy with SSL
docker compose -f compose.prod.yml --profile production up -d
```

### Custom Configuration (Optional)
```bash
# Download environment template
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/.env.example

# Customize settings (optional)
cp .env.example .env
# Edit .env with your preferences

# Deploy with custom config
docker compose up -d
```

### High Performance Mode
```bash
# Optimized for heavy workloads
docker run -d \
  --name ffprobe-api-hp \
  -p 8080:8080 \
  -e WORKER_POOL_SIZE=16 \
  -e MAX_FILE_SIZE=107374182400 \
  -e PROCESSING_TIMEOUT=1200 \
  --memory=8g \
  --cpus=4 \
  rendiffdev/ffprobe-api:latest
```

## 🧪 Test Your Deployment

### 1. Health Check
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","service":"ffprobe-api"}
```

### 2. Basic Analysis
```bash
# Analyze a sample video URL
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url":"https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4"}' \
  http://localhost:8080/api/v1/probe/url
```

### 3. AI Analysis (if enabled)
```bash
# Wait for AI models to download (~5 minutes first time)
docker compose logs ollama

# Test AI-powered analysis
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url":"https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4", "include_llm":true}' \
  http://localhost:8080/api/v1/probe/url
```

## 🔍 Monitoring & Logs

### View Logs
```bash
# API logs
docker compose logs -f ffprobe-api

# AI service logs
docker compose logs -f ollama

# All services
docker compose logs -f
```

### Service Status
```bash
# Check all containers
docker compose ps

# Resource usage
docker stats
```

## 🛠️ Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Use different port
docker run -d -p 8081:8080 rendiffdev/ffprobe-api:latest
```

#### Out of Memory (with AI)
```bash
# Disable AI for lower memory usage
docker run -d \
  -p 8080:8080 \
  -e ENABLE_LOCAL_LLM=false \
  rendiffdev/ffprobe-api:latest
```

#### AI Models Download Failed
```bash
# Manual model download
docker compose exec ollama ollama pull gemma3:270m
docker compose exec ollama ollama pull phi3:mini
```

### Performance Tuning
```bash
# Check current settings
curl http://localhost:8080/api/v1/system/stats

# Adjust worker pool
docker run -d \
  -p 8080:8080 \
  -e WORKER_POOL_SIZE=12 \
  rendiffdev/ffprobe-api:latest
```

## 📚 Next Steps

1. **[API Documentation](../docs/api/README.md)** - Complete API reference
2. **[Quality Control Features](../QC_ANALYSIS_LIST.md)** - All 20+ QC categories
3. **[Production Deployment](DEPLOYMENT_GUIDE.md)** - SSL, monitoring, scaling
4. **[Development Setup](../README.md)** - Build from source

## 🆘 Support

- **Docker Hub**: [rendiffdev/ffprobe-api](https://hub.docker.com/r/rendiffdev/ffprobe-api)
- **GitHub Issues**: [Report problems](https://github.com/rendiffdev/ffprobe-api/issues)
- **Documentation**: [Full docs](../README.md)

---

**🎉 Your FFprobe API is ready! Start analyzing videos immediately at `http://localhost:8080`**