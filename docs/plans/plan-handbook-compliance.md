# Plan: Handbook compliance

> Source PRD: n/a. Source: the handbook compliance audit of 2026-09-27 (findings G01–G56) and the owner's rulings on each item.

## Goal

jotti follows `~/r/handbook` wherever no project reason stands against it. Each difference that remains is intended and written down. Every item below was decided by the owner; dropped items are listed so that nobody reopens them.

## Architectural decisions

- **Agent instructions**: `CLAUDE.md` is only `@AGENTS.md`, as in `~/r/gyva`. `AGENTS.md` is English and holds only jotti-specific facts and rules; nothing the global `~/.claude/CLAUDE.md` already states. No Copilot, GitHub-agent or cloud-session content.
- **Language**: code, comments, commits, agent instructions and dev scaffolding are English. UI strings, `docs/` and operator-facing scaffolding (`.env.example`) are German. `AGENTS.md` states this in one line.
- **Push policy**: feature branches and `main` are pushed; never force. A GitHub ruleset on `main` blocks force-push and deletion, with no PR requirement. The repo allows squash merges only.
- **Reverse proxy**: Caddy everywhere. jotti.rocks moves from nginx + certbot to the Caddy build of `reverse-proxy/Dockerfile` with a static Caddyfile. nginx remains only for static serving inside the frontend and website images.
- **Postgres**: stays on 17 until an upgrade path exists (new line in `docs/decisions.md`).
- **Go tests**: no build tag on unit tests; `//go:build integration` stays.
- **Frontend tests**: fake at the `BackendClient` boundary with a real QueryClient; never `vi.mock` the project's own modules.
- **Class helper**: the `cn` package (shadcn upstream since 2026-09-03), approved as a new dependency.
- **Go formatting**: goimports (CI, hook), gopls defaults in the editor; no gofumpt.

## Inventory

- `AGENTS.md`, `CLAUDE.md`, `.claude/settings.json` — agent config to rewrite
- `.claude/workflows/jotti-full-audit.js`, `.claude/workflows/review-phase.js` — deleted
- `.github/copilot-instructions.md`, `.github/instructions/` — deleted
- `.github/workflows/{ci,release,security-scans,fuzz}.yml`, `.github/dependabot.yml` — CI hardening
- `database/migrate/Dockerfile` — golang-migrate download
- `scripts/check-pins.sh` — pin gate, extended to action SHAs
- `scripts/lib.sh — read_env(), tracked_text_files(), wait_for_healthy()` — shared shell helpers
- `scripts/prod-restore.sh`, `packaging/windows/jotti-restore.cmd` — restore paths
- `scripts/prod-backup-verify.sh`, `scripts/prod-update.sh` — drill and update
- `scripts/prod-harden.sh`, `docs/leitfaden/self-hosting.md`, `docs/leitfaden/aktualisieren-backups.md` — operator hardening
- `docker-compose.rocks.yml`, `docker-compose.initial-cert.yml`, `reverse-proxy/nginx.rocks.conf`, `reverse-proxy/nginx.initial-cert.conf`, `scripts/rocks-init.sh`, `docs/jotti-rocks-infra.md` — jotti.rocks stack
- `resolver/Dockerfile`, `reverse-proxy/Dockerfile` — root containers
- `scripts/check-build-tags.sh`, `Makefile` targets `test`, `fuzz`, `lint-backend`, `check-backend` — unit tag plumbing
- `frontend/src/lib/Backend.ts — BackendClient`, `frontend/src/lib/utils.ts — cn()` — test boundary and class helper
- `frontend/src/service/` — split by type today
- `e2e/package.json` — no ESLint, TypeScript 7.0.2
- `scripts/check-prose.sh` — current-state gate; gains the prose caps
- `docs/plans/plan-release-v1.0.0.md` — sections 4–5 hold backlog items

## Resolved decisions

