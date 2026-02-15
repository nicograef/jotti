# Copilot Instructions — jotti

## Projekt

jotti ist ein Bestell- und Kassensystem für Vereine. Single-Tenant Webapp deployed via Docker Compose.

## Architektur

- **Backend**: Go 1.25, stdlib `net/http`, PostgreSQL via `pgx/v5`, JWT-Auth, structured logging (`zerolog`)
- **Frontend**: React 19, Vite 7, TypeScript 5.9 (strict), Tailwind CSS 4, shadcn/ui (new-york), Zod 4
- **Datenbank**: PostgreSQL 17, Migrationen via `golang-migrate`
- **Reverse Proxy**: nginx, HTTPS via Let's Encrypt

## Bereiche: Admin vs. Service

### Projektstand & Anforderungen

Der vollständige Anforderungskatalog liegt in `ANFORDERUNGEN.md` (50 Anforderungen: 21 umgesetzt, 2 teilweise, 27 offen).

**Bekannte teilweise Umsetzungen:**

- **Produktkategorien in Service-UI (#21)**: `ProductCategory` existiert im Datenmodell, aber `ProductList.tsx` zeigt flache Liste. → Nach `category` gruppieren.
- **Stornierung nur für Admins (#22)**: `cancel-table-variants` ist für beide Rollen offen. → Rollenprüfung einbauen, Frontend: Stornierungstab nur für Admins.

**Nächste offene Must-haves (siehe `ANFORDERUNGEN.md` für Details):**

- #23 Tisch-Schnellsuche (FAB/Suchfeld in `TableSelectionPage.tsx`)
- #24 Übersicht eigene Bestellungen (`POST /service/get-user-orders`)
- #25 Tisch-Umbuchung (neuer Event-Typ `table.order-transferred:v1`)
- #26 Umsatz pro Bediener (`POST /admin/get-revenue-by-user`)
- #27 Bondruck (ESC/POS oder Web-Print mit Print-CSS)
- #31 Freibon mit freier Preiseingabe
- #33 Offline-Fähigkeit (`vite-plugin-pwa`, Service Worker, IndexedDB-Queue)

**Nice-to-haves:** Rückgeldberechnung (#37), Freitext-Notiz pro Position (#42), Label pro Bestellung (#36), Reporting/Export (#38–40).

### Admin (`/admin/*`) — nur Rolle `admin`

Verwaltung: Produkte + Varianten, Tische, Benutzer. Backend-Routen in `api/admin.go`, Frontend unter `src/admin/`.

### Service (`/service/*`) — Rollen `admin` + `service`

Betrieb: Tisch auswählen, bestellen, liefern, bezahlen, stornieren, Verlauf. Backend-Routen in `api/service.go`, Frontend unter `src/service/`.

### Auth (`/auth/*`) — kein JWT

Login und Passwort setzen. Backend-Routen in `api/auth.go`.

## Backend-Konventionen

### Architektur (pro Domain-Modul)

```
api/<domain>/http/       → HTTP-Handler (Request parsen, Response formatieren)
api/<domain>/application/ → Application-Service (Use-Case-Orchestrierung)
domain/<domain>/          → Domain-Modelle, Validierung, Business-Logik
repository/<domain>_repo/ → SQL-Persistenz (pgx)
```

### Patterns

- **Alle API-Endpunkte sind POST-only** — keine GET/PUT/DELETE
- **CQRS**: Getrennte `Command`- und `Query`-Structs pro Repository mit separaten Interfaces
- **Event Sourcing** (nur für Tisch-Operationen): Bestellungen, Bezahlungen, Lieferungen, Stornierungen als immutable Events in `events`-Tabelle
- **CRUD** für: Benutzer, Produkte, Tische (relationale Tabellen)
- **Validierung**: `zog`-Schemas (ähnlich Zod) für alle Eingaben
- **Fehlerformat**: `{code: string, details?: string}` als JSON
- **Soft-Deletes**: Status `deleted`, in Queries gefiltert
- **Geldbeträge**: immer in Cent (int), keine Floats

### Event-Typen

- `table.order-placed:v1` — Bestellung aufgegeben
- `table.payment-registered:v1` — Bezahlung registriert
- `table.variants-delivered:v1` — Varianten ausgeliefert
- `table.variants-canceled:v1` — Varianten storniert
- `table.snapshot:v1` — Materialisierter Snapshot

### Geplante Event-Typen (noch nicht implementiert)

- `table.order-transferred:v1` — Bestellung auf anderen Tisch umbucht (Anforderung #25)
- `table.variants-prepared:v1` — Varianten zubereitet / abholbereit (Anforderung #35, #45)
- `table.variants-status-changed:v1` — Zubereitungsstatus geändert (Anforderung #46)

### Event-Subject-Format

`"table:<tableID>"` — z.B. `"table:42"`

### State-Rekonstruktion aus Events

- Balance = Summe(Bestellungen) - Summe(Bezahlungen) - Summe(Stornierungen)
- Unbezahlt = bestellt - bezahlt - storniert
- Ungeliefert = bestellt - geliefert - storniert

### Auth

- JWT HS256, 12h Gültigkeit, Claims: `sub` (userID), `role` (admin|service)
- Middleware extrahiert `userID` aus JWT und setzt in Request-Context
- Passwörter: Argon2id-Hashing
- Rollen: `admin` (Vollzugriff), `service` (Bestell-/Kassierbetrieb)

### Neue Endpunkte hinzufügen

1. Domain-Modell in `domain/` definieren (mit zog-Schema)
2. Repository-Interface und -Implementierung in `repository/`
3. Application-Service in `api/<domain>/application/`
4. HTTP-Handler in `api/<domain>/http/`
5. Route in `api/admin.go` oder `api/service.go` registrieren

### Tests

- Unit-Tests: `//go:build unit` Tag
- Integrationstests: `//go:build integration` Tag, benötigen laufende PostgreSQL
- Mocks: pro Repository in `repository/<domain>_repo/mock.go`
- Test ausführen: `go test -tags=unit -race ./...`

## Frontend-Konventionen

### Architektur (pro Feature)

```
src/admin/<feature>/     → Admin-Seiten und -Komponenten
src/service/             → Service-Seiten (Tischauswahl, Tisch-Workflow)
src/lib/                 → Backend-Clients, Auth, Utilities
src/components/ui/       → shadcn/ui-Komponenten
src/components/common/   → Gemeinsame Komponenten (Formulare, EmptyState)
```

### Patterns

- **Alle API-Aufrufe über `Backend`-Singleton** (`src/lib/Backend.ts`) — nie direkt `fetch`
- **`BackendClient`-Interface** aus `src/lib/Backend.ts` — alle Domain-Backend-Klassen verwenden dieses Interface per Dependency Injection statt eigene zu definieren
- **401-Interceptor** — `Backend.post()` erkennt 401-Responses automatisch, loggt den Benutzer aus und leitet zu `/login` weiter
- **Zod-Validierung** für alle Request- und Response-Bodies (Runtime-Typsicherheit)
- **Domain-spezifische Backend-Klassen** (z.B. `ProductBackend`, `TableBackend`) mit Dependency Injection
- **State Management**: Kein globaler Store — React `useState`/`useEffect` + Custom Hooks
- **`useFetch<T>()`-Hook** (`src/lib/useFetch.ts`) — generischer Data-Fetching-Hook mit `loading`, `data`, `error`, `reload`, `setData`. Alle Feature-Hooks (`useAllProducts`, `useActiveTables`, etc.) basieren darauf.
- **Route Guards**: als React Router `loader`-Funktionen (`AdminGuard`, `ServiceGuard`, `AuthRedirect`)
- **Drawer-Pattern**: Alle wichtigen Aktionen (Bestellen, Bezahlen, Stornieren, Liefern) öffnen einen Bottom-Sheet-Drawer mit Zusammenfassung vor Bestätigung. Gemeinsame Hilfsfunktionen (`selectVariants`, `calculateTotalPrice`) in `src/service/components/table/drawerUtils.ts`
- **Error Toasts**: Alle mutativen Aktionen zeigen `toast.error('Aktion fehlgeschlagen')` bei Fehlern (Sonner)
- **UI-Sprache**: Deutsch. Code-Sprache: Englisch.
- **Backend ist Single Source of Truth für Daten-Filterung** — keine redundante Filterlogik im Frontend. Vor dem Hinzufügen von Frontend-Filtern prüfen, ob das Backend die Daten bereits korrekt aufbereitet.

### UI-Komponenten

- shadcn/ui (Radix-basiert) + eigene Compound Components (`Item`, `Field`)
- Icons: Lucide React
- Toasts: Sonner
- Drawers: Vaul

### Geldbeträge im Frontend

- Gespeichert/übertragen als Cent (int)
- Anzeige als Euro über `formatCents()` aus `src/lib/utils.ts` — nie inline formatieren
- Eingabe über spezielle Euro-Eingabefelder mit Umrechnung (`centsToPrice`/`priceToCents` in `FormFields.tsx`)

### Neue Seiten hinzufügen

1. Model + Zod-Schema in `src/admin/<feature>/Model.ts` oder `src/service/`
2. Backend-Client mit `BackendClient`-Interface aus `@/lib/Backend`
3. Hook via `useFetch<T>()` aus `@/lib/useFetch`
4. Komponenten in gleichem Verzeichnis
5. Route in `src/routes.ts` registrieren

### Styling

- Tailwind CSS 4 (`@tailwindcss/vite`), keine `tailwind.config.js`
- CSS-Variablen für Farben in `src/index.css` (Violet/Indigo-Schema)
- Dark Mode via `dark`-Klasse
- `cn()` Utility für bedingte Klassen (`clsx` + `tailwind-merge`)
- Path-Alias: `@/` → `./src/`

## Datenbank

### Schema-Enums

- `UserRole`: `admin`, `service`
- `EntityStatus`: `active`, `inactive`, `deleted`
- `ProductCategory`: `food`, `beverage`, `other`

### Tabellen

- `users` (id, name, username, password_hash, onetime_password_hash, role, status, created_at)
- `tables` (id, name, status, created_at)
- `products` (id, name, category, created_at)
- `product_variants` (id, product_id FK, name, price_cents, status, created_at)
- `events` (id, user_id FK, type, subject, timestamp, data JSONB) — **append-only, kein UPDATE/DELETE**

### Migrationen

- Liegen in `database/migrations/` (SQL-Dateien, nummeriert)
- Werden automatisch beim Container-Start via `migrate`-Container angewendet
- Neue Migration: `<nummer>_<name>.up.sql` und `<nummer>_<name>.down.sql`

## Dokumentation synchron halten

Bei jeder Änderung am Projekt (neue Endpunkte, Seiten, Dependencies, Architektur, Versionen etc.) müssen folgende Dateien aktualisiert werden, sofern betroffen:

- `AGENTS.md` — Tech-Stack, Verzeichnisstruktur, Regeln, Event-Referenz, DB-Schema
- `.github/copilot-instructions.md` — Architektur, Konventionen, Patterns, DB-Tabellen
- `README.md` — Feature-Liste, Architektur-Tabelle
- `DEVELOPMENT.md` — Build-/Test-/Deploy-Befehle, Konfigurationsdateien

Alle Angaben zu Frameworks, Versionen und Tabellen müssen mit dem tatsächlichen Code übereinstimmen. Prüfe nach Änderungen, ob die Dokumentation noch korrekt ist.

## Entwicklung

```bash
# Dev-Stack starten
docker compose -f docker-compose.dev.yml up --build -d

# Frontend: http://localhost
# API: http://localhost/api

# Unit-Tests
cd backend && go test -tags=unit -race ./...

# Lint
cd frontend && pnpm lint
```
