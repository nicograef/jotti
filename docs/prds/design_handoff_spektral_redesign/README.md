# Handoff: Spektral-Redesign des jotti App-Frontends

## Overview
Überträgt die Spektral-Markensprache der neuen jotti-Website (`website/src/styles/brand.css`) **subtil und dezent** auf das App-Frontend (`frontend/`): Login, Service-Bereich (Meine Tische, Tischansicht), Admin-Bereich (Übersicht, Kassentag, Produkte, Benutzer) und die gemeinsamen Komponenten in `common/` bzw. `components/`. Zusätzlich definiert es ein Motion-Inventar aus subtilen UX-Animationen und vereinheitlicht die Typografie (Inter + Space Grotesk) über App, Website und Doku.

Begleitdokumente in diesem Ordner:
- **PRD.md** — Product Requirements Document (Ziele, Anforderungen, Akzeptanzkriterien)
- **UMBAU-PLAN.md** — phasenweiser Implementierungsplan für Claude Code

## About the Design Files
Die Datei `jotti Spektral-Redesign.dc.html` (+ `ios-frame.jsx`, `support.js`) ist eine **Design-Referenz in HTML** — ein Prototyp, der Look und Verhalten zeigt, **kein Produktionscode**. Aufgabe ist es, diese Designs **im bestehenden Frontend** nachzubauen: React 19 + TypeScript + Tailwind CSS 4 + shadcn/ui (Verzeichnis `frontend/`), unter Beibehaltung der dortigen Muster (CVA-Varianten, CSS-Custom-Properties in `index.css`, `cn()`-Utility, Radix-Primitives).

## Fidelity
**High-fidelity.** Farben, Typografie, Abstände und Interaktionen sind final gemeint. Die Screens wurden aus dem echten Quellcode (`frontend/src/**`) rekonstruiert; Abweichungen im Prototyp (vereinfachte Icons, statische Daten) sind irrelevant — maßgeblich sind bestehende App-Layouts **plus die unten beschriebenen Deltas**. Es ist ein *Restyling-Layer*, kein Layout-Umbau: Bestehende Komponentenstruktur, Semantik und WCAG-Entscheidungen (AA-Kontraste, Disabled-Token, Warn-Amber) bleiben unangetastet.

## Kernprinzip: Wo das Spektrum auftritt — und wo nicht
**Ja (nur hier):**
1. **Wortmarke „jotti“** — Spektralverlauf als Text-Fill (Login-Karte, Admin-Sidebar, mobiler Admin-Header)
2. **Hintergrund-Glows** — weiche, stark geblurte Radial-Gradients hinter Seitenköpfen (Admin), hinter der Login-Karte und in leeren/Erfolgs-Zuständen
3. **Ladezustände** — Skeleton-Shimmer mit leichter Spektral-Tönung; Fortschrittsbalken (Export, TSE-Setup) mit Spektral-Fill
4. **Hairline-Akzente** — 2 px Spektral-Linie oben auf der Hero-Kennzahlkarte („Kassierter Umsatz“); 3 px Spektral-Marker links am aktiven Sidebar-Nav-Eintrag

**Nein (bleibt wie heute):** Buttons, Badges, Status-Semantik (primär-Grün / destruktiv-Rot / warn-Amber), Fließtext, Icons, Diagramme. Emerald `--primary` bleibt die Primärfarbe.

## Screens / Views

### 1. Login (`pages/LoginPage.tsx`, `common/AuthLayout.tsx`, `common/LoginForm.tsx`)
- **Layout unverändert:** zentrierte Karte `max-w-sm` auf `bg-primary/5`, Footer „Entwickelt von Nico Gräf“.
- **Delta A — Wortmarke:** `<h1>jotti</h1>` (bisher `text-4xl font-extrabold`) wird 38 px, `Space Grotesk` 700, mit Spektral-Text-Fill:
  `background: var(--spectral); background-clip: text; -webkit-background-clip: text; color: transparent;`
- **Delta B — Glows:** Im `AuthLayout` drei absolut positionierte, dekorative Kreise (aria-hidden, `pointer-events:none`):
  - teal: 320×320 px, oben links versetzt, `radial-gradient(circle, color-mix(in oklab, var(--sp-teal) 55%, transparent), transparent 65%)`, `filter: blur(60px)`, Opacity `0.25`
  - violet: 340×340 px, unten rechts, sp-violet 45 %, blur 64 px, Opacity `0.20`
  - orange: 200×200 px, mittig rechts, sp-orange 40 %, blur 56 px, Opacity `0.18`
  - Optionale Drift-Animation: `translate` ±16 px, 14–22 s, ease-in-out, infinite, alternierend; unter `prefers-reduced-motion` aus.
