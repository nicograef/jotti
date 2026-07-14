# PRD: Spektral-Redesign des jotti App-Frontends

**Status:** Entwurf zur Umsetzung · **Bereich:** `frontend/` (React 19, TS, Tailwind 4, shadcn/ui) · **Referenz:** `design_handoff_spektral_redesign/README.md` + Prototyp `jotti Spektral-Redesign.dc.html`

## 1. Hintergrund & Problem
Die jotti-Website wurde neu gestaltet: bunte Spektralpalette (8 `--sp-*`-Töne), Spektralverläufe, Space Grotesk für Headings (`website/src/styles/brand.css`). Das App-Frontend (Login, Service, Admin) wirkt dagegen markenfern: Wortmarke ist schwarzer Text, keine Verbindung zur Spektral-Sprache, Ladezustände sind grauer Standard-Pulse, Interaktionen geben kaum Feedback (Erfolg nur als Toast).

## 2. Ziele
1. **Markenkohärenz:** Website-Spektrum dezent in der App sichtbar machen — erkennbar dieselbe Marke, ohne die Arbeits-UI bunt zu machen.
2. **UX-Feedback:** Subtile Animationen, die Bediensicherheit erhöhen (Erfolg beim Kassieren, Statuswechsel, Ladezustände), nicht dekorieren.
3. **Typo-Vereinheitlichung:** Inter (UI/Body) + Space Grotesk (Wortmarke/Headings) überall — App, Website, Doku.

**Nicht-Ziele:** Kein Layout-Umbau, keine neuen Features, keine Änderung der Farbsemantik (Grün/Rot/Amber), keine Website-/Doku-Änderungen (dort bereits umgesetzt).

## 3. Leitplanken (verbindlich)
- Spektrum **nur** an vier Stellen: Wortmarke, Hintergrund-Glows, Ladezustände (Shimmer/Progress), Hairline-Akzente (Hero-Karte, aktiver Nav-Eintrag). Nirgendwo sonst.
- Emerald `--primary` bleibt Primärfarbe; Status-Semantik unverändert.
- Alle WCAG-Entscheidungen des Bestands bleiben (AA-Text, Non-Text-Kontrast, Disabled-Token, Warn-Amber nach ADR 04). Glows/Hairlines sind dekorativ (`aria-hidden`) und dürfen Textkontrast nicht unter AA drücken (Opacity ≤ .25 hinter Text).
- `prefers-reduced-motion: reduce` deaktiviert sämtliche Animationen.
- Keine neuen Dependencies außer `@fontsource-variable/space-grotesk`. Keine Animations-Library; CSS + rAF genügen.
- Performance: Glows sind statische, geblurte Elemente (kein Repaint-Loop außer optionalem Drift am Login); Zähl-Animationen nur bei Wertänderung.

## 4. Anforderungen

### F1 — Token-Fundament
`--sp-*`-Farben und `--spectral` (Light/Dark) aus `brand.css` nach `frontend/src/index.css` übernehmen; Keyframes (`shimmer`, `fadeUp`, `pop`, `pulsedot`, `drift`) + Reduced-Motion-Regel ergänzen. Single Source: Werte müssen mit der Website identisch bleiben (Kommentar mit Quellverweis).

### F2 — Wortmarke
Wiederverwendbare Komponente `<Wortmarke size>` (Spektral-Text-Fill, Space Grotesk 700). Einsatz: LoginForm (38 px), AdminSidebar (26 px), AdminMobileHeader (14 px → hier genügt Fill auf bestehender Größe).

### F3 — Login-Glows
Drei dekorative Glow-Kreise im AuthLayout (teal/violet/orange, Werte im README), optionaler Drift-Loop.

### F4 — Admin-Header-Glow
`<HeaderGlow farben>`-Komponente hinter `AdminPageHeader`; pro Seite definierte Farbpaare (Übersicht teal+violet, Kassentag orange+teal, Produkte green+blue, Benutzer blue+red, übrige Seiten teal+violet). `print:hidden`.

