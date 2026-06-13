# Custom Caddy image for the local / LAN stack (docker-compose.local.yml).
#
# Builds Caddy with the acme-dns DNS provider module (dns.providers.acmedns),
# required for DNS-01 wildcard certificates via lokal.jotti.rocks. The module is
# compiled in already so the trusted-certificate phase only changes configuration.
#
# prod/rocks deployments keep stock nginx and do not use this image.

# --- Build Stage ---
FROM caddy:2.11.4-builder AS builder

RUN xcaddy build \
  --with github.com/caddy-dns/acmedns@v0.7.0

# --- Deploy Stage ---
FROM caddy:2.11.4

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY Caddyfile.local /etc/caddy/Caddyfile
