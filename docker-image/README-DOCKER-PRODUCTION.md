# FFprobe API - Production Docker Infrastructure

## 🏗️ Production-Grade Docker Infrastructure

This directory contains a comprehensive, enterprise-ready Docker infrastructure for deploying FFprobe API in production environments. The infrastructure has been completely optimized for security, performance, scalability, and operational excellence.

### 📊 Infrastructure Overview

```
docker-image/
├── Dockerfile.optimized              # Multi-stage production Dockerfile
├── compose.production.optimized.yaml # Production Docker Compose
├── build-optimized.sh               # Enhanced build automation
├── deploy-production.sh             # Production deployment script
├── config/                          # Production configurations
│   ├── prometheus/                  # Monitoring configuration
│   ├── grafana/                    # Dashboard configuration
│   └── traefik/                    # Reverse proxy settings
├── security/                       # Security policies & configs
│   ├── docker-security.yaml        # Security overlay
│   ├── seccomp-profile.json       # System call filtering
│   └── ...                        # Additional security configs
└── scripts/                        # Operational scripts
    ├── secrets-manager.sh          # Secrets management
    ├── backup/                     # Backup automation
    └── monitoring/                 # Health checks
```

## 🚀 Quick Start

### 1. Build Production Image

```bash
# Build with security scanning and multi-arch support
./docker-image/build-optimized.sh \
  --target production \
  --multi-arch \
  --scan \
  --sbom

# Alternative: Basic production build
./docker-image/build-optimized.sh --target production
```

### 2. Generate Secrets

```bash
# Generate all production secrets
./docker-image/scripts/secrets-manager.sh generate

# Verify secrets are created
./docker-image/scripts/secrets-manager.sh verify
```

### 3. Deploy to Production

```bash
# Complete production deployment
./docker-image/deploy-production.sh \
  --mode compose \
  --environment production \
  --domain api.yourcompany.com \
  --enable-ssl \
  --enable-monitoring \
  --enable-backup \
  --deploy

# Docker Swarm deployment
./docker-image/deploy-production.sh \
  --mode swarm \
  --environment production \
  --deploy
```

## 🛠️ Build System

### Multi-Stage Dockerfile Features

- **Security Hardened**: Non-root user, minimal attack surface, read-only filesystem
- **Multi-Architecture**: Native support for AMD64 and ARM64
- **Optimized Layers**: Efficient caching and minimal image size
- **Multiple Targets**: Production, development, test, minimal, security-scan

#### Available Build Targets

| Target | Purpose | Size | Features |
|--------|---------|------|----------|
| `production` | Production deployment | ~150MB | Security hardened, optimized |
| `development` | Local development | ~300MB | Hot reload, debugging tools |
| `test` | CI/CD testing | ~250MB | Test frameworks, coverage |
| `minimal` | Ultra-minimal | ~50MB | Scratch-based, no shell |
| `security-scan` | Security analysis | ~200MB | Embedded scanning tools |

### Build Script Features

```bash
./build-optimized.sh --help  # Full options list

# Security-focused build
./build-optimized.sh \
  --target production \
  --scan \
  --sign \
  --sbom \
  --push

# Multi-architecture build
./build-optimized.sh \
  --multi-arch \
  --registry ghcr.io/yourorg \
  --push

# Development build with custom args
./build-optimized.sh \
  --target development \
  --build-arg GO_VERSION=1.23 \
  --build-arg DEBUG=true
```

## 🏭 Production Deployment

### Docker Compose (Recommended)

The production Docker Compose configuration includes:

- **Load-balanced API** with 3 replicas
- **High-performance Valkey cache** with persistence
- **Ollama AI service** with GPU support
- **Traefik reverse proxy** with SSL termination
- **Prometheus monitoring** with custom metrics
- **Grafana dashboards** with pre-built visualizations
- **Automated backup system** with encryption
- **Distributed tracing** with Jaeger

```bash
# Basic deployment
docker-compose -f docker-image/compose.production.optimized.yaml up -d

# With security overlay
docker-compose \
  -f docker-image/compose.production.optimized.yaml \
  -f docker-image/security/docker-security.yaml \
  up -d

# Specific profiles
docker-compose --profile monitoring up -d  # Monitoring only
docker-compose --profile production up -d  # Full production stack
```

### Docker Swarm (High Availability)

For multi-node deployments with automatic failover:

```bash
# Initialize swarm (on manager node)
docker swarm init

# Create networks
docker network create --driver overlay frontend
docker network create --driver overlay --internal backend

# Deploy stack
docker stack deploy \
  -c docker-image/compose.production.optimized.yaml \
  -c docker-image/security/docker-security.yaml \
  rendiff-probe

# Scale services
docker service scale rendiff-probe_api=5
```

