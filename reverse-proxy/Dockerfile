# Custom image for the LAN stack (docker-compose.local.yml / release) and the
# self-hosted production stack (docker-compose.prod.yml). Bundles Caddy — built with
# the acme-dns DNS provider (DNS-01 wildcard certificates) and caddy-ratelimit (the
# public /api/ rate limit) — plus jotti-reverse-proxy as the entrypoint, which
# renders the Caddyfile and runs Caddy as a child, plus the app's static build,
# which Caddy serves itself.
# The jotti.rocks stack (docker-compose.rocks.yml) builds the `rocks` target, which
# adds the website, and runs the bundled Caddy directly on the static Caddyfile.rocks.
# The build context is the repository root; Dockerfile.dockerignore limits it.

FROM caddy:2.11.7-builder AS caddy-builder

RUN xcaddy build \
  --with github.com/caddy-dns/acmedns@v0.7.0 \
  --with github.com/mholt/caddy-ratelimit@v0.1.0

FROM golang:1.27-alpine AS proxy-builder

# The image tag names the Go minor; the go directive in go.mod names the patch.
ENV GOTOOLCHAIN=auto

WORKDIR /src
COPY reverse-proxy/go.mod reverse-proxy/go.sum ./
RUN go mod download
COPY reverse-proxy/*.go ./
RUN go build -o /go/bin/jotti-reverse-proxy .

FROM node:24-alpine AS app-builder

WORKDIR /app

COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN corepack enable \
  && pnpm install --frozen-lockfile

COPY frontend/ ./

# Das VERSION-Argument wird im RUN konsumiert, damit ein Versionswechsel die
# Build-Schicht invalidiert und nicht das alte Bundle im neuen Image landet.
ARG VERSION=dev
RUN VERSION="${VERSION}" pnpm build

FROM node:24-alpine AS website-builder

WORKDIR /app

COPY website/package.json website/pnpm-lock.yaml website/pnpm-workspace.yaml ./
RUN corepack enable \
  && pnpm install --frozen-lockfile

# Astro reads docs/ as ../docs relative to the project root (/app).
COPY website/ ./
COPY docs/ /docs/
RUN pnpm build

FROM caddy:2.11.7 AS app

COPY --from=caddy-builder /usr/bin/caddy /usr/bin/caddy
COPY --from=proxy-builder /go/bin/jotti-reverse-proxy /usr/local/bin/jotti-reverse-proxy
COPY --from=app-builder /app/dist /srv/app

ENTRYPOINT [ "jotti-reverse-proxy" ]

FROM app AS rocks

COPY --from=website-builder /app/dist /srv/website

# The last stage is the default target: the self-hosted image, without the website.
FROM app
