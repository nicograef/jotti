# Entwicklung & Deployment

## Umgebungsvariablen

```bash
cp .env.example .env
```

Folgende Variablen setzen:

| Variable            | Beschreibung                                         |
| ------------------- | ---------------------------------------------------- |
| `POSTGRES_USER`     | Datenbank-Benutzer (Standard: `admin`)               |
| `POSTGRES_PASSWORD` | Datenbank-Passwort                                   |
| `JWT_SECRET`        | Geheimer Schlüssel für JWT-Signierung (erforderlich) |

JWT-Secret generieren: `openssl rand -base64 32`

## Lokale Entwicklung

```bash
docker compose -f docker-compose.dev.yml up --build -d
```

- Frontend: http://localhost (Vite Dev-Server mit HMR)
- Backend-API: http://localhost/api (nginx Reverse Proxy)
- PostgreSQL: `localhost:5432`

Backend läuft mit `go run`, Frontend mit `pnpm dev` — Änderungen werden automatisch übernommen.

```bash
# Logs
docker compose -f docker-compose.dev.yml logs -f backend-dev

# Stoppen
docker compose -f docker-compose.dev.yml down
```

## Tests

### Unit-Tests (Backend)

```bash
cd backend && go test -tags=unit -race ./...
```

### Integrationstests (Backend)

Benötigt eine laufende PostgreSQL-Instanz mit angewendeten Migrationen:

```bash
./test-integration.sh
```

Oder manuell in CI: `go test -tags=integration -race ./...`

### Linting

```bash
# Backend
cd backend && go vet ./... && goimports -l .

# Frontend
cd frontend && pnpm lint
```

## Production-Deployment

### Erstes Zertifikat (einmalig)

```bash
docker compose -f docker-compose.initial-cert.yml up -d

docker compose -f docker-compose.initial-cert.yml run --rm --entrypoint certbot certbot certonly \
  --webroot -w /var/www/certbot \
  -d jotti.rocks -d www.jotti.rocks \
  --email graef.nico@gmail.com --agree-tos --no-eff-email

docker compose -f docker-compose.initial-cert.yml down
```

### Produktionsstack starten

```bash
docker compose up -d --build
```

Zertifikate werden automatisch alle 24h via Certbot erneuert.

```bash
# Logs
docker compose logs -f backend

# Stoppen
docker compose down
```

## Konfigurationsdateien

| Datei                              | Zweck                          |
| ---------------------------------- | ------------------------------ |
| `docker-compose.yml`               | Produktionsstack               |
| `docker-compose.staging.yml`       | Staging-Stack                  |
| `docker-compose.dev.yml`           | Entwicklung mit Hot Reload     |
| `docker-compose.initial-cert.yml`  | Erstzertifikat (Let's Encrypt) |
| `reverse-proxy/nginx.conf`         | nginx Produktion (HTTPS)       |
| `reverse-proxy/nginx.dev.conf`     | nginx Entwicklung (HTTP)       |
| `reverse-proxy/nginx.staging.conf` | nginx Staging                  |

## CI/CD

GitHub Actions CI (`ci.yml`) führt bei Push/PR auf `main` aus:

- **Backend**: `go vet`, `goimports`, `go build`, Unit-Tests, `golangci-lint`, Integrationstests
- **Frontend**: `pnpm lint`, `pnpm build`

Nur geänderte Pfade werden getestet (via `dorny/paths-filter`).
