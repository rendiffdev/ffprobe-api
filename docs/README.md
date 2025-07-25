# 📚 FFprobe API Documentation

Welcome to the comprehensive FFprobe API documentation. This production-ready REST API provides advanced media file analysis capabilities using FFmpeg/FFprobe with enterprise-grade features.

## 🎯 Overview

The FFprobe API is a comprehensive, enterprise-grade service that offers:

- **🎬 Advanced Media Analysis**: Detailed analysis of video, audio, and image files
- **📊 Quality Assessment**: Industry-standard video quality metrics (VMAF, PSNR, SSIM)
- **📺 HLS Support**: Complete analysis and validation of HLS streaming playlists  
- **📄 Report Generation**: Multi-format reports (PDF, HTML, Excel, Markdown, etc.)
- **☁️ Cloud Storage**: Full integration with AWS S3, Google Cloud Storage, Azure Blob
- **🤖 AI Insights**: AI-powered analysis explanations and recommendations
- **⚡ Batch Processing**: High-performance batch analysis capabilities
- **🔒 Enterprise Security**: Authentication, rate limiting, and audit logging

## 🚀 Quick Start

### Authentication

The API supports multiple authentication methods for maximum flexibility:

#### API Key (Recommended)
```bash
curl -H "X-API-Key: your-api-key" http://localhost:8080/api/v1/health
```

#### JWT Bearer Token
```bash
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8080/api/v1/health
```

#### Basic Authentication (Development)
```bash
curl -u "username:password" http://localhost:8080/api/v1/health
```

### Basic Workflow Examples

#### 1. Analyze a Media File
```bash
curl -X POST "http://localhost:8080/api/v1/probe/file" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "options": {
      "include_streams": true,
      "include_format": true,
      "include_chapters": true,
      "include_programs": true
    }
  }'
```

#### 2. Check Analysis Status  
```bash
curl "http://localhost:8080/api/v1/probe/status/analysis-id" \
  -H "X-API-Key: your-api-key"
```

#### 3. Generate and Download Report
```bash
# Generate report
curl -X POST "http://localhost:8080/api/v1/probe/report" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "your-analysis-id",
    "format": "pdf",
    "template": "professional",
    "include_charts": true
  }'

# Download report
curl "http://localhost:8080/api/v1/probe/download/report-id" \
  -H "X-API-Key: your-api-key" \
  -o report.pdf
```

## 📡 API Endpoints Reference

### 🎬 Core Media Analysis

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/probe/file` | Analyze local media file |
| `POST` | `/api/v1/probe/url` | Analyze media from URL |
| `POST` | `/api/v1/probe/upload` | Upload and analyze file |
| `GET` | `/api/v1/probe/status/{id}` | Get analysis status |
| `GET` | `/api/v1/probe/result/{id}` | Get analysis results |
| `GET` | `/api/v1/probe/analyses` | List all analyses |
| `DELETE` | `/api/v1/probe/{id}` | Delete analysis |

### 📊 Quality Analysis

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/quality/compare` | Compare video quality |
| `POST` | `/api/v1/quality/analyze` | Quality analysis request |
| `GET` | `/api/v1/quality/result/{id}` | Get quality metrics |
| `GET` | `/api/v1/quality/statistics` | Quality statistics |
| `POST` | `/api/v1/quality/validate` | Validate quality metrics accuracy |

### 📺 HLS Streaming Analysis

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/hls/analyze` | Analyze HLS playlist |
| `POST` | `/api/v1/hls/validate` | Validate HLS compliance |
| `GET` | `/api/v1/hls/result/{id}` | Get HLS analysis results |
| `POST` | `/api/v1/hls/segments` | Analyze specific segments |

### 📄 Report Generation

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/reports/generate` | Generate report |
| `GET` | `/api/v1/reports/status/{id}` | Get report status |
| `GET` | `/api/v1/reports/download/{id}` | Download report |
| `GET` | `/api/v1/reports/templates` | List available templates |

