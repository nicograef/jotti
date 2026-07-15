# Plan: Produkt- und Varianten-Statistik im Report

> Source PRD: `docs/prds/prd-produkt-varianten-statistik.md`

## Goal

Der Report zeigt pro Kassensitzung eine Aufschlüsselung der Verkäufe pro Produkt
und Variante — je Zeile zwei getrennte Zahlen: **ausgegebene Menge** (Bestellung
− Korrektur + Direktverkauf) und **Umsatz** (Kassiert + Direktverkauf −
Warenrücknahme/Storno). Gegliedert in Kategorie-Abschnitte (Essen → Getränke →
Sonstiges), pro Variante aufgeschlüsselt und auf Produktebene mit Zwischensumme
gruppiert, Ein-Varianten-Produkte zu einer Zeile zusammengefasst, je Kategorie
nach Menge absteigend sortiert. Angezeigt sowohl in der Tagesabrechnung
(`get-abrechnung`) als auch im Live-Dashboard (`get-live-reporting`). Setzt die
Roadmap-Anforderung **R-05** um.

## Architectural decisions

Dauerhafte Entscheidungen, die für alle Phasen gelten:

- **Endpunkte**: keine neuen. Die bestehenden `admin/get-abrechnung` und
  `admin/get-live-reporting` tragen ein neues Response-Feld `produktStatistik`.
- **Datenquelle**: reine additive Leseauswertung über `kassenjournal`, gefiltert
  nach `kassensitzung_nr`. Kein neues Event, keine Schema-Migration, kein Join auf
  Stammdaten (Reporting-ACL bleibt gewahrt). Positionsdaten kommen aus den
  eingefrorenen Fat-Events.
- **SQL**: eine neue sqlc-Query `GetProduktStatistik` in
  `backend/sqlc/queries/reporting.sql`, die `data->'positionen'` per
  `jsonb_array_elements` entfaltet und **flache Zeilen pro Variante**
  `(kategorie, varianteId, produktName, varianteName, ausgegebeneMenge, umsatzCents)`
  liefert (`GROUP BY kategorie, varianteId, produktName, varianteName`). Kein neuer
  `kj_extract_*`-DB-Helper (die Vorzeichen-Logik steht inline in der Query). Nach
  Query-Änderung: `make sqlc`.
  - Casts wie in den bestehenden Reporting-Queries: `(position->>'varianteId')::int`,
    `(position->>'menge')::int`, `(position->>'einzelpreisCents')::int` sowie
    `COALESCE(SUM(... CASE ...), 0)::int` für beide Summenspalten, damit sqlc
    Go-`int` erzeugt (passend zu den `int`-Domänenfeldern). `kategorie`,
    `produktName`, `varianteName` bleiben `text` — **kein** Enum-Cast (die Query
    trägt keinen `Steuersatz`, anders als `GetUmsatzPositionszeilen`).
- **Event-Mapping** (in der SQL-Query als `CASE`):
  - *Ausgegebene Menge* = Σ `menge`: `bestellung-aufgenommen:v1` (+),
    `bestellung-korrigiert:v1` (−), `direktverkauf-getaetigt:v1` (+).
  - *Umsatz* = Σ `einzelpreisCents × menge`: `zahlung-kassiert:v1` (+),
    `direktverkauf-getaetigt:v1` (+), `stornierung-erteilt:v1` (−),
    `direktverkauf-storniert:v1` (−).
  - `bestellung-umgebucht:v1` wird **nicht** gezählt (Positionen bereits bei der
    Bestellung erfasst).
- **Domänen-Modelle** (`backend/domain/reporting/reporting.go`, keine `json`-Tags):
  - `ProduktStatistikZeile` — **flacher Zeilentyp** (Repo-Ausgabe, Eingabe der
    Gruppierung): `Kategorie string`, `ProduktName string`, `VarianteID int`,
    `VarianteName string`, `AusgegebeneMenge int`, `UmsatzCents int`. Genau die
    Spalten der SQL-Query.
  - `VarianteStatistik` — `VarianteID int`, `VarianteName string`,
    `AusgegebeneMenge int`, `UmsatzCents int`.
  - `ProduktStatistik` — `Kategorie string`, `ProduktName string`,
    `AusgegebeneMenge int` (Zwischensumme), `UmsatzCents int` (Zwischensumme),
    `Varianten []VarianteStatistik`.
  - Neues Feld `ProduktStatistik []ProduktStatistik` auf `ReportingData` **und**
    `LiveReportingData` (die *gruppierte* Form; die flachen Zeilen erscheinen nie
    in der Response).
