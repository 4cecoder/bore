#!/bin/bash

# Bore Server Restore Script
# This script restores backups of configuration, certificates, and logs

set -e

BACKUP_DIR="/opt/bore/backups"
BACKUP_NAME="$1"

if [ -z "$BACKUP_NAME" ]; then
    echo "Usage: $0 <backup_name>"
    echo "Available backups:"
    ls -la "$BACKUP_DIR"/*.tar.gz 2>/dev/null || echo "No backups found"
    exit 1
fi

BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"

if [ ! -f "$BACKUP_PATH" ]; then
    echo "Backup file not found: $BACKUP_PATH"
    exit 1
fi

echo "Starting restore from: $BACKUP_NAME"

# Create temporary directory for extraction
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Extract backup
echo "Extracting backup..."
tar -xzf "$BACKUP_PATH" -C "$TEMP_DIR"

# Stop bore service
echo "Stopping bore service..."
systemctl stop bore || true

# Restore configuration files
if [ -d "$TEMP_DIR/config" ]; then
    echo "Restoring configuration files..."
    cp -r "$TEMP_DIR/config" ./
fi

# Restore certificates
if [ -d "$TEMP_DIR/certs" ]; then
    echo "Restoring certificates..."
    cp -r "$TEMP_DIR/certs" ./
fi

# Restore logs
if [ -d "$TEMP_DIR/logs" ]; then
    echo "Restoring logs..."
    mkdir -p /var/log/bore
    cp "$TEMP_DIR/logs/server.log" /var/log/bore/ 2>/dev/null || true
fi

# Restore database (when implemented)
# if [ -f "$TEMP_DIR/db/bore_db.sql" ]; then
#     echo "Restoring database..."
#     psql -h "$DB_HOST" -U "$DB_USER" "$DB_NAME" < "$TEMP_DIR/db/bore_db.sql"
# fi

# Start bore service
echo "Starting bore service..."
systemctl start bore

echo "Restore completed successfully"