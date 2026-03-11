---
description: "Erstellt einen ausführlichen Implementierungsplan (plan.md) für ein neues Feature oder eine Aufgabe. Recherchiert den Codebase-Kontext, bevor geplant wird."
agent: "agent"
argument-hint: "Beschreibe das Feature oder die Aufgabe..."
---

# Planungsphase

Du bist in der **Planungsphase**. Deine Aufgabe ist es, die nachfolgende Aufgabenbeschreibung zu analysieren, den Codebase-Kontext zu recherchieren und einen ausführlichen Implementierungsplan zu erstellen.

## Vorgehen

1. **Analysiere** die Aufgabenbeschreibung gründlich
2. **Recherchiere** den relevanten Codebase-Kontext — lies betroffene Dateien, verstehe bestehende Patterns und Abhängigkeiten
3. **Verifiziere** dein Verständnis — hinterfrage Annahmen, prüfe ob dein Plan mit den bestehenden Konventionen konsistent ist
4. **Erstelle** den Plan als `plan.md` im Arbeitsverzeichnis (z.B. `docs/agents/plan.md`)

## Regeln

- **Führe KEINE Code-Änderungen durch.** Nur plan.md erstellen/schreiben.
- **Sei gründlich.** Nutze alle verfügbaren Tools um Kontext zu sammeln.
- **Referenziere konkrete Dateien und Code-Stellen**, damit ein Agent in einer neuen Session direkt loslegen kann.
- **Beschreibe das Was UND das Wie** — nicht nur "erstelle Handler", sondern welche Funktion, welche Parameter, welches Pattern (mit Verweis auf existierenden Code als Vorlage).

## Struktur von plan.md

```markdown
# Plan: <Titel>

## Ziel

<Was soll erreicht werden?>

## Kontext

<Relevante bestehende Dateien, Patterns, Abhängigkeiten>

## Implementierungsschritte

<Nummerierte Schritte mit konkreten Datei-Referenzen und Code-Beschreibungen>

## Offene Fragen / Risiken

<Falls vorhanden>

## Referenzen

<Links zu relevanten Dateien im Projekt>
```
