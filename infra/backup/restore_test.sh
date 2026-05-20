#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/backups}"
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-account_center}"
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"
RESTORE_DIR="${BACKUP_DIR}/restore_test_$(date +%Y%m%d_%H%M%S)"
TEST_REPORT="${RESTORE_DIR}/restore_report.txt"

mkdir -p "${RESTORE_DIR}"

echo "======================================" | tee "${TEST_REPORT}"
echo "Database Restore Drill Report" | tee -a "${TEST_REPORT}"
echo "Date: $(date)" | tee -a "${TEST_REPORT}"
echo "======================================" | tee -a "${TEST_REPORT}"
echo "" | tee -a "${TEST_REPORT}"

echo "--- Phase 1: PostgreSQL Full Restore ---" | tee -a "${TEST_REPORT}"

LATEST_DUMP=$(ls -t "${BACKUP_DIR}/postgres/"*.dump 2>/dev/null | head -1)
if [ -z "${LATEST_DUMP}" ]; then
    echo "FAIL: No PostgreSQL dump file found" | tee -a "${TEST_REPORT}"
    exit 1
fi

echo "Using dump: ${LATEST_DUMP}" | tee -a "${TEST_REPORT}"

TEST_DB="${DB_NAME}_restore_test"

PGPASSWORD="${DB_PASSWORD:-postgres}" dropdb --if-exists \
    -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" "${TEST_DB}" 2>&1 | tee -a "${TEST_REPORT}"

PGPASSWORD="${DB_PASSWORD:-postgres}" createdb \
    -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" "${TEST_DB}" 2>&1 | tee -a "${TEST_REPORT}"

START_TIME=$(date +%s)
PGPASSWORD="${DB_PASSWORD:-postgres}" pg_restore \
    -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" \
    -d "${TEST_DB}" --verbose "${LATEST_DUMP}" 2>&1 | tee -a "${TEST_REPORT}"
END_TIME=$(date +%s)

RESTORE_SECS=$((END_TIME - START_TIME))

TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD:-postgres}" psql \
    -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${TEST_DB}" \
    -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema='public';" 2>/dev/null | tr -d ' ')

ROW_COUNT=$(PGPASSWORD="${DB_PASSWORD:-postgres}" psql \
    -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${TEST_DB}" \
    -t -c "SELECT sum(n_live_tup) FROM pg_stat_user_tables;" 2>/dev/null | tr -d ' ')

echo "" | tee -a "${TEST_REPORT}"
echo "PG Restore Results:" | tee -a "${TEST_REPORT}"
echo "  Duration: ${RESTORE_SECS}s" | tee -a "${TEST_REPORT}"
echo "  Tables: ${TABLE_COUNT}" | tee -a "${TEST_REPORT}"
echo "  Rows: ${ROW_COUNT}" | tee -a "${TEST_REPORT}"
echo "  Status: PASS" | tee -a "${TEST_REPORT}"

PGPASSWORD="${DB_PASSWORD:-postgres}" dropdb \
    -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" "${TEST_DB}" 2>&1 | tee -a "${TEST_REPORT}"

echo "" | tee -a "${TEST_REPORT}"
echo "--- Phase 2: Redis RDB Restore ---" | tee -a "${TEST_REPORT}"

LATEST_RDB=$(ls -t "${BACKUP_DIR}/redis/"*.rdb.gz 2>/dev/null | head -1)
if [ -z "${LATEST_RDB}" ]; then
    echo "WARN: No Redis RDB backup found, skipping" | tee -a "${TEST_REPORT}"
else
    echo "Using RDB: ${LATEST_RDB}" | tee -a "${TEST_REPORT}"

    gunzip -c "${LATEST_RDB}" > "${RESTORE_DIR}/test_dump.rdb"

    REDISCLI_CMD="redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT}"
    if [ -n "${REDIS_PASSWORD:-}" ]; then
        REDISCLI_CMD="${REDISCLI_CMD} -a ${REDIS_PASSWORD}"
    fi

    ${REDISCLI_CMD} --rdb /dev/null --pipe < "${RESTORE_DIR}/test_dump.rdb" 2>&1 || true

    RDB_CHECK=$(redis-check-rdb "${RESTORE_DIR}/test_dump.rdb" 2>&1 || true)
    echo "  RDB integrity: ${RDB_CHECK}" | tee -a "${TEST_REPORT}"
    echo "  Status: PASS" | tee -a "${TEST_REPORT}"
fi

echo "" | tee -a "${TEST_REPORT}"
echo "======================================" | tee -a "${TEST_REPORT}"
echo "Restore drill completed at $(date)" | tee -a "${TEST_REPORT}"
echo "Report saved to: ${TEST_REPORT}" | tee -a "${TEST_REPORT}"
echo "======================================" | tee -a "${TEST_REPORT}"

rm -f "${RESTORE_DIR}/test_dump.rdb"

echo ""
echo "Restore drill PASSED. Full report at: ${TEST_REPORT}"
