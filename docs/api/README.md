# FFprobe API Reference

**Professional Video Analysis API with 19 QC Categories**

## Base URL

```
http://localhost:8080
```

## Authentication

Authentication is optional in development mode. In production, the API supports:

### API Key Authentication
```bash
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/v1/probe/file
```

### JWT Token Authentication
```bash
curl -H "Authorization: Bearer your-jwt-token" \
     http://localhost:8080/api/v1/probe/file
```

## API Endpoints

### Health Check

```
GET /health
```

Returns service health status and available QC analysis tools.

**Response:**
```json
{
  "status": "healthy",
  "service": "rendiff-probe-core",
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

### Analyze Video File

```
POST /api/v1/probe/file
Content-Type: multipart/form-data
```

Upload and analyze a video or audio file with comprehensive QC analysis.

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
      "nb_programs": 0,
      "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
      "format_long_name": "QuickTime / MOV",
      "start_time": "0.000000",
      "duration": "60.500000",
      "size": "1048576",
      "bit_rate": "5000000",
      "probe_score": 100
    },
    "streams": [
      {
        "index": 0,
        "codec_name": "h264",
        "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
        "profile": "High",
        "codec_type": "video",
        "width": 1920,
        "height": 1080,
        "coded_width": 1920,
        "coded_height": 1088,
        "display_aspect_ratio": "16:9",
        "pix_fmt": "yuv420p",
        "level": 40,
        "r_frame_rate": "30/1",
        "avg_frame_rate": "30/1",
        "time_base": "1/30000",
        "nb_frames": "1815"
      },
      {
        "index": 1,
        "codec_name": "aac",
        "codec_long_name": "AAC (Advanced Audio Coding)",
        "codec_type": "audio",
        "sample_rate": "48000",
        "channels": 2,
        "channel_layout": "stereo",
        "bits_per_sample": 0,
        "nb_frames": "2833"
      }
    ],
    "enhanced_analysis": {
      "timecode_analysis": {
        "has_timecode": true,
        "is_drop_frame": false,
        "start_timecode": "01:00:00:00"
      },
      "hdr_analysis": {
        "is_hdr_content": false,
        "color_space": "bt709"
      },
      "codec_analysis": {
        "video_codec": "h264",
        "video_profile": "High",
        "audio_codec": "aac"
      },
      "resolution_analysis": {
        "width": 1920,
        "height": 1080,
        "aspect_ratio": "16:9",
        "is_standard_resolution": true
      },
      "data_integrity": {
        "has_errors": false,
        "crc_valid": true,
        "integrity_score": 100
      }
    }
  }
}
```

### Analyze URL

```
POST /api/v1/probe/url
Content-Type: application/json
```

Analyze a video file directly from a URL without uploading.

**Request Body:**
```json
{
  "url": "https://example.com/video.mp4",
  "include_llm": false,
  "timeout": 60
}
```

**Request:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/video.mp4", "include_llm": true}' \
  http://localhost:8080/api/v1/probe/url
```

**Response:**
```json
{
  "status": "success",
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "url": "https://example.com/video.mp4",
  "filename": "video.mp4",
  "analysis": { ... },
  "qc_categories_analyzed": 19,
  "llm_report": "Professional analysis report...",
  "llm_enabled": true,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### HLS Stream Analysis

```
POST /api/v1/probe/hls
Content-Type: application/json
```

Analyze HLS streams for quality, compliance, and performance metrics.

**Request Body:**
```json
{
  "manifest_url": "https://example.com/stream.m3u8",
  "analyze_segments": true,
  "analyze_quality": true,
  "validate_compliance": true,
  "performance_analysis": true,
  "max_segments": 10,
  "include_llm": false
}
```

**Request:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "manifest_url": "https://example.com/stream.m3u8",
    "analyze_segments": true,
    "max_segments": 5
  }' \
  http://localhost:8080/api/v1/probe/hls
