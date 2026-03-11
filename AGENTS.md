# Agent Instructions — jotti

jotti ist ein **kostenloses, quelloffenes Mobile-Kassensystem (mPOS)** für Vereine und gemeinnützige Organisationen. Zielgruppe: eingetragene Vereine (e.V.), gGmbH, gUG, Stiftungen, kirchliche Träger — für temporäre Gastronomie-Veranstaltungen (Vereinsfeste, Weihnachtsmärkte, Maihocks, Konzerte, 2–3 Mal pro Jahr, 5–50 Tische, 5–30 ehrenamtliche Helfer).

Servicekräfte nehmen auf ihren eigenen Smartphones (BYOD) im Browser Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer. Self-hosted per Docker Compose, Source-Available (AGPL-3.0 + Non-Commercial), Mobile-first.

**Bewusst NICHT enthalten:** Kartenzahlung, TSE/KassenSichV, Reservierungen, Warenwirtschaft, Lieferservice, Multi-Standort, CRM, Kiosk-Modus. Diese Reduktion ist gewollt — jedes zusätzliche Feature erhöht Komplexität für ehrenamtliche Teams.

Weiterführende Docs: [Produktbeschreibung](docs/produktbeschreibung.md) · [Anforderungen](docs/requirements.md) · [Implementierungsplan](docs/implementation-plan.md) · [Entwicklung](docs/development.md) · [Ubiquitous Language](docs/language.md) · [Event Storming](docs/event-storming.md) · [Lizenz](docs/lizenz-und-nutzung.md)

## Tech-Stack

| Komponente    | Technologie                                                                |
| ------------- | -------------------------------------------------------------------------- |
| Backend       | Go, stdlib `net/http`, `pgx/v5`, `sqlc`, `zerolog`, `zog`, `golang-jwt/v5` |
| Frontend      | React, Vite, TypeScript (strict), Tailwind CSS 4, shadcn/ui, Zod           |
| Datenbank     | PostgreSQL, `golang-migrate`                                               |
| Infrastruktur | Docker Compose, nginx Reverse Proxy, Let's Encrypt                         |

## Wichtige Regeln

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE.
2. **Geldbeträge sind immer in Cent (int).** Niemals Floats für Geld verwenden.
3. **Event-Sourcing für Tisch-Operationen.** Events sind immutable (append-only). Nie Events updaten oder löschen.
4. **CRUD für Stammdaten** (Benutzer, Produkte, Tische). Soft-Deletes via `status = 'deleted'`.
5. **Validierung mit Schemas.** Backend: `zog`. Frontend: `Zod`. Beide Seiten validieren.
6. **Deutsche Ubiquitous Language.** Fachbegriffe der Domäne sind deutsch (Bestellung, Zahlung, Lieferung, Stornierung, Tisch, Position). Infrastruktur-Code (Auth, Config, DB) bleibt englisch. Alle Benutzer-sichtbaren Strings auf Deutsch. Commits auf Englisch.
7. **Kein globaler State-Store im Frontend.** Nur React Hooks + Singletons.
8. **Frontend API-Aufrufe nur über Backend-Klassen.** Nie direkt `fetch()` verwenden. Alle Domain-Backend-Klassen nutzen das `BackendClient`-Interface aus `src/lib/Backend.ts`.
9. **Dokumentation synchron halten.** Bei Änderungen diese Dateien aktualisieren, sofern betroffen: `AGENTS.md`, `README.md`, `docs/development.md`, `docs/requirements.md`, `docs/implementation-plan.md`, `docs/language.md`.
10. **Backend ist die Single Source of Truth für Daten-Filterung.** Filterung, Aggregation und Aufbereitung gehören ins Backend. Das Frontend zeigt an, was das Backend liefert.

## Bereiche

- **Admin** (`admin`): Routen `/admin/*` (`api/admin.go`), Frontend `src/admin/`, `AdminGuard`. Produkte, Tische, Benutzer verwalten.
- **Service** (`admin` + `senior_service` + `service`): Routen `/service/*` (`api/service.go`), Stornierung über `api/senior_service.go`. Frontend `src/service/`, `ServiceGuard`. Bestellen, Liefern, Kassieren, Stornieren.
- **Auth** (kein JWT): Routen `/auth/*` (`api/auth.go`). Login, Passwort setzen.

## Tests

```bash
cd backend && go test -tags=unit -race ./...   # Unit-Tests
./test-integration.sh                           # Integrationstests
cd frontend && pnpm lint                        # Frontend-Lint
```
