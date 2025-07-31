# 🆘 Troubleshooting Guide

This guide helps you diagnose and fix common issues with the FFprobe API.

## 🔍 Quick Diagnostics

### Health Check Commands

```bash
# Basic service status
docker-compose ps

# API health with authentication
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health

# Service logs
docker-compose logs -f ffprobe-api

# Resource usage
docker stats --no-stream
```

### Environment Verification

```bash
# Check essential environment variables
echo "API_KEY: ${API_KEY:0:20}..." # Show first 20 chars only
echo "ENABLE_AUTH: $ENABLE_AUTH"
echo "API_PORT: $API_PORT"

# Verify .env file
grep -E "^(API_KEY|ENABLE_AUTH|POSTGRES_PASSWORD)" .env
```

## 🚨 Common Issues

### 1. "Connection refused" or "Service unavailable"

**Symptoms:**
```bash
curl: (7) Failed to connect to localhost port 8080: Connection refused
```

**Diagnosis:**
```bash
# Check if services are running
docker-compose ps

# Check port availability
netstat -tulpn | grep :8080

# Check Docker daemon
docker info
```

**Solutions:**
```bash
# Start services
docker-compose up -d

# Check for port conflicts
docker-compose down
docker-compose up -d

# If port is occupied, change port in .env
echo "API_PORT=8081" >> .env
docker-compose up -d
```

### 2. "Invalid API key" or Authentication Errors

**Symptoms:**
```json
{"error": "authentication_failed", "message": "Invalid API key"}
```

**Diagnosis:**
```bash
# Check API key format (should be 79 characters)
echo $API_KEY | wc -c

# Verify key pattern
echo $API_KEY | grep -E '^ffprobe_(test|live)_sk_[a-f0-9]{64}$'

# Test authentication
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health
```

**Solutions:**
```bash
# Generate new API key
export API_KEY="ffprobe_test_sk_$(openssl rand -hex 32)"
echo "API_KEY=$API_KEY" >> .env

# Restart services to pick up new key
docker-compose restart

# Verify authentication is enabled
grep ENABLE_AUTH .env
```

### 3. "File too large" Error

**Symptoms:**
```json
{"error": "file_too_large", "message": "File exceeds maximum size limit"}
```

**Diagnosis:**
```bash
# Check file size
ls -lh your-video.mp4

# Check current limit
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health | jq '.config.max_file_size'
```

**Solutions:**
```bash
# Increase file size limit (example: 10GB)
echo "MAX_FILE_SIZE=10737418240" >> .env
docker-compose restart

# Or compress your video first
ffmpeg -i large-video.mp4 -c:v libx264 -crf 23 compressed-video.mp4
```

### 4. Database Connection Errors

**Symptoms:**
```
pq: password authentication failed for user "ffprobe"
```

**Diagnosis:**
```bash
# Check database service
docker-compose logs postgres

# Test database connection
docker-compose exec postgres pg_isready -U ffprobe
```

**Solutions:**
```bash
# Reset database with fresh password
docker-compose down -v  # WARNING: This deletes data
export DB_PASSWORD="$(openssl rand -hex 16)"
echo "POSTGRES_PASSWORD=$DB_PASSWORD" >> .env
docker-compose up -d

# Wait for database to initialize
sleep 30
docker-compose logs postgres
```

### 5. Video Analysis Fails

**Symptoms:**
```json
{"error": "analysis_failed", "message": "FFprobe execution failed"}
```

**Diagnosis:**
```bash
# Check video file integrity
ffprobe your-video.mp4

# Check available disk space
df -h

# Check service logs
docker-compose logs ffprobe-api | grep ERROR
```

**Solutions:**
```bash
# Test with a simple video
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@test-video.mp4"

# Clear temporary files
docker-compose exec ffprobe-api rm -rf /app/temp/*

# Restart service
docker-compose restart ffprobe-api
```

### 6. AI/LLM Service Issues

**Symptoms:**
```
Failed to connect to Ollama service
```

**Diagnosis:**
```bash
# Check Ollama service
docker-compose logs ollama

# Test Ollama endpoint
curl http://localhost:11434/api/version

# Check model availability
docker-compose exec ollama ollama list
```

**Solutions:**
```bash
# Restart Ollama service
docker-compose restart ollama

# Pull the model manually
docker-compose exec ollama ollama pull phi3:mini

# Disable LLM if not needed
echo "ENABLE_LOCAL_LLM=false" >> .env
docker-compose restart
```

### 7. Performance Issues

**Symptoms:**
- Slow response times
- High CPU/memory usage
- Timeouts

**Diagnosis:**
```bash
# Check resource usage
docker stats --no-stream

# Check processing queue
curl -H "X-API-Key: $API_KEY" http://localhost:8080/metrics | grep queue

# Check disk space
df -h /var/lib/docker
```

**Solutions:**
```bash
# Scale services
docker-compose up -d --scale ffprobe-api=2

# Increase resource limits
echo "API_MEMORY_LIMIT=4G" >> .env
echo "API_CPU_LIMIT=2.0" >> .env
docker-compose up -d

# Clean up old files
docker system prune -f
docker volume prune -f
```

