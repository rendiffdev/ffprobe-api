---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: ['bug', 'needs-triage']
assignees: ''
---

## 🐛 Bug Description
A clear and concise description of what the bug is.

## 🔄 Steps to Reproduce
Steps to reproduce the behavior:

1. Go to '...'
2. Click on '...'
3. Scroll down to '...'
4. See error

## ✅ Expected Behavior
A clear and concise description of what you expected to happen.

## ❌ Actual Behavior
A clear and concise description of what actually happened.

## 📸 Screenshots
If applicable, add screenshots to help explain your problem.

## 🌍 Environment
**Docker Environment:**
- Docker Version: [e.g., 24.0.6]
- Docker Compose Version: [e.g., v2.20.0]
- Host OS: [e.g., Ubuntu 22.04, macOS 13.0, Windows 11]

**Local Development (if applicable):**
- OS: [e.g., Ubuntu 22.04]
- Go Version: [e.g., 1.23]
- FFmpeg Version: [e.g., 6.1]
- PostgreSQL Version: [e.g., 16]
- Redis Version: [e.g., 7]

**API Details:**
- API Version: [e.g., 2.0.0]
- Authentication Method: [e.g., API Key, JWT]
- Endpoint: [e.g., POST /api/v1/probe/file]

## 📋 Configuration
**Environment Variables (remove sensitive data):**
```bash
STORAGE_PROVIDER=s3
ENABLE_AUTH=true
# Add relevant config here
```

**Request Details (if API related):**
```bash
curl -X POST "http://localhost:8080/api/v1/probe/file" \
  -H "X-API-Key: ***" \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/path/to/video.mp4"
  }'
```

## 📝 Logs
**Error Logs:**
```
Paste any relevant log output here (remove sensitive information)
```

**Docker Logs (if applicable):**
```bash
# Command used to get logs
docker compose logs ffprobe-api

# Log output
[paste logs here]
```

## 🔍 Additional Context
Add any other context about the problem here.

**Media File Details (if relevant):**
- File format: [e.g., MP4, MOV]
- File size: [e.g., 1.2GB]
- Video codec: [e.g., H.264]
- Resolution: [e.g., 1920x1080]
- Duration: [e.g., 2 minutes 30 seconds]

**Error Frequency:**
- [ ] This happens every time
- [ ] This happens occasionally
- [ ] This happened only once

**Impact:**
- [ ] Blocks functionality completely
- [ ] Degrades performance
- [ ] Minor inconvenience

## 🔧 Attempted Solutions
What have you tried to resolve this issue?

- [ ] Restarted Docker containers
- [ ] Checked Docker logs
- [ ] Verified environment variables
- [ ] Tested with different files
- [ ] Checked API documentation
- [ ] Searched existing issues

## 🆘 Priority
- [ ] Low - Nice to have fix
- [ ] Medium - Affects some workflows
- [ ] High - Blocks important functionality
- [ ] Critical - Production system down

---

**Note:** Please ensure you've checked the [documentation](../docs/README.md) and [existing issues](https://github.com/rendiffdev/ffprobe-api/issues) before submitting.