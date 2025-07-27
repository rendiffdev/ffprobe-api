# ⚙️ Configuration Guide

Complete configuration reference for FFprobe API deployment.

## 🌍 Environment Variables

### Core Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `HOST` | `0.0.0.0` | HTTP server host |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `ENVIRONMENT` | `development` | Environment (development, staging, production) |

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_DB` | `ffprobe_api` | Database name |
| `POSTGRES_USER` | `ffprobe` | Database username |
| `POSTGRES_PASSWORD` | - | Database password (required) |
| `DB_MAX_OPEN_CONNS` | `25` | Maximum open connections |
| `DB_MAX_IDLE_CONNS` | `10` | Maximum idle connections |

### Redis Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | - | Redis password |
| `REDIS_DB` | `0` | Redis database number |

### Authentication & Security

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_AUTH` | `true` | Enable authentication |
| `API_KEY` | - | Master API key (required, min 32 chars) |
| `JWT_SECRET` | - | JWT signing secret (required, min 32 chars) |
| `TOKEN_EXPIRY_HOURS` | `24` | JWT token expiry time |
| `REFRESH_EXPIRY_HOURS` | `168` | Refresh token expiry time |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_RATE_LIMIT` | `true` | Enable rate limiting |
| `RATE_LIMIT_PER_MINUTE` | `60` | Requests per minute |
| `RATE_LIMIT_PER_HOUR` | `1000` | Requests per hour |
| `RATE_LIMIT_PER_DAY` | `10000` | Requests per day |

### File Handling

| Variable | Default | Description |
|----------|---------|-------------|
| `UPLOAD_DIR` | `/app/uploads` | Upload directory path |
| `REPORTS_DIR` | `/app/reports` | Reports directory path |
| `MAX_FILE_SIZE` | `53687091200` | Max file size (50GB) |
| `TEMP_DIR` | `/tmp` | Temporary files directory |

### FFmpeg Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `FFMPEG_PATH` | `/usr/local/bin/ffmpeg` | FFmpeg binary path |
| `FFPROBE_PATH` | `/usr/local/bin/ffprobe` | FFprobe binary path |
| `VMAF_MODEL_PATH` | `/usr/local/share/vmaf` | VMAF models directory |
| `MAX_CONCURRENT_JOBS` | `4` | Maximum concurrent processing |

## 📁 Directory Structure

### Data Directories
```
data/
├── postgres/          # PostgreSQL data
├── redis/             # Redis data
├── uploads/           # Uploaded files
├── reports/           # Generated reports
├── models/            # VMAF models
├── logs/              # Application logs
├── temp/              # Temporary files
├── cache/             # Cache data
├── backup/            # Backup files
├── prometheus/        # Prometheus data
└── grafana/           # Grafana data
```

### Configuration Files
```
docker/
├── nginx.conf         # Nginx configuration
├── prometheus.yml     # Prometheus configuration
├── grafana/
│   ├── dashboards/    # Grafana dashboards
│   └── provisioning/  # Grafana provisioning
└── init.sql           # Database initialization
```

## 🐳 Docker Configuration

### Environment Files

#### `.env` (Base Configuration)
```bash
# Database
POSTGRES_PASSWORD=your_secure_db_password_here
POSTGRES_USER=ffprobe
POSTGRES_DB=ffprobe_api

# Redis
REDIS_PASSWORD=your_secure_redis_password_here

# Authentication
API_KEY=your_secure_api_key_minimum_32_characters
JWT_SECRET=your_secure_jwt_secret_minimum_32_characters

# Grafana
GRAFANA_USER=admin
GRAFANA_PASSWORD=your_secure_grafana_password

# Data Path
DATA_PATH=./data
```

#### `.env.development`
```bash
# Development overrides
LOG_LEVEL=debug
ENVIRONMENT=development
ENABLE_RATE_LIMIT=false
RATE_LIMIT_PER_MINUTE=1000

# Development passwords (change these!)
POSTGRES_PASSWORD=dev_password_change_this
REDIS_PASSWORD=dev_redis_pass
API_KEY=dev_api_key_change_this_minimum_32_chars
JWT_SECRET=dev_jwt_secret_change_this_minimum_32_chars
GRAFANA_PASSWORD=admin_change_this
```

#### `.env.production`
```bash
# Production overrides
LOG_LEVEL=info
ENVIRONMENT=production
ENABLE_RATE_LIMIT=true

# Use secrets management for these in production
# POSTGRES_PASSWORD=read_from_secrets_manager
# REDIS_PASSWORD=read_from_secrets_manager
# API_KEY=read_from_secrets_manager
# JWT_SECRET=read_from_secrets_manager
```

### Docker Compose Overrides

#### Development (`compose.dev.yml`)
```yaml
services:
  ffprobe-api:
    volumes:
      - .:/app
    environment:
      - LOG_LEVEL=debug
    
  adminer:
    image: adminer:4.8.1
    ports:
      - "8090:8080"
    environment:
      - ADMINER_DEFAULT_SERVER=postgres
```

#### Production (`compose.prod.yml`)
```yaml
services:
  ffprobe-api:
    deploy:
      replicas: 3
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
    
  nginx:
    image: nginx:1.25.3-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./docker/nginx.conf:/etc/nginx/nginx.conf:ro
```

## 🔒 Security Configuration

### SSL/TLS Setup

```bash
# Generate certificates (for development)
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Production: Use Let's Encrypt or your CA
certbot certonly --webroot -w /var/www/html -d yourdomain.com
```

### Nginx Configuration (`docker/nginx.conf`)
```nginx
events {
    worker_connections 1024;
}

