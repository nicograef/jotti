---
description: "Erstellt eine progress.md mit kleinteiliger Task-Liste aus einem bestehenden plan.md. Für iteratives Abarbeiten durch einen Coding Agent."
agent: "agent"
argument-hint: "Pfad zur plan.md (z.B. docs/agents/plan.md)"
---

# Progress-Datei erstellen

Erzeuge aus dem referenzierten Plan eine **progress.md** im selben Verzeichnis. Die Datei enthält eine kleinteilige Task-Liste im Checkbox-Style, die ein Coding Agent Abschnitt für Abschnitt abarbeiten kann.

## Vorgehen

1. **Lies** die referenzierte plan.md vollständig
2. **Unterteile** jeden Implementierungsschritt in möglichst kleine, atomare Tasks
3. **Gruppiere** die Tasks in logische Abschnitte (z.B. nach Schicht: Domain, Repository, Handler, Frontend)
4. **Schreibe** die progress.md

## Regeln

- **Führe KEINE Code-Änderungen durch.** Nur progress.md erstellen/schreiben.
- **Tasks müssen atomar sein** — ein Task = eine klar abgrenzbare Aktion (eine Datei erstellen, eine Funktion schreiben, einen Test hinzufügen)
- **Jeder Task muss unabhängig prüfbar sein** — nach Abschluss muss klar sein, ob er erledigt ist oder nicht

## Struktur von progress.md

Die Datei MUSS mit folgendem Block beginnen:

```markdown
Details siehe plan.md

## Agent Anweisungen

> **Lies diese Anweisungen vollständig, bevor du mit der Arbeit beginnst.**

1. **Ein Abschnitt pro Auftrag.** Bearbeite nur den nächsten nicht-abgeschlossenen Abschnitt (= erster Abschnitt mit offenen `- [ ] Tasks`)
2. **Ein Task nach dem anderen.** Arbeite Tasks innerhalb des Abschnitts sequentiell ab - von oben nach unten.
3. **Sofort abhaken.** Ändere `- [ ]` zu `- [x]` in dieser Datei **unmittelbar** nachdem ein Task erfolgreich erledigt ist. Nicht erst am Ende des Abschnitts, nicht gebündelt - **nach jedem einzelnen Task**.
4. **Abschnitt abschließen.** Wenn alle Tasks eines Abschnitts `[x]` sind, führe die wichtigsten Dev-Scripte und CI-Steps lokal aus: compilation, build, linting, formatting, testing. Stelle sicher, dass es keine Fehler oder Warnings gibt. Erst dann ist der Abschnitt fertig.
5. **Stoppen.** Nach Abschluss eines Abschnitts: **stopp**. Beginne nicht den nächsten Abschnitt, sondern melde, dass der Abschnitt abgeschlossen ist.
6. **Conventional Commit Message schreiben.** Schreibe zu deinen Änderungen bzw. dem Abschnitt eine Conventional Commit Message. Führe kein Commit selbst durch, schreibe nur die Message in den Chat, sodass diese kopiert werden kann.
```

Danach folgen die Abschnitte mit Tasks:

```markdown
## Abschnitt 1: <Titel>

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Abschnitt 2: <Titel>

- [ ] Task 4
- [ ] Task 5
```
