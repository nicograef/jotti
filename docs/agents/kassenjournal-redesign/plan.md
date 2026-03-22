# Plan: Kassenjournal-Redesign — Vollständige Umsetzung

## Agent Anweisungen

> **Lies diese Anweisungen vollständig, bevor du mit der Arbeit beginnst.**

### Kontext laden (vor jedem Abschnitt)

Bevor du einen Abschnitt beanspruchst, lies **genau die Dateien, die im `Kontext:`-Block des Abschnitts aufgelistet sind** — nicht mehr, nicht weniger. Zusätzlich:

- Bereits erstellte/geänderte Dateien aus vorherigen Abschnitten lesen (um nahtlos anzuknüpfen)

Diese Dateien werden in jeder neuen Session erneut gelesen — die Kontext-Beschaffung ist kein eigener Abschnitt, sondern Pflicht vor jeder Arbeit.

### Abschnitt beanspruchen

1. **Lies die gesamte plan.md** — insbesondere den Parallelisierungs-Abschnitt und alle Abschnitts-Überschriften.
2. **Finde den nächsten verfügbaren Abschnitt.** Ein Abschnitt ist verfügbar, wenn:
   - Er offene Tasks hat (`- [ ]`)
   - Er **nicht** mit 🔒 oder ✅ markiert ist
   - Seine Abhängigkeiten erfüllt sind (alle Vorgänger-Abschnitte sind ✅)
3. **Beanspruche den Abschnitt sofort**, indem du `🔒` an die Überschrift anhängst (`## Abschnitt N: Titel` → `## Abschnitt N: Titel 🔒`). Erst danach mit der Arbeit beginnen.
4. **Falls kein verfügbarer Abschnitt existiert: Stoppe sofort, ohne Änderungen vorzunehmen.** Erkläre dem User: welche Abschnitte noch offen sind, warum sie nicht bearbeitet werden können (🔒 = anderer Agent arbeitet daran, oder Abhängigkeiten noch nicht ✅), und welche Vorgänger-Abschnitte zuerst abgeschlossen werden müssen. **Führe keine Änderungen an Dateien durch.**

### Abschnitt abarbeiten

1. **Ein Task nach dem anderen.** Arbeite Tasks innerhalb des Abschnitts sequentiell ab — von oben nach unten.
2. **Sofort abhaken.** Ändere `- [ ]` zu `- [x]` in dieser Datei **unmittelbar** nachdem ein Task erfolgreich erledigt ist. Nicht erst am Ende des Abschnitts, nicht gebündelt — **nach jedem einzelnen Task**. Verwende beim Abhaken immer die **Abschnitts-Überschrift + den vollständigen Task-Text** als Kontext, damit die Ersetzung eindeutig ist.
3. **Abschnitt abschließen.** Wenn du an Code gearbeitet hast: Wenn alle Tasks eines Abschnitts `[x]` sind, führe die wichtigsten Dev-Scripte und CI-Steps lokal aus: compilation, build, linting, formatting, testing. Stelle sicher, dass es keine Fehler oder Warnings gibt. Erst dann ist der Abschnitt fertig. Wenn du an Dokumentation gearbeitet hast: Lese Korrektur, stelle sicher, dass alle Links funktionieren, und dass die Formatierung korrekt ist.
4. **✅ setzen.** Ersetze `🔒` durch `✅` in der Abschnitts-Überschrift (`## Abschnitt N: Titel 🔒` → `## Abschnitt N: Titel ✅`).
5. **Stoppen.** Nach Abschluss eines Abschnitts: **stopp**. Beginne nicht den nächsten Abschnitt, sondern melde, dass der Abschnitt abgeschlossen ist.
6. **Conventional Commit Message schreiben.** Wenn du an Code gearbeitet hast: Schreibe zu deinen Änderungen bzw. dem Abschnitt eine Conventional Commit Message. Führe kein Commit selbst durch, schreibe nur die Message in den Chat, sodass diese kopiert werden kann. Wenn du an Dokumentation gearbeitet hast: Schreibe eine passende Commit Message für die Dokumentationsänderungen.

---

## Parallelisierung

Die folgenden Abschnitte können **parallel** in separaten Chat-Sessions bearbeitet werden:

- **Abschnitte 1–8** sind strikt sequentiell (jeder baut auf dem vorherigen auf):
  1 → 2 → 3 → 4 → 5 → 6 → 7 → 8

- **Abschnitt 9** (Dokumentation) kann **parallel zu Abschnitten 1–8** bearbeitet werden, da er ausschließlich Markdown-Dateien ändert und keine Code-Dateien berührt.

**Zusammenfassung:**

| Pfad A (Code)                 | Pfad B (Dokumentation) |
| ----------------------------- | ---------------------- |
| 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 | 9 (parallel zu allem)  |

**Datei-Überschneidungen:**

- Abschnitt 9 ändert `AGENTS.md`, `README.md`, `docs/*.md`, `.github/instructions/*.md`
- Abschnitte 1–8 ändern `database/`, `backend/`, `frontend/` — **keine Überschneidung** mit Abschnitt 9
- Innerhalb von 1–8 gibt es transitive Abhängigkeiten: Schema (1) → sqlc (2) → Domain (3) → Repository (4) → Application (5) → HTTP/Routing (6) → Frontend (7) → Seed/Tests (8)

---

## Abschnitt 1: DB-Schema — `kassenjournal`, `tisch_session_state`, `kassensitzung_state`

Kontext:

- `database/migrations/01_initial.up.sql:98-196` — aktuelles Schema (`events`, `table_state`, Trigger, Indizes, Grants)
- `docs/redesign.md:183-195` — neues `kassenjournal`-Schema
- `docs/redesign.md:483-511` — neues `kassensitzung_state`-Schema
- `docs/redesign.md:515-537` — neues `tisch_session_state`-Schema

### Tasks

