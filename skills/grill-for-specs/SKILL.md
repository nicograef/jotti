---
name: grill-for-specs
description: >-
  Prevents the agent from making assumptions by forcing structured clarifying
  questions before acting. Use when the user wants thorough spec-gathering,
  disambiguation, or wants the agent to ask before assuming. Applies to any
  task type.
---

# Grill for Specs

Never assume — always ask. Before acting on any task, identify ambiguities and
unknowns, then resolve them through structured questions with the AskQuestion
tool.

## Workflow

Work through up to **3 sequential rounds** of questions. Each round: 2–5
structured questions, drilling deeper based on prior answers.

### Round 1 — Scope & Intent

Identify what is ambiguous or underspecified in the user's request.
Ask 2–3 questions covering the biggest unknowns first.

### Round 2 — Drill Deeper

Based on Round 1 answers, ask 2–3 follow-up questions on remaining gaps,
edge cases, or conflicting constraints.

### Round 3 — Final Confirmation (if needed)

Resolve any last ambiguities. Confirm critical decisions before proceeding.

### After Questions

Update the current plan or document with every decision made.
Then proceed with the task.

## Question Guidelines

1. **Always recommend.** Every question must include a recommendation with
   brief reasoning. Label it clearly (e.g., "recommended" in the option label,
   or a note in the prompt).
2. **Structured over free-text.** Use the AskQuestion tool with concrete
   options. If a question seems open-ended, convert it to multiple-choice with
   an "Other (specify)" escape hatch.
3. **Context before question.** The prompt should explain *why* the question
   matters so the user can make an informed choice.
4. **Group related choices.** Use `allow_multiple: true` when the user may
   legitimately pick more than one option.
5. **Max 5 questions per round.** Prioritise — ask the most impactful
   questions first.

## Escalation

If the user declines to answer or says "just do it":

1. Proceed with the best-guess default for each unanswered question.
2. Document every assumption in the plan or output as a clearly marked callout
   (e.g., a blockquote prefixed with **Assumption:**).
3. Continue with the task.

## Constraints

- Max **3 rounds** of questions — after Round 3, stop asking and proceed.
- Max **5 questions** per round.
- Always use the **AskQuestion tool** when available; fall back to
  conversational questions only if the tool is unavailable.
- Do not repeat questions the user has already answered.
