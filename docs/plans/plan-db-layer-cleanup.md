# Plan: DB-Layer Cleanup (schema, queries, repositories)

> Source PRD: n/a (derived from a `/cleanup` review of the DB schema, SQL queries, and repository Go code)

## Goal

Apply the five accepted cleanup findings from the review of the persistence
layer. All changes are behavior-preserving: naming/formatting fixes, one DRY
refactor of transaction boilerplate, and one SQL parameter-style
reconciliation. No functional behavior, query results, or generated `dbgen`
output should change.

## Architectural decisions

Durable decisions that apply across all phases:

- **`withTx` helper shape**: a private method per transaction-capable repo,
  `func (r Repository) withTx(ctx context.Context, fn func(*dbgen.Queries) error) error`.
  It begins the tx, `defer tx.Rollback()` (keeping the existing
  `//nolint:errcheck` rationale), runs `fn(r.q.WithTx(tx))`, and returns
  `db.Error(tx.Commit())`. Methods that return a value (e.g. the generated
  event ID) capture it via a closure variable. No generics, no shared package
  helper.
- **SQL named-parameter style**: named params use the `@name` shorthand (as in
  `reporting.sql`). Existing positional `$1`/`$2` params are left untouched —
  only the verbose `sqlc.arg(name)` form is reconciled to `@name`.
- **No `dbgen` edits**: `backend/sqlc/dbgen/` is generated (`DO NOT EDIT`). It
  is regenerated via `make sqlc`, never hand-edited.

## Inventory

Transaction-capable repos (store `*sql.DB`, call `BeginTx`) — the only targets
for the `withTx` refactor:

- `backend/repository/kassenjournal_repo/repo.go` — 6 tx blocks:
  `WriteEvent:35-51`, `WriteEventWithDruckauftraege:58-87`,
  `WriteEventWithNachsignierAuftrag:92-128`,
  `WriteEventWithDruckauftraegeUndNachsignierAuftrag:132-173`,
  `WriteUmbuchung:177-199`, `RebuildAllProjections:521-580`.
- `backend/repository/druckauftrag_repo/repo.go` — 2 tx blocks:
  `EnqueueDruckauftraege:56-76`, `MeldeDruckergebnis:121-154`.
- `backend/repository/tse_repo/repo.go` — 1 tx block:
  `QuittiereTSENachsignierAuftrag:81-113`.

Text/format fixes:

- `backend/repository/kassensitzungen_repo/types.go:10` — comment wrongly says
  "table persistence layer" (copy-pasted from `table_repo/types.go:10`).
- `backend/repository/product_repo/repo.go:21,46,73` — three
  `fmt.Errorf("failed to unmarshal variants: %w", err)` wraps; the rest of the
  layer is verb-first with no `failed to` prefix (only 3 such strings in
  `repository/`).
- `backend/sqlc/queries/kassensitzungen.sql:18-26` — `GetKassenbestand`
  COALESCE block; lines 22-23 use a 2-space indent vs 4 elsewhere.

SQL named-param reconciliation:

- `backend/sqlc/queries/tse_nachsignier.sql:28-31` — `sqlc.arg(letzter_fehler)`,
  `sqlc.arg(max_versuche)`, `sqlc.arg(id)`.
- `backend/sqlc/queries/relay.sql:20-22` — same three `sqlc.arg(...)` in
  `IncrementDruckauftragFehlversuch`.
- `backend/sqlc/queries/reporting.sql` — reference style already uses
  `@kassensitzung_nr` / `@user_id`.

Tooling (from `Makefile`): `make sqlc` (`cd backend && sqlc generate`),
`make lint-backend` (go vet + goimports), `make fmt-backend` (goimports),
`make test` (backend unit tests).

## Resolved decisions

- **`withTx` placement**: per-repo private method (not a shared generic in the
  `db` package); return values flow out via closure capture.
- **Param-style scope**: reconcile `sqlc.arg(name)` → `@name` only; do not
  convert existing positional `$1` params.

## Open questions / Risks

- **Item 5 must produce a zero diff in `dbgen`.** `@name` and `sqlc.arg(name)`
  generate byte-identical Go, so `make sqlc` should leave
  `backend/sqlc/dbgen/` unchanged. If it does change, stop and investigate
  before committing — a non-empty diff means the assumption is wrong.
- **Item 4 must preserve error wrapping semantics.** In-tx steps already return
  wrapped errors (e.g. `writeEventInTx`), so the helper returns `fn`'s error
  verbatim; only `BeginTx`/`Commit` failures get `db.Error(...)`, matching the
  current code.

---

## Phase 1: Text-only quick wins

### Context

- `backend/repository/kassensitzungen_repo/types.go:10` — misleading copy-paste
  comment.
- `backend/repository/table_repo/types.go:10` — the correct original comment to
  match the style against.
