# Plan: Admin-Dashboard & Kassenberichte — UX-Umbau (UX-Review 2026-07-11)

> Source PRD: n/a — abgeleitet aus dem UX-Review des Admin-Dashboards (Chat-Session 2026-07-11,
> Praxistest-Feedback 2026-07-09: `praxistest-2026-07-09.md`, Zeile „Admin UI: Vereinfachen …")

## Goal

Das Admin-Dashboard beantwortet die drei Fragen des Kassenwarts direkt auf dem Handy
(375 px Baseline): **Wie viel ist in der Kasse? Steht noch Geld an Tischen aus? Gibt es
auffällige Stornierungen?** Dafür:

1. Entfernen, was keine Entscheidung stützt: Progressbars, „N Zahlungen"-Badges,
   Umsatz-pro-Tisch-Ranking.
2. Live-Dashboard als eine scrollbare Seite in Prioritätsreihenfolge, mit Auto-Refresh.
3. Historische Auswertung als eigene Seite **Kassenberichte** (nur abgeschlossene
   Kassensitzungen) — keine doppelten Zahlen mehr auf einer Seite.
4. Neu (einzige Ergänzung): Stornierungen **pro Servicekraft** als Kontroll-Signal.
5. Mobile-Politur: Above-the-fold, Ladezustand, Abschnitts-Header, Touch-Layout.

## Architectural decisions

Durable decisions that apply across all phases:

- **Routes**: `/admin/auswertung` bleibt das Live-Dashboard und die Landing-Route
  (Index-Redirect in `frontend/src/routes.ts` unverändert). Neu: `/admin/kassenberichte`
  (lazy, `AdminGuard` via Eltern-Route) für abgeschlossene Kassensitzungen, Steuersatz-Tabelle
  und DSFinV-K-Export. Sidebar-Gruppe „Auswertungen" bekommt zwei Einträge:
  **„Live-Dashboard"** (`/admin/auswertung`) und **„Kassenberichte"** (`/admin/kassenberichte`).
- **API**: `POST /admin/get-all-kassensitzungen` wird zu
  `POST /admin/get-abgeschlossene-kassensitzungen` (einziger Consumer ist die
  Reporting-Seite). Liefert nur `status = 'abgeschlossen'` (der transiente Status
  `wird_abgeschlossen` erscheint nicht). Filterung im Backend (Regel 9), neue sqlc-Query
  `GetAbgeschlosseneKassensitzungen`.
- **Kanonische Kennzahlen-Reihenfolge** (Live und Kassenberichte identisch):
  1. Gesamtumsatz (kassiert), 2. Offene Saldi (nur live), 3. Bestellungen,
  4. Direktverkauf, 5. Stornierungen. Die Tiles Bestellungen/Direktverkauf bleiben auf dem
  Live-Dashboard erhalten (explizite Nutzerentscheidung: der Kassenwart will den
  bestellten/„kommenden" Umsatz inkl. noch nicht kassierter Positionen live sehen).
- **Live-Dashboard-Layout**: eine scrollbare Seite ohne Tabs, Blockreihenfolge:
  kompakte Systemwarnungen → Kennzahlen-Tiles → Offene Tische → Stornierungen →
  Servicekräfte. `ScrollableTabsList` entfällt auf dieser Seite.
- **Servicekraft-Zeile (live)**: `Name · kassiert €` plus Status: grünes „Fertig" **oder**
  „Offen: X € · Tisch 11, Zelt A2" (eigene unbezahlte Positionen in Euro, Tischnamen inline,
  gedämpft). Keine Progressbar, kein Zahlungen-Badge, keine Chip-Kacheln. Der Offen-Betrag
  kommt aus der Domäne: `kasse.EigeneArbeitAnTisch`/`kasse.OffeneArbeitTisch` werden additiv
  um `OffenCents` erweitert (Summe `EinzelpreisCents × Menge` der eigenen unbezahlten
  Positionen). Die Service-Seite (Eigene Übersicht, R-06) bleibt unverändert; das Feld ist
  dort nur zusätzlich verfügbar.
- **Storno-Aggregat pro Servicekraft** (Live **und** Kassenberichte): neues Breakdown-Feld
  `stornierungenProServicekraft` (`userId`, `userName`, `name`, `anzahlStornierungen`,
  `stornierungenCents`), im Backend aus den vorhandenen `StornierungDetail`-Zeilen
  aggregiert (Application-Schicht `api/reporting/application`, keine neue SQL-Query nötig).
  Anzeige: roter Marker „N Storno" an der Servicekraft-Zeile + Aggregat-Zeile im
  Stornierungen-Block; die bestehende Detail-Liste (Audit: wer/wann/warum/Bar-Rückgabe)
  bleibt unverändert darunter.
- **Umsatz pro Tisch entfällt ersatzlos** (Live-Tab, Kassenberichte, Backend-Query
  `GetUmsatzProTisch`/DTOs, Domain-Typ `reporting.UmsatzTisch`). Anforderung R-03
  wird gestrichen; Entscheidung als ADR `docs/adrs/02_umsatz-pro-tisch-entfernen.md`
  festgehalten (Präzedenz: ADR 01). „Offene Tische" (Salden) ist davon unberührt und wird
  aufgewertet.
- **Sortierung Offene Tische**: `GetOffeneTischeDetails` sortiert nach `saldo_cents DESC`
  (statt `ORDER BY t.name`, das „Tisch 11" vor „Tisch 2" einsortiert) — größte offene
  Beträge zuerst.
- **Live-Aktualität**: `useLiveReporting` bekommt `refetchInterval: 30_000`; der Seitenkopf
  zeigt „Stand HH:MM" (aus `dataUpdatedAt` von React Query) und einen manuellen
  Refresh-Button.
- **Ein gemeinsames Storno-Item**: eine Komponente `StornoItem` für Live und Kassenberichte
  (heute zwei abweichende Layouts). Zeitstempel als `HH:MM` (Kassensitzung = ein Tag),
  Preise mit `whitespace-nowrap`.
- **Steuersatz-Tabelle mobil**: gestapelte Blöcke je Steuersatz (Brutto/Netto/Steuer als
  beschriftete Zeilen) statt `min-w-[32rem]`-Tabelle mit horizontalem Scroll.

## Inventory

- `frontend/src/admin/reporting/AdminDashboardPage.tsx — AdminDashboardPage()` — heutige
  Kombi-Seite (Live + Historisch + Alerts); wird zum reinen Live-Dashboard.
- `frontend/src/admin/reporting/LiveReportingSection.tsx — LiveReportingSection()` — Tabs,
  Tiles, Progressbars, Zahlungen-Badges, Offene-Tische-Chips, Servicekraft-/Storno-Listen.
- `frontend/src/admin/reporting/ReportingResults.tsx — ReportingResults()` — historische
  Tabs inkl. Steuersatz-Tabelle; zieht auf die Kassenberichte-Seite um.
- `frontend/src/admin/reporting/ReportingFilter.tsx — ReportingFilter()` — Kassensitzungs-
  Select (`w-72`, Label-Truncation).
- `frontend/src/admin/reporting/hooks.ts — useLiveReporting(), useKassensitzungen(),
  useReport(), useDsfinvkExport()` — React-Query-Hooks (kein `refetchInterval`).
- `frontend/src/admin/reporting/utils.ts — pct(), formatOffeneArbeit(), formatLocalTime()`
  — nach dem Umbau teilweise tot.
- `frontend/src/admin/reporting/types.ts — ServicekraftLiveSchema, UmsatzTischSchema,
  LiveReportingDataSchema, ReportingDataSchema` — Zod-Schemas, beide Seiten anpassen.
- `frontend/src/admin/AdminSidebar.tsx — reportingItems` — Sidebar-Eintrag „Dashboard".
- `frontend/src/routes.ts` — Admin-Kinderrouten, Index-Redirect auf `auswertung`.
- `frontend/src/components/ui/progress.tsx — Progress()` — `value || 0` → Indikator bei
  fehlendem `value` unsichtbar (heutiger „Ladebalken").
- `backend/api/reporting/application/query.go — GetLiveReporting(), GetReporting(),
  mergeServicekraefteLive()` — Assemblierung der Breakdowns; Ort für Storno-Aggregat.
- `backend/api/reporting/http/query_handler.go — GetAllKassensitzungenHandler(),
  GetLiveReportingHandler(), GetReportingHandler()` — Response-DTOs (`json`-Tags).
- `backend/api/admin.go` — Routenregistrierung `/get-all-kassensitzungen` etc.
- `backend/sqlc/queries/reporting.sql — GetOffeneTischeDetails, Umsatz-pro-Tisch- und
  Umsatz-pro-Servicekraft-Queries`; `backend/sqlc/queries/kassensitzungen.sql —
  GetAllKassensitzungen`.
- `backend/domain/reporting/reporting.go — LiveReportingData, ReportingData, UmsatzTisch,
  ServicekraftLive`; `backend/domain/kasse/offene_arbeit.go — ComputeEigeneArbeitAnTisch(),
  ComputeOffeneArbeitProServicekraft()`.
- `backend/repository/reporting_repo/repo.go` — Repository hinter den Queries.
- Tests: `frontend/src/admin/reporting/*.test.tsx`,
  `backend/api/reporting/application/query_test.go`,
  `e2e/tests/admin-live-reporting.spec.ts`, `e2e/tests/admin-reporting-fehlerpfade.spec.ts`,
  `e2e/tests/admin-dsfinvk-export.spec.ts`.
- Doku: `docs/anforderungen.md` (R-01, R-03, R-07), `docs/handbuch.md` (Read Models),
  `docs/adrs/README.md`.

## Resolved decisions

- **Zwei Routen** statt einer Kombi-Seite (Nutzerentscheidung): Live-Dashboard bleibt
  Landing; Kassenberichte separat.
- **Single Scroll ohne Tabs** auf dem Live-Dashboard (Nutzerentscheidung).
- **Umsatz pro Tisch überall löschen** — live und historisch, inkl. Backend
  (Nutzerentscheidung; konservativste Variante).
- **Bestellungen- und Direktverkauf-Tiles bleiben live sichtbar** (Nutzerentscheidung):
  der Kassenwart will den bestellten/kommenden Umsatz inkl. unkassierter Positionen sehen.
  Konsequenz: Sub-Labels stellen den Zusammenhang klar (Bestellungen = „Bestellwert, inkl.
  noch nicht kassiert"; Gesamtumsatz = „kassiert, abzüglich Warenrücknahmen"), damit die
  drei Geldzahlen (bestellt / kassiert / offen) nicht wie ein Widerspruch wirken.
- Refresh-Intervall 30 s (Abwägung Aktualität vs. Last; lokaler Server, eine Admin-Session).
- Kassenberichte zeigen nur `abgeschlossen`; `wird_abgeschlossen` ist ein transienter
  Barrierestatus und erscheint nirgends.
- Offene Arbeit pro Servicekraft wird als **ein Euro-Betrag + Tischnamen inline** angezeigt
  (keine Chips, kein „N zu kassieren" — Positionszählung entfällt aus der Admin-Sicht).
- Stornierungen pro Servicekraft ist die **einzige funktionale Ergänzung** des Umbaus;
  Datenbasis existiert bereits (`StornierungDetail.userId`).

## Open questions / Risks

- `e2e/tests/admin-live-reporting.spec.ts` navigiert heute über `getByRole('tab', …)` —
  der Tab-Wegfall (Phase 4) erfordert ein Umschreiben der Spec auf die Blockstruktur, nicht
  nur Selektor-Kosmetik. Eingeplant in Phase 4.
- Das additive Domain-Feld `OffenCents` berührt `EigeneUebersicht` (Service-Dashboard,
  R-06) nicht funktional; dessen UI bleibt bewusst unangetastet (Scope Guard).
- R-04 (Abrechnung pro Servicekraft) bleibt bestehen — nur R-03 (pro Tisch) entfällt.

---

## Phase 1: Mobile-Politur des bestehenden Dashboards

### Context

- `frontend/src/admin/reporting/LiveReportingSection.tsx — LiveReportingSection()` —
  Progressbars, Badges, Header-Zeile, Storno-Liste, Ladezustand.
- `frontend/src/admin/reporting/ReportingResults.tsx — ReportingResults()` — dieselben
  Muster historisch + Steuersatz-Tabelle.
- `frontend/src/admin/reporting/ReportingFilter.tsx — ReportingFilter()` — `w-72`-Select.
- `frontend/src/admin/reporting/AdminDashboardPage.tsx — AdminDashboardPage()` — zwei
  großflächige Alert-Banner vor dem Inhalt.
- `frontend/src/admin/reporting/utils.ts — pct(), formatLocalTime()`.

### What to build

Rein visuelle/deklarative Frontend-Bereinigung der bestehenden Seite, ohne Struktur- oder
API-Änderung — der Praxistest-Kern („keine Progressbars, keine ‚17 Zahlungen'") wird sofort
erfüllt:

- Alle `<Progress>`-Balken aus Servicekraft- und Tisch-Zeilen (live + historisch) entfernen;
  `pct()` löschen.
- Alle „N Zahlungen"-Badges (live + historisch, Servicekräfte + Tische) entfernen.
- Ladezustand: `Loader2`-Spinner (Muster aus `ReportingFilter`) statt wertlosem
  `<Progress>` ohne `value`.
- Seitenkopf einzeilig kompakt („Live-Dashboard" + Datum/Bezeichnung gestapelt statt
  4-zeilig umbrechend).
- Die beiden Alert-Banner (Drucker/TSE) auf je eine kompakte Zeile reduzieren, damit die
  ersten Kennzahlen bei 375×667 in den ersten Viewport rücken.
- Storno-Liste: Zeit als `HH:MM` (statt Datum + Sekunden), Positionspreise mit
  `whitespace-nowrap` (kein „€"-Orphan), „Fertig"-Grün über das Theme-Token `text-primary`
  (Primary ist in beiden Themes das Marken-Grün) statt hartem `text-green-700`.
- Kassensitzungs-Select: volle Breite auf Mobil statt `w-72` (kein Label-Truncating).
- Steuersatz-Tabelle: gestapelte Blöcke je Steuersatz auf Mobil (Brutto/Netto/Steuer als
  beschriftete Zeilen), kein horizontaler Scroll.

### Acceptance criteria

- [x] Bei 375×667 mit offener Kassensitzung: keine Progressbar und kein „Zahlungen"-Badge
      mehr im gesamten Admin-Dashboard (live + historisch).
- [x] Erste Kennzahlen-Kachel ist bei 375×667 trotz aktiver Drucker- und TSE-Warnung ohne
      Scrollen sichtbar.
- [x] Ladezustand zeigt einen sichtbaren, animierten Spinner.
- [x] Steuersatz-Tabelle zeigt bei 375 px alle Beträge (Brutto/Netto/Steuer) ohne
      horizontalen Scroll.
- [x] Storno-Einträge zeigen `HH:MM`; kein Preis bricht zwischen Betrag und „€" um.
- [x] Kassensitzungs-Select zeigt Datum + Bezeichnung ungekürzt (typische Länge wie
      „11.07.2026 (Sommerfest 26 Sonntag)").
- [x] `make check` grün; betroffene Vitest-Specs (`LiveReportingSection.test.tsx`,
      `ReportingResults.test.tsx`, `AdminDashboardPage.test.tsx`) angepasst und grün.

---

## Phase 2: Umsatz pro Tisch ersatzlos entfernen

### Context

- `frontend/src/admin/reporting/LiveReportingSection.tsx — LiveReportingSection()` —
  Tab „Tische" (live).
- `frontend/src/admin/reporting/ReportingResults.tsx — ReportingResults()` — Tab „Tische"
  (historisch).
- `frontend/src/admin/reporting/types.ts — UmsatzTischSchema` und `umsatzProTisch` in
  beiden Data-Schemas.
- `backend/domain/reporting/reporting.go — UmsatzTisch` + Breakdown-Felder.
- `backend/sqlc/queries/reporting.sql — GetUmsatzProTisch`;
  `backend/repository/reporting_repo/repo.go`.
- `backend/api/reporting/http/query_handler.go` — DTO-Felder.
- `docs/anforderungen.md` — R-03; `docs/handbuch.md` — Read-Model-Beschreibung;
  `docs/adrs/README.md` + `docs/adrs/01_ausgabe-bestaetigen.md` als Format-Vorlage.

### What to build

Der komplette vertikale Schnitt „Umsatz pro Tisch" verschwindet aus dem Produkt: beide
Frontend-Tabs, Zod-Schemas, Response-DTOs, Domain-Typ, Repository-Methode und SQL-Query
(danach `make sqlc`). „Offene Tische" (Salden) bleibt unberührt. Dazu die Dokumentation:
R-03 aus `docs/anforderungen.md` streichen (inkl. Verweis in der Reporting-Tabelle),
Read-Model-Beschreibung im `docs/handbuch.md` anpassen und die Produktentscheidung als
`docs/adrs/02_umsatz-pro-tisch-entfernen.md` festhalten (Begründung: Praxistest-Feedback,
beantwortet keine Kassenwart-Frage; Präzedenz ADR 01).

### Acceptance criteria

- [x] Kein Tab/Abschnitt „Tische" mehr im Live-Dashboard und in der historischen
      Auswertung; „Offene Tische" (Salden) unverändert vorhanden.
- [x] `grep -ri "umsatzProTisch\|UmsatzTisch" backend/ frontend/src/` liefert keine Treffer
      mehr (außer ggf. Git-Historie/Plan-Datei).
- [x] Live- und Reporting-Response enthalten das Feld `umsatzProTisch` nicht mehr; die
      Zod-Schemas validieren die neuen Responses.
- [x] R-03 ist aus `docs/anforderungen.md` entfernt, `docs/handbuch.md` beschreibt das
      Read Model ohne Tisch-Umsatz, ADR 02 existiert und ist im ADR-README indexiert.
- [x] `make check` und Backend-Integrationstests (`make verify`) grün; E2E-Specs
      (`admin-live-reporting.spec.ts`, `admin-reporting-fehlerpfade.spec.ts`) angepasst.

---

## Phase 3: Kassenberichte als eigene Seite

### Context

- `frontend/src/routes.ts` — Admin-Kinderrouten (lazy-Muster der bestehenden Einträge).
- `frontend/src/admin/AdminSidebar.tsx — reportingItems` — heute ein Eintrag „Dashboard".
- `frontend/src/admin/reporting/AdminDashboardPage.tsx — AdminDashboardPage()` — enthält
  heute Filter + Export + `ReportingResults` unterhalb des Live-Teils.
- `frontend/src/admin/reporting/ReportingBackend.ts — getAllKassensitzungen()`.
- `backend/sqlc/queries/kassensitzungen.sql — GetAllKassensitzungen`;
  `backend/api/reporting/http/query_handler.go — GetAllKassensitzungenHandler()`;
  `backend/api/admin.go` — Routenregistrierung.
- Storno-Renderer in `LiveReportingSection.tsx` und `ReportingResults.tsx` (zwei Layouts).

### What to build

Neue Seite `/admin/kassenberichte` („Kassenberichte" als H1) mit Kassensitzungs-Select,
DSFinV-K-Export und der historischen Auswertung; das Live-Dashboard verliert den kompletten
„Historische Auswertung"-Abschnitt. Backend: neue sqlc-Query
`GetAbgeschlosseneKassensitzungen` (`WHERE status = 'abgeschlossen'`, Sortierung wie bisher),
Endpoint umbenennen zu `/admin/get-abgeschlossene-kassensitzungen`, alte Query/Route
entfernen. Sidebar: Einträge „Live-Dashboard" und „Kassenberichte" unter „Auswertungen".
Kennzahlen-Kacheln der Kassenberichte in die kanonische Reihenfolge bringen (Gesamtumsatz →
Bestellungen → Direktverkauf → Stornierungen). Gemeinsame `StornoItem`-Komponente
extrahieren und in beiden Seiten verwenden (ein Layout statt zwei). Leerer Zustand der
Kassenberichte (noch keine abgeschlossene Kassensitzung) mit Hinweis + Link zur
Kasse-Seite.

### Acceptance criteria

- [ ] `/admin/auswertung` zeigt ausschließlich Live-Inhalte; `/admin/kassenberichte` zeigt
      Select (nur abgeschlossene Sitzungen), Export und Auswertung.
- [ ] Bei offener Kassensitzung erscheinen deren Zahlen **nur** auf dem Live-Dashboard —
      nirgendwo doppelt.
- [ ] Sidebar zeigt unter „Auswertungen" die Einträge „Live-Dashboard" und
      „Kassenberichte"; aktiver Zustand stimmt je Route.
- [ ] Backend liefert unter `/admin/get-abgeschlossene-kassensitzungen` nur
      `status = 'abgeschlossen'`; der alte Endpoint existiert nicht mehr.
- [ ] Storno-Einträge sehen auf beiden Seiten identisch aus (eine `StornoItem`-Komponente).
- [ ] Kassenberichte ohne abgeschlossene Sitzung zeigen einen erklärenden leeren Zustand.
- [ ] `make verify` grün; `admin-dsfinvk-export.spec.ts` und
      `admin-reporting-fehlerpfade.spec.ts` auf die neue Route umgestellt.

---

## Phase 4: Live-Dashboard als Single-Scroll-Seite

### Context

- `frontend/src/admin/reporting/LiveReportingSection.tsx — LiveReportingSection()` — Tabs,
  Tile-Grid, Offene-Tische-Chips, Servicekraft-Zeilen.
- `frontend/src/admin/reporting/hooks.ts — useLiveReporting()` — kein `refetchInterval`.
- `frontend/src/admin/reporting/utils.ts — formatOffeneArbeit()` — wird ersetzt.
- `backend/sqlc/queries/reporting.sql — GetOffeneTischeDetails` — `ORDER BY t.name`.
- `backend/domain/kasse/offene_arbeit.go — ComputeEigeneArbeitAnTisch(),
  ComputeOffeneArbeitProServicekraft()` — Positionszählung ohne Cents.
- `backend/api/reporting/application/query.go — mergeServicekraefteLive()`.
- `e2e/tests/admin-live-reporting.spec.ts` — navigiert über Tabs.

### What to build

Das Live-Dashboard wird eine scrollbare Seite ohne Tabs, Blockreihenfolge: kompakte
Systemwarnungen → Kennzahlen (kanonische Reihenfolge: Gesamtumsatz „kassiert, abzüglich
Warenrücknahmen" → Offene Saldi → Bestellungen „Bestellwert, inkl. noch nicht kassiert" →
Direktverkauf → Stornierungen) → Offene Tische (nach Saldo absteigend, Backend-Sortierung)
→ Stornierungen (Detail-Liste mit `StornoItem`) → Servicekräfte.

Servicekraft-Zeile neu: Name + kassiert € rechtsbündig; darunter grünes „Fertig" oder
„Offen: X € · Tisch 11, Zelt A2". Dafür Domäne additiv erweitern: `EigeneArbeitAnTisch` und
`OffeneArbeitTisch` bekommen `OffenCents` (Summe `EinzelpreisCents × Menge` der eigenen
unbezahlten Positionen); Application-Schicht und DTO/Zod-Schema reichen den Betrag durch.
`formatOffeneArbeit()` („N zu kassieren") entfällt.

Aktualität: `refetchInterval: 30_000` im Live-Hook, „Stand HH:MM" aus `dataUpdatedAt` im
Seitenkopf, manueller Refresh-Button daneben.

### Acceptance criteria

- [ ] Live-Dashboard hat keine Tabs; alle vier Blöcke sind per Scroll erreichbar, Reihenfolge
      wie oben.
- [ ] Kennzahlen-Reihenfolge beginnt mit Gesamtumsatz; Sub-Labels benennen den Zusammenhang
      bestellt/kassiert/offen wie festgelegt.
- [ ] Offene Tische sind absteigend nach Saldo sortiert (Backend-Response-Reihenfolge; kein
      Frontend-Sort).
- [ ] Servicekraft-Zeile zeigt „Fertig" oder „Offen: X € · <Tischnamen>"; Betrag entspricht
      der Summe der eigenen unbezahlten Positionen (Backend-Unit-Test auf `OffenCents`).
- [ ] Live-Daten aktualisieren sich ohne Interaktion (30-s-Intervall); „Stand HH:MM" und
      Refresh-Button sind vorhanden und funktionieren.
- [ ] `make verify` grün; `admin-live-reporting.spec.ts` auf die Blockstruktur umgeschrieben
      (keine Tab-Selektoren mehr).

---

## Phase 5: Stornierungen pro Servicekraft

### Context

- `backend/api/reporting/application/query.go — GetLiveReporting(), GetReporting()` —
  hat die `StornierungDetail`-Liste (inkl. `UserID`, `UserName`, `Name`, `BetragCents`)
  bereits im Zugriff.
- `backend/domain/reporting/reporting.go — LiveReportingData, ReportingData` — Breakdowns.
- `backend/api/reporting/http/query_handler.go` — Response-DTOs.
- `frontend/src/admin/reporting/types.ts — ServicekraftLiveSchema,
  LiveReportingDataSchema, ReportingDataSchema`.
- Live-Servicekraft-Zeile und Stornierungen-Block aus Phase 4; Kassenberichte-Seite aus
  Phase 3.

### What to build

Backend aggregiert die vorhandene Storno-Detail-Liste pro Servicekraft (Application-Schicht,
kein neues SQL): neues Breakdown-Feld `stornierungenProServicekraft` mit `userId`,
`userName`, `name`, `anzahlStornierungen`, `stornierungenCents` — in Live- und
Reporting-Response identisch. Frontend: auf dem Live-Dashboard erhält jede
Servicekraft-Zeile mit Stornos einen roten Marker („2 Storno"), und der Stornierungen-Block
beginnt mit der Aggregat-Zeile („felix 1 · sophie 1") vor der Detail-Liste. In den
Kassenberichten erscheint dasselbe Aggregat im Servicekräfte-Abschnitt (Spalte/Marker) und
über der Storno-Detail-Liste. Null-Stornos bleiben unmarkiert (kein „0 Storno"-Rauschen).

### Acceptance criteria

- [ ] Live- und Reporting-Response enthalten `stornierungenProServicekraft`; Summe über
      alle Servicekräfte entspricht `anzahlStornierungen`/`gesamtStornierungenCents` der
      Summary (Backend-Unit-Test).
- [ ] Servicekraft-Zeilen mit ≥ 1 Storno zeigen live und in den Kassenberichten einen roten
      „N Storno"-Marker; Zeilen ohne Storno zeigen nichts.
- [ ] Der Stornierungen-Block zeigt vor der Detail-Liste das Aggregat pro Servicekraft;
      die Detail-Liste (wer/wann/warum/Bar-Rückgabe/Positionen) bleibt unverändert.
- [ ] Zod-Schemas validieren die erweiterten Responses; `make verify` grün; E2E-Spec des
      Live-Dashboards prüft den Storno-Marker mit Seed-Daten.
