#!/bin/bash
# FFprobe API - Production Secrets Manager
# Secure secrets generation, rotation, and management for Docker Swarm/Compose

set -euo pipefail

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SECRETS_DIR="${SCRIPT_DIR}/../secrets"
readonly BACKUP_DIR="${SECRETS_DIR}/backup"
readonly LOG_FILE="${SECRETS_DIR}/secrets-audit.log"

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Logging function
log_audit() {
    local level="$1"
    local message="$2"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    echo "[${timestamp}] ${level}: ${message}" | tee -a "${LOG_FILE}"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
    log_audit "INFO" "$1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    log_audit "SUCCESS" "$1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    log_audit "WARNING" "$1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    log_audit "ERROR" "$1"
}

show_usage() {
    cat << EOF
FFprobe API - Production Secrets Manager

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    generate                Generate all secrets for first-time setup
    rotate [SECRET_NAME]    Rotate specific secret or all secrets
    backup                  Create encrypted backup of secrets
    restore [BACKUP_FILE]   Restore secrets from backup
    verify                  Verify all secrets exist and are valid
    clean                   Remove old backup files
    list                    List all managed secrets
    
OPTIONS:
    --dry-run              Show what would be done without executing
    --force                Force operation without confirmation
    --backup-passphrase    Passphrase for encrypted backups
    --help                 Show this help message

EXAMPLES:
    $0 generate                    # Generate all secrets for first setup
    $0 rotate valkey_password      # Rotate only Valkey password
    $0 rotate --force              # Rotate all secrets without confirmation
    $0 backup --backup-passphrase mysecret123
    $0 restore backup-2024-01-15.tar.gz.gpg

SECRETS MANAGED:
    - valkey_password              Redis/Valkey authentication
    - jwt_secret                   JWT token signing
    - csrf_secret                  CSRF protection
    - api_key                      API authentication
    - grafana_password            Grafana admin password
    - backup_encryption_key       Backup encryption key
    - ssl_certificate             SSL certificate (if self-signed)
    - ssl_private_key             SSL private key
    
ENVIRONMENT VARIABLES:
    SECRETS_PASSPHRASE            Master passphrase for secret encryption
    BACKUP_RETENTION_DAYS         Days to keep backup files (default: 30)
    SECRET_LENGTH                 Length for generated secrets (default: 32)

EOF
}

# Initialize secrets directory structure
init_secrets_dir() {
    log_info "Initializing secrets directory structure"
    
    mkdir -p "${SECRETS_DIR}"
    mkdir -p "${BACKUP_DIR}"
    
    # Set secure permissions
    chmod 700 "${SECRETS_DIR}"
    chmod 700 "${BACKUP_DIR}"
    
    # Create audit log if it doesn't exist
    touch "${LOG_FILE}"
    chmod 600 "${LOG_FILE}"
    
    log_success "Secrets directory initialized"
}

# Generate a secure random string
generate_secret() {
    local length="${1:-32}"
    
    if command -v openssl &> /dev/null; then
        openssl rand -base64 "${length}" | tr -d "=+/" | cut -c1-"${length}"
    elif command -v head &> /dev/null && [[ -c /dev/urandom ]]; then
        head -c "${length}" /dev/urandom | base64 | tr -d "=+/" | cut -c1-"${length}"
    else
        log_error "No suitable random generator found"
        return 1
    fi
}

# Generate a secure password
generate_password() {
    local length="${1:-24}"
    
    if command -v openssl &> /dev/null; then
        openssl rand -base64 32 | tr -d "=+/" | cut -c1-"${length}"
    else
        generate_secret "${length}"
    fi
}

# Generate JWT secret with specific requirements
generate_jwt_secret() {
    # JWT secrets should be at least 256 bits (32 bytes)
    openssl rand -hex 32
}

# Generate API key
generate_api_key() {
    # API keys: alphanumeric, 40 characters
    openssl rand -hex 20
}

