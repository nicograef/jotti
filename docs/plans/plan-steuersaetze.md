# Plan: Steuersätze (F-07)

> Source PRD: `docs/prds/prd-steuersaetze.md`

## Goal

Jedes Produkt erhält einen Pflichtfeld-Steuersatz (`regel`/`ermaessigt`/`befreit`/`kombi`). Der Satz wird beim Bestellen in Positionen eingefroren (Fat Event) und im Reporting pro Steuersatz aufgeschlüsselt (Brutto/Netto/Steuer). Ein zentrales `steuer`-Rechenmodul liefert die Brutto→Netto/Steuer-Aufteilung.

## Architectural decisions

- **Enum-Typ (DB):** `CREATE TYPE Steuersatz AS ENUM ('regel', 'ermaessigt', 'befreit', 'kombi')` in `01_initial.up.sql`.
- **DB-Spalte:** `produkte.steuersatz Steuersatz NOT NULL`.
- **Domain-Paket:** `backend/domain/steuer/` (Support-Paket, analog zu `event`/`jwt`). Enthält Enum, Rechnung, zog-Schema.
- **Position-Feld:** `Steuersatz string` im Domain-Struct, `"steuersatz"` JSON-Key in Event-Data.
- **Enrichment:** Serverseitig in `table/application/command.go` und `direktverkauf/application/command.go`, dort wo bereits Produktname/Preis/Kategorie eingefroren werden.
- **Reporting-Query:** Neue SQL-Funktion `kj_extract_umsatz_pro_steuersatz` (JSONB-Unnest der Positions-Arrays).
- **Frontend-Feld:** `steuersatz` im `ProduktSchema`, kategoriebasierter Default im Formular.
- **Preis-Semantik:** Bestehender `preis_cents`/`einzelpreis` = Bruttopreis. Keine Umbenennung.

## Inventory

- `backend/domain/product/product.go:1-45` — Produkt/Kategorie/Status domain model
- `backend/domain/kasse/bestellung.go:10-60` — Position struct + positionEventData + zog-Schema
- `backend/domain/kasse/tisch_session_events.go` — Event-Typen und Data-Structs
- `backend/domain/kasse/direktverkauf_events.go` — Direktverkauf-Event-Structs (shared positionEventData)
- `backend/api/table/application/command.go:349-430` — Bestellung-Enrichment (batch-lookup + fat Position)
- `backend/api/direktverkauf/application/command.go:142-181` — Direktverkauf-Enrichment (gleiche Logik)
- `backend/api/product/http/command_handler.go` — Produkt-HTTP-Handler + Request/Response-DTOs
- `backend/api/product/application/command.go` — Produkt-Application-Service
- `backend/repository/product_repo/repo.go` — Produkt-Repository (CRUD)
- `backend/sqlc/queries/produkte.sql` — sqlc-Queries für Produkte
- `backend/sqlc/queries/reporting.sql` — Reporting-Aggregationsqueries (kj_extract-Funktionen)
- `database/migrations/01_initial.up.sql:56-100` — produkte/produkt_varianten-Tabellen + ProduktKategorie-Enum
- `database/migrations/01_initial.up.sql:263-310` — kj_extract-Funktionen (JSONB-Extraktion)
- `database/seed.sql` — Seed-Produkte und -Events
- `backend/domain/reporting/reporting.go` — Reporting-Domain-Modell (Summary, Breakdowns, ReportingData)
- `backend/api/reporting/application/query.go` — Reporting-Application-Service
- `backend/repository/reporting_repo/repo.go` — Reporting-Repository (GetReporting mit errgroup)
- `frontend/src/admin/products/Produkt.ts` — TS-Typen/Schemas (ProduktSchema, KategorieSchema)
- `frontend/src/admin/products/NewProductDialog.tsx` — Produktanlage-Dialog
- `frontend/src/admin/products/EditProductDialog.tsx` — Produktbearbeitung-Dialog
- `frontend/src/service/table/Bestellung.ts` — TS PositionSchema
- `frontend/src/admin/reporting/ReportingResults.tsx` — Tagesabrechnungs-UI
- `frontend/src/admin/reporting/ReportingBackend.ts` — Reporting-API-Client
- `backend/domain/event/event.go` — Support-Paket-Muster (lean, interface-frei)
- `backend/domain/jwt/jwt.go` — Support-Paket-Muster (lean, interface-frei)

## Resolved decisions

