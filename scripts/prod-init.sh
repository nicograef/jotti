#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — First-Deploy Automation (self-hosted production, Weg B)
#
# Reads the domain and email from .env (no hardcoding), validates the host, then
# starts the pinned production stack. Caddy obtains the Let's Encrypt certificate
# automatically (HTTP-01/TLS-ALPN) — no certbot step. Steps:
#   1. Validate prerequisites (Docker, Compose, .env, JOTTI_DOMAIN/LETSENCRYPT_EMAIL/JOTTI_VERSION)
#   2. Check that the domain resolves (and ideally points to this server)
#   3. Pull the pinned images and start the stack
#   4. Wait for the backend to be healthy and verify HTTPS
#
# Usage: ./scripts/prod-init.sh  (or `make prod-init`)
# =============================================================================

COMPOSE_PROD="docker-compose.prod.yml"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

# parse_semver "v1.2.3" — echoes "1 2 3" and returns 0, or returns 1 when the
# value is not a plain vMAJOR.MINOR.PATCH (e.g. "latest", "dev"). A pre-release
# or build suffix ("1.2.3-rc1", "1.2.3+meta") is trimmed before parsing. Mirrors
# core.parseSemver (windows/starter/core/update.go); kept as a standalone copy.
parse_semver() {
  local s="${1#v}"
  s="${s%%[-+]*}"
  [[ "$s" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]] || return 1
  printf '%s %s %s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}"
}

# Minimum length for secrets; mirrors backend/config.MinSecretLength.
MIN_SECRET_LENGTH=16

