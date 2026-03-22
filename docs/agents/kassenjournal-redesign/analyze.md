# Analyse: Kassenjournal-Redesign — Vollständige Umsetzung

## Ziel

Das [Redesign-Dokument](../../redesign.md) definiert eine neue Core Domain „Kasse" mit dem Kassenjournal als einziger Tabelle für alle finanziellen Geschäftsvorfälle. Ziel dieser Analyse ist es, **alle** notwendigen Änderungen zu identifizieren — sowohl in der Dokumentation als auch im Code —, um das Redesign vollumfänglich umzusetzen. Das Redesign überstimmt alle bestehenden Dokumente, die aktuelle Implementierung und alle bisherigen Konventionen.

Die Analyse gliedert sich in zwei Hauptbereiche:

1. **Dokumentation:** Alle Markdown-Dateien im Repo, die aktualisiert werden müssen
2. **Implementierung:** Alle Code-Dateien (Backend, Frontend, DB, Tests), die geändert werden müssen

---

## Bestandsaufnahme

### 1. Dokumentation — Markdown-Dateien

Das Redesign verändert fundamentale Konzepte: Bounded Contexts werden zusammengelegt, Begriffe bekommen neue Bedeutungen, die Persistenzstrategie ändert sich. Jedes Dokument, das Kassenbetrieb, Kassenführung, Abrechnungskreis, Events, Tisch-Aggregat oder die DB-Struktur referenziert, muss aktualisiert werden.

#### 1.1 `docs/handbuch.md` — Entwickler-Handbuch (UMFANGREICHSTE ÄNDERUNG)

Das Handbuch ist das am stärksten betroffene Dokument. Es definiert das bisherige Domänenmodell, das durch das Redesign grundlegend ersetzt wird.

**§2 Bounded Contexts (Zeilen ~84–113):**

- **Ist:** Vier Bounded Contexts: Kassenbetrieb (Core Domain, ES), Kassenführung (Supporting, Immutable Records), Stammdaten (Supporting, CRUD), Auth (Generic)
- **Soll:** Drei Bounded Contexts: Kasse (Core Domain, ES/Kassenjournal), Stammdaten (Supporting, CRUD), Auth (Generic)
- Die Kontextübersicht (Tabelle), Context Map und Beziehungen zwischen Kontexten müssen komplett neu geschrieben werden
- Die bidirektionale Abhängigkeit Kassenbetrieb↔Kassenführung entfällt (redesign.md §2.1)

**§3 Kassenbetrieb (Core Domain) (Zeilen ~116–310):**

- **Ist:** Tisch als Event-Sourced Aggregat mit Subject `tisch:{id}`, Event-Typen `tisch.{event}:v1`
- **Soll:** Tisch-Session (Abrechnungskreis) als Aggregat im Kasse-Kontext mit Subject `kassensitzung-{YYYYMMDD}-tisch-{id}`, Event-Typen `{event}:v1` ohne `tisch.`-Präfix
- §3.1 (Tisch-Aggregat): Muss zu „Tisch-Abrechnungskreis-Aggregat" werden, Subject-Format ändern, Lebenszyklus erklären (session-scoped statt unbegrenzt)
- §3.2 (Invarianten): Neue Invariante „Kassensitzung-Invariante" (offene KS erforderlich), plus alle Kassensitzung-Invarianten aus redesign.md §3.7
- §3.3 (Domain Events): Event-Typen umbenennen (`tisch.bestellung-aufgenommen:v1` → `bestellung-aufgenommen:v1`), neue Kassensitzung-Events hinzufügen (6 neue Event-Typen aus redesign.md §3.6)
- §3.4 (Event Replay + Snapshots): Anpassen an session-scoped Replay, Kassensitzung-Replay beschreiben

**§5 Kassenführung (Supporting Sub-Domain) (Zeilen ~477–605):**

- **Ist:** Eigenständiger Bounded Context mit Immutable Records
- **Soll:** In die Kasse (Core Domain) integriert, Event-Sourced statt Immutable Records
- §5.1–§5.7 müssen komplett in den neuen Kasse-Kontext integriert oder als Abschnitte unter Kasse neu geschrieben werden
- § 5.2 (Abrechnungskreis) → wird zur Tisch-Session (Abrechnungskreis = pro Tisch pro Kassensitzung)
- § 5.3 (Anfangsbestand) → wird zum Kassensitzung-Event `anfangsbestand-gesetzt:v1`
- § 5.4 (Kassenbestand) → SQL-Aggregation über Kassenjournal (redesign.md §3.9)
- § 5.5 (Kassenbewegungen) → Kassensitzung-Events `kassenbewegung-gebucht:v1`
- § 5.6 (Kassensturz) → Kassensitzung-Events `kassensturz-durchgefuehrt:v1` + `differenz-soll-ist-gebucht:v1`
- § 5.7 (Tagesabschluss) → Kassensitzung-Event `tagesabschluss-erstellt:v1`

**§8 Read Models (Zeilen ~808–843):**

- **Ist:** `table_state`-Projektion (PK: `tisch_id`, unbegrenzt)
- **Soll:** `tisch_session_state` (PK: subject, session-scoped) + `kassensitzung_state`
- Tischübersicht: `SELECT ... FROM tische LEFT JOIN tisch_session_state` statt `SELECT * FROM table_state`

