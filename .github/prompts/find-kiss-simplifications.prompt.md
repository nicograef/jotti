---
description: "Audit the jotti codebase for simplification opportunities, dead code, stale abstractions, and cleanup candidates across database, backend, and frontend."
argument-hint: "Optional focus area, e.g. service frontend, backend repositories, SQL queries, or whole repo"
agent: "agent"
---

# Find KISS Simplification Opportunities in jotti

Use this as **Stage 2** of the repo-quality audit workflow, after cross-layer consistency and before the final UX review.

This prompt is for **structural simplification**: removing unnecessary complexity, stale indirection, redundant round-trips, and low-value code paths.

It is **not** the prompt for:

- frontend/backend payload mismatches or schema drift -> use Stage 1
- running lint/build/test/integration commands -> use Stage 3
- mobile usability, UI consistency, or workflow friction -> use Stage 4

## Context & Motivation

jotti is a mobile POS system for volunteer-run events. It's in active pre-release development with no production users — breaking changes are welcome. The codebase has evolved through several architectural phases (e.g., event sourcing was added, then a synchronous CQRS projection `table_state` was introduced). Some code still reflects earlier patterns that are no longer necessary.

**Example we already found:** The `TablePage` frontend was making 4-5 separate API calls to fetch saldo, unbezahlte Positionen, and ungelieferte Positionen — all reading the **same** `table_state` database row. This was a leftover from before the CQRS projection existed, when each piece of data had to be computed separately from the event log. The fix: consolidate into a single `get-tisch-state` endpoint, lift state to the parent component, and pass data down as props.

**Pattern to look for:** Places where the code does N round-trips / N queries / N abstractions for data that could be served in 1, or where indirection exists that no longer serves a purpose.

The question here is not "is this wrong?" but "is this more complicated than it needs to be?"

## Task

Search the jotti codebase thoroughly for similar simplification opportunities. Focus on:

### 1. Redundant API Round-Trips

- Frontend pages/components that make multiple API calls where fewer would suffice
- Backend endpoints that read the same DB table/row but return different slices of it
- Cases where data is fetched on mount but could be passed as props from a parent that already has it
- Duplicate reads caused by component boundaries rather than real data ownership

### 2. Over-Abstracted Layers

- Interfaces with only one implementation that add indirection without value
- Wrapper functions that just forward to another function without transformation
- Separate files/types that could be colocated or merged
- Repository or application methods that are only called once and do not add meaningful semantics
- Handler, query, or command abstractions that exist only for dependency injection but have no alternate implementation or test value
- Factory or wiring code that adds files but no meaningful decisions

### 3. Stale Patterns

- Code that was written for an earlier architecture (before event sourcing, before table_state projection, before CQRS) and hasn't been updated
- Unused exports, dead code paths, or endpoints that nothing calls
- Hooks or backend methods that duplicate logic available elsewhere
- Old read paths that still compute data dynamically even though the synchronous projection already exists
- Test-only helpers, adapters, or mocks that kept production abstractions alive after the original use case disappeared
- Documentation or comments that still push contributors toward superseded patterns

### 4. Frontend Data Flow

- Components that fetch their own data when a parent already has it (or could easily fetch it)
- Multiple `useFetch` hooks in the same component tree that could be consolidated
- State that's derived from other state but stored/fetched separately
- Backend response slices that are fetched separately even though one parent screen already has, or could cheaply request, the full payload

Only include frontend findings here when the main issue is redundant data flow or unnecessary request structure. If the main issue is user confusion, awkward interaction, missing feedback, or mobile layout quality, that belongs in Stage 4 instead.

### 5. Backend Query Consolidation

- Multiple SQL queries that could be a single query (or a single `table_state` read)
- Application-layer methods that call the same repository method and just extract different fields
- Read models that are computed on-the-fly but could use the projection (or vice versa)
- SQL queries that are generated but never called from repository code
- Repository methods that only forward sqlc calls without adding invariants, mapping, or transaction boundaries

Only include query findings here when the main issue is redundant reads or needless indirection. If the issue is that the query, schema, payload, or mapping is inconsistent or incorrect, that belongs in Stage 1 instead.

### 6. Dead Code and Low-Value Duplication

