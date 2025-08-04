# 📋 Quality Checks & Analysis Features

> **Comprehensive guide to all video quality control checks and analysis capabilities**

This document provides a complete reference for all quality control checks, analysis features, and enhanced parameters available in the FFprobe API.

## 📊 Coverage Overview

| Category | Standard Checks | Enhanced Checks | Total Coverage |
|----------|----------------|-----------------|----------------|
| **Container Format** | 5 | 1 | 100% (6/6) |
| **Video Format** | 14 | 4 | 95% (18/19) |
| **Audio Format** | 6 | 2 | 80% (8/10) |
| **Frame Analysis** | 1 | 4 | 100% (5/5) |
| **Content Analysis** | 0 | 9 | 90% (9/10) |
| **Quality Metrics** | 3 | 0 | 30% (3/10) |
| **Overall** | **29** | **20** | **83% (49/59)** |

---

## 🎬 Standard FFprobe Checks

These checks are performed automatically with every analysis request using standard FFprobe capabilities.

### Container/Format Information
```json
"format": {
  "filename": "video.mp4",
  "nb_streams": 2,
  "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
  "format_long_name": "QuickTime / MOV",
  "duration": "120.5",
  "size": "125829120",
  "bit_rate": "8000000",
  "probe_score": 100
}
```

**Available Checks:**
- ✅ **Container Format** - Format identification and validation
- ✅ **Duration** - Total media duration
- ✅ **File Size** - Physical file size in bytes
- ✅ **Bit Rate** - Overall bitrate
- ✅ **Format Score** - FFprobe confidence score

### Stream Information
```json
"streams": [
  {
    "index": 0,
    "codec_name": "h264",
    "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
    "profile": "High",
    "codec_type": "video",
    "width": 1920,
    "height": 1080,
    "coded_width": 1920,
    "coded_height": 1088,
    "sample_aspect_ratio": "1:1",
    "display_aspect_ratio": "16:9",
    "pix_fmt": "yuv420p",
    "level": 40,
    "color_range": "tv",
    "color_space": "bt709",
    "color_transfer": "bt709",
    "color_primaries": "bt709",
    "field_order": "progressive",
    "r_frame_rate": "30/1",
    "avg_frame_rate": "30/1",
    "time_base": "1/15360",
    "duration": "120.000000",
    "bit_rate": "5000000",
    "nb_frames": "3600"
  }
]
```

**Video Stream Checks:**
- ✅ **Video Codec** - Codec identification (H.264, H.265, etc.)
- ✅ **Resolution** - Width and height dimensions
- ✅ **Frame Rate** - Real and average frame rates
- ✅ **Aspect Ratio** - Sample and display aspect ratios
- ✅ **Pixel Format** - Color format (yuv420p, yuv422p, etc.)
- ✅ **Profile & Level** - Codec profile and level
- ✅ **Color Space** - Color space standard (bt709, bt2020, etc.)
- ✅ **Color Transfer** - Transfer characteristics
- ✅ **Color Primaries** - Color primaries standard
- ✅ **Field Order** - Progressive/interlaced detection
- ✅ **Color Range** - Full/limited color range
- ✅ **Video Bit Rate** - Stream bitrate
- ✅ **Frame Count** - Total number of frames
- ✅ **Bit Depth** - Color bit depth

**Audio Stream Checks:**
- ✅ **Audio Codec** - Codec identification (AAC, MP3, etc.)
- ✅ **Sample Rate** - Audio sample rate
- ✅ **Channels** - Number of audio channels
- ✅ **Channel Layout** - Channel configuration
- ✅ **Audio Bit Rate** - Stream bitrate
- ✅ **Sample Format** - Audio sample format

### Additional Standard Information
- ✅ **Chapters** - Chapter markers and metadata
- ✅ **Programs** - Program information for transport streams
- ✅ **Metadata Tags** - Embedded metadata and tags
- ✅ **Packet Information** - Low-level packet data (optional)

---

## 🔍 Enhanced Analysis Checks

Enhanced checks provide additional quality control parameters and content analysis. Enable with `"content_analysis": true` in API requests.

### Stream Analysis Enhancement
```json
"enhanced_analysis": {
  "stream_counts": {
    "total_streams": 3,
    "video_streams": 1,
    "audio_streams": 2,
    "subtitle_streams": 0,
    "data_streams": 0,
    "attachment_streams": 0
  }
}
```

**Enhanced Stream Checks:**
- ✅ **Stream Counting** - Detailed breakdown by stream type
- ✅ **Closed Caption Detection** - CC stream identification

### Video Analysis Enhancement
```json
"video_analysis": {
  "chroma_subsampling": "4:2:0",
  "matrix_coefficients": "ITU-R BT.709",
  "bit_rate_mode": "CBR",
  "has_closed_captions": false
}
```

**Enhanced Video Checks:**
- ✅ **Chroma Subsampling** - YUV subsampling pattern (4:2:0, 4:2:2, 4:4:4)
- ✅ **Matrix Coefficients** - Color matrix standards mapping
- ✅ **Video Bit Rate Mode** - CBR/VBR detection
- ✅ **Closed Caption Validation** - Accessibility compliance

