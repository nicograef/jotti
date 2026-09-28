#!/usr/bin/env bash
set -euo pipefail

# jotti.rocks — copies the acme-dns SQLite database from the VPS to a local
# directory. Runs on the laptop; the VPS needs sqlite3 and rsync, and the SSH
# user needs read access to the Docker volume.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

usage() {
  cat >&2 <<'EOF'
Usage: ./scripts/rocks-backup.sh <dest>

Writes <dest>/acme-dns-<timestamp>.db after an integrity check on the VPS.

Environment:
  ROCKS_SSH_HOST    SSH target (default: jotti.rocks)
  ROCKS_SSH_OPTS    extra ssh options, e.g. "-p 2222 -i ~/.ssh/other_key"
  ROCKS_ACMEDNS_DB  database path on the VPS
                    (default: /var/lib/docker/volumes/jotti_acme-dns-data/_data/acme-dns.db)
EOF
}

[[ $# -eq 1 ]] || { usage; exit 1; }
DEST="$1"
[[ -d "$DEST" ]] || fatal "Destination directory not found: $DEST"

for tool in ssh rsync; do
  command -v "$tool" &>/dev/null || fatal "$tool is not installed or not on PATH."
done

HOST="${ROCKS_SSH_HOST:-jotti.rocks}"
DB="${ROCKS_ACMEDNS_DB:-/var/lib/docker/volumes/jotti_acme-dns-data/_data/acme-dns.db}"
read -r -a SSH_OPTS <<< "${ROCKS_SSH_OPTS:-}"
TARGET="$DEST/acme-dns-$(date +%Y%m%d-%H%M%S).db"

# ssh joins its arguments into one remote command line, so each is quoted here.
remote() {
  local quoted
  printf -v quoted '%q ' "$@"
  # shellcheck disable=SC2029
  ssh "${SSH_OPTS[@]}" "$HOST" "$quoted"
}

info "Checking $DB on $HOST ..."
remote test -r "$DB" || fatal "Cannot read $DB on $HOST."

REMOTE_TMP="$(remote mktemp /tmp/acme-dns-backup.XXXXXX)"
trap 'remote rm -f "$REMOTE_TMP" || warn "Could not remove $REMOTE_TMP on $HOST."' EXIT

info "Taking an online backup ..."
remote sqlite3 "$DB" ".backup '$REMOTE_TMP'"

CHECK="$(remote sqlite3 "$REMOTE_TMP" "PRAGMA integrity_check;")"
[[ "$CHECK" == "ok" ]] || fatal "integrity_check failed on the backup: $CHECK"
info "integrity_check: ok"

info "Copying to $TARGET ..."
rsync -a -e "ssh ${ROCKS_SSH_OPTS:-}" "$HOST:$REMOTE_TMP" "$TARGET"

info "Backup written: $TARGET"
