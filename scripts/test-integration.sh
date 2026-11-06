#!/usr/bin/env bash
# test-integration.sh — backend integration tests against a throwaway Postgres
#
# Usage:
#   make check-integration
#   TEST_PG_PORT=5433 ./scripts/test-integration.sh   # when 5432 is taken
#   TEST_PKG='./repository/produkt_repo/' TEST_RUN='^TestCreateProdukt$' make check-integration
#
# Environment:
#   TEST_PG_CONTAINER, TEST_PG_PORT   container name and host port; parallel runs need their own
#   TEST_PKG                          space-separated packages relative to backend/, default ./...
#   TEST_RUN                          go test -run pattern; also sets -v
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
TEST_PKG="${TEST_PKG:-./...}"
TEST_RUN="${TEST_RUN:-}"

cleanup() {
  info "Removing $TEST_PG_CONTAINER ..."
  docker rm -f "$TEST_PG_CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

start_throwaway_postgres "$TEST_PG_CONTAINER" "$TEST_PG_PORT"

# -p 1: every package shares this one database, so parallel packages would
# pollute each other's data.
info "Running integration tests ..."
cd "$PROJECT_ROOT/backend"
read -r -a test_pkgs <<<"$TEST_PKG"
run_args=()
if [[ -n "$TEST_RUN" ]]; then
  run_args=(-run "$TEST_RUN" -v)
fi
POSTGRES_HOST=localhost \
POSTGRES_PORT="$TEST_PG_PORT" \
POSTGRES_USER=admin \
POSTGRES_PASSWORD=admin \
POSTGRES_DBNAME=jotti \
JWT_SECRET=test-secret \
RELAY_AUTH_TOKEN=test-relay-token \
go test -tags=integration -count=1 -race -p 1 "${run_args[@]}" "${test_pkgs[@]}"

info "Integration tests passed."