- 4 Phasen (steuer-Paket → Produkt → Position/Events → Reporting), jeweils als eigener vertikaler Schnitt.
- Kein neues Kombi-Menü-Produkt in den Seed-Daten; Kombi wird ausschließlich über Tests abgedeckt.
- Bestehende Seed-Produkte erhalten plausible Steuersätze: Getränke → `regel`, Speisen → `ermaessigt`, Festbändchen → `befreit`.
- Position-Event-Daten werden um `"steuersatz"` erweitert (neuer JSON-Key in positionEventData).
- Rundung: kaufmännisch (half up), Netto = Brutto − Steuer, Invariante Netto + Steuer = Brutto.
- Reporting aggregiert Brutto pro Steuersatz in SQL; Netto/Steuer wird in Go aus dem aggregierten Brutto berechnet.
- Kombi-Auflösung im Reporting: `kombi`-Brutto wird in Go via `steuer.Aufteilen()` auf 7 %- und 19 %-Zeilen verteilt; keine eigene „Kombi"-Zeile im UI.

## Open questions / Risks

- Die Reporting-Query per JSONB-Unnest kann bei sehr hoher Event-Anzahl langsam werden. Für den Ziel-Use-Case (2–3 Feste/Jahr, < 5.000 Events/Kassensitzung) ist das unkritisch.
- Seed-Events müssen manuell um `steuersatz`-Felder in den JSONB-Payloads erweitert werden (kein automatisches Schema-Enforcement auf JSONB).

---

## Phase 1: `steuer`-Paket (Domain Support)

**User stories**: 17, 18, 20

### Context

- `backend/domain/event/event.go` — Support-Paket-Muster: lean Paket mit Typen, Konstruktoren, Validierung, Unit-Tests.
- `backend/domain/jwt/jwt.go` — Weiteres Support-Paket-Beispiel: keine Interfaces, keine `json`-Tags auf Structs.
- `backend/domain/kasse/bestellung.go:60-65` — Bestehendes zog-Schema-Muster für Position (Vorlage für Steuersatz-Validierung).

### What to build

Ein neues Paket `backend/domain/steuer/` mit:

1. **Typ `Steuersatz`** (string-Enum) mit Konstanten `RegelSteuersatz`, `ErmaessigtSteuersatz`, `BefreitSteuersatz`, `KombiSteuersatz`. Methode `Prozent() int` (19/7/0/— für kombi).
2. **Typ `Aufteilung`** mit Feldern `Satz Steuersatz`, `Brutto int`, `Netto int`, `Steuer int`.
3. **Funktion `Aufteilen(brutto int, satz Steuersatz) []Aufteilung`** — die zentrale Rechenlogik. Für regel/ermaessigt/befreit ein Element, für kombi zwei (70 % → 7 %, 30 % → 19 %). Rundung kaufmännisch (round half up), Invariante: Netto + Steuer = Brutto pro Aufteilung, Summe Brutto-Anteile = Gesamt-Brutto bei kombi.
4. **zog-Schema** `SteuersatzSchema` zur Validierung erlaubter Werte (wiederverwendbar in Produkt- und Positions-Validierung).
5. **Unit-Tests** (tabellengetrieben): alle Sätze, Rundungs-Edgecases (500ct/19 %, 500ct/7 %, 0ct, 1ct, ungerade Beträge bei kombi, halbe-Cent-Grenze).

### Acceptance criteria

- [ ] `backend/domain/steuer/steuer.go` existiert mit Typ, Konstanten, `Prozent()`-Methode, `Aufteilen()`-Funktion
- [ ] `backend/domain/steuer/steuer_test.go` existiert mit tabellengetriebenen Tests
- [ ] Invariante Netto + Steuer = Brutto gilt für alle Testfälle (inklusive kombi-Anteile)
- [ ] Kombi-Aufteilung: Summe der Brutto-Anteile = Gesamt-Brutto (kein Rundungsverlust)
- [ ] zog-Schema validiert nur die 4 erlaubten Werte und gibt deutsche Fehlermeldung bei Verstoß
- [ ] `make test` und `make lint` bestehen

---

## Phase 2: Produkt + Steuersatz (Schema → API → UI)

**User stories**: 1, 2, 3, 4, 5, 6, 7, 8, 21, 22

### Context

