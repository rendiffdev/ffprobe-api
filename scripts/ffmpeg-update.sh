#!/bin/bash

# FFmpeg Update Manager Script
# Uses BtbN's FFmpeg builds for consistent, full-featured binaries

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_API="https://api.github.com/repos/BtbN/FFmpeg-Builds/releases"
INSTALL_DIR="${FFMPEG_INSTALL_DIR:-/usr/local/bin}"
BACKUP_DIR="${INSTALL_DIR}/backup"
TEMP_DIR="/tmp/ffmpeg-update-$$"
VERSION_FILE="${INSTALL_DIR}/.ffmpeg-version"

# Logging
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

# Cleanup on exit
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Get current FFmpeg version
get_current_version() {
    if [ -f "$INSTALL_DIR/ffmpeg" ]; then
        version=$("$INSTALL_DIR/ffmpeg" -version 2>&1 | head -1 | grep -oP 'version \K[^ ]+' || echo "unknown")
        echo "$version"
    else
        echo "not_installed"
    fi
}

# Parse version for comparison
parse_version() {
    echo "$1" | sed -E 's/^n?([0-9]+)\.([0-9]+)\.?([0-9]+)?.*$/\1 \2 \3/'
}

# Compare versions
compare_versions() {
    local current="$1"
    local new="$2"
    
    IFS=' ' read -r curr_major curr_minor curr_patch <<< "$(parse_version "$current")"
    IFS=' ' read -r new_major new_minor new_patch <<< "$(parse_version "$new")"
    
    # Set defaults for missing values
    curr_major=${curr_major:-0}
    curr_minor=${curr_minor:-0}
    curr_patch=${curr_patch:-0}
    new_major=${new_major:-0}
    new_minor=${new_minor:-0}
    new_patch=${new_patch:-0}
    
    if [ "$new_major" -gt "$curr_major" ]; then
        echo "major"
    elif [ "$new_major" -eq "$curr_major" ] && [ "$new_minor" -gt "$curr_minor" ]; then
        echo "minor"
    elif [ "$new_major" -eq "$curr_major" ] && [ "$new_minor" -eq "$curr_minor" ] && [ "$new_patch" -gt "$curr_patch" ]; then
        echo "patch"
    else
        echo "none"
    fi
}

# Get latest release info from GitHub
get_latest_release() {
    local release_type="${1:-stable}"
    local api_url="$GITHUB_API/latest"
    
    if [ "$release_type" = "nightly" ]; then
        api_url="$GITHUB_API"
    fi
    
    curl -s "$api_url" | jq -r '
        if type == "array" then .[0] else . end |
        {
            tag: .tag_name,
            published: .published_at,
            assets: [.assets[] | select(.name | contains("linux64-gpl.tar.xz")) | {name: .name, url: .browser_download_url, size: .size}]
        }'
}

# Determine system architecture
get_system_arch() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        linux)
            case "$arch" in
                x86_64) echo "linux64" ;;
                aarch64) echo "linuxarm64" ;;
                *) echo "unsupported" ;;
            esac
            ;;
        darwin)
            case "$arch" in
                x86_64) echo "macos64" ;;
                arm64) echo "macosarm64" ;;
                *) echo "unsupported" ;;
            esac
            ;;
        *)
            echo "unsupported"
            ;;
    esac
}

# Download FFmpeg build
download_ffmpeg() {
    local url="$1"
    local dest="$2"
    
    log_info "Downloading FFmpeg from BtbN..."
    log_info "URL: $url"
    
    # Download with progress bar
    curl -L --progress-bar -o "$dest" "$url"
    
    if [ $? -eq 0 ]; then
        log_success "Download complete"
        return 0
    else
        log_error "Download failed"
        return 1
    fi
}

# Extract archive
extract_archive() {
    local archive="$1"
    local dest="$2"
    
    log_info "Extracting archive..."
    
    case "$archive" in
        *.tar.xz)
            tar -xJf "$archive" -C "$dest"
            ;;
        *.tar.gz)
            tar -xzf "$archive" -C "$dest"
            ;;
        *.zip)
            unzip -q "$archive" -d "$dest"
            ;;
        *)
            log_error "Unsupported archive format"
            return 1
            ;;
    esac
    
    log_success "Extraction complete"
}

