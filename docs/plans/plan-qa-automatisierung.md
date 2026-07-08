# Plan: QA-Automatisierung für v1.0.0 und den laufenden Betrieb

> Quell-PRD: [prd-qa-automatisierung.md](../prds/prd-qa-automatisierung.md)

## Ziel

Die automatisierbaren Teile der v1.0.0-QA als dauerhafte Artefakte bauen
(E2E-Suite, DSFinV-K-Validator, Berechtigungs-Matrix, Scans, Fuzz-Targets,
TSE-Live-Suite, Ops-Smoke-Skript), den manuellen QA-Guide auf echte
Handarbeit reduzieren und die QA einmal vollständig durchführen. Ergebnis:
jeder Pull Request durchläuft denselben Sicherheitsstandard wie das
Release, und die verbleibende Handarbeit passt in einen Nachmittag.

## Architekturentscheidungen

Dauerhafte Entscheidungen, die für alle Phasen gelten:

- **E2E-Verzeichnis**: eigenes Top-Level-Verzeichnis `e2e/` (eigenes
  pnpm-Paket, Playwright + TypeScript). Zwei Playwright-Projekte:
  Desktop-Chromium (Admin) und Mobile-Chromium (Service). Die Suite ist
  BASE_URL-agnostisch (`E2E_BASE_URL`, Default: laufender Dev-Stack auf
  `http://localhost`).
- **E2E-Stack in CI**: eigene `docker-compose.e2e.yml`, die die echten
  Images aus dem Checkout baut (Backend, Frontend-nginx, Migrate,
  Reverse-Proxy) — jeder PR testet die Auslieferungs-Artefakte. Lokal
  läuft dieselbe Suite gegen den Dev-Stack.
- **Seed-Reset**: neuer Endpoint `POST /test/reset-and-seed`, nur
  registriert wenn `JOTTI_ALLOW_SEED=1` (in Prod/Release nie gesetzt,
  gleiche Guard-Logik wie das seed-Subkommando). Leert Kassenjournal,
  Projektionen und Stammdaten und seedet über die bestehende Seed-Engine
  neu. Jede Spec-Datei startet damit.
- **make-Targets**: `test-e2e` (E2E gegen `E2E_BASE_URL`), `fuzz` (lange
  Fuzz-Läufe, lokal), `test-tse-live` wird zur vollen Live-Suite
  ausgebaut (fährt Wegwerf-Postgres + Migrationen nach dem Muster von
  `scripts/test-integration.sh`), `test-tse-live-setup` bleibt separates
  Opt-in. `make verify` bleibt unverändert; E2E ist ein eigenes
  Pflicht-CI-Gate daneben.
- **DSFinV-K-Validator**: eigenständiges Go-Paket im Backend-Modul neben
  `seed/` (Arbeitsname `backend/dsfinvkcheck`; endgültiger Name nach
  language.md in Phase 5). Schnittstelle: Export-ZIP rein, Befundliste
  raus. Keine neue Dependency für DTD-Validierung (siehe Risiken).
- **Berechtigungs-Matrix**: die Routen-Registrierung wird deklarativ.
  Eine Routentabelle (Pfad, erlaubte Rollenmenge) ist die einzige Quelle,
  aus der `SetupRoutes` registriert und der table-driven Integrationstest
  sich speist. Eine Route ohne Deklaration kann nicht existieren.
- **CI-Jobs**: `e2e` (jeder PR, mit Trace/Screenshot-Artefakten bei Rot),
  `vuln-scan-go` (govulncheck über alle Go-Module) und `vuln-scan-npm`
  (pnpm audit) zusätzlich mit wöchentlichem Schedule, damit neue
  Advisories ohne Code-Änderung auffallen.
- **Befund-Report**: `docs/plans/befund-report-qa-v1.0.0.md`
  (Arbeitsdokument, nach Abarbeitung der Fixes löschen).

## Inventar

Bestehende Muster und Anknüpfungspunkte:

- `Makefile:47-64` — Test-Targets; `56-62` die beiden TSE-Live-Targets
  mit `.env.fiskaly-test`-Check; `296-303` check/check-full/verify-Kette
- `.github/workflows/ci.yml:9-46` — paths-filter-Job (neue Filter für
  `e2e/` nötig); `326-334` shellcheck-Job; `336-392`
  Integrationstest-Job (Postgres-Service, `-p 1`); `401-497`
  Upgrade-Pfad-Job (bleibt unangetastet)
- `backend/app/app.go:45-86` — `SetupRoutes` mit Rollen-Middleware je
  Bereich; Zeile 76: Security-Header setzt der Reverse-Proxy, nicht das
  Backend