### 8. Port Already in Use

**Symptoms:**
```
Error starting userland proxy: listen tcp 0.0.0.0:8080: bind: address already in use
```

**Diagnosis:**
```bash
# Find what's using the port
sudo netstat -tulpn | grep :8080
# or
sudo lsof -i :8080
```

**Solutions:**
```bash
# Option 1: Kill the process using the port
sudo kill -9 PID_NUMBER

# Option 2: Use a different port
echo "API_PORT=8081" >> .env
docker-compose down
docker-compose up -d

# Update your API calls to use the new port
curl -H "X-API-Key: $API_KEY" http://localhost:8081/health
```

## 🔧 Advanced Diagnostics

### Debug Mode

Enable debug logging:
```bash
echo "LOG_LEVEL=debug" >> .env
echo "DEBUG_MODE=true" >> .env
docker-compose restart

# View detailed logs
docker-compose logs -f ffprobe-api
```

### Container Access

Access running containers for debugging:
```bash
# Access API container
docker-compose exec ffprobe-api bash

# Access database
docker-compose exec postgres psql -U ffprobe -d ffprobe_api

# Access Ollama
docker-compose exec ollama bash
```

### Health Check Details

Get detailed health information:
```bash
# Full health check
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health | jq .

# Service-specific health
curl -H "X-API-Key: $API_KEY" http://localhost:8080/api/v1/probe/health

# Database health
docker-compose exec postgres pg_isready -d ffprobe_api -U ffprobe
```

### Performance Monitoring

Monitor system performance:
```bash
# CPU and memory usage
docker stats --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"

# Disk usage
docker system df

# Network usage
docker-compose exec ffprobe-api netstat -i
```

## 🚨 Emergency Recovery

### Complete Reset (Nuclear Option)

⚠️ **WARNING: This will delete all data and reset everything**

```bash
# Stop and remove everything
docker-compose down -v --remove-orphans

# Remove all containers and images
docker system prune -af

# Remove data directory (if using local storage)
sudo rm -rf ./data

# Start fresh
cp .env.example .env
# Configure your .env file
docker-compose up -d
```

### Backup Before Recovery

```bash
# Backup database
docker-compose exec postgres pg_dump -U ffprobe ffprobe_api > backup.sql

# Backup configuration
cp .env .env.backup

# Backup analysis data
docker cp $(docker-compose ps -q ffprobe-api):/app/data ./data-backup
```

### Restore from Backup

```bash
# Restore database
docker-compose exec -T postgres psql -U ffprobe -d ffprobe_api < backup.sql

# Restore configuration
cp .env.backup .env

# Restore data
docker cp ./data-backup/. $(docker-compose ps -q ffprobe-api):/app/data/
```

## 📊 Health Check Script

Create an automated health check:

```bash
#!/bin/bash
# health-check.sh

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🏥 FFprobe API Health Check"
echo "=========================="

# Check environment
if [ -z "$API_KEY" ]; then
    echo -e "${RED}❌ API_KEY not set${NC}"
    exit 1
fi

# Check services
echo "1. Checking Docker services..."
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Services are running${NC}"
else
    echo -e "${RED}❌ Services are not running${NC}"
    echo "Run: docker-compose up -d"
    exit 1
fi

# Check API health
echo "2. Checking API health..."
if curl -s -H "X-API-Key: $API_KEY" http://localhost:8080/health | grep -q "healthy"; then
    echo -e "${GREEN}✅ API is healthy${NC}"
else
    echo -e "${RED}❌ API health check failed${NC}"
    exit 1
fi

# Check database
echo "3. Checking database..."
if docker-compose exec -T postgres pg_isready -U ffprobe > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Database is ready${NC}"
else
    echo -e "${RED}❌ Database connection failed${NC}"
    exit 1
fi

# Check disk space
echo "4. Checking disk space..."
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -lt 90 ]; then
    echo -e "${GREEN}✅ Disk space OK ($DISK_USAGE% used)${NC}"
else
    echo -e "${YELLOW}⚠️  Disk space low ($DISK_USAGE% used)${NC}"
fi

echo -e "\n🎉 Health check complete!"
```

Use the health check:
```bash
chmod +x health-check.sh
./health-check.sh
```

## 📞 Getting Additional Help

If these troubleshooting steps don't resolve your issue:

1. **Check GitHub Issues**: [Issues Page](https://github.com/rendiffdev/ffprobe-api/issues)
2. **Contact Support**: [dev@rendiff.dev](mailto:dev@rendiff.dev)
3. **Create a Bug Report** with:
   - Error messages and logs
   - Your .env configuration (without secrets)
   - Docker and system versions
   - Steps to reproduce the issue

3. **Enable Debug Mode** and provide logs:
   ```bash
   echo "LOG_LEVEL=debug" >> .env
   docker-compose restart
   docker-compose logs ffprobe-api > debug.log
   ```

4. **System Information**:
   ```bash
   docker --version
   docker-compose --version
   uname -a
   free -h
   df -h
   ```

Include this information when asking for help to get faster, more accurate support.