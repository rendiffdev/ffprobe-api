# 🎯 API Usage Tutorial

Step-by-step guide to using the FFprobe API effectively.

## Prerequisites

Make sure you have the API running. Choose your deployment:

```bash
# Quick deployment (recommended for testing)
docker compose --profile quick up -d

# Production deployment
docker compose -f compose.yaml -f compose.production.yaml up -d
```

Set up your API key (required for most endpoints):
```bash
export API_KEY="your-api-key-here"
```

## Health Check

Check if the API and ffprobe service are healthy:

```bash
curl -X GET http://localhost:8080/health
# Or with API key if authentication is enabled:
curl -H "X-API-Key: $API_KEY" -X GET http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "service": "ffprobe-api",
  "version": "v1.0.0",
  "database": "healthy",
  "stats": {
    "open_connections": 1,
    "in_use": 0,
    "idle": 1
  }
}
```

## Probe Service Health

Check ffprobe-specific health:

```bash
curl -X GET http://localhost:8080/api/v1/probe/health
```

Response:
```json
{
  "status": "healthy",
  "ffprobe_version": "ffprobe version 4.4.0"
}
```

## Probe Local File

Analyze a local media file:

```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "source_type": "local",
    "async": false
  }'
```

Response:
```json
{
  "analysis_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "completed",
  "analysis": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "file_name": "video.mp4",
    "file_path": "/path/to/video.mp4",
    "file_size": 1048576,
    "content_hash": "abc123...",
    "source_type": "local",
    "status": "completed",
    "ffprobe_data": {
      "format": {
        "filename": "/path/to/video.mp4",
        "nb_streams": 2,
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
          "duration": "120.5"
        },
        {
          "index": 1,
          "codec_name": "aac",
          "codec_type": "audio",
          "channels": 2,
          "sample_rate": "44100"
        }
      ]
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:30Z",
    "processed_at": "2024-01-01T00:00:30Z"
  }
}
```

## Probe Remote URL

Analyze a remote media file:

```bash
curl -X POST http://localhost:8080/api/v1/probe/url \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/video.mp4",
    "async": false,
    "timeout": 60
  }'
```

## Async Probe

Start an async analysis:

```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/large-video.mp4",
    "source_type": "local",
    "async": true
  }'
```

Response:
```json
{
  "analysis_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "processing",
  "message": "Analysis started, check status endpoint for progress"
}
```

## Check Analysis Status

Check the status of an analysis:

```bash
curl -X GET http://localhost:8080/api/v1/probe/status/123e4567-e89b-12d3-a456-426614174000
```

Response:
```json
{
  "analysis_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "completed",
  "analysis": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "status": "completed",
    "ffprobe_data": { ... }
  }
}
```

## Quick Probe

Perform a fast analysis with minimal information:

```bash
curl -X POST http://localhost:8080/api/v1/probe/quick \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "source_type": "local"
  }'
```

## List Analyses

List all analyses for the current user:

```bash
curl -X GET "http://localhost:8080/api/v1/probe/analyses?limit=10&offset=0"
```

Response:
```json
{
  "analyses": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "file_name": "video1.mp4",
      "status": "completed",
      "created_at": "2024-01-01T00:00:00Z"
    },
    {
      "id": "456e7890-e89b-12d3-a456-426614174001",
      "file_name": "video2.mp4",
      "status": "processing",
      "created_at": "2024-01-01T00:01:00Z"
    }
  ],
  "limit": 10,
  "offset": 0,
  "count": 2
}
```

## Delete Analysis

Delete an analysis:

```bash
curl -X DELETE http://localhost:8080/api/v1/probe/analyses/123e4567-e89b-12d3-a456-426614174000
```

Response: 204 No Content

## Custom ffprobe Options

Use custom ffprobe options:

```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "source_type": "local",
    "options": {
      "output_format": "json",
      "show_format": true,
      "show_streams": true,
      "show_chapters": true,
      "select_streams": "v:0",
      "count_frames": true,
      "probe_size": 1048576,
      "analyze_duration": 5000000
    }
  }'
```

## System Information

Get system version information:

```bash
curl -X GET http://localhost:8080/api/v1/system/version
```

Response:
```json
{
  "service": "ffprobe-api",
  "version": "v1.0.0",
  "api_version": "v1",
  "build_time": "2024-01-01T00:00:00Z",
  "commit": "abc123..."
}
```

## System Stats

Get system statistics:

```bash
curl -X GET http://localhost:8080/api/v1/system/stats
```

Response:
```json
{
  "uptime": "1h30m",
  "requests_total": 1042,
  "active_jobs": 3,
  "memory_usage": "128MB",
  "database": {
    "open_connections": 5,
    "in_use": 2,
    "idle": 3
  }
}
```

## Error Handling

All endpoints return standardized error responses:

```json
{
  "error": "Error message",
  "details": "Additional error details"
}
```

Common HTTP status codes:
- 200: Success
- 202: Accepted (for async operations)
- 400: Bad Request (invalid input)
- 404: Not Found
- 500: Internal Server Error
- 503: Service Unavailable (ffprobe not available)

## Request ID Tracking

All responses include a `X-Request-ID` header for tracking requests. You can also provide your own request ID:

```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: my-custom-id-123" \
  -d '{"file_path": "/path/to/video.mp4"}'
```