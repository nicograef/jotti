# Plan: K-04b Stornierung bezahlter Positionen + Auszahlung

## Kontext und Designentscheidungen

### Ausgangslage

- **K-04a** (Stornierung unbezahlter Positionen) ist bereits umgesetzt.
- **K-04b** (Stornierung bezahlter Positionen) ist technisch im Backend fast vollständig: `StornierungErteilen` prüft nur, ob eine Position nicht bereits storniert wurde — der Bezahlstatus spielt für die Stornierungsinvariante keine Rolle. Auch im Frontend berechnet `getStornierbarePositionen` im Historie-Tab die stornier­baren Positionen ohne Bezahlstatus-Filter.
- Hauptlücken: Kommentar-Pflichtfeld, Auszahlungs-Event + UI, Reporting.

### Warum ein eigenes Auszahlungs-Event?

An einem Vereinstisch sitzen über den Tag **mehrere unabhängige Gästegruppen** hintereinander. Wenn Gruppe A vollständig bezahlt hat und eine Position nachträglich storniert wird, entsteht ein negativer Saldo. Dieser muss ausgeglichen werden (Cash geht raus), bevor Gruppe B „auf Kosten" des negativen Saldos bestellt. Das `AuszahlungGeleistet`-Event ist die Quittierung dieses Vorgangs.

### Semantik der Auszahlung

`Auszahlung` ist konzeptionell eine **negative Zahlung**:

- `ZahlungKassiert`: Cash kommt rein → `saldo -= betrag`
- `AuszahlungGeleistet`: Cash geht raus → `saldo += betrag`

### Kein Saldo-Precondition-Check

Die Auszahlung erlaubt einen **freien Betrag** (≥ 1 Cent) und prüft **nicht**, ob der Saldo negativ ist. Grund: Mehrere unabhängige Gästegruppen am Tisch machen einen festen „Saldo muss negativ sein"-Check fachlich falsch. Die Servicekraft behält den Überblick — das Kommentarfeld hilft bei der Nachvollziehbarkeit.

### Kommentar-Pflichtfeld

Stornierungen und Auszahlungen erfordern ab sofort einen Kommentar (mind. 3, max. 100 Zeichen). Für Bestellungen, Kassierungen und Ausgaben bleibt der Kommentar optional.

### Auszahlung nur via serviceleitung/admin

Gleiche Rolle wie die Stornierung: `serviceleitung` und `admin`. Route unter `/serviceleitung/`.

### Einstieg K-04b im Frontend

Stornierung bezahlter Positionen erfolgt über den **Historie-Tab** (via `HistorieStornierungDrawer`) — das ist bereits möglich. Kein neuer UI-Abschnitt im Bezahlen-Tab nötig. Die Auszahlung selbst erscheint als neue Aktion im **Bezahlen-Tab** (immer sichtbar für serviceleitung/admin, nicht nur bei negativem Saldo).

### Reporting: Umsatz-Formel

```
gesamtUmsatzCents = SUM(kassierungen) − SUM(auszahlungen)
```

Neues Summary-Feld `ausstehendAuszahlungenCents` = aktueller Stand aller negativen Tischsaldi (zeitraumunabhängig, wie `offeneSaldiCents`), zeigt offene Auszahlungsschulden des Vereins.

---

## Betroffene Dateien (Übersicht)

| Phase                   | Dateien                                                                                                                                                                                                                                                                               |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 0 Doku                  | `docs/anforderungen.md`                                                                                                                                                                                                                                                               |
| 1 Kommentar-Pflichtfeld | `backend/domain/table/stornierungErteiltEvent.go`, `backend/domain/table/events_test.go`, `frontend/src/service/table/Stornierung.ts`, `frontend/src/service/components/table/StornierungDrawer.tsx`, `frontend/src/service/components/table/HistorieStornierungDrawer.tsx`           |
| 2 Domain Event          | `backend/domain/table/auszahlungGeleistetEvent.go` (neu), `backend/domain/table/events.go`, `backend/domain/table/projection.go`                                                                                                                                                      |
| 3 Backend API           | `backend/api/table/application/command.go`, `backend/api/table/http/command_handler.go`, `backend/api/serviceleitung.go`                                                                                                                                                              |
| 4 Frontend UI           | `frontend/src/service/table/Auszahlung.ts` (neu), `frontend/src/service/table/TischBackend.ts`, `frontend/src/service/components/table/AuszahlungDrawer.tsx` (neu), `frontend/src/service/components/table/Zahlung.tsx`, `frontend/src/service/TablePage.tsx`, Tischkarten-Komponente |
| 5 Historie              | `backend/api/table/http/query_handler.go`, `frontend/src/service/components/table/TischHistorie.tsx`                                                                                                                                                                                  |
| 6 Reporting             | `backend/sqlc/queries/reporting.sql`, `backend/domain/reporting/reporting.go`, `backend/repository/reporting_repo/repo.go`, `backend/api/reporting/http/query_handler.go`, `frontend/src/admin/reporting/types.ts`, `frontend/src/admin/reporting/ReportingResults.tsx`               |

