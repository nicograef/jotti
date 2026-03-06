# Improvement Recommendations — jotti

> Based on analysis of the jotti codebase and cross-referencing patterns from `/workspaces/admin` (infrastructure knowledge base) and `/workspaces/lexiban` (full-stack IBAN validator).

## Methodology

Each recommendation is evaluated against jotti's specific context: a **nonprofit event ordering system** used by volunteers on smartphones, deployed on a single VPS via Docker Compose. The project is at **46% feature completion** (23/50 requirements) with a strong backend foundation and significant frontend gaps.

Recommendations that don't apply to jotti's context — regardless of how good they are in general — are explicitly rejected with reasoning.

---

## 1 · Frontend Testing (Critical)

### Problem

Zero test files exist in the frontend. No test framework is configured. `pnpm test` is not defined. CI runs lint and build but cannot catch runtime regressions, broken interactions, or wrong business logic in the UI.

This is the single biggest risk in the project. The service UI is the **primary interface** — volunteers use it under pressure during live events. A broken drawer, a wrong price display, or a silent swallowed error means lost revenue or wrong charges with no safety net.

### Recommendation

Set up Vitest + React Testing Library. Focus tests on **business-critical paths only** — not on achieving coverage percentages.

**Priority test targets (in order):**

1. **`drawerUtils.ts`** — `selectVariants()`, `calculateTotalPrice()`. Pure functions, easy to test, high impact. Wrong calculations = wrong money.
2. **`formatCents()`** — Displays money everywhere. Must handle 0, negative, large values correctly.
3. **`useFetch` hook** — Core data fetching pattern. Test loading/error/success states.
4. **Auth flow** — `AuthSingleton` login/logout, token handling, 401 interceptor behavior.
5. **Order/Payment/Delivery drawers** — Integration tests verifying the full user flow: select items → confirm → API call triggered with correct payload.

**What NOT to test:** shadcn/ui components (tested upstream), pure layout components, simple prop-passing wrappers.

### Implementation

```bash
cd frontend
pnpm add -D vitest @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
```

Add to `vite.config.ts`:

```ts
test: {
  environment: 'jsdom',
  setupFiles: ['./src/test/setup.ts'],
}
```

Add to `package.json`:

```json
"test": "vitest run",
"test:watch": "vitest"
```

Add `pnpm run test` to the frontend CI job after lint and before build.

### Effort

Small setup (~1h), then incremental test writing per feature.

---

## 2 · React Error Boundary (Critical)

### Problem

No error boundary exists in the component tree. An unhandled exception in any component — a null reference, a failed JSON parse, a malformed API response — crashes the **entire app** with a white screen. During a live event, this means a volunteer's phone becomes useless until they manually reload.

### Recommendation

Add a single top-level error boundary in `App.tsx` with a German-language fallback UI and a "Neu laden" (reload) button. This is ~30 lines of code with outsized impact.

No need for granular per-component boundaries — jotti's pages are independent (route-based), so a top-level catch with a reload is sufficient.

### Effort

Minimal (~30 min).

---

## 3 · Makefile for Developer Experience (Recommended)

### Problem

All dev commands require remembering Docker Compose flags and paths:

```bash
docker compose -f docker-compose.dev.yml up --build -d
cd backend && go test -tags=unit -race ./...
./test-integration.sh
cd frontend && pnpm lint
```

This isn't painful for a solo developer who knows the project, but it creates friction for onboarding (new volunteers, AI agents) and introduces typo risk.

### Recommendation

Add a Makefile with the ~10 most common commands. Both `admin` and `lexiban` use this pattern successfully. The Makefile in `admin/templates/Makefile` is a good starting point.

```makefile
.PHONY: dev down test test-integration lint

dev:
	docker compose -f docker-compose.dev.yml up --build -d

down:
	docker compose -f docker-compose.dev.yml down

logs:
	docker compose -f docker-compose.dev.yml logs -f

test:
	cd backend && go test -tags=unit -race ./...

test-integration:
	./test-integration.sh

lint-backend:
	cd backend && go vet ./... && goimports -l .

lint-frontend:
	cd frontend && pnpm lint

lint: lint-backend lint-frontend

prod-up:
	docker compose up -d --build

prod-down:
	docker compose down
```

**Keep it flat.** No variables, no includes, no clever abstractions. A Makefile is useful only if it's obvious and maintainable.

### What NOT to do

Don't add a Vite dev proxy to remove the nginx dependency (suggested from lexiban analysis). jotti's dev setup already works with nginx reverse proxy, and the production setup relies on nginx too. Having nginx in dev ensures dev/prod parity. Removing it saves nothing and introduces a "works in dev, breaks in prod" risk for CORS, CSP headers, and WebSocket routing.

### Effort

Minimal (~30 min). Update `DEVELOPMENT.md` to reference `make` targets.

---

## 4 · golangci-lint Configuration File (Recommended)

### Problem

CI runs `golangci-lint v2.6.1` but with **default configuration** — no `.golangci.yml` exists. This means:

- Lint rules are implicit and version-dependent (upgrading may silently change behavior)
- No project-specific tuning (e.g., disabling irrelevant linters, setting line length)
- Developers running `golangci-lint` locally may get different results than CI if their version differs

### Recommendation

Add a `.golangci.yml` in `backend/` that explicitly enables/disables linters and pins the configuration. Focus on linters that catch real bugs in jotti's context:

- **errcheck** — unchecked errors (critical for database/HTTP code)
- **govet** — suspicious constructs
- **staticcheck** — Go's most thorough analyzer
- **unused** — dead code
- **ineffassign** — useless assignments
- **gosec** — security issues (SQL injection, hardcoded creds)

