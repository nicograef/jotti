#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — Ops Smoke Test (self-hosted production, Weg B)
#
# Drives the orchestrated production scripts (prod-init.sh, prod-update.sh,
# prod-backup.sh, prod-backup-verify.sh) end-to-end and machine-readably logs
# every step, so the installation/update path is verified repeatedly instead
# of once by hand. No real run happens in this phase (a throwaway host run
# follows later) — this script is reviewed statically here.
#
# Modes:
#   install   prod-init, then set-password with the parsed ADMIN-EINMALPASSWORT
#             one-time code, then login — proves the first-boot roundtrip.
#             Requires a FRESH host (no admin password set yet): prod-init only
#             issues a new one-time password on first bootstrap, so a rerun
#             against an already-initialized host fails at parse-admin-otp.
#   ops       prod-backup, prod-backup-verify, then a prod-update roundtrip
#             (re-applies the pinned JOTTI_VERSION) against a running stack.
#   release   like install, plus one Direktverkauf, one Kassenbeleg and one
#             DSFinV-K-Export via the API, against a pinned VERSION argument.
#             Also requires a FRESH host, for the same reason as "install".
#             Configures a dummy Kassenbeleg-Druckstation (TEST-NET-1 IP, never
#             a real printer) so beleg-drucken can enqueue a Druckauftrag, and
#             polls it until the async TSE-Signatur-Worker reports "eingereiht"
#             (an "ausstehend" 200 alone does not prove a receipt was queued).
#
# Every mode additionally checks the reverse proxy in front of the deployed
# stack: security headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options)
# and the login rate limit (429 after repeated bad logins).
#
# Host provisioning and the TLS/certificate acceptance stay manual (see
# docs/leitfaden.md); this script only drives the already-provisioned host.
#
# NEVER runs prod-restore.sh, `docker compose down -v`, or deletes volumes —
# no destructive step is part of any mode.
#
# Machine-readable log: one TSV line per step on stdout —
#   <unix_ts>\t<step>\t<status>\t<duration_seconds>\t<detail>
# status is one of: ok, fail. The script aborts (set -e discipline: every
# failing step calls `fail_step` which exits 1) at the first failed step.
#
# Configuration:
#   JOTTI_BASE_URL   base URL of the deployed stack (default: https://$JOTTI_DOMAIN,
#                    JOTTI_DOMAIN read from .env)
#   ADMIN_PASSWORD   password to set for the initial admin during "install"/
#                    "release" (default: a generated throwaway password)
#   LOG_FILE         optional path to also append the TSV log to (default: none,
#                    stdout only)
#
# Usage:
#   ./scripts/ops-smoke.sh install             # needs a fresh host
#   ./scripts/ops-smoke.sh ops
#   ./scripts/ops-smoke.sh release VERSION      # e.g. v0.14.0, needs a fresh host
# =============================================================================

COMPOSE_PROD="docker-compose.prod.yml"

# Private (0700, mktemp default) scratch dir for step logs, HTTP bodies and
# the admin JWT — these can contain the admin OTP, password or a valid token,
# so a world-readable fixed /tmp path is avoided. Removed on exit.
SMOKE_TMP="$(mktemp -d "${TMPDIR:-/tmp}/ops-smoke.XXXXXX")"
trap 'rm -rf "$SMOKE_TMP"' EXIT

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

# log_line STEP STATUS DURATION DETAIL — emits one TSV line (and appends to
# LOG_FILE if set). This is the single source of the machine-readable protocol.
log_line() {
  local step="$1" status="$2" duration="$3" detail="${4:-}"
  local line
  line="$(printf '%s\t%s\t%s\t%s\t%s' "$(date +%s)" "$step" "$status" "$duration" "$detail")"
  echo "$line"
  if [[ -n "${LOG_FILE:-}" ]]; then
    echo "$line" >>"$LOG_FILE"
  fi
}

# run_step STEP CMD... — runs CMD, times it, logs ok/fail, and aborts the whole
# script on failure (no destructive cleanup is ever needed: the steps here
# never touch volumes or run prod-restore.sh).
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

