# Complete QC Analysis Categories

## Overview

The FFprobe API provides **20+ comprehensive QC analysis categories** divided into standard technical analysis and advanced professional broadcast analysis.

## Standard Technical Analysis Categories (11)

### 1. StreamCounts
- Total stream count analysis
- Video/audio/subtitle/data stream enumeration
- Stream type distribution analysis

### 2. VideoAnalysis
- Resolution and aspect ratio analysis
- Frame rate and timing analysis
- Color space and bit depth evaluation
- Codec and profile analysis

### 3. AudioAnalysis
- Channel configuration and layout
- Sample rate and bit depth analysis
- Audio codec and compression analysis
- Dynamic range and level analysis

### 4. GOPAnalysis
- Group of Pictures structure analysis
- I/P/B frame distribution
- GOP size and pattern analysis
- Keyframe interval evaluation

### 5. FrameStatistics
- Frame count and distribution
- Frame size statistics
- Temporal analysis patterns
- Frame type analysis

### 6. ContentAnalysis
- Scene change detection
- Motion analysis
- Complexity measurement
- Visual content characteristics

### 7. BitDepthAnalysis
- Bit depth validation and analysis
- Color depth evaluation
- Precision analysis
- HDR compatibility assessment

### 8. ResolutionAnalysis
- Display and storage resolution
- Scaling factor analysis
- Pixel aspect ratio validation
- Resolution standard compliance

### 9. FrameRateAnalysis
- Frame rate accuracy validation
- Variable frame rate detection
- Temporal consistency analysis
- Broadcast standard compliance

### 10. CodecAnalysis
- Codec identification and validation
- Profile and level analysis
- Compression efficiency evaluation
- Compatibility assessment

### 11. ContainerAnalysis
- Container format validation
- Metadata structure analysis
- Muxing pattern evaluation
- Format compliance checking

## Advanced Professional QC Analyzers (9)

### 1. TimecodeAnalysis
**Professional Use**: Broadcast, post-production workflows
- **SMPTE Timecode Detection**: Identifies embedded timecode tracks
- **Drop Frame Analysis**: Validates drop frame vs non-drop frame timecode
- **Timecode Continuity**: Checks for timecode breaks or discontinuities
- **Frame Rate Correlation**: Validates timecode against actual frame rate
- **Start Timecode Extraction**: Captures initial timecode values

### 2. AFDAnalysis (Active Format Description)
**Professional Use**: Broadcast distribution, multi-platform delivery
- **AFD Code Detection**: Identifies AFD signaling in video streams
- **Aspect Ratio Validation**: Validates AFD compliance with content geometry
- **Broadcast Standards**: ITU-R BT.1868 compliance checking
- **Format Conversion**: AFD preservation across format conversions
- **Display Compatibility**: Multi-device display optimization

### 3. TransportStreamAnalysis
**Professional Use**: Broadcast transmission, IPTV, streaming
- **MPEG-TS Structure**: Analyzes transport stream packet structure
- **PID Mapping**: Identifies and maps Program ID streams
- **PSI/SI Analysis**: Program Specific Information validation
- **Error Detection**: Transport errors, continuity counter issues
- **Bandwidth Analysis**: Bitrate distribution and utilization

### 4. EndiannessAnalysis
**Professional Use**: Cross-platform workflows, archival systems
- **Byte Order Detection**: Big-endian vs little-endian analysis
- **Platform Compatibility**: Cross-platform file compatibility assessment
- **Architecture Validation**: Hardware architecture compatibility
- **Binary Format Analysis**: Container endianness validation
- **Migration Assessment**: Legacy format conversion planning

### 5. AudioWrappingAnalysis
**Professional Use**: Professional audio post-production, broadcast
- **Professional Format Detection**: BWF, RF64, AES3 identification
- **Channel Mapping**: Audio channel layout and routing analysis
- **Embedding Standards**: Audio embedding compliance validation
- **Broadcast Audio**: Professional audio format compliance
- **Surround Sound**: Multi-channel audio configuration analysis

### 6. IMFAnalysis (Interoperable Master Format)
**Professional Use**: Netflix delivery, international distribution
- **Package Validation**: IMF package structure compliance
- **Application Profiles**: AS-02, AS-11 DPP validation
- **Composition Playlist**: CPL structure and asset mapping
- **Essence Validation**: Essence container and metadata compliance
- **Delivery Standards**: Netflix and studio delivery requirements

### 7. MXFAnalysis (Material Exchange Format)
**Professional Use**: Professional broadcast workflows, archive systems
- **Format Validation**: Complete MXF structure analysis
- **Operational Patterns**: OP1a, OP1b, OP2a, OP3a validation
- **Essence Containers**: JPEG 2000, ProRes, DNxHD validation
- **Metadata Compliance**: SMPTE metadata standard validation
- **Broadcast Compliance**: Professional broadcast format requirements

