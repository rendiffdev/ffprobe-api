# FFprobe API - Production-Ready Docker Image for Docker Hub
# Multi-stage build optimized for minimal size and maximum performance
# Built with BtbN's FFmpeg binaries and AI-powered analysis capabilities

# Stage 1: Download FFmpeg binaries from BtbN
FROM alpine:3.20 AS ffmpeg-downloader

RUN apk add --no-cache curl tar xz jq ca-certificates

WORKDIR /tmp

# Download the latest stable FFmpeg build from BtbN
RUN curl -s https://api.github.com/repos/BtbN/FFmpeg-Builds/releases/latest | \
    jq -r '.assets[] | select(.name | contains("linux64-gpl.tar.xz") and contains("master")) | .browser_download_url' | \
    head -1 | xargs curl -L -o ffmpeg.tar.xz && \
    tar -xJf ffmpeg.tar.xz && \
    mv ffmpeg-*/bin/ffmpeg /usr/local/bin/ && \
    mv ffmpeg-*/bin/ffprobe /usr/local/bin/ && \
    chmod +x /usr/local/bin/ffmpeg /usr/local/bin/ffprobe

# Verify FFmpeg installation
RUN /usr/local/bin/ffmpeg -version && /usr/local/bin/ffprobe -version

# Download VMAF models for quality analysis
RUN mkdir -p /usr/local/share/vmaf && \
    cd /usr/local/share/vmaf && \
    curl -sLO https://github.com/Netflix/vmaf/raw/master/model/vmaf_v0.6.1.json && \
    curl -sLO https://github.com/Netflix/vmaf/raw/master/model/vmaf_v0.6.1neg.json && \
    curl -sLO https://github.com/Netflix/vmaf/raw/master/model/vmaf_4k_v0.6.1.json && \
    curl -sLO https://github.com/Netflix/vmaf/raw/master/model/vmaf_b_v0.6.3.json

# Stage 2: Build Go application
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git build-base ca-certificates

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with all optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s -X main.version=docker -X main.buildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    -o ffprobe-api cmd/ffprobe-api/main.go

# Stage 3: Final production image
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    bash \
    curl \
    # FFmpeg runtime libraries
    libstdc++ \
    libgcc \
    libgomp \
    # Video processing
    mesa-gl \
    mesa-glapi \
    # Audio processing
    alsa-lib \
    # Subtitle processing
    fontconfig \
    ttf-liberation \
    # Network protocols
    librtmp \
    # Image processing
    libjpeg-turbo \
    libpng \
    libwebp

# Create non-root user
RUN adduser -D -u 1000 ffprobe

# Create necessary directories with proper permissions
RUN mkdir -p /app/data /app/uploads /app/temp /app/cache /app/reports /app/backup && \
    chown -R ffprobe:ffprobe /app

WORKDIR /app

# Copy FFmpeg binaries from downloader stage
COPY --from=ffmpeg-downloader /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg-downloader /usr/local/bin/ffprobe /usr/local/bin/ffprobe
COPY --from=ffmpeg-downloader /usr/local/share/vmaf /usr/local/share/vmaf

# Copy Go application from builder
COPY --from=builder /build/ffprobe-api /app/ffprobe-api

# Copy migrations
COPY --chown=ffprobe:ffprobe migrations /app/migrations

# Create entrypoint script
RUN cat > /app/entrypoint.sh << 'EOF'
#!/bin/bash
set -e

echo "Starting FFprobe API..."
echo "Version: $(ffprobe -version | head -1)"
echo "Build: Docker Hub Production Image"
echo "Date: $(date -u +'%Y-%m-%d %H:%M:%S UTC')"

# Initialize SQLite database
echo "Initializing SQLite database..."
if [ ! -f "/app/data/ffprobe.db" ]; then
    echo "Creating new SQLite database..."
    # Database will be created automatically on first start
fi

# Check cache connection if Valkey is configured
if [ ! -z "$VALKEY_HOST" ]; then
    echo "Valkey cache configured at $VALKEY_HOST:$VALKEY_PORT"
fi

# Start the application
exec /app/ffprobe-api "$@"
EOF

RUN chmod +x /app/entrypoint.sh && \
    chown ffprobe:ffprobe /app/entrypoint.sh

# Set environment variables with sensible defaults for immediate use
ENV FFMPEG_PATH=/usr/local/bin/ffmpeg \
    FFPROBE_PATH=/usr/local/bin/ffprobe \
    VMAF_MODEL_PATH=/usr/local/share/vmaf \
    API_PORT=8080 \
    HOST=0.0.0.0 \
    # Database Configuration (SQLite embedded - zero config)
    DB_TYPE=sqlite \
    DB_PATH=/app/data/ffprobe.db \
    # Cache Configuration (optional Valkey)
    VALKEY_HOST="" \
    VALKEY_PORT=6379 \
    VALKEY_PASSWORD="" \
    # Directory Configuration
    UPLOAD_DIR=/app/uploads \
    TEMP_DIR=/app/temp \
    CACHE_DIR=/app/cache \
    REPORTS_DIR=/app/reports \
    DATA_DIR=/app/data \
    BACKUP_DIR=/app/backup \
    # Performance Configuration
    MAX_FILE_SIZE=10737418240 \
    WORKER_POOL_SIZE=4 \
    PROCESSING_TIMEOUT=300 \
    # Security Configuration (disabled by default for easy setup)
    ENABLE_AUTH=false \
    API_KEY="" \
    JWT_SECRET="" \
    # Rate Limiting (enabled with sensible defaults)
    ENABLE_RATE_LIMIT=true \
    RATE_LIMIT_PER_MINUTE=60 \
    RATE_LIMIT_PER_HOUR=1000 \
    # AI/LLM Configuration (disabled by default)
    ENABLE_LOCAL_LLM=false \
    OLLAMA_URL="" \
    OLLAMA_MODEL="" \
    # Application Configuration
    GO_ENV=production \
    GIN_MODE=release \
    LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${API_PORT}/health || exit 1

# Switch to non-root user
USER ffprobe

# Expose port
EXPOSE 8080

# Volume mount points
VOLUME ["/app/uploads", "/app/reports", "/app/data", "/app/backup"]

# Entry point
ENTRYPOINT ["/app/entrypoint.sh"]