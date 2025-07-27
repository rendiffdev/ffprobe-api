#!/bin/bash

# FFprobe API Configuration Validator
# Validates environment configuration before deployment

set -euo pipefail

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[VALIDATOR] $1${NC}"; }
warn() { echo -e "${YELLOW}[WARNING] $1${NC}"; }
error() { echo -e "${RED}[ERROR] $1${NC}"; }
info() { echo -e "${BLUE}[INFO] $1${NC}"; }

# Configuration file
ENV_FILE="${1:-.env}"

if [[ ! -f "$ENV_FILE" ]]; then
    error "Configuration file '$ENV_FILE' not found!"
    echo "Usage: $0 [env-file]"
    echo "Example: $0 .env"
    exit 1
fi

log "Validating configuration file: $ENV_FILE"

# Source the environment file
set -a
source "$ENV_FILE"
set +a

# Validation counters
ERRORS=0
WARNINGS=0
CHECKS=0

# Helper functions
check_required() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    local description="$2"
    
    ((CHECKS++))
    
    if [[ -z "$var_value" ]]; then
        error "Missing required variable: $var_name ($description)"
        ((ERRORS++))
        return 1
    fi
    
    return 0
}

check_min_length() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    local min_length="$2"
    local description="$3"
    
    ((CHECKS++))
    
    if [[ ${#var_value} -lt $min_length ]]; then
        error "$var_name must be at least $min_length characters ($description)"
        ((ERRORS++))
        return 1
    fi
    
    return 0
}

check_format() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    local pattern="$2"
    local description="$3"
    
    ((CHECKS++))
    
    if [[ -n "$var_value" && ! "$var_value" =~ $pattern ]]; then
        error "$var_name has invalid format ($description)"
        ((ERRORS++))
        return 1
    fi
    
    return 0
}

warn_insecure() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    local insecure_pattern="$2"
    local description="$3"
    
    ((CHECKS++))
    
    if [[ "$var_value" =~ $insecure_pattern ]]; then
        warn "$var_name appears to use default/insecure value ($description)"
        ((WARNINGS++))
        return 1
    fi
    
    return 0
}

echo ""
info "🔍 Starting configuration validation..."
echo ""

# =============================================================================
# Core Configuration Validation
# =============================================================================

log "Checking core configuration..."

check_required "ENVIRONMENT" "Deployment environment"
check_required "LOG_LEVEL" "Logging level"

if [[ -n "${ENVIRONMENT:-}" ]]; then
    if [[ ! "$ENVIRONMENT" =~ ^(development|staging|production)$ ]]; then
        error "ENVIRONMENT must be one of: development, staging, production"
        ((ERRORS++))
    fi
fi

if [[ -n "${LOG_LEVEL:-}" ]]; then
    if [[ ! "$LOG_LEVEL" =~ ^(debug|info|warn|error)$ ]]; then
        error "LOG_LEVEL must be one of: debug, info, warn, error"
        ((ERRORS++))
    fi
fi

# =============================================================================
# Security Configuration Validation
# =============================================================================

log "Checking security configuration..."

# API Key validation
if check_required "API_KEY" "API authentication key"; then
    check_min_length "API_KEY" 32 "Security requirement"
    warn_insecure "API_KEY" "change_this|dev_|test_|example" "Use a secure random key"
fi

# JWT Secret validation
if check_required "JWT_SECRET" "JWT token signing secret"; then
    check_min_length "JWT_SECRET" 32 "Security requirement"
    warn_insecure "JWT_SECRET" "change_this|dev_|test_|secret|jwt" "Use a secure random secret"
fi

# Database password validation
if check_required "POSTGRES_PASSWORD" "Database password"; then
    check_min_length "POSTGRES_PASSWORD" 8 "Security recommendation"
    warn_insecure "POSTGRES_PASSWORD" "change_this|password|admin|dev_|test_" "Use a strong password"
fi

# Redis password validation
if [[ -n "${REDIS_PASSWORD:-}" ]]; then
    check_min_length "REDIS_PASSWORD" 8 "Security recommendation"
    warn_insecure "REDIS_PASSWORD" "change_this|password|redis|dev_|test_" "Use a strong password"
fi

# =============================================================================
# Network Configuration Validation
# =============================================================================

log "Checking network configuration..."

check_required "API_PORT" "API server port"

if [[ -n "${API_PORT:-}" ]]; then
    if [[ ! "$API_PORT" =~ ^[0-9]+$ ]] || [[ "$API_PORT" -lt 1 ]] || [[ "$API_PORT" -gt 65535 ]]; then
        error "API_PORT must be a valid port number (1-65535)"
        ((ERRORS++))
    fi
fi

