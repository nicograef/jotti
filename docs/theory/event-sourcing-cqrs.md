# Event-Sourcing & CQRS — Theorie und Anwendung in jotti

Dieses Dokument dient als theoretisches Nachschlagewerk für Event-Sourcing und CQRS. Es erklärt beide Muster unabhängig voneinander, beschreibt ihre Synergie und zeigt, wie jotti sie konkret einsetzt.

> **Verwandte Dokumente:**
>
> - [ADR: Event-Sourcing](../adr/event-sourcing.md) — Entscheidung für Event-Sourcing vs. CRUD
> - [CQRS in jotti (operativ)](../cqrs.md) — Ist-Zustand und Implementierungsplan
> - [Event-Sourcing vs. CRUD (Vergleich)](../event-sourcing-vs-crud.md) — Detaillierter Alternativenvergleich
> - [DDD Theorie](ddd.md) — Domain-Driven Design Grundlagen
> - [Architektur-Übersicht](README.md) — Index aller Theorie-Dokumente

---

## Inhaltsverzeichnis

1. [Event-Sourcing — Theorie](#1-event-sourcing--theorie)
2. [CQRS — Theorie](#2-cqrs--theorie)
3. [Event-Sourcing + CQRS: Die Kombination](#3-event-sourcing--cqrs-die-kombination)
4. [Event-Sourcing vs. CRUD: Entscheidungsmatrix](#4-event-sourcing-vs-crud-entscheidungsmatrix)
5. [Event-Sourcing in jotti — Ist-Zustand](#5-event-sourcing-in-jotti--ist-zustand)
6. [CQRS in jotti — Ist-Zustand und Ausbaustufen](#6-cqrs-in-jotti--ist-zustand-und-ausbaustufen)
7. [Patterns und Best Practices](#7-patterns-und-best-practices)
8. [Anti-Patterns](#8-anti-patterns)
9. [Referenzen](#9-referenzen)

---

## 1. Event-Sourcing — Theorie

### 1.1 Grundidee

Event-Sourcing ist ein Persistenzmuster, bei dem **nicht der aktuelle Zustand**, sondern die **Folge aller Zustandsänderungen** (Events) gespeichert wird. Der aktuelle Zustand wird durch Abspielen (Replay) aller Events rekonstruiert.

> **Greg Young:** _"Instead of storing just the current state of the data in a domain, use an append-only store to record the full series of actions taken on that data."_

**Analogie:** Ein Bankkonto speichert nicht nur den Kontostand (500€), sondern die vollständige Transaktionshistorie (+1000€ Einzahlung, -300€ Abhebung, -200€ Überweisung). Der Kontostand ist eine **Projektion** der Events.

### 1.2 Kernkonzepte

#### Event Store

Der Event Store ist eine **append-only** Datenbank (oder Tabelle). Events werden nur eingefügt, nie geändert oder gelöscht. Diese Unveränderlichkeit ist die stärkste Garantie des Musters.

```
Event Store (append-only):
┌────┬──────────────────────────┬────────────┬──────────────────────┐
│ ID │ Type                     │ Subject    │ Data (JSONB)         │
├────┼──────────────────────────┼────────────┼──────────────────────┤
│  1 │ bestellung-aufgegeben:v1 │ tisch:1    │ {positionen: [...]}  │
│  2 │ bestellung-aufgegeben:v1 │ tisch:2    │ {positionen: [...]}  │
│  3 │ zahlung-registriert:v1   │ tisch:1    │ {positionen: [...]}  │
│  4 │ produkte-geliefert:v1    │ tisch:1    │ {positionen: [...]}  │
│  5 │ produkte-storniert:v1    │ tisch:2    │ {positionen: [...]}  │
└────┴──────────────────────────┴────────────┴──────────────────────┘
```

#### Event

Ein Event beschreibt eine **vergangene Tatsache**. Es ist immutable und enthält alle Daten, die nötig sind, um die Zustandsänderung zu reproduzieren.

**Eigenschaften eines guten Events:**

| Eigenschaft            | Beschreibung                                                           |
| ---------------------- | ---------------------------------------------------------------------- |
| **Vergangenheitsform** | `BestellungAufgegeben`, nicht `BestellungAufgeben`                     |
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

## 2. CQRS — Theorie

### 2.1 Grundidee

**Command Query Responsibility Segregation** trennt die Verantwortlichkeit für Schreiboperationen (Commands) und Leseoperationen (Queries) auf System-Ebene. CQRS erweitert das CQS-Prinzip (Command Query Separation) von Bertrand Meyer:

> **CQS (Methoden-Ebene):** Eine Methode soll entweder den Zustand ändern (Command) ODER Daten zurückgeben (Query) — nie beides.
>
> **CQRS (System-Ebene):** Das System hat **separate Modelle** für Schreiben und Lesen.

### 2.2 Command Side

Commands drücken eine **Absicht** aus: „Gib diese Bestellung auf", „Registriere diese Zahlung". Commands:

- Ändern den Zustand
- Geben maximal eine ID oder Erfolgsmeldung zurück
- Werden validiert
- Können abgelehnt werden (z.B. ungültige Daten)
- Sind idempotent bei wiederholter Ausführung (idealerweise)

```go
// Command: Absichtserklärung
type BestellungAufgebenCommand struct {
    TischID    int
    Positionen []Position
    Comment    string
}

// Command Handler: Verarbeitung
func (h *Handler) BestellungAufgeben(ctx context.Context, cmd BestellungAufgebenCommand) error {
    // 1. Validieren
    // 2. Business Rules prüfen
    // 3. Event erzeugen und speichern
    return nil
}
```

### 2.3 Query Side

Queries fragen Daten ab, **ohne den Zustand zu ändern**. Queries:

- Haben keine Seiteneffekte
- Geben Daten zurück (Read Model / DTO)
- Können gecacht werden
- Können gegen optimierte Datenstrukturen arbeiten

```go
// Query: Datenanfrage
type GetTischSaldoQuery struct {
    TischID int
}

// Query Handler: Daten lesen
func (h *Handler) GetTischSaldo(ctx context.Context, q GetTischSaldoQuery) (int, error) {
    // Optimiertes Read Model abfragen
    return saldoCents, nil
}
```

### 2.4 CQRS-Ausbaustufen

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

### 2.5 Vor- und Nachteile von CQRS

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

## 3. Event-Sourcing + CQRS: Die Kombination

### 3.1 Warum sie zusammengehören

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

### 3.2 Projektionsstrategien

#### Synchrone Projektion

Das Read Model wird **im selben Request** wie das Event aktualisiert:

```go
func (s *CommandService) BestellungAufgeben(ctx, userID, tischID, ...) error {
    // 1. Event schreiben
    eventRepo.WriteEvent(ctx, event)
    // 2. Read Model aktualisieren (gleiche Transaktion)
    readRepo.UpdateSaldo(ctx, tischID, deltaCents)
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
              ├── Synchron:  Saldo aktualisieren (kritisch)
              └── Asynchron: Tagesstatistik aktualisieren (unkritisch)
```

### 3.3 Konsistenzanforderungen nach Feature

Nicht alle Read Models brauchen starke Konsistenz:

| Feature                     | Konsistenz           | Begründung                                       |
| --------------------------- | -------------------- | ------------------------------------------------ |
| **Saldo**                   | Stark (synchron)     | Fehlerhafte Saldo-Anzeige führt zu Fehlzahlungen |
| **Unbezahlte Positionen**   | Stark (synchron)     | Kassiervorgang muss korrekt sein                 |
| **Ungelieferte Positionen** | Stark (synchron)     | Lieferung muss vollständig sein                  |
| **Tisch-Historie**          | Eventual (asynchron) | Historische Ansicht verträgt kurze Verzögerung   |
| **Tagesabrechnung**         | Eventual (asynchron) | Wird nicht in Echtzeit benötigt                  |
| **Umsatzstatistiken**       | Eventual (asynchron) | Aggregierte Daten, keine Echtzeitanforderung     |

---

## 4. Event-Sourcing vs. CRUD: Entscheidungsmatrix

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

### Anwendung auf jotti

| Bereich                 | Muster         | Begründung                                            |
| ----------------------- | -------------- | ----------------------------------------------------- |
| **Tisch-Operationen**   | Event-Sourcing | Audit Trail, Zustandsrekonstruktion, fachliche Events |
| **Benutzer**            | CRUD           | Einfache Entität, kein Audit nötig                    |
| **Produkte/Varianten**  | CRUD           | Stammdaten, selten geändert                           |
| **Tische (Stammdaten)** | CRUD           | Name/Status, kein Event-Stream nötig                  |

---

## 5. Event-Sourcing in jotti — Ist-Zustand

### 5.1 Event-Schema

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

### 5.2 Event-Typen

| Event                 | Typ-String                       | Daten (JSONB)                                                                      |
| --------------------- | -------------------------------- | ---------------------------------------------------------------------------------- |
| Bestellung aufgegeben | `tisch.bestellung-aufgegeben:v1` | `{positionen, comment, gesamtPreisCents}`                                          |
| Zahlung registriert   | `tisch.zahlung-registriert:v1`   | `{positionen, comment, gesamtPreisCents}`                                          |
| Produkte storniert    | `tisch.produkte-storniert:v1`    | `{positionen, comment, gesamtPreisCents}`                                          |
| Produkte geliefert    | `tisch.produkte-geliefert:v1`    | `{positionen, comment}`                                                            |
| Snapshot              | `tisch.snapshot:v1`              | `{saldoCents, unbezahltePositionen, ungeliefertePositionen, gesamtZahlungenCents}` |

### 5.3 Zustandsrekonstruktion

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

### 5.4 Snapshot-Strategie

- Snapshots sind **selbst Events** (Typ `tisch.snapshot:v1`)
- Werden bei **jedem Command** erstellt (nach Event-Schreiben)
- `ReadEventsWithSnapshot()` liest Events ab dem letzten Snapshot
- Snapshot enthält: SaldoCents, UnbezahltePositionen, UngeliefertePositionen, GesamtZahlungenCents

---

## 6. CQRS in jotti — Ist-Zustand und Ausbaustufen

### 6.1 Ist-Zustand: Stufe 1 (Logische Trennung)

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

### 6.2 Nächste Stufe: Stufe 2 (Synchrone Projektion)

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

---

## 7. Patterns und Best Practices

### 7.1 Event Design

**Self-contained Events:** Events kopieren alle relevanten Daten (z.B. Produktname, Preis) zum Zeitpunkt der Aktion. Keine Referenzen auf veränderliche Stammdaten.

```go
// RICHTIG: Preis und Name im Event eingebettet
type Position struct {
    ID         int    // Varianten-ID (für Referenz)
    Name       string // "Cola 0,5l" — zum Zeitpunkt der Bestellung
    PreisCents int    // 350 — zum Zeitpunkt der Bestellung
    Quantity   int
}

// FALSCH: Nur ID referenzieren
type Position struct {
    VariantID int  // Was, wenn der Preis sich ändert?
    Quantity  int
}
```

### 7.2 Event-Versionierung

Events sind versioniert (`:v1`-Suffix). Bei Schema-Änderungen:

1. **Neue Version definieren** (`:v2`) mit neuem Schema
2. **Upcaster schreiben** — transformiert `:v1`-Events in `:v2`-Format beim Replay
3. **Alte Events niemals ändern** — sie sind immutable

```go
// Upcaster-Beispiel (hypothetisch)
func upcastV1toV2(event Event) Event {
    if event.Type == "tisch.bestellung-aufgegeben:v1" {
        // V1 hatte kein "comment"-Feld → Default setzen
        data := event.Data
        data["comment"] = ""
        return Event{...event, Type: "tisch.bestellung-aufgegeben:v2", Data: data}
    }
    return event
}
```

### 7.3 Idempotenz

Commands sollten idempotent sein — mehrfaches Ausführen desselben Commands hat den gleichen Effekt wie einmaliges Ausführen. Strategien:

- **Idempotency Key** im Request (Client-generierte UUID)
- **Deduplizierung** im Event Store (vor INSERT prüfen)
- **Natürliche Idempotenz** (z.B. "Setze Status auf aktiv" statt "Aktiviere")

### 7.4 Concurrency Control

Bei konkurrierenden Schreibzugriffen auf denselben Event-Stream:

- **Optimistic Concurrency:** Expected Version beim Schreiben prüfen
- **Pessimistic Locking:** `SELECT ... FOR UPDATE` auf Aggregate-Ebene
- **Last-Writer-Wins:** Akzeptabel bei unabhängigen Events (z.B. zwei Bestellungen am selben Tisch = beide gültig)

**In jotti:** Derzeit Last-Writer-Wins — bei einem Vereinsfest mit wenigen gleichzeitigen Servicekräften pro Tisch ist das akzeptabel.

### 7.5 Event Store-Optimierung

| Optimierung         | Beschreibung                       | In jotti                            |
| ------------------- | ---------------------------------- | ----------------------------------- |
| **Snapshots**       | Komprimierter Zustand              | ✅ Ja, als Event-Typ                |
| **Partitionierung** | Events nach Subject partitionieren | ❌ Nicht nötig (geringe Datenmenge) |
| **Archivierung**    | Alte Events in Cold Storage        | ❌ Nicht nötig                      |
| **Indexierung**     | Index auf `(subject, type)`        | ✅ Ja                               |
| **Batch-Reads**     | Events in einem Query laden        | ✅ Ja (`ReadEventsWithSnapshot`)    |

---

## 8. Anti-Patterns

### 8.1 Event als Command

**Problem:** Events beschreiben, was **passieren soll**, statt was **passiert ist**.

```go
// FALSCH: Event als Absichtserklärung
type PlaceOrderEvent struct { ... }  // "Platziere Bestellung" — Imperativ

// RICHTIG: Event als Tatsache
type OrderPlacedEvent struct { ... }  // "Bestellung platziert" — Vergangenheit
```

### 8.2 Zu große Events

**Problem:** Ein Event enthält den gesamten Aggregat-Zustand statt nur die Änderung.

```go
// FALSCH: Gesamtzustand im Event
type TischUpdatedEvent struct {
    AlleBestellungen    []Bestellung
    AlleZahlungen       []Zahlung
    Saldo               int
}

// RICHTIG: Nur die Änderung
type BestellungAufgegebenEvent struct {
    Positionen      []Position
    GesamtPreisCents int
}
```

### 8.3 Events mutieren

**Problem:** Bestehende Events nachträglich ändern (z.B. Preis korrigieren).

**Lösung:** Korrekturen als **neue Events** modellieren (z.B. `StornierungEvent` + neues `BestellungEvent`).

### 8.4 CRUD getarnt als Event-Sourcing

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

### 8.5 Fehlende Snapshot-Strategie

**Problem:** Replay über tausende Events bei jeder Query.

**Lösung:** Snapshots nach N Events oder nach jeder Schreiboperation erstellen.

---

## 9. Referenzen

### Primärquellen

- **Greg Young** (2010): _CQRS Documents_ — cqrs.wordpress.com — Ursprung von CQRS + Event-Sourcing
- **Martin Fowler**: [Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html), [CQRS](https://martinfowler.com/bliki/CQRS.html)
- **Udi Dahan**: [Clarified CQRS](https://udidahan.com/2009/12/09/clarified-cqrs/) — CQRS + DDD
- **AWS Prescriptive Guidance**: [CQRS Pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html)

### Praxisquellen

- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — Event-Sourcing + CQRS in Go-Microservices
- [Event Sourcing Explained (2025)](https://www.baytechconsulting.com/blog/event-sourcing-explained-2025) — Write-Store als Single Source of Truth
- [Event Sourcing vs. CRUD](https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj) — Entscheidungsmatrix

### Projekt-intern

- [ADR: Event-Sourcing](../adr/event-sourcing.md) — Architekturbewertung pro/contra
- [CQRS in jotti (operativ)](../cqrs.md) — Detaillierter Implementierungsplan
- [Event-Sourcing vs. CRUD](../event-sourcing-vs-crud.md) — 8-Tabellen-CRUD-Alternative