### 8. DeadPixelAnalysis
**Professional Use**: Camera QC, acquisition monitoring, content quality
- **Computer Vision Detection**: Automated pixel defect identification
- **Defect Classification**: Dead, stuck, and hot pixel categorization
- **Spatial Analysis**: Pixel defect clustering and distribution
- **Temporal Tracking**: Frame-to-frame defect consistency
- **Quality Impact**: Assessment of visual quality degradation

### 9. PSEAnalysis (Photosensitive Epilepsy Risk)
**Professional Use**: Broadcast safety compliance, content distribution
- **Flash Detection**: Automated flash pattern analysis
- **Risk Assessment**: ITC/Ofcom/Harding FPA compliance testing
- **Pattern Analysis**: Spatial pattern and contrast change detection
- **Broadcast Compliance**: ITU-R BT.1702, EBU Tech 3253 standards
- **Safety Validation**: Viewer safety and regulatory compliance

## Usage by Industry

### Broadcast Television
- **Required**: TimecodeAnalysis, AFDAnalysis, TransportStreamAnalysis, PSEAnalysis
- **Recommended**: MXFAnalysis, AudioWrappingAnalysis

### Streaming Platforms
- **Required**: TransportStreamAnalysis, PSEAnalysis, DeadPixelAnalysis
- **Recommended**: IMFAnalysis, EndiannessAnalysis

### Post-Production
- **Required**: TimecodeAnalysis, MXFAnalysis, AudioWrappingAnalysis
- **Recommended**: DeadPixelAnalysis, EndiannessAnalysis

### Content Acquisition
- **Required**: DeadPixelAnalysis, TimecodeAnalysis
- **Recommended**: PSEAnalysis, AudioWrappingAnalysis

### Archival/Preservation
- **Required**: EndiannessAnalysis, MXFAnalysis, IMFAnalysis
- **Recommended**: TimecodeAnalysis, AudioWrappingAnalysis

## Compliance Standards

### International Standards
- **ITU-R BT.1702**: Photosensitive epilepsy guidance
- **ITU-R BT.1868**: Active Format Description
- **SMPTE ST 377**: Material Exchange Format
- **SMPTE ST 2067**: Interoperable Master Format

### Regional Broadcast Standards
- **FCC** (USA): Broadcast regulations and PSE guidelines
- **Ofcom** (UK): Broadcasting standards and accessibility
- **EBU Tech 3253** (Europe): PSE analysis guidelines
- **ARIB** (Japan): Digital broadcasting standards

### Industry Delivery Standards
- **Netflix**: IMF delivery specifications and technical requirements
- **DPP** (UK): Digital Production Partnership standards
- **AAF**: Advanced Authoring Format workflows
- **BWF**: Broadcast Wave Format audio standards

## API Integration

### Enable All QC Analysis
```json
{
  "file_path": "/path/to/media.mxf",
  "content_analysis": true,
  "options": {
    "enable_enhanced_analysis": true
  }
}
```

### Selective QC Analysis
```json
{
  "file_path": "/path/to/media.mp4",
  "content_analysis": true,
  "qc_categories": [
    "timecode_analysis",
    "pse_analysis", 
    "dead_pixel_analysis"
  ]
}
```

### Response Structure
```json
{
  "enhanced_analysis": {
    "timecode_analysis": { "has_timecode": true, "is_drop_frame": false },
    "afd_analysis": { "has_afd": true, "afd_codes": ["0010"] },
    "transport_stream_analysis": { "is_mpeg_transport_stream": true },
    "mxf_analysis": { "is_mxf_file": true, "mxf_profile": "OP1a" },
    "pse_analysis": { "pse_risk_level": "safe" },
    "dead_pixel_analysis": { "has_dead_pixels": false }
  }
}
```

## Performance Considerations

### Analysis Complexity
- **Basic QC** (Standard categories): ~2-10 seconds
- **Professional QC** (All categories): ~10-60 seconds  
- **Selective QC** (Chosen categories): Variable based on selection

### Resource Requirements
- **CPU**: Multi-threaded analysis for optimal performance
- **Memory**: Scales with file size and analysis depth
- **Storage**: Temporary space for frame extraction and analysis
- **Network**: Efficient streaming for remote file analysis

---

**Total QC Categories**: 20 (11 Standard + 9 Advanced Professional)  
**Compliance Standards**: 10+ international and regional standards  
**Industry Applications**: Broadcast, Streaming, Post-Production, Archival