### GOP (Group of Pictures) Analysis
```json
"gop_analysis": {
  "average_gop_size": 30.0,
  "max_gop_size": 30,
  "min_gop_size": 30,
  "keyframe_count": 120,
  "total_frame_count": 3600,
  "gop_pattern": "Regular (GOP=30)"
}
```

**GOP Structure Checks:**
- ✅ **GOP Size Analysis** - Average, min, max GOP sizes
- ✅ **Keyframe Count** - I-frame detection and counting
- ✅ **GOP Pattern Detection** - Regular/irregular pattern identification
- ✅ **GOP Consistency** - Structure regularity validation

### Frame Statistics Analysis
```json
"frame_statistics": {
  "total_frames": 3600,
  "i_frames": 120,
  "p_frames": 2400,
  "b_frames": 1080,
  "frame_types": {
    "I": 120,
    "P": 2400,
    "B": 1080
  },
  "average_frame_size": 4166.67,
  "max_frame_size": 15000,
  "min_frame_size": 1200
}
```

**Frame-Level Checks:**
- ✅ **Frame Type Distribution** - I/P/B frame counts and ratios
- ✅ **Frame Size Statistics** - Average, min, max frame sizes
- ✅ **Frame Type Analysis** - Detailed frame type breakdown
- ✅ **Compression Efficiency** - Frame size variation analysis

### Audio Analysis Enhancement
```json
"audio_analysis": {
  "bit_rate_mode": "CBR"
}
```

**Enhanced Audio Checks:**
- ✅ **Audio Bit Rate Mode** - CBR/VBR detection for audio streams
- ✅ **Multi-stream Analysis** - Individual audio stream validation

---

## 🎯 Content Analysis (Advanced)

Content analysis uses FFmpeg filters to detect quality issues and broadcast compliance problems.

### Blackness Detection
```json
"black_frames": {
  "detected_frames": 0,
  "percentage": 0.0,
  "threshold": 0.1
}
```
- **Purpose**: Detect black or nearly black frames
- **Method**: FFmpeg `blackdetect` filter
- **Threshold**: Configurable blackness sensitivity
- **Use Case**: Detect transmission problems, editing errors

### Freeze Frame Detection
```json
"freeze_frames": {
  "detected_frames": 2,
  "percentage": 0.05,
  "threshold": 0.001
}
```
- **Purpose**: Detect static/frozen video sections
- **Method**: FFmpeg `freezedetect` filter
- **Threshold**: Motion detection sensitivity
- **Use Case**: Identify encoding problems, source issues

### Audio Clipping Detection
```json
"audio_clipping": {
  "clipped_samples": 0,
  "percentage": 0.0,
  "peak_level_db": -1.2
}
```
- **Purpose**: Detect audio distortion and clipping
- **Method**: FFmpeg `astats` filter with peak analysis
- **Threshold**: Peak level detection above -1dB
- **Use Case**: Audio quality validation, loudness compliance

### Blockiness Analysis
```json
"blockiness": {
  "average_blockiness": 0.05,
  "max_blockiness": 0.15,
  "threshold": 0.1
}
```
- **Purpose**: Measure compression blockiness artifacts
- **Method**: FFmpeg `blockdetect` filter
- **Threshold**: Configurable blockiness sensitivity
- **Use Case**: Compression quality assessment

### Blurriness Analysis
```json
"blurriness": {
  "average_sharpness": 65.2,
  "min_sharpness": 45.8,
  "blur_detected": false
}
```
- **Purpose**: Measure image sharpness and blur
- **Method**: Edge detection with convolution filters
- **Threshold**: Sharpness score analysis
- **Use Case**: Focus quality, encoding sharpness validation

### Interlacing Detection
```json
"interlace_info": {
  "interlace_detected": false,
  "progressive_frames": 3580,
  "interlaced_frames": 20,
  "confidence": 0.006
}
```
- **Purpose**: Detect interlacing artifacts and field order issues
- **Method**: FFmpeg `idet` filter
- **Analysis**: Progressive vs interlaced frame counting
- **Use Case**: Deinterlacing validation, source format detection

### Noise Analysis
```json
"noise_level": {
  "average_noise": 0.8,
  "max_noise": 2.1,
  "noise_profile": "detected"
}
```
- **Purpose**: Measure video noise levels
- **Method**: FFmpeg `signalstats` filter
- **Analysis**: Luminance difference analysis
- **Use Case**: Source quality assessment, denoising validation

### Broadcast Loudness Compliance
```json
"loudness_meter": {
  "integrated_loudness_lufs": -23.0,
  "loudness_range_lu": 7.2,
  "true_peak_dbtp": -1.0,
  "broadcast_compliant": true,
  "standard": "EBU R128"
}
```
- **Purpose**: Broadcast loudness standards compliance
- **Method**: FFmpeg `ebur128` filter
- **Standards**: EBU R128, ATSC A/85, ITU-R BS.1770
- **Measurements**: 
  - Integrated Loudness (LUFS)
  - Loudness Range (LU)
  - True Peak (dBTP)
