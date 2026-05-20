#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-account_center}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql.gz"
WAL_ARCHIVE_DIR="${BACKUP_DIR}/wal_archive"

mkdir -p "${BACKUP_DIR}" "${WAL_ARCHIVE_DIR}"

echo "[$(date)] Starting PostgreSQL backup for ${DB_NAME}..."

PGPASSWORD="${DB_PASSWORD:-postgres}" pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --format=custom \
    --compress=6 \
    --verbose \
    -f "${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.dump"

echo "[$(date)] Full backup completed: ${DB_NAME}_${TIMESTAMP}.dump"

PGPASSWORD="${DB_PASSWORD:-postgres}" pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --format=plain \
    --no-owner \
    --no-privileges" \
    | gzip > "${BACKUP_FILE}"

echo "[$(date)] SQL backup completed: ${BACKUP_FILE}"

find "${BACKUP_DIR}" -name "${DB_NAME}_*.dump" -mtime +${RETENTION_DAYS} -delete
find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql.gz" -mtime +${RETENTION_DAYS} -delete
find "${WAL_ARCHIVE_DIR}" -name "*.wal.gz" -mtime +${RETENTION_DAYS} -delete

echo "[$(date)] Backup cleanup completed (retention: ${RETENTION_DAYS} days)"
echo "[$(date)] Disk usage: $(du -sh ${BACKUP_DIR} | cut -f1)"
