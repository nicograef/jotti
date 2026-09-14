# Custom image for the LAN stack (docker-compose.local.yml / release) and the
# self-hosted production stack (docker-compose.prod.yml). Bundles Caddy — built with
# the acme-dns DNS provider (DNS-01 wildcard certificates) and caddy-ratelimit (the
# public /api/ rate limit) — plus jotti-reverse-proxy as the entrypoint, which
# renders the Caddyfile and runs Caddy as a child.
# The jotti.rocks demo keeps stock nginx and does not use this image.

FROM caddy:2.11.4-builder AS caddy-builder

RUN xcaddy build \
  --with github.com/caddy-dns/acmedns@v0.7.0 \
  --with github.com/mholt/caddy-ratelimit@v0.1.0

FROM golang:1.27.1-alpine AS proxy-builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /go/bin/jotti-reverse-proxy .

FROM caddy:2.11.4

COPY --from=caddy-builder /usr/bin/caddy /usr/bin/caddy
COPY --from=proxy-builder /go/bin/jotti-reverse-proxy /usr/local/bin/jotti-reverse-proxy

ENTRYPOINT [ "jotti-reverse-proxy" ]
