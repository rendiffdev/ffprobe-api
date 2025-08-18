#!/bin/bash
# FFprobe API - Docker Build Script
# Builds Docker images with support for multiple architectures and targets

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
IMAGE_NAME="ffprobe-api"
VERSION="latest"
TARGET="production"
PLATFORM="linux/amd64"
PUSH=false
BUILD_ARGS=""
CACHE=true

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "FFprobe API Docker Build Script"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -n, --name NAME         Docker image name (default: ffprobe-api)"
    echo "  -v, --version VERSION   Image version tag (default: latest)"
    echo "  -t, --target TARGET     Build target (production|development|test|minimal) (default: production)"
    echo "  -p, --platform PLATFORM Build platform (default: linux/amd64)"
    echo "      --multi-arch        Build for multiple architectures (linux/amd64,linux/arm64)"
    echo "      --push              Push image to registry"
    echo "      --no-cache          Disable Docker build cache"
    echo "  -a, --build-arg ARG     Pass build argument (can be used multiple times)"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                          # Build production image"
    echo "  $0 --target development                     # Build development image"
    echo "  $0 --multi-arch --push                      # Build and push multi-arch image"
    echo "  $0 --target minimal --name ffprobe-minimal  # Build minimal image"
    echo "  $0 --build-arg GO_VERSION=1.23              # Build with custom Go version"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -t|--target)
            TARGET="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        --multi-arch)
            PLATFORM="linux/amd64,linux/arm64"
            shift
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --no-cache)
            CACHE=false
            shift
            ;;
        -a|--build-arg)
            BUILD_ARGS="$BUILD_ARGS --build-arg $2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate target
case $TARGET in
    production|development|test|minimal)
        ;;
    *)
        print_error "Invalid target: $TARGET"
        print_error "Valid targets: production, development, test, minimal"
        exit 1
        ;;
esac

# Set full image name
FULL_IMAGE_NAME="${IMAGE_NAME}:${VERSION}"
if [[ "$TARGET" != "production" ]]; then
    FULL_IMAGE_NAME="${IMAGE_NAME}:${VERSION}-${TARGET}"
fi

print_status "Building Docker image: $FULL_IMAGE_NAME"
print_status "Target: $TARGET"
print_status "Platform: $PLATFORM"

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if buildx is available for multi-arch builds
if [[ "$PLATFORM" == *","* ]]; then
    if ! docker buildx version >/dev/null 2>&1; then
        print_error "Docker buildx is required for multi-architecture builds"
        exit 1
    fi
    
    # Create or use existing builder
    if ! docker buildx inspect ffprobe-builder >/dev/null 2>&1; then
        print_status "Creating new buildx builder..."
        docker buildx create --name ffprobe-builder --driver docker-container --bootstrap
    fi
    docker buildx use ffprobe-builder
fi

# Build Docker command
DOCKER_CMD="docker"
if [[ "$PLATFORM" == *","* ]]; then
    DOCKER_CMD="docker buildx"
fi

# Prepare build command
BUILD_CMD="$DOCKER_CMD build"
BUILD_CMD="$BUILD_CMD --file docker-image/Dockerfile"
BUILD_CMD="$BUILD_CMD --target $TARGET"
BUILD_CMD="$BUILD_CMD --tag $FULL_IMAGE_NAME"
BUILD_CMD="$BUILD_CMD --platform $PLATFORM"

# Add build arguments
if [[ -n "$BUILD_ARGS" ]]; then
    BUILD_CMD="$BUILD_CMD $BUILD_ARGS"
fi

# Add cache options
if [[ "$CACHE" == "false" ]]; then
    BUILD_CMD="$BUILD_CMD --no-cache"
else
    BUILD_CMD="$BUILD_CMD --cache-from ${IMAGE_NAME}:cache"
    BUILD_CMD="$BUILD_CMD --cache-to type=inline"
fi

# Add push option for multi-arch builds
if [[ "$PLATFORM" == *","* ]]; then
    if [[ "$PUSH" == "true" ]]; then
        BUILD_CMD="$BUILD_CMD --push"
    else
        BUILD_CMD="$BUILD_CMD --load"
    fi
elif [[ "$PUSH" == "true" ]]; then
    print_warning "Push option ignored for single-arch builds. Use 'docker push' after build."
fi

# Add build context
BUILD_CMD="$BUILD_CMD ."

print_status "Running: $BUILD_CMD"

# Execute build
if eval $BUILD_CMD; then
    print_success "Docker image built successfully: $FULL_IMAGE_NAME"
    
    # Show image details
    if [[ "$PLATFORM" != *","* ]]; then
        print_status "Image details:"
        docker images | grep "$IMAGE_NAME" | head -5
        
        # Show image size
        IMAGE_SIZE=$(docker images --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}" | grep "$FULL_IMAGE_NAME" | awk '{print $2}')
        print_status "Image size: $IMAGE_SIZE"
    fi
    
    # Push single-arch image if requested
    if [[ "$PUSH" == "true" && "$PLATFORM" != *","* ]]; then
        print_status "Pushing image to registry..."
        if docker push "$FULL_IMAGE_NAME"; then
            print_success "Image pushed successfully"
        else
            print_error "Failed to push image"
            exit 1
        fi
    fi
    
else
    print_error "Docker build failed"
    exit 1
fi

print_success "Build completed successfully!"

# Show next steps
echo ""
print_status "Next steps:"
if [[ "$TARGET" == "development" ]]; then
    echo "  docker run -p 8080:8080 -v \$(pwd):/app $FULL_IMAGE_NAME"
elif [[ "$TARGET" == "test" ]]; then
    echo "  docker run --rm $FULL_IMAGE_NAME"
else
    echo "  docker run -p 8080:8080 $FULL_IMAGE_NAME"
fi
echo "  docker-compose up  # or use docker-compose for full stack"