---

## Phase 0 · Dokumentation

> Kann parallel zu allen anderen Phasen umgesetzt werden.

### 0.1 — `docs/anforderungen.md`

**K-04a:** Akzeptanzkriterium ändern:

```diff
- Kommentar optional (max. 100 Zeichen)
+ Kommentar **erforderlich** (mind. 3, max. 100 Zeichen)
```

**K-04b:** Akzeptanzkriterien ergänzen:

```diff
+ - Kommentar für die Stornierung **erforderlich** (mind. 3, max. 100 Zeichen)
+ - Auszahlung ist eine eigenständige, von der Stornierung unabhängige Operation
+ - Auszahlungsbetrag ist frei wählbar (≥ 1 Cent); bei negativem Tischsaldo wird der Betrag im UI vorausgefüllt
+ - Kommentar für die Auszahlung **erforderlich** (mind. 3, max. 100 Zeichen)
+ - Auszahlung wird als unveränderliches `AuszahlungGeleistet`-Event im Kassenjournal gespeichert
+ - Saldo des Tisches wird nach Auszahlung korrekt erhöht (kann positiv, null oder weiterhin negativ sein)
+ - Negativer Saldo wird in der UI prominent hervorgehoben (Tischkarte + Tisch-Detail + Bezahlen-Tab)
+ - Reporting: `GetReportingStats`, `GetUmsatzProServicekraft`, `GetUmsatzProTisch` berücksichtigen Auszahlungen korrekt
```

---

## Phase 1 · Kommentar Pflichtfeld

> Unabhängig von Phase 2–6, kann zuerst umgesetzt werden.
> **Achtung:** Test-Fixtures in `events_test.go` müssen mit angepasst werden.

### 1.1 — stornierungErteiltEvent.go

```go
// vorher
"Kommentar": z.String().Max(100),
// nachher
"Kommentar": z.String().Min(3).Max(100).Required(),
```

Auch `stornierungErteiltV1DataSchema` auf Min-3-Constraint prüfen.

### 1.2 — events_test.go

Alle Aufrufe `NewStornierungErteiltEvent(...)` mit leerem/zu-kurzem Kommentar → z. B. `"Test"`.

### 1.3 — Stornierung.ts

```typescript
// in StornierungSchema und StornierungErteilenSchema
kommentar: z.string().min(3).max(100),
```

### 1.4 — StornierungDrawer.tsx

```typescript
const kommentarInvalid = kommentar.trim().length < 3
// Submit-Button:
disabled={loading || noPositionenSelected || kommentarInvalid}
```

Optionaler Hinweistext unterhalb des Felds, wenn berührt und zu kurz.

### 1.5 — HistorieStornierungDrawer.tsx

Gleiches Kommentar-Enforcement wie 1.4. Error-Handling auf `getActionErrorMessage` mit `byCode`-Mapping upgraden. `Pick<TischBackend, 'stornierungErteilen'>` sicherstellen.

---

## Phase 2 · AuszahlungGeleistet Domain Event

> Muss vor Phase 3 und 5 abgeschlossen sein. Kann parallel zu Phase 1 umgesetzt werden.

### 2.1 — `backend/domain/table/auszahlungGeleistetEvent.go` (neue Datei)

