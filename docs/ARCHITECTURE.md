# Rendiff Probe Architecture

## Overview

Rendiff Probe is a professional video analysis platform built on a clean, layered architecture designed for scalability, maintainability, and production-grade reliability. The system provides both REST API and CLI interfaces for comprehensive media file analysis using FFprobe and FFmpeg.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                        │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐                  │
│  │  REST API       │  │  GraphQL API    │  │  CLI Tool       │                  │
│  │  (HTTP/JSON)    │  │  (Queries)      │  │  (rendiffprobe) │                  │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘                  │
└───────────┼─────────────────────┼─────────────────────┼──────────────────────────┘
            │                     │                     │
┌───────────▼─────────────────────▼─────────────────────▼──────────────────────────┐
│                            API GATEWAY LAYER                                      │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐             │
│  │ Gin Router   │ │ Auth/JWT     │ │ Rate Limiter │ │ CORS/Headers │             │
│  │ Middleware   │ │ Middleware   │ │ Middleware   │ │ Middleware   │             │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘             │
└───────────────────────────────────┬──────────────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────────────┐
│                             HANDLER LAYER                                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │ Probe       │ │ Batch       │ │ HLS         │ │ GraphQL     │ │ WebSocket   │ │
│  │ Handler     │ │ Handler     │ │ Handler     │ │ Handler     │ │ Handler     │ │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ │
└─────────┼───────────────┼───────────────┼───────────────┼───────────────┼────────┘
          │               │               │               │               │
┌─────────▼───────────────▼───────────────▼───────────────▼───────────────▼────────┐
│                            SERVICE LAYER                                          │
│  ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐                  │
│  │ Analysis Service │ │ Report Service   │ │ Secret Rotation  │                  │
│  │ - CreateAnalysis │ │ - PDF Generation │ │ - API Key Mgmt   │                  │
│  │ - ProcessAnalysis│ │ - JSON Export    │ │ - JWT Rotation   │                  │
│  └────────┬─────────┘ └────────┬─────────┘ └──────────────────┘                  │
└───────────┼─────────────────────┼────────────────────────────────────────────────┘
            │                     │
┌───────────▼─────────────────────▼────────────────────────────────────────────────┐
│                            ANALYSIS ENGINE                                        │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                        FFmpeg Integration Layer                             │  │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │  │
│  │  │ FFprobe     │ │ Content     │ │ HDR         │ │ Enhanced    │          │  │
│  │  │ Runner      │ │ Analyzer    │ │ Analyzer    │ │ Analyzers   │          │  │
│  │  │             │ │ (26 parallel│ │             │ │ (19 QC      │          │  │
│  │  │             │ │ goroutines) │ │             │ │ categories) │          │  │
│  │  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘          │  │
│  └─────────┼───────────────┼───────────────┼───────────────┼─────────────────┘  │
│            │               │               │               │                      │
│  ┌─────────▼───────────────▼───────────────▼───────────────▼─────────────────┐  │
│  │                      FFmpeg/FFprobe Binaries                               │  │
│  │  (signalstats, idet, ebur128, astats, blackdetect, freezedetect, etc.)    │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────┘
                                    │
┌───────────────────────────────────▼──────────────────────────────────────────────┐
│                            DATA LAYER                                             │
│  ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐                  │
│  │ SQLite Database  │ │ Valkey/Redis     │ │ File Storage     │                  │
│  │ - Analysis Store │ │ - Rate Limiting  │ │ - Temp Files     │                  │
│  │ - Job Queue      │ │ - API Key Cache  │ │ - Uploads        │                  │
│  │ - Audit Logs     │ │ - Session Cache  │ │ - Reports        │                  │
│  └──────────────────┘ └──────────────────┘ └──────────────────┘                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Client Layer

#### REST API (`rendiff-probe`)
- **Technology**: Gin Web Framework
- **Protocol**: HTTP/HTTPS
- **Formats**: JSON, multipart/form-data
- **Endpoints**:
  - `POST /api/v1/probe/file` - Analyze uploaded file
  - `POST /api/v1/probe/url` - Analyze URL
  - `POST /api/v1/probe/hls` - Analyze HLS stream
  - `POST /api/v1/batch/analyze` - Batch analysis
  - `GET /api/v1/batch/status/:id` - Batch status

#### GraphQL API
- **Schema**: Type-safe queries and mutations
- **Playground**: Interactive GraphiQL interface
- **Features**: Flexible field selection, nested queries

#### CLI Tool (`rendiffprobe-cli`)
- **Commands**: `analyze`, `info`, `categories`, `version`
- **Formats**: `report`, `json`, `text`
- **Features**: Local analysis without server