| Item | Ruling |
| --- | --- |
| G01 | SHA-pin every `uses:` with a `# vX.Y.Z` comment; `check-pins.sh` fails on a tag pin |
| G02 | `ARG MIGRATE_SHA256` + `sha256sum -c` in `database/migrate/Dockerfile`, the only download site; CI builds and runs that image |
| G03 | Manual laptop-run pull script for the acme-dns SQLite; no server timer |
| G04, G21, G22, G23 | `persist-credentials: false`; `timeout-minutes` per job; actionlint; Dependabot `docker-compose` ecosystem |
| G05 | Better Stack uptime and SSL-expiry monitors (owner step); Caddy removes the certbot heartbeat need |
| G06 | `psql -1 -v ON_ERROR_STOP=1`, then `vacuumdb --analyze-in-stages` |
| G07 | Decision line: stay on Postgres 17 |
| G08 | Dev and test Postgres on `127.0.0.1`; Vite stays LAN-bound with a comment |
| G09 | Resolver: non-root `USER` with `NET_BIND_SERVICE`. Reverse proxy: stays root (no `caddy-data` ownership migration) with `cap_drop: ALL`, `cap_add: NET_BIND_SERVICE`, `no-new-privileges`, `read_only` |
| G10, G49 | Link `make prod-harden` from the guide; install unattended-upgrades and a key-only sshd drop-in; `ufw limit`, systemd fail2ban backend |
| G11, G12, G16, G31, G36 | Handbook plugin, both workflows, the Fable rule and the model lines are removed |
| G13, G17, G32 | Removed from `AGENTS.md`; the global rules apply (push `main`, decide before asking, ≤ 3-line report) |
| G14, G33 | Ruleset on `main` (no force-push, no deletion); squash-only merges (owner step) |
| G15 | Git identity in the claude.ai/code environment (owner step) |
| G18, G44, G24 | Drop the unit tag; sentinel errors, stateful fakes, `t.Errorf`; goimports as a `go.mod` tool (sqlc stays outside: cgo) |
| G19 | e2e: TypeScript `~6.0`, typed ESLint with eslint-plugin-playwright (approved dependency) |
| G20 | Fake `BackendClient` for every page test |
| G26, G27, G28 | Per-context `.dockerignore`; pnpm fallback without corepack; `make up` and classed `make help` |
| G29, G30 | `.env.example` complete plus drift gate; index in `docs/README.md` plus gate |
| G34 | Copilot-only facts are dropped |
| G35 | Open items of sections 4–5 move verbatim to `docs/backlog.md` |
| G37 | Comments ≤ 2 sentences; long reasoning moves into the doc the comment links |
| G38 | Prose caps gate all prose except `docs/rechtsquellen/`, LICENSE, TERMS, CLA, SERVICE and the Verfahrensdokumentation |
| G39, G40, G41, G42, G43 | One home per fact; historic residue out; link gate scope; script hygiene; shellcheck in lint, `check-repo` collects all failures |
| G46, G47, G48 | `tsc -b` before tests; `service/<feature>/`; typed exported hooks; `cn` via `npx shadcn@latest migrate cn` |
| G50, G51, G52, G53, G54, G55, G56 | Headers ported to the rocks Caddyfile; Postgres `shm_size`/`stop_grace_period`; rocks waits and healthchecks; kassenjournal row-count drill and image prune; "Laufender Betrieb" section; `read_env` strips `\r` and quotes; `.gitignore` sections |
| Dropped | G25 (moot after G18), G45 (website ESLint and tsconfig flags) |

## Open questions / Risks

- **Read-only proxy (Phase 6).** `read_only: true` needs every path Caddy writes (`/data`, `/config`, the rendered Caddyfile, temp files) on a volume or tmpfs.
- **jotti.rocks cutover (Phase 5).** The switch causes brief downtime, and Caddy requests new certificates. The HTTP-01 challenge needs port 80 free during the switch.
- **cn is 0.x (Phase 9).** One known divergence from tailwind-merge (`bg-gradient-to-*`) does not affect jotti's two uses in `components/ui/tabs.tsx`.
- **Prose caps in German (Phase 13).** Compound-heavy sentences make the 20-word cap tight; splitting leitfaden sentences must keep their content (keep-bar ruling).

## Phase 1: Agent config reset

**Depends on**: none

### What to build

`CLAUDE.md` becomes the single line `@AGENTS.md`. `AGENTS.md` is rewritten in English, modelled on `~/r/gyva/AGENTS.md`. It keeps:
- the product identity;
- the compliance scope;
- the freeze discipline;
- the domain rules (POST-only, cents, event sourcing, CRUD with soft deletes, schemas on both sides, ubiquitous language, no global store, API calls via backend classes, backend filtering, no `json` tags in domain, `sqlc/dbgen` untouched, `make sqlc`, ask before new dependencies and Docker changes, no secrets);
- product conservatism with the D01 precedent;
- the `docs/decisions.md` exception and its "replaced by DNN" rule;
- the `docs/rechtsquellen/` pointer;
- the start and gate commands;
- the one-line language scope.

