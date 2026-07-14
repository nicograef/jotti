# Umbau-Plan: Spektral-Redesign `frontend/`

Phasenweise, jede Phase einzeln shipbar und testbar. Reihenfolge so gewählt, dass Fundament zuerst kommt und jede Folgephase nur konsumiert. Vor Start: README.md + PRD.md lesen, Prototyp `jotti Spektral-Redesign.dc.html` im Browser öffnen (Light/Dark-Toggle oben links, Screen 1c ist interaktiv).

## Phase 0 — Fundament (Tokens, Keyframes, Fonts)
**Dateien:** `frontend/src/index.css`, `frontend/package.json`
1. `--sp-*` + `--spectral` (Light in `:root`, Dark-Overrides in `.dark`) aus `website/src/styles/brand.css` übernehmen — Werte 1:1, Quellkommentar setzen.
2. Keyframes `shimmer`, `fadeUp`, `pop`, `pulsedot`, `drift` + `@media (prefers-reduced-motion: reduce){ *{animation-duration:.01ms!important; transition-duration:.01ms!important} }` in `@layer base`.
3. `npm i @fontsource-variable/space-grotesk`; Import in `index.css`; `--font-heading: 'Space Grotesk Variable', 'Inter Variable', sans-serif` im `@theme inline`-Block (ersetzt das bisherige `--font-heading: var(--font-sans)`).
4. Sichtprüfung: CardTitle (nutzt `font-heading`) rendert Space Grotesk; sonst nichts verändert.

**Test:** App startet, kein visueller Diff außer CardTitle-Font.

## Phase 1 — Wortmarke
**Neu:** `frontend/src/components/common/Wortmarke.tsx` — `<span>` mit `font-heading font-bold`, Spektral-Text-Fill (`bg-[image:var(--spectral)] bg-clip-text text-transparent`), Größe per Prop/className.
**Einsatz:** `LoginForm.tsx` (h1, text-4xl→38px), `AdminSidebar.tsx` (h1, 26px), `AdminLayout.tsx` AdminMobileHeader (14px bold).
**Test:** Text bleibt im DOM (Tests, die auf „jotti“ matchen, laufen weiter).

## Phase 2 — Glows
**Neu:** `frontend/src/components/common/Glow.tsx` (ein Radial-Kreis, Props: farbe/size/position/opacity, `aria-hidden`, `pointer-events-none`) und `frontend/src/admin/components/HeaderGlow.tsx` (Ellipsen-Paar hinter Seitenkopf, Props: `farben: [SpFarbe, SpFarbe]`, `print:hidden`).
1. `AuthLayout.tsx`: Wrapper `relative overflow-hidden`, drei Glows (teal 320/blur60/.25 oben links, violet 340/blur64/.20 unten rechts, orange 200/blur56/.18 mitte rechts), Karte `relative`.
2. Admin-Seiten: HeaderGlow über `AdminPageHeader`-Aufrufern — Übersicht teal+violet, Kassentag orange+teal, Produkte green+blue, Benutzer blue+red, Rest teal+violet.
**Test:** Kontrast-Stichprobe Header-Text (AA), Dark Mode prüfen, Druckansicht Kassenberichte ohne Glow.

## Phase 3 — Ladezustände
1. `ui/skeleton.tsx`: `animate-pulse bg-accent` → Spektral-Shimmer (Rezept im README). Eine Utility-Klasse `skeleton-shimmer` in `index.css` anlegen, da der Gradient zu lang für sinnvolles Inline-Tailwind ist.
2. **Neu:** `components/common/SpektralProgress.tsx` (Track `bg-muted` h-1.5 rounded-full, Fill `bg-[image:var(--spectral)]`), einsetzen wo Fortschritt existiert (TSE-Wizard, Exporte) — nur dort, wo heute schon ein Indikator ist.
**Test:** Alle Skeleton-Verwendungen (TischListSkeleton, ProductListSkeleton, HistorieRowSkeleton, EigeneUebersicht) sichten.