- `backend/api/admin.go:33-69`, `api/service.go`,
  `api/serviceleitung.go:12-32`, `api/auth.go:12-18`,
  `api/relay.go:44-53` — heutige imperative Routen-Registrierung
- `backend/api/middleware/middleware.go` — JWT-, RateLimit-, Recovery-,
  PostOnly-Middleware samt Tests
- `backend/main.go:55-70` — Subkommandos `seed`/`rebuild-projections`
  mit `JOTTI_ALLOW_SEED`-Guard
- `backend/seed/` — Engine (`engine.go`), 3-Tage-Drehbuch
  (`szenario.go:296-300`), Guard (`guard.go`), Fake-TSE (`faketse.go`),
  Integrationstest-Muster (`seed_integration_test.go`)
- `backend/repository/tse_repo/fiskaly_client_live_test.go:21-52` —
  Live-Test-Muster: Skip ohne `FISKALY_TEST_*`, Abbruch wenn Umgebung
  nicht TEST
- `backend/api/fiskal/dsfinvk/` — Export-Erzeuger (archive.go, table.go,
  mapper.go, index.xml, beiliegende `gdpdu-01-09-2004.dtd`)
- `backend/api/fiskal/export/application/export.go` und
  `export/http/handler.go` — Export-Anwendungsdienst und
  Download-Endpoint
- `backend/api/fiskal/signatur/` — TSE-Signatur-Worker und Watchdog
  (gestartet in `app.go:90-94`)
- `backend/api/druck/bondruck/application/escpos/formatter.go` —
  ESC/POS-Encoder (Fuzz-Target)
- `backend/domain/event/event.go:104` — Event-Daten-Deserialisierung,
  die Replay-Kante (Fuzz-Target)
- `backend/domain/kasse/event_json_contract_test.go` —
  Contract-Guard-Muster
- `docker-compose.yml` — Dev-Stack: Frontend auf `:80` (Vite), Backend
  per `go run`, `JOTTI_ALLOW_SEED=1` gesetzt
- `docker-compose.prod.yml` — Prod-Topologie: postgres, migrate,
  backend, frontend (nginx), reverse-proxy (Caddy) auf 80/443
- `reverse-proxy/caddyfile.go:81-128` — generierte Caddyfile:
  `/api/*` → backend:3000, Rest → frontend:80; `caddyfile_test.go` als
  Ort für Header-Assertions
- `frontend/Dockerfile` + `frontend/nginx.conf` — Prod-Frontend-Artefakt
- `frontend/src/routes.ts` — vollständige Routenkarte (Basis für den
  Screen-Sweep in Phase 14)
- `scripts/test-integration.sh` — Wegwerf-Postgres-Muster (Start,
  TCP-Wartelogik, Migrationen, Cleanup-Trap)
- `scripts/prod-init.sh`, `prod-update.sh`, `prod-backup.sh`,
  `prod-backup-verify.sh` — die Ops-Pfade, die der Smoke orchestriert
- `docs/plans/guide-manuelle-qa-v1.0.0.md` — der umzubauende Guide
  (Blöcke 1–8)
- `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/` —
  Spezifikation samt DTD und Beispielen
- `.env.fiskaly-test.example` — Vorlage der gitignorierten
  Credentials-Datei

## Geklärte Entscheidungen

Aus der Klärungsrunde (2026-07-08):

- E2E lokal gegen den Dev-Stack, in CI gegen die aus dem Checkout
  gebauten Prod-Images; die Suite selbst ist BASE_URL-agnostisch.
- Seed-Reset per Test-Endpoint im Backend (`JOTTI_ALLOW_SEED`-gated),
  nicht per docker exec aus Playwright.
- E2E bleibt eigenes Target neben `make verify`; „grün" heißt künftig
  verify plus E2E-Job in CI.
- Der reale Ops-Smoke-Lauf wird in die QA-Durchführung (Phase 14)
  gebündelt; ein einziger Wegwerf-Host deckt Skript-Verifikation und
  QA-Blöcke ab. Die Skript-Phase endet mit shellcheck + Review.
- Phasenzuschnitt mit 14 dünnen Phasen bestätigt.

## Offene Fragen / Risiken

- **Security-Header liegen beim Reverse-Proxy**, nicht im Backend
  (`app.go:76`). Die „schnelle Ebene" aus dem PRD wird deshalb als
  Unit-Test der Caddyfile-Generierung im reverse-proxy-Modul umgesetzt
  (Header-Direktiven im generierten Caddyfile), die reale Ebene prüft
  das Ops-Smoke-Skript am deployten Stack. Die Middleware-Ebene im
  Backend deckt nur das Login-Rate-Limit.
