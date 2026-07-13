# Plan: UI-Audit-Politur und -Härtung

> Source PRD: docs/prds/prd-ui-audit-politur.md

## Goal

Die 26 Befunde des Mehr-Experten-UI-Audits vom 13.07.2026 beheben, indem ihre
sechs systemischen Ursachen (M01–M06) plus vier lose Politur-Bündel bereinigt
werden. Kein neues Feature, keine neue Designsprache — nur Korrektheits- und
Konsistenzfixes an den vom Audit benannten Primitiven. Sequenziert nach Hebel und
Risiko (P0 → P2). Alle Änderungen sind Frontend (TSX/CSS), wenige nutzersichtbare
Strings und mitgezogene Frontend-Tests; nichts berührt Backend, API, DB-Schema,
Event-JSON oder persistierte Daten (freeze-sicher).

## Architectural decisions

Durable decisions, die für alle Phasen gelten:

- **Drawer-Vertrag (ADR 03 bleibt maßgeblich).** `DrawerBody`
  (`frontend/src/components/ui/drawer.tsx — DrawerBody`) ist der einzige
  Scrollbereich (`min-h-0 overflow-y-auto`); `DrawerFooter` ist ein
  nicht-scrollender Flex-Sibling und bleibt immer sichtbar. Gesamtsumme,
  Pflichtfeld und Primäraktion gehören in den sichtbaren (Footer-)Bereich, nicht
  in den scrollenden Body. Kein neues Drawer-Modul.
- **Raster-Regel.** Ein Grid mit ausschließlich breakpoint-definierten Spalten
  (`lg:`/`sm:`/`md:`/`2xl:grid-cols-*` ohne Basis-Track) erhält eine Basis
  `grid-cols-1`. Grids mit bewusster Basis (`grid-cols-2`) bleiben unberührt.
  Abnahme ist verhaltensbasiert: kein horizontaler Seiten-Überlauf
  (`document.scrollingElement.scrollWidth ≤ innerWidth`) bei 390px.
- **Token-Ebene.** Neue benannte CSS-Custom-Properties in
  `frontend/src/index.css` (`:root` **und** `.dark`): ein Disabled-Treatment und
  ein Warn-Treatment (Amber, distinkt von destruktiv-Rot und primär-Grün), beide
  WCAG AA in Light und Dark. Badge- und Karten-Füllung werden an vorhandene bzw.
  diese Tokens gebunden. Keine weiteren neuen Tokens.
- **Neue Test-Dependency (bewusster PRD-Override).** `@axe-core/playwright` wird
  als devDependency in `e2e/package.json` aufgenommen (nicht in `frontend/`), für
  den automatisierten WCAG-AA-Kontrast-Durchlauf. Test-/CI-only, kein
  Runtime-/Bundle-Impact. Der Nutzer hat diesen Override der PRD-Zeile „keine
  neuen Abhängigkeiten“ ausdrücklich freigegeben.
- **ADR.** NEU07 wird als `docs/adrs/04_warn-bestaetigung.md` festgehalten
  (nächste fortlaufende Nummer nach 01/02/03): Warn-Bestätigung für irreversible
  Routine-Aktionen, bindet künftige Bestätigungsdialoge.
- **Ownership-Grenze Finanzamt-Grids.** Die drei Grids der Finanzamt-Sektionen
  (`EinrichtungSection`, `LaeuftAllesSection`, `GutZuWissenSection`) gehören
  Phase 3 (Finanzamt-Layout), nicht dem generischen Raster-Sweep (Phase 2), damit
  dieselben Zeilen nicht in zwei Phasen editiert werden.
- **Kein Backend/Schema/Event.** Ausschließlich Frontend + nutzersichtbare Strings
  + Frontend-Tests. Freeze-Disziplin nicht berührt.

## Inventory

Bestehende Dateien, Muster und Prior-Art-Tests (Pfad + Symbol):

Drawer / Service:
- `frontend/src/components/ui/drawer.tsx — DrawerBody, DrawerFooter, DrawerContent` — ADR-03-Vertrag (`max-h-[85dvh]`, Safe-Area, Body = einziger Scrollbereich).
- `frontend/src/service/components/table/BestellungDrawer.tsx — BestellungDrawer` — Total (`Receipt`) + Kommentar im Body, Buttons im Footer.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx — HistorieStornierungDrawer` — Positionsliste + „Stornierung gesamt“ + Pflicht-`KommentarField` im Body; Button im Footer.
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx — HistorieUmbuchungDrawer` — Positionsliste + „Umbuchung gesamt“ + Ziel-Tisch-`NativeSelect` im Body; Button im Footer.
- `frontend/src/service/components/table/ZahlungDrawer.tsx — ZahlungDrawer` — Kassieren-Referenz; `DockActionSlot`-Zeile im festen Dock.
- `frontend/src/service/components/table/Receipt.tsx — Receipt` — geteilte Positionsliste + „Gesamt“-Zeile.
- `frontend/src/service/components/table/CommentField.tsx — KommentarField` — Pflicht-/Invalid-/Touched-Logik, Hinweistext.
- `frontend/src/service/TablePage.tsx — TablePage` — Tab-Leiste (Bestellen | Kassieren | Historie), `tabsLocked`.
- `frontend/src/service/components/ServiceDock.tsx — ServiceDock, DockActionSlot` — festes Bottom-Dock (`z-40`).

Raster / Preis:
- 12 verifizierte Sweep-Stellen (siehe Phase 2). Referenz für korrektes Muster: `frontend/src/admin/reporting/UebersichtStatusZeile.tsx` (`grid grid-cols-1 gap-3 sm:grid-cols-3`).
- `frontend/src/admin/products/VariantChip.tsx — VariantChip` — Preis inline nach Name (kein feste Spalte).
- `frontend/src/service/components/table/ProductList.tsx — VariantRow` — Preis inline nach Name; `Stepper` rechts via `justify-between`. Genutzt von `Bestellung.tsx` und `Direktverkauf.tsx`.
- `frontend/src/service/components/Stepper.tsx — Stepper` — Mengen-Stepper (`w-7`-Label).

