# FFprobe API Reference

**🧠 AI-Powered Video Analysis API - Beyond Traditional FFprobe**

**The only media analysis API with built-in GenAI intelligence** - transforming raw FFprobe data into actionable professional insights, recommendations, and risk assessments.

**Key GenAI Differentiator:** Every endpoint supports `"include_llm": true` parameter for AI-powered analysis

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

The API supports multiple authentication methods:

### API Key Authentication (Recommended)
```bash
curl -H "X-API-Key: your-api-key" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/v1/probe/file
```

### JWT Token Authentication
```bash
curl -H "Authorization: Bearer your-jwt-token" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/v1/probe/file
```

## Core API Endpoints

### 🎬 Video Analysis

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/probe/file` | POST | Analyze uploaded video file | Required |
| `/probe/url` | POST | Analyze video from URL | Required |
| `/probe/quick` | POST | Fast basic analysis | Required |
| `/probe/hls` | POST | Analyze HLS streams | Required |
| `/probe/status/{id}` | GET | Get analysis status | Required |
| `/probe/analyses` | GET | List user analyses | Required |
| `/probe/analyses/{id}` | DELETE | Delete analysis | Required |

### 🔍 Quality & Comparison

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/quality/compare` | POST | Compare video quality | Required |
| `/quality/analysis/{id}` | GET | Get quality analysis | Required |
| `/quality/statistics` | GET | Quality statistics | Required |
| `/comparisons` | POST | Create video comparison | Required |
| `/comparisons/{id}` | GET | Get comparison results | Required |
| `/comparisons` | GET | List comparisons | Required |

### 📊 Reports & Data

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/reports/analysis` | POST | Generate analysis report | Required |
| `/reports/comparison` | POST | Generate comparison report | Required |
| `/reports/formats` | GET | List report formats | Required |
| `/probe/raw/{id}` | GET | Get raw FFprobe data | Required |
| `/probe/download/{id}` | GET | Download report | Required |

### ⚡ Batch & Streaming

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/batch/analyze` | POST | Batch video processing | Required |
| `/batch/status/{id}` | GET | Get batch status | Required |
| `/stream/analysis` | GET | Stream analysis (WebSocket) | Required |
| `/stream/progress/{id}` | GET | Stream progress | Required |

### 📁 Storage & Upload

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/upload` | POST | Upload video file | Required |
| `/upload/chunk` | POST | Chunked upload | Required |
| `/storage/upload` | POST | Storage upload | Required |
| `/storage/download/{key}` | GET | Download from storage | Required |

### 🔐 Authentication & Keys

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/auth/login` | POST | User login | None |
| `/auth/refresh` | POST | Refresh token | None |
| `/auth/logout` | POST | User logout | Required |
| `/keys/create` | POST | Create API key | Required |
| `/keys/rotate` | POST | Rotate API key | Required |

### 🎯 GraphQL

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/graphql` | POST | GraphQL endpoint | Optional |
| `/graphql/playground` | GET | GraphQL playground (dev) | None |
| `/graphql/schema` | GET | GraphQL schema | None |

## Sample Requests

### Analyze a Video File

```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -F "file=@video.mp4" \
  -F "include_llm=true" \
  -F "quality_analysis=true"
```

**Response:**
```json
{
  "analysis_id": "uuid-here",
  "status": "completed",
  "file_info": {
    "filename": "video.mp4",
    "size": 1048576,
    "duration": 60.5,
    "format": "mp4"
  },
  "video_streams": [
    {
      "index": 0,
      "codec_name": "h264",
      "width": 1920,
      "height": 1080,
      "frame_rate": "30/1",
      "bit_rate": "5000000"
    }
  ],
  "audio_streams": [
    {
      "index": 1,
      "codec_name": "aac",
      "sample_rate": 48000,
      "channels": 2,
      "bit_rate": "128000"
    }
  ],
  "quality_metrics": {
    "vmaf_score": 85.2,
    "psnr": 42.1,
    "ssim": 0.95
  },
  "llm_report": "🧠 EXECUTIVE SUMMARY: Professional HD content ready for broadcast delivery. TECHNICAL ANALYSIS: H.264 encoding at optimal 5Mbps bitrate for 1080p resolution. QUALITY ASSESSMENT: Excellent visual quality (VMAF 85.2), no artifacts detected. RECOMMENDATIONS: 1) Consider HEVC encoding for 40% size reduction while maintaining quality. 2) Audio levels optimal for broadcast standards. 3) Suitable for Netflix, YouTube, and OTT platforms. RISK ASSESSMENT: Low technical risk, fully compliant with industry standards. WORKFLOW INTEGRATION: Ready for immediate delivery pipeline.",
  "llm_enabled": true
}
```

### Compare Two Videos

```bash
curl -X POST http://localhost:8080/api/v1/comparisons \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "original_analysis_id": "uuid-1",
    "modified_analysis_id": "uuid-2",
    "comparison_type": "quality",
    "include_ai_analysis": true
  }'
