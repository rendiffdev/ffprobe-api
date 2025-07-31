# 🚀 FFprobe API Deployment Guide

## 📋 Deployment Options Overview

### 🟢 Simple Deployment (Recommended for Small/Test Organizations)
**File**: `compose.simple.yml`  
**Purpose**: Lightweight, ready-to-deploy setup without monitoring overhead

**What's Included:**
- ✅ FFprobe API service
- ✅ PostgreSQL database
- ✅ Redis cache
- ❌ No Prometheus/Grafana (saves resources)
- ❌ No Ollama LLM (disabled by default)

**Resource Usage:**
- Memory: ~2-3GB total
- CPU: 2-4 cores recommended
- Storage: ~5GB base + uploads

**Command:**
```bash
docker compose -f compose.simple.yml up -d
```

**Perfect for:**
- Small organizations
- Test/staging environments
- Cost-conscious deployments
- Quick demos/proofs of concept

---

### 🟡 Development Deployment
**File**: `compose.yml + compose.dev.yml`  
**Purpose**: Local development with debugging tools

**What's Included:**
- ✅ All simple deployment features
- ✅ Adminer (database GUI)
- ✅ Redis Commander (Redis GUI)
- ✅ Hot reload for development
- ✅ Debug logging enabled

**Command:**
```bash
docker compose -f compose.yml -f compose.dev.yml up -d
```

---

### 🟠 Production Deployment  
**File**: `compose.yml + compose.production.yml`
**Purpose**: Medium-scale production without monitoring

**What's Included:**
- ✅ All simple deployment features
- ✅ Ollama LLM service (AI features)
- ✅ Production-optimized settings
- ✅ Resource limits configured
- ❌ No monitoring stack (keeps it lightweight)

**Resource Usage:**
- Memory: ~6-8GB total
- CPU: 4-6 cores recommended
- Storage: ~15GB base + models + uploads

**Command:**
```bash
docker compose -f compose.yml -f compose.production.yml up -d
```

---

### 🔴 Enterprise Deployment
**File**: `compose.yml + compose.enterprise.yml`
**Purpose**: Full-scale enterprise with monitoring

**What's Included:**
- ✅ All production deployment features
- ✅ **Prometheus monitoring**
- ✅ **Grafana dashboards**
- ✅ Load balancer (Nginx)
- ✅ Message queue (RabbitMQ)
- ✅ Horizontal scaling support
- ✅ Enhanced resource allocation

**Resource Usage:**
- Memory: ~12-16GB total
- CPU: 8+ cores recommended
- Storage: ~30GB base + monitoring data

**Command:**
```bash
docker compose -f compose.yml -f compose.enterprise.yml up -d
```

---

## 🎯 Which Deployment Should You Choose?

### Choose **Simple** if:
- ✅ Small team (< 10 users)
- ✅ Budget/resource constraints
- ✅ Testing or staging environment
- ✅ Basic video analysis needs
- ✅ Don't need monitoring dashboards

### Choose **Production** if:
- ✅ Medium team (10-50 users)
- ✅ Need AI-powered insights
- ✅ Production workload
- ✅ Want cost-effective monitoring (logs only)

### Choose **Enterprise** if:
- ✅ Large team (50+ users)
- ✅ Need comprehensive monitoring
- ✅ High availability requirements  
- ✅ Compliance/audit requirements
- ✅ Performance analytics needed

---

## 🔧 Quick Setup Commands

### Simple Deployment (Recommended Start)
```bash
# 1. Clone repository
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# 2. Set environment variables
cp .env.example .env
# Edit .env with your values

# 3. Deploy
docker compose -f compose.simple.yml up -d

# 4. Verify
curl http://localhost:8080/health
```

### Upgrading from Simple to Production
```bash
# Stop simple deployment
docker compose -f compose.simple.yml down

# Start production deployment (keeps data)
docker compose -f compose.yml -f compose.production.yml up -d
```

### Monitoring Access (Enterprise Only)
- **Grafana**: http://localhost:3000 (admin/admin_change_this)  
- **Prometheus**: http://localhost:9090
- **API**: http://localhost:8080

---

## 💡 Cost Optimization Tips

1. **Start Simple**: Begin with `compose.simple.yml` and upgrade as needed
2. **Disable LLM**: Set `ENABLE_LOCAL_LLM=false` to save 3-4GB memory
3. **Use External Services**: Consider managed PostgreSQL/Redis for production
4. **Resource Limits**: Adjust memory/CPU limits based on your actual usage

---

## 🔒 Security Notes

All deployment options include:
- ✅ Non-root container execution
- ✅ Security headers enabled
- ✅ JWT authentication
- ✅ Rate limiting
- ✅ Input validation
- ✅ SQL injection protection

**Production Checklist:**
- [ ] Change default passwords in `.env`
- [ ] Use strong JWT secrets (32+ characters)
- [ ] Configure SSL/TLS certificates
- [ ] Set up regular backups
- [ ] Monitor logs for security events

---

## 📊 Resource Requirements Summary

| Deployment | Memory | CPU | Storage | Monitoring |
|------------|--------|-----|---------|------------|
| **Simple** | 2-3GB | 2-4 cores | 5GB+ | Logs only |
| **Production** | 6-8GB | 4-6 cores | 15GB+ | Logs only |
| **Enterprise** | 12-16GB | 8+ cores | 30GB+ | Full stack |

---

*For detailed configuration options, see the main [README.md](README.md)*