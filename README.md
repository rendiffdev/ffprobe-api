# 🎬 FFprobe API

> **Professional video analysis API with AI-powered insights and enterprise scaling**

[![Production Ready](https://img.shields.io/badge/production-ready-green.svg)](#-deployment-options)
[![Security Score](https://img.shields.io/badge/security-96%2F100-brightgreen.svg)](#-security-features)
[![API Version](https://img.shields.io/badge/api-v1.0-blue.svg)](#-api-overview)
[![Docker](https://img.shields.io/badge/docker-optimized-blue.svg)](#-what's-included-out-of-the-box)
[![GitHub](https://img.shields.io/badge/GitHub-rendiffdev%2Fffprobe--api-blue?logo=github)](https://github.com/rendiffdev/ffprobe-api)
[![Website](https://img.shields.io/badge/Website-rendiff.dev-blue?logo=web)](https://rendiff.dev)

## 🎯 What This Solves

**Problem**: "I need professional video analysis and can't tell if my optimizations actually improved quality"  
**Solution**: Get comprehensive video analysis with AI insights and before/after comparison validation

### Key Capabilities
- 📹 **Complete Video Analysis**: Technical specs, quality metrics, compliance checking
- 🤖 **AI-Powered Insights**: Professional video engineering assessment with recommendations  
- 📊 **Video Comparison**: Compare original vs modified videos to validate improvements
- 🏗️ **Enterprise Ready**: Production-grade scaling, security, and monitoring

## 🚀 Installation Options

### 1. **Single Server** (Development & Small Scale)
**Perfect for**: Development, testing, small projects  
**Requirements**: 4GB RAM, 2 CPU cores

```bash
# One-command installation
curl -fsSL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/scripts/install-single-server.sh | bash

# Or manual
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api
./scripts/install-single-server.sh
```

**What you get**: All services in lightweight containers (~3.5GB total)

### 2. **Professional Installation** (Interactive Setup)
**Perfect for**: Production deployments, custom configurations

```bash
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api
./scripts/install.sh
```

**Features**: Interactive configuration, deployment type selection, monitoring options

### 3. **Enterprise Scaling** (High Availability)
**Perfect for**: Large scale, high availability, load balancing

```bash
# Enterprise deployment with scaling
docker compose -f compose.yml -f compose.enterprise.yml up -d

# Scale specific services
docker compose -f compose.yml -f compose.enterprise.yml up -d --scale ffprobe-api=3 --scale ffprobe-worker=5
```

## 🎬 Real-World Example: Video Optimization Workflow

**Common scenario**: "I optimized my video, but is it actually better?"

```bash
# 1. Analyze original video
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "file=@original-large-video.mp4" \
  > original_analysis.json

# 2. Your optimization process (external)
ffmpeg -i original-large-video.mp4 -c:v libx264 -crf 23 optimized-video.mp4

# 3. Analyze optimized video  
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "file=@optimized-video.mp4" \
  > optimized_analysis.json

# 4. Get AI-powered comparison verdict
ORIGINAL_ID=$(jq -r '.id' original_analysis.json)
OPTIMIZED_ID=$(jq -r '.id' optimized_analysis.json)

curl -X POST http://localhost:8080/api/v1/comparisons/quick \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"original_analysis_id\": \"$ORIGINAL_ID\",
    \"modified_analysis_id\": \"$OPTIMIZED_ID\",  
    \"include_llm\": true
  }"

# 5. Clear recommendation
# Response: "30% file size reduction with maintained quality → Accept changes ✅"
```

## ✨ Core Features

### 🎬 Professional Video Analysis
- **Complete FFprobe Integration**: All metadata, streams, formats, chapters
- **Quality Metrics**: VMAF, PSNR, SSIM analysis with Netflix-grade models
- **HLS/DASH Support**: Streaming protocol validation and optimization
- **Compliance Checking**: Broadcasting and platform standards validation
- **Batch Processing**: Handle multiple files efficiently

### 🤖 AI-Powered Insights (Zero Configuration)
- **Auto-Configured Local AI**: Phi-3 Mini model (2GB RAM) downloads automatically
- **No Setup Required**: Ollama service starts with `docker compose up -d`
- **Professional Analysis**: 8-section video engineering reports
- **Quality Recommendations**: AI-generated optimization suggestions
- **Smart Fallback**: Optional OpenRouter API when local LLM unavailable

### 📊 Video Comparison System
- **Before/After Validation**: Compare original vs modified videos objectively
- **Quality Improvement Tracking**: Determine if changes actually improved content
- **File Size Analysis**: Optimization impact on storage costs
- **AI-Powered Decisions**: Get clear accept/reject recommendations
- **Comparative Metrics**: Side-by-side technical and quality comparisons

### 🏗️ Production Architecture (Fully Containerized)
- **Zero Setup**: All services auto-configured (PostgreSQL, Redis, FFmpeg, Ollama)
- **Microservices Ready**: Scale API, FFprobe workers, and LLM service independently
- **Security Hardened**: 96/100 security score with comprehensive hardening
- **Monitoring Integrated**: Prometheus metrics, Grafana dashboards, health checks
- **Enterprise Features**: JWT/API key auth, rate limiting, audit logging

## 🏗️ Architecture

### Zero-Configuration Deployment (Default)
```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Client    │───▶│  FFprobe API │───▶│ PostgreSQL │
└─────────────┘    │ (w/ FFmpeg)  │    │(Auto-setup)│
                   └──────┬───────┘    └─────────────┘
                          │                  │
                   ┌──────▼──────┐    ┌──────▼──────┐
                   │   Ollama    │    │   Redis     │
                   │(Phi-3 Mini) │    │ (Caching)   │
                   │(Auto-DL'd)  │    │(Auto-setup)│
                   └─────────────┘    └─────────────┘
```

### Enterprise Scaling (Production)
```
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Client    │───▶│ Load Balancer│───▶│ FFprobe Workers │
└─────────────┘    │    (Nginx)   │    │   (Scalable)    │
                   └──────┬───────┘    └─────────────────┘
                          │                      │
                   ┌──────▼──────┐    ┌──────────▼──────┐
                   │ AI Workers  │    │ Database Cluster│
                   │ (Scalable)  │    │  (PostgreSQL)   │
                   └─────────────┘    └─────────────────┘
```

## 🎁 What's Included Out-of-the-Box

**All services automatically configured with `docker compose up -d`:**

### 🐳 **Core Services** (Zero Setup Required)
- **🎬 FFprobe API Server** - Main application with FFmpeg/FFprobe built-in
- **🐘 PostgreSQL Database** - Persistent data storage with auto-initialization
- **🔴 Redis Cache** - High-performance caching and session storage
- **🤖 Ollama + Phi-3 Mini** - Local AI analysis (auto-downloads 2GB model)
- **📊 Prometheus + Grafana** - Monitoring and visualization dashboards

### ✨ **No Manual Setup Required For:**
- ✅ Database schemas and migrations
- ✅ AI model downloads and configuration
- ✅ FFmpeg/FFprobe tools and codecs
- ✅ Quality assessment libraries (VMAF)
- ✅ Authentication and API key management
- ✅ Rate limiting and security headers
- ✅ Monitoring and health checks
- ✅ Container networking and volumes

**Just run installation script and everything works!**

## 🔧 API Overview

### Core Endpoints

| Endpoint | Method | Description | Example |
|----------|--------|-------------|---------|
| `/api/v1/probe/file` | POST | Analyze uploaded video | Upload and get comprehensive analysis |
| `/api/v1/probe/url` | POST | Analyze video from URL | Analyze remote video files |
| `/api/v1/comparisons/quick` | POST | Compare two videos | Before/after validation |
| `/api/v1/quality/vmaf` | POST | VMAF quality comparison | Objective quality metrics |
| `/health` | GET | API health status | System status check |

### Authentication Required
All endpoints require authentication:

```bash
# API Key (Recommended)
curl -H "X-API-Key: your-api-key" ...

# JWT Token
curl -H "Authorization: Bearer your-jwt-token" ...
```

### Response Format
```json
{
  "id": "analysis-uuid",
  "status": "completed",
  "file_name": "video.mp4",
  "file_size": 104857600,
  "analysis": {
    "format": {
      "duration": "120.5",
      "bit_rate": "5000000",
      "format_name": "mov,mp4,m4a,3gp,3g2,mj2"
    },
    "streams": [/* detailed stream info */]
  },
  "quality_metrics": {
    "vmaf_score": 85.6,
    "psnr": 42.3,
    "ssim": 0.95
  },
  "llm_report": "Professional analysis with recommendations...",
  "created_at": "2024-01-15T10:30:00Z"
}
```

## 📊 Scaling & Performance

### Resource Requirements by Deployment Type

| Deployment | RAM | CPU | Storage | Use Case |
|------------|-----|-----|---------|----------|
| **Single Server** | 4GB | 2 cores | 5GB | Development, small projects |
| **Production** | 8GB | 4 cores | 20GB | Medium production workloads |
| **Enterprise** | 16GB+ | 8+ cores | 50GB+ | High availability, large scale |

### Scaling Examples
```bash
# Light Load: 10-50 requests/min
docker compose up -d

# Medium Load: 100-500 requests/min
docker compose -f compose.yml -f compose.production.yml up -d

# Heavy Load: 1000+ requests/min  
docker compose -f compose.yml -f compose.enterprise.yml up -d \
  --scale ffprobe-api=3 --scale ffprobe-worker=5 --scale ai-worker=2
```

### Performance Metrics
- **Processing Speed**: 1-5 minutes per video (depending on size)
- **Concurrent Jobs**: 2-20 (based on deployment type)
- **API Throughput**: 60-1000 requests/minute (with rate limiting)
- **AI Analysis**: 10-30 seconds per video (local Phi-3 Mini)

## 🔒 Security Features

- **🏆 96/100 Security Score** - Enterprise-grade security hardening
- **🔐 Multi-Auth Support** - JWT tokens, API keys, role-based access control
- **🛡️ Container Security** - Non-root users, read-only filesystems
- **📊 Rate Limiting** - Configurable per-user/IP limits (60/min default)
- **🔍 Audit Logging** - Complete request/response logging
- **🚫 Input Validation** - Comprehensive input sanitization

## 🔍 Monitoring Options

### Self-Hosted (Default)
- **Grafana Dashboard**: `http://localhost:3000` (admin/[generated-password])
- **Prometheus Metrics**: `http://localhost:9090`
- **Health Endpoints**: `http://localhost:8080/health`

### Grafana Cloud (Enterprise)
```bash
# Configure during installation or manually
GRAFANA_CLOUD_URL=https://your-instance.grafana.net
GRAFANA_CLOUD_USERNAME=your-username  
GRAFANA_CLOUD_API_KEY=your-api-key

# Deploy with cloud monitoring
docker compose -f compose.yml -f docker/grafana-cloud.yml up -d
```

### Key Metrics Monitored
- Request rate and response times
- Video processing queue depth
- Quality analysis success rates
- Resource utilization (CPU, memory, disk)
- AI model performance and availability

## 📚 Documentation

### Essential Guides
| Guide | When to Use |
|-------|-------------|
| **[Complete API Guide](docs/api/complete-api-guide.md)** | Detailed API usage with examples |
| **[API Authentication](docs/API_AUTHENTICATION.md)** | Setting up API keys and JWT tokens |
| **[Video Comparison System](docs/COMPARISON_SYSTEM.md)** | Before/after video validation |
| **[Local LLM Setup](docs/tutorials/local-llm-setup.md)** | Zero-configuration AI guide |

### Quick References
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Configuration Guide](docs/deployment/configuration.md)** - Environment setup
- **[Production Deployment](docs/deployment/PRODUCTION_READINESS_CHECKLIST.md)** - Production checklist

## 🆘 Troubleshooting

### Common Issues

**Services won't start**
```bash
# Check system resources
free -h && df -h

# Check Docker
docker --version && docker compose version

# View logs
docker compose logs -f
```

**API authentication fails**
```bash
# Verify API key format (should be 79 characters)
echo $API_KEY | wc -c

# Test health endpoint
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health
```

**AI analysis not working**
```bash
# Check Ollama service
curl http://localhost:11434/api/version

# Verify model downloaded
docker compose exec ollama ollama list
```

### Getting Help
- **📖 Documentation**: [docs/](docs/) directory
- **🐛 GitHub Issues**: [Report Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
- **📧 Contact**: [dev@rendiff.dev](mailto:dev@rendiff.dev)
- **🌐 Website**: [rendiff.dev](https://rendiff.dev)

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup
```bash
# Clone and setup development environment
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# Start development services
./scripts/install-single-server.sh

# Run tests
make test

# Build
make build
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **FFmpeg Team** - For the excellent FFmpeg and FFprobe tools
- **Ollama** - For making local LLM deployment simple
- **Microsoft** - For the efficient Phi-3 Mini model
- **Netflix** - For the VMAF quality assessment library

---

## 🏢 About Rendiff

**FFprobe API** is developed by [**Rendiff**](https://rendiff.dev) - specialists in professional video processing and AI-powered media analysis.

- **🌐 Website**: [rendiff.dev](https://rendiff.dev)
- **📧 Contact**: [dev@rendiff.dev](mailto:dev@rendiff.dev)
- **🐦 Twitter**: [@rendiffdev](https://x.com/rendiffdev)
- **💻 GitHub**: [github.com/rendiffdev](https://github.com/rendiffdev)

**Built with ❤️ for professional video workflows**

---

## 📞 Need Help Getting Started?

1. **First time setup**: Run `./scripts/install-single-server.sh` 
2. **Production deployment**: Run `./scripts/install.sh` for guided setup
3. **API usage**: Check [Complete API Guide](docs/api/complete-api-guide.md)
4. **Enterprise scaling**: See [compose.enterprise.yml](compose.enterprise.yml)
5. **Issues**: Create a [GitHub Issue](https://github.com/rendiffdev/ffprobe-api/issues)

**🎬 Ready to build professional video applications!**