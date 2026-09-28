#!/usr/bin/env bash
set -euo pipefail

# jotti — database restore (self-hosted production).
#
# Restores a pg_dump created by prod-backup.sh into the production database.
# DESTRUCTIVE: the dumps use --clean --if-exists, so objects are dropped and
# re-created; the application services are stopped during the restore. The
# restore runs in one transaction, so a failure leaves the database unchanged.

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
PG_SERVICE="postgres"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

require_docker_stack "$COMPOSE_FILE"

resolve_backup_dir
select_dump "$BACKUP_DIR" "${1:-}"

if (( ${#DUMPS_FOUND[@]} > 0 )); then
  info "Available backups in $BACKUP_DIR:"
  for d in "${DUMPS_FOUND[@]}"; do
    echo "    $d"
  done
fi

# Reject a corrupt archive before the stack is touched; the same check guards
# the write side in prod-backup.sh.
if [[ "$SELECTED" == *.gz ]]; then
  info "Checking the archive (gzip -t) ..."
  if ! gzip -t "$SELECTED"; then
    fatal "Integrity check failed (gzip -t): $SELECTED is corrupt. Nothing was changed."
  fi
fi

# A dump cut at a statement boundary restores without an SQL error; only
# pg_dump's closing comment proves the file is complete.
dump_tail="$(decompress | tail -n 20)"
if ! grep -qx -- '-- PostgreSQL database dump complete' <<<"$dump_tail"; then
  fatal "$SELECTED lacks pg_dump's closing line, so it is truncated. Nothing was changed."
fi

echo ""
warn "This will OVERWRITE the current jotti database with:"
warn "  Dump:    $SELECTED"
warn "  Stack:   $COMPOSE_FILE"
warn "All data created since that backup will be lost."
read -r -p "Continue? Type 'yes' to proceed: " answer
[[ "$answer" == "yes" ]] || fatal "Aborted by user. Nothing was changed."

info "Starting the database ..."
docker compose -f "$COMPOSE_FILE" up -d --wait "$PG_SERVICE"

info "Stopping application services during the restore ..."
docker compose -f "$COMPOSE_FILE" stop backend frontend reverse-proxy

# -1 with ON_ERROR_STOP rolls the whole restore back on the first SQL error;
# psql ignores -1 without -f, hence `-f -` for stdin. The postgres role comes
# from the container's own POSTGRES_USER.
info "Restoring $SELECTED ..."
if ! decompress | docker compose -f "$COMPOSE_FILE" exec -T "$PG_SERVICE" \
       sh -c 'psql -U "$POSTGRES_USER" -d jotti -1 -v ON_ERROR_STOP=1 -f -'; then
  error "Restore failed and was rolled back; the database is unchanged."
  error "Application services are stopped. Restart them: make prod-up"
  exit 1
fi

# A restored database has no planner statistics until autovacuum reaches it.
info "Updating planner statistics ..."
if ! docker compose -f "$COMPOSE_FILE" exec -T "$PG_SERVICE" \
       sh -c 'vacuumdb -U "$POSTGRES_USER" -d jotti --analyze-in-stages'; then
  warn "vacuumdb failed; the data is restored, autovacuum will analyze it later."
fi

info "Restarting the full stack ..."
docker compose -f "$COMPOSE_FILE" up -d

# Force-recreate the reverse-proxy so it re-resolves the freshly restarted
# backend/frontend upstreams. On the rocks stack this clears nginx's cached
# upstream IPs (the 502-after-restore trap); on Caddy stacks it is a no-op.
info "Recreating the reverse-proxy ..."
docker compose -f "$COMPOSE_FILE" up -d --no-deps --force-recreate reverse-proxy

echo ""
info "Restore complete. jotti is running again."
