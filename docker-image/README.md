# FFprobe API - Docker Images

Production-ready Docker images for the FFprobe API with multi-architecture support.

[![Docker Pulls](https://img.shields.io/docker/pulls/rendiffdev/ffprobe-api)](https://hub.docker.com/r/rendiffdev/ffprobe-api)
[![Docker Image Size](https://img.shields.io/docker/image-size/rendiffdev/ffprobe-api/latest)](https://hub.docker.com/r/rendiffdev/ffprobe-api)
[![Docker Image Version](https://img.shields.io/docker/v/rendiffdev/ffprobe-api)](https://hub.docker.com/r/rendiffdev/ffprobe-api)

## Available Images on Docker Hub

All images are available at: **[rendiffdev/ffprobe-api](https://hub.docker.com/r/rendiffdev/ffprobe-api)**

### 🔥 **Recommended: Hybrid Production Image**
```bash
# Multi-architecture with enhanced features
docker run -d -p 8080:8080 rendiffdev/ffprobe-api:hybrid
```

**Features:**
- ✅ Multi-architecture (AMD64 + ARM64)
- ✅ SQLite database with analysis history
- ✅ Rate limiting and request logging
- ✅ Enhanced API endpoints (health, stats, history)
- ✅ Security headers and validation
- ✅ Python Flask with advanced features

**Available tags:** `hybrid`, `production`, `latest`, `stable`, `full-featured`, `v1.0.0`

### 🚀 **Minimal Working Image**
```bash
# Simple Python Flask wrapper
docker run -d -p 8080:8080 rendiffdev/ffprobe-api:minimal
```

**Features:**
- ✅ Lightweight and fast
- ✅ Basic FFprobe functionality
- ✅ Production-ready
- ✅ AMD64 only

**Available tags:** `minimal`, `working`

## Architecture Support

| Image | AMD64 (Linux) | ARM64 (Mac Silicon) |
|-------|---------------|---------------------|
| `hybrid` | ✅ | ✅ |
| `minimal` | ✅ | ❌ |

## Quick Start

### For AMD64 (Linux servers, cloud)
```bash
docker run -d -p 8080:8080 --platform linux/amd64 rendiffdev/ffprobe-api:hybrid
```

### For ARM64 (Apple Silicon Macs)
```bash
docker run -d -p 8080:8080 --platform linux/arm64 rendiffdev/ffprobe-api:hybrid
```

### Auto-detect platform
```bash
docker run -d -p 8080:8080 rendiffdev/ffprobe-api:hybrid
```

## API Endpoints

- **Health Check**: `GET /health`
- **Video Analysis**: `POST /api/v1/probe` (with file upload)
- **Version Info**: `GET /api/v1/version`
- **Statistics**: `GET /api/v1/stats` (hybrid only)
- **History**: `GET /api/v1/history` (hybrid only)

## Environment Variables

```bash
docker run -d -p 8080:8080 \
  -e RATE_LIMIT_PER_MINUTE=200 \
  -e MAX_FILE_SIZE=2147483648 \
  -e LOG_LEVEL=debug \
  -v ./data:/app/data \
  rendiffdev/ffprobe-api:hybrid
```

## Building from Source

### Build Hybrid (Recommended)
```bash
./build-hybrid.sh v1.0.0
```

### Build Minimal
```bash
./build-minimal.sh
```

## Security Features

- 🔒 Non-root user
- 🔒 Security headers
- 🔒 Rate limiting
- 🔒 Input validation
- 🔒 No hardcoded secrets

## Production Deployment

```bash
# With persistent data
docker run -d --name ffprobe-api \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./uploads:/app/uploads \
  -v ./reports:/app/reports \
  --restart unless-stopped \
  rendiffdev/ffprobe-api:hybrid

# Test the deployment
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/v1/probe -F "file=@video.mp4"
```

## Files in this Directory

- `Dockerfile.hybrid` - Multi-architecture production image (recommended)
- `Dockerfile.minimal` - Simple working image
- `build-hybrid.sh` - Build script for hybrid image
- `build-minimal.sh` - Build script for minimal image

---

**Ready for production use!** 🚀