# Plan: Spektral-Redesign des App-Frontends

> Source PRD: docs/prds/prd-spektral-redesign-frontend.md

## Goal

Die Spektral-Markensprache der Website dezent auf das App-Frontend übertragen
(Login, Service, Verwaltung) — als reiner Restyling-Layer. Das Spektrum
erscheint an genau vier Stellen (Wortmarke, Hintergrund-Glows, Ladezustände,
Hairline-Akzente), dazu kommen das verbindliche Motion-Inventar der PRD, die
Typografie-Vereinheitlichung (Space Grotesk als Heading-Schrift) und der
Erfolgs-Pop, der die Erfolgs-Toasts in den drei Service-Buchungsflows ersetzt.
Layouts, Bedienabläufe, Status-Semantik (Grün/Rot/Amber) und
Barrierefreiheits-Entscheidungen bleiben unangetastet. Referenz für Maße,
Farben und Kurven ist der Handoff (`docs/prds/design_handoff_spektral_redesign/`),
bei inhaltlichem Widerspruch gilt die PRD.

## Architectural decisions

Durable decisions, die für alle Phasen gelten:

- **Tokens.** Die acht `--sp-*`-Töne und `--spectral` wandern nach
  `frontend/src/index.css`: Light-Werte in `:root`, Dark-Overrides in `.dark`
  (App-Mechanismus `@custom-variant dark`, nicht der `[data-theme]`-Selektor
  der Website — maßgeblich ist Wertgleichheit). Quelle mit
  Single-Source-Kommentar: `website/src/styles/brand.css`. Die Website
  überschreibt im Dark-Mode nur fünf Töne (red, green, teal, blue, violet) und
  nicht den Verlauf selbst; der Dark-`--spectral` der App wird gemäß
  Handoff-Spezifikation aus den effektiven Dark-Tonwerten rekomponiert.
  `--spectral-v` und die `--sp-*-text`-Varianten der Website werden nicht
  übernommen (kein Konsument in der App).
- **Animations-Inventar zentral.** Keyframes `shimmer`, `fadeUp`, `pop`,
  `pulsedot`, `drift` plus benannte, wiederverwendbare Utilities
  (Arbeitsnamen: `animate-fade-up`, `animate-pop`, `animate-pulsedot`,
  `animate-drift`, `skeleton-shimmer`) einmal in `index.css` — Einsatzstellen
  konsumieren nur, keine Ad-hoc-Arbitrary-Animationen pro Stelle. Eine
  zentrale `@media (prefers-reduced-motion: reduce)`-Regel nullt Animations-
  und Transition-Dauern global (erfasst auch `tw-animate-css`-Bestand).
  Dauern und Easings verbindlich nach dem Motion-Inventar der PRD (dort
  tabelliert, hier nicht dupliziert).
- **Neue Komponenten/Hooks** (Namen verbindlich):
  `frontend/src/components/common/Wortmarke.tsx` (`Wortmarke`),
  `frontend/src/admin/components/HeaderGlow.tsx` (`HeaderGlow`),
  `frontend/src/service/components/ErfolgsPop.tsx` (`ErfolgsPop`),
  `frontend/src/hooks/use-count-up.ts` (`useCountUp`). Die Login-Glows
  entstehen inline im `AuthLayout` (drei dekorative divs) — keine eigene
  Glow-Primitive, es gibt nur zwei ungleiche Formen (Kreis-Trio vs.
  Ellipsen-Paar).
- **HeaderGlow-Integration.** Einmalig in `AdminPageHeader` (optionale
  Farbpaar-Prop, Default teal+violett) statt pro Seite montiert — jede Seite
  mit `AdminPageHeader` bekommt den Glow automatisch, das Farbpaar ist die
  einzige Stellschraube. Mapping: Übersicht teal+violett, Kassentag
  orange+teal, Produkte & Preise grün+blau, Helfer & Zugänge blau+rot; Tische,
  Bondrucker, Berichte & Export, Finanzamt & TSE nutzen den Default
  teal+violett. `TSEEinrichtungPage` hat keinen `AdminPageHeader` und bekommt
  keinen Glow.
- **Typografie.** `@fontsource-variable/space-grotesk` ist die einzige neue
  Dependency (PRD-genehmigt). `--font-heading` im `@theme inline`-Block wird
  von `var(--font-sans)` auf `'Space Grotesk Variable', 'Inter Variable',
  sans-serif` umgehängt; damit greifen `CardTitle`, Dialog-, Drawer-, Sheet-,
  AlertDialog- und Empty-Titel die Schrift automatisch ab (tragen
  `font-heading` bereits). Explizit ergänzt wird `font-heading` nur am
  `AdminPageHeader`-h1 und am Tischtitel-h1 der `TablePage`. Beträge und
  Kennzahl-Werte bleiben Inter (tabellarische Ziffern).