Finanzamt:
- `frontend/src/admin/finanzamt/EinrichtungSection.tsx — EinrichtungSection, SchrittKarte` — 3-Schritt-Grid `lg:grid-cols-3`; ELSTER-UUID im `truncate`-`<code>` mit Kopier-Button; „Als erledigt markieren“ als Text-Link-`<button>` in `flex flex-wrap`.
- `frontend/src/admin/components/WarnKarte.tsx — WarnKarte` — Warn-Karte, aktuell hart `border-destructive/40 bg-destructive/4 text-destructive`.
- `frontend/src/admin/finanzamt/LaeuftAllesSection.tsx`, `GutZuWissenSection.tsx`, `BetreiberForm.tsx` — Sektionen der Finanzamt-Seite.

Dark Mode / Tokens / Primitive:
- `frontend/src/index.css` — einzige Token-Datei; `:root` (Light) + `.dark` (Dark), unabhängig authored. Kein `--disabled`-, kein `--warn`-Token.
- `frontend/src/components/ui/button.tsx — buttonVariants` — `disabled:opacity-50` uniform; `shadow-xs` nur auf `outline`; `default` = grün.
- `frontend/src/components/ui/input.tsx — Input`, `badge.tsx — Badge/badgeVariants`, `card.tsx — Card`.
- `frontend/src/admin/settings/DruckstationConfigPage.tsx — DruckstationConfigPage, AlarmKarte, FehlgeschlagenerDruckauftragRow, DruckstationCard` — „Nochmal drucken“ (default/grün), Drucker-IP-`Input`, Outline-Buttons, unbegrenzte Fehl-Bon-Liste, roher `letzterFehler`.
- `frontend/src/admin/AdminSidebar.tsx — AdminSidebar` — Theme-Toggle (Label aus `isDark`), „Bondrucker“-Label.
- `frontend/src/components/theme-provider.tsx — useTheme` — `isDark`, `setTheme`.
- `frontend/src/components/ui/sidebar.tsx — SidebarMenuButton, sidebarMenuButtonVariants` — `data-active`/`hover:bg-sidebar-accent`.

