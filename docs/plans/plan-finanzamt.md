# Plan: Finanzamt-Seite

> Source PRD: [docs/prds/prd-finanzamt.md](../prds/prd-finanzamt.md)

## Goal

Eine neue Admin-Seite „Finanzamt" unter `/admin/finanzamt` wird der eine Ort
für alle Belange gegenüber dem Finanzamt. Die heute über die Einstellungen
verstreuten fiskalischen Sektionen (Betreiber-Stammdaten, Kassenidentität,
TSE-Ausfalldokumentation) ziehen dorthin um, ergänzt um eine faktische
TSE-Anbindungs-Zeile mit Einstiegspunkt zur TSE-Einrichtung und eine Sektion
„Dokumente und Pflichten". Das Reporting-Dashboard behält nur einen kompakten
Warn-Banner. Die Einstellungen-Seite entfällt. Es ist eine reine
Frontend-Umstrukturierung auf bestehenden Endpunkten: kein Backend-Umbau, keine
Prüfungsbereitschafts- oder Readiness-Bewertung.

## Architectural decisions

Durable Entscheidungen für alle Phasen:

- **Frontend-Routes:** Neue lazy Admin-Routen `/admin/finanzamt` und
  `/admin/tse-einrichtung`, beide hinter `AdminGuard`, analog zu den übrigen
  Admin-Routen. Die Route `/admin/einstellungen` entfällt am Ende.
- **Kein Backend-Umbau:** Die Seite komponiert ausschließlich bestehende
  Endpunkte (`get-betreiber`, `update-betreiber`, `get-kassenidentitaet`,
  `get-tse-nachsignier-auftraege`, `tse-nachsignier-auftrag-zuruecksetzen`,
  `tse-nachsignier-auftrag-verwerfen`, `get-tse-status`,
  `get-tse-konfiguration`, `update-tse-konfiguration`,
  `test-tse-verbindung`). Kein neuer Endpunkt, Kontext oder Schema.
- **Komponenten und Hooks werden verschoben und wiederverwendet, nicht
  dupliziert:** Die heute in `EinstellungenPage.tsx` inline definierten
  Sektionen (Betreiber, Kassenidentität, Nachsignier) und das
  TSE-Konfigurationsformular werden an ihre neuen Orte verschoben; während der
  Übergangsphase (Phase 1) bezieht die noch bestehende Einstellungen-Seite sie
  von dort. Die Hooks bleiben in `frontend/src/admin/settings/hooks.ts`, da sie
  weiter auch von der Druckstationen-Seite genutzt werden.
- **TSE-Anbindung ist faktische Anzeige plus Navigation, kein Urteil:** Die
  Zeile zeigt „konfiguriert ja/nein" und Umgebung aus `get-tse-status` und
  verlinkt auf `/admin/tse-einrichtung`. Ausdrücklich keine Ampel, keine
  Readiness-Bewertung.
- **Kassenidentität:** `angelegtAm` wird als „Anlegedatum" beschriftet, nicht
  als ELSTER-Inbetriebnahmedatum. Ein rechtlich gemeintes Inbetriebnahmedatum
  ist Sache von F-05.
- **Dokumente und Pflichten:** Klartext-Hinweis auf die 10-Jahres-Aufbewahrung
  plus externe Links (`target="_blank"`) auf die Repo-Dokumente
  (Betreiber-Leitfaden, Compliance-Überblick) im öffentlichen
  Source-Available-Repo.

## Inventory

- `frontend/src/routes.ts:131-137`: lazy Admin-Route `einstellungen` (Muster
  für die neuen Routen); `:80-139` Admin-Routenbaum hinter `AdminGuard`
- `frontend/src/admin/AdminSidebar.tsx:39-70`: `adminItems` der Gruppe
  „Verwaltung" inkl. Eintrag „Einstellungen" (`:65-69`); `:1-12` Icon-Importe
  (Landmark ergänzen)
