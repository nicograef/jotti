# CQRS — Theorie

Dieses Dokument dient als theoretisches Nachschlagewerk für Command Query Responsibility Segregation (CQRS). Es erklärt das Muster, seine Ausbaustufen, Read Model Design, Projektionsstrategien, Eventual Consistency und Entscheidungskriterien.

---

## Inhaltsverzeichnis

1. [CQRS — Grundlagen](#1-cqrs--grundlagen)
2. [Read Model Design](#2-read-model-design)
3. [Projektionsstrategien](#3-projektionsstrategien)
4. [Eventual Consistency](#4-eventual-consistency)
5. [CQRS ohne Event-Sourcing](#5-cqrs-ohne-event-sourcing)
6. [Kombination mit Event-Sourcing](#6-kombination-mit-event-sourcing)
7. [Entscheidungsmatrix: CQRS vs. CRUD](#7-entscheidungsmatrix-cqrs-vs-crud)
8. [Anti-Patterns](#8-anti-patterns)
9. [Referenzen](#9-referenzen)

---

## 1. CQRS — Grundlagen

### 1.1 Grundidee: Von CQS zu CQRS

**Command Query Responsibility Segregation** trennt die Verantwortlichkeit für Schreiboperationen (Commands) und Leseoperationen (Queries) auf System-Ebene. CQRS erweitert das CQS-Prinzip (Command Query Separation) von Bertrand Meyer:

> **CQS (Methoden-Ebene):** Eine Methode soll entweder den Zustand ändern (Command) ODER Daten zurückgeben (Query) — nie beides.
>
> **CQRS (System-Ebene):** Das System hat **separate Modelle** für Schreiben und Lesen.

> **Greg Young:** _"CQRS uses the same definition of Commands and Queries that Meyer used and maintains the viewpoint that they should be pure. The fundamental difference is that in CQRS objects are split into two objects, one containing the Commands one containing the Queries."_

> **Martin Fowler:** _"The change that CQRS introduces is to split that conceptual model into separate models for update and display. The rationale is that for many problems, particularly in more complicated domains, having the same conceptual model for commands and queries leads to a more complex model that does neither well."_

Wichtig: CQRS sagt **nichts darüber aus, wo Daten gespeichert werden**. Command Handler und Query Handler können dieselbe Datenbank, dieselben Tabellen verwenden — solange sie konzeptuell getrennt sind. Zwei Datenbanken sind eine mögliche, aber keine notwendige Konsequenz.

### 1.2 Command Side

Commands drücken eine **Absicht** aus: „Gib diese Bestellung auf", „Registriere diese Zahlung". Commands:

- Ändern den Zustand
- Geben maximal eine ID oder Erfolgsmeldung zurück
- Werden validiert (Autorisierung, Geschäftsregeln)
- Können abgelehnt werden (z.B. ungültige Daten, Concurrency-Konflikt)
- Sind idealerweise **idempotent**: Mehrfaches Absenden führt zum selben Ergebnis

```go
// Command: Absichtserklärung (Domänen-Sprache)
type PlaceOrderCommand struct {
    TableID         int
    Items           []OrderItem
    IdempotencyKey  string  // verhindert doppelte Verarbeitung
}

// Command Handler: Validierung + Business Rules + Persistenz
func (h *Handler) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error {
    if err := h.validateCommand(cmd); err != nil {
        return err
    }
    // Idempotenz prüfen
    if h.alreadyProcessed(ctx, cmd.IdempotencyKey) {
        return nil
    }
    // Business Rules prüfen, Event erzeugen und speichern
    return h.eventRepo.WriteEvent(ctx, buildEvent(cmd))
}
```

### 1.3 Query Side

Queries fragen Daten ab, **ohne den Zustand zu ändern**. Queries:

- Haben keine Seiteneffekte
- Geben optimierte Datenstrukturen (Read Models / DTOs) zurück
- Können gecacht werden
- Können gegen separate, leseoptimierte Datenspeicher arbeiten

```go
// Query: Datenanfrage ohne Seiteneffekte
type GetOrderSummaryQuery struct {
    OrderID int
}

// Read Model: auf die Abfrage zugeschnittene Datenstruktur
type OrderSummaryDTO struct {
    OrderID          int
    BalanceCents     int
    OpenItems        []ItemDTO
    LastPaymentAt    *time.Time
}

// Query Handler: liest aus optimiertem Read Model
func (h *Handler) GetOrderSummary(ctx context.Context, q GetOrderSummaryQuery) (OrderSummaryDTO, error) {
    return h.readRepo.FindOrderSummary(ctx, q.OrderID)
}
```

### 1.4 CQRS-Ausbaustufen

CQRS ist kein binäres Muster — es gibt Abstufungen:

#### Stufe 0: Kein CQRS

Ein einziges Modell für Lesen und Schreiben. Klassisches CRUD.

```
Client → Service → Repository → Datenbank
```

#### Stufe 1: Logische Trennung

Separate Command- und Query-Handler, aber **gleiche Datenbank** und gleiches Datenmodell. Bereits hier: Change Tracking in Queries deaktivieren, separate Codemodule für Commands und Queries.

```
Client → Command Handler ─┐
                          ├── Repository → Datenbank
Client ← Query Handler  ──┘
```

#### Stufe 2: Getrennte Modelle (Projektionen)

Command Side schreibt in Write Store, eine **Projektion** synchronisiert Daten in ein Read Model. Abfragen lesen aus dem optimierten Read Model.

```
Client → Command Handler → Write Store (Events/Tabellen)
                               │
                               ▼ Projektion (synchron/asynchron)
                               │
Client ← Query Handler  ← Read Store (Projektions-Tabellen)
```

#### Stufe 3: Getrennte Datenbanken

Write Store und Read Store in **separaten Datenbank-Instanzen** oder -Technologien. Maximale Skalierbarkeit, aber höchste Komplexität.

```
Command Handler → PostgreSQL (Write/Events)
                       │
                       ▼ Async Projektion (LISTEN/NOTIFY, CDC, Message Queue)
                       │
Query Handler   ← Redis / Elasticsearch / Read-Replica / separate PostgreSQL-Instanz
```

### 1.5 Vor- und Nachteile von CQRS

| Vorteil                                 | Beschreibung                                                   |
| --------------------------------------- | -------------------------------------------------------------- |
| **Optimierte Lesemodelle**              | Read Models können exakt auf Abfragen zugeschnitten sein       |
| **Unabhängige Skalierung**              | Lese- und Schreiblast getrennt skalierbar                      |
| **Vereinfachte Modelle**                | Jedes Modell tut nur eine Sache, kein Kompromiss-Design        |
| **Natürliche Event-Sourcing-Ergänzung** | Events = Write Model, Projektionen = Read Model                |
| **Task-basierte UIs**                   | Commands spiegeln Benutzerintentionen wider (keine CRUD-Maske) |
| **Vertical Slices**                     | Jeder Handler ist ein unabhängiger Code-Silo                   |

| Nachteil                          | Beschreibung                                       |
| --------------------------------- | -------------------------------------------------- |
| **Eventual Consistency**          | Read Model kann kurzzeitig veraltet sein           |
| **Mehr Code**                     | Separate Handler, Modelle, Projektionen            |
| **Synchronisationskomplexität**   | Projektion muss zuverlässig synchron bleiben       |
| **Overkill für einfache Domains** | CRUD-Entities profitieren nicht von CQRS           |
| **Mentaler Overhead**             | Entwickler müssen das Muster vollständig verstehen |

---

## 2. Read Model Design

### 2.1 Grundprinzip: Daten für den Lesezweck optimieren

Ein **Read Model** (auch: View Model, Projection, Query Model) ist eine Datenstruktur, die **speziell für eine Abfrage oder eine Gruppe von Abfragen** optimiert wurde. Im Gegensatz zum Write Model, das auf Konsistenz und Geschäftsregeln optimiert ist, ist das Read Model auf **Leseperformance und Einfachheit** optimiert.

> **Udi Dahan:** _"Create an additional data store whose structure mirrors the view model. One table for each view. Then our client could simply SELECT \* FROM MyViewTable and bind the result to the screen."_

Das Grundprinzip: Statt Joins und Transformationen zur Abfragezeit vorab zu berechnen und zu materialisieren.

### 2.2 Denormalisierung

Relationale Datenbanken sind normalisiert, um Redundanz zu vermeiden. Read Models dürfen (und sollten) **bewusst denormalisiert** sein:

```sql
-- Normalisiert (Write Model): 3 Tabellen, 2 Joins nötig
SELECT o.id, u.name, SUM(oi.price) AS total
FROM orders o
JOIN users u ON u.id = o.user_id
JOIN order_items oi ON oi.order_id = o.id
GROUP BY o.id, u.name;

-- Denormalisiert (Read Model): 1 Tabelle, kein Join
SELECT id, customer_name, total_cents FROM order_summaries WHERE id = $1;
```

**Vorteile:** Einfache Queries, keine Joins, niedrige Latenz.
**Nachteile:** Redundante Datenhaltung, Synchronisationsaufwand bei Änderungen.

### 2.3 Materialized Views

**Materialized Views** sind vorberechnete Abfrageergebnisse, die physisch gespeichert werden. PostgreSQL unterstützt sie nativ:

```sql
CREATE MATERIALIZED VIEW order_daily_stats AS
SELECT
    DATE(created_at) AS day,
    COUNT(*)          AS order_count,
    SUM(total_cents)  AS revenue_cents
FROM orders
GROUP BY DATE(created_at);

-- Refresh (manuell oder via Trigger/Scheduler):
REFRESH MATERIALIZED VIEW CONCURRENTLY order_daily_stats;
```

**Anwendungsfälle:** Aggregationen (Statistiken, Dashboards), komplexe Berechnungen, die selten aktualisiert werden müssen.

**Einschränkungen:** `REFRESH MATERIALIZED VIEW` blockiert Reads (außer mit `CONCURRENTLY`). Nicht für Echtzeit-Daten geeignet.

### 2.4 Search Indexes (Elasticsearch)

Für Volltext-Suche, Facettierung und analytische Abfragen reicht eine relationale DB oft nicht aus. **Elasticsearch** (oder OpenSearch) als Read Model:

```
Write: PostgreSQL (Write Store)
           │
           ▼ Async Projektion
           │
Read:  Elasticsearch (Search Index)
```

**Anwendungsfälle:** Produktsuche mit Filtern/Facetten, Volltextsuche in Dokumenten, Log-Analyse, kombinierte Suche über mehrere Entitäten.

**Synchronisierung:** Über Change Data Capture (CDC) oder explizite Projektion nach jedem Write.

### 2.5 Multiple Read Models für verschiedene Use-Cases

Derselbe Event-Stream kann mehrere Read Models bedienen — jedes optimiert für einen anderen Anwendungsfall:

```
events-Tabelle (Write Store)
        │
        ├──▶ order_summaries     (für Bestellübersicht — schnelle SELECT)
        ├──▶ customer_history    (für Kundenprofil — alle Bestellungen)
        ├──▶ daily_stats         (für Dashboard — aggregierte Zahlen)
        └──▶ elasticsearch_index (für Volltextsuche)
```

**Vorteil:** Jede Abfrage erhält genau die Datenstruktur, die sie braucht — keine Kompromisse.

**Wichtig:** Read Models sind **sekundäre Daten**. Die Events sind die einzige Source of Truth. Bei Inkonsistenz kann das Read Model jederzeit aus den Events neu aufgebaut werden (Projektion-Rebuild).

---

## 3. Projektionsstrategien

### 3.1 Synchrone Projektion

Das Read Model wird **innerhalb derselben Transaktion** wie das Event aktualisiert. Keine Eventual Consistency.

```go
func (s *CommandService) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error {
    return s.db.WithTx(ctx, func(tx pgx.Tx) error {
        // 1. Event in Write Store schreiben
        if err := s.eventRepo.WithTx(tx).WriteEvent(ctx, event); err != nil {
            return err
        }
        // 2. Read Model in gleicher Transaktion aktualisieren
        return s.readRepo.WithTx(tx).UpdateOrderSummary(ctx, newSummary)
    })
}
```

**Vorteile:** Starke Konsistenz, kein Sync-Problem, einfaches Error Handling (Transaktion rollt alles zurück).
**Nachteile:** Write-Latenz steigt (mehr DB-Operationen pro Request), enge Kopplung zwischen Write- und Read-Seite.

**Geeignet für:** Kritische Daten, bei denen veraltete Anzeige inakzeptabel ist (Kontostände, Inventar).

### 3.2 Asynchrone Projektion

Das Read Model wird **nach dem Request** in einem separaten Prozess aktualisiert.

#### Variante A: Polling

Ein Background-Worker prüft regelmäßig auf neue Events und aktualisiert das Read Model:

```
Event INSERT → (Worker prüft alle 5s) → Read Model UPDATE
```

**Einfach zu implementieren**, aber Polling-Overhead und Latenz bis zum nächsten Poll-Intervall.

#### Variante B: LISTEN/NOTIFY (PostgreSQL)

PostgreSQL sendet eine Benachrichtigung, sobald ein Event geschrieben wird:

```sql
-- Trigger nach Event INSERT:
CREATE OR REPLACE FUNCTION notify_new_event() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('new_event', row_to_json(NEW)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

```go
// Projektor wartet auf Benachrichtigungen:
conn.Listen(ctx, "new_event")
for {
    notification := conn.WaitForNotification(ctx)
    projectReadModel(notification.Payload)
}
```

**Niedrige Latenz**, kein Polling-Overhead. Nur innerhalb einer PostgreSQL-Instanz einsetzbar.

#### Variante C: Message Queue

Events werden in eine Message Queue (Kafka, RabbitMQ, NATS) publiziert, ein Projektor konsumiert sie:

```
Event INSERT → Outbox → Message Broker → Projektor → Read Model UPDATE
```

**Vorteile:** Skalierbar, entkoppelt, mehrere Konsumenten möglich.
**Nachteile:** Zusätzliche Infrastruktur, komplexeres Deployment.

**Geeignet für:** Nicht-kritische Read Models (Statistiken, Historien), Read Models, die für viele Konsumenten aufgebaut werden.

### 3.3 Hybride Projektion

Kritische Read Models synchron, unkritische asynchron — in derselben Anwendung:

```
Bestellung aufgegeben → Event INSERT
    │
    ├──▶ Synchron (gleiche TX):   Kontostand, offene Positionen (kritisch)
    └──▶ Asynchron (Background):  Tagesstatistik, Kundenhistorie (unkritisch)
```

**Empfehlung:** Mit synchroner Projektion starten und schrittweise zu asynchron migrieren, wenn Skalierungsbedarf entsteht.

### 3.4 Change Data Capture (CDC)

**Change Data Capture** liest den **Write-Ahead Log (WAL)** der Datenbank und reagiert auf Änderungen, ohne dass die Anwendung explizit Events publizieren muss.

```
PostgreSQL WAL → Debezium (CDC Connector) → Kafka Topic → Projektor → Read Model
```

**Debezium** ist das bekannteste Open-Source-CDC-Tool:

```json
// Debezium-Nachricht nach INSERT in orders-Tabelle:
{
  "op": "c",
  "after": { "id": 101, "customer_id": 42, "total_cents": 4200 }
}
```

**Vorteile:**

- Keine Änderungen an der Anwendung nötig (DB-agnostisch)
- Erfasst auch DDL-Änderungen
- Garantierte Delivery (at-least-once via Kafka)

**Nachteile:**

- Komplexe Infrastruktur (Kafka, Debezium Connector)
- Erfordert PostgreSQL Logical Replication aktiviert
- Bindet an DB-Schema (strukturelle Änderungen = Breaking Changes)

**Geeignet für:** Legacy-Systeme ohne Event-Sourcing, die CQRS-Read-Models brauchen. Oder wenn die Write-Seite nicht angepasst werden kann.

### 3.5 Transactional Outbox

Das **Outbox Pattern** löst das Problem, dass Event-Publishing und Datenbankschreiben nicht atomar sind:

```
Ohne Outbox: DB-Schreiben ✓, Event-Publishing ✗ → Inkonsistenz
Mit Outbox:  DB-Schreiben + Outbox-INSERT in einer TX → Background Worker publiziert
```

```sql
-- Outbox-Tabelle:
CREATE TABLE outbox_events (
    id          BIGSERIAL PRIMARY KEY,
    event_type  TEXT NOT NULL,
    payload     JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed   BOOLEAN NOT NULL DEFAULT false
);
```

```go
// In einer Transaktion: Domain-Schreiben + Outbox-INSERT
func (s *Service) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error {
    return s.db.WithTx(ctx, func(tx pgx.Tx) error {
        s.orderRepo.WithTx(tx).Save(ctx, order)
        s.outboxRepo.WithTx(tx).Insert(ctx, OutboxEvent{
            EventType: "order.placed",
            Payload:   marshalOrder(order),
        })
        return nil
    })
}
// Background Worker liest Outbox und publiziert in Message Queue
```

**Garantie:** At-Least-Once Delivery. Der Projektor muss idempotent sein (doppelte Events ignorieren).

**Kombination mit CDC:** Statt eines Background Workers kann CDC (Debezium) die Outbox-Tabelle lesen — das nennt sich **Transactional Outbox + CDC**, eine robuste Kombination für verteilte Systeme.

---

## 4. Eventual Consistency

### 4.1 Grundkonzept

Bei asynchronen Projektionen kann das Read Model **kurzzeitig veraltet** sein. Nach einem Command ist garantiert, dass das Read Model **irgendwann** (eventually) konsistent wird — aber nicht sofort.

> **Werner Vogels (Amazon CTO):** _"Eventually consistent — the storage system guarantees that if no new updates are made to the object, eventually all accesses will return the last updated value."_

Eventual Consistency ist kein Makel, sondern ein bewusster Trade-off: **Höhere Schreibperformance und Skalierbarkeit** gegen **kurze Inkonsistenzfenster**.

### 4.2 Read-Your-Own-Writes

Ein häufiges Problem: Benutzer A schreibt Daten, liest aber sofort danach das veraltete Read Model.

**Strategien:**

```
1. Synchrone Projektion (kein Staleness-Problem)
2. Timestamp-basiert: Client schickt Write-Timestamp, Server wartet auf Sync
3. Version-basiert: Client schickt erwartete Version, Server wartet bis >= Version
4. Optimistic Update im Frontend: UI zeigt sofort den erwarteten Zustand
```

```go
// Version-basiert: Query wartet auf Konsistenz
func (h *QueryHandler) GetOrderSummary(ctx context.Context, q GetOrderSummaryQuery) (OrderSummaryDTO, error) {
    if q.MinVersion > 0 {
        if err := h.waitForVersion(ctx, q.OrderID, q.MinVersion, 500*time.Millisecond); err != nil {
            // Timeout: veraltete Daten zurückgeben mit Staleness-Flag
            return h.readRepo.FindWithStalenessFlag(ctx, q.OrderID)
        }
    }
    return h.readRepo.Find(ctx, q.OrderID)
}
```

### 4.3 Session Consistency & Causal Consistency

**Session Consistency:** Ein Client sieht immer seine eigenen Writes — auch wenn andere Clients noch ältere Daten sehen.

**Causal Consistency:** Wenn Aktion B kausal von Aktion A abhängt, sieht jeder Client B erst nachdem er A gesehen hat.

**Praktische Umsetzung:**

- Client speichert den letzten bekannten Write-Token (Sequenznummer, Timestamp)
- Server-seitig: Sticky Sessions zu einer Replica oder Read-Model-Version prüfen

### 4.4 Compensation / Corrective Events

Statt inkonsistente Zustände zu verhindern, werden sie **nachträglich korrigiert**:

```
Command:      Bestellung aufgegeben (Artikel A reserviert)
Concurrency:  Gleichzeitig: Artikel A vom Lager als nicht verfügbar markiert
Compensation: Saga erkennt Konflikt → sendet "Bestellung-storniert"-Event
```

Das Compensation-Muster ist zentral in **Sagas** (verteilten Transaktionen ohne 2-Phase-Commit).

### 4.5 UI-Strategien

Das Frontend kann Eventual Consistency durch UX-Tricks unsichtbar machen:

| Strategie                  | Beschreibung                                                           | Beispiel                                          |
| -------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------- |
| **Optimistic Update**      | UI zeigt sofort den erwarteten Zustand, rollt bei Fehler zurück        | Like-Button springt sofort auf "Liked"            |
| **Stale-While-Revalidate** | Sofort gecachte Daten zeigen, im Hintergrund aktualisieren             | Produktliste lädt sofort, aktualisiert sich still |
| **Polling**                | Nach einem Command regelmäßig abfragen, bis neuer Zustand sichtbar ist | Nach Zahlung alle 2s neu laden bis "Bezahlt"      |
| **WebSocket-Push**         | Server benachrichtigt Client wenn Read Model aktualisiert              | Echtzeit-Dashboard                                |
| **Pending State anzeigen** | "Ihre Bestellung wird verarbeitet..." bis Projektion abgeschlossen     | Task-basierte UI mit Fortschrittsanzeige          |

---

## 5. CQRS ohne Event-Sourcing

### 5.1 Eigenständiger Wert von CQRS

Ein verbreitetes Missverständnis: CQRS benötige Event-Sourcing. Das ist falsch.

> **Oskar Dudycz:** _"CQRS can be used without Event Sourcing. Projections are a concept that's part of Event-Driven Architecture and essential building block of CQRS. They're not tied to Event Sourcing."_

CQRS liefert **ohne ES** bereits erheblichen Wert:

- **Klare Trennung:** Commands und Queries sind konzeptuell getrennt → bessere Lesbarkeit
- **Optimierte Queries:** Query Handler können direkt gegen optimierte Strukturen arbeiten
- **Unabhängige Skalierung:** Read-Replicas für Queries, primäre DB für Writes
- **Change Tracking:** In Query Handlern deaktiviert (kein ORM-Overhead)

### 5.2 CQRS mit klassischer RDBMS

Stufe 1 und 2 sind ohne ES problemlos umsetzbar:

```
Klassische DB (PostgreSQL)
├── Tabelle: orders         (Write-optimiert, normalisiert)
├── Tabelle: order_items    (Write-optimiert)
│
└── Materialized View: order_summaries  (Read-optimiert, denormalisiert)
    → wird bei jedem Write aktualisiert oder via REFRESH
```

```go
// Separate Handler für Commands und Queries
// Commands: UPDATE/INSERT mit Domain-Validierung
func (h *CommandHandler) PlaceOrder(ctx, cmd) error { ... }

// Queries: direkt gegen materialized view — kein ORM, kein Join
func (h *QueryHandler) GetOrderSummary(ctx, orderID int) (OrderSummaryDTO, error) {
    return h.db.QueryRow(ctx,
        "SELECT id, customer_name, total_cents, item_count FROM order_summaries WHERE id = $1",
        orderID,
    ).Scan(...)
}
```

### 5.3 CQRS mit Read Replicas

Der einfachste Weg zu CQRS Stufe 3 ohne ES: **Read Replicas** der Datenbank als Query-Ziel:

```
Write: Primary PostgreSQL   (alle Commands)
Read:  Read Replica 1       (Query Handler)
       Read Replica 2       (Reporting)
       Read Replica 3       (Analytics)
```

**AWS-Beispiel:** RDS Primary (Writes) + RDS Read Replicas (Queries). Replikationslatenz typischerweise < 100ms.

**Einschränkung:** Das Datenbankschema ist für Writes und Reads gleich — keine vollständige Read-Model-Optimierung. Für spezifische Abfragen lohnen sich separate materialisierte Read Models mehr.

### 5.4 Schrittweise Einführung

CQRS muss nicht auf einmal eingeführt werden. Empfohlener Pfad:

```
Stufe 0 → Stufe 1:  Command/Query Handler trennen (refactoring)
Stufe 1 → Stufe 2:  Read Model für teuerste Queries materialisieren
Stufe 2 → Stufe 3:  Read Store auf separate DB auslagern (nur wenn nötig)
```

**Martin Fowler:** _"CQRS should only be used on specific portions of a system (a Bounded Context) and not the system as a whole."_ — Nicht das gesamte System auf CQRS umstellen, sondern nur dort, wo der Nutzen klar überwiegt.

---

## 6. Kombination mit Event-Sourcing

### 6.1 Warum sie zusammenpassen

Event-Sourcing allein hat ein **Lese-Performance-Problem**: Jede Query muss den Event-Stream replayed werden. CQRS löst dieses Problem durch separate Read Models (Projektionen).

Umgekehrt macht CQRS eine klare Trennung zwischen Write- und Query-Seite — Events sind das natürliche Write Model, Projektionen das natürliche Read Model.

```
┌──────────────────────────────────────────────────────────────┐
│                      Command Side                            │
│                                                              │
│  Client → HTTP Handler → Command Service → Event Repository  │
│                                              │ INSERT Event  │
│                                              ▼               │
│                                    ┌──────────────────┐      │
│                                    │  events-Tabelle   │      │
│                                    │  (append-only)    │      │
│                                    └────────┬─────────┘      │
└─────────────────────────────────────────────┼────────────────┘
                                              │
                                    Projektion (sync/async)
                                              │
┌─────────────────────────────────────────────┼────────────────┐
│                      Query Side             ▼                │
│                                    ┌──────────────────┐      │
│  Client ← HTTP Handler ← Query    │  Read Model       │      │
│                          Service ← │  (z.B. Saldo-     │      │
│                                    │   Tabelle)        │      │
│                                    └──────────────────┘      │
└──────────────────────────────────────────────────────────────┘
```

### 6.2 Projektions-Rebuild

Ein einzigartiger Vorteil der ES+CQRS-Kombination: Read Models können jederzeit **vollständig neu aufgebaut** werden:

```
1. Read Model leeren (TRUNCATE)
2. Alle Events aus dem Event Store replayed
3. Projektor baut Read Model neu auf
→ garantiert konsistentes Ergebnis
```

**Anwendungsfälle:** Bug in Projektion behoben → Read Model neu aufbauen. Neues Read Model für neues Feature → aus bestehenden Events aufbauen.

### 6.3 Konsistenzanforderungen nach Feature

Nicht alle Read Models brauchen starke Konsistenz:

| Feature                   | Konsistenz           | Begründung                                     |
| ------------------------- | -------------------- | ---------------------------------------------- |
| **Kontostand/Saldo**      | Stark (synchron)     | Fehlerhafte Anzeige führt zu Fehlbuchungen     |
| **Offene Positionen**     | Stark (synchron)     | Abrechnungsvorgang muss korrekt sein           |
| **Lieferstatus**          | Stark (synchron)     | Lieferung muss vollständig sein                |
| **Transaktions-Historie** | Eventual (asynchron) | Historische Ansicht verträgt kurze Verzögerung |
| **Tagesabrechnung**       | Eventual (asynchron) | Wird nicht in Echtzeit benötigt                |
| **Umsatzstatistiken**     | Eventual (asynchron) | Aggregierte Daten, keine Echtzeitanforderung   |

Siehe [Event-Sourcing Theorie](event-sourcing.md#10-kombination-mit-cqrs) für die Grundlagen von Event-Sourcing und Snapshots.

---

## 7. Entscheidungsmatrix: CQRS vs. CRUD

### 7.1 Wann CQRS sinnvoll ist

| Kriterium                                      | CRUD          | CQRS Stufe 1 | CQRS Stufe 2+ |
| ---------------------------------------------- | ------------- | ------------ | ------------- |
| Einfache CRUD-Entities ohne Geschäftslogik     | ✅ Bevorzugt  | ❌ Overkill  | ❌ Overkill   |
| Komplexe Domäne mit vielen Geschäftsregeln     | ⚠️ Möglich    | ✅ Empfohlen | ✅ Empfohlen  |
| Lese-/Schreiblast sehr unterschiedlich         | ⚠️ Möglich    | ✅ Empfohlen | ✅ Empfohlen  |
| Reporting/Analytics neben operativem Betrieb   | ⚠️ Kompromiss | ⚠️ Besser    | ✅ Ideal      |
| Event-Sourcing im Einsatz                      | ❌ Aufwändig  | ✅ Empfohlen | ✅ Empfohlen  |
| Multiple Clients mit unterschiedlichen Sichten | ⚠️ Möglich    | ✅ Empfohlen | ✅ Ideal      |
| Team ohne CQRS-Erfahrung                       | ✅ Bevorzugt  | ⚠️ Lernkurve | ❌ Riskant    |
| Startup / MVP / wenig Traffic                  | ✅ Bevorzugt  | ⚠️ Optional  | ❌ Premature  |

### 7.2 Entscheidungs-Flowchart

```
Hat die Domäne komplexe Lese-/Schreibanforderungen?
├─ NEIN → CRUD reicht aus
└─ JA  → Sind Read- und Write-Anforderungen signifikant unterschiedlich?
          ├─ NEIN → CQRS Stufe 1 (logische Trennung)
          └─ JA  → Gibt es Performance-Probleme mit dem aktuellen Modell?
                    ├─ NEIN → CQRS Stufe 1 oder 2, evaluieren
                    └─ JA  → CQRS Stufe 2 (separate Projektionen)
                              └─ Brauchen Queries andere DB-Technologie?
                                  ├─ NEIN → Stufe 2 mit gleicher DB
                                  └─ JA  → Stufe 3 (separate Datenbanken)
```

### 7.3 Signale, die GEGEN CQRS sprechen

- Einfache CRUD-API ohne komplexe Geschäftslogik
- Team ohne Erfahrung mit Eventual Consistency
- Deadlines oder MVP-Druck — CRUD ist deutlich schneller
- Keine signifikante Last-Asymmetrie zwischen Reads und Writes
- Starke Konsistenz überall erforderlich (kein Spielraum für Eventual Consistency)

> **Martin Fowler:** _"Beware that for most systems CQRS adds risky complexity. Using CQRS on a domain that doesn't match it will add complexity, thus reducing productivity and increasing risk."_

---

## 8. Anti-Patterns

### 8.1 CQRS überall anwenden

**Problem:** Das gesamte System auf CQRS umstellen, auch einfache CRUD-Module.

```
// Anti-Pattern: CQRS für einfaches User-CRUD
type CreateUserCommand struct { Name, Email string }
type GetUserQuery struct { ID int }
// Für: name, email, status — keine Geschäftslogik → CRUD reicht
```

**Lösung:** CQRS nur in Bounded Contexts einsetzen, die von der Trennung profitieren. CRUD-Entities bleiben CRUD.

### 8.2 Synchrone Projektionen überall

**Problem:** Alle Read Models synchron aktualisieren — auch nicht-kritische.

```
// Anti-Pattern: Tagesstatistiken synchron im Command aktualisieren
func (s *Service) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error {
    s.eventRepo.WriteEvent(ctx, event)
    s.readRepo.UpdateOrderSummary(ctx, ...)        // kritisch ✅
    s.statsRepo.UpdateDailyRevenue(ctx, ...)        // unkritisch ❌ → verlangsamt Write
    s.analyticsRepo.UpdateCustomerLifetimeValue(ctx, ...) // unkritisch ❌
    return nil
}
```

**Lösung:** Nur kritische Read Models synchron. Alles andere asynchron via Background Worker oder LISTEN/NOTIFY.

### 8.3 Commands als Queries verwenden

**Problem:** Ein Command gibt komplexe Daten zurück — vermischt Command und Query.

```go
// Anti-Pattern:
func (h *Handler) PlaceOrder(ctx, cmd) (OrderDetailDTO, error) {
    // Schreibt AND gibt komplexes Read Model zurück
}
```

**Lösung:** Command gibt minimal zurück (ID, Erfolgsmeldung). Client fragt danach separat mit Query.

### 8.4 Projektion nicht idempotent

**Problem:** Bei asynchroner Projektion kann dieselbe Event-Nachricht mehrfach ankommen (at-least-once delivery). Wenn die Projektion nicht idempotent ist, entstehen Duplikate.

```go
// Anti-Pattern: Nicht idempotente Projektion
func (p *Projektor) Handle(event Event) {
    p.db.Exec("UPDATE stats SET count = count + 1 WHERE day = $1", event.Day)
    // Bei doppelter Verarbeitung: count wird zweimal erhöht ❌
}

// Korrekt: Idempotente Projektion
func (p *Projektor) Handle(event Event) {
    p.db.Exec(`
        INSERT INTO stats (day, count) VALUES ($1, 1)
        ON CONFLICT (day) DO UPDATE SET count = EXCLUDED.count + 1
        WHERE NOT EXISTS (SELECT 1 FROM processed_events WHERE event_id = $2)
    `, event.Day, event.ID)
}
```

**Lösung:** Events mit ihrer ID als processed markieren. `ON CONFLICT DO NOTHING` für einfache Fälle.

### 8.5 Read Model als einzige Source of Truth

**Problem:** Das Read Model wird als primäre Wahrheitsquelle behandelt. Der Event Store oder das Write Model wird vernachlässigt.

**Folge:** Bei einem Bug in der Projektion gehen Daten verloren — kein Rebuild möglich.

**Lösung:** Events (oder das normalisierte Write Model) sind immer die Source of Truth. Das Read Model ist sekundär und jederzeit neu aufbaubar.

---

## 9. Referenzen

### Primärquellen

- **Bertrand Meyer** (1988): _Object-Oriented Software Construction_ — CQS-Prinzip: "A method should either change state (Command) or return data (Query) — never both"
- **Greg Young** (2010): [CQRS Documents](https://cqrs.wordpress.com/) — Ursprung von CQRS + Abgrenzung zu ES
- **Martin Fowler**: [CQRS](https://martinfowler.com/bliki/CQRS.html) — Wann CQRS sinnvoll ist, Warnungen vor Over-Engineering
- **Udi Dahan**: [Clarified CQRS](https://udidahan.com/2009/12/09/clarified-cqrs/) — CQRS + DDD, eigenständiger Wert von CQRS, Query-Datenstrukturen

### Architektur & Cloud

- **AWS Prescriptive Guidance**: [CQRS Pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html) — Cloud-native CQRS, DynamoDB+Aurora, Read Replicas
- **Chris Richardson**: [CQRS Pattern (microservices.io)](https://microservices.io/patterns/data/cqrs.html) — CQRS als Microservice-Pattern, View Databases
- **Microsoft**: [CQRS Pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/cqrs) — Azure-Architektur-Guide zu CQRS

### Praxis & Projektionen

- **Oskar Dudycz**: [CQRS Facts and Myths Explained](https://event-driven.io/en/cqrs_facts_and_myths_explained/) — Mythen über CQRS (zwei Datenbanken, Eventual Consistency, ES-Zwang)
- **Oskar Dudycz**: [Projections and Read Models in Event-Driven Architecture](https://event-driven.io/en/projections_and_read_models_in_event_driven_architecture/) — Projektions-Rebuild, Synchron vs. Asynchron, Idempotenz
- **Oskar Dudycz**: [Outbox, Inbox patterns and delivery guarantees explained](https://event-driven.io/en/outbox_inbox_patterns_and_delivery_guarantees_explained/) — Transactional Outbox, At-Least-Once, Idempotenz
- **Kamil Grzybek**: [Modular Monolith with DDD](https://github.com/kgrzybek/modular-monolith-with-ddd) — Outbox, Inbox, CQRS in einem Monolithen (C# Referenzimplementierung)

### Go-spezifisch

- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — CQRS in Go-Microservices mit DDD und Event-Sourcing

### Bücher

- **Martin Kleppmann** (2017): _Designing Data-Intensive Applications_ — Consistency Models, Derived Data, Stream Processing als Read Model
- **Vaughn Vernon** (2013): _Implementing Domain-Driven Design_ — CQRS + DDD, Command/Query-Trennung auf Aggregate-Ebene

### Cross-Referenzen

- [Event-Sourcing Theorie](event-sourcing.md) — Event-Sourcing Grundlagen, Snapshots, Outbox
