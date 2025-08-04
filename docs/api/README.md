# API Reference

> **Complete REST API documentation for FFprobe API with examples and response formats**

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

FFprobe API supports two authentication methods:

### API Key Authentication (Recommended for services)

```bash
curl -H "X-API-Key: your-api-key" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/v1/probe/file
```

### JWT Token Authentication (Recommended for users)

```bash
curl -H "Authorization: Bearer your-jwt-token" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/v1/probe/file
```

## Core Endpoints

### Video Analysis Endpoints

| Endpoint | Method | Description | Quality Checks |
|----------|--------|-------------|----------------|
| `/probe/file` | POST | Analyze local video file | 29 standard + 20 enhanced |
| `/probe/url` | POST | Analyze video from URL | 29 standard + 20 enhanced |
| `/probe/quick` | POST | Fast basic analysis | 29 standard only |
| `/batch/analyze` | POST | Batch video processing | Configurable |
| `/probe/status/{id}` | GET | Get analysis status | N/A |

### Management Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/probe/analyses` | GET | List user analyses |
| `/probe/analyses/{id}` | DELETE | Delete analysis |
| `/quality/compare` | POST | Quality comparison |
| `/reports/analysis` | POST | Generate reports |

### System Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | System health check |
| `/probe/health` | GET | FFprobe service health |

## Detailed Endpoint Documentation

### POST /api/v1/probe/file

Analyze a local video file with comprehensive quality control checks.

**Request Body:**
```json
{
  "file_path": "/path/to/video.mp4",
  "content_analysis": true,
  "async": false,
  "generate_reports": true,
  "report_formats": ["json", "pdf"],
  "options": {
    "show_format": true,
    "show_streams": true,
    "show_chapters": true,
    "count_frames": true
  }
}
```

**Parameters:**
- `file_path` (required): Path to the video file
- `content_analysis` (optional): Enable enhanced analysis with 20 additional checks
- `async` (optional): Process asynchronously (default: false)
- `generate_reports` (optional): Generate analysis reports
- `report_formats` (optional): Array of formats ["json", "xml", "pdf"]
- `options` (optional): FFprobe options for detailed control

