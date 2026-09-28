#!/usr/bin/env bash
# test-tse-live.sh — TSE live suite against the fiskaly TEST TSS
#
# Usage:
#   make test-tse-live   # needs .env.fiskaly-test (template: .env.fiskaly-test.example)
#
# What it does:
#   1. Starts its own Postgres container on its own port, so it runs beside
#      scripts/test-integration.sh.
#   2. Applies every migration, then runs the live signing tests with the
#      credentials from .env.fiskaly-test.
# It never creates a TSS: that setup run lives only in `make test-tse-live-setup`.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

TSE_PG_CONTAINER="${TSE_PG_CONTAINER:-jotti-pg-tse-live}"
TSE_PG_PORT="${TSE_PG_PORT:-5453}"
PG_IMAGE="postgres:17.11"
DATABASE_URL="postgres://admin:admin@localhost:${TSE_PG_PORT}/jotti?sslmode=disable"

[[ -f "$PROJECT_ROOT/.env.fiskaly-test" ]] ||
  fatal ".env.fiskaly-test is missing. Template: .env.fiskaly-test.example"

cleanup() {
  info "Removing $TSE_PG_CONTAINER ..."
  docker rm -f "$TSE_PG_CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker rm -f "$TSE_PG_CONTAINER" >/dev/null 2>&1 || true

info "Starting $PG_IMAGE on port $TSE_PG_PORT ..."
docker run -d \
  --name "$TSE_PG_CONTAINER" \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=jotti \
  -p "${TSE_PG_PORT}:5432" \
  "$PG_IMAGE" >/dev/null

# pg_isready also answers for the image's temporary socket-only init server, and
# migrate then fails with "connection reset by peer"; a TCP query does not.
info "Waiting for PostgreSQL to accept real connections ..."
ready=""
for ((i = 0; i < 30; i++)); do
  if docker exec -e PGPASSWORD=admin "$TSE_PG_CONTAINER" \
    psql -h 127.0.0.1 -U admin -d jotti -c "SELECT 1" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 2
done
if [[ -z "$ready" ]]; then
  error "PostgreSQL did not become ready in time. Container logs:"
  docker logs "$TSE_PG_CONTAINER"
  exit 1
fi

info "Running database migrations (forward-only) ..."
migrate -path "$PROJECT_ROOT/database/migrations" -database "$DATABASE_URL" up

info "Running TSE live tests against the fiskaly TEST TSS ..."
cd "$PROJECT_ROOT/backend"
# Holds the FISKALY_TEST_* credentials and the admin PUK/PIN: never print them.
set -a
# shellcheck disable=SC1091
. "$PROJECT_ROOT/.env.fiskaly-test"
set +a

# -run leaves out the TSS-creating TestFiskalySetup_LiveVollerDurchlauf.
# JOTTI_TSE_LIVE=1 opts in, so plain integration runs skip even with exported credentials.
JOTTI_TSE_LIVE=1 \
POSTGRES_HOST=localhost \
POSTGRES_PORT="$TSE_PG_PORT" \
POSTGRES_USER=admin \
POSTGRES_PASSWORD=admin \
POSTGRES_DBNAME=jotti \
go test -tags=integration -count=1 -v -p 1 -run 'LiveSigniert|LiveSuite' ./api/fiskal/tse_live/ ./repository/tse_repo/

info "TSE live tests passed."
