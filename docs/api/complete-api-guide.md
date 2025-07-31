# 🎬 Complete API Usage Guide

**Professional video analysis with AI insights** - Everything you need to integrate FFprobe API into your applications.

## 🚀 Quick Start

### 1. Get Your API Key
```bash
# Your API key is in the .env file after installation
grep API_KEY .env

# Or generate a new one
openssl rand -hex 32
```

### 2. Test Connection
```bash
# Health check
curl -H "X-API-Key: YOUR_API_KEY" http://localhost:8080/health

# Expected response
{
  "status": "healthy",
  "service": "ffprobe-api",
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 3. Analyze Your First Video
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "file=@video.mp4" \
  -F "include_llm=true"
```

## 🔐 Authentication

### API Key Authentication (Recommended)
```bash
# Header format
X-API-Key: ffprobe_live_sk_1234567890abcdef...

# Example request
curl -H "X-API-Key: YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     "http://localhost:8080/api/v1/probe/health"
```

### JWT Token Authentication
```bash
# Get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your_password"}'

# Use token in requests
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     "http://localhost:8080/api/v1/probe/health"
```

## 📋 Core Endpoints

### 1. Health & Status

#### **GET** `/health`
System health check (no authentication required)
```bash
curl http://localhost:8080/health
```

#### **GET** `/api/v1/probe/health`
Service-specific health check (authentication required)
```bash
curl -H "X-API-Key: YOUR_API_KEY" \
     http://localhost:8080/api/v1/probe/health
```

**Response:**
```json
{
  "status": "healthy",
  "services": {
    "database": "connected",
    "redis": "connected",
    "ollama": "available",
    "ffmpeg": "installed"
  },
  "system": {
    "memory_usage": "45%",
    "disk_usage": "12%",
    "uptime": "2h 30m"
  }
}
```

### 2. Video Analysis

#### **POST** `/api/v1/probe/file`
Analyze uploaded video file

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "file=@video.mp4" \
  -F "include_llm=true" \
  -F "include_quality=true"
```

**Parameters:**
- `file` (required): Video file to analyze
- `include_llm` (optional): Include AI analysis (default: false)
- `include_quality` (optional): Include quality metrics (default: false)
- `format` (optional): Output format (json, xml) (default: json)

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "file_name": "video.mp4",
  "file_size": 104857600,
  "duration": "00:02:30",
  "created_at": "2024-01-15T10:30:00Z",
  "analysis": {
    "format": {
      "filename": "video.mp4",
      "nb_streams": 2,
      "nb_programs": 0,
      "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
      "format_long_name": "QuickTime / MOV",
      "start_time": "0.000000",
      "duration": "150.000000",
      "size": "104857600",
      "bit_rate": "5592405",
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
        "width": 1920,
        "height": 1080,
        "coded_width": 1920,
        "coded_height": 1080,
        "r_frame_rate": "30/1",
        "avg_frame_rate": "30/1",
        "time_base": "1/15360",
        "duration_ts": 2304000,
        "duration": "150.000000",
        "bit_rate": "5000000",
        "nb_frames": "4500"
      },
      {
        "index": 1,
        "codec_name": "aac",
        "codec_long_name": "AAC (Advanced Audio Coding)",  
        "codec_type": "audio",
        "codec_tag_string": "mp4a",
        "sample_fmt": "fltp",
        "sample_rate": "48000",
        "channels": 2,
        "channel_layout": "stereo",
        "bits_per_sample": 0,
        "duration": "150.000000",
        "bit_rate": "128000"
      }
    ]
  },
  "quality_metrics": {
    "vmaf_score": 85.6,
    "psnr": 42.3,
    "ssim": 0.95,
    "bitrate_efficiency": "good",
    "resolution_quality": "excellent"
  },
  "llm_report": {
    "overall_assessment": "High-quality video with excellent technical specifications...",
    "technical_analysis": {
      "video_quality": "excellent",
      "audio_quality": "good",
      "encoding_efficiency": "very_good"
    },
    "recommendations": [
      "Consider slight bitrate reduction for streaming",
      "Audio levels are well-balanced",
      "Excellent choice of H.264 profile for compatibility"
    ],
    "compliance": {
      "web_ready": true,
      "mobile_compatible": true,
      "streaming_optimized": true
    }
  }
}
```

#### **POST** `/api/v1/probe/url`
Analyze video from URL

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/probe/url \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4",
    "include_llm": true
  }'
```

#### **GET** `/api/v1/analysis/{id}`
Get analysis results by ID

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
     http://localhost:8080/api/v1/analysis/550e8400-e29b-41d4-a716-446655440000
```

### 3. Video Comparison

