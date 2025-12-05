## Overview

This document covers running the application locally for development (hot reload) and deploying to production with HTTPS certificates.

## Environment Configuration

Before starting the application, you need to configure environment variables:

1. **Copy the example file:**

   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` and set:**

   - `POSTGRES_USER` - Database username (default: admin)
   - `POSTGRES_PASSWORD` - Database password (**change this!**)
   - `JWT_SECRET` - Secret key for JWT signing (**required, no default**)

3. **Generate a secure JWT secret:**
   ```bash
   openssl rand -base64 32
   ```

**Important:** The application will **fail to start** if `JWT_SECRET` is not set. This is a security feature.

## Local Development (HTTP, Hot Reload)

Local access uses `localhost` only (no subdomains needed).

**Start dev stack:**

```bash
docker compose -f docker-compose.dev.yml up --build -d
```

- Frontend: http://localhost (SPA served at root)
- Backend API: http://localhost/api (reverse proxied path prefix)

Edit Go or TS/TSX files and refresh your browser. Containers run `go run` (backend) and `pnpm dev` (frontend) with hot reload enabled.

**View logs:**

```bash
docker compose -f docker-compose.dev.yml logs -f backend-dev
docker compose -f docker-compose.dev.yml logs -f frontend-dev
docker compose -f docker-compose.dev.yml logs -f reverse-proxy-dev
```

**Stop dev stack:**

```bash
docker compose -f docker-compose.dev.yml down
```

**Optional local HTTPS (mkcert):**

```bash
mkcert localhost
# Then add a 443 server block to `reverse-proxy/nginx.dev.conf` referencing the generated certs.
```

## Production Deployment

### Initial Certificate Setup (First Time Only)

Use the `docker-compose.initial-cert.yml` file to obtain your first Let's Encrypt certificate. This minimal stack runs only nginx and certbot without the application services.

**Prerequisites:**

- DNS A records for `jotti.rocks` and `www.jotti.rocks` pointing to your server's public IP
- Firewall allows incoming traffic on ports 80 and 443

**Step 1: Start minimal stack**

```bash
docker compose -f docker-compose.initial-cert.yml up -d
```

**Step 2: Obtain Let's Encrypt certificate**

```bash
docker compose -f docker-compose.initial-cert.yml run --rm --entrypoint certbot certbot certonly \
  --webroot -w /var/www/certbot \
  -d jotti.rocks -d www.jotti.rocks \
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

The certificate is now mounted and the full stack (frontend, backend, database, reverse proxy with HTTPS) will start successfully.

### Running Production Stack

After initial certificate setup, bring up the full stack:

```bash
docker compose up -d --build
```

Certificates automatically renew every 12 hours via the certbot service.

**Test certificate renewal:**

```bash
docker compose run --rm --entrypoint certbot certbot renew --dry-run
```

**Rebuild and restart:**

```bash
docker compose up -d --build
```

**View backend logs:**

```bash
docker compose logs -f backend
```

**Stop production stack:**

```bash
docker compose down
```

## Database Access

Connect to the PostgreSQL database:

```bash
psql -h localhost -p 5432 -U ${POSTGRES_USER} -d jotti
```

## Configuration Files

| File                                    | Purpose                                           |
| --------------------------------------- | ------------------------------------------------- |
| `docker-compose.yml`                    | Production stack with full application            |
| `docker-compose.initial-cert.yml`       | Minimal stack for first-time certificate issuance |
| `docker-compose.dev.yml`                | Development stack with hot reload                 |
| `reverse-proxy/nginx.conf`              | Production nginx config with HTTPS                |
| `reverse-proxy/nginx.initial-cert.conf` | Minimal nginx config for certificate issuance     |
| `reverse-proxy/nginx.dev.conf`          | Development nginx config (HTTP only)              |

## Docker Volumes

| Volume               | Mount Path                 | Description                            |
| -------------------- | -------------------------- | -------------------------------------- |
| `certbot-challenges` | `/var/www/certbot`         | ACME challenge files for Let's Encrypt |
| `letsencrypt`        | `/etc/letsencrypt`         | SSL certificates and renewal config    |
| `postgres-data`      | `/var/lib/postgresql/data` | PostgreSQL database files              |

## Troubleshooting

### Certificate Issues

**Certbot connection timeout:**

- Verify DNS: `dig +short jotti.rocks` and `dig +short www.jotti.rocks`
- Check port 80: `sudo netstat -tlnp | grep :80`
- Test external access: `curl -I http://jotti.rocks/.well-known/acme-challenge/test`
- Ensure firewall allows ports 80 and 443
- Use `docker-compose.initial-cert.yml` for first-time setup

**Certificate not found on startup:**

- Ensure certificates were created using initial-cert stack first
- Check volume mount: `docker compose exec reverse-proxy ls -la /etc/letsencrypt/live/jotti.rocks/`

### Application Issues

**Stale frontend assets:**

```bash
docker compose restart frontend
# Or rebuild if environment variables changed
docker compose up -d --build frontend
```

**Database migrations not applied:**

```bash
docker compose logs migrate
```

**Backend errors:**

```bash
docker compose logs -f backend
```

## Security Notes

- **CSP enforced:** Content Security Policy headers are configured in both production and dev nginx configs
  - Dev CSP includes `unsafe-eval` and `unsafe-inline` for Vite HMR
  - Production CSP is strict; add external domains explicitly to relevant directives
- **Rate limiting:** API endpoints limited to 10 requests/second per IP (burst 20)
- **HTTPS only:** Production redirects all HTTP traffic to HTTPS
- **www redirect:** `www.jotti.rocks` automatically redirects to `jotti.rocks` for canonical URL
- **Regular maintenance:** Prune unused Docker volumes periodically to save space
