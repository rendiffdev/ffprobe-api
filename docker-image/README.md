# 🐳 FFprobe API Docker Hub Image

**Production-ready Docker image for instant deployment of AI-powered video analysis API**

[![Docker Pulls](https://img.shields.io/docker/pulls/rendiffdev/ffprobe-api)](https://hub.docker.com/r/rendiffdev/ffprobe-api)
[![Docker Image Size](https://img.shields.io/docker/image-size/rendiffdev/ffprobe-api/latest)](https://hub.docker.com/r/rendiffdev/ffprobe-api)
[![Docker Image Version](https://img.shields.io/docker/v/rendiffdev/ffprobe-api)](https://hub.docker.com/r/rendiffdev/ffprobe-api)

## 🚀 One-Command Deployment (30 seconds)

### ⚡ Instant Start - Zero Configuration Required!
```bash
# Single command - works immediately
docker run -d \
  --name ffprobe-api \
  -p 8080:8080 \
  -v ffprobe_data:/app/data \
  -v ffprobe_uploads:/app/uploads \
  rendiffdev/ffprobe-api:latest

# API ready immediately at http://localhost:8080
curl http://localhost:8080/health
```

### 🎯 Full Stack with AI (Recommended)
```bash
# Download zero-config compose file
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/compose.yml

# Start everything - auto-downloads AI models, sets up cache
docker compose up -d

# Full stack ready: API + SQLite + Valkey + AI
curl http://localhost:8080/health
```

### 🏢 Production with SSL
```bash
# Production-ready deployment
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/compose.prod.yml
echo "DOMAIN=api.yourdomain.com" > .env
echo "ACME_EMAIL=admin@yourdomain.com" >> .env
docker compose -f compose.prod.yml --profile production up -d
```

## 📦 What's Included

**🎯 Zero-Configuration Docker Image - Ready to Use!**

- ✅ **SQLite Database** - Embedded, no external DB needed
- ✅ **Valkey Cache** - Redis-compatible, open source  
- ✅ **AI Analysis** - Auto-downloads Gemma3 & Phi3 models
- ✅ **Latest FFmpeg** - BtbN builds with all codecs
- ✅ **20+ QC Categories** - Professional video quality control
- ✅ **Production Optimized** - 8 workers, 20GB file support
- ✅ **Security Hardened** - Non-root user, rate limiting
- ✅ **Health Monitoring** - Built-in health checks
- ✅ **Multi-Architecture** - AMD64 & ARM64 support
- ✅ **Persistent Storage** - Named volumes for data

## 🎯 Image Variants

| Tag | Description | Size | Use Case |
|-----|-------------|------|----------|
| `latest` | Latest stable release | ~500MB | Production |
| `v1.0.0` | Specific version | ~500MB | Production with version pinning |
| `alpine` | Alpine-based minimal | ~450MB | Resource-constrained environments |
| `dev` | Development build | ~600MB | Development/testing |

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_PORT` | `8080` | API server port |
| `ENABLE_AUTH` | `false` | Enable API authentication |
| `API_KEY` | - | API key for authentication |
| `POSTGRES_HOST` | - | PostgreSQL host (optional) |
| `POSTGRES_DB` | `ffprobe_api` | Database name |
| `POSTGRES_USER` | `ffprobe` | Database user |
| `POSTGRES_PASSWORD` | - | Database password |
| `REDIS_HOST` | - | Redis host (optional) |
| `REDIS_PASSWORD` | - | Redis password |
| `ENABLE_LOCAL_LLM` | `true` | Enable AI analysis |
| `OLLAMA_URL` | `http://ollama:11434` | Ollama service URL |
| `OLLAMA_MODEL` | `gemma3:270m` | AI model to use |
| `WORKER_POOL_SIZE` | `4` | Concurrent processing workers |
| `MAX_FILE_SIZE` | `10737418240` | Max upload size (10GB) |

### Volume Mounts

| Path | Purpose | Required |
|------|---------|----------|
| `/app/uploads` | File uploads | Yes |
| `/app/reports` | Generated reports | Yes |
| `/app/data` | Application data | No |
| `/app/backup` | Backup storage | No |

## 🚢 Deployment Examples

### Standalone API Server
```bash
docker run -d \
  --name ffprobe-api \
  --restart unless-stopped \
  -p 8080:8080 \
  -e ENABLE_AUTH=true \
  -e API_KEY=$(openssl rand -hex 32) \
  -v /path/to/uploads:/app/uploads \
  -v /path/to/reports:/app/reports \
  rendiffdev/ffprobe-api:latest
```

### With PostgreSQL & Redis
```yaml
# compose.yml
version: '3.8'

services:
  api:
    image: rendiffdev/ffprobe-api:latest
    ports:
      - "8080:8080"
    environment:
      - POSTGRES_HOST=postgres
      - POSTGRES_PASSWORD=secure_password
      - REDIS_HOST=redis
      - REDIS_PASSWORD=redis_password
    depends_on:
      - postgres
      - redis
    volumes:
      - ./uploads:/app/uploads
      - ./reports:/app/reports

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=ffprobe_api
      - POSTGRES_USER=ffprobe
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass redis_password
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### With AI Analysis (Ollama)
```yaml
# docker-compose-ai.yml
version: '3.8'

services:
  api:
    image: rendiffdev/ffprobe-api:latest
    ports:
      - "8080:8080"
    environment:
      - ENABLE_LOCAL_LLM=true
      - OLLAMA_URL=http://ollama:11434
      - OLLAMA_MODEL=gemma3:270m
    depends_on:
      - ollama

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    ports:
      - "11434:11434"
    command: serve

volumes:
  ollama_data:
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ffprobe-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ffprobe-api
  template:
    metadata:
      labels:
        app: ffprobe-api
    spec:
      containers:
      - name: ffprobe-api
        image: rendiffdev/ffprobe-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENABLE_AUTH
          value: "true"
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: ffprobe-secrets
              key: api-key
        volumeMounts:
        - name: uploads
          mountPath: /app/uploads
        - name: reports
          mountPath: /app/reports
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: uploads
        persistentVolumeClaim:
          claimName: ffprobe-uploads
      - name: reports
        persistentVolumeClaim:
          claimName: ffprobe-reports
```

## 🔒 Security

The Docker image includes several security features:

- **Non-root user**: Runs as user `ffprobe` (UID 1000)
- **Minimal base image**: Alpine Linux for reduced attack surface
- **No shell access**: Production image has minimal tooling
- **Health checks**: Built-in health monitoring
- **Secure defaults**: Authentication can be enabled via environment variables
- **Read-only filesystem**: Can be run with `--read-only` flag (requires volume mounts)

### Running with Security Options
```bash
docker run -d \
  --name ffprobe-api \
  --security-opt no-new-privileges \
  --cap-drop ALL \
  --cap-add DAC_OVERRIDE \
  --read-only \
  -p 8080:8080 \
  -v /tmp:/tmp \
  -v $(pwd)/uploads:/app/uploads \
  -v $(pwd)/reports:/app/reports \
  rendiffdev/ffprobe-api:latest
```

## 🧪 Testing the Deployment

### Basic Health Check
```bash
curl http://localhost:8080/health
```

### Upload and Analyze Video
```bash
# Upload a video file
curl -X POST \
  -F "file=@sample.mp4" \
  http://localhost:8080/api/v1/probe/file

# With AI analysis
curl -X POST \
  -F "file=@sample.mp4" \
  -F "include_llm=true" \
  http://localhost:8080/api/v1/probe/file
```

### Analyze URL
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/video.mp4"}' \
  http://localhost:8080/api/v1/probe/url
```

## 📊 Resource Requirements

### Minimum Requirements
- **CPU**: 1 core
- **RAM**: 2GB
- **Disk**: 5GB
- **Docker**: 20.10+

### Recommended for Production
- **CPU**: 4+ cores
- **RAM**: 8GB
- **Disk**: 20GB
- **Docker**: 24.0+

### With AI Analysis
- **CPU**: 4+ cores
- **RAM**: 8-16GB (depends on model)
- **Disk**: 20GB+
- **GPU**: Optional but recommended for faster inference

## 🔄 Updating

### Pull Latest Image
```bash
docker pull rendiffdev/ffprobe-api:latest
docker compose down
docker compose up -d
```

### Specific Version
```bash
docker pull rendiffdev/ffprobe-api:v1.0.0
```

## 🐛 Troubleshooting

### Container Won't Start
```bash
# Check logs
docker logs ffprobe-api

# Check health
docker inspect ffprobe-api --format='{{json .State.Health}}'
```

### Permission Issues
```bash
# Fix volume permissions
sudo chown -R 1000:1000 ./uploads ./reports
```

### Port Already in Use
```bash
# Use different port
docker run -p 8081:8080 rendiffdev/ffprobe-api:latest
```

### Out of Memory
```bash
# Increase memory limit
docker run -m 4g rendiffdev/ffprobe-api:latest
```

## 🏗️ Building From Source

```bash
# Clone repository
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api/docker-image

# Build image
docker build -t ffprobe-api:local .

# Run locally built image
docker run -p 8080:8080 ffprobe-api:local
```

## 📚 Documentation

- [Main Documentation](https://github.com/rendiffdev/ffprobe-api)
- [API Reference](https://github.com/rendiffdev/ffprobe-api/blob/main/docs/api/README.md)
- [QC Features](https://github.com/rendiffdev/ffprobe-api/blob/main/QC_ANALYSIS_LIST.md)
- [Docker Compose Guide](https://github.com/rendiffdev/ffprobe-api/blob/main/docs/deployment/modern-docker-compose.md)

## 🤝 Support

- **Issues**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
- **Docker Hub**: [rendiffdev/ffprobe-api](https://hub.docker.com/r/rendiffdev/ffprobe-api)

## 📄 License

MIT License - See [LICENSE](https://github.com/rendiffdev/ffprobe-api/blob/main/LICENSE) file

---

**Ready to analyze media with AI? Pull and run the image now!** 🚀