- **Kein Backend, keine weiteren Dependencies.** Reines Frontend; keine
  Animations-Library (CSS + `requestAnimationFrame` genügen).

## Inventory

- `frontend/src/index.css` — `@theme inline` mit `--font-heading:
  var(--font-sans)`; Token-Blöcke `:root`/`.dark` (oklch); einziger
  `@layer base`; importiert `tw-animate-css` und
  `@fontsource-variable/inter`; bisher keine Keyframes, keine
  Reduced-Motion-Regel.
- `website/src/styles/brand.css` — Quell-Tokens `--sp-*`/`--spectral`
  (`:root`; Dark unter `[data-theme='dark']`).
- `frontend/src/components/common/AuthLayout.tsx — AuthLayout` — gemeinsames
  Layout von `LoginPage`, `PasswordPage`, `ErrorPage`.
- Wortmarken-h1 mit identischem Markup (`text-4xl text-center
  font-extrabold`): `components/common/LoginForm.tsx`,
  `components/common/PasswordForm.tsx`, `pages/HydrateFallbackPage.tsx`
  (ohne AuthLayout), `pages/ErrorPage.tsx`.
- `frontend/src/admin/AdminSidebar.tsx` — Wortmarke-h1 (`text-3xl
  font-extrabold`), `NavGroup`, Kassentag-Chip mit `StatusDot`
  (`kasseOffen`), Footer-Versionszeile `jotti {version}`.
- `frontend/src/admin/AdminLayout.tsx — AdminMobileHeader` — Wortmarke-span
  (`text-sm font-bold`), bereits `print:hidden`.
- `frontend/src/admin/components/AdminPageHeader.tsx` — h1 `text-2xl
  font-bold`; Verwendungsstellen sind genau die acht Admin-Seiten:
  `admin/reporting/LiveReportingSection.tsx` (Übersicht),
  `admin/kasse/KassensitzungPage.tsx` (Kassentag),
  `admin/products/AdminProductsPage.tsx`, `admin/tables/AdminTablesPage.tsx`,
  `admin/users/AdminUsersPage.tsx`,
  `admin/settings/DruckstationConfigPage.tsx`,
  `admin/reporting/KassenberichtePage.tsx`,
  `admin/finanzamt/FinanzamtPage.tsx`.
- `frontend/src/admin/components/StatusDot.tsx — StatusDot` — 7-px-Punkt,
  `zustand: ok | fehler | neutral`, `role="img"` + `aria-label`.
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Hero-Karte
  „Kassierter Umsatz“ (eigene Card mit `ring-1`), vier `SummaryCard`s,
  Live-Zeile mit `StatusDot`; Polling über `useLiveReporting`
  (`admin/reporting/hooks.ts`, `refetchInterval: 30_000`).
- `frontend/src/admin/kasse/KassensitzungPage.tsx — StepperRow,
  EroeffnetKarte` — Schritt-Schiene (`done`/`active`/`inactive`,
  Check-Häkchen); der Wechsel nach „Kasse eröffnen“ ist datengetrieben
  (`refetch` von `useOffeneKassensitzung`). Schritt-Titel sind `CardTitle`s.
- `frontend/src/components/ui/skeleton.tsx — Skeleton` — `animate-pulse
  bg-muted`. Verwendungsstellen: `TischListSkeleton`
  (`service/TableSelectionPage.tsx`), `ProductListSkeleton`
  (`service/components/table/ProductList.tsx`; konsumiert von
  `Bestellung.tsx` und `direktverkauf/Direktverkauf.tsx`),
  `HistorieRowSkeleton` (`service/components/table/HistorieRowSkeleton.tsx`;
  konsumiert von `TischHistorie.tsx` und `DirektverkaufHistorie.tsx`),
  `EigeneUebersicht.tsx`, `SidebarMenuSkeleton` (`ui/sidebar.tsx`).
- `frontend/src/components/ui/button.tsx — buttonVariants` — hat
  `transition-all` und `active:not-aria-[haspopup]:translate-y-px`.
- `frontend/src/components/ui/tabs.tsx — TabsContent` — minimal, ohne
  Animation; `frontend/src/components/ui/sidebar.tsx — SidebarMenuButton` —
  Aktiv-Zustand über `data-active`-Attribut.
