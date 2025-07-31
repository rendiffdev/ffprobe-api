# 🚀 FFprobe API Deployment Guide

## 📋 Deployment Options Overview

### 🟢 Simple Deployment (Recommended for Small/Test Organizations)
**File**: `compose.simple.yml`  
**Purpose**: Complete LLM-powered API setup without monitoring overhead

**What's Included:**
- ✅ FFprobe API service
- ✅ PostgreSQL database  
- ✅ Redis cache
- ✅ **Ollama LLM (enabled by default)** - Essential for AI-powered analysis
- ✅ **OpenRouter fallback** - Automatic fallback for enhanced reliability
- ❌ No Prometheus/Grafana (enterprise-only monitoring)

**Resource Usage:**
- Memory: ~4-5GB total (includes LLM)
- CPU: 2-4 cores recommended
- Storage: ~8GB base + models + uploads

**Command:**
```bash
docker compose -f compose.simple.yml up -d
```

**Perfect for:**
- Small organizations
- Test/staging environments  
- Cost-conscious deployments
- Quick demos with AI features

---

### 🟡 Development Deployment
**File**: `compose.yml + compose.dev.yml`  
**Purpose**: Local development with debugging tools and AI features

**What's Included:**
- ✅ All simple deployment features
- ✅ Adminer (database GUI)
- ✅ Redis Commander (Redis GUI)
- ✅ Hot reload for development
- ✅ Debug logging enabled
- ✅ Full LLM capabilities

**Command:**
```bash
docker compose -f compose.yml -f compose.dev.yml up -d
```

---

### 🟠 Production Deployment  
**File**: `compose.yml + compose.production.yml`
**Purpose**: Medium-scale production with enhanced AI features

**What's Included:**
- ✅ All simple deployment features
- ✅ **Enhanced Ollama setup** - Optimized for production workloads
- ✅ **Multiple LLM models** - Better AI analysis variety
- ✅ Production-optimized settings
- ✅ Resource limits configured
- ✅ **Intelligent LLM fallback** - Local-first, cloud backup
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
**Purpose**: Full-scale enterprise with monitoring and AI intelligence

**What's Included:**
- ✅ All production deployment features
- ✅ **Prometheus monitoring**
- ✅ **Grafana dashboards**
- ✅ Load balancer (Nginx)
- ✅ Message queue (RabbitMQ)
- ✅ **Advanced LLM orchestration** - Multiple models with smart routing
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

## 🤖 AI/LLM Features Across All Deployments

All deployment options include **LLM-powered analysis** by default:

### 🎯 **What's AI-Powered:**
- ✅ **Video Analysis Reports** - Human-readable technical insights
- ✅ **Quality Assessment** - Professional video quality evaluation  
- ✅ **Comparison Analysis** - AI-driven before/after analysis
- ✅ **Technical Recommendations** - FFmpeg optimization suggestions
- ✅ **Format Suitability** - Delivery platform recommendations

### 🔄 **Smart Fallback System:**
1. **Local LLM First** - Uses Ollama for privacy and speed
2. **OpenRouter Fallback** - Automatic cloud backup if local fails
3. **Graceful Degradation** - API continues working without AI if both fail

### ⚙️ **LLM Configuration:**
```bash
# Local LLM (default: enabled)
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=phi3:mini

# OpenRouter fallback (optional)
OPENROUTER_API_KEY=your-api-key-here
```

---

## 🎯 Which Deployment Should You Choose?

### Choose **Simple** if:
- ✅ Small team (< 10 users)
- ✅ Want AI features without complexity
- ✅ Budget/resource constraints
- ✅ Testing or staging environment
- ✅ Don't need monitoring dashboards

### Choose **Production** if:
- ✅ Medium team (10-50 users)
- ✅ Need enhanced AI performance
- ✅ Production workload with AI requirements
- ✅ Want optimized LLM processing

### Choose **Enterprise** if:
- ✅ Large team (50+ users)
- ✅ Need comprehensive monitoring  
- ✅ High availability requirements
- ✅ Advanced AI orchestration needed
- ✅ Compliance/audit requirements

---

## 🔧 Quick Setup Commands

### Simple Deployment (LLM-Powered)
```bash
# 1. Clone repository
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# 2. Set environment variables
cp .env.example .env
# Edit .env with your values

# 3. Deploy with AI features
docker compose -f compose.simple.yml up -d

# 4. Verify (should show LLM status)
curl http://localhost:8080/health
```

### Test AI Features
```bash
# Upload a video and get AI analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -F "file=@test-video.mp4"

# The response will include LLM-generated insights
```

---

## 💡 Cost Optimization Tips

1. **Start Simple**: Begin with `compose.simple.yml` - includes AI without monitoring overhead
2. **Local LLM First**: Uses free Ollama models, only pays for OpenRouter fallback when needed
3. **Smart Resource Limits**: Each deployment tier optimized for different scales
4. **Optional Cloud LLM**: OpenRouter fallback is optional - works great with just local LLM

---

## 📊 Resource Requirements Summary

| Deployment | Memory | CPU | Storage | AI Features | Monitoring |
|------------|--------|-----|---------|-------------|------------|
| **Simple** | 4-5GB | 2-4 cores | 8GB+ | ✅ Full LLM | Logs only |
| **Production** | 6-8GB | 4-6 cores | 15GB+ | ✅ Enhanced LLM | Logs only |
| **Enterprise** | 12-16GB | 8+ cores | 30GB+ | ✅ Advanced LLM | Full monitoring |

---

## 🔒 Security & Privacy

- **Local LLM**: All AI processing can run locally for maximum privacy
- **Encrypted Communication**: All external LLM calls use HTTPS
- **API Key Security**: OpenRouter keys are optional and securely managed
- **No Data Leakage**: Local-first approach means your videos stay on your infrastructure

---

*The FFprobe API is designed to be **AI-first** while maintaining complete flexibility in deployment scale and privacy requirements.*