- **DTD-Validierung in Go**: die Standardbibliothek validiert kein DTD,
  und eine cgo-Bindung (libxml2) kommt nicht in Frage. Der Validator
  implementiert die DTD-Regeln (Elemente, Reihenfolge, Pflichtattribute)
  als eigene Prüfungen gegen die beiliegende DTD als Referenz. Sollte
  sich eine reine Go-Bibliothek anbieten: vorher fragen (Regel 16).
- **Vite-Dev-Server vs. nginx**: Specs könnten lokal grün und in CI rot
  sein (oder umgekehrt). Bei Divergenz gilt der CI-Lauf gegen die
  Prod-Images; wartende Assertions und Testing-Library-Selektoren
  minimieren das Risiko.
- **IDOR-Zuschnitt**: jotti ist ein Ein-Mandanten-System; „fremde
  Objekt-IDs" heißt hier Zugriff über die eigene fachliche Abgrenzung
  hinaus (z. B. Service-Rolle ruft Serviceleitungs-Storno auf) plus
  definiertes Verhalten bei fremden/nicht existenten IDs. Der genaue
  Fallkatalog entsteht in Phase 7 aus der Routentabelle.
- **pnpm audit** kann an nicht behebbaren Advisories hängen bleiben:
  Schwelle high+, dokumentierte Ausnahmen im Workflow, die bestehende
  Supply-Chain-Policy (Mindestalter) bleibt unberührt.
- **fiskaly-TEST-Verfügbarkeit**: die Live-Suite hängt an einer externen
  API; genau deshalb bleibt sie opt-in und außerhalb von CI.

---

## Phase 1: E2E-Fundament (Tracer Bullet)

**User Stories**: 2 (CI bei jedem PR), 3 (ein Befehl lokal), 4 (Seed-Zustand)

### Kontext

- `docker-compose.yml` — Dev-Stack, Frontend auf `:80`, `JOTTI_ALLOW_SEED=1`
- `backend/main.go:55-70` + `backend/seed/guard.go` — Guard-Logik, die der Endpoint übernimmt
- `backend/app/app.go:45-86` — Anschlussstelle für die Test-Route
- `reverse-proxy/caddyfile.go:81-128` — Routing, das die e2e-Compose abbilden muss
- `.github/workflows/ci.yml:9-46` — paths-filter, neuer Filter für `e2e/`

### Was zu bauen ist

Der komplette Durchstich mit genau einer Spec: `e2e/` als eigenes
pnpm-Paket mit Playwright, zwei Viewport-Projekte, Reset über den neuen
`POST /test/reset-and-seed`-Endpoint (nur bei `JOTTI_ALLOW_SEED=1`
registriert, leert und seedet in einer Transaktion), `make test-e2e`
gegen `E2E_BASE_URL`, `docker-compose.e2e.yml` mit den echten Images und
ein CI-Job, der Stack baut, Suite fährt und bei Rot Trace + Screenshot
als Artefakt hochlädt. Die erste Spec: Anmelden als Servicekraft, eine
Bestellung aufnehmen, kassieren, sichtbaren Betrag asserten. Spec-Namen
folgen der Fachsprache aus language.md; keine festen Wartezeiten.

### Akzeptanzkriterien

- [ ] `make test-e2e` läuft lokal gegen den laufenden Dev-Stack und ist grün
- [ ] CI-Job baut die Prod-Images aus dem Checkout, startet den Stack und führt die Suite bei jedem PR aus; Trace/Screenshot als Artefakt bei Fehlschlag
- [ ] Reset-Endpoint existiert nur bei `JOTTI_ALLOW_SEED=1`; ohne Flag ist die Route nicht registriert (Test), Seed-Guards bleiben intakt
- [ ] Die erste Spec startet vom Seed-Zustand und läuft in beiden Viewport-Projekten
- [ ] `make verify` unverändert grün

---

## Phase 2: E2E Service-Kernflows

**User Stories**: 1 (Kernflows), 8 (Handy-Viewport)

### Kontext

- `frontend/src/service/` — TablePage, DirektverkaufPage, Drawers
- `docs/plans/guide-manuelle-qa-v1.0.0.md:27` — Block-3-Liste der Geschäftsvorfälle als Checkliste
- `backend/api/kasse/` — tischgeschaeft, direktverkauf, kassenfuehrung

### Was zu bauen ist

