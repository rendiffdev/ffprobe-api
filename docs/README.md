# FFprobe API Documentation

> **Complete documentation for FFprobe API - Enterprise Video Analysis Platform**

## 📚 Documentation Index

### Getting Started
- **[Quick Start Verification](QUICK_START_VERIFICATION.md)** - Step-by-step setup verification
- **[Deployment Guide](deployment/README.md)** - Complete deployment instructions
- **[Configuration](deployment/configuration.md)** - Environment variables and settings

### API Reference
- **[API Overview](api/README.md)** - Complete REST API documentation
- **[Authentication](api/authentication.md)** - API keys and JWT tokens
- **[Complete API Guide](api/complete-api-guide.md)** - Comprehensive endpoint documentation
- **[Enhanced API Features](api/enhanced_api.md)** - Advanced API capabilities

### Quality Control
- **[Quality Checks Reference](QUALITY_CHECKS.md)** - All 49 quality control parameters
- **[Comparison System](COMPARISON_SYSTEM.md)** - Video comparison and quality validation
- **[API Authentication](API_AUTHENTICATION.md)** - Security and authentication details

### Deployment
- **[Deployment Overview](deployment/README.md)** - Deployment strategies and options
- **[Configuration Guide](deployment/configuration.md)** - Complete configuration reference
- **[Storage Configuration](deployment/storage-configuration.md)** - Storage backend setup
- **[Production Checklist](deployment/production-checklist.md)** - Pre-deployment validation
- **[Production Readiness](deployment/PRODUCTION_READINESS_CHECKLIST.md)** - Comprehensive production guide

### Development
- **[Architecture Overview](development/architecture.md)** - System design and components
- **[Contributing Guidelines](../CONTRIBUTING.md)** - How to contribute
- **[Repository Structure](../REPOSITORY_STRUCTURE.md)** - Project organization

### Operations
- **[Monitoring Guide](operations/monitoring.md)** - Prometheus and logging setup
- **[Troubleshooting Guide](operations/troubleshooting.md)** - Common issues and solutions
- **[Legacy Troubleshooting](TROUBLESHOOTING.md)** - Additional troubleshooting reference

### Tutorials
- **[API Usage Tutorial](tutorials/api_usage.md)** - Getting started with the API
- **[Local LLM Setup](tutorials/local-llm-setup.md)** - AI report generation with Ollama

## 🚀 Quick Navigation

### By Use Case

#### I want to...
- **[Analyze a video file →](api/README.md#post-apiv1probefile)**
- **[Enable all 49 quality checks →](QUALITY_CHECKS.md)**
- **[Deploy to production →](deployment/README.md)**
- **[Set up monitoring →](operations/monitoring.md)**
- **[Configure authentication →](api/authentication.md)**
- **[Troubleshoot issues →](operations/troubleshooting.md)**

### By Role

#### Video Engineer
- [Quality Checks Reference](QUALITY_CHECKS.md)
- [Comparison System](COMPARISON_SYSTEM.md)
- [Enhanced API Features](api/enhanced_api.md)

#### DevOps Engineer
- [Deployment Guide](deployment/README.md)
- [Monitoring Setup](operations/monitoring.md)
- [Production Checklist](deployment/production-checklist.md)

#### Developer
- [API Documentation](api/README.md)
- [Complete API Guide](api/complete-api-guide.md)
- [Architecture Overview](development/architecture.md)

## 📖 Documentation Standards

### Document Structure
Each documentation file follows this structure:
1. **Title & Description** - Clear purpose statement
2. **Prerequisites** - Required knowledge or setup
3. **Content** - Main documentation with examples
4. **Troubleshooting** - Common issues for that topic
5. **Next Steps** - Related documentation links

### Code Examples
All code examples are:
- **Tested** - Verified to work with current version
- **Complete** - Runnable without modification
- **Annotated** - Include helpful comments
- **Practical** - Based on real use cases

### Version Information
- **API Version**: v1
- **FFmpeg Version**: 6.1.1
- **Docker Base**: Alpine 3.19
- **Go Version**: 1.23
- **Documentation Updated**: August 2024

## 🔍 Search Documentation

### By Feature
- [Quality Checks (49 parameters)](QUALITY_CHECKS.md)
- [Video Comparison](COMPARISON_SYSTEM.md)
- [Authentication System](API_AUTHENTICATION.md)
- [Enhanced API Features](api/enhanced_api.md)

### By Technology
- [FFprobe/FFmpeg 6.1](QUALITY_CHECKS.md)
- [PostgreSQL Configuration](deployment/configuration.md#database-configuration)
- [Redis Caching](deployment/configuration.md#redis-configuration)
- [Docker Deployment](deployment/README.md#docker-compose-recommended)
- [Prometheus Monitoring](operations/monitoring.md#prometheus-metrics)

## 📞 Support

### Getting Help
- **GitHub Issues**: [Report bugs or request features](https://github.com/rendiffdev/ffprobe-api/issues)
- **Discussions**: [Community discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
- **Email**: [support@rendiff.dev](mailto:support@rendiff.dev)

### Contributing
- [Contribution Guidelines](../CONTRIBUTING.md)
- [Repository Structure](../REPOSITORY_STRUCTURE.md)

---

*Documentation is continuously updated. For the latest changes, see [GitHub repository](https://github.com/rendiffdev/ffprobe-api).*