#!/usr/bin/env bash
# jotti — shared shell helpers for scripts/*.sh; source, don't execute.
# Every log helper writes to stderr, so a caller's stdout stays free for a
# machine-readable return value (see ops-smoke.sh's TSV protocol).

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1" >&2; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1" >&2; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

# read_env KEY — reads a value from .env without executing the file (passwords
# may contain shell-special characters).
read_env() {
  local key="$1"
  { grep -E "^${key}=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
}

# parse_semver "v1.2.3" — echoes "1 2 3"; returns 1 for a non-semver value
# ("latest", "dev"), a pre-release or build suffix is trimmed first. Mirrors
# core.parseSemver (windows/starter/core/update.go).
parse_semver() {
  local s="${1#v}"
  s="${s%%[-+]*}"
  [[ "$s" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]] || return 1
  printf '%s %s %s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}"
}

# require_docker_stack COMPOSE_FILE [--no-compose-cli] — fatals unless docker,
# COMPOSE_FILE and .env are there. --no-compose-cli drops the `docker compose`
# (v2) requirement for prod-backup-verify.sh, which drives a throwaway container
# with docker run/exec only.
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

# wait_for_healthy CONTAINER — polls the Docker healthcheck for up to 60 s;
# returns 1 if the container is not healthy by then.
wait_for_healthy() {
  local container="$1" status
  for _ in $(seq 1 30); do
    status="$(docker inspect -f '{{.State.Health.Status}}' "$container" 2>/dev/null || echo unknown)"
    [[ "$status" == "healthy" ]] && return 0
    sleep 2
  done
  return 1
}

resolve_backup_dir() {
  BACKUP_DIR="${BACKUP_DIR:-$(read_env BACKUP_DIR)}"
  [[ -n "$BACKUP_DIR" ]] || BACKUP_DIR="./backups"
}

# select_dump BACKUP_DIR [TARGET] — sets SELECTED (TARGET as a path or a name in
# BACKUP_DIR, otherwise the newest dump there) and DUMPS_FOUND (every dump in
# BACKUP_DIR, oldest first) for callers that also list the backups.
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

decompress() {
  if [[ "$SELECTED" == *.gz ]]; then
    gzip -dc "$SELECTED"
  else
    cat "$SELECTED"
  fi
}

# tracked_text_files — the tracked corpus both repo-wide text gates police.
# Excluded: paths frozen by the freeze discipline, rule texts that quote the
# banned words themselves, and generated or vendored files.
tracked_text_files() {
  git ls-files \
    ':(glob,exclude)CHANGELOG.md' \
    ':(glob,exclude)docs/plans/**' \
    ':(glob,exclude)docs/rechtsquellen/**' \
    ':(glob,exclude)database/migrations/**' \
    ':(glob,exclude)backend/sqlc/dbgen/**' \
    ':(glob,exclude)AGENTS.md' \
    ':(glob,exclude).github/copilot-instructions.md' \
    ':(glob,exclude).github/instructions/**' \
    ':(glob,exclude).claude/**' \
    ':(glob,exclude)**/pnpm-lock.yaml' \
    ':(glob,exclude)**/go.sum'
}