- Unused functions, methods, components, hooks, schemas, DTOs, or exports
- Single-use types or files that could be inlined without losing clarity
- Code paths that exist only because of earlier architectural phases and no longer serve production behavior

Only report duplication here when it increases maintenance burden without protecting a real domain boundary. If two definitions are inconsistent across layers, that is Stage 1; if they are simply repetitive but locally clear, explain why they should or should not be collapsed.

### 7. What Not to Report Here

Do **not** use this prompt for:

- Request/response mismatches, JSON field drift, nullability drift, or type mismatches across layers
- Zod vs Go validation mismatches where the core problem is correctness or inconsistency
- Failing tests, lint errors, formatting failures, or build errors
- Mobile layout issues, wording problems, loading-state clarity, or workflow friction as experienced by end users

When a finding clearly belongs elsewhere, say so briefly and point it to the right stage instead of forcing it into this output.

### 8. Audit Method

For every claim, verify it in code before reporting it:

1. Trace usages end-to-end where relevant: frontend call site -> HTTP handler -> application service -> repository -> SQL query.
2. Confirm whether an export, method, endpoint, or query is actually unused before calling it dead code.
3. Distinguish production dead code from code that only exists for tests.
4. When you suspect duplicate logic, compare the concrete files and describe the overlap precisely.
5. Prefer repo evidence over generic clean-code advice.

### 9. Repo-Specific Guardrails

Do not flag the following as problems just because they add structure:

- Event sourcing for table operations itself. The event log remains immutable and intentional.
- The synchronous `table_state` projection. It is a deliberate simplification and should instead be used to spot stale read paths that bypass it.
- POST-only HTTP endpoints. This is a repo rule, not a smell.
- The separation between admin, service, and serviceleitung bounded contexts. Do not suggest merging them.
- The rule that frontend API calls go through backend classes instead of direct `fetch()`.

Look for simplification opportunities inside those constraints, not by removing them.

## Output Format

Group findings by category. For each finding, provide:

1. **Category**: `Redundant API Calls`, `Over-Abstraction`, `Dead Code`, `Stale Pattern`, `Frontend Data Flow`, `Query Consolidation`, `Schema Duplication`, or another precise label
2. **What**: Brief description of the redundancy or complexity
3. **Where**: Specific files and line numbers
4. **Impact**: What becomes simpler, faster, or easier to understand if fixed
5. **Why it exists**: Best guess at the historical reason
6. **Suggested simplification**: Concrete proposal that removes code, indirection, or round-trips
7. **Risk/trade-off**: What could go wrong or what we'd lose
8. **Effort**: `Small` / `Medium` / `Large`
9. **Priority**: `Quick Win`, `Worth Scheduling`, or `Needs Careful Refactor`

After listing all findings, add a final section:

## Recommended Cleanup Backlog

1. **Top 5 Quick Wins**: Small, low-risk changes with immediate payoff
2. **Next Refactor Batch**: Medium changes worth grouping into one cleanup pass
3. **High-Risk or Cross-Cutting Items**: Bigger simplifications that need dedicated attention

If you find no meaningful issue in a category, say so briefly instead of forcing a weak finding.

If you notice a likely issue that belongs in another stage, add a short note at the end:

## Out-of-Scope Notes

- Brief bullet list of items that should be checked with Stage 1, 3, or 4 instead

## Constraints

- Do NOT suggest adding features or new abstractions — this is about _removing_ unnecessary complexity
- Do NOT suggest changes to the event sourcing core (events are immutable, projection is synchronous — this is intentional)
- Do NOT suggest merging bounded contexts (admin vs. service separation is intentional)
- Do NOT recommend deleting code unless you have verified it is unused or clearly superseded
- Do NOT confuse necessary domain duplication with accidental technical duplication; explain the difference when relevant
- Do NOT turn this into a general correctness audit; keep the focus on simplification and cognitive load reduction
- Focus on things that make the codebase simpler for volunteer developers to understand and maintain
- Read the actual code before making claims — no speculative suggestions

## Goal

Produce an audit that is immediately useful as a cleanup backlog for the jotti codebase. Favor findings that reduce cognitive load, remove stale code, collapse redundant request paths, and make the system easier for future contributors to understand without changing intentional architecture.
