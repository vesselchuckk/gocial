#!/bin/sh

set -eu

DB_HOST="${DB_HOST}"
DB_PORT="${DB_PORT}"
DB_USER="${DB_USER}"
DB_PASSWORD="${DB_PASSWORD}"
DB_NAME="${DB_NAME}"

DB_URL="${DB_URL:-postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable}"

#
# Wait for Redis
#
if [ "${REDIS_ENABLED:-false}" = "true" ] || [ -n "${REDIS_ADDR:-}" ]; then
    REDIS_ADDR="${REDIS_ADDR:-redis:6379}"
    REDIS_HOST="${REDIS_ADDR%%:*}"
    REDIS_PORT="${REDIS_ADDR##*:}"

    echo "Waiting for Redis at ${REDIS_HOST}:${REDIS_PORT}..."

    until redis-cli \
        -h "${REDIS_HOST}" \
        -p "${REDIS_PORT}" \
        ${REDIS_PASSWORD:+-a "$REDIS_PASSWORD"} \
        ping >/dev/null 2>&1
    do
        sleep 1
    done

    echo "Redis is ready."
fi

#
# Wait for PostgreSQL
#
echo "Waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}..."

until pg_isready \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" >/dev/null 2>&1
do
    sleep 1
done

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