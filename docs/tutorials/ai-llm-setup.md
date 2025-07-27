# 🤖 AI/LLM Setup Guide

This guide covers setting up AI-powered analysis capabilities in FFprobe API using either local LLMs (Ollama) or cloud LLMs (OpenRouter).

## 🎯 Overview

FFprobe API provides intelligent media analysis through integrated LLM capabilities:

- **Automatic Report Generation**: Every FFprobe analysis automatically generates an AI-powered insights report
- **Flexible Deployment**: Choose local (Ollama), cloud (OpenRouter), or both with fallback
- **Cost-Effective**: Local processing eliminates API costs while cloud provides high availability
- **Privacy-Focused**: Local processing keeps your media metadata private

## 🚀 Quick Setup

### Interactive Installation (Recommended)

The interactive installer will configure LLM options for you:

```bash
cd ffprobe-api
./scripts/setup/install.sh
```

Choose from these LLM configuration options:
1. **🦙 Local LLM only (Ollama)** - Private, no API costs
2. **☁️ Cloud LLM only (OpenRouter)** - Requires API key, always available  
3. **🔄 Local + Cloud fallback (Recommended)** - Best of both worlds
4. **❌ Disable AI features** - Basic analysis only

## 🦙 Local LLM Setup (Ollama)

### Prerequisites
- **8GB+ RAM** (16GB recommended for larger models)
- **Docker and Docker Compose**

### Configuration

The installer automatically configures Ollama, but for manual setup:

```bash
# Environment variables
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=mistral:7b
```

### Available Models

| Model | Size | RAM Required | Best For |
|-------|------|--------------|----------|
| `mistral:7b` | 4.1GB | 8GB+ | General analysis, fast responses |
| `qwen2:7b` | 4.4GB | 8GB+ | Technical analysis, multilingual |
| `phi3:mini` | 2.3GB | 4GB+ | Lightweight, basic insights |
| `gemma2:9b` | 5.5GB | 12GB+ | Advanced analysis, detailed reports |
| `llama3:8b` | 4.7GB | 8GB+ | Comprehensive media analysis |

### Model Management

```bash
# List available models
docker exec ffprobe-ollama ollama list

# Pull a new model
docker exec ffprobe-ollama ollama pull qwen2:7b

# Change active model (update .env and restart)
OLLAMA_MODEL=qwen2:7b
docker compose restart ffprobe-api
```

## ☁️ Cloud LLM Setup (OpenRouter)

### Get API Key

