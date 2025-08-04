# Troubleshooting Guide

> **Comprehensive troubleshooting guide for common issues with FFprobe API**

## Quick Diagnostics

### System Health Check

```bash
# Check all services
curl http://localhost:8080/health

# Check specific components
docker compose ps
docker compose logs --tail=50

# Database connectivity
docker compose exec postgres pg_isready

# Redis connectivity  
docker compose exec redis redis-cli ping
```

## Common Issues

### 🔴 Service Won't Start

#### Symptoms
- Container exits immediately
- Port binding errors
- Service unreachable

#### Diagnosis
```bash
# Check container status
docker compose ps

# View startup logs
docker compose logs ffprobe-api

# Check port availability
netstat -an | grep 8080
lsof -i :8080
```

#### Solutions

**Port already in use:**
```bash
# Change port in .env
API_PORT=8081

# Or kill process using port
kill -9 $(lsof -t -i:8080)
```

**Configuration errors:**
```bash
# Validate environment file
docker compose config

# Check required variables
grep -E "API_KEY|JWT_SECRET|POSTGRES_PASSWORD" .env
```

**Resource constraints:**
```bash
# Increase Docker resources
docker system prune -a
docker compose down
docker compose up -d
```

### 🔐 Authentication Failures

#### Symptoms
- 401 Unauthorized responses
- Invalid token errors
- API key not working

#### Diagnosis
```bash
# Test API key
curl -H "X-API-Key: your-key" http://localhost:8080/api/v1/probe/health

# Check API key format
echo -n "your-api-key" | wc -c  # Should be 32+ characters

# Verify JWT configuration
curl -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"email":"test@example.com","password":"password"}'
```

#### Solutions

**Invalid API key:**
```bash
# Generate new API key
openssl rand -hex 32

# Update environment
API_KEY=new-generated-key
docker compose restart ffprobe-api
```

**JWT token expired:**
```javascript
// Decode JWT to check expiry
const token = "your.jwt.token";
const payload = JSON.parse(atob(token.split('.')[1]));
console.log(new Date(payload.exp * 1000));
```

**CORS issues:**
```bash
# Update CORS configuration
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
docker compose restart ffprobe-api
```

### 💾 Database Connection Issues

#### Symptoms
- "database connection failed" errors
- Slow queries
- Connection pool exhausted

#### Diagnosis
```bash
# Check PostgreSQL status
docker compose exec postgres pg_isready

# View active connections
docker compose exec postgres psql -U postgres -c \
  "SELECT count(*) FROM pg_stat_activity;"

# Check database logs
docker compose logs postgres --tail=100
```

#### Solutions

**Connection refused:**
```bash
# Verify database configuration
docker compose exec ffprobe-api env | grep POSTGRES

# Test direct connection
docker compose exec postgres psql -U postgres ffprobe_api

# Reset database
docker compose down
docker volume rm ffprobe-api_postgres_data
docker compose up -d
```

**Connection pool exhausted:**
```sql
-- Increase connection limit
ALTER SYSTEM SET max_connections = 200;

-- Kill idle connections
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'idle' 
  AND state_change < NOW() - INTERVAL '10 minutes';
```

**Slow queries:**
```sql
-- Find slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Add missing indexes
CREATE INDEX idx_analyses_user_id ON analyses(user_id);
CREATE INDEX idx_analyses_status ON analyses(status);
```

### 📹 FFprobe/FFmpeg Issues

#### Symptoms
- "ffprobe not found" errors
- Analysis failures
- Unsupported format errors

#### Diagnosis
```bash
# Check FFprobe availability
docker compose exec ffprobe-api ffprobe -version

# Test FFprobe directly
docker compose exec ffprobe-api ffprobe /path/to/test.mp4

# Check file permissions
docker compose exec ffprobe-api ls -la /path/to/video/
```

#### Solutions

**FFprobe not found:**
```dockerfile
# Rebuild with FFmpeg
FROM alpine:3.19
RUN apk add --no-cache ffmpeg
```

**Permission denied:**
```bash
# Fix file permissions
chmod 644 /path/to/video.mp4

# Mount with correct permissions
volumes:
  - ./videos:/videos:ro
```

**Format not supported:**
```bash
# Check supported formats
docker compose exec ffprobe-api ffprobe -formats

# Update FFmpeg
docker compose build --no-cache
```

### 🚀 Performance Issues

#### Symptoms
- Slow response times
- High CPU/memory usage
- Timeouts

#### Diagnosis
```bash
# Monitor resource usage
docker stats

# Check application metrics
curl http://localhost:8080/metrics | grep -E "memory|cpu|goroutine"

# Analyze response times
curl -w "@curl-format.txt" -o /dev/null -s \
  http://localhost:8080/api/v1/probe/file
```

#### Solutions

**High memory usage:**
```bash
# Set memory limits
docker compose down
docker compose up -d --memory="2g"

# Optimize configuration
MAX_CONCURRENT_ANALYSES=5
ANALYSIS_TIMEOUT=300
```

**Slow processing:**
```bash
# Enable caching
REDIS_HOST=redis
CACHE_TTL=3600

# Increase worker pool
WORKER_POOL_SIZE=10
```

**Database performance:**
```sql
-- Vacuum and analyze
VACUUM ANALYZE analyses;

-- Update statistics
ANALYZE;
```

### 📊 Quality Metrics Failures

#### Symptoms
- VMAF calculation errors
- Missing quality scores
- Content analysis failures

#### Diagnosis
```bash
# Check enhanced analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-key" \
  -d '{"file_path":"/test.mp4","content_analysis":true}'

# View processing logs
docker compose logs ffprobe-api | grep -i "vmaf\|quality"
```

