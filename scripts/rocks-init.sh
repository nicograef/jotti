#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# jotti.rocks — First-Deploy Script for the project website
#
# Deploys the jotti.rocks setup:
#   - https://jotti.rocks        → static landing page
#   - https://demo.jotti.rocks   → demo app (frontend + backend API)
#   - https://auth.jotti.rocks   → acme-dns API (trusted local TLS)
#
# Uses docker-compose.prod.yml + docker-compose.rocks.yml override.
# Self-hosters should use scripts/prod-init.sh instead.
#
# Usage: ./scripts/rocks-init.sh
# =============================================================================

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
DOMAIN="jotti.rocks"
DOMAIN_WWW="www.jotti.rocks"
DOMAIN_DEMO="demo.jotti.rocks"
DOMAIN_AUTH="auth.jotti.rocks"
EMAIL="graef.nico@gmail.com"

COMPOSE_CERT="docker-compose.initial-cert.yml"
COMPOSE_PROD=(-f docker-compose.prod.yml -f docker-compose.rocks.yml)

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

# ---------------------------------------------------------------------------
# Step 0 — Change to project root
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

if ! grep -qE '^VPS_PUBLIC_IP=.+' .env; then
  fatal "VPS_PUBLIC_IP missing or empty in .env (public IPv4 of this server, needed by resolver + acme-dns). See docs/betrieb/leitfaden-rocks-dns.md."
fi

if ! command -v docker &>/dev/null; then
  fatal "docker is not installed or not on PATH."
fi

if ! docker compose version &>/dev/null; then
  fatal "docker compose (v2) is not available."
fi

if [[ ! -f "$COMPOSE_CERT" ]]; then
  fatal "Missing compose file: $COMPOSE_CERT"
fi
if [[ ! -f docker-compose.prod.yml ]]; then
  fatal "Missing compose file: docker-compose.prod.yml"
fi
if [[ ! -f docker-compose.rocks.yml ]]; then
  fatal "Missing compose file: docker-compose.rocks.yml"
fi

# A DNS lookup tool is required for the resolution preflight below; without one,
# that check would misreport a missing tool as a DNS failure.
if ! command -v host &>/dev/null && ! command -v dig &>/dev/null; then
  fatal "Neither 'host' nor 'dig' found. Install one (e.g. dnsutils / bind-tools) for the DNS preflight."
fi

info "Prerequisites OK."

# ---------------------------------------------------------------------------
# Step 2 — Check DNS resolution
# ---------------------------------------------------------------------------
info "Checking DNS resolution for $DOMAIN..."

if ! host "$DOMAIN" &>/dev/null && ! dig +short "$DOMAIN" 2>/dev/null | grep -q .; then
  fatal "DNS resolution failed for $DOMAIN. Ensure the domain points to this server before continuing."
fi

info "DNS resolution for $DOMAIN: OK"

CERTBOT_DOMAINS="-d $DOMAIN"

if host "$DOMAIN_WWW" &>/dev/null || dig +short "$DOMAIN_WWW" 2>/dev/null | grep -q .; then
  info "DNS resolution for $DOMAIN_WWW: OK"
  CERTBOT_DOMAINS="$CERTBOT_DOMAINS -d $DOMAIN_WWW"
else
  warn "DNS resolution for $DOMAIN_WWW failed. www will not be included in the certificate."
fi

if host "$DOMAIN_DEMO" &>/dev/null || dig +short "$DOMAIN_DEMO" 2>/dev/null | grep -q .; then
  info "DNS resolution for $DOMAIN_DEMO: OK"
  CERTBOT_DOMAINS="$CERTBOT_DOMAINS -d $DOMAIN_DEMO"
else
  warn "DNS resolution for $DOMAIN_DEMO failed. demo will not be included in the certificate."
fi

# auth.jotti.rocks resolves via the resolver/acme-dns stack on this server —
# on a fresh install it only works once the stack is up and the NS delegation
# is set. Expand the certificate later as described in the guide.
if host "$DOMAIN_AUTH" &>/dev/null || dig +short "$DOMAIN_AUTH" 2>/dev/null | grep -q .; then
  info "DNS resolution for $DOMAIN_AUTH: OK"
  CERTBOT_DOMAINS="$CERTBOT_DOMAINS -d $DOMAIN_AUTH"
else
  warn "DNS resolution for $DOMAIN_AUTH failed. auth will not be included in the certificate."
  warn "Expand the certificate after the stack is up — see docs/betrieb/leitfaden-rocks-dns.md."
fi

# ---------------------------------------------------------------------------
# Step 3 — Request initial Let's Encrypt certificate
# ---------------------------------------------------------------------------
info "Starting nginx for ACME challenge..."
docker compose -f "$COMPOSE_CERT" up -d reverse-proxy

sleep 3

info "Requesting certificate from Let's Encrypt..."

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
# Step 5 — Start full production stack with jotti.rocks override
# ---------------------------------------------------------------------------
info "Building and starting production stack..."
docker compose "${COMPOSE_PROD[@]}" up -d --build

info "Waiting for services to start..."
sleep 10

# ---------------------------------------------------------------------------
# Step 6 — Verify deployment
# ---------------------------------------------------------------------------
info "Verifying deployment..."

# Check landing page
HTTPS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "https://$DOMAIN" 2>/dev/null || echo "000")

if [[ "$HTTPS_STATUS" == "200" || "$HTTPS_STATUS" == "301" || "$HTTPS_STATUS" == "302" ]]; then
  info "Landing page HTTPS check: OK (HTTP $HTTPS_STATUS)"
else
  warn "Landing page HTTPS check returned HTTP $HTTPS_STATUS — may not be fully ready yet."
fi

# Check demo app
DEMO_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "https://$DOMAIN_DEMO" 2>/dev/null || echo "000")

if [[ "$DEMO_STATUS" == "200" || "$DEMO_STATUS" == "301" || "$DEMO_STATUS" == "302" ]]; then
  info "Demo app HTTPS check: OK (HTTP $DEMO_STATUS)"
else
  warn "Demo app HTTPS check returned HTTP $DEMO_STATUS — may not be fully ready yet."
fi

# Check acme-dns API
AUTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 10 "https://$DOMAIN_AUTH/health" 2>/dev/null || echo "000")

if [[ "$AUTH_STATUS" == "200" ]]; then
  info "acme-dns API HTTPS check: OK (HTTP $AUTH_STATUS)"
else
  warn "acme-dns API HTTPS check returned HTTP $AUTH_STATUS — expected if auth.jotti.rocks is not yet in the certificate (see docs/betrieb/leitfaden-rocks-dns.md)."
fi

# Check HTTP→HTTPS redirect
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
printf "${GREEN} jotti.rocks — Deployment Complete${NC}\n"
echo "=========================================="
echo ""
echo "  Landing page:  https://$DOMAIN"
echo "  Demo app:      https://$DOMAIN_DEMO"
echo "  acme-dns API:  https://$DOMAIN_AUTH"
echo ""
echo "  Useful commands:"
echo "    make rocks-up     — Rebuild & restart"
echo "    make rocks-down   — Stop all services"
echo "    make rocks-logs   — Follow logs"
echo ""
echo "  Certificates renew automatically every 24h."
echo "=========================================="
