#!/bin/bash
set -e

# FFprobe API - Minimal Working Image Build Script
# Creates a simple working Docker image with Python Flask wrapper

DOCKER_REPO="rendiffdev/ffprobe-api"
VERSION="minimal"
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🚀 Building FFprobe API Minimal Working Image"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Repository: $DOCKER_REPO"
echo "Version: $VERSION"
echo "Build Date: $BUILD_DATE"
echo "Git Commit: $GIT_COMMIT"
echo "Platform: linux/amd64 only"
echo "Features: Python Flask wrapper, FFprobe analysis"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo

echo "📦 Building minimal working image..."
docker buildx build \
  --platform linux/amd64 \
  --tag "$DOCKER_REPO:$VERSION" \
  --tag "$DOCKER_REPO:latest" \
  --tag "$DOCKER_REPO:working" \
  --build-arg BUILD_DATE="$BUILD_DATE" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg VERSION="$VERSION" \
  --push \
  --file Dockerfile.minimal \
  ..

echo
echo "✅ Successfully built and pushed: $DOCKER_REPO:$VERSION"
echo "✅ Successfully built and pushed: $DOCKER_REPO:latest" 
echo "✅ Successfully built and pushed: $DOCKER_REPO:working"

echo
echo "🎉 Minimal Working Image Build Completed!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📦 Available Images:"
echo "   • $DOCKER_REPO:minimal (main minimal image)"
echo "   • $DOCKER_REPO:latest (alias)"
echo "   • $DOCKER_REPO:working (alias)"
echo
echo "🚀 Ready for Immediate Use!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo
echo "📋 Quick Start Commands:"
echo "   # Download and run:"
echo "   docker run -d -p 8080:8080 $DOCKER_REPO:minimal"
echo
echo "   # Check health:"
echo "   curl http://localhost:8080/health"
echo
echo "   # Test video upload:"
echo "   curl -X POST http://localhost:8080/api/v1/probe -F \"file=@video.mp4\""
echo
echo "✨ The minimal working image is now available on Docker Hub!"