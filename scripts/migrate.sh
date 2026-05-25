#!/usr/bin/env sh
# Thin wrapper around golang-migrate so CI/Compose can call a stable script
# even if we swap the underlying tool later.
set -e

DIR="${MIGRATIONS_DIR:-./migrations}"
URL="${DB_URL:-postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable}"
CMD="${1:-up}"

echo "running migrations: $CMD"
migrate -database "$URL" -path "$DIR" "$CMD"
