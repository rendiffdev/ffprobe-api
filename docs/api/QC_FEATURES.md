# QC Features API Guide

## Overview

The FFprobe API provides comprehensive Quality Control (QC) analysis with **20+ professional QC analysis categories** designed for broadcast, streaming, and post-production workflows.

> 📋 **For the complete detailed list of all QC categories**, see [QC_ANALYSIS_LIST.md](../../QC_ANALYSIS_LIST.md) in the root directory.

This guide focuses on **API usage**, **integration patterns**, and **implementation best practices** for QC analysis.

## API Integration

### Enable Advanced QC Analysis

#### Full QC Analysis
```bash
POST /api/v1/probe/file
Content-Type: application/json

{
  "file_path": "/path/to/media.mxf",
  "content_analysis": true,
  "generate_reports": true,
  "report_formats": ["json", "pdf"]
}
```

#### Selective QC Categories
```bash
POST /api/v1/probe/file
Content-Type: application/json

{
  "file_path": "/path/to/media.mp4",
  "content_analysis": true,
  "qc_categories": [
    "timecode_analysis",
    "pse_analysis", 
    "dead_pixel_analysis",
    "mxf_analysis"
  ]
}
```

### GraphQL QC Query
```graphql
query GetQCAnalysis($id: ID!) {
  analysis(id: $id) {
    result {
      enhancedAnalysis {
        timecodeAnalysis {
          hasTimecode
          isDropFrame
          startTimecode
        }
        mxfAnalysis {
          isMXFFile
          mxfProfile
          validationResults {
            overallCompliance
          }
        }
        pseAnalysis {
          pseRiskLevel
          flashAnalysis {
            flashCount
            maxFlashRate
          }
        }
        deadPixelAnalysis {
          hasDeadPixels
          deadPixelCount
          qualityImpactAssessment {
            severityLevel
          }
        }
      }
    }
  }
}
```

## QC Results Structure

### Standard Response Format
```json
{
  "analysis_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "analysis": {
    "result": {
      "enhanced_analysis": {
        "stream_counts": {
          "total_streams": 3,
          "video_streams": 1,
          "audio_streams": 2
        },
        "timecode_analysis": {
          "has_timecode": true,
          "is_drop_frame": false,
          "start_timecode": "01:00:00:00"
        },
        "transport_stream_analysis": {
          "is_mpeg_transport_stream": true,
          "total_pids": 12,
          "errors": []
        },
        "mxf_analysis": {
          "is_mxf_file": true,
          "mxf_profile": "OP1a",
          "validation_results": {
            "overall_compliance": true
          }
        },
        "pse_analysis": {
          "pse_risk_level": "safe",
          "broadcast_compliance": {
            "itc_compliant": true,
            "ofcom_compliant": true
          }
        }
      }
    }
  }
}
```

## Workflow Integration

### Automated QC Pipeline
```bash
# Step 1: Upload and analyze
curl -X POST -F "file=@video.mxf" \
  -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/probe/file

# Step 2: Check analysis status
curl -H "X-API-Key: your-api-key" \
  http://localhost:8080/api/v1/probe/status/{analysis-id}

# Step 3: Generate compliance report
curl -X POST \
  -H "X-API-Key: your-api-key" \
  -d '{"analysis_id": "...", "formats": ["pdf", "json"]}' \
  http://localhost:8080/api/v1/reports/analysis
```

### Batch QC Processing
```bash
POST /api/v1/batch/analyze
Content-Type: application/json

{
  "files": [
    "/path/to/file1.mxf",
    "/path/to/file2.mp4"
  ],
  "qc_profile": "broadcast_compliance",
  "webhook_url": "https://your-app.com/qc-webhook"
}
```

## QC Profiles

### Broadcast Compliance Profile
- Timecode Analysis ✓
- AFD Analysis ✓  
- Transport Stream Analysis ✓
- PSE Analysis ✓
- MXF Validation ✓

### Streaming Platform Profile
- Dead Pixel Detection ✓
- PSE Analysis ✓
- Transport Stream Analysis ✓
- Content Analysis ✓

### Post-Production Profile
- Timecode Analysis ✓
- MXF Validation ✓
- Audio Wrapping Analysis ✓
- Dead Pixel Detection ✓

### Archive/Preservation Profile
- Endianness Analysis ✓
- MXF Validation ✓
- IMF Compliance ✓
- Audio Wrapping Analysis ✓

