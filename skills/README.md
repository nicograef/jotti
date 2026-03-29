# Skills

Agent skills for GitHub Copilot — copy individual skill directories into project repos as needed.

## When to Use Which Skill

| Problem | Skill | Directory |
|---------|-------|-----------|
| Unclear requirements, need to ask questions | **Clarify** | [clarify/](clarify/) |
| Planning a feature from PRD or task description | **Create Plan** | [create-plan/](create-plan/) |
| Writing a product requirements document | **Write a PRD** | [write-a-prd/](write-a-prd/) |
| Executing an existing plan step by step | **Implement Plan** | [implement-plan/](implement-plan/) |
| Building features test-first (red-green-refactor) | **TDD** | [tdd/](tdd/) |
| Code review, cross-layer consistency audit | **Code Audit** | [code-audit/](code-audit/) |
| Mobile UX, UI consistency, workflow friction | **UX Review** | [ux-review/](ux-review/) |
| Exploring multiple API / interface designs | **Design Interface** | [design-interface/](design-interface/) |
| Finding refactoring / deepening opportunities | **Improve Architecture** | [improve-architecture/](improve-architecture/) |
| Syncing a project with handbook best practices | **Handbook Sync** | [handbook-sync/](handbook-sync/) |
| Extracting DDD glossary terms | **Ubiquitous Language** | [ubiquitous-language/](ubiquitous-language/) |
| Extracting text/tables from PDF files | **PDF Extract** | [pdf-extract/](pdf-extract/) |
| Extracting text/tables from Word (.docx) files | **Word Extract** | [docx-extract/](docx-extract/) |
| Extracting data from Excel (.xlsx) workbooks | **Excel Extract** | [xlsx-extract/](xlsx-extract/) |

## Typical Workflow

1. **Clarify** → gather requirements
2. **Write a PRD** → formalise into a document
3. **Create Plan** → break PRD into vertical slices
4. **Implement Plan** → execute slices with **TDD**
5. **Code Audit** → review the result
6. **Improve Architecture** → identify deepening opportunities

## Adding a New Skill

See [.github/instructions/skills.instructions.md](../.github/instructions/skills.instructions.md) for format requirements.
