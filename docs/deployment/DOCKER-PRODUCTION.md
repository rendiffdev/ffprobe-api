# FFprobe API - Production Docker Deployment Guide

## 🏭 Enterprise-Grade Docker Infrastructure

This guide covers the **new production-grade Docker infrastructure** that transforms FFprobe API into an enterprise-ready video processing platform with comprehensive security, monitoring, and operational capabilities.

## 🚀 Quick Start - Production Deployment

### Prerequisites

- **Docker** 24.0+ with Compose
- **8GB RAM** minimum (16GB recommended for full stack)
- **20GB disk space** for production deployment
- **Valid domain name** for SSL certificates
- **Basic understanding** of Docker and container orchestration

### 1. Clone and Prepare

```bash
git clone <your-repo-url>
cd rendiff-probe
```

### 2. Build Production Image

```bash
# Build optimized production image with security scanning
./docker-image/build-optimized.sh \
  --target production \
  --scan \
  --sbom \
  --multi-arch
```

### 3. Generate Secrets

```bash
# Generate all production secrets securely
./docker-image/scripts/secrets-manager.sh generate

# Verify secrets are created
./docker-image/scripts/secrets-manager.sh verify
```

### 4. Configure Environment

```bash
# Set your domain and environment
export DOMAIN=api.yourcompany.com
export SSL_EMAIL=admin@yourcompany.com
export ENVIRONMENT=production
export DATA_PATH=/opt/rendiff-probe/data
```

### 5. Deploy Production Stack

```bash
# Complete production deployment with SSL and monitoring
./docker-image/deploy-production.sh \
  --mode compose \
  --environment production \
  --domain $DOMAIN \
  --enable-ssl \
  --enable-monitoring \
  --enable-backup \
  --deploy
```

### 6. Verify Deployment

```bash
# Check health of all services
./docker-image/deploy-production.sh --health-check

# View service status
docker-compose -f docker-image/compose.production.optimized.yaml ps
```

## 🎯 Production Features

### **🛡️ Security Hardening**
- **Non-root containers** with minimal privileges
- **Read-only filesystems** with tmpfs for writable areas
- **Seccomp profiles** limiting system calls
- **Network encryption** between services
- **Automated secrets management** with rotation
- **Vulnerability scanning** integrated into builds

### **📊 Comprehensive Monitoring**
- **Prometheus** metrics collection with custom rules
- **Grafana** dashboards with FFprobe-specific visualizations
- **Jaeger** distributed tracing for request tracking
- **Automated alerting** for critical issues
- **Health checks** with dependency monitoring

### **⚡ Performance Optimization**
- **Multi-stage builds** reducing image size by 60%
- **Layer caching** for faster builds and deployments
- **Resource-optimized** container configurations
- **Load balancing** with automatic scaling
- **Connection pooling** and cache optimization

### **🔄 Operational Excellence**
- **Zero-downtime deployments** with rolling updates
- **Automated backups** with encryption and retention
- **Log aggregation** with structured JSON logging
- **Service discovery** and health monitoring
- **Disaster recovery** procedures and automation

## 🏗️ Architecture Overview

### Production Stack Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Production Stack                          │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Traefik   │  │ Prometheus  │  │   Grafana   │        │
│  │ (Reverse    │  │ (Metrics)   │  │(Dashboards) │        │
│  │  Proxy)     │  │             │  │             │        │
│  └─────┬───────┘  └─────────────┘  └─────────────┘        │
│        │                                                   │
│        ▼                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ FFprobe API │  │   Valkey    │  │   Ollama    │        │
│  │ (3 replicas)│  │  (Cache)    │  │ (AI Models) │        │
│  │             │  │             │  │             │        │
│  └─────┬───────┘  └─────────────┘  └─────────────┘        │
│        │                                                   │
│        ▼                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   SQLite    │  │   Jaeger    │  │   Backup    │        │
│  │ (Database)  │  │ (Tracing)   │  │  Service    │        │
│  │             │  │             │  │             │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### Service Endpoints

| Service | Internal Port | External URL | Purpose |
|---------|---------------|--------------|---------|
| **FFprobe API** | 8080 | https://api.domain.com | Main API service |
| **Prometheus** | 9090 | https://prometheus.domain.com | Metrics collection |
| **Grafana** | 3000 | https://grafana.domain.com | Dashboards |
| **Jaeger** | 16686 | https://jaeger.domain.com | Distributed tracing |
| **Traefik** | 80/443 | - | Reverse proxy & SSL |

## 🔧 Deployment Modes

### Docker Compose (Recommended)

Best for single-node deployments with full production features:

