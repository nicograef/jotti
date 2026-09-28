#!/usr/bin/env bash
# rocks-backup.sh — copies the jotti.rocks acme-dns SQLite database to the laptop
#
# Usage:
#   make rocks-backup DEST=<dir>   # or: ./scripts/rocks-backup.sh <dir>; environment in usage()
#
# Runs on the laptop; the VPS needs rsync, and the SSH user needs the docker group.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

usage() {
  cat >&2 <<'EOF'
Usage: ./scripts/rocks-backup.sh <dest>

Writes <dest>/acme-dns-<timestamp>.db after an integrity check on the VPS.

Environment:
  ROCKS_SSH_HOST        SSH target (default: jotti.rocks)
  ROCKS_SSH_OPTS        extra ssh options, e.g. "-p 2222 -i ~/.ssh/other_key"
  ROCKS_ACMEDNS_VOLUME  Docker volume holding the database (default: jotti_acme-dns-data)
EOF
}

[[ $# -eq 1 ]] || { usage; exit 1; }
DEST="$1"
[[ -d "$DEST" ]] || fatal "Destination directory not found: $DEST"

for tool in ssh rsync; do
  command -v "$tool" &>/dev/null || fatal "$tool is not installed or not on PATH."
done

HOST="${ROCKS_SSH_HOST:-jotti.rocks}"
VOLUME="${ROCKS_ACMEDNS_VOLUME:-jotti_acme-dns-data}"
TARGET="$DEST/acme-dns-$(date +%Y%m%d-%H%M%S).db"

# All calls share one SSH connection: ufw's `limit` on the VPS refuses a 6th new
# connection within 30 s.
CONTROL_PATH="$(mktemp -u "${TMPDIR:-/tmp}/rocks-backup-ssh.XXXXXX")"
read -r -a SSH_OPTS <<< "${ROCKS_SSH_OPTS:-}"
SSH_OPTS+=(-o ControlMaster=auto -o "ControlPath=$CONTROL_PATH" -o ControlPersist=60)

# ssh joins its arguments into one remote command line, so each is quoted here.
remote() {
  local quoted
  printf -v quoted '%q ' "$@"
  # shellcheck disable=SC2029
  ssh "${SSH_OPTS[@]}" "$HOST" "$quoted"
}

cleanup() {
  if [[ -n "${REMOTE_TMP:-}" ]]; then
    remote rm -rf "$REMOTE_TMP" || warn "Could not remove $REMOTE_TMP on $HOST."
  fi
  ssh "${SSH_OPTS[@]}" -O exit "$HOST" 2>/dev/null || true
}
trap cleanup EXIT

REMOTE_TMP="$(remote mktemp -d /tmp/acme-dns-backup.XXXXXX)"
# shellcheck disable=SC2016 # expands on the VPS
REMOTE_UID_GID="$(remote sh -c 'echo "$(id -u):$(id -g)"')"

# The database file is root-only, so a root container reads it through a read-only
# mount and hands the copy to the SSH user.
info "Taking an online backup of $VOLUME on $HOST ..."
CHECK="$(remote docker run --rm \
  -v "$VOLUME:/data:ro" -v "$REMOTE_TMP:/out" \
  alpine:3.24 sh -c "apk add --no-cache -q sqlite \
    && sqlite3 -readonly /data/acme-dns.db '.backup /out/acme-dns.db' \
    && chown $REMOTE_UID_GID /out/acme-dns.db \
    && sqlite3 /out/acme-dns.db 'PRAGMA integrity_check;'")"
[[ "$CHECK" == "ok" ]] || fatal "integrity_check failed on the backup: $CHECK"
info "integrity_check: ok"

info "Copying to $TARGET ..."
rsync -a -e "ssh ${ROCKS_SSH_OPTS:-} -o ControlPath=$CONTROL_PATH" "$HOST:$REMOTE_TMP/acme-dns.db" "$TARGET"

info "Backup written: $TARGET"
