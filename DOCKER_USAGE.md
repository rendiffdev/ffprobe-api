# FFprobe API - Docker Usage Guide

A simple, ready-to-use Docker image for video file analysis using FFprobe.

## Quick Start

### Pull and Run
```bash
# Download and run the image
docker run -d -p 8080:8080 rendiffdev/ffprobe-api:minimal

# Or use the latest tag
docker run -d -p 8080:8080 rendiffdev/ffprobe-api:latest
```

### Health Check
```bash
curl http://localhost:8080/health
```

### Analyze a Video File
```bash
# Upload and analyze a video file
curl -X POST http://localhost:8080/api/v1/probe -F "file=@your-video.mp4"
```

## Available Images

- `rendiffdev/ffprobe-api:minimal` - Main minimal working image
- `rendiffdev/ffprobe-api:latest` - Alias for minimal
- `rendiffdev/ffprobe-api:working` - Alias for minimal

## API Endpoints

### Health Check
- **URL**: `GET /health`
- **Response**: JSON with status information

```json
{
  "status": "healthy",
  "timestamp": "2025-01-18T12:00:00.000000",
  "version": "1.0.0-minimal",
  "ffprobe_available": true
}
```

### Video Analysis
- **URL**: `POST /api/v1/probe`
- **Content-Type**: `multipart/form-data`
- **Parameter**: `file` (video file to analyze)

```json
{
  "file_id": "uuid-here",
  "filename": "your-video.mp4",
  "analysis": {
    "format": { ... },
    "streams": [ ... ]
  },
  "timestamp": "2025-01-18T12:00:00.000000"
}
```

### Version Information
- **URL**: `GET /api/v1/version`

## Docker Compose Example

```yaml
version: '3.8'

services:
  ffprobe-api:
    image: rendiffdev/ffprobe-api:minimal
    ports:
      - "8080:8080"
    volumes:
      - "./uploads:/app/uploads"
      - "./reports:/app/reports"
      - "./logs:/app/logs"
    environment:
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

## Persistent Storage

The image provides volume mount points for:

- `/app/uploads` - Uploaded files (temporary storage)
- `/app/reports` - Analysis reports 
- `/app/logs` - Application logs

## Platform Support

- **Architecture**: AMD64/x86_64 only
- **Optimized for**: Cloud instances, CI/CD, production deployments
- **Base**: Alpine Linux 3.20

## Features

✅ **Zero Configuration** - Works out of the box  
✅ **Latest FFmpeg** - BtbN static builds included  
✅ **Python Flask API** - Simple REST interface  
✅ **Health Checks** - Docker native health monitoring  
✅ **Security** - Non-root user, minimal attack surface  
✅ **Lightweight** - Alpine Linux base (~150MB)  

## Examples

### Basic Usage
```bash
# Start the service
docker run -d --name ffprobe-api -p 8080:8080 rendiffdev/ffprobe-api:minimal

# Wait for startup
sleep 5

# Check if it's working
curl http://localhost:8080/health

# Analyze a video file
curl -X POST http://localhost:8080/api/v1/probe -F "file=@sample.mp4" | jq .

# Stop the service
docker stop ffprobe-api && docker rm ffprobe-api
```

### With Volume Mounting
```bash
# Create local directories
mkdir -p ./data/{uploads,reports,logs}

# Run with persistent storage
docker run -d --name ffprobe-api \
  -p 8080:8080 \
  -v ./data/uploads:/app/uploads \
  -v ./data/reports:/app/reports \
  -v ./data/logs:/app/logs \
  rendiffdev/ffprobe-api:minimal
```

### Production Deployment
```bash
# Run with resource limits and restart policy
docker run -d --name ffprobe-api \
  -p 8080:8080 \
  --memory=1g \
  --cpus=2 \
  --restart=unless-stopped \
  -v ffprobe-uploads:/app/uploads \
  -v ffprobe-reports:/app/reports \
  -v ffprobe-logs:/app/logs \
  rendiffdev/ffprobe-api:minimal
```

## Troubleshooting

### Container won't start
```bash
# Check logs
docker logs ffprobe-api

# Check if port is available
netstat -tulnp | grep :8080
```

### Health check fails
```bash
# Check container status
docker ps -a

# Check health check details
docker inspect ffprobe-api | jq '.[0].State.Health'
```

### Analysis fails
- Ensure uploaded file is a valid video format
- Check file size (default limit: 100MB)
- Verify FFmpeg supports the codec

## Limitations

- AMD64 architecture only (use on Intel/AMD servers)
- 100MB file size limit
- Basic FFprobe analysis only (no advanced features)
- Single-threaded processing

## Support

This is a simplified working version. For the full-featured implementation:
- Check the GitHub repository
- Review the original Go-based API
- Consider building from source for complete features

---

**Docker Hub**: https://hub.docker.com/r/rendiffdev/ffprobe-api  
**Ready for immediate deployment!** 🚀