### ☁️ Storage Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/storage/upload` | Upload file to storage |
| `GET` | `/api/v1/storage/download/{key}` | Download file |
| `DELETE` | `/api/v1/storage/{key}` | Delete file |
| `POST` | `/api/v1/storage/signed-url` | Get signed URL |
| `GET` | `/api/v1/storage/list` | List stored files |

### 🤖 AI Features

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/genai/ask` | Ask AI about analysis |
| `POST` | `/api/v1/genai/analysis` | Generate AI insights |
| `POST` | `/api/v1/genai/quality-insights/{id}` | AI quality recommendations |

### 🔄 Batch Processing

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/batch/analyze` | Start batch analysis |
| `GET` | `/api/v1/batch/status/{id}` | Get batch status |
| `POST` | `/api/v1/batch/{id}/cancel` | Cancel batch operation |
| `GET` | `/api/v1/batch` | List batch operations |

### 🔧 System & Monitoring

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Basic health check |
| `GET` | `/health/detailed` | Detailed health status |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/api/v1/system/info` | System information |

## 📋 Response Formats

### Standard API Response Structure

All API responses follow a consistent format for predictable integration:

```json
{
  "status": "success|error|processing",
  "message": "Human readable message", 
  "data": {
    // Response data specific to endpoint
  },
  "meta": {
    "request_id": "unique-request-id",
    "timestamp": "2024-01-01T12:00:00Z",
    "api_version": "2.0.0",
    "processing_time": "150ms"
  }
}
```

### Analysis Result Structure

Complete media analysis results include comprehensive metadata:

```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "file_info": {
    "filename": "video.mp4",
    "size": 1048576,
    "format": "mp4",
    "duration": 120.5,
    "created_at": "2024-01-01T12:00:00Z"
  },
  "format": {
    "filename": "video.mp4", 
    "nb_streams": 2,
    "nb_programs": 0,
    "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
    "format_long_name": "QuickTime / MOV",
    "start_time": "0.000000",
    "duration": "120.500000",
    "size": "1048576",
    "bit_rate": "69738",
    "probe_score": 100
  },
  "streams": [
    {
      "index": 0,
      "codec_name": "h264",
      "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10", 
      "profile": "High",
      "codec_type": "video",
      "codec_tag_string": "avc1",
      "codec_tag": "0x31637661",
      "width": 1920,
      "height": 1080,
      "coded_width": 1920,
      "coded_height": 1088,
      "closed_captions": 0,
      "film_grain": 0,
      "has_b_frames": 2,
      "sample_aspect_ratio": "1:1",
      "display_aspect_ratio": "16:9",
      "pix_fmt": "yuv420p",
      "level": 40,
      "color_range": "tv",
      "color_space": "bt709",
      "color_transfer": "bt709",
      "color_primaries": "bt709",
      "chroma_location": "left",
      "field_order": "progressive",
      "refs": 1,
      "is_avc": "true",
      "nal_length_size": "4",
      "r_frame_rate": "30/1",
      "avg_frame_rate": "30/1",
      "time_base": "1/15360",
      "start_pts": 0,
      "start_time": "0.000000",
      "duration_ts": 1851392,
      "duration": "120.500000",
      "bit_rate": "65536",
      "bits_per_raw_sample": "8",
      "nb_frames": "3615"
    }
  ],
  "chapters": [],
  "programs": [],
  "quality_metrics": {
    // Optional quality analysis results
  }
}
```

### Quality Analysis Results

```json
{
  "quality_analysis_id": "quality-uuid",
  "reference_file": "/path/to/original.mp4",
  "distorted_file": "/path/to/compressed.mp4", 
  "metrics": {
    "vmaf": {
      "overall_score": 85.42,
      "min_score": 72.31,
      "max_score": 95.67,
      "mean_score": 85.42,
      "frame_scores": [
        {"frame": 0, "timestamp": 0.0, "score": 85.1},
        {"frame": 1, "timestamp": 0.033, "score": 85.3}
      ]
    },
    "psnr": {
      "overall_score": 42.15,
      "y_channel": 42.89,
      "u_channel": 45.21,
      "v_channel": 44.87
    },
    "ssim": {
      "overall_score": 0.9534,
      "y_channel": 0.9541,
      "u_channel": 0.9598,
      "v_channel": 0.9463
    }
  }
}
```

## ⚠️ Error Handling

### HTTP Status Codes

The API uses standard HTTP status codes with detailed error messages:

| Code | Status | Description |
|------|--------|-------------|
| `200` | OK | Request successful |
| `201` | Created | Resource created successfully |
| `202` | Accepted | Request accepted for processing |
| `400` | Bad Request | Invalid request parameters |
| `401` | Unauthorized | Invalid or missing API key |
| `403` | Forbidden | Insufficient permissions |
| `404` | Not Found | Resource doesn't exist |
| `409` | Conflict | Resource already exists |
| `413` | Payload Too Large | File size exceeds limit |
| `422` | Unprocessable Entity | Validation errors |
| `429` | Too Many Requests | Rate limit exceeded |
| `500` | Internal Server Error | Server error |
| `503` | Service Unavailable | Service temporarily unavailable |

### Error Response Format

```json
{
  "status": "error",
  "error": {
    "code": "INVALID_FILE_FORMAT",
    "message": "The uploaded file format is not supported",
    "details": "Only video and audio files are accepted. Received: text/plain",
    "field": "file_path",
    "timestamp": "2024-01-01T12:00:00Z"
  },
  "meta": {
    "request_id": "req-12345",
    "api_version": "2.0.0"
  }
}
```

### Common Error Codes

| Error Code | Description | Resolution |
|------------|-------------|------------|
| `INVALID_API_KEY` | API key is invalid or expired | Check your API key |
| `RATE_LIMIT_EXCEEDED` | Too many requests | Wait before retrying |
| `FILE_NOT_FOUND` | Specified file doesn't exist | Verify file path |
| `UNSUPPORTED_FORMAT` | File format not supported | Use supported format |
| `FILE_TOO_LARGE` | File exceeds size limit | Reduce file size |
| `PROCESSING_FAILED` | Analysis processing failed | Check file integrity |
| `STORAGE_ERROR` | Cloud storage operation failed | Check storage config |

## 🔧 Rate Limiting

Rate limits are enforced per API key to ensure fair usage:

### Default Limits

| Period | Limit |
|--------|-------|
| **Per Minute** | 60 requests |
| **Per Hour** | 1,000 requests |
| **Per Day** | 10,000 requests |

### Enterprise Limits

| Period | Limit |
|--------|-------|
| **Per Minute** | 300 requests |
| **Per Hour** | 10,000 requests |  
| **Per Day** | 100,000 requests |

### Rate Limit Headers

Rate limit information is included in all response headers:

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1640995200
X-RateLimit-Type: per-minute
```

