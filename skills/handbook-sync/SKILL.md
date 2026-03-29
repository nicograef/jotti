---
name: handbook-sync
description: >-
  Audit a project against the handbook knowledge base and propose targeted
  improvements. Use when the user wants to align a project with handbook
  best practices, sync templates, review CI/Docker/infra setup, or bootstrap
  a new project from handbook patterns. The user provides the absolute path
  to the handbook repo on disk.
---

# Handbook Sync

Compare a target project against the handbook and propose **project-appropriate**
improvements. Never copy verbatim — always adapt to the target project's stack,
conventions, and maturity.

## Input

- **Handbook path** (required): The absolute path to the handbook repo on
  disk, provided by the user in the prompt. Example:
  `Use handbook-sync with handbook at /home/nico/repos/handbook`
- The target project is the current workspace / working directory.
- The target project must have at least a `README.md` or `AGENTS.md` to
  determine its stack and conventions.

If the user does not provide a handbook path, ask for it before proceeding.
Verify the path exists and contains a `README.md` with a handbook structure
before starting the workflow.

## Workflow

### 1. Discover the target project

Read the target project's key files to understand its stack and setup:

- `AGENTS.md`, `.github/copilot-instructions.md` — existing agent instructions
- `Makefile` — available commands
- `docker-compose.yml`, `docker-compose.prod.yml` — container setup
- `.github/workflows/` — CI/CD configuration
- `devcontainer.json` or `.devcontainer/` — dev environment
- `.gitignore`, `.editorconfig` — repo hygiene
- `README.md` — project description and structure

Build a mental model: **What stack? What maturity? What's already configured?**

### 2. Identify relevant handbook areas

Based on the target project's stack, select only the handbook areas that apply.
Skip everything that does not match.

| Handbook area | Relevant when target project uses |
|---|---|
| `templates/Makefile` | Any project (adapt targets to stack) |
| `templates/docker-compose.yml` | Docker |
| `templates/docker-compose.prod.yml` | Docker in production |
| `templates/nginx-tls.conf` | Nginx reverse proxy |
| `templates/ci.yml` | GitHub Actions |
| `templates/devcontainer.json` | Dev Containers / Codespaces |
| `templates/AGENTS.md` | Copilot Agent Mode |
| `templates/copilot-instructions.md` | Copilot (any mode) |
| `guides/docker-setup.md` | Docker |
| `guides/docker-multi-stage-builds.md` | Docker with multi-stage builds |
| `guides/github-actions-cicd.md` | GitHub Actions |
| `guides/copilot-agent-setup.md` | Copilot Agent Mode |
| `guides/go.md` | Go backend |
| `guides/react.md` | React frontend |
| `guides/provision-server.md` | VPS / bare-metal deployment |
| `guides/letsencrypt-docker.md` | TLS with Let's Encrypt |
| `guides/nginx-reverse-proxy.md` | Nginx |
| `guides/postgresql-operations.md` | PostgreSQL |

### 3. Compare and analyse

For each relevant handbook area, compare the handbook reference with the
target project's current state. Categorise findings:

- **Missing** — handbook pattern not present, would add value
- **Outdated** — present but behind handbook (older image tags, deprecated
  flags, missing security headers)
- **Divergent** — different approach that may be intentional — flag, don't
  auto-fix
- **Aligned** — already matches handbook — skip

### 4. Present findings

Output a prioritised table:

| # | Category | Area | Finding | Handbook ref | Effort |
|---|----------|------|---------|-------------|--------|
| 1 | Missing | CI | No lint step in workflow | `templates/ci.yml` | S |
| 2 | Outdated | Docker | Compose uses v3.8 syntax | `templates/docker-compose.yml` | S |
| 3 | Divergent | Makefile | Uses npm scripts instead of Make | `templates/Makefile` | — |

**Rules for presentation:**

- Sort by impact: security > correctness > consistency > convenience.
- Mark **Divergent** items clearly — these need user judgment.
- Include the handbook file reference so the user can review the source.
- Do NOT apply changes yet.

### 5. User selects changes

Wait for the user to pick which findings to apply. The user may:

- Accept individual items by number
- Accept all non-divergent items
- Reject or defer specific items
- Ask for more detail on any item

### 6. Apply changes

For each accepted finding:

1. **Read** the handbook reference file.
2. **Read** the target project's corresponding file (or note its absence).
3. **Adapt** — do not copy. Adjust to the target project's:
   - Language and framework (Go vs Node vs Java)
   - Directory structure
   - Naming conventions
   - Existing patterns and style
4. **Write** the change (edit existing file or create new file).
5. **Verify** — run the project's build/lint/test if available.

After all changes, list what was modified.

## Constraints

- **Adapt, never copy.** Every handbook pattern must be translated to fit the
  target project. A Go project does not get `pnpm` commands. A project without
  Docker does not get Compose files.
- **Respect existing decisions.** If the target project has a working pattern
  that differs from the handbook, flag it as Divergent — do not overwrite
  without explicit user approval.
- **One source of truth.** Do not duplicate handbook content into the target
  project. If the target project needs ongoing reference, add a note in its
  `AGENTS.md` pointing to the handbook path.
- **Minimal changes.** Apply only what the user approved. Do not sneak in
  extra improvements.
- **No version guessing.** If a handbook template references a specific version
  (Docker image tag, action version, tool version), verify it is still current
  via web search before applying.
