# ADR: CQRS-Projektionen für Tisch-Zustand

## Status

**Accepted** — Synchrone Projektion für operative Queries, On-Demand SQL für Reporting. Ersetzt vorherige ADR (Lazy Projection + Background Worker).

## Kontext

### Ausgangslage (vor dieser ADR)

jotti implementierte CQRS Stufe 1 (logische Trennung): Getrennte Command-/Query-Services, separate HTTP-Handler und segregierte Repository-Interfaces. Der Tisch-Zustand wurde bei jedem Lesezugriff durch Event-Replay berechnet, optimiert durch einen Snapshot-as-Event-Mechanismus (`tisch.snapshot:v1`).

### Problemstellung

1. **Snapshot-as-Event ist ein Anti-Pattern.** Das Snapshot-Event (`tisch.snapshot:v1`) war ein Infrastruktur-Artefakt im fachlichen Event Stream. Es vermischte Domänen-Events mit Optimierungs-Artefakten und untergrub die Klarheit des Kassenjournals.
2. **N Replays für Tischübersicht.** Die Tischübersicht (K-05) erforderte bei 50 Tischen 50 separate Event-Replays für die Saldo-Berechnung — konzeptionell unbefriedigend, auch wenn bei jottis Lastprofil (< 200 Events/Tisch) die Performance ausreichte.
3. **Kein Reporting-Pfad.** Tischübergreifende Auswertungen (R-01 bis R-05) erforderten JSONB-Parsing über alle Events — keine vorberechneten Aggregate vorhanden.

## Bewertete Alternativen

### A: Status quo (Snapshot + Event-Replay)

| +                               | -                                          |
| ------------------------------- | ------------------------------------------ |
| Kein Aufwand, existiert bereits | Snapshot-as-Event bleibt unrein            |
| Minimal komplex                 | Tischübersicht = N Replays                 |
| Strong Consistency              | Reporting-SQL wird komplex (JSONB-Parsing) |

### B: Synchrone Projektion (gewählt)

`table_state`-Tabelle, UPSERT in derselben Transaktion wie Event-INSERT.

| +                                                   | -                                              |
| --------------------------------------------------- | ---------------------------------------------- |
| Löst Snapshot-as-Event-Problem                      | Write wird minimal langsamer (ein UPSERT mehr) |
| Vorhersagbare Latenz (Read = 1 SELECT)              | Apply-Funktion muss konsistent mit Events sein |
| Tischübersicht trivial: `SELECT * FROM table_state` | Neuer Code (~150 Zeilen Apply-Funktion)        |
| Strong Consistency ohne Tricks                      |                                                |
| Reporting-Enabler (Saldo vorberechnet)              |                                                |

### C: Lazy Projection

Read-Through-Cache: Beim Lesezugriff prüfen, ob Projektion aktuell ist, fehlende Events bei Bedarf replayed.

| +                                     | -                                       |
| ------------------------------------- | --------------------------------------- |
| Command bleibt „rein" (single INSERT) | Unvorhersagbare Read-Latenz             |
| Self-healing                          | Fast identisch zum Snapshot-Mechanismus |
|                                       | Staleness-Logik, Cache-Invalidierung    |

**Verworfen:** Löst konzeptionell dasselbe Problem wie der bestehende Snapshot-Mechanismus — statt „lade Snapshot + replay Delta" wird es „lade Projektion + prüfe Aktualität + replay Delta". Der Architekturgewinn gegenüber dem Status quo ist marginal.

### D: Background Worker (für Analytik)

Worker pollt `events`-Tabelle und projiziert asynchron auf analytische Tabellen.

| +                                 | -                                     |
| --------------------------------- | ------------------------------------- |
| Skaliert für große Event-Volumina | Over-Engineering für < 10k Events     |
|                                   | Worker-Lifecycle, Checkpoint-Tracking |
|                                   | Eventual Consistency unnötig          |
|                                   | Widerspricht „Radikale Einfachheit"   |

**Verworfen:** Bei max. ~10.000 Events (50 Tische × 200 Events) kann jede Reporting-Query on-demand in Millisekunden berechnet werden. Ein Background Worker mit Checkpoint-Tabelle und Polling-Intervall ist unverhältnismäßig für ein System, das 2–3 Mal pro Jahr für wenige Stunden läuft.

### E: DB-Trigger

Business-Logik (Event-Auswertung, JSONB-Varianten-Mapping) in PL/pgSQL.

**Verworfen:** Schwer wartbar und testbar. Domänenlogik gehört in Go-Code, nicht in SQL-Prozeduren.

## Entscheidung

**Synchrone Projektion (`table_state`) für operative Queries. On-Demand SQL-Aggregation für Reporting. Kein Background Worker.**

### Operative Projektionen — Synchrone Projektion (Write-Through)

Die `table_state`-Tabelle wird in derselben Transaktion wie das Event-INSERT aktualisiert:

```
BEGIN TX
  INSERT INTO events (...)           → event_id
  SELECT * FROM table_state          → aktueller Zustand (oder Zero-Value)
  ApplyEvent(zustand, event)         → neuer Zustand (reine Go-Funktion)
  UPSERT INTO table_state (...)      → neuer Zustand persistiert
COMMIT TX
```

Die `ApplyEvent()`-Funktion (`backend/domain/table/projection.go`) ist eine reine Funktion in der Domain-Schicht ohne DB-Zugriff. Sie verarbeitet die vier Domänen-Event-Typen und berechnet den neuen `TischState` (Saldo, unbezahlte/ungelieferte Positionen, Gesamtzahlungen).

