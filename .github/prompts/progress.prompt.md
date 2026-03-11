---
description: "Erstellt eine progress.md mit kleinteiliger Task-Liste aus einem bestehenden plan.md. Für iteratives Abarbeiten durch einen Coding Agent."
agent: "agent"
argument-hint: "Pfad zur plan.md (z.B. docs/agents/theory-cleanup/plan.md)"
---

# Progress-Datei erstellen

Erzeuge aus dem referenzierten Plan eine **progress.md** im selben Verzeichnis. Die Datei enthält eine kleinteilige Task-Liste im Checkbox-Style, die ein Coding Agent Abschnitt für Abschnitt abarbeiten kann.

## Vorgehen

1. **Lies** die referenzierte plan.md vollständig
2. **Unterteile** jeden Implementierungsschritt in möglichst kleine, atomare Tasks
3. **Gruppiere** die Tasks in logische Abschnitte (z.B. nach Schicht: Domain, Repository, Handler, Frontend)
4. **Analysiere Abhängigkeiten** zwischen Abschnitten: Welche Abschnitte arbeiten an denselben Dateien? Welche sind voneinander unabhängig?
5. **Schreibe** die progress.md mit Parallelisierungs-Hinweisen

## Regeln

- **Führe KEINE Code-Änderungen durch.** Nur progress.md erstellen/schreiben.
- **Tasks müssen atomar sein** — ein Task = eine klar abgrenzbare Aktion (eine Datei erstellen, eine Funktion schreiben, einen Test hinzufügen)
- **Jeder Task muss unabhängig prüfbar sein** — nach Abschluss muss klar sein, ob er erledigt ist oder nicht
- **Keine reinen Kontext-Lade-Abschnitte.** Abschnitte, deren Tasks nur aus "Datei X lesen" oder "Kontext erfassen" bestehen, sind verboten. Kontext-Laden gehört in den Block `### Kontext laden` der Agent-Anweisungen (siehe Template unten) — nicht in einen eigenen Abschnitt. Jeder Abschnitt muss Output produzieren (Dateien erstellen/ändern, Code schreiben, Dokumentation schreiben).

## Struktur von progress.md

Die Datei MUSS mit folgendem Block beginnen:

```markdown
Details siehe plan.md

## Agent Anweisungen

> **Lies diese Anweisungen vollständig, bevor du mit der Arbeit beginnst.**

### Kontext laden (vor jedem Abschnitt)

Bevor du einen Abschnitt beanspruchst, lies **immer** diese Dateien:

1. `plan.md` (im selben Verzeichnis) — Gesamtplan, Kontext und Referenzen
2. Alle in plan.md genannten Referenzdateien, die für den Abschnitt relevant sind
3. Bereits erstellte/geänderte Dateien aus vorherigen Abschnitten (um nahtlos anzuknüpfen)

Diese Dateien werden in jeder neuen Session erneut gelesen — die Kontext-Beschaffung ist kein eigener Abschnitt, sondern Pflicht vor jeder Arbeit.

### Abschnitt beanspruchen

1. **Lies die gesamte progress.md** — insbesondere den Parallelisierungs-Abschnitt und alle Abschnitts-Überschriften.
2. **Finde den nächsten verfügbaren Abschnitt.** Ein Abschnitt ist verfügbar, wenn:
   - Er offene Tasks hat (`- [ ]`)
   - Er **nicht** mit 🔒 oder ✅ markiert ist
   - Seine Abhängigkeiten erfüllt sind (alle Vorgänger-Abschnitte sind ✅)
3. **Beanspruche den Abschnitt sofort**, indem du `🔒` an die Überschrift anhängst (`## Abschnitt N: Titel` → `## Abschnitt N: Titel 🔒`). Erst danach mit der Arbeit beginnen.
4. **Falls kein verfügbarer Abschnitt existiert: Stoppe sofort, ohne Änderungen vorzunehmen.** Erkläre dem User: welche Abschnitte noch offen sind, warum sie nicht bearbeitet werden können (🔒 = anderer Agent arbeitet daran, oder Abhängigkeiten noch nicht ✅), und welche Vorgänger-Abschnitte zuerst abgeschlossen werden müssen. **Führe keine Änderungen an Dateien durch.**

### Abschnitt abarbeiten

1. **Ein Task nach dem anderen.** Arbeite Tasks innerhalb des Abschnitts sequentiell ab — von oben nach unten.
2. **Sofort abhaken.** Ändere `- [ ]` zu `- [x]` in dieser Datei **unmittelbar** nachdem ein Task erfolgreich erledigt ist. Nicht erst am Ende des Abschnitts, nicht gebündelt — **nach jedem einzelnen Task**. Verwende beim Abhaken immer die **Abschnitts-Überschrift + den vollständigen Task-Text** als Kontext, damit die Ersetzung eindeutig ist.
3. **Abschnitt abschließen.** Wenn du an Code gearbeitet hast: Wenn alle Tasks eines Abschnitts `[x]` sind, führe die wichtigsten Dev-Scripte und CI-Steps lokal aus: compilation, build, linting, formatting, testing. Stelle sicher, dass es keine Fehler oder Warnings gibt. Erst dann ist der Abschnitt fertig. Wenn du an Dokumentation gearbeitet hast: Lese Korrektur, stelle sicher, dass alle Links funktionieren, und dass die Formatierung korrekt ist.
4. **✅ setzen.** Ersetze `🔒` durch `✅` in der Abschnitts-Überschrift (`## Abschnitt N: Titel 🔒` → `## Abschnitt N: Titel ✅`).
5. **Stoppen.** Nach Abschluss eines Abschnitts: **stopp**. Beginne nicht den nächsten Abschnitt, sondern melde, dass der Abschnitt abgeschlossen ist.
6. **Conventional Commit Message schreiben.** Wenn du an Code gearbeitet hast: Schreibe zu deinen Änderungen bzw. dem Abschnitt eine Conventional Commit Message. Führe kein Commit selbst durch, schreibe nur die Message in den Chat, sodass diese kopiert werden kann. Wenn du an Dokumentation gearbeitet hast: Schreibe eine passende Commit Message für die Dokumentationsänderungen.
```

Danach folgt der **Parallelisierungs-Abschnitt**, dann die Abschnitte mit Tasks:

```markdown
## Parallelisierung

Die folgenden Abschnitte können **parallel** in separaten Chat-Sessions bearbeitet werden:

- Abschnitte X, Y, Z (keine gemeinsamen Dateien)

Die folgenden Abschnitte haben **Abhängigkeiten**:

- Abschnitt A → muss nach Abschnitt B abgeschlossen sein
- Abschnitt C → muss nach allen vorherigen Abschnitten abgeschlossen sein

---

## Abschnitt 1: <Titel>

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Abschnitt 2: <Titel>

- [ ] Task 4
- [ ] Task 5
```

### Parallelisierungs-Analyse

Für den Parallelisierungs-Abschnitt analysierst du:

1. **Datei-Überschneidungen:** Welche Abschnitte bearbeiten welche Dateien? Abschnitte, die **keine gemeinsamen Dateien** bearbeiten, können parallel laufen.
2. **Logische Abhängigkeiten:** Baut ein Abschnitt auf dem Ergebnis eines anderen auf? (z.B. "Repository muss existieren, bevor der Handler es aufruft")
3. **Fasse zusammen**, welche Abschnitte parallel laufen dürfen und welche sequentiell sein müssen.
