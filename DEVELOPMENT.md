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
make dev
# oder: docker compose -f docker-compose.dev.yml up --build -d
```

- Frontend: http://localhost (Vite Dev-Server mit HMR)
- Backend-API: http://localhost/api (nginx Reverse Proxy)
- PostgreSQL: `localhost:5432`

Backend läuft mit `go run`, Frontend mit `pnpm dev` — Änderungen werden automatisch übernommen.

```bash
# Logs
make logs
# oder: docker compose -f docker-compose.dev.yml logs -f backend-dev

# Stoppen
make down
# oder: docker compose -f docker-compose.dev.yml down
```

## Tests

### Unit-Tests (Backend)

```bash
make test
# oder: cd backend && go test -tags=unit -race ./...
```

### Integrationstests (Backend)

Benötigt eine laufende PostgreSQL-Instanz mit angewendeten Migrationen:

```bash
make test-integration
# oder: ./test-integration.sh
```

Oder manuell in CI: `go test -tags=integration -race ./...`

### Linting

```bash
make lint           # Backend + Frontend
make lint-backend   # nur Backend
make lint-frontend  # nur Frontend

# oder manuell:
cd backend && go vet ./... && goimports -l .
cd frontend && pnpm lint
```

## Production-Deployment

### Erstes Deployment (einmalig)

Das Skript `scripts/prod-init.sh` automatisiert das erste Deployment:

```bash
make init
# oder: ./scripts/prod-init.sh
```

Das Skript führt folgende Schritte aus:

1. Prüft Voraussetzungen (`.env`, Docker, DNS)
2. Startet nginx für die ACME-Challenge
3. Fordert ein Let's-Encrypt-Zertifikat an
4. Startet den vollständigen Produktionsstack
5. Verifiziert, dass HTTPS funktioniert

**Voraussetzungen:**

- `.env` konfiguriert (siehe oben)
- DNS für `jotti.rocks` zeigt auf den Server
- Port 80 ist erreichbar (für ACME-Challenge)

### Manuelle Zertifikatserstellung (alternativ)

Falls das automatisierte Skript nicht verwendet werden soll:

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
make prod-up
# oder: docker compose up -d --build
```

Zertifikate werden automatisch alle 24h via Certbot erneuert.

```bash
# Logs
docker compose logs -f backend

# Stoppen
make prod-down
# oder: docker compose down
```

## Konfigurationsdateien

| Datei                              | Zweck                           |
| ---------------------------------- | ------------------------------- |
| `docker-compose.yml`               | Produktionsstack                |
| `docker-compose.staging.yml`       | Staging-Stack                   |
| `docker-compose.dev.yml`           | Entwicklung mit Hot Reload      |
| `docker-compose.initial-cert.yml`  | Erstzertifikat (Let's Encrypt)  |
| `scripts/prod-init.sh`             | Automatisiertes Erst-Deployment |
| `reverse-proxy/nginx.conf`         | nginx Produktion (HTTPS)        |
| `reverse-proxy/nginx.dev.conf`     | nginx Entwicklung (HTTP)        |
| `reverse-proxy/nginx.staging.conf` | nginx Staging                   |

## CI/CD

GitHub Actions CI (`ci.yml`) führt bei Push/PR auf `main` aus:

- **Backend**: `go vet`, `goimports`, `go build`, Unit-Tests, `golangci-lint`, Integrationstests
- **Frontend**: `pnpm lint`, `pnpm build`

Nur geänderte Pfade werden getestet (via `dorny/paths-filter`).

## Anforderungen & Roadmap

Der vollständige Anforderungskatalog mit 50 Anforderungen (21 umgesetzt, 2 teilweise, 27 offen) liegt in `ANFORDERUNGEN.md`. Dort finden sich auch konkrete Implementierungsvorschläge für jede offene Anforderung.

Für Kodier-Agenten sind die wichtigsten Architekturhinweise und offenen Features in `AGENTS.md` zusammengefasst.