# Production-specific network validation
if [[ "${ENVIRONMENT:-}" == "production" ]]; then
    if [[ -z "${DOMAIN_NAME:-}" ]]; then
        warn "DOMAIN_NAME should be set for production deployment"
        ((WARNINGS++))
    elif [[ -n "${DOMAIN_NAME:-}" ]]; then
        check_format "DOMAIN_NAME" '^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$' "Valid domain name format"
    fi
    
    if [[ -n "${EMAIL:-}" ]]; then
        check_format "EMAIL" '^[^@]+@[^@]+\.[^@]+$' "Valid email format"
    fi
fi

# =============================================================================
# Storage Configuration Validation
# =============================================================================

log "Checking storage configuration..."

check_required "DATA_PATH" "Data storage path"
check_required "MAX_FILE_SIZE" "Maximum file size limit"

if [[ -n "${MAX_FILE_SIZE:-}" ]]; then
    if [[ ! "$MAX_FILE_SIZE" =~ ^[0-9]+$ ]]; then
        error "MAX_FILE_SIZE must be a number (bytes)"
        ((ERRORS++))
    fi
fi

# Cloud storage validation
if [[ "${STORAGE_TYPE:-local}" != "local" ]]; then
    case "${STORAGE_TYPE}" in
        "s3")
            check_required "AWS_ACCESS_KEY_ID" "AWS access key"
            check_required "AWS_SECRET_ACCESS_KEY" "AWS secret key" 
            check_required "S3_BUCKET" "S3 bucket name"
            ;;
        "gcs")
            check_required "GCS_BUCKET" "GCS bucket name"
            check_required "GCP_PROJECT_ID" "GCP project ID"
            ;;
        "azure")
            check_required "AZURE_STORAGE_ACCOUNT" "Azure storage account"
            check_required "AZURE_STORAGE_KEY" "Azure storage key"
            ;;
        *)
            error "Invalid STORAGE_TYPE: ${STORAGE_TYPE}. Must be: local, s3, gcs, azure"
            ((ERRORS++))
            ;;
    esac
fi

# =============================================================================
# Rate Limiting Validation
# =============================================================================

if [[ "${ENABLE_RATE_LIMIT:-}" == "true" ]]; then
    log "Checking rate limiting configuration..."
    
    check_required "RATE_LIMIT_PER_MINUTE" "Rate limit per minute"
    check_required "RATE_LIMIT_PER_HOUR" "Rate limit per hour"
    check_required "RATE_LIMIT_PER_DAY" "Rate limit per day"
    
    # Validate rate limit numbers
    for rate_var in RATE_LIMIT_PER_MINUTE RATE_LIMIT_PER_HOUR RATE_LIMIT_PER_DAY; do
        if [[ -n "${!rate_var:-}" && ! "${!rate_var}" =~ ^[0-9]+$ ]]; then
            error "$rate_var must be a positive number"
            ((ERRORS++))
        fi
    done
fi

# =============================================================================
# Performance Configuration Validation
# =============================================================================

log "Checking performance configuration..."

if [[ -n "${MAX_CONCURRENT_JOBS:-}" ]]; then
    if [[ ! "$MAX_CONCURRENT_JOBS" =~ ^[0-9]+$ ]] || [[ "$MAX_CONCURRENT_JOBS" -lt 1 ]]; then
        error "MAX_CONCURRENT_JOBS must be a positive number"
        ((ERRORS++))
    elif [[ "$MAX_CONCURRENT_JOBS" -gt 16 ]]; then
        warn "MAX_CONCURRENT_JOBS > 16 may cause high resource usage"
        ((WARNINGS++))
    fi
fi

# =============================================================================
# Database Configuration Validation
# =============================================================================

log "Checking database configuration..."

check_required "POSTGRES_HOST" "Database host"
check_required "POSTGRES_PORT" "Database port"
check_required "POSTGRES_DB" "Database name"
check_required "POSTGRES_USER" "Database user"

if [[ -n "${POSTGRES_PORT:-}" ]]; then
    if [[ ! "$POSTGRES_PORT" =~ ^[0-9]+$ ]] || [[ "$POSTGRES_PORT" -lt 1 ]] || [[ "$POSTGRES_PORT" -gt 65535 ]]; then
        error "POSTGRES_PORT must be a valid port number"
        ((ERRORS++))
    fi
fi

# =============================================================================
# SSL/TLS Configuration Validation
# =============================================================================

