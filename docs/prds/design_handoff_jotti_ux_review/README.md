# Handoff: jotti UX-Review — Quick Wins umsetzen

## Überblick
Dieses Paket enthält die Befunde des UX-Reviews (Juli 2026, Basis `nicograef/jotti@main`) als umsetzbare Tickets für Claude Code. Es ist **kein neues Design**, sondern eine Liste gezielter Fixes am bestehenden Frontend — alle Frontend-only, alle im bestehenden Token-/Komponentensystem (shadcn/ui, Tailwind 4, ADR 04/07/08 bleiben gültig).

## Über die beigelegten HTML-Dateien
`UX-Review.dc.html` und `Screens – Ist-Zustand.dc.html` sind **Design-Referenzen in HTML** — Mockups, kein Produktionscode. Nichts daraus kopieren; die Fixes werden direkt im jotti-Frontend (`frontend/src/…`) mit den vorhandenen Komponenten umgesetzt. Die Mockups zeigen nur das Zielbild (Vorher/Nachher pro Fix).

## Fidelity
High-fidelity für die *Vorher*-Zustände (aus dem echten Code rekonstruiert). Die *Nachher*-Mockups sind richtungsweisend: exakte Abstände/Varianten aus dem bestehenden Designsystem ableiten (Button-Variants, `--warn`-Muster, Item/Drawer-Primitives), nicht aus den Mockup-Pixeln.

## Empfohlener Workflow mit Claude Code
1. Im jotti-Repo arbeiten: `git checkout -b ux/quick-wins-1` (ein Branch pro Block aus TICKETS.md).
2. Claude Code im Repo-Root starten; jotti hat eigene Agent-Regeln (`CLAUDE.md`, `AGENTS.md`, `.github/instructions/frontend.instructions.md`) — die gelten weiter und Claude Code liest sie automatisch.
3. Pro Ticket den Prompt-Block aus `TICKETS.md` einfügen. Die Tickets nennen exakte Dateien und Akzeptanzkriterien.
4. Verifizieren: `make dev`, betroffene Flows manuell auf Handy-Breite (390 px) UND ≥1024 px prüfen (ADR 08: beide Container teilen die Abschluss-Komponenten). Bestehende Tests laufen lassen (`frontend`: vitest; E2E-Specs matchen auf Texte — Copy-Änderungen brauchen Test-Anpassung, siehe Ticket-Hinweise).
5. Reihenfolge: Block A (vor 1.0) → Block B (Copy-&-Token-Sweep) → Block C (nach 1.0).

## Dateien in diesem Paket
- `TICKETS.md` — alle Befunde als Claude-Code-Tickets mit Dateipfaden und Akzeptanzkriterien
- `UX-Review.dc.html` — das Review mit Vorher/Nachher-Mockups (im Browser öffnen)
- `Screens – Ist-Zustand.dc.html` — rekonstruierte Referenz-Screens

## Design-Tokens
Keine neuen Tokens nötig, mit einer Ausnahme (F5): ggf. ein `--destructive-solid-foreground` (Dark: dunkles Rot ≈ red-950 auf der red-400-Fläche), nach dem Muster von `--warn`/`--warn-foreground` in `frontend/src/index.css`.
