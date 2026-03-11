# Agent Workflow: Plan → Progress → Implement

## Überblick

Dreistufiger Workflow für Feature-Implementierung mit Copilot Agent Mode:

1. **`/plan`** — Recherche + Implementierungsplan erstellen
2. **`/progress`** — Plan in atomare Tasks + Parallelisierungs-Info aufteilen
3. **`/implement`** — Abschnitt für Abschnitt abarbeiten (auch parallel)

## Verzeichnisstruktur

Jedes Thema bekommt ein eigenes Verzeichnis unter `docs/agents/`:

```
docs/agents/
  theory-cleanup/
    plan.md
    progress.md
  admin-dashboard/
    plan.md
    progress.md
  bestellung-storno/
    plan.md
    progress.md
```

## Workflow

### Session 1: Plan erstellen

```
/plan docs/theory von jotti entkoppeln und als eigenständige Wissensbasis aufbereiten
```

→ Agent erstellt `docs/agents/theory-cleanup/plan.md`

### Session 2: Progress-Datei erstellen

```
/progress docs/agents/theory-cleanup/plan.md
```

→ Agent erstellt `docs/agents/theory-cleanup/progress.md` mit Task-Liste und Parallelisierungs-Hinweisen

### Session 3+: Implementieren (sequentiell)

```
/implement docs/agents/theory-cleanup/progress.md
```

→ Agent arbeitet den nächsten offenen Abschnitt ab, hakt Tasks ab, stoppt nach dem Abschnitt.

### Session 3+: Implementieren (parallel)

Jeder Agent beansprucht automatisch den nächsten freien Abschnitt per 🔒-Marker. **Kein manueller Abschnitt-Parameter nötig** — einfach mehrere Chat-Sessions mit demselben Befehl starten:

| Chat-Session | Befehl                                              |
| ------------ | --------------------------------------------------- |
| Session A    | `/implement docs/agents/theory-cleanup/progress.md` |
| Session B    | `/implement docs/agents/theory-cleanup/progress.md` |
| Session C    | `/implement docs/agents/theory-cleanup/progress.md` |

Jeder Agent:

1. Liest die progress.md
2. Sieht, welche Abschnitte ✅ (fertig) oder 🔒 (in Arbeit) sind
3. Prüft Abhängigkeiten im Parallelisierungs-Abschnitt
4. Markiert den nächsten freien Abschnitt mit 🔒
5. Arbeitet die Tasks ab
6. Ersetzt 🔒 durch ✅ wenn fertig

### Abschnitt-Marker

| Marker        | Bedeutung                                    |
| ------------- | -------------------------------------------- |
| (kein Marker) | Verfügbar                                    |
| 🔒            | In Bearbeitung — von einem Agent beansprucht |
| ✅            | Abgeschlossen — alle Tasks erledigt          |

## Parallel: Verschiedene Features

Da jedes Feature sein eigenes Verzeichnis hat, können auch verschiedene Features parallel bearbeitet werden:

| Chat-Session | Befehl                                               |
| ------------ | ---------------------------------------------------- |
| Session A    | `/implement docs/agents/theory-cleanup/progress.md`  |
| Session B    | `/implement docs/agents/admin-dashboard/progress.md` |

## Hinweise

- **Automatische Koordination:** Agents koordinieren sich über die 🔒/✅-Marker in der progress.md. Kein manuelles Zuweisen nötig.
- **Abhängigkeiten:** Agents prüfen selbst, ob Vorgänger-Abschnitte ✅ sind, bevor sie einen Abschnitt beanspruchen.
- **Race Condition:** Bei gleichzeitigem Start kann es vorkommen, dass zwei Agents denselben Abschnitt beanspruchen wollen. Die Datei-Edit-Operation schlägt dann bei einem fehl (Überschrift bereits geändert). Der Agent erkennt das und wählt den nächsten freien.
- **Checkpoint-Commits:** Nach jedem Abschnitt einen Commit machen — so sind Konflikte leichter zu lösen.