#### Solutions

**VMAF model missing:**
```bash
# Download VMAF models
wget https://github.com/Netflix/vmaf/raw/master/model/vmaf_v0.6.1.json
mkdir -p /app/models
mv vmaf_v0.6.1.json /app/models/
```

**Content analysis timeout:**
```bash
# Increase timeout
CONTENT_ANALYSIS_TIMEOUT=600
FILTER_TIMEOUT=120
```

### 🔄 Batch Processing Issues

#### Symptoms
- Jobs stuck in queue
- Batch failures
- Progress not updating

#### Diagnosis
```bash
# Check job queue
curl http://localhost:8080/api/v1/batch/status

# Monitor worker logs
docker compose logs ffprobe-worker --tail=100

# Check Redis queue
docker compose exec redis redis-cli LLEN job_queue
```

#### Solutions

**Stuck jobs:**
```bash
# Clear job queue
docker compose exec redis redis-cli FLUSHDB

# Restart workers
docker compose restart ffprobe-worker
```

**Worker crashes:**
```go
// Add recovery in worker
defer func() {
    if r := recover(); r != nil {
        log.Printf("Worker panic: %v", r)
        // Re-queue job
    }
}()
```

## Error Messages Reference

### HTTP Status Codes

| Code | Meaning | Common Causes | Solution |
|------|---------|---------------|----------|
| 400 | Bad Request | Invalid input | Check request format |
| 401 | Unauthorized | Missing/invalid auth | Verify credentials |
| 403 | Forbidden | Insufficient permissions | Check user role |
| 404 | Not Found | Resource doesn't exist | Verify resource ID |
| 429 | Too Many Requests | Rate limit exceeded | Wait or increase limit |
| 500 | Internal Error | Server issue | Check logs |
| 503 | Service Unavailable | Service down | Check service health |

### Application Error Codes

```json
{
  "INVALID_FILE_PATH": "File path validation failed",
  "UNSUPPORTED_FORMAT": "Video format not supported",
  "PROCESSING_TIMEOUT": "Analysis exceeded timeout",
  "DATABASE_ERROR": "Database operation failed",
  "FFPROBE_ERROR": "FFprobe execution failed",
  "QUEUE_FULL": "Processing queue at capacity"
}
```

## Debug Mode

### Enable Debug Logging

```bash
# Set debug level
LOG_LEVEL=debug
DEBUG=true

# Restart service
docker compose restart ffprobe-api

# View debug logs
docker compose logs -f ffprobe-api | grep DEBUG
```

### Debug Endpoints

```bash
# Get system info
curl http://localhost:8080/debug/vars

# Profile CPU
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Check goroutines
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

## Recovery Procedures

### Emergency Restart

```bash
#!/bin/bash
# emergency-restart.sh

echo "Starting emergency restart..."

# Stop all services
docker compose down

# Clear temporary data
rm -rf /tmp/ffprobe-*

# Prune Docker resources
docker system prune -f

# Start services
docker compose up -d

# Wait for health
sleep 10
curl http://localhost:8080/health
```

### Database Recovery

```bash
# Backup current state
docker compose exec postgres pg_dump -U postgres ffprobe_api > backup.sql

# Restore from backup
docker compose exec -T postgres psql -U postgres ffprobe_api < backup.sql

# Rebuild indexes
docker compose exec postgres psql -U postgres ffprobe_api -c "REINDEX DATABASE ffprobe_api;"
```

### Cache Reset

```bash
# Clear Redis cache
docker compose exec redis redis-cli FLUSHALL

# Restart with fresh cache
docker compose restart redis ffprobe-api
```

## Monitoring Commands

### Real-time Monitoring

```bash
# Watch service logs
watch 'docker compose logs --tail=20 ffprobe-api'

# Monitor resources
watch docker stats

# Track requests
tail -f logs/access.log | grep POST
```

### Performance Analysis

```bash
# Analyze slow queries
docker compose exec postgres psql -U postgres -c \
  "SELECT * FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;"

# Check cache hit rate
docker compose exec redis redis-cli INFO stats | grep keyspace
```

## Support Escalation

### Level 1: Self-Service
1. Check this troubleshooting guide
2. Review error logs
3. Verify configuration
4. Restart services

### Level 2: Community Support
1. Search [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
2. Post in [Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
3. Check Stack Overflow

### Level 3: Direct Support
1. Create detailed issue report
2. Include logs and configuration
3. Email: support@rendiff.dev

## Diagnostic Checklist

### Before Reporting Issues

- [ ] Service health check passes
- [ ] Latest version deployed
- [ ] Configuration validated
- [ ] Logs reviewed for errors
- [ ] Resources (CPU/Memory) adequate
- [ ] Network connectivity verified
- [ ] Database accessible
- [ ] FFprobe functional
- [ ] Authentication working
- [ ] Reproducible in isolation

### Information to Provide

```markdown
## Issue Report

**Environment:**
- Version: [e.g., 1.0.0]
- OS: [e.g., Ubuntu 22.04]
- Docker: [version]
- Deployment: [Docker Compose/K8s]

**Issue:**
- Description: [What happened]
- Expected: [What should happen]
- Actual: [What actually happened]

**Steps to Reproduce:**
1. [First step]
2. [Second step]

**Logs:**
```
[Include relevant logs]
```

**Configuration:**
```
[Include relevant config]
```
```

---

## Next Steps

- [Monitoring Guide](monitoring.md)
- [Security Guide](security.md)
- [Performance Tuning](../deployment/performance.md)
- [FAQ](../FAQ.md)