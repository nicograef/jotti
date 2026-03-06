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
