#!/bin/bash
# FFprobe API - Production-Optimized Docker Build Script
# Enterprise-grade build automation with security scanning, multi-arch support, and CI/CD integration
# Version: 2.0

set -euo pipefail

# =============================================================================
# Configuration and Constants
# =============================================================================

# Build configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "${SCRIPT_DIR}")"
readonly BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
readonly BUILD_USER="${USER:-builder}"
readonly BUILD_HOST="${HOSTNAME:-localhost}"

# Default values
IMAGE_NAME="${IMAGE_NAME:-rendiff-probe}"
VERSION="${VERSION:-latest}"
TARGET="${TARGET:-production}"
PLATFORM="${PLATFORM:-linux/amd64}"
REGISTRY="${REGISTRY:-}"
PUSH="${PUSH:-false}"
SCAN_SECURITY="${SCAN_SECURITY:-true}"
BUILD_CACHE="${BUILD_CACHE:-true}"
MULTI_ARCH="${MULTI_ARCH:-false}"
PARALLEL_BUILD="${PARALLEL_BUILD:-true}"
BUILD_ARGS="${BUILD_ARGS:-}"

# Security and quality settings
SIGN_IMAGE="${SIGN_IMAGE:-false}"
SBOM_GENERATION="${SBOM_GENERATION:-true}"
VULNERABILITY_SCAN="${VULNERABILITY_SCAN:-true}"
PERFORMANCE_TEST="${PERFORMANCE_TEST:-false}"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# =============================================================================
# Utility Functions
# =============================================================================

print_banner() {
    echo -e "${CYAN}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  FFprobe API - Production-Optimized Docker Build System v2.0"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${NC}"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
FFprobe API - Production-Optimized Docker Build Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    Build Configuration:
        -n, --name NAME           Docker image name (default: rendiff-probe)
        -v, --version VERSION     Image version tag (default: latest)
        -t, --target TARGET       Build target (production|development|test|minimal|security-scan)
        -r, --registry REGISTRY   Container registry URL
        
    Platform & Architecture:
        -p, --platform PLATFORM   Build platform (default: linux/amd64)
        --multi-arch              Build for multiple architectures (amd64,arm64)
        --parallel                Enable parallel builds (default: true)
        
    Build Options:
        --push                    Push image to registry after build
        --no-cache                Disable Docker build cache
        --build-arg ARG=VALUE     Pass build argument (repeatable)
        
    Security & Quality:
        --scan                    Enable security vulnerability scanning
        --sign                    Sign container image (requires cosign)
        --sbom                    Generate Software Bill of Materials
        --performance-test        Run performance benchmarks
        
    Utility:
        --dry-run                 Show what would be executed without running
        -h, --help               Show this help message
        --debug                  Enable debug output

EXAMPLES:
    # Basic production build
    $0 --target production
    
    # Multi-architecture build with push
    $0 --multi-arch --push --registry ghcr.io/company
    
    # Security-focused build with scanning and signing
    $0 --target production --scan --sign --sbom
    
    # Development build with custom build args
    $0 --target development --build-arg GO_VERSION=1.23 --build-arg DEBUG=true
    
    # Complete CI/CD pipeline build
    $0 --target production --multi-arch --push --scan --sign --sbom --performance-test

TARGETS:
    production      Production-ready image with security hardening
    development     Development image with hot reload and debugging tools
    test           Testing image with coverage tools and test dependencies
    minimal        Ultra-minimal image built from scratch
    security-scan  Production image with embedded security scanning tools

ENVIRONMENT VARIABLES:
    IMAGE_NAME              Docker image name
    VERSION                 Image version tag
    REGISTRY                Container registry URL
    BUILD_CACHE             Enable/disable build cache (true/false)
    SCAN_SECURITY           Enable security scanning (true/false)
    DOCKER_BUILDKIT         Enable BuildKit (recommended: 1)
    
EOF
}