Specs für die Servicekraft-Flows im Mobile-Projekt: Bestellen, Ausgeben,
Kassieren (Teil- und Vollzahlung), Direktverkauf und dessen Storno,
Stornierung über die Serviceleitung (geldneutrale Korrektur und
Warenrücknahme), Umbuchen auf einen anderen Tisch, Kassenabschluss.
Innerhalb einer Spec-Datei dürfen Schritte aufeinander aufbauen (der
Kassenabschluss braucht vorherige Umsätze); Spec-Dateien untereinander
bleiben unabhängig. Assertions ausschließlich auf Sichtbares: Beträge,
Positionszustände, Salden, Abschlussmeldung.

### Akzeptanzkriterien

- [ ] Jeder Flow aus User Story 1 hat eine Spec im Handy-Viewport, die vom Seed-Zustand startet
- [ ] Der Kassenabschluss-Test erzeugt seine Umsätze selbst und assertet die sichtbare Abschlussmeldung
- [ ] Beide Storno-Arten und die Teilzahlung sind abgedeckt
- [ ] CI-E2E-Job weiterhin grün

---

## Phase 3: E2E Admin-Flows und Export-Download

**User Stories**: 6 (Admin-Flows), 7 (Export-Download)

### Kontext

- `frontend/src/admin/` — products, tables, users, kasse, reporting, tse, finanzamt, settings
- `backend/api/fiskal/export/http/handler.go` — Download-Endpoint, den der Klickpfad erreicht

### Was zu bauen ist

Specs für die Verwaltungsseiten im Desktop-Projekt: Produkt samt
Variante anlegen/ändern/deaktivieren, Tisch anlegen/ändern, Benutzer
anlegen (inkl. Einmalpasswort-Anzeige) und deaktivieren, Druckstation
verwalten, Kassenführung (Kassensturz/Tagesabschluss aus Admin-Sicht),
Live-Reporting zeigt die Umsätze der Seed-Daten. Dazu der
Export-Download: Klick in der Finanzamt-Ansicht bis zur
heruntergeladenen ZIP-Datei (Playwright-Download-Event, Datei nicht
leer; die inhaltliche Prüfung übernimmt der Validator ab Phase 5).

### Akzeptanzkriterien

- [ ] Jeder Admin-Bereich (Produkte, Tische, Benutzer, Druckstationen, Kassenführung, Reporting) hat mindestens eine Spec mit sichtbaren Assertions
- [ ] Export-Spec lädt eine nicht-leere ZIP-Datei über die Oberfläche herunter
- [ ] Specs laufen im Desktop-Projekt und starten vom Seed-Zustand
- [ ] CI-E2E-Job weiterhin grün

---

## Phase 4: E2E Fehlerpfade

**User Stories**: 5 (sichtbare Fehler statt Leer-Defaults)

### Kontext

- `frontend/src/lib/Backend.ts` — zentrale API-Schicht, deren Fehlerverhalten sichtbar werden muss
- Praxistest-Befund: stille Leer-Defaults wie Saldo 0,00 bei Serverfehlern

### Was zu bauen ist

Specs, die per Playwright-Route-Interception Serverfehler (5xx) und
Netzabbruch (abgebrochene Requests) auf den wichtigsten Screens
simulieren: Tischübersicht, Tisch-Detail, Kassieren-Drawer,
Admin-Dashboard/Reporting. Assertion jeweils: ein sichtbarer
Fehlerzustand (Meldung oder Fehler-Referenz), niemals ein stiller
Leer-Default wie Saldo 0,00 oder eine leere Liste ohne Hinweis.

### Akzeptanzkriterien

- [ ] Serverfehler- und Netzabbruch-Spec je für Service-Screens und Admin-Reporting
- [ ] Jede Spec assertet einen sichtbaren Fehlerzustand; Leer-Defaults gelten als Fehlschlag
- [ ] CI-E2E-Job weiterhin grün

---

## Phase 5: DSFinV-K-Validator — Paket und Strukturprüfung

**User Stories**: 9 (Strukturvalidator), 11 (Integrationstest in verify)

### Kontext

- `backend/api/fiskal/dsfinvk/` — Erzeuger: table.go (Spalten), archive.go (ZIP), index.xml, DTD
- `backend/api/fiskal/export/application/export.go` — erzeugt das ZIP, das der Integrationstest durchschickt
- `backend/seed/` — Seed-Engine mit Fake-TSE für gefüllte TSE-Felder
- `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/` — Spezifikation, Anhänge, DTD
- `scripts/test-integration.sh` + `ci.yml:336-392` — wo der Integrationstest mitläuft