# validate_secret KEY — read KEY from .env and fatal unless it is set, not a known
# .env.example placeholder, and at least MIN_SECRET_LENGTH chars. Mirrors
# backend/config.ValidateSecrets so a weak secret fails before the stack starts.
validate_secret() {
  local key="$1"
  local value
  value="$(read_env "$key")"
  if [[ -z "$value" ]]; then
    fatal "$key is not set in .env. Run 'make init' to generate secure secrets."
  fi
  case "$value" in
    your-256-bit-secret-replace-this-in-production|your-relay-auth-token-replace-this-in-production|your-secure-password-here|admin)
      fatal "$key still uses the .env.example placeholder value. Run 'make init' to generate a real secret." ;;
  esac
  if (( ${#value} < MIN_SECRET_LENGTH )); then
    fatal "$key is too short (${#value} chars); it needs at least $MIN_SECRET_LENGTH characters."
  fi
}

# ---------------------------------------------------------------------------
# Step 0 — Change to project root (script may be called from anywhere)
# ---------------------------------------------------------------------------
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

info "Project root: $PROJECT_ROOT"

# ---------------------------------------------------------------------------
# Step 1 — Validate prerequisites
# ---------------------------------------------------------------------------
info "Checking prerequisites..."

if [[ ! -f .env ]]; then
  fatal ".env file not found. Run 'make init' first."
fi

if ! command -v docker &>/dev/null; then
  fatal "docker is not installed or not on PATH."
fi

if ! docker compose version &>/dev/null; then
  fatal "docker compose (v2) is not available."
fi

if [[ ! -f "$COMPOSE_PROD" ]]; then
  fatal "Missing compose file: $COMPOSE_PROD"
fi

if ! command -v host &>/dev/null && ! command -v dig &>/dev/null; then
  fatal "Neither 'host' nor 'dig' found. Install one (e.g. dnsutils / bind-tools) for the DNS preflight."
fi

DOMAIN="$(read_env JOTTI_DOMAIN)"
EMAIL="$(read_env LETSENCRYPT_EMAIL)"
VERSION="$(read_env JOTTI_VERSION)"

if [[ -z "$DOMAIN" ]]; then
  fatal "JOTTI_DOMAIN is not set in .env. Set it to your public domain (e.g. jotti.meinverein.de)."
fi
if [[ -z "$EMAIL" ]]; then
  fatal "LETSENCRYPT_EMAIL is not set in .env. Set it to a contact email for the Let's Encrypt account."
fi
# Only a pinned release tag (vMAJOR.MINOR.PATCH) is accepted; "latest" or an
# empty value would silently pull a moving image. Compose references the tag as a
# bare ${JOTTI_VERSION} with no default, so an empty value aborts the stack; this
# check turns that into an early, actionable error before anything is pulled.
if ! parse_semver "$VERSION" >/dev/null; then
  error "JOTTI_VERSION in .env is not a pinned release tag (found: '${VERSION:-<empty>}')."
  error "Set it to a release tag like v0.3.1 from https://github.com/nicograef/jotti/releases."
  fatal "Refusing to deploy an unpinned version ('latest' and empty are not allowed)."
fi

# Reject weak or placeholder secrets before starting the stack (mirrors the
# backend startup validation; a known secret means a full auth bypass).
validate_secret JWT_SECRET
validate_secret RELAY_AUTH_TOKEN
validate_secret POSTGRES_PASSWORD

info "Prerequisites OK (domain: $DOMAIN, email: $EMAIL, version: $VERSION)."

# ---------------------------------------------------------------------------
# Step 2 — Check DNS resolution and that it points to this server
# ---------------------------------------------------------------------------
info "Checking DNS resolution for $DOMAIN..."

resolve_a() {
  if command -v dig &>/dev/null; then
    dig +short A "$1" 2>/dev/null | tail -n1
  else
    host -t A "$1" 2>/dev/null | awk '/has address/ { print $NF; exit }'
  fi
}

DOMAIN_IP="$(resolve_a "$DOMAIN")"
if [[ -z "$DOMAIN_IP" ]]; then
  fatal "DNS resolution failed for $DOMAIN. Point the domain at this server before continuing."
fi
info "DNS resolution for $DOMAIN: OK ($DOMAIN_IP)"

# Best-effort: warn (not fatal) if the domain does not resolve to this server's
# public IP. Detecting the own public IP needs an outbound call and may fail.
SERVER_IP="$(curl -fsS --max-time 5 https://api.ipify.org 2>/dev/null || true)"
if [[ -n "$SERVER_IP" && "$SERVER_IP" != "$DOMAIN_IP" ]]; then
  warn "$DOMAIN resolves to $DOMAIN_IP but this server appears to be $SERVER_IP."
  warn "Let's Encrypt issuance will fail unless the domain points to this server."
fi

# ---------------------------------------------------------------------------
# Step 3 — Pull pinned images and start the stack
# ---------------------------------------------------------------------------
info "Pulling pinned images..."
docker compose -f "$COMPOSE_PROD" pull

info "Starting production stack..."
docker compose -f "$COMPOSE_PROD" up -d

# ---------------------------------------------------------------------------
# Step 4 — Wait for the backend, then verify HTTPS
# ---------------------------------------------------------------------------
info "Waiting for the backend to become healthy..."
backend_healthy=false
for _ in $(seq 1 30); do
  status="$(docker inspect -f '{{.State.Health.Status}}' jotti-backend 2>/dev/null || echo unknown)"
  if [[ "$status" == "healthy" ]]; then
    backend_healthy=true
    break
  fi
  sleep 2
done

if [[ "$backend_healthy" != true ]]; then
  error "Backend did not become healthy in time."
  fatal "Check logs with: docker compose -f $COMPOSE_PROD logs -f"
fi
info "Backend healthy."

info "Waiting for HTTPS / certificate issuance (this can take a moment)..."
https_ok=false
for _ in $(seq 1 30); do
  code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "https://$DOMAIN/api/health" 2>/dev/null || echo 000)"
  if [[ "$code" == "200" ]]; then
    https_ok=true
    break
  fi
  sleep 3
done

# HTTP→HTTPS redirect check (informational).
HTTP_STATUS="$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "http://$DOMAIN" 2>/dev/null || echo 000)"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "=========================================="
printf "${GREEN} %s${NC}\n" "jotti — Deployment Complete"
echo "=========================================="
echo ""
echo "  Domain:  https://$DOMAIN"
if [[ "$https_ok" == true ]]; then
  info "HTTPS check: OK (/api/health returned 200)"
else
  warn "HTTPS did not return 200 yet — the certificate may still be issuing."
  warn "Re-check in a minute, or follow logs: docker compose -f $COMPOSE_PROD logs -f reverse-proxy"
fi
if [[ "$HTTP_STATUS" == "308" || "$HTTP_STATUS" == "301" || "$HTTP_STATUS" == "302" ]]; then
  info "HTTP→HTTPS redirect: OK (HTTP $HTTP_STATUS)"
else
  warn "HTTP→HTTPS redirect returned HTTP $HTTP_STATUS"
fi
echo ""
echo "  Useful commands:"
echo "    make prod-up     — Restart the stack"
echo "    make prod-down   — Stop all services"
echo "    make prod-logs   — Follow logs"
echo ""
# First-time setup: grep the admin one-time login code straight from the backend
# logs (analogous to the Windows starter). ANSI-tolerant: match the marker
# substring and extract the 6-digit code; the newest match wins.
otp_code="$(docker compose -f "$COMPOSE_PROD" logs backend 2>/dev/null \
  | grep -a "ADMIN-EINMALPASSWORT" \
  | grep -aoE 'code=[0-9]{6}' \
  | tail -n1 | cut -d= -f2 || true)"
if [[ -n "$otp_code" ]]; then
  info "First-time setup — admin one-time login code (user 'admin'): $otp_code"
  echo "    Log in as 'admin' with this code, then set your own password."
else
  warn "No admin one-time login code found in the logs yet (setup may be complete, or the backend just started)."
  echo "    Re-check with: docker compose -f $COMPOSE_PROD logs backend | grep ADMIN-EINMALPASSWORT"
fi
echo ""
echo "  Caddy renews the certificate automatically."
echo "=========================================="