### Handling Rate Limits

When rate limit is exceeded, implement exponential backoff:

```bash
# Example with retry logic
curl -w "%{http_code}" -H "X-API-Key: key" http://api.example.com/endpoint
# If response is 429, wait and retry with exponential backoff
```

## 📁 File Upload & Processing Limits

### File Size Limits

| Plan | File Size Limit | Batch Size Limit |
|------|------------------|------------------|
| **Free** | 1GB | 10 files |
| **Pro** | 50GB | 100 files |
| **Enterprise** | 500GB | 1,000 files |

### Supported Formats

#### Video Formats
- **Container**: MP4, MOV, AVI, MKV, WebM, FLV, M4V, 3GP, ASF, WMV
- **Codecs**: H.264, H.265, VP9, AV1, MPEG-2, MPEG-4, Theora

#### Audio Formats  
- **Container**: MP3, WAV, FLAC, AAC, OGG, WMA, M4A, AIFF
- **Codecs**: MP3, AAC, FLAC, Opus, Vorbis, PCM

#### Streaming Formats
- **HLS**: M3U8 playlists and TS segments
- **DASH**: MPD manifests and MP4 segments
- **Smooth Streaming**: ISM manifests

### Processing Timeouts

| Operation | Timeout |
|-----------|---------|
| **File Analysis** | 10 minutes |
| **Quality Comparison** | 30 minutes |
| **HLS Analysis** | 5 minutes |
| **Report Generation** | 2 minutes |
| **Batch Processing** | 2 hours |

