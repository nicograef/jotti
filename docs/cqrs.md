# CQRS in jotti

Dieses Dokument beschreibt das Architekturmuster **Command Query Responsibility Segregation (CQRS)**, analysiert dessen aktuelle Nutzung in jotti und zeigt, wie eine vollständige CQRS-Implementierung die bekannten Schwächen des Event-Sourcing-Ansatzes in jotti mindern kann.

> **Verwandte Dokumente:**
> - [Event-Sourcing vs. CRUD](event-sourcing-vs-crud.md) — Hybride Architektur und Event-Sourcing-Implementierung
> - [jotti ohne Event-Sourcing](no-event-sourcing.md) — CRUD-Alternative und Vergleich
> - [ADR: Datenbankzugriff — Entscheidung für sqlc](adr/orm.md) — sqlc als Persistenz-Werkzeug

---

## Inhaltsverzeichnis

1. [CQRS — Theorie und Ursprung](#1-cqrs--theorie-und-ursprung)
2. [CQRS in jotti — Ist-Zustand](#2-cqrs-in-jotti--ist-zustand)
3. [Die Schwächen von Event-Sourcing und wie CQRS hilft](#3-die-schwächen-von-event-sourcing-und-wie-cqrs-hilft)
4. [Implementierungsplan: Synchrone Projektion (einfacher Ansatz)](#4-implementierungsplan-synchrone-projektion-einfacher-ansatz)
5. [Vor- und Nachteile der synchronen Projektion](#5-vor--und-nachteile-der-synchronen-projektion)
6. [Fortgeschrittene Alternativen: Ohne Snapshots und ohne C→Q-Abhängigkeit](#6-fortgeschrittene-alternativen-ohne-snapshots-und-ohne-cq-abhängigkeit)
7. [Domain-bezogene Konsistenz-Anforderungen: Kernfunktionalität vs. Zusatzfeatures](#7-domain-bezogene-konsistenz-anforderungen-kernfunktionalität-vs-zusatzfeatures)
8. [Zusammenfassung und Empfehlung](#8-zusammenfassung-und-empfehlung)
9. [Referenzen](#9-referenzen)

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

## 4. Implementierungsplan: Synchrone Projektion (einfacher Ansatz)

Der Plan beschreibt die Migration von der aktuellen logischen CQRS-Trennung zu einem vollständigen CQRS mit **synchronen Projektionen** als Read Model. Der Ansatz ist bewusst inkrementell und vermeidet einen Big-Bang-Umbau.

> **Hinweis:** Seit der Erstellung dieses Dokuments wurde [sqlc als Persistenz-Werkzeug übernommen](adr/orm.md). Die folgenden Code-Beispiele verwenden daher die sqlc-basierte Repository-Architektur: SQL-Queries werden in `.sql`-Dateien definiert, `sqlc generate` erzeugt typsichere Go-Funktionen, und Repositories wrappen diese generierten Funktionen. Details zur sqlc-Architektur siehe [ADR: Datenbankzugriff](adr/orm.md) und [Datenbank & Persistenz](database.md).

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

#### SQL-Queries (`sqlc/queries/table_state.sql`)

Neue sqlc-Query-Datei für die Projektions-Tabelle:

```sql
-- name: UpsertTableState :exec
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
    updated_at           = now();

-- name: GetTableState :one
SELECT table_id, balance_cents, total_payments_cents,
       unpaid_variants, undelivered_variants, last_event_id
FROM table_state
WHERE table_id = $1;
```

Nach `sqlc generate` werden typsichere Go-Funktionen generiert (`dbgen.UpsertTableState`, `dbgen.GetTableState`).

#### Repository-Wrapper (`repository/table_state_repo/`)

Neues Package `backend/repository/table_state_repo/`:

```go
package table_state_repo

import (
    "context"
    "database/sql"
    "encoding/json"

    "github.com/nicograef/jotti/backend/db"
    "github.com/nicograef/jotti/backend/domain/table"
    "github.com/nicograef/jotti/backend/sqlc/dbgen"
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
    q  *dbgen.Queries
}

func NewRepository(sqlDB *sql.DB) Repository {
    return Repository{DB: sqlDB, q: dbgen.New(sqlDB)}
}

// UpsertTx schreibt oder aktualisiert den Zustand eines Tisches innerhalb
// einer bestehenden Transaktion. Wird in derselben Transaktion wie
// WriteEvent aufgerufen.
func (r Repository) UpsertTx(ctx context.Context, tx *sql.Tx, state TableState) error {
    unpaidJSON, err := json.Marshal(state.UnpaidVariants)
    if err != nil {
        return err
    }
    undeliveredJSON, err := json.Marshal(state.UndeliveredVariants)
    if err != nil {
        return err
    }

    qtx := r.q.WithTx(tx)
    return qtx.UpsertTableState(ctx, dbgen.UpsertTableStateParams{
        TableID:             state.TableID,
        BalanceCents:        state.BalanceCents,
        TotalPaymentsCents:  state.TotalPaymentsCents,
        UnpaidVariants:      unpaidJSON,
        UndeliveredVariants: undeliveredJSON,
        LastEventID:         state.LastEventID,
    })
}

// Get liest den aktuellen Zustand eines Tisches aus der Projektionstabelle.
func (r Repository) Get(ctx context.Context, tableID int) (TableState, error) {
    row, err := r.q.GetTableState(ctx, tableID)
    if err != nil {
        return TableState{}, db.Error(err)
    }

    var state TableState
    state.TableID = row.TableID
    state.BalanceCents = row.BalanceCents
    state.TotalPaymentsCents = row.TotalPaymentsCents
    state.LastEventID = row.LastEventID

    if err := json.Unmarshal(row.UnpaidVariants, &state.UnpaidVariants); err != nil {
        return TableState{}, err
    }
    if err := json.Unmarshal(row.UndeliveredVariants, &state.UndeliveredVariants); err != nil {
        return TableState{}, err
    }

    return state, nil
}

// Upsert schreibt oder aktualisiert den Zustand eines Tisches nicht-transaktional.
// Wird für die Lazy Projection (Alternative B, Abschnitt 6.3) verwendet, bei der
// das Projektions-Update außerhalb einer Transaktion erfolgt und bei Fehlern
// ignoriert werden kann (Self-Healing beim nächsten Read).
func (r Repository) Upsert(ctx context.Context, state TableState) error {
    unpaidJSON, err := json.Marshal(state.UnpaidVariants)
    if err != nil {
        return err
    }
    undeliveredJSON, err := json.Marshal(state.UndeliveredVariants)
    if err != nil {
        return err
    }

    return r.q.UpsertTableState(ctx, dbgen.UpsertTableStateParams{
        TableID:             state.TableID,
        BalanceCents:        state.BalanceCents,
        TotalPaymentsCents:  state.TotalPaymentsCents,
        UnpaidVariants:      unpaidJSON,
        UndeliveredVariants: undeliveredJSON,
        LastEventID:         state.LastEventID,
    })
}
```

Das Repository folgt dem bestehenden sqlc-Pattern: sqlc generiert die typsicheren Query-Funktionen in `dbgen/`, das Repository-Package wrappt diese und bildet sie auf Domain-Typen ab (analog zu `event_repo`, `table_repo`, etc. — siehe [Datenbank & Persistenz](database.md)).

### 4.3 Schritt 3 — Event Store mit Transaktionsunterstützung

Das bestehende `event_repo` muss eine **transaktionale Variante** von `WriteEvent` erhalten, die es ermöglicht, Event-Insert und Projektion-Upsert atomar auszuführen.

In `backend/repository/event_repo/repo.go` zwei neue Methoden ergänzen:

```go
// WriteEventTx schreibt ein Event innerhalb einer bestehenden Transaktion.
// Damit kann der Aufrufer Event-Insert und Projektion-Upsert atomar ausführen.
// Nutzt sqlcs WithTx()-Methode, um die generierten Queries transaktional auszuführen.
func (r Repository) WriteEventTx(ctx context.Context, tx *sql.Tx, e event.Event) (int, error) {
    qtx := r.q.WithTx(tx)
    id, err := qtx.WriteEvent(ctx, dbgen.WriteEventParams{
        UserID:    e.UserID,
        Type:      e.Type,
        Subject:   e.Subject,
        Data:      e.Data,
        Timestamp: e.Time,
    })
    if err != nil {
        return 0, db.Error(err)
    }
    return id, nil
}

// BeginTx startet eine neue Datenbanktransaktion.
func (r Repository) BeginTx(ctx context.Context) (*sql.Tx, error) {
    return r.DB.BeginTx(ctx, nil)
}
```

Die `WithTx()`-Methode wird von sqlc automatisch generiert und erlaubt es, dasselbe `dbgen.Queries`-Struct mit einer bestehenden Transaktion zu verwenden. So wird die bestehende `WriteEvent`-SQL-Query (definiert in `sqlc/queries/events.sql`) wiederverwendet — kein dupliziertes SQL nötig.

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
    UpsertTx(ctx context.Context, tx *sql.Tx, state table_state_repo.TableState) error
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
    if err := c.TableStateRepo.UpsertTx(ctx, tx, newState); err != nil {
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
| sqlc | Query-Datei `sqlc/queries/table_state.sql` + `sqlc generate` | Gering |
| Repository | Neues `table_state_repo` (wrappt sqlc-generierte Funktionen) | Mittel |
| Repository | `event_repo` um `WriteEventTx`/`BeginTx` erweitern (nutzt `q.WithTx()`) | Gering |
| Domain | `ApplyEventToState()` + `ProjectionState` | Mittel |
| Application/Command | `writeEventAndUpdateProjection()` | Mittel |
| Application/Query | Auf `table_state_repo` umstellen | Gering |
| Application/Command | `CreateTableSnapshot()` entfernen | Gering |
| Backfill | Einmaliges Migrationsskript | Mittel |
| Tests | Unit-Tests für `ApplyEventToState()` | Mittel |
| **Gesamt** | | **Mittel** |

---

## 5. Vor- und Nachteile der synchronen Projektion

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

## 6. Fortgeschrittene Alternativen: Ohne Snapshots und ohne C→Q-Abhängigkeit

Die in Abschnitt 4 vorgeschlagene **synchrone Projektion** löst die Lese-Probleme elegant, bringt aber drei Nachteile auf der Schreibseite mit sich:

1. **Zusätzliche Komplexität beim Schreiben**: Jedes Command erfordert nun eine Datenbanktransaktion (Event-Insert + Projektion-Upsert). Bisher war ein einzelnes `INSERT` ausreichend.
2. **Zusätzliche Abhängigkeit (C→Q)**: Der Command-Service hängt von `TableStateRepo` ab — eine zusätzliche Schicht. Bei Fehlern in der Projektion-Logik schlagen Commands fehl, obwohl das Event valide gewesen wäre. Dies verletzt das CQRS-Prinzip der strikten Trennung: die Write-Seite kennt und beeinflusst die Read-Seite.
3. **Backfill erforderlich**: Für bestehende Daten muss die Projektion einmalig befüllt werden — ein Deployment-Schritt mehr.

Dieser Abschnitt untersucht alternative Ansätze, die **Snapshots vollständig eliminieren** und gleichzeitig die **C→Q-Abhängigkeit vermeiden** — d.h. Commands bleiben einfache Event-INSERTs ohne Wissen über das Read Model.

### 6.1 Problemanalyse: Warum entsteht die C→Q-Abhängigkeit?

Die C→Q-Abhängigkeit in Abschnitt 4 entsteht, weil die Projektion **synchron im Command-Pfad** aktualisiert wird:

```
Command → BEGIN TX → INSERT event → READ projection → APPLY event → UPSERT projection → COMMIT
```

Der Command-Service muss dafür:
- Das `TableStateRepo` kennen (zusätzliches Interface)
- Die `ApplyEventToState()`-Logik aufrufen (Domain-Logik im Write-Pfad)
- Bei Projektionsfehlern die gesamte Transaktion rollbacken (Write scheitert wegen Read-Logik)

Dies ist architektonisch problematisch, denn in einem sauberen CQRS-Modell sollte die Write-Seite keine Kenntnis von der Read-Seite haben. Microsoft beschreibt dies im Azure Architecture Center:

> *"Commands update data. Queries retrieve data."* — Die Verantwortlichkeiten sollen getrennt bleiben.

In Event-Sourcing-Systemen ist der Event Store die Single Source of Truth. Projektionen (Read Models) sind **abgeleitete Daten** — sie können jederzeit aus dem Event Log rekonstruiert werden (vgl. Baytech Consulting: *"If a read model becomes corrupted, contains a bug, or needs to be changed, it can be safely deleted and completely rebuilt by replaying the event stream"*). Wenn die Projektion den Write-Pfad blockieren kann, wird diese Eigenschaft untergraben.

### 6.2 Alternative A: PostgreSQL-Trigger-basierte Projektion

#### Konzept

Die Projektionslogik wird **auf Datenbankebene** via PostgreSQL-Trigger implementiert. Ein `AFTER INSERT`-Trigger auf der `events`-Tabelle aktualisiert die `table_state`-Tabelle automatisch bei jedem neuen Event.

```
Command → INSERT event → [Trigger: UPDATE table_state] → Done
Query   → SELECT FROM table_state → Done
```

#### Implementierung

```sql
-- Trigger-Funktion: aktualisiert table_state nach jedem Event-Insert
CREATE OR REPLACE FUNCTION update_table_state_projection()
RETURNS TRIGGER AS $$
DECLARE
    v_table_id INT;
    v_data JSONB;
    v_variants JSONB;
    v_total_cents INT;
BEGIN
    -- Nur table-Events verarbeiten
    IF NEW.subject NOT LIKE 'table:%' THEN
        RETURN NEW;
    END IF;

    v_table_id := CAST(split_part(NEW.subject, ':', 2) AS INT);
    v_data := NEW.data;

    -- Initialen Zustand sicherstellen
    INSERT INTO table_state (table_id, last_event_id)
    VALUES (v_table_id, 0)
    ON CONFLICT (table_id) DO NOTHING;

    CASE NEW.type
        WHEN 'table.order-placed:v1' THEN
            v_total_cents := (v_data->>'totalPriceCents')::INT;
            UPDATE table_state SET
                balance_cents = balance_cents + v_total_cents,
                unpaid_variants = unpaid_variants || (v_data->'variants'),
                undelivered_variants = undelivered_variants || (v_data->'variants'),
                last_event_id = NEW.id,
                updated_at = now()
            WHERE table_id = v_table_id;

        WHEN 'table.payment-registered:v1' THEN
            v_total_cents := (v_data->>'totalPaymentCents')::INT;
            UPDATE table_state SET
                balance_cents = balance_cents - v_total_cents,
                total_payments_cents = total_payments_cents + v_total_cents,
                last_event_id = NEW.id,
                updated_at = now()
            WHERE table_id = v_table_id;
            -- Hinweis: Die Varianten-Reduktion (reduceVariants) — also das
            -- Entfernen bezahlter Varianten aus dem JSONB-Array unpaid_variants —
            -- erfordert in PL/pgSQL elementweisen Vergleich und Manipulation von
            -- JSONB-Arrays (jsonb_array_elements, jsonb_agg mit Filterung).
            -- In Go ist dies eine einfache Slice-Operation, in SQL deutlich komplexer.

        WHEN 'table.variants-canceled:v1' THEN
            v_total_cents := (v_data->>'totalCancelationCents')::INT;
            UPDATE table_state SET
                balance_cents = balance_cents - v_total_cents,
                last_event_id = NEW.id,
                updated_at = now()
            WHERE table_id = v_table_id;

        WHEN 'table.variants-delivered:v1' THEN
            UPDATE table_state SET
                last_event_id = NEW.id,
                updated_at = now()
            WHERE table_id = v_table_id;

        ELSE NULL; -- Unbekannte Events ignorieren (z.B. Snapshots)
    END CASE;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_table_state
    AFTER INSERT ON events
    FOR EACH ROW
    EXECUTE FUNCTION update_table_state_projection();
```

#### Bewertung

| Kriterium | Bewertung |
|-----------|-----------|
| Command-Einfachheit | ✅ Commands bleiben einfache Event-INSERTs — kein Go-Code-Änderung |
| C→Q-Trennung | ✅ Command-Service hat keine Kenntnis der Projektion |
| Konsistenz | ✅ Atomar — Trigger läuft in derselben Transaktion wie INSERT |
| Snapshot-Eliminierung | ✅ Projektion ersetzt Snapshots vollständig |
| Backfill | ⚠️ Einmalig nötig für bestehende Daten (oder Trigger manuell für historische Events triggern) |
| Wartbarkeit | ❌ Business-Logik in PL/pgSQL — schwerer zu testen, debuggen und refactoren als Go-Code |
| JSONB-Manipulation | ❌ Komplexe Varianten-Reduktion (`reduceVariants`) in PL/pgSQL ist aufwendig und fehleranfällig |
| Testbarkeit | ❌ Keine Unit-Tests im Go-Sinne; Integrationstests auf DB-Ebene nötig |

**Fazit**: Für die einfachen Aggregationen (Balance, Summen) funktioniert der Trigger-Ansatz gut. Die **Varianten-Listen-Manipulation** (Elemente aus JSONB-Arrays subtrahieren) ist in PL/pgSQL jedoch deutlich komplexer als in Go und schwer testbar. Für jottis spezifische Anforderungen (varianten-basierte Akkumulation und Reduktion) ist dieser Ansatz **bedingt geeignet**.

### 6.3 Alternative B: Query-seitige Lazy Projection (Read-Through-Cache)

#### Konzept

Die Projektion wird **nicht beim Schreiben** aktualisiert, sondern **beim Lesen** — als Read-Through-Cache. Wenn eine Query den Tischzustand anfragt, prüft die Query-Seite, ob die gespeicherte Projektion aktuell ist (anhand `last_event_id`). Falls nicht, werden nur die fehlenden Events nachgespielt und die Projektion aktualisiert.

```
Command → INSERT event → Done (kein Projektions-Update)
Query   → Ist Projektion aktuell?
            Ja  → SELECT FROM table_state → return
            Nein → Replay fehlende Events → UPDATE table_state → return
```

#### Implementierung

```go
// In query.go — Lazy Projection Pattern
func (q Query) GetTableBalance(ctx context.Context, tableID int) (int, error) {
    state, err := q.ensureProjectionUpToDate(ctx, tableID)
    if err != nil {
        return 0, err
    }
    return state.BalanceCents, nil
}

func (q Query) ensureProjectionUpToDate(ctx context.Context, tableID int) (TableState, error) {
    // 1. Aktuellen Projektion-Zustand lesen (falls vorhanden)
    currentState, err := q.TableStateRepo.Get(ctx, tableID)
    if err != nil && !errors.Is(err, db.ErrNotFound) {
        return TableState{}, ErrDatabase
    }

    // 2. Prüfen, ob neue Events existieren
    subject := "table:" + strconv.Itoa(tableID)
    var events []event.Event
    if currentState.LastEventID > 0 {
        // Nur Events seit dem letzten projizierten Event lesen
        events, err = q.EventRepo.ReadEventsSinceID(ctx, subject, currentState.LastEventID+1)
    } else {
        // Kein Projektion-Eintrag vorhanden: alle Events lesen
        events, err = q.EventRepo.ReadEventsBySubject(ctx, subject)
    }
    if err != nil {
        return TableState{}, ErrDatabase
    }

    // 3. Wenn keine neuen Events: Projektion ist aktuell
    if len(events) == 0 {
        return currentState, nil
    }

    // 4. Fehlende Events auf den Zustand anwenden
    projState := table.ProjectionState{
        BalanceCents:        currentState.BalanceCents,
        TotalPaymentsCents:  currentState.TotalPaymentsCents,
        UnpaidVariants:      currentState.UnpaidVariants,
        UndeliveredVariants: currentState.UndeliveredVariants,
    }
    for _, evt := range events {
        projState, err = table.ApplyEventToState(projState, evt)
        if err != nil {
            return TableState{}, err
        }
    }

    // 5. Projektion aktualisieren (für nachfolgende Queries)
    newState := TableState{
        TableID:             tableID,
        BalanceCents:        projState.BalanceCents,
        TotalPaymentsCents:  projState.TotalPaymentsCents,
        UnpaidVariants:      projState.UnpaidVariants,
        UndeliveredVariants: projState.UndeliveredVariants,
        LastEventID:         events[len(events)-1].ID,
    }
    // Nicht-transaktional: Projektions-Update ist optional.
    // Bei Fehler wird die Projektion beim nächsten Read erneut berechnet.
    _ = q.TableStateRepo.Upsert(ctx, newState)

    return newState, nil
}
```

#### Bewertung

| Kriterium | Bewertung |
|-----------|-----------|
| Command-Einfachheit | ✅ Commands bleiben unverändert — reines Event-INSERT |
| C→Q-Trennung | ✅ Command-Service hat keine Kenntnis der Projektion |
| Konsistenz | ✅ Stark konsistent — Leser sehen immer den aktuellen Zustand (Replay on Read) |
| Snapshot-Eliminierung | ✅ Projektion ersetzt Snapshots vollständig |
| Backfill | ✅ **Nicht nötig** — Lazy Projection befüllt sich beim ersten Lesezugriff selbst |
| Wartbarkeit | ✅ Gesamte Projektionslogik in Go — testbar mit Unit-Tests |
| Selbstheilend | ✅ Fehlerhafte Projektion wird beim nächsten Read automatisch korrigiert |
| Latenz beim Lesen | ⚠️ Erster Lesezugriff nach vielen Writes ist langsamer (einmaliger Replay) |
| Concurrent Writes | ⚠️ Bei gleichzeitigen Reads auf denselben Tisch kann es zu Race Conditions beim Projektions-Update kommen — lösbar durch `SELECT ... FOR UPDATE` oder optimistisches Locking via `last_event_id` |

**Fazit**: Die Lazy Projection löst **alle drei genannten Nachteile** der synchronen Projektion:
- ✅ Keine zusätzliche Schreibkomplexität (Commands bleiben ein INSERT)
- ✅ Keine C→Q-Abhängigkeit (Command-Service kennt die Projektion nicht)
- ✅ Kein Backfill nötig (Projektion befüllt sich selbst beim ersten Read)

Der Trade-off ist eine potenzielle Latenz beim ersten Lesezugriff, die aber in der Praxis gering ist: Tische in jotti haben typischerweise wenige Dutzend Events pro Schicht, und das Replay weniger Events ist in Go extrem schnell.

### 6.4 Alternative C: Asynchroner Background Worker (Polling)

#### Konzept

Ein Hintergrundprozess (Goroutine) pollt periodisch die `events`-Tabelle auf neue Events und aktualisiert die Projektion asynchron. Commands und Queries sind vollständig entkoppelt.

```
Command → INSERT event → Done
                              ↓ (asynchron, z.B. alle 100ms)
Worker  → Poll neue Events → UPDATE table_state
Query   → SELECT FROM table_state → Done
```

#### Implementierung (Skizze)

```go
// projector/worker.go
func StartProjectionWorker(ctx context.Context, eventRepo EventRepo,
    stateRepo TableStateRepo, interval time.Duration) {

    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := projectNewEvents(ctx, eventRepo, stateRepo); err != nil {
                log.Error().Err(err).Msg("Projection worker error")
            }
        }
    }
}

func projectNewEvents(ctx context.Context, eventRepo EventRepo,
    stateRepo TableStateRepo) error {

    // 1. Höchste projizierte Event-ID ermitteln
    lastProjectedID, _ := stateRepo.GetGlobalLastEventID(ctx)

    // 2. Alle neuen Events seit letzter Projektion lesen
    events, err := eventRepo.ReadEventsSinceGlobalID(ctx, lastProjectedID+1)
    if err != nil || len(events) == 0 {
        return err
    }

    // 3. Events nach Tisch gruppieren und Projektionen aktualisieren
    for tableID, tableEvents := range groupByTable(events) {
        state, _ := stateRepo.Get(ctx, tableID)
        for _, evt := range tableEvents {
            state = table.ApplyEventToState(state, evt)
        }
        stateRepo.Upsert(ctx, state)
    }

    return nil
}
```

#### Bewertung

| Kriterium | Bewertung |
|-----------|-----------|
| Command-Einfachheit | ✅ Commands bleiben unverändert — reines Event-INSERT |
| C→Q-Trennung | ✅ Vollständige Entkopplung — Command kennt weder Projektion noch Worker |
| Konsistenz | ❌ **Eventual Consistency** — kurzes Zeitfenster (~100ms), in dem Leser veraltete Daten sehen |
| Snapshot-Eliminierung | ✅ Projektion ersetzt Snapshots vollständig |
| Backfill | ⚠️ Worker kann bestehende Events beim Start nachprojizieren — aber initialer Durchlauf nötig |
| Wartbarkeit | ✅ Projektionslogik in Go — testbar |
| Infrastruktur | ❌ Zusätzliche Goroutine, Lifecycle-Management, Health-Checks, Fehlerbehandlung bei Crashes |
| Skalierbarkeit | ⚠️ Einzelner Worker — bei mehreren Instanzen braucht man Leader Election oder Partitionierung |

**Fazit**: Der Background Worker bietet die **sauberste CQRS-Trennung**, hat aber einen gravierenden Nachteil für **operative Queries** (Balance, offene Positionen) im Kassenbetrieb: **Eventual Consistency**. In einem Kassensystem muss die Servicekraft unmittelbar nach dem Aufgeben einer Bestellung die korrekte Balance sehen. Ein Fenster von 100ms mag technisch akzeptabel erscheinen, ist aber architektonisch riskant — unter Last, bei Fehlern im Worker oder bei Neustarts kann dieses Fenster wachsen. Microsoft und Confluent betonen diese Problematik:

> *"When the read databases and write databases are separated, the read data might not show the most recent changes immediately."* — Microsoft Azure Architecture Center

Für **analytische Projektionen** (Umsatzauswertungen, Produktstatistiken, Tagesabrechnung) ist der Background Worker hingegen die geeignete Wahl: Eventual Consistency ist für retrospektive Auswertungen akzeptabel, und der Worker bietet maximale Entkopplung ohne Einfluss auf den Write-Pfad. Abschnitt 7 beschreibt diese Unterscheidung im Detail.

### 6.5 Vergleich der Alternativen

| Kriterium | Synchrone Projektion (Abschnitt 4) | A: DB-Trigger | B: Lazy Projection | C: Background Worker |
|-----------|--------------------------------------|---------------|---------------------|-----------------------|
| Command bleibt einfach (single INSERT) | ❌ Transaktion nötig | ✅ | ✅ | ✅ |
| Keine C→Q-Abhängigkeit | ❌ Command kennt Read Model | ✅ | ✅ | ✅ |
| Kein Backfill nötig | ❌ Einmaliger Backfill | ❌ Einmaliger Backfill | ✅ Self-healing | ⚠️ Worker-Initialisierung |
| Starke Konsistenz | ✅ Gleiche Transaktion | ✅ Gleiche Transaktion | ✅ Read-Time-Konsistenz | ❌ Eventual Consistency¹ |
| Snapshot-Eliminierung | ✅ | ✅ | ✅ | ✅ |
| Wartbarkeit / Testbarkeit | ✅ Go-Code | ❌ PL/pgSQL | ✅ Go-Code | ✅ Go-Code |
| JSONB-Varianten-Manipulation | ✅ Go-Logik | ❌ Komplex in SQL | ✅ Go-Logik | ✅ Go-Logik |
| Infrastruktur-Overhead | Gering | Gering | Gering | Mittel (Worker-Lifecycle) |
| Implementierungsaufwand | Mittel | Mittel | **Gering** | Hoch |

> ¹ Eventual Consistency ist für **operative Queries** (Balance, offene Positionen) nicht akzeptabel. Für **analytische Projektionen** (Umsatzauswertungen, Tagesabrechnung) ist sie hingegen ausreichend — der Background Worker ist dort die empfohlene Variante (siehe Abschnitt 7).

### 6.6 Empfehlung für jotti

Für jottis spezifischen Kontext — ein Kassensystem mit wenigen gleichzeitigen Benutzern, PostgreSQL als einziger Datenspeicher, Go als Backend-Sprache — ist **Alternative B (Lazy Projection)** die attraktivste Wahl für **operative Projektionen** (Balance, offene Positionen), wenn die Nachteile der synchronen Projektion vermieden werden sollen:

1. **Commands bleiben einfach**: Ein `INSERT INTO events` — kein transaktionaler Overhead, keine neue Abhängigkeit. Der Code in `command.go` bleibt unverändert.
2. **Keine C→Q-Abhängigkeit**: Der Command-Service kennt weder die `table_state`-Tabelle noch das `TableStateRepo`. Die CQRS-Trennung bleibt rein.
3. **Kein Backfill nötig**: Die Projektion befüllt sich selbst beim ersten Lesezugriff. Kein separater Migrations- oder Deployment-Schritt.
4. **Starke Konsistenz**: Da die Projektion beim Lesen aktualisiert wird, sieht der Leser immer den aktuellen Zustand — kein Eventual-Consistency-Problem.
5. **Selbstheilend**: Fehlerhafte Projektionen werden beim nächsten Read automatisch korrigiert — robuster als die synchrone Variante, bei der ein Projektionsfehler den Write-Pfad blockiert.

Der einzige Trade-off — eine leicht erhöhte Latenz beim ersten Read nach mehreren Writes — ist für jottis Lastprofil (wenige Events pro Tisch, wenige Tische gleichzeitig) vernachlässigbar.

**Empfohlene Reihenfolge der Umsetzung (falls Lazy Projection gewählt wird):**

1. Migration: `table_state`-Tabelle anlegen (identisch zu Abschnitt 4.1)
2. sqlc-Queries für `table_state` definieren (identisch zu Abschnitt 4.2)
3. `table_state_repo` implementieren — mit einer nicht-transaktionalen `Upsert`-Methode (ohne `*sql.Tx`)
4. `ApplyEventToState()` in der Domain-Schicht implementieren (identisch zu Abschnitt 4.4)
5. `Query`-Struct um `TableStateRepo` erweitern, `ensureProjectionUpToDate()` implementieren
6. `CreateTableSnapshot()` und Snapshot-Logik entfernen
7. Tests: Unit-Tests für `ApplyEventToState()`, Integration-Tests für Lazy Projection

**Kein `command.go`-Refactoring nötig** — Commands bleiben unverändert.

> **Hinweis**: Für **analytische Projektionen** (Umsatzauswertungen, Tagesabrechnung) ist der Background Worker (Alternative C, Abschnitt 6.4) die empfohlene Variante. Dort ist Eventual Consistency akzeptabel. Diese Unterscheidung wird in Abschnitt 7 ausführlich behandelt.

---

## 7. Domain-bezogene Konsistenz-Anforderungen: Kernfunktionalität vs. Zusatzfeatures

Die bisherige Analyse bewertete alle CQRS-Varianten einheitlich anhand einer einzigen Konsistenz-Anforderung: Strong Consistency für das Kassensystem. Eine differenziertere Betrachtung der Domänen-Grenzen in jotti zeigt, dass nicht alle Projektion-Anwendungsfälle dieselbe Anforderung teilen — und damit unterschiedliche CQRS-Varianten geeignet sind.

### 7.1 Zwei Konsistenzklassen in jotti

| Bereich | Anwendungsfälle | Konsistenz-Anforderung |
|---------|-----------------|------------------------|
| **A) Kernfunktionalität** | Tischbasiertes Kassensystem: Bestellungen aufgeben, Zahlungen registrieren, Stornierungen, Lieferungen, Tischbalance, offene Positionen | **Strong Consistency** (zwingend) |
| **B) Zusatzfeatures** | Umsatzauswertungen, Tagesabrechnung, Lagerbestände, Produktstatistiken, Stornierungsraten | **Eventual Consistency** (ausreichend) |

Diese Unterscheidung entscheidet darüber, welche CQRS-Varianten für welchen Anwendungsfall geeignet sind — und erlaubt es, zuvor pauschal abgelehnte Ansätze für Bereich B neu zu bewerten.

### 7.2 Warum Strong Consistency für den Kassenbetrieb zwingend ist

Im Kassenbetrieb (Bereich A) sehen Servicekräfte unmittelbar nach jeder Aktion den aktualisierten Tischzustand:

- Nach dem Aufgeben einer Bestellung muss die **Balance sofort korrekt** angezeigt werden.
- Nach einer Zahlung müssen die **bezahlten Positionen sofort verschwinden**.
- Nach einer Stornierung müssen die **stornierten Artikel sofort als storniert** gelten.

Ein Verzögerungsfenster — selbst von wenigen Hundert Millisekunden — ist nicht akzeptabel:

1. **Doppelt-Bezahlt-Risiko**: Zeigt die Projektion noch die alte Balance, könnte die Servicekraft irrtümlich eine zweite Zahlung auslösen.
2. **Vertrauensverlust**: Servicekräfte auf mobilen Geräten müssen dem System vertrauen. Sichtbar veraltete Werte untergraben das Vertrauen unmittelbar.
3. **Keine natürliche Toleranz für Verzögerungen**: Im Gegensatz zu analytischen Dashboards ist ein POS-System ein Echtzeit-Werkzeug mit operativen Konsequenzen.

**Fazit**: Für Bereich A ist Strong Consistency obligatorisch — Lazy Projection (Abschnitt 6.3) und synchrone Projektion (Abschnitt 4) sind die richtigen Ansätze.

### 7.3 Warum Eventual Consistency für Zusatzfeatures ausreicht

Analytische Auswertungen (Bereich B) haben grundlegend andere Nutzungsmuster:

- Der Admin öffnet den Umsatz-Report am Abend oder zwischen Schichten — Daten, die sekunden- bis minutenalt sind, sind vollständig akzeptabel.
- Lagerbestands-Tracking reagiert auf Bestellungen — ein kurzes Verzögerungsfenster beeinflusst keine operativen Entscheidungen im Kassenbetrieb.
- Aggregierte Statistiken sind von Natur aus retrospektiv — eine leichte Verzögerung hat keinen Einfluss auf den laufenden Betrieb.

Für Bereich B können **asynchrone Projektionen** (Background Worker, Abschnitt 6.4) bedenkenlos eingesetzt werden. Das Hauptargument gegen diesen Ansatz in Abschnitt 6.4 — *„Eventual Consistency ist für ein Kassensystem nicht akzeptabel"* — gilt für analytische Abfragen nicht.

### 7.4 Bezug zur bestehenden Domain-Architektur von jotti

jotti unterscheidet bei der **Persistenz** bereits konsequent nach Domänen-Grenzen:

| Subdomain | Persistenz-Strategie | Begründung |
|-----------|---------------------|------------|
| **Core Domain**: Kassensystem (Tisch-Operationen) | Event-Sourcing (append-only `events`-Tabelle) | Audit-Trail, vollständige Historie, Replay |
| **Supporting/Generic Subdomains**: Auth, Usermanagement, Produktkatalog, Tisch-Stammdaten | CRUD (keine Historie, kein Audit-Log) | Einfachheit, direkte Abfragen |

Diese Unterscheidung lässt sich auf die **Lese-Seite** der Core Domain übertragen. Aus dem Event Store entstehen zwei Kategorien von CQRS-Projektionen mit unterschiedlichen Anforderungen:

| Projektion | Konsistenz-Anforderung | Geeignete CQRS-Variante |
|-----------|------------------------|------------------------|
| **Operative Projektion** — Balance, Unpaid, Undelivered je Tisch | Strong Consistency | Lazy Projection oder synchrone Projektion |
| **Analytische Projektion** — Tagesumsatz, Produktstatistiken, Stornierungsraten | Eventual Consistency | Asynchroner Background Worker |

### 7.5 Architektur mit zwei Projektionspfaden

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Frontend                                   │
└──────┬──────────────────────────────────────────┬───────────────────────┘
       │ Commands (Schreiben)                     │ Queries (Lesen)
       ▼                                          ▼
┌──────────────────┐         ┌────────────────────────────────────────────┐
│ Command Handler  │         │              Query Handler                 │
│ (schreibt Events)│         │  A) Operativ (Strong)  B) Analytisch (Ev.) │
└──────┬───────────┘         └───────────┬────────────────────┬───────────┘
       │ INSERT                          │ SELECT             │ SELECT
       ▼                                 ▼                    ▼
┌──────────────────┐   ┌────────────────────┐   ┌────────────────────────┐
│   events         │   │  table_state       │   │  daily_revenue         │
│   (append-only)  │   │  (operative        │   │  variant_sales         │
│                  │   │   Projektion,      │   │  (analytische          │
│                  │   │   Strong           │   │   Projektionen,        │
│                  │   │   Consistency)     │   │   Eventual             │
└──────────────────┘   └────────────────────┘   │   Consistency)         │
       │                        ▲               └────────────────────────┘
       │    Lazy Projection      │                          ▲
       │    (beim ersten Read)   │               ┌──────────────────────┐
       └────────────────────────┘               │  Background Worker   │
                                                │  (async, z.B. ~30s)  │
                                                └──────────────────────┘
```

### 7.6 Aktualisierte Empfehlung nach Konsistenz-Klasse

Diese Betrachtung ergänzt — nicht ersetzt — die Empfehlungen in Abschnitt 8. Sie präzisiert, welche CQRS-Variante für welche Projektion geeignet ist:

**Für operative Projektionen (Bereich A — Strong Consistency):**

- **Empfehlung: Lazy Projection** (Abschnitt 6.3) — Strong Consistency, Commands bleiben einfache Event-INSERTs, kein Backfill nötig, selbstheilend.
- Alternative: Synchrone Projektion (Abschnitt 4) — ebenfalls stark konsistent, aber mit C→Q-Abhängigkeit und Backfill-Aufwand.

**Für analytische Projektionen (Bereich B — Eventual Consistency):**

- **Empfehlung: Background Worker** (Abschnitt 6.4) — da Eventual Consistency akzeptabel ist, entfällt das Hauptargument gegen diesen Ansatz. Der Worker pollt die `events`-Tabelle, projiziert neue Events auf analytische Tabellen (z.B. `daily_revenue`, `variant_sales`) und aktualisiert diese asynchron.
- Ein Polling-Intervall von 30–60 Sekunden ist für Reports und Statistiken vollständig akzeptabel.
- Commands bleiben einfache Event-INSERTs — keine C→Q-Abhängigkeit, vollständige Entkopplung.
- **Kein separater Datenspeicher nötig**: Analytische Projektionstabellen liegen in derselben PostgreSQL-Datenbank — die Vorteile des Background Workers entstehen ohne Infrastruktur-Overhead eines separaten Read Stores.

### 7.7 Bezug zu DDD: Konsistenz-Grenzen an Domänen-Grenzen

Diese Zweiteilung entspricht einem etablierten DDD-Prinzip: **Konsistenz-Grenzen folgen Aggregat-Grenzen**. Der operative Tischzustand (Balance, offene Positionen) ist Teil des `Table`-Aggregats der Core Domain — er muss stark konsistent sein. Analytische Auswertungen aggregieren über Tisch-Grenzen hinweg und gehören konzeptuell zu einer separaten *Reporting-Subdomain*, für die Eventual Consistency eine natürliche Wahl ist.

> *„Strong consistency should be confined to within an Aggregate boundary. Between Aggregates and Bounded Contexts, eventual consistency is the norm."* — Vaughn Vernon, *Implementing Domain-Driven Design*

---

## 8. Zusammenfassung und Empfehlung

### Zusammenfassung

jotti implementiert CQRS bereits auf der logischen Ebene: Die Application-Schicht trennt `Command`- und `Query`-Structs konsequent, HTTP-Handler sind aufgeteilt, und Repository-Interfaces sind nach Lese-/Schreibzugriff getrennt. Diese Trennung ist architektonisch korrekt und wertvoll.

Was fehlt, ist die **Nutzung eines separaten Read Models**. Aktuell lesen sowohl Commands als auch Queries aus demselben Event Store — die Query-Seite rekonstruiert den Zustand jedes Mal aus Events. Das ist konzeptuell sauber (keine Redundanz, Quelle der Wahrheit ist der Event Store), aber suboptimal in Bezug auf Leseleistung und Wartbarkeit.

Dieses Dokument beschreibt zwei Wege, den Zustand als Projektion vorzuhalten und damit den Snapshot-Mechanismus abzulösen:

1. **Synchrone Projektion** (Abschnitt 4): Die Projektion wird im Command-Pfad innerhalb derselben Transaktion aktualisiert. Einfach zu verstehen und konsistent, aber mit erhöhter Schreibkomplexität und einer C→Q-Abhängigkeit.
2. **Lazy Projection** (Abschnitt 6.3): Die Projektion wird auf der Query-Seite als Read-Through-Cache aktualisiert. Commands bleiben unverändert, kein Backfill nötig, selbstheilend — aber leicht erhöhte Latenz beim ersten Lesezugriff.

Beide Ansätze nutzen die gleiche `table_state`-Tabelle und die gleiche `ApplyEventToState()`-Logik. Der Unterschied liegt nur darin, **wann und wo** die Projektion aktualisiert wird.

### Empfehlung

**Die Implementierung einer Projektionstabelle (`table_state`) wird empfohlen** — aber mit Prioritätsstufe **mittel**, nicht hoch. Begründung:

1. **Kein unmittelbarer Performance-Engpass**: In der aktuellen Praxis sind Tische klein (wenige Dutzend Events pro Schicht). Der Snapshot-Mechanismus adressiert das Performance-Problem ausreichend. Die Migration ist sinnvoll, aber nicht dringend.

2. **Klarer Architektur-Gewinn**: Der Snapshot-as-Event-Ansatz ist konzeptuell fragwürdig — ein Snapshot ist kein Domain-Event. Eine echte Projektion ist konzeptuell sauberer und führt langfristig zu besser wartbarem Code.

3. **Enabler für zukünftige Features**: Analytische Abfragen (Tagesabrechnung nach Benutzer, Umsatz pro Produkt) werden mit typisierten Projektionen erheblich einfacher implementierbar. Die `table_state`-Projektion ist der erste Schritt in diese Richtung.

4. **Bewusstes Maßhalten**: Das Muster bleibt auf Stufe 2 (*Separate Read Models im selben Store*). Eine Trennung in separate Datenspeicher ist für jottis Größe nicht notwendig und würde mehr Komplexität einführen als lösen — wie Fowler warnt.

**Bevorzugter Ansatz für operative Projektionen (Kernfunktionalität)**: Wenn die Projektion umgesetzt wird, empfiehlt sich die **Lazy Projection** (Abschnitt 6.3), da sie alle Vorteile der synchronen Projektion bietet, ohne die drei Nachteile (Schreibkomplexität, C→Q-Abhängigkeit, Backfill) einzuführen. Die synchrone Projektion (Abschnitt 4) bleibt als Alternative, falls eine noch einfachere Implementierung bevorzugt wird.

**Für analytische Projektionen (Zusatzfeatures)**: Der **Background Worker** (Abschnitt 6.4) ist die empfohlene Variante — Eventual Consistency ist für Umsatzauswertungen und Statistiken akzeptabel (siehe Abschnitt 7.3). Analytische Projektionstabellen (z.B. `daily_revenue`, `variant_sales`) liegen in derselben PostgreSQL-Datenbank und erfordern keine separate Infrastruktur.

**Nicht empfohlen** wird:
- Asynchrone Projektionen via Background Worker für **operative Queries** (Balance, offene Positionen): Eventual Consistency ist für den Kassenbetrieb nicht akzeptabel (für Analysen hingegen geeignet — siehe Abschnitt 7)
- Separate Read-Datenbank (Overhead nicht gerechtfertigt)
- PostgreSQL-Trigger-basierte Projektion (Business-Logik in PL/pgSQL schwer wartbar)
- Vollständige Auflösung des Event Store zugunsten von CRUD (verliert Audit-Trail-Garantien — siehe [no-event-sourcing.md](no-event-sourcing.md))

### Fazit

CQRS ist in jotti bereits als Denkmodell verankert — die Benennung `Command`/`Query`, die Trennung der Structs und Handler zeigen das. Der nächste sinnvolle Schritt ist, dieses Muster mit einer konkreten Projektion zu vervollständigen und damit die bekannten Schwächen des aktuellen Ansatzes (Snapshot-Mechanismus, Lese-Komplexität) elegant zu lösen. Die Implementierung ist überschaubar, rückwärtskompatibel und bringt einen messbaren Gewinn in Sauberkeit und Wartbarkeit. Die sqlc-basierte Repository-Architektur (siehe [ADR: Datenbankzugriff](adr/orm.md)) bietet dabei eine solide Grundlage für die neuen Repository-Schichten.

Die Unterscheidung zwischen Kernfunktionalität (Strong Consistency, Lazy Projection) und Zusatzfeatures (Eventual Consistency, Background Worker) — beschrieben in Abschnitt 7 — erlaubt es zudem, für zukünftige analytische Features die passende CQRS-Variante gezielt einzusetzen, ohne das operative Kassensystem zu belasten. Beide Projektion-Kategorien nutzen dieselbe `events`-Tabelle als Quelle der Wahrheit und arbeiten in derselben PostgreSQL-Datenbank — die Komplexität bleibt überschaubar.

---

## 9. Referenzen

- **Martin Fowler**: [CQRS](https://martinfowler.com/bliki/CQRS.html) — Grundlegende Beschreibung des Musters, Warnung vor übermäßigem Einsatz
- **Martin Fowler**: [CommandQuerySeparation](https://martinfowler.com/bliki/CommandQuerySeparation.html) — Das Grundprinzip hinter CQRS
- **Greg Young**: [CQRS Documents](https://cqrs.wordpress.com/wp-content/uploads/2010/11/cqrs_documents.pdf) (PDF) — Originals Dokument, das CQRS popularisiert hat; Verbindung mit Event Sourcing
- **Udi Dahan**: [Clarified CQRS](https://udidahan.com/2009/12/09/clarified-cqrs/) — Frühe, einflussreiche Beschreibung
- **Wikipedia**: [Command Query Responsibility Segregation](https://en.wikipedia.org/wiki/Command_Query_Responsibility_Segregation) — Übersicht und Geschichte
- **Microsoft**: [CQRS pattern (Azure Architecture)](https://learn.microsoft.com/en-us/azure/architecture/patterns/cqrs) — Praxisnahe Beschreibung mit Varianten, Separate Read/Write Models
- **Microsoft**: [Tactical DDD](https://learn.microsoft.com/en-us/azure/architecture/microservices/model/tactical-domain-driven-design) — Aggregate-Design, Domain Events, Konsistenzgrenzen
- **AWS**: [CQRS Pattern (Prescriptive Guidance)](https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/cqrs-pattern.html) — Eventual Consistency und Datenspeicher-Kombinationen
- **Confluent**: [What is CQRS?](https://www.confluent.io/learn/cqrs/) — Asynchrone Updates, Event-Log-Integration, Projections
- **Baytech Consulting**: [Event Sourcing Explained 2025](https://www.baytechconsulting.com/blog/event-sourcing-explained-2025) — Snapshots, Projections, Aggregate-Rehydration
- **Mia-Platform**: [Understanding Event Sourcing and CQRS](https://mia-platform.eu/blog/understanding-event-sourcing-and-cqrs-pattern/) — Zusammenspiel von Event Sourcing und CQRS
- **GeeksforGeeks**: [CQRS vs. Event Sourcing](https://www.geeksforgeeks.org/system-design/difference-between-cqrs-and-event-sourcing/) — Unterschiede und Gemeinsamkeiten
- **DEV Community**: [Event Sourcing vs. CRUD](https://dev.to/alex_aslam/event-sourcing-vs-crud-when-1000-database-writes-dont-matter-5bpj) — Hybride Ansätze und Entscheidungsframeworks
- **Bertrand Meyer**: *Object-Oriented Software Construction* (1988) — Ursprung von CQS
- **Vaughn Vernon**: *Implementing Domain-Driven Design* (2013) — Aggregat-Grenzen, Konsistenz-Grenzen, Eventual Consistency zwischen Bounded Contexts
- **jotti intern**: [Event-Sourcing vs. CRUD](event-sourcing-vs-crud.md) — Hybride Architektur und Event-Sourcing-Details
- **jotti intern**: [jotti ohne Event-Sourcing](no-event-sourcing.md) — CRUD-Alternative und Nachteile-Analyse
- **jotti intern**: [ADR: Datenbankzugriff — Entscheidung für sqlc](adr/orm.md) — sqlc als Persistenz-Werkzeug
