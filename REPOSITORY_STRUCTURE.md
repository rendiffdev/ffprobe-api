# 📁 Repository Structure

This document provides a comprehensive overview of the FFprobe API repository organization.

## 🏗️ Directory Structure

```
ffprobe-api/
├── 📄 Core Files
│   ├── README.md                           # Main project documentation
│   ├── LICENSE                             # MIT license
│   ├── CHANGELOG.md                        # Version history
│   ├── CONTRIBUTING.md                     # Contributor guidelines
│   ├── SECURITY_AUDIT_REPORT.md            # Security audit results
│   ├── REPOSITORY_STRUCTURE.md             # This file
│   └── .env.example                        # Configuration template
│
├── 🛠️ Build & Configuration
│   ├── Makefile                            # Build automation
│   ├── go.mod                              # Go module definition
│   ├── go.sum                              # Go module checksums
│   ├── .gitignore                          # Git ignore rules
│   └── .dockerignore                       # Docker ignore rules
│
├── 🐳 Docker & Deployment
│   ├── Dockerfile                          # Multi-stage Docker build
│   ├── docker-entrypoint.sh               # Container entry point
│   ├── compose.yml                         # Base Docker Compose
│   ├── compose.dev.yml                     # Development overrides
│   ├── compose.prod.yml                    # Production overrides
│   └── docker/                             # Docker configurations
│       ├── nginx.conf                      # Nginx reverse proxy
│       ├── prometheus.yml                  # Prometheus config
│       └── init.sql                        # Database initialization
│
├── 📜 Scripts & Tools
│   ├── scripts/
│   │   ├── README.md                       # Scripts documentation
│   │   ├── setup/                          # Installation scripts
│   │   │   ├── install.sh                  # Interactive installer
│   │   │   ├── quick-setup.sh              # Quick 3-mode setup
│   │   │   └── validate-config.sh          # Configuration validator
│   │   ├── deployment/                     # Deployment scripts
│   │   │   ├── deploy.sh                   # Production deployer
│   │   │   └── healthcheck.sh              # Health validation
│   │   └── maintenance/                    # Maintenance scripts
│   │       └── backup.sh                   # Backup automation
│
├── 🎬 Application Code
│   ├── cmd/                                # Application entry points
│   │   └── ffprobe-api/
│   │       └── main.go                     # Main application
│   ├── internal/                           # Private application code
│   │   ├── api/                            # API routes
│   │   ├── batch/                          # Batch processing
│   │   ├── config/                         # Configuration
│   │   ├── database/                       # Database layer
│   │   ├── ffmpeg/                         # FFmpeg integration
│   │   ├── handlers/                       # HTTP handlers
│   │   ├── hls/                            # HLS stream processing
│   │   ├── llm/                            # AI/ML integration
│   │   ├── middleware/                     # HTTP middleware
│   │   ├── models/                         # Data models
│   │   ├── quality/                        # Quality analysis
│   │   ├── reports/                        # Report generation
│   │   ├── services/                       # Business logic
│   │   ├── storage/                        # Storage providers
│   │   ├── validator/                      # Input validation
│   │   └── workflows/                      # E2E workflows
│   └── pkg/                                # Public libraries
│       └── logger/                         # Logging utilities
│
├── 🗄️ Database
│   └── migrations/                         # Database migrations
│       ├── 001_initial_schema.up.sql       # Initial schema
│       ├── 001_initial_schema.down.sql     # Schema rollback
│       ├── 005_create_quality_metrics_tables.up.sql
│       └── 005_create_quality_metrics_tables.down.sql
│
├── 🧪 Testing
│   └── tests/                              # Test files
│       ├── handlers_test.go                # Handler tests
│       ├── integration_test.go             # Integration tests
│       ├── services_test.go                # Service tests
│       └── storage_test.go                 # Storage tests
│
├── 📚 Documentation
│   └── docs/                               # Project documentation
│       ├── README.md                       # Documentation index
│       ├── api/                            # API documentation
│       │   ├── authentication.md           # Auth & security
│       │   ├── enhanced_api.md             # Advanced features
│       │   └── openapi.yaml                # API specification
│       ├── architecture/                   # System architecture
│       ├── deployment/                     # Deployment guides
│       │   ├── configuration.md            # Configuration guide
│       │   ├── DOCKER_AUDIT_REPORT.md      # Container security
│       │   └── PRODUCTION_READINESS_CHECKLIST.md
│       └── tutorials/                      # Step-by-step guides
│           └── api_usage.md                # API usage examples
│
└── 🤖 CI/CD
    └── .github/                            # GitHub workflows
        ├── workflows/
        │   └── ci.yml                      # CI/CD pipeline
        ├── ISSUE_TEMPLATE/                 # Issue templates
        │   ├── bug_report.md               # Bug report template
        │   └── feature_request.md          # Feature request template
        └── pull_request_template.md        # PR template
```

## 📋 File Categories

### 🔧 Configuration Files

| File | Purpose | Environment |
|------|---------|-------------|
| `.env.example` | Configuration template | All |
| `compose.yml` | Base Docker Compose | All |
| `compose.dev.yml` | Development overrides | Development |
| `compose.prod.yml` | Production overrides | Production |
| `docker/nginx.conf` | Reverse proxy config | Production |
| `docker/prometheus.yml` | Monitoring config | All |