- [ ] Tabelle `events` in `kassenjournal` umbenennen: Tabellenname, neue Spalte `kassensitzung_nr INT NOT NULL` hinzufügen (nach `data JSONB`), `UNIQUE (subject, version)` beibehalten
- [ ] Alle Indizes umbenennen: `idx_events_user_id` → `idx_kassenjournal_user_id`, `idx_events_subject` → `idx_kassenjournal_subject`, `idx_events_type` → `idx_kassenjournal_type`, `idx_events_subject_type` → `idx_kassenjournal_subject_type`, `idx_events_type_timestamp` → `idx_kassenjournal_type_timestamp`; neuen Index `idx_kassenjournal_ks_nr ON kassenjournal(kassensitzung_nr)` hinzufügen
- [ ] Immutabilitäts-Trigger anpassen: Funktionsname `prevent_event_mutation` → `prevent_kassenjournal_mutation`, alle drei Trigger-Namen (`events_no_update`, `events_no_delete`, `events_no_truncate`) auf `kassenjournal_no_update/delete/truncate` umbenennen, `ON events` → `ON kassenjournal`
- [ ] REVOKE/GRANT auf `kassenjournal` anpassen (statt `events`)
- [ ] Tabelle `table_state` durch `tisch_session_state` ersetzen: PK `subject TEXT` statt `tisch_id INTEGER`, neue Spalten `tisch_id INT NOT NULL REFERENCES tische(id)`, `kassensitzung_nr INT NOT NULL`, FK `last_event_id REFERENCES kassenjournal(id)`, Index `idx_tisch_session_state_ks_nr ON tisch_session_state(kassensitzung_nr)`
- [ ] Neue Tabelle `kassensitzung_state` erstellen: `subject TEXT PRIMARY KEY`, `z_nr INT UNIQUE NOT NULL`, `datum DATE NOT NULL`, `status TEXT NOT NULL CHECK (status IN ('offen', 'abgeschlossen'))`, `last_event_id INT NOT NULL REFERENCES kassenjournal(id)`, `last_event_version INT NOT NULL`
- [ ] Kommentare im Schema aktualisieren: Subject-Format-Beispiele (`kassensitzung-20260322`, `kassensitzung-20260322-tisch-42`), Event-Typ-Beispiele (ohne `tisch.`-Präfix)

---

## Abschnitt 2: SQL-Queries + sqlc-Generierung

Kontext:

- `backend/sqlc/queries/events.sql:1-17` — aktuelle Event-Queries
- `backend/sqlc/queries/table_state.sql:1-22` — aktuelle Projektions-Queries
- `backend/sqlc/queries/reporting.sql:1-102` — Reporting-Queries (Event-Typen, Tabellennamen, Zeitraum-Filter)
- `backend/sqlc/queries/relay.sql:1-7` — Relay-Query
- `backend/sqlc/queries/tables.sql:1-29` — Tischübersicht-Queries (JOIN auf `table_state`)
- `backend/sqlc.yaml:1-23` — sqlc-Konfiguration
- `docs/redesign.md:600-650` — Kassenbestand-Query, Reporting-Migration
- Ergebnis aus Abschnitt 1 (neues DB-Schema)

### Tasks

- [ ] Datei `backend/sqlc/queries/events.sql` löschen und neue Datei `backend/sqlc/queries/kassenjournal.sql` erstellen mit: `WriteEvent` (INSERT INTO `kassenjournal` mit neuer Spalte `kassensitzung_nr`, insgesamt 9 Parameter), `ReadEvent` (SELECT FROM `kassenjournal`), `ReadEventsBySubject` (SELECT FROM `kassenjournal` WHERE subject), `GetMaxVersion` (SELECT FROM `kassenjournal`), `GetDistinctSubjects` (SELECT FROM `kassenjournal`)
- [ ] Datei `backend/sqlc/queries/table_state.sql` löschen und neue Datei `backend/sqlc/queries/tisch_session_state.sql` erstellen mit: `UpsertTischSessionState` (INSERT/ON CONFLICT auf `subject`, Spalten: subject, tisch_id, kassensitzung_nr, saldo_cents, unbezahlte_positionen, ausstehende_positionen, gesamt_zahlungen_cents, last_event_id, last_event_version), `GetTischSessionState` (WHERE subject = $1), `GetTischSessionStatesByKassensitzungNr` (WHERE kassensitzung_nr = $1, für Tischübersicht), `DeleteAllTischSessionState` (DELETE FROM tisch_session_state)
- [ ] Neue Datei `backend/sqlc/queries/kassensitzung_state.sql` erstellen mit: `UpsertKassensitzungState` (INSERT/ON CONFLICT auf subject, Spalten: subject, z_nr, datum, status, last_event_id, last_event_version), `GetOffeneKassensitzung` (WHERE status = 'offen' LIMIT 1), `GetKassensitzungBySubject` (WHERE subject = $1), `GetNextZNr` (SELECT COALESCE(MAX(z_nr), 0) + 1 FROM kassensitzung_state), `DeleteAllKassensitzungState` (DELETE FROM kassensitzung_state)
- [ ] Datei `backend/sqlc/queries/reporting.sql` aktualisieren: Alle `FROM events` → `FROM kassenjournal`; alle Event-Typen `'tisch.bestellung-aufgenommen:v1'` → `'bestellung-aufgenommen:v1'` (analog für zahlung-kassiert, stornierung-erteilt, ausgabe-bestaetigt, auszahlung-geleistet); alle `WHERE timestamp >= @von AND timestamp < @bis` → `WHERE kassensitzung_nr = @kassensitzung_nr`; alle `table_state` → `tisch_session_state`; Subject-Parsing `SPLIT_PART(e.subject, ':', 2)::integer` → `tss.tisch_id` via JOIN auf `tisch_session_state tss ON tss.subject = e.subject`; Query-Parameter von `von`/`bis` auf `kassensitzung_nr` umstellen; `GetEigeneUebersicht` ebenfalls auf `kassenjournal` + `kassensitzung_nr` umstellen
- [ ] Datei `backend/sqlc/queries/relay.sql` aktualisieren: `FROM events` → `FROM kassenjournal`, Event-Typ `'tisch.bestellung-aufgenommen:v1'` → `'bestellung-aufgenommen:v1'`
- [ ] Datei `backend/sqlc/queries/tables.sql` aktualisieren: `GetAktiveTische` — JOIN auf `tisch_session_state` statt `table_state`, mit zusätzlichem Parameter `kassensitzung_nr` (`LEFT JOIN tisch_session_state tss ON tss.tisch_id = t.id AND tss.kassensitzung_nr = $1`); `GetAktiveTischeMitFavoriten` — gleiches Pattern mit `kassensitzung_nr`-Parameter (Position nach user_id-Parameter)
- [ ] `make sqlc` ausführen, um den generierten Code in `backend/sqlc/dbgen/` zu regenerieren; sicherstellen, dass keine Fehler auftreten

---

## Abschnitt 3: Domain-Schicht — `domain/kasse/` + `domain/table/` reduzieren

Kontext:

- `backend/domain/table/events.go:1-133` — Event-Typ-Konstanten, `parseTischIDFromSubject`, `GetHistoryFromEvents`, `accumulatePositionen`, `reduceByPosition`
- `backend/domain/table/projection.go:1-107` — `TischState`, `ApplyEvent`, `ComputeNichtStorniertePositionen`
- `backend/domain/table/bestellung.go:1-91` — `Position`, `Bestellung`, Schemas
- `backend/domain/table/zahlung.go`, `stornierung.go`, `ausgabe.go`, `auszahlung.go` — Domain-Typen
- `backend/domain/table/bestellungAufgenommenEvent.go:1-89` — Event-Data-Struct + `NewBestellungAufgenommenEvent()`
- `backend/domain/table/zahlungKassiertEvent.go`, `stornierungErteiltEvent.go`, `ausgabeBestaetigtEvent.go`, `auszahlungGeleistetEvent.go` — analog
- `backend/domain/table/tisch.go:1-107` — Tisch-Stammdaten, `AktiverTisch`, `AktiverTischMitFavorit`
- `backend/domain/event/event.go:1-112` — Event-Envelope
- `docs/redesign.md:357-426` — Event-Katalog + Event-Data-Strukturen
- `docs/redesign.md:427-478` — Invarianten
- `docs/redesign.md:698-730` — Paketstruktur

