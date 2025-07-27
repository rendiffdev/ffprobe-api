# Enhanced API Documentation

This document describes the enhanced API endpoints structure that includes advanced features like file uploads, batch processing, and real-time streaming analysis.

## Enhanced API Features

### 1. File Upload Support
- Single file upload with size validation
- Chunked upload for large files (50GB+)
- Auto-analysis after upload
- File cleanup options

### 2. Batch Processing
- Analyze multiple files in a single request
- Concurrent processing with progress tracking
- Batch status monitoring and cancellation
- Results aggregation

### 3. Real-time Streaming
- WebSocket-based real-time analysis updates
- Server-Sent Events for progress tracking
- Live stream analysis (RTMP, RTSP, HLS)
- Progress callbacks and status updates

### 4. Enhanced Probe Features
- Advanced ffprobe options support
- Custom timeout and output size limits
- Stream selection and filtering
- Multiple output formats

## API Endpoints Overview

### Core Probe Endpoints (Enhanced)
```
POST   /api/v1/probe/file         - Analyze local file
POST   /api/v1/probe/url          - Analyze remote URL
POST   /api/v1/probe/quick        - Fast analysis with minimal data
GET    /api/v1/probe/status/:id   - Get analysis status
GET    /api/v1/probe/analyses     - List user analyses
DELETE /api/v1/probe/analyses/:id - Delete analysis
GET    /api/v1/probe/health       - Service health check
```

### File Upload Endpoints
```
POST   /api/v1/upload             - Upload single file
POST   /api/v1/upload/chunk       - Upload file chunk
GET    /api/v1/upload/status/:id  - Get upload status
```

### Batch Processing Endpoints
```
POST   /api/v1/batch/analyze      - Create batch analysis
GET    /api/v1/batch/status/:id   - Get batch status
POST   /api/v1/batch/:id/cancel   - Cancel batch
GET    /api/v1/batch              - List batch operations
```

### Streaming Endpoints
```
GET    /api/v1/stream/analysis    - WebSocket for real-time updates
GET    /api/v1/stream/progress/:id - SSE progress stream
POST   /api/v1/stream/live        - Analyze live stream
```

### System Endpoints
```
GET    /health                    - System health (database + ffprobe)
GET    /api/v1/system/version     - Service version information
GET    /api/v1/system/stats       - System statistics
```

## Enhanced Features Examples

### 1. File Upload with Auto-Analysis

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Content-Type: multipart/form-data" \
  -F "file=@large_video.mp4" \
  -F "auto_analyze=true" \
  -F "delete_on_complete=true" \
  -F "async=true"
```

Response:
```json
{
  "id": "upload-123e4567-e89b-12d3-a456-426614174000",
  "file_name": "large_video.mp4",
  "file_size": 5368709120,
  "upload_path": "/tmp/uploads/upload-123_large_video.mp4",
  "analysis_id": "123e4567-e89b-12d3-a456-426614174001",
  "status": "processing",
  "message": "File uploaded and analysis started"
}
```

### 2. Chunked Upload for Large Files

```bash
# Upload first chunk
curl -X POST http://localhost:8080/api/v1/upload/chunk \
  -H "Content-Type: multipart/form-data" \
  -F "chunk=@video.part1" \
  -F "upload_id=chunk-session-123" \
  -F "chunk_number=1" \
  -F "total_chunks=10" \
  -F "file_name=huge_video.mp4"

# Upload subsequent chunks
curl -X POST http://localhost:8080/api/v1/upload/chunk \
  -H "Content-Type: multipart/form-data" \
  -F "chunk=@video.part2" \
  -F "upload_id=chunk-session-123" \
  -F "chunk_number=2" \
  -F "total_chunks=10"
```

### 3. Batch Analysis

```bash
curl -X POST http://localhost:8080/api/v1/batch/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {
        "id": "file1",
        "file_path": "/path/to/video1.mp4",
        "source_type": "local"
      },
      {
        "id": "file2",
        "file_path": "https://example.com/video2.mp4",
        "source_type": "url",
        "options": {
          "timeout": 60,
          "show_format": true,
          "show_streams": true
        }
      }
    ],
    "async": true,
    "options": {
      "output_format": "json",
      "show_format": true,
      "show_streams": true
    }
  }'