#### 1.2 `docs/anforderungen.md` — Anforderungen

**§1 Kassenbetrieb (Zeilen ~68–526):**

- K-07 (Kassenjournal/Historie): Terminologie aktualisieren — „Kassenjournal" statt lose „Historie", Bezug auf Kassensitzung
- Alle K-Anforderungen: „Kassenbetrieb" → „Kasse", Subject-Format-Referenzen aktualisieren

**§5 Reporting (Zeilen ~527–796):**

- R-01 (Tagesabrechnung): Filterung nach `kassensitzung_nr` statt Zeitraum
- R-02 (Datenexport): DSFinV-K-Export-Verweis auf Kassenjournal
- Alle R-Anforderungen: Zeitraum-basierte Filterung → kassensitzungsbezogene Filterung

**§8 Kassenführung (Zeilen ~797–943):**

- **Ist:** Eigenständiger Abschnitt „Kassenführung (Supporting Sub-Domain)"
- **Soll:** Integration in Kasse (Core Domain). Die KF-Anforderungen (KF-01 bis KF-09) bleiben inhaltlich erhalten, gehören aber organisatorisch in den Kasse-Kontext:
  - KF-01 (Abrechnungskreis) → „Kassensitzung eröffnen" (Admin-Aktion, generiert `kassensitzung-eroeffnet:v1`)
  - KF-02 (Anfangsbestand) → Kassensitzung-Event
  - KF-03 (Kassenbestand) → SQL-Aggregation über Kassenjournal
  - KF-04/05/06 (Kassenbewegungen) → Kassensitzung-Events
  - KF-07 (Tagesabschluss) → `tagesabschluss-erstellt:v1`
  - KF-08 (Kassensturz) → `kassensturz-durchgefuehrt:v1` + `differenz-soll-ist-gebucht:v1`

#### 1.3 `docs/language.md` — Ubiquitous Language

**Kassenbetrieb-Begriffe (Zeilen ~111–206):**

- „Tisch" als Domain-Begriff: Muss klargestellt werden, dass der Tisch im Kasse-Kontext als „Tisch-Session" / „Abrechnungskreis" agiert, nicht als Stammdaten-Entität
- Subject-Format: Alle Referenzen von `tisch:{id}` auf `kassensitzung-{YYYYMMDD}-tisch-{id}` aktualisieren
- Event-Typen: `tisch.bestellung-aufgenommen:v1` → `bestellung-aufgenommen:v1` (Präfix entfällt)

**Kassenführung-Begriffe (Zeilen ~209–286):**

- „Abrechnungskreis": Definition aktualisieren → pro Tisch pro Kassensitzung, nicht global
- „Kassensitzung": Neues Konzept, muss als eigener Begriff definiert werden (globaler Betriebstag)
- Neue Begriffe hinzufügen: Kassenjournal, Kassensitzung, Tisch-Session, Kassensitzung-Sperre

**Abweichungen Ist/Soll (Zeilen ~35–52):**

- Aktualisieren: Die bisherigen Renames (Backend ✅, Frontend ⏳) ggf. erweitern um die neuen Umbenennungen

**Namenskonventionen (Zeilen ~17–34):**

- Go-Structs: `domain/kasse/` statt `domain/table/` für Event-Sourcing-Logik
- DB-Tabellen: `kassenjournal` statt `events`, `tisch_session_state` statt `table_state`, `kassensitzung_state` neu

#### 1.4 `docs/diagrams.md` — Diagramme

Nahezu jedes Diagramm ist betroffen:

- **Diagramm 2 (Bounded Context Map, Zeile ~63):** Vier → drei Kontexte, Kasse statt Kassenbetrieb+Kassenführung
- **Diagramm 4 (Tisch-Aggregat Zustandsdiagramm, Zeile ~139):** Neues session-scoped Zustandsdiagramm mit impliziter Erstellung bei erster Bestellung
- **Diagramm 5 (Domain Events + Saldo-Fluss, Zeile ~173):** Event-Typen ohne `tisch.`-Präfix, neue Kassensitzung-Events
- **Diagramm 6 (Bestellvorgang Sequenz, Zeile ~206):** Kassensitzung-Sperre als neuer Schritt im Sequenzdiagramm
- **Diagramm 8 (Kassenführung Lifecycle, Zeile ~344):** Komplett neuschreiben als Kassensitzung-Lifecycle innerhalb der Kasse
- **Diagramm 9 (Kassenbestand-Berechnung, Zeile ~392):** SQL-Aggregation über Kassenjournal statt Cross-Context-Berechnung
- **Diagramm 11 (Schichtenarchitektur, Zeile ~486):** Neue Paketstruktur (`domain/kasse/`, `kassenjournal_repo/`)
- **Diagramm 12 (Event Sourcing + Synchrone Projektion, Zeile ~532):** Zwei Projektionen statt einer, StreamType-Routing
- **Diagramm 15 (API-Bereichsgliederung, Zeile ~678):** Neue Kasse-Endpunkte (KS eröffnen, Kassenbewegung, Kassensturz, Z-Bon)
- **Diagramm 17 (DB-Schema ER-Diagramm, Zeile ~794):** `kassenjournal`, `kassensitzung_state`, `tisch_session_state` statt `events` + `table_state`