It drops everything the global CLAUDE.md states: communication, the git workflow, verify, ask and web-search rules, the current-state rule, and the post-task summary.

Delete `.github/copilot-instructions.md`, `.github/instructions/` and `.claude/workflows/`. Remove the `handbook@nicograef` plugin and its marketplace from `.claude/settings.json`; keep the formatting hook and the attribution block. Move the docs table into a new `docs/README.md` (German), which lists every doc including `lizenzmodell.md`, `jotti-rocks-infra.md`, `leitfaden/` and `prds/`. `AGENTS.md` links to it.

### Acceptance criteria

- [x] `cat CLAUDE.md` prints only `@AGENTS.md`
- [x] `grep -rniE 'copilot|CLAUDE_CODE_REMOTE|fable|cloud-session' AGENTS.md CLAUDE.md .claude/ .github/` finds nothing
- [x] `test ! -e .github/copilot-instructions.md && test ! -e .github/instructions && test ! -e .claude/workflows`
- [x] `jq '.enabledPlugins["handbook@nicograef"], .extraKnownMarketplaces' .claude/settings.json` prints `null null`
- [x] every `docs/*.md` and `docs/*/` appears in `docs/README.md` (checked by a `ls`/`grep` loop)
- [x] `make check-repo` green

## Phase 2: CI supply-chain hardening

**Depends on**: none

### What to build

Every `uses:` in `.github/workflows/*.yml` points to a full commit SHA with a `# vX.Y.Z` comment, and `scripts/check-pins.sh` fails on any tag pin. Every `actions/checkout` sets `persist-credentials: false`; steps that push tags or publish releases get the token explicitly. Every job has `timeout-minutes`. `database/migrate/Dockerfile` verifies a pinned `MIGRATE_SHA256` before extracting golang-migrate. The two CI steps that curl the tarball (`backend-integration-tests`, `upgrade-path`) build that image instead and run it with `--network host`, so the Dockerfile is the only download site. `.github/dependabot.yml` adds the `docker-compose` ecosystem, grouped with `docker`.

### Acceptance criteria

- [x] `grep -hE '^\s*-?\s*uses:' .github/workflows/*.yml | grep -vcE '@[0-9a-f]{40}'` prints `0`
- [x] a scratch workflow line with `@v5` makes `scripts/check-pins.sh` fail
- [x] `grep -c 'persist-credentials: false' .github/workflows/*.yml` equals the checkout count
- [x] every job in every workflow carries `timeout-minutes` (yq or grep check)
- [x] `docker build -f database/migrate/Dockerfile database` fails with a wrong `MIGRATE_SHA256` and succeeds with the pinned one
- [x] `grep -rn 'golang-migrate/migrate/releases' .github` finds nothing

## Phase 3: Restore and backup safety

**Depends on**: none

### What to build

`scripts/prod-restore.sh` and `packaging/windows/jotti-restore.cmd` restore in one transaction (`psql -1 -v ON_ERROR_STOP=1`) and then run `vacuumdb --analyze-in-stages`. `scripts/prod-backup-verify.sh` checks row counts on `kassenjournal` in the restored copy, not just that tables exist. `scripts/prod-update.sh` runs `docker image prune -f` after a verified update. `docker-compose.prod.yml`, `docker-compose.local.yml`, `docker-compose.release.yml` and `docker-compose.e2e.yml` give Postgres `shm_size: 128mb` and `stop_grace_period: 1m`. `docs/decisions.md` gains the next D line: Postgres stays on 17 until an upgrade path exists, and Dependabot ignores its majors.

### Acceptance criteria

- [x] restoring a dump truncated mid-way via `make prod-restore` against the local stack exits non-zero and leaves the previous data intact (row count before = after)
- [x] `make prod-backup-verify` fails when the `kassenjournal` row count in the restored copy is 0
- [x] `grep -c 'shm_size' docker-compose.{prod,local,release,e2e}.yml` prints 1 for each

## Phase 4: Operator hardening path

**Depends on**: none

### What to build

`scripts/prod-harden.sh` installs and enables unattended-upgrades and writes a key-only sshd drop-in. It uses `ufw limit` for SSH and gives the fail2ban jail the systemd backend, as the handbook's `scripts/setup-server.sh` does. `docs/leitfaden/self-hosting.md` adds a hardening step that links `make prod-harden`. `docs/leitfaden/aktualisieren-backups.md` gains a short "Laufender Betrieb" section: uptime check, disk space, reboot after kernel updates.

