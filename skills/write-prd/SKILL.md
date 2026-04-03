---
name: write-prd
description: >-
  Create a PRD through structured clarification, codebase exploration, and
  module design, then save as a local Markdown file. Use when user wants to
  write a PRD, create a product requirements document, or plan a new feature.
---

# Write a PRD

You may skip steps if you don't consider them necessary.

## Workflow

### 1. Gather the problem description

Ask the user for a long, detailed description of the problem they want to
solve and any potential ideas for solutions.

### 2. Explore the codebase

Explore the repo to verify their assertions and understand the current state
of the codebase.

### 3. Clarify ambiguities

Resolve unknowns through **1–3 rounds** of structured questions.

**Rules:**

- **Explore before asking.** If a question can be answered by reading the
  codebase, read the codebase instead of asking the user.
- **Always recommend.** Every question must include a recommended answer.
  Label it clearly (e.g. "(recommended)" in the option label, or a note in
  the prompt) with brief reasoning.
- **Context before question.** The prompt should explain *why* the question
  matters so the user can make an informed choice.
- **Structured over free-text.** Prefer the Ask Question tool when available;
  fall back to conversational questions only if no such tool exists. Present
  concrete options. Convert open-ended questions to multiple-choice with an
  "Other (specify)" escape hatch.
- **Max 5 questions per round.** Prioritise the most impactful unknowns.
- **Stop when resolved.** If all ambiguities are clear after 1 round, stop.
  Continue only if unresolved branches remain.

If the user declines to answer: proceed with recommended defaults and document
each assumption in the PRD as a clearly marked callout (e.g. blockquote
prefixed with **Assumption:**).

### 4. Design modules

Sketch out the major modules you will need to build or modify to complete
the implementation. Actively look for opportunities to extract deep modules
that can be tested in isolation.

A deep module (as opposed to a shallow module) is one which encapsulates a
lot of functionality in a simple, testable interface which rarely changes.

Check with the user that these modules match their expectations. Check with
the user which modules they want tests written for.

### 5. Write the PRD

Once you have a complete understanding of the problem and solution, use the
template below to write the PRD. Save it to `docs/prds/prd-<name>.md`
(create the directory if it doesn't exist). Use a short kebab-case name
derived from the feature (e.g. `prd-user-onboarding.md`).

## PRD Template

```markdown
# PRD: <Feature Name>

## Problem Statement

The problem that the user is facing, from the user's perspective.

## Solution

The solution to the problem, from the user's perspective.

## User Stories

A LONG, numbered list of user stories. Each user story should be in the format
of:

1. As an <actor>, I want a <feature>, so that <benefit>

This list of user stories should be extremely extensive and cover all aspects
of the feature.

## Implementation Decisions

A list of implementation decisions that were made. This can include:

- The modules that will be built/modified
- The interfaces of those modules that will be modified
- Technical clarifications from the developer
- Architectural decisions
- Schema changes
- API contracts
- Specific interactions

Do NOT include specific file paths or code snippets. They may end up being
outdated very quickly.

## Testing Decisions

A list of testing decisions that were made. Include:

- A description of what makes a good test (only test external behavior, not
  implementation details)
- Which modules will be tested
- Prior art for the tests (i.e. similar types of tests in the codebase)

## Out of Scope

A description of the things that are out of scope for this PRD.

## Further Notes

Any further notes about the feature.
```

## Quality

- Before presenting results, run the self-review checklist from AGENTS.md (Quality Principles) — applied to the quality of the PRD artifact. Surface issues in the chat only if found.
- After task completion, include a human-readable summary paragraph alongside the commit message (see AGENTS.md, Git Workflow).