if [[ "${ENVIRONMENT:-}" == "production" && "${SSL_TYPE:-disabled}" != "disabled" ]]; then
    log "Checking SSL/TLS configuration..."
    
    case "${SSL_TYPE}" in
        "letsencrypt")
            if [[ -z "${EMAIL:-}" ]]; then
                error "EMAIL is required for Let's Encrypt certificates"
                ((ERRORS++))
            fi
            if [[ -z "${DOMAIN_NAME:-}" ]]; then
                error "DOMAIN_NAME is required for Let's Encrypt certificates"
                ((ERRORS++))
            fi
            ;;
        "custom")
            check_required "SSL_CERT_PATH" "SSL certificate path"
            check_required "SSL_KEY_PATH" "SSL private key path"
            ;;
        *)
            error "Invalid SSL_TYPE: ${SSL_TYPE}. Must be: letsencrypt, custom, disabled"
            ((ERRORS++))
            ;;
    esac
fi

# =============================================================================
# Monitoring Configuration Validation
# =============================================================================

if [[ "${ENABLE_GRAFANA:-}" == "true" ]]; then
    log "Checking monitoring configuration..."
    
    if [[ -n "${GRAFANA_PASSWORD:-}" ]]; then
        warn_insecure "GRAFANA_PASSWORD" "admin|change_this|password" "Use a strong Grafana password"
    fi
fi

# =============================================================================
# Development Specific Warnings
# =============================================================================

if [[ "${ENVIRONMENT:-}" == "development" ]]; then
    log "Checking development configuration..."
    
    if [[ "${ENABLE_AUTH:-}" == "true" ]]; then
        info "Authentication is enabled in development mode"
    fi
    
    if [[ "${ENABLE_RATE_LIMIT:-}" == "true" ]]; then
        info "Rate limiting is enabled in development mode"
    fi
fi

# =============================================================================
# Production Specific Validation
# =============================================================================

if [[ "${ENVIRONMENT:-}" == "production" ]]; then
    log "Checking production configuration..."
    
    if [[ "${ENABLE_AUTH:-}" != "true" ]]; then
        error "Authentication must be enabled in production"
        ((ERRORS++))
    fi
    
    if [[ "${ENABLE_RATE_LIMIT:-}" != "true" ]]; then
        warn "Rate limiting should be enabled in production"
        ((WARNINGS++))
    fi
    
    if [[ "${LOG_LEVEL:-}" == "debug" ]]; then
        warn "Debug logging is enabled in production (may impact performance)"
        ((WARNINGS++))
    fi
    
    if [[ "${SSL_TYPE:-}" == "disabled" ]]; then
        warn "SSL is disabled in production (not recommended)"
        ((WARNINGS++))
    fi
fi

# =============================================================================
# Validation Summary
# =============================================================================

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "                         🔍 VALIDATION SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

info "Configuration file: $ENV_FILE"
info "Total checks performed: $CHECKS"

if [[ $ERRORS -eq 0 && $WARNINGS -eq 0 ]]; then
    log "✅ Configuration validation PASSED"
    echo "   🎉 All checks passed successfully!"
    echo "   🚀 Configuration is ready for deployment"
elif [[ $ERRORS -eq 0 ]]; then
    warn "⚠️  Configuration validation PASSED with warnings"
    echo "   ⚠️  Found $WARNINGS warning(s) - please review"
    echo "   🚀 Configuration can be deployed with caution"
else
    error "❌ Configuration validation FAILED"
    echo "   ❌ Found $ERRORS error(s) and $WARNINGS warning(s)"
    echo "   🛑 Fix errors before deployment"
fi

echo ""

# Deployment readiness check
if [[ $ERRORS -eq 0 ]]; then
    echo "🎯 DEPLOYMENT READINESS:"
    
    if [[ "${ENVIRONMENT:-}" == "development" ]]; then
        log "✅ Ready for development deployment"
        echo "   🔧 Use: docker compose -f compose.yml -f compose.dev.yml up -d"
    elif [[ "${ENVIRONMENT:-}" == "production" ]]; then
        if [[ $WARNINGS -eq 0 ]]; then
            log "✅ Ready for production deployment"
        else
            warn "⚠️  Ready for production with warnings"
        fi
        echo "   🏭 Use: docker compose -f compose.yml -f compose.prod.yml up -d"
    else
        log "✅ Ready for deployment"
        echo "   🚀 Use appropriate compose configuration for your environment"
    fi
    
    echo ""
    echo "📋 NEXT STEPS:"
    echo "   1. Review any warnings above"
    echo "   2. Test configuration: docker compose config"
    echo "   3. Deploy application: scripts/deploy.sh or docker compose up"
    echo "   4. Verify health: curl http://localhost:${API_PORT:-8080}/health"
else
    echo "🛑 DEPLOYMENT BLOCKED:"
    echo "   ❌ Fix all errors before attempting deployment"
    echo "   📝 Update $ENV_FILE with correct values"
    echo "   🔄 Run validation again: $0 $ENV_FILE"
fi

echo ""

# Exit with appropriate code
exit $ERRORS