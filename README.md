# FFprobe API - Enterprise Video Analysis

**Production-ready video analysis platform with 49 quality control checks, AI insights, and broadcast compliance**

[![Production Ready](https://img.shields.io/badge/production-ready-green.svg)](docs/deployment/README.md)
[![Quality Control](https://img.shields.io/badge/quality_control-49_checks-blue.svg)](docs/QUALITY_CHECKS.md)
[![Docker](https://img.shields.io/badge/docker-optimized-blue.svg)](docs/deployment/README.md)

## Core Capabilities

- **📹 Comprehensive Analysis**: 49 quality control parameters with 83% industry standard coverage
- **🤖 AI-Powered Insights**: Professional video engineering reports with local/cloud LLM integration
- **📊 Industry Metrics**: VMAF, PSNR, SSIM analysis using Netflix-grade quality models
- **🔍 Advanced Detection**: Content analysis for blackness, freeze frames, clipping, broadcast compliance
- **🏗️ Enterprise Architecture**: Scalable microservices with monitoring, security, and Docker optimization

## Quick Start

```bash
# Start with Docker (recommended)
docker compose up -d

# Basic analysis (29 checks)
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: demo-key" \
  -d '{"file_path": "/path/to/video.mp4"}'

# Enhanced analysis (49 checks)
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: demo-key" \
  -d '{"file_path": "/path/to/video.mp4", "content_analysis": true}'
```

**📋 [Complete Setup Guide →](docs/deployment/README.md)**

## System Architecture

**Scalable microservices architecture with enterprise deployment options**

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   Client    │───▶│  FFprobe API │───▶│ PostgreSQL │
└─────────────┘    │   (Go/Gin)   │    │  Database   │
                   └──────┬───────┘    └─────────────┘
                          │
                   ┌──────▼──────┐    ┌─────────────┐
                   │    Redis    │    │ Monitoring  │
                   │  (Caching)  │    │(Prometheus) │  
                   └─────────────┘    └─────────────┘
```

**🏗️ [Architecture Details →](docs/development/architecture.md)**

## Feature Overview

### 🎬 Professional Video Analysis
**Complete metadata extraction with advanced quality control**
- FFprobe integration with 29 standard + 20 enhanced quality parameters
- GOP analysis, chroma subsampling, bitrate mode detection
- Content analysis: blackness, freeze frames, audio clipping detection
- HLS/DASH streaming protocol support with batch processing

### 🤖 AI-Powered Engineering Reports
**Professional insights with local and cloud LLM integration**
- Ollama local LLM support with configurable models
- OpenRouter cloud AI fallback for enhanced capabilities
- Quality recommendations and comparison analysis
- Professional video engineering report generation

### 📊 Industry-Standard Quality Metrics
**Netflix-grade quality assessment with broadcast compliance**
- VMAF (Video Multimethod Assessment Fusion) scoring
- PSNR/SSIM objective quality measurements
- EBU R128 loudness compliance validation
- Frame-by-frame quality tracking and statistics

**📋 [Complete Quality Checks (49 total) →](docs/QUALITY_CHECKS.md)**

### 🏗️ Enterprise Production Features
**Scalable, secure, and monitoring-ready architecture**
- Microservices with independent scaling capabilities
- Security hardened with JWT/API key authentication
- Prometheus monitoring with health checks and logging
- Multi-stage Docker builds with resource optimization

## Security & Authentication

**Enterprise-grade security with multiple authentication methods**
- JWT token authentication with refresh tokens
- API key authentication for service integration
- Role-based access control (Admin, User, Pro, Premium)
- Rate limiting with Redis and CORS configuration
- Input validation and path traversal protection
- Secure password hashing and database prepared statements

**🔒 [Security Documentation →](docs/operations/security.md)**

## API Overview

**RESTful API with comprehensive video analysis endpoints**

| Endpoint | Description | Quality Checks |
|----------|-------------|----------------|
| `POST /api/v1/probe/file` | Analyze local video file | 29 standard + 20 enhanced |
| `POST /api/v1/probe/url` | Analyze video from URL | 29 standard + 20 enhanced |
| `POST /api/v1/probe/quick` | Fast basic analysis | 29 standard only |
| `POST /api/v1/batch/analyze` | Batch processing | Configurable |
| `GET /health` | System health check | N/A |

### Enhanced Analysis
```bash
# Enable all 49 quality checks with content analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -d '{"file_path": "/path/to/video.mp4", "content_analysis": true}'
```

**📖 [Complete API Documentation →](docs/api/README.md)**

## Production Deployment

**Enterprise-ready with multiple deployment configurations**

| Deployment | Resources | Throughput | Use Case |
|------------|-----------|------------|----------|
| Development | 4GB RAM, 2 cores | 2-5 concurrent | Testing, development |
| Production | 8GB RAM, 4 cores | 5-15 concurrent | Medium-scale operations |
| Enterprise | 16GB+ RAM, 8+ cores | 15-50 concurrent | High-volume processing |

**⚙️ [Complete Deployment Guide →](docs/deployment/README.md)**

## Monitoring & Operations

**Production-ready monitoring with Prometheus and structured logging**
- Health check endpoints with comprehensive system validation
- Prometheus metrics for request rates, processing times, error tracking
- Structured JSON logging with request IDs and audit trails
- Optional Grafana Cloud integration for enterprise deployments

**📊 [Monitoring Setup →](docs/deployment/monitoring.md)**

## Configuration

**Comprehensive configuration with validation and environment-based settings**
- Server, database, and authentication configuration
- Security settings with API keys and JWT token management
- Storage configuration with upload and report directories
- Optional LLM integration (Ollama local, OpenRouter cloud)
- Production validation for all configuration parameters

**⚙️ [Configuration Reference →](docs/deployment/configuration.md)**

## Deployment Options

**Multiple deployment strategies for different scales and requirements**
- **Docker Compose**: Recommended for development and production
- **Kubernetes**: Enterprise-grade orchestration with scaling and monitoring
- **Cloud Providers**: AWS, GCP, Azure deployment guides
- **Bare Metal**: Traditional server deployment documentation

**🐳 [Deployment Guide →](docs/deployment/README.md)**

## Testing

**Comprehensive test suite with unit, integration, and load testing**
- Unit tests for core business logic and utilities
- Integration tests for database and external service operations
- API endpoint testing with authentication and error handling
- Load testing for performance validation

**🧪 [Testing Guide →](docs/development/testing.md)**

## Documentation

### 📖 User Guides
- **[API Reference](docs/api/README.md)** - Complete endpoint documentation
- **[Quality Checks](docs/QUALITY_CHECKS.md)** - 49 quality control parameters
- **[Authentication](docs/api/authentication.md)** - JWT and API key setup

### 🚀 Deployment
- **[Deployment Guide](docs/deployment/README.md)** - Complete deployment options
- **[Configuration](docs/deployment/configuration.md)** - Environment and settings
- **[Monitoring](docs/deployment/monitoring.md)** - Production monitoring setup

### 🛠️ Development
- **[Development Setup](docs/development/README.md)** - Local development environment
- **[Architecture](docs/ARCHITECTURE.md)** - System design and components
- **[Contributing](CONTRIBUTING.md)** - Contribution guidelines

## Troubleshooting

**Common issues and performance optimization guidance**
- Authentication and database connection troubleshooting
- FFmpeg and processing failure diagnostics
- Performance tuning for database and memory management
- Scaling recommendations for high-volume deployments

**🔧 [Troubleshooting Guide →](docs/troubleshooting/README.md)**

## Contributing

**We welcome contributions from the community**
- Fork repository and create feature branches
- Follow Go coding standards and security practices
- Include comprehensive tests with all changes
- Container-native development approach

**👥 [Contributing Guidelines →](CONTRIBUTING.md)**

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- **FFmpeg Team** - Excellent FFmpeg and FFprobe tools
- **Gin Framework** - High-performance HTTP web framework  
- **PostgreSQL** - Robust database system
- **Netflix** - VMAF quality assessment library
- **Go Community** - Excellent ecosystem and tools

---

## Support & Contact

- **🐛 Issues**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
- **📧 Email**: [support@rendiff.dev](mailto:support@rendiff.dev)
- **📖 Documentation**: [Complete Documentation](docs/)

## Production Checklist

**Pre-deployment validation checklist**
- [ ] Environment variables and secure API keys configured
- [ ] SSL/TLS certificates and database security setup
- [ ] Monitoring, alerting, and backup procedures enabled
- [ ] Authentication, file uploads, and resource limits tested
- [ ] Log rotation and storage configuration verified

**🚀 [Production Checklist →](docs/deployment/production-checklist.md)**