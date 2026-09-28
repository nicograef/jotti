# AGENTS.md

jotti: a free mobile point-of-sale system (mPOS) for German non-profits (e.V., gGmbH, gUG, foundations, church bodies) at temporary catering events. Typical use: club festivals, Christmas markets, concerts; 2–3 times a year, 5–50 tables, 5–30 volunteer helpers. Service staff take orders, collect payment and cancel per table in the browser of their own smartphones (BYOD). Admins manage products, tables and users. Self-hosted via Docker Compose, mobile-first, under a proprietary source-available licence (non-commercial, usage agreement required).

Non-goals: `docs/produktbeschreibung.md` (product scope) and `docs/anforderungen.md` (excluded features, each with its reason).

**Compliance.** jotti is an electronic recording system under § 1 KassenSichV and needs a TSE under § 146a AO. `docs/compliance.md` holds the TSE, DSFinV-K and ELSTER details. The original texts of the statutes and specs (AO, UStG, KassenSichV, GoBD, DSFinV-K, BSI TR-03153, fiskaly API) lie in `docs/rechtsquellen/`; consult them before the web.

## Rules

- Language: code, comments, commits and agent instructions in English; UI strings, `docs/` and `.env.example` in German.
- Domain terms are German in code too (Bestellung, Zahlung, Ausgabe, Stornierung, Tisch, Position); infrastructure code (auth, config, DB) is English. `docs/language.md` maps each term per layer.
- Every API endpoint is POST; the only exception is `GET /health`.
- Money is an `int` in cents, never a float.
- Kasse operations are event-sourced. The `kassenjournal` table is append-only: never update or delete a row. The projection `tisch_sessions` and the CRUD entity `kassensitzungen` change in the same transaction.
- Master data (users, products, tables) is CRUD with soft deletes via `status = 'deleted'`.
- Both sides validate with schemas: `zog` in the backend, Zod in the frontend.
- The frontend has no global state store, only React hooks and singletons.
- The frontend calls the API only through the backend classes, never through `fetch()`. Each uses the `BackendClient` interface from `frontend/src/lib/Backend.ts`.
- The backend filters, aggregates and shapes data; the frontend shows what it gets.
- Domain structs (`backend/domain/`) carry no `json` tags. Those belong on response DTOs in `api/<domain>/http/` and on event data structs only.
- `backend/sqlc/dbgen/` is generated and never edited; run `make sqlc` after a query change.
- Ask before adding a dependency or changing the Docker or reverse-proxy configuration.
- No secrets or passwords in the code.

## Freeze discipline

Production instances hold data under a retention duty. Persisted data (the existing DB schema, event JSON) is never changed.

- DB schema: changes only as a new additive migration `NN_<name>.up.sql`, numbered in sequence, forward-only, no down migrations. `01_initial.up.sql` is frozen. Rules and reasons: `database/migrations/README.md`.
- Event formats: the event JSON contracts are frozen, guarded by `backend/domain/kasse/event_json_contract_test.go`. A change is a new event version (`:vN`), never an edit in place. Old events are never migrated or reinterpreted.
- Backend API: endpoints and formats may change when the frontend and the print relay follow in the same release. They ship together, so the API has no versions.

## Product conservatism

jotti stays minimal for volunteers who operate it under stress. A feature must justify the complexity it puts on the teams and on the codebase; when in doubt, leave it out.

- Warning signs of feature creep: a status nothing else depends on; recording what paper, a shout or trust already covers; configurability nobody asked for; features built in advance.
- Field feedback beats feature ideas. A feature the field exposes as ballast is a removal candidate. Precedent: D01 in `docs/decisions.md`.
- What meets a real need (compliance, field feedback, the core workflow) is built completely.
- `docs/decisions.md` holds one line per binding decision. A line is never rewritten; a superseded line gets "ersetzt durch DNN". This file is the one exception to the current-state rule.
- A decision line never carries a file path; it names the concept, so a rename or move cannot make it false.

## Start

```sh
bash scripts/setup-dev-tools.sh   # once; then put $(go env GOPATH)/bin on PATH
make init                         # .env with generated secrets
make up                           # dev stack via docker compose, frontend on http://localhost
```

`make help` lists every target. `docs/README.md` says which doc answers which question.

## Gate

- `make check`: the fast gate, without integration tests.
- `make verify`: the full gate; it starts its own Postgres on port 5432.
- `make lint` after a code change.
