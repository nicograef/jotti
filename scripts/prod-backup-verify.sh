#!/usr/bin/env bash
set -euo pipefail

# jotti — backup verify (self-hosted production).
#
# Proves that a pg_dump from prod-backup.sh is restorable: it replays the dump
# into a THROWAWAY postgres container (`docker run --rm`, no stack network, no
# stack volumes) and checks that the restored database has tables. The running
# stack is never touched.

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
       psql -U "$PG_USER" -d jotti -q -v ON_ERROR_STOP=1 >/dev/null; then
  fatal "Restore into the throwaway database failed (see psql errors above)."
fi

TABLE_COUNT="$(docker exec "$CONTAINER" \
  psql -U "$PG_USER" -d jotti -tAc \
  "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'")"

if ! [[ "$TABLE_COUNT" =~ ^[0-9]+$ ]] || (( TABLE_COUNT <= 0 )); then
  fatal "Verify failed: the restored database has no tables (count: ${TABLE_COUNT:-unknown})."
fi

echo ""
info "Verify OK — the dump is restorable."
info "  Dump:   $SELECTED"
info "  Tables: $TABLE_COUNT"
