# PostgreSQL — Theorie und Anwendung

Dieses Dokument ist ein allgemeiner Guide für PostgreSQL: Architektur-Grundlagen, Features, Indexierung, Query-Optimierung, Connection Management, Migration-Strategien und den Vergleich mit Alternativen.

---

## Inhaltsverzeichnis

1. [PostgreSQL-Architektur-Grundlagen](#1-postgresql-architektur-grundlagen)
2. [PostgreSQL als Datenbankwahl](#2-postgresql-als-datenbankwahl)
3. [Schema-Design und Konventionen](#3-schema-design-und-konventionen)
4. [Hybride Persistenz: CRUD + Event Store](#4-hybride-persistenz-crud--event-store)
5. [PostgreSQL-spezifische Features](#5-postgresql-spezifische-features)
6. [Indexing Deep Dive](#6-indexing-deep-dive)
7. [Query Optimization](#7-query-optimization)
8. [Advanced Features](#8-advanced-features)
9. [Connection Management](#9-connection-management)
10. [PostgreSQL als Event Store](#10-postgresql-als-event-store)
11. [Performance-Optimierung](#11-performance-optimierung)
12. [Schema-Migration-Strategien](#12-schema-migration-strategien)
13. [PostgreSQL vs. Alternativen](#13-postgresql-vs-alternativen)
14. [Anti-Patterns und Fallstricke](#14-anti-patterns-und-fallstricke)
15. [Referenzen](#15-referenzen)

---

## 1. PostgreSQL-Architektur-Grundlagen

### 1.1 MVCC — Multi-Version Concurrency Control

PostgreSQL verwendet MVCC (Multi-Version Concurrency Control) statt traditioneller Sperrmechanismen. Jede SQL-Anweisung sieht einen **Snapshot** der Datenbank zum Ausführungszeitpunkt, unabhängig vom aktuellen Zustand.

**Kernprinzip:** Statt Zeilen zu überschreiben, erstellt PostgreSQL neue Versionen (Tuples). Die alte Version bleibt sichtbar für Transaktionen, die vor dem Update gestartet wurden.

```
Transaction A (started before update):   SELECT sieht version 1
Transaction B (UPDATE):                  Erstellt version 2 (version 1 bleibt)
Transaction C (started after commit):    SELECT sieht version 2
```

**Vorteile:**

- Reads blockieren nie Writes, Writes blockieren nie Reads
- Konsistente Snapshots ohne Locks
- Serializable Snapshot Isolation (SSI) für strikteste Isolation

**Konsequenz — Dead Tuples:** Alte Versionen akkumulieren sich auf der Festplatte (Bloat). Vacuum bereinigt diese.

### 1.2 WAL — Write-Ahead Logging

WAL (Write-Ahead Logging) garantiert Datenintegrität bei Crashes. Das Kernprinzip: Änderungen werden **zuerst ins WAL-Log geschrieben**, bevor die eigentlichen Datenpages auf Disk geändert werden.

```
Transaktion commit → WAL-Record flushed → Commit bestätigt → Datapage-Flush (später)
```

**Vorteile:**

- Crash Recovery: WAL replay stellt den Zustand nach einem Absturz wieder her
- Weniger Disk-I/O: Sequentielle Writes ins WAL statt random Writes in Datapages
- Grundlage für Streaming Replication (WAL an Replicas senden)
- Point-in-Time Recovery (PITR): WAL archivieren → zu beliebigem Zeitpunkt recovern

### 1.3 Autovacuum und Bloat-Management

Da MVCC alte Versionen (Dead Tuples) zurücklässt, braucht PostgreSQL regelmäßiges Aufräumen:

| Aufgabe                    | Beschreibung                                             |
| -------------------------- | -------------------------------------------------------- |
| **Dead Tuples entfernen**  | Gibt Disk-Space frei (aber nicht ans OS zurück)          |
| **Statistiken updaten**    | Query-Planner braucht aktuelle Statistiken für `ANALYZE` |
| **Visibility Map updaten** | Beschleunigt Index-Only Scans                            |
| **XID Wraparound**         | Verhindert Transaction-ID Overflow (kritisch!)           |

```sql
-- Autovacuum-Aktivität beobachten
SELECT schemaname, tablename, n_dead_tup, n_live_tup, last_autovacuum
FROM pg_stat_user_tables
ORDER BY n_dead_tup DESC;

-- Manuell triggern (nötig nach großen Bulk-Operationen)
VACUUM ANALYZE events;
```

**Tuning:** `autovacuum_vacuum_scale_factor` (default 20%) und `autovacuum_vacuum_threshold` (default 50 rows) steuern, wann Autovacuum anspringt.

### 1.4 Memory-Konfiguration

Die wichtigsten Memory-Parameter in `postgresql.conf`:

| Parameter              | Default | Empfehlung  | Beschreibung                            |
| ---------------------- | ------- | ----------- | --------------------------------------- |
| `shared_buffers`       | 128 MB  | 25% RAM     | Shared Cache für alle Verbindungen      |
| `work_mem`             | 4 MB    | 4–64 MB     | Pro Sort/Hash-Operation, pro Verbindung |
| `effective_cache_size` | 4 GB    | 50–75% RAM  | Planner-Hint: wie viel Cache verfügbar? |
| `maintenance_work_mem` | 64 MB   | 256 MB–1 GB | VACUUM, CREATE INDEX, ALTER TABLE       |
| `wal_buffers`          | auto    | 64 MB       | WAL-Puffer vor dem Flush                |

---

## 2. PostgreSQL als Datenbankwahl

### Warum PostgreSQL?

PostgreSQL eignet sich besonders für Anwendungen mit Event-Sourcing und strukturierten Stammdaten:

| Feature               | PostgreSQL                            | MySQL                    | Typischer Einsatz                     |
| --------------------- | ------------------------------------- | ------------------------ | ------------------------------------- |
| **JSONB**             | Nativ, indizierbar                    | JSON (nicht indizierbar) | Event-Daten als JSONB speichern       |
| **Custom Enums**      | `CREATE TYPE ... AS ENUM`             | ENUM als Spaltentyp      | Rollen, Status-Werte, Kategorien      |
| **Trigger**           | Vollständig (BEFORE/AFTER/INSTEAD OF) | Eingeschränkt            | Append-only-Garantie für Event Stores |
| **ACID**              | Vollständig                           | Vollständig (InnoDB)     | Transaktionale Konsistenz             |
| **IDENTITY Columns**  | `GENERATED BY DEFAULT AS IDENTITY`    | AUTO_INCREMENT           | Standard-SQL-konforme IDs             |
| **Partielle Indexes** | `WHERE`-Klausel im Index              | Nicht unterstützt        | Status-Filter, Archivierung           |
| **CTEs (WITH)**       | Vollständig, rekursiv                 | Seit MySQL 8             | Komplexe Event-Queries                |
| **LISTEN/NOTIFY**     | Nativ                                 | Nicht vorhanden          | Asynchrone Benachrichtigungen         |
| **Partitioning**      | Declarative (Range, List, Hash)       | Eingeschränkt            | Große Event-Tabellen aufteilen        |
| **Window Functions**  | Vollständig                           | Seit MySQL 8 (begrenzt)  | Analytische Queries                   |
| **RLS**               | Row-Level Security                    | Nicht nativ              | Multi-Tenant-Datenisolation           |

### PostgreSQL-Stärken für Event-Sourcing

- **JSONB** für flexible Event-Daten — unterschiedliche Event-Typen in einer Tabelle
- **Trigger** garantieren Append-only — `BEFORE UPDATE/DELETE/TRUNCATE` auf Events
- **Indexierung von JSONB** — Queries auf Event-Daten möglich (GIN-Index)
- **LISTEN/NOTIFY** — Potenzial für asynchrone Projektionen (Stufe 2 CQRS)
- **Sequentielle IDs** — `GENERATED BY DEFAULT AS IDENTITY` für Event-Reihenfolge

---

## 3. Schema-Design und Konventionen

### Namenskonventionen

| Element     | Konvention           | Beispiel                                      |
| ----------- | -------------------- | --------------------------------------------- |
| Tabellen    | Plural, snake_case   | `users`, `product_variants`, `events`         |
| Spalten     | snake_case           | `user_id`, `created_at`, `preis_cents`        |
| Enums       | PascalCase           | `UserRole`, `EntityStatus`, `ProductCategory` |
| Indexes     | `idx_tabelle_spalte` | `idx_events_subject_type`                     |
| Constraints | `tabelle_spalte_fk`  | `events_user_id_fk`                           |

### Custom Enums

PostgreSQL-Enums für typsichere Status-Werte:

```sql
CREATE TYPE UserRole AS ENUM ('admin', 'manager', 'staff');
CREATE TYPE EntityStatus AS ENUM ('active', 'inactive', 'deleted');
CREATE TYPE ProductCategory AS ENUM ('food', 'beverage', 'other');
```

**Vorteile:**

- DB-seitige Validierung (ungültige Werte → Fehler)
- Kein String-Vergleich nötig
- Typsicher im Application-Code (z. B. via sqlc-generierte Go-Enums)

### Soft-Deletes

Stammdaten verwenden Soft-Deletes via `EntityStatus`:

```sql
-- Nie physisch löschen
UPDATE users SET status = 'deleted' WHERE id = $1;

-- Nur aktive Entities laden
SELECT * FROM users WHERE status != 'deleted';
```

### Timestamps

| Spalte       | Typ                         | Beschreibung               |
| ------------ | --------------------------- | -------------------------- |
| `created_at` | `TIMESTAMPTZ DEFAULT now()` | Erstellungszeitpunkt       |
| `updated_at` | `TIMESTAMPTZ DEFAULT now()` | Letzter Änderungszeitpunkt |
| `timestamp`  | `TIMESTAMPTZ DEFAULT now()` | Event-Zeitstempel          |

**Immer `TIMESTAMPTZ`** (mit Zeitzone), nie `TIMESTAMP`. PostgreSQL speichert intern UTC, konvertiert bei Ausgabe.

---

## 4. Hybride Persistenz: CRUD + Event Store

### Das Zwei-Welten-Modell

Viele Systeme kombinieren zwei Persistenzmuster in **einer** PostgreSQL-Instanz:

```
┌────────────────────────────────────────────────────────────────────┐
│                        PostgreSQL                                  │
│                                                                    │
│  ┌──────────────────────────┐    ┌─────────────────────────────┐  │
│  │   CRUD-Tabellen           │    │   Event Store               │  │
│  │                           │    │                             │  │
│  │   users                   │    │   events (append-only)      │  │
│  │   orders                  │    │   ┌─────────────────────┐   │  │
│  │   products                │    │   │ Trigger: kein        │   │  │
│  │   order_items             │    │   │ UPDATE/DELETE/TRUNCATE│   │  │
│  │                           │    │   └─────────────────────┘   │  │
│  │   Muster: SELECT, INSERT, │    │                             │  │
│  │   UPDATE (kein DELETE)    │    │   Muster: nur INSERT        │  │
│  └──────────────────────────┘    └─────────────────────────────┘  │
│                                                                    │
│  Referenzielle Integrität:                                        │
│  events.user_id → users.id                                        │
│  order_items.product_id → products.id                              │
└────────────────────────────────────────────────────────────────────┘
```

### Wann CRUD, wann Event Store?

| Kriterium           | CRUD-Tabellen                               | Event Store                                |
| ------------------- | ------------------------------------------- | ------------------------------------------ |
| **Datentyp**        | Stammdaten (Benutzer, Produkte, Kategorien) | Operationen (Bestellungen, Zahlungen, ...) |
| **Änderungsmuster** | Update in-place                             | Append-only                                |
| **History**         | Nur aktueller Zustand                       | Vollständige Historie                      |
| **Schema**          | Relationale Spalten                         | JSONB (flexibel)                           |
| **Queries**         | Einfache SELECTs                            | Replay + Aggregation                       |
| **Fremdschlüssel**  | Vollständig                                 | Nur `user_id`                              |

---

## 5. PostgreSQL-spezifische Features

### 5.1 JSONB für Event-Daten

Events speichern ihre Daten als JSONB:

```sql
INSERT INTO events (user_id, type, subject, data)
VALUES (1, 'order.placed:v1', 'order:42',
  '{"items": [{"id": 1, "name": "Espresso", "priceCents": 350, "quantity": 2}],
    "comment": "extra hot",
    "totalCents": 700}'
);
```

**JSONB-Vorteile:**

- Binäres Format (schneller als JSON-Text)
- Indizierbar (GIN-Index)
- Operatoren: `->`, `->>`, `@>`, `?`, `#>`
- Kein Schema-Enforcement (flexible Event-Typen)

**Potenzielle Queries auf JSONB:**

```sql
-- Events mit bestimmtem Produkt finden (für Reporting)
SELECT * FROM events
WHERE data @> '{"items": [{"id": 42}]}';

-- Gesamtpreis aus Event extrahieren
SELECT data->>'totalCents' FROM events
WHERE type = 'order.placed:v1';
```

### 5.2 Trigger für Append-only-Garantie

```sql
-- Events-Tabelle vor Änderungen schützen
CREATE OR REPLACE FUNCTION prevent_event_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Events are immutable. %, % operations are not allowed.',
        TG_OP, TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_event_update
    BEFORE UPDATE ON events
    FOR EACH ROW EXECUTE FUNCTION prevent_event_modification();

CREATE TRIGGER prevent_event_delete
    BEFORE DELETE ON events
    FOR EACH ROW EXECUTE FUNCTION prevent_event_modification();

CREATE TRIGGER prevent_event_truncate
    BEFORE TRUNCATE ON events
    EXECUTE FUNCTION prevent_event_modification();
```

**Warum DB-Level statt App-Level?** Ein Trigger schützt auch vor direkten SQL-Zugriffen (psql, Migrationsscripte, Debugging). Application-Code kann die Regel nicht umgehen.

### 5.3 IDENTITY Columns

```sql
id INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY
```

Vorteile gegenüber `SERIAL`:

- SQL-Standard (nicht PostgreSQL-spezifisch)
- Kein separates Sequence-Objekt
- `BY DEFAULT` erlaubt manuelle ID-Angabe (nützlich für Tests/Seeding)

### 5.4 TIMESTAMPTZ

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

- Speichert intern UTC
- Konvertiert bei Ausgabe in die aktuelle Zeitzone
- Kein Ärger mit Sommer-/Winterzeit
- Go's `time.Time` mappt direkt auf `TIMESTAMPTZ`

---

## 6. Indexing Deep Dive

### 6.1 Index-Typen im Überblick

| Index-Typ   | Algorithmus        | Ideal für                                | Beispiel                               |
| ----------- | ------------------ | ---------------------------------------- | -------------------------------------- |
| **B-Tree**  | Balanced Tree      | Gleichheit, Ranges, Sortierung           | `WHERE id = $1`, `ORDER BY created_at` |
| **Hash**    | Hash-Tabelle       | Nur Gleichheit (`=`)                     | Selten nötig (B-Tree deckt alles ab)   |
| **GIN**     | Inverted Index     | JSONB, Arrays, Volltextsuche             | `WHERE data @> '{"id": 42}'`           |
| **GiST**    | Generalized Search | Geometrie, Volltext, Range-Typen         | PostGIS, `tsvector`, `tsrange`         |
| **BRIN**    | Block Range Index  | Sehr große Tabellen, natürlich geordnet  | `WHERE timestamp BETWEEN $1 AND $2`    |
| **SP-GiST** | Space-Partitioned  | Nicht-balancierte Strukturen (Quadtrees) | Selten in Standard-Anwendungen         |

#### B-Tree (Standard)

Der Standard-Index für fast alle Anwendungsfälle:

```sql
-- Automatisch bei PRIMARY KEY und UNIQUE
CREATE INDEX idx_events_subject ON events (subject);

-- Composite: Reihenfolge wichtig! Führende Spalte = selektivster Wert
CREATE INDEX idx_events_subject_type ON events (subject, type);
-- Unterstützt: WHERE subject = $1              ✅
-- Unterstützt: WHERE subject = $1 AND type = $2 ✅
-- Unterstützt NICHT: WHERE type = $2 allein    ❌ (kein Prefix-Match)
```

#### GIN (Generalized Inverted Index)

Für JSONB-Daten und Arrays:

```sql
-- GIN-Index für JSONB @> Operator
CREATE INDEX idx_events_data ON events USING GIN (data);

-- Dann schnell:
SELECT * FROM events WHERE data @> '{"items": [{"id": 42}]}';
```

**Achtung:** GIN-Indexes sind bei Writes teurer als B-Tree. Nur anlegen, wenn JSONB-Queries häufig vorkommen.

#### BRIN (Block Range Index)

Sehr kleiner Index für append-only Tabellen mit natürlicher Sortierung:

```sql
-- Perfekt für Events: neue Events haben immer größere IDs/Timestamps
CREATE INDEX idx_events_timestamp_brin ON events USING BRIN (timestamp);
```

BRIN speichert nur Min/Max pro Block-Gruppe (128 Pages default). Platzsparend, aber weniger präzise als B-Tree.

### 6.2 Partial Indexes

Indexes nur über eine Teilmenge der Tabelle:

```sql
-- Nur aktive Benutzer indexieren (deleted-Einträge werden nie abgefragt)
CREATE INDEX idx_users_active ON users (username)
WHERE status != 'deleted';

-- Nur unbezahlte Bestellungen (kleine Menge, häufig abgefragt)
CREATE INDEX idx_orders_unpaid ON orders (table_id)
WHERE paid_at IS NULL;
```

**Vorteile:** Kleiner Index (weniger Speicher, schnelleres Schreiben), ideal für ungleiche Datenverteilungen.

### 6.3 Covering Indexes (INCLUDE)

Fügt Spalten zum Index hinzu, ohne sie als Suchspalten zu nutzen:

```sql
-- Query: SELECT username, email FROM users WHERE id = $1
-- Ohne INCLUDE: Index Scan + Heap Fetch
-- Mit INCLUDE: Index-Only Scan (kein Heap-Zugriff nötig)
CREATE INDEX idx_users_id_covering ON users (id) INCLUDE (username, email);
```

### 6.4 Index-Only Scans

PostgreSQL kann Queries direkt aus dem Index beantworten, ohne die eigentliche Tabelle (Heap) zu lesen — wenn alle benötigten Spalten im Index enthalten sind:

```sql
EXPLAIN ANALYZE
SELECT username FROM users WHERE id = $1;
-- → Index Only Scan using idx_users_id_covering (0 Heap Fetches)
```

**Voraussetzung:** Visibility Map muss aktuell sein (Autovacuum oder manuelles `VACUUM`).

### 6.5 Composite vs. Single-Column Indexes

| Szenario                                     | Empfehlung                                         |
| -------------------------------------------- | -------------------------------------------------- |
| `WHERE a = $1`                               | Single-Column Index auf `a`                        |
| `WHERE a = $1 AND b = $2`                    | Composite `(a, b)` — führende Spalte = selektivste |
| `WHERE a = $1 ORDER BY b`                    | Composite `(a, b)` — vermeidet Sort                |
| `WHERE a = $1` und `WHERE b = $1` (getrennt) | Zwei Single-Column Indexes                         |
| `WHERE b = $1` (ohne `a`)                    | Separate Index auf `b` (Composite hilft nicht)     |

---

## 7. Query Optimization

### 7.1 EXPLAIN ANALYZE

Das wichtigste Werkzeug zur Query-Analyse:

```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM events
WHERE subject = 'order:42'
  AND id >= 100
ORDER BY id ASC;
```

**Ausgabe interpretieren:**

```
Index Scan using idx_events_subject_type on events  (cost=0.29..8.31 rows=1 width=234)
                                                    (actual time=0.041..0.043 rows=1 loops=1)
  Index Cond: ((subject = 'order:42') AND (id >= 100))
  Buffers: shared hit=3
Planning Time: 0.121 ms
Execution Time: 0.063 ms
```

| Begriff           | Bedeutung                                     |
| ----------------- | --------------------------------------------- |
| `cost=X..Y`       | Planner-Schätzung: Startkosten..Gesamtkosten  |
| `rows=N`          | Geschätzte Zeilenanzahl (vs. `actual rows=N`) |
| `loops=N`         | Wie oft wurde dieser Node ausgeführt          |
| `Buffers: hit=N`  | N Pages aus Shared-Buffer-Cache (gut!)        |
| `Buffers: read=N` | N Pages von Disk gelesen (teuer!)             |
| `Seq Scan`        | Tabellensequenz-Scan (kein Index genutzt)     |
| `Index Scan`      | Index + Heap-Fetch                            |
| `Index Only Scan` | Nur Index, kein Heap-Fetch (optimal)          |

### 7.2 Common Table Expressions (CTEs)

CTEs strukturieren komplexe Queries. Seit PostgreSQL 12 sind CTEs standardmäßig **non-materialized** (inline optimiert):

```sql
-- Non-materialized CTE (PostgreSQL 12+): Planner optimiert durch
WITH recent_events AS (
    SELECT * FROM events WHERE subject = $1 ORDER BY id DESC LIMIT 100
)
SELECT type, COUNT(*) FROM recent_events GROUP BY type;

-- Materialized CTE: Explizit als Subquery materialisieren
WITH MATERIALIZED snapshot AS (
    SELECT MAX(id) AS last_snapshot_id
    FROM events
    WHERE subject = $1 AND type = 'snapshot:v1'
)
SELECT e.* FROM events e, snapshot s WHERE e.id >= COALESCE(s.last_snapshot_id, 0);
```

**Rekursive CTEs** für hierarchische Daten:

```sql
WITH RECURSIVE category_tree AS (
    SELECT id, name, parent_id, 0 AS depth FROM categories WHERE parent_id IS NULL
    UNION ALL
    SELECT c.id, c.name, c.parent_id, ct.depth + 1
    FROM categories c JOIN category_tree ct ON c.parent_id = ct.id
)
SELECT * FROM category_tree ORDER BY depth, name;
```

### 7.3 Window Functions

Analytische Berechnungen über Zeilengruppen ohne GROUP BY:

```sql
-- Laufende Summe der Zahlungen pro Tisch
SELECT
    subject,
    type,
    timestamp,
    SUM(CAST(data->>'amountCents' AS INT))
        OVER (PARTITION BY subject ORDER BY id) AS running_total
FROM events
WHERE type = 'payment.registered:v1';

-- Rang der Bestellungen nach Größe
SELECT
    id,
    data->>'totalCents' AS total,
    RANK() OVER (ORDER BY CAST(data->>'totalCents' AS INT) DESC) AS rank
FROM events
WHERE type = 'order.placed:v1';
```

### 7.4 Query Planner Statistiken

```sql
-- Aktivieren (einmalig)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Top 10 langsamste Queries
SELECT query, calls, mean_exec_time, total_exec_time, rows
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Tabellen mit vielen Sequential Scans → Index-Kandidaten
SELECT relname, seq_scan, idx_scan, n_live_tup
FROM pg_stat_user_tables
WHERE seq_scan > 100
ORDER BY seq_scan DESC;
```

---

## 8. Advanced Features

### 8.1 LISTEN/NOTIFY

Asynchrone Benachrichtigungen zwischen Datenbankverbindungen — ideal für CQRS-Projektionen:

```sql
-- Producer (z. B. Trigger nach INSERT in events)
CREATE OR REPLACE FUNCTION notify_event_inserted()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('events_channel', row_to_json(NEW)::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER after_event_insert
    AFTER INSERT ON events
    FOR EACH ROW EXECUTE FUNCTION notify_event_inserted();
```

```go
// Consumer (Go mit pgx)
conn, _ := pgx.Connect(ctx, databaseURL)
_, err := conn.Exec(ctx, "LISTEN events_channel")

for {
    notification, err := conn.WaitForNotification(ctx)
    if err != nil { break }
    // notification.Payload = JSON des neuen Events
    processEvent(notification.Payload)
}
```

**Anwendungsfälle:**

- Asynchrone Read-Model-Projektionen (CQRS Stufe 2)
- Cache-Invalidierung bei Stammdaten-Änderungen
- Real-Time Dashboard Updates

**Einschränkungen:** LISTEN/NOTIFY ist In-Memory und nicht persistent. Bei Verbindungsabbruch gehen Benachrichtigungen verloren → für kritische Workflows Message Queue bevorzugen.

### 8.2 Partitioning

Große Tabellen in kleinere physische Einheiten aufteilen:

```sql
-- Range-Partitioning nach Monat (für Event-Archivierung)
CREATE TABLE events (
    id        INT GENERATED BY DEFAULT AS IDENTITY,
    subject   TEXT NOT NULL,
    type      TEXT NOT NULL,
    data      JSONB NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY RANGE (timestamp);

CREATE TABLE events_2024_q1 PARTITION OF events
    FOR VALUES FROM ('2024-01-01') TO ('2024-04-01');
CREATE TABLE events_2024_q2 PARTITION OF events
    FOR VALUES FROM ('2024-04-01') TO ('2024-07-01');

-- Hash-Partitioning für gleichmäßige Verteilung
CREATE TABLE events_large PARTITION BY HASH (subject);
CREATE TABLE events_large_0 PARTITION OF events_large FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE events_large_1 PARTITION OF events_large FOR VALUES WITH (MODULUS 4, REMAINDER 1);
```

**Wann Partitioning?**

- Tabelle > physischer RAM (Faustregel: > 10 Mio. Zeilen)
- Häufige Queries filtern nach Partitions-Key (Partition Pruning)
- Ältere Daten regelmäßig archivieren/löschen (`DETACH PARTITION` ist sofortig)

### 8.3 Materialized Views

Pre-computed Read Models für teure Aggregationen:

```sql
-- Materialized View für Dashboard-Statistiken
CREATE MATERIALIZED VIEW daily_revenue AS
SELECT
    DATE_TRUNC('day', timestamp) AS day,
    SUM(CAST(data->>'totalCents' AS INT)) AS revenue_cents,
    COUNT(*) AS order_count
FROM events
WHERE type = 'order.placed:v1'
GROUP BY DATE_TRUNC('day', timestamp);

CREATE UNIQUE INDEX ON daily_revenue (day);  -- Für CONCURRENTLY refresh

-- Refresh (blockierend oder concurrent)
REFRESH MATERIALIZED VIEW daily_revenue;
REFRESH MATERIALIZED VIEW CONCURRENTLY daily_revenue;  -- Kein Lock, braucht UNIQUE Index
```

**Anwendungsfälle:** Reporting-Queries, Dashboard-Aggregate, Read Models für CQRS.

### 8.4 Row-Level Security (RLS)

Datenisolation auf Zeilenebene — ohne Application-Code:

```sql
-- RLS aktivieren
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

-- Policy: Jeder User sieht nur seine eigenen Bestellungen
CREATE POLICY user_isolation ON orders
    USING (user_id = current_setting('app.current_user_id')::INT);
```

**Anwendungsfall:** Multi-Tenant-Systeme, Datenschutz-konforme Isolation.

### 8.5 Generated Columns

Berechnete Spalten direkt in der DB:

```sql
ALTER TABLE products ADD COLUMN
    preis_euro NUMERIC(10,2) GENERATED ALWAYS AS (preis_cents / 100.0) STORED;

-- Wird automatisch bei INSERT/UPDATE berechnet
INSERT INTO products (name, preis_cents) VALUES ('Espresso', 350);
-- preis_euro = 3.50 (automatisch)
```

---

## 9. Connection Management

### 9.1 pgxpool (Go-intern)

pgxpool verwaltet einen Connection Pool innerhalb des Go-Prozesses:

```go
config, _ := pgxpool.ParseConfig(databaseURL)
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = 1 * time.Hour
config.MaxConnIdleTime = 30 * time.Minute

pool, _ := pgxpool.NewWithConfig(ctx, config)
defer pool.Close()
```

**Best Practices:**

- Pool einmal erstellen, überall teilen (Singleton)
- `pool.Query()` / `pool.QueryRow()` nutzen, nie `pool.Acquire()` manuell
- Connection-Anzahl: `(Anzahl CPU-Cores * 2) + 1` als Startpunkt
- Max Connections = `(PostgreSQL max_connections - 3) / Anzahl App-Instanzen`

### 9.2 PgBouncer (externer Pool)

PgBouncer sitzt als Proxy zwischen Application und PostgreSQL, multiplext viele App-Connections auf wenige DB-Connections:

```
App (1000 connections) → PgBouncer (20 DB-Connections) → PostgreSQL
```

**Modi:**

| Modus           | Beschreibung                                    | Einsatz                         |
| --------------- | ----------------------------------------------- | ------------------------------- |
| **Session**     | Verbindung für gesamte Client-Session           | Mit Prepared Statements         |
| **Transaction** | Verbindung nur während Transaktion              | Höchste Effizienz (kein LISTEN) |
| **Statement**   | Verbindung pro Statement (kein Multi-Statement) | Selten sinnvoll                 |

**Wann PgBouncer?**

- Viele kurzlebige Verbindungen (z. B. Serverless, PHP ohne Persistent Connections)
- PostgreSQL `max_connections` wird überschritten
- Mehrere App-Instanzen teilen eine PostgreSQL-Instanz

### 9.3 Odyssey

Odyssey ist ein moderner, multithreaded Connection Pooler von Yandex (Alternative zu PgBouncer):

- Multithreaded (PgBouncer ist single-threaded)
- Bessere TLS-Unterstützung
- Prometheus-Metriken out-of-the-box
- Einsatz: Hochlast-Umgebungen, wo PgBouncer zum Bottleneck wird

### 9.4 Connection-Sizing-Formel

```
Optimale Verbindungen = (Anzahl CPU-Kerne) * 2 + Disk-Spindles

Für eine 4-Core-Maschine mit SSDs:
  DB max_connections ≈ 9–20

Für 3 App-Instanzen mit pgxpool:
  Pool MaxConns pro Instanz = max_connections / 3
```

PostgreSQL öffnet pro Verbindung einen eigenen Prozess (~5–10 MB RAM). Zu viele Verbindungen führen zu RAM-Druck und Context-Switching-Overhead.

---

## 10. PostgreSQL als Event Store

> Zero-Downtime Deployment-Strategien und Backup-Automation werden in [DevOps & Deployment](devops.md) vertieft.

### Vergleich: PostgreSQL vs. spezialisierte Event Stores

| Kriterium                    | PostgreSQL                         | EventStoreDB                        | Kafka                                  |
| ---------------------------- | ---------------------------------- | ----------------------------------- | -------------------------------------- |
| **Primärer Zweck**           | Relationale DB + Event Store       | Dedizierter Event Store             | Message Broker + Log                   |
| **Setup-Komplexität**        | Gering (bekanntes Tool)            | Mittel                              | Hoch (Zookeeper/KRaft)                 |
| **Ordering**                 | Global (IDENTITY) oder pro Subject | Per Stream nativ                    | Per Partition                          |
| **Replay**                   | SQL-Query                          | Stream Subscription                 | Consumer Group Offset                  |
| **Projections**              | Polling oder LISTEN/NOTIFY         | Nativ (Catch-Up Subscriptions)      | Consumer Groups                        |
| **Retention**                | Manuell (Archivierung)             | Konfigurierbar ($MaxCount, $MaxAge) | Log Compaction, Retention Policy       |
| **Schema-Evolution**         | JSONB + Upcasting in App-Code      | JSONB + Upcasting                   | Schema Registry (Avro, Protobuf)       |
| **Throughput**               | Mittel (10k–100k Events/s)         | Hoch (100k+ Events/s)               | Sehr hoch (Millionen/s)                |
| **Operational Overhead**     | Gering (gemeinsam mit CRUD-Daten)  | Mittel                              | Hoch                                   |
| **Transaktionale Garantien** | Vollständig (ACID)                 | Optimistic Concurrency Control      | At-least-once / Exactly-once (komplex) |

**Empfehlung:** PostgreSQL als Event Store ist optimal für Systeme, die:

- PostgreSQL ohnehin für Stammdaten nutzen
- Moderate Event-Volumina haben (< 1 Mio. Events/Tag)
- Kein separates Infrastruktur-Tool einführen wollen
- ACID-Transaktionen zwischen CRUD und Events benötigen

### Skalierung des Event Stores

Bei wachsenden Event-Tabellen:

1. **BRIN-Index** auf `timestamp` — sehr kleiner Index für append-only Tabellen
2. **Partitioning** nach Zeit — ältere Partitions nach Cold Storage archivieren
3. **Snapshots** — State-Aggregation, damit nicht alle Events neu abgespielt werden müssen
4. **Archivierung** — Events älter als X Monate in separate Tabelle verschieben (`DETACH PARTITION`)

---

## 11. Performance-Optimierung

### 11.1 N+1 Problem vermeiden

Das N+1-Problem entsteht, wenn für jede Zeile einer Hauptquery eine zusätzliche Query ausgeführt wird:

```sql
-- FALSCH: N+1 (1 Query + N Queries für Varianten)
SELECT * FROM products;                                        -- 1 Query
SELECT * FROM product_variants WHERE product_id = $1;         -- N Queries

-- RICHTIG: 1 Query mit Join
SELECT p.*, pv.* FROM products p
JOIN product_variants pv ON pv.product_id = p.id;             -- 1 Query
```

### 11.2 Snapshot-Optimierung für Event-Sourcing

Ohne Snapshots: `O(n)` Events pro Query (n = Gesamtzahl Events pro Subject)

Mit Snapshots: `O(k)` Events pro Query (k = Events seit letztem Snapshot)

```sql
-- Events ab letztem Snapshot laden
SELECT * FROM events
WHERE subject = $1
  AND id >= COALESCE(
    (SELECT MAX(id) FROM events WHERE subject = $1 AND type = $2),
    0
  )
ORDER BY id ASC;
```

### 11.3 Vacuum-Monitoring

```sql
-- Tabellen mit hohem Dead-Tuple-Anteil identifizieren
SELECT schemaname, tablename, n_dead_tup, n_live_tup,
       ROUND(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_pct
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY dead_pct DESC;
```

---

## 12. Schema-Migration-Strategien

### 12.1 Migration-Tools im Vergleich

| Tool               | Sprache | Ansatz               | Highlights                                       |
| ------------------ | ------- | -------------------- | ------------------------------------------------ |
| **golang-migrate** | Go      | Up/Down SQL-Dateien  | Einfach, direkte SQL-Kontrolle, CLI + Library    |
| **Atlas**          | Go      | Deklarativ (HCL/SQL) | Schema Diff, Lint, CI-Integration                |
| **goose**          | Go      | Up/Down SQL/Go       | Go-Migrations für komplexe Datentransformationen |
| **Flyway**         | Java    | Versioniert (SQL)    | Enterprise-Features, breite DB-Unterstützung     |
| **Liquibase**      | Java    | XML/YAML/SQL         | Multi-DB, Rollback-Unterstützung                 |

**Für Go-Backends empfohlen:** golang-migrate (einfach, stabil) oder Atlas (moderner, mehr Features).

### 12.2 golang-migrate

Migrationen in `database/migrations/`:

```
database/migrations/
├── 01_initial.up.sql    # Schema erstellen
├── 01_initial.down.sql  # Schema zurücksetzen
├── 02_feature.up.sql    # Neues Feature
└── 02_feature.down.sql  # Feature zurücksetzen
```

### 12.3 Migrations-Konventionen

| Regel               | Beschreibung                                  |
| ------------------- | --------------------------------------------- |
| **Nummerierung**    | Zweistellig, aufsteigend: `01_`, `02_`, `03_` |
| **Naming**          | `<nr>_<beschreibung>.up.sql` / `.down.sql`    |
| **Atomarität**      | Eine Migration = eine logische Änderung       |
| **Reversibilität**  | Jede `up.sql` hat eine `down.sql`             |
| **Idempotenz**      | `IF NOT EXISTS` / `IF EXISTS` verwenden       |
| **Daten-Migration** | Getrennt von Schema-Migration                 |

### 12.4 Zero-Downtime Migrations (Expand/Contract)

Direkte Schema-Änderungen können laufende Deployments unterbrechen. Das **Expand/Contract Pattern** vermeidet dies:

```
Schritt 1 — Expand:      Neue Spalte hinzufügen (ALTER TABLE ADD COLUMN)
Schritt 2 — Migrate:     Daten in neue Spalte befüllen
Schritt 3 — Deploy App:  Neue App-Version schreibt beide Spalten
Schritt 4 — Contract:    Alte Spalte entfernen (ALTER TABLE DROP COLUMN)
```

```sql
-- Schritt 1: Expand — neue email_address Spalte hinzufügen
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_address TEXT;

-- Schritt 2: Daten migrieren (in kleinen Batches)
UPDATE users SET email_address = email WHERE email_address IS NULL LIMIT 1000;

-- Schritt 4: Contract — nach erfolgreichem Deployment der neuen App-Version
ALTER TABLE users DROP COLUMN IF EXISTS email;
```

**Backward-compatible Änderungen (sicher):**

- Neue Tabelle/Spalte hinzufügen
- Spalte nullable machen
- Neue Enum-Werte hinzufügen
- Index erstellen (CONCURRENTLY)

**Breaking Changes (brauchen Expand/Contract):**

- Spalte umbenennen/löschen
- Spalte NOT NULL machen
- Datentyp ändern
- Enum-Wert entfernen

```sql
-- Index CONCURRENTLY: kein Table Lock, kann ohne Downtime erstellt werden
CREATE INDEX CONCURRENTLY idx_events_new ON events (subject, type);
```

### 12.5 Event-Schema-Evolution

Events sind immutable — Schema-Änderungen erfolgen über **Versionierung**, nicht über ALTER TABLE:

```sql
-- V1 Event (bestehend)
INSERT INTO events (type, data)
VALUES ('order.placed:v1', '{"items": [...]}');

-- V2 Event (neues Schema)
INSERT INTO events (type, data)
VALUES ('order.placed:v2', '{"items": [...], "newField": "..."}');
```

Der Go-Code muss beide Versionen beim Replay unterstützen (Upcasting):

```go
switch event.Type {
case "order.placed:v1":
    var v1 OrderPlacedEventV1
    json.Unmarshal(event.Data, &v1)
    return upcastV1toV2(v1)
case "order.placed:v2":
    var v2 OrderPlacedEventV2
    json.Unmarshal(event.Data, &v2)
    return v2
}
```

---

## 13. PostgreSQL vs. Alternativen

### 13.1 PostgreSQL vs. MySQL/MariaDB

| Aspekt               | PostgreSQL                          | MySQL/MariaDB                      |
| -------------------- | ----------------------------------- | ---------------------------------- |
| **JSONB**            | Nativ, indizierbar, GIN-Index       | JSON-Typ, kein Index               |
| **SQL-Compliance**   | SQL-Standard konform                | Historisch viele Abweichungen      |
| **ACID**             | Vollständig (inkl. DDL)             | Vollständig (InnoDB), DDL implicit |
| **Replication**      | Logical + Physical                  | Binlog-basiert (Physical)          |
| **Erweiterungen**    | PostGIS, TimescaleDB, pgvector...   | Weniger Ecosystem                  |
| **Volltextsuche**    | `tsvector`, `tsquery`               | `FULLTEXT` Index (schwächer)       |
| **Partitioning**     | Declarative (Range, List, Hash)     | Eingeschränkt                      |
| **Window Functions** | Vollständig                         | Seit MySQL 8 (begrenzt)            |
| **Lizenz**           | PostgreSQL License (sehr permissiv) | GPL + proprietäres Oracle-Modell   |

**Empfehlung:** PostgreSQL für neue Projekte. MySQL wenn bestehende Infrastruktur oder spezifisches Tooling es erfordert.

### 13.2 PostgreSQL vs. MongoDB

| Aspekt                 | PostgreSQL                          | MongoDB                             |
| ---------------------- | ----------------------------------- | ----------------------------------- |
| **Datenmodell**        | Relational + JSONB                  | Dokument-orientiert                 |
| **Schema**             | Explizit (mit JSONB-Flexibilität)   | Schema-less (optional Validation)   |
| **ACID**               | Multi-Row, Cross-Table Transactions | Multi-Document Transactions (v4.0+) |
| **Joins**              | Nativ, effizient                    | `$lookup` (teuer, limitiert)        |
| **Indexierung**        | B-Tree, GIN, GiST, BRIN             | B-Tree, Compound, Text, Geo         |
| **Horizontal Scaling** | Vertical + Citus (Extension)        | Nativ (Sharded Cluster)             |
| **Query-Sprache**      | SQL (Standard)                      | MQL (proprietär)                    |

**Wann MongoDB bevorzugen:**

- Vollständig schema-lose Daten ohne relationale Verknüpfungen
- Horizontales Sharding nativ benötigt (> 1 TB Daten)
- Team-Präferenz für dokumentenorientiertes Modell

### 13.3 PostgreSQL vs. NewSQL (CockroachDB, YugabyteDB)

NewSQL kombiniert SQL-Semantik mit horizontaler Skalierung:

| Aspekt                     | PostgreSQL               | CockroachDB / YugabyteDB             |
| -------------------------- | ------------------------ | ------------------------------------ |
| **Kompatibilität**         | Referenz-Implementierung | PostgreSQL-kompatibles Wire-Protocol |
| **Horizontal Scaling**     | Vertical + Citus         | Nativ (geo-distributed)              |
| **Konsistenz**             | ACID (single-node)       | Serializable (distributed)           |
| **Latenz**                 | Niedrig (lokal)          | Höher (Consensus-Protokoll)          |
| **Operational Complexity** | Gering                   | Hoch                                 |

**Empfehlung:** PostgreSQL für ~99% der Anwendungen. NewSQL nur wenn horizontale Skalierung über mehrere Regionen zwingend nötig ist.

---

## 14. Anti-Patterns und Fallstricke

### 14.1 Floats für Geldbeträge

```sql
-- FALSCH
price DECIMAL(10, 2)  -- Verführt zu Float-Konvertierung in Go
amount NUMERIC         -- Gleiche Problematik

-- RICHTIG
preis_cents INT NOT NULL  -- Immer Ganzzahl-Cents
```

### 14.2 Fehlende Indexes

```sql
-- FALSCH: Events ohne Index abfragen
SELECT * FROM events WHERE subject = 'order:42';
-- → Seq Scan über ALLE Events

-- RICHTIG: Index vorhanden
CREATE INDEX idx_events_subject_type ON events (subject, type);
-- → Index Scan, nur relevante Events gelesen
```

### 14.3 Hard Deletes statt Soft Deletes

```sql
-- FALSCH
DELETE FROM users WHERE id = $1;
-- Referenzierte Events verlieren ihren user_id-Bezug

-- RICHTIG
UPDATE users SET status = 'deleted', updated_at = now() WHERE id = $1;
-- Event-History bleibt intakt
```

### 14.4 JSONB ohne Validierung

JSONB akzeptiert beliebiges JSON. Validierung sollte **vor dem INSERT** im Application-Code erfolgen:

```go
// Validierung VOR dem Speichern
event, err := domain.NewOrderPlacedEvent(userID, orderID, items, comment)
if err != nil {
    return err  // Validierungsfehler
}
eventRepo.WriteEvent(ctx, event)  // Nur gültige Events speichern
```

### 14.5 VACUUM FULL in Production vermeiden

`VACUUM FULL` erfordert einen exklusiven Lock und kann die Tabelle für Minuten sperren:

```sql
-- FALSCH für Production (Lock!)
VACUUM FULL events;

-- RICHTIG: Standard VACUUM (läuft parallel zu Queries)
VACUUM ANALYZE events;
```

Autovacuum regelmäßig laufen lassen verhindert die Notwendigkeit von `VACUUM FULL`.

### 14.6 Zu viele Verbindungen

```go
// FALSCH: Verbindung pro Request
conn, _ := pgx.Connect(ctx, databaseURL)
defer conn.Close(ctx)
// → Jeder Request öffnet/schließt eine DB-Connection (teuer!)

// RICHTIG: Pool einmal erstellen, überall teilen
var pool *pgxpool.Pool  // globaler Singleton
```

---

## 15. Referenzen

### PostgreSQL-Dokumentation

- [PostgreSQL Dokumentation](https://www.postgresql.org/docs/current/) — Offizielle Referenz
- [PostgreSQL MVCC](https://www.postgresql.org/docs/current/mvcc-intro.html) — Multi-Version Concurrency Control
- [PostgreSQL WAL](https://www.postgresql.org/docs/current/wal-intro.html) — Write-Ahead Logging
- [PostgreSQL Routine Vacuuming](https://www.postgresql.org/docs/current/routine-vacuuming.html) — Autovacuum, Bloat
- [PostgreSQL Index Types](https://www.postgresql.org/docs/current/indexes-types.html) — B-Tree, GIN, GiST, BRIN
- [PostgreSQL Partial Indexes](https://www.postgresql.org/docs/current/indexes-partial.html) — Index-Teilmengen
- [PostgreSQL LISTEN/NOTIFY](https://www.postgresql.org/docs/current/sql-listen.html) — Asynchrone Benachrichtigungen
- [PostgreSQL Partitioning](https://www.postgresql.org/docs/current/ddl-partitioning.html) — Range, List, Hash
- [PostgreSQL Performance Wiki](https://wiki.postgresql.org/wiki/Performance_Optimization) — Tuning-Checkliste

### Bücher und Artikel

- [Use The Index, Luke](https://use-the-index-luke.com/) — Indexing-Bibel (DB-agnostisch mit SQL-Fokus)
- [Brandur: Postgres Atomicity](https://brandur.org/postgres-atomicity) — PostgreSQL Internals, MVCC Deep Dive
- [DB Performance 101](https://dev.to/ari-ghosh/db-performance-101-a-practical-deep-dive-into-backend-database-optimization-4cag) — Connection Pooling, N+1, Indexing, Query Optimization
- [PostgreSQL vs MySQL (Bytebase)](https://www.bytebase.com/blog/postgres-vs-mysql/) — Feature-Vergleich

### Tooling

- [PgBouncer Dokumentation](https://www.pgbouncer.org/) — Connection Pooling
- [Atlas (Schema Migration)](https://atlasgo.io/) — Moderne deklarative Schema-Migrationen
- [golang-migrate](https://github.com/golang-migrate/migrate) — SQL-Migrationstool für Go
- [pg_partman](https://github.com/pgpartman/pg_partman) — Automatisches Partition Management

### Projekt-intern

- [Go Backend Architektur](go-backend.md) — sqlc-Workflow, SQL-Tooling-Vergleich