### Kubernetes (Enterprise)

Kubernetes deployment with Helm charts:

```bash
# Create namespace
kubectl create namespace rendiff-probe

# Deploy with Helm
helm upgrade --install rendiff-probe ./helm/rendiff-probe \
  --namespace rendiff-probe \
  --set environment=production \
  --set domain=api.yourcompany.com
```

## 🔒 Security

### Security Features

- **Non-root containers** with minimal user privileges
- **Read-only filesystems** with tmpfs for writable areas
- **Seccomp profiles** to limit system calls
- **AppArmor/SELinux** support for additional isolation
- **Network encryption** with overlay networks
- **Secrets management** with automatic rotation
- **Security scanning** integrated into build process
- **Vulnerability monitoring** with Trivy integration

### Secrets Management

```bash
# Generate all secrets
./scripts/secrets-manager.sh generate

# Rotate specific secret
./scripts/secrets-manager.sh rotate jwt_secret

# Create encrypted backup
./scripts/secrets-manager.sh backup --backup-passphrase supersecret123

# List all secrets
./scripts/secrets-manager.sh list
```

### Security Policies

The `security/docker-security.yaml` overlay provides:

- **Capability dropping** (ALL capabilities dropped, minimal added back)
- **Resource limits** to prevent DoS attacks
- **Process limits** (max 100 processes per container)
- **Network policies** with encrypted communications
- **Audit logging** for compliance requirements

## 📊 Monitoring & Observability

### Monitoring Stack

- **Prometheus** for metrics collection
- **Grafana** for visualization and alerting
- **Jaeger** for distributed tracing
- **Custom metrics** for business logic
- **Health checks** with dependency monitoring

### Key Metrics Monitored

| Category | Metrics | Alerts |
|----------|---------|--------|
| Application | Request rate, response time, error rate | High latency, error spikes |
| Infrastructure | CPU, memory, disk, network | Resource exhaustion |
| Business | Video processing rate, quality checks | Processing failures |
| Security | Failed auth attempts, rate limits | Security incidents |

### Accessing Monitoring

```bash
# Grafana dashboard
https://grafana.yourdomain.com
# Default: admin / (see secrets/grafana_password.txt)

# Prometheus metrics
https://prometheus.yourdomain.com

# Jaeger tracing
https://jaeger.yourdomain.com
```

### Custom Dashboards

Pre-built Grafana dashboards include:
- **Application Overview** - Request metrics, errors, performance
- **Infrastructure Monitoring** - System resources, container health
- **Video Processing Analytics** - Processing rates, quality metrics
- **Security Dashboard** - Authentication, rate limiting, threats
- **Business Intelligence** - Usage patterns, capacity planning

## 🔄 Operations

### Deployment Management

```bash
# Health check
./deploy-production.sh --health-check

# Scale API service
./deploy-production.sh --scale 10

# View logs
./deploy-production.sh --logs api

# Create backup
./deploy-production.sh --backup

# Update deployment
./deploy-production.sh --update
```

### Backup & Recovery

Automated backup system includes:
- **Database backups** (SQLite, Valkey dumps)
- **Configuration backups** (secrets, configs)
- **Encrypted storage** with S3 compatible backends
- **Retention policies** with automatic cleanup

```bash
# Manual backup
./scripts/backup/create-backup.sh

# Restore from backup
./scripts/backup/restore-backup.sh backup-2024-01-15.tar.gz.gpg

# Verify backup integrity
./scripts/backup/verify-backup.sh backup-2024-01-15.tar.gz.gpg
```

### Log Management

Centralized logging with structured JSON logs:

```bash
# View application logs
docker-compose logs -f api

# View specific service logs
docker-compose logs -f prometheus

# Export logs for analysis
docker-compose logs --no-color api > api-logs-$(date +%Y%m%d).log
```

## 🔧 Configuration

### Environment Variables

Key production configuration variables:

```bash
# Core settings
ENVIRONMENT=production
DOMAIN=api.yourcompany.com
DATA_PATH=/opt/rendiff-probe/data

# Performance tuning
API_REPLICAS=3
WORKER_POOL_SIZE=16
MAX_CONCURRENT_JOBS=8
VALKEY_MAX_MEMORY=2gb

# Security
ENABLE_AUTH=true
ENABLE_RATE_LIMIT=true
ENABLE_CSRF=true
JWT_SECRET_FILE=/run/secrets/jwt_secret

# Monitoring
METRICS_ENABLED=true
TRACING_ENABLED=true
LOG_LEVEL=info

# AI/LLM
OLLAMA_MODEL=gemma3:270m
OLLAMA_FALLBACK_MODEL=phi3:mini
OLLAMA_PARALLEL=4

# Backup
BACKUP_S3_BUCKET=your-backup-bucket
BACKUP_RETENTION_DAYS=30
```

