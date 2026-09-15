#!/usr/bin/env bash
set -euo pipefail

# jotti — safe update of the self-hosted production stack to the JOTTI_VERSION
# set in .env. Mirrors the Windows starter's update flow
# (windows/starter/main.go): refuse downgrades, take a backup BEFORE any
# migration runs, pull, apply, verify health. If the new version does not come up
# healthy, the operator gets a copy-pasteable rollback path and the script aborts
# non-zero — no data created before the update is lost.

COMPOSE_PROD="docker-compose.prod.yml"
BACKEND_CONTAINER="jotti-backend"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

# is_downgrade TARGET RUNNING — 0 when TARGET is a strictly older semver than
# RUNNING. Either side not semver returns 1: the ordering is unknown then, and
# downgrade protection needs pinned semver versions, not "latest".
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

PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

require_docker_stack "$COMPOSE_PROD"

# Only a pinned release tag (vMAJOR.MINOR.PATCH) is accepted: compose references
# the tag as a bare ${JOTTI_VERSION} with no default, so "latest" would track a
# moving image and defeat the downgrade guard, and an empty value would abort the
# stack.
TARGET_VERSION="$(read_env JOTTI_VERSION)"
if ! parse_semver "$TARGET_VERSION" >/dev/null; then
  error "JOTTI_VERSION in .env is not a pinned release tag (found: '${TARGET_VERSION:-<empty>}')."
  error "Set it to a release tag like vX.Y.Z from https://github.com/nicograef/jotti/releases."
  fatal "Refusing to update against an unpinned version ('latest' and empty are not allowed)."
fi

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
  fatal "To go back, restore a backup instead (see docs/leitfaden/aktualisieren-backups.md)."
else
  info "Updating: $RUNNING_VERSION -> $TARGET_VERSION"
fi

# Locate the dump prod-backup.sh just wrote so it can be offered for rollback.
resolve_backup_dir

info "Taking a pre-update backup..."
"$SCRIPT_DIR/prod-backup.sh"

NEWEST_DUMP="$(find "$BACKUP_DIR" -maxdepth 1 -type f -name 'jotti-*.sql.gz' -printf '%f\n' 2>/dev/null | sort | tail -n1)"
if [[ -z "$NEWEST_DUMP" ]]; then
  fatal "Pre-update backup did not produce a dump in $BACKUP_DIR. Aborting before any change."
fi
PRE_UPDATE_DUMP="$BACKUP_DIR/$NEWEST_DUMP"
info "Pre-update backup ready: $PRE_UPDATE_DUMP"

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

info "Pulling pinned images for $TARGET_VERSION..."
if ! docker compose -f "$COMPOSE_PROD" pull; then
  fatal "docker compose pull failed. Nothing was changed; the previous version is still running."
fi

info "Applying the update (this runs database migrations)..."
if ! docker compose -f "$COMPOSE_PROD" up -d; then
  rollback_guidance
  exit 1
fi

info "Waiting for the backend to become healthy..."
if ! wait_for_healthy "$BACKEND_CONTAINER"; then
  rollback_guidance
  exit 1
fi
info "Backend healthy."

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
