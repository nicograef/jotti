# Plan: Multi-Expert Review Fixes (2026-07-17)

> Source PRD: n/a — derived from the multi-expert review of 2026-07-17 (artifact report: https://claude.ai/code/artifact/99fedf36-91e4-4208-bbfd-658e18897629). Covers all 12 ranked recommendations plus the 10-minute quick wins.

## Goal

Land the actionable, freeze-safe findings from the review as a set of small, independently verifiable slices: restore the "CI == `make verify`" guarantee, fix the one real correctness defect (non-atomic produkt delete), remove a class of projection drift (`SaldoCents`), and take the confirmed code-quality/consistency wins — while deliberately declining the shared-kernel/abstraction campaigns the review flagged as traps.

Every change here is **freeze-safe**: no persisted DB schema, no event-JSON contract, no HTTP URL path, and no wire JSON key changes. Only Go/TS identifiers, config, tests, and internal structure move.

## Architectural decisions

Durable decisions that apply across all phases:

- **Freeze boundary (unchanged).** DB schema, event-JSON contracts (`event_json_contract_test.go`), HTTP URL paths, and response JSON keys are frozen. Renames touch Go/TS identifiers only. Route-handler method names may change; the registered URL path strings must not.
- **produkt naming convention = English verb + German noun.** Rename produkt nouns `Product→Produkt`, `Variant→Variante` across repo/application/http (+ tests), keeping the English verbs (`Create`/`Update`/`Activate`/`Deactivate`/`Get`). Result: `CreateProdukt`, `CreateVariante`, `ActivateVariante`, `GetAllProdukte`, `GetActiveProdukte`, `CreateProduktHandler`, etc. This matches `docs/language.md` rule 5 and produkt's already-correct `DeleteProdukt`/`DeleteVariante`. The `tisch` sibling (`TischErstellen…`) is **not** touched — out of scope for these recommendations.
- **Transaction boundary lives in the repository.** Multi-row writes are wrapped in `db.WithTx` inside a repository method, never orchestrated as separate autocommit calls from the application layer (the established Kasse/user/tse pattern).
- **Shared helpers go in the application layer, never in `domain/`.** `domain/` stays pure (no repo/`context.Context` I/O). The extracted position-enrichment helper is an application-layer package.
- **Deliberately not doing (recommendation #12).** Recorded here and annotated in code (Phase 10), not implemented:
  - Consolidating the per-context `ErrConflict`/guard duplication across the kasse bounded contexts into a shared kernel — the per-package `ErrConflict` sentinels are load-bearing for correct HTTP 409 mapping via `errors.Is`; a shared helper returning a foreign sentinel would silently turn 409 into 500.
  - Introducing React Context in `TSEEinrichtungWizard` (fights rule 7; `umgebung` comes from the befund, not root state).
  - Gutting the vendored shadcn `sidebar.tsx` (re-synced from upstream; unused exports are tree-shaken).
  - Generic `Select`/`TextField` wrappers over the divergent form fields.
  - Folding the docker-compose files into a shared base (`release.yml` must ship standalone; overrides can't subtract the differing keys).

## Inventory

Existing files and symbols the phases build on:

- `backend/.golangci.yml` — linter config; currently no `run.build-tags`.
- `.github/workflows/ci.yml` — `backend-golangci` job uses `golangci-lint-action@v9` with no build tags; `.github/workflows/` also holds `release.yml`, `security-scans.yml`.
- `Makefile` — `fuzz` target runs 4 fuzz targets 90s each (`kein CI-Dauerlauf`); `check-backend` lints with `--build-tags=integration`.
- `backend/api/admin.go`, `backend/api/service.go` — manual DI wiring: `xc := xHTTP.CommandHandler{}; xc.Command = …` two-step per handler (~27 sites).
- `backend/api/fiskal/dsfinvk/dsfinvk.go — storno(b bool)`; call sites in `mapper.go`.
- `backend/domain/event/event.go — New()`, `Validate()` — duplicated field checks.
- `backend/domain/kasse/tisch_session.go — ComputeNichtStorniertePositionen()` (stale comment), `SaldoCents` accumulation in the position-mutating arms; `GesamtZahlungenCents` accumulator.
- `backend/domain/kasse/storno_aufteilung.go — ComputeStornoAufteilung()` — `switch evt.Type` with no `default`.
- `backend/api/stammdaten/produkt/{application,http}/*.go`, `backend/repository/produkt_repo/{repo.go,variant.go,batch.go,mock.go}` — the produkt vertical; `DeleteProdukt()` loops `UpdateVariant` then `UpdateProduct` with no tx; `produkt_repo` holds `db *sql.DB` but has no atomic delete.
- `backend/db/db.go — WithTx(ctx, *sql.DB, func(*dbgen.Queries) error)`.
- `backend/api/kasse/tischgeschaeft/application/command.go — BestellungAufnehmen()` (inlines enrichment); `backend/api/kasse/direktverkauf/application/command.go — enrichPositionen()` (the extracted twin); both `http/command_handler.go` map the enrichment error sentinels.
- `scripts/*.sh` — ~10 scripts (`prod-init`, `prod-update`, `prod-restore`, `prod-backup`, `prod-backup-verify`, `prod-harden`, `rocks-init`, `reset-and-seed`, `ops-smoke`, `setup-dev-tools`) copy-paste the RED/GREEN/YELLOW color block; 6 also copy `read_env`. `ops-smoke.sh` logs to stderr (stdout is its return channel). `scripts/test-tse-live.sh` already uses a `# shellcheck source=` directive.
- Frontend hooks: `frontend/src/admin/{tse,settings,finanzamt,reporting}/hooks.ts` inline raw query-key literals; `frontend/src/admin/kasse/hooks.ts` already does `export const kasseBackend`; `frontend/src/admin/{products,users,tables}/{hooks.ts,*Page.tsx}` each `new XBackend(...)` twice; `frontend/src/admin/finanzamt/hooks.ts — useKassenidentitaet` returns the raw `useQuery` object; `frontend/src/service/components/TischAuswahlDrawer.tsx` wraps `useMutation` inside `useActionSubmit`.
- `frontend/src/service/components/table/{HistorieStornierungDrawer,HistorieUmbuchungDrawer,BestellungAbschluss}.tsx` — duplicated drawer header + total footer; `frontend/src/service/components/table/drawerUtils.ts` — existing helpers; `frontend/src/service/components/table/TischHistorie.tsx — zeilenmodell()`, `detailView()`, `HistorieDetail`.
- `frontend/src/service/product/Produkt.ts`, `frontend/src/service/table/Tisch.ts` — English schema consts (`ProductIdSchema`/`VariantIdSchema`); `frontend/src/admin/products/Produkt.ts`, `frontend/src/admin/tables/Tisch.ts` — German consts + input-form constraints.

## Resolved decisions

- **produkt naming** → English verb + German noun (`CreateProdukt`, `CreateVariante`); nouns only, verbs stay English; `tisch` untouched; JSON/URLs/DB frozen. (see Architectural decisions)
- **Recording the declines (#12)** → "Deliberately not doing" section in this plan + inline code comments at the intentionally-duplicated guards/sentinels (Phase 10). No ADR, no refactor.
- **Frontend schema drift (consistency-conventions-2)** → fix **naming only**: service consts → German (`ProduktIdSchema`/`VarianteIdSchema`/`TischIdSchema`). Do **not** unify the admin (input-form) and service (response-parser) schemas: admin's `.trim()`, `.max(99999)` and `saldoCents.min(0)` are legitimately input-form/invariant constraints and correctly absent from the wire-response parsers. Only add a field to a service schema if service genuinely consumes it and it is missing.
- **`SaldoCents`** → stop accumulating; derive from `UnbezahltePositionen` after each mutation. Keep the frozen DB column; keep `GesamtZahlungenCents` as a real accumulator (not derivable).
- **Enrichment helper home** → a new application-layer package shared by tischgeschaeft and direktverkauf, **not** `domain/kasse`.

## Open questions / Risks

- **Phase 5 (enrichment helper)** is the highest-risk backend change: it touches two hot command paths (`BestellungAufnehmen`, `DirektverkaufTaetigen`), unifies their `…PositionInput` types, and re-homes the `ErrProduktNotFound`/`ErrVarianteNichtAktiv` sentinels plus both HTTP error mappings. The active-status defense-in-depth check must stay byte-identical. Verify with the existing command tests of both contexts before/after.
- **Phase 9 (produkt rename)** is wide (repo/application/http + all produkt tests + `admin.go` route registrations). Mechanical, but confirm no `Product`/`Variant` Go identifier remains and that every registered URL path string is unchanged.
- **Phase 4 (`SaldoCents`)** feeds the saldo-never-negative close gate; the derived value must be byte-identical to today's accumulated value on every path. Pin the equality invariant and run the full integration + fuzz suite.

---

## Phase 1: Restore the CI == `make verify` invariant + activate fuzzing

**Covers**: infra-build-config-2 (rec 1), testing-4 (rec 10). Config only — do first so every later phase is linted identically in CI and locally.

### Context

- `backend/.golangci.yml` — no `run.build-tags`, so the 27 `//go:build integration` files are unlinted in CI while `make check-backend` lints them with `--build-tags=integration`.
- `.github/workflows/ci.yml — backend-golangci` — `golangci-lint-action@v9`, no build tags.
- `Makefile — fuzz` — runs 4 fuzz targets but is invoked by no workflow, so CI only replays seed corpora.

### What to build

Add `run.build-tags: [integration]` to `backend/.golangci.yml` so both the CI action and the Makefile lint the same file set (single source, no CI job edit needed). Add a new scheduled workflow `.github/workflows/fuzz.yml` (weekly cron + `workflow_dispatch`) that runs `make fuzz` and, on failure, uploads any generated `backend/**/testdata/fuzz/**` crashers as an artifact.

### Acceptance criteria

- [ ] `backend/.golangci.yml` sets `run.build-tags: [integration]`; `golangci-lint run` in `./backend` (as CI invokes it) now lints `//go:build integration` files.
- [ ] A deliberately-introduced lint violation in an integration-tagged test file fails both `make check` and the CI `backend-golangci` job (verified, then reverted).
- [ ] `.github/workflows/fuzz.yml` runs `make fuzz` on a weekly `schedule` and on `workflow_dispatch`, and uploads crashers on failure; `actionlint`/CI parses it without error.
- [ ] `make check` stays green.

---

## Phase 2: Mechanical backend quick-wins

**Covers**: arch-backend-5, go-code-quality-6, simplification-razor-6, event-sourcing-domain-5 (rec 8), event-sourcing-domain-2 (rec 11 backend half). Zero behavior change.

### Context

- `backend/api/admin.go`, `backend/api/service.go` — ~27 two-step empty-struct-then-assign handler constructions.
- `backend/api/fiskal/dsfinvk/dsfinvk.go — storno(b bool)` — always called with the literal `false`.
- `backend/domain/event/event.go — New()`, `Validate()` — five duplicated field checks (identical error strings).
- `backend/domain/kasse/tisch_session.go — ComputeNichtStorniertePositionen()` — comment claims it backs storno validation; its only non-test caller is the seed engine.
- `backend/domain/kasse/storno_aufteilung.go — ComputeStornoAufteilung()` — `switch evt.Type` has no `default` arm, unlike its three sibling replays.

### What to build

Collapse the admin/service handler wiring into composite literals (`xc := xHTTP.CommandHandler{Command: xApp.Command{…}}`). Replace `storno(bool)` and its call sites with a named constant `stornoNein = "0"`, dropping the dead parameter. Extract the shared `event.New`/`event.Validate` field checks into one unexported helper both call (`New` skips only the Version/Time checks it hasn't assigned yet). Correct the `ComputeNichtStorniertePositionen` comment to state it is a seed helper. Add `default: return StornoAufteilung{}, false` to `ComputeStornoAufteilung` (matching the function's existing bool-failure convention — defense-in-depth, not a reachable bug).

### Acceptance criteria

- [ ] admin.go/service.go handlers built as single composite literals; route table and behavior unchanged.
- [ ] `storno(bool)` gone; DSFinV-K CSV export is byte-identical (existing golden/mapper tests pass unchanged).
- [ ] `event.New` and `event.Validate` share one field-check helper; event validation behavior and error strings unchanged.
- [ ] `ComputeNichtStorniertePositionen` comment reflects reality; no caller change required.
- [ ] `ComputeStornoAufteilung` has a `default` arm; existing storno tests pass; an unknown event type now yields a safe `false` refusal.
- [ ] `make check` green.

---

## Phase 3: Atomic produkt delete (the one correctness defect)

**Covers**: arch-backend-2 (rec 2).

### Context

- `backend/api/stammdaten/produkt/application/command.go — DeleteProdukt()` — loops variant soft-deletes (`UpdateVariant`) then `UpdateProduct`, each its own autocommit write; a mid-loop failure leaves variants deleted with the product still active.
- `backend/repository/produkt_repo/` — holds `db *sql.DB` but has no atomic delete method; `backend/db/db.go — WithTx()` is the established transaction wrapper.

### What to build

Add a `produkt_repo` method that soft-deletes a product and all its variants inside one `db.WithTx` (status → deleted for the product and each variant in a single transaction). Change `DeleteProdukt` to call it instead of orchestrating the loop. Keep the soft-delete semantics (`status = 'deleted'`) exactly as today.

### Acceptance criteria

- [ ] `produkt_repo` exposes one atomic soft-delete-with-variants method wrapping `db.WithTx`; the application layer no longer loops row writes for delete.
- [ ] An integration test injecting a mid-transaction failure proves atomicity: on failure the product and all its variants remain in their pre-delete state (no partial delete).
- [ ] Happy-path delete still soft-deletes product + variants; existing produkt tests pass.
- [ ] `make verify` (incl. integration) green.

---

## Phase 4: Derive `SaldoCents` from positions

**Covers**: event-sourcing-domain-1 (rec 3).

### Context

- `backend/domain/kasse/tisch_session.go` — `SaldoCents` is accumulated (`+=`/`-=`) in the five position-mutating arms independently of `UnbezahltePositionen`, yet always equals `Σ(EinzelpreisCents × Menge)` over those positions; it feeds the saldo-never-negative close gate. `GesamtZahlungenCents` is a genuine accumulator (not derivable).
- `backend/domain/kasse/replay_fuzz_test.go`, `backend/domain/kasse/tisch_session_test.go` — where the invariant will be pinned.

### What to build

Stop accumulating `SaldoCents`. After each mutation of `UnbezahltePositionen`, set `SaldoCents` once from the positions (`Σ EinzelpreisCents × Menge`). Keep the frozen DB column and keep `GesamtZahlungenCents` accumulating. Add a test asserting `SaldoCents == Σ(EinzelpreisCents × Menge)` over `UnbezahltePositionen` after every event type, and extend the fuzz replay to assert the same invariant.

### Acceptance criteria

- [ ] `SaldoCents` is derived from `UnbezahltePositionen` after each position mutation; no arm accumulates it independently.
- [ ] `GesamtZahlungenCents` still accumulates unchanged.
- [ ] A unit test pins `SaldoCents == Σ(EinzelpreisCents × Menge)`; the fuzz replay asserts it too.
- [ ] Projection values are byte-identical to before on all existing tests; the close gate behaves identically.
- [ ] `make verify` (incl. integration + fuzz seed corpus) green.

---

## Phase 5: Extract the shared position-enrichment helper

**Covers**: go-code-quality-1 (rec 7). Highest-risk backend slice — see Risks.

### Context

- `backend/api/kasse/tischgeschaeft/application/command.go — BestellungAufnehmen()` inlines ~55 lines of position enrichment (dedup variant/product IDs → batch `GetVariantsByIDs`/`GetProductsByIDs` → active check → build `kasse.Position`).
- `backend/api/kasse/direktverkauf/application/command.go — enrichPositionen()` is the byte-for-byte twin (one comment word differs), over `VerkaufPositionInput`; `BestellPositionInput` is the tischgeschaeft counterpart.
- Both `…/http/command_handler.go` map the enrichment sentinels (`ErrProduktNotFound`/`ErrVarianteNichtAktiv`) to HTTP codes.

### What to build

Create a new application-layer package holding: one `PositionInput` type (`{ProduktID, VarianteID, Menge}`), a narrow `produktRepo` interface (the two batch reads), the `EnrichPositionen(ctx, repo, inputs) ([]kasse.Position, error)` function, and the shared `ErrProduktNotFound`/`ErrVarianteNichtAktiv` sentinels. Point both `BestellungAufnehmen` and `DirektverkaufTaetigen` at it; unify their `…PositionInput` types onto the shared one; update both HTTP handlers' error mappings to the shared sentinels; update the affected command/handler tests. Keep the active-status check byte-identical. Do **not** place any of this in `domain/kasse`.

### Acceptance criteria

- [ ] A single enrichment implementation exists in an application-layer package; neither command inlines or privately duplicates it.
- [ ] Both contexts produce identical enrichment behavior, including the active-status refusal and error-to-HTTP mapping.
- [ ] `domain/kasse` is untouched by this change (layering preserved).
- [ ] Existing tischgeschaeft and direktverkauf command + handler tests pass (adjusted only for the shared types/sentinels).
- [ ] `make verify` green.

---

## Phase 6: Shared shell library for ops scripts

**Covers**: infra-build-config-5 (rec 6).

### Context

- ~10 `scripts/*.sh` copy-paste the RED/GREEN/YELLOW/NC color block and `info()/warn()/error()/fatal()`; 6 also copy `read_env` (.env parsing).
- `scripts/ops-smoke.sh` logs to stderr (stdout is its return channel); `scripts/test-tse-live.sh` already uses a `# shellcheck source=` directive; CI runs `shellcheck scripts/*.sh`.

### What to build

Add `scripts/lib.sh` with the color vars, log helpers (writing to **stderr**), and `read_env`. Source it robustly from each consuming script via `SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"; . "$SCRIPT_DIR/lib.sh"`, each with a `# shellcheck source=scripts/lib.sh` directive above the source line. Migrate the ~10 scripts to source it and drop their local copies. Keep prod-only helpers (e.g. `parse_semver`, `validate_secret`) local to their scripts — do not hoist them.

### Acceptance criteria

- [ ] `scripts/lib.sh` exists (color vars + stderr log helpers incl. `fatal` + `read_env`); consuming scripts source it via `BASH_SOURCE` with a `# shellcheck source=` directive.
- [ ] The duplicated color/`read_env` blocks are removed from the migrated scripts; prod-only helpers stay local.
- [ ] `shellcheck scripts/*.sh` (the CI gate) passes with no new SC1090/SC1091.
- [ ] `ops-smoke.sh` still emits logs to stderr and its stdout return-value contract is unchanged.

---

## Phase 7: Frontend hooks consistency + mutation cleanup

**Covers**: frontend-architecture-2 (rec 4), frontend-architecture-4 (rec 8), frontend-architecture-5 (quick win), frontend-architecture-1 (rec 11 frontend half).

### Context

- `frontend/src/admin/{tse,settings,finanzamt,reporting}/hooks.ts` — raw query-key literals inlined in both `useQuery` and `invalidateQueries`; `products/users/kasse` already export co-located key consts.
- `frontend/src/admin/{products,users,tables}/{hooks.ts,*Page.tsx}` — each `new XBackend(BackendSingleton)` in both hooks and page; `admin/kasse/hooks.ts` already does `export const kasseBackend`.
- `frontend/src/admin/finanzamt/hooks.ts — useKassenidentitaet` returns the raw `useQuery` object while siblings return a domain-named wrapper.
- `frontend/src/service/components/TischAuswahlDrawer.tsx` — a `useActionSubmit` `run()` wraps a `useMutation.mutateAsync` for one favorite toggle; the two invalidation key consts already exist.

### What to build

Co-locate every query key in the touched hooks as an exported const, referenced by both the `useQuery` and every `invalidateQueries` call — no inline string keys. Instantiate each domain's `Backend` exactly once (export it from the hooks module, per `kasseBackend`) and import that single instance in the page instead of re-`new`-ing it; drop the prop-drilled second instance. Make `useKassenidentitaet` return the wrapped `{ kassenidentitaet, isPending, … }` shape like its siblings and update its single consumer. Unstack the `TischAuswahlDrawer` mutation: call the favorite toggle directly inside `useActionSubmit`'s `run()`, then `invalidateQueries` on the two existing key consts; remove the `useMutation`.

### Acceptance criteria

- [ ] No inline query-key literals remain in `tse/settings/finanzamt/reporting` hooks; each key is one exported const used by `useQuery` and `invalidateQueries`.
- [ ] Each of products/users/tables exports one `Backend` instance; pages import it; no page calls `new XBackend` a second time.
- [ ] `useKassenidentitaet` returns the wrapped shape; `EinrichtungSection` consumer updated.
- [ ] `TischAuswahlDrawer` favorite toggle runs through `useActionSubmit` only (no nested `useMutation`); refresh still works via `invalidateQueries`.
- [ ] Frontend lint + tests green; user-visible behavior unchanged.

---

## Phase 8: Frontend service-component dedup

**Covers**: frontend-components-4, frontend-components-5 (rec 9).

### Context

- `frontend/src/service/components/table/{HistorieStornierungDrawer,HistorieUmbuchungDrawer}.tsx` — byte-identical drawer title block and `border-t-2 pt-2 font-bold` total footer (the footer is also copied into `BestellungAbschluss.tsx`); `drawerUtils.ts` already holds related helpers.
- `frontend/src/service/components/table/TischHistorie.tsx — zeilenmodell()` and `detailView()` each switch over the same four-case `art` union and re-derive `date`; `HistorieDetail` calls `zeilenmodell(detail)` a second time.

### What to build

Extract a `QuelleDrawerHeader({ quelle })` (the `quelleTitel · relativeTime · userName` block) and a `GesamtZeile({ label, betrag })` (the total footer), and reuse them in both Historie drawers and `BestellungAbschluss`. Fold `TischHistorie`'s `detailView` positionen/total into the single `zeilenmodell` exhaustive switch (or vice-versa) so each entry is mapped once, and pass the resulting model down to `HistorieDetail` instead of recomputing it.

### Acceptance criteria

- [ ] `QuelleDrawerHeader` and `GesamtZeile` exist and are used by both Historie drawers and `BestellungAbschluss`; the duplicated header/footer JSX is gone.
- [ ] `TischHistorie` maps each entry through one exhaustive switch; `HistorieDetail` receives the model rather than re-invoking `zeilenmodell`.
- [ ] Rendered output (row model, detail view, totals) is unchanged; existing `TischHistorie`/drawer tests pass.
- [ ] Frontend lint + tests green.

---

## Phase 9: produkt naming consistency (backend + frontend)

**Covers**: consistency-conventions-1, consistency-conventions-2 (rec 5). Widest churn — do last. See Risks.

### Context

- `backend/api/stammdaten/produkt/{application,http}/*.go` + `backend/repository/produkt_repo/*.go` (+ their tests) + `backend/api/admin.go` route registrations — English `Product`/`Variant` in `CreateProduct`, `CreateVariant`, `ActivateVariant`, `GetAllProducts`, `GetActiveProducts`, `CreateProductHandler`, etc.; `DeleteProdukt`/`DeleteVariante` already German.
- `frontend/src/service/product/Produkt.ts`, `frontend/src/service/table/Tisch.ts` — English schema consts (`ProductIdSchema`/`VariantIdSchema`); `frontend/src/admin/products/Produkt.ts`, `frontend/src/admin/tables/Tisch.ts` — German consts + input-form constraints.

### What to build

Rename the produkt Go identifiers to English-verb + German-noun form (`Product→Produkt`, `Variant→Variante`) across repo, application, http, and their tests; keep the English verbs; update the `admin.go` handler references. Registered URL path strings, request/response JSON keys, sqlc/dbgen, and DB stay untouched. On the frontend, rename the service schema consts to German (`ProduktIdSchema`/`VarianteIdSchema`/`TischIdSchema`) to match language.md and the admin side — **naming only**; do not import admin's input-form constraints (`.trim()`, `.max(99999)`, `saldoCents.min(0)`) into the service response parsers, and add a field to a service schema only if service genuinely consumes it and it is missing.

### Acceptance criteria

- [ ] No `Product`/`Variant` Go identifier remains in the produkt vertical; produkt uses `Create/Update/Activate/Deactivate/Get + Produkt/Variante` uniformly (with the existing `DeleteProdukt`/`DeleteVariante`).
- [ ] Every registered URL path string and response JSON key is unchanged; `event_json_contract_test.go` and HTTP handler tests pass without contract edits.
- [ ] Service frontend schema consts are German; admin (input-form) vs service (response-parser) constraint differences are left as-is per the resolved decision.
- [ ] `tisch` backend identifiers are untouched.
- [ ] `make verify` + frontend lint/tests green.

---

## Phase 10: Annotate the intentional duplication (record the declines)

**Covers**: recommendation #12 (go-code-quality-2/3/4/5, arch-backend-6, frontend-components-1/2/3, infra-build-config-1, simplification-razor-2). No refactor — documentation only.

### Context

- `backend/api/kasse/{tischgeschaeft,direktverkauf,kassenfuehrung,druck/beleg}/application/*.go` — the per-context `ErrConflict`/`ErrKasseNichtGeoeffnet` sentinels, `getOffeneKassensitzungOderFehler` guard, and OCC/idempotency mapping duplicated on purpose (one copy in `druck/beleg` is already commented as intentional).

### What to build

Add a short inline comment at the mirrored guards/sentinels in the kasse command packages noting the duplication is deliberate and load-bearing (per-context `ErrConflict` drives correct 409 mapping via `errors.Is`; a shared kernel would couple separate bounded contexts) — referencing the 2026-07-17 review. Ensure the "Deliberately not doing" list in this plan's Architectural decisions stays the record for the non-code declines (React Context, shadcn sidebar, generic field wrappers, compose base file). No behavior change.

### Acceptance criteria

- [ ] Each mirrored kasse guard/sentinel carries a one-line comment marking it as intentionally per-context (citing the review), so future readers/agents don't re-raise the shared-kernel refactor.
- [ ] No code behavior changes; `make check` green.
- [ ] The declined non-code campaigns remain recorded in this plan's "Deliberately not doing" section.
