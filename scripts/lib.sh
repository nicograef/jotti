#!/usr/bin/env bash
# lib.sh — shared shell helpers for scripts/*.sh; source, don't execute.
#
# Usage:
#   . "$SCRIPT_DIR/lib.sh"
#
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
# may contain shell-special characters). Trims whitespace including a CRLF's \r,
# then one pair of surrounding quotes, as a .env edited on Windows may carry both.
read_env() {
  local key="$1" value
  value="$({ grep -E "^${key}=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  if [[ "$value" =~ ^\"(.*)\"$ || "$value" =~ ^\'(.*)\'$ ]]; then
    value="${BASH_REMATCH[1]}"
  fi
  printf '%s\n' "$value"
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

# require_pinned_version ACTION — prints JOTTI_VERSION from .env; fatals unless it
# is a pinned release tag. Compose references the tag as a bare ${JOTTI_VERSION}
# with no default: "latest" would pull a moving image and an empty value aborts
# the stack, so this turns both into an early, actionable error.
require_pinned_version() {
  local action="$1" version
  version="$(read_env JOTTI_VERSION)"
  if ! parse_semver "$version" >/dev/null; then
    error "JOTTI_VERSION in .env is not a pinned release tag (found: '${version:-<empty>}')."
    error "Set it to a release tag like vX.Y.Z from https://github.com/nicograef/jotti/releases."
    fatal "Refusing to $action an unpinned version ('latest' and empty are not allowed)."
  fi
  printf '%s\n' "$version"
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

# require_rocks_stack — fatals unless this host runs the jotti.rocks stack: .env
# sets VPS_PUBLIC_IP and the container jotti-acme-dns exists. A self-hosted
# server shares project name and volume jotti_postgres-data with the rocks stack,
# so a rocks reset there would delete the production cash journal.
require_rocks_stack() {
  if ! grep -qE '^VPS_PUBLIC_IP=.+' .env 2>/dev/null; then
    fatal "VPS_PUBLIC_IP missing or empty in .env: this is not the jotti.rocks server. See docs/jotti-rocks-infra.md."
  fi
  if ! docker inspect jotti-acme-dns &>/dev/null; then
    fatal "Container jotti-acme-dns not found: this is not the jotti.rocks server."
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

# wait_https_ok URL TRIES — true once URL answers 2xx, after at most TRIES
# attempts 3 s apart.
wait_https_ok() {
  curl -fs -o /dev/null --max-time 5 --retry "$(($2 - 1))" --retry-delay 3 --retry-all-errors "$1"
}

# admin_otp_from_logs COMPOSE_FILE [SINCE] — prints the newest admin one-time
# login code from the backend logs (since SINCE when given), or nothing. The
# marker sits inside the msg field of a JSON log line, so the substring is matched.
admin_otp_from_logs() {
  docker compose -f "$1" logs ${2:+--since "$2"} backend 2>/dev/null |
    grep -a "ADMIN-EINMALPASSWORT" |
    grep -aoE 'code=[0-9]{6}' |
    tail -n1 | cut -d= -f2 || true
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

# start_throwaway_postgres CONTAINER PORT — replaces CONTAINER with a fresh
# Postgres (user, password and database admin/admin/jotti) on 127.0.0.1:PORT,
# waits for it and applies every migration. The caller's EXIT trap removes it.
start_throwaway_postgres() {
  local container="$1" port="$2" image="postgres:17.11" ready="" i
  local migrations
  migrations="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/database/migrations"

  docker rm -f "$container" >/dev/null 2>&1 || true

  info "Starting $image on port $port ..."
  docker run -d \
    --name "$container" \
    -e POSTGRES_USER=admin \
    -e POSTGRES_PASSWORD=admin \
    -e POSTGRES_DB=jotti \
    -p "127.0.0.1:${port}:5432" \
    "$image" >/dev/null

  # pg_isready also answers for the image's temporary socket-only init server, and
  # migrate then fails with "connection reset by peer"; a TCP query does not.
  info "Waiting for PostgreSQL to accept real connections ..."
  for ((i = 0; i < 30; i++)); do
    if docker exec -e PGPASSWORD=admin "$container" \
      psql -h 127.0.0.1 -U admin -d jotti -c "SELECT 1" >/dev/null 2>&1; then
      ready=1
      break
    fi
    sleep 2
  done
  if [[ -z "$ready" ]]; then
    error "PostgreSQL did not become ready in time. Container logs:"
    docker logs "$container" >&2
    exit 1
  fi

  info "Running database migrations (forward-only) ..."
  migrate -path "$migrations" -database "postgres://admin:admin@localhost:${port}/jotti?sslmode=disable" up
}

# load_allowlist FILE KIND — reads a gate's allowlist into ALLOW_KEYS,
# ALLOW_NEEDLES and ALLOW_HITS, skipping blank and comment lines. KIND is
# fragment ("path fragment  # reason", the fragment one word), name
# ("name  # reason") or quote (the text before "#", spaces allowed).
load_allowlist() {
  local entry key needle reason entries=()
  ALLOWLIST_FILE="$1"
  ALLOW_KIND="$2"
  ALLOW_KEYS=()
  ALLOW_NEEDLES=()
  ALLOW_HITS=()
  if [[ -f "$ALLOWLIST_FILE" ]]; then
    mapfile -t entries < <(grep -vE '^[[:space:]]*(#|$)' "$ALLOWLIST_FILE")
  fi
  for entry in "${entries[@]+"${entries[@]}"}"; do
    case "$ALLOW_KIND" in
      fragment)
        read -r key needle reason _ <<<"$entry"
        if [[ -z "$needle" || "${needle:0:1}" == "#" ]]; then
          fatal "$ALLOWLIST_FILE: entry without a fragment: $entry"
        fi
        # Without this the field split would cut a fragment at its first space
        # and exempt more lines than the entry names.
        [[ "${reason:0:1}" == "#" ]] || fatal "$ALLOWLIST_FILE: fragment with a space, or reason missing: $entry"
        ;;
      name)
        read -r key reason _ <<<"$entry"
        needle=""
        [[ "${reason:0:1}" == "#" ]] || fatal "$ALLOWLIST_FILE: name with a space, or reason missing: $entry"
        ;;
      quote)
        key="$(sed -E 's/[[:space:]]*#.*$//; s/[[:space:]]+$//' <<<"$entry")"
        needle=""
        ;;
      *) fatal "load_allowlist: unknown kind $ALLOW_KIND" ;;
    esac
    ALLOW_KEYS+=("$key")
    ALLOW_NEEDLES+=("$needle")
    ALLOW_HITS+=(0)
  done
}

# allowlisted KEY [TEXT] — true when an entry covers KEY (a fragment entry also
# needs its fragment in TEXT); counts the hit for report_stale_allowlist.
allowlisted() {
  local key="$1" text="${2:-}" i
  for i in "${!ALLOW_KEYS[@]}"; do
    [[ "$key" == "${ALLOW_KEYS[$i]}" ]] || continue
    if [[ "$ALLOW_KIND" == fragment && "$text" != *"${ALLOW_NEEDLES[$i]}"* ]]; then
      continue
    fi
    ALLOW_HITS[i]=$((ALLOW_HITS[i] + 1))
    return 0
  done
  return 1
}

# report_stale_allowlist FORMAT — reports every entry that covered nothing,
# FORMAT a printf format for the entry, and prints their count on stdout.
report_stale_allowlist() {
  local format="$1" i label stale=0
  for i in "${!ALLOW_KEYS[@]}"; do
    ((ALLOW_HITS[i] == 0)) || continue
    label="${ALLOW_KEYS[$i]}"
    if [[ "$ALLOW_KIND" == fragment ]]; then
      label+=" ${ALLOW_NEEDLES[$i]}"
    fi
    # shellcheck disable=SC2059
    error "$ALLOWLIST_FILE: $(printf "$format" "$label")"
    stale=$((stale + 1))
  done
  printf '%s\n' "$stale"
}
