#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — Safe Update (self-hosted production, Weg B)
#
# Updates the pinned production stack to the JOTTI_VERSION currently set in .env.
# Mirrors the Windows starter's update flow (windows/starter/main.go): refuse
# downgrades, take a pre-update backup BEFORE any migration runs, pull, apply,
# then verify health. If the new version does not come up healthy, the operator
# gets a clear, copy-pasteable rollback path and the script aborts non-zero —
# no data created before the update is lost. Steps:
#   1. Validate prerequisites (Docker, Compose, .env)
#   2. Determine the running vs. target version; refuse downgrades
#   3. Pre-update backup (calls prod-backup.sh)
#   4. Pull the pinned images and apply (runs migrations)
#   5. Wait for health; on failure print rollback guidance and abort
#
# Update workflow: bump JOTTI_VERSION in .env, then run `make prod-update`.
#
# Usage: ./scripts/prod-update.sh  (or `make prod-update`)
# =============================================================================

COMPOSE_PROD="docker-compose.prod.yml"
BACKEND_CONTAINER="jotti-backend"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

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

# is_downgrade TARGET RUNNING — returns 0 when TARGET is a strictly older semver
# than RUNNING. Returns 1 when it is not older OR when either side is not semver
# (then ordering is unknown, so we do not block — downgrade protection needs
# pinned semver versions, not "latest").
is_downgrade() {
  local t r ta tb tc ra rb rc
  t="$(parse_semver "$1")" || return 1
  r="$(parse_semver "$2")" || return 1
  read -r ta tb tc <<<"$t"
  read -r ra rb rc <<<"$r"
  if (( ta != ra )); then (( ta < ra )) && return 0 || return 1; fi
  if (( tb != rb )); then (( tb < rb )) && return 0 || return 1; fi
  if (( tc != rc )); then (( tc < rc )) && return 0 || return 1; fi
  return 1
}

# ---------------------------------------------------------------------------
# Step 0 — Change to project root (script may be called from anywhere)
# ---------------------------------------------------------------------------
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

# ---------------------------------------------------------------------------
# Step 1 — Validate prerequisites
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

# ---------------------------------------------------------------------------
# Step 2 — Determine running vs. target version and guard against downgrades
# ---------------------------------------------------------------------------
# Target comes from .env (the version the operator bumped to). Only a pinned
# release tag (vMAJOR.MINOR.PATCH) is accepted; "latest" or an empty value would
# silently track a moving image and defeat the downgrade guard. Compose references
# the tag as a bare ${JOTTI_VERSION} with no default, so an empty value aborts the
# stack; this check turns that into an early, actionable error.
TARGET_VERSION="$(read_env JOTTI_VERSION)"
if ! parse_semver "$TARGET_VERSION" >/dev/null; then
  error "JOTTI_VERSION in .env is not a pinned release tag (found: '${TARGET_VERSION:-<empty>}')."
  error "Set it to a release tag like v0.3.1 from https://github.com/nicograef/jotti/releases."
  fatal "Refusing to update against an unpinned version ('latest' and empty are not allowed)."
fi

# Running version is the tag the backend container was created from. Absence of
# the container means there is nothing to update yet.
RUNNING_IMAGE="$(docker inspect -f '{{.Config.Image}}' "$BACKEND_CONTAINER" 2>/dev/null || true)"
if [[ -z "$RUNNING_IMAGE" ]]; then
  fatal "No running jotti stack found (container '$BACKEND_CONTAINER' is absent). Use 'make prod-init' for the first deploy."
fi
RUNNING_VERSION="${RUNNING_IMAGE##*:}"

if [[ "$TARGET_VERSION" == "$RUNNING_VERSION" ]]; then
  warn "Target version ($TARGET_VERSION) equals the running version — re-deploying the same version."
elif is_downgrade "$TARGET_VERSION" "$RUNNING_VERSION"; then
  error "Downgrade refused: JOTTI_VERSION ($TARGET_VERSION) is older than the running version ($RUNNING_VERSION)."
  error "Updates change the database and cannot be undone by downgrading; an older version cannot start on newer data."
  fatal "To go back, restore a backup instead (see docs/leitfaden.md)."
else
  info "Updating: $RUNNING_VERSION -> $TARGET_VERSION"
fi