**Response (Success - 200):**
```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "analysis": {
    "format": {
      "filename": "video.mp4",
      "duration": "120.5",
      "bit_rate": "5000000",
      "format_name": "mov,mp4,m4a,3gp,3g2,mj2"
    },
    "streams": [
      {
        "index": 0,
        "codec_name": "h264",
        "codec_type": "video",
        "width": 1920,
        "height": 1080,
        "r_frame_rate": "30/1",
        "avg_frame_rate": "30/1",
        "pix_fmt": "yuv420p",
        "bit_rate": "4500000"
      },
      {
        "index": 1,
        "codec_name": "aac",
        "codec_type": "audio",
        "sample_rate": "48000",
        "channels": 2,
        "bit_rate": "128000"
      }
    ],
    "enhanced_analysis": {
      "stream_counts": {
        "total_streams": 2,
        "video_streams": 1,
        "audio_streams": 1,
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
        "freeze_frames": {
          "detected_frames": 2,
          "percentage": 0.05
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
    "download_urls": [
      "http://localhost:8080/api/v1/reports/download/550e8400-e29b-41d4-a716-446655440000.json",
      "http://localhost:8080/api/v1/reports/download/550e8400-e29b-41d4-a716-446655440000.pdf"
    ]
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Response (Async - 202):**
```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "message": "Analysis started, check status endpoint for progress"
}
```

**Error Response (400/500):**
```json
{
  "error": "Invalid file path",
  "details": "File does not exist or is not accessible"
}
```

### POST /api/v1/probe/url

Analyze a video from a remote URL.

**Request Body:**
```json
{
  "url": "https://example.com/video.mp4",
  "content_analysis": true,
  "timeout": 300,
  "async": false,
  "generate_reports": false
}
```

**Parameters:**
- `url` (required): Video URL to analyze
- `content_analysis` (optional): Enable enhanced analysis
- `timeout` (optional): Request timeout in seconds
- `async` (optional): Process asynchronously
- `generate_reports` (optional): Generate analysis reports

**Response:** Same format as `/probe/file`

### POST /api/v1/probe/quick

Perform fast basic analysis with minimal information.

**Request Body:**
```json
{
  "file_path": "/path/to/video.mp4",
  "source_type": "local"
}
```

**Response:** Basic analysis with standard 29 checks only (no enhanced analysis).

### GET /api/v1/probe/status/{id}

Get the current status and result of an analysis.

**Parameters:**
- `id` (path): Analysis UUID

**Response:**
```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "analysis": {
    // Full analysis object
  }
}
```

**Status Values:**
- `pending`: Analysis queued but not started
- `processing`: Analysis in progress
- `completed`: Analysis finished successfully
- `failed`: Analysis failed with error

### POST /api/v1/batch/analyze

Process multiple videos in batch.

**Request Body:**
```json
{
  "files": [
    {
      "file_path": "/path/to/video1.mp4",
      "content_analysis": true
    },
    {
      "file_path": "/path/to/video2.mp4",
      "content_analysis": false
    }
  ],
  "callback_url": "https://your-app.com/webhook",
  "batch_options": {
    "parallel_jobs": 3,
    "priority": "normal"
  }
}
```

**Response:**
```json
{
  "batch_id": "batch-550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "total_files": 2,
  "analyses": [
    {
      "analysis_id": "550e8400-e29b-41d4-a716-446655440001",
      "file_path": "/path/to/video1.mp4",
      "status": "pending"
    },
    {
      "analysis_id": "550e8400-e29b-41d4-a716-446655440002", 
      "file_path": "/path/to/video2.mp4",
      "status": "pending"
    }
  ]
}
```

### GET /api/v1/probe/analyses

List analyses for the current user.

**Query Parameters:**
- `limit` (optional): Number of results (default: 20, max: 100)
- `offset` (optional): Pagination offset (default: 0)
- `status` (optional): Filter by status
- `source_type` (optional): Filter by source type

**Response:**
```json
{
  "analyses": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "file_name": "video.mp4",
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z",
      "processed_at": "2024-01-15T10:35:00Z"
    }
  ],
  "limit": 20,
  "offset": 0,
  "count": 1
}
```

### DELETE /api/v1/probe/analyses/{id}

Delete an analysis and its results.

**Parameters:**
- `id` (path): Analysis UUID

**Response:** 204 No Content

### GET /health

System health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "ffprobe_version": "6.1.1",
  "database": "connected",
  "redis": "connected",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Quality Control Features

### Standard Analysis (29 Checks)

Automatically included with every request:

**Container/Format (5 checks):**
- Container format validation
- Duration and file size
- Overall bitrate
- Format confidence score

**Video Streams (14 checks):**
- Codec and profile detection
- Resolution and aspect ratio
- Frame rate and pixel format
- Color space and transfer characteristics
- Progressive/interlaced detection

**Audio Streams (6 checks):**
- Codec and sample rate
- Channel configuration
- Audio bitrate and format

**Additional (4 checks):**
- Chapter markers
- Metadata tags
- Program information
- Stream relationships

### Enhanced Analysis (20 Additional Checks)

Enable with `"content_analysis": true`:

**Stream Analysis (2 checks):**
- Detailed stream counting
- Closed caption detection

**Video Enhancement (4 checks):**
- Chroma subsampling pattern
- Matrix coefficients mapping
- Bitrate mode detection (CBR/VBR)
- GOP structure analysis

**Frame Statistics (4 checks):**
- Frame type distribution (I/P/B)
- Frame size statistics
- Compression efficiency
- Keyframe analysis

**Content Analysis (9 checks):**
- Blackness detection
- Freeze frame detection
- Audio clipping detection
- Blockiness measurement
- Blurriness analysis
- Interlacing artifacts
- Noise level measurement
- Broadcast loudness compliance
- True peak detection

**Audio Enhancement (1 check):**
- Audio bitrate mode detection

## Error Handling

### HTTP Status Codes

- `200 OK`: Request successful
- `202 Accepted`: Async request accepted
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service temporarily unavailable

### Error Response Format

```json
{
  "error": "Brief error description",
  "details": "Detailed error message",
  "code": "ERROR_CODE",
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-550e8400-e29b-41d4-a716-446655440000"
}
```

### Common Error Codes

- `INVALID_FILE_PATH`: File path is invalid or inaccessible
- `UNSUPPORTED_FORMAT`: Video format not supported by FFprobe
- `FILE_TOO_LARGE`: File exceeds maximum size limit
- `PROCESSING_TIMEOUT`: Analysis timed out
- `INVALID_API_KEY`: API key is invalid or expired
- `RATE_LIMIT_EXCEEDED`: Too many requests in time window

## Rate Limiting

### Default Limits

- **API Key**: 1000 requests per hour
- **JWT Token**: 500 requests per hour
- **IP Address**: 100 requests per hour (unauthenticated)

### Rate Limit Headers

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642248000
```

## Webhooks

### Callback Configuration

For async processing, provide a callback URL:

```json
{
  "file_path": "/path/to/video.mp4",
  "async": true,
  "callback_url": "https://your-app.com/webhook"
}
```

### Webhook Payload

```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "file_path": "/path/to/video.mp4",
  "analysis": {
    // Complete analysis results
  },
  "timestamp": "2024-01-15T10:35:00Z"
}
```

## SDKs and Client Libraries

### Official SDKs

- **JavaScript/Node.js**: `npm install ffprobe-api-client`
- **Python**: `pip install ffprobe-api-client`
- **Go**: `go get github.com/rendiffdev/ffprobe-api-go`

### Community SDKs

- **PHP**: Available via Packagist
- **Ruby**: Available via RubyGems
- **Java**: Available via Maven Central

---

## Additional Resources

- [Quality Checks Reference](../QUALITY_CHECKS.md)
- [Authentication Guide](authentication.md)
- [Examples and Tutorials](examples.md)
- [Troubleshooting](../troubleshooting/README.md)

*For API support, see [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)*