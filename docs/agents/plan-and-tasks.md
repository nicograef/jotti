# Session 1

Deine Aufgabe ist, die nachfolgenden Anweisungen strukturiert und mit fundiertem Wissen auszuführen. Nutze alle deine Tools, um dein Verständnis abzusichern, deinen Kontext anzureichern und notwendige Informationen einzuholen (auch von externen Quellen.) Gehe möglichst strukturiert vor. Erstelle dir zunächst einen Plan, verifiziere danach noch einmal, ob dieser zum richtigen Ziel bzw. zum gewünschten Ergebnis führt. Erst dann führe Aktionen aus.

**Wichtig:** Du sollst in dieser Session keine Änderungen ausführen. Erstelle einen ausführlichen Plan, wie man vorgehen muss. Schreibe diesen Plan inklusive Kontext und notwendigen Informationen und Referenzen in eine Datei plan.md. Diese Datei soll so gestaltet werden, dass ein Coding Agent in einer neuen Session direkt damit loslegen kann.

---

# Session 2

Erzeuge eine Todo-/Task-Liste, die den Plan aus #file:plan.md in möglichst kleine iterative Schritte unterteilt. Schreibe alles in eine neue Datei progress.md. Die Datei sollte einen minimalen Kontext enthalten, entsprechende Referenzen auf den ausführlichen Plan und die Task-Liste im Checkbox-Style (- [ ] Task).

Füge dann ganz oben in der progress.md folgendes ein:

```md
Details siehe plan.md

## Agent Anweisungen

> **Lies diese Anweisungen vollständig, bevor du mit der Arbeit beginnst.**

1. **Ein Abschnitt pro Auftrag.** Bearbeite nur den nächsten nicht-abgeschlossenen Abschnitt (= erster Abschnitt mit offenen `- [ ] Tasks`)
2. **Ein Task nach dem anderen.** Arbeite Tasks innerhalb des Abschnitts sequentiell ab - von oben nach unten.
3. **Sofort abhaken.** Ändere `- [ ]` zu `- [x]` in dieser Datei **unmittelbar** nachdem ein task erfolgreich erledigt ist. Nicht erst am Ende des Abschnitts, nicht gebündelt - **nach jedem einzelnen Task**.
4. **Abschnitt abschließen.** Wenn alle Tasks eines Abschnitts `[x]` sind, führe die wichtigsten Dev-Scripte und CI steps lokal aus: compilation, build, linting, formatting, testing. Stelle sicher, dass es keine Fehler oder Warnings gibt. Erst dann ist der Abschnitt fertig.
5. **Stoppen.** Nach Abschluss eines Abschnitts: **stopp**. Beginne nicht den nächsten Abschnitt, sondern melde, dass der Abschnitt abgeschlossen ist.
6. **Conventional Commit Message schreiben.** Schreibe zu deinen Änderungen bzw. dem Abschnitt eine conventional commit message. Führe kein commit selbst durch, schreibe nur die Message in den Chat, sodass diese kopiert werden kann.
```

# Session 3

Implementiere den nächsten offenen Task aus #file:progress.md
