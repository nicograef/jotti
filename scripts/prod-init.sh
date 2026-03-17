#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti — First-Deploy Automation Script
#
# Performs the initial production deployment:
# 1. Validates prerequisites (.env, DNS, Docker)
# 2. Requests the first Let's Encrypt certificate via HTTP-01 challenge
# 3. Starts the full production stack
# 4. Verifies HTTPS is working
#
# Usage: ./scripts/prod-init.sh
# =============================================================================

# ---------------------------------------------------------------------------
# Configuration — adjust these if the domain or email changes
# ---------------------------------------------------------------------------
DOMAIN="jotti.rocks"
DOMAIN_WWW="www.jotti.rocks"
DOMAIN_DEMO="demo.jotti.rocks"
EMAIL="graef.nico@gmail.com"

COMPOSE_CERT="docker-compose.initial-cert.yml"
COMPOSE_PROD="docker-compose.prod.yml"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info()  { printf "${GREEN}[INFO]${NC}  %s\n" "$1"; }
warn()  { printf "${YELLOW}[WARN]${NC}  %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1" >&2; }
fatal() { error "$1"; exit 1; }

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

# .env must exist
if [[ ! -f .env ]]; then
  fatal ".env file not found. Copy .env.example to .env and configure it first."
fi

# Docker must be available
if ! command -v docker &>/dev/null; then
  fatal "docker is not installed or not on PATH."
fi

# Docker Compose must be available (v2 plugin)
if ! docker compose version &>/dev/null; then
  fatal "docker compose (v2) is not available."
fi

# Compose files must exist
if [[ ! -f "$COMPOSE_CERT" ]]; then
  fatal "Missing compose file: $COMPOSE_CERT"
fi
if [[ ! -f "$COMPOSE_PROD" ]]; then
  fatal "Missing compose file: $COMPOSE_PROD"
fi

info "Prerequisites OK."

# ---------------------------------------------------------------------------
# Step 2 — Check DNS resolution
# ---------------------------------------------------------------------------
info "Checking DNS resolution for $DOMAIN..."

if ! host "$DOMAIN" &>/dev/null 2>&1 && ! dig +short "$DOMAIN" 2>/dev/null | grep -q .; then
  fatal "DNS resolution failed for $DOMAIN. Ensure the domain points to this server before continuing."
fi

info "DNS resolution for $DOMAIN: OK"

if host "$DOMAIN_WWW" &>/dev/null 2>&1 || dig +short "$DOMAIN_WWW" 2>/dev/null | grep -q .; then
  info "DNS resolution for $DOMAIN_WWW: OK"
else
  warn "DNS resolution for $DOMAIN_WWW failed. www subdomain will not be included in the certificate."
  DOMAIN_WWW=""
fi

if host "$DOMAIN_DEMO" &>/dev/null 2>&1 || dig +short "$DOMAIN_DEMO" 2>/dev/null | grep -q .; then
  info "DNS resolution for $DOMAIN_DEMO: OK"
else
  warn "DNS resolution for $DOMAIN_DEMO failed. demo subdomain will not be included in the certificate."
  DOMAIN_DEMO=""
fi

# ---------------------------------------------------------------------------
# Step 3 — Request initial Let's Encrypt certificate
# ---------------------------------------------------------------------------
info "Starting nginx for ACME challenge..."
docker compose -f "$COMPOSE_CERT" up -d reverse-proxy

# Wait for nginx to be ready
sleep 3

info "Requesting certificate from Let's Encrypt..."

CERTBOT_DOMAINS="-d $DOMAIN"
if [[ -n "$DOMAIN_WWW" ]]; then
  CERTBOT_DOMAINS="$CERTBOT_DOMAINS -d $DOMAIN_WWW"
fi
if [[ -n "$DOMAIN_DEMO" ]]; then
  CERTBOT_DOMAINS="$CERTBOT_DOMAINS -d $DOMAIN_DEMO"
fi

if ! docker compose -f "$COMPOSE_CERT" run --rm --entrypoint certbot certbot certonly \
  --webroot -w /var/www/certbot \
  $CERTBOT_DOMAINS \
  --email "$EMAIL" --agree-tos --no-eff-email; then
  error "Certbot failed. Cleaning up..."
  docker compose -f "$COMPOSE_CERT" down
  fatal "Certificate request failed. Check the output above for details."
fi

info "Certificate issued successfully."

# ---------------------------------------------------------------------------
# Step 4 — Stop initial-cert stack
# ---------------------------------------------------------------------------
info "Stopping initial certificate stack..."
docker compose -f "$COMPOSE_CERT" down

# ---------------------------------------------------------------------------
# Step 5 — Start full production stack
# ---------------------------------------------------------------------------
info "Building and starting production stack..."
docker compose -f "$COMPOSE_PROD" up -d --build

# Wait for services to come up
info "Waiting for services to start..."
sleep 10

# ---------------------------------------------------------------------------
# Step 6 — Verify deployment
# ---------------------------------------------------------------------------
info "Verifying deployment..."

HTTPS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "https://$DOMAIN" 2>/dev/null || echo "000")

if [[ "$HTTPS_STATUS" == "200" || "$HTTPS_STATUS" == "301" || "$HTTPS_STATUS" == "302" ]]; then
  info "HTTPS check: OK (HTTP $HTTPS_STATUS)"
else
  warn "HTTPS check returned HTTP $HTTPS_STATUS — the site may not be fully ready yet."
  warn "Check logs with: docker compose logs -f"
fi

# Verify HTTP→HTTPS redirect
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "http://$DOMAIN" 2>/dev/null || echo "000")

if [[ "$HTTP_STATUS" == "301" ]]; then
  info "HTTP→HTTPS redirect: OK"
else
  warn "HTTP→HTTPS redirect returned HTTP $HTTP_STATUS (expected 301)"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "=========================================="
printf "${GREEN} jotti — Deployment Complete${NC}\n"
echo "=========================================="
echo ""
echo "  Landing:  https://$DOMAIN"
echo "  App:      https://demo.$DOMAIN"
echo "  Status:   HTTPS $HTTPS_STATUS"
echo ""
echo "  Useful commands:"
echo "    make prod-up     — Rebuild & restart"
echo "    make prod-down   — Stop all services"
echo "    make prod-logs   — Follow logs"
echo ""
echo "  Certificates renew automatically every 24h."
echo "=========================================="
