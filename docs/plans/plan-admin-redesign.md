# Plan: Admin-UI-Redesign „Ein Festtag, ein roter Faden"

> Source PRD: `docs/prds/prd-admin-redesign.md`
> Visuelle Spezifikation: `docs/prds/design_handoff_admin_redesign/` (README plus
> `Jotti Admin Review.dc.html`, Abschnitte 0 und 1a bis 1h). Bei
> Aussehens-Detailfragen gilt das Handoff, bei Scope-Fragen die PRD. Vier
> Handoff-Inhalte sind per PRD abgewählt und dürfen nicht umgesetzt werden:
> Serienanlage für Tische (1d), Druckstations-Status (1g), Bon-Klartext an
> fehlgeschlagenen Druckaufträgen (1g), 60-s-Refresh (es bleibt bei 30 s).

## Goal

Der komplette Admin-Bereich wird nach dem Design-Handoff umgebaut: Sidebar
nach Festablauf mit globalem Status, einheitlicher Seitenkopf statt FAB, acht
Seiten-Redesigns, Deaktivieren vor Löschen mit echten Backend-Schutzregeln,
plus gezielte additive Backend-Erweiterungen (Kassenbestand-Aufschlüsselung,
Geldtransit-Liste, Sitzungs-Metadaten, ELSTER-Flag, Testdruck, Listen-Flags
`hatVerkaeufe`/`saldoCents`). Routen, Guards, Event-Formate und die
Ubiquitous Language bleiben unverändert.

## Architectural decisions

Durable Entscheidungen, die für alle Phasen gelten:

- **Frontend-Routen unverändert**: `/admin/auswertung`, `/admin/kassenberichte`,
  `/admin/produkte`, `/admin/tische`, `/admin/benutzer`, `/admin/kasse`,
  `/admin/druckstationen`, `/admin/finanzamt`, `/admin/tse-einrichtung`
  (`frontend/src/routes.ts`). Nur Sidebar-Labels und -Gruppen ändern sich.
- **Neue gemeinsame Bausteine** unter `frontend/src/admin/components/`:
  - `AdminPageHeader` (H1 + erklärende Unterzeile + Aktions-Slot rechts),
  - `HinweisKarte` (Info-Variante, `bg-sidebar border rounded-lg` mit `Info`-Icon),
  - `WarnKarte` (Destructive-Variante, Rahmen `destructive/40`, Fläche `destructive/4`),
  - `StatusDot` (7-px-Punkt, `--primary` ok / `--destructive` Fehler / muted neutral).
  Kein weiteres Komponenten-Framework. Zählhilfe separat:
  `frontend/src/admin/kasse/ZaehlhilfeDialog.tsx` mit reiner Funktion
  `summeAusStueckzahlen` in `frontend/src/admin/kasse/zaehlhilfe.ts`.
- **Neue/erweiterte Admin-Endpunkte** (alle POST unter `/api/admin`, registriert
  in `backend/api/admin.go`; Antwortfelder deutsch-camelCase nach
  `docs/language.md`):

  | Endpunkt | Änderung |
  | --- | --- |
  | `/get-kassenbestand` | Antwort zusätzlich `anfangsbestandCents`, `bareinnahmenCents`, `einlagenCents`, `entnahmenCents`; bestehendes Summenfeld bleibt. Invariante: Anfangsbestand + Bareinnahmen + Einlagen − Entnahmen = Soll-Bestand (vor Kassensturz). |
  | `/get-geldtransit-liste` (neu) | Request `{kassensitzungNr}`, Antwort Liste mit `zeitpunkt`, `richtung` (`einlage`/`entnahme`), `betragCents`, `kommentar`, `gebuchtVon` (Anzeigename aus `kassenjournal.user_name`). Reine Projektion der `geldtransit-gebucht:v1`-Events. |
  | `/get-abgeschlossene-kassensitzungen` | je Eintrag zusätzlich `umsatzGesamtCents`, `abgeschlossenAm` (aus dem `tagesabschluss-erstellt:v1`-Event). |
  | `/get-abrechnung` | zusätzlich `eroeffnetAm`, `abgeschlossenAm`, `abgeschlossenVon`, `kassensturzDifferenzCents` (aus Eröffnungs-, Tagesabschluss- und Kassensturz-Events). |
  | `/get-all-produkte` | je Produkt zusätzlich `hatVerkaeufe` (bool). |
  | `/get-all-tische` | je Tisch zusätzlich `saldoCents` (int; 0 ohne offenen Saldo oder ohne offene Kassensitzung). |
  | `/testbon-drucken` (neu) | Request `{kategorie}`; erzeugt regulären Druckauftrag über die Outbox. |
  | `/elster-meldung-setzen`, `/elster-meldung-zuruecknehmen` (neu) | Betreiber-Kontext; `/get-betreiber` liefert zusätzlich `elsterGemeldetAm` (Datum oder null). |

- **Neue Fehlercodes**: `produkt_hat_verkaeufe` (Produkt löschen),
  `tisch_saldo_offen` (Tisch deaktivieren/löschen),
  `druckstation_nicht_konfiguriert` (Testdruck ohne Drucker-IP). Alle über
  `helper.MapError` als Client-Fehler plus deutsche Meldung in
  `frontend/src/lib/errorMessages.ts`.
- **Schema** (additive Migrationen nach `database/migrations/README.md`;
  Nummern = Stand bei Planerstellung, maßgeblich ist die nächste freie
  Nummer): `04_testbon_bonart.up.sql` erweitert den CHECK auf
  `druckauftraege.bon_art` um den Wert `testbon`;
  `05_elster_gemeldet_am.up.sql` ergänzt `betreiber` um die nullbare Spalte
  `elster_gemeldet_am DATE`.
- **Keine neuen Event-Typen, keine Event-Format-Änderungen.** Alle neuen
  Queries sind Projektionen über vorhandene Journal-Events; der
  Contract-Guard `backend/domain/kasse/event_json_contract_test.go` bleibt
  unberührt.
- **Schutzregel-Semantik**: „Offener Saldo" eines Tischs bezieht sich auf die
  aktuell offene Kassensitzung (`tisch_sessions` der offenen Sitzung,
  `saldo_cents > 0`). Ohne offene Sitzung ist kein Tisch geschützt.
  „Je verkauft" für Produkte heißt: eine Variante des Produkts kommt in
  mindestens einem `bestellung-aufgenommen:v1`- oder
  `direktverkauf-getaetigt:v1`-Event vor (kassensitzungsübergreifend).
- **Sidebar-Status** speist sich ausschließlich aus vorhandenen Queries
  (`useOffeneKassensitzung`, `useFehlgeschlageneDruckauftraege`,
  `useTSEStatus`, `useTSESignaturQueue`); kein neues Polling, TanStack Query
  dedupliziert.
- **Druck des Tagesberichts** über `window.print` plus Tailwind-`print:`-
  Varianten (Sidebar/Layout/Sitzungsliste `print:hidden`, nur die
  Berichtsspalte druckt); kein separates Stylesheet-File.

## Inventory

Frontend:

