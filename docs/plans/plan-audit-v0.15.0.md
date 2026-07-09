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

- [x] `Config.EnableTestApi` existiert und wird in `Load()` aus
  `JOTTI_ENABLE_TEST_API == "1"` gesetzt.
- [x] `SetupRoutes` registriert die Test-Reset-Area anhand `cfg.EnableTestApi`;
  kein direkter `os.Getenv`-Aufruf mehr für diese Entscheidung.
- [x] `testResetArea` ist mit `RateLimited: true` deklariert.
- [x] `docker-compose.e2e.yml` setzt `JOTTI_ENABLE_TEST_API: "1"`, nicht mehr
  `JOTTI_ALLOW_SEED`; das Reverse-Proxy-Port-Mapping bindet an `127.0.0.1`.
- [x] `docker-compose.rocks.yml`, `docker-compose.local.yml`,
  `docker-compose.yml` behalten `JOTTI_ALLOW_SEED` (CLI); die Kommentare nennen
  nur noch das Subkommando, nicht den HTTP-Endpoint.
- [x] `backend/app/app_test.go` prüft Registrierung/Nicht-Registrierung über
  `JOTTI_ENABLE_TEST_API`; der Kommentar in `matrix_integration_test.go` nennt
  das korrekte Flag.
- [x] `scripts/reset-and-seed.sh` funktioniert unverändert (CLI-Pfad, kein
  Bezug auf das neue Flag).
