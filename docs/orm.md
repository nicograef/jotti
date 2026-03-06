# Database & Backend Persistence

This document describes how jotti implements database access and persistence. jotti uses **[sqlc](https://sqlc.dev/)** — a compile-time SQL code generator — to produce type-safe Go code from raw SQL queries. Queries are defined in `.sql` files, validated against the PostgreSQL schema at generation time, and compiled into Go functions with full type safety. The generated code uses the `database/sql` standard library with the [`pgx/v5`](https://github.com/jackc/pgx) PostgreSQL driver. This approach gives full control over queries, eliminates hand-written scanning boilerplate, provides compile-time SQL validation, and keeps the code dependency-light (sqlc has zero runtime dependencies).

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
  - [Dimension-by-Dimension Analysis (GORM)](#dimension-by-dimension-analysis-gorm-vs-jotti)
  - [Summary (GORM)](#summary)
  - [Recommendation (GORM)](#recommendation)
- [sqlc Evaluation: Is sqlc a Good Fit for jotti?](#sqlc-evaluation-is-sqlc-a-good-fit-for-jotti)
  - [What Is sqlc?](#what-is-sqlc)
  - [How sqlc Works](#how-sqlc-works)
  - [sqlc Overview](#sqlc-overview)
  - [Dimension-by-Dimension Analysis (sqlc)](#dimension-by-dimension-analysis-sqlc-vs-jotti)
  - [Summary (sqlc)](#summary-sqlc)
  - [Recommendation (sqlc)](#recommendation-sqlc)
- [sqlx Evaluation: Is sqlx a Good Fit for jotti?](#sqlx-evaluation-is-sqlx-a-good-fit-for-jotti)
  - [What Is sqlx?](#what-is-sqlx)
  - [sqlx Overview](#sqlx-overview)
  - [Dimension-by-Dimension Analysis (sqlx)](#dimension-by-dimension-analysis-sqlx-vs-jotti)
  - [Summary (sqlx)](#summary-sqlx)
  - [Recommendation (sqlx)](#recommendation-sqlx)
- [Comprehensive Comparison: All Four Approaches](#comprehensive-comparison-all-four-approaches)
  - [Approach Profiles](#approach-profiles)
  - [Feature Matrix](#feature-matrix)
  - [Fit Analysis for jotti](#fit-analysis-for-jotti)
  - [Final Recommendation](#final-recommendation)

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

Each domain has its own repository package under `backend/repository/`, with SQL queries defined in `backend/sqlc/queries/` and generated code in `backend/sqlc/dbgen/`:

```
backend/
  sqlc.yaml                 # sqlc configuration
  sqlc/queries/             # SQL query definitions
    users.sql               # User queries
    tables.sql              # Table queries
    products.sql            # Product + variant queries
    events.sql              # Event sourcing queries
  sqlc/dbgen/               # Generated code (DO NOT EDIT)
    db.go                   # DBTX interface + Queries struct
    models.go               # Database model types + enums
    users.sql.go            # Generated user query functions
    tables.sql.go           # Generated table query functions
    products.sql.go         # Generated product + variant query functions
    events.sql.go           # Generated event query functions
  repository/
    event_repo/             # Event sourcing (append-only)
    product_repo/           # Products + variants (CRUD)
    table_repo/             # Tables (CRUD)
    user_repo/              # Users (CRUD)
```

Every repository package contains:

| File            | Purpose                                                            |
| --------------- | ------------------------------------------------------------------ |
| `types.go`      | `Repository` struct, `NewRepository()` constructor, type converters |
| `repo.go`       | Repository methods wrapping sqlc-generated query functions          |
| `mock.go`       | In-memory mock for unit tests                                      |
| `repo_test.go`  | Integration tests (`//go:build integration`)                       |

The `event_repo` is an exception — it has no separate `types.go` because event data is stored as `JSONB` and the `event.Event` domain type is used directly.

### DB-to-Domain Mapping

sqlc generates type-safe Go structs for each query result (e.g., `dbgen.GetUserRow`, `dbgen.GetAllUsersRow`). Each repository defines thin converter functions that map these generated structs to domain model types:

```go
// user_repo/types.go
func userRowToDomain(row dbgen.GetUserRow) user.User {
    return user.User{
        ID:                  row.ID,
        Name:                row.Name,
        Username:            row.Username,
        Role:                user.Role(row.Role),
        Status:              user.Status(row.Status),
        PasswordHash:        row.PasswordHash.String,
        OnetimePasswordHash: row.OnetimePasswordHash.String,
        CreatedAt:           row.CreatedAt,
    }
}
```

The `Repository` struct wraps sqlc's `Queries` struct, which is initialized via `NewRepository()`:

```go
type Repository struct {
    DB *sql.DB
    q  *dbgen.Queries
}

func NewRepository(db *sql.DB) Repository {
    return Repository{DB: db, q: dbgen.New(db)}
}
```

**Direction of mapping:**
- **Read (DB → Domain):** sqlc-generated function scans row → converter maps to domain model.
- **Write (Domain → DB):** Extract fields from domain model → pass as sqlc `Params` struct → sqlc-generated function executes query.

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

---

## sqlc Evaluation: Is sqlc a Good Fit for jotti?

This section evaluates whether adopting [sqlc](https://github.com/sqlc-dev/sqlc) — a SQL compiler that generates type-safe Go code from SQL queries — would be beneficial for jotti. The evaluation applies the same dimension-by-dimension methodology used for the GORM analysis above.

### What Is sqlc?

sqlc is **not** an ORM. It is a **code generator** that takes a fundamentally different approach to database access: instead of mapping objects to tables at runtime, it **parses your SQL queries at compile time** and generates type-safe Go functions, parameter structs, and result structs from them.

The core philosophy is captured by the [Encore engineering blog](https://encore.dev/blog/go-get-it-001-sqlc): *"sqlc lets you write queries in a plain `.sql` file. Alongside each query you give it a name and specify what it returns: no rows, one row, or many rows. From this sqlc generates a function for each query."*

This inverts the traditional ORM approach:

| Aspect | ORM (e.g., GORM) | sqlc |
| --- | --- | --- |
| **Input** | Go structs with tags | SQL queries + schema |
| **Output** | SQL queries (at runtime, via reflection) | Go code (at compile time, via static analysis) |
| **Source of truth** | Go struct definitions | SQL schema + migration files |
| **Query language** | Chainable Go API (`.Where()`, `.Find()`) | Standard SQL |
| **Type safety** | Runtime (reflection-based scanning) | Compile time (generated code) |
| **When errors are caught** | Runtime (often in production) | Build time (`sqlc generate`) |

As noted in the sqlc documentation: *"sqlc generates type-safe code from SQL. You write queries in SQL. You run sqlc to generate code with type-safe interfaces to those queries. You write application code that calls the generated code."*

### How sqlc Works

sqlc's workflow is a three-step process:

1. **Define schema** — sqlc reads your migration files (e.g., `database/migrations/*.up.sql`) to build an internal model of your database schema. It understands `CREATE TABLE`, `ALTER TABLE`, custom types, indexes, and constraints.

2. **Write annotated SQL queries** — You write standard SQL in `.sql` files with a special comment annotation:
   ```sql
   -- name: GetUserByUsername :one
   SELECT id, name, username, role, status, created_at
   FROM users
   WHERE username = $1 AND status != 'deleted';

   -- name: ListActiveUsers :many
   SELECT id, name, username, role, status, created_at
   FROM users
   WHERE status != 'deleted'
   ORDER BY name;

   -- name: CreateUser :one
   INSERT INTO users (name, username, password_hash, onetime_password_hash, role, status)
   VALUES ($1, $2, $3, $4, $5, $6)
   RETURNING id;
   ```

3. **Generate Go code** — Running `sqlc generate` produces:
   - A `models.go` file with Go structs matching your database tables.
   - A `query.sql.go` file with type-safe functions for each annotated query.
   - A `db.go` file with a `Queries` struct and a `DBTX` interface.

The generated code looks exactly like the hand-written boilerplate you would write yourself — `QueryRowContext`, `row.Scan()`, explicit column lists — but it is **correct by construction** because sqlc validates the SQL against the actual schema.

**The key insight**, as the Encore blog emphasizes: *"What makes sqlc special is that it understands your database structure, and uses that understanding to validate the SQL you write. So while it may look like regular, stringly-typed SQL, it's actually being validated against the actual database table. If you have a typo in a column name, sqlc will give you a compile-time error."*

sqlc achieves this **without connecting to your database** — it parses your migration files and uses PostgreSQL's own parsing library under the hood.

### sqlc Overview

| Feature | Description |
| --- | --- |
| Code generation | Generates Go structs, query functions, and `DBTX` interface from SQL |
| Schema understanding | Parses migration files to build schema model; supports `CREATE TABLE`, `ALTER TABLE`, custom types |
| Query validation | Validates SQL queries against schema at compile time; catches typos, type mismatches, missing columns |
| Return types | `:one` (single row), `:many` (slice), `:exec` (no rows), `:execresult` (`sql.Result`), `:execrows` (row count) |
| Parameter binding | Positional (`$1`, `$2`) or named (`sqlc.arg('name')`) parameters become typed function arguments |
| PostgreSQL support | Full PostgreSQL dialect support via `pg_query_go` (PostgreSQL's actual parser); CTEs, `json_agg`, custom enums, JSONB |
| Driver compatibility | Supports `database/sql`, `pgx/v5`, and `pgx/v4` as output targets via `sql_package` config |
| Configuration | `sqlc.yaml` config file defining engine, query/schema paths, output package, and options |
| Cloud features | Optional `sqlc cloud` for query verification, analysis, and CI integration |
| Multi-database | Supports PostgreSQL, MySQL, and SQLite |

**Typical sqlc configuration for jotti:**
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "../../database/migrations/"
    gen:
      go:
        package: "sqlcgen"
        out: "sqlcgen"
        sql_package: "pgx/v5"
```

**sqlc's strengths:**
- **SQL-first** — You write real SQL, not a DSL or chainable API. The database speaks SQL, and so do you.
- **Compile-time safety** — Column name typos, type mismatches, and schema drift are caught before the code runs.
- **Zero runtime overhead** — Generated code is identical to hand-written `database/sql` / `pgx` code. No reflection, no proxy objects.
- **PostgreSQL-native** — Uses PostgreSQL's own parser. Full support for CTEs, `json_agg()`, custom enums, window functions, and other advanced features.
- **No new query language** — Unlike ORMs, there is nothing new to learn beyond SQL and a few annotation comments.

**sqlc's known limitations:**
- **Static analysis is imperfect.** Complex expressions sometimes require manual type casts (`$1::text`) to help sqlc infer the correct type. As the Encore blog notes: *"its static analysis isn't perfect. Sometimes you have to manually specify the type of an expression or column."*
- **SQL lives in separate files.** Queries are decoupled from the Go code that calls them. For simple queries this improves clarity; for complex queries, the Encore blog warns: *"you have to be careful with naming your queries to ensure the semantics of your query are fully described by the name."*
- **Build step required.** Every schema or query change requires running `sqlc generate`. This must be integrated into the development workflow and CI pipeline.
- **Generated code is not idiomatic for all patterns.** The generated `Queries` struct and `DBTX` interface may not align perfectly with existing repository interfaces; an adapter layer may still be needed.
- **Custom types require configuration.** PostgreSQL custom enums and JSONB types need explicit `overrides` in `sqlc.yaml` to map to the correct Go types.

### Dimension-by-Dimension Analysis: sqlc vs. jotti

#### 1. CRUD Repositories (user_repo, table_repo)

**Current state:** `user_repo` (84 + 35 = 119 lines) and `table_repo` (86 + 28 = 114 lines) implement simple CRUD with hand-written SQL, manual `row.Scan()` calls, and `db*` adapter structs with `toDomain()` converters.

**What sqlc would change:**
- The SQL queries already written in `repo.go` files would move to `.sql` files with annotations.
- sqlc would generate the `row.Scan()` boilerplate, parameter binding, and result structs automatically.
- The `db*` adapter structs (`dbuser`, `dbtable`) and their `toDomain()` methods could potentially be replaced by sqlc-generated model structs — or retained if domain separation is preferred.

**Verdict:** sqlc provides a **clear benefit** here. It eliminates the manual scanning boilerplate while keeping the exact same SQL. The queries remain under developer control, and compile-time validation catches schema drift. Unlike GORM, sqlc does not impose its own lifecycle model — it just generates the boilerplate you would write yourself.

**Friction point:** jotti's `db*` → `toDomain()` pattern (Data Mapper) would need adaptation. Options: (a) use sqlc-generated models directly as domain types (simpler but couples generation to domain), or (b) keep a thin mapping layer from sqlc models to domain models (preserves separation but retains some adapter code). Option (b) is more consistent with jotti's architecture.

#### 2. Product Repository (product_repo) — JSON Aggregation

**Current state:** `product_repo` (286 lines total) uses `json_agg()` + `json_build_object()` + CTEs for single-query product-with-variants fetching.

**What sqlc would change:**
- sqlc **fully supports** CTEs, `json_agg()`, `json_build_object()`, and other PostgreSQL-specific functions because it uses PostgreSQL's own parser.
- The complex aggregation query could be placed in a `.sql` file and sqlc would generate the appropriate Go function and result struct.
- However, the return type of `json_agg()` is `json`/`jsonb`, which sqlc would map to `json.RawMessage` or a configured custom type. The developer would still need to unmarshal the JSON into domain types.

**Verdict:** sqlc handles this **significantly better than GORM**. The CTE + `json_agg()` query works as-is in sqlc because sqlc understands the full PostgreSQL dialect. The main friction is configuring the return type for the JSON aggregation column — but this is a one-time configuration in `sqlc.yaml`, not a fundamental architectural conflict.

#### 3. Event Repository (event_repo) — Event Sourcing

**Current state:** `event_repo` (134 lines) implements append-only writes (`WriteEvent`) and CTE-based snapshot reads (`ReadEventsWithSnapshot`).

**What sqlc would change:**
- `WriteEvent` is a simple `INSERT ... RETURNING id` — sqlc would generate this trivially.
- `ReadEventsWithSnapshot` uses a CTE to find the latest snapshot and read events forward — sqlc supports CTEs natively.
- sqlc generates `exec`/`one`/`many` functions without any assumption about mutability. There are no lifecycle hooks, no update/delete methods unless you write `UPDATE`/`DELETE` SQL. An append-only store is perfectly natural: you simply don't write UPDATE or DELETE queries.

**Verdict:** sqlc is an **excellent fit** for event sourcing. Unlike GORM (which assumes mutable records), sqlc generates exactly the functions you define — nothing more. If you only write `INSERT` and `SELECT` queries, that is all that gets generated. The append-only pattern is fully supported because sqlc imposes no lifecycle model.

#### 4. Domain Model Separation (Data Mapper vs. Active Record)

**Current state:** jotti uses the Data Mapper pattern with private `db*` structs and `toDomain()` converters.

**What sqlc would change:**
- sqlc generates its own model structs in a `models.go` file. These are plain Go structs with no tags, no methods, no framework coupling — just fields matching the database columns.
- These generated models are **not** domain models — they are database-level types. To maintain jotti's Data Mapper separation, a mapping layer from sqlc models to domain models would still be needed.
- However, this mapping layer would be significantly simpler: sqlc models are already well-typed Go structs with correct field types, so the mapping is straightforward field-to-field assignment.

**Verdict:** sqlc is **compatible with the Data Mapper pattern**. Unlike GORM (which pushes Active Record), sqlc generates pure data structs that live in a separate package. The `domain/` layer remains untouched. The `repository/` layer becomes a thin wrapper that calls sqlc-generated functions and maps results to domain types.

#### 5. PostgreSQL-Specific Features

**Current state:** jotti uses custom enums, JSONB, triggers, `IDENTITY` columns, `TIMESTAMPTZ`, and privilege restrictions.

**What sqlc would change:**
- **Custom enums:** sqlc parses `CREATE TYPE ... AS ENUM` from migration files and can generate Go string constants or custom types. This requires `overrides` in `sqlc.yaml` to map to jotti's existing domain types (e.g., `user.Role`).
- **JSONB columns:** Supported natively. sqlc maps `jsonb` to `json.RawMessage` by default, which matches jotti's current `event.Data` field.
- **CTEs:** Fully supported — sqlc uses PostgreSQL's own parser.
- **Triggers:** Irrelevant to sqlc — triggers are database-level enforcement that does not affect query generation.
- **`IDENTITY` and `TIMESTAMPTZ`:** Supported natively.

**Verdict:** sqlc has **excellent PostgreSQL support**. Because it uses PostgreSQL's parser (`pg_query_go`), it understands virtually all PostgreSQL-specific features. This is a fundamental advantage over GORM, which provides a generic ORM layer that often cannot express database-specific functionality.

#### 6. Error Handling

**Current state:** The `db` package maps PostgreSQL error codes to domain sentinel errors.

**What sqlc would change:**
- sqlc generates functions that return standard Go errors. The underlying errors are PostgreSQL driver errors (from `pgx`), which are the same errors jotti already handles.
- The `db.Error()` mapping layer would remain unchanged — it operates on the driver-level errors, not on the query execution layer.

**Verdict:** **Fully compatible.** sqlc does not introduce its own error abstraction. The existing error mapping works as-is.

#### 7. Testing & Mocking

**Current state:** Each repository has a `mock.go` implementing the repository interface. Unit tests use mocks; integration tests use a real database.

**What sqlc would change:**
- sqlc generates a `Queries` struct with a `DBTX` interface. This interface can be mocked using standard Go techniques.
- However, jotti's repository interfaces are domain-specific (e.g., `UserRepository`, `TableRepository`), not generated. To maintain these interfaces, the repository layer would wrap sqlc's `Queries` struct, and mocking would continue at the repository interface level.
- Alternatively, sqlc can be configured to emit interfaces for the generated queries, enabling direct mocking of the generated code.

**Verdict:** **Compatible.** The existing mocking strategy is preserved. sqlc does not complicate mocking — it uses standard Go interfaces, not reflection or proxy objects.

#### 8. Codebase Size & Complexity

| Metric | Current | With sqlc (estimated) |
| --- | --- | --- |
| Repository code (hand-written) | 653 lines (8 files) | ~350 lines (mapping layer only, scanning eliminated) |
| SQL query files | 0 (SQL inline in Go) | ~150 lines (`.sql` files with annotations) |
| Generated code | 0 | ~400 lines (auto-generated, not maintained) |
| Adapter structs | 3 `types.go` files | 0–1 (sqlc generates models; thin mapping may remain) |
| SQL statements | ~29 explicit queries | ~29 (same queries, now in `.sql` files) |
| Direct dependencies | 7 | 7 (sqlc is a build tool, not a runtime dependency) |
| Build tools | `go build` | `go build` + `sqlc generate` |
| Learning curve | SQL + pgx | SQL + pgx + sqlc annotation syntax + sqlc.yaml config |

**Key insight:** sqlc is a **build-time tool**, not a runtime dependency. It does not appear in `go.mod`. The generated code uses `database/sql` or `pgx/v5` directly — the same interfaces jotti already uses. This means zero runtime overhead and no new runtime dependencies.

#### 9. Dependency Philosophy

sqlc aligns well with jotti's dependency-light philosophy:

- **No runtime dependency.** sqlc is a CLI tool run during development/CI. The generated code depends only on `database/sql` or `pgx/v5` — libraries jotti already uses.
- **No reflection.** Generated code is plain Go with explicit `Scan()` calls — no runtime struct inspection.
- **No framework lock-in.** If sqlc is abandoned, the generated code continues to work. Queries can be manually maintained as regular Go code (they are already hand-written-style).

The only "cost" is adding `sqlc generate` to the development workflow and CI pipeline, and maintaining a `sqlc.yaml` configuration file.

### Summary (sqlc)

| Criterion | sqlc Benefit | jotti Fit |
| --- | --- | --- |
| Simple CRUD boilerplate reduction | ✅ Eliminates `Scan()` boilerplate | ✅ Saves ~300 lines of hand-written scanning code |
| JSON aggregation queries | ✅ Full PostgreSQL parser support | ✅ CTE + `json_agg()` works natively |
| Event sourcing (append-only) | ✅ No lifecycle assumptions | ✅ Generate only the queries you write |
| Domain model separation | ✅ Generated models are plain structs | ✅ Compatible with Data Mapper pattern |
| PostgreSQL-specific features | ✅ Uses PostgreSQL's own parser | ✅ Enums, JSONB, CTEs, triggers all supported |
| Soft-delete (three-state enum) | ✅ SQL-level filtering unchanged | ✅ Same `WHERE status != 'deleted'` queries |
| Testing & mocking | ✅ Standard Go interfaces | ✅ Existing mocking strategy preserved |
| Compile-time validation | ✅ Catches schema drift, typos | ✅ Prevents runtime query errors |
| Developer onboarding | ⚠️ Must learn sqlc annotations + config | ⚠️ Small learning curve, but SQL knowledge transfers |
| Database portability | ⚠️ Supports PG/MySQL/SQLite separately | ❌ Not a goal for jotti |
| Dependency footprint | ✅ Zero runtime dependency | ✅ Build tool only, no `go.mod` changes |
| Go idiom alignment | ✅ Generates idiomatic Go code | ✅ Matches Go's explicit, no-magic philosophy |
| Build workflow | ⚠️ Requires `sqlc generate` step | ⚠️ Must be integrated into CI/CD |

### Recommendation (sqlc)

**sqlc is a strong candidate for jotti.**

sqlc aligns remarkably well with jotti's philosophy: SQL-first, PostgreSQL-native, explicit, and dependency-light. It addresses the primary weakness of the current approach — manual `Scan()` boilerplate and the risk of runtime errors from schema drift — without introducing the architectural conflicts that disqualify GORM.

The main trade-off is adding a code generation step to the development workflow. This is a process change, not an architectural compromise. The generated code is idiomatic Go that uses the same interfaces jotti already relies on.

---

## sqlx Evaluation: Is sqlx a Good Fit for jotti?

This section evaluates whether adopting [sqlx](https://github.com/jmoiron/sqlx) — a set of extensions to Go's `database/sql` standard library — would be beneficial for jotti.

### What Is sqlx?

sqlx is a **lightweight extension library** for Go's `database/sql` package. It is not an ORM, not a query builder, and not a code generator. Instead, it wraps the standard `sql.DB`, `sql.Tx`, and `sql.Rows` types with additional convenience methods — most importantly, **automatic struct scanning**.

As the sqlx README describes: *"sqlx is a library which provides a set of extensions on go's standard `database/sql` library. The sqlx versions of `sql.DB`, `sql.TX`, `sql.Stmt`, et al. all leave the underlying interfaces untouched, so that their interfaces are a superset on the standard ones."*

This makes sqlx the most conservative option in this evaluation — it changes the least about how you interact with your database.

| Aspect | `database/sql` (current) | sqlx |
| --- | --- | --- |
| **Query execution** | `QueryRowContext`, `QueryContext`, `ExecContext` | Same, plus `Get`, `Select`, `NamedExec`, `NamedQuery` |
| **Result scanning** | Manual `row.Scan(&field1, &field2, ...)` | Automatic `StructScan`, `Get`, `Select` via `db:"..."` tags |
| **Named parameters** | Not supported (positional `$1`, `$2` only) | `NamedExec`, `NamedQuery` with `:field_name` syntax |
| **Struct mapping** | Manual field-by-field | Automatic via struct tags (`db:"column_name"`) |
| **Transactions** | `sql.Tx` with manual `Begin`/`Commit`/`Rollback` | `sqlx.Tx` with same methods + struct scanning |
| **Connection** | `sql.Open()` | `sqlx.Connect()` (also pings) or `sqlx.Open()` |

### sqlx Overview

| Feature | Description |
| --- | --- |
| `db.Get(&dest, query, args...)` | Execute query, scan single row into struct |
| `db.Select(&dest, query, args...)` | Execute query, scan all rows into slice of structs |
| `StructScan(rows, &dest)` | Scan current row into struct using `db:"..."` tags |
| `NamedExec(query, struct/map)` | Execute with named parameters (`:name` instead of `$1`) |
| `NamedQuery(query, struct/map)` | Query with named parameters |
| `Rebind(query)` | Translate `?` placeholders to driver-specific format (`$1`, `:arg`) |
| `In(query, args...)` | Expand `IN (?)` clauses with slice arguments |
| `MapScan(rows, &map)` | Scan row into `map[string]interface{}` |
| `sqlx.DB`, `sqlx.Tx`, `sqlx.Rows` | Drop-in replacements for `sql.DB`, `sql.Tx`, `sql.Rows` with extra methods |

**Typical sqlx usage (adapted for jotti):**

```go
type dbuser struct {
    ID                  int       `db:"id"`
    Name                string    `db:"name"`
    Username            string    `db:"username"`
    PasswordHash        *string   `db:"password_hash"`
    OnetimePasswordHash *string   `db:"onetime_password_hash"`
    Role                string    `db:"role"`
    Status              string    `db:"status"`
    CreatedAt           time.Time `db:"created_at"`
}

// Current jotti code (manual scanning):
func (r *Repository) GetByUsername(ctx context.Context, username string) (user.User, error) {
    row := r.DB.QueryRowContext(ctx, queryGetByUsername, username)
    var u dbuser
    err := row.Scan(&u.ID, &u.Name, &u.Username, &u.PasswordHash,
        &u.OnetimePasswordHash, &u.Role, &u.Status, &u.CreatedAt)
    if err != nil {
        return user.User{}, db.Error(err)
    }
    return u.toDomain(), nil
}

// With sqlx (automatic struct scanning):
func (r *Repository) GetByUsername(ctx context.Context, username string) (user.User, error) {
    var u dbuser
    err := r.DB.GetContext(ctx, &u, queryGetByUsername, username)
    if err != nil {
        return user.User{}, db.Error(err)
    }
    return u.toDomain(), nil
}
```

**sqlx's strengths:**
- **Minimal learning curve.** If you know `database/sql`, you know sqlx. The API is a superset — nothing is replaced, only extended.
- **Drop-in compatible.** Existing `*sql.DB` code works unchanged. Migration can be incremental — one function at a time.
- **Struct scanning.** Eliminates the most tedious part of hand-written database code: manual `row.Scan()` calls with positional field matching.
- **Named parameters.** Improves readability for INSERT/UPDATE statements with many columns.
- **No magic.** No reflection-based query building, no implicit behaviors, no lifecycle hooks. You still write SQL; sqlx just reduces the scanning boilerplate.
- **Battle-tested.** 16k+ GitHub stars, widely used in production, stable API.

**sqlx's known limitations:**
- **Runtime reflection for struct scanning.** Unlike sqlc (compile-time generation), sqlx uses runtime reflection to map column names to struct fields. Errors (e.g., missing `db:"..."` tag, column name mismatch) are caught at runtime, not compile time.
- **No query validation.** sqlx does not understand your schema. Typos in SQL queries are not caught until execution.
- **No code generation.** You still write the `db*` adapter structs, the `toDomain()` converters, and the SQL queries by hand. sqlx only automates the scanning step.
- **`database/sql`-only.** sqlx wraps `database/sql`, not `pgx/v5` native. jotti already uses `pgx/v5` via the `database/sql` compatibility layer (`pgx/v5/stdlib`), so this is compatible — but it means sqlx cannot take advantage of pgx-native features (e.g., `pgx.CollectRows()`).
- **Maintenance status.** sqlx has had less active development in recent years. The library is stable but not rapidly evolving.

### Dimension-by-Dimension Analysis: sqlx vs. jotti

#### 1. CRUD Repositories (user_repo, table_repo)

**Current state:** Each repository manually scans columns positionally with `row.Scan(&field1, &field2, ...)`.

**What sqlx would change:**
- Replace `row.Scan(&u.ID, &u.Name, &u.Username, ...)` with `db.GetContext(ctx, &u, query, args...)`.
- Replace `rows.Scan(...)` loops with `db.SelectContext(ctx, &dest, query, args...)`.
- The `db*` adapter structs would gain `db:"column_name"` tags (most already have the correct field names, so tags may be minimal).
- The `toDomain()` pattern and SQL queries remain unchanged.

**Verdict:** sqlx provides a **moderate benefit** here. It eliminates ~3–5 lines per query method (the `Scan()` call with all its positional arguments) and reduces the risk of column-order bugs. The benefit is real but smaller than sqlc's, because the adapter structs and SQL still need manual maintenance.

#### 2. Product Repository (product_repo) — JSON Aggregation

**Current state:** Uses `json_agg()` + CTE, scanning JSON into `json.RawMessage` and then unmarshaling.

**What sqlx would change:**
- The complex query remains unchanged — sqlx does not modify or interpret queries.
- For the JSON aggregation column, sqlx's `StructScan` would need a field with type `json.RawMessage` and a matching `db:"..."` tag. The subsequent JSON unmarshaling logic remains the same.
- For the simpler variant queries (`CreateVariant`, `UpdateVariant`), sqlx would reduce scanning boilerplate.

**Verdict:** **Marginal improvement.** The complex JSON aggregation query's main boilerplate is in the JSON unmarshaling, not the `Scan()` call. sqlx helps with the simpler queries but does not address the core complexity.

#### 3. Event Repository (event_repo) — Event Sourcing

**Current state:** Simple INSERT and CTE-based SELECT with `json.RawMessage` for event data.

**What sqlx would change:**
- `WriteEvent` (INSERT ... RETURNING id) could use `db.GetContext()` for slightly cleaner scanning.
- `ReadEventsWithSnapshot` scans multiple rows — sqlx's `SelectContext()` would replace the manual scan loop.
- sqlx has no lifecycle assumptions — it works exactly the same for append-only stores as for mutable tables.

**Verdict:** **Good fit.** sqlx simplifies the row-scanning loop in `ReadEventsWithSnapshot` and imposes no architectural constraints on event sourcing. The improvement is modest because `event_repo` is already relatively simple.

#### 4. Domain Model Separation (Data Mapper vs. Active Record)

**Current state:** Private `db*` structs with `toDomain()` converters.

**What sqlx would change:**
- The `db*` structs remain, but with added `db:"..."` tags for automatic scanning.
- The `toDomain()` pattern is completely unaffected.
- sqlx does not generate models, does not push Active Record, and does not require any structural changes to the domain layer.

**Verdict:** **Fully compatible.** sqlx is the most architecturally conservative option. It enhances the existing `db*` structs with tags and provides scanning shortcuts — nothing more. The Data Mapper pattern is preserved exactly as-is.

#### 5. PostgreSQL-Specific Features

**What sqlx would change:** Nothing. sqlx passes SQL queries through to the driver without interpretation. All PostgreSQL features — custom enums, JSONB, CTEs, triggers, `IDENTITY`, `TIMESTAMPTZ` — continue to work exactly as they do now.

**Verdict:** **Fully compatible.** sqlx is query-agnostic; it does not parse, validate, or transform your SQL.

#### 6. Error Handling

**What sqlx would change:** sqlx returns standard `database/sql` errors. The existing `db.Error()` mapping works unchanged.

**Verdict:** **Fully compatible.** No changes needed.

#### 7. Testing & Mocking

**What sqlx would change:**
- Repository structs would hold `*sqlx.DB` instead of `*sql.DB`. Since `sqlx.DB` embeds `sql.DB`, this is backward-compatible.
- The existing interface-based mocking strategy is unaffected — mocks implement the repository interface, not the database type.

**Verdict:** **Fully compatible.** The mocking approach remains unchanged.

#### 8. Codebase Size & Complexity

| Metric | Current | With sqlx (estimated) |
| --- | --- | --- |
| Repository code | 653 lines (8 files) | ~550 lines (scanning reduced, SQL + adapters unchanged) |
| Adapter structs | 3 `types.go` files | 3 `types.go` files (with added `db:"..."` tags) |
| SQL statements | ~29 explicit queries | ~29 (unchanged) |
| Direct dependencies | 7 | 8 (+`sqlx`) |
| Learning curve | SQL + pgx | SQL + pgx + sqlx API (minimal) |

The net reduction is ~100 lines — roughly the scanning boilerplate. The SQL queries, adapter structs, `toDomain()` converters, and error handling all remain.

#### 9. Dependency Philosophy

sqlx is a lightweight library (~4k lines of code) with no transitive dependencies beyond the standard library. It is a smaller dependency than GORM by an order of magnitude and well-aligned with jotti's dependency-light philosophy.

However, adding any dependency has a cost, and the benefit of sqlx (~100 lines saved) is the smallest of the alternatives evaluated. Additionally, pgx already offers `pgx.CollectRows()` and `pgx.RowToStructByName()` for similar struct scanning — though jotti currently uses pgx through the `database/sql` compatibility layer rather than natively.

### Summary (sqlx)

| Criterion | sqlx Benefit | jotti Fit |
| --- | --- | --- |
| Simple CRUD boilerplate reduction | ✅ Moderate (`Scan()` eliminated) | ⚠️ Saves ~100 lines; adapter structs remain |
| JSON aggregation queries | ⚠️ Marginal (scanning only) | ⚠️ Core complexity is in JSON unmarshaling |
| Event sourcing (append-only) | ✅ No lifecycle assumptions | ✅ Works naturally |
| Domain model separation | ✅ Fully compatible | ✅ Data Mapper pattern preserved exactly |
| PostgreSQL-specific features | ✅ Query-agnostic | ✅ All features work unchanged |
| Soft-delete (three-state enum) | ✅ SQL-level filtering unchanged | ✅ Same queries |
| Testing & mocking | ✅ Backward-compatible | ✅ Existing strategy preserved |
| Compile-time validation | ❌ No schema awareness | ❌ Errors caught at runtime only |
| Developer onboarding | ✅ Minimal learning curve | ✅ `database/sql` superset |
| Database portability | ✅ Multi-driver via `database/sql` | ❌ Not a goal for jotti |
| Dependency footprint | ⚠️ Small dependency | ⚠️ Adds 1 dependency for ~100 lines saved |
| Go idiom alignment | ✅ Idiomatic, no magic | ✅ Matches Go philosophy |
| Build workflow | ✅ No extra build step | ✅ No workflow changes needed |

### Recommendation (sqlx)

**sqlx is a valid but modest improvement for jotti.**

sqlx is the lowest-risk, lowest-reward option. It reduces scanning boilerplate without changing anything about jotti's architecture, SQL, error handling, or testing strategy. The migration would be incremental and non-disruptive — you could adopt it one repository at a time.

However, the benefit is small (~100 lines saved), it does not provide compile-time safety, and it adds a runtime dependency. Given that pgx itself already offers struct scanning via `pgx.CollectRows()` and `pgx.RowToStructByName()`, jotti could achieve similar convenience without adding any new dependency by switching from the `database/sql` compatibility layer to native `pgx/v5`.

---

## Comprehensive Comparison: All Four Approaches

This section compares all evaluated approaches side by side: the current status quo (bare `database/sql` + `pgx/v5`), GORM, sqlc, and sqlx.

### Approach Profiles

**1. Status Quo — Bare `database/sql` + `pgx/v5`**
- Hand-written SQL queries, manual `row.Scan()` calls, explicit adapter structs with `toDomain()` converters.
- Full control, zero abstraction overhead, zero additional dependencies.
- Trade-off: manual boilerplate and no compile-time schema validation.

**2. GORM — Full ORM**
- Object-Relational Mapper that generates SQL from Go struct definitions at runtime.
- Provides CRUD automation, association management, hooks, auto-migrations, and soft deletes.
- Trade-off: heavy abstraction, runtime reflection, Active Record pattern, poor fit for event sourcing and PostgreSQL-specific features.

**3. sqlc — SQL Compiler (Code Generator)**
- Compile-time code generator that produces type-safe Go code from hand-written SQL queries.
- Validates queries against the database schema at build time using PostgreSQL's own parser.
- Trade-off: requires a code generation step in the build workflow; SQL lives in separate files.

**4. sqlx — `database/sql` Extension Library**
- Lightweight wrapper around `database/sql` that adds struct scanning, named parameters, and convenience methods.
- Drop-in compatible; changes only the scanning layer.
- Trade-off: runtime reflection for scanning; no schema validation; modest benefit.

### Feature Matrix

| Feature | Status Quo | GORM | sqlc | sqlx |
| --- | --- | --- | --- | --- |
| **Query language** | Raw SQL | Go API + `db.Raw()` | Raw SQL (annotated) | Raw SQL |
| **Type safety** | Runtime | Runtime (reflection) | Compile time | Runtime (reflection) |
| **Schema validation** | None | Partial (auto-migrate) | Full (parse migrations) | None |
| **Struct scanning** | Manual `Scan()` | Automatic (reflection) | Generated code | Automatic (`db:"..."` tags) |
| **Runtime overhead** | None | Reflection + session mgmt | None (generated code) | Reflection (scanning only) |
| **CTE support** | ✅ Native SQL | ❌ `db.Raw()` only | ✅ Native SQL | ✅ Native SQL |
| **`json_agg()` support** | ✅ Native SQL | ❌ `db.Raw()` only | ✅ Native SQL | ✅ Native SQL |
| **Custom enums** | ✅ Manual mapping | ⚠️ Manual handling | ✅ Config overrides | ✅ Manual mapping |
| **Event sourcing fit** | ✅ No assumptions | ❌ Mutable lifecycle | ✅ No assumptions | ✅ No assumptions |
| **Data Mapper compatible** | ✅ Current pattern | ❌ Active Record | ✅ Generated models | ✅ Tag-based scanning |
| **PostgreSQL-native** | ✅ Full support | ⚠️ Limited | ✅ PG parser | ✅ Query-agnostic |
| **Soft-delete (3-state)** | ✅ SQL-level | ⚠️ Custom scopes | ✅ SQL-level | ✅ SQL-level |
| **Mocking strategy** | Interface-based | Complex (`*gorm.DB`) | Interface-based | Interface-based |
| **Runtime dependencies** | 0 new | +2 (gorm + driver) | 0 new (build tool) | +1 (sqlx) |
| **Build workflow change** | None | None | `sqlc generate` step | None |
| **Migration path** | N/A (current) | Full rewrite | Incremental | Incremental |
| **Boilerplate reduction** | Baseline | ~170 lines (CRUD only) | ~300 lines (all repos) | ~100 lines (scanning) |
| **Learning curve** | SQL only | SQL + GORM API | SQL + annotations | SQL + minimal API |
| **Framework lock-in** | None | High (GORM conventions) | Low (generated code works standalone) | Low (superset of `database/sql`) |

### Fit Analysis for jotti

The following table scores each approach against jotti's specific requirements and constraints. Scores: ✅ good fit, ⚠️ partial fit, ❌ poor fit.

| jotti Requirement | Status Quo | GORM | sqlc | sqlx |
| --- | --- | --- | --- | --- |
| Event sourcing (append-only core) | ✅ | ❌ | ✅ | ✅ |
| PostgreSQL-specific SQL (CTEs, `json_agg`) | ✅ | ❌ | ✅ | ✅ |
| Data Mapper architecture | ✅ | ❌ | ✅ | ✅ |
| Three-state soft deletes | ✅ | ⚠️ | ✅ | ✅ |
| Dependency-light philosophy (7 deps) | ✅ | ❌ | ✅ | ⚠️ |
| Compile-time error detection | ❌ | ❌ | ✅ | ❌ |
| Schema drift protection | ❌ | ⚠️ | ✅ | ❌ |
| Boilerplate reduction | ❌ | ⚠️ | ✅ | ⚠️ |
| No workflow changes | ✅ | ✅ | ⚠️ | ✅ |
| Go idiom alignment (explicit, no magic) | ✅ | ❌ | ✅ | ✅ |
| **Score** | **8 ✅, 2 ❌** | **1 ✅, 4 ❌, 5 ⚠️** | **9 ✅, 1 ⚠️** | **7 ✅, 2 ❌, 1 ⚠️** |

### Final Recommendation

**The recommended approach for jotti: Stay with the status quo, with sqlc as the preferred upgrade path if boilerplate reduction and compile-time safety become priorities.**

#### Reasoning

1. **GORM is not suitable for jotti.** The analysis is unambiguous: GORM's mutable-lifecycle model conflicts with event sourcing, its Active Record pattern conflicts with jotti's Data Mapper architecture, and its abstraction layer cannot express jotti's PostgreSQL-specific queries. The boilerplate savings (~170 lines in CRUD repos only) do not justify the architectural compromises, the dependency weight, or the split between ORM-managed and raw-SQL code paths.

2. **sqlx is a safe but insufficient upgrade.** sqlx reduces scanning boilerplate (~100 lines) with minimal risk and no architectural changes. However, it provides no compile-time safety, no schema validation, and no protection against the most common database-related bugs (column name typos, type mismatches, schema drift). Given that pgx itself offers similar struct scanning features (`pgx.CollectRows()`, `pgx.RowToStructByName()`), adopting sqlx adds a dependency for a benefit that could be achieved without one.

3. **sqlc is the strongest alternative.** It addresses the two main weaknesses of the status quo — manual scanning boilerplate and lack of compile-time validation — without compromising any of jotti's architectural principles. It is SQL-first, PostgreSQL-native, generates idiomatic Go code, adds no runtime dependency, and works seamlessly with event sourcing, CTEs, `json_agg()`, and custom enums. The only cost is a code generation step in the build workflow.

4. **The status quo is viable.** jotti's current approach is clean, explicit, and working. The codebase is small (653 lines of repository code), the team understands the SQL, and the existing tests cover the persistence layer. There is no urgent problem to solve. The status quo's main weakness — manual `Scan()` boilerplate with no compile-time validation — is a developer experience issue, not a correctness or architecture issue.

#### Decision Framework

| If... | Then... |
| --- | --- |
| The current approach works well and the team is productive | **Stay with the status quo.** The codebase is small, and manual SQL is not a bottleneck. |
| Schema drift or scanning bugs become a recurring problem | **Adopt sqlc.** Compile-time validation will prevent these errors systematically. |
| The team wants incremental improvement with minimal disruption | **Consider sqlx** as a stepping stone, or use pgx's native struct scanning features directly. |
| The project grows to 20+ tables with complex CRUD | **Re-evaluate sqlc.** The code generation benefit increases linearly with the number of queries. |
| The project needs to support multiple database engines | **Re-evaluate GORM or sqlc's multi-engine support.** But this is not currently a jotti requirement. |

#### If Adopting sqlc

Should jotti decide to adopt sqlc, the migration path would be:

1. Add `sqlc.yaml` configuration pointing to existing migration files.
2. Create `.sql` query files for each repository, using the existing SQL queries from `repo.go` files.
3. Run `sqlc generate` to produce the initial Go code.
4. Refactor each repository to wrap sqlc's generated `Queries` struct, mapping results to domain types.
5. Add `sqlc generate` to the CI pipeline and Makefile.
6. Remove the hand-written `db*` adapter structs and `Scan()` calls, replaced by sqlc's generated models and functions.

This migration can be done incrementally — one repository at a time — without disrupting the rest of the codebase.