- `frontend/src/routes.ts` — Admin-Routen, `AdminGuard`
- `frontend/src/admin/AdminLayout.tsx` — Layout mit Sidebar + Mobile-Header
- `frontend/src/admin/AdminSidebar.tsx` — Sidebar (Gruppen als Modul-Arrays, `NavGroup`-Helper); Test `AdminSidebar.test.tsx`
- `frontend/src/admin/adminListLayout.ts` — `adminListBottomClearance` (FAB-Scroll-Clearance, entfällt), `adminItemActionButton`
- `frontend/src/admin/reporting/` — `AdminDashboardPage.tsx` (Warn-Banner, `RUECKSTAND_WARN_SEKUNDEN`), `LiveReportingSection.tsx`, `SummaryCard.tsx`, `StornoItem.tsx`, `StornoServicekraft.tsx` (`StornoMarker`, `StornoAggregat`), `KassenberichtePage.tsx`, `ReportingResults.tsx`, `hooks.ts` (`useLiveReporting` mit `refetchInterval: 30_000` und `dataUpdatedAt`, `useAbgeschlosseneKassensitzungen`, `useReport`, `useDsfinvkExport`), `utils.ts` (`formatStand`), `ReportingBackend.ts`
- `frontend/src/admin/products/` — `AdminProductsPage.tsx`, `Products.tsx`, `ProductItem.tsx`, Dialoge (`NewProductDialog`, `EditProductDialog`, `NewVariantDialog`, `EditVariantDialog`), `ProduktBackend.ts`, `Produkt.ts` (`STEUERSATZ_LABEL`), `hooks.ts` (`useAllProdukte`)
- `frontend/src/admin/tables/` — `AdminTablesPage.tsx`, `Tische.tsx`, `TischItem.tsx`, `NewTischDialog.tsx`, `TischBackend.ts`, `hooks.ts` (`useAllTische`)
- `frontend/src/admin/users/` — `AdminUsersPage.tsx`, `Users.tsx`, `UserItem.tsx`, `NewUserDialog.tsx`, `UserBackend.ts` (inkl. `resetPassword`), `hooks.ts` (`useAllUsers`)
- `frontend/src/admin/kasse/` — `KassensitzungPage.tsx`, `KasseAbschliessenSection.tsx` (Ist-Bestand-`EuroField`, Bestätigungs-AlertDialog, `signaturen_ausstehend`-Retry), `KasseBackend.ts`, `hooks.ts` (`useOffeneKassensitzung`, `useKassenbestand`)
- `frontend/src/admin/settings/` — `DruckstationConfigPage.tsx` (inkl. `AlleVerwerfenDialog`), `DruckstationBackend.ts` (`hatBonmodus`, `validateDruckerIp`, `formatDruckauftragReferenz`, `REFERENZ_PRAEFIX_LABEL`), `hooks.ts` (`useDruckstationen`, `useFehlgeschlageneDruckauftraege`)
- `frontend/src/admin/finanzamt/` — `FinanzamtPage.tsx` mit `BetreiberSection.tsx`, `KassenidentitaetSection.tsx`, `SignaturauftraegeSection.tsx`, `TSEAusfalldokumentationSection.tsx`, `TSEAnbindungSection.tsx`, `DokumenteUndPflichtenSection.tsx`; `BetreiberBackend.ts`, `hooks.ts` (`useBetreiber`, `useKassenidentitaet`)
- `frontend/src/admin/tse/hooks.ts` — `useTSEStatus`, `useTSESignaturQueue`, `useTSEStoerungen`
- `frontend/src/admin/hooks.ts` — `useVersion` (Sidebar-Footer)
- `frontend/src/lib/Backend.ts` (`BackendClient`, `BackendError`), `frontend/src/lib/errorMessages.ts` (`commonErrorMessages`, `getActionErrorMessage`)
- `frontend/src/hooks/use-action-submit.ts`, `frontend/src/hooks/use-form-action-submit.ts`
- `frontend/src/components/common/FormFields.tsx` (`EuroField`), `frontend/src/components/common/EmptyState.tsx`, `frontend/src/components/ui/` (u. a. `dropdown-menu`, `tooltip`, `collapsible`, `switch`, `badge`, `alert-dialog`, `empty`)
- `frontend/src/index.css` — Design-Token (Light/Dark)
- Test-Muster: Seitentests mocken die Query-Hooks per `vi.mock('./hooks', …)` + `vi.hoisted` (siehe `KassensitzungPage.test.tsx`, `AdminDashboardPage.test.tsx`); `sonner` global gemockt

Backend:

- `backend/api/admin.go` — `NewAdminApi` (Routen-Registrierung)
- `backend/api/reporting/application/query.go` — `Query.GetLiveReporting`, `GetReporting`, `GetAbgeschlosseneKassensitzungen`; HTTP-DTOs in `backend/api/reporting/http/query_handler.go` (`kassensitzungItem`, `reportingResponse`, `offenerTischResponse`, `servicekraftLiveResponse`)
- `backend/repository/reporting_repo/repo.go` + `backend/sqlc/queries/reporting.sql` (`GetReportingStats`, `GetOffeneSaldi`, `GetOffeneTischeDetails`, `GetUmsatzProServicekraft`, `GetStornierungen`)
- `backend/api/kasse/kassenfuehrung/application/` — `command.go` (`KassensitzungEroeffnen`, `GeldtransitBuchen`, `KasseAbschliessen`), `query.go` (`GetKassenbestand`, `GetOffeneKassensitzung`), `errors.go`; HTTP-Schemas in `http/command_handler.go` (`geldtransitBuchenSchema`)
- `backend/repository/kassenjournal_repo/repo.go` — `GetKassenbestand`; SQL in `backend/sqlc/queries/kassensitzungen.sql` mit den `kj_extract_*`-Funktionen aus `database/migrations/01_initial.up.sql` (`kj_extract_eroeffnung_cents`, `kj_extract_zahlung_cents`, `kj_extract_direktverkauf_cents`, `kj_extract_direktverkauf_storno_cents`, `kj_extract_stornierung_cents`, `kj_extract_geldtransit_cents`, `kj_extract_differenz_cents`)
- `backend/domain/kasse/` — `kassensitzung_events.go` (`GeldtransitGebuchtV1Data`, `KassensturzDurchgefuehrtV1Data`, `TagesabschlussErstelltV1Data`), `bestellung.go` (`PositionEventData` mit `VarianteID`), `event_json_contract_test.go`
- `backend/api/stammdaten/produkt/application/command.go` (`DeleteProdukt`), `query.go` (`GetAllProducts`); DTOs in `http/query_handler.go`
- `backend/api/stammdaten/tisch/application/command.go` (`TischDeaktivieren`, `TischLoeschen` via `applyTischStatusChange`), `errors.go`
- `backend/api/stammdaten/user/http/command_handler.go` — Selbstlösch-Schutz (`cannot_delete_self`)
- `backend/api/stammdaten/betreiber/` — `http/query_handler.go` (`betreiberResponse`), `http/command_handler.go` (`updateBetreiberSchema`); Domain `backend/domain/betreiber/betreiber.go`; Tabelle `betreiber` (Singleton, id=1)
- `backend/api/druck/` — `station/http/handler.go` (Druckstationen-Update inkl. Bonmodus-Validierung), `auftrag/application/` (`GetFehlgeschlageneDruckauftraege`, `RetryDruckauftrag`, `DiscardDruckauftrag`, `DiscardAlleFehlgeschlagenen`), `bondruck/application/escpos/formatter.go` (`FormatKassenbeleg` u. a.), `beleg/application/kassenbeleg_command.go` als Vorbild für Outbox-Enqueue (`druckauftrag_repo.EnqueueDruckauftraege`, `NeuerDruckauftrag`)
- `backend/domain/druckstation/druckstation.go` — `Kategorie` (`essen`, `getraenk`, `sonstiges`, `kassenbeleg`, `abholbon`), `HatBonmodus`, `Bonmodus`
- `backend/api/fiskal/setup/application/query.go` (`GetTSEStatus` → `TSEStatus{Umgebung, IstKonfiguriert}`), `backend/api/fiskal/signatur/application/query.go` (`GetTSESignaturQueueZustand`, `GetTSEStoerungen`)
- Test-Muster: Handler-Tests neben dem Handler (`query_handler_test.go`), Unit-Tests mit Build-Tag `unit`, Integrationstests mit Build-Tag `integration`; Journal-Fixtures in `backend/repository/kassenjournal_repo/repo_test.go` (`newTestEvent`, `insertEventRaw`, `validEroeffnungData` u. a.) und `backend/repository/reporting_repo/repo_test.go` (`insertEvent`)
- `database/migrations/` — vorhanden: `01_initial`, `02_druckauftrag_backoff`, `03_ausgabe_entfernen`; Regeln in `README.md`

