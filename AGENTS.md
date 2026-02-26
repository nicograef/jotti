# Agent Instructions — jotti

Dieses Dokument richtet sich an KI-Coding-Agenten (Copilot, Cursor, Cline, Aider, etc.).

## Projektbeschreibung

jotti ist ein Bestell- und Kassensystem für Vereine und Nonprofit-Veranstaltungen. Servicekräfte nehmen auf Smartphones Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer.

## Tech-Stack

| Komponente    | Technologie                                                        | Versionen in             |
| ------------- | ------------------------------------------------------------------ | ------------------------ |
| Backend       | Go, stdlib `net/http`, `pgx/v5`, `zerolog`, `zog`, `golang-jwt/v5` | `backend/go.mod`         |
| Frontend      | React, Vite, TypeScript (strict), Tailwind CSS, shadcn/ui, Zod     | `frontend/package.json`  |
| Datenbank     | PostgreSQL, `golang-migrate`                                       | `docker-compose.dev.yml` |
| Infrastruktur | Docker Compose, nginx Reverse Proxy, Let's Encrypt                 |                          |

## Projektstand & Anforderungen

Der vollständige Anforderungskatalog mit Implementierungsvorschlägen und aktuellem Status liegt in `ANFORDERUNGEN.md`. Vor jeder Feature-Arbeit dort den aktuellen Stand prüfen.

## Backend-Konventionen

### Fehlerformat

Alle Fehler-Responses: `{"code": "<string>", "details": "<optional>"}` (siehe `api/helper/http.go`).

### Auth

- JWT HS256, 12h Gültigkeit, Claims: `sub` (userID), `role` (admin|senior_service|service)
- Middleware extrahiert `userID` und `role` aus JWT in Request-Context
- Passwörter: Argon2id-Hashing (`domain/user/password.go`)

### State-Rekonstruktion aus Events

- Balance = Summe(Bestellungen) − Summe(Bezahlungen) − Summe(Stornierungen)
- Unbezahlt = bestellt − bezahlt − storniert
- Ungeliefert = bestellt − geliefert − storniert

## Frontend-Konventionen

### UI-Bibliotheken

- **shadcn/ui** (Stil: `new-york`, Radix-basiert)
- **Lucide React** (Icons)
- **Sonner** (Toasts) — alle mutativen Aktionen zeigen `toast.error(...)` bei Fehlern
- **Vaul** (Drawers)

### Patterns

- **401-Interceptor**: `Backend.post()` erkennt 401, loggt aus und leitet zu `/login` weiter — kein manuelles 401-Handling nötig
- **Drawer-Pattern**: Bestellen, Bezahlen, Stornieren, Liefern öffnen Bottom-Sheet-Drawer mit Zusammenfassung. Hilfsfunktionen (`selectVariants`, `calculateTotalPrice`) in `src/service/components/table/drawerUtils.ts`
- **Geldbeträge anzeigen**: `formatCents()` aus `src/lib/utils.ts` — nie inline formatieren

### Styling

- Tailwind CSS 4 via `@tailwindcss/vite` (keine `tailwind.config.js`)
- CSS-Variablen in `src/index.css` (Violet/Indigo-Schema, Dark Mode via `.dark`-Klasse)
- `cn()` Utility aus `src/lib/utils.ts` (`clsx` + `tailwind-merge`)
- Path-Alias: `@/` → `./src/`

## Wichtige Regeln

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE.
2. **Geldbeträge sind immer in Cent (int).** Niemals Floats für Geld verwenden.
3. **Event-Sourcing für Tisch-Operationen.** Events sind immutable (append-only). Nie Events updaten oder löschen.
4. **CRUD für Stammdaten** (Benutzer, Produkte, Tische). Soft-Deletes via `status = 'deleted'`.
5. **Validierung mit Schemas.** Backend: `zog`. Frontend: `Zod`. Beide Seiten validieren.
6. **Deutsche UI, englischer Code.** Alle Benutzer-sichtbaren Strings auf Deutsch. Code, Variablen, Commits auf Englisch.
7. **Kein globaler State-Store im Frontend.** Nur React Hooks + Singletons.
8. **Frontend API-Aufrufe nur über Backend-Klassen.** Nie direkt `fetch()` verwenden. Alle Domain-Backend-Klassen nutzen das `BackendClient`-Interface aus `src/lib/Backend.ts`.
9. **Dokumentation synchron halten.** Bei jeder Änderung am Projekt (neue Endpunkte, neue Seiten, geänderte Architektur, neue Dependencies, Versionsänderungen etc.) müssen folgende Dateien aktualisiert werden, sofern betroffen:
   - `AGENTS.md` — Tech-Stack-Tabelle, Verzeichnisstruktur, Regeln, Event-Referenz, DB-Schema
   - `README.md` — Feature-Liste, Architektur-Tabelle
   - `DEVELOPMENT.md` — Build-/Test-/Deploy-Befehle, Konfigurationsdateien-Tabelle
10. **Backend ist die Single Source of Truth für Daten-Filterung.** Filterung, Aggregation und Aufbereitung gehören ins Backend (SQL/Repository). Das Frontend zeigt an, was das Backend liefert — keine redundante Filterlogik im Frontend duplizieren. Vor dem Hinzufügen von Frontend-Filtern prüfen, ob das Backend die Daten bereits korrekt aufbereitet.

## Bereiche: Admin vs. Service

jotti hat zwei getrennte Bereiche mit unterschiedlichen Rollen und Funktionen:

