#!/usr/bin/env bash
set -e
cd "$(dirname "$0")/.."
source .env

echo "→ Миграция DOWN (init-down.sql)..."

PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f migrations/init-down.sql

echo "→ Миграция DOWN выполнена."
