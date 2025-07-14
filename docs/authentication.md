# Authentication and Security Documentation

This document describes the comprehensive authentication, authorization, and security features implemented in the ffprobe-api.

## Authentication Methods

The API supports multiple authentication methods that can be configured based on your requirements.

### 1. JWT Token Authentication

JSON Web Tokens (JWT) provide secure, stateless authentication with role-based access control.

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user_info": {
    "id": "user-123",
    "username": "admin",
    "roles": ["admin", "user"]
  }
}
```

#### Using JWT Token
```bash
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/path/to/video.mp4"}'
```

#### Token Refresh
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### 2. API Key Authentication

API keys provide simple, persistent authentication suitable for server-to-server communication.

#### Using API Key
```bash
# Header method (recommended)
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/path/to/video.mp4"}'

# Authorization header method
curl -X POST http://localhost:8080/api/v1/probe/file \
  -H "Authorization: ApiKey your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/path/to/video.mp4"}'

# Query parameter method (less secure)
curl -X POST "http://localhost:8080/api/v1/probe/file?api_key=your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"file_path": "/path/to/video.mp4"}'
```

#### Generate API Key
```bash
curl -X POST http://localhost:8080/api/v1/auth/api-key \
  -H "Authorization: Bearer your-jwt-token"
```

Response:
```json
{
  "api_key": "ffprobe_user123_abc123def456...",
  "created_at": 1647123456,
  "user_id": "user-123",
  "note": "Store this key securely. It will not be shown again."
}
```

## Role-Based Access Control (RBAC)

The API implements role-based access control with the following roles:

### User Roles

- **`user`**: Basic access to probe, upload, and analysis features
- **`pro`**: Enhanced rate limits and advanced features
- **`premium`**: Higher rate limits and priority processing
- **`admin`**: Full access to all features and system endpoints

### Role Requirements by Endpoint

| Endpoint Group | Required Roles | Description |
|---------------|----------------|-------------|
| `/api/v1/probe/*` | `user`, `pro`, `premium`, `admin` | Basic probe functionality |
| `/api/v1/upload/*` | `user`, `pro`, `premium`, `admin` | File upload features |
| `/api/v1/batch/*` | `user`, `pro`, `premium`, `admin` | Batch processing |
| `/api/v1/stream/*` | `user`, `pro`, `premium`, `admin` | Streaming analysis |
| `/api/v1/system/*` | `admin` | System administration |
| `/metrics` | `admin` | Prometheus metrics |

### Role-Based Rate Limits

Rate limits are automatically adjusted based on user roles:

| Role | Requests/Min | Requests/Hour | Requests/Day |
|------|-------------|---------------|--------------|
| `user` | 60 | 1,000 | 10,000 |
| `pro` | 180 | 3,000 | 30,000 |
| `premium` | 300 | 5,000 | 50,000 |
| `admin` | 600 | 10,000 | 100,000 |

## Security Features

### 1. Rate Limiting

Comprehensive rate limiting with multiple time windows:

```bash
# Check rate limit headers in response
curl -I http://localhost:8080/api/v1/probe/health \
  -H "X-API-Key: your-api-key"

# Response headers:
# X-RateLimit-Limit: 60
# X-RateLimit-Remaining: 59
# X-RateLimit-Reset: 1647123516
```

Rate limit exceeded response:
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retry_after": 1647123516
}
```

### 2. Security Headers

The API automatically adds comprehensive security headers:

- **Content-Security-Policy**: Prevents XSS attacks
- **X-Frame-Options**: Prevents clickjacking
- **X-Content-Type-Options**: Prevents MIME sniffing
- **X-XSS-Protection**: Browser XSS protection
- **Strict-Transport-Security**: Forces HTTPS (when TLS enabled)
- **Referrer-Policy**: Controls referrer information

### 3. CORS Configuration

Cross-Origin Resource Sharing is configurable:

```bash
# Environment variables
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

### 4. Input Sanitization

All input is automatically sanitized to prevent:
- SQL injection attacks
- XSS attacks
- Script injection
- Null byte injection

### 5. Threat Detection

The API includes basic threat detection for:
- SQL injection patterns
- XSS attempts
- Malicious bot signatures
- Suspicious user agents

Detected threats are logged and blocked:
```json
{
  "error": "Security threat detected",
  "code": "SQL_INJECTION"
}
```

## Configuration

### Environment Variables

```bash
# Authentication
ENABLE_AUTH=true
JWT_SECRET=your-super-secret-jwt-key-change-in-production
TOKEN_EXPIRY_HOURS=24
REFRESH_EXPIRY_HOURS=168
API_KEY=your-api-key

# Rate Limiting
ENABLE_RATE_LIMIT=true
RATE_LIMIT_PER_MINUTE=60
RATE_LIMIT_PER_HOUR=1000
RATE_LIMIT_PER_DAY=10000

# Security
ENABLE_CSRF=false
ALLOWED_ORIGINS=*
TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  ffprobe-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENABLE_AUTH=true
      - JWT_SECRET=${JWT_SECRET}
      - API_KEY=${API_KEY}
      - ENABLE_RATE_LIMIT=true
      - RATE_LIMIT_PER_MINUTE=60
      - ALLOWED_ORIGINS=https://yourdomain.com
    volumes:
      - uploads:/app/uploads
```

## Authentication Endpoints

### Complete Authentication API

```bash
# Login
POST /api/v1/auth/login
# Refresh token
POST /api/v1/auth/refresh
# Logout (requires auth)
POST /api/v1/auth/logout
# Get profile (requires auth)
GET /api/v1/auth/profile
# Change password (requires auth)
POST /api/v1/auth/change-password
# Validate token (requires auth)
GET /api/v1/auth/validate
# Generate API key (requires auth)
POST /api/v1/auth/api-key
# List API keys (requires auth)
GET /api/v1/auth/api-keys
# Revoke API key (requires auth)
DELETE /api/v1/auth/api-keys/:id
```

### User Profile Management

```bash
# Get current user profile
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer your-jwt-token"

# Change password
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Authorization: Bearer your-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "oldpassword",
    "new_password": "newpassword123",
    "confirm_password": "newpassword123"
  }'
```

### API Key Management

```bash
# List all API keys for user
curl -X GET http://localhost:8080/api/v1/auth/api-keys \
  -H "Authorization: Bearer your-jwt-token"

# Revoke an API key
curl -X DELETE http://localhost:8080/api/v1/auth/api-keys/key_1 \
  -H "Authorization: Bearer your-jwt-token"
```

## Monitoring and Metrics

### Authentication Metrics

The API provides Prometheus metrics for authentication:

- `auth_failures_total{reason}` - Authentication failures by reason
- `http_requests_total{method,endpoint,status}` - Request counts
- `rate_limit_exceeded_total{identifier_type}` - Rate limit violations

### Security Monitoring

```bash
# View security metrics
curl http://localhost:8080/metrics \
  -H "Authorization: Bearer admin-token"
```

Key security metrics:
- Failed authentication attempts
- Rate limit violations
- Detected threats
- Active sessions
- API key usage

## Error Codes

### Authentication Errors

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `MISSING_API_KEY` | API key not provided | 401 |
| `INVALID_API_KEY` | API key is invalid | 401 |
| `MISSING_TOKEN` | JWT token not provided | 401 |
| `INVALID_TOKEN` | JWT token is invalid/expired | 401 |
| `MISSING_ROLES` | Role information not found | 403 |
| `INSUFFICIENT_PERMISSIONS` | User lacks required role | 403 |

### Security Errors

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded | 429 |
| `CSRF_TOKEN_INVALID` | CSRF token validation failed | 403 |
| `IP_NOT_WHITELISTED` | IP not in whitelist | 403 |
| `GEO_RESTRICTED` | Access restricted by location | 403 |
| `SQL_INJECTION` | SQL injection attempt detected | 403 |
| `XSS_ATTEMPT` | XSS attempt detected | 403 |
| `MALICIOUS_BOT` | Malicious bot detected | 403 |

## Production Security Checklist

### Required for Production

- [ ] Change default JWT secret
- [ ] Use strong, unique API keys
- [ ] Enable HTTPS/TLS
- [ ] Configure proper CORS origins
- [ ] Set up rate limiting
- [ ] Enable security headers
- [ ] Configure trusted proxies
- [ ] Set up monitoring and alerting
- [ ] Regular security updates
- [ ] Implement password policies

### Optional Enhancements

- [ ] Enable CSRF protection
- [ ] Configure geo-restrictions
- [ ] Set up IP whitelisting
- [ ] Implement 2FA
- [ ] Add OAuth2/OIDC integration
- [ ] Set up WAF (Web Application Firewall)
- [ ] Implement audit logging
- [ ] Add intrusion detection

## Troubleshooting

### Common Issues

1. **Token Expired**
   - Use refresh token to get new access token
   - Check token expiry configuration

2. **Rate Limit Exceeded**
   - Wait for rate limit reset
   - Upgrade user role for higher limits
   - Implement exponential backoff

3. **CORS Issues**
   - Check allowed origins configuration
   - Verify request headers
   - Enable preflight handling

4. **Authentication Failed**
   - Verify credentials
   - Check API key format
   - Validate JWT token structure

### Debug Mode

```bash
# Enable debug logging
LOG_LEVEL=debug

# Check authentication flow
curl -v http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer your-token"
```