- [x] `make verify` grün; E2E-Suite grün (Reset läuft weiter über den HTTP-
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

- [x] Bei vorhandenem Kassensturz **und** mindestens einem Buchungs-Event danach
  liefert `KasseAbschliessen` einen klaren, benannten Fehler und schreibt kein
  Abschluss-Event.
- [x] Bei vorhandenem Kassensturz **ohne** Zwischenbuchung bleibt der Wiederanlauf
  wie bisher (Schritt 1 übersprungen, alter Ist maßgeblich, Abschluss erfolgreich).
- [x] Neuer Regressionstest: Teilfehler nach Schritt 1 → `defer` setzt auf
  `offen` → Zwischenbuchung → Retry schlägt mit dem neuen Fehler fehl.
- [x] Bestehende Wiederanlauf-Tests (sofortiger Retry, Teilfehler nach Schritt 2
  ohne Zwischenbuchung) bleiben grün.
- [x] `make verify` grün.

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

- [x] ops-smoke legt vor `kassensitzung-eroeffnen` einen Betreiber an; der
  fachliche Ablauf erreicht Direktverkauf, Beleg und DSFinV-K-Export.
- [x] `http_post_status` erzeugt für jeden Aufruf syntaktisch gültiges JSON
  (empirisch: kein Trailing-`}` mehr); alle Aufrufer übergeben den Body explizit.
- [x] Die drei Saldo-Specs prüfen den Restsaldo/Ausgleich am Header-Saldo-Element;
  eine bewusst falsche Saldo-Zahl im Test würde jetzt fehlschlagen.
- [x] `settleAlleOffenenTische` verzweigt nicht mehr auf ungewartete
  `isVisible()`/`count()`; die kassenabschluss-Spec ist gegen den Fetch-Race
  robust.
- [x] N29, N30, N39–N43 wie beschrieben umgesetzt.
- [x] E2E-Suite grün; ein Trockenlauf von `scripts/ops-smoke.sh` gegen einen
  frischen Stack erreicht mindestens `kassensitzung-eroeffnen` ohne
  `betreiber_nicht_konfiguriert`.
  (Nachweis: die API-Sequenz des Skripts — OTP, set-password, login,
  update-betreiber, kassensitzung-eroeffnen — wurde 1:1 gegen einen frischen,
  ungeseedeten lokalen Stack abgespielt, alle Schritte 200. Der Lauf deckte
  einen latenten 500er auf: die Event-Validierung lehnte `betragCents: 0` ab
  (zog wertet den Zero-Value als fehlend), obwohl die HTTP-Schicht 0 erlaubt —
  gefixt in `fix(kasse): accept zero-cent …` samt Kassensturz-Ist-Bestand und
  Regressionstests. Der volle Skript-Lauf auf einem Wegwerf-Host bleibt das
  bekannte offene Human-Gate.)

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

- [x] Alle N1–N30 (soweit nicht in Phase 1/3), N32–N38, N44 wie oben umgesetzt.
  (Ausnahme N38: bewusst nicht angewendet — die Angleichung ans
  `_ =`-Wrapper-Muster lässt den bodyclose-Linter am Aufrufer fehlschlagen;
  stattdessen Kommentar an der Ausreißer-Stelle, warum das direkte `defer`
  dort absichtlich steht.)
- [x] N13: `druckauftraege-verwerfen` behält die Antwort; der Client zeigt die
  Anzahl im Toast.
- [x] N14: `tabs.tsx` enthält intern keine deutschen Bezeichner mehr; der
  öffentliche Export ist unverändert.
- [x] N23: `docs/plans/plan-befund-fixes-v1.0.0.md` ist gelöscht.
- [x] N26: Die e2e-Job-Bedingung ist nicht mehr tautologisch verschachtelt, der
  Pfadfilter-Block ist entfernt, der Kommentar ist englisch.
- [x] N28/N44: Ein Typecheck-Gate prüft `e2e/`-TypeScript inkl. `helpers/`;
  `pnpm typecheck` im e2e-Verzeichnis ist grün.
- [x] „unverifiziert" markierte Befunde wurden vor dem Anwenden je Befund am
  Code bestätigt.
- [x] `make verify` und die Frontend-/E2E-Gates grün.

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

- [x] Für jede Lücke G1–G10 liegt ein dokumentierter Befund (bestätigt/widerlegt
  mit Belegstelle) vor (siehe Befunde unten).
- [x] G1 ist als durch Phase 1 erledigt verifiziert.
- [x] Bestätigte, klein umsetzbare Lücken sind gefixt; größere sind als
  Folgeplan mit Scope beschrieben (kein stiller Cap ohne Vermerk).
  (Vermerk: der G3-Hook-Fix ist als konkreter Patch beschrieben, aber nicht
  angewendet — der Agent darf seine eigene Hook-Konfiguration in
  `.claude/settings.json` nicht ändern; Anwendung durch den Maintainer.)
- [x] Die M2-Runbook-/Release-Notes-Zeile ist ergänzt
  (`docs/leitfaden/tse-sonderfaelle.md`, Absatz „TSE unter v0.14.0
  eingerichtet?").
- [x] `make verify` grün.

### Phase-5-Befunde (2026-07-09)

- **G1 — erledigt (durch Phase 1).** `docker-compose.e2e.yml` setzt nur noch
  `JOTTI_ENABLE_TEST_API: "1"` (Z. 69), bindet den Proxy an
  `127.0.0.1:${E2E_HTTP_PORT}:80` (Z. 121); `testResetArea` ist rate-limited.
  Zwei komplette E2E-Läufe (23/23) liefen über genau diesen Pfad.
- **G2 — widerlegt.** `PROXY_HTTP_ONLY` wird strikt geparst (nur
  `1/true/yes`, `main.go:212-219`), keine Prod-/Release-Compose-Datei und
  keine `.env`-Vorlage deklariert die Variable; ohne Verkabelung erreicht auch
  eine Host-Variable den Container nicht. Header-Parität gilt: alle drei Modi
  rendern dasselbe `proxySnippet` (`caddyfile.go:96-137`); einziger
  Unterschied ist der beabsichtigte HSTS-Wert.
- **G3 — teilweise bestätigt, Patch offen.** Pfad-Übergabe im Prettier-Hook
  ist injection-sicher (durchgängig gequotet, kein eval). Aber die
  `node_modules`-Suche läuft bis `/` hoch statt an der Repo-Wurzel zu stoppen —
  außerhalb liegende `node_modules/.bin/prettier` würden ausgeführt (aktuell
  existiert dort keines, daher No-op). Vorgeschlagener Patch für
  `.claude/settings.json` (vom Maintainer anzuwenden, Agent darf die eigene
  Hook-Config nicht ändern): in der while-Schleife vor dem Binary-Test
  `if [ -n "$CLAUDE_PROJECT_DIR" ]; then case "$d" in
  "$CLAUDE_PROJECT_DIR"|"$CLAUDE_PROJECT_DIR"/*) ;; *) break;; esac; fi;`
  einfügen.
- **G4 — widerlegt.** `.gitignore` (Z. 15–17) ignoriert `.env.fiskaly-test`
  inkl. Varianten, die Negation gilt nur der Example-Datei; das Example
  enthält ausschließlich leere Keys; nichts Echtes getrackt, nichts in der
  Historie v0.14.0..HEAD.
- **G5 — teilweise bestätigt, gefixt.** Nur das CLI-Tool `shadcn` wanderte
  nach devDependencies (kein Laufzeit-Impact, Vite bundelt build-time);
  24h-Policy greift auch für e2e (pnpm-11-Default). Gefixt: Dev-Stack-Image
  `docker-compose.yml` von `golang:1.26.0-alpine` auf `1.26.5-alpine`
  (übersehener Rest des Toolchain-Bumps); `e2e/` in `dependabot.yml`
  aufgenommen.
- **G6 — teilweise bestätigt, gefixt.** Verfahrensdoku fachlich gegen Code
  und Rechtsquellen geprüft (processType-Tabelle, Outbox, DSFinV-K-Struktur,
  Append-only, Ausfallpfad: alles korrekt). Einzige Diskrepanz: §5 nannte den
  Export nur für „(abgeschlossene)" Kassensitzungen, real sind auch offene
  exportierbar — Formulierung korrigiert.
- **G7 — Kern widerlegt, Follow-up-Empfehlung.** `jotti seed` (CLI) truncatet
  nichts und bricht bei vorhandenen Kassenjournal-Events vor jedem
  Schreibzugriff ab; der Truncate-Pfad hängt ausschließlich am
  E2E-only-HTTP-Endpoint. Restrisiko außerhalb des Codes:
  `scripts/reset-and-seed.sh --yes` (make local-reset-and-seed) löscht das
  Postgres-Volume am Guard vorbei — Empfehlung als Produktentscheidung offen:
  Bestands-Check (kassenjournal count) vor `docker volume rm` auch bei
  `--yes`.
- **G8 — Kern widerlegt (Validator korrekt), Kommentare präzisiert.** Kein
  Regelverstoß, der einen nicht-konformen jotti-Export durchwinken würde;
  Formatwerte, TSE-Pflichtfelder, Storno-Regel und Steuerschlüssel decken sich
  mit DSFinV-K 2.4. Zwei Regeln sind strenger als die Spec (Pfadverbot,
  Date-Spalten) — für jottis flachen Export folgenlos. Gefixt: zwei
  unpräzise Referenz-Kommentare in `inhalt.go` (TSE_SERIAL-Definition,
  BON_NAME-Konvention).
- **G9 — bestätigt, gefixt.** Ein normaler Integrationslauf hätte mit in der
  Shell exportierten `FISKALY_TEST_*`-Credentials die echte fiskaly-TEST-API
  getroffen (`test-integration.sh` läuft `-tags=integration ./...`, Guard
  prüfte nur Credential-Präsenz). Gefixt: explizites Opt-in
  `JOTTI_TSE_LIVE=1` an allen drei Live-Guards; nur `test-tse-live.sh` und
  `make test-tse-live-setup` setzen es. Ohne Opt-in skippen alle fünf
  Live-Tests (empirisch verifiziert).
- **G10 — teilweise bestätigt, Follow-up-Entscheidung.** Alle sechs geprüften
  Kern-Token-Paare bestehen WCAG AA in Light und Dark (berechnet, z. B.
  foreground/background 19,6:1 bzw. 18,9:1). Verstoß: solide
  Löschen-Buttons (`bg-destructive text-white`, 5 Stellen) erreichen im
  Dark-Mode nur 2,89:1 (`--destructive` = red-400). Fix braucht eine
  Design-Entscheidung: dunkleres Dark-`--destructive` (einfach, senkt aber
  `text-destructive`-Kontrast auf Flächen) oder eigenes
  `--destructive-foreground`-Token (präziser, 5 Call-Sites) — als Folgearbeit
  beschrieben, nicht still gefixt.

Zusätzlicher Befund außerhalb der G-Liste (aus dem Phase-3-Trockenlauf): das
zog-Muster `z.Int().GTE(0).Required()` lehnt den gültigen Wert 0 als „fehlend"
ab. Im Kassenführungs-Fluss gefixt (Eröffnungs-Betrag, Kassensturz-Ist);
dasselbe Muster steht noch auf Summenfeldern weiterer Event-Schemas
(`bestellung.go`, `zahlung.go`, `stornierung.go`, `tisch_session_events.go`,
`direktverkauf_events.go`, `umbuchung.go`) — dort ist 0 über die Domänenpfade
mutmaßlich nicht erreichbar, eine kurze Prüfung pro Feld steht als Folgearbeit
aus.
