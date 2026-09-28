#!/usr/bin/env bash
# prod-backup-verify.sh — proves a production dump is restorable
#
# Usage:
#   make prod-backup-verify
#   ./scripts/prod-backup-verify.sh [DUMP]   # default: the newest dump in BACKUP_DIR
#
# What it does:
#   1. Starts a THROWAWAY Postgres of the stack's version (`docker run --rm`, no
#      stack network, no stack volumes).
#   2. Replays the dump in one transaction.
#   3. Fails unless the restored kassenjournal holds events.
# The running stack is never touched.
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# No docker-compose CLI needed here: this script drives a throwaway
# container via `docker run`/`docker exec`, never `docker compose`.
require_docker_stack "$COMPOSE_FILE" --no-compose-cli

# The throwaway postgres uses the same role the dump was created with, so its
# ownership statements (ALTER ... OWNER TO) resolve.
PG_USER="${POSTGRES_USER:-$(read_env POSTGRES_USER)}"
[[ -n "$PG_USER" ]] || PG_USER="admin"

# Pin the throwaway container to the exact postgres version of the stack so the
# verify never drifts from what actually holds the data.
PG_IMAGE="$(grep -oE 'postgres:[0-9][0-9.]*' "$COMPOSE_FILE" | head -n1)"
[[ -n "$PG_IMAGE" ]] || fatal "Could not read the postgres version from $COMPOSE_FILE."

resolve_backup_dir
select_dump "$BACKUP_DIR" "${1:-}"

# --rm plus no --network and no -p: the container shares no network with the
# stack and publishes no port. -fv on removal takes its anonymous data volume
# with it, so `docker volume ls` stays unchanged.
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

info "Restoring $SELECTED into the throwaway database ..."
if ! decompress | docker exec -i "$CONTAINER" \
       psql -U "$PG_USER" -d jotti -q -1 -v ON_ERROR_STOP=1 -f - >/dev/null; then
  fatal "Restore into the throwaway database failed (see psql errors above)."
fi

# A missing table fails the query, so an empty or schema-only dump fails too.
JOURNAL_COUNT="$(docker exec "$CONTAINER" \
  psql -U "$PG_USER" -d jotti -tAc "SELECT count(*) FROM kassenjournal" || true)"

if ! [[ "$JOURNAL_COUNT" =~ ^[0-9]+$ ]] || (( JOURNAL_COUNT <= 0 )); then
  fatal "Verify failed: the restored kassenjournal holds no events (count: ${JOURNAL_COUNT:-unknown})."
fi

echo ""
info "Verify OK — the dump is restorable."
info "  Dump:                 $SELECTED"
info "  kassenjournal events: $JOURNAL_COUNT"
