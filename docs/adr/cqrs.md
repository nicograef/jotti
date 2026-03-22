# ADR: CQRS-Projektionen für Kasse-Kontext

## Status

**Accepted** — Zwei synchrone Projektionen (`tisch_session_state`, `kassensitzung_state`) für operative Queries, On-Demand SQL für Reporting. Ersetzt vorherige ADR (Lazy Projection + Background Worker) und vorherige Fassung (einzelne `table_state`-Projektion).

## Kontext

### Ausgangslage (vor dieser ADR)

jotti implementierte CQRS Stufe 1 (logische Trennung): Getrennte Command-/Query-Services, separate HTTP-Handler und segregierte Repository-Interfaces. Der Tisch-Zustand wurde bei jedem Lesezugriff durch Event-Replay berechnet, optimiert durch einen Snapshot-as-Event-Mechanismus (`tisch.snapshot:v1`).

### Problemstellung

1. **Snapshot-as-Event ist ein Anti-Pattern.** Das Snapshot-Event (`tisch.snapshot:v1`) war ein Infrastruktur-Artefakt im fachlichen Event Stream. Es vermischte Domänen-Events mit Optimierungs-Artefakten und untergrub die Klarheit des Kassenjournals.
2. **N Replays für Tischübersicht.** Die Tischübersicht (K-06) erforderte bei 50 Tischen 50 separate Event-Replays für die Saldo-Berechnung — konzeptionell unbefriedigend, auch wenn bei jottis Lastprofil (< 200 Events/Tisch) die Performance ausreichte.
3. **Kein Reporting-Pfad.** Tischübergreifende Auswertungen (R-01 bis R-05) erforderten JSONB-Parsing über alle Events — keine vorberechneten Aggregate vorhanden.
4. **Kassensitzungs-Hot-Path.** Im vereinheitlichten Kasse-Kontext muss bei jedem Tisch-Command geprüft werden, ob eine offene Kassensitzung existiert (KS-Sperre). Event Replay pro Zugriff ist nicht mehr akzeptabel — ein direkter SELECT auf eine Projektion ist erforderlich.

## Bewertete Alternativen

### A: Status quo (Snapshot + Event-Replay)

| +                               | -                                          |
| ------------------------------- | ------------------------------------------ |
| Kein Aufwand, existiert bereits | Snapshot-as-Event bleibt unrein            |
| Minimal komplex                 | Tischübersicht = N Replays                 |
| Strong Consistency              | Reporting-SQL wird komplex (JSONB-Parsing) |

### B: Synchrone Projektion (gewählt)

Zwei Projektionstabellen (`tisch_session_state`, `kassensitzung_state`), jeweils UPSERT in derselben Transaktion wie Event-INSERT. Routing über expliziten `StreamType`-Parameter.

| +                                                           | -                                                  |
| ----------------------------------------------------------- | -------------------------------------------------- |
| Löst Snapshot-as-Event-Problem                              | Write wird minimal langsamer (ein UPSERT mehr)     |
| Vorhersagbare Latenz (Read = 1 SELECT)                      | Apply-Funktionen müssen konsistent mit Events sein |
| Tischübersicht trivial: `SELECT * FROM tisch_session_state` | Neuer Code (~200–300 Zeilen Apply-Funktionen)      |
| KS-Sperre trivial: `SELECT * FROM kassensitzung_state`      |                                                    |
| Strong Consistency ohne Tricks                              |                                                    |
| Reporting-Enabler (Saldo vorberechnet)                      |                                                    |

### C: Lazy Projection

Read-Through-Cache: Beim Lesezugriff prüfen, ob Projektion aktuell ist, fehlende Events bei Bedarf replayed.

| +                                     | -                                       |
| ------------------------------------- | --------------------------------------- |
| Command bleibt „rein" (single INSERT) | Unvorhersagbare Read-Latenz             |
| Self-healing                          | Fast identisch zum Snapshot-Mechanismus |
|                                       | Staleness-Logik, Cache-Invalidierung    |

**Verworfen:** Löst konzeptionell dasselbe Problem wie der bestehende Snapshot-Mechanismus — statt „lade Snapshot + replay Delta" wird es „lade Projektion + prüfe Aktualität + replay Delta". Der Architekturgewinn gegenüber dem Status quo ist marginal.

### D: Background Worker (für Analytik)

Worker pollt `kassenjournal`-Tabelle und projiziert asynchron auf analytische Tabellen.

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

**Zwei synchrone Projektionen (`tisch_session_state` + `kassensitzung_state`) für operative Queries. On-Demand SQL-Aggregation für Reporting. Kein Background Worker.**

### Operative Projektionen — Synchrone Projektion (Write-Through)

Die `tisch_session_state`- und `kassensitzung_state`-Tabellen werden in derselben Transaktion wie das Event-INSERT aktualisiert. Ein expliziter `StreamType`-Parameter steuert das Routing:

```
BEGIN TX
  INSERT INTO kassenjournal (...)     → event_id
  StreamType-Routing:
    "tisch-session" →
      SELECT * FROM tisch_session_state  → aktueller Zustand (oder Zero-Value)
      ApplyTischSessionEvent(zustand, event)  → neuer Zustand
      UPSERT INTO tisch_session_state (...)   → persistiert
    "kassensitzung" →
      SELECT * FROM kassensitzung_state  → aktueller Zustand (oder Zero-Value)
      ApplyKassensitzungEvent(zustand, event) → neuer Zustand
      UPSERT INTO kassensitzung_state (...)   → persistiert
COMMIT TX
```

Die `ApplyTischSessionEvent()`- und `ApplyKassensitzungEvent()`-Funktionen (`backend/domain/kasse/`) sind reine Funktionen in der Domain-Schicht ohne DB-Zugriff. Sie verarbeiten die jeweiligen Domänen-Event-Typen und berechnen den neuen Zustand.

