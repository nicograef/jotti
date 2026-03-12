---
description: "Implementiert den nächsten offenen Abschnitt aus einer plan.md. Beansprucht den Abschnitt per 🔒-Marker, sodass parallele Agents sich nicht in die Quere kommen."
agent: "agent"
argument-hint: "Pfad zur plan.md (z.B. docs/agents/theory-cleanup/plan.md)"
---

# Implementierung

Lies die referenzierte plan.md und arbeite **einen Abschnitt** ab.

## Abschnitt auswählen und beanspruchen

1. **Lies die gesamte plan.md** — insbesondere den Parallelisierungs-Abschnitt und alle Abschnitts-Überschriften.
2. **Finde den nächsten verfügbaren Abschnitt.** Ein Abschnitt ist verfügbar, wenn:
   - Er offene Tasks hat (`- [ ]`)
   - Er **nicht** mit 🔒 oder ✅ markiert ist
   - Seine Abhängigkeiten erfüllt sind (alle Vorgänger-Abschnitte sind ✅)
3. **Beanspruche den Abschnitt sofort**, indem du `🔒` an die Abschnitts-Überschrift anhängst:
   - Vorher: `## Abschnitt 4: Phase 1 — event-sourcing.md bereinigen`
   - Nachher: `## Abschnitt 4: Phase 1 — event-sourcing.md bereinigen 🔒`
   - **Erst danach** mit der Arbeit beginnen.
4. **Falls kein verfügbarer Abschnitt existiert: Stoppe sofort, ohne Änderungen vorzunehmen.** Erkläre dem User:
   - Welche Abschnitte noch offen sind (falls vorhanden)
   - Warum sie nicht bearbeitet werden können (🔒 = ein anderer Agent arbeitet daran, oder Abhängigkeiten noch nicht ✅)
   - Welche Vorgänger-Abschnitte zuerst abgeschlossen werden müssen
   - **Führe keine Änderungen an Dateien durch — weder an der plan.md noch an anderen Dateien.**

## Kontext laden

Nachdem du einen Abschnitt beansprucht hast, lade den Kontext:

1. **Lies den `Kontext:`-Block des Abschnitts** — er listet exakt die Dateien und Zeilenbereiche auf, die du für diesen Abschnitt brauchst. Lies genau diese Stellen, nicht mehr.
2. **Lies bereits erstellte/geänderte Dateien** aus vorherigen Abschnitten, um nahtlos anzuknüpfen.

## Abschnitt abarbeiten

**Befolge die Agent-Anweisungen in der plan.md exakt.** Die plan.md enthält die vollständigen Regeln für Task-Abarbeitung, Abhaken und Abschluss — halte dich an diese.

## Abschnitt abschließen

Wenn alle Tasks `[x]` sind und Build/Lint/Tests erfolgreich:

1. **Ersetze 🔒 durch ✅** in der Abschnitts-Überschrift:
   - Vorher: `## Abschnitt 4: Phase 1 — event-sourcing.md bereinigen 🔒`
   - Nachher: `## Abschnitt 4: Phase 1 — event-sourcing.md bereinigen ✅`
2. **Stoppen** — beginne nicht den nächsten Abschnitt
3. **Conventional Commit Message** vorschlagen