# fail_step STEP DETAIL — logs a fail line for a check done inline (not via
# run_step, e.g. an HTTP assertion) and aborts.
fail_step() {
  local step="$1" detail="${2:-}"
  log_line "$step" fail 0 "$detail"
  error "Step '$step' failed: $detail"
  exit 1
}

# ok_step STEP DURATION DETAIL — logs a successful inline check.
ok_step() {
  local step="$1" duration="$2" detail="${3:-}"
  log_line "$step" ok "$duration" "$detail"
}

# http_post_status URL DATA — POST-only API (see AGENTS.md); returns the HTTP
# status code, body written to $SMOKE_TMP/body.json. Every caller passes the
# request body explicitly (empty payloads as '{}').
http_post_status() {
  local url="$1" data="$2"
  curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -d "$data" "$url" 2>"$SMOKE_TMP/curl.log" || echo 000
}

# json_field FIELD — extracts a top-level string/number field from
# $SMOKE_TMP/body.json without a jq dependency (values here are always
# simple scalars: token, id, zNr).
json_field() {
  local field="$1"
  grep -oE "\"${field}\"[[:space:]]*:[[:space:]]*\"?[^,}\"]*\"?" "$SMOKE_TMP/body.json" \
    | head -n1 | sed -E "s/\"${field}\"[[:space:]]*:[[:space:]]*//; s/^\"//; s/\"\$//"
}

# redacted_body — dumps $SMOKE_TMP/body.json for a failure message with any
# "token" field masked, so a JWT never lands in stderr/CI logs (e.g. if the
# login step gets a non-200 status but the body still echoes a token field).
redacted_body() {
  sed -E 's/("token"[[:space:]]*:[[:space:]]*)"[^"]*"/\1"[redacted]"/' "$SMOKE_TMP/body.json" 2>/dev/null
}

# ---------------------------------------------------------------------------
# Step 0 — Arguments, project root, prerequisites
# ---------------------------------------------------------------------------
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

if ! command -v docker &>/dev/null; then
  error "docker is not installed or not on PATH."
  exit 1
fi
if ! docker compose version &>/dev/null; then
  error "docker compose (v2) is not available."
  exit 1
fi
if [[ ! -f "$COMPOSE_PROD" ]]; then
  error "Missing compose file: $COMPOSE_PROD"
  exit 1
fi
if [[ ! -f .env ]]; then
  error ".env file not found. Run 'make init' first."
  exit 1
fi

DOMAIN="$(read_env JOTTI_DOMAIN)"
BASE_URL="${JOTTI_BASE_URL:-https://$DOMAIN}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Sm0ke-Test-$(date +%s)!}"

info "Mode: $MODE"
info "Base URL: $BASE_URL"

# ---------------------------------------------------------------------------
# Reusable blocks
# ---------------------------------------------------------------------------

