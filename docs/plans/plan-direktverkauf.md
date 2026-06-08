# Plan: Direktverkauf

> Quell-PRD: `docs/prds/prd-direktverkauf.md`

## Ziel

Den **Direktverkauf** (Barverkauf an der Theke: bestellen + zahlen + ausgeben in
einem Schritt) als eigenes, schlankes Event-Aggregat umsetzen — strikt getrennt vom
bedienten Tisch, **ohne** eigene Stammdaten-Entität und **ohne** Projektionstabelle.
Jeder Verkauf ist ein eigener Event-Stream (`kassensitzung-{nr}/direktverkauf-{uuid}`)
im append-only Kassenjournal. Verkauf und positionsgenaue Stornierung sind sofort
bargeldwirksam und fließen vollständig in Kassenbestand, Reporting und Tagesabschluss.
Bon-/Belegdruck nutzt die bereits vorhandene Bondruck-Infrastruktur.

## Architekturentscheidungen

Durchgängige, phasenübergreifende Festlegungen:

- **Stream-Typ**: neuer `StreamType` `direktverkauf` (`backend/domain/kasse/stream_type.go`).
  Das Repository-Routing schreibt für diesen Typ **nur** ins Kassenjournal und
  aktualisiert **keine** Projektion.
- **Subject**: `kassensitzung-{nr}/direktverkauf-{uuid}` — ein eigener Stream pro
  Verkauf. `direktverkauf-getaetigt:v1` ist `version = 1`; Teilstornos sind
  `version = 2, 3, …` im selben Stream (OCC über `UNIQUE(subject, version)`).
- **Events** (immutable, stabile JSON-Keys):
  - `direktverkauf-getaetigt:v1` — `verkaufId`, `positionen[]`, `gesamtbetragCents`, `kommentar?`
  - `direktverkauf-storniert:v1` — `stornierungId`, `verkaufId`, `positionen[]` (PositionRef), `gesamtStornierungCents`, `kommentar`
  - `Position` (Fat Event) wird unverändert wiederverwendet.
- **Routen (POST-only)**:
  - `POST /service/direktverkauf-taetigen` — Rollen admin/serviceleitung/service
  - `POST /serviceleitung/direktverkauf-stornieren` — Rollen admin/serviceleitung
  - `POST /service/get-direktverkauf-historie` — Rollen admin/serviceleitung/service
  - `POST /service/beleg-drucken` — erweitert um optionale Verkauf-Referenz (Phase 5)
  - `POST /admin/get-bondruck-einstellungen` / `…/update-bondruck-einstellungen` — erweitert (Phase 4)
- **Schema**:
  - **Keine** neue Tabelle und **keine** Projektion für den Kern. Direktverkauf lebt
    vollständig im `kassenjournal`.
  - Zwei neue `IMMUTABLE`-SQL-Funktionen: `kj_extract_direktverkauf_cents`
    (`direktverkauf-getaetigt:v1` → `gesamtbetragCents`) und
    `kj_extract_direktverkauf_storno_cents` (`direktverkauf-storniert:v1` →
    `gesamtStornierungCents`).
  - Phase 4: `bondruck_einstellungen` wird um `direktverkauf_modus`
    (`kein_bon | abholbon | an_stationen`) und `abholbon_drucker_ip` erweitert.
  - Schemaänderungen direkt in `database/migrations/01_initial.up.sql` (aktive
    Entwicklungsphase, Breaking Changes erlaubt). Danach `make sqlc`.
- **Storno-Validierung ohne Projektion**: reine Domänenfunktion
  `ComputeNichtStornierteVerkaufPositionen(events) ([]Position, error)` per
  On-Demand-Replay des einzelnen Verkauf-Streams.
- **Frontend**: neue `DirektverkaufBackend`-Klasse auf Basis des `BackendClient`.
  Einstieg über eine **Karte/Button auf der Tischauswahl-Seite** (`/service/tische`);
  neue Route `service/direktverkauf` unter `ServiceLayout`, abgesichert per
  `ServiceGuard`. Kein Service-Sidebar (existiert nicht).
- **Validierung beidseitig**: zog (Backend) + Zod (Frontend) für jeden Request.

## Inventar