### 🚀 Deployment Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `scripts/setup/install.sh` | Interactive installer | `make install` |
| `scripts/setup/quick-setup.sh` | Quick 3-mode setup | `make quick-setup` |
| `scripts/setup/validate-config.sh` | Config validation | `make validate` |
| `scripts/deployment/deploy.sh` | Production deployment | `make deploy` |
| `scripts/deployment/healthcheck.sh` | Health validation | `make health-check` |
| `scripts/maintenance/backup.sh` | Backup automation | `make backup` |

### 🏗️ Application Architecture

#### **cmd/**: Application Entry Points
- `ffprobe-api/main.go`: Main application server

#### **internal/**: Private Application Code

| Package | Purpose | Key Components |
|---------|---------|----------------|
| `api/` | HTTP routing | Route definitions |
| `batch/` | Batch processing | Job validation |
| `config/` | Configuration | Environment setup |
| `database/` | Data layer | Repositories, connections |
| `ffmpeg/` | FFmpeg integration | FFprobe wrapper |
| `handlers/` | HTTP handlers | Request processing |
| `hls/` | HLS streaming | Stream analysis |
| `middleware/` | HTTP middleware | Auth, logging, security |
| `models/` | Data models | Structs, validation |
| `quality/` | Quality analysis | VMAF, metrics |
| `reports/` | Report generation | PDF, HTML output |
| `services/` | Business logic | Core functionality |
| `storage/` | Storage providers | Local, S3, GCS, Azure |

#### **pkg/**: Public Libraries
- `logger/`: Structured logging with correlation IDs

### 📚 Documentation Organization

#### **docs/api/**: API Reference
- Complete API documentation with examples
- Authentication and security guides
- OpenAPI specification

#### **docs/deployment/**: Deployment Guides
- Configuration reference
- Docker security audit
- Production readiness checklist

#### **docs/tutorials/**: User Guides
- Step-by-step API usage
- Integration examples
- Best practices

## 🔒 Security Considerations

### **File Permissions**
- **Scripts**: `755` (executable)
- **Configs**: `644` (readable)
- **Secrets**: `600` (owner only)

### **Sensitive Files** (Gitignored)
- `.env` - Environment configuration
- `data/` - Application data
- `ssl/` - SSL certificates
- `backups/` - Database backups

## 🛠️ Development Workflow

### **Getting Started**
```bash
# 1. Clone repository
git clone https://github.com/your-org/ffprobe-api.git

# 2. Quick setup
make quick-setup

# 3. Development workflow
make dev-workflow
```

### **Make Targets**

| Target | Purpose | Usage |
|--------|---------|-------|
| `make install` | Interactive installer | Setup |
| `make validate` | Config validation | Pre-deploy |
| `make test` | Run tests | Development |
| `make build` | Build application | Development |
| `make deploy` | Deploy to production | Operations |
| `make backup` | Create backup | Maintenance |

## 📊 Code Organization Principles

### **Separation of Concerns**
- **cmd/**: Entry points only
- **internal/**: Business logic
- **pkg/**: Reusable libraries
- **scripts/**: Operational tools

### **Dependency Direction**
```
cmd/ → internal/ → pkg/
  ↓       ↓        ↓
tests/   docs/   scripts/
```

### **Layer Architecture**
1. **Presentation**: HTTP handlers, middleware
2. **Business**: Services, validation
3. **Data**: Repositories, models
4. **Infrastructure**: Database, storage, external APIs

## 🔄 CI/CD Integration

### **GitHub Actions** (`.github/workflows/`)
- Automated testing
- Security scanning
- Docker image building
- Deployment automation

### **Quality Gates**
- Unit tests: `>80%` coverage
- Integration tests: All pass
- Security scan: No critical issues
- Lint check: Clean code

## 📈 Scalability Considerations

### **Horizontal Scaling**
- Stateless application design
- External session storage (Redis)
- Load balancer ready (Nginx)

### **Vertical Scaling**
- Resource limits in Docker Compose
- Configurable worker pools
- Memory optimization

## 🗃️ Data Management

### **Persistent Data**
- **Database**: PostgreSQL with migrations
- **Cache**: Redis for sessions/cache
- **Files**: Uploads, reports, models
- **Logs**: Structured JSON logging

### **Backup Strategy**
- Database: Daily automated backups
- Files: Incremental backups
- Configuration: Version controlled

## 🌐 Deployment Environments

### **Development**
- Debug logging
- Hot reload
- Development tools (Adminer, Redis Commander)

### **Staging**
- Production-like setup
- Testing integrations
- Pre-production validation

### **Production**
- Security hardening
- SSL/TLS encryption
- Monitoring and alerting
- Backup automation

---

This repository structure follows industry best practices for:
- **🏗️ Clean Architecture** - Clear separation of concerns
- **🔒 Security** - Defense in depth
- **🚀 DevOps** - Infrastructure as code
- **📚 Documentation** - Comprehensive guides
- **🧪 Testing** - Quality assurance
- **📊 Monitoring** - Observability

**🎬 Ready for enterprise deployment!** 🚀