### Tasks

#### 3a: Neues Paket `domain/kasse/` erstellen

- [ ] Datei `backend/domain/kasse/stream_type.go` erstellen: `StreamType`-Typ als `string` mit zwei Konstanten `StreamTypeKassensitzung StreamType = "kassensitzung"` und `StreamTypeTischSession StreamType = "tisch-session"`
- [ ] Datei `backend/domain/kasse/tisch_session_events.go` erstellen: Event-Typ-Konstanten **ohne** `tisch.`-Präfix (`EventTypeBestellungAufgenommenV1 = "bestellung-aufgenommen:v1"` etc. für alle 5 Tisch-Events); Event-Data-Structs (`bestellungAufgenommenV1Data`, `zahlungKassiertV1Data`, `stornierungErteiltV1Data`, `ausgabeBestaetigtV1Data`, `auszahlungGeleistetV1Data`) mit `json`-Tags — **Vorlage:** bestehende Structs aus `domain/table/bestellungAufgenommenEvent.go` etc., identische Felder; Event-Erstellungsfunktionen (`NewBestellungAufgenommenEvent(subject string, ...)` — Subject wird als Parameter übergeben, nicht mehr intern aus tischID konstruiert); `positionEventData`-Struct (JSON-Repräsentation einer Position für Events) aus `bestellung.go` hierher verschieben
- [ ] Datei `backend/domain/kasse/kassensitzung_events.go` erstellen: 6 neue Event-Typ-Konstanten (`EventTypeKassensitzungEroeffnetV1 = "kassensitzung-eroeffnet:v1"`, `EventTypeAnfangsbestandGesetztV1`, `EventTypeKassenbewegungGebuchtV1`, `EventTypeKassensturzDurchgefuehrtV1`, `EventTypeDifferenzSollIstGebuchtV1`, `EventTypeTagesabschlussErstelltV1`); Event-Data-Structs mit `json`-Tags gemäß redesign.md §3.6 (`kassensitzungEroeffnetV1Data{Datum, Bezeichnung, EroeffnetVon}`, `anfangsbestandGesetztV1Data{BetragCents, GesetztVon}`, `kassenbewegungGebuchtV1Data{BewegungID, Art, BetragCents, Kommentar, GebuchtVon}`, `kassensturzDurchgefuehrtV1Data{SollBestandCents, IstBestandCents, DifferenzCents, DurchgefuehrtVon}`, `differenzSollIstGebuchtV1Data{BetragCents, GebuchtVon}`, `tagesabschlussErstelltV1Data{ZNr, ZeitraumVon, ZeitraumBis, UmsatzGesamtCents, StornierungCents, AuszahlungenCents, GeldtransitCents, ErstelltVon}`); Event-Erstellungsfunktionen für jeden Typ (`NewKassensitzungEroeffnetEvent(subject string, userID int, userName string, data)` etc.)
- [ ] Datei `backend/domain/kasse/tisch_session.go` erstellen: `TischSessionState`-Struct (Subject, TischID, KassensitzungNr, SaldoCents, UnbezahltePositionen, AusstehendePositionen, GesamtZahlungenCents, LastEventID, LastEventVersion) — **Vorlage:** `TischState` aus `projection.go`, erweitert um Subject, TischID, KassensitzungNr; `ApplyEvent(state *TischSessionState, e event.Event)` — identische Logik wie bisheriges `ApplyEvent` aus `projection.go`; `ComputeNichtStorniertePositionen(events []event.Event)` — identisch mit bisherigem Code
- [ ] Datei `backend/domain/kasse/kassensitzung.go` erstellen: `KassensitzungState`-Struct (Subject, ZNr, Datum, Status, LastEventID, LastEventVersion); Status-Konstanten `KassensitzungStatusOffen = "offen"`, `KassensitzungStatusAbgeschlossen = "abgeschlossen"`
- [ ] Datei `backend/domain/kasse/subject.go` erstellen: Helper-Funktionen für Subject-Konstruktion und -Parsing: `KassensitzungSubject(datum string) string` → `"kassensitzung-" + datum`; `TischSessionSubject(datum string, tischID int) string` → `"kassensitzung-" + datum + "-tisch-" + strconv.Itoa(tischID)`; `ParseTischIDFromSubject(subject string) (int, error)` — extrahiert tischID aus `kassensitzung-{YYYYMMDD}-tisch-{id}` (Segment nach letztem `-tisch-`); `ParseDatumFromSubject(subject string) (string, error)` — extrahiert `YYYYMMDD` aus Subject
- [ ] Datei `backend/domain/kasse/bestellung.go` erstellen: `Position`-Struct (PositionID, VarianteID, ProduktName, VarianteName, Kategorie, Einzelpreis, Menge) — identisch mit bisherigem `table.Position`; `Bestellung`-Struct (ID, UserID, TischID, Positionen, GesamtPreisCents, Kommentar, AufgenommenAm); `PositionRef`-Struct (PositionID, Menge); Zod-Schemas für Bestellung/Position — **Vorlage:** `domain/table/bestellung.go`, 1:1 kopieren
- [ ] Datei `backend/domain/kasse/zahlung.go` erstellen: `Zahlung`-Struct — Vorlage: `domain/table/zahlung.go`
- [ ] Datei `backend/domain/kasse/stornierung.go` erstellen: `Stornierung`-Struct — Vorlage: `domain/table/stornierung.go`
- [ ] Datei `backend/domain/kasse/ausgabe.go` erstellen: `Ausgabe`-Struct — Vorlage: `domain/table/ausgabe.go`
- [ ] Datei `backend/domain/kasse/auszahlung.go` erstellen: `Auszahlung`-Struct — Vorlage: `domain/table/auszahlung.go`
- [ ] Datei `backend/domain/kasse/historie.go` erstellen: `HistorieEintrag`-Struct und `GetHistoryFromEvents(events []event.Event)` — Vorlage: `domain/table/events.go:34-133` (Funktion `GetHistoryFromEvents` + `accumulatePositionen` + `reduceByPosition`); Event-Typ-Referenzen auf neue Konstanten ohne `tisch.`-Präfix aktualisieren

#### 3b: `domain/table/` auf Stammdaten reduzieren

- [ ] Alle Event-Sourcing-Dateien aus `backend/domain/table/` löschen: `events.go`, `events_test.go`, `projection.go`, `projection_test.go`, `bestellungAufgenommenEvent.go`, `zahlungKassiertEvent.go`, `stornierungErteiltEvent.go`, `ausgabeBestaetigtEvent.go`, `auszahlungGeleistetEvent.go`, `bestellung.go`, `zahlung.go`, `stornierung.go`, `ausgabe.go`, `auszahlung.go` (nur `tisch.go` bleibt)
- [ ] In `backend/domain/table/tisch.go` anpassen: `AktiverTisch` und `AktiverTischMitFavorit` bleiben (SaldoCents kommt jetzt aus `tisch_session_state`, aber das Feld bleibt gleich); sicherstellen, dass keine Imports auf gelöschte Dateien verweisen; ggf. nicht mehr benötigte Imports entfernen