### SSL/TLS Configuration

Automatic SSL certificate management with Let's Encrypt:

```bash
# Enable SSL with Let's Encrypt
export SSL_EMAIL=admin@yourcompany.com
export DOMAIN=api.yourcompany.com

# Deploy with SSL
./deploy-production.sh --enable-ssl --deploy
```

### Custom Configuration

Override default configurations:

```bash
# Custom Prometheus config
cp config/prometheus/prometheus.yml config/prometheus/prometheus.custom.yml
# Edit config/prometheus/prometheus.custom.yml

# Custom Grafana dashboards
cp -r config/grafana/dashboards config/grafana/dashboards.custom
# Add custom dashboards to config/grafana/dashboards.custom/
```

## 🚨 Troubleshooting

### Common Issues

#### Container Won't Start
```bash
# Check container logs
docker-compose logs api

# Check health status
docker-compose ps

# Verify secrets exist
./scripts/secrets-manager.sh verify
```

#### SSL Certificate Issues
```bash
# Check Traefik logs
docker-compose logs traefik

# Verify domain DNS
nslookup api.yourcompany.com

# Check certificate status
curl -I https://api.yourcompany.com
```

#### Performance Issues
```bash
# Check resource usage
docker stats

# Monitor metrics
curl http://localhost:9090/metrics

# Check application logs
docker-compose logs api | grep -i error
```

#### Database Issues
```bash
# Check SQLite database
sqlite3 /opt/rendiff-probe/data/app/rendiff-probe.db ".tables"

# Check Valkey cache
docker-compose exec valkey valkey-cli ping

# Monitor cache performance
docker-compose exec valkey valkey-cli info stats
```

### Performance Tuning

#### API Service Optimization
```bash
# Increase worker pool
export WORKER_POOL_SIZE=32
export MAX_CONCURRENT_JOBS=16

# Scale API replicas
./deploy-production.sh --scale 5
```

#### Cache Optimization
```bash
# Increase Valkey memory
export VALKEY_MAX_MEMORY=4gb

# Tune cache policies
docker-compose exec valkey valkey-cli config set maxmemory-policy allkeys-lru
```

#### AI Service Optimization
```bash
# Enable GPU support (if available)
export OLLAMA_GPU_ENABLED=true

# Increase parallel processing
export OLLAMA_PARALLEL=8

# Optimize model loading
export OLLAMA_KEEP_ALIVE=30m
```

## 📈 Scaling

### Horizontal Scaling

```bash
# Scale API service
docker service scale rendiff-probe_api=10

# Scale with Docker Compose
docker-compose up -d --scale api=5

# Auto-scaling with resource thresholds
# (Configure in monitoring/alerting system)
```

### Vertical Scaling

```bash
# Increase resource limits
# Edit compose.production.optimized.yaml
deploy:
  resources:
    limits:
      cpus: '4.0'
      memory: 4G
    reservations:
      cpus: '2.0'
      memory: 2G
```

### Database Scaling

For larger deployments, consider:
- **PostgreSQL cluster** instead of SQLite
- **Valkey cluster** for distributed caching
- **Read replicas** for improved performance

## 🔍 Maintenance

### Regular Maintenance Tasks

```bash
# Update container images
./deploy-production.sh --update

# Rotate secrets (monthly)
./scripts/secrets-manager.sh rotate

# Clean old backups
./scripts/secrets-manager.sh clean

# Security scan
./build-optimized.sh --target security-scan --scan

# Update dependencies
docker-compose pull
docker-compose up -d
```

### Health Monitoring

Automated health checks monitor:
- **Service availability** (HTTP endpoints)
- **Database connectivity** (SQLite, Valkey)
- **AI service readiness** (Ollama models)
- **Certificate expiration** (SSL/TLS)
- **Disk space usage** (Storage volumes)
- **Resource utilization** (CPU, memory)

## 📚 Additional Resources

- **[Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)**
- **[Production Security Guide](./SECURITY.md)**
- **[Monitoring Runbook](./MONITORING.md)**
- **[Disaster Recovery Plan](./DISASTER-RECOVERY.md)**
- **[API Documentation](../README.md)**

## 🤝 Support

For production deployment support:
- **Issues**: [GitHub Issues](https://github.com/yourorg/rendiff-probe/issues)
- **Documentation**: [Wiki](https://github.com/yourorg/rendiff-probe/wiki)
- **Security**: security@yourcompany.com

---

**Production Infrastructure v2.0** - Deployed with ❤️ for enterprise-grade video processing