#### 1.5 `docs/roadmap.md` — Roadmap

- **Phase 1 — Kassenführung (Zeilen ~52–127):** Aufgaben § 6–12 müssen umgeschrieben werden:
  - § 6 „Kassenführung — DB-Schema" → „Kasse — DB-Schema (Kassenjournal + Projektionen)"
  - § 7 „Abrechnungskreis eröffnen" → „Kassensitzung eröffnen"
  - Alle Aufgaben: Referenzen auf Immutable Records → Event-Sourcing im Kassenjournal
- Der Umsetzungsplan aus redesign.md §6 (Phase A1, A1.5, A2, B, C) sollte in die Roadmap integriert werden

#### 1.6 `docs/compliance.md` — Compliance

- **DSFinV-K-Export (Zeilen ~490–623):** Referenzen auf `events`-Tabelle → `kassenjournal`, `ABRECHNUNGSKREIS` pro Tisch pro Kassensitzung
- **TSE-Integration (Zeilen ~158–342):** Subject-Format-Referenzen aktualisieren
- **GoBD-Konformität (Zeilen ~343–378):** Kassenjournal ist das Aufzeichnungssystem, nicht die Events-Tabelle

#### 1.7 `docs/tagesabschluss.md` — Tagesabschluss

- Referenzen auf den alten Abrechnungskreis aktualisieren
- Kassensturz-Logik: Zwei Events in einer TX (`kassensturz-durchgefuehrt:v1` + `differenz-soll-ist-gebucht:v1` bei Differenz ≠ 0)
- Z-Bon-Aggregation über `kassenjournal` statt Cross-Context-Berechnung

#### 1.8 `docs/bondruck.md` — Bondruck

- Relay-Poll: `events`-Tabelle → `kassenjournal`
- Event-Typ: `tisch.bestellung-aufgenommen:v1` → `bestellung-aufgenommen:v1`
- Subject-Parsing: `tisch:{id}` → `kassensitzung-{YYYYMMDD}-tisch-{id}`

#### 1.9 ADRs (`docs/adr/`)

- **`docs/adr/event-sourcing.md`:** Event-Typen aktualisieren (kein `tisch.`-Präfix), `events` → `kassenjournal`, `table_state` → `tisch_session_state`, Subject-Format, neue Kassensitzung-Events
- **`docs/adr/cqrs.md`:** Zwei Projektionen statt einer, StreamType-Routing, session-scoped Projektion
- **`docs/adr/bondruck.md`:** Relay-Anpassungen referenzieren

#### 1.10 `AGENTS.md` — Agent-Instruktionen

- Zeile ~34–37: Bounded Contexts aktualisieren (3 statt 4)
- Zeile ~130+: Event-Sourcing für „Tisch-Operationen" → „Kasse-Operationen"
- Zeile ~154: „Events are immutable" → „Kassenjournal is immutable"
- Bereiche-Abschnitt: Neuer Bereich „Kasse" mit Kassensitzung-Endpunkten

#### 1.11 `README.md`

- Zeile ~12: „Kassenbetrieb" wird zu „Kasse" oder bleibt als Feature-Gruppierung
- Zeile ~29: „Kassenführung" beschreiben als integraler Teil der Kasse
- Zeile ~75: „Tisch-Operationen werden via Event Sourcing persistiert. Stammdaten und Kassenführung nutzen immutable Records bzw. klassisches CRUD" → „Alle finanziellen Geschäftsvorfälle (Tisch-Operationen und Kassenführung) werden via Event Sourcing im Kassenjournal persistiert. Stammdaten nutzen klassisches CRUD."

#### 1.12 `docs/produktbeschreibung.md`

- Feature-Beschreibungen aktualisieren, soweit sie Kassenbetrieb/Kassenführung als getrennte Bereiche darstellen

#### 1.13 `.github/instructions/`

- **`backend.instructions.md`:** Paketstruktur aktualisieren (`domain/kasse/`, `kassenjournal_repo/`)
- **`database.instructions.md`:** Tabellenname `kassenjournal` statt `events`
- **`event-sourcing.instructions.md`:** Event-Typen aktualisieren (kein Präfix), Subject-Format, neue Kassensitzung-Events, `kassenjournal` statt `events`, zwei Projektionen
- **`frontend.instructions.md`:** Ggf. neue Kassensitzung-UI-Konzepte

---

### 2. Implementierung — Code-Änderungen

#### 2.1 Datenbank-Schema (`database/migrations/01_initial.up.sql`)

**Zeilen ~102–139 (`events`-Tabelle):**

- Tabelle `events` → `kassenjournal` umbenennen
- Neue Spalte `kassensitzung_nr INT NOT NULL` (redesign.md §3.3)
- Index `idx_kassenjournal_ks_nr` hinzufügen
- Alle Indizes umbenennen (`idx_events_*` → `idx_kassenjournal_*`)
- Immutabilitäts-Trigger anpassen (`events_no_update` → `kassenjournal_no_update` etc.)
- REVOKE/GRANT anpassen
- Comments aktualisieren (Subject-Format, Event-Typ-Beispiele)

**Zeilen ~142–159 (`table_state`-Tabelle):**

