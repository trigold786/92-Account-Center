#!/usr/bin/env bash
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/backups/redis}"
REDIS_HOST="${REDIS_HOST:-redis}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
RETENTION_DAYS="${RETENTION_DAYS:-30}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "${BACKUP_DIR}"

echo "[$(date)] Starting Redis backup..."

REDISCLI_CMD="redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT}"
if [ -n "${REDIS_PASSWORD}" ]; then
    REDISCLI_CMD="${REDISCLI_CMD} -a ${REDIS_PASSWORD}"
fi

${REDISCLI_CMD} CONFIG SET save "900 1 300 10 60 10000"
echo "[$(date)] RDB save policy configured"

${REDISCLI_CMD} CONFIG SET appendonly yes
${REDISCLI_CMD} CONFIG SET appendfsync everysec
echo "[$(date)] AOF persistence enabled (everysec sync)"

${REDISCLI_CMD} BGSAVE
echo "[$(date)] BGSAVE triggered"

TIMEOUT=60
ELAPSED=0
while [ ${ELAPSED} -lt ${TIMEOUT} ]; do
    LASTSAVE=$(${REDISCLI_CMD} LASTSAVE)
    sleep 2
    NEWSAVE=$(${REDISCLI_CMD} LASTSAVE)
    if [ "${LASTSAVE}" != "${NEWSAVE}" ]; then
        echo "[$(date)] BGSAVE completed"
        break
    fi
    ELAPSED=$((ELAPSED + 2))
done

if [ ${ELAPSED} -ge ${TIMEOUT} ]; then
    echo "[$(date)] WARNING: BGSAVE did not complete within ${TIMEOUT}s"
fi

docker cp redis:/data/dump.rdb "${BACKUP_DIR}/dump_${TIMESTAMP}.rdb" 2>/dev/null || \
    cp /data/dump.rdb "${BACKUP_DIR}/dump_${TIMESTAMP}.rdb" 2>/dev/null || \
    echo "[$(date)] WARNING: Could not copy RDB file directly"

if [ -f "${BACKUP_DIR}/dump_${TIMESTAMP}.rdb" ]; then
    gzip "${BACKUP_DIR}/dump_${TIMESTAMP}.rdb"
    echo "[$(date)] RDB backup saved: dump_${TIMESTAMP}.rdb.gz"
fi

find "${BACKUP_DIR}" -name "dump_*.rdb.gz" -mtime +${RETENTION_DAYS} -delete

echo "[$(date)] Redis backup completed"
echo "[$(date)] Disk usage: $(du -sh ${BACKUP_DIR} | cut -f1)"