- **Delta C — Fokus-Ring:** Inputs behalten den Olive-Ring (`--ring`), Einblendung mit `transition: box-shadow .15s, border-color .15s`.

### 2. Service · Meine Tische (`service/TableSelectionPage.tsx`, `MeinTischCard.tsx`, `EigeneUebersicht.tsx`)
- **Layout unverändert** (Header 56 px, Übersicht-Karten, Suche h-11, Gruppen „Noch offen“/„Erledigt“, fixe Bodenleiste „Alle Tische“).
- **Delta — Listen-Eintritt:** Tischkarten erscheinen mit `fadeUp` (opacity 0→1, translateY 7px→0), 450 ms, `cubic-bezier(.2,.7,.3,1)`, Stagger 60 ms pro Karte, `animation-fill-mode: both`. Nur beim ersten Mount der Liste (nicht bei Refetch).
- **Delta — Skeleton:** `ui/skeleton.tsx` erhält Spektral-Shimmer (siehe Design Tokens → Shimmer).

### 3. Service · Tischansicht (`service/TablePage.tsx`, `ProductList.tsx`, `Stepper.tsx`, `Zahlung.tsx`, `ServiceDock.tsx`)
- **Layout unverändert** (Titel 22 px, Badges, Saldo rechts, ServiceDock mit Aktionsbutton + TabsList).
- **Delta — Tab-Wechsel:** `TabsContent` blendet mit `fadeUp` 250 ms ease ein.
- **Delta — Stepper-Press:** Plus/Minus-Buttons `active:scale-[.92]`, `transition: transform .1s`. Menge-Ziffer bei Änderung mit Mini-Pop (scale .5→1, 250 ms) — optional, per `key={menge}` remounten.
- **Delta — Dock-Button:** Mengen-Pill im Aktionsbutton poppt bei Änderung (250 ms). Button-Press `active:translate-y-px active:scale-[.99]`.
- **Delta — Erfolgs-Pop beim Kassieren/Bestellen:** Nach erfolgreicher Buchung Overlay über dem Screen: Backdrop `color-mix(in oklab, var(--background) 55%, transparent)` + `backdrop-filter: blur(6px)`; Kreis 76 px `bg-primary`, weißes Check-Icon 34 px, Ring-Schatten `0 0 0 8px primary/15`; Animation `pop`: scale .5→1.07→1 mit `cubic-bezier(.34,1.56,.64,1)` (Spring), 450 ms; Text darunter (17 px/600) mit fadeUp 350 ms delay 100 ms. Auto-dismiss nach ~1,4 s, **danach** Statuswechsel (Badge, Listen). Ersetzt/ergänzt den bisherigen sonner-Toast im Service-Flow.
- **Delta — Saldo zählt animiert:** `data-slot="tisch-saldo"` animiert bei Änderung numerisch von alt→neu, 700 ms, ease-out-cubic (`1-(1-p)^3`), via rAF; `font-variant-numeric: tabular-nums` verhindert Layout-Shift.
- **Delta — Statuswechsel Badge:** neuer Badge-Wert mounted mit `pop` 350 ms.

### 4. Admin · Sidebar (`admin/AdminSidebar.tsx`, `ui/sidebar.tsx`)
- **Delta — Wortmarke:** wie Login, 26 px Space Grotesk 700, Spektral-Fill.
- **Delta — aktiver Nav-Eintrag:** zusätzlich zum bestehenden `bg-sidebar-accent` ein absolut positionierter Marker: `left:0; top:8px; bottom:8px; width:3px; border-radius:2px; background:var(--spectral); opacity:.6`.
- **Delta — Kasse-offen-Punkt:** StatusDot „ok“ in der Kassentag-Karte pulsiert: `pulsedot` 2,4 s ease-out infinite (box-shadow 0→7px, primary 50 %→0). Nur der Punkt in Sidebar-Karte und „Live“-Anzeige, nicht die Nav-Dots.

