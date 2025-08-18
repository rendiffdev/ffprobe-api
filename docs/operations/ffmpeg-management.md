# FFmpeg Version Management

## Overview

The FFprobe API uses **BtbN's FFmpeg Builds** for consistent, high-quality FFmpeg binaries with all codecs and features enabled. The system includes automatic update checking, stability verification, and safe rollback capabilities.

## Key Features

- **🚀 Latest Stable Builds**: Always uses BtbN's production-ready FFmpeg builds
- **🔍 Automatic Update Checking**: Monitors for new releases daily
- **✅ Stability Verification**: 48-hour stability period before recommending updates
- **⚠️ Major Version Protection**: Requires explicit approval for breaking changes
- **↩️ Safe Rollback**: Automatic backups allow instant rollback if issues occur
- **📊 Version Tracking**: Complete version history and change tracking

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  FFmpeg Manager                  │
├─────────────────────────────────────────────────┤
│                                                   │
│  ┌──────────────┐    ┌──────────────────────┐   │
│  │ Version      │───▶│ BtbN GitHub Releases │   │
│  │ Checker      │    └──────────────────────┘   │
│  └──────────────┘                                │
│         │                                        │
│         ▼                                        │
│  ┌──────────────┐    ┌──────────────────────┐   │
│  │ Stability    │───▶│ • 48hr aging check   │   │
│  │ Verifier     │    │ • Known issues DB    │   │
│  │              │    │ • Binary testing     │   │
│  └──────────────┘    └──────────────────────┘   │
│         │                                        │
│         ▼                                        │
│  ┌──────────────┐    ┌──────────────────────┐   │
│  │ Update       │───▶│ • Download binary    │   │
│  │ Installer    │    │ • Backup current     │   │
│  │              │    │ • Atomic swap       │   │
│  └──────────────┘    └──────────────────────┘   │
│                                                   │
└─────────────────────────────────────────────────┘
```

## API Endpoints

### Check Current Version
```bash
GET /api/v1/admin/ffmpeg/version

Response:
{
  "version": {
    "major": 6,
    "minor": 1,
    "patch": 1,
    "git": "abc123",
    "stable": true,
    "release_date": "2024-01-15T10:00:00Z"
  },
  "status": "current"
}
```

### Check for Updates
```bash
GET /api/v1/admin/ffmpeg/check-updates

Response:
{
  "current": {
    "major": 6,
    "minor": 0,
    "patch": 0
  },
  "available": {
    "major": 6,
    "minor": 1,
    "patch": 1,
    "release_url": "https://github.com/BtbN/FFmpeg-Builds/releases/tag/latest",
    "stable": true
  },
  "update_available": true,
  "is_major_upgrade": false,
  "is_minor_upgrade": true,
  "stability": {
    "stable": true,
    "compatibility": true,
    "warnings": [],
    "checked_at": "2024-01-15T12:00:00Z"
  },
  "recommendation": "Minor upgrade available. Review changelog for new features.",
  "user_approval_required": false
}
```

### Perform Update
```bash
POST /api/v1/admin/ffmpeg/update
Content-Type: application/json

{
  "confirm": true,
  "version": "latest"  // optional, defaults to latest stable
}

Response (Server-Sent Events):
data: {"event": "progress", "percent": 25}
data: {"event": "progress", "percent": 50}
data: {"event": "progress", "percent": 100}
data: {"event": "complete", "data": "FFmpeg updated successfully"}
```

### Rollback Update
```bash
POST /api/v1/admin/ffmpeg/rollback

Response:
{
  "status": "success",
  "message": "Rolled back to version 6.0.0",
  "previous_version": "6.1.1",
  "current_version": "6.0.0"
}
```

## Configuration

### Environment Variables

```bash
# FFmpeg binary locations
FFMPEG_PATH=/usr/local/bin/ffmpeg
FFPROBE_PATH=/usr/local/bin/ffprobe

# Update configuration
FFMPEG_AUTO_UPDATE=false              # Auto-update minor/patch versions
FFMPEG_UPDATE_CHECK_INTERVAL=86400    # Check interval in seconds (24h)
FFMPEG_ALLOW_MAJOR_UPDATES=false      # Allow major version auto-updates

# Stability settings
FFMPEG_STABILITY_PERIOD=172800        # Wait period in seconds (48h)
FFMPEG_TEST_ON_UPDATE=true            # Test binary before installation

# Backup settings
FFMPEG_BACKUP_COUNT=3                 # Number of backups to keep
FFMPEG_BACKUP_DIR=/app/backup/ffmpeg  # Backup directory
```

## Update Strategy

### Version Types

| Update Type | Example | Auto-Update | User Approval | Risk Level |
|------------|---------|-------------|---------------|------------|
| **Patch** | 6.0.0 → 6.0.1 | Yes* | No | Low |
| **Minor** | 6.0.1 → 6.1.0 | Yes* | No | Medium |
| **Major** | 6.1.0 → 7.0.0 | No | Required | High |

*When `FFMPEG_AUTO_UPDATE=true` and stability checks pass

### Stability Checks

1. **Age Check**: Release must be at least 48 hours old
2. **Known Issues**: Check against known problematic versions
3. **Binary Test**: Verify basic functionality before installation
4. **Codec Verification**: Ensure essential codecs are present
5. **API Compatibility**: Check for breaking API changes

### Known Issues Database

The system maintains a list of problematic versions:

```json
{
  "problematic_versions": {
    "6.1.0": "Memory leak in specific filters",
    "5.2.0": "Codec compatibility issues",
    "7.0.0-rc1": "Release candidate, not stable"
  }
}
```

## Manual Management

### Using the Update Script

```bash
# Check for updates
./scripts/ffmpeg-update.sh check

