---
description: "Arbeitet den nächsten offenen Abschnitt aus einer plan.md ab."
agent: "agent"
argument-hint: "Pfad zur plan.md (z.B. docs/agents/bestellung-storno/plan.md)"
---

# Implementierung

Lies die referenzierte plan.md und arbeite **einen Abschnitt** ab.

## Vorgehen

1. **Lies die plan.md** und finde den nächsten Abschnitt mit offenen Tasks (`- [ ]`)
2. **Lies den `Kontext:`-Block** des Abschnitts — er listet die relevanten Dateien auf
3. **Arbeite die Tasks sequentiell ab** — von oben nach unten
4. **Hake jeden Task sofort ab** (`- [ ]` → `- [x]`) nachdem er erledigt ist
5. **Prüfe nach dem letzten Task**: Build, Lint, Tests (`make check`)
6. **Stopp** — beginne nicht den nächsten Abschnitt

## Leitfaden

- Einfache, klare, idiomatische Lösungen bevorzugen
- Keine Performance-Optimierung auf Kosten von Lesbarkeit
- Kleine lokale Duplikation ist erlaubt wenn sie den Code verständlicher macht
- Schlage eine Conventional Commit Message vor
