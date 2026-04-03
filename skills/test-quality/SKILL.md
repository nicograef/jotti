---
name: test-quality
description: >-
  Review and refactor an existing test suite. Use when the user wants to reduce
  test count, remove implementation-detail tests, improve test readability, or
  clean up a test suite that has grown noisy or brittle.
---

# Test Quality Review

## Philosophy

**Goal**: a lean suite where every test earns its place by verifying observable
behavior through the public API. Tests that break on internal refactors without
any behavior change are liabilities — they slow down development and erode trust
in the suite.

**Scope**: this skill improves existing tests. It does not add new tests — that
is the TDD skill's job.

**The only question that matters per test**: _"Does this test break when
behavior changes, or when implementation changes?"_ Only the former is worth
keeping.

See [anti-patterns.md](anti-patterns.md) for a catalog of bad tests and
[evaluation-criteria.md](evaluation-criteria.md) for the tagging rubric.

---

## Workflow

### Step 1 — Discover

Identify everything that needs to be evaluated:

- [ ] Find all test files in the project (`**/*.test.ts`, `*_test.go`,
      `**/*.spec.ts`, `test_*.py`, etc.)
- [ ] Identify the language and test framework in use (Jest, Vitest, Go
      testing, pytest, JUnit, …)
- [ ] Count total test cases and test files
- [ ] Note any test helpers, fixtures, or shared setup files
- [ ] List the files in a compact inventory table (file → test count → framework)

Ask the user if the scope should be limited to specific packages or files before
proceeding.

### Step 2 — Audit

Evaluate every test file. For each test, assign one tag:

| Tag | Meaning |
|-----|---------|
| **Keep** | Tests observable behavior through the public API — no changes needed |
| **Refactor** | Correct intent but coupled to internals or poorly written — rewrite |
| **Delete** | Tests implementation details, is redundant, or tests nothing meaningful |
| **Merge** | Overlaps with another test — combine into one assertion-rich test |

Apply the decision rules from [evaluation-criteria.md](evaluation-criteria.md).

Group findings by file. Do not make changes yet.

### Step 3 — Report

Present a structured summary to the user before touching any code:

```
File: src/checkout/checkout.test.ts  (12 tests)
  Keep     4  — [list test names]
  Refactor 5  — [list test names + 1-line reason]
  Delete   2  — [list test names + 1-line reason]
  Merge    1  — [list test names + target]

File: src/cart/cart.test.ts  (8 tests)
  ...

Total: 20 tests → 12 tests after changes (-40%)
```

After presenting the report:

- Explain the biggest quality wins
- **Ask for explicit confirmation before proceeding**
- If the user disagrees with a tag, update before proceeding

Do not start Step 4 until the user confirms.

### Step 4 — Refactor

Work through changes one file at a time:

- [ ] Apply **Merge** first (reduces total test count, simplifies subsequent work)
- [ ] Apply **Refactor** next (rewrite tests to use public interface only)
- [ ] Apply **Delete** last — confirm once more if deleting more than 3 tests at once
- [ ] Remove dead test helpers and fixtures that are no longer referenced
- [ ] Clean up imports left orphaned by deleted tests

Rules during refactoring:
- Rewrite tests to assert on return values and public state, not on internal calls
- Replace mocks of internal collaborators with real implementations or
  in-memory fakes (see [anti-patterns.md](anti-patterns.md) — Mocking Internals)
- Preserve mocks at true system boundaries (HTTP, DB, email, time, randomness)
- Keep test names as behavior descriptions: "user can checkout with valid cart",
  not "calls processPayment"

### Step 5 — Verify

- [ ] Run the full test suite
- [ ] If tests fail: diagnose whether the failure is a regression (protected
      behavior was removed) or a false signal (test was wrong before too)
- [ ] For genuine regressions: restore the deleted test and re-evaluate
- [ ] Report the final before/after count and any regressions found

---

## Constraints

- **Never delete tests without user confirmation** — always show the report first
- **Never add new tests** — out of scope; redirect to TDD skill if coverage gaps exist
- **Never rewrite a test to make it pass** — if behavior broke, the fix is in
  the implementation, not the test
- **Never mock internal collaborators** — mocks belong at system boundaries only
- **Never keep tests that verify call counts or argument order** on internal
  methods — these are implementation-detail tests
- **Never bypass the public interface** to verify state (e.g. querying the DB
  directly after calling a service method)
- **Preserve integration-style tests** even if they are "slow" — they are the
  most valuable tests in the suite
- **Do not refactor implementation code** during this skill — test code only
