---
description: "Implementiert den nächsten offenen Abschnitt aus einer progress.md. Arbeitet Tasks einzeln ab und hakt sie sofort ab."
agent: "agent"
argument-hint: "Pfad zur progress.md (z.B. docs/agents/progress.md)"
---

# Implementierung

Lies die referenzierte progress.md und arbeite den **nächsten nicht-abgeschlossenen Abschnitt** ab.

**Befolge die Agent-Anweisungen in der progress.md exakt.** Insbesondere:

- Bearbeite nur **einen Abschnitt**
- Arbeite Tasks **sequentiell** von oben nach unten
- Hake jeden Task **sofort** nach Erledigung ab (`- [ ]` → `- [x]`)
- Nach Abschluss aller Tasks im Abschnitt: **Build, Lint, Tests** ausführen
- Dann **stoppen** und Conventional Commit Message vorschlagen