Design-System-Primitive:
- `frontend/src/components/common/EuroInput.tsx — EuroInput` — aktuell separater umrandeter €-Präfix-Kasten.
- `frontend/src/components/common/FormFields.tsx — EuroField, PriceField` — react-hook-form-Wrapper (Cent-Handling).
- `frontend/src/admin/products/EditVariantDialog.tsx — EditVariantDialog` — **Referenz** für Ein-Zeilen-Footer (Ghost-Löschen links, Abbrechen/Speichern rechts).
- `frontend/src/admin/tables/EditTischDialog.tsx — EditTischDialog` — Zwei-Block-Footer + hand-gerollter Disabled-Button.
- `frontend/src/admin/users/EditUserDialog.tsx — EditUserDialog` — Footer mit zweitem Passwort-Reset-Einstieg.
- `frontend/src/admin/users/UserRow.tsx — UserRow` — Zeilen-„···“-Menü mit Passwort-Reset (der beabsichtigte einzige Einstieg).
- `frontend/src/admin/users/UserRolle.tsx — RolleBadge` — drei Varianten pro Rolle (Kommentar: „Design-Handoff 1e“).
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx — KasseAbschliessenSection` — „Kasse endgültig abschließen“ (`variant="destructive"`); €-Input/Soll/Gezählt/Differenz/CTA-Ausrichtung.
- `frontend/src/admin/reporting/LiveReportingSection.tsx — LiveReportingSection` — „Kassierter Umsatz“-Hero als `bg-muted/60`-Div; zugleich Referenz für Listen-Kappung (5 + „Alle N anzeigen“).
- `frontend/src/admin/reporting/SummaryCard.tsx — SummaryCard` — weiße `Card`-Nachbarn.
- `frontend/src/admin/kasse/ZaehlhilfeDialog.tsx — ZaehlhilfeDialog` — rohe `<input type="number">` mit nativem Spinner.

Microcopy:
- `frontend/src/admin/settings/DruckstationBackend.ts — FehlgeschlagenerDruckauftragSchema` — `bonArt` verfügbar (`'arbeitsbon' | 'kassenbeleg' | 'testbon'`).
- `frontend/src/admin/reporting/AdminDashboardPage.tsx — AdminDashboardPage` — aggregiertes „N Bons nicht gedruckt“.
- `frontend/src/service/components/table/Zahlung.tsx`, `HistorieUmbuchungDrawer.tsx` — Plural-Strings „Alle N Positionen auswählen“.
- `frontend/src/service/components/table/TischHistorie.tsx — TischHistorie` — „Beleg drucken“ (Drift ggü. „Kassenbeleg“).
- `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx` — **Referenz** für konsistente „Kassenbeleg“-Terminologie.
- `frontend/src/lib/errorMessages.ts — commonErrorMessages` — friendly Fehlertext-Map (für NEU11).

Prior-Art-Tests (E2E in `e2e/tests/`, Vitest neben den Komponenten):
- `e2e/tests/tischservice-lange-bestellung.mobile.spec.ts` — Sticky-Footer-Regression (nur Bestellen + Kassieren).
- `e2e/tests/umbuchung.mobile.spec.ts`, `stornierung-serviceleitung.mobile.spec.ts`, `kassenabschluss.mobile.spec.ts` — Funktionsflüsse (kein Viewport-Assert).
- `frontend/src/service/components/table/HistorieStornierungDrawer.test.tsx`, `HistorieUmbuchungDrawer.test.tsx`, `Zahlung.test.tsx` — Body/Footer-Split + Plural-Strings (pinnen `/Alle 1 Positionen/`).

## Resolved decisions

Aus dem PRD-Decision-Gate und der Klärungsrunde:

- **Kontrast-Check:** `@axe-core/playwright` als e2e-devDependency (Nutzer-Override von „keine neuen Abhängigkeiten“; test-only).
- **#4 Tisch-Löschen:** Ein-Zeilen-Footer wie `EditVariantDialog` (Ghost-Löschen links, Abbrechen/Speichern rechts).
- **#5 Passwort-Reset:** nur der Zeilen-„···“-Menü-Einstieg bleibt; der Einstieg im Bearbeiten-Dialog entfällt.
- **NEU07 Bestätigungsfarbe:** eigenes Warn-Treatment (Amber), weder destruktiv-rot noch primär-grün; als ADR 04 festgehalten.
- **#8 Preis-Spalte:** gemeinsames Name/Preis/Stepper-Layout über Admin und Service.
- **#2 Button-Schatten:** `shadow-xs` einheitlich auf alle soliden Button-Varianten anwenden (konsistent mit `input`/`card`/`select`, die `shadow-xs` bereits tragen), statt nur auf `outline`.
- **NEU08 Rollen-Badges:** ein gemeinsamer Badge-Variant für alle Rollen; die Rolle wird nur über den Text unterschieden. **Achtung:** das kehrt die als „Design-Handoff 1e“ kommentierte Drei-Varianten-Absicht in `UserRolle.tsx` bewusst um — im Review benennen.
- **Finanzamt-Grids** gehören Phase 3, nicht dem Sweep (Phase 2).
- **NEU06 Disabled-Token** wird in Phase 8 (Dark-AA-Zwang) definiert und in Phase 9 an alle Call-Sites verdrahtet.

## Open questions / Risks

- **iOS-PWA-Regression (Phase 1).** Der Drawer-Umbau kann die aus ADR 03
  bekannte iOS-PWA-Regression zurückbringen. Desktop und Browser-Tab zeigen den
  Fehler nicht. **Human-Gate:** Verifikation auf einem echten iOS-Gerät als
  installierte PWA ist Pflicht (nicht automatisierbar).
- **NEU13 evtl. schon durch Radix gedeckt (Phase 11).** `Drawer` = Radix
  `Dialog.Root` (modal default) markiert Sibling-Inhalte `aria-hidden` und trappt
  Fokus. Das Dock/die Tab-Leiste ist ein fester Sibling. Ob die Tab-Leiste
  faktisch inert ist, ist ungetestet — zuerst im Browser verifizieren, dann nur
  falls nötig explizit `inert`/`aria-hidden`/`pointer-events-none` ergänzen; Test
  in jedem Fall.
- **NEU14 Dark-Highlight (Phase 11).** Der falsche Active-Highlight des
  Theme-Toggles im Dark Mode stammt vermutlich aus persistierendem
  `:active`/`hover:bg-sidebar-accent` statt `data-active` — im Browser
  reproduzieren, nicht blind raten.
- **ReportingResults-Steuertabelle (Phase 2, Edge).**
  `frontend/src/admin/reporting/ReportingResults.tsx` nutzt feste
  `grid-cols-[1.6fr_1fr_1fr_1fr]`-Tabellenzeilen — kein Breakpoint-only-Bug, also
  **kein** `grid-cols-1`. Falls der 390px-Überlauf-Check hier anschlägt, braucht
  die Zeile einen horizontalen Scroll-Container, keinen Einspalter.
- **Dark-Token-Abwägung (Phasen 8/9).** Dark-Token-Änderungen dürfen den
  Light-Mode-Kontrast nicht senken; beide Modi müssen AA halten (vgl. G10 in
  `docs/plans/offene-punkte.md`, das der Token-Pass abschließen statt duplizieren
  soll).

---

## Phase 1: Drawer-Sticky-Footer (P0)

**User stories**: 1 (Total/Pflichtfeld/Aktion ohne Scrollen sichtbar), 12
(Stornierung-Sticky-Komfort wie Kassieren).

**Befunde**: #9, #10, #11, NEU04 (Pflichtfeld-sichtbar-Teil).

### Context

- `frontend/src/components/ui/drawer.tsx — DrawerBody, DrawerFooter` — der einzuhaltende ADR-03-Vertrag; Footer bleibt sichtbar, Body scrollt.
- `frontend/src/service/components/table/BestellungDrawer.tsx — BestellungDrawer` — Total (`Receipt`) liegt im Body; Kommentar optional.
- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx — HistorieStornierungDrawer` — „Stornierung gesamt“ + Pflicht-Kommentar im Body.
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx — HistorieUmbuchungDrawer` — „Umbuchung gesamt“ + Ziel-Tisch-Auswahl im Body.
- `frontend/src/service/components/table/ZahlungDrawer.tsx — ZahlungDrawer` — Kassieren-Referenz.
- `e2e/tests/tischservice-lange-bestellung.mobile.spec.ts` — bestehende Sticky-Footer-Regression (nur Bestellen/Kassieren; hält Button im Viewport).

### What to build

Wichtig: Die Primäraktions-**Buttons** liegen in allen Drawern bereits im
sichtbaren `DrawerFooter`. Was bei langer Liste unter den Fold scrollt, sind die
**Gesamtsumme** und das **Pflichtfeld**. Diese wandern in den sichtbaren
Footer-Bereich (unmittelbar über der Primäraktion), sodass bei scrollendem Body
Gesamtsumme, Pflichtfeld und Aktion zusammen sichtbar bleiben:

- Bestellung: Gesamtsumme in den Footer-Bereich (Kommentar bleibt optional im Body).
- Stornierung: „Stornierung gesamt“ + Pflicht-Kommentar in den Footer-Bereich.
- Umbuchung: „Umbuchung gesamt“ + Ziel-Tisch-Auswahl in den Footer-Bereich.

Der ADR-03-Vertrag bleibt gewahrt: `DrawerBody` bleibt der einzige Scrollbereich,
`max-h-[85dvh]` und Safe-Area unverändert, kein neues Scroll-Nesting im Footer.
Die e2e-Regression wird auf Stornierung und Umbuchung ausgeweitet.

### Acceptance criteria

- [x] Bei einer langen Bestellung, Stornierung und Umbuchung (Body scrollt) sind Gesamtsumme, das jeweilige Pflichtfeld (Kommentar bzw. Ziel-Tisch) und die Primäraktion ohne Scrollen gleichzeitig sichtbar.
- [x] Der ADR-03-Layout-Vertrag bleibt gewahrt: `DrawerBody` einziger Scrollbereich, `max-h-[85dvh]`, Safe-Area, kein zweiter Scrollcontainer im Footer.
- [x] Die mobile e2e-Regression (390px) hält bei langer Liste in Bestellung, Stornierung und Umbuchung Gesamtsumme, Pflichtfeld und Primäraktion im Viewport (bestehender Test für Bestellen/Kassieren erweitert um Stornierung + Umbuchung).
- [x] Die bestehenden Vitest-Body/Footer-Split-Tests bleiben grün bzw. werden an die neue Platzierung angepasst.
- [x] Human-Gate dokumentiert: Prüfung auf echtem iOS-Gerät als installierte PWA (keine iOS-PWA-Regression aus ADR 03).
  > **Human-Gate (nicht automatisierbar).** Der Umbau verschiebt nur Gesamtsumme/Pflichtfeld
  > vom scrollenden `DrawerBody` in den nicht-scrollenden `DrawerFooter`; der ADR-03-Vertrag
  > (`DrawerBody` = einziger Scrollbereich mit `min-h-0 overflow-y-auto`, `DrawerFooter` als
  > nicht-scrollender Flex-Sibling, `max-h-[85dvh]`, `pb-[env(safe-area-inset-bottom)]`) bleibt
  > unverändert — es wurde kein zweiter Scrollcontainer eingeführt und keine Höhen-/Safe-Area-
  > Regel angefasst. Die aus vaul bekannte iOS-PWA-Regression (Scroll-Lock) ist damit
  > strukturell nicht berührt. Ein echter Gerätecheck ist in der Cloud-Session nicht möglich;
  > **vor dem Merge auf einem echten iOS-Gerät als installierte PWA verifizieren** (Bestellung,
  > Stornierung, Umbuchung je mit langer Liste: Body scrollt, Footer mit Summe/Pflichtfeld/
  > Aktion bleibt fixiert und über der Home-Indicator-Safe-Area sichtbar).

---

## Phase 2: Raster-Basisspalten-Sweep (P0)

**User stories**: 3 (kein horizontaler Überlauf auf dem Smartphone).

**Befunde**: #12 sowie der Überlauf-Anteil mehrerer Screens.

### Context

- Referenz für das korrekte Muster: `frontend/src/admin/reporting/UebersichtStatusZeile.tsx` (`grid grid-cols-1 gap-3 sm:grid-cols-3`).
- 12 verifizierte Fix-Stellen (Breakpoint-only ohne Basis-Track):
  - `frontend/src/service/TableSelectionPage.tsx — TischGruppe, TischListSkeleton` (2 Grids `lg:grid-cols-2 2xl:grid-cols-3`).
  - `frontend/src/service/components/table/TischHistorie.tsx — TischHistorie` (`lg:grid-cols-2 2xl:grid-cols-3`).
  - `frontend/src/service/components/direktverkauf/DirektverkaufHistorie.tsx` (2 Grids `lg:grid-cols-2 2xl:grid-cols-3`).
  - `frontend/src/service/components/table/Zahlung.tsx` (3 Grids `lg:grid-cols-2 2xl:grid-cols-3`).
  - `frontend/src/admin/reporting/LiveReportingSection.tsx` (`lg:grid-cols-2`).
  - `frontend/src/admin/reporting/ReportingResults.tsx` (`md:grid-cols-2`-Panel-Zeile).
  - `frontend/src/admin/reporting/KassenberichtePage.tsx` (`md:grid-cols-[280px_1fr]` — fixe px-Spalte, braucht Basis-Track).
  - `frontend/src/admin/users/AdminUsersPage.tsx` (`lg:grid-cols-[1fr_320px]` — fixe px-Spalte, braucht Basis-Track).
- **Nicht anfassen** (bewusste Basis): u. a. `ReportingResults.tsx` `grid-cols-2 lg:grid-cols-4`, `LaufenderBetriebSection.tsx`, `Tische.tsx`, `EigeneUebersicht.tsx`, `DruckstationConfigPage.tsx` `grid-cols-2`, sowie shadcn/ui-Primitive.
- **Ausgenommen (Phase 3):** die drei Finanzamt-Grids (`EinrichtungSection`, `LaeuftAllesSection`, `GutZuWissenSection`).
- **Testabdeckung:** Der 390px-Überlauf-Check (`e2e/tests/tischservice-viewport-ueberlauf.mobile.spec.ts`) deckt die vier Servicekraft-Screens verhaltensbasiert ab (Kassieren, Historie, Tischauswahl, Direktverkauf-Historie). Die vier Admin-Screens (`LiveReportingSection`, `ReportingResults`-Panel-Zeile, `KassenberichtePage`, `AdminUsersPage`) und der `ReportingResults`-Steuertabellen-Randfall (AC #5) sind Desktop-Flächen (`admin-*`-Specs laufen in Desktop Chrome, nicht am 390px-Handy); ihre Überlauf-Sicherheit ist strukturell durch den vorangestellten Basis-Track (`grid-cols-1` erzwingt mobil eine Einzelspalte) begründet, nicht e2e-gemessen.

### What to build

Mechanischer Sweep: jedem der 12 Breakpoint-only-Grids eine Basis `grid-cols-1`
voranstellen; bei den beiden fixen px-Template-Grids sicherstellen, dass die
Template-Spaltendefinition erst am Breakpoint greift und mobil eine
Einzelspalte gilt. Jede geänderte Fläche visuell prüfen, damit bewusste
Mehrspalter nicht zu Einspaltern werden. Eine 390px-Überlauf-Regression über die
betroffenen Screens hinzufügen bzw. erweitern.

### Acceptance criteria

- [x] Bei 390px gibt es keinen horizontalen Seiten-Überlauf (`scrollWidth ≤ innerWidth`) auf Historie, Kassieren, Direktverkauf-Historie und Tischauswahl sowie den übrigen im Sweep berührten Screens.
- [x] Alle 12 Fix-Stellen tragen eine Basis `grid-cols-1`; kein bewusstes `grid-cols-2`-Layout wurde verändert.
- [x] Beträge und Labels sind bei 390px vollständig lesbar (nicht abgeschnitten).
- [x] Eine mobile e2e-Überlaufmessung (390px) deckt die betroffenen Screens ab.
- [x] Falls die `ReportingResults`-Steuertabelle bei 390px überläuft, erhält sie einen horizontalen Scroll-Container (kein `grid-cols-1`); andernfalls bleibt sie unberührt.

---

## Phase 3: Finanzamt-Einrichtung-Layout (P0)

**User stories**: 6 (Finanzamt-Einrichtung vollständig lesen und bedienen).

**Befunde**: #1.

### Context

- `frontend/src/admin/finanzamt/EinrichtungSection.tsx — EinrichtungSection, SchrittKarte` — 3-Schritt-Grid `lg:grid-cols-3`; ELSTER-UUID im `truncate`-`<code>` mit Kopier-Button; „Als erledigt markieren“ als Text-Link-`<button>` in `flex flex-wrap`.
- `frontend/src/admin/components/WarnKarte.tsx — WarnKarte` — fehlt `min-w-0` am Text-Flex-Child; bei `lg:grid-cols-3` an 1024–1440px zu schmal.
- `frontend/src/admin/finanzamt/LaeuftAllesSection.tsx`, `GutZuWissenSection.tsx` — Sektionen darunter (`sm:grid-cols-2`).
- `frontend/src/admin/finanzamt/FinanzamtPage.tsx — FinanzamtPage` — `max-w-4xl`-Wrapper, alle Sektionen teilen die Breite.

### What to build

Das 3-Schritt-Layout so umbauen, dass es die Container-Breite der darunter
liegenden Panels nicht überschreitet oder unterhalb ~1100px vertikal stapelt
(inkl. Basis `grid-cols-1` für den mobilen Track, `min-w-0` an den squeezenden
Flex-Children). Die ELSTER-UUID bekommt ein eigenes, vollbreites Feld mit
Kopier-Button (statt truncated `<code>`), sodass sie abtippbar ist. Die
Aktionszeile bricht so um, dass „Als erledigt markieren“ bei 1440px und 1024px
vollständig sichtbar und klickbar ist. Dark-Mode-Kontrast dieser Karten wird in
Phase 8 auf AA gebracht (Cross-Ref).

### Acceptance criteria

- [x] Der 3-Schritt-Container ist nicht breiter als die Panels darunter, oder die Schritte stapeln unterhalb ~1100px vertikal.
- [x] Die ELSTER-UUID steht in einem eigenen, vollbreiten Feld mit Kopier-Button und ist vollständig lesbar/abtippbar (nicht mehr nur truncated).
- [x] „Als erledigt markieren“ ist bei 1440px und 1024px vollständig sichtbar und klickbar.
- [x] Kein horizontaler Überlauf der Finanzamt-Seite bei 390px (Basis-Track gesetzt).

---

## Phase 4: Disabled-Affordanz in den Service-Drawern (P1)

**User stories**: 2 (Grund bei deaktivierter Aktion sichtbar/geführt).

**Befunde**: NEU04, #10, #11.

### Context

- `frontend/src/service/components/table/HistorieStornierungDrawer.tsx` — Primäraktion `disabled` bei fehlendem Kommentar (`kommentarInvalid`) / keiner Auswahl.
- `frontend/src/service/components/table/HistorieUmbuchungDrawer.tsx` — Primäraktion `disabled` bei fehlendem Ziel-Tisch / keiner Auswahl; heute kein Hinweis am Button.
- `frontend/src/service/components/table/CommentField.tsx — KommentarField` — hat bereits Hinweistext, aber erst nach `touched`.
- Phase 1 hat Pflichtfeld + Aktion in den Footer-Bereich gebracht — der Grund kann dort direkt neben der Aktion erscheinen.

### What to build

Ein kleines gemeinsames Muster mit testbarer Schnittstelle
(Validierungszustand rein, Hinweistext raus): Ist die Primäraktion deaktiviert,
nennt die Oberfläche sichtbar den Grund (z. B. „Kommentar erforderlich“,
„Ziel-Tisch wählen“) und/oder führt zum fehlenden Feld — ohne auf `touched` zu
warten, sodass der Hinweis erreichbar ist, bevor der Nutzer den toten Button
antippt. Angewendet auf Stornierung und Umbuchung (und, wo eine Pflichtbedingung
besteht, Bestellung).

### Acceptance criteria

- [x] Bei deaktivierter Primäraktion in Stornierung und Umbuchung erscheint ein sichtbarer, zur fehlenden Bedingung passender Grund nahe der Aktion.
- [x] Kein stummer, grundloser toter Button mehr in diesen Service-Drawern.
- [x] Der Enabled-Zustand der Aktion folgt der Validierung; der Grundtext verschwindet, sobald die Bedingung erfüllt ist.
- [x] Vitest-Komponententest: bei deaktivierter Aktion wird der Grund gerendert (Sichtbarkeit/Textinhalt, keine Klassennamen-Assertion).

---

## Phase 5: Preis-Spalte (P1)

**User stories**: 4 (Preis in fester Spalte, unabhängig von der Namenslänge).

**Befunde**: #8.

### Context

- `frontend/src/admin/products/VariantChip.tsx — VariantChip` — `<span>{name}</span>` gefolgt von `<strong>{preis} €</strong>` inline; kein feste Spalte.
- `frontend/src/service/components/table/ProductList.tsx — VariantRow` — Name + Preis inline (`flex items-baseline gap-2.5`), `Stepper` rechts. Genutzt von `Bestellung.tsx` und `Direktverkauf.tsx`.
- Referenz-Muster (stapelt statt Spalte): `frontend/src/service/components/PositionAuswahlListe.tsx` — Name `truncate` in `flex-1 min-w-0`, Stepper fix.

### What to build

Ein gemeinsames Name/Preis/Stepper-Layout mit fester Preis-Spaltenposition,
wiederverwendet in Admin (`VariantChip`) und Service (`VariantRow`): der Name
darf umbrechen/`truncate`n (`min-w-0`), der Preis steht an einer festen
Spaltenposition und wird von langen Namen nicht abgelöst; der Stepper bleibt am
rechten Rand. Dasselbe Layout in beiden Bereichen.

### Acceptance criteria

- [x] Der Variantenpreis steht an einer festen Spaltenposition, unabhängig von der Namenslänge; ein langer/umbrechender Name löst den Preis nicht ab.
- [x] Admin (Produkt-Chips) und Service (Bestellen) nutzen dasselbe Name/Preis/Stepper-Layout.
- [x] Bei 390px kein Überlauf und keine abgeschnittenen Preise in diesen Listen.

---

## Phase 6: Microcopy und Terminologie (P1)

**User stories**: 5 (korrekter Plural), 8 (Kassenbeleg ≠ Küche).

**Befunde**: NEU02, NEU03, #7, NEU11.

### Context

- `frontend/src/admin/settings/DruckstationConfigPage.tsx — AlarmKarte` — Titel „… die Küche hat ihn/sie nicht!“ unabhängig von der Bon-Art.
- `frontend/src/admin/settings/DruckstationBackend.ts — FehlgeschlagenerDruckauftragSchema` — `bonArt` (`'arbeitsbon' | 'kassenbeleg' | 'testbon'`) verfügbar, aber ungenutzt.
- `frontend/src/admin/reporting/AdminDashboardPage.tsx — AdminDashboardPage` — aggregiertes „N Bons nicht gedruckt“ (gleiche Ursache).
- `frontend/src/service/components/table/Zahlung.tsx` und `HistorieUmbuchungDrawer.tsx` — je ein String „Alle N Positionen auswählen“ (auch bei N=1).
- Tests, die den kaputten Plural pinnen: `frontend/src/service/components/table/Zahlung.test.tsx`, `HistorieUmbuchungDrawer.test.tsx` (Regex `/Alle 1 Positionen auswählen/`).
- `frontend/src/service/components/table/TischHistorie.tsx — TischHistorie` — „Beleg drucken“/„Beleg …“; **Referenz** `DirektverkaufHistorie.tsx` (durchgängig „Kassenbeleg“).
- `frontend/src/lib/errorMessages.ts — commonErrorMessages` — friendly Fehlertext-Map; `DruckstationConfigPage.tsx` zeigt `letzterFehler` roh.
- Verbindliche Begriffe: `docs/language.md`.

### What to build

- NEU02: Der Fehl-Bon-Warntext richtet sich nach `bonArt` — Arbeitsbon → küchenrelevante Formulierung; Kassenbeleg (Gäste-Beleg) → keine Küchen-Behauptung; Testbon → neutral. Gilt für `AlarmKarte` und die aggregierte Dashboard-Meldung.
- NEU03: Singular „1 Position“, Plural „N Positionen“ in beiden Produktionsstrings; die zwei pinnenden Tests werden mitgezogen (sie kodieren den Bug).
- #7: `TischHistorie` an „Kassenbeleg“ ausrichten, wo es der gesetzliche Gäste-Beleg ist. „Bon“ bleibt dem operativen Arbeitsbon vorbehalten. Das Sidebar-Label „Bondrucker“ und die URL `/druckstationen` bleiben unberührt.
- NEU11: Drucker-Fehlertexte vereinheitlichen; roher `letzterFehler` wird über eine friendly Formulierung geführt statt Techniker-Jargon verbatim zu zeigen.

### Acceptance criteria

- [ ] Ein fehlgeschlagener Kassenbeleg wird nicht als Küchenproblem beschrieben; der Warntext folgt der tatsächlichen Bon-Art (Arbeitsbon/Kassenbeleg/Testbon).
- [ ] Bei einer Position rendert die Oberfläche „1 Position“, bei mehreren „N Positionen“ (Kassieren und Umbuchung).
- [ ] Die zuvor „Alle 1 Positionen“ pinnenden Tests prüfen nun „1 Position“ (Singular) und „N Positionen“ (Plural).
- [ ] `TischHistorie` verwendet „Kassenbeleg“ konsistent mit `DirektverkaufHistorie`; „Bondrucker“-Sidebar-Label und `/druckstationen`-Route unverändert.
- [ ] Drucker-Fehlertexte sind einheitlich formuliert; kein roher Backend-Jargon-String mehr direkt in der Admin-UI.

---

## Phase 7: Fehl-Bon-Liste begrenzen (P1)

**User stories**: 7 (Liste begrenzt/scrollbar, Stationskonfiguration erreichbar).

**Befunde**: #6.

### Context

- `frontend/src/admin/settings/DruckstationConfigPage.tsx — AlarmKarte, FehlgeschlagenerDruckauftragRow` — mappt `druckauftraege` in ein unbegrenztes `flex flex-col gap-2`-Div; drückt die Stationskarten unter den Fold.
- `frontend/src/admin/components/WarnKarte.tsx — WarnKarte` — Wrapper ohne Höhenbegrenzung.
- Referenz: `frontend/src/admin/reporting/LiveReportingSection.tsx` — kappt bei 5 + „Alle N anzeigen“.

### What to build

Die Fehl-Bon-Liste in der Höhe begrenzen und scrollbar machen (`max-h` +
`overflow-y-auto`), sodass die Stationskonfiguration darunter auch bei vielen
Einträgen immer erreichbar bleibt. Microcopy dieses Bereichs kommt aus Phase 6.

### Acceptance criteria

- [ ] Die Fehl-Bon-Liste ist höhenbegrenzt und scrollbar.
- [ ] Bei vielen Einträgen bleibt die Stationskonfiguration darunter ohne Scrollen durch die gesamte Liste erreichbar.
- [ ] Bei wenigen Einträgen entsteht kein leerer/abgeschnittener Eindruck (Höhe passt sich an bzw. Cap greift erst über der Schwelle).

---

## Phase 8: Dark-Mode-Kontrast + WCAG-AA-Gate (P1)

**User stories**: 9 (ausreichende Kontraste auf Recovery-/Compliance-Screens).

**Befunde**: NEU01, NEU05, NEU06 (Dark-Anteil), #1 (Dark-Anteil).

### Context

- `frontend/src/index.css` — `:root`/`.dark`-Token; kein `--disabled`-Token; Disabled = `opacity-50` in `button.tsx`/`input.tsx`.
- `frontend/src/components/ui/button.tsx — buttonVariants` — `default` grün; `outline` mit `dark:bg-input/30`/`dark:border-input` (translucent, kontrastarm); Disabled `opacity-50` über dunklem Primär.
- `frontend/src/components/ui/input.tsx — Input` — `dark:bg-input/30`, translucent-weiße Border.
- `frontend/src/admin/settings/DruckstationConfigPage.tsx` — „Nochmal drucken“ (default/grün), Drucker-IP-`Input`, Outline-Buttons.
- Recovery-/Compliance-Screens: Bondrucker/Druckstationen, Finanzamt (Phase 3), Kassenabschluss.
- `e2e/` — eigene `package.json`/`pnpm-lock.yaml`; Playwright-Setup vorhanden.

### What to build

Die Dark-Mode-Tokens/-Klassen so anpassen, dass grüne Aktionen
(„Nochmal drucken“), Outline-Buttons, Input-/Drucker-IP-Felder und deaktivierte
Primäraktionen im Dark Mode WCAG AA erreichen (4,5:1 für Text, 3:1 für UI-Ränder),
ohne den Light-Mode-Kontrast unter AA zu senken. Dabei wird das
Disabled-Treatment mit einem AA-tauglichen Wert (Token oder abgestimmte Klasse)
definiert, auf das Phase 9 alle Disabled-Primäraktionen verdrahtet
(NEU06-Cross-Ref). `@axe-core/playwright` wird als e2e-devDependency ergänzt und
ein automatisierter WCAG-AA-Durchlauf über die Dark-Mode-Recovery- und
-Compliance-Screens (Bondrucker/Druckstationen, Finanzamt, Kassenabschluss) in
Light und Dark eingerichtet.

### Acceptance criteria

- [ ] Grüne Aktionen, Outline-Buttons, Input-/Drucker-IP-Felder und deaktivierte Primäraktionen erreichen im Dark Mode AA (Text 4,5:1, UI-Ränder 3:1).
- [ ] Der Light-Mode-Kontrast dieser Primitive bleibt bei/über AA (nicht gesenkt).
- [ ] `@axe-core/playwright` ist als devDependency in `e2e/package.json` ergänzt (mit aktualisiertem `e2e/pnpm-lock.yaml`); kein Eintrag in `frontend/`.
- [ ] Ein automatisierter axe-WCAG-AA-Durchlauf bestätigt die Recovery-/Compliance-Screens in Light und Dark grün.
- [ ] Ein AA-taugliches Disabled-Treatment ist definiert und in Phase 9 wiederverwendbar.

---

## Phase 9: Design-Token-Ebene und Primitive (P2)

**User stories**: 10 (konsistente Badges, Karten, Bestätigungsfarben), 11
(irreversible Bestätigung signalisiert „irreversibel“, nicht „gefährlich“).

**Befunde**: NEU06 (Token/Verdrahtung), NEU07 (Warn-Token + ADR), NEU08, NEU10, #2.

### Context

- `frontend/src/index.css` — Ziel für benannte Tokens (Disabled aus Phase 8, neues Warn-Amber).
- `frontend/src/components/ui/button.tsx — buttonVariants` — `disabled:opacity-50` uniform; `shadow-xs` nur `outline`.
- `frontend/src/admin/tables/EditTischDialog.tsx` — hand-gerollter Disabled-Button (`variant="outline" className="w-full text-destructive" disabled`).
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx` — „Kasse endgültig abschließen“ `variant="destructive"` (Button + `AlertDialogAction`).
- `frontend/src/admin/users/UserRolle.tsx — RolleBadge` — drei Varianten pro Rolle (Kommentar „Design-Handoff 1e“).
- `frontend/src/components/ui/badge.tsx — Badge/badgeVariants`.
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — „Kassierter Umsatz“-Hero als `bg-muted/60`; Nachbarn `SummaryCard` (`bg-card`).
- `docs/adrs/` — ADRs 01/02/03; neuer Eintrag 04.

