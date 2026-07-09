# Plan: Audit-Fixes v0.15.0

> Source PRD: docs/plans/audit-v0.15.0.md (Befundliste des Release-Audits v0.14.0..HEAD)

## Goal

Die im Audit `docs/plans/audit-v0.15.0.md` bestätigten Befunde umsetzen: den
einzigen CRITICAL (C1, exponierter DB-Reset), die Correctness-/Test-/Ops-Major
(M1, M3–M6), die 40 Minor-Cleanups (N1–N44, M2 fällt ohne Code weg) sowie eine
abschließende Investigations-Phase für die zehn Coverage-Lücken (G1–G10).
Reihenfolge: Fixes zuerst, Investigation zuletzt.

## Architectural decisions

Durable decisions, die über alle Phasen gelten:

- **Zwei getrennte Flags statt einem.** `JOTTI_ALLOW_SEED` gated künftig
  **ausschließlich** das CLI-Subkommando `jotti seed` (Shell-Zugriff nötig,
  harmlos) und bleibt auf `docker-compose.rocks.yml`, `docker-compose.local.yml`
  und `docker-compose.yml` gesetzt — der Demo-Reset `scripts/reset-and-seed.sh`
  läuft über dieses Subkommando (`docker compose exec backend jotti seed`) und
  bleibt intakt. Ein **neues** Flag `JOTTI_ENABLE_TEST_API` gated
  **ausschließlich** den HTTP-Endpoint `POST /test/reset-and-seed` und wird nur
  in `docker-compose.e2e.yml` gesetzt.
- **Test-API über `config.Config`, nicht `os.Getenv`.** Das neue Flag wird in
  `backend/config/config.go — Load()` als Feld `Config.EnableTestApi` (true ⇔
  `JOTTI_ENABLE_TEST_API == "1"`) geladen; `backend/app/app.go — SetupRoutes()`
  liest `cfg.EnableTestApi` statt `seed.AllowedByEnv(os.Getenv)` (behebt den
  IoC-Bruch aus dem C1-Cluster).
- **Test-Reset-Area ist rate-limited.** Die `testResetArea`
  (`backend/app/routes.go`) bekommt `RateLimited: true` wie auth/relay.
- **E2E-Proxy bindet an `127.0.0.1`.** Das Port-Mapping der Reverse-Proxy in
  `docker-compose.e2e.yml` wird von `0.0.0.0` (implizit) auf `127.0.0.1`
  eingeschränkt (behebt N31/G1).
- **M1 bricht im Konfliktfall hart ab.** Der Kassenabschluss-Wiederanlauf
  verwendet den alten Ist-Bestand nur weiter, wenn nach dem protokollierten
  Kassensturz keine Buchungs-Events im Stream liegen; andernfalls Abbruch mit
  klarem Domänenfehler statt stiller Wiederverwendung.
- **ops-smoke konfiguriert den Betreiber vor der Kassensitzung** und der
  POST-Helper trägt keinen `{}`-Default mehr (jeder Aufrufer übergibt den Body
  explizit).
- **E2E-Saldo-Assertions sind aufs Saldo-Element gescopt** (inline, kein
  Helper), nicht ungescopt per `getByText().first()`.
- **M2 erhält keinen Code.** Die produktive Erstinstallation (2026-07-07) hat
  TSE nie eingerichtet; es gibt keine leere Seriennummer in Produktion, und der
  Code-Bug ist in v0.15.0 bereits behoben. Statt eines Backfills nur eine
  Runbook-/Release-Notes-Zeile (Teil von Phase 5, Doku-Sweep).

## Inventory

- `backend/app/app.go — SetupRoutes()` — hängt `testResetArea` an, wenn
  `seed.AllowedByEnv(os.Getenv)`; hier auf `cfg.EnableTestApi` umstellen.
- `backend/app/routes.go — testResetArea(), mountArea(), Areas()` — deklarative
  Area-Struktur; `RateLimited` fehlt bei der Test-Area.
- `backend/seed/guard.go — AllowSeedEnv, AllowedByEnv()` — bleibt für das CLI-
  Subkommando unverändert.
