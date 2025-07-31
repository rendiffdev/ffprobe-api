# FFprobe Worker Service - Dedicated container for media analysis
FROM alpine:3.20

# Install FFmpeg, FFprobe and runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    bash \
    python3 \
    py3-pip \
    git \
    build-base \
    cmake \
    nasm \
    pkgconfig \
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

# Build libvmaf
WORKDIR /tmp
RUN git clone --depth 1 --branch v3.0.0 https://github.com/Netflix/vmaf.git && \
    cd vmaf/libvmaf && \
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

# Build FFmpeg with libvmaf
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
        --enable-libass \
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
        --enable-libvmaf \
        --enable-libaom \
        --enable-nonfree && \
    make -j$(nproc) && \
    make install && \
    cd / && rm -rf /tmp/*

# Install Go for the worker service
RUN apk add --no-cache go

# Create non-root user
RUN adduser -D -s /bin/sh -u 1001 ffprobe

# Create directories
RUN mkdir -p /app/temp /app/cache && \
    chown -R ffprobe:ffprobe /app

# Set working directory
WORKDIR /app

# Copy worker service code (will be created)
COPY ffprobe-worker /usr/local/bin/ffprobe-worker
RUN chmod +x /usr/local/bin/ffprobe-worker

# Switch to non-root user
USER ffprobe

# Environment variables
ENV FFMPEG_PATH=/usr/local/bin/ffmpeg \
    FFPROBE_PATH=/usr/local/bin/ffprobe \
    VMAF_MODEL_PATH=/usr/local/share/vmaf \
    TEMP_DIR=/app/temp \
    CACHE_DIR=/app/cache

# Expose worker port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Run the worker service
CMD ["/usr/local/bin/ffprobe-worker"]