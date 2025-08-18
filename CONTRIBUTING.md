# Contributing to FFprobe API

Thank you for your interest in contributing to FFprobe API! This document provides guidelines and information for contributors.

> **📋 For detailed technical guidelines, see [CONTRIBUTOR-GUIDELINES.md](CONTRIBUTOR-GUIDELINES.md)**  
> This document covers general contribution workflow and community guidelines.

## 🚀 Quick Start

1. **Fork** the repository
2. **Clone** your fork: `git clone https://github.com/YOUR-USERNAME/ffprobe-api.git`
3. **Add upstream**: `git remote add upstream https://github.com/rendiffdev/ffprobe-api.git`
3. **Read** [Technical Guidelines](CONTRIBUTOR-GUIDELINES.md) for detailed setup instructions
4. **Install** dependencies: `docker compose up -d` 
5. **Run** tests: `go test ./...`
6. **Start** development: `go run cmd/ffprobe-api/main.go`

## 🛠️ Development Setup

### Prerequisites

- Go 1.23+
- Docker & Docker Compose v2.20+
- Git

### Quick Environment Setup

```bash
# Clone and setup
git clone https://github.com/rendiffdev/ffprobe-api.git
cd ffprobe-api

# Start development environment (includes PostgreSQL, Redis, FFmpeg)
docker compose -f compose.yaml -f compose.development.yaml up -d

# Verify setup
curl http://localhost:8080/health

# Run tests
go test ./...
```

