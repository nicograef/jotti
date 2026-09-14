#!/usr/bin/env bash
set -euo pipefail

# jotti — ops smoke test (self-hosted production): drives the production scripts
# (prod-init.sh, prod-update.sh, prod-backup.sh, prod-backup-verify.sh) end to
# end and logs every step machine-readably on stdout, and to LOG_FILE when set:
#   <unix_ts>\t<step>\t<status: ok|fail>\t<duration_seconds>\t<detail>
# The first failed step aborts the run.
#
#   ./scripts/ops-smoke.sh install          # prod-init, set-password, login
#   ./scripts/ops-smoke.sh ops              # backup, backup-verify, update
#   ./scripts/ops-smoke.sh release VERSION  # install plus sale, receipt, export
#
# install and release need a FRESH host: prod-init only issues a one-time admin
# password on first bootstrap, so a rerun fails at parse-admin-otp. Every mode
# also checks the reverse proxy's security headers and login rate limit.
# Host provisioning and the TLS/certificate acceptance stay manual
# (docs/leitfaden/self-hosting.md).
# NEVER runs prod-restore.sh, `docker compose down -v`, or deletes a volume —
# no mode has a destructive step.

COMPOSE_PROD="docker-compose.prod.yml"

# Private (0700, mktemp default) scratch dir for step logs, HTTP bodies and
# the admin JWT — these can contain the admin OTP, password or a valid token,
# so a world-readable fixed /tmp path is avoided. Removed on exit.
SMOKE_TMP="$(mktemp -d "${TMPDIR:-/tmp}/ops-smoke.XXXXXX")"
trap 'rm -rf "$SMOKE_TMP"' EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

log_line() {
  local step="$1" status="$2" duration="$3" detail="${4:-}"
  local line
  line="$(printf '%s\t%s\t%s\t%s\t%s' "$(date +%s)" "$step" "$status" "$duration" "$detail")"
  echo "$line"
  if [[ -n "${LOG_FILE:-}" ]]; then
    echo "$line" >>"$LOG_FILE"
  fi
}

# run_step STEP CMD... — times CMD, logs ok or fail, and aborts on failure.
run_step() {
  local step="$1"
  shift
  local start end duration
  start="$(date +%s)"
  if "$@" >"$SMOKE_TMP/step.log" 2>&1; then
    end="$(date +%s)"
    duration=$((end - start))
    log_line "$step" ok "$duration"
    return 0
  fi
  end="$(date +%s)"
  duration=$((end - start))
  log_line "$step" fail "$duration" "see $SMOKE_TMP/step.log"
  error "Step '$step' failed after ${duration}s. Output:"
  cat "$SMOKE_TMP/step.log" >&2
  exit 1
}

# fail_step STEP DETAIL — logs a fail line for an inline check and aborts.
fail_step() {
  local step="$1" detail="${2:-}"
  log_line "$step" fail 0 "$detail"
  error "Step '$step' failed: $detail"
  exit 1
}

ok_step() {
  local step="$1" duration="$2" detail="${3:-}"
  log_line "$step" ok "$duration" "$detail"
}

# http_post_status URL DATA [AUTH_HEADER] — echoes the HTTP status code; the
# response body lands in $SMOKE_TMP/body.json.
http_post_status() {
  local url="$1" data="$2"
  curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' ${3:+-H "$3"} -d "$data" "$url" 2>"$SMOKE_TMP/curl.log" || echo 000
}

# json_field FIELD — extracts a top-level field from $SMOKE_TMP/body.json
# without a jq dependency; every value read here is a simple scalar.
json_field() {
  local field="$1"
  grep -oE "\"${field}\"[[:space:]]*:[[:space:]]*\"?[^,}\"]*\"?" "$SMOKE_TMP/body.json" \
    | head -n1 | sed -E "s/\"${field}\"[[:space:]]*:[[:space:]]*//; s/^\"//; s/\"\$//"
}

# redacted_body — dumps $SMOKE_TMP/body.json with any "token" field masked, so a
# JWT never lands in stderr or a CI log.
redacted_body() {
  sed -E 's/("token"[[:space:]]*:[[:space:]]*)"[^"]*"/\1"[redacted]"/' "$SMOKE_TMP/body.json" 2>/dev/null
}

MODE="${1:-}"
case "$MODE" in
  install|ops) ;;
  release)
    RELEASE_VERSION="${2:-}"
    [[ -n "$RELEASE_VERSION" ]] || { error "Usage: $0 release VERSION (e.g. v0.14.0)"; exit 1; }
    ;;
  *)
    error "Usage: $0 <install|ops|release> [VERSION]"
    exit 1
    ;;
esac

PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

require_docker_stack "$COMPOSE_PROD"

DOMAIN="$(read_env JOTTI_DOMAIN)"
BASE_URL="${JOTTI_BASE_URL:-https://$DOMAIN}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Sm0ke-Test-$(date +%s)!}"

info "Mode: $MODE"
info "Base URL: $BASE_URL"

