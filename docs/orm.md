# Database & Backend Persistence

This document describes how jotti implements database access and persistence. jotti does **not** use an ORM — instead it uses hand-written SQL queries executed via the `database/sql` standard library with the [`pgx/v5`](https://github.com/jackc/pgx) PostgreSQL driver. This approach gives full control over queries while keeping the code simple and dependency-light.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Database Schema](#database-schema)
  - [Enums](#enums)
  - [Tables](#tables)
  - [Events Table (Event Sourcing)](#events-table-event-sourcing)
  - [Indexes](#indexes)
  - [Migrations](#migrations)
- [Database Connection (`db` Package)](#database-connection-db-package)
  - [Error Mapping](#error-mapping)
  - [NullTime Helper](#nulltime-helper)
  - [Resource Cleanup](#resource-cleanup)
  - [Test Database](#test-database)
- [Repository Layer](#repository-layer)
  - [General Structure](#general-structure)
  - [DB-to-Domain Mapping](#db-to-domain-mapping)
  - [CRUD Repositories](#crud-repositories)
    - [`user_repo`](#user_repo)
    - [`table_repo`](#table_repo)
    - [`product_repo`](#product_repo)
  - [Event Repository (`event_repo`)](#event-repository-event_repo)
- [Mock Repositories](#mock-repositories)
- [Integration Tests](#integration-tests)
- [Dual Persistence Strategy](#dual-persistence-strategy)
- [ORM Evaluation: Is GORM a Good Fit for jotti?](#orm-evaluation-is-gorm-a-good-fit-for-jotti)
  - [What Is an ORM?](#what-is-an-orm)
  - [The Object-Relational Impedance Mismatch](#the-object-relational-impedance-mismatch)
  - [Go and ORMs — A Unique Relationship](#go-and-orms--a-unique-relationship)
  - [GORM Overview](#gorm-overview)
  - [When ORMs ARE the Right Choice](#when-orms-are-the-right-choice)
  - [Dimension-by-Dimension Analysis](#dimension-by-dimension-analysis-gorm-vs-jotti)
  - [Summary](#summary)
  - [Recommendation](#recommendation)

---

## Architecture Overview

The persistence layer follows a **repository pattern** with a clear separation of concerns:

```
HTTP Handler → Application Service → Repository → PostgreSQL
                                         ↕
                                    Domain Model
```

Each repository package:
1. Holds a `Repository` struct with a `*sql.DB` field.
2. Defines private database-specific types (e.g., `dbuser`, `dbtable`) that map to SQL columns.
3. Provides a `toDomain()` method to convert database types into domain models.
4. Executes raw SQL using `database/sql` (`QueryRowContext`, `QueryContext`, `ExecContext`).
5. Maps database errors to domain errors via the shared `db` package.

There are no generated queries, no reflection-based mapping, and no query builders — every SQL statement is written explicitly.

---

## Database Schema

The schema is defined in SQL migration files under `database/migrations/`. The current schema consists of two migrations:

- `01_initial.up.sql` — Creates all tables, enums, indexes, triggers, and seed data.
- `02_add_senior_service_role.up.sql` — Adds `senior_service` to the `UserRole` enum.

### Enums

| Enum              | Values                             | Used by                |
| ----------------- | ---------------------------------- | ---------------------- |
| `UserRole`        | `admin`, `senior_service`, `service` | `users.role`           |
| `EntityStatus`    | `active`, `inactive`, `deleted`    | `users.status`, `tables.status`, `product_variants.status` |
| `ProductCategory` | `food`, `beverage`, `other`        | `products.category`    |

Enums are PostgreSQL custom types created with `CREATE TYPE ... AS ENUM`. They enforce valid values at the database level. The Go domain models mirror these as typed string constants (e.g., `user.AdminRole = "admin"`).

### Tables

#### `users`

Stores system users who perform actions in jotti.

| Column                  | Type            | Nullable | Description                                    |
| ----------------------- | --------------- | -------- | ---------------------------------------------- |
| `id`                    | `INT` (identity) | No       | Primary key, auto-generated                    |
| `name`                  | `TEXT`          | No       | Full name of the user                          |
| `username`              | `TEXT` (unique) | No       | Unique login name                              |
| `password_hash`         | `TEXT`          | Yes      | Argon2id hash; NULL until user sets a password |
| `onetime_password_hash` | `TEXT`          | Yes      | One-time password for onboarding/reset         |
| `role`                  | `UserRole`      | No       | Access role: admin, senior_service, or service |
| `status`                | `EntityStatus`  | No       | active, inactive, or deleted (soft-delete)     |
| `created_at`            | `TIMESTAMPTZ`   | No       | Creation timestamp (UTC)                       |

Indexes: `idx_users_username` (username), `idx_users_status` (status).

#### `tables`

Customers sit at tables and place orders from there.

| Column       | Type            | Nullable | Description                                |
| ------------ | --------------- | -------- | ------------------------------------------ |
| `id`         | `INT` (identity) | No       | Primary key, auto-generated                |
| `name`       | `TEXT` (unique) | No       | Name or number (e.g., "Tisch 1")          |
| `status`     | `EntityStatus`  | No       | active, inactive, or deleted (soft-delete) |
| `created_at` | `TIMESTAMPTZ`   | No       | Creation timestamp (UTC)                   |

Index: `idx_tables_status` (status).

#### `products`

Products that can be ordered by customers.

| Column       | Type              | Nullable | Description                          |
| ------------ | ----------------- | -------- | ------------------------------------ |
| `id`         | `INT` (identity)  | No       | Primary key, auto-generated          |
| `name`       | `TEXT`            | No       | Product name                         |
| `category`   | `ProductCategory` | No       | food, beverage, or other             |
| `created_at` | `TIMESTAMPTZ`     | No       | Creation timestamp (UTC)             |

Products have no status column — their visibility is controlled by the status of their variants.

#### `product_variants`

Variants of products with individual prices (e.g., "Cola 0.3L" and "Cola 0.5L").

| Column       | Type            | Nullable | Description                                    |
| ------------ | --------------- | -------- | ---------------------------------------------- |
| `id`         | `INT` (identity) | No       | Primary key, auto-generated                    |
| `product_id` | `INT` (FK)      | No       | References `products(id)`                      |
| `name`       | `TEXT`          | No       | Variant name (e.g., "0.5L")                   |
| `price_cents`| `INT`           | No       | Price in cents (e.g., 299 for €2.99)          |
| `status`     | `EntityStatus`  | No       | active, inactive, or deleted (soft-delete)     |
| `created_at` | `TIMESTAMPTZ`   | No       | Creation timestamp (UTC)                       |

Index: `idx_product_variants_status` (status).

### Events Table (Event Sourcing)

The `events` table is the core of the event-sourcing system for table operations (orders, payments, deliveries, cancelations). It is **append-only** — rows are only ever inserted and read, never updated or deleted.

| Column      | Type            | Nullable | Description                                     |
| ----------- | --------------- | -------- | ----------------------------------------------- |
| `id`        | `INT` (identity) | No       | Primary key, auto-generated, used for ordering  |
| `user_id`   | `INT` (FK)      | No       | References `users(id)` — the actor               |
| `type`      | `TEXT`          | No       | Event type (e.g., `table.order-placed:v1`)      |
| `subject`   | `TEXT`          | No       | Aggregate key (e.g., `table:42`)                |
| `timestamp` | `TIMESTAMPTZ`   | No       | Event timestamp (UTC)                           |
| `data`      | `JSONB`         | No       | Event payload, versioned by type                |

**Indexes:** `idx_events_user_id`, `idx_events_subject`, `idx_events_type`, `idx_events_subject_type` (composite).

**Immutability enforcement:**

The events table has multiple layers of protection against mutation:

1. **Privilege restriction:** `REVOKE ALL ON TABLE events FROM PUBLIC; GRANT SELECT, INSERT ON TABLE events TO PUBLIC;` — only SELECT and INSERT allowed for non-owner roles.
2. **Trigger-based enforcement** (for ALL roles including table owner):
   - `events_no_update` — `BEFORE UPDATE` trigger raises an exception.
   - `events_no_delete` — `BEFORE DELETE` trigger raises an exception.
   - `events_no_truncate` — `BEFORE TRUNCATE` trigger raises an exception.

All three triggers call `prevent_event_mutation()`, which raises a PostgreSQL exception with the message `events table is append-only: <operation> not allowed`.

**Event types** (defined in `domain/table/events.go`):

| Event Type                          | Description               |
| ----------------------------------- | ------------------------- |
| `table.order-placed:v1`             | An order was placed       |
| `table.payment-registered:v1`       | A payment was registered  |
| `table.variants-canceled:v1`        | Variants were canceled    |
| `table.variants-delivered:v1`       | Variants were delivered   |
| `table.snapshot:v1`                 | A state snapshot          |

### Indexes

The database uses B-tree indexes for common query patterns:

| Index                        | Table              | Column(s)        | Purpose                                      |
| ---------------------------- | ------------------ | ---------------- | --------------------------------------------- |
| `idx_users_username`         | `users`            | `username`       | Fast lookup by username (login)               |
| `idx_users_status`           | `users`            | `status`         | Filter active/inactive users                  |
| `idx_tables_status`          | `tables`           | `status`         | Filter active/inactive tables                 |
| `idx_product_variants_status`| `product_variants` | `status`         | Filter active/inactive variants               |
| `idx_events_user_id`         | `events`           | `user_id`        | Find events by actor                          |
| `idx_events_subject`         | `events`           | `subject`        | Find events for a specific aggregate          |
| `idx_events_type`            | `events`           | `type`           | Find events by type                           |
| `idx_events_subject_type`    | `events`           | `subject`, `type`| Find snapshots for a specific aggregate       |

### Migrations

Migrations are managed with [`golang-migrate`](https://github.com/golang-migrate/migrate) and stored in `database/migrations/`:

```
database/migrations/
  01_initial.up.sql
  01_initial.down.sql
  02_add_senior_service_role.up.sql
  02_add_senior_service_role.down.sql
```

Each migration has an `up` (apply) and `down` (rollback) file. The initial migration wraps all DDL statements in a `BEGIN`/`COMMIT` transaction block.

---

## Database Connection (`db` Package)

The `db` package (`backend/db/`) provides shared utilities for database access. It does **not** manage the connection itself — the `*sql.DB` pool is created in `backend/app/app.go` during application startup and injected into all repositories.

### Error Mapping

The function `db.Error(err)` translates low-level database errors into domain-specific sentinel errors:

| Database Error                        | Mapped To             |
| ------------------------------------- | --------------------- |
| PostgreSQL unique violation (`23505`)  | `db.ErrAlreadyExists` |
| `sql.ErrNoRows`                       | `db.ErrNotFound`      |
| Any other error                       | `db.ErrDatabase`      |

The function `db.ResultError(result)` checks `RowsAffected()` and returns `db.ErrNotFound` if zero rows were affected (used for UPDATE operations).

These sentinel errors allow application services and HTTP handlers to distinguish between "not found" (→ 404), "already exists" (→ 409), and "internal error" (→ 500) without leaking database details.

### NullTime Helper

The `db.NullTime` type is a custom nullable time wrapper that implements:
- `sql.Scanner` — for reading `TIMESTAMPTZ` columns (handles both `NULL` and `time.Time` values).
- `driver.Valuer` — for writing time values to the database.
- `json.Unmarshaler` / `json.Marshaler` — for parsing and encoding time values in JSON.

This is necessary because `sql.NullTime` does not implement `json.Unmarshaler`, which causes issues when scanning JSON results from PostgreSQL's `json_agg()` function. The `product_repo` package uses `json_agg()` to aggregate product variants into JSON arrays, requiring a type that handles both database NULLs and JSON null/timestamp strings.

### Resource Cleanup

The `db.Close(closer, name)` function safely closes an `io.Closer` (typically `*sql.Rows`) and logs any error using `zerolog`. This is used throughout repositories with `defer db.Close(rows, "resource name")`.

### Test Database

The `db.OpenTestDatabase()` function opens a connection to a local PostgreSQL instance (`localhost:5432`, database `jotti`, user/password `admin`) for integration tests. It panics if the connection fails, ensuring fast test failure.

---

## Repository Layer

### General Structure

Each domain has its own repository package under `backend/repository/`:

```
backend/repository/
  event_repo/       # Event sourcing (append-only)
  product_repo/     # Products + variants (CRUD)
  table_repo/       # Tables (CRUD)
  user_repo/        # Users (CRUD)
```

Every repository package contains:

| File            | Purpose                                             |
| --------------- | --------------------------------------------------- |
| `types.go`      | Private DB struct + `toDomain()` converter          |
| `repo.go`       | `Repository` struct + SQL query methods             |
| `mock.go`       | In-memory mock for unit tests                       |
| `repo_test.go`  | Integration tests (`//go:build integration`)        |

The `event_repo` is an exception — it has no `types.go` because event data is stored as `JSONB` and the `event.Event` domain type is used directly.

### DB-to-Domain Mapping

Each CRUD repository defines a private struct that mirrors the database row layout, using `sql.NullString`, `sql.NullTime`, or `db.NullTime` for nullable columns. These structs have a `toDomain()` method that converts them to domain model types:

```go
// user_repo/types.go
type dbuser struct {
    ID                  int            `db:"id"`
    Name                string         `db:"name"`
    Username            string         `db:"username"`
    Role                string         `db:"role"`
    Status              string         `db:"status"`
    PasswordHash        sql.NullString `db:"password_hash"`
    OnetimePasswordHash sql.NullString `db:"onetime_password_hash"`
    CreatedAt           sql.NullTime   `db:"created_at"`
}

func (dp *dbuser) toDomain() user.User {
    return user.User{
        ID:                  dp.ID,
        Name:                dp.Name,
        Username:            dp.Username,
        Role:                user.Role(dp.Role),
        Status:              user.Status(dp.Status),
        PasswordHash:        dp.PasswordHash.String,
        OnetimePasswordHash: dp.OnetimePasswordHash.String,
        CreatedAt:           dp.CreatedAt.Time,
    }
}
```

Note: The `db:"..."` struct tags are purely documentary — they are not used by any reflection-based mapper. All column mapping is done via explicit `rows.Scan()` calls with positional arguments matching the `SELECT` column order.

**Direction of mapping:**
- **Read (DB → Domain):** `Scan()` into db struct → `toDomain()` → return domain model.
- **Write (Domain → DB):** Extract fields from domain model → pass directly as SQL parameters. No intermediate db struct is used for writes.

### CRUD Repositories

The three CRUD repositories (`user_repo`, `table_repo`, `product_repo`) share the same patterns.

#### `user_repo`

Manages the `users` table.

| Method              | SQL Operation   | Description                              |
| ------------------- | --------------- | ---------------------------------------- |
| `GetUser(id)`       | `SELECT ... WHERE id = $1 AND status != 'deleted'` | Single user by ID (excludes soft-deleted) |
| `GetUserByUsername(username)` | `SELECT ... WHERE username = $1 AND status != 'deleted'` | Single user by username (for login) |
| `GetAllUsers()`     | `SELECT ... WHERE status != 'deleted' ORDER BY id ASC` | All non-deleted users (no password hashes) |
| `CreateUser(u)`     | `INSERT ... RETURNING id` | Creates a user, returns generated ID |
| `UpdateUser(u)`     | `UPDATE ... WHERE id = $1` | Updates all mutable fields |

Key details:
- `GetUser` and `GetUserByUsername` scan all columns including password hashes (needed for authentication).
- `GetAllUsers` deliberately omits `password_hash` and `onetime_password_hash` from the SELECT (the listing endpoint does not need secrets).
- All reads filter out `status != 'deleted'` (soft-delete convention).
- `UpdateUser` updates all mutable fields in a single statement. The `db.ResultError()` check on `RowsAffected` returns `ErrNotFound` if the ID does not exist.

#### `table_repo`

Manages the `tables` table.

| Method              | SQL Operation   | Description                              |
| ------------------- | --------------- | ---------------------------------------- |
| `GetTable(id)`      | `SELECT ... WHERE id = $1 AND status != 'deleted'` | Single table by ID |
| `GetAllTables()`    | `SELECT ... WHERE status != 'deleted' ORDER BY id ASC` | All non-deleted tables |
| `GetActiveTables()` | `SELECT ... WHERE status = 'active' ORDER BY id ASC` | Only active tables (for service UI) |
| `CreateTable(t)`    | `INSERT ... RETURNING id` | Creates a table, returns generated ID |
| `UpdateTable(t)`    | `UPDATE ... SET name, status WHERE id = $1` | Updates name and status |

#### `product_repo`

Manages the `products` and `product_variants` tables. This repository is more complex because products and their variants are fetched together using PostgreSQL JSON aggregation.

**Product operations:**

| Method                | SQL Operation   | Description                              |
| --------------------- | --------------- | ---------------------------------------- |
| `GetProduct(id)`      | `SELECT ... WITH json_agg subquery` | Single product with non-deleted variants |
| `GetAllProducts()`    | `SELECT ... WITH CTE + LEFT JOIN` | All products with non-deleted variants |
| `GetActiveProducts()` | `SELECT ... WITH CTE + INNER JOIN` | Products that have at least one active variant |
| `CreateProduct(p)`    | `INSERT ... RETURNING id` | Creates a product, returns generated ID |
| `UpdateProduct(p)`    | `UPDATE ... SET name, category WHERE id = $1` | Updates name and category |

**Variant operations** (in `variant.go`):

| Method                   | SQL Operation   | Description                              |
| ------------------------ | --------------- | ---------------------------------------- |
| `GetVariant(id)`         | `SELECT ... WHERE id = $1 AND status != 'deleted'` | Single variant by ID |
| `CreateVariant(pid, v)`  | `INSERT ... RETURNING id` | Creates a variant for a product |
| `UpdateVariant(v)`       | `UPDATE ... SET name, price_cents, status WHERE id = $1` | Updates variant fields |

**JSON aggregation pattern:**

Products are always fetched with their variants using PostgreSQL's `json_agg()` and `json_build_object()` functions. This avoids the N+1 query problem:

```sql
-- GetAllProducts uses a CTE for efficient aggregation
WITH variant_json AS (
    SELECT product_id,
           json_agg(json_build_object(
               'id', id, 'name', name,
               'price_cents', price_cents,
               'status', status, 'created_at', created_at
           )) AS variants
    FROM product_variants
    WHERE status != 'deleted'
    GROUP BY product_id
)
SELECT p.id, p.name, p.category, p.created_at,
       COALESCE(vj.variants, '[]') AS variants
FROM products p
LEFT JOIN variant_json vj ON vj.product_id = p.id
ORDER BY p.id ASC
```

The resulting JSON array is scanned as `[]byte` and unmarshaled into `[]dbvariant` structs using `json.Unmarshal`. The `COALESCE(..., '[]')` ensures products without variants return an empty JSON array instead of NULL.

`GetActiveProducts` uses `INNER JOIN` instead of `LEFT JOIN` to exclude products that have no active variants at all.

### Event Repository (`event_repo`)

The event repository handles the append-only `events` table used for event sourcing. Unlike the CRUD repositories, it has no `types.go` — the `event.Event` domain model is scanned directly because the `data` column (`JSONB`) is stored as `json.RawMessage` in the domain model.

| Method                                  | SQL Operation   | Description                              |
| --------------------------------------- | --------------- | ---------------------------------------- |
| `WriteEvent(e)`                         | `INSERT ... RETURNING id` | Appends a new event |
| `ReadEvent(id)`                         | `SELECT ... WHERE id = $1` | Single event by ID |
| `ReadEventsBySubject(subject)`          | `SELECT ... WHERE subject = $1 ORDER BY id ASC` | All events for an aggregate |
| `ReadEventsSinceID(subject, fromID)`    | `SELECT ... WHERE subject = $1 AND id >= $2 ORDER BY id ASC` | Events since a given ID (inclusive) |
| `GetLastSnapshotID(subject, type)`      | `SELECT COALESCE(MAX(id), 0) ... WHERE subject AND type` | ID of the most recent snapshot |
| `ReadEventsWithSnapshot(subject, type)` | CTE-based query | Events from the last snapshot onwards |

**Snapshot optimization:**

The `ReadEventsWithSnapshot` method uses a CTE to find the latest snapshot event and only read events from that point forward. This avoids replaying the entire event history:

```sql
WITH last_snapshot AS (
    SELECT COALESCE(MAX(id), 0) AS id
    FROM events
    WHERE subject = $1 AND type = $2
)
SELECT e.id, e.user_id, e.type, e.subject, e.data, e.timestamp
FROM events e, last_snapshot ls
WHERE e.subject = $1 AND e.id >= ls.id
ORDER BY e.id ASC
```

If no snapshot exists, `COALESCE(MAX(id), 0)` returns 0, and all events are included (since all IDs are ≥ 0).

**Shared `scanEvents` helper:**

The private `scanEvents(rows)` function iterates over `*sql.Rows` and builds a `[]event.Event` slice. It is reused by `ReadEventsBySubject`, `ReadEventsSinceID`, and `ReadEventsWithSnapshot`.

---

## Mock Repositories

Each repository package provides a `NewMock(items, err)` constructor that returns an in-memory implementation of the repository interface. Mocks are used in **unit tests** for application services and HTTP handlers, avoiding the need for a database.

Pattern:
```go
// event_repo/mock.go
func NewMock(events []event.Event, err error) *mockRepo {
    eventMap := make(map[int]event.Event)
    for _, e := range events {
        eventMap[e.ID] = e
    }
    return &mockRepo{events: eventMap, err: err}
}
```

The mock stores data in a map and returns the configured error on all operations. This enables testing both success and error scenarios. The mocks implement the same interface as the real repositories but are not bound to `//go:build` tags, so they are always available.

---

## Integration Tests

Each repository has an integration test file with the `//go:build integration` build tag. These tests run against a real PostgreSQL instance and follow a consistent pattern:

1. **Setup:** Open a test database connection, clean relevant tables.
2. **Seed:** Insert test data using repository methods.
3. **Assert:** Call the method under test and verify results.
4. **Teardown:** Clean up tables and close the connection.

```go
func setup(t *testing.T) (Repository, func(t *testing.T)) {
    db := dbpkg.OpenTestDatabase()
    _, err := db.Exec("DELETE FROM tables")
    // ...
    return Repository{DB: db}, func(t *testing.T) {
        _, err = db.Exec("DELETE FROM tables")
        // ...
        db.Close()
    }
}
```

The event repository tests require special handling: before deleting events, the `events_no_delete` trigger must be temporarily disabled (`ALTER TABLE events DISABLE TRIGGER events_no_delete`) and re-enabled afterwards.

Run integration tests with:
```bash
cd backend && go test -tags=integration -race ./...
```

This requires a running PostgreSQL instance with applied migrations (see [development.md](development.md)).

---

## Dual Persistence Strategy

jotti uses two fundamentally different persistence strategies:

### 1. CRUD for Master Data

**Users**, **tables**, **products**, and **product variants** are stored in traditional relational tables with standard Create/Read/Update operations. Deletes are soft-deletes via `status = 'deleted'` — records are never physically removed. All read queries filter out deleted records (`WHERE status != 'deleted'`).

### 2. Event Sourcing for Table Operations

**Orders**, **payments**, **deliveries**, and **cancelations** are stored as immutable events in the `events` table. The current state of a table is reconstructed by replaying its events:

- **Balance** = Σ(order totals) − Σ(payment totals) − Σ(cancelation totals)
- **Unpaid variants** = ordered − paid − canceled
- **Undelivered variants** = ordered − delivered − canceled

Snapshots (`table.snapshot:v1` events) periodically capture the full state, allowing `ReadEventsWithSnapshot` to skip replaying older events.

This dual approach gives the best of both worlds: simple CRUD for rarely-changing master data, and a complete audit trail with temporal queries for the operational core.

---

## ORM Evaluation: Is GORM a Good Fit for jotti?

This section evaluates whether adopting an ORM — specifically [GORM](https://github.com/go-gorm/gorm), the most popular ORM in the Go ecosystem — would be beneficial for jotti. The evaluation draws on established ORM theory, the specifics of Go as a language, GORM's feature set, and a detailed analysis of how each maps onto jotti's actual codebase.

### What Is an ORM?

Object-Relational Mapping (ORM) is a programming technique for converting data between a relational database and the heap memory of a programming language. It creates, in effect, a virtual object database that can be used from within the program. ([Wikipedia: Object–relational mapping](https://en.wikipedia.org/wiki/Object%E2%80%93relational_mapping))

The core challenge is that object-oriented programs and relational databases model data differently:

- **Objects** are nodes in a directed graph — they reference each other via pointers, support encapsulation, and can form deep hierarchies.
- **Relational tuples** are rows in tables linked by foreign keys, operated on by relational algebra, and shared across concurrent processes.

An ORM bridges this gap by providing:

- **Automatic CRUD generation** — `Create()`, `Find()`, `Update()`, `Delete()` methods derived from struct/class definitions.
- **Schema-to-struct mapping** — Annotations (tags, decorators) define column names, types, and constraints; the ORM handles scanning and parameter binding.
- **Relationship management** — Associations (has-one, has-many, belongs-to, many-to-many) with eager/lazy loading.
- **Auto-migrations** — The ORM detects differences between struct definitions and the database schema and generates DDL to synchronize them.
- **Query builder** — Chainable API for building queries programmatically (`Where()`, `Order()`, `Limit()`, `Joins()`).
- **Hooks/callbacks** — Lifecycle events (before/after create, update, delete) for cross-cutting concerns.
- **Transaction management** — Simplified API for wrapping operations in database transactions.

As noted by [AWS](https://aws.amazon.com/what-is/object-relational-mapping/): *"Because the mappings are abstracted, if the database structure ever changes or you migrate to a new database, the ORM can still point to the correct data with minimal updates."* This database portability is one of the primary selling points.

Common ORM frameworks include Hibernate (Java), Entity Framework (.NET/C#), SQLAlchemy (Python), and GORM (Go).

**Trade-offs.** ORMs reduce boilerplate and can accelerate development, but — as Wikipedia notes — *"disadvantages of ORM tools generally stem from the high level of abstraction obscuring what is actually happening in the implementation code."* Furthermore, *"ORMs are limited to their predefined functionality, which may not cover all edge cases or database features. They usually mitigate this limitation by providing users with an interface to write raw queries."*

### The Object-Relational Impedance Mismatch

The fundamental theoretical problem behind ORMs is the **object-relational impedance mismatch** — a set of conceptual and technical difficulties when mapping between relational data stores and object-oriented domain models. ([Wikipedia: Object–relational impedance mismatch](https://en.wikipedia.org/wiki/Object%E2%80%93relational_impedance_mismatch))

Key dimensions of the mismatch:

| Dimension | Objects (OO) | Relational (SQL) |
| --- | --- | --- |
| **Structure** | Directed graph, nesting, composition | Flat tuples, foreign keys, no nesting |
| **Identity** | Memory address / reference equality | Primary key equality |
| **Encapsulation** | Private fields, public interfaces | All columns accessible |
| **Inheritance** | Class hierarchies, polymorphism | No native concept (workarounds exist) |
| **Lifecycle** | Garbage collected on heap | Row insertion/deletion with transactions |
| **Concurrency** | Single-process, in-memory | Multi-process, shared with locking |
| **Types** | Language-specific (string, int, etc.) | SQL-specific (VARCHAR, NUMERIC, JSONB) |
| **Manipulation** | Per-class imperative methods | Standardized declarative operators (SQL) |

ORMs attempt to bridge all of these gaps. The more dimensions that actually differ in your application, the more value — and friction — an ORM provides.

**A critical insight from the impedance mismatch literature:** *"Since the ORM is often specified in terms of configuration, annotations, and restricted domain-specific languages, it lacks the flexibility of a full programming language to resolve the impedance mismatch."* This means that when the ORM's abstractions fail to cover your use case, you fall back to raw SQL — but now inside a framework that was designed to hide SQL.

### Go and ORMs — A Unique Relationship

ORMs were originally designed for traditionally object-oriented languages like Java, C#, and Python, where the impedance mismatch is most pronounced: deep class hierarchies, inheritance, polymorphism, and encapsulated state all clash with flat relational tables.

**Go is fundamentally different.** Go is not a traditionally object-oriented language:

- **No classes or inheritance.** Go uses structs and composition, not class hierarchies.
- **No polymorphism via inheritance.** Polymorphism is achieved via interfaces — small, implicit contracts.
- **No constructors or destructors.** Structs are plain data with optional methods.
- **Flat data structures.** Go structs are essentially named tuples — semantically close to database rows.
- **Explicit error handling.** No exceptions; every error is a returned value.
- **No hidden behavior.** No implicit getters/setters, no proxy objects, no lazy-loading magic.

This means the impedance mismatch is **inherently smaller in Go** than in Java or C#. A Go struct is already close to a database tuple. The "bridging" that an ORM provides is less necessary because the gap is narrower.

This is why the Go ecosystem has gravitated toward SQL-first tools (`sqlc`, `pgx`, `scany`) rather than heavy ORMs. GORM is popular, but it is often criticized in the Go community for introducing implicit behavior that conflicts with Go's philosophy of explicitness.

### GORM Overview

[GORM](https://github.com/go-gorm/gorm) is Go's most widely adopted ORM (36k+ GitHub stars). Its features include:

| Feature | Description |
| --- | --- |
| Struct-based models | Define tables as Go structs with `gorm:"..."` tags |
| Auto-migrations | `db.AutoMigrate(&User{})` creates/alters tables to match structs |
| Associations | Has One, Has Many, Belongs To, Many To Many with Preload/Joins |
| Hooks | Before/After Create/Save/Update/Delete/Find callbacks |
| Transactions | `db.Transaction(func(tx) error { ... })` with nested savepoints |
| SQL builder | Raw SQL support alongside the query builder (`db.Raw()`) |
| Soft deletes | Built-in `gorm.Model` includes `DeletedAt` for soft deletes |
| Batch operations | Batch insert, find-in-batches |
| Plugin system | Database resolver, Prometheus, custom plugins |

**Typical GORM usage:**

```go
type User struct {
    gorm.Model          // ID, CreatedAt, UpdatedAt, DeletedAt
    Name     string
    Username string `gorm:"uniqueIndex"`
    Role     string
}

// Create
db.Create(&User{Name: "Alice", Username: "alice", Role: "admin"})

// Read
var user User
db.Where("username = ?", "alice").First(&user)

// Update
db.Model(&user).Update("role", "service")

// Preload associations
db.Preload("Variants").Find(&products)
```

**GORM's strengths:**
- Significantly reduces CRUD boilerplate for standard entities.
- Association handling (Preload, Joins) solves the N+1 problem for common patterns.
- Built-in soft deletes, timestamps, and hooks for cross-cutting concerns.
- Large community and ecosystem of plugins.

**GORM's known limitations:**
- **Implicit behaviors.** Method chaining can silently accumulate conditions. Session management (new session vs. shared) is a common source of bugs.
- **Reflection-heavy.** GORM relies on runtime reflection for struct-to-column mapping, which obscures errors until runtime rather than compile time.
- **Raw SQL fallback.** Complex queries (CTEs, window functions, JSON aggregation) require `db.Raw()`, breaking the abstraction.
- **PostgreSQL-specific features.** Custom enum types, JSONB operations, and triggers all need manual handling outside GORM's managed scope.
- **Domain model coupling.** GORM encourages using the same struct for both ORM mapping and business logic, which conflicts with clean domain separation.

### When ORMs ARE the Right Choice

Before evaluating GORM against jotti, it is important to acknowledge the kinds of projects where ORMs provide clear value:

1. **Many entities with simple CRUD.** Applications with 20+ tables and predominantly straightforward Create/Read/Update/Delete operations benefit enormously from ORM-generated queries.
2. **Rapid prototyping.** When schema and requirements are in flux, auto-migrations and convention-over-configuration accelerate development.
3. **Database portability.** If the application must support multiple database engines (PostgreSQL, MySQL, SQLite), the ORM abstracts away dialect differences.
4. **Large teams with mixed skill levels.** ORMs provide a uniform API that reduces the need for deep SQL expertise across the team.
5. **Traditional OOP languages.** In Java or C# with deep class hierarchies and rich entity relationships, ORMs bridge a wide impedance gap.
6. **Complex relationship graphs.** Applications with many-to-many relationships, polymorphic associations, and deep eager/lazy loading patterns benefit from ORM relationship management.

### Dimension-by-Dimension Analysis: GORM vs. jotti

The following analysis evaluates GORM against jotti's actual codebase. The repository layer totals **653 lines** across 8 files (including types and variant files), with **7 direct dependencies** in `go.mod`.

#### 1. CRUD Repositories (user_repo, table_repo)

**Current state:** `user_repo` (repo.go: 84 lines, types.go: 35 lines) and `table_repo` (repo.go: 86 lines, types.go: 28 lines) implement simple single-table CRUD with soft deletes. Each has a private `db*` struct, a `toDomain()` converter, and 5 methods with straightforward SQL.

**What GORM would change:**
- Eliminate the `db*` adapter structs — GORM scans directly into annotated structs.
- Replace hand-written `INSERT ... RETURNING id` and `SELECT ... WHERE` with `db.Create()` and `db.Where().First()`.
- Built-in soft-delete support via `gorm.DeletedAt` could partially match jotti's `status = 'deleted'` pattern.

**Verdict:** GORM would reduce boilerplate here, but the savings are modest (~40–50 lines per repo). The current code is already simple and readable. This is where GORM provides its **best case** benefit for jotti.

**Friction point:** jotti uses a **three-state `EntityStatus` enum** (active/inactive/deleted) rather than GORM's binary deleted-at timestamp. GORM's soft-delete model assumes two states (present or soft-deleted via a nullable `DeletedAt` timestamp). Adapting this to jotti's three-state model would require custom scopes or overriding GORM's default behavior — partially negating the convenience. This is a small but concrete example of the impedance mismatch: jotti's domain model does not align with GORM's built-in abstractions.

#### 2. Product Repository (product_repo) — JSON Aggregation

**Current state:** `product_repo` (repo.go: 191 lines, types.go: 47 lines, variant.go: 48 lines = 286 lines total) is the most complex repository. It uses PostgreSQL-specific features:
- `json_agg()` + `json_build_object()` for denormalizing variants into parent products.
- Common Table Expressions (CTEs) for efficient aggregation.
- `COALESCE(..., '[]')` for handling products with no variants.
- Custom `db.NullTime` type to handle JSON timestamp unmarshaling from `json_agg()` output.

**What GORM would change:**
- The `Preload("Variants")` feature could replace the JSON aggregation approach. GORM would issue two queries (one for products, one for variants) and stitch them together in memory.
- The CTE-based single-query approach would be replaced with GORM's N+1-safe but two-query preload.

**Verdict:** This is where GORM would cause a **regression**. The current implementation is a deliberate optimization: a single SQL query returns products with pre-aggregated variants using PostgreSQL's `json_agg()`. This is exactly the kind of database-specific feature that the impedance mismatch literature warns about — an ORM cannot replicate it because it is a relational/SQL capability with no object-oriented equivalent. To use the current SQL with GORM, you would need `db.Raw()`, at which point you are writing raw SQL inside an ORM — losing the abstraction benefit while keeping the dependency cost.

#### 3. Event Repository (event_repo) — Event Sourcing

**Current state:** `event_repo` (repo.go: 134 lines) implements an append-only event store with:
- `WriteEvent()` — simple INSERT.
- `ReadEventsWithSnapshot()` — CTE-based query that finds the latest snapshot and reads forward.
- Direct scanning into the `event.Event` domain model (no adapter struct needed because `data` is `json.RawMessage`).
- Database-level immutability enforcement via triggers (UPDATE/DELETE/TRUNCATE prevented).

**What GORM would change:**
- GORM fundamentally assumes **mutable records** — its entire lifecycle model is Create → Update → Delete. An append-only event store contradicts this assumption.
- GORM's auto-migration would conflict with the trigger-based immutability enforcement (triggers that prevent UPDATE/DELETE/TRUNCATE on the events table).
- The CTE-based snapshot query has no GORM query-builder equivalent — it would require `db.Raw()`.
- GORM's soft-delete feature is meaningless for an event store where records are never deleted.

**Verdict:** GORM is a **fundamentally poor fit** for event sourcing. This is not a minor friction point — it is a category mismatch. The Wikipedia article on impedance mismatch notes that relational databases use *"standardized operators for data manipulation"*, while ORMs layer OO lifecycle semantics on top. Event sourcing rejects the mutable-lifecycle assumption entirely. Using GORM here would mean constantly bypassing the framework for the operational core of the application.

#### 4. Domain Model Separation (Data Mapper vs. Active Record)

**Current state:** jotti implements the **Data Mapper pattern** — a well-established architectural pattern where the domain model and the database schema are independently designed, and a separate mapper layer translates between them. Each repository has private `db*` structs (e.g., `dbuser`, `dbtable`, `dbproduct`, `dbvariant`) that mirror database columns, with `toDomain()` methods that convert to domain models (`user.User`, `table.Table`, etc.).

This is distinct from the **Active Record pattern** — where domain objects contain both business logic and database access methods (e.g., `user.Save()`, `user.FindByUsername()`). GORM encourages the Active Record pattern by design.

**What GORM would change:**
- GORM encourages using the same struct for both ORM mapping and business logic (via `gorm:"..."` tags on domain structs). This is the Active Record approach.
- To maintain jotti's Data Mapper separation, you would need either:
  - (a) Add `gorm:"..."` tags to domain models — mixing persistence concerns into the domain layer.
  - (b) Keep separate ORM models and domain models — which retains the current adapter overhead while adding ORM complexity on top.

**Verdict:** jotti's architecture explicitly follows the Data Mapper pattern, cleanly separating the `domain/` and `repository/` packages. GORM's Active Record design pushes toward coupling them. The Wikipedia impedance mismatch article observes that ORMs often *"expose the properties publicly to work with database columns"*, conflicting with encapsulation — and jotti's `db*` structs exist precisely to preserve this boundary. Maintaining the Data Mapper separation while using GORM would negate most of its convenience.

#### 5. PostgreSQL-Specific Features

**Current state:** jotti leverages several PostgreSQL-specific features:
- Custom enum types (`UserRole`, `EntityStatus`, `ProductCategory`) enforced at the database level.
- `JSONB` columns with `json_agg()` / `json_build_object()`.
- Trigger-based immutability enforcement (three triggers on the events table).
- `IDENTITY` columns (SQL standard auto-increment).
- `TIMESTAMPTZ` with timezone awareness.
- Privilege restriction (`REVOKE ALL ... GRANT SELECT, INSERT`) for defense-in-depth.

**What GORM would change:**
- GORM supports PostgreSQL via the `gorm.io/driver/postgres` driver, but many PostgreSQL-specific features require raw SQL or custom data types.
- Custom enums need manual handling (GORM does not natively manage PostgreSQL `CREATE TYPE ... AS ENUM`).
- JSONB operations require the `datatypes` plugin or raw expressions.
- Triggers and privilege restrictions are outside GORM's scope entirely.

**Verdict:** The AWS article highlights database portability as a key ORM benefit: *"if the database structure ever changes or you migrate to a new database, the ORM can still point to the correct data with minimal updates."* However, jotti is deeply committed to PostgreSQL — custom enums, JSONB, triggers, CTEs, and privilege restrictions are all PostgreSQL-specific features that would not transfer to another database engine. Database portability is not a goal, eliminating one of the primary ORM selling points.

#### 6. Error Handling

**Current state:** The `db` package (65 lines) provides a focused error-mapping layer:
- `db.Error()` translates PostgreSQL error codes (e.g., unique violation `23505`) into domain sentinel errors (`ErrNotFound`, `ErrAlreadyExists`, `ErrDatabase`).
- `db.ResultError()` checks `RowsAffected()` for UPDATE operations.

**What GORM would change:**
- GORM has its own error handling (`gorm.ErrRecordNotFound`, etc.) that would replace the custom error mapping.
- However, mapping from GORM errors to jotti's domain errors would still be needed in application services or handlers.

**Verdict:** Neutral. The error-mapping layer would change form but not disappear. GORM's error handling is adequate but not simpler for jotti's use case.

#### 7. Testing & Mocking

**Current state:** Each repository has a `mock.go` file implementing the repository interface with in-memory data. Unit tests use these mocks; integration tests run against a real PostgreSQL instance. The interface-based mocking is clean and idiomatic Go.

**What GORM would change:**
- GORM-based repositories would be harder to mock because they depend on `*gorm.DB` rather than a simple interface.
- Options: use `go-sqlmock` to mock the SQL driver, or maintain the current interface-based mocking pattern with GORM behind the interface.

**Verdict:** The current approach is cleaner. Go's strength is its interface-based composition — jotti's repository interfaces enable simple, focused mocks without any database driver dependency. GORM would not improve this.

#### 8. Codebase Size & Complexity

| Metric | Current | With GORM (estimated) |
| --- | --- | --- |
| Repository code | 653 lines (8 files) | ~480 lines (CRUD shrinks, event_repo + product_repo mostly unchanged) |
| Adapter structs | 3 types.go files | 0–2 files (GORM tags on models or separate ORM models) |
| SQL statements | ~29 explicit queries | ~12 (CRUD eliminated, complex queries remain as `db.Raw()`) |
| Direct dependencies | 7 | 9 (+`gorm` + `gorm/driver/postgres`) |
| Learning curve | SQL + pgx | SQL + pgx + GORM API + GORM session model + GORM conventions |
| db package utilities | 3 files, ~160 lines | Partially replaced, partially retained |

The net reduction is roughly **~170 lines** of repository code — at the cost of two additional dependencies, a split between ORM-managed and raw-SQL code paths, and a steeper learning curve that now spans two paradigms.

#### 9. Dependency Philosophy

jotti currently has **7 direct dependencies**, all serving clearly scoped purposes:

| Dependency | Purpose |
| --- | --- |
| `pgx/v5` | PostgreSQL driver |
| `zerolog` | Structured logging |
| `zog` | Validation schemas |
| `golang-jwt/v5` | JWT authentication |
| `google/uuid` | UUID generation |
| `golang.org/x/crypto` | Argon2id password hashing |
| `golang.org/x/time` | Rate limiting |

Adding GORM would introduce a major dependency (~20k lines of code including the driver) that brings its own conventions, session management model, implicit behaviors, and reflection-based runtime. This would be the largest dependency in the project by a significant margin, and it would only be used for a subset of the persistence layer (the CRUD repositories).

### Summary

| Criterion | GORM Benefit | jotti Fit |
| --- | --- | --- |
| Simple CRUD boilerplate reduction | ✅ Moderate | ⚠️ Savings real but small (~170 lines) for 2 of 4 repos |
| JSON aggregation queries | ❌ Cannot replace | ❌ Requires `db.Raw()` fallback |
| Event sourcing (append-only) | ❌ Fundamentally incompatible | ❌ Contradicts GORM's mutable-lifecycle model |
| Domain model separation | ❌ Active Record vs. Data Mapper | ❌ Conflicts with jotti's architecture |
| PostgreSQL-specific features | ❌ Limited support | ❌ Raw SQL needed for enums, triggers, JSONB |
| Soft-delete (three-state enum) | ⚠️ Partial mismatch | ⚠️ Custom scopes needed for active/inactive/deleted |
| Testing & mocking | ❌ More complex | ❌ Current interface-based mocking is simpler |
| Auto-migrations | ✅ Convenient for CRUD | ⚠️ Conflicts with trigger-based immutability |
| Developer onboarding | ✅ Familiar API | ⚠️ Must learn GORM quirks + still write raw SQL |
| Database portability | ✅ Core ORM benefit | ❌ Not a goal — jotti is committed to PostgreSQL |
| Dependency footprint | ❌ Large dependency | ❌ jotti is intentionally dependency-light (7 deps) |
| Go idiom alignment | ❌ Reflection + implicit behavior | ❌ Conflicts with Go's explicit, simple philosophy |
| Impedance mismatch severity | ✅ Bridges OO↔relational gap | ⚠️ Gap is small in Go — structs ≈ tuples |

### Recommendation

**Do not adopt GORM for jotti.**

After studying ORM theory, the object-relational impedance mismatch, GORM's specific feature set, and jotti's codebase in detail, the conclusion is clear: GORM is not a good fit for this project.

**The core reasoning:**

1. **The impedance mismatch is small in Go.** ORMs were designed to bridge the gap between deeply object-oriented languages (Java, C#) and relational databases. Go structs are flat, composable data types — essentially named tuples — that map naturally to database rows. The `db*` struct → `toDomain()` pattern in jotti's repositories is explicit, lightweight, and idiomatic. The gap that GORM would bridge barely exists.

2. **Event sourcing fundamentally contradicts ORM assumptions.** The `event_repo` handles the operational core of jotti (orders, payments, deliveries, cancelations). It is append-only, uses CTE-based snapshot optimization, and is protected by database triggers that prevent mutation. GORM's entire model — Create/Update/Delete lifecycle, auto-migration, soft deletes — assumes mutable records. Using GORM here would mean fighting the framework on every operation.

3. **jotti's most complex SQL cannot be expressed in GORM.** The `product_repo` uses PostgreSQL's `json_agg()` with CTEs to efficiently fetch products with their variants in a single query. This is a deliberate optimization that has no GORM query-builder equivalent. The event repo's `ReadEventsWithSnapshot()` uses a CTE to find the latest snapshot event. Both would require `db.Raw()` — writing raw SQL inside an ORM, paying the dependency cost while getting none of the abstraction benefit.

4. **jotti's Data Mapper architecture conflicts with GORM's Active Record design.** jotti cleanly separates `domain/` (business logic, validation, typed constants) from `repository/` (SQL, scanning, database types). GORM pushes toward placing `gorm:"..."` tags on domain structs, coupling persistence to the domain layer. Maintaining the current separation while using GORM would require keeping separate ORM models — adding complexity without eliminating the adapter layer.

5. **Database portability is irrelevant.** jotti is committed to PostgreSQL. Custom enum types, JSONB, triggers, CTEs, privilege restrictions — these are all PostgreSQL-specific features integral to jotti's design. The primary selling point of ORMs, as described by [AWS](https://aws.amazon.com/what-is/object-relational-mapping/), is that you can *"migrate to a new database"* with *"minimal updates"*. jotti has no such requirement.

6. **The savings do not justify the cost.** GORM would reduce ~170 lines of CRUD boilerplate across the `user_repo` and `table_repo`. In exchange, it adds a major dependency (~20k lines), introduces a split architecture (ORM for CRUD, raw SQL for everything else), requires learning GORM's session model and implicit behaviors, and complicates mocking. The cost-benefit ratio is unfavorable.

**If targeted boilerplate reduction is desired**, consider lightweight alternatives that complement jotti's SQL-first approach:

- **[sqlc](https://sqlc.dev/)** — Generates type-safe Go code from SQL queries at compile time. Eliminates manual `Scan()` calls and adapter structs while keeping full SQL control. Best fit for jotti's philosophy: write SQL, get type-safe Go code.
- **[scany](https://github.com/georgysavva/scany)** — A lightweight library that automates `rows.Scan()` into structs using field tags. No query builder, no migrations, no lifecycle hooks — just less scanning boilerplate. Compatible with `pgx/v5`.
- **[pgx` built-in features](https://github.com/jackc/pgx)** — `pgx/v5` already supports `pgx.CollectRows()` and `pgx.RowToStructByName()` for automated struct scanning. jotti could use these directly without any new dependency.

All three options preserve jotti's SQL-first, PostgreSQL-native, dependency-light approach while addressing the only real GORM benefit: reduced CRUD boilerplate.
