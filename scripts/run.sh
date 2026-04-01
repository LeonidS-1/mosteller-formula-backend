#!/usr/bin/env bash
set -e
cd "$(dirname "$0")/.."
source .env
echo "→ Запуск приложения..."
go run ./cmd/app/
