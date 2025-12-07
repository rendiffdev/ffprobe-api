# Complete QC Analysis Categories

## Overview

The Rendiff Probe API provides **19 top-level QC categories** with **26 parallel content analyzers** covering **121 industry-standard parameters** for professional video analysis. All analyzers use real FFmpeg filters (signalstats, idet, ebur128, astats, etc.) for accurate broadcast and streaming quality control.

## Content Analysis (26 Parallel Analyzers)

The Content Analysis category runs 26 analyzers in parallel for comprehensive real-time quality assessment:

### Video Quality Analyzers
| Analyzer | FFmpeg Filter | Parameters | Description |
|----------|---------------|------------|-------------|
| Baseband Analysis | signalstats | YMIN, YMAX, YAVG, UMIN, UMAX, VMIN, VMAX, BRNG, TOUT, VREP | Broadcast legal range checking (Y: 16-235, C: 16-240) |
| Video Quality Score | signalstats + entropy | Sharpness, Contrast, Entropy, Color, Noise | Objective quality scoring (0-100) |
| Blockiness Analysis | signalstats | Compression artifact detection | DCT block boundary artifacts |
| Blurriness Analysis | signalstats | Focus/motion blur metrics | Sharpness and edge detection |
| Noise Analysis | signalstats | SNR, temporal noise | Signal-to-noise ratio measurement |
| Line Error Detection | signalstats TOUT/VREP/BRNG | Luma/chroma line errors, DigiBeta errors | Horizontal line defect detection |

### Video Content Analyzers
| Analyzer | FFmpeg Filter | Parameters | Description |
|----------|---------------|------------|-------------|
| Black Frame Detection | blackdetect | Duration, threshold, positions | Dead or black frame identification |
| Freeze Frame Detection | freezedetect | Duration, positions | Static frame detection |
| Letterbox Detection | cropdetect | Aspect ratio, pillarbox/letterbox | Content boundary analysis |
| Color Bars Detection | signalstats | SMPTE bars, EBU bars | Test pattern identification |
| Safe Area Analysis | cropdetect | Title/action safe margins | Broadcast safe area compliance |
| Temporal Complexity | signalstats YDIF + scene | Motion complexity, scene changes | Content complexity metrics |
| Field Dominance | idet | TFF/BFF, interlace detection | Field order and interlace analysis |
| Differential Frames | signalstats YDIF | Frame difference anomalies | Jump cut and flash detection |
| Interlace Analysis | idet | Progressive/interlaced ratio | Scan type detection |

### Audio Analyzers
| Analyzer | FFmpeg Filter | Parameters | Description |
|----------|---------------|------------|-------------|
| Loudness Metering | ebur128 | Integrated, Momentary, Short-term, LRA, True Peak | EBU R128 compliance |
| Audio Clipping | astats | Peak levels, clip count | Digital clipping detection |
| Silence Detection | silencedetect | Duration, positions | Mute/silence identification |
| Phase Correlation | aphasemeter | L/R phase, mono compatibility | Stereo phase analysis |
| Channel Mapping | astats | Channel layout, routing | Multi-channel configuration |
| Audio Frequency | astats | Spectrum analysis, anomalies | Frequency range analysis |
| Test Tone Detection | astats | 1kHz tone, calibration signals | Audio test pattern detection |

### Additional Analyzers
| Analyzer | FFmpeg Filter | Parameters | Description |
|----------|---------------|------------|-------------|
| HDR Analysis | signalstats + metadata | HDR10, Dolby Vision, HLG, MaxCLL, MaxFALL | High dynamic range validation |
| Timecode Continuity | metadata | SMPTE timecode, discontinuities | Timecode stream analysis |
| Dropout Detection | signalstats | Signal loss, corruption | Video signal dropout detection |

---

## Top-Level QC Categories (19)

The following categories are currently implemented and validated:

### 1. AFD Analysis
**Professional Use**: Broadcast distribution, multi-platform delivery
- **AFD Code Detection**: Identifies AFD signaling in video streams
- **Aspect Ratio Validation**: Validates AFD compliance with content geometry
- **Broadcast Standards**: ITU-R BT.1868 compliance checking
- **Display Compatibility**: Multi-device display optimization

