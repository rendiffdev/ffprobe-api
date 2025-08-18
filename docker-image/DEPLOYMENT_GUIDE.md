# 🚀 Docker Hub Deployment Guide

Complete guide for deploying the FFprobe API Docker image from Docker Hub.

## 🎯 Prerequisites

- Docker 24.0+ installed
- Docker Compose v2.20+
- 4GB+ RAM (8GB+ recommended with AI)
- 10GB+ free disk space
- Internet connection for pulling images

## 📦 Available Images

The FFprobe API is available on Docker Hub at: **`rendiffdev/ffprobe-api`**

### Image Tags

| Tag | Description | Use Case |
|-----|-------------|----------|
| `latest` | Latest stable release | Production |
| `v1.0.0` | Specific version | Version pinning |
| `alpine` | Minimal Alpine build | Resource-constrained |
| `dev` | Development build | Testing/development |

## 🚀 Quick Deployment Options

### Option 1: Single Container (Simple)

Perfect for testing or simple deployments without database persistence.

```bash
# Pull and run the image
docker pull rendiffdev/ffprobe-api:latest

# Run with basic settings
docker run -d \
  --name ffprobe-api \
  --restart unless-stopped \
  -p 8080:8080 \
  -v $(pwd)/uploads:/app/uploads \
  -v $(pwd)/reports:/app/reports \
  rendiffdev/ffprobe-api:latest

# Test the deployment
curl http://localhost:8080/health
```

### Option 2: Docker Compose (Recommended)

Complete stack with database, cache, and AI capabilities.

```bash
# Download the deployment files
mkdir ffprobe-api && cd ffprobe-api
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/compose.yml
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/.env.example

# Configure environment
cp .env.example .env
# Edit .env with your settings (see Configuration section below)

# Deploy the stack
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f ffprobe-api
```

### Option 3: Production with SSL (Advanced)

Full production setup with SSL termination and monitoring.

```bash
# Download production compose file
curl -O https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/docker-image/compose.prod.yml

# Configure SSL environment
cp .env.example .env.prod
# Set DOMAIN and ACME_EMAIL in .env.prod

# Deploy with production profile
docker compose -f compose.prod.yml --profile production up -d
```

## ⚙️ Configuration

### Environment Variables

Edit the `.env` file to configure your deployment:

```bash
# Basic Configuration
API_PORT=8080
ENABLE_AUTH=false

# Database (required for persistence)
POSTGRES_PASSWORD=your_secure_password_here
REDIS_PASSWORD=your_redis_password_here

# AI Analysis (optional but recommended)
ENABLE_LOCAL_LLM=true
OLLAMA_MODEL=gemma3:270m

# Performance Tuning
WORKER_POOL_SIZE=8
MAX_FILE_SIZE=10737418240
PROCESSING_TIMEOUT=300

# Security (enable for production)
ENABLE_AUTH=true
API_KEY=your_api_key_here
JWT_SECRET=your_jwt_secret_here
```

### Generate Secure Secrets

```bash
# Generate API key
openssl rand -hex 32

# Generate JWT secret
openssl rand -hex 32

# Generate database password
openssl rand -hex 16
```

### Volume Configuration

The container uses these important volume mounts:

```yaml
volumes:
  - ./uploads:/app/uploads      # File uploads (required)
  - ./reports:/app/reports      # Generated reports (required)
  - ./data:/app/data           # Application data (optional)
  - ./backup:/app/backup       # Backup storage (optional)
```

Create these directories with proper permissions:

```bash
mkdir -p uploads reports data backup
chmod 755 uploads reports data backup
```

## 🏗️ Deployment Scenarios

### Development Environment

```bash
# Quick development setup
docker run -d \
  --name ffprobe-dev \
  -p 8080:8080 \
  -e ENABLE_AUTH=false \
  -e ENABLE_RATE_LIMIT=false \
  -v $(pwd)/uploads:/app/uploads \
  rendiffdev/ffprobe-api:latest
```

### Production Environment

```bash
# Production deployment with all services
version: '3.8'
services:
  api:
    image: rendiffdev/ffprobe-api:latest
    restart: always
    environment:
      - ENABLE_AUTH=true
      - API_KEY=${API_KEY}
      - POSTGRES_HOST=postgres
      - REDIS_HOST=redis
      - ENABLE_LOCAL_LLM=true
    depends_on:
      - postgres
      - redis
      - ollama
    volumes:
      - ./uploads:/app/uploads
      - ./reports:/app/reports
      - ./backup:/app/backup
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3

  postgres:
    image: postgres:16-alpine
    restart: always
    environment:
      POSTGRES_DB: ffprobe_api
      POSTGRES_USER: ffprobe
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backup/postgres:/backup

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data

  ollama:
    image: ollama/ollama:latest
    restart: always
    volumes:
      - ollama_data:/root/.ollama
    deploy:
      resources:
        limits:
          memory: 4G

volumes:
  postgres_data:
  redis_data:
  ollama_data:
```