**Warum synchron statt lazy?**

- **Vorhersagbare Latenz** für alle Reads — kein implizites Write-on-Read
- **Gleiche Komplexität** wie Lazy (beide brauchen `ApplyEvent()`), aber einfacheres Konsistenzmodell
- **Strong Consistency** ohne Staleness-Detection oder Rebuild-Pfad

**CQRS-Trennung bleibt intakt:** Der Projektor ist ein internes Detail von `EventRepo.WriteEvent()`. Der Command-Service ruft weiterhin nur `WriteEvent()` auf — das UPSERT auf `table_state` ist transparent. Der Query-Service liest direkt aus `table_state` über `ReadTableState()`.

### Kassenjournal (Historie) — Event-Replay (Stufe 1)

Die Historie _ist_ der Event Stream. `GetTischHistorie()` liest weiterhin alle Events via `ReadEventsBySubject()` und formatiert sie über `GetHistoryFromEvents()`. Kein Read Model nötig.

### Reporting (R-01–R-05) — On-Demand SQL-Aggregation

| Aspekt     | Entscheidung                                                      |
| ---------- | ----------------------------------------------------------------- |
| Strategie  | SQL-Queries über `events`-Tabelle + `table_state`                 |
| Konsistenz | Strong (immer aktuell, on-demand berechnet)                       |
| Fallback   | Materialized Views bei Bedarf (unwahrscheinlich bei < 10k Events) |

Bei ~10.000 Events ist jede Aggregation in < 100ms erledigt. Kein Background Worker, keine Eventual Consistency, keine Checkpoint-Tabelle.

### CQRS-Stufen im Überblick

| Bereich                  | CQRS-Stufe                         | Strategie                         | Konsistenz |
| ------------------------ | ---------------------------------- | --------------------------------- | ---------- |
| Kassenbetrieb (operativ) | **Stufe 2** — Synchrone Projektion | `table_state`, UPSERT in Event-TX | Strong     |
| Kassenjournal (Historie) | **Stufe 1** — Event-Replay         | Event Stream = Read Model         | Strong     |
| Reporting (R-01–R-05)    | **Stufe 1** — On-Demand SQL        | `events` + `table_state`          | Strong     |
| Stammdaten (CRUD)        | **Stufe 0** — Kein CQRS            | Kein Event-Sourcing               | Strong     |

### Architektur

```
┌──────────────────────────────────────────────────────────────┐
│                          Frontend                             │
└──────┬──────────────────────────────────┬────────────────────┘
       │ Commands (Schreiben)             │ Queries (Lesen)
       ▼                                  ▼
┌──────────────────┐         ┌───────────────────────────────┐
│ Command Handler  │         │        Query Handler          │
│ (schreibt Events)│         │  Operativ      │   Historie   │
└──────┬───────────┘         └──────┬─────────┴──────┬───────┘
       │                            │                │
       ▼                            ▼                ▼
┌──────────────────────┐   ┌──────────────┐   ┌────────────┐
│   EventRepo.         │   │ table_state  │   │  events    │
│   WriteEvent()       │   │ (Projektion) │   │  (Stream)  │
│                      │   └──────────────┘   └────────────┘
│   BEGIN TX           │          ▲
│    INSERT event  ────┼──────────┤
│    ApplyEvent()      │   synchron in
│    UPSERT state  ────┘   selber TX
│   COMMIT TX          │
└──────────────────────┘
```

## Konsequenzen

### Positiv

- **Strong Consistency überall** — Projektion in derselben TX wie Event-INSERT, kein Stale State, kein Eventual-Consistency-Problem.
- **Ein Projektionsmechanismus** — Nur synchrone Projektion, kein zweiter Mechanismus (Lazy, Worker) zu pflegen.
- **Snapshot-Ablösung** — `tisch.snapshot:v1` wird nicht mehr erzeugt. Der Event Stream enthält nur noch fachliche Domänen-Events. Das Kassenjournal ist konzeptionell sauber.
- **Triviale Reads** — Tischübersicht = 1 SELECT auf `table_state`. Saldo, unbezahlte/ungelieferte Positionen direkt verfügbar.
- **Reporting-Enabler** — Vorberechneter Saldo und Positionen in `table_state` vereinfachen tischübergreifende Aggregation.
- **Selbstheilend** — Bei Inkonsistenz kann `table_state` jederzeit aus Events reberechnet werden (alle Events bleiben append-only erhalten).

### Negativ

- **Write-Pfad minimal komplexer** — 1 INSERT + 1 UPSERT pro Transaktion (statt nur 1 INSERT).
- **Apply-Funktion muss konsistent sein** — `ApplyEvent()` ist die einzige Zustandsberechnung. Testabdeckung ist essenziell.
- **JSONB in Projektionstabelle** — `unbezahlte_positionen` und `ungelieferte_positionen` sind JSONB-Spalten. Typsicherheit nur auf Go-Ebene, nicht auf DB-Ebene.

## Referenzen

- [ADR: Event-Sourcing für Tisch-Operationen](event-sourcing.md) — Persistenz-Strategie, Event-Modell, Append-Only-Garantie
- [Handbuch §3](../handbuch.md) — Domain-Modell, Tisch-Aggregat, Invarianten
- [Anforderungen](../anforderungen.md) — K-05 (Tischübersicht), K-06 (Kassenjournal), R-01–R-05 (Reporting)