- **Repo → Gruppierung → Response (Plumbing, ausdrücklich festgelegt):** die
  `computeUmsatzProSteuersatz`-In-place-Analogie gilt hier **nicht**, weil Ein- und
  Ausgabe unterschiedliche Formen haben (flache Zeilen → verschachtelte
  Produkte). Deshalb:
  1. Eine dedizierte Repo-Methode `GetProduktStatistik(ctx, kassensitzungNr)
     ([]reporting.ProduktStatistikZeile, error)` auf dem `reportingRepo`-Interface
     liefert die flachen Zeilen (SQL-Rows → `ProduktStatistikZeile`).
  2. Die Anwendungsschicht ruft sie in `Query.GetReporting` **und**
     `Query.GetLiveReporting` auf und weist
     `data.ProduktStatistik = gruppiereProduktStatistik(zeilen)` zu — analog dazu,
     wie `GetLiveReporting` bereits mehrere Repos orchestriert. Kein transientes
     Feld auf der Domäne/Response; ein zusätzlicher On-Demand-Query-Roundtrip ist
     für einen Admin-Report vernachlässigbar.
- **Gruppierung/Sortierung**: eine reine Funktion `gruppiereProduktStatistik`
  in der Anwendungsschicht (`backend/api/reporting/application/`, neben
  `computeUmsatzProSteuersatz` in `query.go`) baut aus den flachen
  `ProduktStatistikZeile`-Werten die Produkt-Hierarchie, berechnet Zwischensummen
  und sortiert: Kategorien fest Essen → Getränke → Sonstiges; Produkte je Kategorie
  nach `AusgegebeneMenge` absteigend; Varianten je Produkt nach `AusgegebeneMenge`
  absteigend; `ProduktName`/`VarianteName` als stabiler Tiebreaker. Backend liefert
  die Liste fertig sortiert. Isoliert unit-testbar (flache Eingabe → gruppierte,
  sortierte Ausgabe).
- **JSON-Contract** (HTTP-Response-DTO in
  `backend/api/reporting/http/query_handler.go` und Live-Pendant): Feld
  `produktStatistik` = Array von `{ kategorie, produktName, ausgegebeneMenge,
  umsatzCents, varianten: [{ varianteId, varianteName, ausgegebeneMenge,
  umsatzCents }] }`.
- **Frontend**: neue Zod-Schemas `VarianteStatistikSchema` /
  `ProduktStatistikSchema` in `frontend/src/admin/reporting/types.ts`, eingehängt
  in `ReportingDataSchema` und `LiveReportingDataSchema`. Anzeige als neue
  Komponente, eingebunden in `ReportingResults.tsx` (Abrechnung) und
  `LiveReportingSection.tsx` (Live). Ein-Varianten-Produkte werden in der
  Darstellung zu einer Zeile zusammengefasst (reine Präsentation; Datenmodell
  bleibt Produkt + eine Varianten-Zeile).

## Inventory

- `backend/sqlc/queries/reporting.sql` — bestehende Reporting-Queries; Muster
  `GetUmsatzPositionszeilen` (entfaltet `data->'positionen'`, liefert unaggregierte
  Zeilen) ist die direkte Vorlage.
- `backend/repository/reporting_repo/repo.go — Repository.GetReporting()` /
  `Repository.GetLiveReporting()` — feuern die Reporting-Queries parallel per
  `errgroup` und mappen Rows → Domäne. Vorlage für die neue dedizierte Methode
  `Repository.GetProduktStatistik()`, die SQL-Rows → `[]ProduktStatistikZeile` mappt.
- `backend/domain/reporting/reporting.go` — Read-Model-Typen (`ReportingData`,
  `LiveReportingData`, `UmsatzServicekraft`, `StornierungPosition`); Vorlage für
  die neuen Typen und den neuen Feld-Anbau.
- `backend/api/reporting/application/query.go — Query.GetReporting()` /
  `Query.GetLiveReporting()` und `computeUmsatzProSteuersatz()` — Vorlage für die
  reine Gruppierungs-/Sortierfunktion und deren Aufruf.
- `backend/api/reporting/http/query_handler.go — toReportingResponse()`,
  `reportingResponse`, `umsatzSteuersatzResponse`, `stornierungPosition` — DTO- und
  Mapper-Muster; Live-Pendant im selben File (`liveSummaryResponse` ff.).