### 2. API Gateway Layer

#### Middleware Stack
```go
Middleware Pipeline:
├── Recovery (panic handling)
├── RequestID (tracing)
├── Logging (structured logs)
├── CORS (cross-origin)
├── RateLimit (per-user/tenant)
├── Authentication (JWT/API Key)
└── Validation (request validation)
```

#### Rate Limiting
- Per-minute, per-hour, per-day windows
- Tenant-aware limits
- Role-based quotas (admin, premium, pro, user)
- Graceful cleanup with TTL-based expiration

#### Authentication
- JWT tokens with rotation support
- API key authentication
- Multi-tenant isolation
- RBAC (Role-Based Access Control)

### 3. Handler Layer

Handlers translate HTTP requests into service calls:

| Handler | Responsibility |
|---------|---------------|
| `ProbeHandler` | Single file/URL analysis |
| `BatchHandler` | Parallel batch processing |
| `HLSHandler` | HLS stream analysis |
| `GraphQLHandler` | GraphQL query resolution |
| `WebSocketHandler` | Real-time progress updates |
| `APIKeyHandler` | API key management |

### 4. Service Layer

Business logic encapsulation:

```go
Service Layer:
├── AnalysisService
│   ├── CreateAnalysis()
│   ├── ProcessAnalysis()
│   ├── GetAnalysis()
│   └── ListAnalyses()
├── ReportService
│   ├── GeneratePDFReport()
│   ├── GenerateJSONReport()
│   └── ExportAnalysis()
├── SecretRotationService
│   ├── GenerateAPIKey()
│   ├── RotateAPIKey()
│   ├── RotateJWTSecret()
│   └── CleanupExpiredKeys()
└── VMAFModelService
    └── Quality assessment models
```

### 5. Analysis Engine

The core of Rendiff Probe - a sophisticated FFmpeg integration layer.

#### FFprobe Runner
- Process management with context cancellation
- Structured output parsing (JSON format)
- Timeout handling
- Error recovery

#### Content Analyzer (26 Parallel Analyzers)
All analyzers run concurrently using goroutines with proper cleanup:

```go
Content Analyzers:
├── Video Quality
│   ├── Baseband Analysis (signalstats)
│   ├── Video Quality Score
│   ├── Blockiness Detection
│   ├── Blurriness Analysis
│   ├── Noise Analysis
│   └── Line Error Detection
├── Video Content
│   ├── Black Frame Detection (blackdetect)
│   ├── Freeze Frame Detection (freezedetect)
│   ├── Letterbox Detection (cropdetect)
│   ├── Color Bars Detection
│   ├── Safe Area Analysis
│   ├── Temporal Complexity
│   ├── Field Dominance (idet)
│   ├── Differential Frames
│   └── Interlace Analysis
├── Audio Analysis
│   ├── Loudness Metering (ebur128)
│   ├── Audio Clipping (astats)
│   ├── Silence Detection (silencedetect)
│   ├── Phase Correlation (aphasemeter)
│   ├── Channel Mapping
│   ├── Audio Frequency
│   └── Test Tone Detection
└── Additional
    ├── HDR Analysis
    ├── Timecode Continuity
    └── Dropout Detection
```

#### Enhanced Analyzers (19 QC Categories)

| Category | Standards | Purpose |
|----------|-----------|---------|
| AFD Analysis | ITU-R BT.1868 | Active Format Description |
| Dead Pixel Detection | Computer Vision | Camera QC |
| PSE Flash Analysis | ITC/Ofcom, ITU-R BT.1702 | Epilepsy safety |
| HDR Analysis | HDR10, Dolby Vision, HLG | HDR validation |
| Audio Wrapping | BWF, RF64, AES3 | Professional audio |
| Endianness Detection | - | Cross-platform |
| Codec Analysis | - | Format validation |
| Container Validation | MP4, MKV, MOV | Structure analysis |
| Resolution Analysis | - | Display optimization |
| Frame Rate Analysis | Broadcast standards | Temporal accuracy |
| Bitdepth Analysis | 8/10/12-bit | Color depth |
| Timecode Analysis | SMPTE 12M | TC continuity |
| MXF Analysis | SMPTE ST 377 | Broadcast format |
| IMF Compliance | SMPTE ST 2067 | Distribution |
| Transport Stream | MPEG-TS | Broadcast transmission |
| Content Analysis | Multiple | 26 sub-analyzers |
| Enhanced Analysis | - | Quality metrics |
| Stream Disposition | Section 508 | Accessibility |
| Data Integrity | CRC32, MD5 | Error detection |

### 6. Data Layer