Bestehende Dateien, Muster und Verträge als Prior Art:

**Domäne (`backend/domain/kasse/`)**

- `stream_type.go` — `StreamType` mit aktuell `StreamTypeKassensitzung`,
  `StreamTypeTischSession`.
- `subject.go:8-18` — `KassensitzungSubject`, `TischSessionSubject`; `:20-51` Parser.
- `bestellung.go:7-15` — `Position`; `:18-28` `positionEventData` (JSON-Keys);
  `:50-63` `positionSchema`; `:66-69` `PositionRef`.
- `tisch_session_events.go:8-13` — Event-Typ-Konstanten; `:110-215` Event-Creation
  mit zog-Validierung (Muster für `NewBestellungAufgenommenEvent` etc.).
- `tisch_session.go:68-101` — `ComputeNichtStorniertePositionen`; `:103-132`
  `accumulatePositionen` / `reduceByPosition`.
- `historie.go:26-59` — `GetHistoryFromEvents`.
- Tests: `tisch_session_test.go`, `historie_test.go`, `subject_test.go`.

**Application / API**

- `api/table/application/command.go:325-410` — `BestellungAufnehmen` (offene KS prüfen,
  Positionen via Batch-Fetch anreichern, Event schreiben); `:108-116`
  `GetOffeneKassensitzungOderFehler`; `:138-165` `writeEvent` (OCC via `GetMaxVersion`);
  `:236-246` `computeNichtStorniertePositionen`; `:471-510` `StornierungErteilen`;
  `:207-228` `validatePositionRefs`; `:420-440` `resolvePositions`; `:140-149`
  `enqueueArbeitsbonDruckauftraege`.
- `api/table/http/command_handler.go:277-313` — Bestellung-Handler + zog-Schema +
  Error-Mapping; `:409-437` Storno-Handler; `:376` Kassenbeleg-Handler.
- `api/service.go:20-50` — Service-Wiring; `api/serviceleitung.go:1-20` —
  Serviceleitung-Wiring.
- `api/middleware/middleware.go:150-191` — `NewJwtMiddleware` (Rollenprüfung).
- `app/app.go:30-45` — Routen-/Rollenbindung (`/service/*`, `/serviceleitung/*`).
- `api/helper/http.go` — `ReadAndValidateBody`, `SendEmptyResponse`, `MapError`.

**Repository / DB**

- `repository/kassenjournal_repo/repo.go:32-77` — `WriteEvent(ctx, e, streamType, ksNr)`;
  `:52-63` Routing-`switch`; `:79-101` `handleKassensitzungEvent`; `:104-150`
  `handleTischSessionEvent`; `:204-224` `ReadEventsBySubject`; `:229-237`
  `GetMaxVersion`; `:341-347` `GetKassenbestand`.
- `database/migrations/01_initial.up.sql:124-137` — `kassenjournal` (mit
  `UNIQUE(subject, version)`); `:156-169` Immutabilitäts-Trigger; `:263-302`
  `kj_extract_*`-Funktionen; `:218` `druckstationen`; `:247` `druckauftraege`;
  `:366-376` `bondruck_einstellungen`.
- `sqlc/queries/kassensitzungen.sql:16-30` — `GetKassenbestand`;
  `sqlc/queries/reporting.sql:14-31` — `GetReportingStats`.
- `domain/reporting/reporting.go:23-46` — `Summary`; `api/reporting/http/query_handler.go:59-63`
  Reporting-Response-DTO.
