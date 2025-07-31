# 📚 FFprobe API Documentation

Welcome to the comprehensive documentation for the FFprobe API. This directory contains everything you need to successfully deploy, configure, and use the API.

## 🗂️ Documentation Structure

### 📖 [API Reference](api/)
Complete API documentation with endpoints, request/response schemas, and examples.

| Document | Description |
|----------|-------------|
| [🔐 Authentication](api/authentication.md) | API key and JWT authentication |
| [⚡ Enhanced API](api/enhanced_api.md) | Advanced features and capabilities |
| [📋 OpenAPI Spec](api/openapi.yaml) | Machine-readable API specification |

### 🚀 [Deployment Guide](deployment/)
Production deployment instructions and configuration.

| Document | Description |
|----------|-------------|
| [🔍 Docker Audit](deployment/DOCKER_AUDIT_REPORT.md) | Container security and optimization |
| [✅ Production Checklist](deployment/PRODUCTION_READINESS_CHECKLIST.md) | Pre-deployment validation |
| [⚙️ Configuration](deployment/configuration.md) | Environment variables and settings |

### 🏗️ [Architecture](architecture/)
System design, components, and technical architecture.

| Document | Description |
|----------|-------------|
| [🏛️ System Overview](architecture/system-overview.md) | High-level architecture |
| [🔄 Data Flow](architecture/data-flow.md) | Request processing flow |
| [⚖️ Scaling](architecture/scaling.md) | Horizontal and vertical scaling |

### 🎓 [Tutorials](tutorials/)
Step-by-step guides and practical examples.

| Document | Description |
|----------|-------------|
| [📝 API Usage](tutorials/api_usage.md) | Getting started with the API |
| [🎬 Video Analysis](tutorials/video-analysis.md) | Video processing examples |
| [📊 Quality Assessment](tutorials/quality-assessment.md) | VMAF and quality metrics |

## 🚀 Quick Navigation

### For Developers
- [🔧 Development Setup](../README.md#-development)
- [📋 API Reference](api/)
- [🧪 Testing Guide](../tests/)
- [📜 Scripts Documentation](../scripts/README.md)

### For DevOps/SRE
- [🐳 Container Deployment](deployment/)
- [📊 Monitoring Setup](../docker/prometheus.yml)
- [🔒 Security Configuration](deployment/PRODUCTION_READINESS_CHECKLIST.md)
- [📁 Repository Structure](../REPOSITORY_STRUCTURE.md)
- [🛡️ Security Audit](../SECURITY_AUDIT_REPORT.md)

### For Installation & Setup
- [🎯 Interactive Installer](../scripts/setup/install.sh)
- [⚡ Quick Setup](../scripts/setup/quick-setup.sh)
- [✅ Configuration Validator](../scripts/setup/validate-config.sh)
- [🚀 Production Deployment](../scripts/deployment/deploy.sh)

### For Integration
- [🎯 API Examples](../README.md#-api-examples)
- [🔐 Authentication](api/authentication.md)
- [⚡ Advanced Features](api/enhanced_api.md)

## 🆘 Getting Help

### Documentation Issues
- **Missing Information**: Open an issue with the `documentation` label
- **Outdated Content**: Submit a PR with corrections
- **New Examples**: Contribute tutorials in the `tutorials/` directory

### Technical Support
- **Bug Reports**: [GitHub Issues](https://github.com/rendiffdev/ffprobe-api/issues)
- **Feature Requests**: [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
- **Contact**: [dev@rendiff.dev](mailto:dev@rendiff.dev)
- **Community Chat**: [Discord/Slack Channel]

## 📝 Contributing to Documentation

We welcome documentation contributions! Please follow these guidelines:

### Documentation Standards
- **Clear Structure**: Use headings, lists, and tables
- **Code Examples**: Include working code snippets
- **Screenshots**: Add visual aids when helpful
- **Links**: Cross-reference related sections

### File Organization
```
docs/
├── api/              # API reference documentation
├── architecture/     # System design and architecture
├── deployment/       # Deployment and operations
└── tutorials/        # Step-by-step guides

scripts/
├── setup/            # Installation and configuration
├── deployment/       # Production deployment tools
└── maintenance/      # Backup and maintenance tools
```

### Writing Style
- **Audience**: Write for your intended audience (developers, operators, etc.)
- **Clarity**: Use simple, clear language
- **Examples**: Include practical examples
- **Updates**: Keep content current with code changes

## 🔄 Documentation Updates

This documentation is automatically updated with each release. For the latest information:

- **Development**: Check the `main` branch documentation
- **Stable**: Use the documentation from the latest release tag
- **API Changes**: Review the [CHANGELOG.md](../CHANGELOG.md)

---

**📖 Happy Reading!** 

Need help? Check our [support channels](#-getting-help) or contribute to make this documentation better.