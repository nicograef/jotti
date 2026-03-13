---
description: "Run the full staged jotti repo-quality audit in the recommended order: cross-layer consistency, readability-first simplification review, executable verification, and mobile UX review."
argument-hint: "Optional scope, e.g. whole repo, auth, admin CRUD, service flow, reporting, or start at a specific stage"
agent: "agent"
---

# Run Repo Quality Audit for jotti

Use this prompt when you want the **recommended full audit order** without having to remember which prompt should come first.

This is the **workflow prompt** for the staged repo-quality audit.

## Default Order

Run the audit in this order unless the user explicitly asks for a subset or a different starting point:

1. **Stage 1: Cross-Layer Consistency**
2. **Stage 2: KISS Simplification**
3. **Stage 3: Full Repo Verification**
4. **Stage 4: Mobile UX Consistency**

## Stage Roles

### Stage 1: Cross-Layer Consistency

Use [.github/prompts/find-cross-layer-consistency-issues.prompt.md](.github/prompts/find-cross-layer-consistency-issues.prompt.md).

Goal:

- find mismatches between frontend, backend, database, SQL, Go structs, TypeScript types, schemas, and request/response payloads

Ask here:

- do the layers still agree?

### Stage 2: KISS Simplification

Use [.github/prompts/find-kiss-simplifications.prompt.md](.github/prompts/find-kiss-simplifications.prompt.md).

Goal:

- find readability and consistency improvements: stale abstractions, low-value indirection, dead code, and structural complexity that can be removed safely

Ask here:

- what is harder to read or reason about than it should be?

### Stage 3: Full Repo Verification

Use [.github/prompts/run-full-repo-verification.prompt.md](.github/prompts/run-full-repo-verification.prompt.md).

Goal:

- run the executable repo checks and summarize failures or confirm the repo passes its local quality gate

Ask here:

- does the repo currently pass its checks?

### Stage 4: Mobile UX Consistency

Use [.github/prompts/review-mobile-ux-consistency.prompt.md](.github/prompts/review-mobile-ux-consistency.prompt.md).

Goal:

- find workflow friction, UI inconsistency, unclear wording, missing feedback, and mobile-first problems for real volunteers using phones

Ask here:

- what will confuse, slow down, or trip up users?

## How to Run the Workflow

### Default behavior

If the user does not narrow the scope, audit the whole repo.

For each stage:

1. follow that stage's scope strictly
2. do not blur stage boundaries
3. collect only verified findings
4. keep the result concise and decision-oriented

### Handoff rules

- If Stage 1 finds correctness or alignment issues, keep them in Stage 1 even if they also create complexity.
- If Stage 2 finds readability or maintainability issues in otherwise correct structures, keep them in Stage 2.
- If Stage 3 finds failing commands, report them there rather than restating them as architecture or UX issues.
- If Stage 4 notices a likely payload mismatch or stale abstraction, note it briefly as out-of-scope and point it back to the right earlier stage.

### When to stop early

- If the user asks for only one stage, run only that stage.
- If Stage 1 uncovers severe correctness drift, you may note that Stage 2 and Stage 4 findings should be interpreted cautiously until those issues are fixed.
- If Stage 3 is blocked by environment problems, report that clearly and continue only with read-only stages when useful.

## Final Output Structure

Return results in this order:

1. **Stage 1 Summary**
2. **Stage 2 Summary**
3. **Stage 3 Summary**
4. **Stage 4 Summary**
5. **Recommended Next Actions**

For each stage summary:

- include only the most important findings
- mention if the stage found nothing significant
- mention if some likely issues were deliberately left for another stage

For **Recommended Next Actions**, prioritize:

1. correctness bugs first
2. then readability/consistency simplification wins
3. then failing verification steps
4. then UX improvements

## Constraints

- Do not collapse all stages into one undifferentiated audit.
- Do not repeat the same finding across multiple stages unless the user explicitly asks for overlap.
- Prefer the existing stage prompts over inventing a new taxonomy.
- Keep the workflow easy to run again later with the same expectations.
- In Stage 2, do not prioritize performance-only optimization or cleverness over readability and idiomatic simplicity.