# step_install — prod-init, set-password with the parsed OTP, then login.
# prod-init.sh already waits for the backend health check and HTTPS itself, so
# this step only needs to add the login roundtrip on top.
step_install() {
  # Recorded before prod-init runs so the OTP grep below only sees log lines
  # from this run: on a non-fresh host, bootstrap skips re-issuing a one-time
  # password (ActionSkip), and an unscoped grep would otherwise pick up a
  # stale, already-consumed code from a previous run.
  # Trailing Z: docker compose logs --since interprets a zone-less timestamp
  # as the CLIENT's local time, not UTC. On a host with TZ ahead of UTC (e.g.
  # Europe/Berlin, the usual setup for German VPS), an unzoned value would
  # make --since point 1-2h into the future (in UTC), filtering out the
  # ADMIN-EINMALPASSWORT line prod-init just emitted.
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

# step_ops — prod-backup, prod-backup-verify, prod-update roundtrip.
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
  status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d '{"vereinsname":"Ops-Smoke-Verein","strasse":"Teststraße 1","plz":"12345","ort":"Teststadt"}' "$BASE_URL/api/admin/update-betreiber" 2>"$SMOKE_TMP/curl.log" || echo 000)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "update-betreiber" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "update-betreiber" "$duration"

  start="$(date +%s)"
  status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d '{"bezeichnung":"ops-smoke","betragCents":0}' "$BASE_URL/api/admin/kassensitzung-eroeffnen" 2>"$SMOKE_TMP/curl.log" || echo 000)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "kassensitzung-eroeffnen" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "kassensitzung-eroeffnen" "$duration"

  start="$(date +%s)"
  status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d '{"name":"Ops-Smoke-Produkt","kategorie":"sonstiges","steuersatz":"regel"}' "$BASE_URL/api/admin/create-produkt" 2>"$SMOKE_TMP/curl.log" || echo 000)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "create-produkt" "expected 200, got $status: $(redacted_body)"
  fi
  local produkt_id
  produkt_id="$(json_field id)"
  ok_step "create-produkt" "$duration" "id=$produkt_id"

  start="$(date +%s)"
  status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d "$(printf '{"produktId":%s,"name":"Standard","preisCents":250}' "$produkt_id")" "$BASE_URL/api/admin/create-variante" 2>"$SMOKE_TMP/curl.log" || echo 000)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "create-variante" "expected 200, got $status: $(redacted_body)"
  fi
  local variante_id
  variante_id="$(json_field id)"
  ok_step "create-variante" "$duration" "id=$variante_id"

  start="$(date +%s)"
  status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d "$(printf '{"id":%s}' "$variante_id")" "$BASE_URL/api/admin/activate-variante" 2>"$SMOKE_TMP/curl.log" || echo 000)"
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
  status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d '{"kategorie":"kassenbeleg","druckerIp":"192.0.2.1"}' "$BASE_URL/api/admin/update-druckstationen" 2>"$SMOKE_TMP/curl.log" || echo 000)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "update-druckstationen" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "update-druckstationen" "$duration"

  local verkauf_id
  verkauf_id="$(cat /proc/sys/kernel/random/uuid 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')"
  start="$(date +%s)"
  status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
    -X POST -H 'Content-Type: application/json' -H "$auth_header" \
    -d "$(printf '{"verkaufId":"%s","positionen":[{"produktId":%s,"varianteId":%s,"menge":1}],"kommentar":"ops-smoke"}' "$verkauf_id" "$produkt_id" "$variante_id")" \
    "$BASE_URL/api/service/direktverkauf-taetigen" 2>"$SMOKE_TMP/curl.log" || echo 000)"
  end="$(date +%s)"; duration=$((end - start))
  if [[ "$status" != "200" ]]; then
    fail_step "direktverkauf-taetigen" "expected 200, got $status: $(redacted_body)"
  fi
  ok_step "direktverkauf-taetigen" "$duration" "verkaufId=$verkauf_id"

  # beleg-drucken meldet "ausstehend" (200, kein Druckauftrag), solange die
  # TSE-Signatur des Verkaufs noch nicht vom asynchronen Signatur-Worker
  # quittiert wurde; die UI ruft in diesem Fall denselben Endpunkt erneut auf.
  # Nur "eingereiht" beweist einen tatsächlich angelegten Druckauftrag, daher
  # hier auf "eingereiht" pollen statt ein beliebiges 200 zu akzeptieren.
  local beleg_status="" attempt
  start="$(date +%s)"
  for attempt in $(seq 1 20); do
    status="$(curl -sS -o "$SMOKE_TMP/body.json" -w '%{http_code}' --max-time 10 \
      -X POST -H 'Content-Type: application/json' -H "$auth_header" \
      -d "$(printf '{"verkaufId":"%s"}' "$verkauf_id")" "$BASE_URL/api/service/beleg-drucken" 2>"$SMOKE_TMP/curl.log" || echo 000)"
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

# step_security_headers — CSP, HSTS, X-Frame-Options, X-Content-Type-Options
# on the deployed reverse proxy (Caddy).
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

# step_login_rate_limit — hammers /api/auth/login with bad credentials until
# the reverse proxy (or the app-level throttle) answers 429.
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

# ---------------------------------------------------------------------------
# Run the selected mode
# ---------------------------------------------------------------------------
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
      # Always rewrite the line, even when the previous value was empty: an
      # unconditional restore keeps no lasting release pin in .env. If there was
      # no JOTTI_VERSION line to begin with, the pinning sed was a no-op too, so
      # this substitution simply matches nothing.
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
