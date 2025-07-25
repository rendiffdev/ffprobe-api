#!/bin/sh
# Production health check script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_NAME="${POSTGRES_DB:-ffprobe_api}"
DB_USER="${POSTGRES_USER:-ffprobe}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"

echo "🏥 FFprobe API Health Check"
echo "=========================="

# Check API health
echo -n "API Health: "
if curl -f -s "${API_URL}/health" > /dev/null; then
    echo -e "${GREEN}✓ Healthy${NC}"
else
    echo -e "${RED}✗ Unhealthy${NC}"
    exit 1
fi

# Check database connectivity
echo -n "Database: "
if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Connected${NC}"
else
    echo -e "${RED}✗ Disconnected${NC}"
    exit 1
fi

# Check Redis connectivity
echo -n "Redis: "
if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Connected${NC}"
else
    echo -e "${RED}✗ Disconnected${NC}"
    exit 1
fi

# Check FFmpeg availability
echo -n "FFmpeg: "
if command -v ffmpeg > /dev/null 2>&1; then
    VERSION=$(ffmpeg -version | head -n1)
    echo -e "${GREEN}✓ Available${NC} ($VERSION)"
else
    echo -e "${RED}✗ Not found${NC}"
    exit 1
fi

# Check FFprobe availability
echo -n "FFprobe: "
if command -v ffprobe > /dev/null 2>&1; then
    VERSION=$(ffprobe -version | head -n1)
    echo -e "${GREEN}✓ Available${NC} ($VERSION)"
else
    echo -e "${RED}✗ Not found${NC}"
    exit 1
fi

# Check disk space
echo -n "Disk Space: "
DISK_USAGE=$(df -h /app | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -lt 90 ]; then
    echo -e "${GREEN}✓ OK${NC} (${DISK_USAGE}% used)"
else
    echo -e "${YELLOW}⚠ Warning${NC} (${DISK_USAGE}% used)"
fi

# Check memory usage
echo -n "Memory: "
if command -v free > /dev/null 2>&1; then
    MEM_USAGE=$(free | grep Mem | awk '{print int($3/$2 * 100)}')
    if [ "$MEM_USAGE" -lt 90 ]; then
        echo -e "${GREEN}✓ OK${NC} (${MEM_USAGE}% used)"
    else
        echo -e "${YELLOW}⚠ Warning${NC} (${MEM_USAGE}% used)"
    fi
else
    echo -e "${YELLOW}⚠ Cannot check${NC}"
fi

echo "=========================="
echo -e "${GREEN}All systems operational!${NC}"
exit 0