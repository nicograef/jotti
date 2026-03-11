# Event-Sourcing — Theorie

Dieses Dokument dient als theoretisches Nachschlagewerk für Event-Sourcing. Es erklärt das Muster, seine Kernkonzepte, Technologieoptionen, fortgeschrittene Patterns und Entscheidungskriterien gegenüber CRUD.

---

## Inhaltsverzeichnis

1. [Grundidee & Paradigmenwechsel](#1-grundidee--paradigmenwechsel)
2. [Event Store: Konzept und Technologien](#2-event-store-konzept-und-technologien)
3. [Event Design](#3-event-design)
4. [State Reconstruction (Replay)](#4-state-reconstruction-replay)
5. [Snapshots](#5-snapshots)
6. [Event-Schema-Evolution](#6-event-schema-evolution)
7. [Fortgeschrittene Patterns](#7-fortgeschrittene-patterns)
8. [Event-Sourcing vs. CRUD: Entscheidungsmatrix](#8-event-sourcing-vs-crud-entscheidungsmatrix)
9. [Reale Fallstudien](#9-reale-fallstudien)
10. [Kombination mit CQRS](#10-kombination-mit-cqrs)
11. [Anti-Patterns](#11-anti-patterns)
12. [Referenzen](#12-referenzen)

---

## 1. Grundidee & Paradigmenwechsel

### 1.1 Von Snapshot zu Narration

Event-Sourcing ist ein Persistenzmuster, bei dem **nicht der aktuelle Zustand**, sondern die **Folge aller Zustandsänderungen** (Events) gespeichert wird. Der aktuelle Zustand wird durch Abspielen (Replay) aller Events rekonstruiert.

> **Greg Young:** _"Instead of storing just the current state of the data in a domain, use an append-only store to record the full series of actions taken on that data."_
>
> **Martin Fowler:** _"Capture all changes to an application state as a sequence of events."_

**Analogie Buchführung (Accounting Ledger):** Ein Buchhalter erfasst alle Buchungen als unveränderliche Einträge im Hauptbuch — niemals werden Einträge gelöscht, Korrekturen sind neue Gegenbuchungen. Der Kontostand ist jederzeit aus den Buchungen rekonstruierbar. Traditionelle Datenbanken gleichen dagegen einer Tafel, die immer wieder überschrieben wird: Der aktuelle Stand ist sichtbar, die Geschichte ist verloren.

| Paradigma                 | Speicherinhalt                     | Frage                             |
| ------------------------- | ---------------------------------- | --------------------------------- |
| **CRUD (State-Oriented)** | Aktueller Zustand (überschreibend) | _Wie ist der Zustand jetzt?_      |
| **Event-Sourcing**        | Vollständige Ereignishistorie      | _Wie ist der Zustand entstanden?_ |

### 1.2 Paradigmenwechsel im Schreiben

Event-Sourcing transformiert den kritischsten Datenbankvorgang: Ein potenziell komplexes `UPDATE`, das Locks erfordert und Konflikte erzeugt, wird zu einem einfachen, nicht-destruktiven `INSERT` in ein Append-only-Log. Das reduziert Write-Contention dramatisch und vereinfacht die Transaktionslogik auf der Schreibseite — auf Kosten erhöhter Komplexität auf der Leseseite.

---

## 2. Event Store: Konzept und Technologien

### 2.1 Konzept

Der **Event Store** ist das Herzstück eines event-gesourcten Systems. Er ist ein Append-only-Log mit drei fundamentalen Garantien:

1. **Immutability** — Einmal geschriebene Events können nicht geändert oder gelöscht werden
2. **Ordering** — Events innerhalb eines Streams sind streng chronologisch geordnet
3. **Append-Only** — Neue Einträge werden nur angefügt, nie überschrieben

```
Event Store — Stream "order:101":
┌────┬────────────────────────────┬──────────────────────────────┐
│ ID │ Type                       │ Data (JSONB)                 │
├────┼────────────────────────────┼──────────────────────────────┤
│  1 │ order.placed:v1            │ {items: [...], total: 4200}  │
│  2 │ order.payment-received:v1  │ {amount: 4200}               │
│  3 │ order.items-shipped:v1     │ {items: [...], tracking: "1Z999AA1"}│
└────┴────────────────────────────┴──────────────────────────────┘
```

### 2.2 Event Store-Technologien im Vergleich

| Technologie       | Typ                        | Stärken                                                                   | Schwächen                                                  | Geeignet für                                     |
| ----------------- | -------------------------- | ------------------------------------------------------------------------- | ---------------------------------------------------------- | ------------------------------------------------ |
| **EventStoreDB**  | Spezialisierter ES-Store   | Native Streams, Subscriptions, Projektions-Engine, Optimistic Concurrency | Eigenes Betriebsmodell, Community Edition limitiert        | Dedizierte ES-Systeme, Microservices             |
| **PostgreSQL**    | Relationale DB (generisch) | Bekannt, ACID, JSONB, Trigger für Immutability, keine neue Infrastruktur  | Keine native Stream-Abstraktion, manuelle Partitions       | Monolithen, Teams mit SQL-Expertise              |
| **Apache Kafka**  | Distributed Log / Broker   | Extreme Skalierbarkeit, Retention, Consumer Groups, Replay                | Kein primärer ES-Store, keine Aggregate-Granularität       | Event Streaming, Microservices mit hohem Volumen |
| **DynamoDB**      | NoSQL (AWS)                | Serverless, skalierbar, niedrige Latenz bei einfachen Zugriffen           | Kein natürliches Ordering über Partitionen, Vendor Lock-in | AWS-native Systeme, globale Skalierung           |
| **Marten (.NET)** | PostgreSQL-Wrapper         | PostgreSQL-basiert, LINQ-Queries, Document Store + ES kombiniert          | .NET-only, kein Go-Support                                 | .NET-Systeme auf PostgreSQL                      |
| **Axon (Java)**   | ES-Framework (Java)        | Vollständiges ES+CQRS+Saga-Framework, Command Bus, Event Bus              | Stark opinionated, Java-only, Lizenzkosten möglich         | Java/Spring-Enterprise-Anwendungen               |

**Entscheidungsregel:** Für den Einstieg und Teams ohne ES-Infrastruktur ist **PostgreSQL als Event Store** der pragmatischste Ansatz. EventStoreDB lohnt sich ab dem Punkt, wo die erweiterten Features (persistente Subscriptions, Server-side Projektionen) den Betriebsaufwand rechtfertigen. Kafka ist kein primärer Event Store, sondern ein Event-Transport — eignet sich aber als sekundärer Kanal für Integration Events.

---

## 3. Event Design

### 3.1 Thin Events vs. Fat Events

Ein fundamentales Designproblem ist die Granularität des Event-Inhalts:

| Typ            | Inhalt                                          | Vorteile                             | Nachteile                                      |
| -------------- | ----------------------------------------------- | ------------------------------------ | ---------------------------------------------- |
| **Thin Event** | Nur das Notwendigste (z.B. IDs, Schlüsselwerte) | Kleiner, weniger Kopplung            | Consumer müssen Daten nachladen (Join-Problem) |
| **Fat Event**  | Alle relevanten Kontextdaten eingebettet        | Self-contained, kein Nachladen nötig | Größer, Redundanz bei Duplikaten               |

**Empfehlung für Domain Events:** Fat Events — alle relevanten Daten (Name, Preis, Status) werden zum Zeitpunkt der Aktion eingebettet. Bei **Integration Events** (systemübergreifend) kann Thin sinnvoll sein, wenn der Consumer sowieso die aktuelle Version braucht.

```go
// Fat Event (RICHTIG für Domain Events)
type OrderPlacedEvent struct {
    OrderID    int
    CustomerID int
    Items      []OrderItem  // Name + Preis eingebettet
    TotalCents int
    PlacedAt   time.Time
}

type OrderItem struct {
    ProductID  int
    Name       string  // "Espresso" — zum Zeitpunkt der Bestellung
    PriceCents int     // 350 — zum Zeitpunkt der Bestellung
    Quantity   int
}

// Thin Event (problematisch: Was wenn Preis sich ändert?)
type OrderPlacedEventThin struct {
    OrderID    int
    ProductIDs []int  // Consumer muss Preise separat nachladen
}
```

### 3.2 Domain Events vs. Integration Events

| Typ                   | Scope                           | Lebensdauer | Consumer                          | Beispiel                         |
| --------------------- | ------------------------------- | ----------- | --------------------------------- | -------------------------------- |
| **Domain Event**      | Innerhalb eines Bounded Context | Dauerhaft   | Gleicher BC, interne Projektionen | `OrderPlaced`, `PaymentReceived` |
| **Integration Event** | Systemübergreifend              | Kurzlebig   | Andere Services/BCs               | `OrderConfirmed` (an Logistics)  |

Domain Events sind die Grundlage des Event Stores. Integration Events werden typischerweise aus Domain Events abgeleitet und über einen Message Broker veröffentlicht — über das **Outbox Pattern** (→ Abschnitt 7.2).

### 3.3 Eigenschaften eines guten Events

| Eigenschaft            | Beschreibung                                                          |
| ---------------------- | --------------------------------------------------------------------- |
| **Vergangenheitsform** | `OrderPlaced`, nicht `PlaceOrder`                                     |
| **Self-contained**     | Alle Daten im Event, keine Referenzen auf veränderliche Stammdaten    |
| **Versioniert**        | Schema-Evolution durch Versionsnummer (`:v1`, `:v2`)                  |
| **Geordnet**           | Sequenzielle ID oder Zeitstempel für kausale Ordnung                  |
| **Granular**           | Nur die relevante Zustandsänderung, kein gesamter Aggregat-Zustand    |
| **Fachlich sinnvoll**  | Beschreibt Geschäftsereignis, kein technisches Implementierungsdetail |

---

## 4. State Reconstruction (Replay)

### 4.1 Das Apply/Fold-Pattern

Zustandsrekonstruktion folgt dem funktionalen **Fold/Reduce-Pattern**: Ein leerer Anfangszustand wird schrittweise durch Anwendung jedes Events transformiert.

```
State₀ = Anfangszustand (leer)
State₁ = apply(State₀, Event₁)
State₂ = apply(State₁, Event₂)
...
Stateₙ = fold(State₀, [Event₁, Event₂, ..., Eventₙ])
```

**Go-Implementierung:**

```go
// Apply-Funktion: reine Funktion ohne Seiteneffekte
// Fehlerbehandlung für json.Unmarshal aus Kürze weggelassen
func Apply(state OrderState, event Event) OrderState {
    switch event.Type {
    case "order.placed:v1":
        var data OrderPlacedData
        _ = json.Unmarshal(event.Data, &data)
        state.Items = data.Items
        state.TotalCents = data.TotalCents
        state.Status = "pending"
    case "order.payment-received:v1":
        state.Status = "paid"
    case "order.items-shipped:v1":
        state.Status = "shipped"
    }
    return state
}

// Fold: Aggregation aller Events
func Rehydrate(events []Event) OrderState {
    state := OrderState{}  // leerer Anfangszustand
    for _, e := range events {
        state = Apply(state, e)
    }
    return state
}
```

**Schlüsselprinzip:** Die `Apply`-Funktion ist eine **reine Funktion** — kein I/O, keine Seiteneffekte. Das macht sie einfach testbar und deterministisch.

### 4.2 Commands vs. Events

```
Command (Absicht):    PlaceOrder → Validierung → Ablehnen (falls ungültig)
                                                ↓ (falls gültig)
Event (Tatsache):                         OrderPlaced → Event Store
```

Commands können abgelehnt werden. Events sind unveränderliche Tatsachen und werden nie abgelehnt.

---

## 5. Snapshots

### 5.1 Notwendigkeit

Bei langen Event-Streams wird das Replay teuer. Ein Aggregat mit 10.000 Events muss bei jeder Operation alle 10.000 Events laden und verarbeiten.

```
Ohne Snapshot:  [1] [2] [3] ... [9999] [10000]  → 10.000 Events laden
Mit Snapshot:   [SNAPSHOT@9990] [9991] ... [10000]  → Snapshot + 10 Events
```

### 5.2 Snapshot-Strategien

| Strategie              | Beschreibung                                           | Geeignet für                                     |
| ---------------------- | ------------------------------------------------------ | ------------------------------------------------ |
| **N-Events-Strategie** | Snapshot nach jedem N-ten Event (z.B. alle 100 Events) | Stabile, vorhersehbare Aggregat-Größen           |
| **Zeitbasiert**        | Snapshot in festen Zeitintervallen (z.B. täglich)      | Langlebige Aggregates mit regelmäßiger Aktivität |
| **Bei jedem Write**    | Snapshot nach jeder Schreiboperation                   | Häufig gelesene, selten geschriebene Aggregates  |
| **On-Demand**          | Snapshot manuell oder bei Bedarf (z.B. nach Migration) | Einmalige Optimierungen                          |

### 5.3 Snapshot als Event vs. separater Store

**Option A — Snapshot als eigener Event-Typ:**

- Snapshot wird als normales Event in den Event Store geschrieben
- Keine separate Infrastruktur nötig
- Nachvollziehbar in der Event-Historie

**Option B — Separater Snapshot-Store:**

- Snapshots in eigener Tabelle/Collection
- Event Store bleibt „sauber" mit nur Domain Events
- Flexiblere Snapshot-Strategie ohne Stream-Pollution

---

## 6. Event-Schema-Evolution

Das schwierigste Problem in event-gesourcten Systemen: Events sind immutable und können jahrelang im Store liegen. Business-Anforderungen ändern sich — Event-Schemas müssen sich anpassen.

### 6.1 Strategien im Überblick

| Strategie                 | Beschreibung                                                           | Aufwand   | Eignung                               |
| ------------------------- | ---------------------------------------------------------------------- | --------- | ------------------------------------- |
| **Weak Schema**           | Optionale Felder, Defaults für fehlende Werte                          | Gering    | Kleine, additive Änderungen           |
| **Upcasting**             | Alte Events werden beim Lesen on-the-fly in neuere Version konvertiert | Mittel    | Nicht-breaking Schema-Änderungen      |
| **Event Versioning**      | Neue Versionen parallel (`:v1`, `:v2`), Handler für beide              | Mittel    | Breaking Changes mit Migration        |
| **Lazy Migration**        | Events werden beim Lesen migriert und zurückgeschrieben                | Hoch      | Schrittweise Migration langer Streams |
| **Stream Transformation** | Komplette Neuschreibung aller Events in neues Format                   | Sehr hoch | Fundamental andere Datenstruktur      |

### 6.2 Upcasting (empfohlen für die meisten Fälle)

Beim Upcasting transformiert eine Upcaster-Funktion ein altes Event-Format beim Lesen in das neue Format — ohne die gespeicherten Events zu ändern.

```go
// Event-Versionierung: v1 → v2
// v1: kein "comment"-Feld
// v2: "comment"-Feld hinzugefügt

func upcastOrderPlaced(event Event) Event {
    if event.Type == "order.placed:v1" {
        var data map[string]interface{}
        _ = json.Unmarshal(event.Data, &data) // Fehlerbehandlung aus Kürze weggelassen

        // Fehlendes Feld mit Default befüllen
        if _, ok := data["comment"]; !ok {
            data["comment"] = ""
        }

        newData, _ := json.Marshal(data) // Fehlerbehandlung aus Kürze weggelassen
        return Event{
            ID:        event.ID,
            Type:      "order.placed:v2",  // hochgestuft
            Subject:   event.Subject,
            Timestamp: event.Timestamp,
            Data:      newData,
        }
    }
    return event
}

// Upcaster in der Read-Pipeline registrieren
func ReadAndUpcast(events []Event) []Event {
    result := make([]Event, len(events))
    for i, e := range events {
        result[i] = upcastOrderPlaced(e)
    }
    return result
}
```

**Grundregel:** Alte Events niemals direkt ändern. Upcaster laufen beim Lesen — transparent für den Rest der Anwendung.

### 6.3 Backward/Forward Compatibility

| Kompatibilität          | Beschreibung                                        | Mittel                                |
| ----------------------- | --------------------------------------------------- | ------------------------------------- |
| **Backward Compatible** | Neuer Code kann alte Events verarbeiten             | Optionale Felder, Defaults, Upcasting |
| **Forward Compatible**  | Alter Code kann neue Events ignorieren/überspringen | Unbekannte Felder ignorieren          |

**Schema Registry:** In verteilten Systemen mit vielen Services und Event-Typen verwaltet eine **Schema Registry** (z.B. Confluent Schema Registry für Avro/Protobuf) die Schemas zentral und erzwingt Kompatibilitätschecks vor dem Deployment.

### 6.4 Serialisierungsformate

| Format       | Typ        | Vorteile                                      | Nachteile                               |
| ------------ | ---------- | --------------------------------------------- | --------------------------------------- |
| **JSON**     | Text       | Lesbar, flexibel, weit verbreitet             | Kein eingebautes Schema, größer         |
| **JSONB**    | Binär-JSON | JSON-Features + PostgreSQL-native Indexierung | PostgreSQL-spezifisch                   |
| **Protobuf** | Binär      | Kompakt, typsicher, Schema-Evolution          | Binär (nicht direkt lesbar), Build-Step |
| **Avro**     | Binär      | Schema Registry, gute Kafka-Integration       | Komplexer Setup                         |

---

## 7. Fortgeschrittene Patterns

### 7.1 Saga Pattern (Process Manager)

**Problem:** Komplexe Geschäftsvorgänge umfassen mehrere Aggregates oder Services. Eine atomare Transaktion über Servicegrenzen ist mit 2PC nicht praktikabel.

**Lösung:** Eine **Saga** zerlegt den Vorgang in eine Sequenz lokaler Transaktionen. Jede lokale Transaktion veröffentlicht ein Event, das den nächsten Schritt auslöst. Bei einem Fehler werden **Kompensations-Transaktionen** ausgeführt.

```
Saga: Bestellung aufgeben (Order → Payment → Inventory)

1. OrderService:      Order.Placed  →  Event: OrderCreated
2. PaymentService:    Payment.Reserved  →  Event: PaymentReserved
3. InventoryService:  Stock.Reserved  →  Event: StockReserved
4. OrderService:      Order.Confirmed  →  Event: OrderConfirmed

Bei Fehler in Schritt 3 (kein Lager):
3b. InventoryService:   →  Event: StockReservationFailed
4b. PaymentService:     Kompensation: Payment.Released
5b. OrderService:       Kompensation: Order.Cancelled
```

**Choreography vs. Orchestration:**

| Ansatz            | Beschreibung                                                   | Vorteile                            | Nachteile                           |
| ----------------- | -------------------------------------------------------------- | ----------------------------------- | ----------------------------------- |
| **Choreography**  | Jeder Service reagiert auf Events anderer Services             | Kein zentraler Coordinator          | Schwer zu debuggen, impliziter Flow |
| **Orchestration** | Zentraler Process Manager steuert den Ablauf (Commands senden) | Expliziter Flow, leicht zu debuggen | Single Point of Failure möglich     |

### 7.2 Outbox Pattern

**Problem:** Beim Verarbeiten eines Commands muss der Service zwei Aktionen atomisch ausführen:

1. Datenbank aktualisieren (Event speichern)
2. Event an Message Broker senden

Schlägt eine der beiden Aktionen fehl, entsteht Inkonsistenz.

**Lösung:** Das **Outbox Pattern** speichert das zu sendende Event als Teil derselben DB-Transaktion in einer `outbox`-Tabelle. Ein separater Prozess (Message Relay) liest die Outbox und sendet die Events an den Broker.

```
Transaktion (atomar):
  INSERT INTO events (type, subject, data) VALUES (...)
  INSERT INTO outbox (event_id, destination, payload) VALUES (...)

Message Relay (asynchron):
  SELECT * FROM outbox WHERE sent_at IS NULL
  → Events an Message Broker senden
  → UPDATE outbox SET sent_at = now() WHERE id = ...
```

```
┌───────────────────────────────────────────────┐
│ Datenbank                                     │
│  events (Event Store)                         │
│  outbox (zu sendende Nachrichten)             │
└───────────────────────────┬───────────────────┘
                            │ Message Relay (Polling/CDC)
                            ▼
                    ┌──────────────┐
                    │ Message Broker│
                    │ (Kafka, NATS) │
                    └──────────────┘
```

**Implementierungsvarianten:**

- **Polling Publisher:** Relay fragt regelmäßig die Outbox-Tabelle ab
- **Transaction Log Tailing / CDC:** Relay liest das Datenbank-WAL (Write-Ahead Log) via Debezium oder ähnlichen Tools

### 7.3 Inbox Pattern

**Problem:** Message Broker liefern Nachrichten mindestens einmal (**at-least-once delivery**). Event-Handler müssen idempotent sein, oder doppelte Events führen zu Fehlverhalten.

**Lösung:** Das **Inbox Pattern** verwaltet eine `inbox`-Tabelle, die bereits verarbeitete Event-IDs speichert. Vor der Verarbeitung wird geprüft, ob das Event bereits verarbeitet wurde.

```sql
-- Inbox-Tabelle
CREATE TABLE inbox (
    message_id  TEXT PRIMARY KEY,  -- eindeutige ID vom Broker
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Idempotente Verarbeitung
INSERT INTO inbox (message_id)
VALUES ($1)
ON CONFLICT (message_id) DO NOTHING
RETURNING message_id;
-- Wenn kein Row zurückgegeben: Event bereits verarbeitet → überspringen
```

### 7.4 Idempotenz & Concurrency Control

**Idempotenz-Strategien:**

| Strategie               | Beschreibung                                             |
| ----------------------- | -------------------------------------------------------- |
| **Idempotency Key**     | Client sendet UUID im Request; Server prüft auf Duplikat |
| **Natural Idempotency** | „Setze Status auf X" statt „Erhöhe Status um 1"          |
| **Inbox Pattern**       | Message-ID in Inbox-Tabelle tracken (→ 7.3)              |

**Concurrency Control:**

| Strategie                  | Beschreibung                                                                  |
| -------------------------- | ----------------------------------------------------------------------------- |
| **Optimistic Concurrency** | Expected Version beim Schreiben prüfen; Conflict → Retry                      |
| **Pessimistic Locking**    | `SELECT ... FOR UPDATE` auf Aggregat-Ebene vor dem Schreiben                  |
| **Last-Writer-Wins**       | Letzter Write gewinnt (akzeptabel bei unabhängigen Events am selben Aggregat) |

```go
// Optimistic Concurrency: Erwartete Version mitschicken
func WriteEvent(subject string, expectedVersion int, event Event) error {
    result, err := db.Exec(`
        INSERT INTO events (subject, type, data, version)
        SELECT $1, $2, $3, $4
        WHERE (SELECT MAX(version) FROM events WHERE subject = $1) = $4 - 1
    `, subject, event.Type, event.Data, expectedVersion+1)
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return ErrConcurrencyConflict  // Version hat sich geändert → Retry
    }
    return nil
}
```

---

## 8. Event-Sourcing vs. CRUD: Entscheidungsmatrix

### 8.1 Ausführliche Entscheidungsmatrix

| Kriterium                     | Event-Sourcing bevorzugen                        | CRUD bevorzugen                            |
| ----------------------------- | ------------------------------------------------ | ------------------------------------------ |
| **Audit Trail**               | Pflicht (Compliance, wer/was/wann)               | Nicht relevant                             |
| **Temporal Queries**          | Zustand zu beliebigem Zeitpunkt nötig            | Nur aktueller Zustand relevant             |
| **Geschäftsregeln**           | Komplex, zeitabhängig, regelbasiert              | Einfach, statisch                          |
| **Schreibmuster**             | Append-only, viele kleine Änderungen             | Full Updates, seltene Änderungen           |
| **Lesemuster**                | Aggregationen, Zeitreihen, Projektionen          | Einfache Lookups, Filtern, Sortieren       |
| **Debugging & Fehleranalyse** | Exakter Replay von Produktionsproblemen nötig    | Log-Dateien ausreichend                    |
| **Schema-Evolution**          | Häufig, fachlich getrieben                       | Selten, stabile Tabellen                   |
| **Eventual Consistency**      | Akzeptabel / gewünscht                           | Strenge Konsistenz erforderlich            |
| **Datenmenge pro Aggregat**   | Überschaubar (< 10.000 Events, mit Snapshots)    | Unbegrenzt                                 |
| **Undo/Redo**                 | Erforderlich (Stornierungen, Korrekturen)        | Nicht benötigt                             |
| **Event-Streaming**           | Integration mit anderen Services über Events     | Kein Event-Streaming geplant               |
| **Team-Erfahrung**            | Team kennt ES-Patterns                           | CRUD-Erfahrung, keine ES-Lernbereitschaft  |
| **Systemkomplexität**         | Komplexes Domain-Modell rechtfertigt ES-Overhead | Einfaches System, ES wäre Over-Engineering |
| **DSGVO/Datenlöschung**       | Kryptografisches Erasure implementierbar         | Standard-DELETE ausreichend                |

### 8.2 Hybridstrategie (empfohlen)

In den meisten Systemen ist **kein Entweder-oder** notwendig. Die optimale Strategie kombiniert beide Muster:

| Domäne                          | Muster         | Begründung                                              |
| ------------------------------- | -------------- | ------------------------------------------------------- |
| **Finanztransaktionen**         | Event-Sourcing | Nachvollziehbarkeit, Compliance, unveränderliche Ledger |
| **Bestellverwaltung**           | Event-Sourcing | Audit Trail, Statusübergänge, Undo/Redo                 |
| **Komplexe Workflows**          | Event-Sourcing | Zeitabhängige Geschäftsregeln, Saga-Integration         |
| **Benutzerverwaltung**          | CRUD           | Einfache Entität, kein Audit Trail nötig                |
| **Stammdaten (Katalog)**        | CRUD           | Selten geändert, stabile Tabellen                       |
| **Konfiguration/Einstellungen** | CRUD           | Einfacher Zustand, keine Historisierung nötig           |

### 8.3 Entscheidungsfluss

```
Ist ein vollständiger Audit Trail Pflicht?
  → Ja: Event-Sourcing
  → Nein: Weiter ↓

Ist Zeitreise (historischer Zustand) notwendig?
  → Ja: Event-Sourcing
  → Nein: Weiter ↓

Gibt es komplexe, zeitabhängige Geschäftsregeln?
  → Ja: Event-Sourcing evaluieren
  → Nein: CRUD

Hat das Team ES-Erfahrung?
  → Nein + einfaches System: CRUD
  → Ja oder Lernbereitschaft + geschäftlicher Mehrwert: Event-Sourcing
```

---

## 9. Reale Fallstudien

### 9.1 Banking & Finanzsysteme

**Kontext:** Kernbankensysteme verwalten Konten und Transaktionen.

**Warum Event-Sourcing?**

- Regulatorische Anforderungen verlangen lückenlose Audit Trails
- Kontostand ist korrekturdurchführbar: falsche Buchung wird durch Gegenbuchung korrigiert, nie überschrieben
- Zeitreise: Kontoauszug zum Stichtag muss exakt rekonstruierbar sein

**Muster:** Jede Kontobewegung (Einzahlung, Auszahlung, Überweisung) ist ein unveränderlicher Event. Der Kontostand ist eine Projektion (CQRS). Sagas koordinieren mehrstufige Überweisungen über Service-Grenzen.

### 9.2 E-Commerce (Bestellverwaltung)

**Kontext:** Plattformen wie Amazon verarbeiten Millionen von Bestellungen.

**Warum Event-Sourcing?**

- Bestellstatus durchläuft viele Zustände: `Placed → Paid → Shipped → Delivered → Returned`
- Jeder Schritt muss nachvollziehbar sein (Kundenservice, Compliance)
- Stornierungen und Rücksendungen sind Kompensations-Events, keine Deletes
- Sagas koordinieren: Bestellung → Zahlung → Lager → Versand

**Muster:** Outbox Pattern für zuverlässige Event-Veröffentlichung zwischen Microservices. Projektionen für Kundenbereich (Bestellhistorie), Reporting, Analytics.

### 9.3 Logistik & Supply Chain

**Kontext:** Tracking von Paketen, Container, Sendungen.

**Warum Event-Sourcing?**

- Positionshistorie ist intrinsisch event-basiert: jeder Scan ist ein Event
- SLA-Überwachung benötigt exakte Zeitstempel jedes Zustandswechsels
- Debugging: bei Lieferproblemen muss exakter Weg rekonstruiert werden

**Muster:** Event-Stream pro Sendungs-ID. Projektionen für aktuelle Position (Read Model). Temporal Queries für SLA-Reports.

### 9.4 Non-Profit Gastronomie-Kassensystem

**Kontext:** Mobiles Kassensystem für temporäre Gastronomie-Veranstaltungen (Vereinsfeste, Märkte).

**Warum Event-Sourcing?**

- Tisch-Zustand (Bestellungen, Zahlungen, Stornierungen) ist natürlich event-basiert
- Kassenbericht muss am Ende nachvollziehbar sein
- Stornierungen sind explizite Ereignisse, keine Löschungen

> Für ein konkretes Beispiel eines Event-Sourcing-basierten POS-Systems (Gastronomie) siehe [POS-Systeme & Gastronomie-Domäne](pos.md).

---

## 10. Kombination mit CQRS

Event-Sourcing und CQRS ergänzen sich natürlich: Event-Sourcing löst das Write-Problem optimal (append-only, immutable), aber erzeugt ein Leseproblem (Replay für jeden Query). CQRS löst das Leseproblem durch separate Read Models (Projektionen).

```
Write Side (Event-Sourcing):
  Command → Command Handler → Aggregate rehydrieren
                            → Business Rules prüfen
                            → Event(s) in Event Store schreiben

Event Store → Projektion (synchron oder asynchron) → Read Store

Read Side (CQRS):
  Query → Query Handler → Read Store (optimiertes Read Model) → Response
```

**Warum sie zusammenpassen:**

| Event-Sourcing Problem                     | CQRS-Lösung                                          |
| ------------------------------------------ | ---------------------------------------------------- |
| Replay für jeden Read-Zugriff              | Read Store mit materialisierten Projektionen         |
| Events sind kein Query-freundliches Format | Denormalisierte Read Models für schnelle Queries     |
| Eventual Consistency auf der Leseseite     | Explizites Read/Write-Modell macht Tradeoff sichtbar |

**→ Ausführliche Darstellung in [CQRS Theorie](cqrs.md#6-kombination-mit-event-sourcing)**, insbesondere:

- Ausbaustufen (Stufe 0–3)
- Projektionsstrategien (synchron, asynchron, CDC)
- Eventual Consistency Strategien
- Read Model Design (Denormalisierung, Materialized Views)

---

## 11. Anti-Patterns

### 11.1 Event als Command

**Problem:** Events beschreiben, was **passieren soll**, statt was **passiert ist**.

```go
// FALSCH: Event als Absichtserklärung
type PlaceOrderEvent struct { ... }  // "Platziere Bestellung" — Imperativ

// RICHTIG: Event als Tatsache
type OrderPlacedEvent struct { ... }  // "Bestellung platziert" — Vergangenheit
```

### 11.2 Zu große Events (God Events)

**Problem:** Ein Event enthält den gesamten Aggregat-Zustand statt nur die Änderung.

```go
// FALSCH: Gesamtzustand im Event
type OrderUpdatedEvent struct {
    AllItems    []Item
    AllPayments []Payment
    Balance     int
}

// RICHTIG: Nur die Änderung
type ItemAddedToOrderEvent struct {
    Item       Item
    NewTotalCents int
}
```

### 11.3 Events mutieren

**Problem:** Bestehende Events nachträglich ändern (z.B. Preis korrigieren, Typo im Event-Typ fixen).

**Lösung:** Korrekturen als **neue Kompensations-Events** modellieren. Schema-Änderungen über Upcasting handhaben (→ Abschnitt 6.2).

### 11.4 CRUD getarnt als Event-Sourcing

**Problem:** Nur CRUD-Operationen in Events verpacken, ohne fachlichen Mehrwert.

```go
// FALSCH: CRUD als Events — kein fachlicher Gewinn
type UserCreatedEvent struct { Name, Role string }
type UserUpdatedEvent struct { Name, Role string }
type UserDeletedEvent struct { ID int }

// RICHTIG: Für einfache Entitäten reicht CRUD
INSERT INTO users (name, role) VALUES (...);
UPDATE users SET name = ... WHERE id = ...;
```

### 11.5 Fehlende Snapshot-Strategie

**Problem:** Replay über tausende Events bei jeder Query; akzeptable Latenz nur bei kurzen Streams.

**Lösung:** Explizite Snapshot-Strategie von Anfang an definieren — auch wenn sie zunächst „bei jedem Write" ist.

### 11.6 Event-Sourcing überall erzwingen

**Problem:** ES für alle Entitäten im System einsetzen, auch wenn kein Mehrwert.

**Lösung:** Hybridstrategie — ES nur dort wo Audit Trail, Zeitreise oder komplexe Geschäftsregeln einen echten Mehrwert bieten (→ Abschnitt 8.2).

---

## 12. Referenzen

### Primärquellen

- **Greg Young** (2010): _CQRS Documents_ — [cqrs.wordpress.com](https://cqrs.wordpress.com/) — Ursprung von CQRS + Event-Sourcing
- **Martin Fowler** (2005): [Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html) — Grundlegende Definition und Konzepte
- **Martin Kleppmann** (2017): _Designing Data-Intensive Applications_ — Kapitel zu Event Sourcing, Immutability und Stream Processing

### Praxisquellen

- [Event Sourcing Explained (BayTech 2025)](https://www.baytechconsulting.com/blog/event-sourcing-explained-2025) — Paradigm Shift, Core Mechanics, Event Store Anatomy, Snapshots
- [Event Sourcing vs. CRUD (dev.to)](https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj) — Entscheidungsmatrix, Praxisbeispiele
- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — Event-Sourcing + CQRS in Go-Microservices
- [Oskar Dudycz: Event-Driven.io](https://event-driven.io/en/) — Praxis-Blog: Outbox, Snapshots, Schema-Evolution, Upcasting
- [Chris Richardson: Microservices Patterns](https://microservices.io/patterns/) — Saga, Outbox, Inbox Pattern
- [EventStoreDB Docs](https://developers.eventstore.com/) — Spezialisierter Event Store, Subscriptions

### Projekt-intern

- [CQRS Theorie](cqrs.md) — Command Query Responsibility Segregation
- [Event-Sourcing vs. CRUD: Entscheidungsmatrix](#8-event-sourcing-vs-crud-entscheidungsmatrix) — Entscheidungsmatrix und Hybridstrategie