### Acceptance criteria

- [x] `shellcheck scripts/prod-harden.sh` clean
- [x] `prod-harden.sh` run in a throwaway Debian/Ubuntu container (or VM) leaves `unattended-upgrades` enabled, `sshd -T | grep -i passwordauthentication` = `no`, `ufw status` shows `LIMIT` for 22
- [x] `grep -n 'prod-harden' docs/leitfaden/self-hosting.md` finds the step
- [x] `make check-repo` green (links, prose)

## Phase 5: jotti.rocks on Caddy

**Depends on**: none

### What to build

The rocks stack's reverse proxy uses the Caddy build from `reverse-proxy/Dockerfile` (it already includes caddy-ratelimit), run with a static `reverse-proxy/Caddyfile.rocks` instead of the rendered config. The file covers:
- the same hosts: `jotti.rocks`, `www`, `demo`, `auth`;
- redirects, including the `/docs/` map;
- the HSTS, CSP and other security headers, with Referrer-Policy aligned to the Caddy prod stack (G50);
- the rate limits: API 10 r/s with burst 20, and acme-dns `/register` at 1 per minute;
- automatic HTTP-01 certificates.

The certbot service, `docker-compose.initial-cert.yml`, `reverse-proxy/nginx.rocks.conf` and `reverse-proxy/nginx.initial-cert.conf` are deleted with every reference. `scripts/rocks-init.sh` replaces fixed sleeps with `wait_for_healthy()`. The auxiliary rocks services get healthchecks, including a DNS probe on the resolver. Postgres in `docker-compose.rocks.yml` gets `shm_size`/`stop_grace_period`.

The new `scripts/rocks-backup.sh <dest>`, run from the laptop, does four things:
- SSHes to the VPS;
- takes `sqlite3 .backup` of the acme-dns database;
- runs `PRAGMA integrity_check`;
- rsyncs the dated file to `<dest>` and removes the remote temp file.

`docs/jotti-rocks-infra.md` describes the Caddy setup, the backup script with a restore drill, and the Better Stack monitors.

### Acceptance criteria

- [x] `docker compose -f docker-compose.rocks.yml config -q` passes; `grep -rn 'certbot\|nginx.rocks\|initial-cert' --exclude-dir=node_modules --exclude=CHANGELOG.md .` finds nothing outside `frontend/`/`website/` nginx files
- [x] `caddy validate --config reverse-proxy/Caddyfile.rocks --adapter caddyfile` passes (inside the built image)
- [x] local rocks stack with Caddy's internal CA: `curl -skI` shows the four hosts' expected status, the redirects and the headers; the 2nd `/register` call within a minute returns 429
- [x] `scripts/rocks-backup.sh` against a local container with SSH (or the VPS, owner-run) yields a file whose `integrity_check` is `ok`
- [x] `shellcheck scripts/rocks-*.sh` clean; `make check-repo` green

## Phase 6: Container hardening and build contexts

**Depends on**: 3, 5

### What to build

`resolver/Dockerfile` runs as a dedicated non-root user whose binary gets `cap_net_bind_service`. The reverse proxy stays root, so existing `caddy-data` volumes need no migration. Every compose file that runs either service adds `cap_drop: [ALL]`, `cap_add: [NET_BIND_SERVICE]` and `security_opt: [no-new-privileges:true]`. The proxy also gets `read_only: true`, with tmpfs or volumes for every path it writes. `.dockerignore` files are added to the backend, resolver, reverse-proxy and database build contexts; `frontend/.dockerignore` gains `.env.*` and a final newline.

### Acceptance criteria

- [x] `docker run --rm --entrypoint id <resolver image>` prints a non-zero uid
- [x] `docker inspect` of the running proxy shows `CapDrop [ALL]`, `CapAdd [NET_BIND_SERVICE]` and `ReadonlyRootfs true`
- [x] the local stack starts from scratch and serves HTTPS on 443; the resolver answers on 53 (`dig @127.0.0.1`)
- [x] updating from the previous release keeps the existing certificates (scripted local run)
- [x] `docker build` of each context shows no `.env*` or `node_modules` in the context (`--progress=plain` context size before/after)

## Phase 7: Dev interface