## ☁️ Cloud Storage Integration

### Supported Providers

#### AWS S3
```bash
# Environment Configuration
export STORAGE_PROVIDER=s3
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...
export S3_BUCKET=my-ffprobe-bucket

# Usage
curl -X POST "http://localhost:8080/api/v1/storage/upload" \
  -H "X-API-Key: your-api-key" \
  -F "file=@video.mp4" \
  -F "key=videos/video.mp4"
```

#### Google Cloud Storage  
```bash
# Environment Configuration
export STORAGE_PROVIDER=gcs
export GCP_PROJECT_ID=my-project
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
export GCS_BUCKET=my-ffprobe-bucket

# Usage with signed URL
curl -X POST "http://localhost:8080/api/v1/storage/signed-url" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "videos/video.mp4",
    "operation": "write",
    "expires_in": 3600
  }'
```

#### Azure Blob Storage
```bash
# Environment Configuration  
export STORAGE_PROVIDER=azure
export AZURE_STORAGE_ACCOUNT=mystorageaccount
export AZURE_STORAGE_KEY=...
export AZURE_CONTAINER=ffprobe-files

# Direct analysis from Azure
curl -X POST "http://localhost:8080/api/v1/probe/url" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://mystorageaccount.blob.core.windows.net/container/video.mp4",
    "storage_auth": {
      "provider": "azure",
      "credentials": "..."
    }
  }'
```

## 📊 Quality Metrics Reference

### VMAF (Video Multi-Method Assessment Fusion)

Industry-standard perceptual video quality metric developed by Netflix.

**Range**: 0-100 (higher is better)  
**Models Available**:
- `vmaf_v0.6.1` - Standard model
- `vmaf_v0.6.1neg` - Enhanced for low quality
- `vmaf_4k_v0.6.1` - Optimized for 4K content
- `vmaf_mobile` - Mobile viewing conditions

**Usage**:
```bash
curl -X POST "http://localhost:8080/api/v1/quality/compare" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "reference_file": "original.mp4",
    "distorted_file": "compressed.mp4", 
    "metrics": ["vmaf"],
    "vmaf_model": "vmaf_v0.6.1",
    "vmaf_options": {
      "subsample": 1,
      "pool_method": "mean"
    }
  }'
```

### PSNR (Peak Signal-to-Noise Ratio)

Traditional objective quality metric measuring signal fidelity.

**Range**: 0-100 dB (higher is better)  
**Components**: Y (luma), U, V (chroma)

**Quality Guidelines**:
- `> 40 dB` - Excellent quality
- `30-40 dB` - Good quality  
- `20-30 dB` - Fair quality
- `< 20 dB` - Poor quality

### SSIM (Structural Similarity Index)

Perceptual quality metric based on structural information.

**Range**: 0-1 (higher is better)  
**Variants**: SSIM, MS-SSIM (multi-scale)

**Quality Guidelines**:
- `> 0.95` - Excellent quality
- `0.90-0.95` - Good quality
- `0.80-0.90` - Fair quality  
- `< 0.80` - Poor quality

## 🔄 Batch Processing

### Batch Analysis Request

Process multiple files efficiently with parallel processing:

```bash
curl -X POST "http://localhost:8080/api/v1/batch/analyze" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {
        "id": "file1",
        "file_path": "/path/to/video1.mp4",
        "options": {
          "include_streams": true,
          "include_format": true
        }
      },
      {
        "id": "file2", 
        "file_path": "s3://bucket/video2.mp4",
        "options": {
          "include_quality_metrics": true,
          "quality_reference": "/path/to/reference.mp4"
        }
      }
    ],
    "async": true,
    "priority": "high",
    "callback_url": "https://your-app.com/webhook/batch-complete"
  }'
```

