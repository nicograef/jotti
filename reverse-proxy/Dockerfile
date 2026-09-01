# Custom image for the LAN stack (docker-compose.local.yml / release) and the
# self-hosted production stack (docker-compose.prod.yml).
#
# Bundles two binaries:
#  - Caddy, built with the acme-dns DNS provider module (dns.providers.acmedns)
#    for DNS-01 wildcard certificates via lokal.jotti.rocks (LAN mode) and the
#    caddy-ratelimit module for the public /api/ rate limit (public mode).
#  - jotti-reverse-proxy, the container entrypoint: in LAN mode it ensures the
#    install state (acme-dns registration), renders the Caddyfile and serves the
#    local status page; in public mode (JOTTI_DOMAIN set) it renders a public
#    Caddyfile with automatic Let's Encrypt. Either way it runs Caddy as a child.
#    Only dependency: rsc.io/qr (QR code on the status page).
#
# The jotti.rocks demo keeps stock nginx and does not use this image.

# --- Caddy build stage (xcaddy with the acmedns + ratelimit modules, pinned) ---
FROM caddy:2.11.4-builder AS caddy-builder

RUN xcaddy build \
  --with github.com/caddy-dns/acmedns@v0.7.0 \
  --with github.com/mholt/caddy-ratelimit@v0.1.0

# --- Entrypoint build stage (jotti-reverse-proxy, stdlib + rsc.io/qr only) ---
FROM golang:1.27.0-alpine AS proxy-builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /go/bin/jotti-reverse-proxy .

# --- Deploy stage ---
FROM caddy:2.11.4

COPY --from=caddy-builder /usr/bin/caddy /usr/bin/caddy
COPY --from=proxy-builder /go/bin/jotti-reverse-proxy /usr/local/bin/jotti-reverse-proxy

ENTRYPOINT [ "jotti-reverse-proxy" ]
