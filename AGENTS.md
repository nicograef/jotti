# Agent Instructions — jotti

jotti ist ein **kostenloses, quelloffenes Mobile-Kassensystem (mPOS)** für Vereine und gemeinnützige Organisationen. Zielgruppe: eingetragene Vereine (e.V.), gGmbH, gUG, Stiftungen, kirchliche Träger — für temporäre Gastronomie-Veranstaltungen (Vereinsfeste, Weihnachtsmärkte, Maihocks, Konzerte, 2–3 Mal pro Jahr, 5–50 Tische, 5–30 ehrenamtliche Helfer).

Servicekräfte nehmen auf ihren eigenen Smartphones (BYOD) im Browser Bestellungen auf, bestätigen die Ausgabe, kassieren und stornieren — alles pro Tisch. Admins verwalten Produkte, Tische und Benutzer. Self-hosted per Docker Compose, proprietäre Source-Available-Lizenz (Non-Commercial, Nutzungsvereinbarung erforderlich), Mobile-first.

**Bewusst NICHT enthalten:** Kartenzahlung, Reservierungen, Warenwirtschaft, Lieferservice, Multi-Standort, CRM, Kiosk-Modus. Diese Reduktion ist gewollt — jedes zusätzliche Feature erhöht Komplexität für ehrenamtliche Teams.

**Compliance-Roadmap (TSE/KassenSichV):** jotti ist ein elektronisches Aufzeichnungssystem im Sinne von § 1 KassenSichV und unterliegt damit der TSE-Pflicht nach § 146a AO. Die TSE-Integration (fiskaly Cloud-TSE, DSFinV-K-Export, ELSTER-Meldepflicht) wird schrittweise implementiert — siehe `docs/roadmap.md` und `docs/compliance.md`.

## Referenzdokumente

Die folgenden Dokumente beschreiben jotti vollständig. Sie werden **nicht automatisch geladen** (zu groß). Bevor du eine Aufgabe beginnst, prüfe ob du Kontext aus einem dieser Dokumente brauchst — und lies dann gezielt den relevanten Abschnitt per `read_file`, nicht das ganze Dokument.

### Anforderungen — `docs/anforderungen.md`

Alle funktionalen und querschnittlichen Anforderungen mit Akzeptanzkriterien, Priorisierung (Must/Should/Nice-to-have) und Status (✅/🔲/🚫).

- **§1 Kassenbetrieb:** K-01 Bestellung aufnehmen, K-02 Zahlung kassieren, K-03 Ausgabe bestätigen, K-04 Stornierung, K-05 Auszahlung, K-06 Tischübersicht, K-07 Kassenjournal, K-08–K-14 (Umbuchung, Rückgeld, Schnellsuche, Bondruck, KDS, Ausgabestationen)
- **§2 Stammdaten:** S-01 Produktverwaltung, S-02 Tischverwaltung, S-03 Benutzerverwaltung
- **§3 Auth:** A-01 Login, A-02 Passwort setzen, A-03 Logout
- **§4 Querschnitt:** Q-01 Mobile-first, Q-02 Mehrbenutzerfähigkeit, Q-03 Validierung, Q-04 Datenintegrität, Q-05–Q-08 (Offline, HTTPS, Rate Limiting, Security Headers)
- **§5 Reporting:** Tagesabrechnung, Abrechnung pro Tisch/Servicekraft, Datenexport
- **§6 Bewusste Abgrenzung:** Won't-haves mit Begründung
- **§7 Fiskalkonformität:** F-01–F-08 (Seriennummer, TSE-Adapter, Belegausgabe, DSFinV-K, ELSTER, Abrechnungskreis, Steuersätze, GoBD Hash-Chain)

→ Lesen bei: neue Features implementieren, Akzeptanzkriterien prüfen, Rollen/Berechtigungen klären, Compliance-Anforderungen umsetzen.

### Entwickler-Handbuch — `docs/handbuch.md`

Architektur, Bounded Contexts, Domain-Modelle, Invarianten, Event-Sourcing-Details.

