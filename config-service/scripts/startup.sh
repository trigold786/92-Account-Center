#!/bin/bash
# ============================================================
# Config Service Startup Script
# ============================================================
# This script runs database migrations before starting
# the config-service application.
#
# Usage:
#   ./scripts/startup.sh
#
# Environment variables:
#   DB_HOST        - Database host (default: localhost)
#   DB_PORT        - Database port (default: 5432)
#   DB_USER        - Database user (default: postgres)
#   DB_PASSWORD    - Database password (default: postgres)
#   DB_NAME        - Database name (default: account_center)
#   PORT           - Service port (default: 30315)
# ============================================================

set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-account_center}"
PORT="${PORT:-30315}"

MIGRATIONS_DIR="$(cd "$(dirname "$0")/../db-migrations" && pwd)"

echo "=== Config Service Startup ==="
echo "Database: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "Migrations: ${MIGRATIONS_DIR}"
echo "Port: ${PORT}"

# Run goose migrations
echo "==> Running database migrations..."
goose -dir "${MIGRATIONS_DIR}" \
  postgres "host=${DB_HOST} port=${DB_PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=disable" \
  up

MIGRATION_EXIT=$?
if [ ${MIGRATION_EXIT} -ne 0 ]; then
  echo "ERROR: Migration failed with exit code ${MIGRATION_EXIT}"
  exit 1
fi

echo "==> Migrations completed successfully"

# Start the service
echo "==> Starting config service on port ${PORT}..."
exec /app/config-service
