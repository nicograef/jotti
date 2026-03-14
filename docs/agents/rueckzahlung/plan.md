# Plan: K-04b Stornierung bezahlter Positionen + Auszahlung

Vollständige Implementierung laut `docs/agents/plan-rueckzahlung.md`.
Jede Phase entspricht einer Agent-Session. Marker: (kein) = frei · 🔒 = in Bearbeitung · ✅ = fertig.

Parallelisierungsmöglichkeiten:

- Phase 0 (Doku) und Phase 1 (Kommentar) sind unabhängig voneinander und von Phase 2–6
- Phase 2 muss vor Phase 3, 5 und 6 abgeschlossen sein
- Phase 3 muss vor Phase 4 abgeschlossen sein
- Phase 5 und Phase 6 können parallel zu Phase 4 laufen (sobald Phase 2 fertig)

---

## Phase 0 · Dokumentation

> Unabhängig von allen anderen Phasen. Kann jederzeit umgesetzt werden.

**Kontext:**

- `docs/agents/plan-rueckzahlung.md` §Phase 0 (vollständig lesen)
- `docs/anforderungen.md`

**Tasks:**

- [ ] `docs/anforderungen.md`: K-04a Akzeptanzkriterium „Kommentar optional (max. 100 Zeichen)" → „Kommentar **erforderlich** (mind. 3, max. 100 Zeichen)"
- [ ] `docs/anforderungen.md`: K-04b Akzeptanzkriterien ergänzen (Kommentar-Pflicht, Auszahlung als eigenständige Operation, freier Betrag ≥ 1 Cent mit UI-Vorausfüllung bei negativem Saldo, Kommentar-Pflicht für Auszahlung, `AuszahlungGeleistet`-Event im Kassenjournal, korrekter Saldo nach Auszahlung, Negativsaldo-Prominenz im UI, Reporting berücksichtigt Auszahlungen)

---

## Phase 1 · Kommentar-Pflichtfeld

> Unabhängig von Phase 2–6. Kann parallel zu Phase 0 und Phase 2 umgesetzt werden.
> Achtung: Test-Fixtures in `events_test.go` müssen mit angepasst werden.

**Kontext:**