### High Availability Setup

```bash
# Load balancer with multiple API instances
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - api1
      - api2
      - api3

  api1: &api
    image: rendiffdev/ffprobe-api:latest
    environment:
      - POSTGRES_HOST=postgres
      - REDIS_HOST=redis
    volumes:
      - shared_uploads:/app/uploads
      - shared_reports:/app/reports

  api2: *api
  api3: *api

  postgres:
    image: postgres:16-alpine
    # Configure with replication for HA

  redis:
    image: redis:7-alpine
    # Configure Redis Cluster for HA

volumes:
  shared_uploads:
  shared_reports:
```

## 🔍 Health Monitoring

### Health Checks

The image includes built-in health checks:

```bash
# Check container health
docker inspect ffprobe-api --format='{{json .State.Health}}'

# Manual health check
curl http://localhost:8080/health
```

### Monitoring Endpoints

```bash
# API health
GET /health

# Metrics (if enabled)
GET /metrics

# Service info
GET /info
```

### Log Management

```bash
# View logs
docker compose logs -f ffprobe-api

# Log rotation (add to compose.yml)
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

## 🚨 Troubleshooting

### Common Issues

#### 1. Container Won't Start

```bash
# Check logs
docker logs ffprobe-api

# Common fixes:
# - Check port availability: lsof -i :8080
# - Verify volume permissions: ls -la uploads/
# - Check environment variables: docker exec ffprobe-api env
```

#### 2. Out of Memory

```bash
# Increase container memory limit
docker run -m 4g rendiffdev/ffprobe-api:latest

# Or in compose.yml:
deploy:
  resources:
    limits:
      memory: 4G
```

#### 3. FFmpeg Errors

```bash
# Test FFmpeg inside container
docker exec ffprobe-api ffmpeg -version

# Check VMAF models
docker exec ffprobe-api ls -la /usr/local/share/vmaf/
```

#### 4. Database Connection Issues

```bash
# Check database connectivity
docker exec ffprobe-api ping postgres

# Verify credentials
docker exec postgres psql -U ffprobe -d ffprobe_api -c "\l"
```

#### 5. AI/LLM Not Working

```bash
# Check Ollama service
curl http://localhost:11434/api/version

# Pull models manually
docker exec ollama ollama pull gemma3:270m

# Check model list
docker exec ollama ollama list
```

### Performance Tuning

```bash
# Optimize for your workload
environment:
  # Increase workers for CPU-bound tasks
  - WORKER_POOL_SIZE=16
  
  # Adjust timeout for large files
  - PROCESSING_TIMEOUT=600
  
  # Tune upload limits
  - MAX_FILE_SIZE=53687091200  # 50GB
  
  # Memory optimization
  - GOMEMLIMIT=4GiB
```

### Security Hardening

```bash
# Run with additional security options
docker run \
  --security-opt no-new-privileges \
  --cap-drop ALL \
  --cap-add DAC_OVERRIDE \
  --read-only \
  -p 8080:8080 \
  -v /tmp:/tmp \
  -v $(pwd)/uploads:/app/uploads \
  rendiffdev/ffprobe-api:latest
```

## 🔄 Updates and Maintenance

### Updating the Image

```bash
# Pull latest version
docker pull rendiffdev/ffprobe-api:latest

# Stop and remove old container
docker compose down

# Start with new image
docker compose up -d

# Clean up old images
docker image prune
```

### Backup and Restore

```bash
# Backup data
docker exec postgres pg_dump -U ffprobe ffprobe_api > backup.sql
tar -czf uploads-backup.tar.gz uploads/
tar -czf reports-backup.tar.gz reports/

# Restore data
docker exec -i postgres psql -U ffprobe ffprobe_api < backup.sql
tar -xzf uploads-backup.tar.gz
tar -xzf reports-backup.tar.gz
```

### Database Migrations

Migrations run automatically on container start. To run manually:

```bash
# Database initialization is automatic with SQLite
# No manual migrations needed - schema created on first startup
```

## 📊 Production Checklist

- [ ] Set secure passwords for all services
- [ ] Enable authentication (`ENABLE_AUTH=true`)
- [ ] Configure SSL/TLS termination
- [ ] Set up log rotation
- [ ] Configure backup strategy
- [ ] Set resource limits
- [ ] Enable monitoring and alerting
- [ ] Test disaster recovery procedures
- [ ] Review security settings
- [ ] Configure firewall rules

## 🤝 Support

- **Issues**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **Documentation**: [Main Docs](https://github.com/rendiffdev/ffprobe-api)
- **Docker Hub**: [rendiffdev/ffprobe-api](https://hub.docker.com/r/rendiffdev/ffprobe-api)

---

**Ready to deploy? Choose your deployment option and get started!** 🚀