- Tabelle `table_state` → `tisch_session_state` umbenennen
- PK: `tisch_id INTEGER` → `subject TEXT`
- Neue Spalten: `tisch_id` (FK, nicht mehr PK), `kassensitzung_nr`
- FK anpassen: `REFERENCES kassenjournal(id)` statt `REFERENCES events(id)`
- Index `idx_tisch_session_state_ks_nr` hinzufügen

**Neue Tabelle `kassensitzung_state`:**

- Schema aus redesign.md §3.8: `subject TEXT PK`, `z_nr INT UNIQUE NOT NULL`, `datum DATE`, `status TEXT CHECK(...)`, `last_event_id`, `last_event_version`

**Zeilen ~163–170 (`tisch_favoriten`):**

- FK-Referenz bleibt auf `tische(id)` — keine Änderung

#### 2.2 Seed-Daten (`database/seed.sql`)

- **Zeilen ~160–473 (Events):** Komplett neu schreiben:
  - Tabelle `events` → `kassenjournal`
  - Subject-Format: `'tisch:1'` → `'kassensitzung-20260320-tisch-1'` (etc.)
  - Event-Typen: `'tisch.bestellung-aufgenommen:v1'` → `'bestellung-aufgenommen:v1'` (etc.)
  - Neue Spalte `kassensitzung_nr` befüllen
  - Kassensitzung-Events hinzufügen (Eröffnung, Anfangsbestand)
  - `table_state` → `tisch_session_state` mit angepasstem Schema
  - Neue `kassensitzung_state`-Einträge

#### 2.3 SQL-Queries (`backend/sqlc/queries/`)

**`events.sql` → `kassenjournal.sql`:**

- Datei umbenennen
- `WriteEvent`: Tabelle `events` → `kassenjournal`, neue Spalte `kassensitzung_nr`
- `ReadEvent`, `ReadEventsBySubject`, `GetMaxVersion`, `GetDistinctSubjects`: Tabellenname anpassen

**`table_state.sql` → `tisch_session_state.sql`:**

- Datei umbenennen
- `UpsertTableState` → `UpsertTischSessionState`: PK `subject` statt `tisch_id`, neue Spalten
- `GetTableState` → `GetTischSessionState`: WHERE auf `subject`
- `GetTableStatesByTischIDs` → ggf. `GetTischSessionStatesByKassensitzungNr`
- Neue Query: `GetTischSessionStatesBySitzungNr` für Tischübersicht

**Neue Datei `kassensitzung_state.sql`:**

- `UpsertKassensitzungState`: INSERT/UPDATE Kassensitzung-Projektion
- `GetOffeneKassensitzung`: WHERE status = 'offen'
- `GetKassensitzungBySubject`: Exakter Lookup

**`tables.sql` (Zeilen 12–19):**

- `GetAktiveTische`: JOIN auf `tisch_session_state` statt `table_state`, mit `kassensitzung_nr`-Filter
- `GetAktiveTischeMitFavoriten`: Gleiches JOIN-Pattern

**`reporting.sql` (Zeilen 1–101):**

- Alle Queries: `FROM events` → `FROM kassenjournal`
- Alle Event-Typen: `'tisch.bestellung-aufgenommen:v1'` → `'bestellung-aufgenommen:v1'` (etc.)
- `GetReportingStats`: `WHERE timestamp >= @von AND timestamp < @bis` → `WHERE kassensitzung_nr = @kassensitzungNr`
- `GetUmsatzProServicekraft`: Gleiches Filterkriterium
- `GetUmsatzProTisch`: Subject-Parsing `SPLIT_PART(e.subject, ':', 2)` → Neues Parsing oder JOIN auf `tisch_session_state.tisch_id`
- `GetStornierungen`: Gleiches Subject-Parsing + Filterung
- `GetOffeneTische`, `GetOffeneSaldi`, `GetAusstehendAuszahlungen`: `table_state` → `tisch_session_state` mit `kassensitzung_nr`-Filter
- `GetEigeneUebersicht`: `FROM events` → `FROM kassenjournal`, Event-Typen aktualisieren

**`relay.sql` (Zeilen 1–5):**

- `FROM events` → `FROM kassenjournal`
- Event-Typ: `'tisch.bestellung-aufgenommen:v1'` → `'bestellung-aufgenommen:v1'`

#### 2.4 sqlc-Konfiguration (`backend/sqlc.yaml`)

- Ggf. Query-Dateinamen anpassen nach Umbenennung

#### 2.5 Domain-Schicht (`backend/domain/`)

**`domain/event/event.go` (Zeilen 1–100):**