### What to build

- Warn-Token (Amber) in `frontend/src/index.css` (`:root` + `.dark`, beide AA), distinkt von destruktiv-rot und primär-grün; die Bestätigung „Kasse endgültig abschließen“ nutzt es statt `variant="destructive"`. ADR `docs/adrs/04_warn-bestaetigung.md` dokumentiert das Muster.
- NEU06: das in Phase 8 definierte Disabled-Treatment als einziges Disabled-Token für Primäraktionen verdrahten (nicht Mint gegen Rosa); hand-gerollte Disabled-Stile (u. a. in `EditTischDialog`) darauf umstellen.
- NEU08: alle Rollen-Badges auf denselben Badge-Variant vereinheitlichen; Rolle nur über Text unterscheiden (kehrt „Design-Handoff 1e“ bewusst um — im Review benennen).
- NEU10: die „Kassierter Umsatz“-Karte nutzt dieselbe Karten-Füllung wie die weißen Nachbar-Karten (statt `bg-muted/60`).
- #2: `shadow-xs` einheitlich über die soliden Button-Varianten (konsistent mit `input`/`card`/`select`).

### Acceptance criteria

- [ ] Ein Warn-Amber-Token existiert in Light und Dark (beide AA); „Kasse endgültig abschließen“ nutzt das Warn-Treatment, nicht destruktiv-rot.
- [ ] `docs/adrs/04_warn-bestaetigung.md` ist angelegt und beschreibt die Warn-Bestätigung für irreversible Routine-Aktionen.
- [ ] Deaktivierte Primäraktionen nutzen durchgängig ein einziges Disabled-Treatment (kein Mint-gegen-Rosa); hand-gerollte Disabled-Stile sind ersetzt.
- [ ] Alle Rollen-Badges nutzen denselben Badge-Variant; die Rolle wird nur über den Text unterschieden.
- [ ] Die „Kassierter Umsatz“-Karte wirkt zwischen den weißen Karten nicht grau/deaktiviert (einheitliche Füllung).
- [ ] `shadow-xs` ist über die soliden Button-Varianten einheitlich (nicht nur auf `outline`).