- `database/migrations/01_initial.up.sql:56-70` — produkte-Tabelle, ProduktKategorie-Enum (Vorlage für neuen Enum-Typ)
- `backend/sqlc/queries/produkte.sql` — bestehende Insert/Update/Select-Queries
- `backend/domain/product/product.go` — Produkt-Domain-Modell (kein `json`-Tag)
- `backend/api/product/http/command_handler.go` — createProductRequest/Response (Vorlage für neues Feld)
- `backend/api/product/application/command.go` — CreateProduct/UpdateProduct (Steuersatz durchreichen)
- `backend/repository/product_repo/repo.go` — CreateProduct/UpdateProduct (SQL-Parameter erweitern)
- `frontend/src/admin/products/Produkt.ts` — ProduktSchema, CreateProduktSchema (Zod)
- `frontend/src/admin/products/NewProductDialog.tsx` — Produktformular mit Kategorie-Auswahl

### What to build

End-to-end: DB-Enum + Spalte → sqlc-Queries → Repository → Domain-Modell → Application-Service → HTTP-Handler/DTO → Frontend-Schema + Formular.

1. **DB:** Neuer Enum-Typ `Steuersatz` + neue NOT-NULL-Spalte `steuersatz` auf `produkte`.
2. **sqlc:** Queries für CreateProdukt/UpdateProdukt/Get\* um `steuersatz` erweitern. `make sqlc`.
3. **Repository:** CreateProduct/UpdateProduct nehmen Steuersatz entgegen.
4. **Domain:** `Produkt` erhält Feld `Steuersatz steuer.Steuersatz`. Erzeugung/Aktualisierung validieren.
5. **Application:** CreateProduct/UpdateProduct nehmen Steuersatz-Parameter entgegen.
6. **HTTP:** Request-DTO erhält `steuersatz`-Feld, Response-DTO (produktDTO) weist ihn aus. Validierung via zog.
7. **Frontend:** `ProduktSchema` + `CreateProduktSchema` um `steuersatz` erweitern. Formular zeigt Select mit deutschen Labels. Kategoriebasierter Default: essen → ermaessigt, getraenk/sonstiges → regel. Beim Bearbeiten: gespeicherter Wert vorausgewählt.
8. **Seed:** `database/seed.sql` — bestehende Produkte erhalten passende Steuersätze.

### Acceptance criteria

- [ ] DB-Schema enthält `Steuersatz`-Enum und `produkte.steuersatz NOT NULL`
- [ ] Produkt anlegen ohne Steuersatz → Validierungsfehler (Backend + Frontend)
- [ ] Produkt anlegen mit `steuersatz: "ermaessigt"` → erfolgreich, Wert in DB und Response
- [ ] Produkt bearbeiten → Steuersatz änderbar
- [ ] Frontend: Kategorie-Wechsel setzt Default (essen → ermaessigt, getraenk → regel)
- [ ] Frontend: beim Bearbeiten ist der gespeicherte Satz vorausgewählt
- [ ] Seed-Daten laden fehlerfrei (`make dev`)
- [ ] `make check` besteht (lint + unit tests + build)

---

## Phase 3: Position & Events (Enrichment + Seed)

**User stories**: 9, 10, 11, 12, 19

### Context

- `backend/domain/kasse/bestellung.go:10-38` — Position + positionEventData (shared fat structure)
- `backend/domain/kasse/bestellung.go:60-65` — positionSchema (zog-Validierung)
- `backend/api/table/application/command.go:360-395` — Bestellung-Enrichment (Produkt-Lookup → fat Position)
- `backend/api/direktverkauf/application/command.go:142-181` — Direktverkauf-Enrichment (gleiche Logik)
- `frontend/src/service/table/Bestellung.ts` — TS PositionSchema (Frontend-Validierung)
- `database/seed.sql` — Seed-Events mit JSONB-Positionen

### What to build

1. **Domain:** `Position` struct erhält `Steuersatz string`. `positionEventData` erhält `Steuersatz string \`json:"steuersatz"\``. Konvertierungsfunktionen (`toPositionEventData`/`fromPositionEventData`) bleiben strukturelle Typkonvertierung. `positionSchema`(zog) erhält Pflichtfeld`Steuersatz`mit Validierung gegen die 4 erlaubten Werte (nutzt`steuer.SteuersatzSchema` oder equivalent OneOf).
2. **Enrichment:** In `table/application/command.go` und `direktverkauf/application/command.go` wird beim Aufbau der fat Position zusätzlich `Steuersatz: string(produkt.Steuersatz)` aus dem geladenen Produkt übernommen.
3. **Frontend:** `PositionSchema` in `Bestellung.ts` erhält `steuersatz`-Feld (Zod enum).
4. **Seed-Events:** JSONB-Payloads in `seed.sql` um `"steuersatz"` in jeder Position ergänzen (konsistent mit den Seed-Produkten aus Phase 2).