### 5. Admin · Übersicht (`reporting/LiveReportingSection.tsx`, `SummaryCard.tsx`, `UebersichtStatusZeile.tsx`)
- **Layout unverändert** (AdminPageHeader, 3er-Statuszeile, 5er-Kennzahlen-Grid, Offene Tische/Team, Storno-Collapsible).
- **Delta — Header-Glow:** hinter dem `AdminPageHeader` ein dekoratives Element: ca. 460×200 px, zwei Ellipsen-Gradients (sp-teal 40 % @ 20 %/40 %, sp-violet 28 % @ 70 %/60 %), `blur(52px)`, Opacity `0.18`, `pointer-events:none`, hinter dem Text (Header-Wrapper `position:relative`). Pro Admin-Seite eine eigene, dezent variierte Farbkombination (Übersicht teal+violet, Kassentag orange+teal, Produkte green+blue, Benutzer blue+red).
- **Delta — Hero-Karte:** „Kassierter Umsatz“ erhält 2 px Spektral-Hairline als oberste Kante (`position:absolute; inset:0 0 auto 0; height:2px; background:var(--spectral); opacity:.6`) + `overflow:hidden` auf der Karte. Wert nowrap, `tabular-nums`.
- **Delta — Live-Puls:** StatusDot neben „Live · aktualisiert …“ pulsiert (s. o.).
- **Delta — Zahlen-Update:** Beim Refetch zählen geänderte Kennzahlen animiert (700 ms ease-out), kein Fade.

### 6. Admin · Kassentag (`kasse/KassensitzungPage.tsx` + Sections)
- **Layout unverändert** (StepperRow-Schiene, Karten, Kacheln, Bewegungsliste, Warn-Button amber).
- **Delta — Header-Glow** (orange+teal, s. o.).
- **Delta — Statuswechsel „Kasse öffnet“:** Nach Eröffnung wechselt Schritt 1 zur „erledigt“-Karte; die neue Karte mounted mit fadeUp 450 ms, das Häkchen im Kreis mit `pop` 350 ms.

### 7. Admin · Produkte & Preise / Helfer & Zugänge
- **Layout unverändert** (Chip-Zeilen bzw. Tabellen-Grid, Switches, Icon-Buttons).
- **Delta — Header-Glow** (je eigene Farbkombination).
- **Delta — Micro-Feedback:** Icon-Buttons/Chips Hover `bg-muted`, Press `scale(.96)` 100 ms; Switch behält Radix-Transition.

## Interactions & Behavior — Motion-Inventar (verbindlich)
| Animation | Dauer | Easing | Einsatz |
|---|---|---|---|
| Button-Press | 100 ms | linear (transform) | alle Buttons: `translateY(1px) scale(.96–.99)` |
| Hover-Transitions | 150 ms | ease | bg/border/box-shadow |
| Fokus-Ring | 150 ms | ease | Inputs, Buttons |
| Tab-/Detail-Wechsel `fadeUp` | 250 ms | ease | TabsContent, Drawer-Inhalte |
| Statuswechsel `pop` | 350 ms | ease | Badges, Häkchen |
| Listen-Eintritt `fadeUp` | 450 ms, Stagger 60 ms | cubic-bezier(.2,.7,.3,1) | Karten-/Zeilenlisten beim ersten Mount |
| Erfolgs-Pop | 450 ms | cubic-bezier(.34,1.56,.64,1) | Kassieren/Bestellen-Bestätigung, auto-dismiss 1,4 s |
| Zahlen zählen | 700 ms | ease-out-cubic | Saldo, Kennzahlen |
| Skeleton-Shimmer | 1,6 s loop | linear | alle Skeletons |
| Live-Puls `pulsedot` | 2,4 s loop | ease-out | „Kasse offen“-/Live-Punkt |
| Glow-Drift | 14–22 s loop | ease-in-out | Login-Glows (optional) |

**Global:** `@media (prefers-reduced-motion: reduce)` deaktiviert alle Animationen/Transitions (Duration ≈ 0). Loop-Animationen (Shimmer, Puls, Drift) laufen über `animation-play-state`, damit sie zentral pausierbar sind.

## State Management
Kein neuer globaler State. Lokal:
- Erfolgs-Pop: `{ open: boolean, text: string }` + 1,4-s-Timeout; Status-Update (Badge/Listen) erst nach Dismiss.
- Zahlen-Animation: Hook `useCountUp(targetCents, 700)` (rAF, ease-out-cubic), rendert Zwischenwerte lokal.
- Listen-Stagger: nur beim ersten Mount (z. B. `useRef`-Flag), nicht bei Query-Refetch.

## Design Tokens

