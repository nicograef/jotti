#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — Database Backup (self-hosted production, Weg B)
#
# Pulls a full pg_dump from the running production postgres into a timestamped,
# gzip-compressed file in a host directory and rotates old dumps. Mirrors the
# Windows pre-update backup (cmd/starter/backup.go): same --clean --if-exists
# dump and same "keep newest N" rotation, so a later restore re-creates the
# objects cleanly. Steps:
#   1. Validate prerequisites (Docker, Compose, .env)
#   2. pg_dump the running stack into BACKUP_DIR/jotti-YYYYMMDD-HHMMSS.sql.gz
#   3. Rotate to the newest BACKUP_KEEP dumps
#
# Configuration (environment overrides .env, which overrides the defaults):
#   BACKUP_DIR   target directory on the host (default: ./backups)
#   BACKUP_KEEP  number of dumps to retain (default: 14; <=0 keeps all)
#
# Usage: ./scripts/prod-backup.sh  (or `make prod-backup`)
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
# Step 1 — Validate prerequisites and resolve configuration
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

# Environment wins, then .env, then the built-in default.
BACKUP_DIR="${BACKUP_DIR:-$(read_env BACKUP_DIR)}"
[[ -n "$BACKUP_DIR" ]] || BACKUP_DIR="./backups"
BACKUP_KEEP="${BACKUP_KEEP:-$(read_env BACKUP_KEEP)}"
[[ -n "$BACKUP_KEEP" ]] || BACKUP_KEEP="14"

if ! [[ "$BACKUP_KEEP" =~ ^-?[0-9]+$ ]]; then
  fatal "BACKUP_KEEP must be an integer (got: $BACKUP_KEEP)."
fi

mkdir -p "$BACKUP_DIR"

# ---------------------------------------------------------------------------
# Step 2 — Dump the database
# ---------------------------------------------------------------------------
# The postgres role lives in the container's own POSTGRES_USER env, so the dump
# never drifts from whatever .env configured. Local socket connections inside
# the container are trust-authenticated, so no password is needed.
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUTFILE="$BACKUP_DIR/jotti-$TIMESTAMP.sql.gz"
TMPFILE="$OUTFILE.partial"

cleanup() { rm -f "$TMPFILE"; }
trap cleanup EXIT

info "Dumping database to $OUTFILE ..."
if ! docker compose -f "$COMPOSE_PROD" exec -T "$PG_SERVICE" \
       sh -c 'pg_dump --clean --if-exists -U "$POSTGRES_USER" -d jotti' \
     | gzip -c > "$TMPFILE"; then
  fatal "pg_dump failed. Is the stack running? Start it with: make prod-up"
fi
mv "$TMPFILE" "$OUTFILE"

info "Backup created: $OUTFILE ($(du -h "$OUTFILE" | cut -f1))"

# ---------------------------------------------------------------------------
# Step 3 — Rotate old dumps (keep the newest BACKUP_KEEP)
# ---------------------------------------------------------------------------
# Timestamped names sort lexicographically == chronologically, so the oldest
# beyond BACKUP_KEEP are the leading entries. A non-positive keep deletes
# nothing — a misconfiguration must never wipe all backups.
if (( BACKUP_KEEP <= 0 )); then
  warn "BACKUP_KEEP=$BACKUP_KEEP (<=0): keeping all backups, no rotation."
else
  mapfile -t dumps < <(find "$BACKUP_DIR" -maxdepth 1 -type f -name 'jotti-*.sql.gz' -printf '%f\n' | sort)
  total=${#dumps[@]}
  if (( total > BACKUP_KEEP )); then
    for ((i = 0; i < total - BACKUP_KEEP; i++)); do
      info "Rotating old backup: ${dumps[i]}"
      rm -f "$BACKUP_DIR/${dumps[i]}"
    done
  fi
fi

echo ""
info "Done. Copy backups off this server regularly (10-year retention; see docs/betrieb/leitfaden-betreiber.md)."
