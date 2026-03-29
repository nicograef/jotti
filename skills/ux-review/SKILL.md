---
name: ux-review
description: >-
  Review a frontend for mobile UX, UI consistency, workflow friction, and
  terminology drift. Use when the user wants a UX audit of a mobile-first
  web app or needs to find usability problems on small screens.
---

# Mobile UX Review

Find UX problems that slow down users on smartphones during real-world usage.

## Review Areas

### Workflow Friction

- Too many taps for frequent actions
- Hidden or unclear next steps
- Missing feedback after save, submit, or destructive actions
- Error states that are hard to recover from

### Mobile-First Quality

- Components that break on narrow screens
- Dense tables or forms without mobile fallback
- Touch targets that are too small or too close together
- Important actions below the visible fold

### UI Consistency

- Same concept labelled differently across screens
- Similar actions with different button labels or placements
- Inconsistent loading, empty, and error states

### Domain Language

- Terminology drift in UI labels vs. backend/domain terms
- Labels that are technically correct but unclear to end users

## Output

Per finding: **Category** → **What** → **Where** (file:lines) →
**User Impact** → **Suggestion** → **Effort** (S/M/L).

At the end: quick wins first, then consistency fixes to batch.

## Quality

- Before presenting results, run the self-review checklist from AGENTS.md (Quality Principles) — applied to the quality of the review artifact. Surface issues in the chat only if found.
- After task completion, include a human-readable summary paragraph alongside the commit message (see AGENTS.md, Git Workflow).
