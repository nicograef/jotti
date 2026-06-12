#!/usr/bin/env bash
set -euo pipefail

# Resets only the production database volume, then recreates the prod stack and
# seeds demo data via the `jotti seed` subcommand (guard and projection rebuild included).

COMPOSE_FILE="docker-compose.prod.yml"
DB_VOLUME="jotti_postgres-data"
PG_SERVICE="postgres"
BACKEND_SERVICE="backend"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

usage() {
  cat <<'EOF'
Usage: ./scripts/prod-reset-and-seed.sh [--yes]

Resets production DB data and reloads seed data without touching SSL volumes.

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

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

[[ -f "$COMPOSE_FILE" ]] || fatal "Missing $COMPOSE_FILE"

command -v docker >/dev/null 2>&1 || fatal "docker not found in PATH"
docker compose version >/dev/null 2>&1 || fatal "docker compose (v2) is not available"

if [[ "$ASSUME_YES" != "true" ]]; then
  echo ""
  warn "This will DELETE all production DB data in volume: $DB_VOLUME"
  warn "SSL certificate volumes (letsencrypt, certbot-challenges) are NOT touched."
  read -r -p "Continue? Type 'yes' to proceed: " answer
  [[ "$answer" == "yes" ]] || fatal "Aborted by user"
fi

info "Stopping production stack..."
docker compose -f "$COMPOSE_FILE" down

if docker volume inspect "$DB_VOLUME" >/dev/null 2>&1; then
  info "Removing database volume: $DB_VOLUME"
  docker volume rm "$DB_VOLUME"
else
  warn "Database volume $DB_VOLUME not found; continuing"
fi

info "Starting production stack..."
docker compose -f "$COMPOSE_FILE" up -d --build

info "Waiting for postgres service to become healthy..."
for _ in $(seq 1 60); do
  status="$(docker compose -f "$COMPOSE_FILE" ps --format json "$PG_SERVICE" | sed -n 's/.*"Health":"\([^"]*\)".*/\1/p')"
  if [[ "$status" == "healthy" ]]; then
    break
  fi
  sleep 2
done

status="$(docker compose -f "$COMPOSE_FILE" ps --format json "$PG_SERVICE" | sed -n 's/.*"Health":"\([^"]*\)".*/\1/p')"
[[ "$status" == "healthy" ]] || fatal "Postgres is not healthy (status: ${status:-unknown})"

info "Seeding demo data via seed subcommand (guard + projection rebuild included)..."
docker compose -f "$COMPOSE_FILE" exec -T "$BACKEND_SERVICE" jotti seed

echo ""
info "Done. Production DB reset + seed completed."
info "SSL certificates were not modified."