http {
    upstream ffprobe_api {
        server ffprobe-api:8080;
    }
    
    server {
        listen 80;
        server_name yourdomain.com;
        return 301 https://$server_name$request_uri;
    }
    
    server {
        listen 443 ssl http2;
        server_name yourdomain.com;
        
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        
        location / {
            proxy_pass http://ffprobe_api;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

## 📊 Monitoring Configuration

### Prometheus Configuration (`docker/prometheus.yml`)
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'ffprobe-api'
    static_configs:
      - targets: ['ffprobe-api:8080']
    metrics_path: /metrics
    scrape_interval: 30s
    
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:5432']
      
  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']
```

### Grafana Dashboards

Dashboard configuration is automatically provisioned in `docker/grafana/`.

## 🚀 Deployment Configurations

### Interactive Installation (Recommended)
```bash
# Run the interactive installer
make install
# OR directly: ./scripts/setup/install.sh
```

### Quick Setup (3 modes)
```bash
# Quick setup script
make quick-setup
# OR directly: ./scripts/setup/quick-setup.sh

# Choose from:
# 1. 🔧 Development - Local development with debugging
# 2. 🧪 Demo - Basic auth, sample configuration
# 3. 🏭 Production - Full security, SSL ready
```

### Manual Configuration
```bash
# Copy and customize configuration
cp .env.example .env
# Edit .env with your preferences

# Validate configuration
make validate
# OR directly: ./scripts/setup/validate-config.sh

# Development deployment
docker compose -f compose.yml -f compose.dev.yml up

# Production deployment
docker compose -f compose.yml -f compose.prod.yml up -d
```

### Kubernetes Configuration

#### ConfigMap (`k8s/configmap.yaml`)
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ffprobe-config
data:
  LOG_LEVEL: "info"
  ENVIRONMENT: "production"
  POSTGRES_HOST: "postgres-service"
  REDIS_HOST: "redis-service"
```

#### Secret (`k8s/secret.yaml`)
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ffprobe-secrets
type: Opaque
stringData:
  POSTGRES_PASSWORD: "your-secure-password"
  API_KEY: "your-secure-api-key"
  JWT_SECRET: "your-secure-jwt-secret"
```

## 🔧 Advanced Configuration

### Performance Tuning

```bash
# Database connections
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=20
DB_CONN_MAX_LIFETIME=5m

# Processing limits
MAX_CONCURRENT_JOBS=8
WORKER_POOL_SIZE=16
PROCESSING_TIMEOUT=300s

# Memory limits
UPLOAD_MEMORY_LIMIT=8GB
PROCESSING_MEMORY_LIMIT=16GB
```

### Logging Configuration

```bash
# Structured logging
LOG_FORMAT=json
LOG_OUTPUT=stdout
LOG_ROTATION=true
LOG_MAX_SIZE=100MB
LOG_MAX_AGE=30
LOG_MAX_BACKUPS=10

# Request logging
LOG_REQUESTS=true
LOG_REQUEST_BODY=false
LOG_RESPONSE_BODY=false
```

### Caching Configuration

```bash
# Redis caching
CACHE_TTL=3600
CACHE_NAMESPACE=ffprobe:
CACHE_COMPRESSION=true

# Memory cache
MEMORY_CACHE_SIZE=1GB
MEMORY_CACHE_TTL=300s
```

## 🧪 Testing Configuration

### Test Environment
```bash
# Test database
TEST_POSTGRES_HOST=localhost
TEST_POSTGRES_PORT=5433
TEST_POSTGRES_DB=ffprobe_test
TEST_POSTGRES_USER=test
TEST_POSTGRES_PASSWORD=test

# Disable external dependencies
ENABLE_AUTH=false
ENABLE_RATE_LIMIT=false
```

### Load Testing
```bash
# Performance test configuration
LOAD_TEST_CONCURRENT_USERS=100
LOAD_TEST_DURATION=300s
LOAD_TEST_RAMP_UP=30s
```

## 📋 Configuration Validation

### Automated Validation

The repository includes comprehensive validation tools:

```bash
# Run all validation checks
make prod-ready

# Individual validation steps
make validate          # Configuration validation
make docker-update     # Docker Compose syntax check
make security          # Security scan
make test-all          # Complete test suite
```

### Pre-deployment Checklist

- [ ] All required environment variables set
- [ ] Database connection tested
- [ ] Redis connection tested
- [ ] SSL certificates valid
- [ ] File permissions correct
- [ ] Resource limits appropriate
- [ ] Security headers configured
- [ ] Monitoring endpoints accessible
- [ ] Configuration validation passed
- [ ] Security audit completed
- [ ] All tests passing

### Configuration Validation

Use the built-in validation script to check your configuration:

```bash
# Validate current .env file
make validate

# Validate specific configuration file
./scripts/setup/validate-config.sh .env.production

# Full deployment validation
make prod-ready
```

The validation script checks:
- ✅ Required environment variables
- ✅ Security requirements (key lengths, patterns)
- ✅ Network configuration (ports, domains)
- ✅ Database and Redis connectivity
- ✅ File permissions and directories
- ✅ Production-specific requirements

## 🆘 Troubleshooting

### Common Configuration Issues

1. **Database Connection Failed**
   - Check `POSTGRES_*` variables
   - Verify network connectivity
   - Check firewall rules

2. **Authentication Not Working**
   - Verify `API_KEY` length (min 32 chars)
   - Check `JWT_SECRET` configuration
   - Validate token expiry settings

3. **File Upload Issues**
   - Check `UPLOAD_DIR` permissions
   - Verify `MAX_FILE_SIZE` setting
   - Check disk space

4. **Performance Issues**
   - Tune `MAX_CONCURRENT_JOBS`
   - Adjust database connection pool
   - Monitor resource usage