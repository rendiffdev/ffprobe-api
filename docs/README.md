# FFprobe API Documentation

Welcome to the documentation for **FFprobe API** - a professional video analysis platform with 19 quality control categories.

## Documentation Index

### Getting Started
- **[Main README](../README.md)** - Project overview and quick start
- **[Quick Start Guide](../README.md#quick-start)** - Get running in minutes

### API Reference
- **[REST API Documentation](api/README.md)** - Complete endpoint reference
- **[QC Analysis List](QC_ANALYSIS_LIST.md)** - All 19 quality control categories

### Deployment
- **[Docker Production Guide](../docker-image/README-DOCKER-PRODUCTION.md)** - Production deployment

### Project Information
- **[Changelog](../CHANGELOG.md)** - Version history
- **[TODO](../TODO.md)** - Roadmap and planned features

---

## Quick Navigation

### For Developers
Start with [API Reference](api/README.md) to understand available endpoints.

### For DevOps
Start with [Docker Production Guide](../docker-image/README-DOCKER-PRODUCTION.md) for deployment.

### For Video Engineers
Start with [QC Analysis List](QC_ANALYSIS_LIST.md) to understand analysis capabilities.

---

## Key Features

### Professional Quality Control
- **19 QC Categories**: Comprehensive broadcast quality analysis
- **Industry Standards**: SMPTE, ITU compliance validation
- **Latest FFmpeg**: BtbN builds with all codecs

### API Endpoints
| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Service health check |
| `/api/v1/probe/file` | POST | Analyze uploaded file |
| `/api/v1/probe/url` | POST | Analyze file from URL |
| `/api/v1/probe/hls` | POST | Analyze HLS stream |
| `/api/v1/batch/analyze` | POST | Batch processing |
| `/api/v1/batch/status/:id` | GET | Batch job status |
| `/api/v1/ws/progress/:id` | WS | Real-time progress |
| `/api/v1/graphql` | POST/GET | GraphQL API |
| `/admin/ffmpeg/version` | GET | FFmpeg version info |

### Deployment Options
| Mode | Memory | Best For |
|------|--------|----------|
| Minimal | 2-3GB | Development, testing |
| Quick | 4GB | Demos, quick testing |
| Production | 8GB+ | Production workloads |

---

## Features (v2.0.0)

### Implemented Features

- [x] URL-based file analysis
- [x] HLS stream analysis
- [x] GraphQL endpoint
- [x] Batch processing
- [x] LLM-powered insights
- [x] WebSocket progress streaming

### Planned Features

- [ ] Webhook callbacks for async processing
- [ ] DASH stream analysis
- [ ] File comparison endpoint
- [ ] Custom QC rule definitions

See [TODO.md](../TODO.md) for the complete roadmap.

---

## Getting Help

- **[GitHub Issues](https://github.com/rendiffdev/rendiff-probe/issues)** - Bug reports
- **Documentation** - Check the docs folder for guides

---

## Documentation Structure

```
docs/
├── README.md              # This file
├── QC_ANALYSIS_LIST.md    # QC categories (19)
└── api/
    └── README.md          # API reference
```

---

**Built for the video processing community**