#### **POST** `/api/v1/comparisons/quick`
Quick comparison between two videos

**Request:**
```bash  
curl -X POST http://localhost:8080/api/v1/comparisons/quick \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "original_analysis_id": "uuid1",
    "modified_analysis_id": "uuid2", 
    "include_llm": true
  }'
```

**Response:**
```json
{
  "id": "comparison-uuid",
  "comparison_type": "quality_improvement",
  "status": "completed",
  "summary": {
    "quality_verdict": "improvement",
    "recommended_action": "accept",
    "confidence_score": 0.85
  },
  "metrics_comparison": {
    "file_size": {
      "original": 104857600,
      "modified": 73400320,
      "percentage_change": -30.0,
      "verdict": "significant_reduction"
    },
    "quality_scores": {
      "vmaf": {
        "original": 85.6,
        "modified": 84.2,
        "change": -1.4,
        "verdict": "minimal_loss"
      }
    },
    "bitrate": {
      "original": 5000000,
      "modified": 3500000,
      "percentage_change": -30.0
    }
  },
  "llm_assessment": "The optimization successfully reduced file size by 30% while maintaining excellent visual quality. The VMAF score decrease of 1.4 points is negligible and won't be noticeable to viewers. Recommended action: Accept these changes."
}
```

#### **GET** `/api/v1/comparisons/{id}`
Get comparison results

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
     http://localhost:8080/api/v1/comparisons/comparison-uuid
```

### 4. Quality Analysis

#### **POST** `/api/v1/quality/vmaf`
VMAF quality analysis between two videos

```bash
curl -X POST http://localhost:8080/api/v1/quality/vmaf \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "reference=@original.mp4" \
  -F "distorted=@compressed.mp4"
```

**Response:**
```json
{
  "vmaf_score": 85.6,
  "psnr": 42.3,
  "ssim": 0.95,
  "ms_ssim": 0.97,
  "frame_scores": [85.2, 85.8, 86.1, ...],
  "quality_assessment": "excellent",
  "recommendations": [
    "Quality is excellent for streaming",
    "No visible artifacts detected"
  ]
}
```

### 5. Batch Operations

#### **POST** `/api/v1/batch/analyze`
Analyze multiple files in batch

```bash
curl -X POST http://localhost:8080/api/v1/batch/analyze \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "files[]=@video1.mp4" \
  -F "files[]=@video2.mp4" \
  -F "include_llm=true"
```

**Response:**
```json
{
  "batch_id": "batch-uuid",
  "status": "processing",
  "total_files": 2,
  "completed": 0,
  "estimated_completion": "2024-01-15T10:45:00Z",
  "results": []
}
```

#### **GET** `/api/v1/batch/{id}/status`
Get batch processing status

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
     http://localhost:8080/api/v1/batch/batch-uuid/status
```

## 🛠️ Advanced Features

### Stream Analysis

#### **POST** `/api/v1/streams/analyze`
Analyze specific streams within a video

```bash
curl -X POST http://localhost:8080/api/v1/streams/analyze \
  -H "X-API-Key: YOUR_API_KEY" \
  -F "file=@video.mkv" \
  -F "stream_indexes=0,1,2"
```

### HLS Analysis

#### **POST** `/api/v1/hls/analyze`
Analyze HLS streaming manifests

```bash
curl -X POST http://localhost:8080/api/v1/hls/analyze \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "manifest_url": "https://example.com/playlist.m3u8",
    "analyze_segments": true
  }'
```

### Custom Analysis

#### **POST** `/api/v1/custom/analyze`
Run custom FFprobe commands

```bash
curl -X POST http://localhost:8080/api/v1/custom/analyze \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "file_url": "https://example.com/video.mp4",
    "ffprobe_args": ["-show_frames", "-select_streams", "v:0"],
    "output_format": "json"
  }'
```

## 📊 Response Formats

### Standard Response Structure
```json
{
  "status": "success|error|processing",
  "data": { /* response data */ },
  "message": "Human readable message",
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-uuid"
}
```