- `frontend/src/service/TablePage.tsx` — Tischtitel-h1 (`text-[22px]`),
  Status-Badge (`n unbezahlt` / `Alles bezahlt`),
  `data-slot="tisch-saldo"` (bereits `tabular-nums`), `tabsLocked`,
  `reload()` (Refetch von State + Historie).
- Erfolgs-Callbacks mit `toast.success` (sonner, Toaster in `App.tsx`):
  `service/components/table/Bestellung.tsx` („Bestellung wurde
  aufgenommen.“), `service/components/table/Zahlung.tsx` („Zahlung
  erfolgreich.“), `service/components/direktverkauf/Direktverkauf.tsx`
  („Verkauf abgeschlossen.“) — danach feuern die reload-Callbacks
  (`TablePage.reload` bzw. `reloadHistorie` in `DirektverkaufPage.tsx`).
  Fehler-Toasts zentral in `hooks/use-action-submit.ts — useActionSubmit`.
- `frontend/src/service/components/Stepper.tsx — Stepper` — 44-px-Buttons
  über `ui/button`; `service/components/table/DockActionButton.tsx` —
  Mengen-Pill im Dock-Aktionsbutton.
- `frontend/src/service/TableSelectionPage.tsx` — Tischkarten-Liste
  (`TischGruppe` → `MeinTischCard`), kein Polling, Refetch nur explizit.
- `frontend/src/lib/utils.ts — formatCents` — Cent→EUR-Formatierung.
- Druckpfad: `admin/reporting/ReportingResults.tsx` (einziger
  `window.print()`-Aufruf, Bericht selbst) +
  `admin/reporting/KassenberichtePage.tsx` (`print:hidden`-Wrapper);
  `AdminSidebar`/`AdminMobileHeader` sind bereits `print:hidden`.
- ADRs: `docs/adrs/05_spektral-branding-website.md` (Frontend-Ausnahme:
  „Das Produkt-Frontend (`frontend/`) behält sein bestehendes,
  zurückhaltendes Design“), `docs/adrs/README.md` (Nygard-Format,
  Namenskonvention `NN_<kebab-case>.md`, Ablösung per neuem ADR).
- Tests: `service/beleg.test.ts` (einziger `vi.useFakeTimers`-Präzedenzfall
  im Frontend), `service/TablePage.test.tsx`, `Zahlung.test.tsx` +
  `ZahlungDrawer.test.tsx` (prüft `toast.error` im Fehlerfall),
  `Bestellung.test.tsx` + `BestellungDrawer.test.tsx`,
  `Direktverkauf.test.tsx`, `admin/AdminSidebar.test.tsx` (StatusDot via
  `getByRole('img', …)`), `admin/reporting/LiveReportingSection.test.tsx`,
  `admin/kasse/KassensitzungPage.test.tsx`. Kein Test assertet
  `toast.success` in den drei Buchungsflows.

## Resolved decisions

- PRD-Klärungen mit dem Betreiber (2026-07-14) gelten unverändert: voller
  Handoff-Umfang; Erfolgs-Pop ersetzt Erfolgs-Toasts im Service inklusive
  Direktverkauf, mit Tap-to-dismiss; kein Fortschrittsbalken (bestätigt:
  im Frontend existiert keine Progress-Komponente); Login-Glow-Drift wird
  umgesetzt; Doku bleibt bei Inter.
- Erfolgs-Pop nur in den drei Buchungsflows (Bestellen, Kassieren,
  Direktverkauf). Die übrigen Service-Erfolgs-Toasts (Umbuchung „Bestellung
  umgebucht.“, Belegdruck in `service/beleg.ts`) und sämtliche
  Verwaltungs-Toasts bleiben; Fehler-Toasts (`useActionSubmit`) unverändert.
  Die Pop-Texte übernehmen die bestehenden Toast-Texte.
- Wortmarke zusätzlich zu den drei PRD-Einsatzorten auch in `PasswordForm`,
  `ErrorPage` und `HydrateFallbackPage`: alle drei rendern heute das
  wortgleiche h1-Markup wie der Login, und `PasswordPage`/`ErrorPage` teilen
  das `AuthLayout` (bekommen dessen Glows automatisch) — eine schwarze
  Wortmarke neben spektralen Glows wäre ein sichtbarer Konsistenzbruch. Die
  Footer-Versionszeile der Sidebar bleibt schlichter Text (Versionsangabe,
  keine Wortmarke).