# Generate or update a single secret
generate_single_secret() {
    local secret_name="$1"
    local secret_file="${SECRETS_DIR}/${secret_name}.txt"
    local backup_file="${BACKUP_DIR}/${secret_name}-$(date +%Y%m%d-%H%M%S).txt"
    
    # Backup existing secret if it exists
    if [[ -f "${secret_file}" ]]; then
        log_info "Backing up existing ${secret_name}"
        cp "${secret_file}" "${backup_file}"
    fi
    
    log_info "Generating ${secret_name}"
    
    case "${secret_name}" in
        "valkey_password"|"grafana_password")
            generate_password 24 > "${secret_file}"
            ;;
        "jwt_secret")
            generate_jwt_secret > "${secret_file}"
            ;;
        "csrf_secret")
            generate_secret 32 > "${secret_file}"
            ;;
        "api_key")
            generate_api_key > "${secret_file}"
            ;;
        "backup_encryption_key")
            generate_secret 64 > "${secret_file}"
            ;;
        *)
            log_warning "Unknown secret type: ${secret_name}. Using default generator."
            generate_secret 32 > "${secret_file}"
            ;;
    esac
    
    # Set secure permissions
    chmod 600 "${secret_file}"
    
    log_success "Generated ${secret_name}"
}

# Generate SSL certificates (self-signed for development)
generate_ssl_certificates() {
    local domain="${SSL_DOMAIN:-localhost}"
    local cert_file="${SECRETS_DIR}/ssl_certificate.pem"
    local key_file="${SECRETS_DIR}/ssl_private_key.pem"
    
    if [[ "${GENERATE_SSL:-false}" == "true" ]]; then
        log_info "Generating self-signed SSL certificates for ${domain}"
        
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout "${key_file}" \
            -out "${cert_file}" \
            -subj "/C=US/ST=State/L=City/O=Organization/CN=${domain}" \
            -addext "subjectAltName=DNS:${domain},DNS:*.${domain},IP:127.0.0.1"
        
        chmod 600 "${key_file}"
        chmod 644 "${cert_file}"
        
        log_success "SSL certificates generated"
    else
        log_info "Skipping SSL certificate generation (set GENERATE_SSL=true to enable)"
    fi
}

# Generate all secrets
generate_all_secrets() {
    log_info "Generating all secrets for FFprobe API"
    
    local secrets=(
        "valkey_password"
        "jwt_secret"
        "csrf_secret"
        "api_key"
        "grafana_password"
        "backup_encryption_key"
    )
    
    for secret in "${secrets[@]}"; do
        generate_single_secret "${secret}"
    done
    
    generate_ssl_certificates
    
    log_success "All secrets generated successfully"
}

# Rotate secrets
rotate_secrets() {
    local secret_name="${1:-}"
    
    if [[ -n "${secret_name}" ]]; then
        log_info "Rotating secret: ${secret_name}"
        generate_single_secret "${secret_name}"
    else
        log_info "Rotating all secrets"
        
        if [[ "${FORCE:-false}" != "true" ]]; then
            echo -n "This will rotate ALL secrets. Continue? (y/N): "
            read -r response
            if [[ "${response}" != "y" && "${response}" != "Y" ]]; then
                log_info "Secret rotation cancelled"
                return 0
            fi
        fi
        
        generate_all_secrets
    fi
    
    log_success "Secret rotation completed"
}

# Create encrypted backup
create_backup() {
    local backup_file="backup-$(date +%Y-%m-%d-%H%M%S).tar.gz"
    local encrypted_backup="${backup_file}.gpg"
    local backup_path="${BACKUP_DIR}/${encrypted_backup}"
    
    log_info "Creating encrypted backup of secrets"
    
    # Create tar archive of secrets
    tar -czf "/tmp/${backup_file}" -C "${SECRETS_DIR}" \
        --exclude='backup' \
        --exclude='*.log' \
        .
    
    # Encrypt with GPG
    if [[ -n "${BACKUP_PASSPHRASE:-}" ]]; then
        gpg --batch --yes --passphrase "${BACKUP_PASSPHRASE}" \
            --symmetric --cipher-algo AES256 \
            --output "${backup_path}" \
            "/tmp/${backup_file}"
        
        # Clean up temporary file
        rm "/tmp/${backup_file}"
        
        log_success "Encrypted backup created: ${backup_path}"
    else
        log_error "BACKUP_PASSPHRASE not set. Cannot create encrypted backup."
        return 1
    fi
}