```

Response:
```json
{
  "batch_id": "batch-123e4567-e89b-12d3-a456-426614174000",
  "status": "processing",
  "total": 2,
  "completed": 0,
  "failed": 0
}
```

### 4. Real-time Analysis via WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/stream/analysis');

ws.onopen = function() {
  // Start analysis
  ws.send(JSON.stringify({
    type: 'analyze',
    id: 'req-1',
    data: {
      url: 'rtmp://live.example.com/stream',
      options: {
        show_format: true,
        show_streams: true,
        timeout: 30
      },
      interval: 5
    }
  }));
};

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  
  switch(message.type) {
    case 'update':
      const update = JSON.parse(message.data);
      console.log('Analysis update:', update);
      break;
    case 'error':
      console.error('Error:', message.error);
      break;
  }
};
```

### 5. Progress Tracking via Server-Sent Events

```javascript
const eventSource = new EventSource('/api/v1/stream/progress/123e4567-e89b-12d3-a456-426614174000');

eventSource.addEventListener('progress', function(event) {
  const progress = JSON.parse(event.data);
  console.log('Progress:', progress.progress * 100 + '%');
});

eventSource.addEventListener('complete', function(event) {
  console.log('Analysis completed!');
  eventSource.close();
});
```

### 6. Live Stream Analysis

```bash
curl -X POST http://localhost:8080/api/v1/stream/live \
  -H "Content-Type: application/json" \
  -d '{
    "url": "rtmp://live.example.com/stream/key",
    "options": {
      "show_format": true,
      "show_streams": true,
      "timeout": 30
    },
    "interval": 10
  }'
```

### 7. Advanced ffprobe Options

```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "options": {
      "output_format": "json",
      "show_format": true,
      "show_streams": true,
      "show_frames": true,
      "select_streams": "v:0,a:0",
      "read_intervals": "10%+20%",
      "count_frames": true,
      "probe_size": 10485760,
      "analyze_duration": 10000000,
      "log_level": "error",
      "pretty_print": true,
      "timeout": 300
    }
  }'
```

## Configuration

### Environment Variables

```bash
# Upload Configuration
UPLOAD_DIR=/var/uploads
MAX_FILE_SIZE=53687091200  # 50GB

# FFmpeg Configuration
FFMPEG_PATH=/usr/local/bin/ffmpeg
FFPROBE_PATH=/usr/local/bin/ffprobe

# API Configuration
API_KEY=your-secret-key
API_PORT=8080
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  ffprobe-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - UPLOAD_DIR=/app/uploads
      - MAX_FILE_SIZE=53687091200
      - POSTGRES_HOST=postgres
      - POSTGRES_DB=ffprobe_api
    volumes:
      - uploads:/app/uploads
    depends_on:
      - postgres

volumes:
  uploads:
```

## Error Handling

All endpoints return standardized error responses:

```json
{
  "error": "Description of the error",
  "details": "Additional technical details"
}
```

### Common HTTP Status Codes

- `200 OK` - Request successful
- `202 Accepted` - Request accepted for async processing
- `400 Bad Request` - Invalid request data
- `404 Not Found` - Resource not found
- `413 Payload Too Large` - File or batch too large
- `415 Unsupported Media Type` - Invalid file format
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service (ffprobe) not available

## Rate Limiting and Security

### Request ID Tracking
All requests include a `X-Request-ID` header for tracing:

```bash
curl -H "X-Request-ID: my-trace-id" http://localhost:8080/api/v1/probe/health
```

### File Security
- File type validation based on extension
- Maximum file size limits
- Automatic cleanup options
- Secure temporary file handling

### WebSocket Security
- Origin validation (configure for production)
- Connection timeout handling
- Message size limits
- Automatic cleanup of stale connections

## Performance Considerations

### File Upload
- Chunked upload for files > 1GB
- Concurrent chunk processing
- Resume capability for interrupted uploads
- Automatic file cleanup

### Batch Processing
- Configurable concurrency limits (default: 5)
- Progress tracking per file
- Cancellation support
- Memory-efficient processing

### Streaming Analysis
- Real-time progress updates
- Connection keepalive
- Automatic timeout handling
- Error recovery mechanisms

## Monitoring and Observability

### Health Checks
- Service health: `/health`
- FFprobe availability: `/api/v1/probe/health`
- Database connectivity included

### Metrics
- Request counting and timing
- File upload statistics
- Batch processing metrics
- WebSocket connection tracking

### Logging
- Structured JSON logging
- Request ID correlation
- Performance metrics
- Error tracking with context