**Warum synchron statt lazy?**

- **Vorhersagbare Latenz** für alle Reads — kein implizites Write-on-Read
- **Gleiche Komplexität** wie Lazy (beide brauchen `ApplyEvent()`), aber einfacheres Konsistenzmodell
- **Strong Consistency** ohne Staleness-Detection oder Rebuild-Pfad

**CQRS-Trennung bleibt intakt:** Der Projektor ist ein internes Detail von `KassenjournalRepo.WriteEvent()`. Der Command-Service ruft weiterhin nur `WriteEvent()` auf — das UPSERT auf die jeweilige Projektionstabelle ist transparent. Die Query-Services lesen direkt aus `tisch_session_state` bzw. `kassensitzung_state`.

### Kassenjournal (Historie) — Event-Replay (Stufe 1)

Die Historie _ist_ der Event Stream. `GetTischHistorie()` liest weiterhin alle Events via `ReadEventsBySubject()` und formatiert sie über `GetHistoryFromEvents()`. Kein Read Model nötig.

### Reporting (R-01–R-05) — On-Demand SQL-Aggregation

| Aspekt     | Entscheidung                                                      |
| ---------- | ----------------------------------------------------------------- |
| Strategie  | SQL-Queries über `kassenjournal`-Tabelle + `tisch_session_state`  |
| Konsistenz | Strong (immer aktuell, on-demand berechnet)                       |
| Filter     | `kassensitzung_nr` statt Zeitraum (`von`/`bis`)                   |
| Fallback   | Materialized Views bei Bedarf (unwahrscheinlich bei < 10k Events) |

Bei ~10.000 Events ist jede Aggregation in < 100ms erledigt. Kein Background Worker, keine Eventual Consistency, keine Checkpoint-Tabelle.

### CQRS-Stufen im Überblick

| Bereich                  | CQRS-Stufe                         | Strategie                                                         | Konsistenz |
| ------------------------ | ---------------------------------- | ----------------------------------------------------------------- | ---------- |
| Kasse (operativ)         | **Stufe 2** — Synchrone Projektion | `tisch_session_state` + `kassensitzung_state`, UPSERT in Event-TX | Strong     |
| Kassenjournal (Historie) | **Stufe 1** — Event-Replay         | Event Stream = Read Model                                         | Strong     |
| Reporting (R-01–R-05)    | **Stufe 1** — On-Demand SQL        | `kassenjournal` + `tisch_session_state`                           | Strong     |
| Stammdaten (CRUD)        | **Stufe 0** — Kein CQRS            | Kein Event-Sourcing                                               | Strong     |

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
┌────────────────────────┐   ┌─────────────────────┐  ┌──────────────┐
│ KassenjournalRepo.     │   │ tisch_session_state │  │ kassenjournal│
│ WriteEvent()           │   │ (Tisch-Projektion)  │  │ (Stream)     │
│                        │   └─────────────────────┘  └──────────────┘
│ BEGIN TX               │          ▲
│  INSERT kassenjournal  │          │
│  StreamType routing ───┤── synchron in
│  ApplyEvent()          │   selber TX
│  UPSERT state      ───┘
│ COMMIT TX              │   ┌─────────────────────┘
└────────────────────────┘   │ kassensitzung_state
                               │ (KS-Projektion)
                               └────────────────────
```

## Konsequenzen

### Positiv

- **Strong Consistency überall** — Projektionen in derselben TX wie Event-INSERT, kein Stale State, kein Eventual-Consistency-Problem.
- **Ein Projektionsmechanismus** — Nur synchrone Projektion mit StreamType-Routing, kein zweiter Mechanismus (Lazy, Worker) zu pflegen.
- **Snapshot-Ablösung** — `tisch.snapshot:v1` wird nicht mehr erzeugt. Der Event Stream enthält nur noch fachliche Domänen-Events. Das Kassenjournal ist konzeptionell sauber.
- **Triviale Reads** — Tischübersicht = 1 SELECT auf `tisch_session_state`. KS-Sperre = 1 SELECT auf `kassensitzung_state`. Saldo, unbezahlte/ausstehende Positionen direkt verfügbar.
- **Reporting-Enabler** — Vorberechneter Saldo und Positionen in `tisch_session_state` vereinfachen tischübergreifende Aggregation. Filter über `kassensitzung_nr`.
- **Selbstheilend** — Bei Inkonsistenz können die Projektionen jederzeit aus dem Kassenjournal reberechnet werden (alle Events bleiben append-only erhalten).

### Negativ

- **Write-Pfad minimal komplexer** — 1 INSERT + 1 UPSERT pro Transaktion (statt nur 1 INSERT). Bei Kassensitzung-Events zusätzlich Routing-Logik.
- **Apply-Funktionen müssen konsistent sein** — `ApplyTischSessionEvent()` und `ApplyKassensitzungEvent()` sind die einzigen Zustandsberechnungen. Testabdeckung ist essenziell.
- **JSONB in Projektionstabelle** — `unbezahlte_positionen` und `ausstehende_positionen` in `tisch_session_state` sind JSONB-Spalten. Typsicherheit nur auf Go-Ebene, nicht auf DB-Ebene.

## Referenzen

- [ADR: Event-Sourcing für Kasse-Operationen](event-sourcing.md) — Persistenz-Strategie, Event-Modell, Append-Only-Garantie
- [Handbuch §3](../handbuch.md) — Domain-Modell, Tisch-Session, Invarianten
- [Anforderungen](../anforderungen.md) — K-06 (Tischübersicht), K-07 (Kassenjournal), R-01–R-05 (Reporting)
