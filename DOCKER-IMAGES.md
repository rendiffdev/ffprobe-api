# Docker Images

The `docker-image/` directory has been moved to `.gitignore` to prevent large build files from being committed to the repository.

## Available Docker Images on Docker Hub

All production-ready Docker images are available on Docker Hub at:
**`rendiffdev/ffprobe-api`**

### Railway Deployment Images:
- `rendiffdev/ffprobe-api:railway-v2` ⭐ **RECOMMENDED**
- `rendiffdev/ffprobe-api:railway-fixed`
- `rendiffdev/ffprobe-api:latest-railway`

### Platform-Specific Images:
- `rendiffdev/ffprobe-api:amd64` - AMD64 architecture
- `rendiffdev/ffprobe-api:arm64` - ARM64 architecture  
- `rendiffdev/ffprobe-api:latest` - Latest multi-arch build

## Features in All Images:
- ✅ 19 QC Analysis Categories
- ✅ Local LLM Integration (Mandatory)
- ✅ Native FFprobe Capabilities
- ✅ Production-Ready Configuration
- ✅ Cloud Deployment Compatible

## Docker Build Files:
The Docker build files and configurations are maintained separately and deployed directly to Docker Hub. Contact the maintainers if you need access to the build configurations.

## Usage:
```bash
# Pull and run the latest Railway-compatible image
docker pull rendiffdev/ffprobe-api:railway-v2
docker run -p 8080:8080 rendiffdev/ffprobe-api:railway-v2
```