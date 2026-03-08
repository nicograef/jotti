# CQRS — Theorie

Dieses Dokument dient als theoretisches Nachschlagewerk für Command Query Responsibility Segregation (CQRS). Es erklärt das Muster, seine Ausbaustufen, die Kombination mit Event-Sourcing und Projektionsstrategien. Ein projektspezifisches Anwendungsbeispiel findet sich im [Appendix](#appendix-anwendungsbeispiel-jotti).

> **Verwandte Dokumente:**
>
> - [Event-Sourcing Theorie](event-sourcing.md) — Event-Sourcing Grundlagen
> - [DDD Theorie](ddd.md) — Domain-Driven Design Grundlagen
> - [CQRS in jotti (operativ)](../cqrs.md) — Ist-Zustand und Implementierungsplan
> - [Architektur-Übersicht](README.md) — Index aller Theorie-Dokumente

---

## Inhaltsverzeichnis

1. [CQRS — Grundlagen](#1-cqrs--grundlagen)
2. [Event-Sourcing + CQRS: Die Kombination](#2-event-sourcing--cqrs-die-kombination)
3. [Appendix: Anwendungsbeispiel (jotti)](#appendix-anwendungsbeispiel-jotti)
4. [Referenzen](#4-referenzen)

---

## 1. CQRS — Grundlagen

### 1.1 Grundidee

**Command Query Responsibility Segregation** trennt die Verantwortlichkeit für Schreiboperationen (Commands) und Leseoperationen (Queries) auf System-Ebene. CQRS erweitert das CQS-Prinzip (Command Query Separation) von Bertrand Meyer:

> **CQS (Methoden-Ebene):** Eine Methode soll entweder den Zustand ändern (Command) ODER Daten zurückgeben (Query) — nie beides.
>
> **CQRS (System-Ebene):** Das System hat **separate Modelle** für Schreiben und Lesen.

### 1.2 Command Side

Commands drücken eine **Absicht** aus: „Gib diese Bestellung auf", „Registriere diese Zahlung". Commands:

- Ändern den Zustand
- Geben maximal eine ID oder Erfolgsmeldung zurück
- Werden validiert
- Können abgelehnt werden (z.B. ungültige Daten)
- Sind idempotent bei wiederholter Ausführung (idealerweise)

```go
// Command: Absichtserklärung
type PlaceOrderCommand struct {
    TableID  int
    Items    []OrderItem
    Comment  string
}

// Command Handler: Verarbeitung
func (h *Handler) PlaceOrder(ctx context.Context, cmd PlaceOrderCommand) error {
    // 1. Validieren
    // 2. Business Rules prüfen
    // 3. Event erzeugen und speichern
    return nil
}
```

### 1.3 Query Side

Queries fragen Daten ab, **ohne den Zustand zu ändern**. Queries:

- Haben keine Seiteneffekte
- Geben Daten zurück (Read Model / DTO)
- Können gecacht werden
- Können gegen optimierte Datenstrukturen arbeiten

```go
// Query: Datenanfrage
type GetOrderBalanceQuery struct {
    OrderID int
}

// Query Handler: Daten lesen
func (h *Handler) GetOrderBalance(ctx context.Context, q GetOrderBalanceQuery) (int, error) {
    // Optimiertes Read Model abfragen
    return balanceCents, nil
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

Separate Command- und Query-Klassen/Handler, aber **gleiche Datenbank** und gleiches Modell.

```
Client → Command Handler → Repository → Datenbank ← Repository ← Query Handler ← Client
```

#### Stufe 2: Getrennte Modelle (Projektionen)

Command Side schreibt in Write Store, eine **Projektion** synchronisiert Daten in ein Read Model.

```
Client → Command Handler → Write Store (Events)
                               │
                               ▼ Projektion (synchron/asynchron)
                               │
Client ← Query Handler  ← Read Store (Projektions-Tabellen)
```

#### Stufe 3: Getrennte Datenbanken

Write Store und Read Store in **separaten Datenbank-Instanzen** (oder DB-Technologien).

```
Command Handler → PostgreSQL (Events)
                       │
                       ▼ Async Projektion (z.B. via LISTEN/NOTIFY)
                       │
Query Handler   ← Redis / Materialized View / andere DB
```

### 1.5 Vor- und Nachteile von CQRS

| Vorteil                                 | Beschreibung                                             |
| --------------------------------------- | -------------------------------------------------------- |
| **Optimierte Lesemodelle**              | Read Models können exakt auf Abfragen zugeschnitten sein |
| **Unabhängige Skalierung**              | Lese- und Schreiblast getrennt skalierbar                |
| **Vereinfachte Modelle**                | Jedes Modell tut nur eine Sache                          |
| **Natürliche Event-Sourcing-Ergänzung** | Events = Write Model, Projektionen = Read Model          |

| Nachteil                          | Beschreibung                                 |
| --------------------------------- | -------------------------------------------- |
| **Eventual Consistency**          | Read Model kann kurzzeitig veraltet sein     |
| **Mehr Code**                     | Separate Handler, Modelle, Projektionen      |
| **Synchronisationskomplexität**   | Projektion muss zuverlässig synchron bleiben |
| **Overkill für einfache Domains** | CRUD-Entities profitieren nicht von CQRS     |

---

## 2. Event-Sourcing + CQRS: Die Kombination

### 2.1 Warum sie zusammengehören

Event-Sourcing allein hat ein Performance-Problem beim Lesen: Jede Query muss Events replayed werden. CQRS löst dieses Problem durch **separate Read Models (Projektionen)**:

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

### 2.2 Projektionsstrategien

#### Synchrone Projektion

Das Read Model wird **im selben Request** wie das Event aktualisiert:

```go
func (s *CommandService) PlaceOrder(ctx context.Context, orderID int, items []Item) error {
    // 1. Event schreiben
    eventRepo.WriteEvent(ctx, event)
    // 2. Read Model aktualisieren (gleiche Transaktion)
    readRepo.UpdateBalance(ctx, orderID, deltaCents)
    return nil
}
```

**Vorteile:** Starke Konsistenz, kein Sync-Problem
**Nachteile:** Write-Latenz steigt, enge Kopplung

#### Asynchrone Projektion

Das Read Model wird **nach dem Request** aktualisiert (via Polling, Trigger oder Message Queue):

```
Event INSERT → LISTEN/NOTIFY → Projektor → Read Model UPDATE
```

**Vorteile:** Write-Latenz bleibt niedrig, lose Kopplung
**Nachteile:** Eventual Consistency, komplexeres Error Handling

#### Hybride Projektion

Kritische Read Models synchron, unkritische asynchron:

```
Bestellung → Event INSERT
              ├── Synchron:  Kontostand aktualisieren (kritisch)
              └── Asynchron: Tagesstatistik aktualisieren (unkritisch)
```

### 2.3 Konsistenzanforderungen nach Feature

Nicht alle Read Models brauchen starke Konsistenz:

| Feature                  | Konsistenz           | Begründung                                               |
| ------------------------ | -------------------- | -------------------------------------------------------- |
| **Kontostand/Saldo**     | Stark (synchron)     | Fehlerhafte Anzeige führt zu Fehlbuchungen               |
| **Offene Positionen**    | Stark (synchron)     | Abrechnungsvorgang muss korrekt sein                     |
| **Lieferstatus**         | Stark (synchron)     | Lieferung muss vollständig sein                          |
| **Transaktions-Historie**| Eventual (asynchron) | Historische Ansicht verträgt kurze Verzögerung           |
| **Tagesabrechnung**      | Eventual (asynchron) | Wird nicht in Echtzeit benötigt                          |
| **Umsatzstatistiken**    | Eventual (asynchron) | Aggregierte Daten, keine Echtzeitanforderung             |

> **Hinweis:** CQRS ist auch ohne Event-Sourcing einsetzbar — mit klassischen CRUD-Datenbanken. Doch die Kombination ist besonders synergetisch, da Events das natürliche Write Model bilden und Projektionen das Leseperformance-Problem lösen. Siehe [Event-Sourcing — Theorie](event-sourcing.md) für die Event-Sourcing-Grundlagen.

---

## Appendix: Anwendungsbeispiel (jotti)

Dieser Abschnitt zeigt, wie CQRS konkret in jotti — einem Non-Profit-POS-System für Vereinsfeste — eingesetzt wird.

### Ist-Zustand: Stufe 1 (Logische Trennung)

jotti implementiert CQRS auf **Stufe 1** — logische Trennung in Command und Query Handler:

```
api/table/application/
├── command.go    // BestellungAufgeben, ZahlungRegistrieren, ...
└── query.go      // GetTischSaldo, GetTischUnbezahlt, ...

api/table/http/
├── command.go    // HTTP Handler für Commands
└── query.go      // HTTP Handler für Queries
```

**Was funktioniert:** Commands und Queries sind klar getrennt. Commands schreiben Events. Queries replayed Events (mit Snapshot-Optimierung).

**Was fehlt:** Es gibt kein separates Read Model. Queries müssen immer den Event-Stream replayed — selbst mit Snapshots ist das aufwändiger als ein einfacher `SELECT`.

### Nächste Stufe: Stufe 2 (Synchrone Projektion)

Empfohlen für jotti. Details siehe [CQRS in jotti (operativ)](../cqrs.md).

**Konzept:** Bei jedem Command wird zusätzlich zum Event ein Read Model (Projektions-Tabelle) aktualisiert:

```sql
CREATE TABLE tisch_zustand (
    tisch_id                  INT PRIMARY KEY REFERENCES tables(id),
    saldo_cents               INT NOT NULL DEFAULT 0,
    gesamt_zahlungen_cents    INT NOT NULL DEFAULT 0,
    unbezahlte_positionen     JSONB NOT NULL DEFAULT '[]',
    ungelieferte_positionen   JSONB NOT NULL DEFAULT '[]',
    last_event_id             INT REFERENCES events(id),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Vorteile:**

- Queries werden zu einfachen `SELECT`-Statements
- Keine Snapshot-Logik in Go nötig
- Event-Stream bleibt Single Source of Truth
- Bei Inkonsistenz: Read Model aus Events neu aufbauen

### Konsistenzanforderungen in jotti

| Feature                     | Konsistenz           | Begründung                                       |
| --------------------------- | -------------------- | ------------------------------------------------ |
| **Saldo**                   | Stark (synchron)     | Fehlerhafte Saldo-Anzeige führt zu Fehlzahlungen |
| **Unbezahlte Positionen**   | Stark (synchron)     | Kassiervorgang muss korrekt sein                 |
| **Ungelieferte Positionen** | Stark (synchron)     | Lieferung muss vollständig sein                  |
| **Tisch-Historie**          | Eventual (asynchron) | Historische Ansicht verträgt kurze Verzögerung   |
| **Tagesabrechnung**         | Eventual (asynchron) | Wird nicht in Echtzeit benötigt                  |
| **Umsatzstatistiken**       | Eventual (asynchron) | Aggregierte Daten, keine Echtzeitanforderung     |

---

## 4. Referenzen

### Primärquellen

- **Greg Young** (2010): _CQRS Documents_ — cqrs.wordpress.com — Ursprung von CQRS + Event-Sourcing
- **Martin Fowler**: [CQRS](https://martinfowler.com/bliki/CQRS.html)
- **Udi Dahan**: [Clarified CQRS](https://udidahan.com/2009/12/09/clarified-cqrs/) — CQRS + DDD
- **Bertrand Meyer** (1988): _Object-Oriented Software Construction_ — CQS-Prinzip
- **AWS Prescriptive Guidance**: [CQRS Pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html)

### Praxisquellen

- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — Event-Sourcing + CQRS in Go-Microservices

### Projekt-intern

- [Event-Sourcing Theorie](event-sourcing.md) — Event-Sourcing Grundlagen
- [CQRS in jotti (operativ)](../cqrs.md) — Detaillierter Implementierungsplan
