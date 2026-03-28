---
description: "Recherchiert den Codebase-Kontext und erstellt eine plan.md mit Bestandsaufnahme und atomarer Task-Liste für ein Feature oder eine Aufgabe."
agent: "agent"
argument-hint: "Beschreibe das Feature oder die Aufgabe..."
---

# Plan erstellen

Recherchiere den Codebase-Kontext und erstelle eine **plan.md** mit Bestandsaufnahme und Task-Liste.

## Vorgehen

1. **Analysiere** die Aufgabenbeschreibung
2. **Recherchiere** den relevanten Codebase-Kontext — lies betroffene Dateien, verstehe bestehende Patterns
3. **Leite einen Slug** aus der Aufgabe ab (z.B. `admin-dashboard`, `bestellung-storno`)
4. **Erstelle** die Datei `docs/agents/<slug>/plan.md`

## Regeln

- **Keine Code-Änderungen.** Nur plan.md erstellen.
- **Präzise Referenzen.** Jede Erkenntnis mit Dateipfad und Zeilennummern belegen (z.B. `backend/api/product/http/handler.go:42-58`).
- **Tasks müssen atomar sein** — ein Task = eine klar abgrenzbare Aktion.
- **Keine reinen Kontext-Lade-Abschnitte.** Jeder Abschnitt muss Output produzieren (Dateien erstellen/ändern).
- **Breaking Changes erlaubt.** Schema direkt in `01_initial.up.sql` ändern, keine Migrations-Strategien.
- **Readability-first.** Einfache, klare, idiomatische Lösungen bevorzugen.

## Struktur

```markdown
# Plan: <Titel>

## Ziel

<Was soll erreicht werden?>

## Bestandsaufnahme

<Relevante bestehende Dateien, Patterns, Abhängigkeiten — jeweils mit Dateipfad:Zeilen>

## Offene Fragen / Risiken

<Falls vorhanden>

---

## Abschnitt 1: <Titel>

Kontext:

- `pfad/datei.go:10-45` — <warum relevant>

- [ ] Task 1
- [ ] Task 2

## Abschnitt 2: <Titel>

Kontext:

- `pfad/datei.go:50-80` — <warum relevant>

- [ ] Task 3
- [ ] Task 4
```
