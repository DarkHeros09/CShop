#!/bin/sh

set -e

echo "run db migration"
migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the up migration"
exec "$@"