### Batch Status Monitoring

```bash
# Get batch status
curl "http://localhost:8080/api/v1/batch/status/batch-id" \
  -H "X-API-Key: your-api-key"

# Response includes progress information
{
  "batch_id": "batch-uuid",
  "status": "processing", 
  "total": 100,
  "completed": 45,
  "failed": 2,
  "in_progress": 5,
  "progress_percentage": 47.0,
  "estimated_completion": "2024-01-01T13:30:00Z",
  "results": [
    {
      "file_id": "file1",
      "status": "completed",
      "analysis_id": "analysis-uuid-1"
    }
  ]
}
```

## 🤖 AI-Powered Features

### AI Analysis Insights

Get human-readable explanations of analysis results:

```bash
curl -X POST "http://localhost:8080/api/v1/genai/analysis" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "your-analysis-id",
    "focus_areas": ["quality", "technical", "optimization"],
    "detail_level": "comprehensive"
  }'

# Response includes AI-generated insights
{
  "analysis_id": "uuid",
  "insights": {
    "summary": "This video shows excellent technical quality with...",
    "quality_assessment": "The video maintains consistent quality with...", 
    "technical_details": "Encoded with H.264 using optimal settings...",
    "recommendations": [
      "Consider increasing bitrate for higher quality",
      "Audio could benefit from noise reduction"
    ]
  }
}
```

### Interactive Q&A

Ask specific questions about your media:

```bash
curl -X POST "http://localhost:8080/api/v1/genai/ask" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "your-analysis-id",
    "question": "What streaming bitrates would work best for this content?"
  }'

# AI provides contextual answers
{
  "question": "What streaming bitrates would work best for this content?",
  "answer": "Based on the analysis, I recommend a bitrate ladder of 500kbps, 1.5Mbps, 3Mbps, and 6Mbps for optimal streaming across different devices and network conditions...",
  "confidence": 0.92
}
```

## 📄 Report Generation

### Available Report Formats

| Format | Extension | Use Case |
|--------|-----------|----------|
| **PDF** | `.pdf` | Professional reports, archival |
| **HTML** | `.html` | Web viewing, interactive charts |
| **Excel** | `.xlsx` | Data analysis, spreadsheets |
| **Markdown** | `.md` | Documentation, GitHub |
| **CSV** | `.csv` | Data export, processing |
| **JSON** | `.json` | API integration, automation |

### Report Templates

| Template | Description | Best For |
|----------|-------------|----------|
| `standard` | Basic analysis overview | General use |
| `professional` | Comprehensive with branding | Client reports |
| `technical` | Detailed technical specs | Engineering teams |
| `executive` | High-level summary | Management |
| `comparison` | Side-by-side quality comparison | Quality analysis |

### Generate Custom Report

```bash
curl -X POST "http://localhost:8080/api/v1/reports/generate" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "your-analysis-id",
    "format": "pdf",
    "template": "professional",
    "options": {
      "include_charts": true,
      "include_thumbnails": true,
      "include_metadata": true,
      "branding": {
        "company_name": "Your Company",
        "logo_url": "https://your-company.com/logo.png"
      }
    },
    "sections": [
      "executive_summary",
      "technical_details", 
      "quality_metrics",
      "recommendations"
    ]
  }'
```

## 📈 Monitoring & Metrics

### Prometheus Metrics

The API exposes comprehensive metrics for monitoring:

```bash
# Available metrics endpoint
curl http://localhost:8080/metrics

# Key metrics include:
# - ffprobe_requests_total
# - ffprobe_request_duration_seconds
# - ffprobe_active_analyses
# - ffprobe_quality_analysis_duration
# - ffprobe_storage_operations_total
# - ffprobe_error_rate
```

