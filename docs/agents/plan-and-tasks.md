# Agent Workflow: Analyze → Plan → Implement

## Überblick

Dreistufiger Workflow für Feature-Implementierung mit Copilot Agent Mode:

1. **`/analyze`** — Recherche + Analyse erstellen
2. **`/plan`** — Analyse in atomare Tasks + Parallelisierungs-Info aufteilen
3. **`/implement`** — Abschnitt für Abschnitt abarbeiten (auch parallel)

## Verzeichnisstruktur

Jedes Thema bekommt ein eigenes Verzeichnis unter `docs/agents/`:

```
docs/agents/
  theory-cleanup/
    analyze.md
    plan.md
  admin-dashboard/
    analyze.md
    plan.md
  bestellung-storno/
    analyze.md
    plan.md
```

## Workflow

### Session 1: Analyse erstellen

```
/analyze docs/theory von jotti entkoppeln und als eigenständige Wissensbasis aufbereiten
```

→ Agent erstellt `docs/agents/theory-cleanup/analyze.md`

### Session 2: Plan erstellen

```
/plan docs/agents/theory-cleanup/analyze.md
```

→ Agent erstellt `docs/agents/theory-cleanup/plan.md` mit Task-Liste und Parallelisierungs-Hinweisen

### Session 3+: Implementieren (sequentiell)

```
/implement docs/agents/theory-cleanup/plan.md
```

→ Agent arbeitet den nächsten offenen Abschnitt ab, hakt Tasks ab, stoppt nach dem Abschnitt.

### Session 3+: Implementieren (parallel)

Jeder Agent beansprucht automatisch den nächsten freien Abschnitt per 🔒-Marker. **Kein manueller Abschnitt-Parameter nötig** — einfach mehrere Chat-Sessions mit demselben Befehl starten:

| Chat-Session | Befehl                                          |
| ------------ | ----------------------------------------------- |
| Session A    | `/implement docs/agents/theory-cleanup/plan.md` |
| Session B    | `/implement docs/agents/theory-cleanup/plan.md` |
| Session C    | `/implement docs/agents/theory-cleanup/plan.md` |

Jeder Agent:

1. Liest die plan.md
2. **Wählt Abschnitt** — sieht, welche Abschnitte ✅ (fertig) oder 🔒 (in Arbeit) sind, prüft Abhängigkeiten
3. **Markiert** den nächsten freien Abschnitt mit 🔒
4. **Lädt Kontext** — liest genau die Dateien aus dem `Kontext:`-Block des Abschnitts
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

| Chat-Session | Befehl                                           |
| ------------ | ------------------------------------------------ |
| Session A    | `/implement docs/agents/theory-cleanup/plan.md`  |
| Session B    | `/implement docs/agents/admin-dashboard/plan.md` |

## Hinweise

- **Kontext pro Abschnitt:** Jeder Abschnitt in der plan.md hat einen `Kontext:`-Block, der exakt auflistet, welche Dateien und Zeilenbereiche gelesen werden müssen — der Agent liest nur das. Es gibt keine separaten „Kontext-Lade-Abschnitte“. Jeder Abschnitt muss echten Output produzieren.
- **Automatische Koordination:** Agents koordinieren sich über die 🔒/✅-Marker in der plan.md. Kein manuelles Zuweisen nötig.
- **Abhängigkeiten:** Agents prüfen selbst, ob Vorgänger-Abschnitte ✅ sind, bevor sie einen Abschnitt beanspruchen.
- **Race Condition:** Bei gleichzeitigem Start kann es vorkommen, dass zwei Agents denselben Abschnitt beanspruchen wollen. Die Datei-Edit-Operation schlägt dann bei einem fehl (Überschrift bereits geändert). Der Agent erkennt das und wählt den nächsten freien.
- **Checkpoint-Commits:** Nach jedem Abschnitt einen Commit machen — so sind Konflikte leichter zu lösen.