# Parse command line arguments
parse_arguments() {
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
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            --multi-arch)
                MULTI_ARCH=true
                PLATFORM="linux/amd64,linux/arm64"
                shift
                ;;
            --push)
                PUSH=true
                shift
                ;;
            --no-cache)
                BUILD_CACHE=false
                shift
                ;;
            --build-arg)
                if [[ -n "${BUILD_ARGS}" ]]; then
                    BUILD_ARGS="${BUILD_ARGS} --build-arg $2"
                else
                    BUILD_ARGS="--build-arg $2"
                fi
                shift 2
                ;;
            --scan)
                SCAN_SECURITY=true
                shift
                ;;
            --sign)
                SIGN_IMAGE=true
                shift
                ;;
            --sbom)
                SBOM_GENERATION=true
                shift
                ;;
            --performance-test)
                PERFORMANCE_TEST=true
                shift
                ;;
            --parallel)
                PARALLEL_BUILD=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --debug)
                set -x
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# Validation Functions
# =============================================================================

validate_environment() {
    log_step "Validating build environment"
    
    # Check Docker availability
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check Docker daemon
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    # Check BuildKit support
    if [[ "${DOCKER_BUILDKIT:-}" != "1" ]]; then
        log_warning "DOCKER_BUILDKIT not enabled. Setting DOCKER_BUILDKIT=1"
        export DOCKER_BUILDKIT=1
    fi
    
    # Check for buildx if multi-arch
    if [[ "${MULTI_ARCH}" == "true" || "${PLATFORM}" == *","* ]]; then
        if ! command -v docker &> /dev/null || ! docker buildx version &> /dev/null; then
            log_error "docker buildx is required for multi-architecture builds"
            exit 1
        fi
    fi
    
    # Validate target
    case "${TARGET}" in
        production|development|test|minimal|security-scan)
            ;;
        *)
            log_error "Invalid target: ${TARGET}"
            log_error "Valid targets: production, development, test, minimal, security-scan"
            exit 1
            ;;
    esac
    
    # Check for security tools if scanning enabled
    if [[ "${SCAN_SECURITY}" == "true" ]]; then
        if ! command -v trivy &> /dev/null; then
            log_warning "Trivy not found. Installing via Docker for security scanning"
        fi
    fi
    
    # Check for cosign if signing enabled
    if [[ "${SIGN_IMAGE}" == "true" ]]; then
        if ! command -v cosign &> /dev/null; then
            log_error "cosign not found. Required for image signing"
            exit 1
        fi
    fi
    
    log_success "Environment validation completed"
}

validate_build_context() {
    log_step "Validating build context"
    
    # Check if we're in the right directory
    if [[ ! -f "${PROJECT_ROOT}/go.mod" ]]; then
        log_error "go.mod not found. Please run from project root or check paths"
        exit 1
    fi
    
    # Check for Dockerfile
    if [[ ! -f "${SCRIPT_DIR}/Dockerfile.optimized" ]]; then
        log_error "Optimized Dockerfile not found at ${SCRIPT_DIR}/Dockerfile.optimized"
        exit 1
    fi
    
    # Check for sensitive files that shouldn't be in build context
    local sensitive_files=(".env" "*.key" "*.pem" "*.p12")
    for pattern in "${sensitive_files[@]}"; do
        if compgen -G "${PROJECT_ROOT}/${pattern}" > /dev/null; then
            log_warning "Sensitive files matching '${pattern}' found in build context"
        fi
    done
    
    log_success "Build context validation completed"
}

# =============================================================================
# Build Functions
# =============================================================================

setup_buildx() {
    if [[ "${MULTI_ARCH}" == "true" ]]; then
        log_step "Setting up Docker Buildx for multi-architecture builds"
        
        local builder_name="ffprobe-builder-optimized"
        
        # Create or use existing builder
        if ! docker buildx inspect "${builder_name}" &> /dev/null; then
            log_info "Creating new buildx builder: ${builder_name}"
            docker buildx create \
                --name "${builder_name}" \
                --driver docker-container \
                --use \
                --bootstrap
        else
            log_info "Using existing buildx builder: ${builder_name}"
            docker buildx use "${builder_name}"
        fi
        
        # Verify builder platforms
        log_info "Available platforms:"
        docker buildx inspect --bootstrap
    fi
}