---

## Phase 10: Geld-Eingabe und Dialog-Konsistenz (P2)

**User stories**: 10 (konsistente Geld-Eingaben und Dialog-Footer), 11
(Kassenabschluss-Ausrichtung).

**Befunde**: #3, #4, #5, NEU07 (Ausrichtungs-Anteil).

### Context

- `frontend/src/components/common/EuroInput.tsx — EuroInput` — separater umrandeter €-Präfix-Kasten (`border-r-0`-Span + `border-l-0`-Input). Genutzt (via `PriceField`/`EuroField`) in `NewVariantDialog`, `EditVariantDialog`, `KasseAbschliessenSection`.
- `frontend/src/admin/tables/EditTischDialog.tsx — EditTischDialog` — Zwei-Block-Footer; **Referenz** `frontend/src/admin/products/EditVariantDialog.tsx` (Ein-Zeilen-Footer, `sm:justify-between`, Ghost-Löschen links).
- `frontend/src/admin/users/EditUserDialog.tsx — EditUserDialog` — Footer mit zweitem Passwort-Reset-Einstieg; der beabsichtigte Einstieg ist `frontend/src/admin/users/UserRow.tsx` (Zeilen-„···“-Menü). `HelferPanels.tsx`-Hilfetext beschreibt das Zeilen-Menü als einzigen Weg.
- `frontend/src/admin/kasse/KasseAbschliessenSection.tsx` — €-Input, Soll/Gezählt/Differenz (`ml-auto`-Block) und CTA in getrennten Flex-Zeilen.

