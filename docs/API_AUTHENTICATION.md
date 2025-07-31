# 🔐 API Authentication Guide

## Overview

The FFprobe API supports two authentication methods to secure your video analysis workflows:

1. **API Key Authentication** ⭐ **(Recommended)** - Simple, secure, perfect for production
2. **JWT Token Authentication** - For user-based applications with login/logout

## 🚀 Quick Start

### Option 1: API Key Authentication (Recommended)

**Step 1: Generate API Key**
```bash
# Generate secure API key
export API_KEY="ffprobe_test_sk_$(openssl rand -hex 32)"
echo "Generated API Key: $API_KEY"
```

**Step 2: Configure Environment**
```bash
# Add to .env file
echo "ENABLE_AUTH=true" >> .env
echo "API_KEY=$API_KEY" >> .env
```

**Step 3: Use API Key**
```bash
# Test authentication
curl -H "X-API-Key: $API_KEY" http://localhost:8080/health

# Analyze video with authentication
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@your-video.mp4"
```

### Option 2: JWT Token Authentication

**Step 1: Setup JWT Secret**
```bash
# Generate JWT secret
export JWT_SECRET="$(openssl rand -hex 32)"
echo "JWT_SECRET=$JWT_SECRET" >> .env
```

**Step 2: Login to Get Token**
```bash
# Default admin login (change password in production!)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "change-this-password"
  }'

# Response includes access_token
```

**Step 3: Use JWT Token**
```bash
# Store token from login response
export JWT_TOKEN="your-jwt-token-from-login"

# Use token in requests
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -F "file=@your-video.mp4"
```

## 🔑 API Key Management

### 1. Environment Setup

Create your environment file:

```bash
# .env
API_KEY=ffprobe_live_sk_1234567890abcdef1234567890abcdef12345678
JWT_SECRET=your-jwt-secret-key-minimum-32-characters-long
ENABLE_AUTH=true
```

**API Key Format**: `ffprobe_[env]_sk_[32-hex-chars]`
- `env`: `live` (production), `test` (development)
- `sk`: Secret Key identifier
- `32-hex-chars`: Random hexadecimal string

### 2. Generate API Keys

#### Option A: Manual Generation

```bash
# Generate secure API key
openssl rand -hex 32

# Format for production
echo "ffprobe_live_sk_$(openssl rand -hex 32)"

# Format for development  
echo "ffprobe_test_sk_$(openssl rand -hex 32)"
```

#### Option B: Using the API (After initial setup)

```bash
# Generate new API key via API
curl -X POST http://localhost:8080/api/v1/auth/api-key \
  -H "X-API-Key: your-existing-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Integration",
    "permissions": ["read", "write"],
    "expires_at": "2025-12-31T23:59:59Z"
  }'

# Response
{
  "id": "key_abc123",
  "key": "ffprobe_live_sk_def456...",
  "name": "Production Integration", 
  "permissions": ["read", "write"],
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

### 3. API Key Operations

#### List All Keys

```bash
curl -X GET http://localhost:8080/api/v1/auth/api-keys \
  -H "X-API-Key: your-master-key"

# Response
{
  "keys": [
    {
      "id": "key_abc123",
      "name": "Production Integration",
      "permissions": ["read", "write"],
      "created_at": "2024-01-15T10:30:00Z",
      "expires_at": "2025-12-31T23:59:59Z",
      "last_used": "2024-01-20T14:22:00Z"
    }
  ]
}
```

#### Revoke API Key

```bash
curl -X DELETE http://localhost:8080/api/v1/auth/api-keys/key_abc123 \
  -H "X-API-Key: your-master-key"

