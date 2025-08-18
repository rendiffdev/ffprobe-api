# FFprobe API

**AI-Powered Video Analysis API - Beyond Traditional FFprobe**

🧠 **The only media analysis API with built-in GenAI intelligence** - transforming raw FFprobe data into actionable professional insights, recommendations, and risk assessments.

**Why choose FFprobe API over direct FFmpeg/FFprobe?**
- 🎯 **GenAI Analysis**: AI-powered interpretation of technical data into professional insights
- 🔍 **Intelligent Risk Assessment**: AI identifies safety, compliance, and technical risks
- 📊 **Smart Recommendations**: GenAI suggests specific FFmpeg commands and workflow improvements
- 🏆 **Professional QC**: 20+ advanced quality control categories beyond basic FFprobe
- 💡 **Executive Summaries**: AI translates technical data for non-technical stakeholders

[![Production Ready](https://img.shields.io/badge/production-ready-green.svg)](PRODUCTION_READINESS_REPORT.md)
[![QC Analysis](https://img.shields.io/badge/QC-20%20Categories-blue.svg)](#advanced-quality-control-features)
[![Docker](https://img.shields.io/badge/docker-ready--to--deploy-blue.svg)](docker-image/QUICK_START.md)

## 🧠 Core GenAI Differentiators

### **FFprobe API: Enhanced Video Analysis**

| Standard Workflow | FFprobe API Enhancement |
|-------------------|------------------------|
| Technical data output | 🎯 **AI-interpreted insights** |
| Analysis workflow | 🤖 **Automated risk assessment** |
| Raw metrics | 💡 **Smart optimization suggestions** |
| Technical format | 📝 **Executive-friendly summaries** |
| Individual file processing | 🔄 **Workflow integration recommendations** |

### **🚀 GenAI-Powered Features**

- **🧠 AI Technical Analysis**: LLM interprets FFprobe data into professional assessment
- **⚠️ Risk Assessment**: AI identifies PSE risks, compliance issues, technical problems
- **🎯 Smart Recommendations**: GenAI suggests specific FFmpeg commands for optimization
- **📊 Quality Insights**: AI evaluates suitability for different delivery platforms
- **🏢 Executive Summaries**: Technical findings translated for management/clients
- **🔍 Issue Detection**: AI spots problems human analysts might miss

### **🛠️ Advanced Technical Features**

- **Advanced Quality Control**: 20+ professional QC analysis categories including timecode, AFD, MXF validation, dead pixel detection, PSE analysis
- **Latest FFmpeg**: Always uses latest stable BtbN builds with all codecs
- **Professional Reports**: Comprehensive technical analysis with quality metrics and compliance validation
- **Multiple Formats**: Supports all video/audio formats that FFmpeg supports
- **REST & GraphQL API**: Complete RESTful and GraphQL interfaces with OpenAPI documentation
- **Zero Config**: Runs out-of-the-box with sensible defaults
- **Production Ready**: Full monitoring, backups, and security features

## 🚀 Quick Start

### Smart Installation with System Requirements Checking

**🤖 The setup script automatically validates your system meets the requirements for your chosen deployment mode:**

```bash
# 🎯 Interactive setup with automatic system validation (recommended)
curl -fsSL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/setup.sh | bash

# ⚡ Non-interactive modes with automatic requirements checking:
curl -fsSL setup.sh | bash -s -- --quick      # 3GB RAM, 8GB disk
curl -fsSL setup.sh | bash -s -- --minimal    # 2GB RAM, 6GB disk
curl -fsSL setup.sh | bash -s -- --production # 8GB RAM, 20GB disk
curl -fsSL setup.sh | bash -s -- --development # 4GB RAM, 15GB disk

# 🔧 Manual setup (no requirements checking)
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api
make quick
```

**✨ What the smart installer checks:**
- 📊 RAM: Deployment-specific minimum (2-8GB)
- 💾 Disk space: Mode-specific requirements (6-20GB) 
- 🖥️ CPU cores: Sufficient processing power (1-4 cores)
- 🔌 Network ports: Required ports available
- 🐳 Docker: Installation and container capabilities
- 🌐 Internet: Connection for downloading images

Your API is now running at **http://localhost:8080**

### 🧠 Test GenAI Analysis (The Core USP)
```bash
# Check health
curl http://localhost:8080/health

# Traditional analysis (like basic FFprobe)
curl -X POST -F "file=@video.mp4" http://localhost:8080/api/v1/probe/file

# 🎆 GenAI-powered analysis (THE DIFFERENTIATOR)
curl -X POST \
  -F "file=@video.mp4" \
  -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file

# Get AI insights from the analysis
curl http://localhost:8080/api/v1/analysis/{id} | jq '.llm_report'
```

**💫 What you get with GenAI analysis:**
- Professional quality assessment in plain English
- Specific FFmpeg optimization commands
- Risk assessment for PSE/compliance issues  
- Delivery platform recommendations
- Executive summary for stakeholders

## 📋 System Requirements

- **Docker** 24.0+ with Compose
- **2GB RAM** minimum (4GB recommended)
- **5GB disk space** for models and data
- **Internet connection** for initial setup
- **No external database required** - SQLite embedded by default

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   FFprobe API   │───▶│     SQLite      │    │     Valkey      │
│   (Latest)      │    │ (Embedded DB)   │    │ (Redis-compatible)│
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
- **Ultra-lightweight**: Only 3 core services
- **Services**: API + Valkey + Ollama (SQLite embedded)
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
- Database/Valkey admin tools
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
- **Database**: SQLite with WAL mode for production performance

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
- **SQLite**: Embedded database (zero configuration)
- **Valkey 8**: High-performance caching (Redis-compatible, open source)
- **Ollama**: Local AI processing (Gemma 3 270M + Phi-3 Mini)
- **FFprobe API**: Core video analysis service

### **Production Components** (Production Only)
- **Traefik v3**: Combined reverse proxy + automatic SSL
- **Prometheus**: Metrics and monitoring
- **Grafana**: Dashboards and visualization
- **Backup Service**: Automated data protection

### **Development Tools** (Development Only)
- **SQLite Browser**: Database administration
- **Valkey Commander**: Cache administration  
- **File Browser**: Upload management


**Resource Savings**: ~300MB RAM, faster startup, fewer containers to manage

## 📖 API Documentation

### Core Endpoints

#### Health Check
```bash
GET /health
```

#### 🧠 GenAI-Powered Analysis (Core USP)
```bash
# THE DIFFERENTIATOR: AI-powered analysis
POST /api/v1/probe/file
Content-Type: application/json

{
  "file_path": "/path/to/video.mp4",
  "include_llm": true,          // 🎆 Enable GenAI analysis
  "content_analysis": true,
  "generate_reports": true,
  "report_formats": ["json", "pdf"]
}
```

**GenAI Response Includes:**
```json
{
  "analysis_id": "uuid",
  "llm_report": "🧠 EXECUTIVE SUMMARY: Professional HD content suitable for broadcast. Video shows excellent technical quality with H.264 encoding at 1920x1080. RECOMMENDATIONS: Consider re-encoding to HEVC for 40% smaller files while maintaining quality. RISKS: No safety concerns detected.",
  "llm_enabled": true
}
```

#### Traditional Analysis (Like Basic FFprobe)
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

#### URL Analysis with GenAI
```bash
POST /api/v1/probe/url
Content-Type: application/json

{
  "url": "https://example.com/video.mp4",
  "include_llm": true,        // 🧠 Enable AI analysis
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

## 🧠 GenAI Analysis Examples (Core USP)

### 🎯 Why GenAI Analysis Changes Everything

**Traditional FFprobe Output:**
```json
{
  "codec_name": "h264",
  "width": 1920,
  "height": 1080,
  "bit_rate": "5000000"
}
```

**FFprobe API with GenAI Output:**
```json
{
  "llm_report": "EXECUTIVE SUMMARY: Professional HD broadcast content ready for delivery. Technical Analysis: H.264 encoding at optimal bitrate (5Mbps) for 1080p resolution. Quality Assessment: Excellent visual quality with no artifacts detected. Recommendations: 1) Consider HEVC encoding for 40% size reduction while maintaining quality. 2) Add closed captions for accessibility compliance. 3) Suitable for Netflix, YouTube, and broadcast distribution. Risk Assessment: Low technical risk, compliant with industry standards. Workflow Integration: Ready for immediate delivery pipeline integration."
}
```

### 🎥 Real-World GenAI Use Cases

#### 🚨 Safety Risk Detection
```bash
# Analyze content for PSE risks
curl -X POST \
  -F "file=@flashing_video.mp4" \
  -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file

# AI Response:
"CRITICAL ALERT: High photosensitive epilepsy risk detected. 
 Flashing patterns exceed safe thresholds (>3Hz). 
 REQUIRED ACTIONS: Add PSE warning, consider content modification."
```

#### 🏆 Quality Optimization
```bash
# Get optimization recommendations
curl -X POST \
  -F "file=@large_video.mp4" \
  -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file

# AI Response:
"OPTIMIZATION OPPORTUNITIES: File is 2.5GB for 10min duration. 
 RECOMMENDED: ffmpeg -i input.mp4 -c:v libx265 -crf 23 -c:a copy output.mp4 
 RESULT: 60% smaller file, same visual quality."
```

#### 📄 Executive Reporting
```bash
# Generate stakeholder-friendly reports
curl -X POST \
  -F "file=@corporate_video.mp4" \
  -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file

# AI Response:
"CLIENT REPORT: Your video meets all technical requirements for 
 social media distribution. Optimized for YouTube, Instagram, and 
 TikTok. No technical issues detected. Ready for immediate publication."
```

### 🛠️ HLS Analysis with GenAI
```bash
# Analyze HLS streams with AI insights
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "source": "/path/to/hls/directory/",
    "analyze_segments": true,
    "include_llm": true
  }' \
  http://localhost:8080/api/v1/probe/hls

# AI analyzes all .ts chunks and provides:
# - Quality ladder optimization suggestions
# - ABR streaming recommendations  
# - Platform compatibility assessment
# - Bandwidth efficiency insights
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
make valkey-shell       # Access Valkey

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

## 📚 Documentation

- **[🚀 Quick Start (Docker)](docker-image/QUICK_START.md)** - One-command deployment
- **[📖 Complete Documentation](docs/README.md)** - Full documentation index  
- **[📡 API Reference](docs/api/README.md)** - REST and GraphQL APIs
- **[🔍 QC Features](QC_ANALYSIS_LIST.md)** - All 20+ quality control categories
- **[🏢 Production Guide](docs/deployment/README.md)** - Enterprise deployment
- **[🤖 AI Setup](docs/tutorials/local-llm-setup.md)** - Local AI analysis setup

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