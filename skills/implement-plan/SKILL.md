---
name: implement-plan
description: >-
  Work through the next open section of a plan.md, completing tasks
  sequentially and checking them off. Use when the user wants to execute
  an existing implementation plan one section at a time.
---

# Implement Plan

Read the referenced plan.md and work through **one section** at a time.

## Workflow

1. **Read the plan.md** and find the next section with open tasks (`- [ ]`)
2. **Read the `Context:` block** of that section — it lists the relevant files
3. **Work through the tasks sequentially** — top to bottom
4. **Check off each task immediately** (`- [ ]` → `- [x]`) after completing it
5. **After the last task**: run the project's build, lint, and test suite
6. **Stop** — do not start the next section

## Guidelines

- Prefer simple, clear, idiomatic solutions
- No performance optimisation at the cost of readability
- Small local duplication is fine when it makes the code more understandable
- Suggest a Conventional Commit message when done