- `backend/repository/product_repo/repo.go:13-90` — the three `failed to`
  wraps live in `GetProduct`, `GetAllProducts`, `GetActiveProducts`.
- `backend/sqlc/queries/kassensitzungen.sql:18-26` — `GetKassenbestand`.

### What to build

Three independent, behavior-neutral edits:

1. Fix the `kassensitzungen_repo/types.go` doc comment so it describes the
   kassensitzungen repository, not "table".
2. Drop the `failed to ` prefix from the three variant-unmarshal error wraps in
   `product_repo/repo.go`, leaving them verb-first (`"unmarshal variants: %w"`)
   to match the layer's dominant style.
3. Re-indent the two misaligned lines in `GetKassenbestand` to 4 spaces so the
   COALESCE term list lines up.

### Acceptance criteria

- [x] `kassensitzungen_repo/types.go:10` no longer mentions "table".
- [x] No `failed to` strings remain in `backend/repository/` production code
      (the three variant-unmarshal wraps in `product_repo`; remaining matches are
      unrelated `t.Fatalf("failed to …")` test-helper messages, left untouched).
- [x] `GetKassenbestand`'s COALESCE terms are uniformly indented.
- [x] `make lint-backend` and `make test` pass. The SQL re-indent regenerated
      `dbgen/kassensitzungen.sql.go` (whitespace-only in the embedded query;
      query results unchanged) — sqlc embeds verbatim SQL text, so source and
      generated had to stay in sync.

---

## Phase 2: Extract a per-repo `withTx` helper

### Context

- `backend/repository/kassenjournal_repo/repo.go:24-26` — `Repository` holds
  `db *sql.DB` and `q *dbgen.Queries`; 6 methods (lines listed in Inventory)
  repeat the `BeginTx / defer Rollback / WithTx / Commit` block.
- `backend/repository/druckauftrag_repo/repo.go:47-54` — same struct shape;
  2 methods.
- `backend/repository/tse_repo/repo.go:53-60` — same struct shape; 1 method.
- `backend/db/db.go:31-47` — `db.Error` already normalizes nil and pg errors,
  so `db.Error(tx.Commit())` is safe.

### What to build

Add a private `withTx(ctx, fn func(*dbgen.Queries) error) error` method to each
of the three transaction-capable repos and rewrite the 9 methods to call it,
moving each method's in-tx body into the callback. Methods returning a value
(event ID in the `kassenjournal_repo` writers; rebuilt-count in
`RebuildAllProjections`) declare a local variable before the call and assign it
inside the callback. The helper centralizes the `//nolint:errcheck` rollback
deferral and the `db.Error` wrapping of begin/commit. Each method must do
exactly what it did before — same inserts, same projection updates, same
rollback-on-error, same returned values and errors.

### Acceptance criteria

- [x] Each of `kassenjournal_repo`, `druckauftrag_repo`, `tse_repo` has one
      `withTx` helper; no raw `r.db.BeginTx(` remains in their methods — the
      sole remaining `BeginTx` per repo lives inside that helper (verified with
      `grep -rn 'BeginTx' backend/repository/`).
- [x] Return values (event IDs, rebuilt count) and error wrapping are
      unchanged for every refactored method. The closure captures the value into
      a local; on any error (begin/fn/commit) the method returns the zero value,
      matching the pre-refactor `return 0, …` paths exactly.
- [x] `make fmt-backend`, `make lint-backend` (incl. that the rollback
      `errcheck` lint stays satisfied) pass.
- [x] `make test` passes. The three repos have no `-tags=unit` tests; their
      write/transaction paths are covered by `make test-integration`, which also
      stayed green with no test edits.

---

## Phase 3: Reconcile SQL named-parameter style

### Context

- `backend/sqlc/queries/tse_nachsignier.sql:28-31` and
  `backend/sqlc/queries/relay.sql:20-22` — the only `sqlc.arg(...)` uses.
- `backend/sqlc/queries/reporting.sql` — established `@name` reference style.
- `backend/sqlc/dbgen/` — generated output that must not change.

### What to build

Replace the `sqlc.arg(letzter_fehler)`, `sqlc.arg(max_versuche)`, and
`sqlc.arg(id)` occurrences in `tse_nachsignier.sql` and `relay.sql` with the
`@letzter_fehler`, `@max_versuche`, `@id` shorthand. Leave all positional
`$1`/`$2` params untouched. Run `make sqlc` to confirm the change regenerates
identical Go.

### Acceptance criteria

- [ ] No `sqlc.arg(` remains in `backend/sqlc/queries/` (verify with
      `grep -rn 'sqlc.arg' backend/sqlc/queries/`).
- [ ] `make sqlc` leaves `backend/sqlc/dbgen/` with **no** diff
      (`git status backend/sqlc/dbgen/` clean).
- [ ] `make lint-backend` and `make test` pass.
