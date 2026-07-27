# Plan: Storno-Zuordnung in der Abrechnung pro Servicekraft

> Source PRD: `docs/prds/prd-storno-zuordnung-servicekraft.md`

## Goal

Ein Storno wird in Reporting und Abrechnung der Servicekraft zugeordnet, deren
Vorgang er rückgängig macht — nicht dem Benutzer, der ihn ausgelöst hat. Die
Aufschlüsselung pro Servicekraft wird dabei von einer Brutto-Kassiert-Summe zu
einem Netto-Abrechnungsbetrag („Abzugeben"), gegen den der Kassenwart das
abgegebene Bargeld direkt prüfen kann.

Rein lesende Änderung: keine neuen Events, keine Migration, keine Änderung an
Kassenjournal, TSE-Signatur oder DSFinV-K-Export.

## Architectural decisions

Durable decisions that apply across all phases:

- **Zuordnungsregel (Kern der gesamten Arbeit).** Jede Storno-Art löst einen
  anderen Verweis auf, jeweils innerhalb derselben Kassensitzung:
  - `stornierung-erteilt:v1` → `zahlung-kassiert:v1` mit derselben `zahlungId`
    → dessen Akteur ist der **Kassierer**. Einwertig.
  - `direktverkauf-storniert:v1` → `direktverkauf-getaetigt:v1` mit derselben
    `verkaufId` → dessen Akteur ist der **Verkäufer**. Einwertig.
  - `bestellung-korrigiert:v1` → je Positions-ID das
    `bestellung-aufgenommen:v1`, dessen Positions-Array diese ID enthält →
    dessen Akteur ist der **Besteller**. **Mehrwertig.**
  - Lässt sich ein Verweis nicht auflösen, fällt die Zeile auf den Akteur
    zurück. Die Liste der Betroffenen ist damit nie leer.
- **Ein Auflösungsort.** Die Auflösung passiert genau einmal, in der
  `GetStornierungen`-Query, und liefert je Storno-Zeile eine Liste betroffener
  Servicekräfte. Alle Aggregate pro Servicekraft (Rücknahme-Betrag,
  Storno-Anzahl) entstehen daraus in der Anwendungsschicht — nach dem Muster,
  das `aggregateStornierungenProServicekraft()` heute schon verwendet. Einzige
  Ausnahme: `GetEigeneUebersicht` (Phase 3) hat keine Detailzeilen zur
  Verfügung und macht den `zahlungId`-Join selbst.
- **Begriff „Rücknahme".** Für die kassenwirksame Warenrücknahme
  (`stornierung-erteilt:v1`) gilt durchgängig **Rücknahme** — in Feldnamen
  (`ruecknahmenCents`, `anzahlRuecknahmen`), in Bezeichnern und in den
  UI-Labels. Nicht „Rückgabe" und nicht „Storno-Betrag"; „Bar-Rückgabe" bleibt
  ausschließlich das bestehende Badge an der Storno-Detailzeile.
- **Zusammengelegte Breakdown-Liste.** `breakdowns.umsatzProServicekraft` und
  `breakdowns.stornierungenProServicekraft` werden zu **einer** Liste
  `breakdowns.abrechnungProServicekraft`. Eine Person hat genau eine Zeile mit
  Geld **und** Storno-Anzahl; die Zuordnung zweier paralleler Listen per
  `userId` im Frontend entfällt.
- **JSON-Contract** (Frontend und Backend werden gemeinsam ausgeliefert, keine
  API-Versionierung nötig):
  - `servicekraftRef`: `{ userId, userName, name }` — geteilte Referenz für
    Akteur und Betroffene.
  - `abrechnungServicekraft`: `{ userId, userName, name, kassiertCents,
    anzahlZahlungen, ruecknahmenCents, anzahlStornierungen, abzugebenCents }`.
    `anzahlStornierungen` ist der kombinierte Zähler über beide
    Tisch-Storno-Arten (Rücknahmen und geldneutrale Korrekturen).
  - `stornierungDetail`: `akteur: servicekraftRef` und
    `betroffene: servicekraftRef[]` ersetzen die heutigen flachen Felder
    `userId` / `userName` / `name`.
  - `eigeneUebersicht` erhält zusätzlich `ruecknahmenCents`,
    `anzahlRuecknahmen` und `abzugebenCents`. Hier zählen **nur** kassenwirksame
    Rücknahmen — eine geldneutrale Korrektur ändert nichts an dem, was die
    Servicekraft abzugeben hat, und hat auf dieser Seite nichts zu suchen.
- **Go-Domänenmodelle** (`backend/domain/reporting/reporting.go`):
  `UmsatzServicekraft` → `AbrechnungServicekraft`; `StornierungServicekraft`
  entfällt ersatzlos; `StornierungDetail` trägt `Akteur` und
  `Betroffene []ServicekraftRef`; `ServicekraftLive` und `EigeneUebersicht`
  bekommen die neuen Geldfelder. Domain-Structs bleiben ohne `json`-Tags.
- **Direktverkauf bleibt aus jeder Servicekraft-Aggregation heraus** (eigene
  Kasse). Ein Direktverkauf-Storno bekommt seine Betroffenen-Auflösung nur für
  die Detailzeile.
- **Invariante:** `abzugebenCents` ist nie negativ. Begründung in
  `backend/domain/kasse/storno_aufteilung.go — ComputeStornoAufteilung()`: Eine
  Rücknahme kann nur Positionen der referenzierten Zahlung zurücknehmen, und
  `zahlungRest.rest` verhindert doppelte Rücknahme. Pro Zahlung gilt daher
  Σ Rücknahmen ≤ Zahlbetrag — und weil beide Seiten demselben Kassierer
  zugeordnet werden, auch pro Person.
- **Keine neuen Indizes.** Weder auf `(data->>'zahlungId')` noch auf
  Positions-IDs. Eine Kassensitzung umfasst wenige tausend Events, die Queries
  laufen on-demand für genau eine Sitzung.
- **`make sqlc` nach jeder Query-Änderung**, `make lint` nach Code-Änderungen.
  `backend/sqlc/dbgen/` wird nie von Hand editiert.

## Inventory

**Backend — Queries**

- `backend/sqlc/queries/reporting.sql — GetStornierungen` — liefert heute je
  Storno-Event den Akteur (`e.user_id`); hier entsteht die Betroffenen-Auflösung.
- `backend/sqlc/queries/reporting.sql — GetUmsatzProServicekraft` — summiert
  `zahlung-kassiert:v1` je Akteur, ohne jeden Storno-Abzug; klammert
  Direktverkäufe bereits bewusst aus.
- `backend/sqlc/queries/reporting.sql — GetEigeneUebersicht` — KPIs einer
  Servicekraft, gefiltert auf `user_id`.

**Backend — Repository**

- `backend/repository/reporting_repo/repo.go — toStornierungen()` — parst die
  Storno-Positionen aus dem Event-JSONB; hier kommen die Betroffenen dazu.
- `backend/repository/reporting_repo/repo.go — toUmsatzServicekraft()` —
  Mapping der Breakdown-Zeilen.
- `backend/repository/reporting_repo/repo.go — GetReporting()`,
  `GetLiveReporting()` — errgroup über die Teil-Queries.
- `backend/repository/reporting_repo/repo.go — GetEigeneUebersicht()`.

**Backend — Anwendungsschicht**

- `backend/api/reporting/application/query.go — aggregateStornierungenProServicekraft()`
  — gruppiert heute nach dem Akteur; wird zur Aggregation über die Betroffenen.
- `backend/api/reporting/application/query.go — mergeServicekraefteLive()` —
  führt Umsatz mit offener eigener Arbeit zusammen (`ComputeOffeneArbeitProServicekraft`).

**Backend — HTTP**

- `backend/api/reporting/http/query_handler.go — toUmsatzServicekraftList()`,
  `toStornierungenProServicekraft()`, `toStornierungDetail()`,
  `toServicekraftLive()`, `GetEigeneUebersichtHandler()` — Response-DTOs.
- `backend/api/service.go` — Route `/get-eigene-uebersicht`.

**Backend — Domäne**

- `backend/domain/reporting/reporting.go — UmsatzServicekraft`,
  `StornierungDetail`, `StornierungServicekraft`, `ServicekraftLive`,
  `EigeneUebersicht`.
- `backend/domain/kasse/storno_aufteilung.go — ComputeStornoAufteilung()` —
  FIFO-Aufteilung; Grundlage der Nicht-negativ-Invariante.
- `backend/domain/kasse/event_json_contract_test.go` — Guard, der grün bleiben
  muss (Beleg, dass nichts an den Events geändert wurde).

**Frontend**

- `frontend/src/admin/reporting/types.ts` — Zod-Schemas
  (`UmsatzServicekraftSchema`, `StornierungServicekraftSchema`,
  `StornierungDetailSchema`, `ServicekraftLiveSchema`).
- `frontend/src/admin/reporting/ReportingResults.tsx` — Abschnitt „Umsatz pro
  Servicekraft"; baut heute `stornoAnzahlByUserId` aus der zweiten Liste.
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Team-Liste,
  Storno-Collapsible.
- `frontend/src/admin/reporting/StornoItem.tsx` — Detailzeile eines Stornos.
- `frontend/src/admin/reporting/StornoServicekraft.tsx — StornoMarker()`,
  `StornoAggregat()`.
- `frontend/src/admin/reporting/utils.ts — formatServicekraft()` — „username
  (Klarname)".
- `frontend/src/service/components/EigeneUebersicht.tsx — EigeneUebersichtKarten()`
  — heute zwei Kacheln (`StatSpalte`) in einem `grid-cols-2` mit `divide-x`.
- `frontend/src/service/TableSelectionPage.tsx` — rendert die eigene Übersicht
  über der Tischliste.
- `frontend/src/service/table/Tisch.ts — EigeneUebersichtSchema`.

**Tests (Prior Art)**

- `backend/repository/reporting_repo/repo_test.go` — Integrationstests mit
  echter DB; Helfer `cleanDB()`, `createUser()`, `createKassensitzung()`;
  `TestGetReporting_IncludesBeideStornoArten` ist die nächste Verwandte.
- `backend/api/reporting/application/query_test.go — TestGetReporting_AggregiertStornierungenProServicekraft`,
  `TestGetLiveReporting_MergesServicekraefteByUserID`.
- `frontend/src/admin/reporting/ReportingResults.test.tsx`,
  `LiveReportingSection.test.tsx`.

## Resolved decisions

- **Eine Rücknahme wird dem Kassierer der referenzierten Zahlung angerechnet**,
  nicht dem Besteller und nicht dem Stornierenden. Der Bargeldfluss folgt der
  Zahlung; die `zahlungId` macht die Zuordnung eindeutig, auch wenn ein Storno
  Positionen mehrerer Bestellungen umfasst.
- **„Abzugeben" (Kassiert − Rücknahmen) wird die Hauptzahl** im Tagesbericht und
  in der Live-Team-Liste; Kassiert und Rücknahmen stehen als Nebenzeile darunter.
- **Direktverkauf bleibt aus der Servicekraft-Abrechnung heraus** — eigene
  Kasse, getrennt vom Tischservice. Ein DV-Storno bekommt in der Detailzeile
  dieselbe Zuordnungsregel (Verkäufer über `verkaufId`), fließt aber in keine
  Servicekraft-Summe ein.
- **Die geldneutrale Korrektur erzeugt nur einen Kontroll-Marker**, keinen
  Betrag, zugeordnet auf den Besteller der Positionen. Umfasst sie Positionen
  mehrerer Besteller, zählt sie **bei jedem** von ihnen.
- **Der Storno-Marker bleibt ein kombinierter Zähler** („2 Storno"), ohne
  Unterscheidung nach Storno-Art. Welche Art vorlag, steht in der
  Storno-Detailliste darunter. Damit bleibt die Team-Zeile mobil schlank.
- **Die „Eigene Übersicht" behält ihre zwei Kacheln.** Sie ist die
  Betriebs-Seite während des Abends, keine Abrechnung — eine dritte Kachel wäre
  fast immer leer, weil Stornos selten sind. Stattdessen erscheint **nur bei
  einer zugeordneten Rücknahme** eine Hinweiszeile unter den Kacheln, die den
  Abzug erklärt statt ihn nur zu beziffern.

## Open questions / Risks

- **Marker-Summe > Kopfkennzahl.** Zählt eine Korrektur bei mehreren
  Bestellern, kann die Summe aller Marker die Storno-Anzahl in der
  Summary-Kachel übersteigen. Als Kontroll-Signal ist das gewollt. Das Frontend
  darf den Marker deshalb nirgends als Aufteilung der Kopfkennzahl darstellen
  (keine „x von y"-Formulierung, keine Prozentangabe).
- **Rückwirkung auf abgeschlossene Sitzungen.** Weil die Zuordnung zur Lesezeit
  entsteht, ändert sich die Aufschlüsselung auch für bereits abgeschlossene
  Kassensitzungen. Ein alter Ausdruck weicht dann von einem neuen ab. Für
  Berichte, die keine fiskalische Ausfertigung sind, ist das unkritisch — es
  gehört aber in die Commit-Message, damit es niemanden überrascht.
- **Positions-Auflösung über JSONB.** Die Besteller-Auflösung durchsucht die
  `positionen`-Arrays der `bestellung-aufgenommen`-Events einer Sitzung ohne
  Index. Erwartet unkritisch (wenige tausend Events, on-demand, eine Sitzung).
  Fällt in Phase 2 eine spürbare Latenz auf, ist ein Index eine rein additive
  Migration und ein eigener Vorgang — nicht Teil dieses Plans.

---

## Phase 1: Storno-Zuordnung auflösen und in der Detailliste zeigen

**User stories**: 3, 5 (und die Zuordnungsgrundlage für 2, 4, 6)

### Context

- `backend/sqlc/queries/reporting.sql — GetStornierungen` — liefert heute
  `e.user_id` / `e.user_name` und joint `users` für den Klarnamen; hier kommen
  die drei Verweis-Auflösungen dazu.
- `backend/repository/reporting_repo/repo.go — toStornierungen()` — parst
  `kommentar` und `positionen` aus `data`; die Betroffenen kommen als eigene
  Spalte und werden hier ins Domänenmodell übersetzt.
- `backend/domain/reporting/reporting.go — StornierungDetail` — trägt heute
  flache Akteurs-Felder.
- `backend/api/reporting/http/query_handler.go — toStornierungDetail()`.
- `frontend/src/admin/reporting/StornoItem.tsx` — rendert heute
  `formatServicekraft(storno.userName, storno.name)`.
- `frontend/src/admin/reporting/types.ts — StornierungDetailSchema`.

### What to build

Die Storno-Query löst je Event die betroffenen Servicekräfte auf und liefert
sie als eigene Spalte neben dem unverändert mitgeführten Akteur. Rücknahme und
Direktverkauf-Storno ergeben genau eine betroffene Person (Join über
`zahlungId` bzw. `verkaufId` auf das Ursprungs-Event derselben Kassensitzung);
die geldneutrale Korrektur ergibt eine oder mehrere (je Positions-ID das
`bestellung-aufgenommen`-Event, dessen Positions-Array sie enthält,
dedupliziert). Findet ein Join nichts, ist der Akteur die betroffene Person.

Domänenmodell, Response-DTO und Zod-Schema tragen ab jetzt `akteur` und
`betroffene` statt der flachen Benutzerfelder. Die Detailzeile im Frontend
nennt die betroffenen Servicekräfte; weicht der Akteur von allen Betroffenen
ab, folgt „storniert von <Akteur>" als gedämpfter Zusatz.

`docs/language.md` bekommt die zwei neuen Begriffe: **Kassierer (kassierende
Servicekraft)** und **Storno-Zuordnung**; **Besteller** wird als für die
Korrektur-Zuordnung mitverwendet ergänzt.

### Acceptance criteria

- [x] `GetStornierungen` liefert je Storno-Zeile Akteur und eine nicht-leere
      Liste betroffener Servicekräfte (jeweils Benutzer-ID, eingefrorener
      Username, live aufgelöster Klarname).
- [x] Integrationstest: Serviceleitung nimmt eine von einer Servicekraft
      kassierte Zahlung zurück → betroffen ist die Servicekraft, Akteur ist die
      Serviceleitung.
- [x] Integrationstest: Ein Storno über zwei Zahlungen verschiedener Kassierer
      erzeugt zwei Events (FIFO) → jedes nennt seinen eigenen Kassierer.
- [x] Integrationstest: Geldneutrale Korrektur über Positionen zweier Besteller
      → beide sind als betroffen gelistet, jeder genau einmal.
- [ ] Integrationstest: Korrektur einer zuvor umgebuchten Position → betroffen
      ist der ursprüngliche Besteller, nicht der Umbucher.
      **Offen — die Annahme des PRD trifft nicht zu:**
      `kasse.NewBestellungUmgebuchtEvents()` vergibt auf dem Zieltisch *frische*
      Positions-IDs, die in keinem `bestellung-aufgenommen:v1` vorkommen. Die
      Positions-Auflösung findet daher nichts und fällt planmäßig auf den Akteur
      zurück (dokumentiert in
      `TestGetStornierungen_KorrekturUmgebuchterPositionFaelltAufAkteurZurueck`).
      Für den ursprünglichen Besteller wäre eine Abstammungs-Auflösung über die
      Umbuchungs-Event-Paare nötig (Produktentscheidung; die Projektion selbst
      stempelt bei der Umbuchung den Umbucher als Besteller auf die Positionen).
- [x] Integrationstest: Direktverkauf-Storno durch einen anderen Benutzer →
      betroffen ist der ursprüngliche Verkäufer.
- [x] Der Klarname wird auch für soft-gelöschte Benutzer aufgelöst, wie bisher
      (Verhalten aus `TestGetReporting_ResolvesKlarnameIncludingSoftDeleted`
      gilt für Akteur und Betroffene).
- [x] Die Storno-Detailzeile im Tagesbericht und im Live-Dashboard zeigt die
      betroffenen Servicekräfte; „storniert von <Akteur>" erscheint genau dann,
      wenn der Akteur nicht selbst betroffen ist.
- [x] `docs/language.md` führt **Kassierer** und **Storno-Zuordnung**.
- [x] `backend/domain/kasse/event_json_contract_test.go` ist unverändert grün.
- [x] `make check` läuft durch.

---

## Phase 2: Abrechnung pro Servicekraft mit „Abzugeben"

**User stories**: 1, 2, 4, 6

### Context

- `backend/sqlc/queries/reporting.sql — GetUmsatzProServicekraft` — bleibt
  inhaltlich die Kassiert-Summe, wird passend umbenannt.
- `backend/api/reporting/application/query.go — aggregateStornierungenProServicekraft()`
  — aggregiert heute nach Akteur; wird zur Aggregation über die Betroffenen aus
  Phase 1 und speist die zusammengelegte Liste.
- `backend/api/reporting/application/query.go — mergeServicekraefteLive()` —
  führt die Zahlen mit der offenen eigenen Arbeit zusammen; Personen ohne
  kassierten Umsatz erscheinen bereits heute.
- `backend/domain/reporting/reporting.go — UmsatzServicekraft`,
  `StornierungServicekraft`, `ServicekraftLive`.
- `backend/api/reporting/http/query_handler.go — toUmsatzServicekraftList()`,
  `toStornierungenProServicekraft()`, `toServicekraftLive()`.
- `frontend/src/admin/reporting/ReportingResults.tsx` — Abschnitt „Umsatz pro
  Servicekraft" samt `stornoAnzahlByUserId`.
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Team-Liste und
  `StornoAggregat`.
- `frontend/src/admin/reporting/StornoServicekraft.tsx — StornoMarker()`,
  `StornoAggregat()`.

### What to build

Die zwei parallelen Breakdown-Listen werden zu einer
`abrechnungProServicekraft`. Pro Person: Kassiert (Tischzahlungen nach Akteur,
unverändert), Rücknahmen (Summe der ihr zugeordneten Rücknahmen aus den
Storno-Zeilen der Phase 1), Abzugeben (Kassiert − Rücknahmen) und die
Storno-Anzahl als kombinierter Zähler über beide Tisch-Storno-Arten.
Direktverkäufe und Direktverkauf-Stornos bleiben außen vor. Eine Person
erscheint, sobald sie kassiert hat, einen Storno zugeordnet bekommt oder — im
Live-Dashboard — offene Arbeit hat; sortiert wird nach Abzugeben absteigend.

Im Tagesbericht heißt der Abschnitt „Abrechnung pro Servicekraft" mit der
Unterzeile, dass Direktverkäufe nicht enthalten sind; pro Zeile steht Abzugeben
als Hauptzahl, darunter Kassiert und Rücknahmen. In der Live-Team-Liste
dieselbe Hauptzahl, die Rücknahmen nur eingeblendet, wenn sie ungleich null
sind. Der Storno-Marker sitzt jetzt bei der betroffenen Person; `StornoAggregat`
speist sich aus derselben Liste (Einträge mit Storno-Anzahl > 0).

`docs/anforderungen.md` (R-04) und der Reporting-Abschnitt in
`docs/handbuch.md` werden auf den Abrechnungs-Saldo und die Zuordnungsregel
nachgezogen.

### Acceptance criteria

- [x] Die Response trägt `breakdowns.abrechnungProServicekraft`;
      `breakdowns.umsatzProServicekraft` und
      `breakdowns.stornierungenProServicekraft` existieren nicht mehr — weder in
      der Abrechnungs- noch in der Live-Response.
- [x] Integrationstest: Serviceleitung nimmt stellvertretend zurück →
      Rücknahmen und Abzugeben ändern sich bei der Servicekraft, die
      Serviceleitung bleibt unverändert.
- [x] Integrationstest: Die Servicekraft nimmt selbst zurück → dasselbe
      Ergebnis; die Zuordnung ist keine Sonderregel für Vertretungsfälle.
- [x] Integrationstest: Servicekraft A kassiert, was B bestellt hat, danach
      Rücknahme → Rücknahmen bei A, B unverändert.
- [x] Integrationstest: Geldneutrale Korrektur → Storno-Anzahl beim Besteller
      steigt, Kassiert / Rücknahmen / Abzugeben bleiben unverändert.
- [x] Integrationstest: Direktverkauf und Direktverkauf-Storno verändern keine
      Zeile der Abrechnung pro Servicekraft.
- [x] Integrationstest: Vollständige Rücknahme einer Zahlung → Abzugeben ist
      null, nicht negativ (Invariante).
- [x] Anwendungsschicht-Test: Die Summe aller Abzugeben-Beträge entspricht dem
      kassierten Tischservice-Umsatz der Sitzung abzüglich aller Rücknahmen —
      Direktverkäufe auf beiden Seiten ausgenommen.
- [x] Eine Person mit ausschließlich zugeordneten Stornos und ohne eigenes
      Kassieren erscheint in der Liste.
- [x] Frontend-Test: Der Tagesbericht zeigt pro Zeile Abzugeben als Hauptzahl
      sowie Kassiert und Rücknahmen; der Storno-Marker steht bei der
      betroffenen, nicht bei der stornierenden Person.
- [x] Frontend-Test: Die Live-Team-Liste blendet die Rücknahmen nur bei einem
      Betrag ungleich null ein.
- [x] `docs/anforderungen.md` (R-04) und `docs/handbuch.md` beschreiben den
      Abrechnungs-Saldo und die Zuordnungsregel.
- [ ] `make verify` läuft durch.
      **Nur in Teilen belegt:** In der Umsetzungsumgebung fehlt Docker, `make
      verify` bricht daher an seinem Docker-Schritt ab. Verifiziert wurden
      stattdessen `make check` (grün) und die vollständige Integrationstest-Suite
      gegen die lokale PostgreSQL-Instanz (`go test -tags=integration ./...`,
      alle Pakete grün) sowie `pnpm test` (455 Tests grün).

---

## Phase 3: Rücknahme-Hinweis in der eigenen Übersicht der Servicekraft

**User stories**: 7

### Context

- `backend/sqlc/queries/reporting.sql — GetEigeneUebersicht` — aggregiert heute
  `bestellung-aufgenommen` und `zahlung-kassiert` gefiltert auf `user_id`; hier
  fehlen die Detailzeilen aus Phase 1, der `zahlungId`-Join passiert deshalb in
  dieser Query selbst.
- `backend/domain/reporting/reporting.go — EigeneUebersicht`.
- `backend/api/reporting/http/query_handler.go — GetEigeneUebersichtHandler()`.
- `backend/api/service.go` — Route `/get-eigene-uebersicht`.
- `frontend/src/service/table/Tisch.ts — EigeneUebersichtSchema`.
- `frontend/src/service/components/EigeneUebersicht.tsx — EigeneUebersichtKarten()`
  — zwei `StatSpalte` in einem `grid-cols-2` mit `divide-x`, plus
  Skeleton-Zustand.
- `frontend/src/service/TableSelectionPage.tsx` — rendert die Karten über der
  Tischliste; hier landet die Hinweiszeile im Fluss.

### What to build

Die Query ermittelt zusätzlich die Rücknahmen, die dieser Servicekraft über die
`zahlungId`-Zuordnung zufallen (Anzahl und Betrag), sowie den daraus
resultierenden Abzugeben-Betrag — dieselbe Regel wie im Admin-Bericht, nur für
genau einen Benutzer und ohne Umweg über die Detailzeilen. Geldneutrale
Korrekturen bleiben hier außen vor: Sie ändern nichts an dem, was abzugeben ist.

Das Service-Dashboard behält seine zwei Kacheln unverändert (Bestellungen |
Kassiert, brutto). Nur wenn dieser Servicekraft mindestens eine Rücknahme
zugeordnet ist, erscheint unter dem Kachel-Block eine Hinweiszeile, die den
Abzug in ganzen Sätzen erklärt und den abzugebenden Betrag nennt — genug Raum,
damit die Servicekraft versteht, warum sie weniger abgibt, statt nur eine
kleinere Zahl zu sehen. Ohne Rücknahme ist die Seite pixelgleich zu heute.

`docs/anforderungen.md` (R-06) und der Read-Model-Eintrag „Eigene Übersicht" in
`docs/handbuch.md` werden nachgezogen.

### Acceptance criteria

- [ ] Die Response von `/service/get-eigene-uebersicht` enthält
      `ruecknahmenCents`, `anzahlRuecknahmen` und `abzugebenCents`.
- [ ] Integrationstest: Nimmt jemand anderes eine von dieser Servicekraft
      kassierte Zahlung zurück, sinkt ihr `abzugebenCents` und
      `anzahlRuecknahmen` steigt; `zahlungenCents` bleibt unverändert.
- [ ] Integrationstest: Nimmt diese Servicekraft eine von jemand anderem
      kassierte Zahlung zurück, bleiben alle drei Felder unverändert.
- [ ] Integrationstest: Eine geldneutrale Korrektur verändert keines der drei
      Felder.
- [ ] `abzugebenCents` einer Servicekraft stimmt mit ihrer Zeile in
      `breakdowns.abrechnungProServicekraft` derselben Kassensitzung überein.
- [ ] Frontend-Test: Ohne Rücknahme zeigt das Service-Dashboard genau die zwei
      bisherigen Kacheln und keine Hinweiszeile.
- [ ] Frontend-Test: Mit Rücknahme erscheint die Hinweiszeile mit Anzahl,
      zurückgegebenem Betrag und abzugebendem Betrag; die Kacheln bleiben
      unverändert.
- [ ] Die Hinweiszeile ist ab 360 px Breite ohne Überlauf lesbar; der
      Lade-Skeleton bleibt zweispaltig wie bisher.
- [ ] `docs/anforderungen.md` (R-06) und `docs/handbuch.md` beschreiben die
      erweiterte eigene Übersicht.
- [ ] `make verify` läuft durch.