For detailed setup instructions, see [Technical Guidelines](CONTRIBUTOR-GUIDELINES.md#development-environment).

## 🔄 Development Workflow

### 1. Find Something to Work On

- Check [Good First Issues](https://github.com/rendiffdev/ffprobe-api/labels/good%20first%20issue) for beginners
- Browse [Help Wanted](https://github.com/rendiffdev/ffprobe-api/labels/help%20wanted) issues
- Review our [Project Roadmap](https://github.com/rendiffdev/ffprobe-api/projects)
- Propose new features via [GitHub Discussions](https://github.com/rendiffdev/ffprobe-api/discussions)

### 2. Create a Feature Branch

```bash
git checkout -b feature/amazing-feature
# or
git checkout -b fix/bug-description
# or
git checkout -b docs/documentation-update
```

### 3. Make Changes

- Follow our [Code Standards](CONTRIBUTOR-GUIDELINES.md#code-standards)
- Add tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 4. Commit Changes

We use [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Features
git commit -m "feat: add video quality comparison endpoint"
git commit -m "feat(storage): add Azure Blob Storage support"

# Bug fixes
git commit -m "fix: resolve memory leak in large file processing"
git commit -m "fix(auth): handle expired JWT tokens correctly"

# Documentation
git commit -m "docs: update API documentation for HLS endpoints"
git commit -m "docs(readme): add Docker deployment examples"

# Tests
git commit -m "test: add integration tests for storage providers"
git commit -m "test(quality): add VMAF accuracy validation tests"

# Refactor
git commit -m "refactor: optimize database query performance"
git commit -m "refactor(handlers): simplify error handling logic"
```

### 5. Push and Create PR

```bash
git push origin feature/amazing-feature
```

Then create a Pull Request using our [PR template](.github/pull_request_template.md).

## 📝 Code Standards

### Overview

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` and `goimports` for formatting
- Run `golangci-lint` for static analysis
- Maintain 85%+ test coverage
- Follow our [detailed code standards](CONTRIBUTOR-GUIDELINES.md#code-standards)

### Quick Style Check

```bash
# Format code
go fmt ./...

# Check for issues
golangci-lint run

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🧪 Testing Guidelines

### Test Types

1. **Unit Tests**: Test individual functions/methods
2. **Integration Tests**: Test component interactions  
3. **End-to-End Tests**: Test complete workflows
4. **Security Tests**: Test input validation and security

### Running Tests

```bash
# All tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...

# Specific package
go test ./internal/services

# Integration tests
go test -tags=integration ./tests
```

For detailed testing patterns and examples, see [Testing Guidelines](CONTRIBUTOR-GUIDELINES.md#testing-guidelines).

## 🔒 Security Guidelines

### Security-First Development

- **Input Validation**: Validate all user inputs
- **SQL Injection Prevention**: Use parameterized queries only
- **File Upload Security**: Validate file types, sizes, and content
- **Authentication**: Implement proper auth checks
- **Secrets Management**: Never commit secrets to git

### Security Checklist

- [ ] Input validation for all endpoints
- [ ] Proper error handling without information leakage
- [ ] Authentication/authorization checks
- [ ] Rate limiting implementation
- [ ] SQL injection prevention
- [ ] XSS protection
- [ ] CSRF protection
- [ ] Security headers

For detailed security guidelines, see [Security Guidelines](CONTRIBUTOR-GUIDELINES.md#security-guidelines).

## 🐛 Bug Reports

### Before Reporting

1. Check [existing issues](https://github.com/rendiffdev/ffprobe-api/issues)
2. Ensure you're using the latest version
3. Test in a clean environment (Docker)

### Bug Report Template

Use our [bug report template](.github/ISSUE_TEMPLATE/bug_report.md) or include:

```markdown
**Bug Description**
A clear description of what the bug is.

**Steps to Reproduce**
1. Go to '...'
2. Click on '...'
3. See error

**Expected Behavior**
What you expected to happen.

**Environment**
- OS: [e.g., Ubuntu 22.04]
- Go version: [e.g., 1.23]
- Docker version: [e.g., 24.0.6]
- API version: [e.g., 2.0.0]

**Logs**
Include relevant log output (remove sensitive information).
```

## 💡 Feature Requests

### Before Requesting

1. Check [existing discussions](https://github.com/rendiffdev/ffprobe-api/discussions)
2. Review our [roadmap](https://github.com/rendiffdev/ffprobe-api/projects)
3. Consider if this fits the project scope

### Feature Request Template

Use our [feature request template](.github/ISSUE_TEMPLATE/feature_request.md) or include:

```markdown
**Feature Description**
A clear description of the feature you'd like to see.

**Problem Statement**
What problem does this solve? Who would benefit?

**Proposed Solution**
How you envision this feature working.

**Alternatives Considered**
Other approaches you've considered.

**Implementation Notes**
Any technical considerations or constraints.
```

## 🚀 Pull Request Guidelines

### Before Submitting

1. Read our [Technical Guidelines](CONTRIBUTOR-GUIDELINES.md)
2. Ensure your branch is up-to-date with main
3. Run all tests locally
4. Self-review your changes

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Tests added for new functionality
- [ ] All tests pass locally
- [ ] Documentation updated
- [ ] Security considerations addressed
- [ ] Breaking changes documented

### PR Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and security scans
2. **Code Review**: Maintainers review code quality and design
3. **Testing**: Changes are tested in staging environment
4. **Approval**: At least one maintainer approval required
5. **Merge**: Squash and merge to main branch

## 📦 Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v3.0.0): Breaking changes
- **MINOR** (v2.1.0): New features (backward compatible)
- **PATCH** (v2.0.1): Bug fixes (backward compatible)

### Release Types

- **🚀 Major Release**: New major version with breaking changes
- **✨ Minor Release**: New features and enhancements
- **🐛 Patch Release**: Bug fixes and security patches
- **🔥 Hotfix**: Critical security or bug fixes

## 🤝 Community Guidelines

### Code of Conduct

We follow the [Contributor Covenant](https://www.contributor-covenant.org/). In summary:

- **Be respectful and inclusive**
- **Welcome newcomers and help them learn**
- **Focus on constructive feedback**
- **Respect different viewpoints and experiences**
- **No harassment, discrimination, or inappropriate behavior**

### Communication Channels

- **🐛 GitHub Issues**: Bug reports and feature requests
- **💬 GitHub Discussions**: General questions and community chat
- **🔄 Pull Requests**: Code contributions and reviews
- **📧 Security**: dev@rendiff.dev (security issues only)

### Response Times

- **Issues**: We aim to respond within 48 hours
- **Pull Requests**: Initial review within 72 hours
- **Security Issues**: Response within 24 hours

## 🎯 Areas for Contribution

### 🔥 High Priority

- **Performance optimizations** for large file processing
- **Additional cloud storage providers** (Backblaze, DigitalOcean)
- **Enhanced security features** (OAuth2, SAML)
- **Client libraries** (Python, JavaScript, Java)
- **Advanced monitoring** and alerting

### ⭐ Medium Priority

- **Web dashboard** for analysis management
- **Advanced analytics** and reporting features
- **Webhook integrations** for external systems
- **Advanced quality metrics** (custom algorithms)
- **Multi-language support** for error messages

### 🌱 Beginner Friendly

- **Documentation improvements** and examples
- **Test coverage** enhancements
- **Code cleanup** and refactoring
- **Bug fixes** and small improvements
- **Docker image optimizations**

### 🆕 Good First Issues

Look for issues labeled with:
- [`good first issue`](https://github.com/rendiffdev/ffprobe-api/labels/good%20first%20issue)
- [`help wanted`](https://github.com/rendiffdev/ffprobe-api/labels/help%20wanted)
- [`documentation`](https://github.com/rendiffdev/ffprobe-api/labels/documentation)

## 📚 Resources

### Documentation

- **📋 [Technical Contributor Guidelines](CONTRIBUTOR-GUIDELINES.md)** - Detailed technical guide
- **📖 [API Documentation](./docs/README.md)** - Complete API reference
- **🐳 [Docker Documentation](./docs/docker.md)** - Container deployment guide
- **🔒 [Security Documentation](./docs/security.md)** - Security best practices

### Examples and Tutorials

- **🚀 [Quick Start Examples](./examples/)** - Get started quickly
- **⚙️ [Configuration Examples](./examples/config/)** - Common configurations
- **🔧 [Integration Examples](./examples/integrations/)** - Third-party integrations

### Development Tools

- **🛠️ [VS Code Extensions](./docs/vscode.md)** - Recommended extensions
- **🐛 [Debugging Guide](./docs/debugging.md)** - Debugging techniques
- **📊 [Performance Profiling](./docs/profiling.md)** - Performance optimization

## ❓ Getting Help

### Self-Help Resources

1. **📖 Check Documentation**: Start with our comprehensive docs
2. **🔍 Search Issues**: Someone might have had the same problem
3. **💬 Browse Discussions**: Check community Q&A
4. **🧪 Try Examples**: Run our example code

### Community Support

1. **💬 GitHub Discussions**: Ask questions and share ideas
2. **🐛 GitHub Issues**: Report bugs and request features  
3. **📧 Email**: dev@rendiff.dev (for security issues)

### Getting Faster Help

- **Be specific**: Include error messages, logs, and environment details
- **Provide context**: What were you trying to achieve?
- **Share code**: Minimal reproduction examples help a lot
- **Use formatting**: Use markdown code blocks for logs and code

## 🎉 Recognition

We appreciate all contributions! Contributors are recognized through:

- **📝 Changelog mentions** for significant contributions
- **⭐ GitHub profile** highlighting in our README
- **🏆 Annual contributor awards** for outstanding contributions
- **📢 Social media shoutouts** for major features

## 📝 License

By contributing to FFprobe API, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

<div align="center">

**🎬 Thank you for contributing to FFprobe API!**

**⭐ Star the project • 🔄 Share with others • 🤝 Join our community**

</div>