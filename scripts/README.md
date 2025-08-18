# 📁 Scripts Directory

This directory contains all operational scripts for the FFprobe API project, organized by category.

## 📂 Directory Structure

```
scripts/
├── setup/              # Installation and configuration scripts
├── deployment/         # Deployment and operational scripts  
├── maintenance/        # Backup and maintenance scripts
└── README.md          # This file
```

## 🔧 Setup Scripts

| Script | Description | Usage |
|--------|-------------|-------|
| [`setup-ollama.sh`](setup/setup-ollama.sh) | **Ollama setup** - Configure local AI models | `./scripts/setup/setup-ollama.sh` |
| [`../setup.sh`](../setup.sh) | **Smart installer** - System requirements check & deployment | `./setup.sh` |
| [`setup/validate-config.sh`](setup/validate-config.sh) | **Configuration validator** - Validates env files | `./scripts/setup/validate-config.sh [.env]` |

### 🎯 Smart Installer Features
The main `setup.sh` script provides comprehensive deployment with system validation:

- **🔧 Deployment Mode Selection** (Quick/Minimal/Production/Development)
- **⚙️ System Requirements Validation** (RAM, CPU, disk, ports)
- **🐳 Docker Capability Testing** (container runtime validation)
- **🔐 Automatic Security Configuration** (API keys, passwords)
- **📊 Optional Monitoring Setup** (Prometheus, Grafana)
- **🤖 AI Model Integration** (Ollama with local LLM)

## 🚀 Deployment Scripts

| Script | Description | Usage |
|--------|-------------|-------|
| [`deployment/deploy.sh`](deployment/deploy.sh) | **Production deployer** - Full deployment automation | `./scripts/deployment/deploy.sh deploy production v1.0.0` |
| [`deployment/healthcheck.sh`](deployment/healthcheck.sh) | **Health checker** - Service health validation | `./scripts/deployment/healthcheck.sh` |

### 🏭 Deployment Commands
```bash
# Deploy to production
./scripts/deployment/deploy.sh deploy production v1.0.0

# Deploy to staging
./scripts/deployment/deploy.sh deploy staging latest

# Check deployment status
./scripts/deployment/deploy.sh status production

# Rollback deployment
./scripts/deployment/deploy.sh rollback production
```

## 🔧 Maintenance Scripts

| Script | Description | Usage |
|--------|-------------|-------|
| [`maintenance/backup.sh`](maintenance/backup.sh) | **Backup system** - Database and file backups | `./scripts/maintenance/backup.sh` |

### 💾 Backup Operations
```bash
# Create full backup
./scripts/maintenance/backup.sh

# Restore from backup
./scripts/maintenance/backup.sh restore 20241127_120000

# List available backups
./scripts/maintenance/backup.sh list
```

## 🚀 Quick Start Guide

### 1. **New Installation**
```bash
# Run smart installer with system validation (recommended)
./setup.sh

# OR non-interactive mode
./setup.sh --quick  # or --minimal, --production, --development
```

### 2. **Existing Configuration**
```bash
# Validate existing config
./scripts/setup/validate-config.sh .env

# Deploy with existing config
./scripts/deployment/deploy.sh deploy production latest
```

### 3. **Development Workflow**
```bash
# Quick development setup
./scripts/setup/quick-setup.sh
# Select option 1 (Development)

# Deploy changes
./scripts/deployment/deploy.sh deploy development latest

# Check health
./scripts/deployment/healthcheck.sh
```

## 🛡️ Security Notes

- **🔐 Scripts handle sensitive data** - API keys, passwords, certificates
- **🔒 Files are created with secure permissions** (600 for configs)
- **🚫 No secrets are logged** - Passwords are masked in output
- **✅ Configuration validation** - Prevents insecure deployments

## 📋 Requirements

### System Requirements
- **Docker & Docker Compose v2.20+**
- **Bash 4.0+** (macOS may need `brew install bash`)
- **curl, wget, openssl** for various operations
- **sudo privileges** for system configuration

### Optional Tools
- **yq** - For YAML processing (improves Docker Compose updates)
- **jq** - For JSON processing
- **certbot** - For Let's Encrypt SSL certificates

## 🔍 Script Details

### Interactive Installer (`setup/install.sh`)
The most comprehensive setup option with full configuration collection:

**Features:**
- ✅ System requirements checking
- ✅ Deployment mode selection (5 modes)
- ✅ Security configuration with validation
- ✅ Network and SSL setup
- ✅ Cloud storage integration
- ✅ Resource allocation
- ✅ Monitoring configuration
- ✅ Post-install verification

**Usage:**
```bash
./scripts/setup/install.sh
# Follow the interactive prompts
```

### Quick Setup (`setup/quick-setup.sh`)
Simplified setup for rapid deployment:

**Modes:**
1. **🔧 Development** - Local dev, no auth, debug logging
2. **🧪 Demo** - Basic auth, sample configuration
3. **🏭 Production** - Full security, SSL ready

**Usage:**
```bash
./scripts/setup/quick-setup.sh
# Select mode 1, 2, or 3
```

### Configuration Validator (`setup/validate-config.sh`)
Comprehensive validation before deployment:

**Checks:**
- ✅ Required environment variables
- ✅ Security requirements (key lengths, patterns)
- ✅ Network configuration (ports, domains)
- ✅ Production-specific requirements
- ✅ Performance recommendations

**Usage:**
```bash
./scripts/setup/validate-config.sh          # Validates .env
./scripts/setup/validate-config.sh .env.prod # Validates specific file
```

## 🆘 Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   chmod +x scripts/setup/install.sh
   chmod +x scripts/**/*.sh
   ```

2. **Docker Not Found**
   ```bash
   # Install Docker first
   # https://docs.docker.com/get-docker/
   ```

3. **Script Fails on macOS**
   ```bash
   # Install GNU bash
   brew install bash
   # Use: /usr/local/bin/bash scripts/setup/install.sh
   ```

4. **SSL Certificate Issues**
   ```bash
   # Check domain DNS settings
   # Ensure ports 80/443 are open
   # Run: ./scripts/setup/validate-config.sh
   ```

### Getting Help

- **📖 Documentation**: [`../docs/`](../docs/)
- **🔧 Configuration Guide**: [`../docs/deployment/configuration.md`](../docs/deployment/configuration.md)
- **🎯 API Examples**: [`../docs/tutorials/api_usage.md`](../docs/tutorials/api_usage.md)
- **🔒 Security Guide**: [`../SECURITY_AUDIT_REPORT.md`](../SECURITY_AUDIT_REPORT.md)

## 📊 Script Compatibility

| Script | Linux | macOS | Windows WSL |
|--------|-------|-------|-------------|
| install.sh | ✅ | ✅ | ✅ |
| quick-setup.sh | ✅ | ✅ | ✅ |
| validate-config.sh | ✅ | ✅ | ✅ |
| deploy.sh | ✅ | ✅ | ✅ |
| backup.sh | ✅ | ✅ | ✅ |

---

**🎬 Ready to deploy your FFprobe API!** Choose the script that best fits your needs and get started. 🚀