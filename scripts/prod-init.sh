#!/usr/bin/env bash
# prod-init.sh — first deploy of the self-hosted production stack
#
# Usage:
#   make prod-init   # reads JOTTI_DOMAIN, LETSENCRYPT_EMAIL, JOTTI_VERSION from .env
#
# What it does:
#   1. Validates .env: a pinned version and strong, non-placeholder secrets.
#   2. Checks that the domain resolves to this server.
#   3. Pulls and starts the pinned stack and waits for backend health and HTTPS;
#      Caddy obtains the Let's Encrypt certificate itself (HTTP-01/TLS-ALPN).
#   4. Prints the admin one-time login code from the backend logs.
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
PROD_BACKEND_CONTAINER="${PROD_BACKEND_CONTAINER:-jotti-backend}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"

# Minimum length for secrets; mirrors backend/config.MinSecretLength.
MIN_SECRET_LENGTH=16

# validate_secret KEY — mirrors backend/config.ValidateSecrets, so a weak or
# placeholder secret fails before the stack starts.
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

PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

info "Project root: $PROJECT_ROOT"

info "Checking prerequisites..."

require_docker_stack "$COMPOSE_FILE"

if ! command -v host &>/dev/null && ! command -v dig &>/dev/null; then
  fatal "Neither 'host' nor 'dig' found. Install one (e.g. dnsutils / bind-tools) for the DNS preflight."
fi

DOMAIN="$(read_env JOTTI_DOMAIN)"
EMAIL="$(read_env LETSENCRYPT_EMAIL)"

if [[ -z "$DOMAIN" ]]; then
  fatal "JOTTI_DOMAIN is not set in .env. Set it to your public domain (e.g. jotti.meinverein.de)."
fi
if [[ -z "$EMAIL" ]]; then
  fatal "LETSENCRYPT_EMAIL is not set in .env. Set it to a contact email for the Let's Encrypt account."
fi
VERSION="$(require_pinned_version deploy)"

# A known or placeholder secret means a full auth bypass.
validate_secret JWT_SECRET
validate_secret RELAY_AUTH_TOKEN
validate_secret POSTGRES_PASSWORD

info "Prerequisites OK (domain: $DOMAIN, email: $EMAIL, version: $VERSION)."

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

# Best-effort: detecting the own public IP needs an outbound call, so a mismatch
# warns instead of aborting.
SERVER_IP="$(curl -fsS --max-time 5 https://api.ipify.org 2>/dev/null || true)"
if [[ -n "$SERVER_IP" && "$SERVER_IP" != "$DOMAIN_IP" ]]; then
  warn "$DOMAIN resolves to $DOMAIN_IP but this server appears to be $SERVER_IP."
  warn "Let's Encrypt issuance will fail unless the domain points to this server."
fi

info "Pulling pinned images..."
docker compose -f "$COMPOSE_FILE" pull

info "Starting production stack..."
docker compose -f "$COMPOSE_FILE" up -d

info "Waiting for the backend to become healthy..."
if ! wait_for_healthy "$PROD_BACKEND_CONTAINER"; then
  error "Backend did not become healthy in time."
  fatal "Check logs with: docker compose -f $COMPOSE_FILE logs -f"
fi
info "Backend healthy."

info "Waiting for HTTPS / certificate issuance (this can take a moment)..."
https_ok=false
wait_https_ok "https://$DOMAIN/api/health" 30 && https_ok=true

HTTP_STATUS="$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "http://$DOMAIN" 2>/dev/null || echo 000)"

echo ""
echo "=========================================="
printf "${GREEN} %s${NC}\n" "jotti — Deployment Complete"
echo "=========================================="
echo ""
echo "  Domain:  https://$DOMAIN"
if [[ "$https_ok" == true ]]; then
  info "HTTPS check: OK (/api/health answered)"
else
  warn "HTTPS did not return 200 yet — the certificate may still be issuing."
  warn "Re-check in a minute, or follow logs: docker compose -f $COMPOSE_FILE logs -f reverse-proxy"
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
# Admin one-time login code from the backend logs (like the Windows starter).
otp_code="$(admin_otp_from_logs "$COMPOSE_FILE")"
if [[ -n "$otp_code" ]]; then
  info "First-time setup — admin one-time login code (user 'admin'): $otp_code"
  echo "    Log in as 'admin' with this code, then set your own password."
else
  warn "No admin one-time login code found in the logs yet (setup may be complete, or the backend just started)."
  echo "    Re-check with: docker compose -f $COMPOSE_FILE logs backend | grep ADMIN-EINMALPASSWORT"
fi
echo ""
echo "  Caddy renews the certificate automatically."
echo "=========================================="
