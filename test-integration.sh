#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTAINER_NAME="jotti-postgres-test"
DATABASE_URL="postgres://admin:admin@localhost:5432/jotti?sslmode=disable"

cleanup() {
  echo ""
  echo "🧹 Cleaning up..."

  if command -v migrate >/dev/null 2>&1; then
    migrate -path "$ROOT_DIR/database/migrations" -database "$DATABASE_URL" down -all >/dev/null 2>&1 || true
  fi

  docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
  docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
}

trap cleanup EXIT

echo "🧪 Starting integration test environment..."

docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

echo "🐘 Starting PostgreSQL..."
docker run -d \
  --name "$CONTAINER_NAME" \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=jotti \
  -p 5432:5432 \
  --health-cmd "pg_isready -U admin -d jotti" \
  --health-interval 2s \
  --health-timeout 5s \
  --health-retries 10 \
  postgres:17

echo "⏳ Waiting for PostgreSQL to be ready..."
until docker exec "$CONTAINER_NAME" pg_isready -U admin -d jotti >/dev/null 2>&1; do
  sleep 2
done

echo "✅ PostgreSQL ready!"

echo "🔄 Running database migrations..."
migrate -path "$ROOT_DIR/database/migrations" -database "$DATABASE_URL" up

echo "🔄 Verifying migration roundtrip (down -all, up)..."
migrate -path "$ROOT_DIR/database/migrations" -database "$DATABASE_URL" down -all
migrate -path "$ROOT_DIR/database/migrations" -database "$DATABASE_URL" up

echo "✅ Migrations complete!"
echo ""

echo "🏃 Running integration tests..."
# -p 1 serializes package test binaries: all integration tests share a single
# Postgres instance, so running packages in parallel pollutes each other's data.
cd "$ROOT_DIR/backend"
POSTGRES_HOST=localhost \
POSTGRES_PORT=5432 \
POSTGRES_USER=admin \
POSTGRES_PASSWORD=admin \
POSTGRES_DBNAME=jotti \
JWT_SECRET=test-secret \
RELAY_AUTH_TOKEN=test-relay-token \
go test -tags=integration -count=1 -race -p 1 ./...

echo "✅ Integration tests passed!"