```bash
# Basic production deployment
docker-compose -f docker-image/compose.production.optimized.yaml up -d

# With security hardening
docker-compose \
  -f docker-image/compose.production.optimized.yaml \
  -f docker-image/security/docker-security.yaml \
  up -d

# Monitoring stack only
docker-compose --profile monitoring up -d
```

### Docker Swarm (High Availability)

For multi-node clusters with automatic failover:

```bash
# Initialize swarm
docker swarm init

# Create overlay networks
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

Enterprise container orchestration with Helm:

```bash
# Create namespace
kubectl create namespace rendiff-probe

# Deploy with Helm (requires helm charts)
helm upgrade --install rendiff-probe ./helm/rendiff-probe \
  --namespace rendiff-probe \
  --set environment=production \
  --set domain=api.yourcompany.com
```

## 🔒 Security Configuration

### Security Features

The production Docker infrastructure includes comprehensive security measures:

#### **Container Security**
- **Non-root execution** with dedicated service users
- **Read-only root filesystems** preventing runtime modifications
- **Minimal base images** reducing attack surface
- **Capability dropping** removing unnecessary privileges
- **Process limits** preventing resource exhaustion

#### **Network Security**
- **Encrypted overlay networks** for service communication
- **Service isolation** with internal networks
- **TLS termination** at the edge with automatic certificates
- **Rate limiting** and DDoS protection
- **CORS and security headers** for web security

#### **Secrets Management**
```bash
# Generate all secrets
./docker-image/scripts/secrets-manager.sh generate

# Rotate specific secret
./docker-image/scripts/secrets-manager.sh rotate jwt_secret

# Create encrypted backup
./docker-image/scripts/secrets-manager.sh backup --backup-passphrase supersecret123

