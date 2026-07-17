#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — Database Backup (self-hosted production, Weg B)
#
# Pulls a full pg_dump from the running production postgres into a timestamped,
# gzip-compressed file in a host directory and rotates old dumps. Mirrors the
# Windows pre-update backup (windows/starter/backup.go): same --clean --if-exists
# dump and same "keep newest N" rotation, so a later restore re-creates the
# objects cleanly. Steps:
#   1. Validate prerequisites (Docker, Compose, .env)
#   2. pg_dump the running stack into BACKUP_DIR/jotti-YYYYMMDD-HHMMSS.sql.gz
#   3. Rotate to the newest BACKUP_KEEP dumps
#
# Configuration (environment overrides .env, which overrides the defaults):
#   BACKUP_DIR       target directory on the host (default: ./backups)
#   BACKUP_KEEP      number of dumps to retain (default: 14; <=0 keeps all)
#   COMPOSE_FILE     compose file to dump from (default: docker-compose.prod.yml)
#   BACKUP_PING_URL  optional URL pinged after a successful backup (dead man's
#                    switch); a failed ping is a warning, never an error
#
# Usage: ./scripts/prod-backup.sh  (or `make prod-backup`)
# =============================================================================

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
PG_SERVICE="postgres"

# ---------------------------------------------------------------------------
# Step 0 — Change to project root (script may be called from anywhere)
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"
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
if [[ ! -f "$COMPOSE_FILE" ]]; then
  fatal "Missing compose file: $COMPOSE_FILE"
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

BACKUP_PING_URL="${BACKUP_PING_URL:-$(read_env BACKUP_PING_URL)}"

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
if ! docker compose -f "$COMPOSE_FILE" exec -T "$PG_SERVICE" \
       sh -c 'pg_dump --clean --if-exists -U "$POSTGRES_USER" -d jotti' \
     | gzip -c > "$TMPFILE"; then
  fatal "pg_dump failed. Is the stack running? Start it with: make prod-up"
fi

# A truncated pipe or a broken gzip stream must never masquerade as a backup.
# Verify the compressed dump before promoting the .partial file; on failure the
# EXIT trap discards it, so no .sql.gz is ever left behind.
if ! gzip -t "$TMPFILE"; then
  fatal "Integrity check failed (gzip -t): the dump is corrupt and was discarded."
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

# ---------------------------------------------------------------------------
# Step 4 — Success ping (dead man's switch)
# ---------------------------------------------------------------------------
# Only after a fully successful, integrity-checked dump: ping an optional
# monitor so it can alarm when the ping ever stops arriving. A failed ping is a
# warning, never a script error — the backup itself already succeeded.
if [[ -n "$BACKUP_PING_URL" ]]; then
  if curl -fsS --connect-timeout 5 --max-time 10 "$BACKUP_PING_URL" >/dev/null 2>&1; then
    info "Success ping sent to BACKUP_PING_URL."
  else
    warn "Success ping to BACKUP_PING_URL failed (backup is fine)."
  fi
fi

echo ""
info "Done. Copy backups off this server regularly (10-year retention; see docs/leitfaden.md)."
