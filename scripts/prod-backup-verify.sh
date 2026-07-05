#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — Backup Verify (self-hosted production, Weg B)
#
# Proves that a pg_dump created by prod-backup.sh is actually recoverable by
# restoring it into a THROWAWAY postgres container and checking that the
# restored database has tables. It never touches the running stack: the
# container runs via `docker run --rm` on the default bridge (no stack network,
# no stack volumes) and is discarded on exit. Steps:
#   1. Validate prerequisites and pick the dump (argument or newest in BACKUP_DIR)
#   2. Start a throwaway postgres (same pinned version as the stack) and wait
#   3. Stream the dump into psql (ON_ERROR_STOP) and count the restored tables
#   4. Print a short summary (dump, table count, result) and exit accordingly
#
# Configuration:
#   BACKUP_DIR    directory to read dumps from (default: ./backups)
#   COMPOSE_FILE  compose file the postgres version is read from
#                 (default: docker-compose.prod.yml)
#
# Usage: ./scripts/prod-backup-verify.sh [DUMP_FILE]  (or `make prod-backup-verify`)
#   DUMP_FILE  optional path or filename in BACKUP_DIR; defaults to the newest.
# =============================================================================

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"

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
  { grep -E "^${key}=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
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
if [[ ! -f "$COMPOSE_FILE" ]]; then
  fatal "Missing compose file: $COMPOSE_FILE"
fi
if [[ ! -f .env ]]; then
  fatal ".env file not found. Run 'make init' first."
fi

# The throwaway postgres uses the same role the dump was created with, so its
# ownership statements (ALTER ... OWNER TO) resolve. Env wins, then .env, then
# the shipped default.
PG_USER="${POSTGRES_USER:-$(read_env POSTGRES_USER)}"
[[ -n "$PG_USER" ]] || PG_USER="admin"

# Pin the throwaway container to the exact postgres version of the stack so the
# verify never drifts from what actually holds the data.
PG_IMAGE="$(grep -oE 'postgres:[0-9][0-9.]*' "$COMPOSE_FILE" | head -n1)"
[[ -n "$PG_IMAGE" ]] || fatal "Could not read the postgres version from $COMPOSE_FILE."

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

# ---------------------------------------------------------------------------
# Step 2 — Start a throwaway postgres (isolated from the running stack)
# ---------------------------------------------------------------------------
# --rm plus no --network and no -p: the container shares no network with the
# stack and publishes no port. -fv on removal takes its anonymous data volume
# (the image declares one) with it, so `docker volume ls` stays unchanged.
# Nothing about the live stack is touched.
CONTAINER="jotti-backup-verify-$$"
cleanup() { docker rm -fv "$CONTAINER" &>/dev/null || true; }
trap cleanup EXIT

info "Starting throwaway $PG_IMAGE for the verify ..."
docker run --rm -d --name "$CONTAINER" \
  -e POSTGRES_USER="$PG_USER" \
  -e POSTGRES_PASSWORD=verify \
  -e POSTGRES_DB=jotti \
  "$PG_IMAGE" >/dev/null

# Probe over TCP, not the socket: the entrypoint's temporary init server listens
# on the socket only, so a socket probe would report "ready" mid-initialisation
# and the restore would race the real server's restart. TCP answers only once
# the real server is up. Restore then uses the socket (trust-authenticated).
ready=""
for ((i = 0; i < 30; i++)); do
  if docker exec "$CONTAINER" pg_isready -h 127.0.0.1 -p 5432 -U "$PG_USER" &>/dev/null; then
    ready=1
    break
  fi
  sleep 1
done
[[ -n "$ready" ]] || fatal "Throwaway postgres did not become ready in time."

# ---------------------------------------------------------------------------
# Step 3 — Restore the dump and count the tables
# ---------------------------------------------------------------------------
decompress() {
  if [[ "$SELECTED" == *.gz ]]; then
    gzip -dc "$SELECTED"
  else
    cat "$SELECTED"
  fi
}

info "Restoring $SELECTED into the throwaway database ..."
if ! decompress | docker exec -i "$CONTAINER" \
       psql -U "$PG_USER" -d jotti -q -v ON_ERROR_STOP=1 >/dev/null; then
  fatal "Restore into the throwaway database failed (see psql errors above)."
fi

TABLE_COUNT="$(docker exec "$CONTAINER" \
  psql -U "$PG_USER" -d jotti -tAc \
  "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'")"

if ! [[ "$TABLE_COUNT" =~ ^[0-9]+$ ]] || (( TABLE_COUNT <= 0 )); then
  fatal "Verify failed: the restored database has no tables (count: ${TABLE_COUNT:-unknown})."
fi

# ---------------------------------------------------------------------------
# Step 4 — Summary
# ---------------------------------------------------------------------------
echo ""
info "Verify OK — the dump is restorable."
info "  Dump:   $SELECTED"
info "  Tables: $TABLE_COUNT"
