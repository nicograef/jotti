# Agent Instructions — jotti

Dieses Dokument richtet sich an KI-Coding-Agenten (Copilot, Cursor, Cline, Aider, etc.).

## Projektbeschreibung

jotti ist ein Bestell- und Kassensystem für Vereine und Nonprofit-Veranstaltungen. Servicekräfte nehmen auf Smartphones Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer.

## Tech-Stack

| Komponente    | Technologie                                                                 |
| ------------- | --------------------------------------------------------------------------- |
| Backend       | Go 1.25, stdlib `net/http`, `pgx/v5`, `zerolog`, `zog`, `golang-jwt/v5`     |
| Frontend      | React 19, Vite 7, TypeScript 5.9 (strict), Tailwind CSS 4, shadcn/ui, Zod 4 |
| Datenbank     | PostgreSQL 17, `golang-migrate`                                             |
| Infrastruktur | Docker Compose, nginx Reverse Proxy, Let's Encrypt                          |

## Projektstand & Anforderungen

Der vollständige Anforderungskatalog mit Implementierungsvorschlägen liegt in `ANFORDERUNGEN.md`. Aktueller Stand:

| Status       | Anzahl |
| ------------ | ------ |
| ✅ Umgesetzt | 23     |
| ❌ Offen     | 27     |
| **Gesamt**   | **50** |

### Nächste offene Must-haves

| #   | Anforderung                              | Implementierungshinweise                                                                                                    |
| --- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| 23  | Tisch-Schnellsuche per Shortcut          | FAB oder Suchfeld in `TableSelectionPage.tsx`, filtert Tischliste oder navigiert direkt zu `/service/table/:id`             |
| 24  | Übersicht eigene Bestellungen mit Status | Neue Service-Seite `/service/orders`, neuer Endpunkt `POST /service/get-user-orders` (Events nach `user_id` aggregieren)    |
| 25  | Bestellungen auf anderen Tisch umbuchen  | Neuer Event-Typ `table.order-transferred:v1`, Endpunkt `POST /service/transfer-table-order` (Storno + Order in Transaktion) |
| 26  | Umsatz pro Bediener (Tagesabrechnung)    | Neuer Admin-Endpunkt `POST /admin/get-revenue-by-user`, `payment-registered`-Events nach `user_id` aggregieren              |
| 27  | Bons drucken (formatiert)                | ESC/POS oder Web-Print (`window.print()` mit Print-CSS für 58mm/80mm), `Receipt.tsx` als Vorlage                            |
| 31  | Freibon mit freier Preiseingabe          | Spezielle Position mit `variant_id = null` + eigenem Preis/Bezeichnung im Event-Data, oder Freibon-Variante                 |
| 33  | Offline-Fähigkeit                        | `vite-plugin-pwa`, Service Worker, IndexedDB-Queue, Cache-Invalidation — hohe Komplexität                                   |

### Offene Nice-to-haves

- **Rückgeldberechnung (#37)**: Rein clientseitig im `PaymentDrawer` — Eingabefeld "Erhalten", Anzeige Rückgeld.
- **Freitext-Notiz pro Position (#42)**: `LineItem`-Struct um `note: string` erweitern, in `ProductList.tsx` Notiz-Icon pro Variante.
- **Bezeichnung/Name pro Bestellung (#36)**: Optionales `label`-Feld in `OrderPlacedEvent.data`, Textfeld im OrderDrawer.
- **Reporting (#38–40)**: Admin-Seite "Tagesabrechnung" mit Umsatz pro Bediener, Gesamtumsatz, CSV/Excel-Export.

Details und vollständige Implementierungsvorschläge: siehe `ANFORDERUNGEN.md`.

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
   - `.github/copilot-instructions.md` — Architektur, Konventionen, Patterns, DB-Tabellen
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
- **Endpunkte**: `create-product`, `update-product`, `create-variant`, `update-variant`, `activate-variant`, `deactivate-variant`, `get-all-products`, `create-table`, `update-table`, `activate-table`, `deactivate-table`, `get-all-tables`, `create-user`, `update-user`, `activate-user`, `deactivate-user`, `reset-password`, `get-all-users`

### Service-Bereich (Rollen: `admin` + `senior_service` + `service`)

Bestell- und Kassierbetrieb am Tisch. Für Servicekräfte, Serviceleitung und Admins zugänglich.

- **Backend**: Routen in `api/service.go` unter `/service/*`, JWT-Middleware erlaubt Rollen `admin`, `senior_service` und `service`. Stornierung (`cancel-table-variants`) läuft über eigene `api/senior_service.go` mit Middleware nur für `admin` und `senior_service`.
- **Frontend**: Seiten unter `src/service/` mit `ServiceGuard` (React Router Loader)
- **Funktionen**: Tisch auswählen, Bestellungen aufgeben, Lieferungen bestätigen, Bezahlungen registrieren, Stornierungen, Tisch-Verlauf einsehen
- **Endpunkte**: `get-active-products`, `get-active-tables`, `get-table`, `place-table-order`, `register-table-payment`, `cancel-table-variants`, `deliver-table-variants`, `get-table-history`, `get-table-balance`, `get-table-unpaid-variants`, `get-table-undelivered-variants`

### Auth-Bereich (kein JWT erforderlich)

- **Backend**: Routen in `api/auth.go` unter `/auth/*`
- **Endpunkte**: `login`, `set-password`

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

Events für Tisch-Operationen. Subject-Format: `"table:<id>"`.

| Event-Typ                     | Beschreibung              |
| ----------------------------- | ------------------------- |
| `table.order-placed:v1`       | Bestellung aufgegeben     |
| `table.payment-registered:v1` | Bezahlung registriert     |
| `table.variants-delivered:v1` | Varianten geliefert       |
| `table.variants-canceled:v1`  | Varianten storniert       |
| `table.snapshot:v1`           | Materialisierter Snapshot |

State wird durch Replay aller Events rekonstruiert. Snapshots optimieren Lesezugriffe.

### Geplante Event-Typen (noch nicht implementiert)

| Event-Typ                          | Beschreibung                         | Anforderung |
| ---------------------------------- | ------------------------------------ | ----------- |
| `table.order-transferred:v1`       | Bestellung auf anderen Tisch umbucht | #25         |
| `table.variants-prepared:v1`       | Varianten zubereitet / abholbereit   | #35, #45    |
| `table.variants-status-changed:v1` | Zubereitungsstatus geändert          | #46         |

## Datenbank-Schema

Enums: `UserRole(admin, senior_service, service)`, `EntityStatus(active, inactive, deleted)`, `ProductCategory(food, beverage, other)`

Tabellen: `users`, `tables`, `products`, `product_variants`, `events` (append-only)

Neue Migration: `database/migrations/<nr>_<name>.up.sql` + `.down.sql`
