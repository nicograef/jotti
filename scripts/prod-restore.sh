#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — Database Restore (self-hosted production, Weg B)
#
# Restores a pg_dump created by prod-backup.sh back into the production
# database. DESTRUCTIVE: it overwrites the current data with the chosen dump
# (the dumps use --clean --if-exists, so objects are dropped and re-created).
# Application services are stopped during the restore so no writes interfere.
# Steps:
#   1. Validate prerequisites and pick the dump (argument or newest in BACKUP_DIR)
#   2. Confirm the destructive action
#   3. Stop app services, restore via psql, restart the stack
#
# Configuration:
#   BACKUP_DIR   directory to read dumps from (default: ./backups)
#
# Usage: ./scripts/prod-restore.sh [DUMP_FILE]  (or `make prod-restore`)
#   DUMP_FILE  optional path or filename in BACKUP_DIR; defaults to the newest.
# =============================================================================

COMPOSE_PROD="docker-compose.prod.yml"
PG_SERVICE="postgres"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

# read_env KEY — read a single value from .env without executing the file
# (passwords may contain shell-special characters). Returns the last match,
# trimmed of surrounding whitespace.
read_env() {
  local key="$1"
  grep -E "^${key}=" .env 2>/dev/null | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
}

# ---------------------------------------------------------------------------
# Step 0 — Change to project root (script may be called from anywhere)
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# ---------------------------------------------------------------------------
# Step 1 — Validate prerequisites and select the dump
# ---------------------------------------------------------------------------
if ! command -v docker &>/dev/null; then
  fatal "docker is not installed or not on PATH."
fi
if ! docker compose version &>/dev/null; then
  fatal "docker compose (v2) is not available."
fi
if [[ ! -f "$COMPOSE_PROD" ]]; then
  fatal "Missing compose file: $COMPOSE_PROD"
fi
if [[ ! -f .env ]]; then
  fatal ".env file not found. Run 'make init' first."
fi

BACKUP_DIR="${BACKUP_DIR:-$(read_env BACKUP_DIR)}"
[[ -n "$BACKUP_DIR" ]] || BACKUP_DIR="./backups"

mapfile -t dumps < <(find "$BACKUP_DIR" -maxdepth 1 -type f \
  \( -name 'jotti-*.sql' -o -name 'jotti-*.sql.gz' \) -printf '%f\n' 2>/dev/null | sort)

TARGET="${1:-}"
if [[ -n "$TARGET" ]]; then
  if [[ -f "$TARGET" ]]; then
    SELECTED="$TARGET"
  elif [[ -f "$BACKUP_DIR/$TARGET" ]]; then
    SELECTED="$BACKUP_DIR/$TARGET"
  else
    fatal "Backup not found: $TARGET"
  fi
else
  (( ${#dumps[@]} > 0 )) || fatal "No backups found in $BACKUP_DIR. Pass a dump file explicitly."
  SELECTED="$BACKUP_DIR/${dumps[-1]}"
fi

if (( ${#dumps[@]} > 0 )); then
  info "Available backups in $BACKUP_DIR:"
  for d in "${dumps[@]}"; do
    echo "    $d"
  done
fi

# ---------------------------------------------------------------------------
# Step 2 — Confirm the destructive action
# ---------------------------------------------------------------------------
echo ""
warn "This will OVERWRITE the current jotti database with:"
warn "  $SELECTED"
warn "All data created since that backup will be lost."
read -r -p "Continue? Type 'yes' to proceed: " answer
[[ "$answer" == "yes" ]] || fatal "Aborted by user. Nothing was changed."

# ---------------------------------------------------------------------------
# Step 3 — Stop app services, restore, restart
# ---------------------------------------------------------------------------
info "Starting the database ..."
docker compose -f "$COMPOSE_PROD" up -d --wait "$PG_SERVICE"

info "Stopping application services during the restore ..."
docker compose -f "$COMPOSE_PROD" stop backend frontend reverse-proxy

# Stream the dump (decompressing on the fly when gzip-compressed) into psql. The
# postgres role comes from the container's own POSTGRES_USER env; ON_ERROR_STOP
# aborts on the first SQL error instead of limping on with a half-restored DB.
decompress() {
  if [[ "$SELECTED" == *.gz ]]; then
    gzip -dc "$SELECTED"
  else
    cat "$SELECTED"
  fi
}

info "Restoring $SELECTED ..."
if ! decompress | docker compose -f "$COMPOSE_PROD" exec -T "$PG_SERVICE" \
       sh -c 'psql -U "$POSTGRES_USER" -d jotti -v ON_ERROR_STOP=1'; then
  error "Restore failed — the database may be in an inconsistent state."
  error "Application services are stopped. Inspect, fix, then restart: make prod-up"
  exit 1
fi

info "Restarting the full stack ..."
docker compose -f "$COMPOSE_PROD" up -d

echo ""
info "Restore complete. jotti is running again."