```

**Response:**
```json
{
  "status": "success",
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "manifest_url": "https://example.com/stream.m3u8",
  "analysis": {
    "playlist_type": "master",
    "variants": [...],
    "segments": [...],
    "compliance": {...}
  },
  "processing_time": "2.5s",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Batch Processing

#### Start Batch Job
```
POST /api/v1/batch/analyze
Content-Type: application/json
```

Process multiple files or URLs in parallel.

**Request Body:**
```json
{
  "files": ["/path/to/video1.mp4", "/path/to/video2.mp4"],
  "urls": ["https://example.com/video3.mp4"],
  "include_llm": false
}
```

**Response:**
```json
{
  "status": "accepted",
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "total": 3,
  "message": "Batch job started",
  "status_url": "/api/v1/batch/status/550e8400-e29b-41d4-a716-446655440000",
  "ws_url": "/api/v1/ws/progress/550e8400-e29b-41d4-a716-446655440000"
}
```

#### Get Batch Status
```
GET /api/v1/batch/status/:id
```

Get the status and results of a batch job.

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "total": 3,
  "completed": 2,
  "failed": 1,
  "results": [
    {"type": "file", "path": "/path/to/video1.mp4", "status": "success", "analysis": {...}},
    {"type": "url", "url": "https://example.com/video3.mp4", "status": "success", "analysis": {...}},
    {"type": "file", "path": "/path/to/video2.mp4", "status": "failed", "error": "File not found"}
  ],
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

### WebSocket Progress

```
GET /api/v1/ws/progress/:id
```

Connect via WebSocket to receive real-time progress updates for batch jobs.

**Message Format:**
```json
{
  "type": "progress",
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "progress": 66.7,
  "message": "Processed file: video2.mp4",
  "status": "processing",
  "timestamp": "2024-01-15T10:32:00Z"
}
```

**JavaScript Example:**
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/progress/job-id');
ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log(`Progress: ${update.progress}% - ${update.message}`);
};
```

### GraphQL API

```
POST /api/v1/graphql
GET  /api/v1/graphql  # GraphiQL interactive interface
```

Flexible query interface for advanced integrations.

**Example Query:**
```graphql
query {
  health {
    status
    version
  }
}
```

**Example Mutation:**
```graphql
mutation {
  analyzeURL(url: "https://example.com/video.mp4", include_llm: true) {
    id
    filename
    status
    streams {
      index
      codec_name
      codec_type
      width
      height
    }
    format {
      format_name
      duration
      bit_rate
    }
    llm_report
    llm_enabled
    timestamp
  }
}
```

**cURL Example:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ health { status version } }"}' \
  http://localhost:8080/api/v1/graphql
```

### LLM-Powered Insights

Add `include_llm=true` to any analysis endpoint to receive AI-generated professional reports.

**With File Upload:**
```bash
curl -X POST \
  -F "file=@video.mp4" \
  -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file
```

**With URL:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/video.mp4", "include_llm": true}' \
  http://localhost:8080/api/v1/probe/url