#### 3c: Tests für `domain/kasse/`

- [ ] Datei `backend/domain/kasse/tisch_session_test.go` erstellen: Tests für `ApplyEvent` — Vorlage: `domain/table/projection_test.go` (alle Test-Cases übernehmen, Event-Typen auf neue Konstanten anpassen, Subject-Format auf `kassensitzung-20260322-tisch-42`); Tests für `ComputeNichtStorniertePositionen`
- [ ] Datei `backend/domain/kasse/historie_test.go` erstellen: Tests für `GetHistoryFromEvents` — Vorlage: `domain/table/events_test.go` (Event-Typen + Subjects anpassen)
- [ ] Datei `backend/domain/kasse/subject_test.go` erstellen: Tests für `KassensitzungSubject`, `TischSessionSubject`, `ParseTischIDFromSubject`, `ParseDatumFromSubject`

---

## Abschnitt 4: Repository-Schicht — `kassenjournal_repo/` (ersetzt `event_repo/`)

Kontext:

- `backend/repository/event_repo/repo.go:1-318` — aktuelles Repository (WriteEvent, ReadTableState, RebuildAllProjections, etc.)
- `backend/repository/event_repo/mock.go:1-130` — Mock-Repository
- `backend/repository/event_repo/repo_test.go:1-570` — Integrationstests
- `docs/redesign.md:539-580` — Write-Through-Pseudocode, StreamType-Routing
- Ergebnis aus Abschnitt 2 (generierter sqlc-Code in `sqlc/dbgen/`)
- Ergebnis aus Abschnitt 3 (`domain/kasse/` Typen)

### Tasks

- [ ] Neues Verzeichnis `backend/repository/kassenjournal_repo/` erstellen
- [ ] Datei `backend/repository/kassenjournal_repo/repo.go` erstellen: **Vorlage:** `event_repo/repo.go`. Struct `Repo` mit `db *sql.DB`. Interface-Methoden:
  - `WriteEvent(ctx, tx, event event.Event, streamType kasse.StreamType, kassensitzungNr int) (int, error)` — Routing: INSERT INTO kassenjournal (mit `kassensitzung_nr`), dann je nach `streamType`: `"kassensitzung"` → UPSERT `kassensitzung_state` (bei `kassensitzung-eroeffnet:v1` INSERT mit `z_nr` aus `GetNextZNr`, bei `tagesabschluss-erstellt:v1` SET status='abgeschlossen', sonst nur last_event_id/version aktualisieren); `"tisch-session"` → UPSERT `tisch_session_state` (identisch mit bisherigem Pattern, aber mit neuem Schema: subject als PK, tisch_id + kassensitzung_nr)
  - `ReadEvent(ctx, id int)` — identisch, nur Tabellenname
  - `ReadEventsBySubject(ctx, subject string)` — identisch, nur Tabellenname
  - `GetMaxVersion(ctx, subject string)` — identisch
  - `ReadTischSessionState(ctx, subject string)` — liest `tisch_session_state` nach Subject (statt `table_state` nach tischID). Gibt Zero-Value zurück wenn kein Eintrag existiert.
  - `GetOffeneKassensitzung(ctx)` — liest `kassensitzung_state` WHERE status='offen'
  - `GetBestellungEventsSinceCursor(ctx, cursor int)` — identisch, Tabellenname + Event-Typ anpassen
  - `RebuildAllProjections(ctx)` — DELETE FROM `tisch_session_state` + DELETE FROM `kassensitzung_state`, dann alle Events replayed: je nach Subject-Muster → `parseTischID(subject)` für Tisch-Events, UPSERT entsprechende Projektion; `kassensitzung-{YYYYMMDD}` (ohne `-tisch-`) → UPSERT `kassensitzung_state`
  - Helper: `parseTischID(subject string)` — delegiert an `kasse.ParseTischIDFromSubject()`; `toTischSessionState()` — JSON-Unmarshaling analog zu bisherigem `toTischState()`
- [ ] Datei `backend/repository/kassenjournal_repo/mock.go` erstellen: Mock-Struct mit denselben Interface-Methoden für Unit-Tests — Vorlage: `event_repo/mock.go`, erweitert um neue Methoden (`GetOffeneKassensitzung`, `ReadTischSessionState`)
- [ ] Datei `backend/repository/kassenjournal_repo/repo_test.go` erstellen: Integrationstests — Vorlage: `event_repo/repo_test.go`. Tests anpassen: Subject-Format `kassensitzung-20260322-tisch-1`, Event-Typen ohne `tisch.`-Präfix, `kassensitzung_nr`-Parameter, StreamType-Parameter. Neue Tests: WriteEvent mit `StreamTypeKassensitzung` → `kassensitzung_state` UPSERT prüfen; WriteEvent mit `StreamTypeTischSession` → `tisch_session_state` UPSERT prüfen; `RebuildAllProjections` mit beiden Projektionen; `GetOffeneKassensitzung` → offene/keine offene KS
- [ ] Verzeichnis `backend/repository/event_repo/` löschen (repo.go, mock.go, repo_test.go)

---

## Abschnitt 5: Application-Schicht — Commands + Queries anpassen

Kontext:

- `backend/api/table/application/command.go:1-483` — Tisch-Commands (writeEvent, loadTischState, Bestellung/Zahlung/Stornierung/Ausgabe/Auszahlung)
- `backend/api/table/application/query.go:1-136` — Tisch-Queries (GetTischState, GetTischHistorie, GetAktiveTische)
- `backend/api/table/application/errors.go` — Fehler-Definitionen
- `backend/api/table/application/command_test.go` — Command-Unit-Tests
- `backend/api/table/application/query_test.go` — Query-Unit-Tests
- `backend/api/reporting/application/query.go:1-46` — Reporting-Queries (GetReporting mit Zeitraum)
- `backend/api/relay/application/query.go:1-59` — Relay-Queries (GetDruckAuftraege)
- `backend/api/relay/application/print.go:1-100` — Drucklogik (Subject-Parsing)
- `backend/domain/reporting/` — Reporting-Typen (Zeitraum)
- Ergebnis aus Abschnitt 3 (`domain/kasse/` Typen, Subject-Helpers)
- Ergebnis aus Abschnitt 4 (`kassenjournal_repo/` Interface)

### Tasks

#### 5a: Tisch-Command-Service anpassen

