# Rendiff Probe

**Professional Video Analysis API & CLI - Powered by FFprobe**

A production-ready REST API and CLI tool for comprehensive video/audio file analysis, built on top of FFprobe with 19 professional QC categories and 26 content analyzers covering 121 industry-standard parameters.

[![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)](https://go.dev/)
[![QC Parameters](https://img.shields.io/badge/QC-121%20Parameters-blue.svg)](#quality-control-features)
[![Content Analyzers](https://img.shields.io/badge/Analyzers-26%20Parallel-green.svg)](#content-analysis)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](#quick-start)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## Acknowledgements

**Rendiff Probe uses [FFprobe](https://ffmpeg.org/ffprobe.html)** from the FFmpeg project as its core media analysis engine. FFprobe is a powerful multimedia stream analyzer that provides detailed information about media files.

- FFprobe is part of the [FFmpeg Project](https://ffmpeg.org/)
- FFprobe is licensed under the [LGPL/GPL license](https://ffmpeg.org/legal.html)
- This project wraps FFprobe functionality with enhanced QC analysis, REST API, and CLI interfaces

We are grateful to the FFmpeg community for developing and maintaining such a robust media analysis tool.

---

## Features

### Core Capabilities
- **Comprehensive FFprobe Analysis**: Full format, stream, frame, and packet analysis via FFprobe
- **121 Industry-Standard QC Parameters**: Complete broadcast and streaming quality control
- **26 Parallel Content Analyzers**: Real-time analysis using FFmpeg filters (signalstats, idet, astats, etc.)
- **REST API** (`rendiff-probe`): HTTP interface for video analysis
- **CLI Tool** (`rendiffprobe-cli`): Command-line tool for local analysis
- **GraphQL API**: Flexible query interface for advanced integrations
- **URL & HLS Analysis**: Direct URL probing and HLS stream analysis
- **Batch Processing**: Process multiple files/URLs in parallel
- **WebSocket Progress**: Real-time progress updates for long operations
- **LLM-Powered Insights**: AI-generated professional analysis reports
- **Docker Ready**: Production-ready containerized deployment
- **SQLite Embedded**: Zero-configuration database
- **Valkey/Redis Caching**: High-performance result caching

### Quality Control Analysis
Professional broadcast and streaming QC analysis covering:

**Header/Format Analysis**: Container validation, codec profiles, resolution, frame rate, bit depth, endianness
**Video Quality**: Baseband analysis (YMIN/YMAX/YAVG), gamut checking, blockiness, blurriness, noise, line errors
**Video Content**: Black frames, freeze frames, letterboxing, color bars, safe areas, field dominance, temporal complexity
**Audio Analysis**: EBU R128 loudness, clipping, silence, phase correlation, channel mapping, frequency analysis
**Broadcast Compliance**: HDR (HDR10/Dolby Vision/HLG), MXF validation, IMF compliance, transport stream analysis, timecode continuity
**Safety & Accessibility**: PSE flash detection, AFD analysis, stream disposition, data integrity

## Quick Start

### Prerequisites
- Docker 24.0+ with Docker Compose (for API)
- Go 1.24+ (for CLI)
- FFprobe installed (for CLI)
- 2GB RAM minimum (4GB recommended)

### Installation

```bash
# Clone the repository
git clone https://github.com/rendiffdev/rendiff-probe.git
cd rendiff-probe

# Quick start API (development mode)
make quick

# Or build CLI for local use
go build -o rendiffprobe-cli ./cmd/rendiffprobe-cli
```

### Using the CLI (`rendiffprobe-cli`)

```bash
# Analyze a video file with full QC report
rendiffprobe-cli analyze video.mp4 --format report

# Get JSON output for automation
rendiffprobe-cli analyze video.mp4 --format json --output result.json

# Quick file info
rendiffprobe-cli info video.mp4

# List all QC categories
rendiffprobe-cli categories
```

### Using the API (`rendiff-probe`)

Your API is now running at **http://localhost:8080**

```bash
# Check health
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "service": "rendiff-probe",
  "version": "2.0.0",
  "powered_by": "FFprobe (FFmpeg)",
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
  "ffprobe_validated": true
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
  "service": "rendiff-probe",
  "powered_by": "FFprobe (FFmpeg)",
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
  "ffprobe_validated": true
}
```

### Analyze File

```bash
POST /api/v1/probe/file
Content-Type: multipart/form-data
```

Upload a video/audio file for comprehensive analysis using FFprobe.

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

### GraphQL API

```bash
POST /api/v1/graphql
GET  /api/v1/graphql  # GraphiQL interface
```

Query and mutate via GraphQL for flexible data access.

## CLI Tool (`rendiffprobe-cli`)

The CLI provides the same powerful analysis capabilities without requiring a running API server.

### Commands

| Command | Description |
|---------|-------------|
| `analyze` | Full QC analysis with all 26 content analyzers |
| `categories` | List available QC analysis categories |
| `info` | Quick file information (basic metadata) |
| `version` | Show version information |

### Output Formats

- **report**: Human-readable comprehensive QC report
- **json**: Machine-readable JSON output
- **text**: Concise text summary

### Examples

```bash
# Full comprehensive report
rendiffprobe-cli analyze video.mp4 --format report

# JSON for automation/scripting
rendiffprobe-cli analyze video.mp4 --format json

# Save output to file
rendiffprobe-cli analyze video.mp4 --format json --output result.json

# Analyze multiple files
rendiffprobe-cli analyze video1.mp4 video2.mp4 --format text

# Quick metadata check
rendiffprobe-cli info video.mp4

# Set timeout for large files
rendiffprobe-cli analyze large_video.mp4 --timeout 300
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

### Production
```bash
make prod
```
- Full monitoring stack
- Authentication enabled
- Automated backups
- Best for: Production deployments

## Architecture

> **[View Full Architecture Documentation](docs/ARCHITECTURE.md)** - Detailed system design, component diagrams, and implementation patterns.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                     │
│    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│    │  REST API   │     │  GraphQL    │     │   CLI Tool  │                   │
│    └──────┬──────┘     └──────┬──────┘     └──────┬──────┘                   │
└───────────┼────────────────────┼────────────────────┼─────────────────────────┘
            │                    │                    │
┌───────────▼────────────────────▼────────────────────▼─────────────────────────┐
│                           MIDDLEWARE LAYER                                     │
│  [Recovery] → [RequestID] → [Logging] → [RateLimit] → [Auth] → [Validation]  │
└───────────────────────────────────┬───────────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼───────────────────────────────────────────┐
│                            SERVICE LAYER                                       │
│   ┌────────────────┐   ┌────────────────┐   ┌────────────────┐               │
│   │ AnalysisService│   │ ReportService  │   │ SecretRotation │               │
│   └───────┬────────┘   └────────────────┘   └────────────────┘               │
└───────────┼───────────────────────────────────────────────────────────────────┘
            │
┌───────────▼───────────────────────────────────────────────────────────────────┐
│                         ANALYSIS ENGINE                                        │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │                   Content Analyzer (26 Parallel Goroutines)             │  │
│  │   ┌────────────┬────────────┬────────────┬────────────┬────────────┐   │  │
│  │   │   Video    │   Audio    │   HDR      │  Broadcast │  Integrity │   │  │
│  │   │  Quality   │  Analysis  │  Analysis  │ Compliance │  Analysis  │   │  │
│  │   └─────┬──────┴─────┬──────┴─────┬──────┴─────┬──────┴─────┬──────┘   │  │
│  │         │            │            │            │            │           │  │
│  │   ┌─────▼────────────▼────────────▼────────────▼────────────▼──────┐   │  │
│  │   │              FFmpeg Filters (signalstats, idet, ebur128, etc)  │   │  │
│  │   └────────────────────────────────────────────────────────────────┘   │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                    │                                           │
│  ┌─────────────────────────────────▼───────────────────────────────────────┐  │
│  │                      FFprobe / FFmpeg Binaries                          │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼───────────────────────────────────────────┐
│                             DATA LAYER                                         │
│   ┌────────────────┐   ┌────────────────┐   ┌────────────────┐               │
│   │ SQLite (Store) │   │ Valkey (Cache) │   │ File Storage   │               │
│   └────────────────┘   └────────────────┘   └────────────────┘               │
└───────────────────────────────────────────────────────────────────────────────┘
```

### Key Components

| Layer | Components | Purpose |
|-------|------------|---------|
| **Client** | REST API, GraphQL, CLI | Multiple interface options |
| **Middleware** | Auth, RateLimit, Logging | Request processing pipeline |
| **Service** | Analysis, Reports, Secrets | Business logic encapsulation |
| **Analysis Engine** | 26 concurrent analyzers | FFmpeg-based quality analysis |
| **Data** | SQLite, Valkey, FileStore | Persistence and caching |

## Quality Control Features

### 26 Content Analyzers (121 Parameters)

| Category | Analyzers | FFmpeg Filters | Standards |
|----------|-----------|----------------|-----------|
| **Video Quality** | Baseband, Quality Score, Blockiness, Blurriness, Noise, Line Errors | signalstats, entropy | EBU R103 |
| **Video Content** | Black Frames, Freeze Frames, Letterbox, Color Bars, Safe Areas, Temporal Complexity, Field Dominance, Differential Frames | blackdetect, freezedetect, cropdetect, idet | ITU-R BT.814 |
| **Audio Quality** | Loudness (EBU R128), Clipping, Silence, Phase, Channel Mapping, Frequency Analysis | ebur128, astats, volumedetect | EBU R128, ITU-R BS.1770 |
| **HDR/Color** | HDR10, Dolby Vision, HLG, Color Space, Gamut | signalstats BRNG | Rec.2020, SMPTE ST 2086 |
| **Broadcast** | Timecode Continuity, MXF Validation, IMF Compliance, Transport Stream | - | SMPTE 12M, ST 377, ST 2067 |
| **Safety** | PSE Flash Detection, AFD Analysis, Test Tone Detection | - | ITU-R BT.1702, Ofcom |
| **Format** | Container, Codec, Resolution, Frame Rate, Bit Depth, Endianness | - | - |
| **Integrity** | Data Integrity, Stream Disposition, Dropout Detection | - | CRC32, MD5 |

### 19 Top-Level QC Categories

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
| Content Analysis | 26 parallel sub-analyzers | See above |
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
| `DB_PATH` | `./data/rendiff-probe.db` | SQLite database path |
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

# Build
go build -o rendiff-probe ./cmd/rendiff-probe
go build -o rendiffprobe-cli ./cmd/rendiffprobe-cli

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

## Documentation

### Core Documentation

| Document | Description | Audience |
|----------|-------------|----------|
| **[User Manual](docs/USER_MANUAL.md)** | Complete guide for using the API and CLI | End Users |
| **[Developer Guide](docs/DEVELOPER_GUIDE.md)** | Contributing to and extending the codebase | Developers |
| **[Architecture](docs/ARCHITECTURE.md)** | System design and component diagrams | Architects |

### Reference Documentation

| Document | Description |
|----------|-------------|
| **[QC Analysis List](docs/QC_ANALYSIS_LIST.md)** | All 19 QC categories with 121 parameters |
| **[API Reference](docs/api/)** | OpenAPI/Swagger documentation |
| **[Changelog](CHANGELOG.md)** | Version history and release notes |

### Quick Links

- [Installation Guide](docs/USER_MANUAL.md#installation)
- [CLI Quick Start](docs/USER_MANUAL.md#cli-tool-guide)
- [API Quick Start](docs/USER_MANUAL.md#rest-api-guide)
- [Contributing](docs/DEVELOPER_GUIDE.md#contributing)

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

### Third-Party Licenses

- **FFprobe/FFmpeg**: Licensed under LGPL/GPL - see [FFmpeg License](https://ffmpeg.org/legal.html)

## Support

- **Issues**: [GitHub Issues](https://github.com/rendiffdev/rendiff-probe/issues)
- **Documentation**: [docs/](docs/)

---

**Rendiff Probe** - Professional Video Analysis, Powered by FFprobe

Built for the video processing community
