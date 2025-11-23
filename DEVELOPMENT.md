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

## Production Stack

### Initial Certificate Setup (First Time Only)

Use the initial-cert compose file to obtain your first Let's Encrypt certificate. This runs only nginx + certbot without the application services.

**Step 1: Start minimal stack for certificate issuance**

```bash
docker compose -f docker-compose.initial-cert.yml up -d
```

**Step 2: Obtain Let's Encrypt certificate**

After DNS A record for `jotti.rocks` points to your server and port 80 is accessible:

```bash
docker compose -f docker-compose.initial-cert.yml run --rm --entrypoint certbot certbot certonly \
  --webroot -w /var/www/certbot \
  -d jotti.rocks \
  --email graef.nico@gmail.com --agree-tos --no-eff-email
```

**Step 3: Stop initial-cert stack**

```bash
docker compose -f docker-compose.initial-cert.yml down
```

**Step 4: Start full production stack**

```bash
docker compose up -d --build
```

The certificate is now in place and the full stack (frontend, backend, database, reverse proxy with HTTPS) will start successfully.

### Running Production Stack

Bring up full stack (after certificates are obtained):

```bash
docker compose up -d --build
```

Automatic renewal runs every 12h. Test renewal:

```bash
docker compose run --rm --entrypoint certbot certbot renew --dry-run
```

### Troubleshooting Certificate Issues

If certbot fails with connection timeout:

- Verify DNS A record points to your server: `dig +short jotti.rocks`
- Check port 80 is open: `sudo netstat -tlnp | grep :80`
- Test external access: `curl -I http://jotti.rocks/.well-known/acme-challenge/test`
- Check firewall rules allow incoming HTTP (port 80)
- Ensure you're using the initial-cert compose file for first-time setup

Adding/transitioning old subdomains (optional redirects):

1. Keep existing certificates for `app.jotti.rocks` / `api.jotti.rocks` until clients migrate.
2. Add an HTTP-only server block redirecting those hosts to `https://jotti.rocks$request_uri` (already present in updated `nginx.conf`).
3. After traffic drops, remove the SANs / extra certs and clean up DNS.

## Directory & Volumes

- Production reverse proxy config: `reverse-proxy/nginx.conf`
- Initial certificate nginx config: `reverse-proxy/nginx.initial-cert.conf`
- Dev proxy config: `reverse-proxy/nginx.dev.conf`
- Production compose: `docker-compose.yml`
- Initial certificate compose: `docker-compose.initial-cert.yml`
- Dev compose: `docker-compose.dev.yml`
- Certbot challenges volume: `certbot-challenges` → `/var/www/certbot`
- Live certs volume: `letsencrypt` → `/etc/letsencrypt/live/<domain>`
- Postgres data: `postgres-data`

## Common Commands

Rebuild & start production stack:

```bash
docker compose up -d --build
```

Tail backend logs (prod):

```bash
docker compose logs -f backend
```

DB psql (local default compose):

```bash
psql -h localhost -p 5432 -U ${POSTGRES_USER} -d jotti
```

Tear down prod stack:

```bash
docker compose down
```

## Troubleshooting

- **Certbot connection timeout**: Firewall blocking port 80, or DNS not pointing to server. See certificate troubleshooting above.
- Certbot challenges failing: ensure port 80 reachable and DNS A records correct.
- Permissions issues on certs: verify volumes mounted read/write for certbot and read-only for nginx.
- Stale frontend assets: restart `frontend` or rebuild if environment variables changed.
- Database migrations not applied: check `migrate` container logs.

## Security Notes

- CSP enforced in production (`reverse-proxy/nginx.conf`) and dev (`reverse-proxy/nginx.dev.conf`).
  - Dev CSP includes `unsafe-eval` / `unsafe-inline` for Vite HMR; remove for staging/prod hardened builds.
  - Add external domains explicitly to relevant directives (`script-src`, `style-src`, `font-src`, `connect-src`).
- Regularly prune unused Docker volumes to save space.
