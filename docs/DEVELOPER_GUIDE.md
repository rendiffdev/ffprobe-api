# Rendiff Probe Developer Guide

**Comprehensive documentation for developers contributing to or extending Rendiff Probe**

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Project Structure](#project-structure)
3. [Development Environment](#development-environment)
4. [Building the Project](#building-the-project)
5. [Code Architecture](#code-architecture)
6. [Adding New Features](#adding-new-features)
7. [Testing](#testing)
8. [Code Style and Standards](#code-style-and-standards)
9. [Debugging](#debugging)
10. [Performance Optimization](#performance-optimization)
11. [Security Considerations](#security-considerations)
12. [Contributing](#contributing)

---

## Getting Started

### Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.24+ | Core language |
| FFmpeg/FFprobe | 6.0+ | Media analysis engine |
| Docker | 24.0+ | Containerization |
| Docker Compose | 2.0+ | Multi-container orchestration |
| Make | Any | Build automation |
| Git | 2.0+ | Version control |

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/rendiffdev/rendiff-probe.git
cd rendiff-probe

# Install dependencies
go mod download

# Verify FFprobe is installed
ffprobe -version

# Run tests to verify setup
go test ./...

# Build the binaries
make build
```

---

## Project Structure

```
rendiff-probe/
├── cmd/                          # Application entry points
│   ├── rendiff-probe/           # REST API server
│   │   └── main.go
│   └── rendiffprobe-cli/        # CLI tool
│       └── main.go
├── internal/                     # Private application code
│   ├── api/                     # API definitions and routing
│   │   ├── router.go           # Gin router setup
│   │   ├── routes.go           # Route definitions
│   │   └── graphql/            # GraphQL schema and resolvers
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── db/                      # Database layer
│   │   ├── sqlite.go           # SQLite implementation
│   │   └── migrations/         # Database migrations
│   ├── ffmpeg/                  # FFmpeg integration layer
│   │   ├── probe.go            # FFprobe wrapper
│   │   ├── content_analyzer.go # 26 parallel analyzers
│   │   ├── enhanced_analyzer.go# Enhanced QC analysis
│   │   ├── hdr_analyzer.go     # HDR analysis
│   │   └── hls_analyzer.go     # HLS stream analysis
│   ├── handlers/                # HTTP request handlers
│   │   ├── probe.go            # File/URL analysis
│   │   ├── batch.go            # Batch processing
│   │   ├── hls.go              # HLS endpoint
│   │   ├── graphql.go          # GraphQL handler
│   │   └── websocket.go        # WebSocket handler
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go             # Authentication
│   │   ├── ratelimit.go        # Rate limiting
│   │   ├── cors.go             # CORS handling
│   │   └── logging.go          # Request logging
│   ├── models/                  # Data models
│   │   ├── analysis.go         # Analysis result models
│   │   ├── request.go          # Request models
│   │   └── response.go         # Response models
│   └── services/                # Business logic layer
│       ├── analysis.go         # Analysis service
│       ├── report.go           # Report generation
│       └── secret_rotation.go  # Secret management
├── docker-image/                # Docker configuration
│   ├── Dockerfile
│   ├── compose.yaml
│   └── build-docker.sh
├── docs/                        # Documentation
│   ├── ARCHITECTURE.md
│   ├── DEVELOPER_GUIDE.md
│   ├── USER_MANUAL.md
│   └── api/                    # API documentation
├── scripts/                     # Build and utility scripts
├── test/                        # Integration tests
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── Makefile                     # Build automation
└── README.md                    # Project overview
```

---

## Development Environment

### Environment Variables

```bash
# Core Configuration
export PORT=8080                           # API server port
export LOG_LEVEL=debug                     # Log level (debug|info|warn|error)
export GIN_MODE=debug                      # Gin framework mode

# FFmpeg Configuration
export FFPROBE_PATH=/usr/local/bin/ffprobe # Path to FFprobe binary
export FFMPEG_PATH=/usr/local/bin/ffmpeg   # Path to FFmpeg binary

# Database Configuration
export DB_PATH=./data/rendiff-probe.db     # SQLite database path
export CLOUD_MODE=false                    # Enable cloud features

# Cache Configuration
export VALKEY_URL=localhost:6379           # Valkey/Redis URL
export CACHE_TTL=3600                      # Cache TTL in seconds

# Authentication (Production)
export JWT_SECRET=your-secret-key          # JWT signing secret
export API_KEY=your-api-key                # API key for authentication

# Rate Limiting
export RATE_LIMIT_RPM=60                   # Requests per minute
export RATE_LIMIT_RPH=1000                 # Requests per hour
export RATE_LIMIT_RPD=10000                # Requests per day
```

### IDE Setup

#### VS Code

Recommended extensions:
- Go (golang.go)
- Docker (ms-azuretools.vscode-docker)
- REST Client (humao.rest-client)
- GitLens (eamodio.gitlens)

`.vscode/settings.json`:
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-v", "-race"],
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

#### GoLand

- Enable Go modules integration
- Set GOROOT and GOPATH correctly
- Configure golangci-lint as external tool

---

## Building the Project

### Build Commands

```bash
# Build all binaries
make build

# Build API server only
go build -o rendiff-probe ./cmd/rendiff-probe

# Build CLI tool only
go build -o rendiffprobe-cli ./cmd/rendiffprobe-cli

# Build with race detection (for testing)
go build -race -o rendiff-probe-race ./cmd/rendiff-probe

# Build for production (with optimizations)
CGO_ENABLED=1 go build -ldflags="-s -w" -o rendiff-probe ./cmd/rendiff-probe

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o rendiff-probe-linux ./cmd/rendiff-probe
```

### Docker Build

```bash
# Build Docker image
make docker-build

# Or manually
docker build -f docker-image/Dockerfile -t rendiff-probe:latest .

# Build with specific tag
docker build -f docker-image/Dockerfile -t rendiff-probe:v2.0.0 .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t rendiff-probe:latest .
```

---

## Code Architecture

### Layer Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        HANDLERS                              │
│  Receive HTTP requests, validate input, call services        │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                        SERVICES                              │
│  Business logic, orchestration, data transformation          │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                    ANALYSIS ENGINE                           │
│  FFprobe/FFmpeg integration, concurrent analysis             │
└────────────────────────────┬────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                      DATA LAYER                              │
│  SQLite, Valkey cache, file storage                          │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

#### 1. Content Analyzer (`internal/ffmpeg/content_analyzer.go`)

The core analysis engine with 26 parallel analyzers:

```go
type ContentAnalyzer struct {
    ffprobePath string
    ffmpegPath  string
    logger      zerolog.Logger
}

// AnalyzeContent runs all 26 analyzers concurrently
func (ca *ContentAnalyzer) AnalyzeContent(ctx context.Context, filePath string) (*ContentAnalysis, error) {
    // Uses WaitGroup for coordination
    // Buffered channels for results
    // Context cancellation for cleanup
}
```

**Concurrency Pattern:**
```go
// Safe goroutine launch pattern
launchAnalyzer := func(name string, analyze func(context.Context, string) (func(), error)) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        select {
        case <-analyzeCtx.Done():
            return
        default:
        }
        result, err := analyze(analyzeCtx, filePath)
        // Context-aware channel send
        select {
        case <-analyzeCtx.Done():
            return
        default:
            if err != nil {
                errorChan <- err
            } else if result != nil {
                resultChan <- result
            }
        }
    }()
}
```

#### 2. Rate Limiter (`internal/middleware/ratelimit.go`)

Multi-window rate limiting with role-based quotas:

```go
type RateLimitMiddleware struct {
    config   RateLimitConfig
    counters *RateCounter
    mu       sync.RWMutex
    logger   zerolog.Logger
    done     chan struct{}  // Graceful shutdown
}

// Role-based limits
func (rl *RateLimitMiddleware) getLimitsForRole(c *gin.Context) RoleLimits {
    // admin:   600/min, 10000/hr, 100000/day
    // premium: 300/min, 5000/hr,  50000/day
    // pro:     180/min, 3000/hr,  30000/day
    // user:    60/min,  1000/hr,  10000/day
}
```

#### 3. Batch Handler (`internal/handlers/batch.go`)

Parallel batch processing with status tracking:

```go
type BatchHandler struct {
    analysisService *services.AnalysisService
    logger          zerolog.Logger
}

// Batch job storage with automatic cleanup
var (
    batchStore        = make(map[uuid.UUID]*BatchStatusResponse)
    batchMutex        sync.RWMutex
    batchCleanupOnce  sync.Once
    batchRetentionTTL = 24 * time.Hour
    batchMaxEntries   = 1000
)
```

### Adding a New Analyzer

To add a new content analyzer:

1. **Define the analysis function** in `content_analyzer.go`:

```go
func (ca *ContentAnalyzer) analyzeMyNewMetric(ctx context.Context, filePath string) (*MyNewMetricResult, error) {
    // Build FFmpeg command
    args := []string{
        "-i", filePath,
        "-vf", "your_filter_here",
        "-f", "null", "-",
    }

    // Execute with context
    cmd := exec.CommandContext(ctx, ca.ffmpegPath, args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("my_new_metric analysis failed: %w", err)
    }

    // Parse output
    result := &MyNewMetricResult{}
    // ... parsing logic

    return result, nil
}
```

2. **Add to the analyzer launch** in `AnalyzeContent`:

```go
// In AnalyzeContent function
launchAnalyzer("myNewMetric", func(ctx context.Context, path string) (func(), error) {
    result, err := ca.analyzeMyNewMetric(ctx, path)
    if err != nil {
        return nil, err
    }
    return func() {
        analysis.MyNewMetric = result
    }, nil
})
```

3. **Update the ContentAnalysis struct**:

```go
type ContentAnalysis struct {
    // ... existing fields
    MyNewMetric *MyNewMetricResult `json:"my_new_metric,omitempty"`
}
```

4. **Update the analyzer count** constant:
```go
const numAnalyzers = 27  // Increment from 26
```

---

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run specific package
go test -v ./internal/ffmpeg/...

# Run specific test
go test -v -run TestAnalyzeContent ./internal/ffmpeg/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
go test -bench=. ./internal/ffmpeg/...
```

### Writing Tests

#### Unit Test Example

```go
func TestContentAnalyzer_AnalyzeBlackFrames(t *testing.T) {
    // Setup
    ca := NewContentAnalyzer("/usr/bin/ffprobe", "/usr/bin/ffmpeg", zerolog.Nop())

    // Create test file or use fixture
    testFile := "testdata/video_with_black.mp4"

    // Execute
    ctx := context.Background()
    result, err := ca.analyzeBlackFrames(ctx, testFile)

    // Assert
    require.NoError(t, err)
    assert.True(t, result.HasBlackFrames)
    assert.Greater(t, result.TotalBlackDuration, 0.0)
}
```

#### Integration Test Example

```go
func TestProbeHandler_AnalyzeFile(t *testing.T) {
    // Setup test server
    router := setupTestRouter()

    // Create test request
    body := new(bytes.Buffer)
    writer := multipart.NewWriter(body)
    part, _ := writer.CreateFormFile("file", "test.mp4")
    io.Copy(part, testVideoReader)
    writer.Close()

    req := httptest.NewRequest("POST", "/api/v1/probe/file", body)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Execute
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)

    var response ProbeResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.NotEmpty(t, response.AnalysisID)
}
```

### Test Fixtures

Place test media files in `testdata/` directory:
```
testdata/
├── video_1080p.mp4
├── video_4k_hdr.mp4
├── video_with_black.mp4
├── audio_stereo.wav
├── hls_manifest.m3u8
└── corrupted.mp4
```

---

## Code Style and Standards

### Go Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `golangci-lint` for linting

```bash
# Format code
gofmt -w .

# Run linter
golangci-lint run

# Fix common issues
golangci-lint run --fix
```

### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Package | lowercase | `ffmpeg`, `handlers` |
| Exported function | PascalCase | `AnalyzeContent` |
| Private function | camelCase | `parseOutput` |
| Constants | PascalCase or SCREAMING_CASE | `MaxTimeout`, `DEFAULT_PORT` |
| Interface | PascalCase, -er suffix | `Analyzer`, `Reader` |
| Struct | PascalCase | `ContentAnalysis` |

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to analyze content: %w", err)
}

// Use custom error types for specific cases
type AnalysisError struct {
    Code    string
    Message string
    Cause   error
}

func (e *AnalysisError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
```

### Logging

Use structured logging with zerolog:

```go
// Good
logger.Info().
    Str("file", filePath).
    Int("duration_ms", duration).
    Msg("Analysis completed")

// Avoid
logger.Info().Msgf("Analysis completed for %s in %d ms", filePath, duration)
```

---

## Debugging

### Debug Mode

```bash
# Enable debug logging
export LOG_LEVEL=debug
export GIN_MODE=debug

# Run with verbose output
./rendiff-probe 2>&1 | tee debug.log
```

### Profiling

```go
import _ "net/http/pprof"

// In main.go, add pprof endpoints
go func() {
    http.ListenAndServe("localhost:6060", nil)
}()
```

```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Goroutine leak | Missing cleanup | Use `defer cancel()` with context |
| Race condition | Concurrent map access | Use sync.RWMutex |
| Memory spike | Large file processing | Use streaming/buffered I/O |
| FFprobe timeout | Long analysis | Increase timeout, optimize filters |

---

## Performance Optimization

### Concurrency Best Practices

```go
// Use buffered channels to prevent blocking
resultChan := make(chan Result, numWorkers)

// Use context for cancellation
ctx, cancel := context.WithTimeout(parent, 120*time.Second)
defer cancel()

// Use WaitGroup for coordination
var wg sync.WaitGroup
wg.Add(numWorkers)
go func() {
    wg.Wait()
    close(resultChan)
}()
```

### Memory Efficiency

```go
// Use bufio.Scanner instead of strings.Split for large outputs
func forEachLine(output []byte, fn func(line string) bool) {
    scanner := bufio.NewScanner(bytes.NewReader(output))
    for scanner.Scan() {
        if !fn(scanner.Text()) {
            return
        }
    }
}

// Reuse buffers where possible
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}
```

### Caching Strategy

```go
// Use Valkey/Redis for frequently accessed data
func (s *Service) GetCachedResult(key string) (*Result, error) {
    // Check cache first
    cached, err := s.cache.Get(key)
    if err == nil {
        return cached, nil
    }

    // Compute result
    result, err := s.computeResult()
    if err != nil {
        return nil, err
    }

    // Store in cache
    s.cache.Set(key, result, s.cacheTTL)
    return result, nil
}
```

---

## Security Considerations

### Input Validation

```go
// Validate file paths to prevent traversal
func validatePath(path string) error {
    clean := filepath.Clean(path)
    if strings.Contains(clean, "..") {
        return errors.New("path traversal detected")
    }
    return nil
}

// Validate URLs
func validateURL(rawURL string) error {
    u, err := url.Parse(rawURL)
    if err != nil {
        return err
    }
    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("invalid URL scheme")
    }
    return nil
}
```

### Command Injection Prevention

```go
// Use exec.CommandContext with explicit arguments
cmd := exec.CommandContext(ctx, "ffprobe",
    "-v", "quiet",
    "-print_format", "json",
    "-show_format",
    "-show_streams",
    filePath,  // Never use shell interpolation
)

// Never do this
// cmd := exec.Command("sh", "-c", "ffprobe " + filePath)
```

### Secret Management

```go
// Use environment variables for secrets
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    log.Fatal("JWT_SECRET environment variable required")
}

// Never log secrets
logger.Info().
    Str("user", userID).
    // Str("token", token) // NEVER log tokens
    Msg("User authenticated")
```

---

## Contributing

### Pull Request Process

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/rendiff-probe.git
   cd rendiff-probe
   git remote add upstream https://github.com/rendiffdev/rendiff-probe.git
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Changes**
   - Follow code style guidelines
   - Add tests for new functionality
   - Update documentation as needed

4. **Run Tests**
   ```bash
   go test -race ./...
   golangci-lint run
   ```

5. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

6. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a Pull Request on GitHub.

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style (formatting)
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

Examples:
```
feat(analyzer): add PSE flash detection

fix(ratelimit): resolve race condition in counter increment

docs(readme): update installation instructions
```

### Code Review Checklist

- [ ] Code follows project style guidelines
- [ ] Tests pass with race detector enabled
- [ ] New functionality has test coverage
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced
- [ ] Performance impact is acceptable
- [ ] Error handling is appropriate
- [ ] Logging is appropriate (no sensitive data)

---

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md)
- [User Manual](USER_MANUAL.md)
- [QC Analysis List](QC_ANALYSIS_LIST.md)
- [API Reference](api/)
- [Changelog](../CHANGELOG.md)

---

*For questions or support, please open an issue on GitHub.*
