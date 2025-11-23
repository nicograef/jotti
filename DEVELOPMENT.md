## Overview

This document covers running the application locally (hot reload) and running the production-style reverse proxy stack with HTTPS certificates.

## Local Development (HTTP, hot reload)

Local access uses `localhost` only (no subdomains needed). Optionally add an entry for a custom name e.g. `jotti.local` if desired.

Start dev stack:

```bash
docker compose -f docker-compose.dev.yml up -d
```

Frontend: http://localhost (SPA served at root).  
Backend API: http://localhost/api (reverse proxied path prefix).

Edit Go or TS/TSX files and refresh; containers run `go run` (backend) and `pnpm dev` (frontend). No extra reload tooling required.

Logs:

```bash
docker compose -f docker-compose.dev.yml logs -f backend-dev
docker compose -f docker-compose.dev.yml logs -f frontend-dev
docker compose -f docker-compose.dev.yml logs -f reverse-proxy-dev
```

Stop:

```bash
docker compose -f docker-compose.dev.yml down
```

Dev proxy config: `reverse-proxy/nginx.dev.conf`.

Optional local HTTPS (mkcert):

```bash
mkcert localhost
# Then add a 443 server block to `reverse-proxy/nginx.dev.conf` referencing the generated certs.
```

## Production Reverse Proxy Stack

Bring up full stack (includes Certbot renewal loop):

```bash
docker compose -f docker-compose.reverse-proxy.yml up -d --build
```

Initial certificate issuance (single-domain after DNS A record for `jotti.rocks` points to host):

```bash
docker compose -f docker-compose.reverse-proxy.yml up -d reverse-proxy
docker compose -f docker-compose.reverse-proxy.yml run --rm certbot certbot certonly \
  --webroot -w /var/www/certbot \
  -d jotti.rocks \
  --email admin@jotti.rocks --agree-tos --no-eff-email
docker compose -f docker-compose.reverse-proxy.yml restart reverse-proxy
```

Automatic renewal runs every 12h. Test renewal:

```bash
docker compose -f docker-compose.reverse-proxy.yml run --rm certbot certbot renew --dry-run
```

## Directory & Volumes

- Reverse proxy config: `reverse-proxy/nginx.conf`
- Dev proxy config: `reverse-proxy/nginx.dev.conf`
- Certbot challenges volume: `certbot-challenges` → `/var/www/certbot`
- Live certs volume: `letsencrypt` → `/etc/letsencrypt/live/<domain>`
- Postgres data: `postgres-data`

## Common Commands

Rebuild & start production stack:

```bash
docker compose -f docker-compose.reverse-proxy.yml up -d --build
```

Tail backend logs (prod):

```bash
docker compose -f docker-compose.reverse-proxy.yml logs -f backend
```

DB psql (local default compose):

```bash
psql -h localhost -p 5432 -U ${POSTGRES_USER} -d jotti
```

Tear down prod stack:

```bash
docker compose -f docker-compose.reverse-proxy.yml down
```

## Troubleshooting

- Certbot challenges failing: ensure port 80 reachable and DNS A records correct.
- Permissions issues on certs: verify volumes mounted read/write for certbot and read-only for nginx.
- Stale frontend assets: restart `frontend` or rebuild if environment variables changed.
- Database migrations not applied: check `migrate` container logs.

## Security Notes

- CSP enforced in production (`reverse-proxy/nginx.conf`) and dev (`reverse-proxy/nginx.dev.conf`).
  - Dev CSP includes `unsafe-eval` / `unsafe-inline` for Vite HMR; remove for staging/prod hardened builds.
  - Add external domains explicitly to relevant directives (`script-src`, `style-src`, `font-src`, `connect-src`).
- Regularly prune unused Docker volumes to save space.