# step_install — prod-init already waits for backend health and HTTPS, so this
# step only adds the OTP parse, set-password and login roundtrip.
step_install() {
  # Recorded before prod-init runs so the OTP grep below only sees this run's
  # log lines: on a non-fresh host bootstrap skips re-issuing a one-time
  # password, and an unscoped grep would pick up a stale, consumed code.
  # Trailing Z matters: `docker compose logs --since` reads a zone-less
  # timestamp as the CLIENT's local time, so on a host ahead of UTC the filter
  # would point into the future and drop the line prod-init just emitted.
  local since
  since="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

  run_step "prod-init" "$SCRIPT_DIR/prod-init.sh"

  local start end duration
  start="$(date +%s)"
  local otp
  otp="$(docker compose -f "$COMPOSE_PROD" logs --since "$since" backend 2>/dev/null \
    | grep -a "ADMIN-EINMALPASSWORT" \
    | grep -aoE 'code=[0-9]{6}' \
    | tail -n1 | cut -d= -f2 || true)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ -z "$otp" ]]; then
    fail_step "parse-admin-otp" "ADMIN-EINMALPASSWORT marker not found in backend logs"
  fi
  ok_step "parse-admin-otp" "$duration"

  start="$(date +%s)"
  local status
  status="$(http_post_status "$BASE_URL/api/auth/set-password" \
    "$(printf '{"username":"admin","password":"%s","onetimePassword":"%s"}' "$ADMIN_PASSWORD" "$otp")")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "set-password" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "set-password" "$duration"

  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/auth/login" \
    "$(printf '{"username":"admin","password":"%s"}' "$ADMIN_PASSWORD")")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "login" "expected 200, got $status: $(redacted_body)"
  fi
  ADMIN_TOKEN="$(json_field token)"
  if [[ -z "$ADMIN_TOKEN" ]]; then
    fail_step "login" "no token in response body"
  fi
  ok_step "login" "$duration"
}

step_ops() {
  run_step "prod-backup" "$SCRIPT_DIR/prod-backup.sh"
  run_step "prod-backup-verify" "$SCRIPT_DIR/prod-backup-verify.sh"
  run_step "prod-update" "$SCRIPT_DIR/prod-update.sh"
}

# step_sale_receipt_export TOKEN — one Direktverkauf, one Kassenbeleg print,
# one DSFinV-K export, all via the (POST-only) API. Needs an open
# Kassensitzung and one active product variant, both created here.
step_sale_receipt_export() {
  local token="$1"
  local auth_header="Authorization: Bearer $token"
  local start end duration status

  # Betreiber-Stammdaten setzen (Voraussetzung für kassensitzung-eroeffnen):
  # auf einem frischen Host ist die betreiber-Tabelle leer, und ohne konfigurierten
  # Betreiber liefert kassensitzung-eroeffnen 400 betreiber_nicht_konfiguriert.
  # Nur die Pflichtfelder (vereinsname, strasse, plz, ort); steuernummer/ustId
  # bleiben optional.
  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/admin/update-betreiber" \
    '{"vereinsname":"Ops-Smoke-Verein","strasse":"Teststraße 1","plz":"12345","ort":"Teststadt"}' "$auth_header")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "update-betreiber" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "update-betreiber" "$duration"

  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/admin/kassensitzung-eroeffnen" \
    '{"bezeichnung":"ops-smoke","betragCents":0}' "$auth_header")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "kassensitzung-eroeffnen" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "kassensitzung-eroeffnen" "$duration"

  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/admin/create-produkt" \
    '{"name":"Ops-Smoke-Produkt","kategorie":"sonstiges","steuersatz":"regel"}' "$auth_header")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "create-produkt" "expected 200, got $status: $(redacted_body)"
  fi
  local produkt_id
  produkt_id="$(json_field id)"
  ok_step "create-produkt" "$duration" "id=$produkt_id"

  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/admin/create-variante" \
    "$(printf '{"produktId":%s,"name":"Standard","preisCents":250}' "$produkt_id")" "$auth_header")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "create-variante" "expected 200, got $status: $(redacted_body)"
  fi
  local variante_id
  variante_id="$(json_field id)"
  ok_step "create-variante" "$duration" "id=$variante_id"

  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/admin/activate-variante" \
    "$(printf '{"id":%s}' "$variante_id")" "$auth_header")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "activate-variante" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "activate-variante" "$duration"

  # Kassenbeleg-Druckstation anlegen (Voraussetzung für beleg-drucken):
  # 192.0.2.1 ist eine TEST-NET-1-Adresse (RFC 5737), also nie ein echter
  # Drucker — der Handler legt bei nicht-leerer druckerIp trotzdem einen
  # Druckauftrag in der DB an, ohne den Drucker tatsächlich zu erreichen.
  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/admin/update-druckstationen" \
    '{"kategorie":"kassenbeleg","druckerIp":"192.0.2.1"}' "$auth_header")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "update-druckstationen" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "update-druckstationen" "$duration"

  local verkauf_id
  verkauf_id="$(cat /proc/sys/kernel/random/uuid 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')"
  start="$(date +%s)"
  status="$(http_post_status "$BASE_URL/api/service/direktverkauf-taetigen" \
    "$(printf '{"verkaufId":"%s","positionen":[{"produktId":%s,"varianteId":%s,"menge":1}],"kommentar":"ops-smoke"}' "$verkauf_id" "$produkt_id" "$variante_id")" "$auth_header")"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "direktverkauf-taetigen" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "direktverkauf-taetigen" "$duration" "verkaufId=$verkauf_id"

  # beleg-drucken meldet "ausstehend" (200, kein Druckauftrag), solange der
  # asynchrone Signatur-Worker die TSE-Signatur nicht quittiert hat. Nur
  # "eingereiht" beweist einen angelegten Druckauftrag, daher darauf pollen.
  local beleg_status="" attempt
  start="$(date +%s)"
  for attempt in $(seq 1 20); do
    status="$(http_post_status "$BASE_URL/api/service/beleg-drucken" \
      "$(printf '{"verkaufId":"%s"}' "$verkauf_id")" "$auth_header")"
    if [[ "$status" != "200" ]]; then
      fail_step "beleg-drucken" "expected 200, got $status: $(redacted_body)"
    fi
    beleg_status="$(json_field status)"
    if [[ "$beleg_status" == "eingereiht" ]]; then
      break
    fi
    sleep 1
  done
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$beleg_status" != "eingereiht" ]]; then
    fail_step "beleg-drucken" "expected status=eingereiht within ${attempt} attempts, last status field: $beleg_status"
  fi
  ok_step "beleg-drucken" "$duration" "attempts=$attempt"

  start="$(date +%s)"
  status="$(curl -sS -o "$SMOKE_TMP/export.zip" -w '%{http_code}' --max-time 15 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d '{"kassensitzungNr":0}' "$BASE_URL/api/admin/export/dsfinvk" 2>"$SMOKE_TMP/curl.log" || echo 000)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "export-dsfinvk" "expected 200, got $status"
  fi
  ok_step "export-dsfinvk" "$duration" "bytes=$(wc -c <"$SMOKE_TMP/export.zip")"
}