calculate_build_metadata() {
    log_step "Calculating build metadata"
    
    # Git information
    if command -v git &> /dev/null && [[ -d "${PROJECT_ROOT}/.git" ]]; then
        GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
        GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
        GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "")
        GIT_DIRTY=$(git diff --quiet 2>/dev/null || echo "-dirty")
    else
        GIT_COMMIT="unknown"
        GIT_BRANCH="unknown"
        GIT_TAG=""
        GIT_DIRTY=""
    fi
    
    # Version handling
    if [[ -n "${GIT_TAG}" ]]; then
        VERSION="${GIT_TAG}"
    elif [[ "${VERSION}" == "latest" && "${GIT_COMMIT}" != "unknown" ]]; then
        VERSION="${GIT_BRANCH}-${GIT_COMMIT}${GIT_DIRTY}"
    fi
    
    # Full image name
    if [[ -n "${REGISTRY}" ]]; then
        FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${VERSION}"
    else
        FULL_IMAGE_NAME="${IMAGE_NAME}:${VERSION}"
    fi
    
    # Add target suffix for non-production builds
    if [[ "${TARGET}" != "production" ]]; then
        FULL_IMAGE_NAME="${FULL_IMAGE_NAME}-${TARGET}"
    fi
    
    log_info "Image: ${FULL_IMAGE_NAME}"
    log_info "Target: ${TARGET}"
    log_info "Platform: ${PLATFORM}"
    log_info "Git: ${GIT_BRANCH}@${GIT_COMMIT}${GIT_DIRTY}"
}

build_image() {
    log_step "Building Docker image"
    
    local dockerfile="${SCRIPT_DIR}/Dockerfile.optimized"
    local build_cmd="docker"
    
    # Use buildx for multi-arch or if explicitly requested
    if [[ "${MULTI_ARCH}" == "true" || "${PLATFORM}" == *","* ]]; then
        build_cmd="docker buildx"
    fi
    
    # Construct build command
    local cmd="${build_cmd} build"
    cmd+=" --file ${dockerfile}"
    cmd+=" --target ${TARGET}"
    cmd+=" --platform ${PLATFORM}"
    cmd+=" --tag ${FULL_IMAGE_NAME}"
    
    # Add build arguments
    cmd+=" --build-arg VERSION=${VERSION}"
    cmd+=" --build-arg COMMIT=${GIT_COMMIT}"
    cmd+=" --build-arg BUILD_DATE=${BUILD_DATE}"
    cmd+=" --build-arg BUILD_USER=${BUILD_USER}"
    cmd+=" --build-arg BUILD_HOST=${BUILD_HOST}"
    
    # Add custom build args
    if [[ -n "${BUILD_ARGS}" ]]; then
        cmd+=" ${BUILD_ARGS}"
    fi
    
    # Cache configuration
    if [[ "${BUILD_CACHE}" == "true" ]]; then
        cmd+=" --cache-from type=registry,ref=${FULL_IMAGE_NAME}-cache"
        cmd+=" --cache-to type=registry,ref=${FULL_IMAGE_NAME}-cache,mode=max"
    else
        cmd+=" --no-cache"
    fi
    
    # Output configuration
    if [[ "${MULTI_ARCH}" == "true" ]]; then
        if [[ "${PUSH}" == "true" ]]; then
            cmd+=" --push"
        else
            log_warning "Multi-arch builds require --push. Adding --push automatically"
            cmd+=" --push"
            PUSH=true
        fi
    else
        cmd+=" --load"
    fi
    
    # SBOM generation
    if [[ "${SBOM_GENERATION}" == "true" ]]; then
        cmd+=" --sbom=true"
        cmd+=" --provenance=true"
    fi
    
    # Build context
    cmd+=" ${PROJECT_ROOT}"
    
    log_info "Build command: ${cmd}"
    
    # Execute build
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "DRY RUN: Would execute: ${cmd}"
    else
        local build_start=$(date +%s)
        if eval "${cmd}"; then
            local build_end=$(date +%s)
            local build_duration=$((build_end - build_start))
            log_success "Image built successfully in ${build_duration}s: ${FULL_IMAGE_NAME}"
        else
            log_error "Docker build failed"
            exit 1
        fi
    fi
}

