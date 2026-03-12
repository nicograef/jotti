---
description: "Erstellt eine ausführliche Analyse (analyze.md) für ein neues Feature oder eine Aufgabe. Recherchiert den Codebase-Kontext und dokumentiert alle relevanten Erkenntnisse."
agent: "agent"
argument-hint: "Beschreibe das Feature oder die Aufgabe..."
---

# Analysephase

Du bist in der **Analysephase**. Deine Aufgabe ist es, die nachfolgende Aufgabenbeschreibung zu analysieren, den Codebase-Kontext zu recherchieren und eine ausführliche Analyse zu erstellen.

## Vorgehen

1. **Analysiere** die Aufgabenbeschreibung gründlich
2. **Recherchiere** den relevanten Codebase-Kontext — lies betroffene Dateien, verstehe bestehende Patterns und Abhängigkeiten
3. **Verifiziere** dein Verständnis — hinterfrage Annahmen, prüfe ob deine Analyse mit den bestehenden Konventionen konsistent ist
4. **Leite einen kurzen, prägnanten Slug** aus der Aufgabenbeschreibung ab (z.B. `theory-cleanup`, `admin-dashboard`, `bestellung-storno`). Nur Kleinbuchstaben, Ziffern und Bindestriche.
5. **Erstelle** die Analyse als `docs/agents/<slug>/analyze.md`

## Regeln

- **Führe KEINE Code-Änderungen durch.** Nur analyze.md erstellen/schreiben.
- **Sei gründlich.** Nutze alle verfügbaren Tools um Kontext zu sammeln.
- **Präzise Referenzen sind Pflicht.** Jede Erkenntnis muss mit Dateipfad, Abschnitt und Zeilennummern belegt sein (z.B. `backend/api/product/http/handler.go:42-58`). Ein Agent in einer neuen Session muss ohne eigene Recherche direkt die richtige Stelle finden können.
- **Keine Implementierungsplanung.** Beschreibe den Ist-Zustand, Patterns und Abhängigkeiten — aber keine konkreten Implementierungsschritte. Das ist Aufgabe der Planungsphase (`/plan`).
- **Ein Thema = ein Verzeichnis.** Jedes Feature/Thema bekommt sein eigenes Unterverzeichnis unter `docs/agents/`. So können mehrere Analysen parallel existieren, ohne sich gegenseitig zu überschreiben.

## Struktur von analyze.md

```markdown
# Analyse: <Titel>

## Ziel

<Was soll erreicht werden?>

## Bestandsaufnahme

<Relevante bestehende Dateien, Patterns, Abhängigkeiten — jeweils mit Dateipfad:Zeilen>

### <Themenbereich 1> (z.B. Bestehende Handler, Domain-Modell, Frontend-Komponenten)

<Erkenntnisse mit präzisen Referenzen>

### <Themenbereich N>

<...>

## Offene Fragen / Risiken

<Falls vorhanden>

## Referenzen

<Alle referenzierten Dateien mit Zeilennummern als kompakte Liste>
```
