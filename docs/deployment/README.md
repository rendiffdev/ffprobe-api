# Deployment Guide

> **Complete guide for deploying FFprobe API in development, production, and enterprise environments**

## Overview

FFprobe API supports multiple deployment strategies to meet different scales and requirements:

- **Development**: Quick setup for testing and development
- **Production**: Medium-scale deployment with monitoring
- **Enterprise**: High-volume deployment with clustering and advanced monitoring

## Quick Start

### Prerequisites

- Docker and Docker Compose 
- 4GB+ RAM, 2+ CPU cores (minimum)
- 10GB+ available disk space

### Basic Deployment

```bash
# Clone repository
git clone <your-repo-url>
cd ffprobe-api

# Start all services
docker compose -f docker-image/compose.yaml up -d

# Verify deployment
curl http://localhost:8080/health
```

## Deployment Options

### 1. Development Setup

**Best for**: Local development, testing, small-scale usage

```bash
# Basic services only
docker compose -f docker-image/compose.yaml up -d

# Verify services
docker compose -f docker-image/compose.yaml ps
```

**Resources**: 4GB RAM, 2 CPU cores, 10GB storage

### 2. Production Setup

**Best for**: Medium-scale production deployments

```bash
# Production configuration with monitoring
docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml up -d

# Scale API instances
docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml up -d --scale ffprobe-api=2
```

**Resources**: 8GB RAM, 4 CPU cores, 50GB storage  
**Features**: Load balancing, monitoring, backup automation

### 3. Enterprise Setup

**Best for**: High-volume production with advanced monitoring

```bash
# Enterprise deployment with full monitoring stack
docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml up -d \
  --scale ffprobe-api=3 \
  --scale ollama=2
```

**Resources**: 16GB+ RAM, 8+ CPU cores, 100GB+ storage  
**Features**: Clustering, advanced monitoring, alerting, backup automation

## Performance Scaling

### Resource Requirements

| Deployment | RAM | CPU | Storage | Concurrent Jobs | Throughput |
|------------|-----|-----|---------|-----------------|------------|
| Development | 4GB | 2 cores | 10GB | 2-5 | 10-50 req/min |
| Production | 8GB | 4 cores | 50GB | 5-15 | 60-200 req/min |
| Enterprise | 16GB+ | 8+ cores | 100GB+ | 15-50 | 200-1000 req/min |

### Scaling Configuration

```bash
# Horizontal scaling
docker compose -f docker-image/compose.yaml up -d --scale ffprobe-api=3

# Resource limits (production)
docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml up -d

# Enterprise with monitoring
docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml up -d
```

## Environment Configuration

### Required Environment Variables

```bash
# Server Configuration
API_PORT=8080
API_HOST=0.0.0.0
BASE_URL=http://localhost:8080

# Database Configuration  
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=ffprobe_api
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-secure-password

# Authentication
API_KEY=your-32-char-api-key-here
JWT_SECRET=your-32-char-jwt-secret-here
```

### Optional Configuration

```bash
# LLM Integration
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=mistral:7b
OPENROUTER_API_KEY=your-openrouter-key

# Storage
MAX_FILE_SIZE=53687091200  # 50GB
UPLOAD_DIR=/app/uploads
REPORTS_DIR=/app/reports

# Security
ENABLE_RATE_LIMIT=true
RATE_LIMIT_PER_MINUTE=60
ALLOWED_ORIGINS=*
```

## Cloud Deployment

### Docker Compose (Recommended)

```yaml
version: '3.8'
services:
  ffprobe-api:
    image: ffprobe-api:latest
    ports:
      - "8080:8080"
    environment:
      - POSTGRES_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: ffprobe_api
      POSTGRES_PASSWORD: secure-password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
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
        image: ffprobe-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: POSTGRES_HOST
          value: "postgres-service"
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
```

## Cloud Provider Guides

### AWS Deployment

- **ECS**: Container orchestration with auto-scaling
- **EKS**: Kubernetes-based deployment
- **EC2**: Traditional server deployment
- **RDS**: Managed PostgreSQL database
- **ElastiCache**: Managed Redis

### Google Cloud Platform

- **Cloud Run**: Serverless container deployment
- **GKE**: Kubernetes Engine deployment
- **Compute Engine**: VM-based deployment
- **Cloud SQL**: Managed PostgreSQL
- **Memorystore**: Managed Redis

### Microsoft Azure

- **Container Instances**: Simple container deployment
- **AKS**: Azure Kubernetes Service
- **App Service**: Platform-as-a-Service deployment
- **PostgreSQL**: Managed database service
- **Redis Cache**: Managed Redis service

## Security Considerations

### Production Security Checklist

- [ ] Change default API keys and JWT secrets
- [ ] Enable HTTPS with SSL/TLS certificates
- [ ] Configure firewall rules and network security
- [ ] Set up proper backup and disaster recovery
- [ ] Enable monitoring and alerting
- [ ] Configure log rotation and retention
- [ ] Implement proper access controls
- [ ] Regular security updates and patches

### Network Security

```bash
# Firewall configuration
ufw allow 22    # SSH
ufw allow 80    # HTTP
ufw allow 443   # HTTPS
ufw deny 8080   # Block direct API access (use reverse proxy)
```

### Reverse Proxy Configuration

```nginx
# Nginx configuration
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Monitoring and Observability

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/health

# Detailed system status
curl http://localhost:8080/api/v1/system/status
```

### Prometheus Metrics

Available at `http://localhost:8080/metrics`:

- Request rates and response times
- Processing queue status
- Database connection health
- System resource utilization

### Logging

```bash
# View application logs
docker compose -f docker-image/compose.yaml logs -f ffprobe-api

# Monitor specific service
docker compose -f docker-image/compose.yaml logs -f postgres
```

## Backup and Recovery

### Database Backup

```bash
# Create backup
docker compose -f docker-image/compose.yaml exec postgres pg_dump -U postgres ffprobe_api > backup.sql

# Restore backup
docker compose -f docker-image/compose.yaml exec -T postgres psql -U postgres ffprobe_api < backup.sql
```

### Configuration Backup

```bash
# Backup configuration
cp .env .env.backup
tar -czf config-backup.tar.gz docker/configs/
```

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8080, 5432, 6379 are available
2. **Memory issues**: Increase Docker memory allocation
3. **Permission errors**: Check file permissions and Docker user
4. **Database connection**: Verify PostgreSQL service is running

### Performance Tuning

```bash
# Monitor resource usage
docker stats

# Check processing queue
curl http://localhost:8080/api/v1/batch/status

# Database performance
docker compose exec postgres psql -U postgres -c "SELECT * FROM pg_stat_activity;"
```

## Migration and Updates

### Version Updates

```bash
# Pull latest images
docker compose -f docker-image/compose.yaml pull

# Restart with new version
docker compose -f docker-image/compose.yaml up -d

# Verify update
curl http://localhost:8080/health
```

### Database Migrations

Database migrations run automatically on startup. For manual migration:

```bash
# Run migrations manually
docker compose -f docker-image/compose.yaml exec ffprobe-api ./migrate -path ./migrations -database "postgres://..." up
```

---

## Next Steps

- [Monitoring Setup](../operations/monitoring.md) 
- [Security Guide](../operations/security.md)
- [Production Readiness Report](../PRODUCTION_READINESS_REPORT.md)
- [Troubleshooting Guide](../../README.md#troubleshooting)