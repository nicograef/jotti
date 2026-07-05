#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — First-Deploy Automation (self-hosted production, Weg B)
#
# Reads the domain and email from .env (no hardcoding), validates the host, then
# starts the pinned production stack. Caddy obtains the Let's Encrypt certificate
# automatically (HTTP-01/TLS-ALPN) — no certbot step. Steps:
#   1. Validate prerequisites (Docker, Compose, .env, JOTTI_DOMAIN/LETSENCRYPT_EMAIL)
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
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

# read_env KEY — read a single value from .env without executing the file
# (passwords may contain shell-special characters). Returns the last match,
# trimmed of surrounding whitespace.
read_env() {
  local key="$1"
  { grep -E "^${key}=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
}

# ---------------------------------------------------------------------------
# Step 0 — Change to project root (script may be called from anywhere)
# ---------------------------------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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

if [[ -z "$DOMAIN" ]]; then
  fatal "JOTTI_DOMAIN is not set in .env. Set it to your public domain (e.g. jotti.meinverein.de)."
fi
if [[ -z "$EMAIL" ]]; then
  fatal "LETSENCRYPT_EMAIL is not set in .env. Set it to a contact email for the Let's Encrypt account."
fi

info "Prerequisites OK (domain: $DOMAIN, email: $EMAIL)."

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
echo "    make prod-up     — Pull & restart"
echo "    make prod-down   — Stop all services"
echo "    make prod-logs   — Follow logs"
echo ""
echo "  Caddy renews the certificate automatically."
echo "=========================================="
