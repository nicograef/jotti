#!/usr/bin/env bash
# reset-and-seed.sh — resets the jotti.rocks demo database and re-seeds it
#
# Usage:
#   make rocks-reset-and-seed   # or: ./scripts/reset-and-seed.sh [--yes]
#
# Deletes the demo stack's database volume, restarts the stack and runs
# `jotti seed`; the certificate volume stays. NOT for self-hosted production.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

usage() {
  cat <<'EOF'
Usage: ./scripts/reset-and-seed.sh [--yes]

Resets the jotti.rocks demo stack's DB data (docker-compose.rocks.yml) and
reloads seed data; the certificate volume stays.

Options:
  --yes    Skip interactive confirmation prompt
  -h       Show this help
  --help   Show this help
EOF
}

ASSUME_YES="false"
for arg in "$@"; do
  case "$arg" in
    --yes)
      ASSUME_YES="true"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fatal "Unknown option: $arg"
      ;;
  esac
done

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.rocks.yml}"
DB_VOLUME="${DB_VOLUME:-jotti_postgres-data}"
BACKEND_SERVICE="${BACKEND_SERVICE:-backend}"

PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

require_docker_stack "$COMPOSE_FILE"
require_rocks_stack

if [[ "$ASSUME_YES" != "true" ]]; then
  echo ""
  warn "This will DELETE all jotti.rocks demo DB data in volume: $DB_VOLUME"
  warn "The certificate volume (caddy-data) is NOT touched."
  read -r -p "Continue? Type 'yes' to proceed: " answer
  [[ "$answer" == "yes" ]] || fatal "Aborted by user"
fi

info "Stopping jotti.rocks demo stack..."
docker compose -f "$COMPOSE_FILE" down

if docker volume inspect "$DB_VOLUME" >/dev/null 2>&1; then
  info "Removing database volume: $DB_VOLUME"
  docker volume rm "$DB_VOLUME"
else
  warn "Database volume $DB_VOLUME not found; continuing"
fi

# The backend waits for a healthy postgres and the finished migrate service, so
# `up -d` returns only once the seed can run.
info "Starting jotti.rocks demo stack..."
docker compose -f "$COMPOSE_FILE" up -d --build

info "Seeding demo data via seed subcommand (guard + projection rebuild included)..."
docker compose -f "$COMPOSE_FILE" exec -T "$BACKEND_SERVICE" jotti seed

echo ""
info "Done. jotti.rocks demo DB reset + seed completed."
info "The certificate volume (caddy-data) is NOT touched."