- Beträge und Kennzahl-Werte bleiben Inter, obwohl der Prototyp sie in Space
  Grotesk zeigt: PRD („Inter bleibt UI- und Fließtextschrift“) und die
  Typografie-Begründung des Handoffs selbst (tabellarische Ziffern für
  Kassenbeträge) schlagen das Prototyp-Detail.
- Glows nur am Login und hinter den Admin-Seitenköpfen; die im
  Handoff-README zusätzlich erwähnten Glows in Leer-/Erfolgszuständen
  entfallen (die Vier-Stellen-Regel der PRD zählt sie nicht auf).
- Der im Handoff als optional markierte Mengen-Ziffer-Mini-Pop am Stepper
  entfällt (nicht im PRD-Motion-Inventar); der Mengen-Pill-Pop am
  `DockActionButton` bleibt (explizites Handoff-Delta, 250 ms).
- ADR 06 (`docs/adrs/06_spektral-branding-app.md`) löst die
  Frontend-Ausnahme aus ADR 05 ab; ADR 05 bleibt akzeptiert und erhält nur
  einen Ablöse-Vermerk, die README-Tabelle wird ergänzt.
- Das Handoff-Bundle wird erst nach bestandener Sichtabnahme gelöscht — es
  ist die Abnahme-Referenz (Prototyp + Screenshots).

## Open questions / Risks

- Tailwind-4-Mechanik für die benannten Animations-Utilities
  (`--animate-*`-Theme-Tokens vs. `@layer utilities`) wird in Phase 1
  festgelegt; Keyframes müssen global registriert sein, sonst greifen
  Animations-Utilities nicht (Handoff-Hinweis).
- jsdom hat kein echtes rAF-Timing: `useCountUp` liefert ohne
  Animationsumgebung sofort den Endwert (PRD-Testentscheidung);
  `ErfolgsPop`-Tests laufen mit `vi.useFakeTimers` (Präzedenz
  `service/beleg.test.ts`).
- Glow-Kontrast im Dark Mode: Deckkraft ggf. innerhalb der
  Handoff-Bandbreite anheben (.22–.28); Leitplanke bleibt AA für Text über
  Glows, Sichtprüfung je Seite in Hell und Dunkel.
- Der Pop verzögert den Refetch um bis zu ~1,4 s; Tap-to-dismiss löst ihn
  sofort aus und begrenzt das Fenster. Schnelle Folgeaktionen sehen bis
  dahin den alten Stand — bewusste PRD-Entscheidung (Statuswechsel erst
  nach dem Schließen).
- Unter reduzierter Bewegung stehen auch Bestands-Animationen still
  (Drawer-Übergänge, `animate-spin`-Spinner) — von der PRD gedeckt
  („sämtliche Animationen“), in Phase 1 einmal gegenprüfen, dass kein
  Zustand dadurch unverständlich wird.

---

## Phase 1: Fundament — Tokens, Animations-Inventar, Typografie

**User stories**: 13, 14

### Context

- `frontend/src/index.css` — `@theme inline` (`--font-heading`), `:root`/`.dark`-Token-Blöcke
- `website/src/styles/brand.css` — Quell-Tokens (wertidentisch übernehmen)
- `frontend/package.json` — neue Dependency `@fontsource-variable/space-grotesk`
- `frontend/src/admin/components/AdminPageHeader.tsx` — h1 bekommt `font-heading`
- `frontend/src/service/TablePage.tsx` — Tischtitel-h1 bekommt `font-heading`

### What to build

Das gesamte Fundament, von dem alle Folgephasen nur konsumieren: die
`--sp-*`-Töne und `--spectral` (Light in `:root`, Dark-Overrides in `.dark`,
Dark-Verlauf aus effektiven Dark-Tönen rekomponiert) mit
Quellverweis-Kommentar; die fünf Keyframes und benannten
Animations-Utilities; die zentrale Reduced-Motion-Regel; Space Grotesk als
Heading-Schrift (Paket installieren, in `index.css` importieren,
`--font-heading` umhängen, `font-heading` an den zwei Seiten-h1 ergänzen).
Sichtbarer Effekt dieser Phase ist ausschließlich der Schriftwechsel der
Überschriften und Kartentitel.

### Acceptance criteria

- [x] `--sp-*` und `--spectral` stehen in `:root` und `.dark` wertidentisch
      zur Website (Dark-Verlauf rekomponiert), mit Quellkommentar auf
      `website/src/styles/brand.css`.
- [x] Keyframes `shimmer`/`fadeUp`/`pop`/`pulsedot`/`drift` und die benannten
      Utilities sind zentral in `index.css` definiert und in Komponenten
      verwendbar.
