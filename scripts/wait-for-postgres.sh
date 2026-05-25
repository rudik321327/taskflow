#!/usr/bin/env sh
# Blocks until Postgres accepts connections — used by Compose entrypoints
# so migrations and the app don't race the database.
set -e

host="${1:-postgres}"
port="${2:-5432}"
user="${DB_USER:-taskflow}"

echo "waiting for postgres at $host:$port..."
until pg_isready -h "$host" -p "$port" -U "$user" >/dev/null 2>&1; do
  sleep 1
done
echo "postgres is ready"