#### SQLite Database
- Embedded, zero-configuration
- Analysis storage with full-text search
- Migration support
- Transaction handling

#### Valkey/Redis Cache
- Rate limit counters
- API key validation cache
- Session storage
- Result caching

#### File Storage
- Temporary upload handling
- Report output storage
- Configurable paths

## Concurrency Model

### Goroutine Management

The system uses a structured concurrency model:

```go
// Pattern: WaitGroup + Context + Buffered Channels
func (ca *ContentAnalyzer) AnalyzeContent(ctx context.Context, filePath string) {
    analyzeCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
    defer cancel()

    var wg sync.WaitGroup
    resultChan := make(chan func(), numAnalyzers)
    errorChan := make(chan error, numAnalyzers)

    // Launch analyzers with proper cleanup
    launchAnalyzer := func(name string, analyze func(context.Context, string) (func(), error)) {
        wg.Add(1)
        go func() {
            defer wg.Done()
            select {
            case <-analyzeCtx.Done():
                return
            default:
            }
            // ... analysis logic
        }()
    }

    // Close channels when done
    go func() {
        wg.Wait()
        close(resultChan)
        close(errorChan)
    }()
}
```

### Thread Safety

- Mutex protection for shared state
- Read-write locks where appropriate
- Lock ordering to prevent deadlocks
- Channel-based communication

## Request Flow

```
1. Client Request
       │
       ▼
2. Gin Router
       │
       ▼
3. Middleware Chain
   ├── Recovery
   ├── RequestID
   ├── Logging
   ├── RateLimit
   └── Auth
       │
       ▼
4. Handler
   └── Request validation
       │
       ▼
5. Service
   └── Business logic
       │
       ▼
6. FFmpeg Runner
   └── Execute ffprobe/ffmpeg
       │
       ▼
7. Content Analyzers (parallel)
   └── 26 concurrent analyses
       │
       ▼
8. Result Aggregation
       │
       ▼
9. Response
```

## Deployment Architecture

### Docker Deployment

```yaml
services:
  rendiff-probe:
    image: rendiff-probe:latest
    environment:
      - PORT=8080
      - DB_PATH=/data/rendiff.db
      - VALKEY_URL=valkey:6379
    volumes:
      - ./data:/data
    ports:
      - "8080:8080"
    depends_on:
      - valkey

  valkey:
    image: valkey/valkey:latest
    volumes:
      - valkey-data:/data
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rendiff-probe
spec:
  replicas: 3
  selector:
    matchLabels:
      app: rendiff-probe
  template:
    spec:
      containers:
      - name: rendiff-probe
        image: rendiff-probe:latest
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
```

## Performance Characteristics

| Operation | Typical Time | Notes |
|-----------|--------------|-------|
| Basic Probe | 1-3s | Metadata extraction |
| Standard QC | 10-30s | All content analyzers |
| Full Analysis | 30-60s | All 121 parameters |
| Batch (10 files) | 2-5min | Parallel processing |

## Security Architecture

### Authentication Flow

```
Client → API Key/JWT → Middleware → Handler → Service
                 │
                 ▼
         Rate Limit Check
                 │
                 ▼
         Permission Check
                 │
                 ▼
         Tenant Isolation
```

### Data Protection
- Input validation at all boundaries
- Path traversal prevention
- SQL injection protection
- XSS prevention in reports

## Monitoring & Observability

### Metrics
- Request latency histograms
- Error rates by endpoint
- Analysis completion rates
- QC failure rates by category

### Logging
- Structured JSON logs
- Request/response tracing
- Error stack traces
- Audit logging

### Health Checks
- `/health` - Service health
- Database connectivity
- FFprobe availability
- Cache connectivity

## Error Handling

### Error Categories
```go
ErrorCode:
├── BadRequest (400)
│   ├── InvalidInput
│   └── ValidationError
├── Unauthorized (401)
│   └── AuthenticationFailed
├── Forbidden (403)
│   └── PermissionDenied
├── NotFound (404)
│   └── ResourceNotFound
├── TooManyRequests (429)
│   └── RateLimitExceeded
└── InternalError (500)
    ├── AnalysisFailed
    └── SystemError
```

### Graceful Degradation
- QC failures don't prevent basic analysis
- Partial results returned with clear status
- Circuit breakers for external services
- Fallback to basic analysis when advanced features unavailable

## Future Considerations

### Scalability
- Horizontal scaling with shared cache
- Worker pool for batch processing
- Message queue for async jobs

### Features
- VMAF quality assessment
- Machine learning insights
- Real-time streaming analysis
- Multi-region deployment

---

*This architecture is designed for professional video analysis workflows in broadcast, streaming, and post-production environments.*