- **§1 Überblick:** Systemvision, Designziele, bewusste Abgrenzung
- **§2 Bounded Contexts:** Kontextübersicht (Kasse, Stammdaten, Auth), Beziehungen (ACL, Fat Events)
- **§3 Kasse (Core Domain):** Tisch-Session, Kassensitzung, Invarianten (Saldo, Ausgabe-, Bezahl-, Stornierungsinvariante, KS-Sperre), Domain Events (BestellungAufgenommen, AusgabeBestaetigt, ZahlungKassiert, StornierungErteilt, KassensitzungEroeffnet u. a.), Kassenjournal, Synchrone Projektion (tisch_session_state), CRUD-Entität (kassensitzungen)
- **§4 Stammdaten:** Produkt-Aggregat (Varianten, Kategorien), Tisch-Stammdaten, Benutzer-Aggregat, CRUD-Persistenz
- **§5 Auth und Rollen:** Berechtigungsmatrix, Onboarding-Ablauf
- **§6 Architekturprinzipien:** Schichtenarchitektur, API-Design, Frontend-Architektur, Validierung, Geldbeträge, OCC, Sicherheit
- **§7 Read Models:** Service-Read-Models (Tisch-Saldo, Unbezahlt, Ausstehend, Historie)

→ Lesen bei: Architekturentscheidungen, Invarianten prüfen, Event-Strukturen verstehen, neue Schichten/Endpunkte entwerfen.

### Ubiquitous Language — `docs/language.md`

Verbindliche Referenz für Fachbegriffe, Code-Repräsentationen und Namenskonventionen.

- **Sprachkonventionen:** Domänenbegriffe deutsch, Infrastruktur englisch, UI deutsch, Commits englisch
- **Namenskonventionen pro Schicht:** Go-Structs, TS-Typen, JSON-Keys, API-Pfade, DB-Tabellen, Frontend-Routen
- **Abweichungen Ist/Soll:** Aktueller Rename-Status (Backend ✅, Frontend ⏳)
- **Begriffsdefinitionen:** Tisch, Bestellung, Position, Ausgabe, Zahlung, Stornierung — jeweils mit Go-Struct, TS-Typ, JSON-Keys, API-Pfad, Frontend-Komponente, UI-Labels

→ Lesen bei: Benennungen klären, neue Felder/Typen benennen, Ist/Soll-Abweichungen prüfen.

### Produktbeschreibung — `docs/produktbeschreibung.md`

Produktidentität, Positionierung, Zielgruppe, Abgrenzung, Marketingtexte.

- **§1–§3:** Claim, Positionierung, Zielgruppe (Personas: Thomas/Admin, Maria/Service, Felix/Serviceleitung)
- **§4–§5:** Problemstellung, Lösung
- **§6–§8:** Kernfeatures, Abgrenzung vs. kommerzielle POS, USPs
- **§9–§10:** Lizenz/Kosten, Marketingtexte

→ Lesen bei: README/Docs/Marketing-Texte anpassen, Abgrenzungsfragen, Zielgruppe verstehen.

### Compliance — `docs/compliance.md`

Rechtliche Grundlagen, Compliance-Anforderungen und Betreiberpflichten.

- **§1 Einleitung:** Rollentrennung Entwickler/Betreiber, Pflichten nach § 146a AO
- **§2 Rechtliche Grundlagen:** KassenSichV, GoBD, DSFinV-K, ELSTER/ERiC
- **§3 Anforderungen an den Entwickler:** TSE-Adapter-Interface, DSFinV-K, Seriennummer, Steuersätze
- **§4 Betreiberpflichten:** BYOT (Cloud-TSE), ELSTER-Meldung, GoBD-Aufbewahrung
- **§5 Architekturprinzipien:** Adapter-Pattern, BYOD-Offline-Blocking

→ Lesen bei: Compliance-Features implementieren, Betreiber-Dokumentation schreiben, rechtliche Anforderungen prüfen.

### Roadmap — `docs/roadmap.md`

Phasenplan für die schrittweise Compliance-Implementierung.

- **Phase 0 — Baseline:** Aktueller Stand (Event-Sourcing, Kassenjournal, Bondrucker)
- **Phase 1 — Compliance-Grundlage:** Seriennummer, Steuersätze, Abrechnungskreis, Belegausgabe, ELSTER-Anleitung
- **Phase 2 — TSE-Integration:** TSEClient (fiskaly), StartTransaction/FinishTransaction, DSFinV-K-Export
- **Phase 3 — Erweiterungen:** eBeleg, ERiC-Schnittstelle, GoBD Hash-Chain

