---
description: "Review jotti's admin, auth, and service frontend for mobile UX, UI consistency, terminology drift, validation behavior, loading states, error handling, and workflow friction."
argument-hint: "Optional focus area, e.g. login, table page, admin product flow, service checkout, or whole frontend"
agent: "agent"
---

# Review Mobile UX Consistency in jotti

Use this as **Stage 4** of the repo-quality audit workflow.

This prompt is for product-facing review of the frontend code: mobile-first behavior, consistency of wording and flows, and obvious UX defects. It is not a visual redesign prompt.

Use this prompt when the core question is: **what will confuse, slow down, or trip up real volunteers using the app on phones?**

It is **not** the prompt for:

- frontend/backend payload or schema mismatches -> use Stage 1
- removing stale abstractions, unnecessary indirection, or redundant request structure -> use Stage 2
- running repo verification commands -> use Stage 3

## Goal

Find verified UX/UI issues that make jotti harder to use for volunteer teams on phones during busy events.

The question here is not "is this architecture elegant?" but "is this flow clear, fast, and resilient for the user?"

## Focus Areas

### 1. Workflow Friction

- Too many taps for common service tasks
- Hidden or unclear next actions
- Hard-to-recover error states
- Missing feedback after save, order, delivery, payment, or cancellation

### 2. Mobile-First Quality

- Components likely to break on narrow screens
- Dense tables or forms without a mobile fallback
- Controls that are too small, too close, or require precision taps
- Important actions below the fold without context

### 3. UI Consistency

- Same concept labeled differently across screens
- Similar actions using different button labels or placements
- Inconsistent loading, empty, and error states
- Inconsistent validation or inline help

### 4. Domain Language Quality

- German domain terminology drift in UI labels
- Mismatch between UI wording and backend/domain concepts
- Labels that are technically correct but likely unclear for volunteers

### 5. Frontend / Backend Interaction Friction

- Screens that appear to need more data than the backend provides cleanly
- UI workarounds caused by awkward payload shapes
- Validation or error handling gaps visible at the component boundary

Only include frontend/backend interaction findings here when the user-visible symptom is the main problem. If the core issue is a payload mismatch, type drift, or cross-layer inconsistency, that belongs in Stage 1. If the core issue is redundant data flow or needless abstraction, that belongs in Stage 2.

## Audit Method

For each finding:

1. Verify it from code, not taste alone.
2. Trace the relevant component tree and backend interaction when needed.
3. Explain the user impact in the context of a busy event workflow.
4. Prefer concrete, small improvements over vague design opinions.

## What Not to Report Here

Do **not** use this prompt for:

- request/response drift, JSON field mismatches, or schema/type disagreement across layers
- factory files, wrapper layers, or query abstractions that are merely unnecessary but not user-visible
- failing lint, format, test, build, or integration commands

When a finding clearly belongs elsewhere, say so briefly and point it to the right stage instead of forcing it into this output.

## Output Format

For each finding, provide:

1. **Category**: `Workflow Friction`, `Mobile Layout`, `Terminology`, `Feedback / Error State`, `Validation UX`, or another precise label
2. **What**: the concrete issue
3. **Where**: exact files and line numbers
4. **User impact**: why this matters in real use
5. **Suggested improvement**: pragmatic change, not redesign theater
6. **Effort**: `Small` / `Medium` / `Large`
7. **Priority**: `Quick Win`, `Worth Scheduling`, or `Needs Careful Refactor`

Then add:

## UX Backlog

1. **Fast usability wins**
2. **Consistency fixes to batch together**
3. **Areas that already look solid**

If you notice likely issues that belong in another stage, add a short section:

## Out-of-Scope Notes

- Brief bullet list of items that should be checked with Stage 1, 2, or 3 instead

## Constraints

- Do not suggest broad redesigns unless the current flow is clearly broken.
- Stay within the repo's existing architecture and product scope.
- Do not turn this into a general architecture or correctness audit; keep the focus on user-visible clarity, speed, and error resistance.
- Focus on clarity, speed, and error resistance for volunteers on smartphones.