- [x] Die zentrale `prefers-reduced-motion`-Regel nullt Animationen und
      Transitions global (DevTools-Emulation: auch bestehende Drawer-/
      Spinner-Animationen stehen).
- [x] `CardTitle`, Dialog-/Drawer-Titel, `AdminPageHeader`-h1 und
      Tischtitel rendern Space Grotesk; Fließtext, Buttons, Beträge und die
      Doku bleiben Inter.
- [x] Kein weiterer visueller Diff; `make check` grün.

---

## Phase 2: Wortmarke

**User stories**: 7 (Wortmarke), 14

### Context

- `frontend/src/components/common/Wortmarke.tsx` — neu
- Einsatzorte: `components/common/LoginForm.tsx`,
  `components/common/PasswordForm.tsx`, `pages/HydrateFallbackPage.tsx`,
  `pages/ErrorPage.tsx`, `admin/AdminSidebar.tsx` (SidebarHeader),
  `admin/AdminLayout.tsx — AdminMobileHeader`

### What to build

Wiederverwendbare `Wortmarke`-Komponente: „jotti“ als echter, selektierbarer
Text in Space Grotesk 700 mit Spektralverlauf als Text-Füllung
(`bg-clip-text` + transparente Textfarbe, Verlauf aus `--spectral`), Größe
über className der Einsatzstelle. Die Überschriften-Semantik der
Einsatzstellen bleibt erhalten (wo heute ein h1 steht, bleibt ein h1).
Einsatz an allen sechs Stellen: Login-Karte (38 px-Äquivalent), Passwort-,
Fehler- und Hydrate-Seite identisch zum Login, Admin-Sidebar (26 px),
mobiler Admin-Kopf (bestehende Größe).

### Acceptance criteria

- [x] „jotti“ erscheint an allen sechs Stellen mit Spektralverlauf in Space
      Grotesk und bleibt als Text im DOM auffindbar (kein Bild).
- [x] Login-, Passwort-, Fehler- und Hydrate-Seite sind optisch identisch
      behandelt.
- [x] Dark Mode nutzt den Dark-Verlauf (automatisch über das Token).
- [x] Bestehende Tests, die auf den Text „jotti“ matchen, laufen unverändert;
      die Footer-Versionszeile der Sidebar ist unverändert schlichter Text.
- [x] `make check` grün.

---

## Phase 3: Hintergrund-Glows

**User stories**: 7 (Glows), 8, 12

### Context

- `frontend/src/components/common/AuthLayout.tsx — AuthLayout` — drei Login-Glows inline
- `frontend/src/admin/components/HeaderGlow.tsx` — neu (Ellipsen-Paar)
- `frontend/src/admin/components/AdminPageHeader.tsx` — Integrationspunkt mit Farbpaar-Prop
- Farbpaar-Zuordnung an den acht Aufrufern (siehe Architectural decisions)
- `frontend/src/admin/reporting/KassenberichtePage.tsx` / `ReportingResults.tsx` — Druck-Gegenprobe

### What to build

Login: Wrapper des `AuthLayout` wird `relative overflow-hidden`, dahinter
drei dekorative, stark geblurte Farbkreise (teal 320 px / blur 60 / Opacity
.25 oben links, violett 340 px / blur 64 / .20 unten rechts, orange 200 px /
blur 56 / .18 mittig rechts) mit langsamer Drift (translate ±16 px, 14–22 s,
ease-in-out, alternierend, je Kreis versetzt). Admin: `HeaderGlow` als
Ellipsen-Paar (blur 52, Opacity .18, ca. 460×200 px) hinter dem
`AdminPageHeader`, integriert über die Farbpaar-Prop mit Default
teal+violett; Farbpaare nach Mapping. Alle Glows sind `aria-hidden`,
`pointer-events-none` und `print:hidden`.

### Acceptance criteria

- [x] Login (und über das `AuthLayout` auch Passwort- und Fehlerseite) zeigt
      die drei driftenden Glows hinter der Karte; unter reduzierter Bewegung
      steht die Drift.
- [x] Alle acht Admin-Seiten zeigen den Seitenkopf-Glow mit dem Farbpaar aus
      dem Mapping.
- [x] Glows sind für Screenreader und Accessibility-Queries unsichtbar und
      klickdurchlässig; bestehende Tests grün.
- [x] Kontrast-Stichprobe: Text über Glows erfüllt AA in Hell und Dunkel
      (Deckkraft ggf. innerhalb der Handoff-Bandbreite justiert).
- [x] Druckvorschau der Kassenberichte ist frei von Glows.
- [x] Kein Layout-Shift und keine horizontalen Scrollbalken durch
      Glow-Überhang.
