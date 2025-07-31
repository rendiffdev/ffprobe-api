# ✅ Quick Start Verification Guide

This guide helps you verify that all the commands in the README actually work.

## 🧪 Test Environment Setup

### Step 1: Verify Prerequisites

```bash
# Check Docker and Docker Compose
docker --version
docker-compose --version

# Check required tools
openssl version
curl --version
jq --version || echo "jq not installed - install with: sudo apt-get install jq"
```

### Step 2: Test Repository Clone and Setup

```bash
# Test clone
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# Verify essential files exist
ls -la README.md docker-compose.yml .env.example Dockerfile.production

# Test environment setup
cp .env.example .env
echo "✅ Environment file created"
```

### Step 3: Test API Key Generation

```bash
# Generate and verify API key format
export API_KEY="ffprobe_test_sk_$(openssl rand -hex 32)"
echo "Generated API Key: $API_KEY"

# Verify key length (should be 79 characters: ffprobe_test_sk_ + 64 hex chars)
echo $API_KEY | wc -c

# Add to environment
echo "API_KEY=$API_KEY" >> .env
echo "ENABLE_AUTH=true" >> .env
```

### Step 4: Test Service Startup

```bash
# Start services in background
docker-compose up -d

# Wait for services to start
sleep 30

# Check service status
docker-compose ps
```

### Step 5: Test Health Endpoint

```bash
# Test without authentication (should fail if auth is enabled)
curl -w "HTTP Status: %{http_code}\n" http://localhost:8080/health

# Test with authentication (should succeed)
curl -H "X-API-Key: $API_KEY" -w "HTTP Status: %{http_code}\n" http://localhost:8080/health

# Expected response: {"status":"healthy","service":"ffprobe-api",...}
```

## 🎬 Test Video Analysis

### Step 6: Create Test Video

```bash
# Create a simple test video using FFmpeg
docker run --rm -v $(pwd):/work jrottenberg/ffmpeg:4.4-alpine \
  -f lavfi -i testsrc=duration=10:size=320x240:rate=30 \
  -f lavfi -i sine=frequency=1000:duration=10 \
  -c:v libx264 -c:a aac -shortest /work/test-video.mp4

# Verify test video was created
ls -lh test-video.mp4
```

### Step 7: Test Video Analysis API

```bash
# Test basic video analysis
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@test-video.mp4" \
  -w "HTTP Status: %{http_code}\n" \
  > analysis_result.json

# Check if analysis succeeded
cat analysis_result.json | jq '.status'

# Save analysis ID for comparison test
ANALYSIS_ID=$(cat analysis_result.json | jq -r '.id')
echo "Analysis ID: $ANALYSIS_ID"
```

### Step 8: Test Video Comparison (if two analyses exist)

```bash
# Create a second test video (slightly different)
docker run --rm -v $(pwd):/work jrottenberg/ffmpeg:4.4-alpine \
  -f lavfi -i testsrc=duration=10:size=320x240:rate=25 \
  -f lavfi -i sine=frequency=800:duration=10 \
  -c:v libx264 -c:a aac -shortest /work/test-video-2.mp4

# Analyze second video
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@test-video-2.mp4" \
  > analysis_result_2.json

ANALYSIS_ID_2=$(cat analysis_result_2.json | jq -r '.id')

# Test comparison
curl -X POST http://localhost:8080/api/v1/comparisons/quick \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"original_analysis_id\": \"$ANALYSIS_ID\",
    \"modified_analysis_id\": \"$ANALYSIS_ID_2\",
    \"include_llm\": false
  }" \
  > comparison_result.json

# Check comparison result
cat comparison_result.json | jq '.summary.quality_verdict // "Processing..."'
```

## 📊 Verification Checklist

Run this complete verification:

```bash
#!/bin/bash
# verification-script.sh

echo "🧪 FFprobe API Verification Script"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✅ $1 is installed${NC}"
        return 0
    else
        echo -e "${RED}❌ $1 is not installed${NC}"
        return 1
    fi
}

test_api_endpoint() {
    local endpoint=$1
    local expected_status=$2
    local description=$3
    
    local actual_status=$(curl -s -w "%{http_code}" -H "X-API-Key: $API_KEY" "$endpoint" -o /dev/null)
    
    if [ "$actual_status" -eq "$expected_status" ]; then
        echo -e "${GREEN}✅ $description (Status: $actual_status)${NC}"
        return 0
    else
        echo -e "${RED}❌ $description (Expected: $expected_status, Got: $actual_status)${NC}"
        return 1
    fi
}

echo "1. Checking prerequisites..."
check_command docker
check_command docker-compose
check_command curl
check_command openssl
check_command jq

echo -e "\n2. Checking environment setup..."
if [ -f ".env" ]; then
    echo -e "${GREEN}✅ .env file exists${NC}"
else
    echo -e "${YELLOW}⚠️  .env file not found, copying from example${NC}"
    cp .env.example .env
fi

# Generate API key if not set
if [ -z "$API_KEY" ]; then
    export API_KEY="ffprobe_test_sk_$(openssl rand -hex 32)"
    echo "API_KEY=$API_KEY" >> .env
    echo -e "${GREEN}✅ Generated API key${NC}"
fi

echo -e "\n3. Testing Docker services..."
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Services are running${NC}"
else
    echo -e "${YELLOW}⚠️  Starting services...${NC}"
    docker-compose up -d
    sleep 30
fi

echo -e "\n4. Testing API endpoints..."
test_api_endpoint "http://localhost:8080/health" 200 "Health endpoint"
test_api_endpoint "http://localhost:8080/metrics" 200 "Metrics endpoint"

echo -e "\n5. Testing authentication..."
# Test without auth (should fail with 401)
local no_auth_status=$(curl -s -w "%{http_code}" "http://localhost:8080/api/v1/probe/health" -o /dev/null)
if [ "$no_auth_status" -eq "401" ]; then
    echo -e "${GREEN}✅ Authentication is properly enforced${NC}"
else
    echo -e "${YELLOW}⚠️  Authentication might be disabled${NC}"
fi

echo -e "\n6. Creating test video..."
if [ ! -f "test-video.mp4" ]; then
    docker run --rm -v $(pwd):/work jrottenberg/ffmpeg:4.4-alpine \
      -f lavfi -i testsrc=duration=5:size=320x240:rate=30 \
      -f lavfi -i sine=frequency=1000:duration=5 \
      -c:v libx264 -c:a aac -shortest /work/test-video.mp4 &> /dev/null
    
    if [ -f "test-video.mp4" ]; then
        echo -e "${GREEN}✅ Test video created${NC}"
    else
        echo -e "${RED}❌ Failed to create test video${NC}"
    fi
else
    echo -e "${GREEN}✅ Test video already exists${NC}"
fi

echo -e "\n7. Testing video analysis..."
if [ -f "test-video.mp4" ]; then
    curl -X POST http://localhost:8080/api/v1/probe/file \
      -H "X-API-Key: $API_KEY" \
      -F "file=@test-video.mp4" \
      -s > test_analysis.json
    
    if [ -s "test_analysis.json" ] && jq -e '.id' test_analysis.json > /dev/null; then
        echo -e "${GREEN}✅ Video analysis working${NC}"
        echo "Analysis ID: $(jq -r '.id' test_analysis.json)"
    else
        echo -e "${RED}❌ Video analysis failed${NC}"
        echo "Response: $(cat test_analysis.json)"
    fi
fi

echo -e "\n🎉 Verification complete!"
echo "Check the results above. Green checkmarks (✅) indicate working features."
```

### Run the Verification Script

```bash
# Make script executable and run
chmod +x verification-script.sh
./verification-script.sh
```

## 🔧 Common Issues and Fixes

### Issue: Services won't start
```bash
# Check ports
netstat -tulpn | grep :8080

# Check Docker daemon
sudo systemctl status docker

# Check logs
docker-compose logs
```

### Issue: Authentication fails
```bash
# Verify API key format
echo $API_KEY | grep -E '^ffprobe_(test|live)_sk_[a-f0-9]{64}$'

# Check .env file
grep API_KEY .env

# Restart services after .env changes
docker-compose restart
```

### Issue: Video analysis fails
```bash
# Check file size limits
ls -lh test-video.mp4

# Check available disk space
df -h

# Check service logs
docker-compose logs ffprobe-api
```

### Issue: Database connection errors
```bash
# Check database service
docker-compose logs postgres

# Check database health
docker-compose exec postgres pg_isready

# Reset database if needed
docker-compose down -v
docker-compose up -d
```

## 📝 Expected Test Results

After running the verification, you should see:

✅ **All prerequisites installed**
✅ **Services running (docker-compose ps shows 'Up')**  
✅ **Health endpoint returns 200**
✅ **Authentication enforced (401 without API key)**
✅ **Test video created successfully**
✅ **Video analysis returns analysis ID**
✅ **All API endpoints respond correctly**

If any step fails, check the troubleshooting section above or create an issue with the error details.

---

This verification guide ensures that users can successfully follow the README instructions and get a working FFprobe API setup.