### Health Check Endpoints

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed health with dependencies
curl http://localhost:8080/health/detailed

# Response includes system status
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "2.0.0",
  "uptime": "72h30m15s",
  "dependencies": {
    "database": "healthy",
    "redis": "healthy", 
    "storage": "healthy",
    "ffmpeg": "healthy"
  }
}
```

## 🔐 Security & Authentication

### API Key Management

```bash
# Generate new API key
curl -X POST "http://localhost:8080/api/v1/auth/keys" \
  -H "Authorization: Bearer admin-token" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Application",
    "permissions": ["read", "write", "admin"],
    "expires_at": "2025-12-31T23:59:59Z"
  }'

# List API keys
curl "http://localhost:8080/api/v1/auth/keys" \
  -H "Authorization: Bearer admin-token"

# Revoke API key
curl -X DELETE "http://localhost:8080/api/v1/auth/keys/key-id" \
  -H "Authorization: Bearer admin-token"
```

### JWT Token Authentication

```bash
# Login to get JWT token
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your-username",
    "password": "your-password"
  }'

# Use JWT token
curl -H "Authorization: Bearer jwt-token" \
  http://localhost:8080/api/v1/probe/analyses
```

## 📚 SDK & Client Libraries

### Official SDKs

#### Go SDK
```go
import "github.com/ffprobe-api/go-sdk"

client := ffprobe.NewClient("your-api-key")
analysis, err := client.AnalyzeFile("/path/to/video.mp4")
```

#### Python SDK  
```python
from ffprobe_api import FFprobeClient

client = FFprobeClient("your-api-key")
analysis = client.analyze_file("/path/to/video.mp4")
```

#### Node.js SDK
```javascript
const FFprobeAPI = require('ffprobe-api');

const client = new FFprobeAPI('your-api-key');
const analysis = await client.analyzeFile('/path/to/video.mp4');
```

#### Java SDK
```java
import dev.rendiff.ffprobe.FFprobeClient;

FFprobeClient client = new FFprobeClient("your-api-key");
Analysis analysis = client.analyzeFile("/path/to/video.mp4");
```

### Community SDKs

- **PHP**: `composer require ffprobe-api/php-sdk`
- **Ruby**: `gem install ffprobe-api`
- **Rust**: `cargo add ffprobe-api`
- **C#**: `dotnet add package FFprobeAPI.Client`

## 🔗 Integration Examples

### Webhook Integration

Set up webhooks to receive real-time updates:

```bash
# Configure webhook endpoint
curl -X POST "http://localhost:8080/api/v1/webhooks" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhook/ffprobe",
    "events": ["analysis.completed", "analysis.failed", "batch.completed"],
    "secret": "webhook-secret-for-verification"
  }'

# Webhook payload example
{
  "event": "analysis.completed",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "analysis_id": "uuid",
    "status": "completed",
    "file_path": "/path/to/video.mp4",
    "processing_time": "45.2s"
  }
}
```

### CI/CD Pipeline Integration

```yaml
# GitHub Actions example
name: Video Quality Check
on: [push]
jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Analyze video quality
        run: |
          curl -X POST "${{ secrets.FFPROBE_API_URL }}/api/v1/quality/compare" \
            -H "X-API-Key: ${{ secrets.FFPROBE_API_KEY }}" \
            -H "Content-Type: application/json" \
            -d '{
              "reference_file": "tests/reference.mp4",
              "distorted_file": "output/compressed.mp4",
              "metrics": ["vmaf", "psnr"]
            }'
```

## 📋 OpenAPI Specification

### Interactive Documentation

Access interactive API documentation:

- **Swagger UI**: `http://localhost:8080/docs/swagger-ui/`
- **ReDoc**: `http://localhost:8080/docs/redoc/`
- **OpenAPI JSON**: `http://localhost:8080/docs/openapi.json`
- **OpenAPI YAML**: `http://localhost:8080/docs/openapi.yaml`

### Download Specification