→ Lesen bei: nächste Compliance-Schritte planen, Feature-Priorisierung, Implementation von F-01 bis F-08.

## Tech-Stack

| Komponente    | Technologie                                                                     |
| ------------- | ------------------------------------------------------------------------------- |
| Backend       | Go 1.26, stdlib `net/http`, `pgx/v5`, `sqlc`, `zerolog`, `zog`, `golang-jwt/v5` |
| Frontend      | React 19, Vite 7, TypeScript 5.9 (strict), Tailwind CSS 4, shadcn/ui, Zod 4     |
| Datenbank     | PostgreSQL 17, `golang-migrate`                                                 |
| Runtime       | Node 24+, pnpm 10+                                                              |
| Infrastruktur | Docker Compose, nginx Reverse Proxy, Let's Encrypt                              |

## Befehle

Alle Befehle werden über das **Makefile** im Projekt-Root ausgeführt:

| Befehl                  | Beschreibung                                 |
| ----------------------- | -------------------------------------------- |
| `make test`             | Backend Unit-Tests                           |
| `make test-frontend`    | Frontend Tests (Vitest)                      |
| `make test-all`         | Alle Unit-Tests (Backend + Frontend)         |
| `make test-integration` | Integrationstests                            |
| `make lint`             | Backend + Frontend Linting                   |
| `make fmt`              | Backend + Frontend Formatierung              |
| `make build`            | Backend + Frontend kompilieren               |
| `make check`            | Schnelle Repo-Prüfung ohne DB-Integration    |
| `make check-full`       | Vollständige Prüfung inkl. Integrationstests |
| `make verify`           | Alias für `make check-full`                  |
| `make sqlc`             | sqlc Code generieren (nach Query-Änderungen) |
| `make dev`              | Dev-Stack starten (Docker Compose)           |
| `make down`             | Dev-Stack stoppen                            |
| `make prod-init`        | Ersteinrichtung Produktion (Zertifikate)     |
| `make prod-up`          | Produktions-Stack starten                    |
| `make prod-down`        | Produktions-Stack stoppen                    |

Siehe `make help` für die vollständige Liste.

## Aktive Entwicklungsphase

jotti befindet sich in aktiver Entwicklung (Pre-Release). **Breaking Changes sind ausdrücklich erwünscht** — es gibt keine produktiven Instanzen und keine Nutzer, auf die Rücksicht genommen werden muss.

- **DB-Schema:** Änderungen direkt in `database/migrations/01_initial.up.sql` vornehmen. Keine neuen Migrationsdateien, keine Down-Migrationen. Dev-DB bei Bedarf neu aufsetzen (`make down && make dev`).
- **Backend-API:** Endpunkte, Request-/Response-Formate und JSON-Keys direkt ändern. Keine API-Versionierung, keine Migrations-Strategien.
- **Event-Formate:** Event-Data-Strukturen und JSON-Keys direkt ändern. Kein Dual-Read, kein Custom `UnmarshalJSON` für alte Daten. Alte Events werden nicht migriert.
- **Frontend:** Typen, Schemas und Komponenten direkt an geänderte Backend-Datenformate anpassen.

