# Security Documentation

> **Comprehensive security guide for FFprobe API**

## Security Overview

FFprobe API implements multiple layers of security to protect your video analysis infrastructure:

- **Authentication**: JWT tokens and API keys
- **Authorization**: Role-based access control (RBAC)
- **Data Protection**: Encryption and secure storage
- **Network Security**: TLS, CORS, rate limiting
- **Input Validation**: Comprehensive sanitization

## Authentication Methods

### API Key Authentication

**Generation:**
```bash
# Generate secure API key (32+ characters)
openssl rand -hex 32
```

**Configuration:**
```bash
API_KEY=your-generated-32-character-api-key
```

**Usage:**
```bash
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/v1/probe/file
```

### JWT Token Authentication

**Configuration:**
```bash
JWT_SECRET=your-32-character-jwt-secret
TOKEN_EXPIRY_HOURS=24
REFRESH_EXPIRY_HOURS=168
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"email":"user@example.com","password":"secure-password"}'
```

**Usage:**
```bash
curl -H "Authorization: Bearer your-jwt-token" \
     http://localhost:8080/api/v1/probe/file
```

## Role-Based Access Control

### User Roles

| Role | Permissions | Use Case |
|------|------------|----------|
| `admin` | Full system access | System administrators |
| `user` | Standard API access | Regular users |
| `pro` | Enhanced features | Professional users |
| `premium` | All features | Enterprise users |
| `viewer` | Read-only access | Monitoring systems |

### Permission Matrix

| Endpoint | Admin | User | Pro | Premium | Viewer |
|----------|-------|------|-----|---------|--------|
| `/probe/file` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `/probe/url` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `/batch/analyze` | ✅ | ❌ | ✅ | ✅ | ❌ |
| `/quality/compare` | ✅ | ❌ | ✅ | ✅ | ❌ |
| `/admin/*` | ✅ | ❌ | ❌ | ❌ | ❌ |

## Security Configuration

### Password Requirements

```go
// Minimum password requirements
- Length: 8+ characters
- Complexity: Mixed case, numbers, special characters
- Hashing: bcrypt with cost factor 10
```

### Session Management

```bash
# Session configuration
SESSION_TIMEOUT=3600        # 1 hour
MAX_SESSIONS_PER_USER=5     # Concurrent sessions
ENABLE_SESSION_REFRESH=true # Auto-refresh tokens
```

## Network Security

### TLS/HTTPS Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    ssl_certificate /etc/ssl/certs/cert.pem;
    ssl_certificate_key /etc/ssl/private/key.pem;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
}
```

### CORS Configuration

```bash
# CORS settings
ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
ALLOWED_METHODS=GET,POST,PUT,DELETE
ALLOWED_HEADERS=Content-Type,Authorization,X-API-Key
EXPOSE_HEADERS=X-Request-ID,X-Rate-Limit-Remaining
```

### Rate Limiting

```bash
# Rate limit configuration
ENABLE_RATE_LIMIT=true
RATE_LIMIT_PER_MINUTE=60    # Per IP/User
RATE_LIMIT_PER_HOUR=1000
RATE_LIMIT_PER_DAY=10000
RATE_LIMIT_BURST=10         # Burst capacity
```

## Input Validation

### File Upload Security

```go
// File upload validation
- Max size: 50GB (configurable)
- Allowed extensions: mp4, mkv, avi, mov, webm
- MIME type validation
- Path traversal prevention
- Virus scanning (optional)
```

### Path Traversal Protection

```go
// Secure path handling
func ValidatePath(path string) error {
    // Remove path traversal attempts
    clean := filepath.Clean(path)
    
    // Ensure within allowed directory
    if !strings.HasPrefix(clean, UPLOAD_DIR) {
        return ErrInvalidPath
    }
    
    return nil
}
```

## Data Protection

### Database Security

```sql
-- User permissions
GRANT SELECT, INSERT, UPDATE ON analyses TO ffprobe_user;
REVOKE DELETE ON analyses FROM ffprobe_user;

-- Row-level security
ALTER TABLE analyses ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_analyses ON analyses
    FOR ALL
    USING (user_id = current_user_id());
```

### Encryption

```bash
# Encryption configuration
ENCRYPT_AT_REST=true         # Database encryption
ENCRYPT_IN_TRANSIT=true      # TLS for all connections
ENCRYPT_SENSITIVE_DATA=true  # Field-level encryption
```

### Secure Storage

```bash
# Storage security
SECURE_FILE_PERMISSIONS=0640  # Read for owner/group
SECURE_DIR_PERMISSIONS=0750   # Execute for owner/group
ENABLE_FILE_ENCRYPTION=true   # Encrypt uploaded files
```

## Audit Logging

### Audit Events

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "event_type": "authentication",
  "user_id": "user-123",
  "ip_address": "192.168.1.100",
  "action": "login",
  "result": "success",
  "metadata": {
    "method": "jwt",
    "user_agent": "Mozilla/5.0..."
  }
}
```

### Logged Events

- Authentication attempts (success/failure)
- Authorization failures
- API key usage
- Administrative actions
- Data access/modification
- Security configuration changes

## Security Headers

### Response Headers

```http
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
```

## Vulnerability Management

### Security Updates

```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy

# Docker base image updates
docker pull alpine:latest
docker build --no-cache -t ffprobe-api .
```

### Security Scanning

```bash
# Container scanning
trivy image ffprobe-api:latest

# Code scanning
gosec ./...

# Dependency scanning
snyk test
```

## Incident Response

### Security Incident Procedure

1. **Detection**: Monitor logs and alerts
2. **Containment**: Isolate affected systems
3. **Investigation**: Analyze logs and traces
4. **Remediation**: Apply fixes and patches
5. **Recovery**: Restore normal operations
6. **Post-mortem**: Document and improve

### Emergency Contacts

- Security Team: security@rendiff.dev
- On-call Engineer: Use PagerDuty
- Management Escalation: Define chain

## Security Best Practices

### Development

- Never commit secrets to version control
- Use environment variables for sensitive data
- Implement proper error handling
- Validate all user input
- Use prepared statements for SQL
- Keep dependencies updated

### Deployment

- Use least privilege principle
- Enable all security features
- Regular security audits
- Penetration testing
- Security training for team
- Incident response drills

### Monitoring

- Real-time security alerts
- Anomaly detection
- Failed authentication tracking
- Suspicious activity monitoring
- Regular log analysis
- Security metrics dashboard

## Compliance

### Standards

- **OWASP Top 10**: Protection against common vulnerabilities
- **GDPR**: Data privacy and protection
- **SOC 2**: Security controls and procedures
- **ISO 27001**: Information security management

### Data Privacy

```bash
# Privacy configuration
ENABLE_DATA_ANONYMIZATION=true
DATA_RETENTION_DAYS=90
ENABLE_RIGHT_TO_DELETE=true
ENABLE_DATA_EXPORT=true
```

---

## Security Checklist

- [ ] Changed default credentials
- [ ] Configured strong passwords
- [ ] Enabled TLS/HTTPS
- [ ] Set up firewall rules
- [ ] Configured rate limiting
- [ ] Enabled audit logging
- [ ] Set up monitoring alerts
- [ ] Tested backup/recovery
- [ ] Reviewed access controls
- [ ] Updated all dependencies

---

## Next Steps

- [Production Checklist](../deployment/production-checklist.md)
- [Monitoring Guide](monitoring.md)
- [Troubleshooting](troubleshooting.md)
- [API Authentication](../api/authentication.md)