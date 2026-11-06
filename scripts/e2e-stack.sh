#!/usr/bin/env bash
# e2e-stack.sh — runs a command against a throwaway docker-compose.e2e.yml stack
#
# Usage:
#   ./scripts/e2e-stack.sh make test-e2e
#   E2E_STACK_PORT=8081 E2E_STACK_PROJECT=jotti-e2e-p1 ./scripts/e2e-stack.sh make test-e2e
#
# Environment:
#   E2E_STACK_PORT      host port of the reverse proxy, default 8080
#   E2E_STACK_PROJECT   compose project name, default jotti-e2e; parallel runs need their own
#
# What it does:
#   1. Builds and starts the stack and waits until the proxy answers HTTP 200.
#   2. Runs the command with E2E_BASE_URL pointing at the stack.
#   3. Removes the stack with its volumes on exit; the exit code is the command's.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

if (($# == 0)); then
  fatal "Usage: $0 <command> [args ...]"
fi

E2E_STACK_PORT="${E2E_STACK_PORT:-8080}"
E2E_STACK_PROJECT="${E2E_STACK_PROJECT:-jotti-e2e}"
COMPOSE=(docker compose -p "$E2E_STACK_PROJECT" -f "$PROJECT_ROOT/docker-compose.e2e.yml")
BASE_URL="http://localhost:${E2E_STACK_PORT}"

cleanup() {
  info "Removing stack $E2E_STACK_PROJECT ..."
  "${COMPOSE[@]}" down -v >/dev/null 2>&1 || true
}
trap cleanup EXIT

info "Starting stack $E2E_STACK_PROJECT on port $E2E_STACK_PORT ..."
E2E_HTTP_PORT="$E2E_STACK_PORT" "${COMPOSE[@]}" up -d --build

ready=""
code=""
for i in $(seq 1 60); do
  code="$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/" || true)"
  if [[ "$code" == "200" ]]; then
    ready=1
    break
  fi
  info "Waiting for the stack ($i/60), status ${code:-none} ..."
  sleep 2
done
if [[ -z "$ready" ]]; then
  error "Stack did not answer HTTP 200 in time; container status and logs follow."
  "${COMPOSE[@]}" ps
  "${COMPOSE[@]}" logs
  exit 1
fi

info "Stack is ready at $BASE_URL."
# Last command: its exit code is the script's, and the EXIT trap keeps it.
E2E_BASE_URL="$BASE_URL" "$@"
