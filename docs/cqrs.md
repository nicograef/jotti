# CQRS in jotti

Dieses Dokument beschreibt das Architekturmuster **Command Query Responsibility Segregation (CQRS)**, analysiert dessen aktuelle Nutzung in jotti und zeigt, wie eine vollständige CQRS-Implementierung die bekannten Schwächen des Event-Sourcing-Ansatzes in jotti mindern kann.

> **Verwandte Dokumente:**
> - [Event-Sourcing vs. CRUD](event-sourcing-vs-crud.md) — Hybride Architektur und Event-Sourcing-Implementierung
> - [jotti ohne Event-Sourcing](no-event-sourcing.md) — CRUD-Alternative und Vergleich

---

## Inhaltsverzeichnis

1. [CQRS — Theorie und Ursprung](#1-cqrs--theorie-und-ursprung)
2. [CQRS in jotti — Ist-Zustand](#2-cqrs-in-jotti--ist-zustand)
3. [Die Schwächen von Event-Sourcing und wie CQRS hilft](#3-die-schwächen-von-event-sourcing-und-wie-cqrs-hilft)
4. [Implementierungsplan: Vollständiges CQRS für jotti](#4-implementierungsplan-vollständiges-cqrs-für-jotti)
5. [Vor- und Nachteile der vorgeschlagenen Lösung](#5-vor--und-nachteile-der-vorgeschlagenen-lösung)
6. [Zusammenfassung und Empfehlung](#6-zusammenfassung-und-empfehlung)
7. [Referenzen](#7-referenzen)

---

## 1. CQRS — Theorie und Ursprung

### 1.1 Grundidee

**Command Query Responsibility Segregation (CQRS)** ist ein Architekturmuster, das die Verantwortlichkeit für *Schreiboperationen* (Commands) und *Leseoperationen* (Queries) in einem System strikt trennt. CQRS erweitert damit das ältere Prinzip der **Command Query Separation (CQS)** von Bertrand Meyer, das auf Methodenebene definiert: Eine Methode soll entweder den Zustand ändern (Command) oder Daten zurückgeben (Query) — aber nie beides.

> **Martin Fowler** schreibt dazu (martinfowler.com/bliki/CQRS.html):
> *"At its heart is the notion that you can use a different model to update information than the model you use to read information."*

CQRS hebt diese Trennung von der Methoden- auf die Service-/Systemebene: Command-Seite und Query-Seite erhalten eigene Modelle, eigene Schichten und können sogar auf eigenen Datenspeichern operieren.

### 1.2 Geschichte und Herkunft

- **Bertrand Meyer (1988)**: Formuliert CQS in *Object-Oriented Software Construction* — das konzeptionelle Fundament.
- **Udi Dahan (2008/2009)**: Überträgt das Prinzip auf Service-Ebene und veröffentlicht "Clarified CQRS" (udidahan.com). Er kombiniert CQRS mit Domain-Driven Design (DDD) und Service-orientierten Architekturen.
- **Greg Young (2010)**: Prägt den Begriff "CQRS" in seiner heutigen Form und popularisiert die Verbindung mit **Event Sourcing** ("CQRS Documents", cqrs.wordpress.com). Young zeigt, dass Event Sourcing und CQRS natürliche Partner sind: das Event Log ist das Write Model, Projektionen sind das Read Model.

### 1.3 Kernkonzepte

#### Command (Schreiben)

Ein Command ist eine Absichtserklärung: *"Platziere diese Bestellung"*, *"Registriere diese Zahlung"*. Commands:

- **Ändern den Zustand** des Systems (Seiteneffekte erlaubt)
- Geben **keinen Datensatz zurück** (allenfalls eine ID oder einen Erfolgsindikator)
- Werden **validiert** bevor sie ausgeführt werden
- Können **abgelehnt** werden (Fehler, ungültige Daten)

#### Query (Lesen)

Eine Query fragt Daten ab: *"Was ist der aktuelle Kontostand?"*, *"Welche Varianten sind noch offen?"*. Queries:

- **Ändern keinen Zustand** (keine Seiteneffekte)
- Geben **Daten zurück** (Read Model)
- Können gegen **separate, optimierte** Datenspeicher arbeiten
- Können **gecacht** werden, da sie idempotent sind

#### Das CQRS-Modell

```
┌─────────────────────────────────────────────────────────────┐
│                        Client (Frontend)                    │
└──────────┬──────────────────────────────────────┬───────────┘
           │ Commands                             │ Queries
           ▼                                      ▼
┌──────────────────────┐              ┌───────────────────────┐
│   Command Handler    │              │    Query Handler       │
│  (schreibt Events)   │              │   (liest Projektion)   │
└──────────┬───────────┘              └──────────┬────────────┘
           │ INSERT                              │ SELECT
           ▼                                      ▼
┌──────────────────────┐              ┌───────────────────────┐
│   Write Store        │  Projektion  │    Read Store          │
│   (events-Tabelle)   │ ──────────► │   (Projektions-DB)     │
│   append-only        │              │   optimiert für Lesen  │
└──────────────────────┘              └───────────────────────┘
```

### 1.4 CQRS und Event Sourcing — eine natürliche Verbindung

Martin Fowler beschreibt CQRS als natürlichen Partner von Event Sourcing:

> *"CQRS fits well with event-based programming models. It's common to see CQRS systems split into separate services communicating with Event Collaboration. This allows these services to easily take advantage of Event Sourcing."*

Der Grund: Beim Event Sourcing ist das Write Model der **Event Store** (append-only). Der aktuelle Zustand muss aus Events rekonstruiert werden — das ist inhärent eine Query-seitige Aufgabe. CQRS gibt dieser Aufgabe einen eigenen Ort: die **Projektion** (Read Model), die den aktuellen Zustand vorhält, ohne bei jeder Abfrage Events neu durchlaufen zu müssen.

### 1.5 CQRS-Varianten (Abstufungen)

CQRS ist kein Alles-oder-Nichts-Muster. Es gibt mehrere Abstufungsstufen:

| Stufe | Beschreibung | Komplexität |
|-------|-------------|-------------|
| **Logische Trennung** | Command/Query als separate Klassen oder Interfaces, gleiche DB | Gering |
| **Separate Read Models** | Eigene Query-Objekte, Projektionen im selben Store | Mittel |
| **Getrennte Datenspeicher** | Write DB + Read DB, asynchrone Synchronisierung | Hoch |
| **Vollständig verteiltes CQRS** | Separate Services, Message Broker, eventual consistency | Sehr hoch |

Fowler warnt ausdrücklich:

> *"You should be very cautious about using CQRS. Many information systems fit well with the notion of an information base that is updated in the same way that it's read, adding CQRS to such a system can add significant complexity."*

---

## 2. CQRS in jotti — Ist-Zustand

### 2.1 Was bereits umgesetzt ist

jotti implementiert CQRS bereits auf der **logischen Ebene** in der Application-Schicht — allerdings nur für den Tisch-Bereich (Event Sourcing). Die CRUD-Bereiche (Benutzer, Produkte, Tische-Stammdaten) sind nicht CQRS-artig getrennt.

#### Separate Command/Query-Structs (Application Layer)

```
backend/api/table/application/
├── command.go      # Command-Struct: schreibt Events
└── query.go        # Query-Struct: liest Events + rekonstruiert Zustand
```

**`Command`-Struct** (`command.go`) enthält ausschließlich zustandsändernde Operationen:

```go
type Command struct {
    TableRepo tableRepoCommand
    EventRepo eventRepoCommand
}

func (c Command) PlaceTableOrder(...)      // schreibt table.order-placed:v1
func (c Command) RegisterTablePayment(...) // schreibt table.payment-registered:v1
func (c Command) CancelTableVariants(...)  // schreibt table.variants-canceled:v1
func (c Command) DeliverTableVariants(...) // schreibt table.variants-delivered:v1
func (c Command) CreateTableSnapshot(...)  // schreibt table.snapshot:v1
```

**`Query`-Struct** (`query.go`) enthält ausschließlich lesende Operationen:

```go
type Query struct {
    TableRepo tableRepoQuery
    EventRepo eventRepoQuery
}

func (q Query) GetTableBalance(...)           // liest Events → berechnet Balance
func (q Query) GetTableHistory(...)           // liest Events → gibt History zurück
func (q Query) GetTableUnpaidVariants(...)    // liest Events → offene Positionen
func (q Query) GetTableUndeliveredVariants(...)// liest Events → ungelieferte Positionen
```

#### Separate Interfaces

Die Interfaces für das Event-Repository sind nach Command/Query getrennt:

```go
// Nur für Commands benötigt:
type eventRepoCommand interface {
    WriteEvent(ctx context.Context, event event.Event) (int, error)
    ReadEventsWithSnapshot(...)  // für CreateTableSnapshot
}

// Nur für Queries benötigt:
type eventRepoQuery interface {
    ReadEventsBySubject(...)
    ReadEventsWithSnapshot(...)
}
```

Dies folgt dem **Interface Segregation Principle (ISP)**: Command-Objekte sehen nicht die Query-Methoden des Repositories und umgekehrt.

#### Separate HTTP-Handler

```
backend/api/table/http/
├── command_handler.go  # CommandHandler: POST-Endpoints für Schreiboperationen
└── query_handler.go    # QueryHandler: POST-Endpoints für Leseoperationen
```

Die HTTP-Routen sind ebenfalls getrennt:

```go
// service.go — Commands und Queries nutzen separate Handler-Instanzen:
tc := table.NewCommandHandler(db)  // Command-seitig
r.HandleFunc("/place-table-order", tc.PlaceTableOrderHandler())
r.HandleFunc("/register-table-payment", tc.RegisterTablePaymentHandler())

tq := table.NewQueryHandler(db)    // Query-seitig
r.HandleFunc("/get-table-balance", tq.GetTableBalanceHandler())
r.HandleFunc("/get-table-history", tq.GetTableHistoryHandler())
```

### 2.2 Was noch fehlt

Obwohl die **logische CQRS-Trennung** vorhanden ist, fehlen zwei wesentliche Aspekte eines vollständigen CQRS:

1. **Kein separates Read Model**: Sowohl Commands als auch Queries greifen auf dieselbe `events`-Tabelle zu. Queries müssen immer noch alle Events laden und den Zustand in Go rekonstruieren — sie profitieren nicht von einem vorberechneten, optimierten Read Store.

2. **Keine event-getriebene Projektion**: Es gibt keinen Mechanismus, der beim Schreiben eines Events automatisch eine Projektion aktualisiert. Stattdessen gibt es den `CreateTableSnapshot`-Command, der manuell aufgerufen werden muss und selbst wieder Events schreibt.

### 2.3 Zusammenfassung Ist-Zustand

| Aspekt | Status | Anmerkung |
|--------|--------|-----------|
| Logische Command/Query-Trennung (Application) | ✅ Vorhanden | `Command`- und `Query`-Structs getrennt |
| Separate HTTP-Handler | ✅ Vorhanden | `CommandHandler` und `QueryHandler` |
| Interface Segregation | ✅ Vorhanden | Separate Repo-Interfaces |
| Separates Read Model | ❌ Fehlt | Queries lesen denselben Event Store |
| Event-getriebene Projektionen | ❌ Fehlt | Kein automatisches Projektion-Update |
| Separate Datenspeicher | ❌ Fehlt | Single DB für Read und Write |

jotti befindet sich auf CQRS-Stufe 1 (*Logische Trennung*) und hat klare Ansatzpunkte, um auf Stufe 2 (*Separate Read Models*) zu migrieren.

---

## 3. Die Schwächen von Event-Sourcing und wie CQRS hilft

Das Dokument [Event-Sourcing vs. CRUD](event-sourcing-vs-crud.md) identifiziert mehrere Nachteile des Event-Sourcing-Ansatzes in jotti. CQRS — konkret in Form von **Projektionen** als Read Model — adressiert diese direkt:

### 3.1 Nachteil: Höhere Lese-Komplexität

**Problem**: Jede Abfrage (Balance, offene Positionen, Verlauf) erfordert das Laden und sequenzielle Durchlaufen aller Events eines Tisches. Die Zustandsrekonstruktion ist algorithmisch teurer als ein direkter Datenbankzugriff.

**CQRS-Lösung**: Eine **Projektionstabelle** hält den aktuellen Zustand des Tisches bereits berechnet vor. Eine Balanceabfrage wird dann zu einem einfachen `SELECT balance_cents FROM table_state WHERE table_id = $1` — kein Event-Replay erforderlich.

```
Aktuell:  Query → load N events → iterate → reconstruct state → return
Mit CQRS: Query → SELECT from projection → return
```

### 3.2 Nachteil: Snapshot-Mechanismus nötig

**Problem**: Der aktuelle Snapshot-Mechanismus ist ein Workaround: Ein Snapshot ist selbst ein Event (`table.snapshot:v1`), das manuell über `CreateTableSnapshot()` ausgelöst werden muss. Das führt zu:
- Zusätzlichem Code für Snapshot-Erstellung und -Erkennung
- Manueller Auslösung (wer, wann, wie oft?)
- Snapshots als Teil des Event Logs — was konzeptuell fragwürdig ist (ein Snapshot *ist* kein Domain-Event)

**CQRS-Lösung**: Eine **synchrone Projektion** (aktualisiert beim Schreiben jedes Events) ersetzt den Snapshot vollständig. Nach jedem Command (`PlaceTableOrder`, `RegisterTablePayment`, etc.) wird die Projektions-Tabelle atomisch aktualisiert. Der Zustand ist immer aktuell — kein separater Snapshot-Mechanismus nötig.

### 3.3 Nachteil: Erschwertes Querying für Analysen

**Problem**: Ad-hoc-Abfragen (z.B. Umsatz pro Produkt, Tagesabrechnung) erfordern JSONB-Parsing in SQL oder eigene Projektionslogik in der Anwendung, da alle Daten in `data JSONB` stecken.

**CQRS-Lösung**: Analytische Projektionen mit **typisierten Spalten** können für spezifische Reporting-Anforderungen erstellt werden. Beispiel: Eine `daily_revenue`-View oder eine `variant_sales`-Projektionstabelle mit expliziten `variant_id INT`, `quantity INT`, `price_cents INT`-Spalten ermöglicht direkte `GROUP BY`-Abfragen.

### 3.4 Nachteil: JSONB-Daten nicht typsicher auf DB-Ebene

**Problem**: Die Query-Seite muss immer JSONB deserialisieren und validieren. Fehler in historischen Events (z.B. fehlendes Feld durch einen alten Bug) können zur Laufzeit die Zustandsrekonstruktion fehlschlagen lassen.

**CQRS-Lösung**: Das Read Model hat **typisierte Spalten** (`INT`, `TEXT`, `TIMESTAMPTZ`). Die Projektion wird beim Schreiben populiert — zu diesem Zeitpunkt ist der Event-Data bereits validiert. Lesezugriffe auf die Projektion benötigen keine JSONB-Deserialisierung mehr.

### 3.5 Nachteil: Konzeptuelle Hürde

**Problem**: Entwickler müssen Zustandsrekonstruktion, Snapshot-Logik und Event-Versionierung verstehen, bevor sie effektiv mit dem System arbeiten können.

**CQRS-Lösung**: Die Query-Seite wird **radikal vereinfacht**: Ein neuer Entwickler, der nur Daten lesen will, muss die Event-Sourcing-Details nicht verstehen. Er arbeitet gegen eine normale relationale Projektionstabelle. Nur Entwickler, die Commands implementieren, müssen das Event-Modell verstehen.

### 3.6 Übersicht: Event-Sourcing-Nachteile vs. CQRS-Lösungen

| Nachteil (Event Sourcing) | CQRS-Lösung | Aufwand |
|---------------------------|-------------|---------|
| Höhere Lese-Komplexität | Projektionstabelle mit vorberechnetem Zustand | Mittel |
| Snapshot-Mechanismus nötig | Synchrone Projektion ersetzt Snapshots | Mittel |
| Erschwertes Querying | Typisierte Projektionen für Analytics | Mittel |
| JSONB nicht typsicher (Query-Seite) | Read Model mit typisierten Spalten | Gering |
| Konzeptuelle Hürde | Vereinfachte Query-Seite | Gering |
| Keine referenzielle Integrität (Write-Seite) | **Nicht adressiert** — bleibt Validierungs-Aufgabe der Anwendung | — |
| Doppelte Validierung (Write-Seite) | **Nicht adressiert** — bleibt bei Event-Sourcing strukturell | — |

CQRS löst primär die **Lese-seitigen** Nachteile. Schreib-seitige Nachteile (fehlende referenzielle Integrität, JSONB) bleiben erhalten — sie sind inhärenter Teil des Event-Sourcing-Write-Modells.

---

## 4. Implementierungsplan: Vollständiges CQRS für jotti

Der Plan beschreibt die Migration von der aktuellen logischen CQRS-Trennung zu einem vollständigen CQRS mit **synchronen Projektionen** als Read Model. Der Ansatz ist bewusst inkrementell und vermeidet einen Big-Bang-Umbau.

### Designentscheidung: Synchrone vs. asynchrone Projektion

Es gibt zwei grundlegende Ansätze:

**Asynchrone Projektion**: Ein separater Prozess (Event Handler, Message Consumer) liest neue Events und aktualisiert Projektionen. Dies ermöglicht maximale Skalierbarkeit, führt aber zu *eventual consistency* — Leser sehen möglicherweise kurzzeitig veraltete Daten.

**Synchrone Projektion**: Die Projektion wird innerhalb derselben Datenbanktransaktion wie das Event-Insert aktualisiert. Starke Konsistenz — Leser sehen immer den aktuellen Zustand. Einfacher zu implementieren und zu testen.

**Für jotti ist die synchrone Projektion die richtige Wahl**, weil:
- Servicekräfte unmittelbar nach einer Bestellung die korrekte Balance sehen müssen (keine Verzögerungen akzeptabel)
- Die Last ist überschaubar (keine Hochlast-Skalierungsanforderungen)
- Die Implementierung einfacher ist und weniger Infrastruktur erfordert

### 4.1 Schritt 1 — Neue Migrationsdatei: `table_state`-Projektionstabelle

Neue Migration `database/migrations/03_add_table_state_projection.up.sql`:

```sql
-- Read Model: aktueller Zustand je Tisch
CREATE TABLE table_state (
    table_id              INT PRIMARY KEY REFERENCES tables(id),
    balance_cents         INT NOT NULL DEFAULT 0,
    total_payments_cents  INT NOT NULL DEFAULT 0,
    unpaid_variants       JSONB NOT NULL DEFAULT '[]',
    undelivered_variants  JSONB NOT NULL DEFAULT '[]',
    last_event_id         INT NOT NULL DEFAULT 0,  -- Konsistenz-Check
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE table_state IS
    'CQRS Read Model: vorberechneter Zustand jedes Tisches. '
    'Wird synchron beim Schreiben jedes Events aktualisiert. '
    'Darf nur über die Projektion-Logik geschrieben werden.';
```

Down-Migration `03_add_table_state_projection.down.sql`:

```sql
DROP TABLE IF EXISTS table_state;
```

**Warum JSONB für `unpaid_variants` und `undelivered_variants`?**

Diese Listen sind variabel lang und haben keine feste Kardinalität pro Tisch. Typisierte separate Tabellen (`table_state_unpaid_items`) wären möglich, würden aber mehrere Tabellen erfordern. Da die Query-Seite diese Daten ohnehin als `[]LineItem`-Slice zurückgibt und JSONB für Listen in PostgreSQL gut unterstützt wird, ist JSONB hier ein vertretbarer Kompromiss. Der entscheidende Vorteil gegenüber dem aktuellen Ansatz: Die Projektion wird **beim Schreiben** populiert, nicht beim Lesen — JSONB-Deserialisierung findet nur einmal statt (beim Schreiben), nicht bei jeder Abfrage.

### 4.2 Schritt 2 — Neues Repository: `table_state_repo`

Neues Package `backend/repository/table_state_repo/`:

```go
package table_state_repo

import (
    "context"
    "database/sql"
    "encoding/json"

    "github.com/nicograef/jotti/backend/domain/table"
)

type TableState struct {
    TableID              int
    BalanceCents         int
    TotalPaymentsCents   int
    UnpaidVariants       []table.LineItem
    UndeliveredVariants  []table.LineItem
    LastEventID          int
}

type Repository struct {
    DB *sql.DB
}

// Upsert schreibt oder aktualisiert den Zustand eines Tisches atomar.
// Wird innerhalb derselben Transaktion aufgerufen wie WriteEvent.
func (r Repository) Upsert(ctx context.Context, tx *sql.Tx, state TableState) error {
    unpaidJSON, err := json.Marshal(state.UnpaidVariants)
    if err != nil {
        return err
    }
    undeliveredJSON, err := json.Marshal(state.UndeliveredVariants)
    if err != nil {
        return err
    }

    _, err = tx.ExecContext(ctx, `
        INSERT INTO table_state
            (table_id, balance_cents, total_payments_cents,
             unpaid_variants, undelivered_variants, last_event_id, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, now())
        ON CONFLICT (table_id) DO UPDATE SET
            balance_cents        = EXCLUDED.balance_cents,
            total_payments_cents = EXCLUDED.total_payments_cents,
            unpaid_variants      = EXCLUDED.unpaid_variants,
            undelivered_variants = EXCLUDED.undelivered_variants,
            last_event_id        = EXCLUDED.last_event_id,
            updated_at           = now()
    `, state.TableID, state.BalanceCents, state.TotalPaymentsCents,
       unpaidJSON, undeliveredJSON, state.LastEventID)

    return err
}

// Get liest den aktuellen Zustand eines Tisches aus der Projektionstabelle.
func (r Repository) Get(ctx context.Context, tableID int) (TableState, error) {
    var state TableState
    var unpaidJSON, undeliveredJSON []byte

    err := r.DB.QueryRowContext(ctx, `
        SELECT table_id, balance_cents, total_payments_cents,
               unpaid_variants, undelivered_variants, last_event_id
        FROM table_state
        WHERE table_id = $1
    `, tableID).Scan(
        &state.TableID, &state.BalanceCents, &state.TotalPaymentsCents,
        &unpaidJSON, &undeliveredJSON, &state.LastEventID,
    )
    if err != nil {
        return TableState{}, err
    }

    if err := json.Unmarshal(unpaidJSON, &state.UnpaidVariants); err != nil {
        return TableState{}, err
    }
    if err := json.Unmarshal(undeliveredJSON, &state.UndeliveredVariants); err != nil {
        return TableState{}, err
    }

    return state, nil
}
```

### 4.3 Schritt 3 — Event Store mit Transaktionsunterstützung

Das bestehende `event_repo` muss eine **transaktionale Variante** von `WriteEvent` erhalten, die es ermöglicht, Event-Insert und Projektion-Upsert atomar auszuführen:

In `backend/repository/event_repo/repo.go` eine neue Methode ergänzen:

```go
// WriteEventTx schreibt ein Event innerhalb einer bestehenden Transaktion.
// Damit kann der Aufrufer Event-Insert und Projektion-Upsert atomar ausführen.
func (r Repository) WriteEventTx(ctx context.Context, tx *sql.Tx, e event.Event) (int, error) {
    var id int
    err := tx.QueryRowContext(ctx,
        `INSERT INTO events (user_id, type, subject, data, timestamp)
         VALUES ($1, $2, $3, $4, $5) RETURNING id`,
        e.UserID, e.Type, e.Subject, e.Data, e.Time,
    ).Scan(&id)
    return id, err
}

// BeginTx startet eine neue Transaktion.
func (r Repository) BeginTx(ctx context.Context) (*sql.Tx, error) {
    return r.DB.BeginTx(ctx, nil)
}
```

### 4.4 Schritt 4 — Projektion-Service (Domain Layer)

Ein neuer Projektion-Service in `backend/domain/table/projection.go` kapselt die Zustandsberechnung für die Projektion:

```go
package table

// ApplyEventToState wendet ein einzelnes Event auf einen bestehenden TableState an
// und gibt den neuen Zustand zurück. Reine Funktion — keine Seiteneffekte.
func ApplyEventToState(state ProjectionState, evt event.Event) (ProjectionState, error) {
    switch EventType(evt.Type) {
    case EventTypeOrderPlacedV1:
        order, err := buildOrderFromEvent(evt)
        if err != nil {
            return state, err
        }
        state.BalanceCents += order.TotalPriceCents
        state.UnpaidVariants = accumulateVariants(state.UnpaidVariants, order.Variants)
        state.UndeliveredVariants = accumulateVariants(state.UndeliveredVariants, order.Variants)

    case EventTypePaymentRegisteredV1:
        payment, err := buildPaymentFromEvent(evt)
        if err != nil {
            return state, err
        }
        state.BalanceCents -= payment.TotalPaymentCents
        state.TotalPaymentsCents += payment.TotalPaymentCents
        state.UnpaidVariants = reduceVariants(state.UnpaidVariants, payment.Variants)

    case EventTypeVariantsCanceledV1:
        cancelation, err := buildCancelationFromEvent(evt)
        if err != nil {
            return state, err
        }
        state.BalanceCents -= cancelation.TotalCancelationCents
        state.UnpaidVariants = reduceVariants(state.UnpaidVariants, cancelation.Variants)
        state.UndeliveredVariants = reduceVariants(state.UndeliveredVariants, cancelation.Variants)

    case EventTypeVariantsDeliveredV1:
        delivery, err := buildDeliveryFromEvent(evt)
        if err != nil {
            return state, err
        }
        state.UndeliveredVariants = reduceVariants(state.UndeliveredVariants, delivery.Variants)
    }

    return state, nil
}

// ProjectionState hält den vorberechneten Zustand eines Tisches.
type ProjectionState struct {
    BalanceCents        int
    TotalPaymentsCents  int
    UnpaidVariants      []LineItem
    UndeliveredVariants []LineItem
}
```

### 4.5 Schritt 5 — Command-Service mit synchroner Projektion

Die Commands in `backend/api/table/application/command.go` werden erweitert, um nach jedem Event die Projektion synchron zu aktualisieren:

```go
type TableStateRepo interface {
    Upsert(ctx context.Context, tx *sql.Tx, state table_state_repo.TableState) error
    Get(ctx context.Context, tableID int) (table_state_repo.TableState, error)
}

type EventRepoTx interface {
    WriteEventTx(ctx context.Context, tx *sql.Tx, e event.Event) (int, error)
    BeginTx(ctx context.Context) (*sql.Tx, error)
}

type Command struct {
    TableRepo      tableRepoCommand
    EventRepo      EventRepoTx
    TableStateRepo TableStateRepo
}

func (c Command) PlaceTableOrder(ctx context.Context, userID, tableID int,
    variants []table.LineItem, comment string) error {

    evt, err := table.NewOrderPlacedEvent(userID, tableID, variants, comment)
    if err != nil {
        return err
    }

    return c.writeEventAndUpdateProjection(ctx, tableID, evt)
}

// writeEventAndUpdateProjection schreibt ein Event und aktualisiert die
// Projektion atomar in einer Transaktion.
func (c Command) writeEventAndUpdateProjection(ctx context.Context,
    tableID int, evt event.Event) error {

    tx, err := c.EventRepo.BeginTx(ctx)
    if err != nil {
        return ErrDatabase
    }
    defer func() {
        // Rollback ist nach erfolgreichem Commit ein No-op (gibt sql.ErrTxDone zurück).
        // Der Rückgabewert wird hier bewusst ignoriert, da ein Rollback-Fehler nach
        // einem fehlgeschlagenen Commit keine zusätzliche Fehlerbehandlung erfordert.
        _ = tx.Rollback()
    }()

    // 1. Event schreiben
    eventID, err := c.EventRepo.WriteEventTx(ctx, tx, evt)
    if err != nil {
        return ErrDatabase
    }

    // 2. Aktuellen Projektion-Zustand lesen
    currentState, err := c.TableStateRepo.Get(ctx, tableID)
    if err != nil {
        if !errors.Is(err, sql.ErrNoRows) {
            // Echter Datenbankfehler — nicht stillschweigend ignorieren
            return ErrDatabase
        }
        // Tisch hat noch keinen Projektion-Eintrag: leerer Ausgangszustand
        currentState = table_state_repo.TableState{TableID: tableID}
    }

    // 3. Event auf Zustand anwenden
    projState := table.ProjectionState{
        BalanceCents:        currentState.BalanceCents,
        TotalPaymentsCents:  currentState.TotalPaymentsCents,
        UnpaidVariants:      currentState.UnpaidVariants,
        UndeliveredVariants: currentState.UndeliveredVariants,
    }
    newProjState, err := table.ApplyEventToState(projState, evt)
    if err != nil {
        return err
    }

    // 4. Projektion aktualisieren
    newState := table_state_repo.TableState{
        TableID:             tableID,
        BalanceCents:        newProjState.BalanceCents,
        TotalPaymentsCents:  newProjState.TotalPaymentsCents,
        UnpaidVariants:      newProjState.UnpaidVariants,
        UndeliveredVariants: newProjState.UndeliveredVariants,
        LastEventID:         eventID,
    }
    if err := c.TableStateRepo.Upsert(ctx, tx, newState); err != nil {
        return ErrDatabase
    }

    return tx.Commit()
}
```

### 4.6 Schritt 6 — Query-Service auf Projektion umstellen

Die Queries in `backend/api/table/application/query.go` werden auf das Read Model umgestellt:

```go
type tableStateRepoQuery interface {
    Get(ctx context.Context, tableID int) (table_state_repo.TableState, error)
}

type Query struct {
    TableRepo      tableRepoQuery
    EventRepo      eventRepoQuery       // Weiterhin für GetTableHistory
    TableStateRepo tableStateRepoQuery  // Neu: Read Model
}

// GetTableBalance liest direkt aus der Projektion — kein Event-Replay.
func (q Query) GetTableBalance(ctx context.Context, tableID int) (int, error) {
    state, err := q.TableStateRepo.Get(ctx, tableID)
    if err != nil {
        return 0, ErrDatabase
    }
    return state.BalanceCents, nil
}

// GetTableUnpaidVariants liest direkt aus der Projektion.
func (q Query) GetTableUnpaidVariants(ctx context.Context, tableID int) ([]t.LineItem, error) {
    state, err := q.TableStateRepo.Get(ctx, tableID)
    if err != nil {
        return nil, ErrDatabase
    }
    return state.UnpaidVariants, nil
}

// GetTableUndeliveredVariants liest direkt aus der Projektion.
func (q Query) GetTableUndeliveredVariants(ctx context.Context, tableID int) ([]t.LineItem, error) {
    state, err := q.TableStateRepo.Get(ctx, tableID)
    if err != nil {
        return nil, ErrDatabase
    }
    return state.UndeliveredVariants, nil
}

// GetTableHistory liest weiterhin Events — der vollständige Verlauf
// ist ein Event-Sourcing-Merkmal und nicht Teil der Projektion.
func (q Query) GetTableHistory(ctx context.Context, tableID int) ([]any, error) {
    subject := "table:" + strconv.Itoa(tableID)
    events, err := q.EventRepo.ReadEventsBySubject(ctx, subject)
    if err != nil {
        return nil, ErrDatabase
    }
    return t.GetHistoryFromEvents(events)
}
```

### 4.7 Schritt 7 — Snapshot-Mechanismus ablösen

Mit der synchronen Projektion ist der Snapshot-Mechanismus obsolet:

- `CreateTableSnapshot()` aus `command.go` entfernen
- `EventTypeSnapshotV1` und `NewSnapshotEvent()`, `buildSnapshotFromEvent()` können deprecated werden (oder für historische Kompatibilität erhalten bleiben)
- `ReadEventsWithSnapshot()` im `event_repo` kann vereinfacht werden (nur noch für `GetTableHistory` relevant)
- Der Snapshot-`case` in `GetBalanceFromEvents()`, `GetUnpaidVariantsFromEvents()` etc. kann für `GetTableHistory()` erhalten bleiben

Die `GetBalanceFromEvents()`, `GetUnpaidVariantsFromEvents()` etc. werden **nur noch für `GetTableHistory()`** verwendet (da der Verlauf weiterhin aus Events rekonstruiert wird). Für alle anderen Queries ist die Projektion der einzige Datenpfad.

### 4.8 Schritt 8 — Initialisierungsmigration (Backfill)

Für bestehende Produktionsdaten muss die `table_state`-Tabelle initial befüllt werden:

```sql
-- Backfill: table_state für alle Tische aus vorhandenen Events berechnen.
-- Dieser Schritt wird einmalig beim Deployment der Migration ausgeführt.
-- Danach übernimmt die synchrone Projektion.

-- Hinweis: Da die Daten in JSONB liegen, ist ein Backfill-Skript in Go
-- (oder ein einmaliger Go-CLI-Befehl) empfehlenswert:
-- go run ./cmd/backfill-table-state/main.go
```

Ein Go-Backfill-Skript (`cmd/backfill-table-state/main.go`) würde:
1. Alle Tische laden
2. Für jeden Tisch alle Events laden
3. Den Zustand aus Events berechnen (mit den bestehenden `GetBalanceFromEvents()` etc.)
4. Den berechneten Zustand in `table_state` schreiben

### 4.9 Übersicht der Änderungen

| Bereich | Änderung | Aufwand |
|---------|----------|---------|
| Datenbank | Migration: `table_state`-Tabelle | Gering |
| Repository | Neues `table_state_repo` | Mittel |
| Repository | `event_repo` um `WriteEventTx`/`BeginTx` erweitern | Gering |
| Domain | `ApplyEventToState()` + `ProjectionState` | Mittel |
| Application/Command | `writeEventAndUpdateProjection()` | Mittel |
| Application/Query | Auf `table_state_repo` umstellen | Gering |
| Application/Command | `CreateTableSnapshot()` entfernen | Gering |
| Backfill | Einmaliges Migrationsskript | Mittel |
| Tests | Unit-Tests für `ApplyEventToState()` | Mittel |
| **Gesamt** | | **Mittel** |

---

## 5. Vor- und Nachteile der vorgeschlagenen Lösung

### Vorteile

1. **Einfachere Queries**: Balance, unbezahlte und ungelieferte Positionen sind direkte Datenbankabfragen ohne Event-Replay. Latenz sinkt von O(n) auf O(1).

2. **Snapshot-Mechanismus entfällt**: Kein manuelles `CreateTableSnapshot()`, keine Snapshot-Events, keine Erkennung von Snapshot-Events in der Rekonstruktionslogik. Der Code wird schlanker.

3. **Starke Konsistenz**: Durch die synchrone Projektion in derselben Transaktion ist der Read Store immer konsistent mit dem Write Store — kein Window für inkonsistente Lesezugriffe.

4. **Query-Seite einfacher verständlich**: Neue Entwickler können die `GetTableBalance`-Query verstehen, ohne das Event-Sourcing-Modell zu kennen.

5. **Erweiterbar für Analytics**: Zusätzliche Projektionen (z.B. `variant_sales` für Tagesabrechnung) können nach demselben Muster hinzugefügt werden.

6. **Rückwärtskompatibel**: Der Event Store bleibt unverändert. Die Projektion ist ein zusätzliches, sekundäres Modell.

### Nachteile

1. **Zusätzliche Komplexität beim Schreiben**: Jedes Command erfordert nun eine Datenbanktransaktion (Event-Insert + Projektion-Upsert). Bisher war ein einzelnes `INSERT` ausreichend.

2. **Zusätzliche Abhängigkeit**: Der Command-Service hängt nun von `TableStateRepo` ab — eine zusätzliche Schicht. Bei Fehlern in der Projektion-Logik schlagen Commands fehl (obwohl das Event valide gewesen wäre).

3. **Backfill erforderlich**: Für bestehende Daten muss die Projektion einmalig befüllt werden — ein Deployment-Schritt mehr.

4. **JSONB für Variants-Listen bleibt**: Die Listen `unpaid_variants` und `undelivered_variants` in der `table_state`-Tabelle bleiben JSONB. Typisierte separate Tabellen wären möglich, erhöhen aber die Komplexität.

5. **Projektion kann desynchronisieren**: Bei Bugs in `ApplyEventToState()` kann die Projektion vom tatsächlichen Event-Zustand abweichen. Ein Konsistenz-Check (z.B. periodischer Vergleich von `last_event_id` mit dem Event Store) wäre sinnvoll.

### Vergleichstabelle: Vorher / Nachher

| Aspekt | Aktuell (CQRS Stufe 1) | Vorgeschlagen (CQRS Stufe 2) |
|--------|------------------------|------------------------------|
| Balance-Query | Events laden + iterieren O(n) | Direkter DB-Zugriff O(1) |
| Unpaid/Undelivered-Query | Events laden + rekonstruieren O(n) | Direkter DB-Zugriff O(1) |
| History-Query | Events laden + transformieren | Events laden + transformieren (unverändert) |
| Schreiboperationen | Single INSERT | Transaktion: INSERT + UPSERT |
| Snapshot | Manuell auslösen, Events schreiben | Entfällt |
| Konsistenzmodell | Eventual (Snapshot manuell) | Strong (synchrone Projektion) |
| Neue Entwickler | Müssen ES verstehen | Query-Seite unabhängig von ES |

---

## 6. Zusammenfassung und Empfehlung

### Zusammenfassung

jotti implementiert CQRS bereits auf der logischen Ebene: Die Application-Schicht trennt `Command`- und `Query`-Structs konsequent, HTTP-Handler sind aufgeteilt, und Repository-Interfaces sind nach Lese-/Schreibzugriff getrennt. Diese Trennung ist architektonisch korrekt und wertvoll.

Was fehlt, ist die **Nutzung eines separaten Read Models**. Aktuell lesen sowohl Commands als auch Queries aus demselben Event Store — die Query-Seite rekonstruiert den Zustand jedes Mal aus Events. Das ist konzeptuell sauber (keine Redundanz, Quelle der Wahrheit ist der Event Store), aber suboptimal in Bezug auf Leseleistung und Wartbarkeit.

Die vorgeschlagene Erweiterung — eine **synchrone Projektionstabelle `table_state`** — ist die kleinstmögliche Ergänzung, die den größten Nutzen bringt:

- Sie ersetzt den unnatürlichen Snapshot-as-Event-Mechanismus durch ein sauberes, immer-aktuelles Read Model.
- Sie macht Balance- und Positions-Abfragen trivial.
- Sie bewahrt den vollständigen Event Log als unveränderliche Quelle der Wahrheit.
- Sie hält die Schreiboperationen transaktional konsistent.

### Empfehlung

**Die Implementierung der synchronen Projektion (`table_state`) wird empfohlen** — aber mit Prioritätsstufe **mittel**, nicht hoch. Begründung:

1. **Kein unmittelbarer Performance-Engpass**: In der aktuellen Praxis sind Tische klein (wenige Dutzend Events pro Schicht). Der Snapshot-Mechanismus adressiert das Performance-Problem ausreichend. Die Migration ist sinnvoll, aber nicht dringend.

2. **Klarer Architektur-Gewinn**: Der Snapshot-as-Event-Ansatz ist konzeptuell fragwürdig — ein Snapshot ist kein Domain-Event. Die synchrone Projektion ist konzeptuell sauberer und führt langfristig zu besser wartbarem Code.

3. **Enabler für zukünftige Features**: Analytische Abfragen (Tagesabrechnung nach Benutzer, Umsatz pro Produkt) werden mit typisierten Projektionen erheblich einfacher implementierbar. Die `table_state`-Projektion ist der erste Schritt in diese Richtung.

4. **Bewusstes Maßhalten**: Das Muster bleibt auf Stufe 2 (*Separate Read Models im selben Store*). Eine Trennung in separate Datenspeicher oder asynchrone Projektionen ist für jottis Größe nicht notwendig und würde mehr Komplexität einführen als lösen — wie Fowler warnt.

**Nicht empfohlen** wird:
- Asynchrone Projektionen (Eventual Consistency ist für ein Kassensystem nicht akzeptabel)
- Separate Read-Datenbank (overhead nicht gerechtfertigt)
- Vollständige Auflösung des Event Store zugunsten von CRUD (verliert Audit-Trail-Garantien — siehe [no-event-sourcing.md](no-event-sourcing.md))

### Fazit

CQRS ist in jotti bereits als Denkmodell verankert — die Benennung `Command`/`Query`, die Trennung der Structs und Handler zeigen das. Der nächste sinnvolle Schritt ist, dieses Muster mit einer konkreten Projektion zu vervollständigen und damit die bekannten Schwächen des aktuellen Ansatzes (Snapshot-Mechanismus, Lese-Komplexität) elegant zu lösen. Die Implementierung ist überschaubar, rückwärtskompatibel und bringt einen messbaren Gewinn in Sauberkeit und Wartbarkeit.

---

## 7. Referenzen

- **Martin Fowler**: [CQRS](https://martinfowler.com/bliki/CQRS.html) — Grundlegende Beschreibung des Musters, Warnung vor übermäßigem Einsatz
- **Martin Fowler**: [CommandQuerySeparation](https://martinfowler.com/bliki/CommandQuerySeparation.html) — Das Grundprinzip hinter CQRS
- **Greg Young**: [CQRS Documents](https://cqrs.wordpress.com/wp-content/uploads/2010/11/cqrs_documents.pdf) (PDF) — Originals Dokument, das CQRS popularisiert hat; Verbindung mit Event Sourcing
- **Udi Dahan**: [Clarified CQRS](https://udidahan.com/2009/12/09/clarified-cqrs/) — Frühe, einflussreiche Beschreibung
- **Wikipedia**: [Command Query Responsibility Segregation](https://en.wikipedia.org/wiki/Command_Query_Responsibility_Segregation) — Übersicht und Geschichte
- **Microsoft**: [CQRS pattern (Azure Architecture)](https://learn.microsoft.com/en-us/azure/architecture/patterns/cqrs) — Praxisnahe Beschreibung mit Varianten
- **Bertrand Meyer**: *Object-Oriented Software Construction* (1988) — Ursprung von CQS
- **jotti intern**: [Event-Sourcing vs. CRUD](event-sourcing-vs-crud.md) — Hybride Architektur und Event-Sourcing-Details
- **jotti intern**: [jotti ohne Event-Sourcing](no-event-sourcing.md) — CRUD-Alternative und Nachteile-Analyse