```

**LLM Report Contains:**
- Basic media overview
- Video stream details
- Audio stream analysis
- Technical issues detection
- FFmpeg optimization recommendations
- Non-technical summary
- Delivery readiness assessment

### FFmpeg Version Management (Admin)

#### Get Current Version
```
GET /admin/ffmpeg/version
```

Returns the current FFmpeg/FFprobe version information.

**Response:**
```json
{
  "version": "6.1.1",
  "build_date": "2024-01-15",
  "configuration": "--enable-gpl --enable-libx264..."
}
```

#### Check for Updates
```
POST /admin/ffmpeg/check
```

Checks if FFmpeg updates are available.

#### Update FFmpeg
```
POST /admin/ffmpeg/update
```

Triggers an FFmpeg update (requires proper permissions).

## Error Responses

The API uses standard HTTP status codes:

| Status | Description |
|--------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found |
| 413 | Payload Too Large - File exceeds limit |
| 429 | Too Many Requests - Rate limited |
| 500 | Internal Server Error |

**Error Response Format:**
```json
{
  "error": "Error description",
  "details": "Additional details about the error"
}
```

## Rate Limits

Default rate limits (configurable):

| Window | Limit |
|--------|-------|
| Per Minute | 60 requests |
| Per Hour | 1000 requests |
| Per Day | 10000 requests |

## Supported File Formats

The API supports all formats that FFmpeg/FFprobe supports, including:

### Video Containers
- MP4, MOV, MKV, AVI, WebM
- MXF (Material Exchange Format)
- MPEG-TS (Transport Stream)
- FLV, WMV, ASF

### Video Codecs
- H.264/AVC, H.265/HEVC
- VP8, VP9, AV1
- ProRes, DNxHD, DNxHR
- MPEG-2, MPEG-4

### Audio Formats
- AAC, MP3, WAV, FLAC
- AC3, EAC3, DTS
- PCM (various bit depths)
- Opus, Vorbis

## QC Analysis Categories

The API performs 19 quality control analysis categories automatically:

1. AFD Analysis
2. Dead Pixel Detection
3. PSE Flash Analysis
4. HDR Analysis
5. Audio Wrapping Analysis
6. Endianness Detection
7. Codec Analysis
8. Container Validation
9. Resolution Analysis
10. Frame Rate Analysis
11. Bitdepth Analysis
12. Timecode Analysis
13. MXF Analysis
14. IMF Compliance
15. Transport Stream Analysis
16. Content Analysis
17. Enhanced Analysis
18. Stream Disposition Analysis
19. Data Integrity Analysis

See [QC Analysis List](../QC_ANALYSIS_LIST.md) for detailed information on each category.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | API server port |
| `LOG_LEVEL` | `info` | Logging level |
| `FFPROBE_PATH` | `ffprobe` | Path to FFprobe binary |
| `MAX_FILE_SIZE` | `5GB` | Maximum upload file size |
| `ANALYSIS_TIMEOUT` | `5m` | Analysis timeout duration |

## Examples

### Basic File Analysis
```bash
curl -X POST \
  -F "file=@sample.mp4" \
  http://localhost:8080/api/v1/probe/file
```

### With Authentication
```bash
curl -X POST \
  -H "X-API-Key: your-api-key" \
  -F "file=@sample.mp4" \
  http://localhost:8080/api/v1/probe/file
```

### Health Check
```bash
curl http://localhost:8080/health | jq
```

## API Summary

### All Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Service health and feature status |
| `/api/v1/probe/file` | POST | Analyze uploaded file |
| `/api/v1/probe/url` | POST | Analyze file from URL |
| `/api/v1/probe/hls` | POST | Analyze HLS stream |
| `/api/v1/batch/analyze` | POST | Start batch processing |
| `/api/v1/batch/status/:id` | GET | Get batch job status |
| `/api/v1/ws/progress/:id` | WS | Real-time progress updates |
| `/api/v1/graphql` | POST/GET | GraphQL API / GraphiQL |
| `/admin/ffmpeg/version` | GET | FFmpeg version info |
| `/admin/ffmpeg/check` | POST | Check for updates |
| `/admin/ffmpeg/update` | POST | Update FFmpeg |

### Implemented Features (v2.0.0)

- [x] URL-based file analysis (`POST /api/v1/probe/url`)
- [x] HLS stream analysis (`POST /api/v1/probe/hls`)
- [x] Batch processing (`POST /api/v1/batch/analyze`)
- [x] GraphQL endpoint (`POST /api/v1/graphql`)
- [x] WebSocket progress streaming
- [x] LLM-powered insights

### Planned Features

- [ ] Webhook callbacks for async processing
- [ ] DASH stream analysis
- [ ] File comparison endpoint
- [ ] Custom QC rule definitions

---

**For complete QC analysis documentation, see [QC Analysis List](../QC_ANALYSIS_LIST.md)**