### 2. Dead Pixel Detection
**Professional Use**: Camera QC, acquisition monitoring, content quality
- **Computer Vision Detection**: Automated pixel defect identification
- **Defect Classification**: Dead, stuck, and hot pixel categorization
- **Quality Impact Assessment**: Visual quality degradation analysis

### 3. PSE Flash Analysis
**Professional Use**: Broadcast safety compliance, content distribution
- **Flash Detection**: Automated flash pattern analysis
- **Risk Assessment**: ITC/Ofcom/Harding FPA compliance testing
- **Broadcast Compliance**: ITU-R BT.1702, EBU Tech 3253 standards
- **Safety Validation**: Viewer safety and regulatory compliance

### 4. HDR Analysis
**Professional Use**: HDR content validation, streaming platform delivery
- **HDR Standard Detection**: HDR10, Dolby Vision, HLG identification
- **Color Space Validation**: Rec.2020, P3 color gamut analysis
- **Metadata Validation**: HDR metadata compliance checking

### 5. Audio Wrapping Analysis
**Professional Use**: Professional audio post-production, broadcast
- **Professional Format Detection**: BWF, RF64, AES3 identification
- **Channel Mapping**: Audio channel layout and routing analysis
- **Embedding Standards**: Audio embedding compliance validation

### 6. Endianness Detection
**Professional Use**: Cross-platform workflows, archival systems
- **Byte Order Detection**: Big-endian vs little-endian analysis
- **Platform Compatibility**: Cross-platform file compatibility assessment
- **Architecture Validation**: Hardware architecture compatibility

### 7. Codec Analysis
**Professional Use**: Format validation, compression analysis
- **Codec Identification**: Codec validation and profile analysis
- **Compression Efficiency**: Quality vs bitrate evaluation
- **Compatibility Assessment**: Platform and device compatibility

### 8. Container Validation
**Professional Use**: Format compliance, workflow compatibility
- **Container Format Validation**: MP4, MKV, AVI structure analysis
- **Metadata Structure**: Container metadata compliance
- **Muxing Pattern Evaluation**: Stream interleaving analysis

### 9. Resolution Analysis
**Professional Use**: Display optimization, quality validation
- **Display Resolution**: Storage vs display resolution analysis
- **Aspect Ratio Validation**: PAR/DAR compatibility checking
- **Resolution Standards**: Format compliance validation

### 10. Frame Rate Analysis
**Professional Use**: Temporal analysis, broadcast compliance
- **Frame Rate Accuracy**: Temporal consistency validation
- **Variable Frame Rate Detection**: VFR pattern analysis
- **Broadcast Standards**: Frame rate compliance checking

### 11. Bitdepth Analysis
**Professional Use**: Color depth validation, HDR compatibility
- **Bit Depth Validation**: 8-bit, 10-bit, 12-bit analysis
- **Color Precision**: Dynamic range assessment
- **HDR Compatibility**: High bit depth validation

### 12. Timecode Analysis
**Professional Use**: Broadcast, post-production workflows
- **SMPTE Timecode Detection**: Embedded timecode track identification
- **Drop Frame Analysis**: Drop frame vs non-drop frame validation
- **Timecode Continuity**: Discontinuity detection
- **Frame Rate Correlation**: Timecode accuracy validation

### 13. MXF Analysis
**Professional Use**: Professional broadcast workflows, archive systems
- **Format Validation**: Complete MXF structure analysis
- **Operational Patterns**: OP1a, OP1b, OP2a, OP3a validation
- **Essence Containers**: Professional codec validation
- **Metadata Compliance**: SMPTE standard validation

### 14. IMF Compliance
**Professional Use**: Netflix delivery, international distribution
- **Package Validation**: IMF package structure compliance
- **Application Profiles**: AS-02, AS-11 DPP validation
- **Composition Playlist**: CPL structure validation
- **Delivery Standards**: Studio delivery requirements

