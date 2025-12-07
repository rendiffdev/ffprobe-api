# Rendiff Probe User Manual

**Complete guide for using the Rendiff Probe API and CLI tool**

---

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [CLI Tool Guide](#cli-tool-guide)
5. [REST API Guide](#rest-api-guide)
6. [GraphQL API Guide](#graphql-api-guide)
7. [Understanding Analysis Results](#understanding-analysis-results)
8. [Quality Control Categories](#quality-control-categories)
9. [Batch Processing](#batch-processing)
10. [HLS Stream Analysis](#hls-stream-analysis)
11. [Configuration](#configuration)
12. [Troubleshooting](#troubleshooting)
13. [FAQ](#faq)

---

## Introduction

Rendiff Probe is a professional video and audio analysis platform that provides comprehensive quality control (QC) analysis for media files. Built on top of FFprobe and FFmpeg, it offers:

- **121 industry-standard QC parameters** across 19 categories
- **26 parallel content analyzers** for fast processing
- **Multiple interfaces**: REST API, GraphQL, and CLI
- **Broadcast-grade analysis** meeting EBU, ITU, and SMPTE standards

### Who Is This For?

- **Video Engineers**: QC validation for broadcast and streaming
- **Post-Production Teams**: Quality assurance before delivery
- **Streaming Platforms**: Automated content analysis
- **Archivists**: Media file validation and cataloging
- **Developers**: Integration into media workflows

---

## Installation

### Option 1: Docker (Recommended for API)

```bash
# Clone repository
git clone https://github.com/rendiffdev/rendiff-probe.git
cd rendiff-probe

# Start with Docker Compose
make quick

# Verify installation
curl http://localhost:8080/health
```

### Option 2: Build from Source

**Prerequisites:**
- Go 1.24 or later
- FFmpeg/FFprobe 6.0 or later

```bash
# Clone repository
git clone https://github.com/rendiffdev/rendiff-probe.git
cd rendiff-probe

# Build API server
go build -o rendiff-probe ./cmd/rendiff-probe

# Build CLI tool
go build -o rendiffprobe-cli ./cmd/rendiffprobe-cli

# Verify FFprobe is installed
ffprobe -version
```

### Option 3: Pre-built Binaries

Download from [Releases](https://github.com/rendiffdev/rendiff-probe/releases):

```bash
# macOS (Intel)
curl -LO https://github.com/rendiffdev/rendiff-probe/releases/latest/download/rendiffprobe-cli-darwin-amd64
chmod +x rendiffprobe-cli-darwin-amd64
mv rendiffprobe-cli-darwin-amd64 /usr/local/bin/rendiffprobe-cli

# macOS (Apple Silicon)
curl -LO https://github.com/rendiffdev/rendiff-probe/releases/latest/download/rendiffprobe-cli-darwin-arm64
chmod +x rendiffprobe-cli-darwin-arm64
mv rendiffprobe-cli-darwin-arm64 /usr/local/bin/rendiffprobe-cli

# Linux
curl -LO https://github.com/rendiffdev/rendiff-probe/releases/latest/download/rendiffprobe-cli-linux-amd64
chmod +x rendiffprobe-cli-linux-amd64
sudo mv rendiffprobe-cli-linux-amd64 /usr/local/bin/rendiffprobe-cli
```

---

## Quick Start

### CLI Quick Start

```bash
# Analyze a video file
rendiffprobe-cli analyze video.mp4

# Get JSON output
rendiffprobe-cli analyze video.mp4 --format json

# Quick file info
rendiffprobe-cli info video.mp4
```

### API Quick Start

```bash
# Start the server
./rendiff-probe

# Analyze a file via API
curl -X POST \
  -F "file=@video.mp4" \
  http://localhost:8080/api/v1/probe/file

# Analyze a URL
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/video.mp4"}' \
  http://localhost:8080/api/v1/probe/url
```

---

## CLI Tool Guide

The CLI tool (`rendiffprobe-cli`) provides powerful analysis capabilities without requiring a running server.

### Available Commands

| Command | Description |
|---------|-------------|
| `analyze` | Full QC analysis with all 26 content analyzers |
| `info` | Quick file information (basic metadata only) |
| `categories` | List all available QC analysis categories |
| `version` | Show version information |

### Analyze Command

The `analyze` command performs comprehensive quality control analysis.

**Syntax:**
```bash
rendiffprobe-cli analyze <file> [flags]
```

**Flags:**
| Flag | Description | Default |
|------|-------------|---------|
| `--format`, `-f` | Output format: `report`, `json`, `text` | `report` |
| `--output`, `-o` | Output file path | stdout |
| `--timeout`, `-t` | Analysis timeout in seconds | 120 |
| `--verbose`, `-v` | Enable verbose output | false |

**Examples:**

```bash
# Full comprehensive report
rendiffprobe-cli analyze video.mp4 --format report

# Machine-readable JSON for automation
rendiffprobe-cli analyze video.mp4 --format json

# Save to file
rendiffprobe-cli analyze video.mp4 --format json --output result.json

# Concise text summary
rendiffprobe-cli analyze video.mp4 --format text

# Analyze with extended timeout for large files
rendiffprobe-cli analyze large_video.mp4 --timeout 300

# Analyze multiple files
rendiffprobe-cli analyze video1.mp4 video2.mp4 video3.mp4
```

### Info Command

Quick metadata extraction without full QC analysis.

```bash
# Basic file info
rendiffprobe-cli info video.mp4

# Output example:
# File: video.mp4
# Duration: 00:01:30.500
# Container: MP4 (mov,mp4,m4a,3gp,3g2,mj2)
# Video: h264, 1920x1080, 30fps, 5.0 Mbps
# Audio: aac, 48000 Hz, stereo, 128 kbps
```

### Categories Command

List all available QC categories and their descriptions.

```bash
rendiffprobe-cli categories

# Output:
# Available QC Categories:
# 1. AFD Analysis - Active Format Description (ITU-R BT.1868)
# 2. Dead Pixel Detection - Pixel defect analysis
# 3. PSE Flash Analysis - Epilepsy safety (ITC/Ofcom)
# ... (19 categories total)
```

### Output Formats

#### Report Format (Default)

Human-readable comprehensive report:

```
================================================================================
                         RENDIFF PROBE - VIDEO QC REPORT
================================================================================

FILE INFORMATION
--------------------------------------------------------------------------------
Filename:     video.mp4
Duration:     00:01:30.500
File Size:    125.5 MB
Container:    MP4 (mov,mp4,m4a,3gp,3g2,mj2)

VIDEO STREAM
--------------------------------------------------------------------------------
Codec:        H.264 (High Profile, Level 4.1)
Resolution:   1920x1080 (16:9)
Frame Rate:   29.97 fps
Bit Rate:     5,000 kbps
Bit Depth:    8-bit
Color Space:  YUV 4:2:0

AUDIO STREAM
--------------------------------------------------------------------------------
Codec:        AAC-LC
Sample Rate:  48000 Hz
Channels:     Stereo (2.0)
Bit Rate:     128 kbps

QUALITY CONTROL ANALYSIS
--------------------------------------------------------------------------------

Video Quality
  Baseband Analysis:
    YMIN: 16 (OK)
    YMAX: 235 (OK)
    YAVG: 125.4
    SATMIN: 0
    SATMAX: 180
  Quality Score: 85.2/100
  Blockiness: Low (2.3)
  Blurriness: Low (1.8)

Content Analysis
  Black Frames: None detected
  Freeze Frames: None detected
  Letterboxing: Not detected
  Color Bars: None

Audio Analysis
  Loudness (EBU R128):
    Integrated: -24.0 LUFS (OK)
    True Peak: -1.5 dBTP (OK)
    LRA: 12.5 LU
  Clipping: None detected
  Silence: None detected
  Phase: In-phase (correlation: 0.95)

SUMMARY
--------------------------------------------------------------------------------
Overall Status: PASS
Warnings: 0
Errors: 0

================================================================================
```

#### JSON Format

Structured JSON for programmatic access:

```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "video.mp4",
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": 90.5,
  "format": {
    "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
    "bit_rate": "5000000",
    "size": "131596288"
  },
  "streams": [
    {
      "codec_type": "video",
      "codec_name": "h264",
      "width": 1920,
      "height": 1080,
      "r_frame_rate": "30000/1001"
    }
  ],
  "content_analysis": {
    "black_frames": {
      "has_black_frames": false,
      "total_duration": 0
    },
    "loudness": {
      "integrated_loudness": -24.0,
      "true_peak": -1.5,
      "loudness_range": 12.5
    }
  },
  "qc_status": "PASS"
}
```

#### Text Format

Concise summary:

```
video.mp4: PASS
Duration: 1:30.500 | 1920x1080 @ 29.97fps | H.264/AAC
Loudness: -24.0 LUFS | Peak: -1.5 dBTP | No issues detected
```

---

## REST API Guide

### Base URL

```
http://localhost:8080/api/v1
```

### Authentication

For development, authentication is disabled. For production, use JWT or API keys:

```bash
# JWT Token
curl -H "Authorization: Bearer <your-jwt-token>" http://localhost:8080/api/v1/probe/url

# API Key
curl -H "X-API-Key: <your-api-key>" http://localhost:8080/api/v1/probe/url
```

### Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health check |
| POST | `/api/v1/probe/file` | Analyze uploaded file |
| POST | `/api/v1/probe/url` | Analyze file from URL |
| POST | `/api/v1/probe/hls` | Analyze HLS stream |
| POST | `/api/v1/batch/analyze` | Start batch analysis |
| GET | `/api/v1/batch/status/:id` | Get batch job status |
| POST | `/api/v1/graphql` | GraphQL endpoint |
| GET | `/api/v1/graphql` | GraphiQL interface |

### Health Check

```bash
GET /health
```

**Response:**
```json
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
  "qc_tools": [
    "AFD Analysis",
    "Dead Pixel Detection",
    "PSE Flash Analysis",
    "HDR Analysis",
    "Content Analysis"
  ],
  "ffprobe_validated": true
}
```

### Analyze File Upload

```bash
POST /api/v1/probe/file
Content-Type: multipart/form-data
```

**Request:**
```bash
curl -X POST \
  -F "file=@video.mp4" \
  -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file
```

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| file | file | Yes | Media file to analyze |
| include_llm | bool | No | Include AI-generated insights |

**Response:**
```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "filename": "video.mp4",
  "size": 131596288,
  "content_type": "video/mp4",
  "result": {
    "format": { ... },
    "streams": [ ... ],
    "content_analysis": { ... },
    "enhanced_analysis": { ... }
  },
  "llm_insights": "Professional analysis summary...",
  "processing_time_ms": 15234
}
```

### Analyze URL

```bash
POST /api/v1/probe/url
Content-Type: application/json
```

**Request:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/video.mp4",
    "include_llm": true
  }' \
  http://localhost:8080/api/v1/probe/url
```

**Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| url | string | Yes | URL of media file |
| include_llm | bool | No | Include AI-generated insights |
| timeout | int | No | Request timeout in seconds |

### Error Responses

All error responses follow this format:

```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "details": "Additional details if available"
}
```

**Common Error Codes:**
| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_INPUT` | 400 | Invalid request parameters |
| `FILE_TOO_LARGE` | 400 | File exceeds size limit |
| `UNSUPPORTED_FORMAT` | 400 | Unsupported media format |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `ANALYSIS_FAILED` | 500 | Analysis processing error |

### Rate Limiting

Rate limits are returned in response headers:

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Requests allowed per minute |
| `X-RateLimit-Remaining` | Requests remaining |
| `X-RateLimit-Reset` | Unix timestamp when limit resets |
| `Retry-After` | Seconds until you can retry |

**Default Limits:**
| Role | Per Minute | Per Hour | Per Day |
|------|------------|----------|---------|
| User | 60 | 1,000 | 10,000 |
| Pro | 180 | 3,000 | 30,000 |
| Premium | 300 | 5,000 | 50,000 |
| Admin | 600 | 10,000 | 100,000 |

---

## GraphQL API Guide

### Endpoint

```
POST /api/v1/graphql
GET  /api/v1/graphql  (GraphiQL Interface)
```

### Query Examples

**Get Analysis by ID:**
```graphql
query GetAnalysis($id: ID!) {
  analysis(id: $id) {
    id
    filename
    duration
    format {
      formatName
      bitRate
    }
    streams {
      codecType
      codecName
      width
      height
    }
    contentAnalysis {
      blackFrames {
        hasBlackFrames
        totalDuration
      }
      loudness {
        integratedLoudness
        truePeak
      }
    }
  }
}
```

**Analyze URL:**
```graphql
mutation AnalyzeURL($url: String!) {
  analyzeUrl(url: $url) {
    analysisId
    filename
    result {
      format {
        formatName
        duration
      }
    }
  }
}
```

**List Analyses:**
```graphql
query ListAnalyses($limit: Int, $offset: Int) {
  analyses(limit: $limit, offset: $offset) {
    id
    filename
    createdAt
    status
  }
}
```

### Using GraphiQL

1. Open `http://localhost:8080/api/v1/graphql` in your browser
2. The GraphiQL interface provides:
   - Interactive query builder
   - Schema documentation
   - Query history
   - Variable editor

---

## Understanding Analysis Results

### Format Information

Basic container and file information:

```json
{
  "format": {
    "filename": "/path/to/video.mp4",
    "nb_streams": 2,
    "nb_programs": 0,
    "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
    "format_long_name": "QuickTime / MOV",
    "start_time": "0.000000",
    "duration": "90.500000",
    "size": "131596288",
    "bit_rate": "11632000",
    "probe_score": 100
  }
}
```

### Stream Information

Details for each stream (video, audio, subtitle, data):

```json
{
  "streams": [
    {
      "index": 0,
      "codec_type": "video",
      "codec_name": "h264",
      "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
      "profile": "High",
      "level": 41,
      "width": 1920,
      "height": 1080,
      "coded_width": 1920,
      "coded_height": 1088,
      "pix_fmt": "yuv420p",
      "color_range": "tv",
      "color_space": "bt709",
      "r_frame_rate": "30000/1001",
      "avg_frame_rate": "30000/1001",
      "bit_rate": "10000000"
    },
    {
      "index": 1,
      "codec_type": "audio",
      "codec_name": "aac",
      "codec_long_name": "AAC (Advanced Audio Coding)",
      "profile": "LC",
      "sample_rate": "48000",
      "channels": 2,
      "channel_layout": "stereo",
      "bit_rate": "128000"
    }
  ]
}
```

### Content Analysis Results

Detailed QC analysis results:

```json
{
  "content_analysis": {
    "black_frames": {
      "has_black_frames": false,
      "frames": [],
      "total_duration": 0,
      "average_duration": 0
    },
    "freeze_frames": {
      "has_freeze_frames": false,
      "frames": [],
      "total_duration": 0
    },
    "baseband_analysis": {
      "ymin": 16,
      "ymax": 235,
      "yavg": 125.4,
      "umin": 64,
      "umax": 192,
      "vmin": 64,
      "vmax": 192,
      "satmin": 0,
      "satmax": 180,
      "satavg": 45.2,
      "out_of_range_pixels": 0
    },
    "loudness": {
      "integrated_loudness": -24.0,
      "true_peak": -1.5,
      "loudness_range": 12.5,
      "sample_peak": -3.2
    },
    "silence_detection": {
      "has_silence": false,
      "segments": [],
      "total_duration": 0
    },
    "audio_clipping": {
      "has_clipping": false,
      "clips_detected": 0,
      "max_peak": 0.85
    },
    "phase_correlation": {
      "average_correlation": 0.95,
      "min_correlation": 0.82,
      "out_of_phase": false
    }
  }
}
```

---

## Quality Control Categories

### 1. Video Quality Analysis

**Baseband Analysis (signalstats)**
- YMIN/YMAX/YAVG: Luminance range (legal: 16-235)
- UMIN/UMAX, VMIN/VMAX: Chroma range
- SATMIN/SATMAX: Saturation levels
- Out-of-range pixel detection

**Quality Metrics**
- Video Quality Score (0-100)
- Blockiness detection
- Blurriness analysis
- Noise measurement
- Line error detection

### 2. Video Content Analysis

| Analyzer | FFmpeg Filter | Description |
|----------|---------------|-------------|
| Black Frame Detection | blackdetect | Identifies black segments |
| Freeze Frame Detection | freezedetect | Detects static frames |
| Letterbox Detection | cropdetect | Identifies black bars |
| Color Bars Detection | Custom | Detects test patterns |
| Safe Area Analysis | Custom | Title/action safe zones |
| Field Dominance | idet | Interlace field order |
| Temporal Complexity | Custom | Motion analysis |

### 3. Audio Analysis

| Analyzer | Standard | Description |
|----------|----------|-------------|
| Loudness Metering | EBU R128 | Integrated, momentary, short-term |
| True Peak | ITU-R BS.1770 | Inter-sample peak detection |
| Clipping Detection | astats | Digital clipping events |
| Silence Detection | silencedetect | Silent segments |
| Phase Correlation | aphasemeter | Stereo phase issues |
| Channel Mapping | Custom | Channel configuration |
| Test Tone Detection | Custom | Reference tone identification |

### 4. HDR Analysis

- HDR10 metadata validation
- Dolby Vision detection
- HLG (Hybrid Log-Gamma) support
- MaxCLL/MaxFALL values
- Color volume analysis

### 5. Broadcast Compliance

| Category | Standard | Description |
|----------|----------|-------------|
| Timecode Analysis | SMPTE 12M | TC continuity validation |
| MXF Validation | SMPTE ST 377 | OP1a/OP-Atom compliance |
| IMF Compliance | SMPTE ST 2067 | CPL/OPL validation |
| Transport Stream | MPEG-TS | PID/PCR analysis |

### 6. Safety & Accessibility

- **PSE Flash Detection**: Epilepsy safety per ITC/Ofcom guidelines
- **AFD Analysis**: Active Format Description (ITU-R BT.1868)
- **Stream Disposition**: Accessibility track identification

---

## Batch Processing

Process multiple files in parallel for efficiency.

### Start Batch Job

```bash
POST /api/v1/batch/analyze
Content-Type: application/json
```

**Request:**
```json
{
  "items": [
    {"url": "https://example.com/video1.mp4"},
    {"url": "https://example.com/video2.mp4"},
    {"url": "https://example.com/video3.mp4"}
  ],
  "options": {
    "include_llm": false,
    "parallel_limit": 5
  }
}
```

**Response:**
```json
{
  "batch_id": "b550e840-e29b-41d4-a716-446655440000",
  "status": "processing",
  "total_items": 3,
  "completed": 0,
  "failed": 0
}
```

### Check Batch Status

```bash
GET /api/v1/batch/status/:batch_id
```

**Response:**
```json
{
  "batch_id": "b550e840-e29b-41d4-a716-446655440000",
  "status": "completed",
  "total_items": 3,
  "completed": 3,
  "failed": 0,
  "results": [
    {
      "item_id": 0,
      "status": "success",
      "analysis_id": "a1234...",
      "filename": "video1.mp4"
    },
    {
      "item_id": 1,
      "status": "success",
      "analysis_id": "a5678...",
      "filename": "video2.mp4"
    },
    {
      "item_id": 2,
      "status": "success",
      "analysis_id": "a9012...",
      "filename": "video3.mp4"
    }
  ],
  "processing_time_ms": 45230
}
```

### Batch Status Values

| Status | Description |
|--------|-------------|
| `pending` | Job queued, not started |
| `processing` | Currently analyzing files |
| `completed` | All items processed |
| `failed` | Job failed (see errors) |
| `partial` | Some items failed |

---

## HLS Stream Analysis

Analyze HTTP Live Streaming (HLS) manifests and segments.

### Analyze HLS Stream

```bash
POST /api/v1/probe/hls
Content-Type: application/json
```

**Request:**
```json
{
  "manifest_url": "https://example.com/stream/master.m3u8",
  "analyze_segments": true,
  "analyze_quality": true,
  "validate_compliance": true,
  "max_segments": 10
}
```

**Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| manifest_url | string | Required | HLS manifest URL |
| analyze_segments | bool | true | Analyze individual segments |
| analyze_quality | bool | true | Run QC on segments |
| validate_compliance | bool | false | Check HLS compliance |
| max_segments | int | 5 | Max segments to analyze |

**Response:**
```json
{
  "manifest_url": "https://example.com/stream/master.m3u8",
  "type": "master",
  "variants": [
    {
      "bandwidth": 5000000,
      "resolution": "1920x1080",
      "codecs": "avc1.640028,mp4a.40.2",
      "url": "1080p/playlist.m3u8"
    },
    {
      "bandwidth": 2500000,
      "resolution": "1280x720",
      "codecs": "avc1.64001f,mp4a.40.2",
      "url": "720p/playlist.m3u8"
    }
  ],
  "segment_analysis": [
    {
      "segment_url": "1080p/segment_001.ts",
      "duration": 6.006,
      "analysis": { ... }
    }
  ],
  "compliance": {
    "valid": true,
    "issues": []
  }
}
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server port |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `FFPROBE_PATH` | `ffprobe` | Path to FFprobe binary |
| `FFMPEG_PATH` | `ffmpeg` | Path to FFmpeg binary |
| `DB_PATH` | `./data/rendiff-probe.db` | SQLite database path |
| `VALKEY_URL` | `valkey:6379` | Valkey/Redis URL |
| `MAX_FILE_SIZE` | `5368709120` | Max upload size (5GB) |
| `ANALYSIS_TIMEOUT` | `120` | Analysis timeout in seconds |

### Security Configuration

| Variable | Description |
|----------|-------------|
| `JWT_SECRET` | JWT signing secret (required in production) |
| `API_KEY` | API key for authentication |
| `CORS_ORIGINS` | Allowed CORS origins (comma-separated) |
| `RATE_LIMIT_RPM` | Requests per minute limit |

### Docker Compose Configuration

```yaml
# compose.yaml
services:
  rendiff-probe:
    image: rendiff-probe:latest
    environment:
      - PORT=8080
      - LOG_LEVEL=info
      - DB_PATH=/data/rendiff-probe.db
      - VALKEY_URL=valkey:6379
    volumes:
      - ./data:/data
      - ./uploads:/uploads
    ports:
      - "8080:8080"
    depends_on:
      - valkey

  valkey:
    image: valkey/valkey:latest
    volumes:
      - valkey-data:/data

volumes:
  valkey-data:
```

---

## Troubleshooting

### Common Issues

#### FFprobe Not Found

```
Error: FFprobe binary not found
```

**Solution:**
```bash
# Install FFmpeg (includes FFprobe)
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt install ffmpeg

# Set path if not in PATH
export FFPROBE_PATH=/usr/local/bin/ffprobe
```

#### Analysis Timeout

```
Error: Analysis timed out after 120 seconds
```

**Solution:**
- Increase timeout: `--timeout 300` (CLI) or `ANALYSIS_TIMEOUT=300` (API)
- Check if file is accessible
- Verify sufficient system resources

#### Rate Limit Exceeded

```json
{"error": "Rate limit exceeded", "code": "RATE_LIMIT_EXCEEDED"}
```

**Solution:**
- Wait for `Retry-After` seconds
- Request higher rate limits
- Implement request queuing

#### File Too Large

```json
{"error": "File too large", "code": "FILE_TOO_LARGE"}
```

**Solution:**
- Use URL analysis instead of upload
- Adjust `MAX_FILE_SIZE` environment variable
- Split large files into smaller segments

### Debug Mode

Enable debug logging for troubleshooting:

```bash
# CLI
rendiffprobe-cli analyze video.mp4 --verbose

# API
export LOG_LEVEL=debug
./rendiff-probe
```

### Checking Service Health

```bash
# API health check
curl http://localhost:8080/health | jq .

# Check FFprobe availability
ffprobe -version

# Check Valkey connectivity
redis-cli -h localhost -p 6379 ping
```

---

## FAQ

### General Questions

**Q: What file formats are supported?**

A: Rendiff Probe supports all formats supported by FFprobe, including:
- Video: MP4, MKV, MOV, AVI, MXF, TS, WebM, ProRes, DNxHD
- Audio: WAV, MP3, AAC, FLAC, PCM, BWF
- Containers: HLS (m3u8), DASH (mpd)

**Q: How long does analysis take?**

A: Analysis time depends on file size and content:
- Basic probe: 1-3 seconds
- Standard QC (all analyzers): 10-30 seconds
- Full analysis with all 121 parameters: 30-60 seconds

**Q: Can I analyze remote files without downloading?**

A: Yes, use the URL endpoint. FFprobe can analyze remote files via HTTP/HTTPS without full download.

### Technical Questions

**Q: How many files can I process in parallel?**

A: Batch processing supports up to 100 files per request. Parallel limit is configurable (default: 5).

**Q: Is the analysis cached?**

A: Yes, results are cached in Valkey/Redis with configurable TTL. Identical requests return cached results.

**Q: What are the system requirements?**

A: Minimum requirements:
- 2GB RAM
- 2 CPU cores
- 10GB disk space
- FFmpeg 6.0+

Recommended:
- 4GB RAM
- 4 CPU cores
- SSD storage

### Integration Questions

**Q: How do I integrate with my existing workflow?**

A: Options include:
1. REST API for HTTP integrations
2. GraphQL for flexible queries
3. CLI for shell scripts and automation
4. WebSocket for real-time progress

**Q: Is there a webhook for completed analyses?**

A: Webhook support is planned. Currently, use polling or WebSocket for status updates.

**Q: Can I customize the QC thresholds?**

A: Custom thresholds are planned for a future release. Currently, thresholds follow broadcast standards (EBU, ITU, SMPTE).

---

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md)
- [Developer Guide](DEVELOPER_GUIDE.md)
- [QC Analysis List](QC_ANALYSIS_LIST.md)
- [API Reference](api/)
- [Changelog](../CHANGELOG.md)

---

## Support

- **Issues**: [GitHub Issues](https://github.com/rendiffdev/rendiff-probe/issues)
- **Documentation**: [docs/](./README.md)
- **Community**: [Discussions](https://github.com/rendiffdev/rendiff-probe/discussions)

---

*Rendiff Probe - Professional Video Analysis, Powered by FFprobe*