# Restore from backup
restore_backup() {
    local backup_file="$1"
    
    if [[ ! -f "${backup_file}" ]]; then
        log_error "Backup file not found: ${backup_file}"
        return 1
    fi
    
    log_info "Restoring secrets from backup: ${backup_file}"
    
    # Decrypt and extract
    if [[ -n "${BACKUP_PASSPHRASE:-}" ]]; then
        gpg --batch --yes --passphrase "${BACKUP_PASSPHRASE}" \
            --decrypt "${backup_file}" | \
            tar -xzf - -C "${SECRETS_DIR}"
        
        # Fix permissions
        chmod 700 "${SECRETS_DIR}"
        find "${SECRETS_DIR}" -name "*.txt" -exec chmod 600 {} \;
        
        log_success "Secrets restored from backup"
    else
        log_error "BACKUP_PASSPHRASE not set. Cannot decrypt backup."
        return 1
    fi
}

# Verify secrets exist and are valid
verify_secrets() {
    log_info "Verifying secrets"
    
    local secrets=(
        "valkey_password"
        "jwt_secret"
        "csrf_secret"
        "api_key"
        "grafana_password"
        "backup_encryption_key"
    )
    
    local missing_secrets=()
    local invalid_secrets=()
    
    for secret in "${secrets[@]}"; do
        local secret_file="${SECRETS_DIR}/${secret}.txt"
        
        if [[ ! -f "${secret_file}" ]]; then
            missing_secrets+=("${secret}")
        elif [[ ! -s "${secret_file}" ]]; then
            invalid_secrets+=("${secret}")
        elif [[ $(stat -c %a "${secret_file}") != "600" ]]; then
            log_warning "Incorrect permissions for ${secret}"
            chmod 600 "${secret_file}"
        fi
    done
    
    if [[ ${#missing_secrets[@]} -gt 0 ]]; then
        log_error "Missing secrets: ${missing_secrets[*]}"
        return 1
    fi
    
    if [[ ${#invalid_secrets[@]} -gt 0 ]]; then
        log_error "Invalid (empty) secrets: ${invalid_secrets[*]}"
        return 1
    fi
    
    log_success "All secrets verified"
}

# Clean old backups
clean_backups() {
    local retention_days="${BACKUP_RETENTION_DAYS:-30}"
    
    log_info "Cleaning backups older than ${retention_days} days"
    
    find "${BACKUP_DIR}" -name "*.gpg" -type f -mtime +"${retention_days}" -delete
    find "${BACKUP_DIR}" -name "*.txt" -type f -mtime +"${retention_days}" -delete
    
    log_success "Old backups cleaned"
}

# List all secrets
list_secrets() {
    log_info "Managed secrets:"
    
    local secrets=(
        "valkey_password"
        "jwt_secret"
        "csrf_secret"
        "api_key"
        "grafana_password"
        "backup_encryption_key"
        "ssl_certificate"
        "ssl_private_key"
    )
    
    for secret in "${secrets[@]}"; do
        local secret_file="${SECRETS_DIR}/${secret}.txt"
        local pem_file="${SECRETS_DIR}/${secret}.pem"
        
        if [[ -f "${secret_file}" ]]; then
            local size=$(stat -c%s "${secret_file}")
            local modified=$(stat -c%y "${secret_file}")
            echo "  ✓ ${secret} (${size} bytes, modified: ${modified})"
        elif [[ -f "${pem_file}" ]]; then
            local size=$(stat -c%s "${pem_file}")
            local modified=$(stat -c%y "${pem_file}")
            echo "  ✓ ${secret} (${size} bytes, modified: ${modified})"
        else
            echo "  ✗ ${secret} (missing)"
        fi
    done
}

# Main function
main() {
    local command="${1:-}"
    shift || true
    
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --backup-passphrase)
                BACKUP_PASSPHRASE="$2"
                shift 2
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                break
                ;;
        esac
    done
    
    # Initialize directory structure
    init_secrets_dir
    
    case "${command}" in
        generate)
            generate_all_secrets
            ;;
        rotate)
            rotate_secrets "$@"
            ;;
        backup)
            create_backup
            ;;
        restore)
            restore_backup "$@"
            ;;
        verify)
            verify_secrets
            ;;
        clean)
            clean_backups
            ;;
        list)
            list_secrets
            ;;
        *)
            log_error "Unknown command: ${command}"
            show_usage
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"