1. Visit [OpenRouter](https://openrouter.ai/)
2. Sign up for an account
3. Generate an API key
4. Fund your account (pay-per-use pricing)

### Configuration

```bash
# Environment variables
OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxx
ENABLE_LOCAL_LLM=false  # Optional: disable local for cloud-only
```

### Pricing (Approximate)

| Model | Input Cost | Output Cost | Best For |
|-------|------------|-------------|----------|
| GPT-4o | $5/1M tokens | $15/1M tokens | High-quality analysis |
| GPT-3.5 Turbo | $0.5/1M tokens | $1.5/1M tokens | Cost-effective |
| Claude 3 Haiku | $0.25/1M tokens | $1.25/1M tokens | Fast, economical |

## 🔄 Hybrid Setup (Recommended)

Configure both local and cloud for optimal reliability:

```bash
# Enable both options
ENABLE_LOCAL_LLM=true
OLLAMA_URL=http://ollama:11434
OLLAMA_MODEL=mistral:7b
OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxxx
```

**Fallback Logic:**
1. Try local LLM first (faster, private, free)
2. Fall back to cloud LLM if local fails
3. Return analysis without LLM report if both fail

## 📊 LLM Report Features

### Automatic Generation

Every FFprobe analysis automatically includes:

```json
{
  "analysis": {
    "id": "uuid-here",
    "ffprobe_data": { ... },
    "llm_report": "AI-generated insights about the media file...",
    "status": "completed"
  }
}
```

### Report Content

LLM reports include:

- **Technical Summary**: Codec, resolution, bitrate analysis
- **Quality Assessment**: Potential issues or optimizations
- **Compatibility Notes**: Platform and device compatibility
- **Recommendations**: Encoding suggestions
- **Metadata Insights**: Analysis of embedded information

### Sample LLM Report

```text
📊 MEDIA ANALYSIS REPORT

🎬 Technical Specifications:
- Video: H.264 (AVC), 1920x1080, 29.97fps, 8.5 Mbps
- Audio: AAC, 44.1kHz, Stereo, 128 kbps
- Duration: 2:34 minutes
- File Size: 163.4 MB

✅ Quality Assessment:
- Excellent video quality with professional encoding
- Audio levels are well-balanced
- No apparent artifacts or compression issues

🔧 Recommendations:
- Consider H.265 for 30-40% size reduction
- Current settings ideal for streaming platforms
- Compatible with all modern devices

📱 Compatibility:
- iOS/Safari: ✅ Full support
- Android: ✅ Full support  
- Smart TVs: ✅ Widely compatible
- Web browsers: ✅ Universal support
```

## 🔧 Configuration Options

### Environment Variables

```bash
# LLM Configuration
ENABLE_LOCAL_LLM=true              # Enable/disable local LLM
OLLAMA_URL=http://ollama:11434     # Ollama service URL
OLLAMA_MODEL=mistral:7b            # Model to use
OPENROUTER_API_KEY=                # OpenRouter API key

# Advanced Options
LLM_TIMEOUT=300                    # LLM generation timeout (seconds)
LLM_MAX_TOKENS=2048               # Maximum response tokens
LLM_TEMPERATURE=0.7               # Response creativity (0.0-1.0)
```

### Docker Compose Override

For custom Ollama configuration:

```yaml
# compose.override.yml
services:
  ollama:
    environment:
      - OLLAMA_MAX_LOADED_MODELS=2
      - OLLAMA_NUM_PARALLEL=4
    volumes:
      - ./ollama-models:/root/.ollama
    deploy:
      resources:
        limits:
          memory: 12G
        reservations:
          memory: 8G
```

## 🚨 Troubleshooting

### Common Issues

#### Local LLM Issues

**Problem**: "Ollama service unavailable"
```bash
# Check Ollama status
docker compose logs ollama

# Restart Ollama service
docker compose restart ollama

# Verify model is loaded
docker exec ffprobe-ollama ollama list
```

**Problem**: "Out of memory" errors
```bash
# Use smaller model
OLLAMA_MODEL=phi3:mini

# Or increase Docker memory limits
# Docker Desktop: Settings > Resources > Memory
```

#### Cloud LLM Issues

**Problem**: "Invalid API key"
```bash
# Verify API key format
echo $OPENROUTER_API_KEY
# Should start with 'sk-or-v1-'

# Test API key
curl -H "Authorization: Bearer $OPENROUTER_API_KEY" \
     https://openrouter.ai/api/v1/models
```

**Problem**: "Rate limit exceeded"
- Check your OpenRouter account credits
- Consider upgrading your plan
- Enable local LLM as fallback

### Health Checks

```bash
# Check LLM service health
curl -H "X-API-Key: your-key" \
     http://localhost:8080/api/v1/genai/health

# Test LLM generation
curl -X POST -H "X-API-Key: your-key" \
     -H "Content-Type: application/json" \
     -d '{"analysis_id": "test"}' \
     http://localhost:8080/api/v1/genai/analysis
```

## 📈 Performance Optimization

### Local LLM Performance

```bash
# Monitor resource usage
docker stats ollama

# Optimize for your hardware
# CPU-optimized (faster response)
OLLAMA_MODEL=phi3:mini

# GPU-optimized (if available)
# Add GPU support to docker-compose.yml
```

### Cost Optimization

For cloud LLMs:

1. **Use local LLM as primary** with cloud fallback
2. **Choose cost-effective models** (Claude 3 Haiku, GPT-3.5)
3. **Monitor usage** through OpenRouter dashboard
4. **Set spending limits** in your OpenRouter account

## 🔒 Security Considerations

### Local LLM Security

- ✅ **Data Privacy**: All processing happens locally
- ✅ **No External Calls**: Media metadata never leaves your infrastructure
- ✅ **Isolated Processing**: Ollama runs in isolated container

### Cloud LLM Security

- ⚠️ **Data Transmission**: Metadata sent to cloud provider
- ✅ **API Key Security**: Stored as environment variable
- ✅ **HTTPS**: All communications encrypted
- ⚠️ **Provider Policies**: Subject to cloud provider data policies

### Best Practices

1. **Use local LLM for sensitive content**
2. **Rotate API keys regularly**
3. **Monitor API usage**
4. **Review cloud provider data policies**

## 📚 Advanced Usage

### Custom Prompts

Modify LLM prompts in the source code:

```go
// internal/services/llm.go
func (s *LLMService) GenerateAnalysis(ctx context.Context, analysis *models.Analysis) (string, error) {
    prompt := `Analyze this media file data and provide insights:
    
    Technical Specs: %s
    Format Info: %s
    
    Focus on: quality assessment, compatibility, optimization recommendations.`
    
    // ... rest of implementation
}
```

### Multiple Model Support

Configure different models for different analysis types:

```bash
# Quick analysis
OLLAMA_MODEL_QUICK=phi3:mini

# Detailed analysis  
OLLAMA_MODEL_DETAILED=mistral:7b

# Technical analysis
OLLAMA_MODEL_TECHNICAL=qwen2:7b
```

## 🆘 Support

For LLM-related issues:

1. **Check logs**: `docker compose logs ollama ffprobe-api`
2. **Verify configuration**: Review environment variables
3. **Test connectivity**: Use health check endpoints
4. **Resource monitoring**: Check memory and CPU usage
5. **Community support**: GitHub Discussions

---

**🚀 Ready to get started?** Run the interactive installer and select your preferred LLM configuration!

```bash
./scripts/setup/install.sh
```