- **Use Case**: Broadcast delivery, streaming platform compliance

---

## 🎨 Quality Metrics Integration

### VMAF (Video Multimethod Assessment Fusion)
```json
"quality_metrics": {
  "vmaf": {
    "overall_score": 85.6,
    "min_score": 72.1,
    "max_score": 92.3,
    "mean_score": 85.6,
    "std_deviation": 4.2
  }
}
```
- **Purpose**: Perceptual video quality assessment
- **Method**: Netflix VMAF models
- **Range**: 0-100 (higher is better)
- **Use Case**: Encoding optimization, quality validation

### PSNR (Peak Signal-to-Noise Ratio)
```json
"psnr": {
  "overall_score": 42.3,
  "y_component": 42.8,
  "u_component": 44.1,
  "v_component": 43.9
}
```
- **Purpose**: Objective quality measurement
- **Method**: Mathematical signal comparison
- **Unit**: Decibels (dB)
- **Use Case**: Technical quality validation

### SSIM (Structural Similarity Index)
```json
"ssim": {
  "overall_score": 0.95,
  "y_component": 0.96,
  "u_component": 0.94,
  "v_component": 0.95
}
```
- **Purpose**: Perceptual similarity measurement
- **Method**: Structural comparison algorithm
- **Range**: 0-1 (closer to 1 is better)
- **Use Case**: Compression quality assessment

---

## 🚀 Usage Examples

### Basic Analysis
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4"
  }'
```

### Enhanced Analysis with Content Checks
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4",
    "content_analysis": true,
    "generate_reports": true,
    "report_formats": ["json", "pdf"]
  }'
```

### Async Processing for Large Files
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/large_video.mp4",
    "content_analysis": true,
    "async": true
  }'
```

---

## 📈 Performance Considerations

### Standard Analysis Performance
- **Speed**: ~30 seconds for typical video files
- **Resource Usage**: Low CPU and memory usage
- **Scalability**: Highly concurrent processing

### Enhanced Analysis Performance
- **Speed**: +60-180 seconds additional processing time
- **Resource Usage**: Higher CPU usage for filter processing
- **Scalability**: Moderate concurrency due to FFmpeg filter overhead
- **Recommendation**: Use async processing for enhanced analysis

### Content Analysis Performance
- **Speed**: Variable based on video duration and complexity
- **Resource Usage**: Intensive CPU usage for content filters
- **Memory**: Streaming processing to minimize memory usage
- **Optimization**: Parallel filter execution where possible

---

## 🔧 Configuration Options

### Analysis Precision
```bash
# Environment variables for enhanced analysis
ENABLE_ENHANCED_ANALYSIS=true
CONTENT_ANALYSIS_TIMEOUT=300  # 5 minutes
FRAME_ANALYSIS_LIMIT=1000     # Analyze first 1000 frames for GOP
```

### Quality Thresholds
```bash
# Configurable quality thresholds
BLACKNESS_THRESHOLD=0.1       # 10% threshold for blackness
FREEZE_THRESHOLD=0.001        # Motion detection sensitivity
SHARPNESS_THRESHOLD=50.0      # Blur detection threshold
```

### Performance Tuning
```bash
# Performance optimization
MAX_CONCURRENT_ANALYSIS=5     # Limit concurrent enhanced analysis
FILTER_TIMEOUT=60            # Per-filter timeout in seconds
ENABLE_FILTER_PARALLEL=true  # Parallel filter execution
```

---

## 🎯 Quality Control Standards Compliance

### Broadcast Standards
- **EBU R128**: Loudness measurement and compliance
- **ITU-R BT.709**: Color space and matrix coefficients
- **ITU-R BT.2020**: HDR and wide color gamut support
- **SMPTE**: Professional video standards

### Streaming Platform Compliance
- **Netflix**: VMAF quality scoring
- **YouTube**: Content quality guidelines
- **Amazon Prime**: Technical delivery specifications
- **Hulu**: Quality control requirements

### Professional Standards
- **ATSC**: Advanced Television Systems Committee
- **DVB**: Digital Video Broadcasting
- **ISDB**: Integrated Services Digital Broadcasting

---

## 📋 Future Enhancements

### Planned Features
- **HDR Metadata Analysis**: HDR10, Dolby Vision validation
- **Professional Format Support**: IMF, MXF compliance
- **Advanced AI Quality**: Machine learning quality assessment
- **Real-time Analysis**: Live stream quality monitoring

### Integration Roadmap
- **Quality Dashboards**: Real-time quality metrics visualization
- **Automated Workflows**: Quality-based processing decisions
- **Alert Systems**: Quality threshold violation notifications
- **Compliance Reporting**: Automated standards compliance reports

---

*Last updated: $(date)*  
*For technical support and questions, see [README.md](../README.md#-support--contact)*