## Resolved decisions

Die PRD hat den Scope bereits fixiert. Folgende Restentscheidungen wurden bei
der Planung getroffen (autonome Planung, daher als dokumentierte Annahmen):

- **Assumption: Ein Plan-Dokument mit 12 Phasen** statt mehrerer Plan-Dateien
  (die PRD erwartet „mehrere Implementierungspläne"). Jede Phase ist
  unabhängig shipbar; das erfüllt denselben Zweck und folgt der Repo-Praxis
  (Service-Redesign). Bei Bedarf lässt sich der Plan trivially splitten.
- **Assumption: Testbon bekommt die eigene `bon_art` `testbon`** (Migration
  erweitert den CHECK) statt `arbeitsbon` wiederzuverwenden; Referenz-Format
  `testdruck:<kategorie>`, Anzeige über eine Erweiterung von
  `formatDruckauftragReferenz`. Korrekter und in der Alarm-Karte nicht
  irreführend.
- **Assumption: `elsterGemeldetAm` wird serverseitig auf das aktuelle Datum
  gesetzt** (Befehl ohne Datums-Parameter); Zurücknehmen setzt NULL. Zwei
  getrennte Endpunkte analog zum `activate`/`deactivate`-Muster.
- **Assumption: `/get-geldtransit-liste` nimmt `kassensitzungNr`** (Symmetrie
  zu `/get-kassenbestand`) statt die offene Sitzung implizit aufzulösen.
- Saldo-Schutz und `saldoCents` beziehen sich auf die offene Kassensitzung
  (siehe Architectural decisions); Tische mit Rest-Saldo aus abgeschlossenen
  Sitzungen sind nicht gesperrt (sonst wären sie dauerhaft unlöschbar).
- Kein Guard für das Löschen einzelner Varianten: Events denormalisieren
  Name/Preis (Berichte bleiben korrekt), das Produkt selbst bleibt über
  `hatVerkaeufe` geschützt, Löschen ist ohnehin ein Soft-Delete. Nur der
  Produkt-Lösch-Guard aus der PRD wird umgesetzt.
- Leerzustände bleiben je Seite beim vorhandenen Muster (`EmptyState` bzw.
  `Empty`); die bestehende Uneinheitlichkeit der beiden Empty-Komponenten
  wird nicht bereinigt (PRD: „Leerzustände bleiben beim bestehenden
  Empty-Muster", Scope Guard).

## Open questions / Risks

- `hatVerkaeufe` prüft per EXISTS über JSONB-Positionen des Kassenjournals;
  bei Vereins-Skala unkritisch, wird aber bei jedem Laden der Produktliste
  ausgeführt. Falls messbar langsam: Prüfung nur im Lösch-Guard plus
  Fehlercode-Fallback im Frontend.
- Die Sidebar ruft die Status-Hooks künftig auf jeder Admin-Seite auf (bisher
  nur Dashboard/Finanzamt). Kein neues Polling, aber mehr parallele Queries
  beim Seitenwechsel; akzeptiert.
- `window.print` mit `print:`-Klassen braucht eine manuelle Sichtprüfung
  (Browser-Unterschiede beim Seitenumbruch); Teil der Abnahme in Phase 12.
- Die Checklisten-Bedingung „Vereinsdaten erledigt" (Phase 11) leitet sich
  aus den Pflichtfeldern der Betreiber-Query ab; es gibt bewusst kein neues
  „vollständig"-Flag im Backend.

---

## Phase 1: Sidebar nach Festablauf

**User stories**: 1, 2

### Context

- `frontend/src/admin/AdminSidebar.tsx` — Gruppen-Arrays und `NavGroup`; Umbau-Ziel
- `frontend/src/admin/AdminSidebar.test.tsx` — bestehender Test, wird erweitert
- `frontend/src/admin/hooks.ts — useVersion()` — Footer-Version
- `frontend/src/admin/kasse/hooks.ts — useOffeneKassensitzung()` — Event-Status-Chip und Kassentag-Punkt
- `frontend/src/admin/settings/hooks.ts — useFehlgeschlageneDruckauftraege()` — roter Punkt Bondrucker
- `frontend/src/admin/tse/hooks.ts — useTSEStatus(), useTSESignaturQueue()` — Warnpunkt Finanzamt
- `frontend/src/admin/reporting/AdminDashboardPage.tsx — RUECKSTAND_WARN_SEKUNDEN` — bestehende Schwelle (60 s) für den TSE-Fehlerzustand, wiederverwenden statt duplizieren
- Handoff Abschnitt 0 (`AdminSidebarNeu.dc.html`) — visuelle Spezifikation

### What to build

Die Sidebar wird nach Handoff-Abschnitt 0 umgebaut: Gruppen „Heute"
(Übersicht, Kassentag), „Vorbereitung" (Produkte & Preise, Tische, Helfer &
Zugänge, Bondrucker), „Nach dem Fest" (Berichte & Export, Finanzamt & TSE),
„Service" (Zum Service-Bereich); Footer mit Theme-Toggle, Abmelden und
Versionszeile. Im Header unter der Wortmarke ein Event-Status-Chip
(Bezeichnung der offenen Sitzung, „Kasse offen · seit HH:MM" mit grünem
Punkt; ohne Sitzung grauer Punkt „Kasse geschlossen"), verlinkt auf
`/admin/kasse`. Statuspunkte an Menüpunkten: Kassentag grün bei offener
Kasse, Bondrucker rot bei fehlgeschlagenen Druckaufträgen, Finanzamt & TSE
rot bei nicht konfigurierter TSE, Rückstand über der bestehenden
60-s-Schwelle oder fehlgeschlagenen Signaturaufträgen. Dafür entsteht der
Baustein `StatusDot` unter `frontend/src/admin/components/`. Alle Routen und
das aktive-Route-Verhalten bleiben unverändert; es ändern sich nur Labels,
Gruppen und Icons (ausschließlich Lucide, laut Handoff).

### Acceptance criteria

- [x] Sidebar zeigt die vier Gruppen mit den neuen Labels und Icons laut Handoff; alle Links führen auf die unveränderten Routen
- [x] Event-Status-Chip zeigt beide Zustände (offen mit Bezeichnung und Uhrzeit, geschlossen) und verlinkt auf `/admin/kasse`
- [x] Statuspunkte erscheinen an Kassentag (grün bei offener Kasse), Bondrucker (rot bei fehlgeschlagenen Bons) und Finanzamt & TSE (rot bei TSE-Problemen) und fehlen im Normalzustand ohne Befund
- [x] Die TSE-Warnschwelle nutzt dieselbe Konstante wie das Dashboard (keine duplizierte Zahl)
- [x] `StatusDot` existiert als gemeinsamer Baustein und wird für Chip und Menüpunkte verwendet
- [x] `AdminSidebar.test.tsx` deckt Gruppen/Labels, Chip-Zustände und mindestens einen Statuspunkt-Fall ab; `make check` grün

---

## Phase 2: Einheitlicher Seitenkopf, FAB-Abschaffung, Karten-Bausteine

**User stories**: 3

### Context

- `frontend/src/admin/adminListLayout.ts — adminListBottomClearance` — entfällt ersatzlos
- `frontend/src/admin/products/NewProductDialog.tsx`, `frontend/src/admin/tables/NewTischDialog.tsx`, `frontend/src/admin/users/NewUserDialog.tsx` — FAB-Trigger (fixed-position-Wrapper), wandern in den Seitenkopf
- Seiten mit eigenem H1 bzw. ohne H1: `AdminProductsPage.tsx`, `AdminTablesPage.tsx`, `AdminUsersPage.tsx`, `KassenberichtePage.tsx`, `KassensitzungPage.tsx`, `LiveReportingSection.tsx` (H1 „Live-Dashboard"), `DruckstationConfigPage.tsx` und `FinanzamtPage.tsx` (bisher ganz ohne H1)
- Handoff Abschnitte 1a–1h — finale H1-Titel und Unterzeilen je Seite

### What to build

Drei gemeinsame Bausteine unter `frontend/src/admin/components/`:
`AdminPageHeader` (H1 24 px/700, Unterzeile 14 px muted, Aktions-Slot
rechts), `HinweisKarte` und `WarnKarte` (Varianten laut Handoff-Token;
Verwendung ab Phase 3). Anschließend bekommen alle acht Admin-Seiten den
`AdminPageHeader` mit den finalen Titeln und Unterzeilen aus dem Handoff
(„Übersicht", „Berichte & Export", „Produkte & Preise", „Tische", „Helfer &
Zugänge", „Kassentag", „Bondrucker", „Finanzamt & TSE"). Die drei
Anlegen-FABs (Produkt, Tisch, Benutzer) wandern als Primary-Buttons in den
Aktions-Slot; der fixed-position-Wrapper und `adminListBottomClearance`
werden ersatzlos gelöscht. Die Seitenkörper bleiben in dieser Phase sonst
unverändert (Redesign folgt je Seite in den späteren Phasen); dynamische
Unterzeilen-Anteile (etwa Produkt-/Variantenzahlen) kommen erst mit dem
jeweiligen Seiten-Redesign.

### Acceptance criteria

- [x] `AdminPageHeader`, `HinweisKarte`, `WarnKarte` existieren als Bausteine mit den Token aus dem Handoff (keine neuen Farben/Radii)
- [x] Alle acht Seiten zeigen H1 + Unterzeile über `AdminPageHeader`; Druckstationen- und Finanzamt-Seite haben damit erstmals ein H1
- [x] „+ Neues Produkt", „+ Neuer Tisch", „+ Neuer Helfer" öffnen die bestehenden Dialoge aus dem Seitenkopf; kein FAB und keine Scroll-Clearance mehr im Admin
- [x] `adminListBottomClearance` ist aus dem Code entfernt
- [x] Bestehende Seitentests laufen angepasst grün (`make check`)

---

## Phase 3: Übersicht (1a)

**User stories**: 4, 5, 6, 7

### Context

- `frontend/src/admin/reporting/AdminDashboardPage.tsx` — heutige Warn-Banner (werden durch die Status-Zeile ersetzt), `RUECKSTAND_WARN_SEKUNDEN`
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — heutiger Dashboard-Körper (`SummaryCard`-Grid, Offene Tische, Stornierungen, Servicekräfte)
- `frontend/src/admin/reporting/hooks.ts — useLiveReporting()` — Datenquelle, `refetchInterval: 30_000` bleibt, `dataUpdatedAt` für „aktualisiert HH:MM"
- `frontend/src/admin/reporting/utils.ts — formatStand()` — Zeitanzeige
- `frontend/src/admin/reporting/StornoItem.tsx`, `StornoServicekraft.tsx — StornoAggregat` — bestehende Storno-Darstellung, wird eingeklappte Zeile
- `frontend/src/admin/kasse/hooks.ts — useKassenbestand()` — Soll-Bestand in der Kasse-Statuszelle
- Handoff Abschnitt 1a — Layout, Copy, Token

### What to build

Die Übersicht wird nach Handoff 1a umgebaut: Im Seitenkopf rechts die
Live-Anzeige („● Live · aktualisiert HH:MM" aus `dataUpdatedAt`) plus
Outline-Button „Jetzt" (manueller `refetch`); das Intervall bleibt bei den
bestehenden 30 s (PRD-Abweichung vom Handoff). Die beiden Alert-Banner
werden durch die dreispaltige Status-Zeile ersetzt (Kasse mit „seit HH:MM"
und Soll-Bestand, TSE mit Warteschlangen-Text, Drucker), Fehlerzustände als
rote Zelle mit „Beheben"-Button, der zu `/admin/druckstationen` bzw.
`/admin/finanzamt` navigiert (Logik und Schwellen der heutigen Banner
übernehmen). Kennzahlen: Hero-Karte „Kassierter Umsatz" mit erklärender
Unterzeile plus vier Nebenkarten (Noch offen, Bestellt gesamt,
Direktverkauf, Storniert). Darunter die Zwei-Spalten-Ansicht Offene Tische
(nach fünf Zeilen „Alle n anzeigen") und Team (Abrechnungsstatus je
Servicekraft, Stornos annotiert). Stornierungen werden zur kompakten,
per Default eingeklappten Zeile (Zusammenfassung via `StornoAggregat`,
Aufklappen zeigt die bestehende `StornoItem`-Liste inline). Leerzustand ohne
Kassensitzung bleibt beim bestehenden `Empty`-Muster mit Link zu
`/admin/kasse`. Nur vorhandene Daten, kein Backend-Change.

### Acceptance criteria

- [x] Status-Zeile zeigt Kasse/TSE/Drucker mit Normal- und Fehlerzustand; „Beheben" navigiert zur jeweiligen Seite; die alten Banner sind entfernt
- [x] Hero-Kennzahl „Kassierter Umsatz" und die vier Nebenkarten zeigen die Werte aus `useLiveReporting` mit den erklärenden Unterzeilen aus dem Handoff
- [x] Offene Tische und Team stehen nebeneinander; die Tisch-Liste kürzt nach fünf Einträgen mit „Alle n anzeigen"
- [x] Storno-Zeile ist eingeklappt, zeigt Zusammenfassung und expandiert zur bestehenden Detail-Liste
- [x] Auto-Refresh bleibt bei 30 s; „aktualisiert HH:MM" und der Jetzt-Button funktionieren
- [x] Seitentests (`AdminDashboardPage.test.tsx` und Nachfolger der `LiveReportingSection`-Tests) decken Status-Zeile (ok/Fehler), Hero-Zahl und Storno-Aufklappen ab; `make check` grün

---

## Phase 4: Produkte & Preise (1c) mit Lösch-Guard

**User stories**: 12, 13

### Context

- `backend/api/stammdaten/produkt/application/command.go — DeleteProdukt()` — bekommt den Guard
- `backend/api/stammdaten/produkt/application/query.go — GetAllProducts()` + `http/query_handler.go` — Produktliste, bekommt `hatVerkaeufe`
- `backend/sqlc/queries/produkte.sql` — Ort für die EXISTS-Query über das Kassenjournal
- `backend/domain/kasse/bestellung.go — PositionEventData` — Positionen referenzieren `VarianteID` (Grundlage der Verkauft-Prüfung)
- `frontend/src/admin/products/` — `Products.tsx`, `ProductItem.tsx`, Dialoge, `ProduktBackend.ts`, `Produkt.ts` (`STEUERSATZ_LABEL`)
- `frontend/src/admin/settings/hooks.ts — useDruckstationen()` — Stations-Zusatz im Kategorie-Label („Bons an Station ‚Essen'")
- `frontend/src/lib/errorMessages.ts` — neuer Fehlercode-Text
- Handoff Abschnitt 1c — Preislisten-Layout, Varianten-Chips, „···"-Menü, Hinweis-Karte

### What to build

Backend: Die Produktliste liefert je Produkt `hatVerkaeufe` (EXISTS über
`bestellung-aufgenommen:v1`- und `direktverkauf-getaetigt:v1`-Events, deren
Positionen eine Variante des Produkts referenzieren). `DeleteProdukt` prüft
dieselbe Bedingung und lehnt mit dem Fehlercode `produkt_hat_verkaeufe` ab
(Backend als Single Source of Truth). Handler- und Integrationstests: Guard
je Fehlerfall (Produkt mit Bestellung, mit Direktverkauf, ohne Verkäufe)
inklusive Fehlercode; Listen-Flag gegen ein per Events aufgebautes Journal.

Frontend: Die Produktseite wird zur Preisliste nach Kategorie (Reihenfolge
Essen, Getränke, Sonstiges; Abschnitts-Label mit Steuersatz- und
Stations-Zusatz). Produktzeilen mit Varianten-Chips (Preis + Mini-Switch für
`aktiviereVariante`/`deaktiviereVariante`, Chip-Klick öffnet
`EditVariantDialog`, Ghost-Pill „+ Variante"), rechts Pencil
(`EditProductDialog`) und „···"-Menü (DropdownMenu) mit Umbenennen, Alle
Varianten deaktivieren, Löschen. Löschen bleibt hinter dem bestehenden
AlertDialog und ist bei `hatVerkaeufe` deaktiviert mit Tooltip „Produkte mit
Verkäufen können nur deaktiviert werden". Unten die `HinweisKarte`
(„Ausverkauft? Schalter aus statt löschen …"). Unterzeile im Kopf wird
dynamisch („n Produkte · m Varianten · Änderungen wirken sofort …").
Kategorie-Icons und „Erstellt am" entfallen.

### Acceptance criteria

- [x] `/get-all-produkte` liefert `hatVerkaeufe`; `/delete-produkt` lehnt Produkte mit Verkäufen mit `produkt_hat_verkaeufe` ab (Backend-Tests je Fall grün, inkl. Journal-Fixture-Test für das Flag) <!-- sales_test.go (Integration) läuft real in Phase 12 make verify; Unit-Guard-Tests grün, SQL gegen echte DB verifiziert -->
- [x] Produktseite zeigt die Preisliste gruppiert nach Kategorie mit Varianten-Chips; Switch schaltet Varianten ohne Dialog
- [x] Löschen sitzt nur im „···"-Menü, ist bei Verkäufen deaktiviert (Tooltip) und vom Backend erzwungen; der Fehlercode hat eine deutsche Meldung in `errorMessages.ts` <!-- Abweichung: dauerhaft sichtbares Begründungs-Label statt Hover-Tooltip (Touch-App; Handoff kritisiert Hover-Tooltips) -->
- [x] „Alle Varianten deaktivieren" schaltet alle aktiven Varianten des Produkts ab
- [x] Neuer Seitentest für die Produktseite (Gruppierung, Switch-Aktion, deaktiviertes Löschen samt Tooltip) nach dem bestehenden Mock-Muster; `make check` grün

---

## Phase 5: Tische (1d) mit Saldo-Schutz

**User stories**: 14, 15

### Context

- `backend/api/stammdaten/tisch/application/command.go — TischDeaktivieren(), TischLoeschen()` — bekommen den Guard
- `backend/api/stammdaten/tisch/application/query.go — GetAllTische()` + `http/query_handler.go` — Tischliste, bekommt `saldoCents`
- `backend/sqlc/queries/tische.sql` — Ort für den Join gegen `tisch_sessions` der offenen Kassensitzung
- `backend/sqlc/queries/tisch_sessions.sql` — bestehende Projektions-Queries als Vorbild
- `frontend/src/admin/tables/` — `Tische.tsx`, `TischItem.tsx`, `NewTischDialog.tsx`, `TischBackend.ts`
- Handoff Abschnitt 1d — Kachel-Grid, Präfix-Gruppierung, Hinweis-Karte (ohne Serienanlage, per PRD abgewählt)

### What to build

Backend: Die Tischliste liefert je Tisch `saldoCents` aus der
`tisch_sessions`-Projektion der offenen Kassensitzung (0 ohne offenen Saldo
oder ohne offene Sitzung). `TischDeaktivieren` und `TischLoeschen` lehnen
Tische mit offenem Saldo mit dem Fehlercode `tisch_saldo_offen` ab. Tests je
Fehlerfall (deaktivieren/löschen mit Saldo, ohne Saldo, ohne offene Sitzung)
inklusive Fehlercode; `saldoCents` gegen ein per Events aufgebautes Journal.

Frontend: Die Tischseite wird zum Kachel-Grid, gruppiert nach Namens-Präfix
(alles vor der letzten Zahl; Rest unter „Weitere"; Gruppierung als reine,
isoliert getestete Funktion im `tables`-Ordner). Kacheln zeigen Name,
Mini-Switch und Statustext; bei offenem Saldo statt „aktiv" den Betrag
(„{Betrag} € offen"), Switch deaktiviert mit Tooltip. Kachel-Klick öffnet
den bestehenden Edit-Dialog (Umbenennen, Löschen mit gleicher Schutzregel
und AlertDialog). Einzel-Anlage über den bestehenden Dialog aus dem
Seitenkopf; ausdrücklich keine Serienanlage. Unten die `HinweisKarte` zum
Saldo-Schutz; Unterzeile im Kopf dynamisch („n Tische · m aktiv · …").

### Acceptance criteria

- [x] `/get-all-tische` liefert `saldoCents` (offene Sitzung); `/deactivate-tisch` und `/delete-tisch` lehnen bei offenem Saldo mit `tisch_saldo_offen` ab (Backend-Tests je Fall grün)
- [x] Tischseite zeigt das präfix-gruppierte Kachel-Grid; die Gruppierungsfunktion ist isoliert getestet
- [x] Kacheln mit offenem Saldo zeigen den Betrag, Switch und Löschen sind deaktiviert (Tooltip) und vom Backend erzwungen; Fehlercode-Meldung in `errorMessages.ts` <!-- Abweichung wie Phase 4: dauerhaft sichtbarer Grund statt Hover-Tooltip (Touch-App) -->
- [x] Es gibt keine Serienanlage (PRD-Abweichung vom Handoff eingehalten)
- [x] Neuer Seitentest für die Tischseite (Gruppierung, Saldo-Schutz samt Tooltip, Deaktivieren-Aktion); `make check` grün

---

## Phase 6: Helfer & Zugänge (1e)

**User stories**: 16, 17, 18

### Context

- `frontend/src/admin/users/` — `Users.tsx`, `UserItem.tsx`, `NewUserDialog.tsx` (inkl. `UserCreatedDialog`/Einmalpasswort-Flow), `UserBackend.ts` (`resetPassword`, `activateUser`, `deactivateUser`)
- `backend/api/stammdaten/user/http/command_handler.go` — bestehender Selbstlösch-Schutz (`cannot_delete_self`), bleibt unverändert
- Auth-Kontext für „das bist du": aktueller Benutzer aus dem bestehenden Auth-Singleton (siehe `AdminGuard` in `frontend/src/routes.ts`)
- Handoff Abschnitt 1e — Tabelle, Rollen-Badges, Panels

### What to build

Reines Frontend-Redesign nach Handoff 1e: Benutzer als Tabelle (Name +
Login, Rolle, Status, Aktionen) im Grid `1fr 320px` neben zwei Panels. Rollen
als beschriftete Badges (Admin default, Serviceleitung outline, Service
secondary) statt Stern-Symbolen. Status als Mini-Switch
(`activateUser`/`deactivateUser`). Aktionen: Pencil (Edit-Dialog) und
„···"-Menü mit „Passwort zurücksetzen" (bestehender
`PasswordResetDialog`-Flow, jetzt direkt an der Zeile) und „Löschen…"
(AlertDialog). Am eigenen Account: Badge „das bist du", kein Löschen im Menü
(der Backend-Schutz `cannot_delete_self` bleibt die zweite Verteidigung).
Rechte Spalte: Panel „So kommt ein Helfer rein" (drei nummerierte
Onboarding-Schritte, Einmalpasswort-Verfahren) und Panel „Was Rollen dürfen"
(Rechte-Erklärung plus Passwort-Reset-Hinweis), beide mit der Copy aus dem
Handoff. Kein Backend-Change.

### Acceptance criteria

- [x] Benutzerliste als Tabelle mit Rollen-Badges, Status-Switch und „···"-Menü; Stern-Symbole sind ersetzt
- [x] Passwort-Reset ist über das Zeilen-Menü in zwei Klicks erreichbar und nutzt den bestehenden Dialog-Flow
- [x] Eigener Account zeigt „das bist du" und bietet kein Löschen an
- [x] Onboarding- und Rollen-Panel stehen neben der Tabelle mit der Handoff-Copy
- [x] Neuer Seitentest für die Benutzerseite (Badges, das-bist-du-Fall, Reset-Aktion im Menü); `make check` grün

---

## Phase 7: Berichte & Export (1b) mit Sitzungs-Metadaten

**User stories**: 8, 9, 10, 11

### Context

- `backend/api/reporting/application/query.go — GetAbgeschlosseneKassensitzungen(), GetReporting()` — beide Antworten werden erweitert
- `backend/api/reporting/http/query_handler.go — kassensitzungItem, reportingResponse` — DTOs
- `backend/sqlc/queries/kassensitzungen.sql`, `backend/sqlc/queries/reporting.sql` — Queries; Metadaten kommen aus den Journal-Events `kassensitzung-eroeffnet:v1` (Zeitstempel), `kassensturz-durchgefuehrt:v1` (`DifferenzCents`), `tagesabschluss-erstellt:v1` (`UmsatzGesamtCents`, Zeitstempel, `user_name`)
- `frontend/src/admin/reporting/` — `KassenberichtePage.tsx` (heutiges Select mit Status-Emojis), `ReportingResults.tsx` (heutige Tabs), `StornoItem.tsx`, `hooks.ts` (`useAbgeschlosseneKassensitzungen`, `useReport`, `useDsfinvkExport`)
- `frontend/src/admin/products/Produkt.ts — STEUERSATZ_LABEL` — Steuersatz-Labels für die Tabelle
- Handoff Abschnitt 1b — Sitzungsliste, Berichtskopf, Tabelle, Export-Block

### What to build

Backend: `/get-abgeschlossene-kassensitzungen` liefert je Sitzung zusätzlich
`umsatzGesamtCents` und `abgeschlossenAm`; `/get-abrechnung` liefert
zusätzlich `eroeffnetAm`, `abgeschlossenAm`, `abgeschlossenVon` und
`kassensturzDifferenzCents`. Alles reine Projektionen aus den vorhandenen
Journal-Events, keine Schema-Änderung. Tests gegen ein per Events
aufgebautes Journal (Fixtures nach dem Muster in
`backend/repository/reporting_repo/repo_test.go`).

Frontend: Die Berichte-Seite bekommt das Grid `280px 1fr`: links die
Sitzungsliste als Karten (Datum, Nr., Bezeichnung, Gesamtumsatz;
Auswahl-Zustand laut Handoff; die offene Sitzung als nicht wählbarer Eintrag
„● offen … läuft — siehe Übersicht", Klick führt zu `/admin/auswertung`;
Status-Emojis entfallen ersatzlos). Rechts der Bericht ohne Tabs, alles
untereinander: formaler Berichtskopf („Tagesbericht Nr. … " mit
Eröffnungs-/Abschlusszeit, abschließendem Benutzer und
Kassensturz-Differenz), vier Kennzahl-Kacheln, Steuersatz-Tabelle
(Brutto/Netto/Steuer, Labels aus `STEUERSATZ_LABEL`), daneben die zwei
Mini-Listen (Umsatz pro Servicekraft, Stornierungen mit bestehender
Darstellung). Oben rechts der „Drucken"-Button (`window.print`; per
Tailwind-`print:`-Klassen druckt nur die Berichtsspalte, Sidebar und
Sitzungsliste sind ausgeblendet). Unter dem Bericht der Export-Block „Für
Steuerberater & Finanzamt" mit Erklärtext und Primary-Button „Archiv
herunterladen (ZIP)" über den bestehenden `useDsfinvkExport`.

### Acceptance criteria

- [x] Sitzungsliste zeigt abgeschlossene Sitzungen mit Datum, Nr., Bezeichnung und Umsatz; die offene Sitzung ist sichtbar, nicht wählbar und führt zur Übersicht; keine Status-Emojis mehr
- [x] Berichtskopf zeigt Nr., Eröffnungs- und Abschlusszeit, abschließenden Benutzer und Kassensturz-Differenz aus den neuen Feldern
- [x] Steuersatz-Tabelle, Servicekraft-Umsätze und Stornierungen stehen ohne Tabs untereinander
- [x] Export-Block erklärt das DSFinV-K-Archiv und lädt es über den bestehenden Export herunter
- [x] Drucken gibt nur die Berichtsspalte aus (manuelle Sichtprüfung, `print:`-Klassen vorhanden) <!-- print:hidden auf Sidebar, Mobile-Header, Sitzungsliste, Export-Block UND generischem Seitenkopf; Sichtprüfung Phase 12 -->
- [x] Backend-Tests für beide erweiterten Antworten grün (Journal-Fixtures, zusätzlich volle Integration gegen echtes Postgres gelaufen); Seitentest (`KassenberichtePage.test.tsx`) deckt Sitzungsliste inkl. offener Sitzung und Berichtskopf ab; `make check` grün

---

## Phase 8: Kassentag I — Stepper, Soll-Bestand-Aufschlüsselung, Bewegungsliste (1f)

**User stories**: 19, 20, 21

### Context

- `backend/api/kasse/kassenfuehrung/application/query.go — GetKassenbestand()` — wird um Komponenten erweitert
- `backend/repository/kassenjournal_repo/repo.go — GetKassenbestand()` + `backend/sqlc/queries/kassensitzungen.sql` — bestehende Summen-Query mit den `kj_extract_*`-Funktionen; die Komponenten werten dieselben Extraktoren einzeln aus (Bareinnahmen = Zahlungen + Direktverkauf − geldwirksame Stornos; Einlagen/Entnahmen getrennt nach `richtung` der `geldtransit-gebucht:v1`-Events)
- `backend/api/kasse/kassenfuehrung/http/` — Handler und zog-Schemas (`geldtransitBuchenSchema` als Muster)
- `frontend/src/admin/kasse/` — `KassensitzungPage.tsx` (heutige Sections), `KasseBackend.ts` (`geldtransitBuchen`), `hooks.ts` (`useOffeneKassensitzung`, `useKassenbestand`), `KassensitzungPage.test.tsx`
- Handoff Abschnitt 1f, Schritte 1 und 2 — Stepper, Aufschlüsselungs-Kacheln, Bewegungsliste

### What to build

Backend: `/get-kassenbestand` liefert zusätzlich `anfangsbestandCents`,
`bareinnahmenCents`, `einlagenCents`, `entnahmenCents` (bestehendes
Summenfeld bleibt). Neuer Endpunkt `/get-geldtransit-liste`
(`{kassensitzungNr}` → Liste mit `zeitpunkt`, `richtung`, `betragCents`,
`kommentar`, `gebuchtVon`), Projektion der `geldtransit-gebucht:v1`-Events
aus dem Kassenjournal. Tests: Komponenten und Liste gegen ein per Events
aufgebautes Journal; die Summen-Invariante (Komponenten ergeben den
bestehenden Soll-Bestand) wird explizit geprüft.

Frontend: Die Kassentag-Seite wird zum vertikalen 3-Schritte-Stepper.
Schritt 1 „Kasse eröffnet" als flache Karte (Zeit, Benutzer,
Anfangsbestand); im Leerzustand ohne Sitzung wird Schritt 1 zur aktiven
Karte mit dem bestehenden Eröffnen-Formular (inkl. TSE-Warn-Dialog),
Schritte 2 und 3 ausgegraut. Schritt 2 „Laufender Betrieb" als aktive Karte:
Soll-Bestand groß mit Stand-Zeit, darunter die vier Aufschlüsselungs-Kacheln
(Anfangsbestand, + Bareinnahmen, + Einlagen, − Entnahmen) und die Liste
„Heutige Kassenbewegungen" mit zwei Buttons „+ Geld einlegen" / „− Geld
entnehmen", die das bestehende Geldtransit-Formular als Dialog mit
vorbelegter Richtung öffnen. Schritt 3 zeigt in dieser Phase die bestehende
Abschluss-Section unverändert als dritte Stepper-Karte (Redesign in
Phase 9). Seitenkopf mit dynamischem Titel („Kassentag Nr. … — …") und
formatiertem Datum.

### Acceptance criteria

- [x] `/get-kassenbestand` liefert die vier Komponenten; ein Backend-Test weist die Summen-Invariante gegen den bestehenden Soll-Bestand nach (Journal mit Zahlungen, Direktverkauf, geldwirksamem Storno, Einlage und Entnahme)
- [x] `/get-geldtransit-liste` liefert die Buchungen der Sitzung mit Zeitpunkt, Richtung, Betrag, Kommentar und Anzeigename (Backend-Tests grün)
- [x] Kassentag zeigt den 3-Schritte-Stepper; ohne offene Sitzung ist Schritt 1 das aktive Eröffnen-Formular, Schritte 2–3 sind ausgegraut
- [x] Schritt 2 zeigt Soll-Bestand, Aufschlüsselungs-Kacheln und die Bewegungsliste; Einlage/Entnahme lassen sich über die Buttons mit vorbelegter Richtung buchen und erscheinen nach dem Buchen in der Liste
- [x] Seitentest deckt Stepper-Zustände (offene Sitzung vs. Leerzustand) und die Bewegungsliste ab; `make check` grün

---

## Phase 9: Kassentag II — Abschluss-Schritt mit Zählhilfe und Warnung (1f)

**User stories**: 22, 23

### Context

- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx` — bestehender Abschluss (EuroField, Bestätigungs-AlertDialog mit Z-Bon-Vorschau, `signaturen_ausstehend`-Retry); wird zum Inhalt von Schritt 3 umgebaut
- `frontend/src/admin/reporting/hooks.ts — useLiveReporting()` — offene Tische für die Warnung (wird dort bereits für die Z-Bon-Vorschau genutzt)
- `frontend/src/components/common/FormFields.tsx — EuroField` — Ist-Bestand-Feld
- Handoff Abschnitt 1f, Schritt 3 — Live-Rechnung, Warn-Box, Zählhilfe, Fußzeile

### What to build

Schritt 3 des Steppers wird nach Handoff umgebaut: Erklärtext, bei offenen
Tischen eine `WarnKarte` („n Tische sind noch offen (Betrag) …" aus
`useLiveReporting`), Formularzeile mit dem EuroField „Gezählter Ist-Bestand",
Outline-Button „Zählhilfe öffnen" und der Live-Rechnung Soll / Gezählt /
Differenz (Differenz = Soll − Ist, bei jeder Eingabe aktualisiert, negativ
in Rot). Fußzeile mit Hinweis („Kleine Differenzen sind normal …") und dem
Button „Kasse endgültig abschließen…", der den bestehenden
Bestätigungs-AlertDialog samt Z-Bon-Vorschau und
`signaturen_ausstehend`-Retry unverändert als zweite Stufe öffnet. Neu dazu
die Zählhilfe: `ZaehlhilfeDialog` als rein clientseitige Komponente
(Stückzahl je Nennwert von 1 ct bis 200 €, Summenanzeige, Übernahme in das
Ist-Bestand-Feld), Kern als reine Funktion `summeAusStueckzahlen`
(Stückzahlen → Cent-Summe) in `frontend/src/admin/kasse/zaehlhilfe.ts`.
Kein Backend-Change.

### Acceptance criteria

- [x] Live-Rechnung zeigt Soll, Gezählt und Differenz und aktualisiert bei jeder Eingabe; negative Differenz erscheint rot
- [x] Bei offenen Tischen erscheint die Warnung mit Anzahl und Betrag; ohne offene Tische fehlt sie
- [x] Zählhilfe summiert Stückzahlen je Nennwert und übernimmt die Summe ins Ist-Bestand-Feld; `summeAusStueckzahlen` und der Dialog sind isoliert getestet (Eingaben → Summe → Übernahme)
- [x] Der Abschluss läuft weiter über den bestehenden AlertDialog inklusive `signaturen_ausstehend`-Retry (Seitentest deckt den Retry-Fall weiter ab)
- [x] Differenz-Rechnung und Warnung sind über den Kassentag-Seitentest abgedeckt; `make check` grün

---

## Phase 10: Bondrucker (1g) mit Testdruck

**User stories**: 24, 25, 26

### Context

- `backend/api/druck/station/` — Stations-Handler (Ort für den Testdruck-Befehl); `backend/domain/druckstation/druckstation.go` (`Kategorie`, `HatBonmodus`)
- `backend/api/druck/beleg/application/kassenbeleg_command.go` — Vorbild für Outbox-Enqueue (`druckauftrag_repo.NeuerDruckauftrag`, `EnqueueDruckauftraege`)
- `backend/api/druck/bondruck/application/escpos/formatter.go` — ESC/POS-Builder; neue einfache Funktion für den Testbon (Stationsname, Zeitstempel)
- `database/migrations/` — neue Migration `04_testbon_bonart.up.sql` (CHECK auf `druckauftraege.bon_art` um `testbon` erweitern)
- `frontend/src/admin/settings/DruckstationConfigPage.tsx` — heutige Seite inkl. `AlleVerwerfenDialog`; `DruckstationBackend.ts` (`formatDruckauftragReferenz`, `REFERENZ_PRAEFIX_LABEL`, `validateDruckerIp`, `hatBonmodus`), `DruckstationConfigPage.test.tsx`
- Handoff Abschnitt 1g — Alarm-Karte, Stationskarten, Bonmodus-Options-Karten (ohne Stations-Status und ohne Bon-Klartext, per PRD abgewählt)

### What to build

Backend: Neuer Befehl `/testbon-drucken` (`{kategorie}`): liest die Station,
lehnt ohne konfigurierte Drucker-IP mit `druckstation_nicht_konfiguriert`
ab, baut über eine neue ESC/POS-Funktion einen einfachen Testbon
(Stationsname, Zeitstempel) und reiht ihn als regulären Druckauftrag mit
`bon_art = 'testbon'` und Referenz `testdruck:<kategorie>` in die Outbox
ein. Es gibt keinen eigenen Status-Rückkanal; ein fehlgeschlagener Testbon
erscheint wie jeder Auftrag in den fehlgeschlagenen Druckaufträgen. Additive
Migration erweitert den `bon_art`-CHECK. Handler-Test plus
Integrationstest (Auftrag landet in der Outbox; Fehlerfall ohne IP).

Frontend: Die Bondrucker-Seite wird nach Handoff 1g umgebaut: Zuoberst (nur
wenn vorhanden) die Alarm-Karte mit den fehlgeschlagenen Bons — je Auftrag
die bestehende Referenz-Darstellung über `formatDruckauftragReferenz`
(erweitert um das `testdruck`-Präfix) plus Fehlertext und Versuche, direkt
daran „Nochmal drucken" und „Verwerfen", bei mehreren „Alle verwerfen" mit
bestehendem Bestätigungs-Dialog. Darunter die Stationskarten (Grid, ohne
Status-Zeile): Kurzbeschreibung, IP-Feld (speichert on-blur oder per Enter
mit Erfolgs-Toast; der Speichern-Button je Feld entfällt,
`validateDruckerIp` bleibt), Outline-Button „Testbon", und für Stationen mit
`hatBonmodus` die zwei erklärenden Bonmodus-Options-Karten („Pro Position —
je Gericht ein Abreiß-Bon" / „Pro Bestellung — ein Sammelbon"). Nicht
konfigurierte Stationen zusammengefasst als gestrichelte Karte mit „Drucker
zuweisen".

### Acceptance criteria

- [x] `/testbon-drucken` erzeugt einen Outbox-Auftrag mit `bon_art = 'testbon'` und Referenz `testdruck:<kategorie>`; ohne IP kommt `druckstation_nicht_konfiguriert` (Backend-Tests grün, Migration angewendet)
- [x] Alarm-Karte steht zuoberst, zeigt Referenz-Darstellung, Fehlertext und Versuche und bietet Nochmal-drucken/Verwerfen direkt am Auftrag; kein Bon-Klartext, kein Stations-Status (PRD-Abweichungen eingehalten)
- [x] Testbon-Button je konfigurierter Station löst den neuen Endpunkt aus (Erfolgs-Toast; Fehlercode-Meldung in `errorMessages.ts`)
- [x] Drucker-IP speichert on-blur und per Enter mit Toast; kein Speichern-Button je Feld mehr
- [x] Bonmodus erscheint als zwei erklärende Options-Karten und schaltet über das bestehende Update
- [x] `DruckstationConfigPage.test.tsx` deckt Alarm-Karte, Testbon-Aktion und IP-Speichern ab; `make check` grün

---

## Phase 11: Finanzamt & TSE (1h) mit ELSTER-Flag

**User stories**: 27, 28, 29

### Context

- `database/migrations/` — neue Migration `05_elster_gemeldet_am.up.sql` (`betreiber` + nullbare Spalte `elster_gemeldet_am DATE`)
- `backend/domain/betreiber/betreiber.go`, `backend/api/stammdaten/betreiber/` — Struct, Query- und Command-Handler (Muster `updateBetreiberSchema`); neue Befehle `/elster-meldung-setzen` und `/elster-meldung-zuruecknehmen`, `/get-betreiber` liefert `elsterGemeldetAm`
- `frontend/src/admin/finanzamt/` — `FinanzamtPage.tsx` und alle Sections (`BetreiberSection`, `KassenidentitaetSection` mit Seriennummer + ELSTER-Hinweis, `SignaturauftraegeSection` mit den vier Roh-Metriken, `TSEAusfalldokumentationSection` mit der Störungsliste, `TSEAnbindungSection` mit Wizard-Link, `DokumenteUndPflichtenSection` mit externen Links), `BetreiberBackend.ts`, `hooks.ts`
- `frontend/src/admin/tse/hooks.ts — useTSEStatus(), useTSESignaturQueue(), useTSEStoerungen()` — Ampel und Collapsible-Inhalte
- Handoff Abschnitt 1h — Checkliste, Ampel, Collapsibles, Gut-zu-wissen-Karte

### What to build

Backend: Additive Migration für `elster_gemeldet_am`; zwei neue Befehle im
Betreiber-Kontext (`/elster-meldung-setzen` setzt serverseitig das aktuelle
Datum, `/elster-meldung-zuruecknehmen` setzt NULL, damit ein Fehlklick
korrigierbar bleibt); `/get-betreiber` liefert das Feld mit. Handler-Tests
für Setzen, Zurücknehmen und die erweiterte Query.

Frontend: Die Finanzamt-Seite wird nach Handoff 1h zu drei Karten umgebaut.
Karte 1 „Einrichtung — x von 3 Schritten erledigt" als Checkliste:
Vereinsdaten (erledigt, wenn die Pflichtfelder der Betreiber-Query gefüllt
sind; Bearbeiten expandiert das bestehende Betreiber-Formular),
TSE aktiv (aus `useTSEStatus`; unerledigt mit Button zum bestehenden Wizard
`/admin/tse-einrichtung`), Kassenmeldung (solange offen: rote Karte mit
Fristtext nach § 146a Abs. 4 AO, Seriennummer als Code-Pill mit Copy-Button
aus der bestehenden Kassenidentität, Link „Anleitung öffnen" und Aktion „Als
erledigt markieren"; danach grün „Gemeldet am {Datum}" mit
Zurücknehmen-Möglichkeit). Karte 2 „Läuft alles?" als Klartext-Ampel („Ja —
TSE signiert normal" bzw. rote Variante nach der bestehenden Banner-Logik
mit 60-s-Schwelle), darunter die zwei Panels Signatur-Warteschlange und
Störungsprotokoll mit Klartext-Zusammenfassung; die vier Roh-Metriken und
die bestehende Störungsliste wandern in Collapsibles („Technische Details",
„Protokoll ansehen"). Karte 3 „Gut zu wissen" mit den bestehenden externen
Links (Leitfaden, Aufbewahrungspflicht). Die manuelle Experten-Konfiguration
bleibt unverändert auf der TSE-Einrichtungsseite.

### Acceptance criteria

- [x] Migration und Befehle funktionieren: Setzen persistiert das Datum, Zurücknehmen löscht es, `/get-betreiber` liefert `elsterGemeldetAm` (Backend-Tests grün)
- [x] Checkliste zeigt „x von 3" korrekt für alle Kombinationen aus Vereinsdaten/TSE/Meldung; jeder unerledigte Schritt bietet die Behebung direkt an (Formular, Wizard-Link, Als-erledigt-markieren)
- [x] Kassenmeldung zeigt offen den Fristtext mit Paragraf und Seriennummer-Pill, erledigt „Gemeldet am {Datum}" mit Korrektur-Möglichkeit; die rote Warnung verschwindet nach dem Abhaken
- [x] „Läuft alles?" zeigt den grünen Normalzustand als Klartext und den roten Fehlerzustand nach bestehender Logik; Roh-Metriken und Störungsliste sind in Collapsibles erreichbar
- [x] Neuer Seitentest für die Finanzamt-Seite (Checklisten-Zustände, Als-erledigt-Flow, Ampel-Zustände, Collapsible-Inhalt) nach dem bestehenden Mock-Muster; `make check` grün

---

## Phase 12: Abschluss — Gegenprüfung und Abnahme

**User stories**: — (Qualitätssicherung über alle Phasen)

### Context

- `docs/plans/offene-punkte.md` — Audit-Rest „G10 / solide Destructive-Buttons im Admin" (die PRD verlangt die Gegenprüfung nach der Umsetzung)
- `docs/prds/prd-admin-redesign.md` — Abschnitte „Beschlossene Vereinfachungen" und „Out of Scope" als Prüfliste
- `Makefile` — `make verify`

### What to build

Abschließende Gegenprüfung über den gesamten Umbau: (1) Der G10-Punkt in
`docs/plans/offene-punkte.md` wird gegen den neuen Stand geprüft — die
destruktiven Aktionen sitzen jetzt in „···"-Menüs und AlertDialogen; trifft
die Befund-Grundlage (solide `bg-destructive`-Buttons auf den genannten
Seiten) nicht mehr zu, wird der Punkt abgehakt, sonst wird der Rest benannt.
(2) Prüfung der PRD-Abweichungen: keine Serienanlage, kein
Druckstations-Status, kein Bon-Klartext, Refresh bei 30 s, keine neuen
Dependencies, Icons nur Lucide, keine Event-Format-Änderungen (Contract-Test
unberührt). (3) Voller `make verify`. (4) Abnahme-Screenshots aller acht
Seiten in Light und Dark (Playwright-Chromium headless, wie bei früheren
Abnahmen) als Grundlage für Nicos Sichtabnahme; der Handoff-Ordner bleibt
bis zur Abnahme liegen.

### Acceptance criteria

- [ ] G10-Punkt in `offene-punkte.md` ist gegengeprüft und abgehakt oder mit konkretem Restbestand aktualisiert
- [ ] Alle vier PRD-Vereinfachungen und die Out-of-Scope-Grenzen sind im Code verifiziert (kurze Checkliste im Chat)
- [ ] `make verify` (inkl. Integrationstests) ist grün
- [ ] Screenshots aller acht Admin-Seiten in Light und Dark liegen zur Abnahme bereit
