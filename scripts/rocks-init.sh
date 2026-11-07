#!/usr/bin/env bash
# rocks-init.sh — first deploy of the jotti.rocks stack (docker-compose.rocks.yml)
#
# Usage:
#   make rocks-init   # self-hosters use scripts/prod-init.sh instead
#
# Serves:
#   https://jotti.rocks       static landing page
#   https://demo.jotti.rocks  demo app (frontend + backend API)
#   https://auth.jotti.rocks  acme-dns API (trusted local TLS)
# Caddy obtains every certificate itself (HTTP-01) and retries a name until it
# resolves to this server.
set -euo pipefail

DOMAIN="jotti.rocks"
DOMAIN_WWW="www.jotti.rocks"
DOMAIN_DEMO="demo.jotti.rocks"
DOMAIN_AUTH="auth.jotti.rocks"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.rocks.yml}"
CONTAINERS=(jotti-backend jotti-acme-dns jotti-resolver jotti-reverse-proxy)

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

require_dns_tool

info "Prerequisites OK."

if [[ -z "$(resolve_a "$DOMAIN")" ]]; then
  fatal "DNS resolution failed for $DOMAIN. Ensure the domain points to this server before continuing."
fi
info "DNS resolution for $DOMAIN: OK"

# auth.jotti.rocks resolves via the resolver/acme-dns stack on this server, so
# on a fresh install it resolves only once the stack is up and the NS
# delegation is set.
for name in "$DOMAIN_WWW" "$DOMAIN_DEMO" "$DOMAIN_AUTH"; do
  if [[ -n "$(resolve_a "$name")" ]]; then
    info "DNS resolution for $name: OK"
  else
    warn "DNS resolution for $name failed. Caddy retries its certificate until it resolves."
  fi
done

info "Building and starting the stack..."
docker compose -f "$COMPOSE_FILE" up -d --build --remove-orphans

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

# expect_status LABEL URL WANT [HINT] — warns unless URL answers HTTP WANT; HINT
# replaces the "(expected WANT)" suffix.
expect_status() {
  local label="$1" url="$2" want="$3" hint="${4:-}" status
  status="$(status_of "$url")"
  if [[ "$status" == "$want" ]]; then
    info "$label: OK"
  elif [[ -n "$hint" ]]; then
    warn "$label returned HTTP $status — $hint"
  else
    warn "$label returned HTTP $status (expected $want)."
  fi
}

expect_status "Landing page HTTPS check" "https://$DOMAIN" 200
expect_status "Demo app HTTPS check" "https://$DOMAIN_DEMO" 200
expect_status "acme-dns API HTTPS check" "https://$DOMAIN_AUTH/health" 200 \
  "expected until the delegation for $DOMAIN_AUTH is active (see docs/jotti-rocks-infra.md)."
expect_status "HTTP→HTTPS redirect" "http://$DOMAIN" 308

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