- `api/kasse/application/command.go:200-270` — Tagesabschluss-Sperre (Kassensturz +
  „alle Tisch-Sessions Saldo 0" via `GetTischSessionsByKassensitzungNr`).

**Bondruck (bereits umgesetzt)**

- `api/bondruck/application/arbeitsbon_policy.go:36` — `CreateArbeitsbonAuftraegeFromEvent`
  (reagiert auf `bestellung-aufgenommen:v1`, gruppiert nach Kategorie).
- `api/bondruck/application/escpos/formatter.go:28` `FormatPositionBon`; `:83`
  `FormatSammelBon`; `:144` `FormatKassenbeleg`.
- `api/table/application/kassenbeleg_command.go:52` — `KassenbelegDrucken` (On-Demand,
  per `zahlungID`).
- `repository/druckauftrag_repo/repo.go` — `EnqueueDruckauftraege`,
  `GetOffeneDruckauftraege`; `repository/druckstation_repo/repo.go` —
  `GetKonfigurierteDruckstationen`.
- `sqlc/queries/bondruck_einstellungen.sql:1-10` — Get/Upsert; `api/admin.go:123,128`
  — Admin-Endpunkte.

**Frontend**

- `src/lib/Backend.ts:46` — `BackendClient`-Interface (`post<T>`); `:57` Implementierung.
- `src/lib/AuthBackend.ts:34-57`, `src/service/product/ProduktBackend.ts:7-17`,
  `src/service/table/TischBackend.ts:41-58` — Backend-Klassen-Muster.
- `src/service/ServiceLayout.tsx` — Header (keine Sidebar);
  `src/service/TableSelectionPage.tsx` — Einstiegsseite `/service/tische`.
- `src/service/TablePage.tsx:45-70` — Tabs Bestellen/Ausgabe/Zahlung;
  `src/service/components/table/ProductList.tsx:60-88` — Produktauswahl nach Kategorie;
  `src/service/components/table/drawerUtils.ts:21-40` — `calculateZahlungsbetraege`
  (Rückgeld K-10, rein clientseitig).
- `src/routes.ts:24-40` — Guards (`ServiceGuard`); `:135-157` Service-Routen;
  `src/routes.test.ts` — Guard-Tests.
- `src/service/table/Bestellung.ts:1-40` — `PositionSchema` (Zod);
  `src/service/table/Tisch.ts:1-28`.
- `src/admin/reporting/ReportingResults.tsx` — `SummaryCard` im „Übersicht"-Tab;
  `src/admin/reporting/ReportingBackend.ts`, `types.ts`.
- `src/admin/settings/EinstellungenPage.tsx` — Bondruck-Sektion
  (`useBondruckEinstellungen`); `src/admin/settings/DruckstationConfigPage.tsx`;
  `src/lib/EinstellungenBackend.ts` — `BondruckEinstellungen`-Typ.

**Docs**

- `docs/language.md` — Fachbegriffe / Namenskonventionen pro Schicht.
- `docs/handbuch.md` §3.3 (Subject/OCC), §4.6 (Bondruck, ~Zeile 657/694).
- `docs/anforderungen.md` — aktuell höchste Nummer K-22.
- `docs/compliance.md`, `docs/produktbeschreibung.md`.

## Geklärte Entscheidungen

- **Umfang**: voller PRD-Umfang in einem Plan; Bondruck-Integration als letzte
  Phasen (4–5), damit sie aufschiebbar bleibt.
- **Navigation**: Einstieg über eine prominente Karte/Button auf der
  Tischauswahl-Seite (`/service/tische`) — kein Service-Sidebar.
- **Tests**: jede Phase liefert ihre eigenen Tests (Backend immer, Frontend wo
  zutreffend; zog + Zod).
- **Doku**: Aktualisierungen direkt in die jeweils passende Phase eingeflochten.

## Offene Punkte / Risiken

- **`make sqlc` nach Query-Änderungen** (Phasen 3–4) ist zwingend; generierter Code in
  `sqlc/dbgen/` wird nie von Hand editiert.
- **DSFinV-K-/TSE-Feinheiten** (`ABRECHNUNGSKREIS`) werden erst in der TSE-Phase final
  festgelegt; für Revisionssicherheit genügt das append-only Event bereits jetzt — kein
  Blocker für diesen Plan.
- **Dev-DB-Reset** nach Migrationsänderung: `make down && make dev`.

---

## Phase 1: Direktverkauf tätigen (Ende-zu-Ende, Happy Path)

**User Stories**: 1, 2, 3, 4, 5, 6, 7, 8, 9, 24, 25

### Kontext

- `backend/domain/kasse/stream_type.go` — Stream-Typen erweitern.
- `backend/domain/kasse/subject.go:8-18` — Subject-Builder als Muster.
- `backend/domain/kasse/bestellung.go:7-15,50-63` — `Position` + Schema wiederverwenden.
- `backend/domain/kasse/tisch_session_events.go:110-145` — Event-Creation-Muster (zog,
  UUID-/Positions-ID-Erzeugung, Summenbildung).
- `backend/api/table/application/command.go:325-410,108-116,138-165` —
  Bestellfluss (offene KS, Batch-Fetch, OCC-Write) als Vorlage.
- `backend/api/table/http/command_handler.go:277-313` — Handler/DTO/zog/Error-Mapping.
- `backend/api/service.go:20-50`, `backend/app/app.go:30-45` — Wiring + Rollenbindung.
- `backend/repository/kassenjournal_repo/repo.go:32-77,52-63` — `WriteEvent` + Routing.
- `frontend/src/lib/Backend.ts:46`, `frontend/src/service/product/ProduktBackend.ts:7-17`
  — Backend-Klassen-Muster.
- `frontend/src/service/TableSelectionPage.tsx`, `frontend/src/routes.ts:135-157` —
  Einstieg + Routen.
- `frontend/src/service/components/table/ProductList.tsx:60-88`,
  `drawerUtils.ts:21-40` — Produktauswahl + Rückgeld zur Wiederverwendung.

### Was gebaut wird

Ein Vereinsmitglied (admin/serviceleitung/service) öffnet über eine Karte auf
`/service/tische` die neue Direktverkauf-Seite, stellt aus den aktiven Produkten eine
Bestellung mit Mengen zusammen, sieht die laufende Summe (und optional das Rückgeld),
und schließt den Verkauf mit **einer** Bestätigung ab. Das Backend prüft die offene
Kassensitzung, reichert die Positionen zu Fat-Event-Positionen an und schreibt **genau
ein** `direktverkauf-getaetigt:v1`-Event in einen neuen, verkaufseigenen Stream
(`kassensitzung-{nr}/direktverkauf-{uuid}`, `version = 1`). Das Repository-Routing
schreibt für den neuen Stream-Typ ausschließlich ins Kassenjournal — keine Projektion.
Nach Erfolg leert sich die Eingabe für den nächsten Gast. Der Tisch-Ablauf bleibt
vollständig getrennt (eigene Events, eigene Endpunkte).

### Akzeptanzkriterien

- [x] Neuer `StreamType` `direktverkauf`; Repository-Routing schreibt dafür **nur** ins
      Kassenjournal und legt **keine** `tisch_sessions`-Zeile an.
- [x] `DirektverkaufSubject(zNr, verkaufUUID)` erzeugt/parst
      `kassensitzung-{nr}/direktverkauf-{uuid}`.
- [x] `direktverkauf-getaetigt:v1` wird per zog validiert; `gesamtbetragCents` ist die
      konsistente Summe der Positionen; Positions-IDs (UUID) werden serverseitig erzeugt.
- [x] `POST /service/direktverkauf-taetigen` (POST-only) schreibt **genau ein** Event;
      ohne offene Kassensitzung → HTTP 409.
- [x] Endpunkt ist für admin/serviceleitung/service erreichbar; Request-DTO:
      `positionen[]` (`produktId`, `varianteId`, `menge`), `kommentar?` (max. 100) —
      **kein** `verkaufsstelleId`.
- [x] Frontend: `DirektverkaufBackend.direktverkaufTaetigen` über `BackendClient`
      (kein direktes `fetch()`); Route `service/direktverkauf` unter `ServiceGuard`;
      Einstiegskarte auf `/service/tische`.
- [x] Verkaufen-Seite rendert die kombinierte Oberfläche (Produktauswahl +
      Abschluss-Button + optionale Rückgeldanzeige) statt der Tabs Bestellen/Bezahlen;
      nach Erfolg ist die Eingabe geleert.
- [x] Tests: Domäne (ein Event, konsistente Summe, Positions-IDs), Command (keine
      offene KS → Fehler; Happy Path ein Event), HTTP, Frontend (genau ein Backend-Aufruf,
      Reset, kombinierte Oberfläche).
- [x] Doku: `docs/language.md` (Direktverkauf, Verkauf; keine `Verkaufsstelle`-Entität),
      `docs/anforderungen.md` (neu **K-24 · Direktverkauf**),
      `docs/handbuch.md` (Stream-Typ, Subject, Event), `docs/produktbeschreibung.md`
      (Direktverkauf personalbedient enthalten; SB-Kiosk weiterhin ausgeschlossen).

---

## Phase 2: Historie + positionsgenaue Stornierung (Ende-zu-Ende)

**User Stories**: 16, 17, 18, 19, 26

### Kontext

- `backend/domain/kasse/tisch_session.go:68-132` — `ComputeNichtStorniertePositionen`
  - Helfer als strukturelle Vorlage (auf den einzelnen Verkauf-Stream angewandt).
- `backend/domain/kasse/historie.go:26-59` — `GetHistoryFromEvents` als Vorlage.
- `backend/api/table/application/command.go:471-510,236-246,207-228,420-440` —
  Storno-Fluss (Replay laden, validieren, OCC-Write).
- `backend/api/serviceleitung.go:1-20`, `backend/app/app.go:30-45` — Wiring + Rollen
  (`/serviceleitung/*` = admin/serviceleitung).
- `backend/api/table/http/command_handler.go:409-437` — Storno-Handler/DTO/zog.
- `frontend/src/service/TablePage.tsx`, Historie-Komponenten als Muster.

### Was gebaut wird

Serviceleitung/Admin sehen die kompakte Direktverkauf-Historie der offenen
Kassensitzung (eine Zeile pro Verkauf) und können einen Verkauf **positionsgenau**
stornieren. Die Stornierung lädt den einzelnen Verkauf-Stream per `ReadEventsBySubject`,
validiert die angeforderten Positionen gegen `ComputeNichtStornierteVerkaufPositionen`
(nur noch nicht stornierte Positionen, höchstens die ursprüngliche Menge) und schreibt
`direktverkauf-storniert:v1` mit `version = maxVersion + 1` (OCC) in denselben Stream.
Mehrere Teilstornos pro Verkauf sind möglich. Das zurückgegebene Bargeld wird durch die
Stornierung selbst kassenwirksam — **keine** separate Auszahlungsbuchung. Die Rolle
`service` kann keine Stornierungen auslösen.

### Akzeptanzkriterien

- [x] `direktverkauf-storniert:v1` (zog-validiert) mit `verkaufId`,
      `positionen[]` (PositionRef: `positionId`, `menge`), `gesamtStornierungCents`,
      `kommentar` (Pflicht, min. 3, max. 100).
- [x] `ComputeNichtStornierteVerkaufPositionen(events)` liefert nach mehreren
      Teilstornos korrekt die Restmengen; Vollstorno → leere Restmenge.
- [x] `POST /serviceleitung/direktverkauf-stornieren` schreibt das Storno-Event mit
      `version = maxVersion + 1`; Storno über verfügbare Menge hinaus → Fehler;
      nicht-existenter Verkauf → Fehler; OCC-Konflikt → HTTP 409; Rolle `service`
      → abgewiesen.
- [x] `POST /service/get-direktverkauf-historie` liefert die kompakte Historie der
      offenen KS (eine Zeile pro Verkauf) über ein Response-DTO der HTTP-Schicht.
- [x] Frontend: Historie kompakt mit positionsgenauer Storno-Aktion, sichtbar nur für
      serviceleitung/admin (`DirektverkaufBackend.direktverkaufStornieren`,
      `…getDirektverkaufHistorie`).
- [x] Tests: Domäne (Teil-/Vollstorno-Replay), Command (Über-Storno, nicht-existent,
      OCC-409, Rolle service abgewiesen), HTTP, Frontend (Historie + Storno).
- [x] Doku: `docs/language.md` (Direktverkauf-Stornierung), `docs/handbuch.md`
      (Storno-Replay, Kassenwirksamkeit ohne separate Auszahlung).

---

## Phase 3: Kassenwirksamkeit + Reporting-Kennzahl

**User Stories**: 20, 21, 22, 23, 27

### Kontext

- `database/migrations/01_initial.up.sql:263-302` — bestehende `kj_extract_*`-Funktionen
  als Muster.
- `backend/sqlc/queries/kassensitzungen.sql:16-30` — `GetKassenbestand`;
  `backend/sqlc/queries/reporting.sql:14-31` — `GetReportingStats`.
- `backend/domain/reporting/reporting.go:23-46`,
  `backend/api/reporting/http/query_handler.go:59-63` — Summary + Response-DTO.
- `backend/api/kasse/application/command.go:200-270` — Tagesabschluss-Sperre
  (nur Tisch-Saldo; Direktverkauf hat keine Projektion/keinen Saldo).
- `frontend/src/admin/reporting/ReportingResults.tsx` — `SummaryCard` im
  „Übersicht"-Tab; `ReportingBackend.ts`, `types.ts`.

### Was gebaut wird

Direktverkäufe und ihre Stornierungen fließen vollständig in die Kassenführung:
`GetKassenbestand` rechnet `+ direktverkauf-getaetigt` und `− direktverkauf-storniert`,
sodass der Soll-Bestand (und damit Kassensturz/Z-Bon) korrekt ist. `GetReportingStats`
erhöht den Gesamtumsatz um Verkäufe und mindert ihn um Stornos und liefert zusätzlich
**eine** aggregierte Kennzahl (Anzahl Direktverkäufe + Direktverkauf-Umsatz) — ohne
Gruppierung pro Theke. Im Admin-Dashboard erscheint diese Kennzahl als eigene
`SummaryCard`. Der Tagesabschluss bleibt unberührt: Direktverkäufe blockieren ihn nie
(keine offenen Salden).

### Akzeptanzkriterien

- [ ] Zwei neue `IMMUTABLE`-SQL-Funktionen `kj_extract_direktverkauf_cents` und
      `kj_extract_direktverkauf_storno_cents` in der Migration; `make sqlc` ausgeführt.
- [ ] `GetKassenbestand` enthält `+ direktverkauf-getaetigt`,
      `− direktverkauf-storniert`.
- [ ] `GetReportingStats`: Gesamtumsatz +Verkauf / −Storno; zusätzlich aggregierte
      Kennzahl (Anzahl + Umsatz Direktverkauf); Response-DTO erweitert.
- [ ] Admin-Dashboard zeigt die Direktverkauf-Kennzahl als eigene `SummaryCard`.
- [ ] Tagesabschluss wird durch Direktverkäufe nie blockiert (Test bestätigt).
- [ ] Tests: Integration (Kassenbestand steigt bei Verkauf, sinkt bei Storno; Kennzahl
      zählt korrekt; Reporting-Stats), Reporting-Query/HTTP, Frontend (`SummaryCard`).
- [ ] Doku: `docs/compliance.md` (einzelner Geschäftsvorfall, Kassenwirksamkeit von
      Verkauf und Storno), `docs/handbuch.md` (Kassenbestand-/Reporting-Erweiterung),
      `docs/anforderungen.md` (K-24-Akzeptanzkriterien).

---

## Phase 4: Bondruck — `direktverkauf_modus` + Abholbon + Stations-Routing

**User Stories**: 10, 11, 12, 13, 15

### Kontext

- `database/migrations/01_initial.up.sql:366-376` — `bondruck_einstellungen`-Singleton.
- `backend/sqlc/queries/bondruck_einstellungen.sql:1-10`,
  `backend/api/admin.go:123,128` — Get/Update-Endpunkte.
- `backend/api/bondruck/application/arbeitsbon_policy.go:36` —
  `CreateArbeitsbonAuftraegeFromEvent` (erweitern auf `direktverkauf-getaetigt:v1`).
- `backend/api/bondruck/application/escpos/formatter.go:28,83` — Formatter-Muster für
  den neuen Abholbon-Formatter.
- `backend/api/table/application/command.go:140-149` — Arbeitsbon-Wiring im Command.
- `backend/repository/druckstation_repo/repo.go`,
  `backend/repository/druckauftrag_repo/repo.go` — Stationen + Outbox.
- `frontend/src/admin/settings/EinstellungenPage.tsx`,
  `frontend/src/lib/EinstellungenBackend.ts` — Admin-Konfiguration.

### Was gebaut wird

Der Betreiber konfiguriert in den Admin-Einstellungen den `direktverkauf_modus`
(`kein_bon | abholbon | an_stationen`) sowie die Abholbon-Drucker-IP. Beim
`direktverkauf-getaetigt:v1`-Event reiht die erweiterte Arbeitsbon-Policy die passenden
Druckaufträge in die bestehende Outbox: `an_stationen` → Positionen nach Kategorie an
die Druckstationen (identische Logik wie beim Tisch), `abholbon` → **ein** kombinierter
Abholbon (festes Label „Direktverkauf", keine Preise) an die Abholbon-IP, `kein_bon` →
keine Outbox-Zeile. Die Policy wird im `DirektverkaufTaetigen`-Command verdrahtet.

### Akzeptanzkriterien

- [ ] `bondruck_einstellungen` erweitert um `direktverkauf_modus` (Enum) und
      `abholbon_drucker_ip` (IPv4, leer = nicht konfiguriert); `make sqlc` ausgeführt.
- [ ] Get/Update-Endpunkte + Admin-UI um `direktverkauf_modus` und Abholbon-IP erweitert;
      beidseitig validiert (Enum + IPv4).
- [ ] Arbeitsbon-Policy reagiert auf `direktverkauf-getaetigt:v1`: `an_stationen` →
      Station-Bons nach Kategorie; `abholbon` → genau **ein** Abholbon mit festem Label
      „Direktverkauf" ohne Preise; `kein_bon` → keine Outbox-Zeile.
- [ ] Policy ist im `DirektverkaufTaetigen`-Command verdrahtet (nach erfolgreichem
      Event-Write).
- [ ] Tests: Policy je `direktverkauf_modus` (Station-Bons, genau ein Abholbon, keine
      Outbox-Zeile), Abholbon-Formatter, Einstellungen-Validierung (zog + Zod).
- [ ] Doku: `docs/handbuch.md` §4.6 (`direktverkauf_modus`, Abholbon),
      `docs/language.md` (Abholbon).

---

## Phase 5: Direktverkauf-Kassenbeleg (fiskalischer Beleg auf Anforderung)

**User Stories**: 14

### Kontext

- `backend/api/table/application/kassenbeleg_command.go:52` — `KassenbelegDrucken`
  (um eine Verkauf-Referenz erweitern).
- `backend/api/bondruck/application/escpos/formatter.go:144` — `FormatKassenbeleg`
  (Datenquelle Verkauf statt Tischzahlung).
- `backend/api/table/http/command_handler.go:376` — `/service/beleg-drucken`-Handler/DTO.
- `frontend/src/service/` — Verkaufen-/Historie-Komponenten für den Auslöser.

### Was gebaut wird

Ein Gast kann auf Anforderung **zusätzlich** zum Abholbon einen fiskalischen
Kassenbeleg für einen Direktverkauf erhalten. Der bestehende On-Demand-Kassenbeleg-Command
akzeptiert dafür eine Verkauf-Referenz (`verkaufId`); derselbe Formatter und dieselbe
Outbox erzeugen genau einen Druckauftrag, nur die Datenquelle ist der Verkauf statt
einer Tischzahlung. Fehlt die Kassenbeleg-Drucker-IP, schlägt der Aufruf mit klarer
Fehlermeldung fehl (bestehendes Verhalten).

### Akzeptanzkriterien

- [ ] `KassenbelegDrucken` akzeptiert eine Verkauf-Referenz und lädt das
      `direktverkauf-getaetigt:v1`-Event als Datenquelle.
- [ ] `POST /service/beleg-drucken` um die Verkauf-Referenz erweitert (DTO + zog),
      unabhängig vom Abholbon.
- [ ] Genau ein `bon_art = 'kassenbeleg'`-Druckauftrag pro Anforderung; fehlende
      Drucker-IP → klare Fehlermeldung.
- [ ] Frontend: Auslöser „Kassenbeleg drucken" auf der Verkaufen-Erfolgsansicht und/oder
      in der Historie (über `DirektverkaufBackend`).
- [ ] Tests: Command (genau ein Auftrag für einen echten Verkauf; Drucker-nicht-konfiguriert
      → Fehler), HTTP, Frontend-Auslöser.
- [ ] Doku: `docs/handbuch.md` (Direktverkauf-Kassenbeleg), `docs/compliance.md`
      (Storno als eigener Geschäftsvorfall, 1:1-TSE-Mapping-Ausblick).