### F5 — Ladezustände
`ui/skeleton.tsx` auf Spektral-Shimmer umstellen (ersetzt `animate-pulse` global). Spektral-Progressbar-Komponente für Export/TSE-Assistent.

### F6 — Hairline-Akzente
2-px-Spektral-Hairline auf der Hero-Kennzahlkarte der Übersicht; 3-px-Marker am aktiven `SidebarMenuButton`. Beide Opacity .6.

### F7 — Motion: Micro-Feedback
Button-Press (100 ms translate+scale) global über `ui/button.tsx` (bestehendes `active:translate-y-px` um `scale` ergänzen); Stepper-Press .92; Fokus-/Hover-Transitions 150 ms.

### F8 — Motion: Übergänge & Status
TabsContent fadeUp 250 ms; Listen-Eintritt fadeUp 450 ms mit 60-ms-Stagger (nur Erst-Mount: Meine Tische, Positionslisten); Badge-/Häkchen-Statuswechsel pop 350 ms; „Kasse öffnet“-Kartenwechsel fadeUp.

### F9 — Motion: Erfolgs-Pop
Overlay-Bestätigung nach Kassieren/Bestellen im Service (Spec im README: Spring-Pop 450 ms, auto-dismiss 1,4 s, danach Statuswechsel). Komponente `service/components/ErfolgsPop.tsx`.

### F10 — Motion: Zahlen zählen
Hook `useCountUp` (700 ms, ease-out-cubic, rAF, tabular-nums) für Tisch-Saldo (`data-slot="tisch-saldo"`) und Übersicht-Kennzahlen bei Refetch.

### F11 — Live-Puls
`StatusDot` erhält optionale `puls`-Prop (pulsedot 2,4 s) für „Kasse offen“ (Sidebar-Karte) und „Live“-Anzeige der Übersicht.

### F12 — Typografie
`@fontsource-variable/space-grotesk` einführen; `--font-heading` in `index.css` auf Space Grotesk mappen. H1/AdminPageHeader/CardTitle nutzen `font-heading` (CardTitle tut es bereits). Fließtext bleibt Inter.

## 5. Akzeptanzkriterien
- [ ] Light & Dark: Spektrum erscheint ausschließlich an den vier definierten Stellen; Screens gleichen dem Prototyp (1a–1g).
- [ ] Wortmarke in Login, Sidebar, Mobile-Header spektral; Text bleibt selektierbar/lesbar (kein Bild).
- [ ] Skeletons shimmern spektral getönt; kein grauer `animate-pulse` mehr sichtbar.
- [ ] Kassieren im Service zeigt Erfolgs-Pop, Saldo zählt animiert herunter, Badge wechselt danach mit Pop.
- [ ] Alle Animationen entsprechen der Inventar-Tabelle (Dauer/Easing); `prefers-reduced-motion` schaltet sie ab (manuell prüfbar via DevTools-Emulation).
- [ ] Kontrast-Stichprobe: Text über Glows erfüllt AA (Glow-Opacity ≤ .25 hinter Text).
- [ ] Keine Layout-Shifts durch Zahlen-Animation (tabular-nums) oder Glows (absolute Positionierung).
- [ ] `npm run lint`, `npm run test`, bestehende Vitest-Suites grün; Snapshot-/DOM-Tests ggf. angepasst (neue dekorative Elemente sind `aria-hidden`).
- [ ] Druckansichten (Kassenberichte) frei von Glows/Animationen.

## 6. Offene Punkte
- Ersetzt der Erfolgs-Pop den sonner-Toast im Service vollständig oder ergänzt er ihn? (Empfehlung: ersetzt ihn dort; Admin behält Toasts.)
- Doku/Starlight-Theme: Headings dort bereits Space Grotesk — nur prüfen, nichts ändern.