# Test FFmpeg binary
test_ffmpeg() {
    local binary="$1"
    
    log_info "Testing FFmpeg binary..."
    
    # Basic version check
    if ! "$binary" -version &>/dev/null; then
        log_error "FFmpeg binary test failed: version check"
        return 1
    fi
    
    # Check for essential codecs
    local codecs_output=$("$binary" -codecs 2>/dev/null)
    local essential_codecs=("h264" "hevc" "aac" "mp3" "opus")
    
    for codec in "${essential_codecs[@]}"; do
        if ! echo "$codecs_output" | grep -q "$codec"; then
            log_warning "Essential codec missing: $codec"
        fi
    done
    
    # Check for essential filters
    local filters_output=$("$binary" -filters 2>/dev/null)
    local essential_filters=("scale" "overlay" "crop" "fps")
    
    for filter in "${essential_filters[@]}"; do
        if ! echo "$filters_output" | grep -q "$filter"; then
            log_warning "Essential filter missing: $filter"
        fi
    done
    
    log_success "FFmpeg binary tests passed"
    return 0
}

# Backup current installation
backup_current() {
    if [ ! -f "$INSTALL_DIR/ffmpeg" ]; then
        log_info "No existing installation to backup"
        return 0
    fi
    
    log_info "Backing up current FFmpeg installation..."
    
    mkdir -p "$BACKUP_DIR"
    
    # Backup with timestamp
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_suffix="backup_${timestamp}"
    
    for binary in ffmpeg ffprobe; do
        if [ -f "$INSTALL_DIR/$binary" ]; then
            cp "$INSTALL_DIR/$binary" "$BACKUP_DIR/${binary}.${backup_suffix}"
        fi
    done
    
    # Save version info
    get_current_version > "$BACKUP_DIR/version.${backup_suffix}"
    
    log_success "Backup complete: $BACKUP_DIR"
}

# Install new binaries
install_binaries() {
    local source_dir="$1"
    
    log_info "Installing new FFmpeg binaries..."
    
    # Find the extracted directory
    local ffmpeg_dir=$(find "$source_dir" -maxdepth 1 -type d -name "ffmpeg-*" | head -1)
    
    if [ -z "$ffmpeg_dir" ]; then
        log_error "Could not find extracted FFmpeg directory"
        return 1
    fi
    
    # Install binaries
    for binary in ffmpeg ffprobe; do
        if [ -f "$ffmpeg_dir/bin/$binary" ]; then
            cp "$ffmpeg_dir/bin/$binary" "$INSTALL_DIR/$binary"
            chmod +x "$INSTALL_DIR/$binary"
            log_success "Installed: $binary"
        else
            log_warning "Binary not found: $binary"
        fi
    done
    
    # Save version info
    get_current_version > "$VERSION_FILE"
    date -u +"%Y-%m-%d %H:%M:%S UTC" >> "$VERSION_FILE"
    
    log_success "Installation complete"
}

# Rollback to previous version
rollback() {
    log_info "Rolling back to previous version..."
    
    # Find most recent backup
    local latest_backup=$(ls -t "$BACKUP_DIR"/ffmpeg.backup_* 2>/dev/null | head -1)
    
    if [ -z "$latest_backup" ]; then
        log_error "No backup found to rollback to"
        return 1
    fi
    
    local backup_suffix=$(basename "$latest_backup" | sed 's/ffmpeg\.//')
    
    # Restore binaries
    for binary in ffmpeg ffprobe; do
        local backup_file="$BACKUP_DIR/${binary}.${backup_suffix}"
        if [ -f "$backup_file" ]; then
            cp "$backup_file" "$INSTALL_DIR/$binary"
            chmod +x "$INSTALL_DIR/$binary"
            log_success "Restored: $binary"
        fi
    done
    
    log_success "Rollback complete"
}

# Check for updates
check_updates() {
    local current_version=$(get_current_version)
    
    log_info "Current FFmpeg version: $current_version"
    log_info "Checking for updates from BtbN builds..."
    
    local latest_info=$(get_latest_release "stable")
    local latest_tag=$(echo "$latest_info" | jq -r '.tag')
    local published=$(echo "$latest_info" | jq -r '.published')
    
    log_info "Latest available: $latest_tag (published: $published)"
    
    # Check version difference
    local upgrade_type=$(compare_versions "$current_version" "$latest_tag")
    
    case "$upgrade_type" in
        major)
            log_warning "MAJOR upgrade available: $current_version → $latest_tag"
            log_warning "Major upgrades may contain breaking changes. Manual review recommended."
            echo "major"
            ;;
        minor)
            log_info "Minor upgrade available: $current_version → $latest_tag"
            echo "minor"
            ;;
        patch)
            log_info "Patch upgrade available: $current_version → $latest_tag"
            echo "patch"
            ;;
        none)
            log_success "You are on the latest version"
            echo "none"
            ;;
    esac
}