### Error Response Structure
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_FILE_FORMAT",
    "message": "Unsupported file format: .avi",
    "details": "Only MP4, MOV, MKV, and WebM formats are supported"
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-uuid"
}
```

## 🚨 Error Handling

### Common Error Codes

| Code | Description | Solution |
|------|-------------|----------|
| `AUTHENTICATION_FAILED` | Invalid API key | Check API key format and validity |
| `FILE_TOO_LARGE` | File exceeds size limit | Reduce file size or increase limit |
| `INVALID_FILE_FORMAT` | Unsupported format | Use supported formats (MP4, MOV, etc.) |
| `ANALYSIS_TIMEOUT` | Processing timeout | Try with smaller file or contact support |
| `RATE_LIMIT_EXCEEDED` | Too many requests | Wait before retry or upgrade plan |
| `SERVICE_UNAVAILABLE` | AI service down | Retry without AI or wait for recovery |

### Retry Logic Example
```bash
#!/bin/bash
analyze_video() {
    local file=$1
    local retries=3
    local delay=5
    
    for ((i=1; i<=retries; i++)); do
        response=$(curl -s -w "%{http_code}" \
            -H "X-API-Key: $API_KEY" \
            -F "file=@$file" \
            http://localhost:8080/api/v1/probe/file)
        
        http_code="${response: -3}"
        if [ "$http_code" = "200" ]; then
            echo "${response%???}" # Remove HTTP code
            return 0
        elif [ "$http_code" = "429" ]; then
            echo "Rate limited, waiting $delay seconds..."
            sleep $delay
            delay=$((delay * 2))
        else
            echo "Error: HTTP $http_code"
            return 1
        fi
    done
    echo "Failed after $retries attempts"
    return 1
}
```

## 📈 Rate Limits

### Default Limits
- **Per minute**: 60 requests
- **Per hour**: 1000 requests  
- **Per day**: 10000 requests
- **Concurrent uploads**: 4 files

### Headers
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1642250400
```

## 🔧 Configuration

### Request Headers
```bash
# Required
X-API-Key: your-api-key

# Optional  
Content-Type: application/json
Accept: application/json
User-Agent: YourApp/1.0
X-Request-ID: custom-request-id
```

### Query Parameters
```bash
# Pagination
?page=1&limit=50

# Filtering
?status=completed&include_ai=true

# Sorting
?sort=created_at&order=desc
```

## 🧪 Testing & Examples

### Test with Sample Files
```bash
# Download test videos
wget https://sample-videos.com/zip/10/mp4/SampleVideo_1280x720_1mb.mp4

# Basic analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@SampleVideo_1280x720_1mb.mp4"

# With AI analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@SampleVideo_1280x720_1mb.mp4" \
  -F "include_llm=true"
```

### SDKs and Libraries

#### Python Example
```python
import requests

class FFprobeAPI:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.headers = {'X-API-Key': api_key}
    
    def analyze_file(self, file_path, include_llm=False):
        with open(file_path, 'rb') as f:
            files = {'file': f}
            data = {'include_llm': include_llm}
            response = requests.post(
                f"{self.base_url}/api/v1/probe/file",
                headers=self.headers,
                files=files,
                data=data
            )
        return response.json()

# Usage
api = FFprobeAPI('http://localhost:8080', 'your-api-key')
result = api.analyze_file('video.mp4', include_llm=True)
print(result)
```

#### JavaScript Example
```javascript
class FFprobeAPI {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
    }

    async analyzeFile(file, includeLLM = false) {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('include_llm', includeLLM);

        const response = await fetch(`${this.baseUrl}/api/v1/probe/file`, {
            method: 'POST',
            headers: {
                'X-API-Key': this.apiKey
            },
            body: formData
        });

        return await response.json();
    }
}

// Usage
const api = new FFprobeAPI('http://localhost:8080', 'your-api-key');
const fileInput = document.getElementById('video-file');
const result = await api.analyzeFile(fileInput.files[0], true);
console.log(result);
```

## 🔍 Monitoring & Debugging

### Health Monitoring
```bash
# Continuous health check
watch -n 30 'curl -s -H "X-API-Key: $API_KEY" http://localhost:8080/api/v1/probe/health | jq'

# Service status
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/v1/status
```

### Request Tracing
```bash
# Add request ID for tracing
curl -H "X-API-Key: $API_KEY" \
     -H "X-Request-ID: trace-123" \
     http://localhost:8080/api/v1/probe/health
```

### Webhook Notifications
```bash
# Configure webhooks for analysis completion
curl -X POST http://localhost:8080/api/v1/webhooks \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhook",
    "events": ["analysis.completed", "analysis.failed"]
  }'
```

## 📚 Additional Resources

- **OpenAPI Spec**: `/api/v1/docs` (Swagger UI)
- **Postman Collection**: `docs/api/ffprobe-api.postman_collection.json`
- **SDK Documentation**: `docs/sdks/`
- **Webhook Reference**: `docs/webhooks.md`
- **Rate Limiting**: `docs/rate-limits.md`

---

## 🆘 Support

- **Documentation**: [docs/](../README.md)
- **GitHub Issues**: [Report bugs](https://github.com/rendiffdev/ffprobe-api/issues)
- **Email**: [dev@rendiff.dev](mailto:dev@rendiff.dev)
- **Health Status**: Check `/health` endpoint

**🎬 Ready to build amazing video applications!**