## Error Handling

### QC Analysis Errors
```json
{
  "error": "QC analysis failed",
  "code": "QC_ANALYSIS_ERROR",
  "details": {
    "failed_categories": ["mxf_analysis"],
    "error_messages": {
      "mxf_analysis": "File is not a valid MXF format"
    },
    "successful_categories": ["timecode_analysis", "pse_analysis"],
    "partial_results": true
  }
}
```

### Graceful Degradation
- QC failures don't prevent basic analysis
- Partial results returned with clear status
- Detailed error reporting for troubleshooting
- Fallback to basic analysis when advanced features unavailable

### Validation Levels
- **Critical**: Issues that prevent content distribution
- **Major**: Issues that may impact quality or compliance  
- **Minor**: Recommendations for optimization
- **Informational**: Technical details and insights

## Performance Considerations

### Analysis Performance
- **Quick QC** (Essential categories): ~2-5 seconds
- **Standard QC** (Most categories): ~10-30 seconds
- **Comprehensive QC** (All categories): ~30-60 seconds
- **Custom QC** (Selected categories): Variable

### Resource Requirements
- **CPU**: Multi-threaded analysis for optimal performance
- **Memory**: Scales with file size and analysis complexity
- **Storage**: Temporary space for frame extraction and analysis
- **Network**: Efficient streaming for remote file analysis

### Optimization Tips
```json
{
  "performance_settings": {
    "parallel_analysis": true,
    "frame_sampling_rate": 0.1,
    "max_analysis_time": 300,
    "enable_caching": true
  }
}
```

## Compliance Standards

### Supported Standards
- **ITU-R BT.1702**: Photosensitive epilepsy guidance
- **ITU-R BT.1868**: Active Format Description  
- **SMPTE ST 377**: Material Exchange Format
- **SMPTE ST 2067**: Interoperable Master Format
- **EBU Tech 3253**: European PSE guidelines

### Regional Compliance
- **FCC** (USA): Broadcast regulations
- **Ofcom** (UK): Broadcasting standards
- **ARIB** (Japan): Digital broadcasting standards
- **EBU** (Europe): Broadcasting guidelines

## Best Practices

### QC Implementation
1. **Define Requirements**: Determine essential vs optional QC categories
2. **Set Thresholds**: Configure pass/fail criteria for each category
3. **Automate Workflows**: Integrate QC into production pipelines
4. **Monitor Performance**: Track analysis times and resource usage
5. **Regular Updates**: Keep QC standards current with industry changes

### Quality Assurance
- Regular validation against reference content
- Calibration with industry-standard QC tools
- Continuous monitoring of analysis accuracy
- User feedback integration for improvements

### Integration Patterns

#### Microservices Architecture
```yaml
# compose.yml
services:
  qc-api:
    image: ffprobe-api:latest
    environment:
      - QC_PROFILE=broadcast_compliance
      - ENABLE_ADVANCED_QC=true
  
  workflow-engine:
    depends_on:
      - qc-api
    environment:
      - QC_ENDPOINT=http://qc-api:8080/api/v1/probe/file
```

#### Webhook Integration
```javascript
// Webhook handler for QC results
app.post('/qc-webhook', (req, res) => {
  const { analysis_id, status, qc_results } = req.body;
  
  if (status === 'completed') {
    // Process QC results
    const criticalIssues = qc_results.critical_findings || [];
    if (criticalIssues.length > 0) {
      // Handle critical QC failures
      rejectContent(analysis_id, criticalIssues);
    } else {
      // Approve content for distribution
      approveContent(analysis_id);
    }
  }
  
  res.status(200).json({ received: true });
});
```

## Monitoring and Alerts

### QC Metrics
- Analysis completion rate
- QC failure rates by category
- Processing time per category
- Compliance pass/fail rates

### Alert Configuration
```json
{
  "qc_alerts": {
    "high_pse_risk": {
      "condition": "pse_risk_level >= 'high'",
      "action": "immediate_notification"
    },
    "mxf_compliance_failure": {
      "condition": "mxf_analysis.overall_compliance == false",
      "action": "workflow_halt"
    },
    "dead_pixel_threshold": {
      "condition": "dead_pixel_count > 10",
      "action": "quality_review_required"
    }
  }
}
```

---

This guide covers API integration for QC analysis. For detailed information about each QC category, see the [Complete QC Analysis List](../../QC_ANALYSIS_LIST.md).