Disable noisy linters that don't add value for jotti's size: `wsl`, `gofumpt` (goimports already used), `exhaustive` (overkill for small enums).

### Effort

Small (~1h). Run locally, fix any new findings, commit config.

---

## 5 · Production Hardening: Resource Limits (Recommended)

### Problem

Neither `docker-compose.yml` (prod) nor `docker-compose.dev.yml` define CPU/memory limits. On a single VPS, a runaway process (e.g., event replay on a large table, or a memory leak) could consume all resources and take down the entire system — including the database.

### Recommendation

Add `deploy.resources.limits` to the production compose file for backend and frontend services:

```yaml
deploy:
  resources:
    limits:
      memory: 256M # backend
      cpus: "1.0"
```

PostgreSQL should keep its own limits via `shared_buffers` and `work_mem` in a custom conf, not via Docker limits (to avoid OOM kills on the database).

Don't bother with resource limits in `docker-compose.dev.yml` — they create friction during development without benefit.

### Effort

Small (~30 min).

---

## 6 · First-Deploy Automation Script (Nice-to-have)

### Problem

jotti has four Docker Compose files (dev, initial-cert, staging, prod) and the cert bootstrap process requires specific manual steps documented across `DEVELOPMENT.md`. The `admin` project solved this with `prod-init.sh` — a script that checks DNS, requests the initial cert, and starts the full stack.

### Recommendation

Port `admin/scripts/prod-init.sh` pattern to jotti. One script that:

1. Checks DNS resolution for the configured domain
2. Runs `docker-compose.initial-cert.yml` to get the first cert
3. Switches to full production compose
4. Verifies HTTPS is working

This is primarily useful if jotti will be deployed to **more than one VPS** (e.g., different events). For a single permanent server, it's a one-time manual step.

### Effort

Medium (~2-3h). Requires testing against actual domain.

---

## Rejected Recommendations

These patterns were identified from `admin` and `lexiban` but **do not apply** to jotti:

### ❌ Vite Dev Proxy (from lexiban)

**Reason:** jotti's entire stack runs in Docker Compose with nginx handling routing. Dev/prod parity is more valuable than removing the nginx container. lexiban uses Vite proxy because its local dev runs without Docker (`make be` + `make fe`). jotti's architecture is different.

### ❌ Feature-Module Frontend Structure (from lexiban)

**Reason:** lexiban has a single feature (IBAN validation). jotti has two distinct areas (admin + service) with separate route guards, different roles, and different use cases. The current `admin/` + `service/` split maps directly to the domain. Reorganizing into feature modules would break the clear admin/service separation that mirrors the backend route structure.

### ❌ Architecture Decision Records (from lexiban)

**Reason:** ADRs are valuable for teams making decisions collaboratively. jotti is a small project with decisions documented implicitly in `AGENTS.md` and `ANFORDERUNGEN.md`. Adding formal ADR documents would create documentation overhead without a reviewing audience. The existing `AGENTS.md` already serves as the "why" document for architectural choices.

### ❌ API Documentation / OpenAPI Spec

**Reason:** jotti's API is internal — consumed only by its own frontend. There are no third-party consumers. The endpoints are documented in `AGENTS.md` and the route files are self-documenting. Maintaining a separate OpenAPI spec would be pure overhead with no consumer.

### ❌ E2E Tests (Playwright/Cypress)

**Reason:** Not yet. jotti needs **unit tests first** (Recommendation 1). E2E tests are expensive to write and maintain, require browser automation infrastructure, and are fragile. For a nonprofit event tool with a small team, the ROI of E2E tests is low compared to testing critical business logic (price calculations, variant selection, payment flows) with fast unit/integration tests. Revisit when the core unit test suite is established and feature completion exceeds 80%.

### ❌ Dependency Scanning (Dependabot/Snyk)

**Reason:** jotti has 7 Go dependencies and 23 npm packages — all mainstream, well-maintained libraries. The attack surface from dependencies is minimal. GitHub's built-in security alerts (enabled by default on GitHub repos) already cover this. Adding Dependabot creates noise (automated PRs for patch versions) without meaningful security improvement at this scale.

### ❌ Bundle Size Analysis in CI

**Reason:** jotti is a mobile-first app used on a local WiFi network at events. Bundle size matters, but the app is small (shadcn/ui + React, no heavy libraries). Monitoring bundle size in CI adds complexity for a problem that doesn't exist yet. If the bundle grows noticeably (>1MB gzipped), add it then.

### ❌ Pre-commit Hooks

**Reason:** CI already enforces goimports, go vet, golangci-lint, and ESLint with zero-warning policy. Pre-commit hooks duplicate this enforcement locally. For a small project, the feedback loop from CI (push → check → fix) is fast enough. Pre-commit hooks add setup friction for every developer machine.

---

## Summary

| #   | Recommendation             | Priority     | Effort  | Impact                         |
| --- | -------------------------- | ------------ | ------- | ------------------------------ |
| 1   | Frontend Testing (Vitest)  | **Critical** | Medium  | Catches regressions in UI      |
| 2   | React Error Boundary       | **Critical** | Minimal | Prevents white-screen crashes  |
| 3   | Makefile                   | Recommended  | Minimal | Better DX, easier onboarding   |
| 4   | golangci-lint Config       | Recommended  | Small   | Explicit, reproducible linting |
| 5   | Production Resource Limits | Recommended  | Small   | Prevents resource exhaustion   |
| 6   | First-Deploy Script        | Nice-to-have | Medium  | One-command production setup   |

**Execution order:** 2 → 1 → 3 → 4 → 5 → 6. Start with the error boundary (30 min), then set up Vitest and write the first tests alongside feature work. The remaining items can be done in any order as time permits.
