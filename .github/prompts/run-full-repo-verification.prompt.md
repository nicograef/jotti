---
description: "Run jotti's full verification commands, including linting, formatting checks, unit tests, builds, and integration tests, and summarize failures with likely root causes."
argument-hint: "Optional target, e.g. make check, make check-full, make verify, backend only, frontend only"
agent: "agent"
---

# Run Full Repo Verification for jotti

Use this as **Stage 3** of the repo-quality audit workflow.

This prompt is for executing the repo's verification commands and producing a high-signal failure report. It is not a code-change prompt unless the user explicitly asks for fixes afterward.

## Default Behavior

Unless the user overrides it, run:

1. `make verify`

If that fails, continue with targeted follow-up commands only as needed to isolate the problem more precisely. Prefer existing Make targets first.

## Goals

- Verify whether the repo currently passes its intended local quality gate
- Surface exact failing steps, commands, and likely root causes
- Distinguish code failures from environment/tooling failures
- Avoid noisy raw logs; summarize the important parts

## Verification Scope

Include, as applicable:

- formatting checks
- linting
- backend unit tests
- frontend unit tests
- backend build
- frontend build
- backend integration tests against a real database

## Reporting Rules

If a command fails, report:

1. **Step**: which target or command failed
2. **Type**: `Format`, `Lint`, `Unit Test`, `Integration Test`, `Build`, or `Environment`
3. **Where**: exact files and lines if available
4. **Failure summary**: the smallest useful error summary
5. **Likely root cause**: concrete explanation, not generic guesswork
6. **Next debugging step**: the best follow-up command or code area to inspect

If all checks pass, say so explicitly and mention any residual gaps, for example the absence of browser-level E2E tests.

## Constraints

- Do not edit code.
- Do not hide environment problems; separate them from product defects.
- Prefer `make verify`, `make check-full`, `make check`, and other existing repo commands over ad-hoc command sequences.
- Keep the final output concise and decision-oriented.
