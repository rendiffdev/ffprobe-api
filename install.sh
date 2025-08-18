#!/bin/bash

# FFprobe API - One-Line Installation Script
# Works on any OS with Docker - Zero configuration required
# Usage: curl -fsSL https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/install.sh | bash

set -e

# One-liner installer that downloads and runs the full setup
SETUP_URL="https://raw.githubusercontent.com/rendiffdev/ffprobe-api/main/setup.sh"

echo "🚀 FFprobe API - Quick Installation Starting..."
echo ""

# Download and run setup script
curl -fsSL "$SETUP_URL" | bash -s -- --quick

echo ""
echo "✅ Installation complete! Your FFprobe API is ready at http://localhost:8080"