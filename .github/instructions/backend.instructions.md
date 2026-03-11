---
description: "Use when working on Go backend code, API handlers, middleware, repositories, domain models, or application services."
applyTo: "backend/**"
---

# Backend-Konventionen

## Verzeichnisstruktur

```
backend/
  main.go                       # Einstiegspunkt
  sqlc.yaml                     # sqlc-Konfiguration
  sqlc/queries/                 # SQL-Queries für sqlc
  sqlc/dbgen/                   # Generierter Code (NICHT EDITIEREN)
  api/service.go                # Service-Routen (Kassenbetrieb)
  api/senior_service.go         # Senior-Service-Routen (Stornierung)
  api/admin.go                  # Admin-Routen (Verwaltung)
  api/auth.go                   # Auth-Routen (Login, Passwort setzen)
  api/<domain>/http/            # HTTP-Handler
  api/<domain>/application/     # Application-Services
  api/middleware/               # JWT-Auth, Rate-Limiting, Logging
  api/helper/                   # HTTP-Hilfsfunktionen (JSON-Parsing, Response)
  domain/<domain>/              # Domain-Modelle und Business-Logik
  repository/<domain>_repo/     # Datenbank-Zugriff (sqlc-basiert)
  config/                       # Konfiguration aus Umgebungsvariablen
  app/                          # App-Struct (Dependency Wiring)
  db/                           # Datenbank-Verbindung und Fehler-Mapping
```

## Fehlerformat

Alle Fehler-Responses: `{"code": "<string>", "details": "<optional>"}` (siehe `api/helper/http.go`).

## Auth

- JWT HS256, 12h Gültigkeit, Claims: `sub` (userID), `role` (admin|senior_service|service)
- Middleware extrahiert `userID` und `role` aus JWT in Request-Context
- Passwörter: Argon2id-Hashing (`domain/user/password.go`)

## State-Rekonstruktion aus Events

- Saldo = Summe(Bestellungen) − Summe(Zahlungen) − Summe(Stornierungen)
- UnbezahltePositionen = bestellt − bezahlt − storniert
- UngeliefertePositionen = bestellt − geliefert − storniert

## Tests

Unit-Tests mit `//go:build unit` Tag. Ausführen: `go test -tags=unit -race ./...`