- [x] `make check` grün.

---

## Phase 4: Ladezustände — Spektral-Shimmer

**User stories**: 6

### Context

- `frontend/src/index.css` — `skeleton-shimmer`-Utility (Gradient-Rezept aus dem Handoff)
- `frontend/src/components/ui/skeleton.tsx — Skeleton` — `animate-pulse` ersetzen
- Verwendungsstellen zur Sichtung: siehe Inventory (Skeleton)

### What to build

Die `skeleton-shimmer`-Utility nach Handoff-Rezept (Verlauf aus `--muted`
mit leichter teal-/violett-Tönung via `color-mix`, `background-size: 220%`,
Keyframe `shimmer` 1,6 s linear infinite) in `index.css`; `ui/skeleton.tsx`
tauscht `animate-pulse` dagegen aus. Damit shimmern alle Skeletons der App
einheitlich spektral — keine Änderungen an den Verwendungsstellen nötig.

### Acceptance criteria

- [x] Die Skeleton-Komponente shimmert spektral getönt in Hell und Dunkel;
      `grep animate-pulse frontend/src` liefert keine Treffer mehr.
- [x] Alle Verwendungsstellen gesichtet: `TischListSkeleton`,
      `ProductListSkeleton` (Bestellen + Direktverkauf),
      `HistorieRowSkeleton` (beide Historien), `EigeneUebersicht`,
      `SidebarMenuSkeleton`.
- [x] Unter reduzierter Bewegung steht der Shimmer (statischer Platzhalter
      bleibt sichtbar).
- [x] `make check` grün.

---

## Phase 5: Hairline-Akzente

**User stories**: 8, 10

### Context

- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Hero-Karte „Kassierter Umsatz“
- `frontend/src/components/ui/sidebar.tsx — SidebarMenuButton` bzw. `admin/AdminSidebar.tsx — NavGroup` — aktiver Nav-Eintrag (`data-active`)

### What to build

Zwei gedämpfte Spektral-Hairlines: eine 2-px-Linie als oberste Kante der
Hero-Kennzahlkarte (dekoratives absolut positioniertes Element, Verlauf aus
`--spectral`, Opacity .6, Karte wird `relative overflow-hidden`; der Wert
bleibt `whitespace-nowrap tabular-nums`) und ein 3-px-Marker links am
aktiven Sidebar-Navigationseintrag (oben/unten 8 px eingerückt, abgerundet,
Opacity .6), ausgelöst über den bestehenden `data-active`-Zustand.

### Acceptance criteria

- [x] Die Hero-Karte trägt die Spektral-Kante; die vier Nebenkarten und alle
      übrigen Karten nicht.
- [x] Der aktive Sidebar-Eintrag zeigt den Marker zusätzlich zum bestehenden
      Hintergrund; alle Nav-Einträge geprüft, Icons und Text unverschoben.
- [x] Beide Akzente sind dekorativ (`aria-hidden`), Hell und Dunkel geprüft;
      im Druckpfad der Kassenberichte tauchen sie nicht auf.
- [x] `make check` grün.

---

## Phase 6: Micro-Feedback — Press & Fokus

**User stories**: 4

### Context

- `frontend/src/components/ui/button.tsx — buttonVariants` — bestehendes `active:translate-y-px`, `transition-all`
- `frontend/src/service/components/Stepper.tsx — Stepper`
- `frontend/src/service/components/table/DockActionButton.tsx` — Mengen-Pill

### What to build

Press-Feedback nach Motion-Inventar: alle Buttons senken und skalieren beim
Drücken leicht (bestehendes `translate-y-px` um `scale` ~.99 ergänzt,
Icon-Größen ~.96, 100 ms transform), die Stepper-Buttons deutlich stärker
(scale ~.92). Die Mengen-Pill im Dock-Aktionsbutton poppt bei
Mengenänderung (Remount über `key` + `pop`, 250 ms gemäß Handoff-Delta).
Hover-/Fokus-Transitions mit 150 ms sind über `transition-all` an den
Buttons bereits vorhanden; ergänzt wird nur die Einblendung des Fokus-Rings
an den Inputs (Transition auf `box-shadow` und `border-color`, 150 ms,
Handoff-Delta C).

### Acceptance criteria

- [x] Buttons geben beim Drücken spürbares Press-Feedback, Stepper stärker;
      Disabled-Verhalten unverändert (keine `pointer-events`-Regression).