- Struct `Event` bleibt strukturell identisch (redesign.md §3.13: „Event-Envelope bleibt unverändert")
- Ggf. Kommentare aktualisieren (Subject-Format-Beispiele)

**`domain/table/` — Aufspaltung:**

Aktuell enthält `domain/table/` sowohl Stammdaten-Logik als auch Event-Sourcing-Logik. Die Event-Sourcing-Logik muss nach `domain/kasse/` verschoben werden.

**Dateien, die in `domain/table/` BLEIBEN (nur Stammdaten):**

- `tisch.go` (Zeilen 1–100): Bleibt — Tisch als reine Stammdaten-Entität (CRUD). Typen `AktiverTisch` und `AktiverTischMitFavorit` müssen angepasst werden: `SaldoCents` kommt jetzt aus `tisch_session_state`, nicht aus `table_state`

**Dateien, die nach `domain/kasse/` VERSCHOBEN werden:**

- `events.go` (Zeilen 1–200): Event-Typen, `HistorieEintrag`, `accumulatePositionen`, `reduceByPosition`
- `projection.go` (Zeilen 1–100): `TischState`, `ApplyEvent`, `ComputeNichtStorniertePositionen`
- `projection_test.go`
- `events_test.go`
- `bestellung.go` (Zeilen 1–100): `Position`, `Bestellung`, Schemas
- `bestellungAufgenommenEvent.go`: Event-Erstellung — Subject-Format muss ändern
- `zahlungKassiertEvent.go`: Gleiches
- `stornierungErteiltEvent.go`: Gleiches
- `ausgabeBestaetigtEvent.go`: Gleiches
- `auszahlungGeleistetEvent.go`: Gleiches
- `zahlung.go`, `stornierung.go`, `ausgabe.go`: Domain-Typen für Events

**Neue Dateien in `domain/kasse/`:**

- `kassensitzung.go`: Kassensitzung-State, Replay, Invarianten
- `kassensitzung_events.go`: 6 neue Event-Typen + Data-Structs (redesign.md §3.6)
- `kassenbestand.go`: Soll-Berechnung (ggf. nur als SQL-Query in der Repo-Schicht)
- `invarianten.go`: Geteilte Positions-Logik, Saldo-Berechnung

**Event-Typ-Änderungen (in `domain/table/events.go` Zeilen 33–37, zukünftig `domain/kasse/`):**

- `"tisch.bestellung-aufgenommen:v1"` → `"bestellung-aufgenommen:v1"`
- `"tisch.zahlung-kassiert:v1"` → `"zahlung-kassiert:v1"`
- `"tisch.stornierung-erteilt:v1"` → `"stornierung-erteilt:v1"`
- `"tisch.ausgabe-bestaetigt:v1"` → `"ausgabe-bestaetigt:v1"`
- `"tisch.auszahlung-geleistet:v1"` → `"auszahlung-geleistet:v1"`

**Subject-Format-Änderungen (in allen Event-Erstellungsfunktionen):**

- Aktuell: `"tisch:" + strconv.Itoa(tischID)` (z.B. in `bestellungAufgenommenEvent.go:49`)
- Neu: Subject wird vom Application Service übergeben, Format `kassensitzung-{YYYYMMDD}-tisch-{tischID}`

#### 2.6 Repository-Schicht (`backend/repository/`)

**`event_repo/` → `kassenjournal_repo/`:**

- Paket und Verzeichnis umbenennen
- `repo.go` (Zeilen 1–200):
  - `WriteEvent`: StreamType-Routing implementieren (redesign.md §3.8)
    - `"kassensitzung"` → INSERT Kassenjournal + UPSERT `kassensitzung_state`
    - `"tisch-session"` → INSERT Kassenjournal + UPSERT `tisch_session_state`
  - `parseTischID`: Muss neues Subject-Format parsen (`kassensitzung-{YYYYMMDD}-tisch-{id}`)
  - `ReadTableState` → `ReadTischSessionState`: Liest nach Subject statt tischID
  - `toTischState`: Anpassen an neues Schema
  - `RebuildAllProjections` (indirekt referenziert in `main.go:95`): Muss beide Projektionen rebuilden
- `mock.go`: Anpassen an neue Interface-Methoden
- `repo_test.go`: Alle Tests anpassen (neue Subject-Formate, Event-Typen, Projektions-Schema)

**`table_repo/repo.go` (Zeilen 1–300):**

- `GetActiveTables`, `GetActiveTablesWithFavorites`: JOIN-Query ändert sich (auf `tisch_session_state` mit `kassensitzung_nr`)
- `GetTableStatesByIDs`: Wird zu `GetTischSessionStatesByKassensitzungNr` oder ähnlich

#### 2.7 Application-Schicht (`backend/api/`)

**`api/table/application/command.go` (Zeilen 1–200):**

- `writeEvent`: Subject wird anders konstruiert (inkl. Kassensitzung-Datum)
- `loadTischState`: Muss Kassensitzung-Sperre prüfen (KS-Projektion abfragen → `GetOffeneKassensitzung()`)
- Subject-Konstruktion: `"tisch:" + strconv.Itoa(tischID)` → `"kassensitzung-" + ksDate + "-tisch-" + strconv.Itoa(tischID)`
- Interface `eventRepo`: Methoden anpassen (WriteEvent mit StreamType)
- Zeilen ~306, 367, 390, 445, 472: Alle Subject-Konstruktionen ändern

**`api/table/application/query.go` (Zeilen 1–200):**

- `GetTischState`: Liest jetzt über Subject statt tischID
- `GetTischHistorie`: Subject-Konstruktion Zeile ~121: `"tisch:" + strconv.Itoa(tischID)` ändern
- `GetAktiveTischeMitFavoriten`: Muss `kassensitzung_nr` übergeben

**Neue Datei `api/kasse/` (oder Erweiterung von `api/table/`):**

- Application Service für Kassensitzung: `KassensitzungEroeffnen`, `AnfangsbestandSetzen`, `KassenbewegungBuchen`, `KassensturzDurchfuehren`, `TagesabschlussErstellen`
- HTTP Handler mit neuen Endpunkten

**`api/reporting/application/query.go` (Zeilen 1–50):**

- `GetReporting`: Parameter von `zeitraum reporting.Zeitraum` → `kassensitzungNr int`
- Interface `reportingRepo`: Methoden-Signatur anpassen

**`api/relay/application/query.go` (Zeilen 1–59):**

- `GetBestellungEventsSinceCursor`: Tabelle + Event-Typ ändern

**`api/relay/application/print.go` (Zeilen 1–100):**

- `parseTischName` (Zeile ~89): `tisch:{id}` → neues Subject-Parsing für `kassensitzung-{YYYYMMDD}-tisch-{id}`
- Ggf. Tischname aus Stammdaten laden statt aus Subject ableiten

#### 2.8 HTTP-Handler (`backend/api/*/http/`)

**`api/table/http/command_handler.go`:**

- Keine strukturelle Änderung für bestehende Endpunkte (API-Endpunkte bleiben gleich laut redesign.md §3.13)

**`api/table/http/query_handler.go`:**

- TischState-Response muss ggf. angepasst werden (neue Felder?)

**`api/reporting/http/query_handler.go`:**

- Request-Format: Von `von`/`bis` (Zeitraum) auf `kassensitzungNr`

**Neue Handler in `api/kasse/http/handler.go`:**

- `KassensitzungEroeffnenHandler`
- `AnfangsbestandSetzenHandler`
- `KassenbewegungBuchenHandler`
- `KassensturzDurchfuehrenHandler`
- `TagesabschlussErstellenHandler`

#### 2.9 Reporting-Domain (`backend/domain/reporting/`)

- `reporting.go`: `Zeitraum` struct wird ggf. ersetzt durch `KassensitzungNr int` als Filterkriterium
- Neue Report-Typen für Kassensitzung-basiertes Reporting

#### 2.10 API-Routing (`backend/api/service.go`, `admin.go`, `serviceleitung.go`)

**`service.go` (Zeilen 1–60):**

- Imports: `event_repo` → `kassenjournal_repo`
- `NewServiceApi`: Repository-Instanziierung anpassen

**`admin.go` (Zeilen 1–90):**

- Imports anpassen
- Neue Kassensitzung-Endpunkte registrieren (KS eröffnen, Anfangsbestand, etc.)

**`serviceleitung.go` (Zeilen 1–30):**

- Imports anpassen

**`relay.go`:**

- Import anpassen

#### 2.11 App-Setup (`backend/app/app.go`, `backend/main.go`)

**`main.go` (Zeilen 1–120):**

- Import: `event_repo` → `kassenjournal_repo`
- Zeile ~95: Rebuild-Projections Meldung anpassen
- `rebuildProjections`: Muss beide Projektionen rebuilden (`kassensitzung_state` + `tisch_session_state`)

**`app.go` (Zeilen 1–100):**

- Route `/admin/` und `/service/`: Ggf. neue Endpunkte einbinden

#### 2.12 Frontend

**Typ-Definitionen:**

- `frontend/src/service/table/Tisch.ts`: `TischState` bleibt strukturell ähnlich, aber die Backend-API könnte zusätzliche Felder liefern
- Ggf. neue Types für Kassensitzung-State

**Backend-Services:**

- `frontend/src/service/table/TischBackend.ts`: API-Endpunkte bleiben laut redesign.md §3.13 gleich → minimale Frontend-Änderungen
- Neue Kassensitzung-Backend-Klasse für Admin-UI (KS eröffnen, Kassenbewegung, etc.)

**Komponenten:**

- Neue Admin-Seiten: Kassensitzung verwalten, Kassenbewegungen, Kassensturz, Z-Bon
- Service-UI: Hinweis „Kasse ist noch nicht geöffnet" wenn keine offene KS (HTTP 409)

**Reporting:**

- `frontend/src/admin/`: Reporting-Queries mit `kassensitzungNr` statt Zeitraum

#### 2.13 Relay (`cmd/relay/`)

- Event-Typ-Erkennung anpassen
- Subject-Parsing für neues Format

#### 2.14 Docker / Makefile

- `Makefile`: `rebuild-projections` Target anpassen (neue Projektions-Logik)
- Keine Docker-Compose-Änderungen erwartet

---

### 3. Beziehungen und Abhängigkeiten

Die Änderungen haben folgende Abhängigkeitskette:

```
1. DB-Schema (01_initial.up.sql)
   ↓
2. sqlc-Queries (queries/*.sql) + make sqlc
   ↓
3. Domain-Schicht (domain/kasse/ neu, domain/table/ reduzieren)
   ↓
4. Repository-Schicht (kassenjournal_repo/ statt event_repo/)
   ↓
5. Application-Services (command.go, query.go)
   ↓
6. HTTP-Handler + API-Routing
   ↓
7. Frontend (Types, Backend-Services, Komponenten)
   ↓
8. Relay (Subject-Parsing, Event-Typ-Erkennung)
   ↓
9. Seed-Daten + Tests
   ↓
10. Dokumentation (alle Markdown-Dateien)
```

Die Dokumentation kann parallel zu den Code-Änderungen aktualisiert werden, sollte aber idealerweise **zuerst** erfolgen (da das Redesign-Dokument die neue Truth definiert und die übrigen Docs in Einklang gebracht werden müssen).

---

## Offene Fragen / Risiken

### 1. Kassensitzung-Pflicht für bestehende Tisch-Operationen

Das Redesign erfordert, dass **jeder** schreibende Tisch-Vorgang eine offene Kassensitzung voraussetzt (redesign.md §3.7). Das bedeutet:

- Alle bestehenden Command-Endpunkte (Bestellung, Zahlung, Ausgabe, Stornierung, Auszahlung) müssen die Kassensitzung-Sperre prüfen
- Der Application Service muss die offene KS kennen, um das Subject zu konstruieren
- **Risiko:** Breaking Change für alle Tisch-Operationen — gesamter Service-Flow ist betroffen

### 2. Subject-Konstruktion im Application Service

Das neue Subject-Format `kassensitzung-{YYYYMMDD}-tisch-{tischID}` erfordert, dass der Application Service das Datum der offenen Kassensitzung kennt. Das Datum kommt aus `kassensitzung_state`. Das bedeutet:

- Jeder Tisch-Command muss zuerst `GetOffeneKassensitzung()` aufrufen
- Das Datum wird aus der Projektion gelesen und ins Subject eingebaut
- **Frage:** Wird das Datum als `YYYYMMDD`-String gespeichert oder aus einem `DATE`-Feld formatiert?

### 3. Reihenfolge: Dokumentation vor Code oder parallel?

Das Redesign-Dokument fordert, dass Dokumentation „zuerst" aktualisiert wird. Allerdings ändern sich durch die Implementierung ggf. Details (z.B. exakte Methodennamen, Zeilennummern).

- **Empfehlung:** Dokumentation zuerst als konzeptuelle Aktualisierung, Code-Referenzen (Zeilennummern) nachträglich anpassen.

### 4. Seed-Daten Umfang

`database/seed.sql` enthält ~300 Event-Einträge mit dem alten Format. Eine manuelle Konvertierung aller Events ist aufwendig.

- **Empfehlung:** Seed-Daten komplett neu generieren (nicht manuell konvertieren)

### 5. Event-Data-Struct JSON-Keys

Die Event-Data-Structs (z.B. `bestellungAufgenommenV1Data`) haben `json`-Tags. Diese müssen unverändert bleiben, wenn die Datenstruktur gleich bleibt. Nur der Event-Type und das Subject ändern sich.

- **Erkenntnis:** Die `json`-Keys in Event-Data-Structs bleiben stabil (redesign.md §3.13: „Fat Events bleiben gleich")

### 6. `kassensitzung_nr` bei Tisch-Events

Beim Schreiben eines Tisch-Events muss die `kassensitzung_nr` mitgegeben werden. Diese kommt aus `kassensitzung_state`. Das bedeutet:

- Der Application Service liest `kassensitzung_state` VOR dem Event-Write
- Die `kassensitzung_nr` wird an das Repository übergeben
- **Event-Struct:** Das `Event`-Struct hat aktuell kein `KassensitzungNr`-Feld. Entweder wird es ergänzt oder als separater Parameter an `WriteEvent` übergeben.

### 7. Reporting-API Breaking Change

Die Umstellung von Zeitraum-basierter auf kassensitzungsbezogener Filterung ist ein Breaking Change der Reporting-API. Das Frontend muss statt `von`/`bis` eine `kassensitzungNr` senden.

- **Frontend-Impact:** Reporting-Dialog muss eine Kassensitzung auswählen statt einen Zeitraum einzugeben

### 8. Frontend-Umfang für Kassensitzung-Management

Die Admin-UI für Kassensitzung (Eröffnen, Anfangsbestand, Kassenbewegungen, Kassensturz, Z-Bon) ist eine komplett neue Feature-Gruppe. Der Frontend-Aufwand ist erheblich.

---

## Referenzen

| Datei                                                 | Zeilen                                          | Relevanz                                                     |
| ----------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------ |
| `docs/redesign.md`                                    | 1–866                                           | Source of Truth für das neue Design                          |
| `docs/handbuch.md`                                    | 84–113, 116–310, 477–605, 808–843               | Bounded Contexts, Core Domain, Kassenführung, Read Models    |
| `docs/anforderungen.md`                               | 68–526, 527–796, 797–943                        | Kassenbetrieb, Reporting, Kassenführung                      |
| `docs/language.md`                                    | 35–52, 111–206, 209–286                         | Abweichungen, Kassenbetrieb-Begriffe, Kassenführung-Begriffe |
| `docs/diagrams.md`                                    | 63, 139, 173, 206, 344, 392, 486, 532, 678, 794 | Alle betroffenen Diagramme                                   |
| `docs/roadmap.md`                                     | 52–127, 159–283                                 | Kassenführung-Phase, TSE-Phase                               |
| `docs/compliance.md`                                  | 158–342, 343–378, 490–623                       | TSE, GoBD, DSFinV-K                                          |
| `docs/tagesabschluss.md`                              | 1–726                                           | Tagesabschluss-Dokumentation                                 |
| `docs/bondruck.md`                                    | 1–Ende                                          | Relay/Bondruck-Dokumentation                                 |
| `docs/adr/event-sourcing.md`                          | 1–Ende                                          | Event-Sourcing ADR                                           |
| `docs/adr/cqrs.md`                                    | 1–Ende                                          | CQRS ADR                                                     |
| `AGENTS.md`                                           | 34–37, 130, 154                                 | Agent-Instruktionen                                          |
| `README.md`                                           | 12, 29, 75                                      | Projektbeschreibung                                          |
| `.github/instructions/backend.instructions.md`        | 1–Ende                                          | Backend-Konventionen                                         |
| `.github/instructions/database.instructions.md`       | 1–Ende                                          | Datenbank-Konventionen                                       |
| `.github/instructions/event-sourcing.instructions.md` | 1–Ende                                          | Event-Sourcing-Referenz                                      |
| `database/migrations/01_initial.up.sql`               | 102–170                                         | DB-Schema (events, table_state)                              |
| `database/seed.sql`                                   | 160–473                                         | Seed-Daten mit altem Format                                  |
| `backend/domain/event/event.go`                       | 1–100                                           | Event-Envelope                                               |
| `backend/domain/table/events.go`                      | 33–37, 40–42                                    | Event-Typen, Subject-Parsing                                 |
| `backend/domain/table/projection.go`                  | 1–100                                           | TischState, ApplyEvent                                       |
| `backend/domain/table/tisch.go`                       | 1–100                                           | Tisch-Stammdaten                                             |
| `backend/domain/table/bestellungAufgenommenEvent.go`  | 49                                              | Subject-Konstruktion `"tisch:"`                              |
| `backend/domain/table/ausgabeBestaetigtEvent.go`      | 36                                              | Subject-Konstruktion                                         |
| `backend/domain/table/zahlungKassiertEvent.go`        | 39                                              | Subject-Konstruktion                                         |
| `backend/domain/table/stornierungErteiltEvent.go`     | 39                                              | Subject-Konstruktion                                         |
| `backend/domain/table/auszahlungGeleistetEvent.go`    | 55                                              | Subject-Konstruktion                                         |
| `backend/domain/table/bestellung.go`                  | 1–100                                           | Position, Bestellung                                         |
| `backend/repository/event_repo/repo.go`               | 1–200                                           | WriteEvent, ReadTableState, parseTischID                     |
| `backend/repository/event_repo/mock.go`               | 1–130                                           | Mock-Repository                                              |
| `backend/repository/event_repo/repo_test.go`          | 1–570                                           | Event-Repo Tests                                             |
| `backend/repository/table_repo/repo.go`               | 1–300                                           | Tisch-CRUD + table_state Queries                             |
| `backend/repository/table_repo/types.go`              | 1–Ende                                          | Typ-Konvertierungen                                          |
| `backend/repository/reporting_repo/repo.go`           | 1–Ende                                          | Reporting-Repository                                         |
| `backend/api/table/application/command.go`            | 1–200                                           | Tisch-Commands                                               |
| `backend/api/table/application/query.go`              | 1–200                                           | Tisch-Queries                                                |
| `backend/api/table/application/errors.go`             | 1–Ende                                          | Fehler-Definitionen                                          |
| `backend/api/table/http/command_handler.go`           | 1–Ende                                          | HTTP-Handler Commands                                        |
| `backend/api/table/http/query_handler.go`             | 1–Ende                                          | HTTP-Handler Queries                                         |
| `backend/api/reporting/application/query.go`          | 1–50                                            | Reporting-Queries                                            |
| `backend/api/reporting/http/query_handler.go`         | 1–Ende                                          | Reporting-Handler                                            |
| `backend/api/relay/application/query.go`              | 1–59                                            | Relay-Queries                                                |
| `backend/api/relay/application/print.go`              | 1–100                                           | Relay-Drucklogik                                             |
| `backend/api/relay/http/handler.go`                   | 1–70                                            | Relay-Handler                                                |
| `backend/api/service.go`                              | 1–60                                            | Service-API Routing                                          |
| `backend/api/admin.go`                                | 1–90                                            | Admin-API Routing                                            |
| `backend/api/serviceleitung.go`                       | 1–30                                            | Serviceleitung-Routing                                       |
| `backend/app/app.go`                                  | 1–100                                           | App-Setup, Routing                                           |
| `backend/main.go`                                     | 1–120                                           | Main, Rebuild-Projections                                    |
| `backend/sqlc/queries/events.sql`                     | 1–17                                            | Event-SQL-Queries                                            |
| `backend/sqlc/queries/table_state.sql`                | 1–22                                            | Projektions-Queries                                          |
| `backend/sqlc/queries/reporting.sql`                  | 1–101                                           | Reporting-Queries                                            |
| `backend/sqlc/queries/relay.sql`                      | 1–5                                             | Relay-Query                                                  |
| `backend/sqlc/queries/tables.sql`                     | 12–19                                           | Tischübersicht-Queries                                       |
| `frontend/src/service/table/Tisch.ts`                 | 11–45                                           | Tisch-Types                                                  |
| `frontend/src/service/table/TischBackend.ts`          | 38–150                                          | API-Aufrufe                                                  |
| `frontend/src/service/table/hooks.ts`                 | 19–78                                           | React Hooks                                                  |
| `frontend/src/routes.ts`                              | 50–81                                           | Service-Routen                                               |