# Update to latest (skip major versions)
./scripts/ffmpeg-update.sh update

# Force update including major versions
./scripts/ffmpeg-update.sh update --allow-major

# Rollback to previous version
./scripts/ffmpeg-update.sh rollback

# Show current version
./scripts/ffmpeg-update.sh version

# Test current installation
./scripts/ffmpeg-update.sh test
```

### Docker Build Options

```bash
# Use latest BtbN build (recommended)
docker build -f Dockerfile.btbn -t ffprobe-api:latest .

# Use specific FFmpeg version
docker build \
  --build-arg FFMPEG_VERSION=6.1.1 \
  -f Dockerfile.btbn \
  -t ffprobe-api:6.1.1 .

# Use nightly builds (not recommended for production)
docker build \
  --build-arg FFMPEG_BUILD=nightly \
  -f Dockerfile.btbn \
  -t ffprobe-api:nightly .
```

## Best Practices

### Production Deployment

1. **Disable Auto-Updates in Production**
   ```bash
   FFMPEG_AUTO_UPDATE=false
   ```

2. **Test Updates in Staging First**
   - Deploy update to staging environment
   - Run comprehensive test suite
   - Monitor for 24-48 hours
   - Deploy to production if stable

3. **Schedule Maintenance Windows**
   - Plan major updates during low-traffic periods
   - Notify users in advance
   - Have rollback plan ready

4. **Monitor After Updates**
   - Check error rates
   - Monitor performance metrics
   - Verify codec functionality
   - Test critical workflows

### Update Decision Matrix

| Scenario | Recommendation | Action |
|----------|---------------|--------|
| Security patch available | Update immediately | Auto-update or manual |
| Minor version with new features | Test in staging first | Manual update after testing |
| Major version available | Extensive testing required | Manual update with approval |
| Current version has known issues | Update to fixed version | Priority update |
| Stable for >30 days | No action needed | Continue monitoring |

## Troubleshooting

### Common Issues

#### Update Fails to Download
```bash
# Check network connectivity
curl -I https://github.com/BtbN/FFmpeg-Builds/releases/latest

# Verify GitHub API access
curl https://api.github.com/rate_limit

# Manual download
wget https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz
```

#### Binary Test Fails
```bash
# Check binary directly
/usr/local/bin/ffmpeg -version

# Verify library dependencies
ldd /usr/local/bin/ffmpeg

# Check available codecs
/usr/local/bin/ffmpeg -codecs
```

#### Rollback Required
```bash
# Automatic rollback
./scripts/ffmpeg-update.sh rollback

# Manual rollback
cp /app/backup/ffmpeg/ffmpeg.backup_* /usr/local/bin/ffmpeg
cp /app/backup/ffmpeg/ffprobe.backup_* /usr/local/bin/ffprobe
chmod +x /usr/local/bin/ffmpeg /usr/local/bin/ffprobe
```

### Health Checks

```bash
# Verify FFmpeg installation
curl http://localhost:8080/api/v1/admin/ffmpeg/version

# Check codec support
ffmpeg -codecs 2>/dev/null | grep -E "h264|hevc|aac|opus"

# Test basic operation
ffmpeg -f lavfi -i testsrc=duration=1:size=320x240:rate=30 -f null -
```

## Security Considerations

1. **Verify Downloads**: All downloads are verified against GitHub's SSL certificates
2. **Binary Testing**: New binaries are tested in isolation before installation
3. **Atomic Updates**: Updates use atomic file operations to prevent corruption
4. **Backup Retention**: Multiple backups ensure recovery options
5. **Access Control**: Update endpoints require admin authentication

## Performance Impact

- **Update Check**: <100ms, cached for 24 hours
- **Download Time**: Varies by connection (typically 1-5 minutes for ~100MB)
- **Installation**: <10 seconds with atomic swap
- **Rollback**: <5 seconds from local backup

## Monitoring

### Metrics to Track

```prometheus
# FFmpeg version info
ffmpeg_version_info{version="6.1.1", major="6", minor="1", patch="1"}

# Update availability
ffmpeg_update_available{type="minor"} 1

# Last update check
ffmpeg_last_update_check_timestamp 1704931200

# Update operations
ffmpeg_updates_total{status="success"} 5
ffmpeg_updates_total{status="failed"} 1
ffmpeg_rollbacks_total 2

# Binary health
ffmpeg_binary_healthy 1
```

### Alerting Rules

```yaml
- alert: FFmpegMajorUpdateAvailable
  expr: ffmpeg_update_available{type="major"} == 1
  for: 1h
  annotations:
    summary: "Major FFmpeg update available"
    description: "Review changelog and plan update"

- alert: FFmpegBinaryUnhealthy
  expr: ffmpeg_binary_healthy == 0
  for: 5m
  annotations:
    summary: "FFmpeg binary health check failed"
    description: "Immediate attention required"
```

## Conclusion

The FFmpeg version management system ensures you always have access to the latest stable FFmpeg builds while maintaining safety and control over updates. The combination of automatic checking, stability verification, and user-controlled major updates provides the perfect balance between staying current and maintaining stability.