### Acceptance criteria

- [ ] Neue Bestellung enthält `steuersatz` in jeder Position des Events (verifizierbar in DB)
- [ ] Der Steuersatz wird serverseitig aus dem Produkt ermittelt (nicht vom Client gesendet)
- [ ] Alle 6 positionsführenden Event-Typen enthalten `steuersatz` in ihren Positionen
- [ ] Seed-Events enthalten konsistente `steuersatz`-Werte
- [ ] `make check` besteht
- [ ] Position-Validierung lehnt fehlenden/ungültigen Steuersatz ab

---

## Phase 4: Reporting (SQL → Backend → Frontend)

**User stories**: 13, 14, 15, 16

### Context

- `backend/sqlc/queries/reporting.sql:16-39` — GetReportingStats (aggregiert kassenjournal, Vorbild für neue Query)
- `database/migrations/01_initial.up.sql:263-310` — bestehende kj_extract-Funktionen (JSONB-Extraktion)
- `backend/domain/reporting/reporting.go` — Reporting-Domain-Modell (Summary, Breakdowns, ReportingData)
- `backend/repository/reporting_repo/repo.go:37-125` — GetReporting (errgroup-Pattern, parallele Queries)
- `backend/api/reporting/application/query.go` — Reporting-Application-Service
- `frontend/src/admin/reporting/ReportingResults.tsx` — Tagesabrechnungs-UI (Tabs, Cards)
- `frontend/src/admin/reporting/ReportingBackend.ts` — API-Endpunkt `admin/get-abrechnung`

### What to build

1. **SQL:** Neue Query `GetUmsatzProSteuersatz` — entfaltet Positions-Arrays aus kassierten Events (`zahlung-kassiert:v1`, `direktverkauf-getaetigt:v1`, minus `direktverkauf-storniert:v1`), gruppiert nach `steuersatz`, summiert `einzelpreis × menge` als Brutto pro Satz. Gefiltert auf Kassensitzung.
2. **Domain:** Neuer Typ `UmsatzSteuersatz` mit `Satz steuer.Steuersatz`, `BruttoCents int`, `NettoCents int`, `SteuerCents int`. `ReportingData` erhält Feld `UmsatzProSteuersatz []UmsatzSteuersatz`.
3. **Application/Repository:** Repository ruft die neue Query auf (parallel in errgroup). Application-Layer empfängt die Brutto-pro-Satz-Werte, ruft `steuer.Aufteilen()` auf jedem auf, um Netto/Steuer abzuleiten. Kombi-Brutto wird via `Aufteilen()` auf die 7 %- und 19 %-Zeilen verteilt (keine eigene Kombi-Zeile).
4. **HTTP:** Response-DTO für Reporting um `umsatzProSteuersatz`-Array erweitern (Satz-Label, Brutto, Netto, Steuer).
5. **Frontend:** Neuer Abschnitt „Umsatz nach Steuersatz" in der Tagesabrechnungs-Seite. Tabelle mit Zeilen pro vorkommendem Satz (Label, Brutto, Netto, Steuer). Steuerbefreit (0 %) erscheint mit Steuerbetrag 0. Typen/Schema aktualisieren.

### Acceptance criteria

- [ ] Tagesabrechnung zeigt Aufschlüsselung pro Steuersatz (19 %, 7 %, 0 %)
- [ ] Pro Steuersatz: Brutto, Netto, Steuerbetrag korrekt (Netto + Steuer = Brutto)
- [ ] Basis = Zahlungen + Direktverkäufe − Direktverkauf-Stornos (konsistent mit PRD)
- [ ] Steuerbefreiter Umsatz erscheint als eigene Zeile mit Steuerbetrag 0
- [ ] Kombi-Umsatz fließt anteilig in 7 %- und 19 %-Zeilen ein (keine Kombi-Zeile)
- [ ] `make check` besteht (inkl. Reporting-Integrationstests falls vorhanden)
- [ ] Frontend zeigt nur die Daten an, die das Backend liefert (keine clientseitige Berechnung)
