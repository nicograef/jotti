---
description: "Implementiert den nächsten offenen Abschnitt aus einer progress.md. Beansprucht den Abschnitt per 🔒-Marker, sodass parallele Agents sich nicht in die Quere kommen."
agent: "agent"
argument-hint: "Pfad zur progress.md (z.B. docs/agents/theory-cleanup/progress.md)"
---

# Implementierung

Lies die referenzierte progress.md und arbeite **einen Abschnitt** ab.

## Kontext laden

Bevor du einen Abschnitt beanspruchst, verschaffe dir den nötigen Kontext:

1. **Lies `plan.md`** im selben Verzeichnis wie die progress.md — sie enthält den Gesamtplan, relevante Datei-Referenzen und Kontext-Informationen.
2. **Lies Referenzdateien**, die in der plan.md genannt werden und für die anstehende Arbeit relevant sind.
3. **Lies bereits erstellte/geänderte Dateien** aus vorherigen Abschnitten, um nahtlos anzuknüpfen.

Dieser Schritt ist in jeder neuen Session nötig, da Kontext zwischen Sessions nicht erhalten bleibt.

## Abschnitt auswählen und beanspruchen

1. **Lies die gesamte progress.md** — insbesondere den Parallelisierungs-Abschnitt und alle Abschnitts-Überschriften.
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
   - **Führe keine Änderungen an Dateien durch — weder an der progress.md noch an anderen Dateien.**

## Abschnitt abarbeiten

**Befolge die Agent-Anweisungen in der progress.md exakt.** Insbesondere:

- Arbeite Tasks **sequentiell** von oben nach unten
- Hake jeden Task **sofort** nach Erledigung ab (`- [ ]` → `- [x]`)
- Verwende beim Abhaken immer die **Abschnitts-Überschrift + den vollständigen Task-Text** als Kontext, damit die Ersetzung in der Datei eindeutig ist
- Nach Abschluss aller Tasks im Abschnitt: **Build, Lint, Tests** ausführen

## Abschnitt abschließen

Wenn alle Tasks `[x]` sind und Build/Lint/Tests erfolgreich:

1. **Ersetze 🔒 durch ✅** in der Abschnitts-Überschrift:
   - Vorher: `## Abschnitt 4: Phase 1 — event-sourcing.md bereinigen 🔒`
   - Nachher: `## Abschnitt 4: Phase 1 — event-sourcing.md bereinigen ✅`
2. **Stoppen** — beginne nicht den nächsten Abschnitt
3. **Conventional Commit Message** vorschlagen