# =============================================================================
# Security and Quality Functions
# =============================================================================

scan_vulnerabilities() {
    if [[ "${SCAN_SECURITY}" != "true" ]]; then
        return 0
    fi
    
    log_step "Scanning for security vulnerabilities"
    
    # Use Trivy for vulnerability scanning
    local scan_cmd="docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
        aquasec/trivy:latest image \
        --exit-code 1 \
        --severity HIGH,CRITICAL \
        --format table \
        ${FULL_IMAGE_NAME}"
    
    log_info "Running security scan: ${scan_cmd}"
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "DRY RUN: Would scan image for vulnerabilities"
    else
        if eval "${scan_cmd}"; then
            log_success "Security scan passed: No HIGH or CRITICAL vulnerabilities found"
        else
            log_error "Security scan failed: HIGH or CRITICAL vulnerabilities detected"
            log_error "Please review the scan results and address vulnerabilities before proceeding"
            exit 1
        fi
    fi
}

sign_image() {
    if [[ "${SIGN_IMAGE}" != "true" ]]; then
        return 0
    fi
    
    log_step "Signing container image"
    
    local sign_cmd="cosign sign --yes ${FULL_IMAGE_NAME}"
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "DRY RUN: Would sign image with cosign"
    else
        if eval "${sign_cmd}"; then
            log_success "Image signed successfully"
        else
            log_error "Image signing failed"
            exit 1
        fi
    fi
}

generate_sbom() {
    if [[ "${SBOM_GENERATION}" != "true" ]]; then
        return 0
    fi
    
    log_step "Generating Software Bill of Materials (SBOM)"
    
    local sbom_file="${PROJECT_ROOT}/sbom-${VERSION}.json"
    local sbom_cmd="docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
        anchore/syft:latest \
        ${FULL_IMAGE_NAME} \
        -o spdx-json=${sbom_file}"
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "DRY RUN: Would generate SBOM"
    else
        if eval "${sbom_cmd}"; then
            log_success "SBOM generated: ${sbom_file}"
        else
            log_warning "SBOM generation failed"
        fi
    fi
}

run_performance_tests() {
    if [[ "${PERFORMANCE_TEST}" != "true" ]]; then
        return 0
    fi
    
    log_step "Running performance benchmarks"
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "DRY RUN: Would run performance tests"
        return 0
    fi
    
    # Basic container startup time test
    log_info "Testing container startup time..."
    local start_time=$(date +%s%N)
    
    if docker run --rm --name "perf-test-$$" "${FULL_IMAGE_NAME}" /app/rendiff-probe --version &> /dev/null; then
        local end_time=$(date +%s%N)
        local startup_ms=$(( (end_time - start_time) / 1000000 ))
        log_success "Container startup time: ${startup_ms}ms"
        
        if [[ ${startup_ms} -gt 5000 ]]; then
            log_warning "Startup time exceeds 5000ms. Consider optimizing image size or startup sequence"
        fi
    else
        log_warning "Performance test failed"
    fi
}

# =============================================================================
# Push and Registry Functions
# =============================================================================

