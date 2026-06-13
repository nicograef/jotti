# Custom image for the local / LAN stack (docker-compose.local.yml).
#
# Bundles two binaries:
#  - Caddy, built with the acme-dns DNS provider module (dns.providers.acmedns)
#    for DNS-01 wildcard certificates via lokal.jotti.rocks.
#  - jotti-local-proxy, the container entrypoint: it ensures the install state
#    (acme-dns registration), renders the Caddyfile, serves the local status page
#    and runs Caddy as a child. Only dependency: rsc.io/qr (QR code on the page).
#
# prod/rocks deployments keep stock nginx and do not use this image.

# --- Caddy build stage (xcaddy with the acmedns module, pinned) ---
FROM caddy:2.11.4-builder AS caddy-builder

RUN xcaddy build \
  --with github.com/caddy-dns/acmedns@v0.7.0

# --- Entrypoint build stage (jotti-local-proxy, stdlib only) ---
FROM golang:1.26.0-alpine AS proxy-builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /go/bin/jotti-local-proxy .

# --- Deploy Stage ---
FROM caddy:2.11.4

COPY --from=caddy-builder /usr/bin/caddy /usr/bin/caddy
COPY --from=proxy-builder /go/bin/jotti-local-proxy /usr/local/bin/jotti-local-proxy

ENTRYPOINT [ "jotti-local-proxy" ]