### What to build

- #3: `EuroInput` auf eine nahtlose €-Affordanz umstellen (kein separater umrandeter Präfix-Kasten; das € liegt innerhalb desselben umrandeten Feldes), Wert weiter in Cent. Alle bestehenden Call-Sites profitieren automatisch, da sie bereits über `EuroField`/`PriceField` laufen.
- #4: `EditTischDialog` übernimmt den Ein-Zeilen-Footer wie `EditVariantDialog` (Ghost-Löschen links, Abbrechen/Speichern rechts), ersetzt die getrennte Vollbreiten-Zeile.
- #5: den zweiten Passwort-Reset-Einstieg im `EditUserDialog`-Footer entfernen; nur der Zeilen-„···“-Menü-Einstieg bleibt (deckt sich mit der Hilfe-Karte).
- NEU07 (Ausrichtung): €-Eingabe, Soll/Gezählt-Differenz und CTA im Kassenabschluss visuell zusammenführen, sodass die Bestätigung nahe den Zahlen steht, die sie bestätigt.

### Acceptance criteria

- [ ] `EuroInput` zeigt eine nahtlose €-Affordanz (kein separater umrandeter Präfix-Kasten); Cent-rein/formatiert-und-Cent-raus bleibt unverändert (bestehende `EuroInput`-Unit-Tests grün bzw. angepasst).
- [ ] Die Geld-Eingabe ist über Neue Variante, Variante bearbeiten und Kasse abschließen identisch.
- [ ] `EditTischDialog` nutzt den Ein-Zeilen-Footer wie `EditVariantDialog`.
- [ ] Es gibt nur noch einen Passwort-Reset-Einstieg (Zeilen-„···“-Menü); der Einstieg im Bearbeiten-Dialog ist entfernt.
- [ ] Im Kassenabschluss sind €-Eingabe, Soll/Gezählt-Differenz und CTA visuell zusammengeführt/ausgerichtet.