## Wichtige Regeln

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE.
2. **Geldbeträge sind immer in Cent (int).** Niemals Floats für Geld verwenden.
3. **Event-Sourcing für Kasse-Operationen.** Das Kassenjournal (`kassenjournal`-Tabelle) ist immutable (append-only). Nie Einträge im Kassenjournal updaten oder löschen. Eine synchrone Projektion (`tisch_session_state`) und eine CRUD-Entität (`kassensitzungen`) werden in derselben Transaktion aktualisiert.
4. **CRUD für Stammdaten** (Benutzer, Produkte, Tische). Soft-Deletes via `status = 'deleted'`.
5. **Validierung mit Schemas.** Backend: `zog`. Frontend: `Zod`. Beide Seiten validieren.
6. **Deutsche Ubiquitous Language.** Fachbegriffe der Domäne sind deutsch (Bestellung, Zahlung, Ausgabe, Stornierung, Tisch, Position). Infrastruktur-Code (Auth, Config, DB) bleibt englisch. Alle Benutzer-sichtbaren Strings auf Deutsch. Commits auf Englisch.
7. **Kein globaler State-Store im Frontend.** Nur React Hooks + Singletons.
8. **Frontend API-Aufrufe nur über Backend-Klassen.** Nie direkt `fetch()` verwenden. Alle Domain-Backend-Klassen nutzen das `BackendClient`-Interface aus `src/lib/Backend.ts`.
9. **Backend ist die Single Source of Truth für Daten-Filterung.** Filterung, Aggregation und Aufbereitung gehören ins Backend. Das Frontend zeigt an, was das Backend liefert.
10. **Domain-Modelle tragen keine `json`-Tags.** Die Domain-Schicht (`domain/`) kennt kein HTTP und keine Serialisierung. `json`-Tags gehören ausschließlich in Response-DTOs der HTTP-Schicht (`api/<domain>/http/`) und in Event-Data-Structs (für Event-Store-Persistenz). Domain-Structs werden nie direkt als API-Response serialisiert.

## Bereiche

- **Admin** (`admin`): Routen `/admin/*` (`api/admin.go`), Frontend `src/admin/`, `AdminGuard`. Produkte, Tische, Benutzer verwalten. Kassensitzung eröffnen/verwalten (künftig).
- **Service** (`admin` + `serviceleitung` + `service`): Routen `/service/*` (`api/service.go`), Stornierung über `api/serviceleitung.go`. Frontend `src/service/`, `ServiceGuard`. Bestellen, Ausgabe bestätigen, Kassieren, Stornieren.
- **Auth** (kein JWT): Routen `/auth/*` (`api/auth.go`). Login, Passwort setzen.

## Grenzen

- ✅ **Immer:** Beide Seiten validieren (zog + Zod), Tests mitliefern, Events immutable behandeln
- ✅ **Immer:** `make sqlc` nach Query-Änderungen, `make lint` nach Code-Änderungen
- ✅ **Immer:** Response-DTOs in der HTTP-Schicht definieren, Domain-Modelle nie direkt serialisieren
- ⚠️ **Erst fragen:** Neue Dependencies hinzufügen, Docker/Nginx-Konfiguration ändern
- 🚫 **Niemals:** `sqlc/dbgen/` editieren (generierter Code)
- 🚫 **Niemals:** Einträge im Kassenjournal updaten oder löschen
- 🚫 **Niemals:** Floats für Geldbeträge verwenden
- 🚫 **Niemals:** `json`-Tags auf Domain-Structs in `domain/` setzen (Ausnahme: Event-Data-Structs für Event Store)
- 🚫 **Niemals:** Domain-Modelle direkt als HTTP-Response durchschleusen
- 🚫 **Niemals:** Direkt `fetch()` im Frontend verwenden
- 🚫 **Niemals:** GET/PUT/DELETE-Endpunkte erstellen
- 🚫 **Niemals:** Secrets oder Passwörter in Code committen

## Git-Workflow

- **Commit-Messages:** Conventional Commits auf Englisch (`feat:`, `fix:`, `refactor:`, `docs:`, `test:`)
- **Kein auto-commit.** Agent schlägt Commit-Message vor, User führt Commit durch.
- **Kein `--force` push oder `--no-verify`.**

## Lokale Qualitaetspruefung

Fuer reproduzierbare lokale Checks (CI-nah):

```bash
bash scripts/setup-dev-tools.sh
make check
make verify
```

- `make check`: schneller Gesamt-Check ohne Integrationstests
- `make verify`: voller Check inkl. Integrationstests

### Troubleshooting fehlende Tools

Wenn `make check` oder `make verify` frueh mit `Fehlendes Tool: ...` abbricht:

1. `bash scripts/setup-dev-tools.sh` erneut ausfuehren.
2. Sicherstellen, dass Go-Bin-Verzeichnis im `PATH` liegt:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

3. Tool-Versionen pruefen:

```bash
go version
node --version
pnpm --version
goimports -V
golangci-lint --version
```
