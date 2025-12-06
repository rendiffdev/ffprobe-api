# FFprobe API - Production-Ready Multi-Stage Docker Build
# Supports both AMD64 and ARM64 architectures
# Based on latest FFmpeg from BtbN/ffmpeg-builds

# =============================================================================
# Build Stage - Go Application
# =============================================================================
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev

# Set working directory
WORKDIR /build

# Copy dependency files
COPY go.mod ./

# Copy source code first to get all dependencies
COPY . .

# Initialize go.sum and download dependencies  
RUN go mod tidy && \
    go mod download && \
    if [ -f go.sum ]; then go mod verify; fi

# Build application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o ffprobe-api \
    ./cmd/ffprobe-api

# =============================================================================
# FFmpeg 8.0 Stage - Using Official Static Builds
# =============================================================================
FROM linuxserver/ffmpeg:8.0-cli AS ffmpeg

# Verify version
RUN ffmpeg -version | head -1

# =============================================================================
# Production Stage - Final Image
# =============================================================================
FROM alpine:3.19 AS production

# Install runtime dependencies (without ffmpeg, we'll copy our custom build)
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    sqlite \
    dumb-init \
    su-exec \
    shadow \
    libass \
    freetype \
    libvpx \
    libvorbis \
    opus \
    libtheora \
    libwebp \
    x264-libs \
    x265-libs \
    fdk-aac \
    lame \
    openssl \
    libva \
    v4l-utils-libs

# Create app user and group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy FFmpeg 8.0 binaries from linuxserver/ffmpeg
COPY --from=ffmpeg /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg /usr/local/bin/ffprobe /usr/local/bin/ffprobe

# Copy application binary from builder stage
COPY --from=builder /build/ffprobe-api /app/

# Verify FFmpeg 8.0 installation and show version
RUN /usr/local/bin/ffmpeg -version | head -1 && \
    /usr/local/bin/ffprobe -version | head -1

# Copy migration files (if they exist)
# COPY --from=builder /build/migrations /app/migrations/

# Create required directories
RUN mkdir -p \
    /app/data \
    /app/uploads \
    /app/reports \
    /app/temp \
    /app/cache \
    /app/logs

# Set proper ownership
RUN chown -R appuser:appgroup /app

# Create health check script
RUN echo '#!/bin/sh' > /app/healthcheck.sh && \
    echo 'curl -f http://localhost:8080/health || exit 1' >> /app/healthcheck.sh && \
    chmod +x /app/healthcheck.sh

# Expose port
EXPOSE 8080

# Set environment variables (non-sensitive)
ENV GO_ENV=production \
    GIN_MODE=release \
    API_PORT=8080 \
    API_HOST=0.0.0.0 \
    LOG_LEVEL=info \
    DB_TYPE=sqlite \
    DB_PATH=/app/data/ffprobe.db \
    UPLOAD_DIR=/app/uploads \
    REPORTS_DIR=/app/reports \
    FFMPEG_PATH=/usr/local/bin/ffmpeg \
    FFPROBE_PATH=/usr/local/bin/ffprobe

# Security settings - PRODUCTION DEFAULTS (can be overridden at runtime)
# WARNING: Set CLOUD_MODE=true only for development/testing
ENV ENABLE_AUTH=true \
    ENABLE_RATE_LIMIT=true \
    ENABLE_CIRCUIT_BREAKER=true \
    CLOUD_MODE=false \
    SKIP_AUTH_VALIDATION=false \
    REQUIRE_LLM=true \
    ENABLE_LOCAL_LLM=true \
    OLLAMA_URL=http://ollama:11434 \
    OLLAMA_MODEL=gemma3:270m

# Required secrets - MUST be set at runtime
# JWT_SECRET - Required, at least 32 characters
# API_KEY - Required, at least 32 characters
# VALKEY_PASSWORD - Required for cache authentication

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD ["/app/healthcheck.sh"]

# Use dumb-init for proper signal handling
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Switch to non-root user
USER appuser:appgroup

# Start application
CMD ["/app/ffprobe-api"]

# =============================================================================
# Development Stage - Hot Reload Support
# =============================================================================
FROM golang:1.23-alpine AS development

# Install development dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev \
    curl \
    air

# Install FFmpeg (development version)
RUN apk add --no-cache ffmpeg

# Set working directory
WORKDIR /app

# Install Air for hot reloading
RUN go install github.com/cosmtrek/air@latest

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Expose port
EXPOSE 8080

# Environment for development
ENV GO_ENV=development \
    GIN_MODE=debug \
    API_PORT=8080 \
    API_HOST=0.0.0.0 \
    LOG_LEVEL=debug

# Health check for development
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Development command with hot reload
CMD ["air", "-c", ".air.toml"]

# =============================================================================
# Test Stage - For CI/CD Testing
# =============================================================================
FROM golang:1.23-alpine AS test

# Install test dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    gcc \
    musl-dev \
    sqlite-dev \
    ffmpeg

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod download

# Run tests
RUN go test -v -race -coverprofile=coverage.out ./...

# Generate test report
RUN go tool cover -html=coverage.out -o coverage.html

# =============================================================================
# Minimal Stage - Smallest Possible Image
# =============================================================================
FROM scratch AS minimal

# Copy ca-certificates for HTTPS
COPY --from=alpine:3.19 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=alpine:3.19 /usr/share/zoneinfo /usr/share/zoneinfo

# Copy application binary
COPY --from=builder /build/ffprobe-api /ffprobe-api

# Copy FFmpeg 8.0 binaries and libraries
COPY --from=ffmpeg /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg /usr/local/bin/ffprobe /usr/local/bin/ffprobe
COPY --from=ffmpeg /usr/local/lib/libav*.so* /usr/local/lib/
COPY --from=ffmpeg /usr/local/lib/libsw*.so* /usr/local/lib/

# Environment
ENV PATH="/usr/local/bin:$PATH"
ENV LD_LIBRARY_PATH="/usr/local/lib:$LD_LIBRARY_PATH"
ENV FFMPEG_PATH=/usr/local/bin/ffmpeg
ENV FFPROBE_PATH=/usr/local/bin/ffprobe

# Expose port
EXPOSE 8080

# Start application
ENTRYPOINT ["/ffprobe-api"]