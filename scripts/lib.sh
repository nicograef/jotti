#!/usr/bin/env bash
# jotti — shared shell helpers for scripts/*.sh: color vars, log helpers,
# .env reading, semver parsing, the docker/compose/.env preflight, and
# backup selection/decompression for the prod-*.sh scripts. Source, don't
# execute:
#
#   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#   # shellcheck source=scripts/lib.sh
#   . "$SCRIPT_DIR/lib.sh"
#
# All log helpers write to stderr, so a caller's stdout stays free for a
# machine-readable return value (see ops-smoke.sh's TSV protocol).

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1" >&2; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1" >&2; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

# read_env KEY — read a single value from .env without executing the file
# (passwords may contain shell-special characters). Returns the last match,
# trimmed of surrounding whitespace.
read_env() {
  local key="$1"
  { grep -E "^${key}=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
}

# parse_semver "v1.2.3" — echoes "1 2 3" and returns 0, or returns 1 when the
# value is not a plain vMAJOR.MINOR.PATCH (e.g. "latest", "dev"). A pre-release
# or build suffix ("1.2.3-rc1", "1.2.3+meta") is trimmed before parsing. Mirrors
# core.parseSemver (windows/starter/core/update.go).
parse_semver() {
  local s="${1#v}"
  s="${s%%[-+]*}"
  [[ "$s" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]] || return 1
  printf '%s %s %s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}"
}

# require_docker_stack COMPOSE_FILE [--no-compose-cli] — fatals unless docker
# is on PATH, COMPOSE_FILE exists and .env exists. Also requires the `docker
# compose` (v2) CLI unless --no-compose-cli is passed: prod-backup-verify.sh
# reads COMPOSE_FILE for the pinned postgres version but drives a throwaway
# container via `docker run`/`docker exec`, never `docker compose`, so it does
# not need the CLI installed.
require_docker_stack() {
  local compose_file="$1"
  local need_compose_cli=1
  [[ "${2:-}" == "--no-compose-cli" ]] && need_compose_cli=0

  if ! command -v docker &>/dev/null; then
    fatal "docker is not installed or not on PATH."
  fi
  if (( need_compose_cli )) && ! docker compose version &>/dev/null; then
    fatal "docker compose (v2) is not available."
  fi
  if [[ ! -f "$compose_file" ]]; then
    fatal "Missing compose file: $compose_file"
  fi
  if [[ ! -f .env ]]; then
    fatal ".env file not found. Run 'make init' first."
  fi
}

# resolve_backup_dir — sets the global BACKUP_DIR: the environment wins, then
# .env, then the built-in default "./backups".
resolve_backup_dir() {
  BACKUP_DIR="${BACKUP_DIR:-$(read_env BACKUP_DIR)}"
  [[ -n "$BACKUP_DIR" ]] || BACKUP_DIR="./backups"
}

# select_dump BACKUP_DIR [TARGET] — resolves the dump to use: TARGET as given
# (a path, or a filename under BACKUP_DIR) when set, otherwise the newest dump
# found in BACKUP_DIR. Sets the globals SELECTED (the resolved path) and
# DUMPS_FOUND (every dump found in BACKUP_DIR, oldest to newest — callers that
# also list the available backups read this instead of re-scanning the dir).
select_dump() {
  local dir="$1"
  local target="${2:-}"

  mapfile -t DUMPS_FOUND < <(find "$dir" -maxdepth 1 -type f \
    \( -name 'jotti-*.sql' -o -name 'jotti-*.sql.gz' \) -printf '%f\n' 2>/dev/null | sort)

  if [[ -n "$target" ]]; then
    if [[ -f "$target" ]]; then
      SELECTED="$target"
    elif [[ -f "$dir/$target" ]]; then
      SELECTED="$dir/$target"
    else
      fatal "Backup not found: $target"
    fi
  else
    (( ${#DUMPS_FOUND[@]} > 0 )) || fatal "No backups found in $dir. Pass a dump file explicitly."
    SELECTED="$dir/${DUMPS_FOUND[-1]}"
  fi
}

# decompress SELECTED — streams SELECTED to stdout, gunzipping on the fly when
# its name ends in .gz. Reads the global SELECTED set by select_dump.
decompress() {
  if [[ "$SELECTED" == *.gz ]]; then
    gzip -dc "$SELECTED"
  else
    cat "$SELECTED"
  fi
}