- `backend/config/config.go — Config, Load()` — neues Feld `EnableTestApi`.
- `backend/api/test/handler.go — NewHandler(), ResetAndSeedHandler()` — der
  Handler selbst bleibt; nur seine Registrierung ändert sich.
- `backend/app/app_test.go` — Test, dass die Route nur bei gesetztem Flag
  registriert wird (pinnt aktuell `JOTTI_ALLOW_SEED`); auf das neue Flag ziehen.
- `backend/app/matrix_integration_test.go` — Kommentarverweis auf
  `JOTTI_ALLOW_SEED` bzgl. der Test-Route; angleichen.
- `backend/seed/writer.go — writeSeed()` — `SET LOCAL session_replication_role
  = replica` (N11: nach `SeedTruncateAll` zurücksetzen).
- `backend/api/kasse/kassenfuehrung/application/command.go — KasseAbschliessen(),
  findeVorhandenenKassensturz()` — M1-Fixort; `findeVorhandenenKassensturz`
  liest `ReadEventsBySubject` in Stream-Reihenfolge.
- `scripts/ops-smoke.sh — http_post_status(), step_sale_receipt_export(),
  restore_jotti_version()` — M3, M4, N29, N30.
- `backend/api/stammdaten/betreiber/http/command_handler.go` —
  `update-betreiber`-Request (`vereinsname`, `strasse`, `plz`, `ort` Pflicht;
  `steuernummer`, `ustId` optional) für den M3-Dummy.
- `e2e/support/servicekraft.ts — settleAlleOffenenTische(), bestellePosition(),
  oeffneTisch()` — M6, N39.
- `e2e/support/seed.ts — resetAndSeed()` — ruft `POST /api/test/reset-and-seed`;
  Endpoint-Pfad bleibt, nur das Compose-Flag ändert sich.
- `docker-compose.e2e.yml` — `JOTTI_ALLOW_SEED` → `JOTTI_ENABLE_TEST_API`,
  Port-Bind auf `127.0.0.1`.
- `docker-compose.rocks.yml`, `docker-compose.local.yml`, `docker-compose.yml` —
  `JOTTI_ALLOW_SEED` bleibt (CLI-Reset), Kommentar präzisieren.
- `.github/workflows/ci.yml — changes-Job (e2e-Pfadfilter), e2e-Job` — N26, N28.
- `e2e/tsconfig.json` — `include` (N44), gekoppelt an das Typecheck-Gate N28.

## Resolved decisions

- **C1** → Belange trennen (neues `JOTTI_ENABLE_TEST_API` nur für den HTTP-
  Endpoint, nur in `docker-compose.e2e.yml`; `JOTTI_ALLOW_SEED` bleibt für das
  CLI-Subkommando auf rocks/local/base). Zusätzlich: Rate-Limit auf der Test-
  Area, IoC-Fix über `config.Config`, e2e-Bind auf `127.0.0.1`.
- **M1** → bei Konflikt (Buchungen nach dem Sturz) mit klarem Fehler abbrechen.
- **M2** → kein Code; nur eine Runbook-/Release-Notes-Zeile.
- **M3** → `update-betreiber`-Schritt in `step_sale_receipt_export` **vor**
  `kassensitzung-eroeffnen`.
- **M4** → den `${2:-{}}`-Default **entfernen**; jeder Aufrufer übergibt den
  Body explizit (auch leere Fälle als `'{}'`).
- **M5** → jede Saldo-Assertion **inline** aufs Header-Saldo-Element scopen
  (kein geteilter Helper).
- **M6** → die fragilen `isVisible()`/`count()`-Zweige durch expect-basierte
  Auto-Wait-Assertions ersetzen, wo der Zustand deterministisch ist.
