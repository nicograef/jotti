#!/usr/bin/env bash
set -euo pipefail

# jotti.rocks — first deploy of the project website stack
# (docker-compose.rocks.yml). Self-hosters use scripts/prod-init.sh instead.
#   https://jotti.rocks       → static landing page
#   https://demo.jotti.rocks  → demo app (frontend + backend API)
#   https://auth.jotti.rocks  → acme-dns API (trusted local TLS)
# Caddy obtains every certificate itself (HTTP-01) and retries a name until it
# resolves to this server.

DOMAIN="jotti.rocks"
DOMAIN_WWW="www.jotti.rocks"
DOMAIN_DEMO="demo.jotti.rocks"
DOMAIN_AUTH="auth.jotti.rocks"

COMPOSE_FILE="docker-compose.rocks.yml"
CONTAINERS=(jotti-backend jotti-frontend jotti-website jotti-acme-dns jotti-resolver jotti-reverse-proxy)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib.sh
. "$SCRIPT_DIR/lib.sh"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

info "Project root: $PROJECT_ROOT"

info "Checking prerequisites..."

require_docker_stack "$COMPOSE_FILE"

if ! grep -qE '^VPS_PUBLIC_IP=.+' .env; then
  fatal "VPS_PUBLIC_IP missing or empty in .env (public IPv4 of this server, needed by resolver + acme-dns). See docs/jotti-rocks-infra.md."
fi

# A DNS lookup tool is required for the resolution preflight below; without one,
# that check would misreport a missing tool as a DNS failure.
if ! command -v host &>/dev/null && ! command -v dig &>/dev/null; then
  fatal "Neither 'host' nor 'dig' found. Install one (e.g. dnsutils / bind-tools) for the DNS preflight."
fi

info "Prerequisites OK."

resolves() {
  host "$1" &>/dev/null || dig +short "$1" 2>/dev/null | grep -q .
}

if ! resolves "$DOMAIN"; then
  fatal "DNS resolution failed for $DOMAIN. Ensure the domain points to this server before continuing."
fi
info "DNS resolution for $DOMAIN: OK"

# auth.jotti.rocks resolves via the resolver/acme-dns stack on this server, so
# on a fresh install it resolves only once the stack is up and the NS
# delegation is set.
for name in "$DOMAIN_WWW" "$DOMAIN_DEMO" "$DOMAIN_AUTH"; do
  if resolves "$name"; then
    info "DNS resolution for $name: OK"
  else
    warn "DNS resolution for $name failed. Caddy retries its certificate until it resolves."
  fi
done

info "Building and starting the stack..."
docker compose -f "$COMPOSE_FILE" up -d --build

for container in "${CONTAINERS[@]}"; do
  if wait_for_healthy "$container"; then
    info "$container: healthy"
  else
    warn "$container is not healthy yet — check 'make rocks-logs'."
  fi
done

info "Verifying deployment..."

# Retries cover Caddy still obtaining a certificate right after the start.
status_of() {
  curl -s -o /dev/null -w "%{http_code}" --max-time 10 \
    --retry 5 --retry-delay 5 --retry-all-errors "$1" 2>/dev/null || echo "000"
}

HTTPS_STATUS="$(status_of "https://$DOMAIN")"
if [[ "$HTTPS_STATUS" == "200" ]]; then
  info "Landing page HTTPS check: OK"
else
  warn "Landing page HTTPS check returned HTTP $HTTPS_STATUS (expected 200)."
fi

DEMO_STATUS="$(status_of "https://$DOMAIN_DEMO")"
if [[ "$DEMO_STATUS" == "200" ]]; then
  info "Demo app HTTPS check: OK"
else
  warn "Demo app HTTPS check returned HTTP $DEMO_STATUS (expected 200)."
fi

AUTH_STATUS="$(status_of "https://$DOMAIN_AUTH/health")"
if [[ "$AUTH_STATUS" == "200" ]]; then
  info "acme-dns API HTTPS check: OK"
else
  warn "acme-dns API HTTPS check returned HTTP $AUTH_STATUS — expected until the delegation for $DOMAIN_AUTH is active (see docs/jotti-rocks-infra.md)."
fi

HTTP_STATUS="$(status_of "http://$DOMAIN")"
if [[ "$HTTP_STATUS" == "308" ]]; then
  info "HTTP→HTTPS redirect: OK"
else
  warn "HTTP→HTTPS redirect returned HTTP $HTTP_STATUS (expected 308)."
fi

echo ""
echo "=========================================="
printf "${GREEN} %s${NC}\n" "jotti.rocks — Deployment Complete"
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
echo "  Caddy renews the certificates automatically."
echo "=========================================="