# Main update function
update_ffmpeg() {
    local force="${1:-false}"
    local allow_major="${2:-false}"
    
    # Check current version
    local current_version=$(get_current_version)
    
    if [ "$current_version" = "not_installed" ]; then
        log_info "FFmpeg not installed, performing fresh installation..."
        force="true"
    fi
    
    # Get latest release info
    local latest_info=$(get_latest_release "stable")
    local latest_tag=$(echo "$latest_info" | jq -r '.tag')
    local arch=$(get_system_arch)
    
    if [ "$arch" = "unsupported" ]; then
        log_error "Unsupported system architecture"
        exit 1
    fi
    
    # Find appropriate asset
    local asset_url=$(echo "$latest_info" | jq -r ".assets[] | select(.name | contains(\"${arch}-gpl\")) | .url" | head -1)
    
    if [ -z "$asset_url" ]; then
        log_error "No suitable FFmpeg build found for $arch"
        exit 1
    fi
    
    # Check if update is needed
    if [ "$force" != "true" ] && [ "$current_version" != "not_installed" ]; then
        local upgrade_type=$(compare_versions "$current_version" "$latest_tag")
        
        if [ "$upgrade_type" = "none" ]; then
            log_success "Already on the latest version"
            exit 0
        fi
        
        if [ "$upgrade_type" = "major" ] && [ "$allow_major" != "true" ]; then
            log_warning "Major upgrade available but not allowed (use --allow-major to proceed)"
            exit 0
        fi
    fi
    
    # Create temp directory
    mkdir -p "$TEMP_DIR"
    
    # Backup current installation
    backup_current
    
    # Download new version
    local archive_file="$TEMP_DIR/ffmpeg.tar.xz"
    if ! download_ffmpeg "$asset_url" "$archive_file"; then
        log_error "Failed to download FFmpeg"
        exit 1
    fi
    
    # Extract archive
    if ! extract_archive "$archive_file" "$TEMP_DIR"; then
        log_error "Failed to extract archive"
        exit 1
    fi
    
    # Test new binary before installation
    local test_binary=$(find "$TEMP_DIR" -path "*/bin/ffmpeg" | head -1)
    if [ -n "$test_binary" ] && ! test_ffmpeg "$test_binary"; then
        log_error "New FFmpeg binary failed tests"
        exit 1
    fi
    
    # Install new binaries
    if ! install_binaries "$TEMP_DIR"; then
        log_error "Installation failed, attempting rollback..."
        rollback
        exit 1
    fi
    
    # Verify installation
    local new_version=$(get_current_version)
    log_success "FFmpeg updated successfully: $current_version → $new_version"
}

# Parse command line arguments
case "${1:-}" in
    check)
        check_updates
        ;;
    update)
        shift
        force="false"
        allow_major="false"
        
        while [ $# -gt 0 ]; do
            case "$1" in
                --force) force="true" ;;
                --allow-major) allow_major="true" ;;
                *) log_error "Unknown option: $1"; exit 1 ;;
            esac
            shift
        done
        
        update_ffmpeg "$force" "$allow_major"
        ;;
    rollback)
        rollback
        ;;
    version)
        version=$(get_current_version)
        echo "FFmpeg version: $version"
        if [ -f "$VERSION_FILE" ]; then
            echo "Installed: $(tail -1 "$VERSION_FILE")"
        fi
        ;;
    test)
        if [ -f "$INSTALL_DIR/ffmpeg" ]; then
            test_ffmpeg "$INSTALL_DIR/ffmpeg"
        else
            log_error "FFmpeg not installed"
            exit 1
        fi
        ;;
    *)
        cat << EOF
FFmpeg Update Manager - BtbN Builds

Usage: $0 [command] [options]

Commands:
    check       Check for available updates
    update      Update FFmpeg to the latest version
    rollback    Rollback to the previous version
    version     Show current FFmpeg version
    test        Test FFmpeg installation

Options for 'update':
    --force         Force update even if on latest version
    --allow-major   Allow major version upgrades

Examples:
    $0 check                    # Check for updates
    $0 update                   # Update to latest (skip major)
    $0 update --allow-major     # Update including major versions
    $0 rollback                 # Rollback to previous version

Environment Variables:
    FFMPEG_INSTALL_DIR    Installation directory (default: /usr/local/bin)

EOF
        exit 1
        ;;
esac