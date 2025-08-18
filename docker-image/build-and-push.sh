#!/bin/bash
set -e

# FFprobe API - Docker Hub Build and Push Script
# Builds and pushes production-ready Docker images to Docker Hub

DOCKER_REPO="rendiffdev/ffprobe-api"
VERSION=${1:-latest}
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🚀 Building FFprobe API Docker Images"
echo "Repository: $DOCKER_REPO"
echo "Version: $VERSION"
echo "Build Date: $BUILD_DATE"
echo "Git Commit: $GIT_COMMIT"
echo

# Build multi-platform image
echo "📦 Building multi-platform Docker image..."
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag "$DOCKER_REPO:$VERSION" \
  --tag "$DOCKER_REPO:latest" \
  --build-arg BUILD_DATE="$BUILD_DATE" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg VERSION="$VERSION" \
  --push \
  --file Dockerfile \
  ..

echo "✅ Successfully built and pushed: $DOCKER_REPO:$VERSION"

# Build standalone image
echo "📦 Building standalone Docker image..."
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag "$DOCKER_REPO:standalone" \
  --build-arg BUILD_DATE="$BUILD_DATE" \
  --build-arg GIT_COMMIT="$GIT_COMMIT" \
  --build-arg VERSION="$VERSION-standalone" \
  --push \
  --file Dockerfile.standalone \
  ..

echo "✅ Successfully built and pushed: $DOCKER_REPO:standalone"

# Tag additional versions
if [ "$VERSION" != "latest" ]; then
    echo "🏷️  Creating additional tags..."
    
    # Tag major version (e.g., v1.0.0 -> v1)
    MAJOR_VERSION=$(echo "$VERSION" | sed -E 's/^v?([0-9]+)\..*/v\1/')
    if [ "$MAJOR_VERSION" != "$VERSION" ]; then
        docker buildx imagetools create \
          "$DOCKER_REPO:$VERSION" \
          --tag "$DOCKER_REPO:$MAJOR_VERSION"
        echo "✅ Tagged: $DOCKER_REPO:$MAJOR_VERSION"
    fi
    
    # Tag minor version (e.g., v1.0.0 -> v1.0)
    MINOR_VERSION=$(echo "$VERSION" | sed -E 's/^v?([0-9]+\.[0-9]+)\..*/v\1/')
    if [ "$MINOR_VERSION" != "$VERSION" ] && [ "$MINOR_VERSION" != "$MAJOR_VERSION" ]; then
        docker buildx imagetools create \
          "$DOCKER_REPO:$VERSION" \
          --tag "$DOCKER_REPO:$MINOR_VERSION"
        echo "✅ Tagged: $DOCKER_REPO:$MINOR_VERSION"
    fi
fi

# Verify images
echo "🔍 Verifying published images..."
docker run --rm "$DOCKER_REPO:latest" /app/ffprobe-api --version || echo "⚠️  Version check failed"

echo
echo "🎉 Build and push completed successfully!"
echo "📋 Available images:"
echo "   - $DOCKER_REPO:latest (main image)"
echo "   - $DOCKER_REPO:standalone (all-in-one)"
if [ "$VERSION" != "latest" ]; then
    echo "   - $DOCKER_REPO:$VERSION"
    [ "$MAJOR_VERSION" != "$VERSION" ] && echo "   - $DOCKER_REPO:$MAJOR_VERSION"
    [ "$MINOR_VERSION" != "$VERSION" ] && [ "$MINOR_VERSION" != "$MAJOR_VERSION" ] && echo "   - $DOCKER_REPO:$MINOR_VERSION"
fi
echo
echo "🚀 Ready for deployment! Users can now run:"
echo "   docker compose up -d"
echo "   # or"
echo "   docker run -d -p 8080:8080 $DOCKER_REPO:latest"