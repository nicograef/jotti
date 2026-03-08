# Event-Sourcing — Theorie

Dieses Dokument dient als theoretisches Nachschlagewerk für Event-Sourcing. Es erklärt das Muster, seine Kernkonzepte, Entscheidungskriterien gegenüber CRUD, Patterns und Anti-Patterns. Ein projektspezifisches Anwendungsbeispiel findet sich im [Appendix](#appendix-anwendungsbeispiel-jotti).

> **Verwandte Dokumente:**
>
> - [CQRS Theorie](cqrs.md) — Command Query Responsibility Segregation
> - [DDD Theorie](ddd.md) — Domain-Driven Design Grundlagen
> - [ADR: Event-Sourcing](../adr/event-sourcing.md) — Entscheidung für Event-Sourcing vs. CRUD
> - [Event-Sourcing vs. CRUD (Vergleich)](../event-sourcing-vs-crud.md) — Detaillierter Alternativenvergleich
> - [Architektur-Übersicht](README.md) — Index aller Theorie-Dokumente

---

## Inhaltsverzeichnis

1. [Event-Sourcing — Grundlagen](#1-event-sourcing--grundlagen)
2. [Event-Sourcing vs. CRUD: Entscheidungsmatrix](#2-event-sourcing-vs-crud-entscheidungsmatrix)
3. [Kombination mit CQRS](#3-kombination-mit-cqrs)
4. [Patterns und Best Practices](#4-patterns-und-best-practices)
5. [Anti-Patterns](#5-anti-patterns)
6. [Appendix: Anwendungsbeispiel (jotti)](#appendix-anwendungsbeispiel-jotti)
7. [Referenzen](#7-referenzen)

---

## 1. Event-Sourcing — Grundlagen

### 1.1 Grundidee

Event-Sourcing ist ein Persistenzmuster, bei dem **nicht der aktuelle Zustand**, sondern die **Folge aller Zustandsänderungen** (Events) gespeichert wird. Der aktuelle Zustand wird durch Abspielen (Replay) aller Events rekonstruiert.

> **Greg Young:** _"Instead of storing just the current state of the data in a domain, use an append-only store to record the full series of actions taken on that data."_

**Analogie:** Ein Bankkonto speichert nicht nur den Kontostand (500€), sondern die vollständige Transaktionshistorie (+1000€ Einzahlung, -300€ Abhebung, -200€ Überweisung). Der Kontostand ist eine **Projektion** der Events.

### 1.2 Kernkonzepte

#### Event Store

Der Event Store ist eine **append-only** Datenbank (oder Tabelle). Events werden nur eingefügt, nie geändert oder gelöscht. Diese Unveränderlichkeit ist die stärkste Garantie des Musters.

```
Event Store (append-only):
┌────┬────────────────────────────┬────────────┬──────────────────────┐
│ ID │ Type                       │ Subject    │ Data (JSONB)         │
├────┼────────────────────────────┼────────────┼──────────────────────┤
│  1 │ order.placed:v1            │ order:101  │ {items: [...]}       │
│  2 │ order.placed:v1            │ order:102  │ {items: [...]}       │
│  3 │ order.payment-received:v1  │ order:101  │ {amount: 4200}       │
│  4 │ order.items-shipped:v1     │ order:101  │ {items: [...]}       │
│  5 │ order.item-returned:v1     │ order:102  │ {items: [...]}       │
└────┴────────────────────────────┴────────────┴──────────────────────┘
```

#### Event

Ein Event beschreibt eine **vergangene Tatsache**. Es ist immutable und enthält alle Daten, die nötig sind, um die Zustandsänderung zu reproduzieren.

**Eigenschaften eines guten Events:**

| Eigenschaft            | Beschreibung                                                           |
| ---------------------- | ---------------------------------------------------------------------- |
| **Vergangenheitsform** | `OrderPlaced`, nicht `PlaceOrder`                                      |
| **Self-contained**     | Alle Daten im Event, keine externen Referenzen auf veränderliche Daten |
| **Versioniert**        | Schema-Evolution durch Versionsnummer (`:v1`, `:v2`)                   |
| **Geordnet**           | Sequenzielle ID oder Zeitstempel für kausale Ordnung                   |
| **Kleinstmöglich**     | Nur die relevante Zustandsänderung, kein redundanter Kontext           |

#### State Reconstruction (Replay)

Der aktuelle Zustand wird durch **Abspielen aller Events** in Reihenfolge berechnet:

```
State₀ = Anfangszustand (leer)
State₁ = apply(State₀, Event₁)   // Bestellung: +50€
State₂ = apply(State₁, Event₂)   // Zahlung: -30€
State₃ = apply(State₂, Event₃)   // Stornierung: -10€
→ Aktueller Zustand: Saldo = 10€
```

#### Snapshots

Problem: Bei vielen Events wird das Replay langsam. Lösung: **Snapshots** speichern einen komprimierten Zustand zu einem bestimmten Zeitpunkt. Beim Replay muss nur ab dem letzten Snapshot abgespielt werden.

```
Events:  [1] [2] [3] [4] [5] [SNAPSHOT@5] [6] [7] [8]
                                    ↑
                            Zustand nach Event 5
                            (komprimiert gespeichert)

Replay für aktuellen Zustand:
  Statt: Event 1→2→3→4→5→6→7→8  (8 Events)
  Nur:   SNAPSHOT@5 → 6→7→8      (Snapshot + 3 Events)
```

### 1.3 Vorteile von Event-Sourcing

| Vorteil                           | Beschreibung                                             |
| --------------------------------- | -------------------------------------------------------- |
| **Vollständiger Audit Trail**     | Jede Aktion ist dokumentiert — wer, was, wann            |
| **Zeitreisen (Temporal Queries)** | Zustand zu jedem Zeitpunkt rekonstruierbar               |
| **Keine Daten gehen verloren**    | Events werden nie gelöscht oder überschrieben            |
| **Einfache Schreiboperationen**   | Ein einzelnes INSERT pro Aktion                          |
| **Schema-Flexibilität**           | Neue Event-Typen hinzufügen, ohne bestehende zu ändern   |
| **Debugging**                     | Event-Stream zeigt exakt, wie ein Zustand entstanden ist |
| **Natürliche CQRS-Basis**         | Events als Write Model, Projektionen als Read Model      |

### 1.4 Nachteile von Event-Sourcing

| Nachteil                     | Beschreibung                                                            |
| ---------------------------- | ----------------------------------------------------------------------- |
| **Leseperformance**          | Replay ist langsamer als ein direkter DB-Select                         |
| **Snapshots nötig**          | Ohne Snapshots skaliert Replay nicht                                    |
| **Eventual Consistency**     | Read Models können veraltet sein                                        |
| **Komplexität**              | Mehr Code für Zustandsrekonstruktion                                    |
| **Event-Schema-Evolution**   | Alte Events müssen mit neuer Logik kompatibel bleiben                   |
| **Kein einfaches Reporting** | SQL-Queries über Events sind komplex                                    |
| **DSGVO/Datenlöschung**      | Events sind immutable — personenbezogene Daten sind schwer zu entfernen |

---

## 2. Event-Sourcing vs. CRUD: Entscheidungsmatrix

### Wann Event-Sourcing?

| Kriterium                   | Event-Sourcing bevorzugen        | CRUD bevorzugen                      |
| --------------------------- | -------------------------------- | ------------------------------------ |
| **Audit Trail**             | Pflicht (wer, was, wann)         | Nicht relevant                       |
| **Geschäftsregeln**         | Komplex, zeitabhängig            | Einfach, statisch                    |
| **Datenhistorie**           | Zustand zu jedem Zeitpunkt nötig | Nur aktueller Zustand relevant       |
| **Schreibmuster**           | Append-only, Events              | CRUD (Create, Read, Update, Delete)  |
| **Lesemuster**              | Aggregationen, Zeitreihen        | Einfache Lookups, Filtern, Sortieren |
| **Schema-Evolution**        | Häufig, neue Event-Typen         | Selten, stabile Tabellen             |
| **Datenmenge pro Aggregat** | Überschaubar (< 10.000 Events)   | Unbegrenzt                           |
| **Team-Erfahrung**          | Event-Sourcing bekannt           | CRUD-Erfahrung                       |

### Typische Anwendung

| Bereich                      | Muster         | Begründung                                                   |
| ---------------------------- | -------------- | ------------------------------------------------------------ |
| **Transaktionsoperationen**  | Event-Sourcing | Audit Trail, Zustandsrekonstruktion, fachliche Events        |
| **Benutzerverwaltung**       | CRUD           | Einfache Entität, kein Audit nötig                           |
| **Stammdaten (Katalog)**     | CRUD           | Stammdaten, selten geändert                                  |
| **Finanztransaktionen**      | Event-Sourcing | Nachvollziehbarkeit und Compliance                           |
| **Konfiguration/Einstellung**| CRUD           | Einfacher Zustand, keine Historie nötig                      |

---

## 3. Kombination mit CQRS

Event-Sourcing allein hat ein Performance-Problem beim Lesen: Jede Query muss den Event-Stream replayed. **CQRS (Command Query Responsibility Segregation)** löst dieses Problem durch separate Read Models (Projektionen), die aus dem Event-Stream aktualisiert werden.

```
Command Side: Client → Command Handler → Event Store (append-only)
                                              │
                                    Projektion (sync/async)
                                              │
Query Side:   Client ← Query Handler  ← Read Store (Projektions-Tabellen)
```

Die Kombination von Event-Sourcing und CQRS ist das gängigste Einsatzmuster: Events bilden das Write Model, Projektionen bilden optimierte Read Models.

**→ Ausführliche Darstellung in [CQRS — Theorie](cqrs.md)**

---

## 4. Patterns und Best Practices

### 4.1 Event Design

**Self-contained Events:** Events kopieren alle relevanten Daten (z.B. Produktname, Preis) zum Zeitpunkt der Aktion. Keine Referenzen auf veränderliche Stammdaten.

```go
// RICHTIG: Preis und Name im Event eingebettet
type OrderItem struct {
    ID         int    // Artikel-ID (für Referenz)
    Name       string // "Espresso" — zum Zeitpunkt der Bestellung
    PriceCents int    // 350 — zum Zeitpunkt der Bestellung
    Quantity   int
}

// FALSCH: Nur ID referenzieren
type OrderItem struct {
    ProductID int  // Was, wenn der Preis sich ändert?
    Quantity  int
}
```

### 4.2 Event-Versionierung

Events sind versioniert (`:v1`-Suffix). Bei Schema-Änderungen:

1. **Neue Version definieren** (`:v2`) mit neuem Schema
2. **Upcaster schreiben** — transformiert `:v1`-Events in `:v2`-Format beim Replay
3. **Alte Events niemals ändern** — sie sind immutable

```go
// Upcaster-Beispiel
func upcastV1toV2(event Event) Event {
    if event.Type == "order.placed:v1" {
        // V1 hatte kein "comment"-Feld → Default setzen
        data := event.Data
        data["comment"] = ""
        return Event{...event, Type: "order.placed:v2", Data: data}
    }
    return event
}
```

### 4.3 Idempotenz

Commands sollten idempotent sein — mehrfaches Ausführen desselben Commands hat den gleichen Effekt wie einmaliges Ausführen. Strategien:

- **Idempotency Key** im Request (Client-generierte UUID)
- **Deduplizierung** im Event Store (vor INSERT prüfen)
- **Natürliche Idempotenz** (z.B. "Setze Status auf aktiv" statt "Aktiviere")

### 4.4 Concurrency Control

Bei konkurrierenden Schreibzugriffen auf denselben Event-Stream:

- **Optimistic Concurrency:** Expected Version beim Schreiben prüfen
- **Pessimistic Locking:** `SELECT ... FOR UPDATE` auf Aggregate-Ebene
- **Last-Writer-Wins:** Akzeptabel bei unabhängigen Events (z.B. zwei unabhängige Bestellungen am selben Aggregat = beide gültig)

### 4.5 Event Store-Optimierung

| Optimierung         | Beschreibung                         |
| ------------------- | ------------------------------------ |
| **Snapshots**       | Komprimierter Zustand nach N Events  |
| **Partitionierung** | Events nach Subject partitionieren   |
| **Archivierung**    | Alte Events in Cold Storage          |
| **Indexierung**     | Index auf `(subject, type)`          |
| **Batch-Reads**     | Events in einem Query laden          |

---

## 5. Anti-Patterns

### 5.1 Event als Command

**Problem:** Events beschreiben, was **passieren soll**, statt was **passiert ist**.

```go
// FALSCH: Event als Absichtserklärung
type PlaceOrderEvent struct { ... }  // "Platziere Bestellung" — Imperativ

// RICHTIG: Event als Tatsache
type OrderPlacedEvent struct { ... }  // "Bestellung platziert" — Vergangenheit
```

### 5.2 Zu große Events

**Problem:** Ein Event enthält den gesamten Aggregat-Zustand statt nur die Änderung.

```go
// FALSCH: Gesamtzustand im Event
type OrderUpdatedEvent struct {
    AllItems    []Item
    AllPayments []Payment
    Balance     int
}

// RICHTIG: Nur die Änderung
type OrderPlacedEvent struct {
    Items      []Item
    TotalCents int
}
```

### 5.3 Events mutieren

**Problem:** Bestehende Events nachträglich ändern (z.B. Preis korrigieren).

**Lösung:** Korrekturen als **neue Events** modellieren (z.B. `ItemCancelledEvent` + neues `OrderPlacedEvent`).

### 5.4 CRUD getarnt als Event-Sourcing

**Problem:** Nur CRUD-Operationen in Events verpacken, ohne fachlichen Mehrwert.

```go
// FALSCH: CRUD als Events — kein fachlicher Gewinn
type UserCreatedEvent { Name, Role }
type UserUpdatedEvent { Name, Role }
type UserDeletedEvent { ID }

// RICHTIG: Für User reicht CRUD
INSERT INTO users (name, role) VALUES (...);
UPDATE users SET name = ... WHERE id = ...;
```

### 5.5 Fehlende Snapshot-Strategie

**Problem:** Replay über tausende Events bei jeder Query.

**Lösung:** Snapshots nach N Events oder nach jeder Schreiboperation erstellen.

---

## Appendix: Anwendungsbeispiel (jotti)

Dieser Abschnitt zeigt, wie Event-Sourcing konkret in jotti — einem Non-Profit-POS-System für Vereinsfeste — eingesetzt wird.

### Event-Schema

Events werden in einer einzigen PostgreSQL-Tabelle gespeichert:

```sql
CREATE TABLE events (
    id        INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
    user_id   INT REFERENCES users(id) NOT NULL,
    type      TEXT NOT NULL,       -- 'tisch.bestellung-aufgegeben:v1'
    subject   TEXT NOT NULL,       -- 'tisch:42'
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now(),
    data      JSONB NOT NULL
);
```

**Immutabilitäts-Garantie:** DB-Trigger verhindern `UPDATE`, `DELETE` und `TRUNCATE` auf der `events`-Tabelle.

### Event-Typen

| Event                 | Typ-String                       | Daten (JSONB)                                                                      |
| --------------------- | -------------------------------- | ---------------------------------------------------------------------------------- |
| Bestellung aufgegeben | `tisch.bestellung-aufgegeben:v1` | `{positionen, comment, gesamtPreisCents}`                                          |
| Zahlung registriert   | `tisch.zahlung-registriert:v1`   | `{positionen, comment, gesamtPreisCents}`                                          |
| Produkte storniert    | `tisch.produkte-storniert:v1`    | `{positionen, comment, gesamtPreisCents}`                                          |
| Produkte geliefert    | `tisch.produkte-geliefert:v1`    | `{positionen, comment}`                                                            |
| Snapshot              | `tisch.snapshot:v1`              | `{saldoCents, unbezahltePositionen, ungeliefertePositionen, gesamtZahlungenCents}` |

### Zustandsrekonstruktion

Die Rekonstruktion erfolgt in Go-Code (`domain/table/events.go`):

```go
// Saldo berechnen
func GetSaldoFromEvents(events []event.Event) int {
    saldo := 0
    for _, e := range events {
        switch e.Type {
        case EventTypeSnapshotV1:
            saldo = snapshot.SaldoCents  // Snapshot als Basis
        case EventTypeBestellungAufgegebenV1:
            saldo += bestellung.GesamtPreisCents
        case EventTypeZahlungRegistriertV1:
            saldo -= zahlung.GesamtPreisCents
        case EventTypeProdukteStorniertV1:
            saldo -= stornierung.GesamtPreisCents
        }
    }
    return saldo
}
```

**Muster:** Accumulate + Reduce über den Event-Stream. Snapshots setzen den Startwert.

### Snapshot-Strategie

- Snapshots sind **selbst Events** (Typ `tisch.snapshot:v1`)
- Werden bei **jedem Command** erstellt (nach Event-Schreiben)
- `ReadEventsWithSnapshot()` liest Events ab dem letzten Snapshot
- Snapshot enthält: SaldoCents, UnbezahltePositionen, UngeliefertePositionen, GesamtZahlungenCents

### Event-Sourcing vs. CRUD in jotti

| Bereich                 | Muster         | Begründung                                            |
| ----------------------- | -------------- | ----------------------------------------------------- |
| **Tisch-Operationen**   | Event-Sourcing | Audit Trail, Zustandsrekonstruktion, fachliche Events |
| **Benutzer**            | CRUD           | Einfache Entität, kein Audit nötig                    |
| **Produkte/Varianten**  | CRUD           | Stammdaten, selten geändert                           |
| **Tische (Stammdaten)** | CRUD           | Name/Status, kein Event-Stream nötig                  |

### Concurrency Control in jotti

Derzeit Last-Writer-Wins — bei einem Vereinsfest mit wenigen gleichzeitigen Servicekräften pro Tisch ist das akzeptabel.

### Event Store-Optimierung in jotti

| Optimierung         | Status                              |
| ------------------- | ----------------------------------- |
| **Snapshots**       | ✅ Ja, als Event-Typ                |
| **Partitionierung** | ❌ Nicht nötig (geringe Datenmenge) |
| **Archivierung**    | ❌ Nicht nötig                      |
| **Indexierung**     | ✅ Ja, Index auf `(subject, type)` |
| **Batch-Reads**     | ✅ Ja (`ReadEventsWithSnapshot`)    |

---

## 7. Referenzen

### Primärquellen

- **Greg Young** (2010): _CQRS Documents_ — cqrs.wordpress.com — Ursprung von CQRS + Event-Sourcing
- **Martin Fowler**: [Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html)

### Praxisquellen

- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — Event-Sourcing + CQRS in Go-Microservices
- [Event Sourcing Explained (2025)](https://www.baytechconsulting.com/blog/event-sourcing-explained-2025) — Write-Store als Single Source of Truth
- [Event Sourcing vs. CRUD](https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj) — Entscheidungsmatrix

### Projekt-intern

- [CQRS Theorie](cqrs.md) — Command Query Responsibility Segregation
- [ADR: Event-Sourcing](../adr/event-sourcing.md) — Architekturbewertung pro/contra
- [Event-Sourcing vs. CRUD](../event-sourcing-vs-crud.md) — 8-Tabellen-CRUD-Alternative