- [ ] In `backend/api/table/application/command.go`: Import `event_repo` → `kassenjournal_repo`, Import `domain/table` → `domain/kasse` (für Event-Typen, Position, Bestellung etc.); `eventRepo`-Interface aktualisieren: `WriteEvent(ctx, tx, event, streamType, kassensitzungNr)`, `ReadTischSessionState(subject)` statt `ReadTableState(tischID)`, `GetOffeneKassensitzung(ctx)` hinzufügen, `GetMaxVersion(subject)`, alle `ReadEventsBySubject(subject)`
- [ ] In `command.go`: Neue Hilfsmethode `getOffeneKassensitzungOderFehler(ctx) (*kasse.KassensitzungState, error)` — ruft `EventRepo.GetOffeneKassensitzung(ctx)` auf, gibt `ErrKasseNichtGeoeffnet` (HTTP 409) zurück wenn nil
- [ ] In `command.go`: `loadTischState(ctx, tischID)` anpassen → ruft zuerst `getOffeneKassensitzungOderFehler` auf, konstruiert Subject via `kasse.TischSessionSubject(ks.Datum, tischID)`, liest dann `ReadTischSessionState(subject)` statt `ReadTableState(tischID)`; gibt Subject + KS-Nr + State zurück
- [ ] In `command.go`: `writeEvent(ctx, event, streamType, kassensitzungNr)` anpassen — übergibt `streamType` und `kassensitzungNr` an Repo
- [ ] In `command.go`: `computeNichtStorniertePositionen(ctx, subject)` — Subject statt tischID, delegiert an `kasse.ComputeNichtStorniertePositionen`
- [ ] In `command.go`: `BestellungAufnehmen()` anpassen — Subject-Konstruktion über `kasse.TischSessionSubject(ks.Datum, tischID)`, Event via `kasse.NewBestellungAufgenommenEvent(subject, ...)`, `writeEvent` mit `StreamTypeTischSession` + `ks.ZNr`
- [ ] In `command.go`: `ZahlungKassieren()`, `StornierungErteilen()`, `AusgabeBestaetigen()`, `AuszahlungLeisten()` — gleiche Anpassungen wie `BestellungAufnehmen`: KS laden, Subject konstruieren, Event-Erstellung mit neuem Funktionsnamen (z.B. `kasse.NewZahlungKassiertEvent(subject, ...)`), `StreamTypeTischSession` + `ks.ZNr`
- [ ] In `errors.go`: Neuen Fehler `ErrKasseNichtGeoeffnet` hinzufügen (für HTTP 409, wenn keine offene Kassensitzung)

#### 5b: Tisch-Query-Service anpassen

- [ ] In `backend/api/table/application/query.go`: Imports aktualisieren (`domain/kasse` statt `domain/table` wo nötig); `eventRepo`-Interface aktualisieren analog zu command.go
- [ ] In `query.go`: `GetAktiveTische()` und `GetAktiveTischeMitFavoriten()` — `tableRepo`-Interface muss `kassensitzungNr` akzeptieren; ermittle offene KS (oder 0 wenn keine offen) und übergebe KS-Nr an `GetActiveTables(ctx, ksNr)` / `GetActiveTablesWithFavorites(ctx, userID, ksNr)`
- [ ] In `query.go`: `GetTischState(tischID)` — ermittle offene KS, konstruiere Subject, lese `ReadTischSessionState(subject)` statt `ReadTableState(tischID)`
- [ ] In `query.go`: `GetTischHistorie(tischID)` — ermittle offene KS, konstruiere Subject, `ReadEventsBySubject(subject)` mit neuem Subject-Format; `kasse.GetHistoryFromEvents()` statt `table.GetHistoryFromEvents()`
- [ ] In `query.go`: `GetMeineTischeState(userID)` — analog zu `GetTischState`, aber für alle Favoriten; Subject-Konstruktion pro Tisch, `ReadTischSessionState(subject)`

#### 5c: Neuer Kassensitzung-Application-Service

- [ ] Neues Verzeichnis `backend/api/kasse/application/` erstellen
- [ ] Datei `backend/api/kasse/application/command.go` erstellen: `Command`-Struct mit `EventRepo eventRepo`; Interface `eventRepo` (WriteEvent, GetMaxVersion, GetOffeneKassensitzung, ReadEventsBySubject); Methoden:
  - `KassensitzungEroeffnen(ctx, userID, userName, datum, bezeichnung)` — Prüfe keine offene KS existiert, konstruiere Subject `kasse.KassensitzungSubject(datum)`, erstelle Event `kasse.NewKassensitzungEroeffnetEvent(subject, ...)`, WriteEvent mit `StreamTypeKassensitzung` + z_nr (wird im Repo berechnet)
  - `AnfangsbestandSetzen(ctx, userID, userName, betragCents)` — Prüfe offene KS existiert, Replay KS-Events → prüfe kein Anfangsbestand bereits gesetzt (Invariante), WriteEvent `anfangsbestand-gesetzt:v1`
  - `KassenbewegungBuchen(ctx, userID, userName, art, betragCents, kommentar)` — Prüfe offene KS, WriteEvent `kassenbewegung-gebucht:v1`
  - `KassensturzDurchfuehren(ctx, userID, userName, istBestandCents)` — Prüfe offene KS, berechne Soll-Bestand (SQL-Query), erstelle Kassensturz-Event, bei Differenz ≠ 0 zweites Event `differenz-soll-ist-gebucht:v1` in derselben TX
  - `TagesabschlussErstellen(ctx, userID, userName)` — Prüfe offene KS, prüfe Kassensturz durchgeführt (Replay), prüfe alle Tisch-AKs Saldo = 0 (Query auf `tisch_session_state`), WriteEvent `tagesabschluss-erstellt:v1`
- [ ] Datei `backend/api/kasse/application/query.go` erstellen: `Query`-Struct mit `EventRepo eventRepo`; Methoden:
  - `GetOffeneKassensitzung(ctx)` — liest Projektion
  - `GetKassenbestand(ctx, kassensitzungNr)` — SQL-Query über Kassenjournal (redesign.md §3.9)

#### 5d: Reporting-Application-Service anpassen

- [ ] In `backend/api/reporting/application/query.go`: `GetReporting(ctx, kassensitzungNr int)` statt `GetReporting(ctx, zeitraum reporting.Zeitraum)`; `reportingRepo`-Interface: Query-Methoden mit `kassensitzungNr` statt `zeitraum`
- [ ] In `backend/domain/reporting/`: `Zeitraum`-Struct entfernen oder durch `kassensitzungNr`-basierte Filterung ersetzen; Reporting-Typen (`ReportingData`, etc.) bleiben, Eingabeparameter ändern sich

#### 5e: Relay-Application-Service anpassen

- [ ] In `backend/api/relay/application/query.go`: Import `event_repo` → `kassenjournal_repo`; `eventRepo`-Interface: `GetBestellungEventsSinceCursor` (Query bleibt funktional gleich, nur Tabellen-/Event-Typname geändert)
- [ ] In `backend/api/relay/application/print.go`: `parseTischName()` / Subject-Parsing anpassen — neues Format `kassensitzung-{YYYYMMDD}-tisch-{id}` statt `tisch:{id}`; nutze `kasse.ParseTischIDFromSubject()` + TableRepo-Lookup für Tischname

