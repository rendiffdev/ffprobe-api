# 🎬 FFprobe API - Production Ready

> **Enterprise-grade video analysis API with AI-powered insights and professional scaling**

[![Production Ready](https://img.shields.io/badge/production-ready-green.svg)](#-production-readiness)
[![Security Hardened](https://img.shields.io/badge/security-hardened-brightgreen.svg)](#-security-features)
[![API Version](https://img.shields.io/badge/api-v1.0-blue.svg)](#-api-overview)
[![Docker Optimized](https://img.shields.io/badge/docker-optimized-blue.svg)](#-deployment-options)

## 🎯 What This Solves

**Problem**: Professional video analysis with scalable, secure, and production-ready infrastructure  
**Solution**: Complete FFprobe integration with AI insights, quality metrics, and enterprise features

### Key Capabilities
- 📹 **Complete Video Analysis**: Technical specs, quality metrics, HLS/DASH support
- 🔍 **Enhanced Quality Control**: 16 additional QC parameters including GOP analysis, content detection
- 🤖 **AI-Powered Insights**: Professional video engineering assessment with recommendations  
- 📊 **Quality Comparison**: VMAF, PSNR, SSIM analysis with before/after validation
- 🏗️ **Production Grade**: Hardened security, scalable architecture, comprehensive monitoring

## 🚀 Quick Start

### Development Setup
```bash
# Clone repository
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# Start with Docker Compose
docker compose up -d

# Verify services
curl http://localhost:8080/health
```

### Simple Deployment (Small/Test Organizations)
```bash
# Complete LLM-powered setup without monitoring overhead
docker compose -f compose.simple.yml up -d

# Verify services (includes AI/LLM status)
curl http://localhost:8080/health
```

### Production Deployment
```bash
# Production-ready configuration with Ollama LLM
cp .env.example .env
# Edit .env with your production values

docker compose -f compose.yml -f compose.production.yml up -d
```

### Enterprise Deployment  
```bash
# Full monitoring stack with Prometheus/Grafana
docker compose -f compose.yml -f compose.enterprise.yml up -d
```

## 🏗️ Architecture

### Core Services
```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Client    │───▶│  FFprobe API │───▶│ PostgreSQL │
└─────────────┘    │   (Go/Gin)   │    │  Database   │
                   └──────┬───────┘    └─────────────┘
                          │
                   ┌──────▼──────┐    ┌─────────────┐
                   │    Redis    │    │   Monitoring│
                   │  (Caching)  │    │(Prometheus) │
                   └─────────────┘    └─────────────┘
```

### Enterprise Scaling
```
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐
│Load Balancer│───▶│ API Cluster  │───▶│ Worker Pool     │
│   (Nginx)   │    │ (Scalable)   │    │ (FFprobe/LLM)   │
└─────────────┘    └──────┬───────┘    └─────────────────┘
                          │
                   ┌──────▼──────┐    ┌─────────────────┐
                   │ Database    │    │ Storage Layer   │
                   │ (PostgreSQL)│    │ (Local/Cloud)   │
                   └─────────────┘    └─────────────────┘
```

## ✨ Core Features

### 🎬 Professional Video Analysis
- **Complete FFprobe Integration**: All metadata, streams, formats, chapters
- **Enhanced Quality Control**: 16 additional QC parameters (GOP analysis, chroma subsampling, bitrate mode detection)
- **Content Analysis**: Blackness detection, freeze frames, audio clipping, interlacing artifacts
- **Quality Metrics**: VMAF, PSNR, SSIM analysis with Netflix-grade models
- **HLS/DASH Support**: Streaming protocol validation and optimization
- **Batch Processing**: Handle multiple files efficiently with progress tracking
- **Raw Data Access**: Direct access to FFprobe JSON output

### 🤖 AI-Powered Insights
- **Local LLM Support**: Ollama integration with configurable models
- **Professional Analysis**: Comprehensive video engineering reports
- **Quality Recommendations**: AI-generated optimization suggestions
- **Comparison Reports**: Intelligent before/after analysis
- **OpenRouter Integration**: Fallback to cloud AI services

### 📊 Quality Assessment System
- **VMAF Analysis**: Netflix Video Multimethod Assessment Fusion
- **Perceptual Metrics**: PSNR, SSIM, and custom quality models
- **Frame-by-Frame Analysis**: Detailed quality tracking over time
- **Comparison Engine**: Objective before/after quality validation
- **Quality Statistics**: Comprehensive quality reporting and trends

### 🏗️ Production Architecture
- **Microservices Design**: Independently scalable components
- **Database Migrations**: Automated schema management with PostgreSQL
- **Security Hardened**: Authentication, authorization, input validation
- **Monitoring Ready**: Prometheus metrics, health checks, logging
- **Container Optimized**: Multi-stage builds, non-root users, resource limits

## 🔒 Security Features

### Authentication & Authorization
- **JWT Token Authentication**: Secure session management with refresh tokens
- **API Key Authentication**: Service-to-service authentication
- **Role-Based Access Control**: Admin, user, pro, premium roles
- **Account Lockout**: Automatic protection against brute force attacks

### Security Hardening
- **Input Validation**: Comprehensive validation for all endpoints
- **Path Traversal Protection**: Secure file upload handling
- **CORS Configuration**: Configurable cross-origin resource sharing
- **Rate Limiting**: Per-user and per-IP rate limiting with Redis
- **Security Headers**: HSTS, XSS protection, content type validation

### Data Protection
- **Password Hashing**: bcrypt with salt for secure password storage
- **Secure File Handling**: Upload sanitization and path validation
- **Database Security**: Prepared statements, connection pooling
- **Error Handling**: Consistent error responses without information leakage

## 🔧 API Overview

### Core Endpoints

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|---------------|
| `/api/v1/probe/file` | POST | Analyze uploaded video (supports `content_analysis: true`) | API Key/JWT |
| `/api/v1/probe/url` | POST | Analyze video from URL (supports `content_analysis: true`) | API Key/JWT |
| `/api/v1/probe/quick` | POST | Fast basic analysis | API Key/JWT |
| `/api/v1/batch/analyze` | POST | Batch video processing | API Key/JWT |
| `/api/v1/quality/compare` | POST | Quality comparison | API Key/JWT |
| `/api/v1/comparisons` | POST | Create comparison | API Key/JWT |
| `/api/v1/reports/analysis` | POST | Generate reports | API Key/JWT |
| `/health` | GET | System health check | None |

### Authentication Methods

```bash
# API Key (Recommended for services)
curl -H "X-API-Key: your-api-key" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/v1/probe/file

# JWT Token (Recommended for users)
curl -H "Authorization: Bearer your-jwt-token" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/v1/probe/file
```

### Enhanced Analysis Request
```bash
# Standard analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "content_analysis": false
  }'

# Enhanced analysis with content analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "content_analysis": true
  }'
```

### Response Format
```json
{
  "analysis_id": "uuid-v4",
  "status": "completed",
  "analysis": {
    "format": {
      "duration": "120.5",
      "bit_rate": "5000000",
      "format_name": "mov,mp4,m4a,3gp,3g2,mj2"
    },
    "streams": [
      {
        "codec_name": "h264",
        "width": 1920,
        "height": 1080,
        "r_frame_rate": "30/1"
      }
    ],
    "enhanced_analysis": {
      "stream_counts": {
        "video_streams": 1,
        "audio_streams": 2,
        "subtitle_streams": 0
      },
      "video_analysis": {
        "chroma_subsampling": "4:2:0",
        "matrix_coefficients": "ITU-R BT.709",
        "bit_rate_mode": "CBR",
        "has_closed_captions": false
      },
      "gop_analysis": {
        "average_gop_size": 30.0,
        "keyframe_count": 120,
        "gop_pattern": "Regular (GOP=30)"
      },
      "frame_statistics": {
        "total_frames": 3600,
        "i_frames": 120,
        "p_frames": 2400,
        "b_frames": 1080
      },
      "content_analysis": {
        "black_frames": {
          "detected_frames": 0,
          "percentage": 0.0
        },
        "loudness_meter": {
          "integrated_loudness_lufs": -23.0,
          "broadcast_compliant": true,
          "standard": "EBU R128"
        }
      }
    },
    "quality_metrics": {
      "vmaf_score": 85.6,
      "psnr": 42.3,
      "ssim": 0.95
    }
  },
  "reports": {
    "formats": ["json", "pdf"],
    "download_urls": ["..."]
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

## 📊 Production Readiness

### Resource Requirements

| Deployment Type | RAM | CPU | Storage | Concurrent Jobs |
|-----------------|-----|-----|---------|-----------------|
| **Development** | 4GB | 2 cores | 10GB | 2-5 |
| **Production** | 8GB | 4 cores | 50GB | 5-15 |
| **Enterprise** | 16GB+ | 8+ cores | 100GB+ | 15-50 |

### Scaling Configuration

```bash
# Simple Setup (Small/Test Organizations) - LLM-Powered
docker compose -f compose.simple.yml up -d

# Development/Testing - Full AI Features
docker compose -f compose.yml -f compose.dev.yml up -d

# Production (Medium Load)
docker compose -f compose.yml -f compose.production.yml up -d

# Enterprise (Heavy Load + Monitoring)
docker compose -f compose.yml -f compose.enterprise.yml up -d \
  --scale ffprobe-api=3 \
  --scale ffprobe-worker=5 \
  --scale llm-service=2
```

### Performance Metrics
- **API Throughput**: 60-1000 requests/minute (with rate limiting)
- **Processing Speed**: 30 seconds - 5 minutes per video (size dependent)
- **Concurrent Processing**: 2-50 simultaneous analyses
- **Database Performance**: Connection pooling, prepared statements
- **Memory Usage**: Optimized for container environments

## 🔍 Monitoring & Observability

### Built-in Monitoring
- **Health Checks**: Comprehensive endpoint monitoring
- **Prometheus Metrics**: Request rates, processing times, error rates
- **Database Monitoring**: Connection health, query performance
- **Resource Tracking**: CPU, memory, and disk utilization

### Logging
- **Structured Logging**: JSON format with request IDs
- **Error Tracking**: Detailed error logging with context
- **Audit Logging**: Authentication and authorization events
- **Performance Logging**: Request/response timing and metrics

### Alerting (Optional)
```bash
# Configure Grafana Cloud integration
export GRAFANA_CLOUD_URL="https://your-instance.grafana.net"
export GRAFANA_CLOUD_USERNAME="your-username"
export GRAFANA_CLOUD_API_KEY="your-api-key"

# Monitoring is included in enterprise setup
docker compose -f compose.yml -f compose.enterprise.yml up -d
```

## 🛠️ Configuration

### Environment Variables

```bash
# Server Configuration
API_PORT=8080
API_HOST=localhost
BASE_URL=http://localhost:8080
LOG_LEVEL=info

# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=ffprobe_api
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-secure-password

# Authentication
API_KEY=your-32-char-api-key
JWT_SECRET=your-32-char-jwt-secret
TOKEN_EXPIRY_HOURS=24
REFRESH_EXPIRY_HOURS=168

# Security
ENABLE_AUTH=true
ENABLE_RATE_LIMIT=true
RATE_LIMIT_PER_MINUTE=60
ALLOWED_ORIGINS=*

# Storage
UPLOAD_DIR=/app/uploads
REPORTS_DIR=/app/reports
MAX_FILE_SIZE=53687091200  # 50GB

# FFmpeg Tools
FFMPEG_PATH=ffmpeg
FFPROBE_PATH=ffprobe

# Optional: LLM Configuration
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=mistral:7b
OPENROUTER_API_KEY=your-openrouter-key
```

### Production Configuration Validation

The application includes comprehensive configuration validation:
- **Required Fields**: Validates all mandatory configuration values
- **Security Settings**: Ensures API keys and JWT secrets meet minimum requirements
- **Directory Validation**: Checks directory existence and write permissions
- **Port Validation**: Validates port ranges and availability
- **Token Expiry**: Validates token expiry relationships
- **CORS Origins**: Validates allowed origins format

## 🐳 Deployment Options

### Docker Compose (Recommended)

```yaml
# Basic deployment
version: '3.8'
services:
  ffprobe-api:
    image: ffprobe-api:latest
    ports:
      - "8080:8080"
    environment:
      - POSTGRES_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: ffprobe_api
      POSTGRES_PASSWORD: secure-password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
```

### Kubernetes (Enterprise)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ffprobe-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ffprobe-api
  template:
    metadata:
      labels:
        app: ffprobe-api
    spec:
      containers:
      - name: ffprobe-api
        image: ffprobe-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: POSTGRES_HOST
          value: "postgres-service"
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
```

## 🧪 Testing

### Running Tests
```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# API tests
make test-api

# Load testing
make test-load
```

### Test Coverage
- **Unit Tests**: Core business logic and utilities
- **Integration Tests**: Database operations and external services
- **API Tests**: Endpoint functionality and error handling
- **Security Tests**: Authentication and authorization flows

## 📚 Documentation

### API Documentation
- **[API Reference](docs/api/README.md)** - Complete API documentation
- **[Authentication Guide](docs/API_AUTHENTICATION.md)** - JWT and API key setup
- **[Quality Metrics](docs/QUALITY_METRICS.md)** - VMAF and quality analysis

### Deployment Guides
- **[Deployment Guide](DEPLOYMENT_GUIDE.md)** - Complete deployment options (Simple/Production/Enterprise)
- **[Repository Structure](REPOSITORY_STRUCTURE.md)** - Complete repository organization guide
- **[Production Readiness](PRODUCTION_AUDIT_REPORT.md)** - Security audit and production checklist

### Development
- **[Development Setup](docs/development/SETUP.md)** - Local development environment
- **[Contributing Guide](CONTRIBUTING.md)** - Contribution guidelines
- **[Architecture](docs/ARCHITECTURE.md)** - System design and components

## 🆘 Troubleshooting

### Common Issues

**Authentication Failures**
```bash
# Check API key format (should be 32+ characters)
echo $API_KEY | wc -c

# Verify JWT secret is set
curl -H "Authorization: Bearer $JWT_TOKEN" http://localhost:8080/api/v1/auth/validate
```

**Database Connection Issues**
```bash
# For simple deployment
docker compose -f compose.simple.yml exec postgres pg_isready

# For production/enterprise
docker compose exec postgres pg_isready

# Verify connection string
docker compose logs ffprobe-api | grep -i database
```

**Processing Failures**
```bash
# Check FFmpeg availability
docker compose exec ffprobe-api ffprobe -version

# Monitor processing queue
curl http://localhost:8080/api/v1/batch/status
```

### Performance Optimization

**Database Performance**
- Connection pooling is configured automatically
- Database migrations run on startup
- Indexes are optimized for common queries

**Memory Management**
- Goroutine contexts prevent resource leaks
- File uploads are streamed to prevent memory issues
- Database connections are properly pooled

**Scaling Recommendations**
- Scale API containers horizontally
- Use Redis for session storage and caching
- Consider read replicas for high-read workloads

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run the test suite
5. Submit a pull request

### Code Standards
- Go modules for dependency management
- Comprehensive error handling
- Security-first development
- Container-native design
- Comprehensive testing

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **FFmpeg Team** - For the excellent FFmpeg and FFprobe tools
- **Gin Framework** - For the high-performance HTTP web framework  
- **PostgreSQL** - For the robust database system
- **Netflix** - For the VMAF quality assessment library
- **Go Community** - For the excellent ecosystem and tools

---

## 📞 Support & Contact

- **🐛 Issues**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
- **📧 Email**: [support@rendiff.dev](mailto:support@rendiff.dev)
- **📖 Documentation**: [docs/](docs/) directory

## 🚀 Production Checklist

Before deploying to production:

- [ ] Configure environment variables (`.env` file)
- [ ] Set secure API keys and JWT secrets (32+ characters)
- [ ] Configure database with strong passwords
- [ ] Set up SSL/TLS certificates
- [ ] Configure monitoring and alerting
- [ ] Set appropriate resource limits
- [ ] Test authentication and authorization
- [ ] Verify file upload limits and storage
- [ ] Set up backup procedures
- [ ] Configure log rotation

**Ready for production deployment!** 🎬