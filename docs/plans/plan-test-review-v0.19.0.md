# Plan: Test review and release readiness v0.19.0

**Goal:** Every scenario and use case has a correct test at the cheapest level. CI, e2e and release workflows are correct. `main` is ready for the v0.19.0 tag.

**Files:** test files under `backend/`, `frontend/src/`, `e2e/`, `windows/`, `reverse-proxy/`, `resolver/`; `.github/workflows/`; `scripts/`; `CHANGELOG.md`; `docs/leitfaden/`.

## Checklist

- [ ] Review: domain, seed and event contracts
- [ ] Review: api/kasse
- [ ] Review: fiskal, TSE and DSFinV-K
- [ ] Review: druck, stammdaten, reporting, auth and repositories
- [ ] Review: frontend tests
- [ ] Review: e2e suite
- [ ] Review: workflows, scripts, Windows and proxy tests, release readiness
- [ ] Verify the findings
- [ ] Fix the confirmed findings
- [ ] Gates: `make check`, `make lint`, integration tests, e2e
- [ ] Release prep v0.19.0, CI green on `main`