push_image() {
    if [[ "${PUSH}" != "true" || "${MULTI_ARCH}" == "true" ]]; then
        return 0  # Multi-arch builds push automatically
    fi
    
    log_step "Pushing image to registry"
    
    if [[ -z "${REGISTRY}" ]]; then
        log_error "Registry not specified. Cannot push image"
        exit 1
    fi
    
    local push_cmd="docker push ${FULL_IMAGE_NAME}"
    
    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "DRY RUN: Would push image to registry"
    else
        if eval "${push_cmd}"; then
            log_success "Image pushed successfully: ${FULL_IMAGE_NAME}"
        else
            log_error "Failed to push image"
            exit 1
        fi
    fi
}

# =============================================================================
# Information and Cleanup Functions
# =============================================================================

show_image_info() {
    if [[ "${DRY_RUN:-false}" == "true" || "${MULTI_ARCH}" == "true" ]]; then
        return 0
    fi
    
    log_step "Image information"
    
    # Show image details
    docker images --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | \
        grep -E "(REPOSITORY|${IMAGE_NAME})" | head -6
    
    # Show image layers (if not multi-arch)
    log_info "Image layers:"
    docker history "${FULL_IMAGE_NAME}" --format "table {{.CreatedBy}}\t{{.Size}}" | head -10
    
    # Show image metadata
    if command -v jq &> /dev/null; then
        log_info "Image metadata:"
        docker inspect "${FULL_IMAGE_NAME}" | jq -r '.[0].Config.Labels // {}'
    fi
}

cleanup_build_cache() {
    log_step "Cleaning up build cache"
    
    # Remove dangling images
    if docker images -f "dangling=true" -q | grep -q .; then
        log_info "Removing dangling images..."
        docker images -f "dangling=true" -q | xargs docker rmi || true
    fi
    
    # Prune build cache (keep recent)
    docker builder prune -f --keep-storage 10GB || true
    
    log_success "Build cache cleanup completed"
}

show_next_steps() {
    log_step "Build completed successfully!"
    
    echo ""
    echo -e "${GREEN}🎉 BUILD SUMMARY${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "📦 Image:     ${CYAN}${FULL_IMAGE_NAME}${NC}"
    echo -e "🎯 Target:    ${YELLOW}${TARGET}${NC}"
    echo -e "🏗️  Platform:  ${BLUE}${PLATFORM}${NC}"
    echo -e "🔧 Git:       ${PURPLE}${GIT_BRANCH}@${GIT_COMMIT}${GIT_DIRTY}${NC}"
    echo -e "⏰ Built:     ${BUILD_DATE}"
    echo ""
    
    echo -e "${GREEN}🚀 NEXT STEPS${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    case "${TARGET}" in
        production)
            echo "🏭 Production deployment:"
            echo "   docker run -p 8080:8080 ${FULL_IMAGE_NAME}"
            echo "   docker-compose -f compose.production.optimized.yaml up"
            ;;
        development)
            echo "🔧 Development usage:"
            echo "   docker run -p 8080:8080 -v \$(pwd):/app ${FULL_IMAGE_NAME}"
            ;;
        test)
            echo "🧪 Testing completed. Review coverage reports."
            ;;
        minimal)
            echo "📦 Minimal image ready for high-density deployments"
            ;;
    esac
    
    if [[ "${PUSH}" == "true" ]]; then
        echo "☁️  Image available at: ${REGISTRY}"
    fi
    
    echo ""
}

# =============================================================================
# Main Execution
# =============================================================================

main() {
    print_banner
    
    # Parse command line arguments
    parse_arguments "$@"
    
    # Validation phase
    validate_environment
    validate_build_context
    
    # Setup phase
    setup_buildx
    calculate_build_metadata
    
    # Build phase
    build_image
    
    # Quality assurance phase
    scan_vulnerabilities
    generate_sbom
    run_performance_tests
    
    # Deployment phase
    push_image
    sign_image
    
    # Information and cleanup
    show_image_info
    cleanup_build_cache
    show_next_steps
}

# Trap cleanup on exit
trap 'cleanup_build_cache' EXIT

# Execute main function with all arguments
main "$@"