---

## Phase 11: Form-Feinschliff und Interaktion (P2)

**User stories**: 10 (ruhige, vorhersehbare Oberfläche).

**Befunde**: NEU09, NEU12, NEU13, NEU14.

### Context

- `frontend/src/service/components/Stepper.tsx — Stepper` — Minus bei Menge 0: `disabled:opacity-100` + `border-dashed`, wirkt antippbar.
- `frontend/src/admin/kasse/ZaehlhilfeDialog.tsx — ZaehlhilfeDialog` — rohe `<input type="number">` mit nativem Spinner (Stück-Zählung, kein Euro).
- `frontend/src/service/TablePage.tsx — TablePage` (Tab-Leiste), `frontend/src/service/components/ServiceDock.tsx — ServiceDock` (festes Dock), `frontend/src/components/ui/drawer.tsx — Drawer` (Radix `Dialog.Root`, modal default).
- `frontend/src/admin/AdminSidebar.tsx — AdminSidebar` — Theme-Toggle-Label aus `isDark`; `frontend/src/components/theme-provider.tsx — useTheme`; `frontend/src/components/ui/sidebar.tsx — SidebarMenuButton, sidebarMenuButtonVariants` (`data-active`/`hover:bg-sidebar-accent`).

### What to build

- NEU09: den Minus-Stepper bei Menge 0 eindeutig deaktiviert darstellen (kein geisterhafter, voll-deckender gestrichelter Kreis, der antippbar wirkt).
- NEU12: die Zählhilfe zeigt keine nativen Browser-Number-Spinner mehr; konsistente Formular-Darstellung wie anderswo (Wert bleibt Stückzahl).
- NEU13: zuerst verifizieren, ob die Tab-Leiste bei offenem Drawer faktisch inert ist (Radix-Default). Falls nicht, sie bei offenem Drawer nicht interaktiv und nicht fokussierbar machen (`inert`/`aria-hidden`/`pointer-events-none`); in jedem Fall Regressionstest ergänzen.
- NEU14: der Theme-Umschalter bekommt ein stabiles Label (nicht aus `isDark` abgeleitet) und im Dark Mode kein fälschliches Active-Page-Highlight (Ursache im Browser reproduzieren: `:active`/`hover`-Persistenz vs. `data-active`).

### Acceptance criteria

- [ ] Der Minus-Stepper bei Menge 0 ist eindeutig als deaktiviert erkennbar und wirkt nicht antippbar.
- [ ] Die Zählhilfe zeigt keine nativen Number-Spinner-Pfeile mehr.
- [ ] Bei offenem Service-Drawer ist die Tab-Leiste nicht interaktiv und nicht per Tastatur fokussierbar (durch Test belegt).
- [ ] Der Theme-Umschalter hat ein stabiles Label unabhängig vom aktiven Theme.
- [ ] Der Theme-Umschalter erhält im Dark Mode kein Active-Page-Highlight.
