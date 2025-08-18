# Modern Docker Compose Deployment Guide

This guide covers the latest Docker Compose features and best practices implemented in FFprobe API deployment.

## Overview

FFprobe API uses the latest [Docker Compose Application Model](https://docs.docker.com/compose/intro/compose-application-model/) with modern features including:

- **Compose Specification 3.8+** with latest features
- **Service Profiles** for different deployment modes
- **Modern Health Checks** with detailed configuration
- **Advanced Networking** with custom IPAM
- **Resource Management** with limits and reservations
- **Security Hardening** with least privilege principles
- **Extensions and Includes** for modular configuration

## Deployment Modes Using Profiles

### Minimal Profile (NEW - Ultra-Lightweight)
```bash
# Absolute minimal setup - 4 core services only
docker compose --profile minimal up -d

# Or using Make
make minimal
```

**Services**: API, PostgreSQL, Redis, Ollama (Gemma 3 270M only)  
**Features**: Essential services only, ~2-3GB RAM, fastest startup  
**Use Case**: Development, testing, resource-constrained environments

### Quick Start Profile
```bash
# Quick setup for immediate use
docker compose --profile quick up -d

# Or using Make
make quick
```

**Services**: API, PostgreSQL, Redis, Ollama (Gemma 3 270M only)  
**Features**: No authentication, development settings, fast startup  
**Use Case**: Testing, demos, quick prototyping

### Development Profile
```bash
# Full development environment with tools
docker compose -f compose.yaml -f compose.development.yaml --profile development up -d

# Or using Make
make dev
```

**Services**: All quick services + development tools  
**Additional Tools**:
- **File Browser** (http://localhost:8083) - File management
- **Adminer** (http://localhost:8081) - Database admin
- **Redis Commander** (http://localhost:8082) - Redis admin
- **File Browser** (http://localhost:8083) - File management

**Features**: Hot reload, debug ports, verbose logging, mounted source code

### Production Profile
```bash
# Production deployment with monitoring and security
docker compose -f compose.yaml -f compose.production.yaml --profile production up -d

# Or using Make
make prod
```

**Services**: All core services + production infrastructure  
**Additional Components**:
- **Traefik** - Combined reverse proxy + automatic SSL (replaces Nginx)
- **Prometheus** - Metrics collection
- **Grafana** - Monitoring dashboards
- **Backup Service** - Automated backups

## Modern Compose Features Used

### 1. Service Profiles
```yaml
services:
  api:
    # ... service definition
    profiles:
      - api
      - quick
      - full
```

**Benefits**:
- Selective service deployment
- Environment-specific configurations
- Reduced resource usage
- Faster startup times

### 2. Advanced Health Checks
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 60s
  start_interval: 5s  # New feature!
```

**New Features**:
- `start_interval`: Faster checks during startup
- Enhanced dependency management
- Graceful failure handling

### 3. Modern Dependency Management
```yaml
depends_on:
  postgres:
    condition: service_healthy
    restart: true  # Auto-restart on dependency failure
  redis:
    condition: service_healthy
    restart: true
```

**Benefits**:
- Automatic service recovery
- Proper startup sequencing
- Cascading health checks

### 4. Resource Management
```yaml
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '0.5'
      memory: 512M
  restart_policy:
    condition: on-failure
    delay: 5s
    max_attempts: 3
    window: 120s
```

**Features**:
- CPU and memory limits
- Resource reservations
- Advanced restart policies
- Rolling update configuration

### 5. Modern Volume Configuration
```yaml
volumes:
  postgres_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: ${DATA_PATH:-./data}/postgres
    labels:
      - "com.ffprobe-api.volume=database"
      - "com.ffprobe-api.backup=daily"
```

**Benefits**:
- Flexible storage backends
- Metadata and labeling
- Backup scheduling hints
- Path customization

### 6. Advanced Networking
```yaml
networks:
  ffprobe-network:
    driver: bridge
    enable_ipv6: false
    ipam:
      driver: default
      config:
        - subnet: 172.20.0.0/16
          gateway: 172.20.0.1
          ip_range: 172.20.240.0/20
    driver_opts:
      com.docker.network.bridge.name: ffprobe-br0
      com.docker.network.driver.mtu: 1500
```

**Features**:
- Custom IP address management
- Network segmentation
- MTU optimization
- Bridge naming

### 7. Extensions and Reusability
```yaml
# Reusable configurations
x-common-variables: &common-variables
  TZ: ${TZ:-UTC}
  LANG: en_US.UTF-8

x-restart-policy: &restart-policy
  restart: unless-stopped

x-logging: &default-logging
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"
```

**Benefits**:
- DRY principle
- Consistent configurations
- Easy maintenance
- Template reuse

### 8. Include Directive
```yaml
include:
  - path: compose.override.yaml
    required: false
  - path: compose.${GO_ENV:-development}.yaml
    required: false
  - path: compose.local.yaml
    required: false
```

**Benefits**:
- Modular configuration
- Environment-specific overrides
- Local customizations
- Optional includes

## Environment-Specific Configurations

### Quick Start Environment Variables
```bash
# Minimal configuration for quick start
GO_ENV=development
POSTGRES_PASSWORD=quickstart123
REDIS_PASSWORD=quickstart123
OLLAMA_MODEL=gemma3:270m
```

### Development Environment Variables
```bash
# Development-specific settings
GO_ENV=development
DEV_ENABLE_DEBUG=true
DEV_DISABLE_AUTH=true
DEV_OLLAMA_MODEL=gemma3:270m
LOG_LEVEL=debug
```

### Production Environment Variables
```bash
# Production configuration (use .env file)
GO_ENV=production
DOMAIN=your-domain.com
ACME_EMAIL=admin@your-domain.com
POSTGRES_PASSWORD=secure_random_password
REDIS_PASSWORD=secure_random_password
GRAFANA_PASSWORD=secure_random_password
BACKUP_ENCRYPTION_KEY=encryption_key
```

## Service Management Commands

### Basic Operations
```bash
# Start all services
docker compose up -d

# Start specific profile
docker compose --profile quick up -d
docker compose --profile development up -d
docker compose --profile production up -d

# Stop services
docker compose stop

# Remove everything
docker compose down -v
```

### Profile-Specific Management
```bash
# Start only database services
docker compose --profile database up -d

# Start API without AI services
docker compose --profile api --profile database --profile cache up -d

# Start monitoring stack only
docker compose -f compose.yaml -f compose.production.yaml --profile monitoring up -d
```

### Health and Status
```bash
# Check service health
docker compose ps

# View logs
docker compose logs -f

# Service-specific logs
docker compose logs -f api
docker compose logs -f postgres
```

### Development Workflow
```bash
# Start development environment
make dev

# Access development tools
open http://localhost:8081  # Adminer (DB admin)
open http://localhost:8082  # Redis Commander
open http://localhost:8083  # File Browser
open http://localhost:8083  # File Browser

# View real-time logs
make logs-api
```

### Production Workflow
```bash
# Deploy production environment
make prod

# Access production services
open https://grafana.your-domain.com  # Grafana dashboards
open https://prometheus.your-domain.com  # Metrics
open https://api.your-domain.com  # API endpoint

# Monitor health
make health

# Create backup
make backup
```

## Performance Optimizations

### Container Optimizations
```yaml
# Tmpfs for temporary files
tmpfs:
  - /tmp:size=100M,noexec,nosuid,nodev

# Init system for proper signal handling
init: true

# Resource limits and reservations
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 2G
    reservations:
      cpus: '0.5'
      memory: 512M
```

### Network Optimizations
```yaml
# Custom MTU for better performance
driver_opts:
  com.docker.network.driver.mtu: 1500

# Network segmentation for security
networks:
  frontend:
    # External-facing services
  backend:
    # Internal services only
```

### Storage Optimizations
```yaml
# SSD-optimized volume
volumes:
  fast_storage:
    driver: local
    driver_opts:
      type: none
      o: bind,rw
      device: /mnt/ssd/data
```

## Security Best Practices

### Container Security
```yaml
# Security hardening
security_opt:
  - no-new-privileges:true
cap_drop:
  - ALL
cap_add:
  - CHOWN
  - SETUID
  - SETGID
```

### Network Security
```yaml
# Internal networks only
networks:
  internal:
    internal: true  # No external access

# Custom firewall rules
driver_opts:
  com.docker.network.bridge.enable_icc: "false"
```

### Secrets Management
```yaml
# Use Docker secrets for sensitive data
secrets:
  postgres_password:
    external: true
  redis_password:
    external: true

services:
  postgres:
    secrets:
      - postgres_password
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
```

## Monitoring and Observability

### Health Check Monitoring
```yaml
# Comprehensive health checks
healthcheck:
  test: ["CMD", "./health-check.sh"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 60s
  start_interval: 5s
```

### Logging Configuration
```yaml
# Structured logging
logging:
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"
    labels: "service,version,environment"
    tag: "{{.ImageName}}|{{.Name}}"
```

### Metrics Collection
```yaml
# Prometheus metrics
labels:
  - "prometheus.scrape=true"
  - "prometheus.port=9090"
  - "prometheus.path=/metrics"
```

## Troubleshooting

### Common Issues

#### Services won't start
```bash
# Check Docker Compose version
docker compose version

# Validate configuration
docker compose config

# Check service logs
docker compose logs service-name
```

#### Profile not found
```bash
# List available profiles
docker compose config --profiles

# Verify profile syntax
docker compose -f compose.yaml config --profiles
```

#### Health checks failing
```bash
# Check health status
docker compose ps

# Run health check manually
docker compose exec service-name curl -f http://localhost:8080/health
```

### Best Practices for Debugging

1. **Use `docker compose config`** to validate syntax
2. **Check service dependencies** with `depends_on`
3. **Monitor resource usage** with `docker stats`
4. **Use structured logging** for better debugging
5. **Implement proper health checks** for all services

## Migration from Older Compose Versions

### From docker-compose to docker compose
```bash
# Old command
docker-compose up -d

# New command
docker compose up -d
```

### From version 2.x to 3.8+
```yaml
# Old syntax
version: '2.4'

# New syntax (no version needed)
# version specification is deprecated
```

### From simple to profile-based deployment
```bash
# Old monolithic approach
docker compose up -d

# New profile-based approach
docker compose --profile quick up -d
docker compose --profile development up -d
docker compose --profile production up -d
```

This modern Docker Compose setup provides a robust, scalable, and maintainable deployment solution for FFprobe API across all environments.