### Neue CSS-Custom-Properties (in `frontend/src/index.css`, `:root` + `.dark`)
Quelle: `website/src/styles/brand.css` — Werte identisch übernehmen.
```css
:root {
  --sp-red:#d24a2a; --sp-orange:#c8781e; --sp-yellow:#8f9a2c; --sp-green:#4f9636;
  --sp-teal:#1f9b8a; --sp-blue:#2f6fc4; --sp-indigo:#4a4fc0; --sp-violet:#8b3fc0;
  --spectral: linear-gradient(100deg,#d24a2a,#c8781e,#4f9636,#1f9b8a,#2f6fc4,#8b3fc0);
}
.dark {
  --sp-red:#e2603f; --sp-green:#5fb045; --sp-teal:#28b8a3; --sp-blue:#4b86dd; --sp-violet:#a457dd;
  --spectral: linear-gradient(100deg,#e2603f,#c8781e,#5fb045,#28b8a3,#4b86dd,#a457dd);
}
```
Alle bestehenden Tokens (`--primary` emerald, Olive-Neutrals, `--destructive`, `--warn`, `--disabled`, Radius `0.45rem`) bleiben unverändert.

### Keyframes
```css
@keyframes shimmer { 0%{background-position:150% 0} 100%{background-position:-50% 0} }
@keyframes fadeUp  { from{opacity:0;transform:translateY(7px)} to{opacity:1;transform:none} }
@keyframes pop     { 0%{opacity:0;transform:scale(.5)} 60%{opacity:1;transform:scale(1.07)} 100%{transform:scale(1)} }
@keyframes pulsedot{ 0%{box-shadow:0 0 0 0 rgb(16 185 129/.5)} 70%{box-shadow:0 0 0 7px transparent} 100%{box-shadow:0 0 0 0 transparent} }
@keyframes drift   { 0%,100%{transform:translate(0,0)} 50%{transform:translate(16px,-12px)} }
```

### Shimmer-Skeleton (ersetzt `animate-pulse` in `ui/skeleton.tsx`)
```css
background: linear-gradient(90deg, var(--muted) 20%,
  color-mix(in oklab, var(--sp-teal) 14%, var(--muted)) 40%,
  color-mix(in oklab, var(--sp-violet) 10%, var(--muted)) 55%,
  var(--muted) 80%);
background-size: 220% 100%;
animation: shimmer 1.6s linear infinite;
```

### Fortschrittsbalken (Export/TSE)
Track: `bg-muted`, Höhe 6 px, radius voll. Fill: `background: var(--spectral)`.

### Typografie (vereinheitlicht über App, Website, Doku)
- **Inter** (Variable, 100–900): UI & Body überall. Bereits in der App (`@fontsource-variable/inter`).
- **Space Grotesk** (Variable, 300–700): **nur** Wortmarke & Überschriften (H1/H2, CardTitle nutzt bereits `font-heading`). In `index.css`: `--font-heading: 'Space Grotesk Variable', 'Inter Variable', sans-serif` + `@fontsource-variable/space-grotesk` als Dependency.
- Begründung: beide SIL OFL (frei, auch kommerziell), selbst gehostet (kein Google-Request, DSGVO), Inter hat tabellarische Ziffern für Beträge. Keine dritte Schrift.

### Glow-Rezept (wiederverwendbar, z. B. `admin/components/HeaderGlow.tsx`)
`position:absolute; top:-70px; left:-40px; width:~420–460px; height:~180–200px;` zwei `radial-gradient(ellipse …)` aus je zwei `--sp-*`-Farben (28–40 % via color-mix), `filter: blur(52px)`, `opacity: .18`, `pointer-events:none`, `aria-hidden`, Container `position:relative; overflow:hidden` (bzw. Seiten-Wrapper). Print: `print:hidden`.

## Assets
Keine Bild-Assets. Icons: bestehende lucide-react. Fonts: `@fontsource-variable/inter` (vorhanden) + `@fontsource-variable/space-grotesk` (neu); Website hostet dieselben Fonts bereits unter `website/public/fonts/`.

## Files
- `jotti Spektral-Redesign.dc.html` — interaktiver Design-Prototyp (alle Screens, Light/Dark-Toggle oben links; Screen 1c interaktiv: bestellen, kassieren, Erfolgs-Pop, Saldo-Animation)
- `ios-frame.jsx`, `support.js` — Laufzeit-Hilfen des Prototyps (nur fürs Öffnen im Browser nötig)
- `PRD.md`, `UMBAU-PLAN.md` — siehe oben
- `PR-VORLAGE.md` — Branch-Name, PR-Titel/-Beschreibung und Checkliste für den Pull Request
- `screenshots/alle-screens-light.jpg`, `screenshots/alle-screens-dark.jpg` — Gesamtansicht aller Screens in beiden Themes
