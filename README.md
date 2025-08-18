# FFprobe API

**Professional video analysis API with comprehensive Quality Control (QC) features**

Complete media analysis solution with 20+ professional QC analysis categories and AI-powered insights.

[![Production Ready](https://img.shields.io/badge/production-ready-green.svg)](PRODUCTION_READINESS_REPORT.md)
[![QC Analysis](https://img.shields.io/badge/QC-20%20Categories-blue.svg)](#advanced-quality-control-features)
[![Docker](https://img.shields.io/badge/docker-latest%20compose-blue.svg)](docs/deployment/modern-docker-compose.md)

## ✨ Features

- **Advanced Quality Control**: 20+ professional QC analysis categories including timecode, AFD, MXF validation, dead pixel detection, PSE analysis
- **Latest FFmpeg**: Always uses latest stable BtbN builds with all codecs
- **AI-Enhanced Analysis**: Optional LLM integration for intelligent insights and risk assessment
- **Professional Reports**: Comprehensive technical analysis with quality metrics and compliance validation
- **Multiple Formats**: Supports all video/audio formats that FFmpeg supports
- **REST & GraphQL API**: Complete RESTful and GraphQL interfaces with OpenAPI documentation
- **Zero Config**: Runs out-of-the-box with sensible defaults
- **Production Ready**: Full monitoring, backups, and security features

## 🚀 Quick Start

### One-Command Installation
```bash
# Option 1: Interactive setup (recommended)
curl -fsSL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/setup.sh | bash

# Option 2: Ultra-quick install
curl -fsSL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/install.sh | bash

# Option 3: Clone and run
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api
make quick
```

Your API is now running at **http://localhost:8080**

### Test It
```bash
# Check health
curl http://localhost:8080/health

# Analyze a video
curl -X POST -F "file=@video.mp4" http://localhost:8080/api/v1/probe/file
```

## 📋 System Requirements

- **Docker** 24.0+ with Compose
- **4GB RAM** minimum (6GB recommended)
- **10GB disk space** for models and data
- **Internet connection** for initial setup

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   FFprobe API   │───▶│   PostgreSQL    │    │     Redis       │
│   (Latest)      │    │   (Database)    │    │   (Cache)       │
└─────────┬───────┘    └─────────────────┘    └─────────────────┘
          │
          ▼
┌─────────────────┐    ┌─────────────────┐
│     Ollama      │    │   FFmpeg        │
│  (AI Models)    │    │  (BtbN Latest)  │
│                 │    │                 │
│ • Gemma3 270M   │    │ • All Codecs    │
│ • Phi-3 Mini    │    │ • VMAF Support  │
└─────────────────┘    └─────────────────┘
```

## 🛠️ Deployment Modes

### Minimal Deployment (Recommended for Testing)
```bash
make minimal
```
- **Ultra-lightweight**: Only 4 core services
- **Services**: API + PostgreSQL + Redis + Ollama
- **Memory**: ~2-3GB total
- **Perfect for**: Development, testing, resource-constrained environments

### Quick Start (Development)
```bash
make quick
```
- Ready in 2 minutes
- No authentication
- Gemma 3 270M for fast analysis
- Perfect for demos and quick testing

### Development Environment
```bash
make dev
```
- Hot reload and debugging
- Database/Redis admin tools
- File browser interface for uploads
- Development-focused tooling

### Production Deployment
```bash
make prod
```
- Full security and monitoring
- Prometheus + Grafana dashboards
- Automated backups
- **Traefik**: Combined reverse proxy + automatic SSL
- Enterprise-ready infrastructure

## 🔍 Advanced Quality Control Features

The FFprobe API provides **comprehensive professional QC analysis** with industry-standard compliance checking.

📋 **[Complete QC Analysis List](QC_ANALYSIS_LIST.md)** - Detailed breakdown of all 20+ QC categories

### QC Analysis Categories Overview

#### Standard Technical Analysis (11 Categories)
- Stream analysis and counting
- Video/audio technical validation  
- Codec and container compliance
- Frame and GOP structure analysis
- Bit depth and resolution validation

#### Advanced Professional QC (9 Categories)
- **Timecode Analysis**: SMPTE timecode parsing, drop frame detection
- **Active Format Description (AFD)**: Broadcast signaling compliance
- **Transport Stream Analysis**: MPEG-TS PID analysis and error detection
- **Endianness Detection**: Binary format analysis and platform compatibility
- **Audio Wrapping**: Professional audio format detection and validation
- **IMF Compliance**: Interoperable Master Format validation (Netflix standard)
- **MXF Format Validation**: Material Exchange Format compliance checking
- **Dead Pixel Detection**: Computer vision-based pixel defect analysis
- **Photosensitive Epilepsy (PSE) Risk**: Automated PSE safety analysis

### AI-Enhanced Analysis (Optional)
- **Risk Assessment**: Automated technical, compliance, and safety risk evaluation
- **Quality Scoring**: Overall QC score with critical findings identification
- **Workflow Integration**: Intelligent recommendations for production pipelines
- **Compliance Insights**: Broadcast standards validation (ITU, FCC, EBU, ATSC)

## 🔒 Security Features

- **API Key Authentication**: Secure access control
- **Rate Limiting**: Prevent abuse
- **Input Validation**: Comprehensive file validation
- **Secure Defaults**: No sensitive data exposure
- **Container Security**: Minimal attack surface

## ⚡ Optimized Component Architecture

### **Essential Components** (All Deployments)
- **PostgreSQL 16**: Primary database
- **Redis 7**: High-performance caching
- **Ollama**: Local AI processing (Gemma 3 270M + Phi-3 Mini)
- **FFprobe API**: Core video analysis service

### **Production Components** (Production Only)
- **Traefik v3**: Combined reverse proxy + automatic SSL
- **Prometheus**: Metrics and monitoring
- **Grafana**: Dashboards and visualization
- **Backup Service**: Automated data protection

### **Development Tools** (Development Only)
- **Adminer**: Database administration
- **Redis Commander**: Redis administration  
- **File Browser**: Upload management


**Resource Savings**: ~150MB RAM, faster startup, fewer containers to manage

## 📖 API Documentation

### Core Endpoints

#### Health Check
```bash
GET /health
```

#### File Analysis with Advanced QC
```bash
POST /api/v1/probe/file
Content-Type: application/json

{
  "file_path": "/path/to/video.mp4",
  "content_analysis": true,
  "generate_reports": true,
  "report_formats": ["json", "pdf"]
}
```

#### URL Analysis
```bash
POST /api/v1/probe/url
Content-Type: application/json

{
  "url": "https://example.com/video.mp4",
  "content_analysis": true,
  "timeout": 300
}
```

#### GraphQL Query
```graphql
query AnalyzeMedia($input: AnalysisInput!) {
  analyzeMedia(input: $input) {
    id
    status
    result {
      enhancedAnalysis {
        timecodeAnalysis { hasTimecode isDropFrame }
        mxfAnalysis { isMXFFile validationResults }
        pseAnalysis { riskLevel violations }
      }
    }
  }
}
```

### Response Format
```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "analysis": {
    "file_info": {
      "filename": "video.mp4",
      "size": 1048576,
      "duration": 60.5
    },
    "result": {
      "streams": [...],
      "format": {...},
      "enhanced_analysis": {
        "timecode_analysis": {
          "has_timecode": true,
          "is_drop_frame": false,
          "start_timecode": "01:00:00:00"
        },
        "mxf_analysis": {
          "is_mxf_file": true,
          "mxf_profile": "OP1a",
          "validation_results": {
            "overall_compliance": true
          }
        },
        "pse_analysis": {
          "pse_risk_level": "safe",
          "flash_analysis": {...}
        },
        "llm_enhanced_report": {
          "overall_qc_score": 95.5,
          "critical_findings": [],
          "risk_assessment": {
            "overall_risk_level": "low"
          }
        }
      }
    }
  }
}
```

## 🔧 Management Commands

```bash
# Service management
make start              # Start all services
make stop               # Stop all services
make restart            # Restart all services
make health             # Check service health
make logs               # View all logs

# Development
make dev                # Development environment
make shell              # Access API container
make db-shell           # Access database
make redis-shell        # Access Redis

# Maintenance
make update             # Update to latest versions
make backup             # Create backup
make clean              # Clean everything
```

## 📊 Monitoring & Observability

### Health Checks
- Service health endpoints
- Dependency health validation
- Model availability checking
- Resource usage monitoring

### Metrics (Production)
- Request rate and latency
- Processing queue depth
- Resource utilization
- Error rates and types

### Dashboards (Production)
- Grafana visualizations
- Real-time service status
- Performance trending
- Alert management
## 🔄 Updates & Maintenance

### Automatic Updates
- **FFmpeg**: Latest stable BtbN builds
- **AI Models**: Model version management
- **Security**: Regular security updates
- **Dependencies**: Container base image updates

### Manual Commands
```bash
# Update everything
make update

# Update specific components
./scripts/ffmpeg-update.sh check
./scripts/ffmpeg-update.sh update --allow-major

# Model management
docker compose exec ollama ollama list
docker compose exec ollama ollama pull gemma3:270m
```

## 🐛 Troubleshooting

### Common Issues

#### Services won't start
```bash
# Check Docker
docker --version
docker compose version

# View logs
make logs

# Reset everything
make reset
```

#### Port conflicts
```bash
# Check ports
lsof -i :8080
lsof -i :5432

# Use custom ports
export API_PORT=8081
export POSTGRES_PORT=5433
make quick
```

#### Models not downloading
```bash
# Check Ollama
curl http://localhost:11434/api/version

# Manual download
docker compose exec ollama ollama pull gemma3:270m
```

## 📚 Additional Documentation

- [Docker Compose Guide](docs/deployment/modern-docker-compose.md)
- [FFmpeg Management](docs/operations/ffmpeg-management.md)
- [AI Model Setup](docs/tutorials/local-llm-setup.md)
- [Production Deployment](docs/deployment/README.md)
- [API Reference](docs/api/README.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)

---

**Built with ❤️ for the video processing community**