- `docs/agents/plan-rueckzahlung.md` §Phase 1 (vollständig lesen)
- `backend/domain/table/stornierungErteiltEvent.go`
- `backend/domain/table/events_test.go`
- `frontend/src/service/table/Stornierung.ts`
- `frontend/src/service/components/table/StornierungDrawer.tsx`
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx`

**Tasks:**

- [ ] `stornierungErteiltEvent.go`: `"Kommentar": z.String().Max(100)` → `z.String().Min(3).Max(100).Required()` in `stornierungErteiltV1DataSchema`
- [ ] `events_test.go`: alle `NewStornierungErteiltEvent(...)`-Aufrufe mit leerem oder zu kurzem Kommentar auf gültigen Kommentar ändern (mind. 3 Zeichen, z. B. `"Test"`)
- [ ] `Stornierung.ts`: `kommentar`-Feld in `StornierungSchema` und `StornierungErteilenSchema` auf `z.string().min(3).max(100)` setzen
- [ ] `StornierungDrawer.tsx`: `kommentarInvalid = kommentar.trim().length < 3` berechnen; Submit-Button `disabled` wenn `loading || noPositionenSelected || kommentarInvalid`; optionalen Hinweistext unter dem Feld wenn berührt und zu kurz
- [ ] `HistorieStornierungDrawer.tsx`: gleiches Kommentar-Enforcement wie StornierungDrawer; Error-Handling auf `getActionErrorMessage` mit `byCode`-Mapping; `Pick<TischBackend, 'stornierungErteilen'>` sicherstellen
- [ ] `make test` + `make test-frontend` ausführen — muss grün sein

---

## Phase 2 · AuszahlungGeleistet Domain Event

> Muss vor Phase 3, 5 und 6 abgeschlossen sein. Kann parallel zu Phase 1 umgesetzt werden.

**Kontext:**

- `docs/agents/plan-rueckzahlung.md` §Phase 2 (vollständig lesen)
- `backend/domain/table/events.go`
- `backend/domain/table/projection.go`
- `backend/domain/table/events_test.go`
- `backend/domain/table/projection_test.go`
- Für Referenz bestehende Event-Dateien: `backend/domain/table/zahlungKassiertEvent.go` (oder aktuellen Namen prüfen)

**Tasks:**

- [ ] `backend/domain/table/auszahlungGeleistetEvent.go` (neue Datei erstellen): Domänen-Struct `Auszahlung` (ID, UserID, TischID, BetragCents, Kommentar, GeleistetAm); Event-Data-Struct `auszahlungGeleistetV1Data` mit json-Tags; Schema `BetragCents >= 1`, `Kommentar min(3) max(100) required`, `AuszahlungID` UUID; Builder `NewAuszahlungGeleistetEvent(userID, userName, tischID, betragCents, kommentar)`; `buildAuszahlungFromEvent(event)` → `Auszahlung`
- [ ] `events.go`: Neue Konstante `EventTypeAuszahlungGeleistetV1 EventType = "tisch.auszahlung-geleistet:v1"`; neue Art `HistorieEintragAuszahlung HistorieEintragArt = "auszahlung"`; neues Feld `Auszahlung *Auszahlung` in `HistorieEintrag`; Case in `GetHistoryFromEvents` für `buildAuszahlungFromEvent`; `continue`-Case in `ComputeNichtStorniertePositionen`
- [ ] `projection.go`: neuen Case in `ApplyEvent` für `EventTypeAuszahlungGeleistetV1` → `state.SaldoCents += data.BetragCents`; `continue`-Case in `ComputeNichtStorniertePositionen`
- [ ] `events_test.go` + `projection_test.go`: Tests für neues Event ergänzen
- [ ] `make test` ausführen — muss grün sein

---

## Phase 3 · Backend Command + API

> Benötigt Phase 2.

**Kontext:**

- `docs/agents/plan-rueckzahlung.md` §Phase 3 (vollständig lesen)
- `backend/api/table/application/command.go`
- `backend/api/table/http/command_handler.go`
- `backend/api/serviceleitung.go`
- Referenz: bestehende Command-Methoden (z. B. `StornierungErteilen`) für Muster

**Tasks:**

- [ ] `command.go`: neue Methode `AuszahlungLeisten` — Tisch-Aktiv-Check via `loadTischState`, kein Saldo-Precondition-Check, `NewAuszahlungGeleistetEvent` aufrufen, `writeEvent` mit OCC
- [ ] `command_handler.go`: Interface `command` um `AuszahlungLeisten` erweitern; Request-Struct `{ tischId, betragCents, kommentar }`; `AuszahlungLeistenHandler()` implementieren mit Validierung und Fehler-Handling
- [ ] `serviceleitung.go`: neue Route `POST /serviceleitung/auszahlung-leisten` mit `middleware.Chain(commandHandler.AuszahlungLeistenHandler(), ...)`
- [ ] `make build` ausführen — Backend muss fehlerfrei kompilieren

---

## Phase 4 · Frontend Auszahlung UI

> Benötigt Phase 3.

**Kontext:**

- `docs/agents/plan-rueckzahlung.md` §Phase 4 (vollständig lesen)
- `frontend/src/service/table/TischBackend.ts`
- `frontend/src/service/components/table/Zahlung.tsx`
- `frontend/src/service/TablePage.tsx`
- Für Struktur-Referenz: `frontend/src/service/components/table/StornierungDrawer.tsx`
- Für Tischkarten-Komponente: Dateinamen in `frontend/src/service/components/` prüfen

**Tasks:**

- [ ] `frontend/src/service/table/Auszahlung.ts` (neue Datei erstellen): `AuszahlungSchema` (id, userId, tischId, betragCents ≥ 1, kommentar min(3), geleistetAm); `AuszahlungLeistenSchema` (tischId, betragCents ≥ 1, kommentar min(3) max(100))
- [ ] `TischBackend.ts`: neue Methode `auszahlungLeisten(cmd: AuszahlungLeisten)` → `POST /serviceleitung/auszahlung-leisten`
- [ ] `frontend/src/service/components/table/AuszahlungDrawer.tsx` (neue Datei erstellen): freies Betrags-Eingabefeld (Euro, intern Cent), vorausgefüllt mit `Math.abs(saldoCents) / 100` wenn Saldo negativ; Kommentar-Pflichtfeld (min 3, max 100); Submit disabled bis `betragCents >= 1 && kommentar.trim().length >= 3`; Trigger-Button „Auszahlung" (`variant="outline"`); Submit-Button „Auszahlung leisten"; Error-Handling via `getActionErrorMessage`
- [ ] `Zahlung.tsx`: `saldoCents`-Prop hinzufügen; `AuszahlungDrawer` für `AuthSingleton.canCancel` immer anzeigen; Banner bei `saldoCents < 0` (rote Farbe, Text „Auszahlung ausstehend: X €"); `onAuszahlungGeleistet` → `reloadState()` + `reloadHistorie()`
- [ ] `TablePage.tsx`: `saldoCents` aus `state` an `Zahlung` weitergeben; negativen Saldo im Header mit roter Farbe + Badge „Auszahlung ausstehend" anzeigen; `onAuszahlungGeleistet` verdrahten
- [ ] Tischkarten-Komponente: negativen Saldo mit roter Farbe/Badge auf der Übersichtskarte anzeigen
- [ ] `make test-frontend` + `make lint` ausführen — muss grün sein

---

## Phase 5 · TischHistorie — Auszahlung anzeigen

> Benötigt Phase 2. Kann parallel zu Phase 4 umgesetzt werden.

**Kontext:**

- `docs/agents/plan-rueckzahlung.md` §Phase 5 (vollständig lesen)
- `backend/api/table/http/query_handler.go`
- `frontend/src/service/components/table/TischHistorie.tsx`
- `frontend/src/service/table/Auszahlung.ts` (aus Phase 4, ggf. zuerst prüfen ob bereits erstellt)

**Tasks:**

- [ ] `query_handler.go`: DTO `auszahlung` (id, userId, tischId, betragCents, kommentar, geleistetAm); Mapper `toAuszahlung(a t.Auszahlung) auszahlung`; Case `t.HistorieEintragAuszahlung` in `toHistorie()` ergänzen
- [ ] `TischHistorie.tsx`: `Auszahlung`-Typ zur Union in `historie`-Prop ergänzen; Erkennung via Diskriminator `art === 'auszahlung'`; `HistoryItem` mit Titel „Auszahlung −X €", Datum, Kommentar; Details-Drawer: Betrag + Kommentar (keine Positionsliste)
- [ ] `make test-frontend` + `make lint` ausführen — muss grün sein

---

## Phase 6 · Reporting

> Benötigt Phase 2. Kann parallel zu Phase 4 und 5 umgesetzt werden. `make sqlc` nach SQL-Änderungen ausführen.

**Kontext:**

- `docs/agents/plan-rueckzahlung.md` §Phase 6 (vollständig lesen)
- `backend/sqlc/queries/reporting.sql`
- `backend/domain/reporting/reporting.go`
- `backend/repository/reporting_repo/repo.go`
- `backend/api/reporting/http/query_handler.go`
- `frontend/src/admin/reporting/types.ts`
- `frontend/src/admin/reporting/ReportingResults.tsx`

**Tasks:**

- [ ] `reporting.sql`: `GetReportingStats` um `gesamt_auszahlungen_cents` ergänzen (Kassierungen − Auszahlungen), Event-Typ `'tisch.auszahlung-geleistet:v1'` einbeziehen; `GetUmsatzProServicekraft` + `GetUmsatzProTisch` um `auszahlungen_cents` ergänzen (CASE-WHEN für neuen Event-Typ); neue Query `GetAusstehendAuszahlungen`: `SUM(ABS(saldo_cents)) WHERE saldo_cents < 0` aus `table_state`
- [ ] `make sqlc` ausführen
- [ ] `reporting.go`: `Summary` um `GesamtAuszahlungenCents` und `AusstehendAuszahlungenCents` ergänzen; `UmsatzServicekraft` und `UmsatzTisch` um `AuszahlungenCents` ergänzen
- [ ] `repo.go`: neue Parallel-Query für `GetAusstehendAuszahlungen` einbinden; alle neuen Felder mappen
- [ ] `query_handler.go` (Reporting HTTP): Response-DTOs um `gesamtAuszahlungenCents`, `ausstehendAuszahlungenCents`, `auszahlungenCents` (pro Servicekraft/Tisch) ergänzen
- [ ] `types.ts`: Zod-Schemas um neue Felder ergänzen
- [ ] `ReportingResults.tsx`: Limitation-Tooltips bei Servicekräfte- und Tische-Tab entfernen; neue Summary-Cards „Auszahlungen" + „Ausstehende Auszahlungen"; `auszahlungenCents` als Sub-Info in Servicekräfte- und Tische-Tab
- [ ] `make test` + `make test-frontend` + `make lint` ausführen — muss grün sein
