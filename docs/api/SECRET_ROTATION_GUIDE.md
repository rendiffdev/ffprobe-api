# Secret Rotation & Rate Limiting Guide

## Overview

This guide covers the implementation of secure API key rotation, JWT secret rotation, and per-user/tenant rate limiting for the FFprobe API platform.

## Table of Contents

1. [API Key Management](#api-key-management)
2. [JWT Secret Rotation](#jwt-secret-rotation)
3. [Per-User/Tenant Rate Limiting](#per-usertenant-rate-limiting)
4. [Database Schema](#database-schema)
5. [API Endpoints](#api-endpoints)
6. [Best Practices](#best-practices)
7. [Monitoring & Alerts](#monitoring--alerts)

---

## API Key Management

### Key Features

- **Secure Generation**: 256-bit cryptographically secure random keys
- **Bcrypt Hashing**: Keys are hashed before storage
- **Automatic Rotation**: 90-day rotation cycle with 7-day grace period
- **Per-Key Rate Limits**: Customizable rate limits per API key
- **Audit Logging**: Complete lifecycle tracking

### Key Format

```
ffprobe_[environment]_sk_[64-character-hex-string]
```

Example:
```
ffprobe_production_sk_a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890
```

### Creating an API Key

```bash
curl -X POST https://api.example.com/api/v1/keys/create \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production API Key",
    "permissions": ["read", "write"],
    "expires_in_days": 90
  }'
```

Response:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Production API Key",
    "key": "ffprobe_production_sk_...", // Only shown once!
    "key_prefix": "a1b2c3d4",
    "expires_at": "2025-05-09T10:00:00Z",
    "rotation_due": "2025-05-09T10:00:00Z",
    "rate_limits": {
      "per_minute": 60,
      "per_hour": 1000,
      "per_day": 10000
    }
  },
  "message": "API key created successfully. Please save the key securely - it will not be shown again."
}
```

### Rotating an API Key

```bash
curl -X POST https://api.example.com/api/v1/keys/rotate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "key_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

Response:
```json
{
  "success": true,
  "data": {
    "old_key_id": "550e8400-e29b-41d4-a716-446655440000",
    "new_key_id": "660e9500-f39c-52e5-b827-557766551111",
    "new_key": "ffprobe_production_sk_...", // Only shown once!
    "key_prefix": "b2c3d4e5",
    "grace_period_ends": "2025-02-16T10:00:00Z"
  },
  "message": "API key rotated successfully. The old key will remain valid for 7 days."
}
```

---

## JWT Secret Rotation

### Features

- **Versioned Secrets**: Each secret has a unique version number
- **Zero-Downtime Rotation**: Old tokens remain valid during grace period
- **Automatic Cleanup**: Expired secrets are automatically removed
- **Admin-Only Access**: Only administrators can rotate JWT secrets

### Rotating JWT Secret

```bash
curl -X POST https://api.example.com/api/v1/admin/rotate-jwt \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN"
```

Response:
```json
{
  "success": true,
  "data": {
    "version": 2,
    "algorithm": "HS256",
    "rotated_at": "2025-02-09T10:00:00Z",
    "expires_at": "2025-05-09T10:00:00Z"
  },
  "message": "JWT secret rotated successfully. Existing tokens remain valid during grace period."
}
```

---

## Per-User/Tenant Rate Limiting

### Rate Limit Hierarchy

1. **API Key Limits** (highest priority)
2. **User Limits** (overrides tenant)
3. **Tenant Limits** (overrides defaults)
4. **Default Limits** (fallback)

### Default Rate Limits

| Level | Requests/Minute | Requests/Hour | Requests/Day |
|-------|----------------|---------------|--------------|
| Free | 20 | 100 | 500 |
| Standard | 60 | 1000 | 10000 |
| Premium | 120 | 5000 | 50000 |
| Enterprise | Custom | Custom | Custom |

### Setting Custom Rate Limits

#### For an API Key
```bash
curl -X POST https://api.example.com/api/v1/admin/rate-limits \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "key_id": "550e8400-e29b-41d4-a716-446655440000",
    "rate_limit_rpm": 120,
    "rate_limit_rph": 5000,
    "rate_limit_rpd": 50000
  }'
```

#### For a User
```bash
curl -X POST https://api.example.com/api/v1/admin/rate-limits \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_123",
    "rate_limit_rpm": 100,
    "rate_limit_rph": 2000,
    "rate_limit_rpd": 20000
  }'
```

#### For a Tenant
```bash
curl -X POST https://api.example.com/api/v1/admin/rate-limits \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "acme_corp",
    "rate_limit_rpm": 200,
    "rate_limit_rph": 10000,
    "rate_limit_rpd": 100000
  }'
```

### Rate Limit Headers

All API responses include rate limit information:

```
X-RateLimit-Limit-Minute: 60
X-RateLimit-Limit-Hour: 1000
X-RateLimit-Limit-Day: 10000
X-RateLimit-Remaining-Minute: 45
X-RateLimit-Remaining-Hour: 876
X-RateLimit-Remaining-Day: 9234
X-RateLimit-Reset-Minute: 1707475260
X-RateLimit-Reset-Hour: 1707478800
X-RateLimit-Reset-Day: 1707523200
```

### Rate Limit Exceeded Response

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Limit: 60 req/min",
  "retry_after": 1707475260,
  "limits": {
    "per_minute": 60,
    "per_hour": 1000,
    "per_day": 10000
  },
  "current": {
    "per_minute": 61,
    "per_hour": 234,
    "per_day": 1234
  }
}
```

---

## Database Schema

### Key Tables

1. **api_keys**: Stores API keys with rotation metadata
2. **jwt_secrets**: Versioned JWT signing secrets
3. **api_key_rotation_log**: Audit log for key lifecycle
4. **tenant_rate_limits**: Per-tenant rate configurations
5. **user_rate_limits**: Per-user rate configurations
6. **rate_limit_usage**: Usage tracking for analytics

### Migration

Run the migration to create the necessary tables:

```bash
migrate -path migrations -database "postgres://user:pass@localhost/db" up 7
```

---

## API Endpoints

### Key Management Endpoints

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| POST | `/api/v1/keys/create` | Create new API key | User |
| POST | `/api/v1/keys/rotate` | Rotate existing key | User |
| GET | `/api/v1/keys/list` | List user's keys | User |
| DELETE | `/api/v1/keys/{id}` | Revoke API key | User |

### Admin Endpoints

| Method | Endpoint | Description | Required Role |
|--------|----------|-------------|---------------|
| POST | `/api/v1/admin/rotate-jwt` | Rotate JWT secret | Admin |
| POST | `/api/v1/admin/rate-limits` | Set custom rate limits | Admin |
| GET | `/api/v1/admin/rotation-status` | Check rotation due | Admin |
| POST | `/api/v1/admin/cleanup` | Cleanup expired keys | Admin |

---

## Best Practices

### Security Best Practices

1. **Never Log Keys**: Never log full API keys, only prefixes
2. **Secure Storage**: Store keys in secure vaults (HashiCorp Vault, AWS Secrets Manager)
3. **Regular Rotation**: Rotate keys every 90 days minimum
4. **Monitor Usage**: Track unusual usage patterns
5. **Immediate Revocation**: Revoke compromised keys immediately

### Rotation Schedule

```yaml
# Recommended rotation schedule
api_keys:
  standard: 90 days
  high_security: 30 days
  grace_period: 7 days

jwt_secrets:
  rotation: 90 days
  grace_period: 24 hours

monitoring:
  check_interval: daily
  alert_before: 7 days
```

### Implementation Checklist

- [ ] Enable automatic rotation monitoring
- [ ] Set up alerts for keys nearing expiration
- [ ] Implement key backup procedures
- [ ] Document rotation procedures
- [ ] Train team on rotation process
- [ ] Set up audit log monitoring
- [ ] Configure rate limit monitoring
- [ ] Implement usage analytics

---

## Monitoring & Alerts

### Key Metrics to Monitor

1. **Keys nearing expiration** (< 7 days)
2. **Failed rotation attempts**
3. **Rate limit violations**
4. **Unusual usage patterns**
5. **Key creation/deletion frequency**

### Sample Prometheus Alerts

```yaml
groups:
  - name: secret_rotation
    rules:
      - alert: APIKeyNearExpiration
        expr: days_until_expiration < 7
        for: 1h
        annotations:
          summary: "API key {{ $labels.key_id }} expires in {{ $value }} days"
      
      - alert: HighRateLimitViolations
        expr: rate(rate_limit_violations[5m]) > 10
        for: 5m
        annotations:
          summary: "High rate limit violations for {{ $labels.tenant_id }}"
      
      - alert: JWTSecretRotationDue
        expr: jwt_secret_age_days > 85
        for: 1h
        annotations:
          summary: "JWT secret rotation due (age: {{ $value }} days)"
```

### Logging Examples

```json
// Successful key creation
{
  "level": "info",
  "timestamp": "2025-02-09T10:00:00Z",
  "event": "api_key_created",
  "user_id": "user_123",
  "tenant_id": "acme_corp",
  "key_prefix": "a1b2c3d4",
  "expires_at": "2025-05-09T10:00:00Z"
}

// Rate limit exceeded
{
  "level": "warn",
  "timestamp": "2025-02-09T10:00:00Z",
  "event": "rate_limit_exceeded",
  "user_id": "user_123",
  "tenant_id": "acme_corp",
  "limit_type": "per_minute",
  "limit": 60,
  "current": 61
}

// Key rotation
{
  "level": "info",
  "timestamp": "2025-02-09T10:00:00Z",
  "event": "api_key_rotated",
  "old_key_prefix": "a1b2c3d4",
  "new_key_prefix": "b2c3d4e5",
  "grace_period_ends": "2025-02-16T10:00:00Z"
}
```

---

## Troubleshooting

### Common Issues

#### 1. Rate Limit Not Applied
- Check Redis connectivity
- Verify cache key format
- Check rate limit hierarchy

#### 2. Key Rotation Fails
- Check database permissions
- Verify maximum active keys limit
- Check transaction isolation level

#### 3. JWT Validation Fails After Rotation
- Ensure grace period is active
- Check token version compatibility
- Verify secret caching

### Debug Commands

```bash
# Check rate limit for user
redis-cli HGETALL "user:user_123:limits"

# Check API key metadata
redis-cli HGETALL "apikey:a1b2c3d4:meta"

# View rotation log
psql -c "SELECT * FROM api_key_rotation_log ORDER BY performed_at DESC LIMIT 10;"

# Check keys due for rotation
psql -c "SELECT id, name, rotation_due FROM api_keys WHERE rotation_due < NOW() + INTERVAL '7 days';"
```

---

## Migration Path

### From Static Keys to Rotating Keys

1. **Phase 1**: Deploy rotation infrastructure
2. **Phase 2**: Issue new rotating keys to all users
3. **Phase 3**: Monitor usage of old keys
4. **Phase 4**: Deprecate old keys with notice
5. **Phase 5**: Revoke old keys

### Timeline Example

```
Week 1-2: Deploy and test rotation system
Week 3-4: Issue new keys to pilot users
Week 5-8: Gradual rollout to all users
Week 9-10: Monitor and support transition
Week 11-12: Deprecate old keys
Week 13: Revoke old keys
```

---

## Support

For issues or questions regarding secret rotation:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review logs in `/var/log/rendiff-probe/`
3. Contact security team for urgent issues
4. File a ticket for non-urgent requests

---

*Last Updated: 2025-02-09*  
*Version: 1.0.0*