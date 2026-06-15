#!/usr/bin/env bash
set -euo pipefail

# Resets a single stack's database volume, then recreates the stack and seeds
# demo data via the `jotti seed` subcommand (guard and projection rebuild included).
#
# Supported stacks:
#   local  — LAN stack (docker-compose.local.yml): the smallest setup, a single
#            device with HTTPS via Caddy. The Caddy data volume (caddy-data,
#            certificates + internal CA) is NOT touched, so no browser re-trust is
#            needed after a reset.
#   rocks  — jotti.rocks demo/staging (docker-compose.prod.yml + docker-compose.rocks.yml).
#            NOT for self-hosted production. The SSL volumes (letsencrypt,
#            certbot-challenges) are NOT touched.

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
Usage: ./scripts/reset-and-seed.sh <local|rocks> [--yes]

Resets one stack's DB data and reloads seed data without touching its TLS/SSL volumes.

Stacks:
  local    LAN stack (docker-compose.local.yml); Caddy cert volume preserved
  rocks    jotti.rocks demo/staging (prod + rocks override); SSL volumes preserved

Options:
  --yes    Skip interactive confirmation prompt
  -h       Show this help
  --help   Show this help
EOF
}

STACK=""
ASSUME_YES="false"
for arg in "$@"; do
  case "$arg" in
    local|rocks)
      [[ -z "$STACK" ]] || fatal "Stack already set to '$STACK'; unexpected argument: $arg"
      STACK="$arg"
      ;;
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

[[ -n "$STACK" ]] || { usage; fatal "Missing required stack argument (local|rocks)"; }

case "$STACK" in
  local)
    COMPOSE_FILES=(-f docker-compose.local.yml)
    DB_VOLUME="jotti-local_postgres-data"
    STACK_LABEL="local (LAN)"
    TLS_NOTE="The Caddy data volume (caddy-data, TLS certificates) is NOT touched."
    ;;
  rocks)
    COMPOSE_FILES=(-f docker-compose.prod.yml -f docker-compose.rocks.yml)
    DB_VOLUME="jotti_postgres-data"
    STACK_LABEL="jotti.rocks demo"
    TLS_NOTE="SSL certificate volumes (letsencrypt, certbot-challenges) are NOT touched."
    ;;
esac

PG_SERVICE="postgres"
BACKEND_SERVICE="backend"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# Every "-f <file>" entry in COMPOSE_FILES must exist.
for i in "${!COMPOSE_FILES[@]}"; do
  [[ "${COMPOSE_FILES[$i]}" == "-f" ]] || continue
  compose_file="${COMPOSE_FILES[$((i + 1))]}"
  [[ -f "$compose_file" ]] || fatal "Missing compose file: $compose_file"
done

command -v docker >/dev/null 2>&1 || fatal "docker not found in PATH"
docker compose version >/dev/null 2>&1 || fatal "docker compose (v2) is not available"

if [[ "$ASSUME_YES" != "true" ]]; then
  echo ""
  warn "This will DELETE all $STACK_LABEL DB data in volume: $DB_VOLUME"
  warn "$TLS_NOTE"
  read -r -p "Continue? Type 'yes' to proceed: " answer
  [[ "$answer" == "yes" ]] || fatal "Aborted by user"
fi

info "Stopping $STACK_LABEL stack..."
docker compose "${COMPOSE_FILES[@]}" down

if docker volume inspect "$DB_VOLUME" >/dev/null 2>&1; then
  info "Removing database volume: $DB_VOLUME"
  docker volume rm "$DB_VOLUME"
else
  warn "Database volume $DB_VOLUME not found; continuing"
fi

info "Starting $STACK_LABEL stack..."
docker compose "${COMPOSE_FILES[@]}" up -d --build

info "Waiting for postgres service to become healthy..."
status=""
for _ in $(seq 1 60); do
  status="$(docker compose "${COMPOSE_FILES[@]}" ps --format json "$PG_SERVICE" | sed -n 's/.*"Health":"\([^"]*\)".*/\1/p')"
  if [[ "$status" == "healthy" ]]; then
    break
  fi
  sleep 2
done

[[ "$status" == "healthy" ]] || fatal "Postgres is not healthy (status: ${status:-unknown})"

info "Seeding demo data via seed subcommand (guard + projection rebuild included)..."
docker compose "${COMPOSE_FILES[@]}" exec -T "$BACKEND_SERVICE" jotti seed

echo ""
info "Done. $STACK_LABEL DB reset + seed completed."
info "$TLS_NOTE"