- Domänen-Struct `Auszahlung` (ID, UserID, TischID, BetragCents, Kommentar, GeleistetAm)
- Event-Data-Struct `auszahlungGeleistetV1Data` (json-Tags für Event Store)
- Schema: `BetragCents >= 1`, `Kommentar min(3) max(100) required`, `AuszahlungID` UUID
- `NewAuszahlungGeleistetEvent(userID, userName, tischID, betragCents, kommentar)` Builder
- `buildAuszahlungFromEvent(event)` → `Auszahlung`

### 2.2 — events.go

- Neue Konstante: `EventTypeAuszahlungGeleistetV1 EventType = "tisch.auszahlung-geleistet:v1"`
- Neue Art: `HistorieEintragAuszahlung HistorieEintragArt = "auszahlung"`
- Neues Feld in `HistorieEintrag`: `Auszahlung *Auszahlung`
- Case in `GetHistoryFromEvents`: `buildAuszahlungFromEvent` aufrufen
- `continue`-Case in `ComputeNichtStorniertePositionen` für den neuen Event-Typ

### 2.3 — projection.go

In `ApplyEvent` neuen Case ergänzen:

```go
case string(EventTypeAuszahlungGeleistetV1):
    // json.Unmarshal data
    state.SaldoCents += data.BetragCents
```

`continue`-Case in `ComputeNichtStorniertePositionen` (gleich wie 2.2).

---

## Phase 3 · Backend Command + API

> Benötigt Phase 2.

### 3.1 — command.go

Neue Methode `AuszahlungLeisten`:

- Tisch-Aktiv-Check via `loadTischState`
- **Kein Saldo-Precondition-Check** (mehrere Gästegruppen)
- `NewAuszahlungGeleistetEvent` aufrufen
- `writeEvent` mit OCC

### 3.2 — command_handler.go

- Interface `command` um `AuszahlungLeisten` erweitern
- Request: `{ tischId, betragCents, kommentar }`
- `AuszahlungLeistenHandler()` implementieren

### 3.3 — serviceleitung.go

```go
mux.Handle("POST /serviceleitung/auszahlung-leisten",
    middleware.Chain(commandHandler.AuszahlungLeistenHandler(), ...))
```

---

## Phase 4 · Frontend Auszahlung UI

> Benötigt Phase 3.

### 4.1 — `frontend/src/service/table/Auszahlung.ts` (neue Datei)

- `AuszahlungSchema` (id, userId, tischId, betragCents ≥ 1, kommentar min(3), geleistetAm)
- `AuszahlungLeistenSchema` (tischId, betragCents ≥ 1, kommentar min(3) max(100))

### 4.2 — TischBackend.ts

Neue Methode `auszahlungLeisten` → `POST serviceleitung/auszahlung-leisten`.

### 4.3 — `frontend/src/service/components/table/AuszahlungDrawer.tsx` (neue Datei)

Orientiert sich strukturell an `StornierungDrawer.tsx`:

- **Freies Betrags-Eingabefeld** (Euro, intern Cent); vorausgefüllt mit `Math.abs(saldoCents) / 100` wenn Saldo negativ
- Kommentar-Pflichtfeld (min 3, max 100)
- Submit disabled bis `betragCents >= 1 && kommentar.trim().length >= 3`
- Trigger-Button: `"Auszahlung"` (semantisch von Stornierung unterscheidbar, z. B. `variant="outline"`)
- Submit-Button: `"Auszahlung leisten"` (variant `"secondary"` oder neutral)
- Error-Handling via `getActionErrorMessage`

### 4.4 — Zahlung.tsx

- `saldoCents` als Prop hinzufügen (kommt aus `state` in `TablePage`)
- `AuszahlungDrawer` für `AuthSingleton.canCancel` **immer** anzeigen
- Banner wenn `saldoCents < 0`:
  ```tsx
  <div className="rounded-md border border-destructive bg-destructive/10 p-3 text-sm text-destructive">
    Auszahlung ausstehend: {formatCents(Math.abs(saldoCents))} €
  </div>
  ```
- `onAuszahlungGeleistet` → `reloadState()` + `reloadHistorie()`

### 4.5 — Negativer Saldo Prominenz

**TablePage.tsx:** Saldo-Anzeige im Header, wenn negativ → rote Farbe + Badge `"Auszahlung ausstehend"`.

**Tischkarten-Komponente:** Negativer Saldo mit roter Farbe/Badge auf der Übersichtskarte.

---

## Phase 5 · TischHistorie — Auszahlung anzeigen

