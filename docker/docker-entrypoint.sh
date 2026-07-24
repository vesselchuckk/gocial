#!/bin/sh

set -eu

: "${DB_HOST:?DB_HOST is required}"
: "${DB_PORT:?DB_PORT is required}"
: "${DB_USER:?DB_USER is required}"
: "${DB_PASSWORD:?DB_PASSWORD is required}"
: "${DB_NAME:?DB_NAME is required}"

DB_URL="${DB_URL:-postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable}"

WAIT_TIMEOUT="${WAIT_TIMEOUT:-60}"

wait_for() {
    local name="$1"
    shift
    elapsed=0

    until "$@" >/dev/null 2>&1; do
        elapsed=$((elapsed + 1))
        if [ $elapsed -ge $WAIT_TIMEOUT ]; then
            echo "Timeout after ${WAIT_TIMEOUT} seconds waiting for ${name}"
            exit 1
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
}

#
# Wait for Redis
#
if [ "${REDIS_ENABLED:-false}" = "true" ] || [ -n "${REDIS_ADDR:-}" ]; then
    REDIS_ADDR="${REDIS_ADDR:-redis:6379}"
    REDIS_HOST="${REDIS_ADDR%%:*}"
    REDIS_PORT="${REDIS_ADDR##*:}"

    echo "Waiting for Redis at ${REDIS_HOST}:${REDIS_PORT}..."

    export REDISCLI_AUTH="${REDIS_PASSWORD:-}"

    wait_for "Redis" redis-cli -h "${REDIS_HOST}" -p "${REDIS_PORT}" ping

    unset REDISCLI_AUTH

    echo "Redis is ready."
fi

#
# Wait for PostgreSQL
#
echo "Waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}..."

wait_for "PostgreSQL" pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}"

echo "PostgreSQL is ready."

#
# Apply migrations
#
MIGRATION_DIR="/app/migrations"

if [ -d "${MIGRATION_DIR}" ]; then
    echo "Applying migrations..."

    migrate \
      -path "${MIGRATION_DIR}" \
      -database "${DB_URL}" \
      up
      
    echo "Migrations completed."
else
    echo "Migration directory not found. Skipping."
fi

#
# Start application
#
echo "Starting API..."

exec /app/api