#### 5f: Tests aktualisieren

- [ ] In `backend/api/table/application/command_test.go`: Alle Tests anpassen — Mock-Interface aktualisieren, Event-Typen ohne `tisch.`-Präfix, Subject-Format neu, `kassensitzungNr`-Parameter, Mock für `GetOffeneKassensitzung` hinzufügen
- [ ] In `backend/api/table/application/query_test.go`: Analog — neue Subject-Formate, Imports aktualisieren
- [ ] Tests für den neuen `api/kasse/application/` Service: mindestens Basis-Tests für `KassensitzungEroeffnen`, Invariantenprüfung (keine doppelte offene KS), `TagesabschlussErstellen` (Saldo-Sperre)

---

## Abschnitt 6: HTTP-Handler + Routing

Kontext:

- `backend/api/table/http/command_handler.go` — Tisch-Command-Handler
- `backend/api/table/http/query_handler.go:1-426` — Tisch-Query-Handler + Response-DTOs
- `backend/api/reporting/http/query_handler.go:1-263` — Reporting-Handler (Request: von/bis)
- `backend/api/relay/http/handler.go` — Relay-Handler
- `backend/api/service.go` — Service-Routing
- `backend/api/admin.go` — Admin-Routing
- `backend/api/serviceleitung.go` — Serviceleitung-Routing
- `backend/api/relay.go` — Relay-Routing
- `backend/app/app.go:1-71` — App-Setup
- `backend/main.go:1-101` — Main, rebuild-projections
- Ergebnis aus Abschnitt 5 (Application-Services)

### Tasks

#### 6a: Bestehende Handler + DTOs anpassen

- [ ] In `backend/api/table/http/query_handler.go`: Response-DTOs bleiben strukturell gleich; Imports von `domain/table` auf `domain/kasse` aktualisieren wo nötig (für `Position`, `Bestellung`, `HistorieEintrag` etc.); Mapping-Funktionen (`toBestellung`, `toZahlung`, etc.) an neue Paketnamen anpassen
- [ ] In `backend/api/table/http/command_handler.go`: Imports aktualisieren (Command-Struct nutzt jetzt `domain/kasse`-Typen); wenn `Position` oder `PositionRef` zu `kasse.Position` / `kasse.PositionRef` geworden sind, Mapping anpassen
- [ ] In `backend/api/reporting/http/query_handler.go`: Request-Struct ändern — von `{ von: time.Time, bis: time.Time }` auf `{ kassensitzungNr: int }`; Validierung anpassen (kassensitzungNr > 0); Aufruf `query.GetReporting(ctx, kassensitzungNr)` statt `query.GetReporting(ctx, zeitraum)`

#### 6b: Neuen Kassensitzung-Handler erstellen

- [ ] Neues Verzeichnis `backend/api/kasse/http/` erstellen
- [ ] Datei `backend/api/kasse/http/handler.go` erstellen: `Handler`-Struct mit `Command *application.Command, Query *application.Query`; Handler-Funktionen (alle POST):
  - `KassensitzungEroeffnenHandler` — Request: `{ datum: string, bezeichnung: string }`, Validierung, ruft `Command.KassensitzungEroeffnen()`
  - `AnfangsbestandSetzenHandler` — Request: `{ betragCents: int }`, ruft `Command.AnfangsbestandSetzen()`
  - `KassenbewegungBuchenHandler` — Request: `{ art: string, betragCents: int, kommentar: string }`, ruft `Command.KassenbewegungBuchen()`
  - `KassensturzDurchfuehrenHandler` — Request: `{ istBestandCents: int }`, ruft `Command.KassensturzDurchfuehren()`
  - `TagesabschlussErstellenHandler` — Request: `{}` (leer), ruft `Command.TagesabschlussErstellen()`
  - `GetOffeneKassensitzungHandler` — Query-Handler, ruft `Query.GetOffeneKassensitzung()`
  - `GetKassenbestandHandler` — Request: `{ kassensitzungNr: int }`, ruft `Query.GetKassenbestand()`
  - Alle Handler folgen dem bestehenden Pattern aus `table/http/command_handler.go` (Request parsen → validieren → Service aufrufen → Response senden)

#### 6c: Routing aktualisieren

- [ ] In `backend/api/admin.go`: Imports von `event_repo` → `kassenjournal_repo`; neuen Kasse-Handler instanziieren + neue Routen registrieren: `/admin/kassensitzung-eroeffnen`, `/admin/anfangsbestand-setzen`, `/admin/kassenbewegung-buchen`, `/admin/kassensturz-durchfuehren`, `/admin/tagesabschluss-erstellen`, `/admin/get-offene-kassensitzung`, `/admin/get-kassenbestand`
- [ ] In `backend/api/service.go`: Import `event_repo` → `kassenjournal_repo`; Repository-Instanziierung anpassen
- [ ] In `backend/api/serviceleitung.go`: Import anpassen
- [ ] In `backend/api/relay.go`: Import `event_repo` → `kassenjournal_repo`; Repository-Instanziierung anpassen

#### 6d: Main + App anpassen

- [ ] In `backend/main.go`: Import `event_repo` → `kassenjournal_repo`; `rebuild-projections` Subcommand: `kassenjournal_repo.New(db).RebuildAllProjections(ctx)` statt `event_repo`; Log-Meldung: „Kassensitzung- und Tisch-Session-Projektionen" statt „Table-State-Projektionen"
- [ ] In `backend/app/app.go`: Imports anpassen; in `SetupRoutes()`: `NewAdminApi` erhält `kassenjournal_repo`-Instanz für Kasse-Handler
- [ ] In `backend/repository/table_repo/repo.go`: Alle JOINs auf `table_state` → `tisch_session_state` aktualisieren; `GetActiveTables(ctx, kassensitzungNr)` — neuer Parameter für den JOIN; `GetActiveTablesWithFavorites(ctx, userID, kassensitzungNr)` — analog; `GetTableStatesByIDs` entfernen oder durch `GetTischSessionStatesByKassensitzungNr` ersetzen

#### 6e: Compilation + Lint prüfen

- [ ] `make build` ausführen — sicherstellen, dass Backend kompiliert
- [ ] `make lint` ausführen — sicherstellen, dass keine Lint-Fehler existieren

---

## Abschnitt 7: Frontend-Anpassungen

Kontext:

- `frontend/src/service/table/Tisch.ts:1-45` — Tisch-Types + Zod-Schemas
- `frontend/src/service/table/TischBackend.ts:1-150` — API-Aufrufe
- `frontend/src/service/table/hooks.ts:1-78` — React Hooks
- `frontend/src/admin/reporting/` — Reporting-Seiten (ReportingBackend.ts, types.ts, hooks.ts, ReportingFilter.tsx, ReportingResults.tsx, AdminDashboardPage.tsx)
- `frontend/src/routes.ts` — Routen
- Ergebnis aus Abschnitt 6 (neue/geänderte API-Endpunkte)