## Phase 4 — Hairlines & Sidebar-Marker
1. Übersicht-Hero-Karte (`LiveReportingSection.tsx`): Karte `relative overflow-hidden` + `<span aria-hidden className="absolute inset-x-0 top-0 h-0.5 bg-[image:var(--spectral)] opacity-60"/>`; Wert `whitespace-nowrap tabular-nums`.
2. `ui/sidebar.tsx` `SidebarMenuButton`: bei `data-active` Marker-Span (3 px, links, spectral, opacity .6) — via `before:`-Utilities oder Kind-Element in `AdminSidebar`s NavGroup.
**Test:** Aktiv-Zustand aller Nav-Einträge, keine Verschiebung des Icons.

## Phase 5 — Micro-Feedback (global)
1. `ui/button.tsx`: `active:not-aria-[haspopup]:translate-y-px` → zusätzlich `active:scale-[.99]` (Icon-Sizes `.96`); Transition ist mit `transition-all` schon da.
2. `Stepper.tsx`: Buttons `active:scale-[.92] transition-transform duration-100`.
3. `DockActionButton.tsx`: Mengen-Pill mit `key={anzahl}` + `animate-[pop_.25s_ease_both]`.
**Test:** Vitest der Button-/Stepper-Suites; kein `pointer-events`-Regressions bei disabled.

## Phase 6 — Übergänge & Statuswechsel
1. `ui/tabs.tsx` `TabsContent`: `data-[state=active]:animate-[fadeUp_.25s_ease_both]`.
2. Listen-Eintritt: In `TableSelectionPage` (MeinTischCard-Grid) und `Zahlung`-Positionsliste Stagger via `style={{animationDelay: `${i*60}ms`}}` + `animate-[fadeUp_.45s_cubic-bezier(.2,.7,.3,1)_both]`; Erst-Mount-Flag (useRef), damit Refetch nicht erneut animiert.
3. Badge-Statuswechsel (`TablePage`: „n unbezahlt“/„Alles bezahlt“): `key` auf Wert + pop 350 ms. Kassentag Schritt-1-Wechsel: fadeUp auf EroeffnetKarte, pop aufs Häkchen.
**Test:** Tab-Wechsel unter tabsLocked unverändert; keine Animation bei reinem Daten-Refetch.

## Phase 7 — Erfolgs-Pop & Zahlen zählen (Service)
1. **Neu:** `service/components/ErfolgsPop.tsx` — Overlay (Spec README/F9), gesteuert über `{open, text}`; Portal in den Screen-Container.
2. Einbauen in `ZahlungDrawer`-Erfolgspfad („Zahlung kassiert“) und `BestellungDrawer` („Bestellung aufgenommen“); dortige `toast.success`-Aufrufe im Service ersetzen; Status-Refetch/`onZahlungKassiert` erst nach Dismiss (1,4 s) auslösen bzw. UI-Wechsel verzögern.
3. **Neu:** `hooks/use-count-up.ts` — `useCountUp(cents, 700)`; einsetzen am Tisch-Saldo (`TablePage`, `data-slot="tisch-saldo"`) und in `LiveReportingSection`-Kennzahlen.
**Test:** TablePage-/Zahlung-Vitest anpassen (Timer mocken); rAF unter jsdom guarden (Fallback: sofortiger Endwert).

## Phase 8 — Live-Puls & Feinschliff
1. `StatusDot.tsx`: Prop `puls?: boolean` → `animate-[pulsedot_2.4s_ease-out_infinite]`; aktivieren in AdminSidebar-Kassentag-Karte und Übersicht-„Live“-Zeile.
2. Gesamtdurchgang gegen Prototyp (Light + Dark, alle Screens 1a–1g), Reduced-Motion-Check, Lint/Test/Build.

## Risiken / Hinweise
- **Tailwind 4 Arbitrary-Animations:** `animate-[fadeUp_…]` braucht die Keyframes aus Phase 0 (global, nicht in `@theme`). Alternativ Utilities in `@layer utilities` definieren (`.anim-fade-up` etc.) — konsistenter bei Mehrfachnutzung.
- **jsdom-Tests:** rAF/`performance.now` für useCountUp mocken; ErfolgsPop-Timeout mit `vi.useFakeTimers`.
- **`background-clip:text`:** benötigt `-webkit-`-Prefix (Tailwind `bg-clip-text` erledigt das).
- **Glows im Dark Mode:** Opacity ggf. leicht anheben (.22–.28), Sichtprüfung.
- **Doppel-Feedback vermeiden:** Wo ErfolgsPop läuft, keinen Toast mehr feuern.