# Response
{
  "message": "API key revoked successfully",
  "revoked_at": "2024-01-20T15:00:00Z"
}
```

### 4. Permission Levels

| Permission | Description | Endpoints |
|------------|-------------|-----------|
| `read` | View analyses and results | `GET /api/v1/*` |
| `write` | Create new analyses | `POST /api/v1/probe/*` |
| `delete` | Delete analyses | `DELETE /api/v1/*` |
| `admin` | Full system access | All endpoints |

#### Example: Read-only Key

```bash
curl -X POST http://localhost:8080/api/v1/auth/api-key \
  -H "X-API-Key: your-admin-key" \
  -d '{
    "name": "Dashboard Viewer",
    "permissions": ["read"],
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

## 🛡️ JWT Authentication

### 1. User Login

```bash
# Login request
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "secure-password"
  }'

# Response
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "expires_at": "2024-01-15T11:30:00Z"
}
```

### 2. Using JWT Tokens

```bash
# Store token
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIs..."

# Make authenticated requests
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -F "file=@video.mp4"
```

### 3. Token Refresh

```bash
# Refresh expired token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your-refresh-token"
  }'

# New tokens response
{
  "access_token": "new-access-token",
  "refresh_token": "new-refresh-token",
  "expires_in": 3600
}
```

### 4. Token Validation

```bash
# Validate current token
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer $JWT_TOKEN"

# Response
{
  "valid": true,
  "expires_at": "2024-01-15T11:30:00Z",
  "user": {
    "id": "user_123",
    "username": "admin",
    "role": "admin"
  }
}
```

## 🔧 Configuration Options

### Environment Variables

```bash
# Authentication Settings
ENABLE_AUTH=true                 # Enable/disable authentication
API_KEY=your-api-key            # Master API key
JWT_SECRET=your-jwt-secret      # JWT signing secret (32+ chars)

# Token Expiration
TOKEN_EXPIRY=24                 # JWT token expiry (hours)
REFRESH_EXPIRY=168              # Refresh token expiry (hours)

# Rate Limiting (per API key/user)
RATE_LIMIT_PER_MINUTE=100       # Requests per minute
RATE_LIMIT_PER_HOUR=1000        # Requests per hour
RATE_LIMIT_PER_DAY=10000        # Requests per day
```

### Docker Compose Configuration

```yaml
# docker-compose.yml
services:
  ffprobe-api:
    environment:
      - ENABLE_AUTH=true
      - API_KEY=ffprobe_live_sk_${API_KEY_SUFFIX}
      - JWT_SECRET=${JWT_SECRET}
      - TOKEN_EXPIRY=24
      - REFRESH_EXPIRY=168
    env_file:
      - .env.production
```

## 🚨 Security Best Practices

### 1. API Key Security

✅ **Do**:
- Use environment variables for API keys
- Generate long, random keys (32+ characters)
- Set expiration dates for keys
- Use different keys for different environments
- Rotate keys regularly (every 90 days)
- Use read-only keys when possible

❌ **Don't**:
- Hardcode keys in source code
- Share keys in plain text
- Use the same key across environments
- Commit keys to version control
- Use weak or predictable keys

### 2. Production Setup

```bash
# Generate production secrets
export API_KEY="ffprobe_live_sk_$(openssl rand -hex 32)"
export JWT_SECRET="$(openssl rand -hex 32)"

# Store in secure environment file
cat > .env.production << EOF
API_KEY=${API_KEY}
JWT_SECRET=${JWT_SECRET}
ENABLE_AUTH=true
TOKEN_EXPIRY=8
REFRESH_EXPIRY=24
RATE_LIMIT_PER_MINUTE=60
EOF

# Secure the file
chmod 600 .env.production
```

### 3. Monitoring and Logging

The API automatically logs:
- Authentication attempts (success/failure)
- API key usage patterns
- Rate limit violations
- Token expiration events

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "API key authentication successful",
  "key_id": "key_abc123",
  "ip": "192.168.1.100",
  "endpoint": "/api/v1/probe/file"
}
```

## 📱 Integration Examples

### 1. Frontend Application (JWT)

```javascript
// Login and store tokens
async function login(username, password) {
  const response = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });
  
  const tokens = await response.json();
  localStorage.setItem('access_token', tokens.access_token);
  localStorage.setItem('refresh_token', tokens.refresh_token);
  return tokens;
}

// Make authenticated API calls
async function analyzeVideo(file) {
  const token = localStorage.getItem('access_token');
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch('/api/v1/probe/file', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: formData
  });
  
  return response.json();
}
```

### 2. Backend Service (API Key)

```python
import requests
import os

class FFprobeClient:
    def __init__(self):
        self.api_key = os.getenv('FFPROBE_API_KEY')
        self.base_url = os.getenv('FFPROBE_API_URL', 'http://localhost:8080')
        
    def analyze_file(self, file_path):
        headers = {'X-API-Key': self.api_key}
        
        with open(file_path, 'rb') as f:
            files = {'file': f}
            response = requests.post(
                f'{self.base_url}/api/v1/probe/file',
                headers=headers,
                files=files
            )
        
        return response.json()

# Usage
client = FFprobeClient()
result = client.analyze_file('video.mp4')
```

### 3. CLI Tool

```bash
#!/bin/bash
# ffprobe-cli.sh

API_KEY="${FFPROBE_API_KEY}"
BASE_URL="${FFPROBE_API_URL:-http://localhost:8080}"

if [ -z "$API_KEY" ]; then
    echo "Error: FFPROBE_API_KEY environment variable not set"
    exit 1
fi

# Analyze video file
analyze_video() {
    local file="$1"
    
    curl -X POST "${BASE_URL}/api/v1/probe/file" \
        -H "X-API-Key: ${API_KEY}" \
        -F "file=@${file}" \
        -s | jq '.'
}

# Usage: ./ffprobe-cli.sh video.mp4
analyze_video "$1"
```

## 🔍 Troubleshooting

### Common Issues

#### 1. "Invalid API Key" Error

```bash
# Check API key format
echo $API_KEY | grep -E '^ffprobe_(live|test)_sk_[a-f0-9]{64}$'

# Verify key is active
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "X-API-Key: $API_KEY"
```

#### 2. "Token Expired" Error

```bash
# Check token expiration
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer $JWT_TOKEN"

# Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -d '{"refresh_token": "your-refresh-token"}'
```

#### 3. Rate Limit Exceeded

```bash
# Check rate limit headers
curl -I http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY"

# Headers include:
# X-RateLimit-Limit: 100
# X-RateLimit-Remaining: 95
# X-RateLimit-Reset: 1642241400
```

### Error Response Format

```json
{
  "error": "authentication_failed",
  "message": "Invalid API key format",
  "code": 401,
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req_abc123"
}
```

## 📋 Complete Workflow Example

### Step 1: Initial Setup

```bash
# Generate API key
export API_KEY="ffprobe_live_sk_$(openssl rand -hex 32)"

# Start the API server
docker-compose up -d
```

### Step 2: Verify Authentication

```bash
# Test authentication
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "X-API-Key: $API_KEY"
```

### Step 3: Analyze Videos

```bash
# Analyze original video
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@original.mp4" \
  > original_analysis.json

# Analyze modified video  
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: $API_KEY" \
  -F "file=@modified.mp4" \
  > modified_analysis.json
```

### Step 4: Compare Results

```bash
# Extract analysis IDs
ORIGINAL_ID=$(jq -r '.id' original_analysis.json)
MODIFIED_ID=$(jq -r '.id' modified_analysis.json)

# Create comparison
curl -X POST http://localhost:8080/api/v1/comparisons/quick \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"original_analysis_id\": \"$ORIGINAL_ID\",
    \"modified_analysis_id\": \"$MODIFIED_ID\",
    \"include_llm\": true
  }"
```

## 🎯 API Key Naming Convention

Use descriptive names for API keys:

```bash
# Production keys
ffprobe_live_sk_production_web_app_2024
ffprobe_live_sk_batch_processor_main
ffprobe_live_sk_monitoring_system

# Development keys  
ffprobe_test_sk_local_development
ffprobe_test_sk_integration_tests
ffprobe_test_sk_staging_environment
```

## 📊 Usage Monitoring

Track API usage with built-in metrics:

```bash
# Get usage statistics
curl -X GET http://localhost:8080/api/v1/auth/usage \
  -H "X-API-Key: $API_KEY"

# Response
{
  "key_id": "key_abc123",
  "requests_today": 150,
  "requests_this_month": 4500,
  "rate_limit_hits": 5,
  "last_request": "2024-01-15T14:30:00Z",
  "quota_remaining": 95500
}
```

---

This comprehensive authentication guide covers all aspects of API security for the FFprobe API. Choose the authentication method that best fits your use case and follow the security best practices for production deployments.