- `backend/domain/kasse/bestellung.go — Position`, `PositionEventData`,
  `Position.Bezeichnung()` — Positions-Felder (`VarianteID`, `ProduktName`,
  `VarianteName`, `Kategorie`, `EinzelpreisCents`, `Menge`) und die
  Produkt-plus-Variante-Beschriftung.
- `backend/domain/kasse/event_json_contract_test.go` — eingefrorene JSON-Keys der
  Positionen (`varianteId`, `produktName`, `varianteName`, `kategorie`,
  `einzelpreisCents`, `menge`); Referenz, nicht zu ändern.
- `frontend/src/admin/reporting/types.ts` — Zod-Schemas
  (`ReportingDataSchema`, `LiveReportingDataSchema`, `UmsatzSteuersatzSchema`,
  `StornierungPositionSchema`); Vorlage und Einhängepunkt.
- `frontend/src/admin/reporting/ReportingResults.tsx — ReportingResults()` — Body
  der Abrechnung (Kennzahlen, „Umsatz nach Steuersatz", Servicekraft-/Storno-Listen);
  Einhängepunkt für den neuen Abschnitt.
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Live-Dashboard-Body;
  zweiter Einhängepunkt.
- `frontend/src/admin/reporting/ReportingBackend.ts` — `getReporting()` /
  `getLiveReporting()`; unverändert im Aufruf, validieren die erweiterten Schemas.
- `frontend/src/admin/reporting/utils.ts`, `@/lib/utils — formatEuro` — Cent-Format.
- Tests als Prior Art: `backend/repository/reporting_repo/repo_test.go`,
  `backend/api/reporting/application/query_test.go`,
  `backend/api/reporting/application/query_export_konsistenz_test.go`,
  `frontend/src/admin/reporting/ReportingResults.test.tsx`,
  `frontend/src/admin/reporting/LiveReportingSection.test.tsx`.

## Resolved decisions

- **Zwei getrennte Grundlagen** für Menge (Produktion) und Umsatz (Einnahmen) —
  bewusst nicht ineinander umrechenbar; erklärender Hinweis im Report.
- **Direktverkauf** zählt in beide Zahlen.
- **Kategorie-Abschnitte** Essen → Getränke → Sonstiges; keine Kategorie-Summenzeile.
- **Sortierung nach ausgegebener Menge absteigend** je Kategorie.
- **Ein-Varianten-Produkte** als eine Zeile (Beschriftung `produktName varianteName`).
- **Kein neues Event, keine Migration, kein Stammdaten-Join.**
- Gruppierung/Sortierung im **Backend** (Single Source of Truth für Aufbereitung).

## Open questions / Risks

- **Sortierrichtung**: absteigend nach Menge (Meistverkaufte oben) — festgelegt.
- **Konsistenz-Invariante**: Σ aller Produkt-`UmsatzCents` muss dem kassierten
  Gesamtumsatz (= Σ `UmsatzProSteuersatz`-Brutto) entsprechen — dieselbe
  Positions-Basis. Einziger realer Bruchweg: die `WHERE type IN (…)`-Menge der
  neuen Query driftet von `GetUmsatzPositionszeilen` ab. Abgesichert im
  Repo-Integrationstest (Phase 1), nicht im Anwendungs-Unit-Test.

---

## Phase 1: Produkt-/Varianten-Statistik in der Tagesabrechnung

**User stories**: 1, 2, 3, 4, 6

### Context

- `backend/sqlc/queries/reporting.sql` — `GetUmsatzPositionszeilen` als
  Query-Vorlage (Entfaltung von `data->'positionen'`).
- `backend/domain/reporting/reporting.go` — Read-Model-Typen; neue Typen +
  Feld-Anbau an `ReportingData`.
- `backend/repository/reporting_repo/repo.go — Repository.GetReporting()` —
  Row→Domäne-Mapping-Muster; die neue Query kommt als **dedizierte Methode**
  `Repository.GetProduktStatistik()` hinzu (nicht in die `GetReporting`-`errgroup`
  eingereiht), aufgerufen von der Anwendungsschicht.
- `backend/api/reporting/application/query.go — Query.GetReporting()`,
  `computeUmsatzProSteuersatz()` — Aufruf- und Muster-Vorlage für die neue reine
  Gruppierungs-/Sortierfunktion.
- `backend/api/reporting/http/query_handler.go — toReportingResponse()`,
  `reportingResponse` — DTO + Mapper.
- `frontend/src/admin/reporting/types.ts — ReportingDataSchema` — Zod-Einhängepunkt.
- `frontend/src/admin/reporting/ReportingResults.tsx — ReportingResults()` —
  Anzeige-Einhängepunkt.
- `frontend/src/admin/reporting/ReportingResults.test.tsx` — Test-Vorlage.

### What to build

Ein durchgehender vertikaler Schnitt für die Tagesabrechnung: eine neue
sqlc-Query aggregiert die eingefrorenen Positionen einer Kassensitzung zu flachen
Varianten-Zeilen mit ausgegebener Menge und Umsatz (Vorzeichen laut Event-Mapping
oben). Eine dedizierte Repo-Methode `GetProduktStatistik` mappt die Rows in
`[]ProduktStatistikZeile`. Die reine Funktion `gruppiereProduktStatistik` in der
Anwendungsschicht gruppiert diese zu Kategorie-Abschnitten und Produkten mit
Zwischensummen und sortiert sie (Kategorie-Reihenfolge fest, Menge absteigend,
Name-Tiebreaker); `Query.GetReporting` weist das Ergebnis
`data.ProduktStatistik` zu. Es wird als `produktStatistik`-Feld der
`get-abrechnung`-Response serialisiert, im Frontend per Zod validiert und in
`ReportingResults.tsx` als neuer Abschnitt „Verkäufe pro Produkt" angezeigt:
Kategorie-Überschriften, Produktzeile mit Zwischensumme und Variantenzeilen
darunter, Ein-Varianten-Produkte als eine Zeile, Spalten „Ausgegeben" und „Umsatz"
(`formatEuro`), erklärender Ein-Satz-Hinweis, leerer Zustand, druckfreundlich als
Teil des Z-Bons.

### Acceptance criteria

- [x] Für eine Kassensitzung mit Bestellungen, Korrekturen, Zahlungen,
  Warenrücknahmen, Direktverkäufen und Direktverkauf-Stornos liefert die Auswertung
  je Variante die korrekte ausgegebene Menge (Bestellung − Korrektur +
  Direktverkauf) und den korrekten Umsatz (Kassiert + Direktverkauf − Storno);
  Umbuchungen verändern beide Zahlen nicht.
- [x] Warenrücknahme/Direktverkauf-Storno mindern nur den Umsatz, nicht die Menge;
  eine geldneutrale Korrektur mindert nur die Menge, nicht den Umsatz.
- [x] Die Ausgabe ist in Kategorie-Abschnitte (Essen → Getränke → Sonstiges)
  gegliedert; je Produkt gibt es eine Zwischensumme über seine Varianten; Produkte
  und Varianten sind nach ausgegebener Menge absteigend sortiert (Name-Tiebreaker).
- [x] `gruppiereProduktStatistik` ist als reine Funktion isoliert unit-getestet
  (`query_test.go`, Muster `computeUmsatzProSteuersatz`): flache
  `ProduktStatistikZeile`-Eingabe → korrekte Kategorie-Reihenfolge, Produkt-Gruppen,
  Zwischensummen, absteigende Mengensortierung und stabiler Tiebreaker.
- [x] Σ aller Produkt-`umsatzCents` entspricht der Summe der
  `umsatzProSteuersatz`-Bruttowerte derselben Kassensitzung. Diese Invariante wird
  im **Repo-Integrationstest** (`repo_test.go`, echtes Postgres, dieselben Events
  speisen `GetProduktStatistik` und `GetUmsatzPositionszeilen`) geprüft — dort, wo
  ein Drift der `WHERE type IN (…)`-Mengen real auffällt; ein
  Anwendungs-Unit-Test mit handgebauten Eingaben kann diesen SQL-Drift nicht fangen.
- [x] `admin/get-abrechnung` liefert das Feld `produktStatistik` mit
  verschachtelten `varianten`; das Frontend validiert es per Zod ohne Fehler.
- [x] `ReportingResults.tsx` zeigt den Abschnitt „Verkäufe pro Produkt" mit
  Kategorie-Überschriften, Zwischensummen, den Spalten „Ausgegeben"/„Umsatz", dem
  erklärenden Hinweis und einem leeren Zustand bei einer Sitzung ohne Verkäufe;
  Ein-Varianten-Produkte erscheinen als eine Zeile.
- [x] `make sqlc`, `make lint` und die betroffenen Backend- und Frontend-Tests
  laufen grün; `sqlc/dbgen/` wurde nicht von Hand editiert.

---

## Phase 2: Produkt-/Varianten-Statistik im Live-Dashboard

**User stories**: 5, 6

### Context

- `backend/repository/reporting_repo/repo.go — Repository.GetProduktStatistik()` —
  in Phase 1 erstellt; wird im Live-Pfad von `Query.GetLiveReporting` unverändert
  mitbenutzt.
- `backend/api/reporting/application/query.go — Query.GetLiveReporting()` — ruft
  die in Phase 1 erstellte Gruppierungs-/Sortierfunktion für die Live-Daten auf.
- `backend/domain/reporting/reporting.go — LiveReportingData` — Feld-Anbau
  (Typen aus Phase 1 wiederverwendet).
- `backend/api/reporting/http/query_handler.go` — Live-Response-DTO (Pendant zu
  `toReportingResponse`) um `produktStatistik` erweitern.
- `frontend/src/admin/reporting/types.ts — LiveReportingDataSchema` — Zod-Einhängepunkt.
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Anzeige-Einhängepunkt.
- `frontend/src/admin/reporting/LiveReportingSection.test.tsx` — Test-Vorlage.

### What to build

Der in Phase 1 gebaute Backend-Kern (Query, `GetProduktStatistik`-Repo-Methode,
Domäne-Typen, `gruppiereProduktStatistik`) wird in den Live-Pfad eingehängt:
`Query.GetLiveReporting` ruft dieselbe Repo-Methode und dieselbe
Gruppierungsfunktion auf und weist `data.ProduktStatistik` zu (keine Duplikat-
Logik); das Feld hängt an `LiveReportingData` und wird in der
`get-live-reporting`-Response als `produktStatistik` serialisiert. Im Frontend wird
dieselbe Anzeige-Komponente aus Phase 1 in `LiveReportingSection.tsx` eingebunden,
sodass die Statistik der offenen Kassensitzung in Echtzeit erscheint.

### Acceptance criteria

- [x] `admin/get-live-reporting` liefert für die offene Kassensitzung das Feld
  `produktStatistik` mit identischer Struktur und identischen Zahlen-Regeln wie die
  Abrechnung (dieselbe Gruppierungs-/Sortierfunktion, keine Duplikat-Logik).
- [x] Das Live-Dashboard zeigt den Abschnitt „Verkäufe pro Produkt" mit denselben
  Kategorie-Abschnitten, Zwischensummen und Spalten wie die Abrechnung; ein leerer
  Zustand erscheint, solange noch keine Verkäufe vorliegen.
- [x] In einer offenen Sitzung mit noch unbezahlten Bestellungen erscheint die
  ausgegebene Menge > 0 bei Umsatz 0 (bestellt/rausgegeben, aber noch nicht
  kassiert) — konsistent mit den Kennzahl-Regeln.
- [x] `make lint` und die betroffenen Backend-/Frontend-Tests laufen grün.

---

## Phase 3: Dokumentation nachziehen

**User stories**: — (Doku-Pflege gemäß AGENTS.md)

### Context

- `docs/anforderungen.md` — Roadmap-Tabelle (R-05) und Reporting-Funktionsumfang
  (R-01 ff.).
- `docs/handbuch.md` — §7.2 Admin-Ansichten (Reporting), Read-Model-Übersicht und
  Endpunkt-Tabelle (`get-abrechnung`/`get-live-reporting`).
- `docs/language.md` — Read-Model-Begriffe (`Summary`, `Breakdowns`,
  `UmsatzServicekraft` …).

### What to build

Die Doku wird an das umgesetzte Feature angepasst: R-05 wandert in
`docs/anforderungen.md` aus der Roadmap in den Reporting-Funktionsumfang; die neuen
Read-Model-Typen (`ProduktStatistik`, `VarianteStatistik`) und das erweiterte
Response-Feld `produktStatistik` werden in der Read-Model-Übersicht und der
Endpunkt-Beschreibung von `docs/handbuch.md` §7.2 sowie in der Begriffsliste von
`docs/language.md` ergänzt.

### Acceptance criteria

- [x] R-05 steht nicht mehr in der Roadmap, sondern als umgesetzte Anforderung im
  Reporting-Funktionsumfang von `docs/anforderungen.md`.
- [x] `docs/handbuch.md` §7.2 nennt `produktStatistik` als Bestandteil der
  `get-abrechnung`- und `get-live-reporting`-Antwort und führt die neuen
  Read-Model-Typen in der Übersicht.
- [x] `docs/language.md` enthält Einträge für `ProduktStatistik` und
  `VarianteStatistik`.
- [x] Keine toten Verweise; Begriffe konsistent mit den im Code verwendeten Namen.
