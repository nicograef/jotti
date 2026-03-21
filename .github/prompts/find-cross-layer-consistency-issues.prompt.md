---
description: "Audit jotti end-to-end for cross-layer consistency issues between frontend, backend, database schema, SQL queries, Go structs, TypeScript types, and request/response payloads."
argument-hint: "Optional focus area, e.g. auth, service flow, admin CRUD, reporting, or whole repo"
agent: "agent"
---

# Find Cross-Layer Consistency Issues in jotti

Use this as **Stage 1** of the repo-quality audit workflow.

This prompt is for finding mismatches across layers, not for broad simplification advice and not for browser automation.

Use this prompt when the core question is: **do the layers still agree on what the system is and how data moves through it?**

It is **not** the prompt for:

- removing stale abstractions, low-value indirection, or redundant round-trips when correctness is not in doubt -> use Stage 2
- running repo verification commands -> use Stage 3
- reviewing mobile usability, interaction friction, or UI clarity -> use Stage 4

## Goal

Audit whether jotti's layers still agree with each other:

- database schema vs SQL queries
- SQL result shapes vs repository mapping
- Go domain structs / DTOs / handlers vs JSON payloads
- frontend backend classes vs backend endpoints
- Zod schemas vs Go validation and handler expectations
- TypeScript types vs actual backend responses
- naming and ubiquitous language across DB, Go, JSON, TypeScript, and UI

The outcome should be a concrete backlog of verified inconsistencies, drift, and likely bugs.

The question here is not "is this overly complex?" but "are these layers still saying the same thing?"

## What to Check

### 1. Request / Response Alignment

- Frontend request bodies vs backend handler request structs
- Frontend response parsing vs backend response JSON shape
- Missing, extra, renamed, nullable, or optional fields
- Status values, enum-like strings, IDs, timestamps, and cent-based money fields

### 2. Type and Schema Drift

- TypeScript types that no longer match Go structs or JSON payloads
- Zod schemas that validate a different shape than backend code accepts or returns
- Duplicate payload schemas that evolved independently
- Inconsistent naming between domain concepts and transport fields

### 3. Database / SQL / Repository Alignment

- SQL queries that assume columns or joins differently than the schema defines them
- sqlc-generated expectations vs repository mapping code
- Migrations, queries, and application code that disagree on nullability, defaults, or status values
- Read models and projections whose shape no longer matches callers' assumptions

### 4. Cross-Layer Invariants

- Money represented anywhere other than cents / integers
- Tisch, Bestellung, Ausgabe, Zahlung, Stornierung, Auszahlung terminology drift
- Backend-only filtering or aggregation that the frontend silently reimplements differently
- Validation rules enforced on one side but not the other

Only include invariant findings here when the main issue is disagreement or correctness drift across layers. If the same behavior is implemented consistently but with unnecessary indirection or redundant requests, that belongs in Stage 2 instead.

### 5. Real Flow Tracing

Trace representative flows end to end where relevant:

1. frontend call site
2. frontend backend/service class
3. HTTP handler
4. application service
5. repository
6. SQL query / schema

Prefer fewer, deeper verified findings over many speculative ones.

### 6. What Not to Report Here

Do **not** use this prompt for:

- wrapper layers, factory files, or abstractions that are merely unnecessary but still consistent
- duplicate reads, extra API calls, or split data flow where the payloads are still correct
- failing lint, format, test, build, or integration commands
- wording, layout, loading-state clarity, or workflow friction where the main issue is user experience rather than layer mismatch

When a finding clearly belongs elsewhere, say so briefly and point it to the right stage instead of forcing it into this output.

## Repo-Specific Guardrails

- Respect POST-only endpoints.
- Treat the backend as the single source of truth for filtering and aggregation.
- Do not suggest direct `fetch()` in the frontend.
- Do not flag event sourcing or the synchronous `table_state` projection as architectural mistakes.
- Do not recommend backwards-compatibility shims. Breaking changes are acceptable in this repo.
- When proposing fixes, prefer simple, explicit, idiomatic alignment changes over clever or optimization-driven rewrites.

## Output Format

For each finding, provide:

1. **Category**: `Payload Mismatch`, `Type Drift`, `Schema Drift`, `SQL / Repository Mismatch`, `Naming Drift`, `Validation Gap`, or another precise label
2. **What**: the concrete inconsistency
3. **Where**: exact files and line numbers across all affected layers
4. **Evidence chain**: show the traced path from caller to DB or vice versa
5. **Why it matters**: bug risk, maintenance cost, or invalid assumption
6. **Suggested fix**: the smallest coherent alignment change
7. **Risk / trade-off**: what to verify if the fix is applied
8. **Effort**: `Small` / `Medium` / `Large`
9. **Priority**: `Quick Win`, `Worth Scheduling`, or `Needs Careful Refactor`

Then add:

## Cross-Layer Alignment Backlog

1. **Immediate bug risks**
2. **Drift to fix before bigger refactors**
3. **Areas that look consistent**

If you notice likely issues that belong in another stage, add a short section:

## Out-of-Scope Notes

- Brief bullet list of items that should be checked with Stage 2, 3, or 4 instead

## Constraints

- Read actual code before making claims.
- Do not propose feature work.
- Do not report hypothetical issues without file evidence.
- Do not turn this into a general cleanup audit; keep the focus on correctness, agreement, and cross-layer drift.
- If a suspected issue is actually intentional, say so briefly and move on.
