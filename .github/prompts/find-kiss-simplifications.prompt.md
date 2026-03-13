# Find KISS Simplification Opportunities in jotti

## Context & Motivation

jotti is a mobile POS system for volunteer-run events. It's in active pre-release development with no production users — breaking changes are welcome. The codebase has evolved through several architectural phases (e.g., event sourcing was added, then a synchronous CQRS projection `table_state` was introduced). Some code still reflects earlier patterns that are no longer necessary.

**Example we already found:** The `TablePage` frontend was making 4-5 separate API calls to fetch saldo, unbezahlte Positionen, and ungelieferte Positionen — all reading the **same** `table_state` database row. This was a leftover from before the CQRS projection existed, when each piece of data had to be computed separately from the event log. The fix: consolidate into a single `get-tisch-state` endpoint, lift state to the parent component, and pass data down as props.

**Pattern to look for:** Places where the code does N round-trips / N queries / N abstractions for data that could be served in 1, or where indirection exists that no longer serves a purpose.

## Task

Search the jotti codebase thoroughly for similar simplification opportunities. Focus on:

### 1. Redundant API Round-Trips

- Frontend pages/components that make multiple API calls where fewer would suffice
- Backend endpoints that read the same DB table/row but return different slices of it
- Cases where data is fetched on mount but could be passed as props from a parent that already has it

### 2. Over-Abstracted Layers

- Interfaces with only one implementation that add indirection without value
- Wrapper functions that just forward to another function without transformation
- Separate files/types that could be colocated or merged

### 3. Stale Patterns

- Code that was written for an earlier architecture (before event sourcing, before table_state projection, before CQRS) and hasn't been updated
- Unused exports, dead code paths, or endpoints that nothing calls
- Hooks or backend methods that duplicate logic available elsewhere

### 4. Frontend Data Flow

- Components that fetch their own data when a parent already has it (or could easily fetch it)
- Multiple `useFetch` hooks in the same component tree that could be consolidated
- State that's derived from other state but stored/fetched separately

### 5. Backend Query Consolidation

- Multiple SQL queries that could be a single query (or a single `table_state` read)
- Application-layer methods that call the same repository method and just extract different fields
- Read models that are computed on-the-fly but could use the projection (or vice versa)

## Output Format

For each finding, provide:

1. **What**: Brief description of the redundancy/complexity
2. **Where**: Specific files and line numbers
3. **Why it exists**: Best guess at the historical reason
4. **Suggested simplification**: Concrete proposal
5. **Risk/trade-off**: What could go wrong or what we'd lose
6. **Effort**: Small / Medium / Large

## Constraints

- Do NOT suggest adding features or new abstractions — this is about _removing_ unnecessary complexity
- Do NOT suggest changes to the event sourcing core (events are immutable, projection is synchronous — this is intentional)
- Do NOT suggest merging bounded contexts (admin vs. service separation is intentional)
- Focus on things that make the codebase simpler for volunteer developers to understand and maintain
- Read the actual code before making claims — no speculative suggestions
