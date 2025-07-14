# Multi-stage build for FFprobe API with FFmpeg and libvmaf support
# Stage 1: Build dependencies and compile FFmpeg with libvmaf
FROM alpine:3.19 AS ffmpeg-builder

# Install build dependencies
RUN apk add --no-cache \
    build-base \
    cmake \
    git \
    nasm \
    pkgconfig \
    python3 \
    python3-dev \
    meson \
    ninja \
    yasm \
    zlib-dev \
    openssl-dev \
    opus-dev \
    libvorbis-dev \
    lame-dev \
    fdk-aac-dev \
    x264-dev \
    x265-dev \
    libvpx-dev \
    libaom-dev \
    libass-dev \
    freetype-dev \
    libtheora-dev \
    libwebp-dev

# Set working directory
WORKDIR /tmp

# Build libvmaf
RUN git clone --depth 1 --branch v3.0.0 https://github.com/Netflix/vmaf.git && \
    cd vmaf && \
    cd libvmaf && \
    meson setup build --buildtype=release --default-library=static && \
    ninja -vC build && \
    ninja -vC build install

# Download VMAF models
RUN mkdir -p /usr/local/share/vmaf && \
    cd /usr/local/share/vmaf && \
    wget -q https://github.com/Netflix/vmaf/raw/master/model/vmaf_v0.6.1.json && \
    wget -q https://github.com/Netflix/vmaf/raw/master/model/vmaf_v0.6.1neg.json && \
    wget -q https://github.com/Netflix/vmaf/raw/master/model/vmaf_4k_v0.6.1.json && \
    wget -q https://github.com/Netflix/vmaf/raw/master/model/vmaf_b_v0.6.3.json

# Build FFmpeg with libvmaf support
RUN git clone --depth 1 --branch n6.1 https://github.com/FFmpeg/FFmpeg.git ffmpeg && \
    cd ffmpeg && \
    ./configure \
        --prefix=/usr/local \
        --enable-gpl \
        --enable-version3 \
        --enable-static \
        --disable-shared \
        --disable-debug \
        --disable-ffplay \
        --disable-indev=sndio \
        --disable-outdev=sndio \
        --cc=gcc \
        --enable-fontconfig \
        --enable-frei0r \
        --enable-gnutls \
        --enable-libass \
        --enable-libbluray \
        --enable-libfdk-aac \
        --enable-libfreetype \
        --enable-libmp3lame \
        --enable-libopus \
        --enable-libtheora \
        --enable-libvorbis \
        --enable-libvpx \
        --enable-libwebp \
        --enable-libx264 \
        --enable-libx265 \
        --enable-libxml2 \
        --enable-libvmaf \
        --enable-libaom \
        --enable-nonfree && \
    make -j$(nproc) && \
    make install

# Stage 2: Go application builder
FROM golang:1.21-alpine AS go-builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(date +%Y%m%d-%H%M%S)" \
    -o ffprobe-api \
    ./cmd/ffprobe-api

# Stage 3: Final runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    postgresql-client \
    redis \
    python3 \
    && rm -rf /var/cache/apk/*

# Copy FFmpeg binaries and VMAF models from builder
COPY --from=ffmpeg-builder /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg-builder /usr/local/bin/ffprobe /usr/local/bin/ffprobe
COPY --from=ffmpeg-builder /usr/local/share/vmaf /usr/local/share/vmaf

# Copy Go application from builder
COPY --from=go-builder /app/ffprobe-api /usr/local/bin/ffprobe-api

# Create non-root user
RUN adduser -D -s /bin/sh -u 1001 ffprobe

# Create directories for uploads, reports, and models
RUN mkdir -p /app/uploads /app/reports /app/models /app/logs && \
    chown -R ffprobe:ffprobe /app

# Set working directory
WORKDIR /app

# Create default configuration directory
RUN mkdir -p /app/config

# Copy example configuration
COPY .env.example /app/config/env.example

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Switch to non-root user
USER ffprobe

# Set environment variables
ENV FFMPEG_PATH=/usr/local/bin/ffmpeg
ENV FFPROBE_PATH=/usr/local/bin/ffprobe
ENV VMAF_MODEL_PATH=/usr/local/share/vmaf
ENV UPLOAD_DIR=/app/uploads
ENV REPORTS_DIR=/app/reports
ENV LOG_LEVEL=info
ENV API_PORT=8080

# Run the application
CMD ["/usr/local/bin/ffprobe-api"]