### 15. Transport Stream Analysis
**Professional Use**: Broadcast transmission, IPTV, streaming
- **MPEG-TS Structure**: Transport stream packet analysis
- **PID Mapping**: Program ID stream identification
- **PSI/SI Analysis**: Program Specific Information validation
- **Error Detection**: Transport errors and continuity issues

### 16. Content Analysis
**Professional Use**: Content characterization, scene analysis
- **Scene Change Detection**: Temporal content transitions
- **Motion Analysis**: Content motion characteristics
- **Complexity Measurement**: Visual complexity assessment

### 17. Enhanced Analysis
**Professional Use**: Advanced quality metrics, AI-powered insights
- **Quality Scoring**: Overall content quality assessment
- **Risk Assessment**: Technical and compliance risk evaluation
- **Workflow Integration**: Pipeline optimization recommendations

### 18. Stream Disposition Analysis
**Professional Use**: Accessibility compliance, multi-language content validation
- **Accessibility Features**: SDH subtitles, audio descriptions, forced subtitles detection
- **Stream Role Analysis**: Main, alternate, commentary, and descriptive stream identification
- **Language Distribution**: Multi-language content validation and compliance
- **ADA Compliance**: Section 508 and WCAG accessibility standards validation
- **Broadcast Standards**: Stream disposition compliance for broadcast delivery

### 19. Data Integrity Analysis
**Professional Use**: File integrity validation, broadcast compliance, quality assurance
- **Error Detection**: Comprehensive format, bitstream, and packet error detection
- **Hash Validation**: CRC32, MD5 data integrity verification
- **Corruption Detection**: Automated file corruption and damage assessment
- **Broadcast Compliance**: Professional broadcast delivery standards validation
- **Quality Scoring**: Overall data integrity scoring (0-100 scale)

## Usage by Industry

### Broadcast Television
- **Required**: Timecode Analysis, AFD Analysis, Transport Stream Analysis, PSE Flash Analysis
- **Recommended**: MXF Analysis, Audio Wrapping Analysis

### Streaming Platforms
- **Required**: Transport Stream Analysis, PSE Flash Analysis, Dead Pixel Detection, HDR Analysis
- **Recommended**: IMF Compliance, Endianness Detection

### Post-Production
- **Required**: Timecode Analysis, MXF Analysis, Audio Wrapping Analysis
- **Recommended**: Dead Pixel Detection, Codec Analysis

### Content Acquisition
- **Required**: Dead Pixel Detection, Timecode Analysis, HDR Analysis
- **Recommended**: PSE Flash Analysis, Audio Wrapping Analysis

### Archival/Preservation
- **Required**: Endianness Detection, MXF Analysis, IMF Compliance
- **Recommended**: Timecode Analysis, Container Validation

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
    "pse_flash_analysis", 
    "dead_pixel_detection",
    "hdr_analysis"
  ]
}
```

### Response Structure
```json
{
  "enhanced_analysis": {
    "afd_analysis": { "has_afd": true, "afd_codes": ["0010"] },
    "dead_pixel_detection": { "has_dead_pixels": false },
    "pse_flash_analysis": { "pse_risk_level": "safe" },
    "hdr_analysis": { "is_hdr_content": true, "hdr_standard": "HDR10" },
    "timecode_analysis": { "has_timecode": true, "is_drop_frame": false },
    "transport_stream_analysis": { "is_mpeg_transport_stream": true },
    "mxf_analysis": { "is_mxf_file": true, "mxf_profile": "OP1a" },
    "enhanced_analysis": { "overall_qc_score": 95.5 }
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

**Total QC Categories**: 19 top-level + 26 content analyzers
**Total Parameters**: 121 industry-standard parameters
**FFmpeg Filters Used**: signalstats, idet, ebur128, astats, blackdetect, freezedetect, cropdetect, silencedetect, aphasemeter, entropy
**Compliance Standards**: EBU R128, ITU-R BS.1770, ITU-R BT.1702, SMPTE 12M/ST 377/ST 2067, Rec.2020
**Industry Applications**: Broadcast, Streaming, Post-Production, Archival, QC Workflows