```

**Response:**
```json
{
  "comparison_id": "uuid-comparison",
  "status": "completed",
  "original_file": {
    "filename": "original.mp4",
    "quality_score": 75.2
  },
  "modified_file": {
    "filename": "optimized.mp4",
    "quality_score": 85.1
  },
  "improvement": {
    "quality_delta": 9.9,
    "size_reduction": "15%",
    "is_improvement": true
  },
  "ai_assessment": "The modified version shows significant quality improvement..."
}
```

### Generate Analysis Report

```bash
curl -X POST http://localhost:8080/api/v1/reports/analysis \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "analysis_id": "uuid-here",
    "format": "pdf",
    "include_recommendations": true,
    "include_charts": true
  }'
```

## Error Responses

The API uses standard HTTP status codes and returns JSON error responses:

```json
{
  "error": {
    "code": "INVALID_FILE_FORMAT",
    "message": "Unsupported video format",
    "details": "Only MP4, AVI, MOV, and MKV formats are supported"
  },
  "request_id": "req-12345678"
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_API_KEY` | 401 | Invalid or missing API key |
| `INSUFFICIENT_PERMISSIONS` | 403 | User lacks required permissions |
| `FILE_TOO_LARGE` | 413 | File exceeds maximum size limit |
| `INVALID_FILE_FORMAT` | 400 | Unsupported video format |
| `PROCESSING_FAILED` | 500 | Video analysis failed |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |

## Rate Limits

| Limit Type | Default | Scope |
|------------|---------|-------|
| Per Minute | 60 requests | Per API key |
| Per Hour | 1000 requests | Per API key |
| Per Day | 10000 requests | Per API key |

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1640995200
```

## AI-Powered Features

### LLM Analysis
- **Models**: Gemma 3 270M (primary), Phi-3 Mini (fallback)
- **Response Time**: 0.5-3 seconds
- **Analysis Sections**: 8 professional sections including quality assessment, optimization recommendations

### Quality Metrics
- **VMAF**: Video Multi-method Assessment Fusion scores
- **PSNR**: Peak Signal-to-Noise Ratio
- **SSIM**: Structural Similarity Index
- **Custom**: Blockiness, blur, noise detection

## WebSocket Endpoints

### Real-time Analysis Progress
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/stream/analysis');

ws.onmessage = function(event) {
  const progress = JSON.parse(event.data);
  console.log(`Progress: ${progress.percentage}%`);
};
```

## GraphQL Schema

The API provides a comprehensive GraphQL schema for complex queries:

```graphql
query GetAnalysisWithComparisons($id: ID!) {
  analysis(id: $id) {
    id
    filename
    status
    videoStreams {
      codecName
      width
      height
      bitRate
    }
    qualityMetrics {
      vmafScore
      psnrAvg
      ssimAvg
    }
    comparisons {
      id
      originalFile
      modifiedFile
      qualityImprovement
    }
  }
}
```

---

## 🚀 Quick Start Integration

```bash
# 1. Get API key
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password"}'

# 2. Analyze video
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -F "file=@video.mp4"

# 3. Get results
curl http://localhost:8080/api/v1/probe/analyses \
  -H "X-API-Key: your-api-key"
```

## SDKs and Libraries

- **JavaScript/Node.js**: Coming soon
- **Python**: Coming soon  
- **Go**: Native API client examples available
- **cURL**: Complete examples in this documentation

---

**For more examples and advanced usage, see our [API Usage Tutorial](../tutorials/api_usage.md)**