### Tasks

#### 7a: Reporting-Frontend auf `kassensitzungNr` umstellen

- [ ] In `frontend/src/admin/reporting/types.ts`: Reporting-Request-Typ von `{ von: string, bis: string }` auf `{ kassensitzungNr: number }` ändern; Zod-Schemas anpassen
- [ ] In `frontend/src/admin/reporting/ReportingBackend.ts`: API-Aufruf für Abrechnung: Payload `{ kassensitzungNr }` statt `{ von, bis }`
- [ ] In `frontend/src/admin/reporting/hooks.ts`: Hook-Parameter von Zeitraum auf kassensitzungNr umstellen
- [ ] In `frontend/src/admin/reporting/ReportingFilter.tsx`: Filter-UI von Datums-Eingabe auf KS-Auswahl umstellen (Dropdown/Select mit offener Kassensitzung oder Nummer-Eingabe); neuen API-Aufruf zum Laden der offenen Kassensitzung einbauen
- [ ] In `frontend/src/admin/reporting/AdminDashboardPage.tsx` + `ReportingResults.tsx`: Anpassungen an geändertes Filter-/Daten-Format

#### 7b: Kassensitzung-Admin-UI (Grundgerüst)

- [ ] Neues Verzeichnis `frontend/src/admin/kasse/` erstellen
- [ ] Datei `frontend/src/admin/kasse/types.ts` erstellen: Zod-Schemas + Types für `KassensitzungState` (`{ subject, zNr, datum, status }`), `Kassenbestand` (`{ sollBestandCents }`)
- [ ] Datei `frontend/src/admin/kasse/KasseBackend.ts` erstellen: Backend-Klasse — Vorlage: `TischBackend.ts`-Pattern; Methoden: `kassensitzungEroeffnen(datum, bezeichnung)`, `anfangsbestandSetzen(betragCents)`, `kassenbewegungBuchen(art, betragCents, kommentar)`, `kassensturzDurchfuehren(istBestandCents)`, `tagesabschlussErstellen()`, `getOffeneKassensitzung()`, `getKassenbestand(kassensitzungNr)`
- [ ] Datei `frontend/src/admin/kasse/hooks.ts` erstellen: `useOffeneKassensitzung()` — Hook zum Laden der offenen KS-Projektion; `useKassenbestand(kassensitzungNr)` — Hook zum Laden des Soll-Bestands
- [ ] Datei `frontend/src/admin/kasse/KassensitzungPage.tsx` erstellen: Grundgerüst-Seite für Kassensitzung-Verwaltung: Anzeige offene KS (oder „Keine Kassensitzung geöffnet"), Button „Kassensitzung eröffnen", Anfangsbestand setzen, Kassenbewegung buchen, Kassensturz, Tagesabschluss
- [ ] In `frontend/src/routes.ts`: Neue Route `/admin/kasse` → `KassensitzungPage`; Import und Routing-Eintrag hinzufügen
- [ ] In Admin-Sidebar (`frontend/src/admin/AdminSidebar.tsx`): Neuen Menüpunkt „Kasse" mit Link zu `/admin/kasse` hinzufügen

#### 7c: Service-UI — Kasse-nicht-geöffnet-Handling

- [ ] In den Service-Hooks/Error-Handling: HTTP 409 von Tisch-Commands als „Kasse ist noch nicht geöffnet"-Hinweis anzeigen (Toast/Alert); `TischBackend.ts` oder globales Error-Handling erweitern

#### 7d: Frontend-Tests

- [ ] Vorhandene Frontend-Tests (`hooks.test.ts`, `routes.test.ts`) aktualisieren: neue Routen, geänderte Typen

---

## Abschnitt 8: Seed-Daten + Integrationstests + Abschluss

Kontext:

- `database/seed.sql:1-3259` — aktuelle Seed-Daten (Events mit altem Format, table_state)
- `backend/repository/event_repo/repo_test.go` — war in Abschnitt 4 bereits als `kassenjournal_repo/repo_test.go` neugeschrieben
- `backend/api/health/health_integration_test.go` — Health-Check-Integrationstest
- `test-integration.sh` — Integrations-Test-Script
- Alle Ergebnisse aus Abschnitten 1–7

### Tasks

- [ ] `database/seed.sql` komplett neuschreiben: Kassensitzung-Events einfügen (Eröffnung für 3 Tage: 20260320, 20260321, 20260322; Anfangsbestand pro KS); Subject-Format `kassensitzung-{YYYYMMDD}-tisch-{id}` statt `tisch:{id}`; Event-Typen ohne `tisch.`-Präfix; neue Spalte `kassensitzung_nr` befüllen; `tisch_session_state` statt `table_state` (mit subject als PK, tisch_id, kassensitzung_nr); `kassensitzung_state`-Einträge für die 3 KS (2 abgeschlossen, 1 offen); Tagesabschluss-Events für Tag 1 + 2
- [ ] `make test` ausführen — alle Backend-Unit-Tests müssen grün sein
- [ ] `make test-frontend` ausführen — alle Frontend-Tests müssen grün sein
- [ ] `make lint` ausführen — keine Lint-Fehler
- [ ] `make fmt` ausführen — Formatierung prüfen
- [ ] `make build` ausführen — Build erfolgreich
- [ ] `make check` ausführen — Gesamtprüfung ohne DB bestanden

---

## Abschnitt 9: Dokumentation aktualisieren 🔒

Kontext:

- `docs/redesign.md:1-866` — Source of Truth für das neue Design
- `docs/handbuch.md:84-113, 116-310, 477-605, 808-843` — Bounded Contexts, Core Domain, Kassenführung, Read Models
- `docs/anforderungen.md:68-526, 527-796, 797-943` — Kassenbetrieb, Reporting, Kassenführung
- `docs/language.md:35-52, 111-206, 209-286` — Abweichungen, Begriffe
- `docs/diagrams.md:63, 139, 173, 206, 344, 392, 486, 532, 678, 794` — Diagramme
- `docs/roadmap.md:52-127` — Kassenführung-Phase
- `docs/compliance.md:158-342, 343-378, 490-623` — TSE, GoBD, DSFinV-K
- `docs/tagesabschluss.md` — Tagesabschluss
- `docs/bondruck.md` — Bondruck/Relay
- `docs/adr/event-sourcing.md`, `docs/adr/cqrs.md`, `docs/adr/bondruck.md` — ADRs
- `AGENTS.md` — Agent-Instruktionen
- `README.md` — Projektbeschreibung
- `.github/instructions/backend.instructions.md`, `database.instructions.md`, `event-sourcing.instructions.md`, `frontend.instructions.md` — Coding-Konventionen

### Tasks

#### 9a: Entwickler-Handbuch (`docs/handbuch.md`)

- [x] §2 Bounded Contexts komplett neuschreiben: Drei statt vier Kontexte (Kasse, Stammdaten, Auth); Kontextübersicht-Tabelle, Context Map neu; bidirektionale Abhängigkeit entfällt; Kasse = Core Domain mit Event-Sourcing (Kassenjournal)
- [x] §3 Kassenbetrieb → §3 Kasse (Core Domain) umschreiben: Tisch-Session (Abrechnungskreis) als Aggregat statt Tisch-Aggregat; Subject-Format `kassensitzung-{YYYYMMDD}-tisch-{id}`; Event-Typen ohne `tisch.`-Präfix; Kassensitzung-Invariante ergänzen; 6 neue Kassensitzung-Events dokumentieren; Event-Replay session-scoped
- [x] §5 Kassenführung (Supporting Sub-Domain) komplett in §3 Kasse integrieren: Abrechnungskreis = Tisch-Session; Kassensitzung-Events statt Immutable Records; Kassenbestand als SQL-Aggregation; Kassensturz als Zwei-Event-Muster
- [x] §8 Read Models aktualisieren: `tisch_session_state` (PK: subject, session-scoped) + `kassensitzung_state`; Tischübersicht als JOIN-Query; neue Projektionstabellen dokumentieren

#### 9b: Anforderungen (`docs/anforderungen.md`)

- [x] §1 „Kassenbetrieb" → „Kasse" umbenennen; K-07 (Historie) → Kassenjournal-Terminologie; Subject-Format-Referenzen aktualisieren
- [x] §5 Reporting: Filterung nach `kassensitzung_nr` statt Zeitraum; R-01 Tagesabrechnung: Bezug auf Kassensitzung; R-02 DSFinV-K-Export auf Kassenjournal verweisen
- [x] §8 Kassenführung als eigenständigen Abschnitt auflösen und in Kasse-Kontext integrieren: KF-01 → Kassensitzung eröffnen, KF-02 → Anfangsbestand, KF-03 → Kassenbestand, KF-04/05/06 → Kassenbewegungen, KF-07 → Tagesabschluss, KF-08 → Kassensturz

#### 9c: Ubiquitous Language (`docs/language.md`)

- [x] Kassenbetrieb-Begriffe aktualisieren: Tisch = Stammdaten vs. Tisch-Session = Kasse-Aggregat; Subject-Format `kassensitzung-{YYYYMMDD}-tisch-{id}`; Event-Typen ohne Präfix
- [x] Kassenführung-Begriffe aktualisieren: Abrechnungskreis = pro Tisch pro Kassensitzung; Neue Begriffe: Kassenjournal, Kassensitzung, Tisch-Session, Kassensitzung-Sperre
- [x] Namenskonventionen aktualisieren: Go-Structs `domain/kasse/`, DB-Tabellen `kassenjournal`, `tisch_session_state`, `kassensitzung_state`
- [x] Abweichungen Ist/Soll aktualisieren: Neue Umbenennungen dokumentieren

#### 9d: Diagramme (`docs/diagrams.md`)

- [x] Diagramm 2 (Context Map): Vier → drei Kontexte (Kasse, Stammdaten, Auth)
- [x] Diagramm 4 (Tisch-Aggregat Zustandsdiagramm): Session-scoped, implizite Erstellung bei erster Bestellung
- [x] Diagramm 5 (Domain Events + Saldo-Fluss): Event-Typen ohne Präfix, neue KS-Events
- [x] Diagramm 6 (Bestellvorgang Sequenz): Kassensitzung-Sperre als zusätzlicher Prüfschritt
- [x] Diagramm 8 (Kassenführung Lifecycle): Komplett neuschreiben als Kassensitzung-Lifecycle
- [x] Diagramm 9 (Kassenbestand): SQL-Aggregation über Kassenjournal
- [x] Diagramm 11 (Schichtenarchitektur): Paketstruktur `domain/kasse/`, `kassenjournal_repo/`
- [x] Diagramm 12 (Event Sourcing + Projektion): Zwei Projektionen, StreamType-Routing
- [x] Diagramm 15 (API-Bereichsgliederung): Neue Kasse-Endpunkte (KS eröffnen, Kassenbewegung, etc.)
- [x] Diagramm 17 (DB-Schema ER): `kassenjournal`, `kassensitzung_state`, `tisch_session_state`

#### 9e: Roadmap (`docs/roadmap.md`)

- [x] Phase 1 Aufgaben umschreiben: „Kassenführung — DB-Schema" → „Kasse — DB-Schema (Kassenjournal + Projektionen)"; „Abrechnungskreis eröffnen" → „Kassensitzung eröffnen"; Immutable Records → Event-Sourcing im Kassenjournal; Umsetzungsplan aus redesign.md §6 integrieren

#### 9f: Compliance, Tagesabschluss, Bondruck

- [x] `docs/compliance.md`: `events`-Tabelle → `kassenjournal`; `ABRECHNUNGSKREIS` pro Tisch pro KS; Subject-Format aktualisieren; GoBD-Referenzen auf Kassenjournal
- [x] `docs/tagesabschluss.md`: Abrechnungskreis-Referenzen aktualisieren; Kassensturz als Zwei-Event-Muster; Z-Bon über Kassenjournal-Aggregation
- [x] `docs/bondruck.md`: `events` → `kassenjournal`; Event-Typ `bestellung-aufgenommen:v1` (ohne Präfix); Subject-Parsing für neues Format

#### 9g: ADRs

- [x] `docs/adr/event-sourcing.md`: Event-Typen ohne Präfix; `events` → `kassenjournal`; `table_state` → `tisch_session_state`; Subject-Format; neue KS-Events
- [x] `docs/adr/cqrs.md`: Zwei Projektionen statt einer; StreamType-Routing; session-scoped Projektion
- [x] `docs/adr/bondruck.md`: Relay-Anpassungen referenzieren (neues Subject-Format, Event-Typ)

#### 9h: AGENTS.md, README.md, Instruktionen

- [x] `AGENTS.md`: Bounded Contexts (3 statt 4); Event-Sourcing für „Kasse-Operationen" statt „Tisch-Operationen"; „Kassenjournal is immutable" statt „Events are immutable"; Bereiche-Abschnitt erweitern um Kasse-Endpunkte
- [x] `README.md`: Kassenbetrieb/Kassenführung-Referenzen vereinheitlichen; Persistenz-Beschreibung auf Kassenjournal + CRUD aktualisieren
- [x] `.github/instructions/backend.instructions.md`: Paketstruktur (`domain/kasse/`, `kassenjournal_repo/`)
- [x] `.github/instructions/database.instructions.md`: Tabellenname `kassenjournal` statt `events`; neue Projektions-Tabellen
- [x] `.github/instructions/event-sourcing.instructions.md`: Event-Typen ohne Präfix; Subject-Format; `kassenjournal` statt `events`; zwei Projektionen; StreamType-Routing; neue KS-Events
- [x] `.github/instructions/frontend.instructions.md`: Kassensitzung-UI-Konzepte (falls relevant); Verweis auf neue Admin-Kasse-Seite