```bash
# Download OpenAPI specification
curl http://localhost:8080/docs/openapi.yaml > ffprobe-api.yaml
curl http://localhost:8080/docs/openapi.json > ffprobe-api.json

# Generate client code using OpenAPI Generator
openapi-generator-cli generate \
  -i ffprobe-api.yaml \
  -g python \
  -o ./python-client
```

## 🆘 Support & Resources

### Documentation Resources

- **📚 [API Documentation](https://github.com/your-org/ffprobe-api/docs)** - This comprehensive guide
- **🚀 [Quick Start Guide](https://github.com/your-org/ffprobe-api#quick-start)** - Get started in minutes  
- **📋 [Technical Guidelines](../CONTRIBUTOR-GUIDELINES.md)** - For developers
- **🐳 [Docker Guide](./docker.md)** - Container deployment
- **🔒 [Security Guide](./security.md)** - Security best practices
- **📊 [Monitoring Guide](./monitoring.md)** - Observability setup

### Example Code & Tutorials

- **🎯 [API Examples](../examples/)** - Ready-to-use code samples
- **⚙️ [Configuration Examples](../examples/config/)** - Common configurations  
- **🔧 [Integration Examples](../examples/integrations/)** - Third-party integrations
- **📱 [Mobile App Examples](../examples/mobile/)** - iOS and Android integration

### Community & Support

- **🌐 Website**: [https://rendiff.dev](https://rendiff.dev)
- **📧 Email Support**: [dev@rendiff.dev](mailto:dev@rendiff.dev)
- **🐛 Bug Reports**: [GitHub Issues](https://github.com/your-org/ffprobe-api/issues)
- **💬 Community**: [GitHub Discussions](https://github.com/your-org/ffprobe-api/discussions)
- **📢 Updates**: [Twitter @RendiffDev](https://twitter.com/RendiffDev)

### Enterprise Support

For enterprise customers, we offer:

- **🎯 Priority Support**: 24/7 technical support
- **🏢 Custom Deployment**: On-premise and hybrid cloud
- **📞 Phone Support**: Direct access to engineering team
- **📋 SLA Guarantees**: 99.9% uptime guarantee
- **🎓 Training**: Technical training and onboarding

Contact [enterprise@rendiff.dev](mailto:enterprise@rendiff.dev) for more information.

## 📊 Changelog

### v2.0.0 (Latest)
- ✨ **Added**: Cloud storage integration (S3, GCS, Azure)
- ✨ **Added**: AI-powered analysis insights and Q&A
- ✨ **Added**: Enhanced HLS analysis and validation
- ✨ **Added**: Multi-format report generation (PDF, Excel, HTML)
- ✨ **Added**: Batch processing capabilities
- ✨ **Added**: Quality metrics accuracy validation
- ✨ **Added**: Comprehensive security audit and hardening
- 🚀 **Improved**: Performance optimizations for large files
- 🚀 **Improved**: Enhanced rate limiting and monitoring
- 🚀 **Improved**: WebSocket support for real-time updates
- 🐛 **Fixed**: Memory optimization for concurrent processing
- 🔒 **Security**: Input validation and injection prevention

### v1.5.0
- ✨ **Added**: Advanced quality metrics (VMAF, PSNR, SSIM)
- ✨ **Added**: Professional report templates
- 🚀 **Improved**: FFmpeg integration and error handling
- 🐛 **Fixed**: Database connection pooling issues

### v1.0.0
- 🎉 **Initial Release**: Basic media analysis functionality
- ✨ **Features**: RESTful API with JWT authentication
- ✨ **Features**: Basic quality metrics and reporting
- ✨ **Features**: Docker containerization

---

<div align="center">

**🎬 FFprobe API - Professional Media Analysis Made Simple**

**⭐ [Star on GitHub](https://github.com/your-org/ffprobe-api) • 📚 [View Documentation](https://github.com/your-org/ffprobe-api/docs) • 💬 [Join Community](https://github.com/your-org/ffprobe-api/discussions)**

</div>