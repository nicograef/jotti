---
name: create-plan
description: >-
  Create an implementation plan from a PRD or a task description. Researches
  codebase context, clarifies ambiguities, and outputs a phased plan with
  vertical slices and acceptance criteria. Use when the user wants to plan a
  feature, break down a PRD, create an implementation plan, or mentions
  "tracer bullets".
---

# Create Plan

Create a phased implementation plan from **either** a PRD **or** a task
description. Output is a Markdown file in `docs/plans/`.

## Workflow

### 1. Determine the entry point

- **PRD provided** (file or in conversation context) → skip to step 3.
- **Task description only** → continue to step 2.

If a PRD exists but is not yet in context, ask the user to paste it or point
you to the file.

### 2. Clarify ambiguities (task-description path only)

Resolve unknowns through **1–3 rounds** of structured questions before
planning.

**Rules:**

- **Explore before asking.** If a question can be answered by reading the
  codebase, read the codebase instead of asking the user.
- **Always recommend.** Every question must include a recommended answer with
  brief reasoning.
- **Structured over free-text.** Use concrete options. Convert open-ended
  questions to multiple-choice with an "Other (specify)" escape hatch.
- **Max 5 questions per round.** Prioritise the most impactful unknowns.
- **Stop when resolved.** If all ambiguities are clear after 1 round, stop.
  Continue only if unresolved branches remain.

If the user declines to answer: proceed with recommended defaults and document
each assumption as a clearly marked callout (e.g. blockquote prefixed with
**Assumption:**).

### 3. Research the codebase

Read affected files, understand existing patterns, integration layers, and
the current architecture.

### 4. Identify architectural decisions

Before slicing, identify high-level decisions that are unlikely to change
throughout implementation:

- Route structures / URL patterns
- Database schema shape
- Key data models
- Authentication / authorization approach
- Third-party service boundaries

These go in the plan header so every phase can reference them.

### 5. Draft vertical slices

Break the work into **tracer bullet** phases. Each phase is a thin vertical
slice that cuts through ALL integration layers end-to-end.

**Slice rules:**

- Each slice delivers a narrow but COMPLETE path through every layer (schema,
  API, UI, tests).
- A completed slice is demoable or verifiable on its own.
- Prefer many thin slices over few thick ones.
- Do NOT include specific file names, function names, or implementation
  details that are likely to change as later phases are built.
- DO include durable decisions: route paths, schema shapes, data model names.

For small tasks (refactors, config changes, single-module work), a **single
phase** is perfectly valid.

### 6. Validate with the user

Present the proposed breakdown as a numbered list. For each phase show:

- **Title**: short descriptive name
- **User stories covered** (if working from a PRD): which user stories this
  addresses.

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Should any phases be merged or split further?

Iterate until the user approves the breakdown.

### 7. Write the plan file

Derive a slug from the task (e.g. `admin-dashboard`, `order-cancel`).
Create the file `docs/plans/plan-<slug>.md` (create the directory if it
doesn't exist).

## Rules

- **No code changes.** Only create the plan file.
- **Precise references.** Back every finding with file path and line numbers
  (e.g. `backend/api/product/http/handler.go:42-58`).
- **Readability-first.** Prefer simple, clear, idiomatic solutions.

## Quality

- Before presenting results, run the self-review checklist from AGENTS.md (Quality Principles) — applied to the quality of the plan artifact. Surface issues in the chat only if found.
- After task completion, include a human-readable summary paragraph alongside the commit message (see AGENTS.md, Git Workflow).

## Plan Template

```markdown
# Plan: <Title>

> Source PRD: <relative path to PRD file, or "n/a" if from task description>

## Goal

<What should be achieved?>

## Architectural decisions

Durable decisions that apply across all phases:

- **Routes**: ...
- **Schema**: ...
- **Key models**: ...
- (add/remove sections as appropriate; omit entirely for small tasks)

## Inventory

<Relevant existing files, patterns, dependencies — each with file path:lines>

## Resolved decisions

<Decisions made during the clarification phase — one bullet per decision>

## Open questions / Risks

<If any — otherwise omit>

---

## Phase 1: <Title>

**User stories**: <list from PRD, or omit if from task description>

### Context

- `path/file.go:10-45` — <why relevant>

### What to build

A concise description of this vertical slice. Describe the end-to-end
behavior, not layer-by-layer implementation.

### Acceptance criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

---

## Phase 2: <Title>

**User stories**: <list from PRD>

### Context

- `path/file.go:50-80` — <why relevant>

### What to build

...

### Acceptance criteria

- [ ] ...

<!-- Repeat for each phase -->
```