step_security_headers() {
  local start end duration headers
  start="$(date +%s)"
  headers="$(curl -sS -D - -o /dev/null --max-time 10 "$BASE_URL/api/health" 2>"$SMOKE_TMP/curl.log" || true)"
  end="$(date +%s)"; duration=$((end - start))

  local missing=""
  for h in "Content-Security-Policy" "Strict-Transport-Security" "X-Frame-Options" "X-Content-Type-Options"; do
    if ! grep -qi "^${h}:" <<<"$headers"; then
      missing="$missing $h"
    fi
  done

  if [[ -n "$missing" ]]; then
    fail_step "security-headers" "missing:$missing"
  fi
  ok_step "security-headers" "$duration"
}

step_login_rate_limit() {
  local start end duration status got_429=false
  start="$(date +%s)"
  for _ in $(seq 1 40); do
    status="$(http_post_status "$BASE_URL/api/auth/login" '{"username":"ops-smoke-does-not-exist","password":"wrong"}')"
    if [[ "$status" == "429" ]]; then
      got_429=true
      break
    fi
  done
  end="$(date +%s)"; duration=$((end - start))

  if [[ "$got_429" != true ]]; then
    fail_step "login-rate-limit" "no 429 seen after 40 attempts (last status: $status)"
  fi
  ok_step "login-rate-limit" "$duration"
}

case "$MODE" in
  install)
    step_install
    step_security_headers
    step_login_rate_limit
    ;;
  ops)
    step_ops
    step_security_headers
    step_login_rate_limit
    ;;
  release)
    # Pinning JOTTI_VERSION mutates the versioned .env; restore the previous
    # value on exit (success or failure) so the smoke run has no lasting
    # side effect on the host's configuration.
    PREVIOUS_JOTTI_VERSION="$(read_env JOTTI_VERSION)"
    restore_jotti_version() {
      # Always rewrite the line, even when the previous value was empty: if
      # there was no JOTTI_VERSION line, the pinning sed matched nothing either.
      sed -i.bak "s/^JOTTI_VERSION=.*/JOTTI_VERSION=$PREVIOUS_JOTTI_VERSION/" .env && rm -f .env.bak
      rm -rf "$SMOKE_TMP"
    }
    trap restore_jotti_version EXIT

    info "Pinning JOTTI_VERSION=$RELEASE_VERSION for the release smoke run..."
    sed -i.bak "s/^JOTTI_VERSION=.*/JOTTI_VERSION=$RELEASE_VERSION/" .env && rm -f .env.bak
    step_install
    step_sale_receipt_export "$ADMIN_TOKEN"
    step_security_headers
    step_login_rate_limit
    ;;
esac

echo "" >&2
info "Ops smoke ($MODE) completed — all steps ok."
