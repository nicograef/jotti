---
description: "Audit the jotti codebase for readability, consistency, and simplification opportunities with a focus on clean, idiomatic code across database, backend, and frontend."
argument-hint: "Optional focus area, e.g. service frontend, backend repositories, SQL queries, or whole repo"
agent: "agent"
---

# Find Readability-First KISS Improvements in jotti

Use this as **Stage 2** of the repo-quality audit workflow, after cross-layer consistency and before the final UX review.

This prompt is for **readability-first simplification**: removing unnecessary complexity, stale indirection, and low-value code paths while preserving simple, explicit, idiomatic code.

It is **not** the prompt for:

- frontend/backend payload mismatches or schema drift -> use Stage 1
- running lint/build/test/integration commands -> use Stage 3
- mobile usability, UI consistency, or workflow friction -> use Stage 4

## Context & Motivation

jotti is a mobile POS system for volunteer-run events. It's in active pre-release development with no production users — breaking changes are welcome. The codebase has evolved through several architectural phases (e.g., event sourcing was added, then a synchronous CQRS projection `table_state` was introduced). Some code still reflects earlier patterns that are no longer necessary.

**Example we already found:** The `TablePage` frontend previously spread one conceptual view state across multiple requests and local fetch paths. This increased cognitive load. The simplification was to make data flow explicit in one place (`get-tisch-state` in the parent) and pass values down clearly.

**Pattern to look for:** Places where the code is harder to read than necessary due to unclear flow, stale indirection, naming drift, or unnecessary abstraction.

The question here is not "is this wrong?" but "is this easy to read, consistent, and idiomatic?"

## Task

Search the jotti codebase thoroughly for similar simplification opportunities. Focus on:

### 1. Readability and Local Clarity

- Long, nested, or branched logic that can be simplified without changing behavior
- Code paths that hide intent instead of making it obvious
- Repeated patterns where small helper extraction improves clarity (without creating abstraction layers)
- Inconsistent naming for the same concept within one layer

### 2. Over-Abstracted Layers

- Interfaces with only one implementation that add indirection without value
- Wrapper functions that just forward to another function without transformation
- Separate files/types that could be colocated or merged
- Repository or application methods that are only called once and do not add meaningful semantics
- Handler, query, or command abstractions that exist only for dependency injection but have no alternate implementation or test value
- Factory or wiring code that adds files but no meaningful decisions

When in doubt, prefer explicit and straightforward code over indirection.

### 3. Stale Patterns

- Code that was written for an earlier architecture (before event sourcing, before table_state projection, before CQRS) and hasn't been updated
- Unused exports, dead code paths, or endpoints that nothing calls
- Hooks or backend methods that duplicate logic available elsewhere
- Old read paths that still compute data dynamically even though the synchronous projection already exists
- Test-only helpers, adapters, or mocks that kept production abstractions alive after the original use case disappeared
- Documentation or comments that still push contributors toward superseded patterns

### 4. Consistency and Idiomatic Style

- Inconsistent coding style across similar modules (same layer, different conventions)
- Inconsistent error handling and validation flow that makes behavior harder to predict
- Non-idiomatic patterns when a clearer local idiom already exists in the codebase
- Places where small duplication is actually clearer than forced DRY, and vice versa

Only include frontend findings here when the main issue is code readability/consistency. If the main issue is user confusion, awkward interaction, missing feedback, or mobile layout quality, that belongs in Stage 4 instead.

### 5. Backend and SQL Simplicity

- SQL queries that are hard to understand due to unnecessary complexity
- Repository methods that only forward sqlc calls without adding domain value
- Query/service boundaries where intent is unclear and can be made explicit
- SQL queries that are generated but never called from repository code

Only include database/query findings here when the main issue is readability/maintainability. If the issue is query/schema/payload correctness drift, that belongs in Stage 1.

### 6. Dead Code and Duplication with Judgment

- Unused functions, methods, components, hooks, schemas, DTOs, or exports
- Single-use types or files that could be inlined without losing clarity
- Code paths that exist only because of earlier architectural phases and no longer serve production behavior

Do not collapse duplication just to be DRY. Keep duplication when it improves local clarity and domain readability.

Only report duplication here when it increases maintenance burden without protecting a real domain boundary. If two definitions are inconsistent across layers, that is Stage 1; if they are simply repetitive but locally clear, explain why they should or should not be collapsed.

### 7. What Not to Report Here

Do **not** use this prompt for:

- Request/response mismatches, JSON field drift, nullability drift, or type mismatches across layers
- Zod vs Go validation mismatches where the core problem is correctness or inconsistency
- Failing tests, lint errors, formatting failures, or build errors
- Mobile layout issues, wording problems, loading-state clarity, or workflow friction as experienced by end users
- Performance-only optimizations (parallelization, batching, caching, micro-optimizations) unless they also clearly simplify code and improve readability

When a finding clearly belongs elsewhere, say so briefly and point it to the right stage instead of forcing it into this output.

### 8. Audit Method

For every claim, verify it in code before reporting it:

1. Trace usages end-to-end where relevant: frontend call site -> HTTP handler -> application service -> repository -> SQL query.
2. Confirm whether an export, method, endpoint, or query is actually unused before calling it dead code.
3. Distinguish production dead code from code that only exists for tests.
4. When you suspect duplicate logic, compare the concrete files and describe the overlap precisely.
5. Prefer repo evidence over generic clean-code advice.
6. Prefer the smallest simplification that makes intent clearer.

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

1. **Category**: `Readability`, `Consistency`, `Over-Abstraction`, `Dead Code`, `Stale Pattern`, `Local Duplication`, `Query Simplicity`, or another precise label
2. **What**: Brief description of the redundancy or complexity
3. **Where**: Specific files and line numbers
4. **Impact**: What becomes simpler and easier to understand if fixed
5. **Why it exists**: Best guess at the historical reason
6. **Suggested simplification**: Concrete proposal that improves readability and consistency with minimal change
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

- Do NOT suggest adding features or clever abstractions — this is about simpler, clearer code
- Do NOT suggest changes to the event sourcing core (events are immutable, projection is synchronous — this is intentional)
- Do NOT suggest merging bounded contexts (admin vs. service separation is intentional)
- Do NOT recommend deleting code unless you have verified it is unused or clearly superseded
- Do NOT confuse necessary domain duplication with accidental technical duplication; explain the difference when relevant
- Do NOT turn this into a general correctness audit; keep the focus on simplification and cognitive load reduction
- Focus on things that make the codebase simpler for volunteer developers to read, understand, and maintain
- Do NOT optimize for performance as the primary goal
- Prefer explicit and idiomatic code over over-DRY or abstraction-heavy proposals
- Read the actual code before making claims — no speculative suggestions

## Goal

Produce an audit that is immediately useful as a cleanup backlog for the jotti codebase. Favor findings that reduce cognitive load, remove stale code, and make the system easier to read and reason about without changing intentional architecture.
