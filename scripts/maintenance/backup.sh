#!/bin/sh
# Database backup script for production

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/app/backup}"
DB_HOST="${POSTGRES_HOST:-postgres}"
DB_PORT="${POSTGRES_PORT:-5432}"
DB_NAME="${POSTGRES_DB:-ffprobe_api}"
DB_USER="${POSTGRES_USER:-ffprobe}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/ffprobe_api_backup_${TIMESTAMP}.sql"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"

# Create backup directory if not exists
mkdir -p "$BACKUP_DIR"

echo "🔄 Starting database backup..."
echo "Timestamp: ${TIMESTAMP}"

# Perform backup
PGPASSWORD="${POSTGRES_PASSWORD}" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --verbose \
    --no-owner \
    --no-acl \
    --format=custom \
    --file="$BACKUP_FILE"

# Compress backup
gzip "$BACKUP_FILE"
BACKUP_FILE="${BACKUP_FILE}.gz"

# Get file size
SIZE=$(ls -lh "$BACKUP_FILE" | awk '{print $5}')

echo "✅ Backup completed successfully!"
echo "File: $BACKUP_FILE"
echo "Size: $SIZE"

# Clean old backups
echo "🧹 Cleaning old backups (older than ${RETENTION_DAYS} days)..."
find "$BACKUP_DIR" -name "ffprobe_api_backup_*.sql.gz" -mtime +${RETENTION_DAYS} -delete

# List remaining backups
echo "📦 Current backups:"
ls -lh "$BACKUP_DIR"/ffprobe_api_backup_*.sql.gz 2>/dev/null || echo "No backups found"

# Upload to cloud storage if configured
if [ -n "$BACKUP_S3_BUCKET" ]; then
    echo "☁️ Uploading to S3..."
    aws s3 cp "$BACKUP_FILE" "s3://${BACKUP_S3_BUCKET}/database-backups/$(basename $BACKUP_FILE)"
    echo "✅ Upload completed!"
fi

exit 0