### Was zu bauen ist

Das Validator-Paket mit schmaler Schnittstelle (Export-ZIP rein,
Befundliste raus) und den Strukturregeln: Dateinamen englisch und
kleingeschrieben, CSV-Regeln (Semikolon, CRLF, Komma-Dezimaltrennung,
Header-Zeile, Spaltenreihenfolge exakt nach Spezifikation), index.xml
deklariert nur vorhandene Tabellen und entspricht der DTD-Struktur.
Jede Prüfregel zitiert ihre Fundstelle in der DSFinV-K-2.4 als
Kommentar. Unit-Tests mit guten und absichtlich kaputten Fixtures. Ein
Integrationstest erzeugt aus Seed-Daten (Fake-TSE) einen echten Export
und muss befundfrei sein — damit läuft die Prüfung in jedem
verify-Lauf und im CI-Integrationsjob.

### Akzeptanzkriterien

- [ ] Validator-Paket mit Schnittstelle „ZIP rein, Befundliste raus" und eigenen Unit-Tests (gute + kaputte Fixtures)
- [ ] Strukturregeln decken Dateinamen, CSV-Format, Spaltenreihenfolge und index.xml/DTD ab; jede Regel mit Fundstellen-Kommentar
- [ ] Integrationstest Seed → Export → Validator ist befundfrei und läuft unter dem integration-Tag
- [ ] `make verify` grün, CI-Integrationsjob grün

---

## Phase 6: DSFinV-K-Validator — Inhaltsprüfung

**User Stories**: 10 (Inhaltsregeln)

### Kontext

- `docs/plans/guide-manuelle-qa-v1.0.0.md:48-62` — Block-6-Checkliste als Regelkatalog
- `docs/compliance.md` §6.6 (Storno-Abbildung), §6.4 (Bediener-Felder)
- `backend/seed/szenario.go:296-300` — das Drehbuch muss alle geprüften Fälle enthalten (Storno, Kombi, Tagesabschluss)

### Was zu bauen ist

