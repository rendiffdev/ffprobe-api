# FFprobe API - Technical Contributor Guidelines

Welcome to the FFprobe API project! This guide will help you get started as a technical contributor and provide you with everything you need to know about the codebase, development practices, and contribution workflow.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Project Architecture](#project-architecture)
- [Development Environment](#development-environment)
- [Code Standards](#code-standards)
- [Testing Guidelines](#testing-guidelines)
- [Security Guidelines](#security-guidelines)
- [API Design Principles](#api-design-principles)
- [Database Guidelines](#database-guidelines)
- [Performance Considerations](#performance-considerations)
- [Contribution Workflow](#contribution-workflow)
- [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or later
- Docker and Docker Compose
- FFmpeg/FFprobe installed locally (for development)
- PostgreSQL (via Docker)
- Redis (via Docker)

### Getting Started
```bash
# 1. Clone the repository
git clone https://github.com/your-org/ffprobe-api.git
cd ffprobe-api

# 2. Start development environment
docker-compose up -d postgres redis

# 3. Install dependencies
go mod download

# 4. Run database migrations
go run cmd/migrate/main.go up

# 5. Start the development server
go run cmd/ffprobe-api/main.go

# 6. Test the API
curl http://localhost:8080/health
```

### First Contribution Checklist
- [ ] Read this entire document
- [ ] Set up development environment
- [ ] Run existing tests: `go test ./...`
- [ ] Make a small test change and verify it works
- [ ] Check out our [Good First Issues](https://github.com/your-org/ffprobe-api/labels/good%20first%20issue)

## 🏗️ Project Architecture

### Directory Structure
```
ffprobe-api/
├── cmd/                    # Application entry points
│   ├── ffprobe-api/       # Main API server
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── api/              # API routing and middleware
│   ├── handlers/         # HTTP request handlers
│   ├── services/         # Business logic layer
│   ├── database/         # Database layer (repositories)
│   ├── models/           # Data models and types
│   ├── ffmpeg/           # FFmpeg/FFprobe integration
│   ├── quality/          # Video quality analysis
│   ├── hls/              # HLS streaming analysis
│   ├── storage/          # File storage providers
│   ├── reports/          # Report generation
│   ├── middleware/       # HTTP middleware
│   ├── config/           # Configuration management
│   └── validator/        # Input validation utilities
├── tests/                # Integration and end-to-end tests
├── docker/               # Docker configuration files
├── scripts/              # Utility scripts
└── docs/                 # Documentation
```

### Layer Architecture
```
┌─────────────────────────────────────────┐
│           HTTP Handlers                 │ ← API endpoints
├─────────────────────────────────────────┤
│           Services Layer                │ ← Business logic
├─────────────────────────────────────────┤
│        Repository Layer                 │ ← Data access
├─────────────────────────────────────────┤
│         Database Layer                  │ ← PostgreSQL
└─────────────────────────────────────────┘
```

### Core Components

#### 1. **FFmpeg Integration** (`internal/ffmpeg/`)
- **Purpose**: FFprobe execution and result parsing
- **Key Files**:
  - `ffprobe.go` - Core FFprobe wrapper
  - `builder.go` - Command builder pattern
  - `validation.go` - Input validation
  - `types.go` - FFprobe data structures

#### 2. **Quality Analysis** (`internal/quality/`)
- **Purpose**: Video quality metrics (VMAF, PSNR, SSIM)
- **Key Files**:
  - `analyzer.go` - Quality analysis engine
  - `types.go` - Quality metric types
  - `validation.go` - Quality request validation
  - `accuracy_validator.go` - Metric accuracy testing

#### 3. **Storage Systems** (`internal/storage/`)
- **Purpose**: Multi-provider file storage
- **Supported Providers**: Local, S3, Google Cloud, Azure
- **Key Files**:
  - `interface.go` - Storage provider interface
  - `factory.go` - Provider factory
  - Individual provider implementations

#### 4. **HLS Analysis** (`internal/hls/`)
- **Purpose**: HLS streaming analysis
- **Key Files**:
  - `analyzer.go` - HLS stream analyzer
  - `parser.go` - Manifest parser
  - `validation.go` - HLS validation rules

## 🛠️ Development Environment

### Environment Variables
Create a `.env` file in the project root:
```bash
# Database
DATABASE_URL=postgres://ffprobe:password@localhost:5432/ffprobe_db

# Redis
REDIS_URL=redis://localhost:6379

# FFmpeg
FFMPEG_PATH=/usr/local/bin/ffmpeg
FFPROBE_PATH=/usr/local/bin/ffprobe

# Storage (optional)
AWS_ACCESS_KEY_ID=your_key
AWS_SECRET_ACCESS_KEY=your_secret
S3_BUCKET=your_bucket

# API Keys (optional)
OPENROUTER_API_KEY=your_openrouter_key

# Server
PORT=8080
LOG_LEVEL=debug
```

### Docker Development
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Rebuild and restart
docker-compose up --build -d

# Clean restart
docker-compose down -v && docker-compose up -d
```

### Local Development
```bash
# Install development tools
go install github.com/cosmtrek/air@latest  # Hot reload
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Start with hot reload
air

# Run linting
golangci-lint run

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 Code Standards

### Go Style Guidelines
We follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html).

#### Naming Conventions
```go
// ✅ Good
type AnalysisService struct {
    ffprobe *ffmpeg.FFprobe
    logger  zerolog.Logger
}

func (s *AnalysisService) ProcessFile(ctx context.Context, filePath string) error {
    // Implementation
}

// ❌ Bad
type analysisservice struct {
    FFProbe *ffmpeg.FFprobe
    Logger  zerolog.Logger
}

func (s *analysisservice) processfile(ctx context.Context, filepath string) error {
    // Implementation
}
```

#### Error Handling
```go
// ✅ Good - Wrap errors with context
func (s *AnalysisService) ProcessFile(ctx context.Context, filePath string) (*models.Analysis, error) {
    if err := s.validateFile(filePath); err != nil {
        return nil, fmt.Errorf("file validation failed: %w", err)
    }
    
    result, err := s.ffprobe.Probe(ctx, filePath)
    if err != nil {
        return nil, fmt.Errorf("ffprobe execution failed: %w", err)
    }
    
    return result, nil
}

// ❌ Bad - Generic error messages
func (s *AnalysisService) ProcessFile(ctx context.Context, filePath string) (*models.Analysis, error) {
    if err := s.validateFile(filePath); err != nil {
        return nil, err  // No context
    }
    
    result, err := s.ffprobe.Probe(ctx, filePath)
    if err != nil {
        return nil, fmt.Errorf("error: %v", err)  // Generic message
    }
    
    return result, nil
}
```

#### Logging Standards
```go
// ✅ Good - Structured logging
func (h *ProbeHandler) AnalyzeFile(c *gin.Context) {
    analysisID := uuid.New()
    
    h.logger.Info().
        Str("analysis_id", analysisID.String()).
        Str("file_path", filePath).
        Str("user_id", userID).
        Msg("Starting file analysis")
        
    // ... processing ...
    
    h.logger.Info().
        Str("analysis_id", analysisID.String()).
        Dur("processing_time", processingTime).
        Msg("Analysis completed successfully")
}

// ❌ Bad - Unstructured logging
func (h *ProbeHandler) AnalyzeFile(c *gin.Context) {
    log.Printf("Starting analysis for file: %s", filePath)
    
    // ... processing ...
    
    log.Printf("Analysis completed in %v", processingTime)
}
```

### Package Organization
- **One responsibility per package**
- **Interfaces in consumer packages**
- **No circular dependencies**
- **Clear package naming**

```go
// ✅ Good package structure
package services

import (
    "github.com/rendiffdev/ffprobe-api/internal/models"
    "github.com/rendiffdev/ffprobe-api/internal/database"
    "github.com/rendiffdev/ffprobe-api/internal/ffmpeg"
)

type AnalysisService struct {
    repo    database.Repository  // Interface from database package
    ffprobe ffmpeg.FFprobe      // Concrete type from ffmpeg package
    logger  zerolog.Logger
}
```

## 🧪 Testing Guidelines

### Test Categories

#### 1. Unit Tests
```go
// File: internal/services/analysis_service_test.go
func TestAnalysisService_ProcessFile(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *models.Analysis
        wantErr bool
    }{
        {
            name:    "valid video file",
            input:   "testdata/sample.mp4",
            wantErr: false,
        },
        {
            name:    "nonexistent file",
            input:   "testdata/nonexistent.mp4",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := setupTestService(t)
            
            got, err := service.ProcessFile(context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.NotNil(t, got)
        })
    }
}
```

#### 2. Integration Tests
```go
// File: tests/integration_test.go
func TestAPIIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Start test server
    server := setupTestServer(t, db)
    defer server.Close()
    
    // Test API endpoints
    resp, err := http.Post(server.URL+"/api/v1/probe", "application/json", body)
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

#### 3. End-to-End Tests
```go
// File: tests/e2e_test.go
func TestCompleteWorkflow(t *testing.T) {
    // Test complete analysis workflow
    // 1. Upload file
    // 2. Start analysis
    // 3. Check status
    // 4. Get results
    // 5. Generate report
}
```

### Test Data Management
```
tests/
├── testdata/
│   ├── sample.mp4          # Small test video
│   ├── sample.mov          # Different format
│   ├── corrupt.mp4         # Corrupted file for error testing
│   └── manifest.m3u8       # HLS test data
├── fixtures/
│   ├── analysis_response.json
│   └── quality_metrics.json
└── helpers/
    ├── test_server.go      # Test server setup
    └── test_data.go        # Test data generators
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestAnalysisService_ProcessFile ./internal/services

# Run integration tests only
go test -tags=integration ./tests

# Run with verbose output
go test -v ./...
```

## 🔒 Security Guidelines

### Input Validation
Always validate and sanitize user inputs:

```go
// ✅ Good - Comprehensive validation
func ValidateFilePath(filePath string) error {
    if strings.TrimSpace(filePath) == "" {
        return fmt.Errorf("file path cannot be empty")
    }
    
    // Check path length
    if len(filePath) > 2000 {
        return fmt.Errorf("file path too long")
    }
    
    // Check for dangerous characters
    dangerousChars := []string{";", "&", "|", "`", "$", "(", ")", "<", ">"}
    for _, char := range dangerousChars {
        if strings.Contains(filePath, char) {
            return fmt.Errorf("file path contains dangerous character: %s", char)
        }
    }
    
    // Check for path traversal
    if strings.Contains(filePath, "..") {
        return fmt.Errorf("file path contains path traversal")
    }
    
    return nil
}
```

### SQL Injection Prevention
```go
// ✅ Good - Use parameterized queries
func (r *Repository) GetAnalysisByID(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
    query := `SELECT id, file_name, status FROM analyses WHERE id = $1`
    
    var analysis models.Analysis
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &analysis.ID,
        &analysis.FileName,
        &analysis.Status,
    )
    
    return &analysis, err
}

// ❌ Bad - String concatenation
func (r *Repository) GetAnalysisByID(ctx context.Context, id string) (*models.Analysis, error) {
    query := fmt.Sprintf("SELECT * FROM analyses WHERE id = '%s'", id)  // VULNERABLE!
    // ... rest of implementation
}
```

### API Security
```go
// Rate limiting middleware
func RateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 60) // 60 requests per minute
    
    return gin.HandlerFunc(func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    })
}

// Input size limiting
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
        c.Next()
    })
}
```

## 🎯 API Design Principles

### RESTful Conventions
```
# Resource Collections
GET    /api/v1/analyses           # List analyses
POST   /api/v1/analyses           # Create analysis

# Individual Resources  
GET    /api/v1/analyses/{id}      # Get analysis
PUT    /api/v1/analyses/{id}      # Update analysis
DELETE /api/v1/analyses/{id}      # Delete analysis

# Sub-resources
GET    /api/v1/analyses/{id}/reports     # Get reports for analysis
POST   /api/v1/analyses/{id}/reports     # Create report

# Actions (when REST isn't sufficient)
POST   /api/v1/analyses/{id}/cancel      # Cancel analysis
POST   /api/v1/batch/analyze             # Batch operations
```

### Response Format Standards
```go
// Success Response
type APIResponse struct {
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

// Error Response
type ErrorResponse struct {
    Error   string      `json:"error"`
    Details string      `json:"details,omitempty"`
    Code    string      `json:"code,omitempty"`
}

// Pagination Meta
type Meta struct {
    Total       int `json:"total"`
    Page        int `json:"page"`
    PerPage     int `json:"per_page"`
    TotalPages  int `json:"total_pages"`
}
```

### Status Code Guidelines
```go
// ✅ Good - Appropriate status codes
func (h *AnalysisHandler) GetAnalysis(c *gin.Context) {
    id := c.Param("id")
    
    analysis, err := h.service.GetAnalysis(ctx, id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            c.JSON(http.StatusNotFound, ErrorResponse{
                Error: "Analysis not found",
            })
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Error: "Internal server error",
        })
        return
    }
    
    c.JSON(http.StatusOK, APIResponse{
        Data: analysis,
    })
}
```

## 🗄️ Database Guidelines

### Migration Management
```sql
-- migrations/001_create_analyses_table.up.sql
CREATE TABLE analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_name VARCHAR(500) NOT NULL,
    file_path VARCHAR(2000) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_analyses_status ON analyses(status);
CREATE INDEX idx_analyses_created_at ON analyses(created_at);
```

```sql
-- migrations/001_create_analyses_table.down.sql
DROP TABLE analyses;
```

### Repository Pattern
```go
type Repository interface {
    CreateAnalysis(ctx context.Context, analysis *models.Analysis) error
    GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error)
    UpdateAnalysisStatus(ctx context.Context, id uuid.UUID, status models.AnalysisStatus) error
    ListAnalyses(ctx context.Context, filters ListFilters) ([]*models.Analysis, error)
}

type PostgresRepository struct {
    db *sql.DB
}

func (r *PostgresRepository) CreateAnalysis(ctx context.Context, analysis *models.Analysis) error {
    query := `
        INSERT INTO analyses (id, file_name, file_path, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)`
    
    _, err := r.db.ExecContext(ctx, query,
        analysis.ID,
        analysis.FileName,
        analysis.FilePath,
        analysis.Status,
        analysis.CreatedAt,
        analysis.UpdatedAt,
    )
    
    return err
}
```

## ⚡ Performance Considerations

### Database Optimization
```go
// ✅ Good - Use prepared statements for repeated queries
type Repository struct {
    getAnalysisStmt *sql.Stmt
}

func NewRepository(db *sql.DB) (*Repository, error) {
    stmt, err := db.Prepare("SELECT id, file_name, status FROM analyses WHERE id = $1")
    if err != nil {
        return nil, err
    }
    
    return &Repository{
        getAnalysisStmt: stmt,
    }, nil
}

// ✅ Good - Implement pagination
func (r *Repository) ListAnalyses(ctx context.Context, limit, offset int) ([]*models.Analysis, error) {
    query := `
        SELECT id, file_name, status, created_at 
        FROM analyses 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2`
    
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    // ... handle results
}
```

### Caching Strategy
```go
// Redis caching for expensive operations
func (s *AnalysisService) GetAnalysis(ctx context.Context, id uuid.UUID) (*models.Analysis, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("analysis:%s", id.String())
    if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
        var analysis models.Analysis
        if err := json.Unmarshal([]byte(cached), &analysis); err == nil {
            return &analysis, nil
        }
    }
    
    // Cache miss - get from database
    analysis, err := s.repo.GetAnalysis(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    if data, err := json.Marshal(analysis); err == nil {
        s.redis.Set(ctx, cacheKey, data, 10*time.Minute)
    }
    
    return analysis, nil
}
```

### Concurrent Processing
```go
// ✅ Good - Use worker pools for batch processing
func (s *BatchService) ProcessBatch(ctx context.Context, files []string) error {
    const maxWorkers = 5
    semaphore := make(chan struct{}, maxWorkers)
    
    var wg sync.WaitGroup
    errChan := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(filePath string) {
            defer wg.Done()
            
            semaphore <- struct{}{}        // Acquire
            defer func() { <-semaphore }() // Release
            
            if err := s.processFile(ctx, filePath); err != nil {
                errChan <- err
            }
        }(file)
    }
    
    wg.Wait()
    close(errChan)
    
    // Check for errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## 🔄 Contribution Workflow

### 1. Issue Creation
Before starting work:
- Check existing issues and PRs
- Create an issue describing the problem/feature
- Wait for maintainer feedback on approach
- Get issue assigned to you

### 2. Development Process
```bash
# 1. Create feature branch
git checkout -b feature/add-new-metric

# 2. Make changes with frequent commits
git add .
git commit -m "feat: add VMAF quality metric support

- Implement VMAF analyzer in quality package
- Add VMAF configuration options
- Update quality service to support VMAF
- Add comprehensive tests for VMAF analysis

Closes #123"

# 3. Keep branch updated
git fetch origin
git rebase origin/main

# 4. Push changes
git push origin feature/add-new-metric
```

### 3. Pull Request Guidelines
```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] New tests added for new functionality
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No new security vulnerabilities introduced
```

### 4. Code Review Process
- All PRs require at least one approval
- Address all feedback before merging
- Maintain clean commit history
- Update documentation as needed

## 🐛 Troubleshooting

### Common Development Issues

#### 1. FFmpeg/FFprobe Not Found
```bash
# Linux/Ubuntu
sudo apt-get update
sudo apt-get install ffmpeg

# macOS
brew install ffmpeg

# Verify installation
ffprobe -version
```

#### 2. Database Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps

# View PostgreSQL logs
docker-compose logs postgres

# Reset database
docker-compose down -v
docker-compose up -d postgres
```

#### 3. Go Module Issues
```bash
# Clean module cache
go clean -modcache

# Verify dependencies
go mod verify

# Update dependencies
go mod tidy
```

#### 4. Docker Build Issues
```bash
# Clear Docker cache
docker system prune -a

# Rebuild without cache
docker-compose build --no-cache

# Check Docker logs
docker-compose logs api
```

### Testing Issues

#### 1. Test Data Missing
```bash
# Create test data directory
mkdir -p tests/testdata

# Generate test video (requires FFmpeg)
ffmpeg -f lavfi -i testsrc=duration=10:size=320x240:rate=1 -c:v libx264 tests/testdata/sample.mp4
```

#### 2. Race Condition in Tests
```go
// ✅ Good - Proper test isolation
func TestConcurrentAnalysis(t *testing.T) {
    t.Parallel()  // Allow parallel execution
    
    // Use unique test data for each test
    testFile := createTempTestFile(t)
    defer os.Remove(testFile)
    
    // Test implementation
}
```

### Performance Issues

#### 1. Slow FFprobe Execution
- Check FFprobe arguments for efficiency
- Verify file accessibility
- Monitor system resources
- Consider using FFprobe's `-select_streams` for large files

#### 2. Database Performance
```sql
-- Check slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Analyze table statistics
ANALYZE analyses;

-- Check index usage
SELECT schemaname, tablename, attname, n_distinct, correlation 
FROM pg_stats 
WHERE tablename = 'analyses';
```

## 📚 Knowledge Base

### Frequently Asked Questions

#### Q: How do I add a new quality metric?
1. Add metric type to `internal/quality/types.go`
2. Implement analyzer in `internal/quality/analyzer.go`
3. Add validation in `internal/quality/validation.go`
4. Update service layer in `internal/services/quality_service.go`
5. Add API endpoints in `internal/handlers/quality_handler.go`
6. Write comprehensive tests

#### Q: How do I add a new storage provider?
1. Implement `Provider` interface in `internal/storage/`
2. Add provider to factory in `internal/storage/factory.go`
3. Add configuration options in `internal/config/config.go`
4. Add validation in `internal/storage/validation.go`
5. Write integration tests

#### Q: How do I handle breaking API changes?
1. Use API versioning (`/api/v1/`, `/api/v2/`)
2. Maintain backward compatibility when possible
3. Document deprecation timeline
4. Provide migration guide
5. Update client libraries

#### Q: How do I optimize for large files?
1. Use streaming where possible
2. Implement progress reporting
3. Add timeouts and cancellation
4. Consider chunked processing
5. Monitor memory usage

### Useful Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [FFmpeg Documentation](https://ffmpeg.org/documentation.html)
- [Docker Documentation](https://docs.docker.com/)

### Team Communication

- **Slack**: #ffprobe-api-dev
- **Issues**: GitHub Issues for bugs and features
- **Discussions**: GitHub Discussions for questions
- **Wiki**: Project wiki for detailed documentation

### Getting Help

1. Check this documentation first
2. Search existing issues
3. Ask in team Slack channel
4. Create a GitHub discussion
5. Reach out to maintainers directly

---

**Happy Contributing! 🎉**

If you have questions or suggestions for improving this guide, please open an issue or submit a PR.