# List managed secrets
./docker-image/scripts/secrets-manager.sh list
```

#### **Compliance Features**
- **Audit logging** for compliance requirements
- **Security scanning** integrated into CI/CD
- **Vulnerability monitoring** with automated updates
- **Policy enforcement** with admission controllers
- **Compliance frameworks** support (SOC2, PCI-DSS, GDPR)

## 📊 Monitoring & Alerting

### Monitoring Stack

The production deployment includes a comprehensive monitoring solution:

#### **Prometheus Metrics**
- Application performance metrics
- Infrastructure resource usage
- Business logic metrics
- Custom recording rules for efficiency

#### **Grafana Dashboards**
Pre-built dashboards for:
- **Application Overview** - Request rates, errors, performance
- **Infrastructure Monitoring** - CPU, memory, disk, network
- **Video Processing Analytics** - Processing rates, quality metrics
- **Security Dashboard** - Authentication, rate limits, threats

#### **Jaeger Tracing**
- Distributed request tracing
- Performance bottleneck identification
- Service dependency mapping
- Error propagation tracking

#### **Automated Alerting**
```yaml
# Example alert rules
- alert: FFprobeAPIDown
  expr: up{job="rendiff-probe"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "FFprobe API instance is down"

- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High error rate detected"
```

### Accessing Monitoring

```bash
# Grafana dashboard (admin/password from secrets)
https://grafana.yourdomain.com

# Prometheus metrics
https://prometheus.yourdomain.com

# Jaeger tracing
https://jaeger.yourdomain.com

# API health check
curl https://api.yourdomain.com/health
```

## 🔄 Operations & Maintenance

### Service Management

```bash
# Scale API service
./docker-image/deploy-production.sh --scale 10

# View logs
./docker-image/deploy-production.sh --logs api

# Health check all services
./docker-image/deploy-production.sh --health-check

# Update deployment
./docker-image/deploy-production.sh --update
```

### Backup & Recovery

#### **Automated Backups**
```bash
# Backup schedule runs automatically
# Default: Daily at 2 AM UTC

# Manual backup
./docker-image/scripts/backup/create-backup.sh

# List backups
ls -la /opt/rendiff-probe/data/backups/
```

#### **Backup Restoration**
```bash
# Restore from backup
./docker-image/scripts/backup/restore-backup.sh backup-2024-01-15.tar.gz.gpg

# Verify backup integrity
./docker-image/scripts/backup/verify-backup.sh backup-2024-01-15.tar.gz.gpg
```

### Log Management

```bash
# View application logs
docker-compose logs -f api

# View all services logs
docker-compose logs -f

# Export logs for analysis
docker-compose logs --no-color > production-logs-$(date +%Y%m%d).log

# Monitor logs in real-time
docker-compose logs -f | grep ERROR
```

### Updates & Maintenance

```bash
# Update container images
docker-compose pull
docker-compose up -d

# Rotate secrets (monthly)
./docker-image/scripts/secrets-manager.sh rotate

# Security scan
./docker-image/build-optimized.sh --target security-scan --scan

# Clean old backups
./docker-image/scripts/secrets-manager.sh clean
```

## 📈 Scaling & Performance

### Horizontal Scaling

```bash
# Scale API service
docker service scale rendiff-probe_api=10

# Scale with Docker Compose
docker-compose up -d --scale api=5

# Auto-scaling configuration in compose file
deploy:
  mode: replicated
  replicas: 3
  resources:
    limits:
      cpus: '2.0'
      memory: 2G
```

### Performance Tuning

#### **API Service Optimization**
```bash
# Environment variables for performance
export WORKER_POOL_SIZE=32
export MAX_CONCURRENT_JOBS=16
export REQUEST_TIMEOUT=30s
```

#### **Cache Optimization**
```bash
# Valkey performance tuning
export VALKEY_MAX_MEMORY=4gb
export VALKEY_POOL_SIZE=20
```

#### **Resource Allocation**
```yaml
# In compose file
deploy:
  resources:
    limits:
      cpus: '4.0'
      memory: 4G
    reservations:
      cpus: '2.0'
      memory: 2G
```

## 🚨 Troubleshooting

### Common Issues

#### **Container Won't Start**
```bash
# Check container logs
docker-compose logs api

# Check resource usage
docker stats

# Verify secrets exist
./docker-image/scripts/secrets-manager.sh verify

# Check disk space
df -h /opt/rendiff-probe/data
```

#### **SSL Certificate Issues**
```bash
# Check Traefik logs
docker-compose logs traefik

# Verify domain DNS
nslookup api.yourcompany.com

# Check certificate status
curl -I https://api.yourcompany.com

# Manual certificate generation
openssl s_client -connect api.yourcompany.com:443 -servername api.yourcompany.com
```

#### **Performance Issues**
```bash
# Monitor metrics
curl http://localhost:9090/metrics | grep ffprobe

# Check database performance
sqlite3 /opt/rendiff-probe/data/app/rendiff-probe.db ".schema"

# Analyze slow queries
docker-compose logs api | grep "slow query"

# Resource utilization
docker stats --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
```

#### **Network Issues**
```bash
# Check network connectivity
docker network ls
docker network inspect frontend

# Test service communication
docker-compose exec api curl http://valkey:6379

# Port availability
lsof -i :8080
netstat -tlnp | grep :8080
```

### Health Monitoring

The production deployment includes comprehensive health checks:

```bash
# Service health endpoints
curl http://localhost:8080/health          # API health
curl http://localhost:9090/-/ready         # Prometheus
curl http://localhost:3000/api/health      # Grafana
curl http://localhost:11434/api/version    # Ollama

# Database connectivity
docker-compose exec valkey valkey-cli ping
sqlite3 /opt/rendiff-probe/data/app/rendiff-probe.db ".tables"

# Resource monitoring
docker system df
docker system events --since="1h"
```

## 🌐 Production Checklist

Before going live, ensure:

### **Infrastructure**
- [ ] Docker and Docker Compose installed
- [ ] Sufficient resources allocated (8GB+ RAM, 20GB+ disk)
- [ ] Valid domain name configured
- [ ] DNS records pointing to your server
- [ ] Firewall configured (ports 80, 443, 22)

### **Security**
- [ ] All secrets generated and secured
- [ ] SSL certificates working
- [ ] Security scanning passed
- [ ] Firewall rules configured
- [ ] Regular security updates scheduled

### **Monitoring**
- [ ] Prometheus collecting metrics
- [ ] Grafana dashboards accessible
- [ ] Alerting rules configured
- [ ] Log aggregation working
- [ ] Health checks responding

### **Backup & Recovery**
- [ ] Automated backups configured
- [ ] Backup encryption working
- [ ] Recovery procedures tested
- [ ] Retention policies set
- [ ] Off-site backup storage configured

### **Performance**
- [ ] Load testing completed
- [ ] Resource limits configured
- [ ] Scaling policies defined
- [ ] Performance baselines established
- [ ] Monitoring thresholds set

## 📞 Support & Resources

### Documentation
- **[Main README](../../README.md)** - Project overview
- **[API Documentation](../api/README.md)** - REST and GraphQL APIs
- **[Security Guide](./SECURITY.md)** - Security best practices
- **[Disaster Recovery](./DISASTER-RECOVERY.md)** - Recovery procedures

### Community
- **Issues**: [GitHub Issues](https://github.com/yourorg/rendiff-probe/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourorg/rendiff-probe/discussions)
- **Wiki**: [Project Wiki](https://github.com/yourorg/rendiff-probe/wiki)

### Professional Support
For enterprise support and consulting:
- **Email**: support@yourcompany.com
- **Security**: security@yourcompany.com
- **Professional Services**: Available for custom deployments and integrations

---

**Production Docker Infrastructure v2.0** - Built for enterprise-scale video processing with security, monitoring, and operational excellence.