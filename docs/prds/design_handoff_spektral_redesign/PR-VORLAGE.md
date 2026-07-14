# PR-Vorlage (für Claude Code)

Branch: `design/spektral-redesign-frontend`

## Titel
feat(frontend): Spektral-Redesign — Marken-Akzente & Motion-Inventar

## Beschreibung
Überträgt die Spektral-Markensprache der Website dezent auf das App-Frontend (Login, Service, Admin, common) und führt ein einheitliches Motion-Inventar ein. Spezifikation: `design_handoff_spektral_redesign/` (README, PRD, UMBAU-PLAN, Screenshots Light/Dark, interaktiver Prototyp).

**Spektrum nur an vier Stellen:** Wortmarke, Hintergrund-Glows (Login, Admin-Seitenköpfe), Ladezustände (Shimmer/Progress), Hairline-Akzente (Hero-Karte, aktiver Nav-Eintrag). Grün bleibt Primärfarbe, Status-Semantik unverändert.

**Motion:** Button-Press 100 ms · Tab-/Listen-fadeUp 250/450 ms · Erfolgs-Pop beim Kassieren 450 ms Spring · Saldo/Kennzahlen zählen 700 ms · Skeleton-Shimmer 1,6 s · Live-Puls 2,4 s · `prefers-reduced-motion` schaltet alles ab.

**Typografie:** Inter (UI/Body) + Space Grotesk (Wortmarke/Headings) via `@fontsource-variable/space-grotesk`.

## Checkliste (aus PRD §5)
- [ ] Spektrum nur an den vier definierten Stellen (Light + Dark)
- [ ] Erfolgs-Pop + animierter Saldo beim Kassieren
- [ ] Skeletons shimmern spektral, kein grauer Pulse mehr
- [ ] Reduced-Motion-Check, AA-Kontrast über Glows
- [ ] Lint/Tests/Build grün, Druckansichten ohne Glows

Umsetzung phasenweise nach `UMBAU-PLAN.md` (Phase 0–8), gern als Commit je Phase.
