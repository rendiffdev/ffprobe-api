# FFprobe API Documentation

Welcome to the FFprobe API documentation. This comprehensive REST API provides media file analysis capabilities using FFmpeg/FFprobe.

## Overview

The FFprobe API is a production-ready service that offers:

- **Media Analysis**: Detailed analysis of video, audio, and image files
- **Quality Assessment**: Video quality metrics (VMAF, PSNR, SSIM)
- **HLS Support**: Analysis and validation of HLS streaming playlists
- **Report Generation**: Multi-format reports (PDF, HTML, Excel, etc.)
- **Cloud Storage**: Integration with AWS S3, Google Cloud Storage, Azure Blob
- **AI Insights**: AI-powered analysis explanations and recommendations

## Quick Start

### Authentication

The API supports multiple authentication methods:

```bash
# API Key (recommended)
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/health

# JWT Bearer Token
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8080/api/v1/health
```

### Basic Usage

1. **Analyze a media file**:
```bash
curl -X POST "http://localhost:8080/api/v1/probe/file" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "options": {
      "include_streams": true,
      "include_format": true
    }
  }'
```

2. **Check analysis status**:
```bash
curl "http://localhost:8080/api/v1/probe/status/analysis-id" \
  -H "X-API-Key: your-api-key"
```

3. **Generate a report**:
```bash
curl -X POST "http://localhost:8080/api/v1/probe/report" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "your-analysis-id",
    "format": "pdf"
  }'
```

## API Endpoints

### Core Analysis
- `POST /api/v1/probe/file` - Analyze media file
- `POST /api/v1/probe/url` - Analyze media from URL
- `GET /api/v1/probe/status/{id}` - Get analysis status
- `GET /api/v1/probe/analyses` - List all analyses

### Quality Analysis
- `POST /api/v1/probe/compare` - Compare video quality
- `GET /api/v1/quality/analysis/{id}` - Get quality metrics
- `GET /api/v1/quality/statistics` - Quality statistics

### HLS Analysis
- `POST /api/v1/probe/hls` - Analyze HLS playlist
- `POST /api/v1/probe/hls/validate` - Validate HLS playlist
- `GET /api/v1/probe/hls/{id}` - Get HLS analysis results

### Reports
- `POST /api/v1/probe/report` - Generate report
- `GET /api/v1/probe/report/{id}` - Get report status
- `GET /api/v1/probe/download/{id}` - Download report

### Storage
- `POST /api/v1/storage/upload` - Upload file
- `GET /api/v1/storage/download/{key}` - Download file
- `DELETE /api/v1/storage/{key}` - Delete file
- `POST /api/v1/storage/signed-url` - Get signed URL

### AI Features
- `POST /api/v1/ask` - Ask AI about analysis
- `POST /api/v1/genai/analysis` - Generate AI insights

## Response Formats

All API responses follow a consistent format:

```json
{
  "status": "success|error",
  "message": "Human readable message",
  "data": {
    // Response data
  },
  "request_id": "unique-request-id",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Analysis Result Structure

```json
{
  "analysis_id": "uuid",
  "status": "completed",
  "result": {
    "format": {
      "filename": "video.mp4",
      "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
      "duration": "120.5",
      "size": "1048576",
      "bit_rate": "1000000"
    },
    "streams": [
      {
        "index": 0,
        "codec_name": "h264",
        "codec_type": "video",
        "width": 1920,
        "height": 1080,
        "r_frame_rate": "30/1",
        "duration": "120.5"
      }
    ]
  }
}
```

## Error Handling

The API uses standard HTTP status codes and provides detailed error messages:

```json
{
  "error": "invalid_request",
  "message": "The file path is required",
  "code": 400,
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req-123"
}
```

Common error codes:
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (invalid API key)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

## Rate Limiting

Rate limits are enforced per API key:
- **60 requests per minute**
- **1000 requests per hour**
- **10000 requests per day**

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1640995200
```

## File Upload Limits

- **Maximum file size**: 50GB
- **Supported formats**: All FFmpeg supported formats
- **Timeout**: 10 minutes for analysis

## Cloud Storage Integration

Configure cloud storage providers:

### AWS S3
```bash
export STORAGE_PROVIDER=s3
export STORAGE_BUCKET=my-bucket
export STORAGE_REGION=us-east-1
export STORAGE_ACCESS_KEY=your-access-key
export STORAGE_SECRET_KEY=your-secret-key
```

### Google Cloud Storage
```bash
export STORAGE_PROVIDER=gcs
export STORAGE_BUCKET=my-bucket
export GCP_SERVICE_ACCOUNT_JSON='{"type":"service_account",...}'
```

### Azure Blob Storage
```bash
export STORAGE_PROVIDER=azure
export STORAGE_BUCKET=my-container
export AZURE_STORAGE_ACCOUNT=account-name
export AZURE_STORAGE_KEY=account-key
```

## Quality Metrics

### VMAF (Video Multi-Method Assessment Fusion)
- Range: 0-100 (higher is better)
- Industry standard for video quality assessment
- Models available: v0.6.1, v0.6.1neg, 4K, B

### PSNR (Peak Signal-to-Noise Ratio)
- Range: 0-100 dB (higher is better)
- Traditional quality metric
- Computed per color channel (Y, U, V)

### SSIM (Structural Similarity Index)
- Range: 0-1 (higher is better)
- Perceptual quality metric
- Available: SSIM, MS-SSIM

## SDK and Libraries

Official SDKs available:
- **Go**: `go get github.com/ffprobe-api/go-sdk`
- **Python**: `pip install ffprobe-api`
- **Node.js**: `npm install ffprobe-api`
- **Java**: Maven/Gradle dependency

## Examples

### Video Quality Comparison
```bash
curl -X POST "http://localhost:8080/api/v1/probe/compare" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "reference_file": "/path/to/original.mp4",
    "distorted_file": "/path/to/compressed.mp4",
    "metrics": ["vmaf", "psnr", "ssim"]
  }'
```

### HLS Playlist Analysis
```bash
curl -X POST "http://localhost:8080/api/v1/probe/hls" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "manifest_url": "https://cdn.example.com/playlist.m3u8",
    "analyze_segments": true,
    "segment_limit": 5
  }'
```

### Generate PDF Report
```bash
curl -X POST "http://localhost:8080/api/v1/probe/report" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "your-analysis-id",
    "format": "pdf",
    "template": "detailed"
  }'
```

### AI-Powered Analysis
```bash
curl -X POST "http://localhost:8080/api/v1/ask" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "your-analysis-id",
    "question": "What are the potential quality issues with this video?"
  }'
```

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:
- YAML: `/docs/openapi.yaml`
- JSON: `/docs/openapi.json`
- Interactive UI: `/docs/swagger-ui/`

## Support

- **Documentation**: [GitHub Repository](https://github.com/rendiffdev/ffprobe-api)
- **Website**: [https://rendiff.dev](https://rendiff.dev)
- **GitHub Issues**: [https://github.com/rendiffdev/ffprobe-api/issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **Email Support**: dev@rendiff.dev

## Changelog

### v2.0.0
- Added cloud storage integration
- AI-powered analysis insights
- Enhanced HLS support
- Multi-format report generation
- Improved rate limiting
- WebSocket streaming support

### v1.0.0
- Initial release
- Basic media analysis
- Quality metrics (VMAF, PSNR, SSIM)
- RESTful API
- JWT authentication