**Depends on**: 2

### What to build

`make dev` is renamed `make up` (every reference in docs and scripts follows). `make help` renders targets by class with the handbook's `make-help.awk`, copied to `scripts/make-help.awk`, so the `prod-*` targets stand apart. `make lint` includes `check-shell`. `check-repo` runs every gate and reports all failures at the end. actionlint is pinned in `scripts/setup-dev-tools.sh` and runs in `make check` and the `repo-checks` job. `scripts/setup-dev-tools.sh` falls back to `npm install -g pnpm@<pinned>` without corepack and loses its `CLAUDE_CODE_REMOTE` branch. `.gitignore` is re-sectioned after the handbook template.

### Acceptance criteria

- [x] `make up` starts the dev stack; `grep -rn 'make dev\b' --exclude-dir=node_modules .` finds nothing
- [x] `make help` output groups targets under class headings
- [x] a deliberate failure in two gates makes `make check-repo` report both
- [x] a workflow syntax error makes `make check` fail via actionlint
- [x] `bash scripts/setup-dev-tools.sh` succeeds on this laptop (Node 26, no corepack)

## Phase 8: Scripts and repo gates

**Depends on**: 1, 3, 4, 5, 7

### What to build

Script hygiene across `scripts/`: the handbook header, `[[ ]]`, `${VAR:-default}` instead of hard-coded ports and container names, and `lib.sh` helpers instead of emoji `echo`. `read_env()` strips `\r` and one pair of surrounding quotes. `tracked_text_files()` drops the Copilot exclusion. The link gate gets its own file list, which covers `AGENTS.md` and `.claude/**`, and a per-line allowlist. Dev and test Postgres bind to `127.0.0.1` in `docker-compose.yml`, `scripts/test-integration.sh` and `scripts/test-tse-live.sh`; Vite keeps its LAN bind with a one-line comment on why. `.env.example` gains `PROXY_LE_STAGING` and `ACMEDNS_BASE_URL` as commented keys. Two new gates run in `check-repo`: every variable the compose files read appears in `.env.example`, and every doc appears in `docs/README.md`.

### Acceptance criteria

- [x] `shellcheck scripts/*.sh` clean; `grep -rn '\[ ' scripts/*.sh` finds no single-bracket tests
- [x] `read_env` returns `x` for `KEY="x"\r` (fixture test in `check-shell` or a bats-free shell test)
- [x] `docker compose up db` then `ss -ltn | grep 5432` shows `127.0.0.1:5432` only
- [x] removing a key from `.env.example`, or a row from `docs/README.md`, makes `make check-repo` fail
- [x] a broken link in `AGENTS.md` makes the link gate fail

## Phase 9: Frontend structure

**Depends on**: none

### What to build

`frontend/src/service/` is organised by feature (`service/<feature>/`), like `admin/`, as pure moves with import updates. Exported hooks get explicit return types. `npx shadcn@latest migrate cn` replaces clsx + tailwind-merge with the `cn` package; `lib/utils.ts` re-exports `cn`.

### Acceptance criteria

- [x] `ls frontend/src/service` shows feature folders and no `components/` type folder
- [x] `grep -n "clsx\|tailwind-merge" frontend/package.json frontend/src -r` finds nothing
- [x] `make check-frontend` green

## Phase 10: Frontend tests at the API boundary

**Depends on**: 9

### What to build

A test helper renders a page with a real QueryClient from `createQueryClient` and a fake `BackendClient` that answers per endpoint. Every page test that `vi.mock`s one of the project's own modules is rewritten to use it and to assert what the user sees. Third-party mocks stay where needed.

### Acceptance criteria

- [x] `grep -rn "vi.mock('@/\|vi.mock('\./\|vi.mock('\.\./" frontend/src` finds nothing
- [x] `make test-frontend` green; the test count does not drop

## Phase 11: e2e lint and frontend gate order

**Depends on**: 2, 7

### What to build

`e2e/package.json` pins `typescript ~6.0` and adds typed ESLint (`typescript-eslint` recommendedTypeChecked plus `eslint-plugin-playwright`), wired into `make lint` and the `e2e` CI job. The frontend gate runs `tsc -b` before the tests, locally and in `frontend-ci`, and uses frozen installs locally.

### Acceptance criteria

- [x] a floating `page.click()` without `await` in an e2e test makes `make lint` fail
- [x] `scripts/check-pins.sh` reports one TypeScript line across packages
- [x] a type error in `frontend/src` fails `make check-frontend` before any test runs

