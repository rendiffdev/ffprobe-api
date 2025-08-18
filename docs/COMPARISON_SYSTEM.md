# 📊 Video Comparison System

## Overview

The Video Comparison System allows you to analyze improvements between an original video and its modified version after optimization. This addresses the critical business requirement of validating whether video processing improvements are actually beneficial.

## 🎯 Business Problem Solved

**Before**: "I modified my video to fix issues, but how do I know if it's actually better?"

**After**: Comprehensive comparison analysis with:
- ✅ Quality improvement validation
- ✅ File size optimization analysis  
- ✅ AI-powered professional assessment
- ✅ Actionable recommendations
- ✅ Pass/fail decision support

## 🚀 Quick Start

### Prerequisites
- API authentication set up (see [API Authentication Guide](api/authentication.md))
- Two video analyses to compare

### Basic Comparison Workflow

```bash
# 1. Set your API key
export API_KEY="your-api-key"

# 2. Create comparison
curl -X POST http://localhost:8080/api/v1/comparisons/quick \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "original_analysis_id": "uuid-of-original",
    "modified_analysis_id": "uuid-of-modified", 
    "include_llm": true
  }'

# 3. Get results
curl -X GET http://localhost:8080/api/v1/comparisons/{comparison-id}/report \
  -H "X-API-Key: $API_KEY"
```

## 📋 Complete Workflow Example

### Step 1: Analyze Original Video
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@original-video.mp4"
# Save the returned analysis ID
```

### Step 2: Process Your Video (External)
```bash
# Example: Re-encode with better settings
ffmpeg -i original-video.mp4 -c:v libx264 -crf 23 modified-video.mp4
```

### Step 3: Analyze Modified Video
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@modified-video.mp4"
# Save the returned analysis ID
```

### Step 4: Compare Results
```bash
curl -X POST http://localhost:8080/api/v1/comparisons \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "original_analysis_id": "original-uuid",
    "modified_analysis_id": "modified-uuid",
    "comparison_type": "full_analysis",
    "include_llm": true
  }'
```

## 📊 Understanding Results

### Quality Verdicts

| Verdict | Score Range | Meaning | Action |
|---------|-------------|---------|--------|
| `significant_improvement` | 30+ points | Major quality gains | ✅ Accept |
| `improvement` | 10-30 points | Noticeable gains | ✅ Accept |
| `minimal_change` | -10 to +10 | Small changes | ⚠️ Review |
| `regression` | -30 to -10 | Quality decreased | ❌ Optimize |
| `significant_regression` | -30+ points | Major quality loss | ❌ Reject |

### Example Result Analysis

```json
{
  "summary": {
    "overall_improvement": 25.5,
    "quality_verdict": "improvement", 
    "recommended_action": "accept"
  },
  "file_size": {
    "percentage_change": -30.0,
    "compression_ratio": 0.7
  },
  "quality_score": {
    "overall_score": 87.5,
    "video_score": 85.0,
    "compression_score": 95.0
  }
}
```

**Interpretation**: 30% file size reduction with maintained quality → Accept changes ✅

## 🎯 Comparison Types

### Full Analysis (`full_analysis`)
Complete comparison including:
- Video/audio quality metrics
- File size optimization
- Format changes
- AI professional assessment

### Quality Focus (`quality`)
Detailed quality analysis:
- VMAF, PSNR, SSIM metrics
- Perceptual quality assessment
- Frame-by-frame analysis

### Optimization (`optimization`)
File size and efficiency focus:
- Compression effectiveness
- Quality vs size tradeoffs
- Storage cost impact

## 🚨 Common Use Cases

### 1. Post-Processing Validation
After noise reduction, color correction, stabilization:
```bash
POST /api/v1/comparisons/quick
{
  "comparison_type": "quality",
  "include_llm": true
}
```

### 2. Encoding Optimization  
Testing different codec settings:
```bash
POST /api/v1/comparisons
{
  "comparison_type": "optimization",
  "threshold": 85.0
}
```

### 3. Quality Assurance
Systematic validation before delivery:
```bash
POST /api/v1/comparisons
{
  "comparison_type": "full_analysis",
  "include_llm": true,
  "threshold": 90.0
}
```

## 💡 Best Practices

1. **Always Compare**: Create comparisons for any video modification
2. **Set Thresholds**: Use appropriate quality thresholds for your content
3. **Use AI Assessment**: Enable LLM analysis for complex decisions
4. **Document Results**: Keep comparison history for learning

## 🔍 API Reference

### Create Comparison
```bash
POST /api/v1/comparisons
{
  "original_analysis_id": "uuid",
  "modified_analysis_id": "uuid", 
  "comparison_type": "full_analysis",
  "include_llm": true,
  "threshold": 85.0
}
```

### Get Results
```bash
GET /api/v1/comparisons/{id}
GET /api/v1/comparisons/{id}/report?format=detailed
```

### List Comparisons
```bash
GET /api/v1/comparisons?limit=20&offset=0
```

## 📈 Success Examples

### File Optimization Success
```
Original: 150MB → Modified: 87MB (-42%)
Quality: Maintained (92/100)
Result: $2,400/year storage savings ✅
```

### Quality Enhancement Success
```
VMAF Score: 65 → 89 (+37% improvement)  
File Size: Minimal increase (+5%)
Result: Broadcast-ready content ✅
```

For complete API authentication setup, see the [API Authentication Guide](api/authentication.md).