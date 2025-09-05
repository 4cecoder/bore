#!/bin/bash

# Bore Server Backup Script
# This script creates backups of configuration, certificates, and logs

set -e

BACKUP_DIR="/opt/bore/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="bore_backup_${TIMESTAMP}"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"

# Create backup directory
mkdir -p "$BACKUP_PATH"

echo "Starting bore backup: $BACKUP_NAME"

# Backup configuration files
if [ -d "config" ]; then
    echo "Backing up configuration files..."
    cp -r config "$BACKUP_PATH/"
fi

# Backup certificates
if [ -d "certs" ]; then
    echo "Backing up certificates..."
    cp -r certs "$BACKUP_PATH/"
fi

# Backup logs (if they exist)
if [ -f "/var/log/bore/server.log" ]; then
    echo "Backing up logs..."
    mkdir -p "$BACKUP_PATH/logs"
    cp /var/log/bore/server.log "$BACKUP_PATH/logs/"
fi

# Backup database (when implemented)
# if [ -n "$DB_HOST" ]; then
#     echo "Backing up database..."
#     mkdir -p "$BACKUP_PATH/db"
#     pg_dump -h "$DB_HOST" -U "$DB_USER" "$DB_NAME" > "$BACKUP_PATH/db/bore_db.sql"
# fi

# Create archive
echo "Creating backup archive..."
cd "$BACKUP_DIR"
tar -czf "${BACKUP_NAME}.tar.gz" "$BACKUP_NAME"

# Remove uncompressed backup
rm -rf "$BACKUP_NAME"

# Clean up old backups (keep last 30 days)
echo "Cleaning up old backups..."
find "$BACKUP_DIR" -name "bore_backup_*.tar.gz" -mtime +30 -delete

echo "Backup completed: ${BACKUP_PATH}.tar.gz"
echo "Backup size: $(du -h "${BACKUP_PATH}.tar.gz" | cut -f1)"