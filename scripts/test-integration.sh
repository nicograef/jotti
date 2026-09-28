#!/usr/bin/env bash
# test-integration.sh — backend integration tests against a throwaway Postgres
#
# Usage:
#   make check-integration
#   TEST_PG_PORT=5433 ./scripts/test-integration.sh   # when 5432 is taken
#
# What it does:
#   1. Starts a fresh Postgres container and waits until a TCP query passes.
#   2. Applies every migration in database/migrations.
#   3. Runs `go test -tags=integration` serially and removes the container on exit.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

TEST_PG_CONTAINER="${TEST_PG_CONTAINER:-jotti-postgres-test}"
TEST_PG_PORT="${TEST_PG_PORT:-5432}"
PG_IMAGE="postgres:17.11"
DATABASE_URL="postgres://admin:admin@localhost:${TEST_PG_PORT}/jotti?sslmode=disable"

cleanup() {
  info "Removing $TEST_PG_CONTAINER ..."
  docker rm -f "$TEST_PG_CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker rm -f "$TEST_PG_CONTAINER" >/dev/null 2>&1 || true

info "Starting $PG_IMAGE on port $TEST_PG_PORT ..."
docker run -d \
  --name "$TEST_PG_CONTAINER" \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=jotti \
  -p "${TEST_PG_PORT}:5432" \
  "$PG_IMAGE" >/dev/null

# pg_isready also answers for the image's temporary socket-only init server, and
# migrate then fails with "connection reset by peer"; a TCP query does not.
info "Waiting for PostgreSQL to accept real connections ..."
ready=""
for ((i = 0; i < 30; i++)); do
  if docker exec -e PGPASSWORD=admin "$TEST_PG_CONTAINER" \
    psql -h 127.0.0.1 -U admin -d jotti -c "SELECT 1" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 2
done
if [[ -z "$ready" ]]; then
  error "PostgreSQL did not become ready in time. Container logs:"
  docker logs "$TEST_PG_CONTAINER"
  exit 1
fi

info "Running database migrations (forward-only) ..."
migrate -path "$PROJECT_ROOT/database/migrations" -database "$DATABASE_URL" up

# -p 1: every package shares this one database, so parallel packages would
# pollute each other's data.
info "Running integration tests ..."
cd "$PROJECT_ROOT/backend"
POSTGRES_HOST=localhost \
POSTGRES_PORT="$TEST_PG_PORT" \
POSTGRES_USER=admin \
POSTGRES_PASSWORD=admin \
POSTGRES_DBNAME=jotti \
JWT_SECRET=test-secret \
RELAY_AUTH_TOKEN=test-relay-token \
go test -tags=integration -count=1 -race -p 1 ./...

info "Integration tests passed."