- [x] Die Mengen-Pill im Dock-Button poppt bei Änderung.
- [x] Unter reduzierter Bewegung wechseln Zustände sofort, ohne Übergänge.
- [x] Bestehende Button-/Stepper-/Drawer-Tests grün; `make check` grün.

---

## Phase 7: Übergänge & Statuswechsel

**User stories**: 5

### Context

- `frontend/src/components/ui/tabs.tsx — TabsContent`
- `frontend/src/service/TableSelectionPage.tsx` — Tischkarten-Liste (`TischGruppe` → `MeinTischCard`)
- `frontend/src/service/components/table/Zahlung.tsx` — Positionsliste (`PositionItem`)
- `frontend/src/service/TablePage.tsx` — Status-Badge (`n unbezahlt` / `Alles bezahlt`), `tabsLocked`
- `frontend/src/admin/kasse/KassensitzungPage.tsx — StepperRow, EroeffnetKarte`

### What to build

Sanfte Übergänge nach Motion-Inventar: `TabsContent` blendet beim
Aktivieren mit fadeUp (250 ms) ein. Listen-Eintritt mit fadeUp (450 ms,
Stagger 60 ms) ausschließlich beim ersten Aufbau — ein Erst-Mount-Flag
stellt sicher, dass Daten-Refetches nie erneut animieren; Einsatz an der
Tischkarten-Liste und der Zahlungs-Positionsliste. Statuswechsel poppen
(350 ms): das Status-Badge der Tischansicht bei Wertwechsel, und nach
„Kasse eröffnen“ erscheint die `EroeffnetKarte` mit fadeUp, ihr Häkchen mit
Pop.

### Acceptance criteria

- [x] Tab-Wechsel in Tischansicht und Direktverkauf blenden sanft ein;
      `tabsLocked`-Verhalten unverändert.
- [x] Tischkarten und Zahlungs-Positionen staggern nur beim ersten Aufbau;
      ein Refetch (z. B. nach einer Buchung) animiert nicht erneut.
- [x] Badge-Wechsel poppt; Texte und Semantik unverändert (Tests grün).
- [x] Nach dem Eröffnen der Kasse erscheint die erledigt-Karte mit fadeUp,
      das Häkchen poppt.
- [x] Unter reduzierter Bewegung erscheint alles sofort, ohne Flackern.
- [x] `make check` grün.

---

## Phase 8: Erfolgs-Pop

**User stories**: 1, 2

### Context

- `frontend/src/service/components/ErfolgsPop.tsx` — neu
- `frontend/src/service/components/table/Bestellung.tsx` / `Zahlung.tsx` /
  `direktverkauf/Direktverkauf.tsx` — Erfolgs-Callbacks (heute `toast.success`)
- `frontend/src/service/TablePage.tsx — reload()` und
  `frontend/src/service/DirektverkaufPage.tsx` (reloadHistorie) — nachgelagerte Refetches
- `frontend/src/service/beleg.test.ts` — Fake-Timer-Präzedenz

### What to build

`ErfolgsPop` als unübersehbares Overlay über dem Screen: geblurter
Backdrop, 76-px-Kreis in Primärgrün mit weißem Häkchen und Ring-Schatten,
Spring-Pop (450 ms), Text darunter mit verzögertem fadeUp; gesteuert über
`{ open, text }`, Auto-Dismiss nach ~1,4 s plus Tap-to-dismiss (gesamtes
Overlay antippbar), `role="status"`, damit Screenreader die Bestätigung wie
bisher den Toast ankündigen. Gehostet je einmal in `TablePage` (Bestellen +
Kassieren) und `DirektverkaufPage` (Direktverkauf). Die drei
Erfolgs-Callbacks öffnen den Pop statt `toast.success` zu feuern; die
nachgelagerten reload-/refetch-Callbacks laufen erst beim Dismiss (Auto
oder Tap), sodass sichtbare Statuswechsel dem Pop folgen. Die Texte bleiben
die bisherigen Toast-Texte.

### Acceptance criteria

- [x] Pop erscheint nach erfolgreichem Bestellen, Kassieren und
      Direktverkauf; in diesen drei Flows feuert kein Erfolgs-Toast mehr.
- [x] Auto-Dismiss nach Ablauf der Anzeigedauer, Tap schließt sofort; der
      nachgelagerte Statuswechsel/Refetch wird erst nach dem Schließen
      ausgelöst — getestet mit `vi.useFakeTimers`.
- [x] Fehlerpfad unverändert (`toast.error`, Drawer bleibt offen; bestehende
      `ZahlungDrawer`-Tests grün).
