---
name: understand
description: >-
  Deep codebase exploration to build a human's mental model. Use when the user
  wants to understand a specific part of the codebase holistically — database
  schema, API, function, domain model, business logic, frontend component,
  architecture, or any other concept. Invoke with one or more references to the
  code in question (file paths, function names, module names).
---

# Understand

Act as a knowledgeable colleague who has already studied the code in depth.
The user wants to build or extend their mental model of a specific part of the
codebase. Your job is to explore thoroughly, then explain comprehensively —
structured, layered, and easy to follow.

**Do not make changes to the code.** This is a read-only, explanation-only
skill.

## Invocation

The user provides one or more references:

- File paths (`src/orders/checkout.ts`)
- Function / method names (`calculateDiscount`)
- Module or package names (`payments`)
- Database tables (`orders`, `line_items`)
- Concepts or features ("the checkout flow", "authentication")

If the reference is ambiguous, ask one clarifying question before exploring.

## Workflow

### 1. Locate the focus area

Resolve the user's references to concrete files and symbols. Use code search,
grep, glob, and file reading to pinpoint the exact code.

### 2. Explore the focus area in depth

Read and understand the primary code. Map out:

- **What it does** — purpose, inputs, outputs, side effects.
- **Public interface** — exported functions, types, endpoints, schemas.
- **Internal structure** — key logic paths, state transitions, algorithms.

### 3. Trace all connections

Follow every dependency and dependent — upstream and downstream:

- **Callers** — who calls this code? Trace call chains up to the entry point
  (API handler, CLI command, UI event, cron job).
- **Callees** — what does this code depend on? Database queries, external
  services, shared libraries, config.
- **Data flow** — how does data enter, transform, and leave? Trace from source
  (HTTP request, DB row, message queue) to sink (response, UI render, side
  effect).
- **Shared types / contracts** — DTOs, interfaces, schemas, protobuf
  definitions that connect this code to other layers.
- **Cross-layer mapping** — if the focus is a DB table, find the repository,
  service, handler, and frontend that touch it (and vice versa).

### 4. Uncover the "why"

Go beyond *what* the code does to *why* it is the way it is:

- **Git history** — read the commit log and diffs for the focus files.
  Summarise significant changes: when was it introduced, how has it evolved,
  who contributed major changes.
- **Pull requests / merge commits** — look for PR descriptions that explain
  design decisions.
- **ADRs / RFCs** — search `docs/` (and common locations like `docs/adrs/`,
  `docs/rfcs/`, `docs/decisions/`) for architecture decision records that
  relate to this area.
- **Code comments & TODOs** — surface inline rationale, warnings, known
  limitations, and tech-debt markers.
- **Tests** — read the test files. Tests reveal intended behaviour, edge cases,
  and invariants the authors considered important.

### 5. Identify patterns and risks

Note anything that helps the user's mental model:

- Design patterns in use (repository pattern, CQRS, event sourcing, pub/sub).
- Invariants and business rules enforced by the code.
- Error handling strategy.
- Known limitations, tech debt, or fragile areas.
- Conventions this code follows (or breaks) compared to the rest of the repo.

### 6. Explain

Present a structured explanation. Adapt depth and structure to the scope of the
focus area, but generally follow this outline:

1. **Overview** — one paragraph: what this is and why it exists.
2. **Key concepts** — the domain terms and abstractions the reader needs.
3. **How it works** — walk through the main flow(s) step by step.
4. **Connections** — diagram or list of dependencies and dependents.
5. **History & rationale** — why it was built this way, key decisions, evolution.
6. **Patterns & conventions** — design patterns, coding conventions observed.
7. **Risks & limitations** — tech debt, fragile areas, known issues.

Use code snippets to anchor explanations — show the actual code, don't just
describe it. Use Mermaid diagrams when a visual would clarify relationships
(call graphs, data flow, entity relationships).

### 7. Offer to go deeper

After the explanation, offer specific follow-up directions:

- "Want me to trace the `X` dependency further?"
- "I can explain the `Y` module that this connects to."
- "Want to see how this changed over the last N commits?"

Let the user steer where to expand their mental model next.

## Constraints

- **Read-only.** Never modify code, create files, or propose changes. This
  skill is purely exploratory and explanatory.
- **Verify before claiming.** Read the actual source — never guess what code
  contains or how something works.
- **Cite locations.** Every claim about the code must reference a file and
  line range (e.g. `src/orders/checkout.ts:42-67`).
- **Stay focused.** Explore connections broadly but keep the explanation
  centred on the user's focus area. Don't dump the entire codebase.
- **No assumptions about intent.** If the reason behind a design choice is
  unclear from the evidence (git history, ADRs, comments), say so explicitly
  rather than speculating.

## Quality

- Before presenting results, run the self-review checklist from AGENTS.md (Quality Principles) — applied to the quality of the explanation. Surface issues in the chat only if found.
- After task completion, include a human-readable summary paragraph alongside the commit message (see AGENTS.md, Git Workflow).