## Phase 12: Go tests without the unit tag

**Depends on**: 2, 7

### What to build

`//go:build unit` is removed from every test and fake; the fakes move to untagged test-helper files or packages. `scripts/check-build-tags.sh` requires a tag only on integration tests. `Makefile`, `ci.yml` and `fuzz.yml` drop `-tags=unit`, and golangci-lint runs once with `--build-tags=integration`. Tests use sentinel errors with `errors.Is` instead of string matching, stateful fakes instead of call counters, and `t.Errorf` for non-fatal assertions. goimports becomes a `go.mod` tool called via `go tool goimports`, replacing its pins in `ci.yml` and `setup-dev-tools.sh`. sqlc stays pinned outside `go.mod`: its cgo parser would enter the backend module graph.

### Acceptance criteria

- [x] `grep -rln 'go:build unit' backend` finds nothing
- [x] `cd backend && go test ./...` runs as many tests as `-tags=unit` ran before (count via `go test -json`)
- [x] `grep -rn 'tags=unit' Makefile .github` finds nothing
- [x] `grep -rn 'Error() ==\|err.Error(), "' backend --include=*_test.go` finds nothing
- [x] `make check-backend` green

## Phase 13: Comments at the cap

**Depends on**: 10, 12

### What to build

Every code comment longer than two sentences is cut to the invariant or the non-obvious why. Reasoning that is a spec (fiscal, TSE, DSFinV-K) moves into the matching `docs/` section, and the comment links it (e.g. `handbuch §3.13`).

### Acceptance criteria

- [x] a scan for `//` blocks of ≥ 4 lines in `backend/` and `frontend/src/` returns none, excluding generated `sqlc/dbgen/`
- [x] `make check` green

## Phase 14: Docs to the prose caps

**Depends on**: 1, 3, 4, 5, 8, 13

### What to build

`scripts/check-prose.sh` gains the handbook's prose caps (sentence ≤ 20 words, paragraph ≤ 3 lines). They exempt `docs/rechtsquellen/`, `LICENSE`, `TERMS.md`, `CLA.md`, `SERVICE.md` and `docs/verfahrensdokumentation.md`. All violations are fixed; leitfaden sentences are split, never shortened in content. Each fact that is stated several times gets one canonical home (non-goals, ELSTER, the TSE decision, the processType mapping); the other places link to it. Historic residue goes: "phasenweise", the ADR aside in `decisions.md`, leftover numbering in `produktbeschreibung.md`, with renumbering and anchors fixed. The open items of sections 4–5 of `docs/plans/plan-release-v1.0.0.md` move verbatim to `docs/backlog.md`, which is listed in `docs/README.md`.

### Acceptance criteria

- [x] `make check-repo` green with the caps on
- [x] a 25-word sentence added to `README.md` makes `scripts/check-prose.sh` fail
- [x] `grep -n '^## [45]\.' docs/plans/plan-release-v1.0.0.md` finds nothing; `docs/backlog.md` holds the open items

## Phase 15: Owner steps

**Depends on**: 5 (deploy), otherwise none

### What to build

These steps are outbound or run on external systems. The owner confirms each one before it runs.
- A GitHub ruleset on `main` that blocks force-push and deletion (`gh api`).
- Squash-only merges in the repo settings.
- The git user name and email in the claude.ai/code environment setup.
- Better Stack monitors for jotti.rocks, demo, auth.jotti.rocks and the resolver, with SSL-expiry alerts.
- Deploy the Caddy rocks stack to the VPS, then run `scripts/rocks-backup.sh` once to an external drive.

### Acceptance criteria

- [x] `gh api repos/nicograef/jotti/rulesets` lists the `main` ruleset with `non_fast_forward` and `deletion`
- [x] `gh api repos/nicograef/jotti --jq '.allow_merge_commit, .allow_rebase_merge'` prints `false false`
- [ ] `curl -sI https://auth.jotti.rocks/health` returns 200 with a Caddy-served certificate
- [ ] the Better Stack monitors show "up"

## Phase 16: CI on main

**Depends on**: all other phases landed on `main`

### What to build

Nothing new. The CI checks of every phase run once, after landing.

### Acceptance criteria

- [x] CI green on `main` after landing, including `upgrade-path`, and `release.yml` green in a `workflow_dispatch` dry run, including its restore step
