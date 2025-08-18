#!/bin/bash
set -e

# FFprobe API - Hybrid Production Multi-Architecture Build Script
# Python Flask + FFmpeg with advanced features

DOCKER_REPO="rendiffdev/ffprobe-api"
VERSION=${1:-"v1.0.0"}
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🚀 Building FFprobe API Hybrid Production Multi-Architecture Images"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Repository: $DOCKER_REPO"
echo "Version: $VERSION"
echo "Build Date: $BUILD_DATE"
echo "Git Commit: $GIT_COMMIT"
echo "Base: Python Flask + Enhanced Features"
echo "Platforms: linux/amd64, linux/arm64"
echo "Features: SQLite, Rate Limiting, Request Logging, Multi-arch"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

# Ensure buildx is available
echo "🔧 Setting up Docker buildx..."
if ! docker buildx inspect multiarch >/dev/null 2>&1; then
    docker buildx create --name multiarch --driver docker-container --bootstrap
    docker buildx use multiarch
else
    docker buildx use multiarch
fi

echo "📦 Building hybrid production multi-architecture image..."
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag "$DOCKER_REPO:hybrid" \
  --tag "$DOCKER_REPO:$VERSION" \
  --tag "$DOCKER_REPO:latest" \
  --tag "$DOCKER_REPO:production" \
  --tag "$DOCKER_REPO:stable" \
  --tag "$DOCKER_REPO:full-featured" \
  --build-arg BUILD_DATE="$BUILD_DATE" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg VERSION="$VERSION" \
  --push \
  --file Dockerfile.hybrid \
  .. \
  --progress=plain

echo
echo "✅ Successfully built and pushed hybrid production images!"
echo

# Create semantic version tags if version is provided
if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "🏷️  Creating semantic version tags..."
    
    # Extract major.minor from version (e.g., v1.2.3 -> v1.2)
    MAJOR_MINOR=$(echo "$VERSION" | sed -E 's/^(v[0-9]+\.[0-9]+)\.[0-9]+$/\1/')
    # Extract major from version (e.g., v1.2.3 -> v1)
    MAJOR=$(echo "$VERSION" | sed -E 's/^(v[0-9]+)\.[0-9]+\.[0-9]+$/\1/')
    
    if [ "$MAJOR_MINOR" != "$VERSION" ]; then
        docker buildx imagetools create \
          "$DOCKER_REPO:$VERSION" \
          --tag "$DOCKER_REPO:$MAJOR_MINOR"
        echo "✅ Tagged: $DOCKER_REPO:$MAJOR_MINOR"
    fi
    
    if [ "$MAJOR" != "$VERSION" ] && [ "$MAJOR" != "$MAJOR_MINOR" ]; then
        docker buildx imagetools create \
          "$DOCKER_REPO:$VERSION" \
          --tag "$DOCKER_REPO:$MAJOR"
        echo "✅ Tagged: $DOCKER_REPO:$MAJOR"
    fi
fi

echo
echo "🎉 Hybrid Production Multi-Architecture Build Completed!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📦 Available Images:"
echo "   • $DOCKER_REPO:hybrid (main hybrid image)"
echo "   • $DOCKER_REPO:production (production alias)"
echo "   • $DOCKER_REPO:latest (latest alias)"
echo "   • $DOCKER_REPO:stable (stable alias)"
echo "   • $DOCKER_REPO:full-featured (full-featured alias)"
echo "   • $DOCKER_REPO:$VERSION (version-specific)"

if [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    [ "$MAJOR_MINOR" != "$VERSION" ] && echo "   • $DOCKER_REPO:$MAJOR_MINOR (major.minor)"
    [ "$MAJOR" != "$VERSION" ] && [ "$MAJOR" != "$MAJOR_MINOR" ] && echo "   • $DOCKER_REPO:$MAJOR (major)"
fi

echo
echo "🏗️  Architecture Support:"
echo "   • linux/amd64 (Intel/AMD servers, cloud instances)"
echo "   • linux/arm64 (Apple Silicon Macs, ARM servers)"
echo
echo "✨ Enhanced Features:"
echo "   • SQLite database with analysis history"
echo "   • Rate limiting (100 requests/minute)"
echo "   • Request logging and statistics"
echo "   • Multi-endpoint API (health, stats, history)"
echo "   • Security headers and CORS protection"
echo "   • Enhanced error handling and logging"
echo "   • Platform detection and reporting"
echo "   • Comprehensive health checks"
echo
echo "🔒 Security Features:"
echo "   • Non-root user (UID 10001)"
echo "   • Security headers enabled"
echo "   • Input validation and sanitization"
echo "   • Rate limiting protection"
echo "   • No hardcoded secrets"
echo
echo "🚀 Ready for Production Deployment!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo
echo "📋 Quick Start Commands:"
echo "   # AMD64 (Linux servers):"
echo "   docker run -d -p 8080:8080 --platform linux/amd64 $DOCKER_REPO:hybrid"
echo
echo "   # ARM64 (Mac Silicon):"
echo "   docker run -d -p 8080:8080 --platform linux/arm64 $DOCKER_REPO:hybrid"
echo
echo "   # Auto-detect platform:"
echo "   docker run -d -p 8080:8080 $DOCKER_REPO:hybrid"
echo
echo "   # With persistent data:"
echo "   docker run -d -p 8080:8080 -v ./data:/app/data $DOCKER_REPO:hybrid"
echo
echo "📡 API Endpoints:"
echo "   • Health: curl http://localhost:8080/health"
echo "   • Version: curl http://localhost:8080/api/v1/version"
echo "   • Stats: curl http://localhost:8080/api/v1/stats"
echo "   • History: curl http://localhost:8080/api/v1/history"
echo "   • Analyze: curl -X POST http://localhost:8080/api/v1/probe -F \"file=@video.mp4\""
echo
echo "✨ Hybrid production images with full features are now available on Docker Hub!"