> Benötigt Phase 2. Kann parallel zu Phase 4 starten.

### 5.1 — query_handler.go

- DTO `auszahlung` (id, userId, tischId, betragCents, kommentar, geleistetAm)
- `toAuszahlung(a t.Auszahlung) auszahlung` Mapper
- Case `t.HistorieEintragAuszahlung` in `toHistorie()`

### 5.2 — TischHistorie.tsx

- `Auszahlung`-Typ zur Union in `historie`-Prop ergänzen
- Erkennung via Diskriminator: `art === 'auszahlung'`
- `HistoryItem` mit Titel `Auszahlung −X €`, Datum, Kommentar
- Details-Drawer: Betrag + Kommentar (keine Positionsliste)

---

## Phase 6 · Reporting

> Benötigt Phase 2 (Event-Typ bekannt). `make sqlc` nach SQL-Änderungen.

### 6.1 — reporting.sql

- `GetReportingStats`: `gesamt_umsatz_cents = kassierungen − auszahlungen`; neues Feld `gesamt_auszahlungen_cents`; Event-Typ `'tisch.auszahlung-geleistet:v1'` in `WHERE event_type IN (...)`
- `GetUmsatzProServicekraft`: CASE-When für `'tisch.zahlung-kassiert:v1'` und `'tisch.auszahlung-geleistet:v1'`; neues Feld `auszahlungen_cents`
- `GetUmsatzProTisch`: analog
- Neue Query `GetAusstehendAuszahlungen`: `SUM(ABS(saldo_cents)) WHERE saldo_cents < 0` aus `table_state`

### 6.2 — `make sqlc`

### 6.3 — reporting.go

- `Summary`: + `GesamtAuszahlungenCents`, `AusstehendAuszahlungenCents`
- `UmsatzServicekraft`: + `AuszahlungenCents`
- `UmsatzTisch`: + `AuszahlungenCents`

### 6.4 — repo.go

Neue Parallel-Query für `GetAusstehendAuszahlungen`; Mapping aller neuen Felder.

### 6.5 — query_handler.go

Response-DTOs: `gesamtAuszahlungenCents`, `ausstehendAuszahlungenCents`, `auszahlungenCents` per Servicekraft/Tisch.

### 6.6 — types.ts

Zod-Schemas um neue Felder ergänzen.

### 6.7 — ReportingResults.tsx

- Limitation-Tooltips (Servicekräfte + Tische) entfernen
- Neue Cards in Übersicht: `Auszahlungen` + `Ausstehende Auszahlungen`
- Servicekräfte/Tische-Tab: `auszahlungenCents` als Sub-Info anzeigen

---

## Parallelisierungs-Hinweise

```
Phase 0 (Doku)         ──────────────────────────────────> jederzeit
Phase 1 (Kommentar)    ──────────────────────────────────> unabhängig
Phase 2 (Domain Event) ──────┐
                             ├─> Phase 3 (Backend API) ─> Phase 4 (Frontend UI)
                             ├─> Phase 5 (Historie)    ─> parallel zu Phase 4
                             └─> Phase 6 (Reporting)   ─> parallel zu Phase 4+5
```

---

## Verifikation

```bash
make test           # Unit-Tests (inkl. angepasste Event-Fixtures)
make test-frontend  # Frontend-Tests
make lint           # Lint
```

**Manuelle Flows:**

- **A** — Stornierung ohne Kommentar: Submit-Button bleibt deaktiviert
- **B** — K-04b: bezahlte Position im Historie-Tab stornieren → negativer Saldo → Banner im Bezahlen-Tab
- **C** — Auszahlung bei negativem Saldo: vorausgefüllter Betrag, Kommentar, bestätigen → Saldo steigt
- **D** — Auszahlung bei positivem Saldo (mehrere Gruppen): freier Betrag eingeben → korrekte Buchung
- **E** — Reporting: Umsatz = Kassierungen − Auszahlungen; neue Cards sichtbar; keine Tooltips mehr

---

## Nicht in Scope

- Teila­uszahlungen mit Positionsreferenz (Betrag ist frei, keine Positionsbindung)
- Stornierung bezahlter Positionen im Bezahlen-Tab (nur über Historie-Tab)
- Offline-Handling für Auszahlungen
