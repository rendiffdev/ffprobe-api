# FFprobe API

**Professional Video Analysis API with Advanced QC Capabilities**

A production-ready REST API for comprehensive video/audio file analysis using FFprobe, with 19 professional quality control analysis categories.

[![Go Version](https://img.shields.io/badge/go-1.25.5-blue.svg)](https://go.dev/)
[![QC Analysis](https://img.shields.io/badge/QC-19%20Categories-blue.svg)](#quality-control-features)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](#quick-start)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Features

### Core Capabilities
- **Comprehensive FFprobe Analysis**: Full format, stream, frame, and packet analysis
- **19 Professional QC Categories**: Industry-standard quality control analysis
- **REST API**: Simple HTTP interface for video analysis
- **GraphQL API**: Flexible query interface for advanced integrations
- **URL & HLS Analysis**: Direct URL probing and HLS stream analysis
- **Batch Processing**: Process multiple files/URLs in parallel
- **WebSocket Progress**: Real-time progress updates for long operations
- **LLM-Powered Insights**: AI-generated professional analysis reports
- **Docker Ready**: Production-ready containerized deployment
- **SQLite Embedded**: Zero-configuration database
- **Valkey/Redis Caching**: High-performance result caching

### Quality Control Analysis
Professional broadcast and streaming QC analysis including:
- AFD (Active Format Description) Analysis
- Dead Pixel Detection
- PSE (Photosensitive Epilepsy) Flash Analysis
- HDR Analysis (HDR10, Dolby Vision, HLG)
- Timecode Analysis (SMPTE)
- MXF Format Validation
- IMF Compliance Checking
- Transport Stream Analysis
- And 10 more categories...

📋 **[Complete QC Analysis List](docs/QC_ANALYSIS_LIST.md)**

## Quick Start

### Prerequisites
- Docker 24.0+ with Docker Compose
- 2GB RAM minimum (4GB recommended)
- 5GB disk space

### Installation

```bash
# Clone the repository
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# Quick start (development mode)
make quick

# Or minimal deployment
make minimal
```

Your API is now running at **http://localhost:8080**

### Verify Installation

```bash
# Check health
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "service": "ffprobe-api",
  "version": "2.0.0",
  "features": {
    "file_probe": true,
    "url_probe": true,
    "hls_analysis": true,
    "batch_processing": true,
    "websocket": true,
    "graphql": true,
    "llm_insights": true
  },
  "qc_tools": ["AFD Analysis", "Dead Pixel Detection", ...],
  "ffmpeg_validated": true
}
```

## API Reference

### Health Check

```bash
GET /health
```

Returns service health status and available QC tools.

**Response:**
```json
{
  "status": "healthy",
  "service": "ffprobe-api-core",
  "qc_tools": [
    "AFD Analysis",
    "Dead Pixel Detection",
    "PSE Flash Analysis",
    "HDR Analysis",
    "Audio Wrapping Analysis",
    "Endianness Detection",
    "Codec Analysis",
    "Container Validation",
    "Resolution Analysis",
    "Frame Rate Analysis",
    "Bitdepth Analysis",
    "Timecode Analysis",
    "MXF Analysis",
    "IMF Compliance",
    "Transport Stream Analysis",
    "Content Analysis",
    "Enhanced Analysis",
    "Stream Disposition Analysis",
    "Data Integrity Analysis"
  ],
  "ffmpeg_validated": true
}
```

### Analyze File

```bash
POST /api/v1/probe/file
Content-Type: multipart/form-data
```

Upload a video/audio file for comprehensive analysis.

**Request:**
```bash
curl -X POST \
  -F "file=@video.mp4" \
  http://localhost:8080/api/v1/probe/file
```

**Response:**
```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "video.mp4",
  "size": 1048576,
  "result": {
    "format": {
      "filename": "/tmp/upload_1234567890_video.mp4",
      "nb_streams": 2,
      "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
      "duration": "60.500000",
      "bit_rate": "5000000"
    },
    "streams": [
      {
        "index": 0,
        "codec_type": "video",
        "codec_name": "h264",
        "width": 1920,
        "height": 1080,
        "r_frame_rate": "30/1"
      },
      {
        "index": 1,
        "codec_type": "audio",
        "codec_name": "aac",
        "sample_rate": "48000",
        "channels": 2
      }
    ],
    "enhanced_analysis": {
      "timecode_analysis": {...},
      "hdr_analysis": {...},
      "codec_analysis": {...},
      "data_integrity": {...}
    }
  }
}
```

### Analyze URL

```bash
POST /api/v1/probe/url
Content-Type: application/json
```

Analyze a video file from a URL without uploading.

**Request:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/video.mp4", "include_llm": true}' \
  http://localhost:8080/api/v1/probe/url
```

### HLS Stream Analysis

```bash
POST /api/v1/probe/hls
Content-Type: application/json
```

Analyze HLS streams for quality, compliance, and performance.

**Request:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "manifest_url": "https://example.com/stream.m3u8",
    "analyze_segments": true,
    "analyze_quality": true,
    "validate_compliance": true,
    "max_segments": 10
  }' \
  http://localhost:8080/api/v1/probe/hls
```

### Batch Processing

```bash
POST /api/v1/batch/analyze    # Start batch job
GET  /api/v1/batch/status/:id # Get job status
```

Process multiple files or URLs in parallel.

**Request:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://example.com/video1.mp4", "https://example.com/video2.mp4"],
    "include_llm": false
  }' \
  http://localhost:8080/api/v1/batch/analyze
```

**Response:**
```json
{
  "status": "accepted",
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "total": 2,
  "status_url": "/api/v1/batch/status/550e8400-e29b-41d4-a716-446655440000",
  "ws_url": "/api/v1/ws/progress/550e8400-e29b-41d4-a716-446655440000"
}
```

### WebSocket Progress

```bash
GET /api/v1/ws/progress/:id
```

Connect via WebSocket to receive real-time progress updates for batch jobs.

### GraphQL API

```bash
POST /api/v1/graphql
GET  /api/v1/graphql  # GraphiQL interface
```

Query and mutate via GraphQL for flexible data access.

**Example Query:**
```graphql
query {
  health {
    status
    version
  }
}

mutation {
  analyzeURL(url: "https://example.com/video.mp4", include_llm: true) {
    id
    filename
    status
    llm_report
  }
}
```

### LLM Insights

Add `include_llm=true` to any analysis endpoint to get AI-powered insights:

```bash
# With file upload
curl -X POST -F "file=@video.mp4" -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file

# With URL
curl -X POST -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/video.mp4", "include_llm": true}' \
  http://localhost:8080/api/v1/probe/url
```

### FFmpeg Version Management (Admin)

```bash
GET  /admin/ffmpeg/version     # Get current FFmpeg version
POST /admin/ffmpeg/check       # Check for updates
POST /admin/ffmpeg/update      # Update FFmpeg
```

## Deployment Modes

### Minimal (Development/Testing)
```bash
make minimal
```
- Core services only: API + Valkey + Ollama
- Memory: ~2-3GB
- Best for: Development, testing

### Quick Start
```bash
make quick
```
- Ready in 2 minutes
- No authentication required
- Best for: Quick testing, demos

### Development
```bash
make dev
```
- Hot reload enabled
- Admin tools included
- Best for: Development

### Production
```bash
make prod
```
- Full monitoring stack
- Authentication enabled
- Automated backups
- Best for: Production deployments

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   FFprobe API   │───▶│     SQLite      │    │     Valkey      │
│   (Go/Gin)      │    │ (Embedded DB)   │    │ (Redis Cache)   │
└────────┬────────┘    └─────────────────┘    └─────────────────┘
         │
         ▼
┌─────────────────┐    ┌─────────────────┐
│     Ollama      │    │     FFmpeg      │
│  (AI Models)    │    │   (BtbN Build)  │
└─────────────────┘    └─────────────────┘
```

## Quality Control Features

### 19 QC Analysis Categories

| Category | Description | Standards |
|----------|-------------|-----------|
| AFD Analysis | Active Format Description | ITU-R BT.1868 |
| Dead Pixel Detection | Pixel defect analysis | Computer Vision |
| PSE Flash Analysis | Epilepsy safety | ITC/Ofcom, ITU-R BT.1702 |
| HDR Analysis | HDR content validation | HDR10, Dolby Vision, HLG |
| Audio Wrapping | Professional audio formats | BWF, RF64, AES3 |
| Endianness Detection | Binary format compatibility | - |
| Codec Analysis | Codec validation | - |
| Container Validation | Format compliance | MP4, MKV, MOV |
| Resolution Analysis | Aspect ratio validation | - |
| Frame Rate Analysis | Temporal accuracy | Broadcast standards |
| Bitdepth Analysis | Color depth validation | 8/10/12-bit |
| Timecode Analysis | SMPTE timecode | SMPTE 12M |
| MXF Analysis | Broadcast format | SMPTE ST 377 |
| IMF Compliance | Distribution format | SMPTE ST 2067 |
| Transport Stream | MPEG-TS analysis | MPEG-TS |
| Content Analysis | Scene/motion analysis | - |
| Enhanced Analysis | Quality metrics | - |
| Stream Disposition | Accessibility | Section 508 |
| Data Integrity | Error/hash validation | CRC32, MD5 |

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API port |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `FFPROBE_PATH` | `ffprobe` | Path to FFprobe binary |
| `DB_PATH` | `./data/ffprobe.db` | SQLite database path |
| `VALKEY_URL` | `valkey:6379` | Valkey/Redis connection |

### Security Configuration

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | JWT signing secret (required in production) |
| `API_KEY` | API key for authentication |
| `RATE_LIMIT_RPM` | Rate limit per minute |

## Management Commands

```bash
# Service Management
make start      # Start all services
make stop       # Stop all services
make restart    # Restart services
make status     # Show status
make logs       # View logs
make health     # Check health

# Development
make test-unit       # Run unit tests
make test-coverage   # Run tests with coverage
make lint            # Run linter

# Maintenance
make update     # Update services
make backup     # Create backup
make clean      # Clean everything
```

## Testing

```bash
# Run all tests
make test-unit

# Run with coverage
make test-coverage

# Run specific package
go test -v ./internal/ffmpeg/...

# Run with race detection
make test-race
```

Current test coverage:
- config: 50.3%
- middleware: 22.3%
- ffmpeg: 6.0%

## Troubleshooting

### Services Won't Start

```bash
# Check Docker
docker --version
docker compose version

# View logs
make logs

# Reset everything
make clean && make quick
```

### Port Conflicts

```bash
# Check port usage
lsof -i :8080

# Use different port
PORT=8081 make quick
```

### FFprobe Validation Fails

```bash
# Check FFprobe in container
docker compose exec api ffprobe -version

# Rebuild container
make clean && make quick
```

## Documentation

- **[QC Analysis List](docs/QC_ANALYSIS_LIST.md)** - All 19 QC categories
- **[Changelog](CHANGELOG.md)** - Version history
- **[TODO](TODO.md)** - Roadmap and tasks

## Roadmap

### Completed Features (v2.0.0)

- [x] GraphQL API endpoint
- [x] URL-based file analysis
- [x] HLS stream analysis endpoint
- [x] Batch processing API
- [x] WebSocket progress streaming
- [x] LLM-powered analysis insights

### Planned Features

- [ ] Webhook callbacks for async processing
- [ ] DASH stream analysis
- [ ] Compare multiple files side-by-side
- [ ] Custom QC rule definitions

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test-unit`)
5. Commit your changes
6. Push to the branch
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **Documentation**: [docs/](docs/)

---

**Built for the video processing community**