- `frontend/src/admin/settings/EinstellungenPage.tsx:23-86`:
  `KassenidentitaetSection` (Label „Inbetriebnahmedatum" bei `:75` wird
  „Anlegedatum"); `:88-232` `BetreiberForm`/`BetreiberSection`; `:234-467`
  `TSEKonfigurationForm`/`TSEKonfigurationSection` (zieht nach
  `tse-einrichtung`); `:469-595` `NachsignierAuftragRow`/`TSENachsignierSection`;
  `:597-609` `EinstellungenPage` (Komposition, entfällt am Ende)
- `frontend/src/admin/settings/hooks.ts:65-157`: `useKassenidentitaet`,
  `useBetreiber`, `useTSEKonfiguration`, `useTSENachsignierAuftraege`,
  `useTSEStatus` (bleiben, werden wiederverwendet); `:18-63` Druckstation-Hooks
  (bleiben für die Druckstationen-Seite)
- `frontend/src/lib/EinstellungenBackend.ts:88-168`: `EinstellungenBackend` mit
  allen genutzten Endpunkt-Methoden; `:37-45` `TSEStatus` (Umgebung,
  `offeneNachsignierungen`, `istKonfiguriert`)
- `frontend/src/admin/reporting/AdminDashboardPage.tsx:21-57`: die zwei
  ausführlichen TSE-Warnblöcke (`showTSEWarning`, `showNachsignierWarning`),
  Auslöser via `useTSEStatus`; der Block bei `:34-39` verlinkt heute auf
  `/admin/einstellungen`
- `docs/betrieb/leitfaden-betreiber.md`, `docs/compliance.md`: Zieldokumente der
  Sektion „Dokumente und Pflichten"

## Resolved decisions

Aus dem Klärungsprozess (2026-06-15, mit User abgestimmt):

- **Minimaler TSE-Umzug in diesem Plan:** Dieser Plan legt
  `/admin/tse-einrichtung` als schlanke Route an und verschiebt das bestehende
  TSE-Konfigurationsformular unverändert dorthin, damit das Löschen der
  Einstellungen sicher ist und die Finanzamt-Verlinkung nie verwaist,
  unabhängig vom Fortschritt des TSE-Wizard-Plans. Der TSE-Wizard baut seinen
  Wizard später auf dieser Seite auf und ersetzt damit seine eigene Phase 2
  („`tse-einrichtung` anlegen, Einstellungen-Sektion zu Status plus Link").
- **Dokumente und Pflichten via externe Repo-Links:** Klartext-Hinweis plus
  externe Links auf die Markdown-Dokumente im öffentlichen Repo. Keine
  In-App-Hilfeseiten, kein Markdown-Rendering.
- **Phasenzuschnitt:** zwei Phasen, additive Phase 1, Cutover in Phase 2.

## Open questions / Risks

- **Ziel-URL der Doku-Links offen:** Im Repo ist (noch) keine öffentliche
  Repo- oder Homepage-URL hinterlegt (`package.json`, `README`). Die konkrete
  Basis-URL ist bei der Umsetzung von Phase 1 einzutragen; die Sektionsstruktur
  steht unabhängig davon fest.
- **Koordination mit dem TSE-Wizard-Plan:** Dieser Plan nimmt die im
  TSE-Wizard-Plan, Phase 2 vorgesehene Anlage von `/admin/tse-einrichtung` und
  den Konfig-Umzug vorweg. Wird der TSE-Wizard später umgesetzt, baut seine
  Phase 2 auf dem hier geschaffenen Stand auf, statt ihn erneut anzulegen.

---

## Phase 1: Finanzamt-Seite, `tse-einrichtung`-Route und Sidebar (additiv)

**User stories**: 1, 2, 3, 4, 5, 6, 8, 9

### Context

- `frontend/src/routes.ts:131-137`: Routen-Muster (lazy Admin-Route)
- `frontend/src/admin/AdminSidebar.tsx:39-70`: `adminItems` (neuer Eintrag
  „Finanzamt")
- `frontend/src/admin/settings/EinstellungenPage.tsx:23-595`: die zu
  verschiebenden Sektionen und das TSE-Konfigurationsformular
- `frontend/src/admin/settings/hooks.ts:65-157`: wiederverwendete Hooks
- `frontend/src/lib/EinstellungenBackend.ts:88-168`: genutzte Endpunkte

### What to build

Zwei neue lazy Admin-Routen entstehen, beide hinter `AdminGuard`:

- `/admin/tse-einrichtung` hostet das bestehende TSE-Konfigurationsformular
  (Speichern, Leeren, Verbindung testen) unverändert. Es wird aus der
  Einstellungen-Seite herausgelöst und hierher verschoben.
- `/admin/finanzamt` ist die neue Seite, mobile-first, Karten stapeln auf
  kleinen Bildschirmen (ab 360 px bedienbar). Sie komponiert:
  - **Betreiber-Stammdaten:** lesen und bearbeiten (Vereinsname, Adresse,
    Steuernummer, USt-ID), wie heute.
  - **Kassenidentität:** read-only, Seriennummer mit Kopier-Funktion; das
    `angelegtAm`-Feld wird als „Anlegedatum" beschriftet, nicht als
    Inbetriebnahmedatum.
  - **TSE-Ausfalldokumentation:** die Nachsignier-Liste samt Recovery-Aktionen
    (zurücksetzen, verwerfen), vollständig umgezogen.
  - **TSE-Anbindung:** eine faktische Status-Zeile (konfiguriert ja/nein,
    Umgebung aus `get-tse-status`) mit Link „Einrichten oder ändern" auf
    `/admin/tse-einrichtung`.
  - **Dokumente und Pflichten:** Klartext-Hinweis zur 10-Jahres-Aufbewahrung
    und externe Links auf Betreiber-Leitfaden und Compliance-Überblick.

Die Seitenleiste erhält in der Gruppe „Verwaltung" einen neuen Eintrag
„Finanzamt" (Landmark-Icon). Die Sektionskomponenten und das TSE-Formular
werden an ihre neuen Orte verschoben und wiederverwendet, nicht dupliziert; die
noch bestehende Einstellungen-Seite bezieht sie übergangsweise von dort. Kein
Backend-Umbau.

### Acceptance criteria

- [ ] `/admin/finanzamt` ist über einen Sidebar-Eintrag „Finanzamt" in der
      Gruppe „Verwaltung" erreichbar
- [ ] Betreiber-Stammdaten laden und speichern funktionieren auf der
      Finanzamt-Seite
- [ ] Kassenidentität wird angezeigt, die Seriennummer ist kopierbar, das
      Datum ist als „Anlegedatum" beschriftet
- [ ] Die TSE-Ausfalldokumentation zeigt die Nachsignier-Vorgänge; einreihen
      (zurücksetzen) und verwerfen funktionieren
- [ ] Die TSE-Anbindungs-Zeile zeigt konfiguriert ja/nein und Umgebung und
      verlinkt auf `/admin/tse-einrichtung`
- [ ] `/admin/tse-einrichtung` hostet das TSE-Konfigurationsformular;
      Speichern, Leeren und Verbindung testen funktionieren dort unverändert
- [ ] „Dokumente und Pflichten" zeigt den 10-Jahres-Hinweis und die externen
      Doku-Links
- [ ] Die Seite ist ab 360 px Breite bedienbar (Karten stapeln)
- [ ] `make lint` (Frontend Lint/Typecheck) grün

---

## Phase 2: Dashboard-Banner und Wegfall der Einstellungen-Seite (Cutover)

**User stories**: 7 (schließt zudem 1 ab)

### Context

- `frontend/src/admin/reporting/AdminDashboardPage.tsx:21-57`: die zwei
  ausführlichen TSE-Warnblöcke (Auslöser via `useTSEStatus`), Link bei `:34-39`
  auf `/admin/einstellungen`
- `frontend/src/routes.ts:131-137`: zu entfernende Route `einstellungen`
- `frontend/src/admin/AdminSidebar.tsx:65-69`: zu entfernender Sidebar-Eintrag
- `frontend/src/admin/settings/EinstellungenPage.tsx:597-609`: zu entfernende
  Page

### What to build

Auf dem Reporting-Dashboard ersetzt ein kompakter Banner die beiden
ausführlichen TSE-Warnblöcke. Der Banner erscheint bei nicht konfigurierter TSE
oder offenen Nachsignierungen (unveränderte Auslöser via `get-tse-status`) und
verlinkt auf die Finanzamt-Seite; nur Darstellung und Detailort ändern sich.

Anschließend werden die Einstellungen-Route, der Sidebar-Eintrag und die
`EinstellungenPage` entfernt. Es bleibt keine Referenz auf
`/admin/einstellungen` übrig.

### Acceptance criteria

- [ ] Das Dashboard zeigt statt der zwei Blöcke einen kompakten Banner, der bei
      nicht konfigurierter TSE oder offenen Nachsignierungen erscheint und auf
      `/admin/finanzamt` verlinkt
- [ ] Die Route `/admin/einstellungen` existiert nicht mehr; der Sidebar-Eintrag
      „Einstellungen" ist entfernt
- [ ] `EinstellungenPage` ist entfernt; keine Datei referenziert mehr
      `/admin/einstellungen`
- [ ] Betreiber, Kassenidentität, Nachsignier und TSE-Konfiguration sind
      ausschließlich über Finanzamt bzw. `tse-einrichtung` erreichbar (keine
      doppelte Darstellung mehr)
- [ ] `make lint` (Frontend Lint/Typecheck) grün
