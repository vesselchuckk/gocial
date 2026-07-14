#!/bin/sh
set -eu

DB_HOST="${DB_HOST:-db}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-admin}"
DB_PASSWORD="${DB_PASSWORD:-qwerty}"
DB_NAME="${DB_NAME:-gosocial}"
DB_URL="${DB_URL:-postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable}"

if [ "${REDIS_ENABLED:-false}" = "true" ] || [ -n "${REDIS_ADDR:-}" ]; then
  REDIS_ADDR="${REDIS_ADDR:-redis:6379}"
  REDIS_HOST="${REDIS_ADDR%%:*}"
  REDIS_PORT="${REDIS_ADDR##*:}"

  echo "Waiting for Redis at ${REDIS_HOST}:${REDIS_PORT}..."
  until redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" ping >/dev/null 2>&1; do
    sleep 1
  done
fi

echo "Waiting for Postgres at ${DB_HOST}:${DB_PORT}..."
until pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" >/dev/null 2>&1; do
  sleep 1
done

MIGRATION_DIR="/app/migrations"
if [ -d "${MIGRATION_DIR}" ]; then
  echo "Applying SQL migrations from ${MIGRATION_DIR}"
  for sql_file in $(find "${MIGRATION_DIR}" -maxdepth 1 -name '*.up.sql' | sort); do
    echo "Applying ${sql_file}"
    PGPASSWORD="${DB_PASSWORD}" psql "${DB_URL}" -v ON_ERROR_STOP=1 -f "${sql_file}"
  done
else
  echo "No migration directory found at ${MIGRATION_DIR}; continuing without migration step"
fi

exec /app/api