### Admin-Bereich (Rolle: `admin`)

Verwaltung von Stammdaten. Nur für Administratoren zugänglich.

- **Backend**: Routen in `api/admin.go` unter `/admin/*`, JWT-Middleware erlaubt nur Rolle `admin`
- **Frontend**: Seiten unter `src/admin/` mit `AdminGuard` (React Router Loader)
- **Funktionen**: Produkte + Varianten erstellen/bearbeiten/aktivieren/deaktivieren, Tische verwalten, Benutzer anlegen/bearbeiten/Passwort zurücksetzen
- **Endpunkte**: siehe `backend/api/admin.go`

### Service-Bereich (Rollen: `admin` + `senior_service` + `service`)

Bestell- und Kassierbetrieb am Tisch. Für Servicekräfte, Serviceleitung und Admins zugänglich.

- **Backend**: Routen in `api/service.go` unter `/service/*`, JWT-Middleware erlaubt Rollen `admin`, `senior_service` und `service`. Stornierung (`cancel-table-variants`) läuft über eigene `api/senior_service.go` mit Middleware nur für `admin` und `senior_service`.
- **Frontend**: Seiten unter `src/service/` mit `ServiceGuard` (React Router Loader)
- **Funktionen**: Tisch auswählen, Bestellungen aufgeben, Lieferungen bestätigen, Bezahlungen registrieren, Stornierungen, Tisch-Verlauf einsehen
- **Endpunkte**: siehe `backend/api/service.go` und `backend/api/senior_service.go`

### Auth-Bereich (kein JWT erforderlich)

- **Backend**: Routen in `api/auth.go` unter `/auth/*`
- **Endpunkte**: siehe `backend/api/auth.go`

## Verzeichnisstruktur

```
backend/
  main.go                       # Einstiegspunkt
  api/service.go                # Service-Routen (Bestell-/Kassierbetrieb)
  api/senior_service.go         # Senior-Service-Routen (Stornierung)
  api/admin.go                  # Admin-Routen (Verwaltung)
  api/auth.go                   # Auth-Routen (Login, Passwort setzen)
  api/<domain>/http/            # HTTP-Handler
  api/<domain>/application/     # Application-Services
  api/middleware/               # JWT-Auth, Rate-Limiting, Logging
  api/helper/                   # HTTP-Hilfsfunktionen (JSON-Parsing, Response)
  domain/<domain>/              # Domain-Modelle und Business-Logik
  repository/<domain>_repo/     # Datenbank-Zugriff (SQL via pgx)
  config/                       # Konfiguration aus Umgebungsvariablen
  app/                          # App-Struct (Dependency Wiring)
  db/                           # Datenbank-Verbindung

frontend/
  src/routes.ts                 # Alle Routen + Guards
  src/App.tsx                   # Root-Komponente
  src/lib/                      # Auth, Backend-Client (BackendClient-Interface), useFetch-Hook, Utilities
  src/admin/                    # Admin-Bereich (Produkte, Tische, Benutzer)
  src/service/                  # Service-Bereich (Tisch-Workflow)
  src/pages/                    # Login, Passwort setzen
  src/components/ui/            # shadcn/ui-Komponenten
  src/components/common/        # Gemeinsame Komponenten

database/
  migrations/                   # SQL-Migrationen (up/down)
  migrate/                      # Migration-Tool (Container)

reverse-proxy/                  # nginx-Konfigurationen
```

## Lokale Entwicklung

```bash
cp .env.example .env            # Umgebungsvariablen konfigurieren
docker compose -f docker-compose.dev.yml up --build -d
# Frontend: http://localhost | API: http://localhost/api
```

## Tests ausführen

```bash
cd backend && go test -tags=unit -race ./...         # Unit-Tests
./test-integration.sh                                 # Integrationstests
cd frontend && pnpm lint                              # Frontend-Lint
```

## Wie füge ich ein neues Feature hinzu?

### Backend (neuer Endpunkt)

1. Domain-Modell + zog-Schema in `domain/<domain>/`
2. Repository-Interface + Implementierung in `repository/<domain>_repo/`
3. Application-Service in `api/<domain>/application/`
4. HTTP-Handler in `api/<domain>/http/`
5. Route registrieren in `api/admin.go` oder `api/service.go`
6. Unit-Test mit `//go:build unit` Tag

### Frontend (neue Seite)

1. Zod-Schema + TypeScript-Typen in Feature-Verzeichnis
2. Backend-Client-Klasse (nutzt `BackendClient`-Interface aus `@/lib/Backend`)
3. Custom Hook via `useFetch<T>()` aus `@/lib/useFetch`
4. React-Komponenten
5. Route in `src/routes.ts` registrieren

## Event-Sourcing-Referenz

Events für Tisch-Operationen. Subject-Format: `"table:<id>"`. State wird durch Replay aller Events rekonstruiert. Snapshots optimieren Lesezugriffe.

Alle Event-Typen und deren Datenstrukturen: siehe `backend/domain/table/events.go` und die zugehörigen `*Event.go`-Dateien im selben Verzeichnis.

## Datenbank-Schema

Tabellen: `users`, `tables`, `products`, `product_variants`, `events` (append-only).

Aktuelles Schema und Spalten: siehe SQL-Migrationen in `database/migrations/` (alle `*.up.sql`-Dateien in Reihenfolge anwenden).

Neue Migration: `database/migrations/<nr>_<name>.up.sql` + `.down.sql`
