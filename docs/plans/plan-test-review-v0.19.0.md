# Plan: Test review and release readiness v0.19.0

**Goal:** Every scenario and use case has a correct test at the cheapest level. CI, e2e and release workflows are correct. `main` is ready for the v0.19.0 tag.

**Files:** test files under `backend/`, `frontend/src/`, `e2e/`, `windows/`, `reverse-proxy/`, `resolver/`; `.github/workflows/`; `scripts/`; `CHANGELOG.md`; `docs/leitfaden/`.

## Decisions (owner, 2026-11-09)

- Pfand leaves station and Zusatzdrucker Arbeitsbons; the print path reads the Pfand flag from Stammdaten. New D-line.
- e2e: the four correctness fixes plus onboarding, role guard, Auto-Pfand and UI order to print job.
- The migrate Dockerfile curl gets `--retry 5 --retry-delay 5 --retry-all-errors`.
- The DSFinV-K export requires a Kassensitzung number; the default-session branch goes.
- Tisch grouping and natural sort stay in the frontend as presentation. New D-line.
- CHANGELOG `[1.0.0]` gains the missing user-facing lines.
- The lead tags v0.19.0 with `make prod-release` once every CI job on main is green.
- Taken without asking: the Umbuchung Abgang shows negative MENGE in lines.csv (DSFinV-K 4.2.8).

## Checklist

- [x] Review: domain, seed and event contracts
- [x] Review: api/kasse
- [x] Review: fiskal, TSE and DSFinV-K
- [x] Review: druck, stammdaten, reporting, auth and repositories
- [x] Review: frontend tests
- [x] Review: e2e suite
- [x] Review: workflows, scripts, Windows and proxy tests, release readiness
- [x] Verify the findings
- [x] Fix the confirmed findings
- [x] Gates: `make check`, `make lint`, integration tests, e2e
- [ ] Release prep v0.19.0, CI green on `main`