- [x] Umbuchungs-, Storno-, Belegdruck- und Verwaltungs-Toasts unverändert.
- [x] Unter reduzierter Bewegung erscheint der Pop ohne Animation und
      verschwindet nach Ablauf bzw. Tap.
- [x] `make check` grün.

---

## Phase 9: Zahlen zählen

**User stories**: 3, 11

### Context

- `frontend/src/hooks/use-count-up.ts` — neu (`useCountUp`)
- `frontend/src/service/TablePage.tsx` — `data-slot="tisch-saldo"`
- `frontend/src/admin/reporting/LiveReportingSection.tsx` + `SummaryCard.tsx` — Hero-Wert und vier Nebenkarten
- `frontend/src/lib/utils.ts — formatCents`

### What to build

`useCountUp`: animiert einen numerischen Wert bei Änderung über 700 ms
(ease-out-cubic, `requestAnimationFrame`) vom alten zum neuen Wert — nicht
beim ersten Rendern; ohne Animationsumgebung (reduzierte Bewegung, fehlendes
rAF, Testumgebung) erscheint sofort der Endwert. Einsatz am Tisch-Saldo
(hat bereits `tabular-nums`) und an den Übersicht-Kennzahlen (Hero-Wert und
die vier Nebenkarten; nur numerische Werte, Formatierung weiterhin über
`formatCents` bzw. die bestehende Darstellung, `tabular-nums` gegen
Layout-Shift).

### Acceptance criteria

- [x] Der Tisch-Saldo zählt bei Änderung animiert und endet exakt am
      Zielwert (Hook-Test).
- [x] Die Übersicht-Kennzahlen zählen beim 30-s-Refetch bzw.
      „Jetzt“-Refresh nur bei tatsächlicher Wertänderung; kein Layout-Shift.
- [x] Ohne Animationsumgebung liefert der Hook sofort den Endwert (Test).
- [x] Bestehende `TablePage`-/`LiveReportingSection`-Tests bleiben grün
      (Endzustände unverändert).
- [x] `make check` grün.

---

## Phase 10: Live-Puls, ADR & Gesamtabnahme

**User stories**: 9, 12, 13

### Context

- `frontend/src/admin/components/StatusDot.tsx — StatusDot`
- `frontend/src/admin/AdminSidebar.tsx` — Kassentag-Chip (`kasseOffen`)
- `frontend/src/admin/reporting/LiveReportingSection.tsx` — Live-Zeile
- `docs/adrs/05_spektral-branding-website.md`, `docs/adrs/README.md` — ADR-Format und -Register

### What to build

`StatusDot` erhält eine optionale `puls`-Prop (Keyframe `pulsedot`, 2,4 s
ease-out infinite als weicher Ring), aktiviert an genau zwei Stellen: am
Kassentag-Chip der Sidebar (nur bei offener Kasse) und am Live-Punkt der
Übersicht; alle übrigen StatusDots bleiben statisch. ADR 06
(`docs/adrs/06_spektral-branding-app.md`, Nygard-Format) hält die Ablösung
der Frontend-Ausnahme aus ADR 05 fest, inklusive der Vier-Stellen-Regel als
neuer Leitplanke und Verweis auf das Motion-Inventar der PRD; ADR 05
bekommt den Ablöse-Vermerk, die README-Tabelle den neuen Eintrag.
Abschließend der Gesamtdurchgang: Sichtvergleich aller Screens gegen den
Prototyp in Hell und Dunkel (Playwright-Screenshots als Abnahmegrundlage
für den Betreiber), Reduced-Motion-Gegenprobe per DevTools-Emulation,
Druck-Stichprobe der Kassenberichte, volle Prüfung.

### Acceptance criteria

- [x] Der Kassentag-Chip pulsiert nur bei offener Kasse, der Live-Punkt der
      Übersicht pulsiert; alle übrigen StatusDots statisch; `aria-label` und
      Rolle unverändert (`AdminSidebar`-Tests grün).
- [x] Unter reduzierter Bewegung steht der Puls.
- [x] `docs/adrs/06_spektral-branding-app.md` geschrieben, ADR 05 mit
      Ablöse-Vermerk, README-Tabelle aktualisiert.
- [x] Screenshot-Satz aller Screens in Hell und Dunkel für die Sichtabnahme
      erstellt; Reduced-Motion- und Druck-Gegenprobe durchgeführt und im
      Chat dokumentiert.
- [x] `make verify` grün.
- [ ] Nach bestandener Sichtabnahme durch den Betreiber (Human-Gate):
      Handoff-Bundle `docs/prds/design_handoff_spektral_redesign/` löschen.