Die Inhaltsregeln auf demselben Paket: Storno-Referenzen (Negativbetrag,
`REF_BON_ID` in references.csv, `BON_STORNO = 0`), Kombi-Steueraufteilung
in lines_vat.csv und transactions_vat.csv (7-%- und 19-%-Anteil),
Bediener-Felder (`BEDIENER_NAME` eingefroren, `BEDIENER_ID` = user_id),
Tagesabschluss-Zeile (AVSonstige mit BON_NAME „Tagesabschluss"),
tse.csv-Stammdaten vollständig gefüllt, Abrechnungskreis in
allocation_groups.csv. Falls das Seed-Drehbuch einen geprüften Fall noch
nicht erzeugt, wird es ergänzt, damit der Integrationstest jede Regel
wirklich ausübt.

### Akzeptanzkriterien

- [ ] Alle Inhaltsregeln aus User Story 10 implementiert, je mit Fundstellen-Kommentar und kaputtem Fixture
- [ ] Das Seed-Szenario übt jede Inhaltsregel nachweislich aus (Storno, Kombi, Tagesabschluss, TSE-Stammdaten)
- [ ] Integrationstest weiterhin befundfrei; `make verify` grün

---

## Phase 7: Berechtigungs-Matrix

**User Stories**: 16 (Matrix), 17 (Vollständigkeits-Guard), 18 (Header + Rate-Limit)

### Kontext

- `backend/app/app.go:45-86` — Rollen-Middleware je Bereich, heutige Registrierung
- `backend/api/admin.go:33-69` und Geschwister — imperative Route-Listen, die in die Tabelle wandern
- `backend/api/middleware/middleware.go` — JWT-Prüfung, RateLimitMiddleware(5)
- `reverse-proxy/caddyfile.go` + `caddyfile_test.go` — Ort der Header-Assertions

### Was zu bauen ist

Die Routen-Registrierung wird deklarativ: eine Routentabelle (Pfad,
Bereich, erlaubte Rollenmenge) als einzige Quelle, aus der `SetupRoutes`
registriert. Darauf ein table-driven Integrationstest, der jede Route
gegen jede Rolle fährt: erlaubte Rollen antworten 2xx oder mit
fachlichem Fehler, verbotene 403, fehlende und ungültige Tokens 401.
Für Routen mit Objektbezug wird der Zugriff über die fachliche
Abgrenzung hinaus geprüft (Fallkatalog aus der Tabelle abgeleitet).
Sonderfälle health/auth/relay (kein JWT) sind in der Tabelle explizit
deklariert. Dazu: Middleware-Integrationstest für das Login-Rate-Limit
(429 nach Überschreitung) und Header-Assertions auf die generierte
Caddyfile (CSP, HSTS, X-Frame-Options, X-Content-Type-Options).

### Akzeptanzkriterien

- [ ] Routentabelle ist die einzige Registrierungsquelle; eine Route ohne Rollendeklaration ist nicht registrierbar bzw. lässt den Test fehlschlagen
- [ ] Matrix-Test prüft jede Route × jede Rolle × 401/403 sowie die Objektbezug-Fälle
- [ ] Login-Rate-Limit-Test (429) und Caddyfile-Header-Test vorhanden
- [ ] Verhalten aller Routen unverändert (bestehende Tests grün); `make verify` grün

---

## Phase 8: Schwachstellen-Scans in CI

**User Stories**: 19 (Dependency-Scans)

### Kontext

- `.github/workflows/ci.yml:9-46` — paths-filter; `go.work` listet alle Go-Module
- Frontend-Supply-Chain-Policy (minimumReleaseAge) bleibt unberührt

### Was zu bauen ist

Zwei CI-Jobs: govulncheck über alle Go-Module (backend, resolver,
reverse-proxy, windows/relay, windows/starter), Fehlschlag bei
erreichbaren Befunden; pnpm audit für das Frontend, Fehlschlag ab
Schweregrad high. Beide laufen bei jedem PR (paths-gefiltert auf die
jeweiligen Module) und zusätzlich wöchentlich per Schedule, damit neue
Advisories ohne Code-Änderung auffallen.

### Akzeptanzkriterien

- [ ] govulncheck-Job über alle Go-Module, rot bei erreichbaren Befunden
- [ ] pnpm-audit-Job, rot ab high; Ausnahmen nur dokumentiert im Workflow
- [ ] Beide Jobs laufen bei PRs und wöchentlich; aktueller Stand ist grün (oder Befunde sind als Ausnahme dokumentiert)

---

## Phase 9: Fuzz-Targets und Parallelzugriffstest

**User Stories**: 20 (Fuzzing), 21 (Parallelzugriff)

### Kontext

- `backend/domain/event/event.go:104` — Event-Daten-Deserialisierung (Replay-Kante)
- `backend/api/fiskal/dsfinvk/table.go` + `mapper.go` — CSV-Encoder
- `backend/api/druck/bondruck/application/escpos/formatter.go` — ESC/POS-Encoder
- `backend/domain/kasse/event_json_contract_test.go` — Korpus-Quelle: echte Event-JSON-Fixtures
- `ci.yml:336-392` — Integrationstests laufen `-p 1`; Parallelität muss innerhalb eines Tests stattfinden

### Was zu bauen ist

Drei Go-native Fuzz-Targets mit Seed-Korpus aus echten Fixtures:
Event-JSON-Deserialisierung (kein Panic, kein stiller Datenverlust beim
Replay), DSFinV-K-CSV-Encoder und ESC/POS-Encoder (kein Panic, keine
kaputten Steuersequenzen). Der Korpus läuft automatisch als Unit-Test
mit; `make fuzz` startet längere Läufe lokal (definierte fuzztime je
Target), kein Dauerlauf in CI. Dazu der Parallelzugriffstest: ein
Integrationstest, in dem zwei nebenläufige Clients denselben Tisch
bedienen (bestellen, ausgeben, kassieren); Assertion auf Konsistenz von
Journal, Projektion und Salden — keine verlorenen Positionen, keine
Doppelbuchung.

### Akzeptanzkriterien

- [ ] Drei Fuzz-Targets mit Seed-Korpus, die als normale Unit-Tests mitlaufen
- [ ] `make fuzz` führt längere Läufe lokal aus
- [ ] Parallelzugriffstest (zwei Clients, gleicher Tisch) unter dem integration-Tag, Konsistenz-Assertions auf Journal und Projektion
- [ ] `make verify` grün

---

## Phase 10: TSE-Live-Suite — Geschäftsvorfälle und Stammdaten

**User Stories**: 12 (alle Geschäftsvorfälle live), 15 (Setup bleibt Opt-in)

### Kontext

- `backend/repository/tse_repo/fiskaly_client_live_test.go:21-52` — Live-Muster (Skip, TEST-Guard)
- `backend/api/fiskal/signatur/` — Worker, der die Aufträge abarbeitet
- `Makefile:56-62` — bestehende Targets; `scripts/test-integration.sh` — Wegwerf-Postgres-Muster
- `.env.fiskaly-test.example` — Credentials-Vorlage; Memory: TSS ce2d00dc, serial_number-Lektion

### Was zu bauen ist

Die Live-Suite als Integrationstests mit Live-Guard: echte Datenbank
(Wegwerf-Postgres wie im Integrationsskript), echter fiskaly-Client
gegen die TEST-TSS, laufender Signatur-Worker. Jeder Geschäftsvorfall
wird durch die Anwendungsdienste ausgelöst und real signiert:
Bestellung, Teil- und Vollzahlung, Direktverkauf und dessen Storno,
Warenrücknahme, geldneutrale Korrektur, Umbuchung, Geldtransit,
Kassendifferenz, Tagesabschluss. Geprüft wird je Vorfall der
abgeschlossene Signaturauftrag, die Signaturdaten im Journal und das
processType-Mapping. Zusätzlich explizit: die Vollständigkeit der
persistierten TSE-Stammdaten (Signaturalgorithmus, Public Key,
Zertifikat, Seriennummer — die serial_number-Lektion vom 08.07.2026).
`make test-tse-live` wird zum Einstiegspunkt der Suite ausgebaut;
`test-tse-live-setup` (legt eine unlöschbare TSS an) bleibt separat.

### Akzeptanzkriterien

- [ ] `make test-tse-live` fährt DB + Suite; ohne `.env.fiskaly-test` bricht das Target ab, ohne Credentials skippen die Tests
- [ ] Guard verhindert Läufe gegen Nicht-TEST-Umgebungen
- [ ] Jeder Geschäftsvorfall aus User Story 12 real signiert, Signatur im Journal und processType-Mapping geprüft
- [ ] Stammdaten-Vollständigkeit explizit assertet
- [ ] Der Setup-Durchlauf bleibt ausschließlich im separaten Opt-in-Target

---

## Phase 11: TSE-Live-Suite — Ausfall, Nachsignierung, Latenz

**User Stories**: 13 (Ausfallpfad), 14 (p95-Latenzmessung)

### Kontext

- `backend/domain/tse/stoerung.go` — Störungsprotokoll
- `docs/plans/guide-manuelle-qa-v1.0.0.md:31-37` — Block-4-Checkliste als Testfall-Katalog
- `docs/verfahrensdokumentation.md` — Zusage p95 < 5 s, die die Messung belegt oder korrigiert

### Was zu bauen ist

Auf der Suite aus Phase 10: der Ausfallpfad, indem der Client zur
Laufzeit auf eine unerreichbare Adresse bzw. ungültige Credentials
umgeschaltet wird. Geprüft wird: Vorgänge bleiben buchbar und warten
nicht auf die TSE, das Störungsprotokoll erfasst den Zeitraum mit
Grund, nach Wiederherstellung läuft die Nachsignierung, das
Abschluss-Gate antwortet korrekt (409 mit Anzahl/Alter bei frisch
ausstehenden Signaturen, erlaubt bei dokumentiertem Ausfall). Dazu die
Latenzmessung: eine definierte Zahl Signaturaufträge im Burst, Ausgabe
von p50/p95; das Ergebnis wird in der Verfahrensdokumentation
nachgeführt (Zusage bestätigen oder anpassen).

### Akzeptanzkriterien

- [ ] Ausfalltest: buchbar während Störung, Störungsprotokoll mit Zeitraum, Nachsignierung nach Wiederherstellung, Abschluss-Gate-Verhalten in beiden Fällen
- [ ] Latenzmessung gibt p50/p95 reproduzierbar aus
- [ ] Verfahrensdokumentation trägt das Messergebnis (bestätigt oder angepasst)

---

## Phase 12: Ops-Smoke-Skript

**User Stories**: 22 (Roundtrip-Skript), 23 (Release-Smoke)

### Kontext

- `scripts/prod-init.sh`, `prod-update.sh`, `prod-backup.sh`, `prod-backup-verify.sh` — die orchestrierten Pfade
- `docker-compose.prod.yml` — Stack, gegen den geprüft wird
- `.github/workflows/ci.yml:326-334` — shellcheck-Job deckt `scripts/*.sh` bereits ab
- Geklärt: realer Lauf erst in Phase 14 auf dem QA-Host

### Was zu bauen ist

`scripts/ops-smoke.sh` mit drei Modi: Erstinstallation (prod-init bis
zum Login-Roundtrip per API mit dem ausgegebenen Einmalpasswort),
Betrieb (prod-backup, prod-backup-verify, prod-update-Roundtrip) und
Release-Smoke (gepinnte Image-Version als Parameter, dann Installation,
Login, ein Verkauf, ein Beleg, ein Export per API). Am deployten Stack
prüft das Skript zusätzlich Security-Header und Login-Rate-Limit durch
den Reverse-Proxy. Jeder Schritt wird maschinenlesbar protokolliert
(Schrittname, Status, Dauer). Host-Provisionierung, destruktives
prod-restore und TLS-Abnahme bleiben ausdrücklich beim Menschen.

### Akzeptanzkriterien

- [ ] Skript mit den drei Modi und maschinenlesbarem Protokoll
- [ ] Header- und Rate-Limit-Prüfung am deployten Stack enthalten
- [ ] shellcheck-frei (CI-Job grün)
- [ ] Review bestätigt: keine destruktiven Schritte, Abbruch bei jedem Fehlschritt

---

## Phase 13: Rest-Guide-Umbau

**User Stories**: 24 (Guide enthält nur noch Handarbeit)

### Kontext

- `docs/plans/guide-manuelle-qa-v1.0.0.md` — Blöcke 1–8
- `docs/plans/plan-v1.0.0-release.md` — Gates verweisen auf den Guide; Checkboxen bleiben dort führend
- PRD-Abschnitt „Solution", Kategorie C — die Liste dessen, was bleibt

### Was zu bauen ist

Der bestehende Guide wird in place zum Rest-Guide: jeder automatisierte
Punkt wird entfernt und durch einen Verweis auf die zuständige Suite
ersetzt (Block 3/4 → TSE-Live-Suite, Block 6 → Validator, Block 7
teilweise → Ops-Smoke und Parallelzugriffstest). Übrig bleiben:
physische Hardware (Bondrucker-Druckbild, QR-Scan mit dem Handy), der
echte Windows-Rechner, fiskaly-Konto samt TEST→LIVE-Umschaltung und
PUK/PIN-Verwahrung, Zwei-Geräte-Test in echt, destruktives
prod-restore, TLS-Abnahme, Usability mit Vereinshelfern, das
IDEA-/Prüftooling-Gegenlesen und die Abnahme-Entscheidungen. Eine
Checkliste, keine Doppelpflege; Verweise aus dem Release-Guide werden
angepasst, seine Gate-Checkboxen bleiben führend.

### Akzeptanzkriterien

- [ ] Guide enthält keinen Punkt mehr, den eine Suite abdeckt; jeder entfernte Punkt verweist auf seine Suite
- [ ] Alle Kategorie-C-Punkte aus dem PRD sind enthalten
- [ ] Release-Guide-Verweise stimmen weiterhin

---

## Phase 14: QA-Durchführung und Befund-Report

**User Stories**: 25 (einmalige Durchführung), 26 (dauerhafter Umfang bestätigt)

### Kontext

- Alle Artefakte aus Phasen 1–13
- `frontend/src/routes.ts` — Routenkarte als Basis des Screen-Sweeps
- Vorbedingungen: Wegwerf-Host (Ubuntu, von Nico gestellt), `.env.fiskaly-test` lokal vorhanden

### Was zu bauen ist

Die einmalige vollständige Durchführung: alle Suiten laufen (verify
inkl. Validator, Matrix, Fuzz-Korpus und Parallelzugriff; E2E gegen den
CI-Stack; Scans), die TSE-Live-Suite inklusive Ausfall- und
Latenz-Block läuft gegen die TEST-TSS, das Ops-Smoke-Skript fährt auf
dem Wegwerf-Host Erstinstallation, Betrieb und Release-Smoke. Zusätzlich
agentengetrieben: ein Screen-Sweep über sämtliche Routen und Zustände
aus `routes.ts` in beiden Viewports (auch Seiten ohne eigene Spec) und
eine heuristische UX-Review. Alle Befunde landen nach Schwere geordnet
im Befund-Report (`docs/plans/befund-report-qa-v1.0.0.md`); die Fixes
selbst sind Folgearbeit außerhalb dieses Plans.

### Akzeptanzkriterien

- [ ] Alle Suiten-Läufe protokolliert (grün oder Befund im Report)
- [ ] TSE-Live-Blöcke vollständig durchgespielt, Latenzergebnis dokumentiert
- [ ] Ops-Smoke in allen drei Modi auf dem Wegwerf-Host gelaufen
- [ ] Screen-Sweep über alle Routen/Zustände und heuristische UX-Review durchgeführt
- [ ] Befund-Report existiert, nach Schwere geordnet, mit Verweis auf Reproduktionsweg je Befund