- **N13** → Antwort behalten und die Zahl im Frontend-Toast anzeigen
  („n Aufträge verworfen").
- **N14** → interne Bezeichner in `frontend/src/components/ui/tabs.tsx`
  anglisieren, öffentlicher Export bleibt.
- **N26** → toten e2e-Pfadfilter entfernen, „läuft immer auf push/PR" belassen,
  Kommentar anglisieren.
- **N28** → Typecheck-Gate für `e2e/` ergänzen (samt `include`-Fix N44).
- **Minor-Bulk** (N1–N12, N15–N25, N27, N29, N30, N32–N44) → wie im Audit
  vorgeschlagen; „unverifiziert" markierte Befunde bei der Umsetzung je Befund
  am Code bestätigen, bevor sie angewendet werden.
- **G1–G10** → eigene Investigations-Phase am Ende; nur bestätigte Lücken werden
  zu Fixes (ggf. Folgeplan).

## Open questions / Risks

- **M1-Mechanik.** Die Konflikterkennung (Buchungs-Events nach dem Sturz) muss
  die abschluss-eigenen Folge-Events (Differenzbuchung, Tagesabschluss) sauber
  vom Buchungsverkehr abgrenzen, sonst schlägt ein legitimer Wiederanlauf nach
  Teilfehler in Schritt 2/3 fälschlich fehl. Genau dafür die Regressionstests in
  Phase 2.
- **G-Phase kann Folgearbeit erzeugen.** Bestätigt sich in G2 (PROXY_HTTP_ONLY),
  G3 (`.claude`-Hook), G8 (DSFinV-K-Spec-Konformität) oder G9 (TSE-Live-Test
  trifft echte API) ein echtes Problem, wird der Fix als eigener Plan geschnitten
  — Phase 5 liefert Befunde, nicht zwingend Fixes.

---

## Phase 1: C1 — Test-Reset-Endpoint absichern

**User stories**: n/a (Security-Fix)

### Context

- `backend/config/config.go — Config, Load()` — neues Feld `EnableTestApi`.
- `backend/app/app.go — SetupRoutes()` — Registrierung auf `cfg.EnableTestApi`
  umstellen.
- `backend/app/routes.go — testResetArea()` — `RateLimited: true` ergänzen.
- `backend/seed/guard.go — AllowSeedEnv` — bleibt, nur noch CLI.
- `backend/app/app_test.go`, `backend/app/matrix_integration_test.go` — Tests
  und Kommentar aufs neue Flag ziehen.
- `docker-compose.e2e.yml`, `docker-compose.rocks.yml`,
  `docker-compose.local.yml`, `docker-compose.yml` — Flags und Bind.

### What to build

Der HTTP-Endpoint `POST /test/reset-and-seed` wird von einem eigenen Flag
`JOTTI_ENABLE_TEST_API` abhängig, das nur `docker-compose.e2e.yml` setzt und das
über `config.Config` (nicht direkt `os.Getenv`) gelesen wird. Der Endpoint wird
zusätzlich rate-limitet. `JOTTI_ALLOW_SEED` behält seine ursprüngliche, sichere
Aufgabe (CLI-Subkommando `jotti seed`) und bleibt auf rocks/local/base — der
Demo-Reset über `scripts/reset-and-seed.sh` funktioniert unverändert. Das
E2E-Compose bindet die Reverse-Proxy nur noch an `127.0.0.1`. Nach dem Fix ist
der DB-Reset über HTTP ausschließlich in der E2E-Umgebung erreichbar, nie mehr
aus dem Internet (jotti.rocks) oder dem Vereins-WLAN (local).

### Acceptance criteria

- [ ] `Config.EnableTestApi` existiert und wird in `Load()` aus
  `JOTTI_ENABLE_TEST_API == "1"` gesetzt.
- [ ] `SetupRoutes` registriert die Test-Reset-Area anhand `cfg.EnableTestApi`;
  kein direkter `os.Getenv`-Aufruf mehr für diese Entscheidung.
- [ ] `testResetArea` ist mit `RateLimited: true` deklariert.
- [ ] `docker-compose.e2e.yml` setzt `JOTTI_ENABLE_TEST_API: "1"`, nicht mehr
  `JOTTI_ALLOW_SEED`; das Reverse-Proxy-Port-Mapping bindet an `127.0.0.1`.
- [ ] `docker-compose.rocks.yml`, `docker-compose.local.yml`,
  `docker-compose.yml` behalten `JOTTI_ALLOW_SEED` (CLI); die Kommentare nennen
  nur noch das Subkommando, nicht den HTTP-Endpoint.
- [ ] `backend/app/app_test.go` prüft Registrierung/Nicht-Registrierung über
  `JOTTI_ENABLE_TEST_API`; der Kommentar in `matrix_integration_test.go` nennt
  das korrekte Flag.
- [ ] `scripts/reset-and-seed.sh` funktioniert unverändert (CLI-Pfad, kein
  Bezug auf das neue Flag).
- [ ] `make verify` grün; E2E-Suite grün (Reset läuft weiter über den HTTP-
  Endpoint, jetzt hinter dem neuen Flag).

---

## Phase 2: M1 — Kassenabschluss-Wiederanlauf gegen Zwischenbuchungen absichern

**User stories**: n/a (Correctness-Fix)

### Context

- `backend/api/kasse/kassenfuehrung/application/command.go — KasseAbschliessen(),
  findeVorhandenenKassensturz()` — der alte Ist-Bestand aus dem vorhandenen
  Sturz wird stumm wiederverwendet, während der Soll frisch berechnet wird.
- `backend/api/kasse/kassenfuehrung/application/kassenabschluss_wiederanlauf_integration_test.go`
  und `command_test.go` — bestehende Wiederanlauf-Tests (nur sofortiger Retry
  ohne Zwischenbuchungen).

### What to build

Der Wiederanlauf erkennt, ob nach dem protokollierten Kassensturz weitere
Buchungs-Events im Stream liegen (also solche, die nicht zum Abschluss selbst
gehören — nicht Kassensturz, Differenzbuchung, Tagesabschluss). Liegt eine
solche Zwischenbuchung vor, bricht der Abschluss mit einem klaren Domänenfehler
ab, statt den veralteten Ist-Bestand zu übernehmen und legitime Umsätze als
Soll-Ist-Differenz zu verbuchen. Ohne Zwischenbuchung bleibt das heutige,
bewusst dokumentierte Verhalten (alter Ist zählt) erhalten.

### Acceptance criteria

- [ ] Bei vorhandenem Kassensturz **und** mindestens einem Buchungs-Event danach
  liefert `KasseAbschliessen` einen klaren, benannten Fehler und schreibt kein
  Abschluss-Event.
- [ ] Bei vorhandenem Kassensturz **ohne** Zwischenbuchung bleibt der Wiederanlauf
  wie bisher (Schritt 1 übersprungen, alter Ist maßgeblich, Abschluss erfolgreich).
- [ ] Neuer Regressionstest: Teilfehler nach Schritt 1 → `defer` setzt auf
  `offen` → Zwischenbuchung → Retry schlägt mit dem neuen Fehler fehl.
- [ ] Bestehende Wiederanlauf-Tests (sofortiger Retry, Teilfehler nach Schritt 2
  ohne Zwischenbuchung) bleiben grün.
- [ ] `make verify` grün.

---

## Phase 3: M3–M6 — Ops-Smoke und E2E-Robustheit

**User stories**: n/a (Test-/Ops-Fixes)

Enthält die Major M3, M4, M5, M6 plus die Minor, die dieselben Dateien berühren:
ops-smoke (N29, N30) und E2E-Specs/Support (N39, N40, N41, N42, N43).

### Context

- `scripts/ops-smoke.sh — step_sale_receipt_export(), http_post_status(),
  restore_jotti_version()` — M3, M4, N29, N30.
- `backend/api/stammdaten/betreiber/http/command_handler.go` — Pflichtfelder des
  M3-Dummy-Betreibers.
- `e2e/support/servicekraft.ts — settleAlleOffenenTische()` — M6, N39.
- `e2e/tests/tischservice-teilzahlung.mobile.spec.ts`,
  `e2e/tests/bestellen-kassieren.spec.ts`,
  `e2e/tests/kassenabschluss.mobile.spec.ts` — M5 (Saldo-Assertions).
- `e2e/tests/kassieren-fehlerpfade.spec.ts` — N39.
- `e2e/tests/direktverkauf-storno.mobile.spec.ts` — N40.
- `e2e/tests/admin-dsfinvk-export.spec.ts` — N42; `e2e/playwright.config.ts` — N43.

### What to build

- **M3**: In `step_sale_receipt_export` vor `kassensitzung-eroeffnen` ein
  `POST /api/admin/update-betreiber` mit minimalen Dummy-Stammdaten
  (`vereinsname`, `strasse`, `plz`, `ort`); der Schritt schlägt bei `!= 200`
  fehl wie die übrigen.
- **M4**: Den `${2:-{}}`-Default in `http_post_status` entfernen; jeder Aufrufer
  übergibt den Body explizit (auch leere Fälle als `'{}'`).
- **M5**: Jede Saldo-Assertion (`0,00 €`/`5,00 €` etc.) in den drei Specs inline
  auf das Saldo-Element im Tisch-Header scopen statt ungescoptem
  `getByText().first()`.
- **M6**: In `settleAlleOffenenTische` (und den gleichartigen Stellen) die
  `isVisible()`/`count()`-Zweige durch expect-basierte Auto-Wait-Assertions
  ersetzen, wo der Zustand deterministisch ist.
- **N29**: Helper-Header-Kommentar auf den echten Namen/Parameter
  (`http_post_status`, POST-fix) korrigieren.
- **N30**: `restore_jotti_version` schreibt auch den Leer-Fall zurück (kein
  liegenbleibender Release-Pin in `.env`).
- **N39**: `oeffneTisch`/`bestellePosition` in den Specs aus
  `support/servicekraft.ts` importieren statt reimplementieren.
- **N40**: No-op `.first()` auf dem bereits `.last()`-kollabierten Locator
  entfernen.
- **N41**: Falsche Viewport-/Tisch-Kommentare korrigieren.
- **N42**: Dynamischen Import auf statischen Top-Level-Import umstellen.
- **N43**: Kommentar „CI ist strenger" korrigieren (Settings gelten global).

### Acceptance criteria

- [ ] ops-smoke legt vor `kassensitzung-eroeffnen` einen Betreiber an; der
  fachliche Ablauf erreicht Direktverkauf, Beleg und DSFinV-K-Export.
- [ ] `http_post_status` erzeugt für jeden Aufruf syntaktisch gültiges JSON
  (empirisch: kein Trailing-`}` mehr); alle Aufrufer übergeben den Body explizit.
- [ ] Die drei Saldo-Specs prüfen den Restsaldo/Ausgleich am Header-Saldo-Element;
  eine bewusst falsche Saldo-Zahl im Test würde jetzt fehlschlagen.
- [ ] `settleAlleOffenenTische` verzweigt nicht mehr auf ungewartete
  `isVisible()`/`count()`; die kassenabschluss-Spec ist gegen den Fetch-Race
  robust.
- [ ] N29, N30, N39–N43 wie beschrieben umgesetzt.
- [ ] E2E-Suite grün; ein Trockenlauf von `scripts/ops-smoke.sh` gegen einen
  frischen Stack erreicht mindestens `kassensitzung-eroeffnen` ohne
  `betreiber_nicht_konfiguriert`.

---

## Phase 4: Minor-Cleanups nach Bereich

**User stories**: n/a (Codequalität)

Reine Cleanups ohne Laufzeit-Korrektheitswirkung, gruppiert nach Bereich. Als
„unverifiziert" markierte Befunde werden vor dem Anwenden je Befund am Code
bestätigt.

### Context

- Backend (Go): `backend/app/routes.go`, `backend/dsfinvkpruefung/*`,
  `backend/seed/*`, `windows/relay/main.go`,
  `backend/api/druck/auftrag/http/handler.go`.
- Frontend: `frontend/src/components/ui/tabs.tsx`,
  `frontend/src/service/*`, `frontend/src/admin/settings/DruckstationConfigPage.tsx`,
  `frontend/src/service/components/table/*Drawer.tsx`.
- Docs: `docs/plans/*`, `docs/prds/prd-praxistest-fixes.md`,
  `database/migrations/README.md`.
- CI: `.github/workflows/ci.yml`, `scripts/setup-dev-tools.sh`,
  `e2e/tsconfig.json`.
- Tests (Lesbarkeit): `backend/dsfinvkpruefung/*_test.go`,
  `backend/api/kasse/kassenfuehrung/application/command_test.go`,
  `backend/app/matrix_integration_test.go`,
  `backend/api/fiskal/tse_live/tse_live_suite_test.go`,
  `backend/api/druck/relay/relay_integration_test.go`,
  `reverse-proxy/caddyfile_test.go`.

### What to build

**Backend (N1–N13):** N1 Doc-Kommentar `mountArea` (letzten Satz streichen);
N2 totes Feld `indexSpalte.DezimalKomma` entfernen; N3 Kommentar auf benannte
Spalten-Map kürzen; N4 `nil` bei ReadAll-Fehler in `pruefung.go` setzen; N5
`istPfad`/`hatEndung` durch `strings.ContainsAny`/`strings.HasSuffix` ersetzen;
N6 `gleicheReihenfolge` durch `slices.Equal` ersetzen; N7 `regelDateinameGrafik`
→ `regelDateinameFremdformat`; N8 Package-Doku um fachliche Inhaltsregeln
ergänzen; N9 Bon-Namen-Umkehrregel ergänzen oder Satz streichen; N10 Demo-
Usernames über Konstanten; N11 `session_replication_role` nach `SeedTruncateAll`
auf `DEFAULT`/`origin` zurücksetzen und Kommentar präzisieren; N12 Relay-Retry-
Kommentar auf `MaxDruckversuche = 6` angleichen; **N13** Sammel-Endpoint-Antwort
behalten und die Zahl im Frontend-Toast anzeigen.

**Frontend (N14–N19):** **N14** interne Bezeichner in `tabs.tsx` anglisieren
(Export bleibt); N15 `onAdd(id, maxMenge)` → `onAdd(id)`; N16
`produktlistenFreiraum` neutral benennen; N17 „Alle verwerfen"-Dialog als
Same-File-Komponente extrahieren; N18 Drawer-Mapping-Helper in `drawerUtils.ts`;
N19 `LadefehlerAlert`-Komponente in `components/common`.

**Docs (N20–N25):** N20 DSFinV-K-Export-Pfad korrigieren; N21 tote Links auf
`plan-v0.14.0-breaking.md` entlinken; N22 „Sammel-Retry" → „Sammel-Verwerfen";
N23 abgeschlossenen `plan-befund-fixes-v1.0.0.md` löschen; N24 offene Checkbox
auf gelöschten Befund-Report aktualisieren; N25 Migrations-README auf
`02_druckauftrag_backoff` aktualisieren.

**CI (N26–N28, N44):** **N26** toten e2e-Pfadfilter entfernen, „läuft immer auf
push/PR" belassen, Kommentar anglisieren; N27 `setup-dev-tools.sh` auf Go 1.26.5
und pnpm-Major 11 angleichen; **N28** Typecheck-Gate für `e2e/` ergänzen samt
**N44** `include`-Fix in `e2e/tsconfig.json` (`helpers/**/*.ts` aufnehmen).

**Test-Lesbarkeit (N32–N38):** N32 `cleanSeedDB`-Spiegel nachziehen oder Anspruch
streichen; N33 Caddyfile-Test auf volle Header angleichen; N34 Helper
`muessePruefen`; N35 `ergebnis, err` → `_, err`; N36 `map[user.Role]string`;
N37 TSE-Live-Suite-Kommentare korrigieren, Credentials durchreichen; N38
Body-Close-Muster angleichen.

### Acceptance criteria

- [ ] Alle N1–N30 (soweit nicht in Phase 1/3), N32–N38, N44 wie oben umgesetzt.
- [ ] N13: `druckauftraege-verwerfen` behält die Antwort; der Client zeigt die
  Anzahl im Toast.
- [ ] N14: `tabs.tsx` enthält intern keine deutschen Bezeichner mehr; der
  öffentliche Export ist unverändert.
- [ ] N23: `docs/plans/plan-befund-fixes-v1.0.0.md` ist gelöscht.
- [ ] N26: Die e2e-Job-Bedingung ist nicht mehr tautologisch verschachtelt, der
  Pfadfilter-Block ist entfernt, der Kommentar ist englisch.
- [ ] N28/N44: Ein Typecheck-Gate prüft `e2e/`-TypeScript inkl. `helpers/`;
  `pnpm typecheck` im e2e-Verzeichnis ist grün.
- [ ] „unverifiziert" markierte Befunde wurden vor dem Anwenden je Befund am
  Code bestätigt.
- [ ] `make verify` und die Frontend-/E2E-Gates grün.

---

## Phase 5: G1–G10 — Coverage-Lücken untersuchen + M2-Doku

**User stories**: n/a (Investigation)

### Context

- `docker-compose.e2e.yml`, `reverse-proxy/caddyfile.go`,
  `reverse-proxy/main.go` — G1, G2.
- `.claude/settings.json` — G3 (PostToolUse-Hook).
- `.env.fiskaly-test.example`, `.gitignore` — G4.
- `frontend/package.json`, `e2e/pnpm-lock.yaml`, Dockerfiles — G5.
- `docs/verfahrensdokumentation.md`, `docs/leitfaden/` — G6.
- `backend/sqlc/queries/seed.sql`, `backend/seed/writer.go` — G7.
- `backend/dsfinvkpruefung/`, `docs/rechtsquellen/` — G8.
- `backend/api/fiskal/tse_live/`, `scripts/test-integration.sh` — G9.
- `frontend/src/index.css` — G10.

### What to build

Jede der zehn Coverage-Lücken wird untersucht und mit einem Befund
(bestätigt / widerlegt, mit Belegstelle) dokumentiert. Nur bestätigte Lücken
werden zu Fixes; umfangreichere Fixes werden als eigener Folgeplan geschnitten,
nicht in dieser Phase erzwungen. Enthält außerdem die M2-Doku-Zeile
(Runbook/Release-Notes: v0.14.0-TSE-Setups vor dem ersten DSFinV-K-Export neu
einrichten).

Untersuchungsfokus je Lücke: G1 e2e-Compose-Verkabelung (durch Phase 1 bereits
adressiert — als erledigt verifizieren); G2 kann `PROXY_HTTP_ONLY` versehentlich
in Prod aktiv werden und gilt die CSP-/Header-Parität wirklich; G3 Injection-/
Traversal-Sicherheit des `prettier`-Hooks; G4 `.env.fiskaly-test`-Ignore-
Negation und keine echten Credentials im Example; G5 shadcn devDependencies-
Verschiebung (Laufzeit-Impact), e2e-Lockfile vs. 24h-Policy, Go-1.26.5-Parität;
G6 fachliche Korrektheit der Compliance-Doku gegen Code und `docs/rechtsquellen/`
(zugleich Herstellerdoku nach BSI TR-03153-1); G7 welche Tabellen der Reset
truncated und ob TSE-signierte Daten weggeworfen werden können; G8 inhaltliche
DSFinV-K-Spec-Konformität der `dsfinvkpruefung`-Regeln; G9 ob ein normaler
Integrationslauf mit gesourcten Credentials die echte fiskaly-TEST-API trifft;
G10 WCAG-AA-Kontrastwerte / Dark-Mode-Regressionen der Farb-Token.

### Acceptance criteria

- [ ] Für jede Lücke G1–G10 liegt ein dokumentierter Befund (bestätigt/widerlegt
  mit Belegstelle) vor.
- [ ] G1 ist als durch Phase 1 erledigt verifiziert.
- [ ] Bestätigte, klein umsetzbare Lücken sind gefixt; größere sind als
  Folgeplan mit Scope beschrieben (kein stiller Cap ohne Vermerk).
- [ ] Die M2-Runbook-/Release-Notes-Zeile ist ergänzt.
- [ ] `make verify` grün.