# ---------------------------------------------------------------------------
# Step 3 — Pre-update backup (before any migration runs)
# ---------------------------------------------------------------------------
# Resolve BACKUP_DIR exactly like prod-backup.sh so we can locate the dump it
# just wrote and offer it for rollback.
BACKUP_DIR="${BACKUP_DIR:-$(read_env BACKUP_DIR)}"
[[ -n "$BACKUP_DIR" ]] || BACKUP_DIR="./backups"

info "Taking a pre-update backup..."
"$SCRIPT_DIR/prod-backup.sh"

NEWEST_DUMP="$(find "$BACKUP_DIR" -maxdepth 1 -type f -name 'jotti-*.sql.gz' -printf '%f\n' 2>/dev/null | sort | tail -n1)"
if [[ -z "$NEWEST_DUMP" ]]; then
  fatal "Pre-update backup did not produce a dump in $BACKUP_DIR. Aborting before any change."
fi
PRE_UPDATE_DUMP="$BACKUP_DIR/$NEWEST_DUMP"
info "Pre-update backup ready: $PRE_UPDATE_DUMP"

# rollback_guidance prints the copy-pasteable path back to the previous version.
rollback_guidance() {
  echo "" >&2
  error "Update failed: the stack did not come up healthy."
  echo "" >&2
  warn "A pre-update backup was taken before any migration ran:"
  warn "  $PRE_UPDATE_DUMP"
  echo "" >&2
  warn "Roll back to the previous version ($RUNNING_VERSION) in two steps:"
  warn "  1. Set JOTTI_VERSION=$RUNNING_VERSION in .env"
  warn "  2. ./scripts/prod-restore.sh $PRE_UPDATE_DUMP"
  warn "     (restores the pre-update database and restarts the previous version)"
  echo "" >&2
  warn "No data created before the update is lost — it is in the backup above."
}

# ---------------------------------------------------------------------------
# Step 4 — Pull the pinned images and apply the update
# ---------------------------------------------------------------------------
info "Pulling pinned images for $TARGET_VERSION..."
if ! docker compose -f "$COMPOSE_PROD" pull; then
  fatal "docker compose pull failed. Nothing was changed; the previous version is still running."
fi

info "Applying the update (this runs database migrations)..."
if ! docker compose -f "$COMPOSE_PROD" up -d; then
  rollback_guidance
  exit 1
fi

# ---------------------------------------------------------------------------
# Step 5 — Wait for the backend to become healthy, then verify HTTPS
# ---------------------------------------------------------------------------
info "Waiting for the backend to become healthy..."
backend_healthy=false
for _ in $(seq 1 30); do
  status="$(docker inspect -f '{{.State.Health.Status}}' "$BACKEND_CONTAINER" 2>/dev/null || echo unknown)"
  if [[ "$status" == "healthy" ]]; then
    backend_healthy=true
    break
  fi
  sleep 2
done

if [[ "$backend_healthy" != true ]]; then
  rollback_guidance
  exit 1
fi
info "Backend healthy."

# Best-effort public health check (the certificate already exists from the
# previous deploy, so this should pass quickly).
DOMAIN="$(read_env JOTTI_DOMAIN)"
https_ok=false
if [[ -n "$DOMAIN" ]]; then
  info "Verifying https://$DOMAIN/api/health ..."
  for _ in $(seq 1 10); do
    code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "https://$DOMAIN/api/health" 2>/dev/null || echo 000)"
    if [[ "$code" == "200" ]]; then
      https_ok=true
      break
    fi
    sleep 3
  done
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "=========================================="
printf "${GREEN} %s${NC}\n" "jotti — Update Complete"
echo "=========================================="
echo ""
echo "  Version: $RUNNING_VERSION -> $TARGET_VERSION"
if [[ -n "$DOMAIN" ]]; then
  echo "  Domain:  https://$DOMAIN"
  if [[ "$https_ok" == true ]]; then
    info "HTTPS check: OK (/api/health returned 200)"
  else
    warn "HTTPS did not return 200 yet — re-check in a minute or follow logs:"
    warn "  docker compose -f $COMPOSE_PROD logs -f reverse-proxy"
  fi
fi
echo ""
echo "  Pre-update backup: $PRE_UPDATE_DUMP"
echo "=========================================="
