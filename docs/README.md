# FFprobe API Documentation

**Complete documentation for the professional video analysis API with advanced QC features**

## 📖 Quick Navigation

### 🚀 Getting Started
- **[Quick Start Guide](../README.md#quick-start)** - Get running in 2 minutes
- **[Docker Deployment](../docker-image/QUICK_START.md)** - Zero-config Docker deployment
- **[Local AI Setup](tutorials/local-llm-setup.md)** - AI-powered analysis setup

### 📡 API Reference
- **[REST API Documentation](api/README.md)** - Complete endpoint reference
- **[GraphQL API Guide](api/GRAPHQL_API_GUIDE.md)** - GraphQL queries and mutations
- **[QC Features Guide](api/QC_FEATURES.md)** - Quality Control analysis integration
- **[OpenAPI Specification](api/openapi.yaml)** - Machine-readable API spec

### 🔧 Advanced Topics
- **[Production Deployment](deployment/README.md)** - Production deployment strategies
- **[Architecture Overview](development/architecture.md)** - System architecture and design
- **[Secret Rotation](api/SECRET_ROTATION_GUIDE.md)** - API key and JWT management
- **[Complete QC Analysis List](../QC_ANALYSIS_LIST.md)** - All 20+ QC categories detailed
- **[GraphQL API Guide](api/GRAPHQL_API_GUIDE.md)** - GraphQL endpoints  
- **[Authentication Guide](api/authentication.md)** - API keys and security
- **[Secret Rotation Guide](api/SECRET_ROTATION_GUIDE.md)** - Security management

### 🏗️ Development & Architecture
- **[System Architecture](development/architecture.md)** - Technical design overview
- **[Video Comparison System](COMPARISON_SYSTEM.md)** - Quality comparison features

### 🔧 Operations & Monitoring
- **[FFmpeg Management](operations/ffmpeg-management.md)** - FFmpeg updates and configuration
- **[Monitoring Setup](operations/monitoring.md)** - Prometheus and Grafana
- **[Security Guide](operations/security.md)** - Security best practices
- **[Troubleshooting](TROUBLESHOOTING.md)** - Common issues and solutions

### 📋 Production Readiness
- **[Production Checklist](deployment/PRODUCTION_READINESS_CHECKLIST.md)** - Pre-deployment validation

---

## 🎯 By Use Case

### "I want to..."

#### **Analyze videos**
- [Upload and analyze a video file →](api/README.md)
- [Compare video quality improvements →](COMPARISON_SYSTEM.md)
- [Enable AI-powered insights →](tutorials/local-llm-setup.md)

#### **Deploy to production**
- [Production deployment guide →](deployment/README.md)
- [Security configuration →](operations/security.md)
- [Monitoring setup →](operations/monitoring.md)

#### **Develop and extend**
- [API development tutorial →](tutorials/api_usage.md)
- [System architecture →](development/architecture.md)
- [Contributing guidelines →](../CONTRIBUTING.md)

#### **Troubleshoot issues**
- [Common problems and solutions →](TROUBLESHOOTING.md)
- [FFmpeg issues →](operations/ffmpeg-management.md)
- [Docker Compose issues →](deployment/modern-docker-compose.md)

---

## 🏆 Key Features

### **AI-Powered Analysis**
- **Dual-Model Setup**: Gemma 3 270M (fast) + Phi-3 Mini (comprehensive)
- **Professional Reports**: 8-section technical analysis
- **Quality Assessment**: VMAF, PSNR, SSIM metrics
- **Smart Recommendations**: FFmpeg optimization suggestions

### **Enterprise Ready**
- **Latest FFmpeg**: BtbN builds with all codecs
- **Production Monitoring**: Prometheus + Grafana
- **Automatic SSL**: Traefik with Let's Encrypt
- **Scalable Architecture**: Modern Docker Compose profiles

### **Developer Friendly**
- **REST + GraphQL APIs**: Complete endpoint coverage
- **Comprehensive Testing**: Unit, integration, and E2E tests
- **Modern Deployment**: Profile-based Docker Compose
- **Detailed Documentation**: Complete guides and references

---

## 📊 Resource Requirements

| Deployment | Memory | CPU | Storage | Use Case |
|------------|--------|-----|---------|----------|
| **Minimal** | 2-3GB | 2 cores | 5GB | Development, testing |
| **Quick** | 4-5GB | 2-4 cores | 8GB | Demos, small teams |
| **Production** | 8-16GB | 8+ cores | 30GB+ | Enterprise deployment |

---

## 🆘 Support

- **Documentation Issues**: Check [troubleshooting guide](TROUBLESHOOTING.md)
- **Bug Reports**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)

---

**Built for the video engineering community** 🎬