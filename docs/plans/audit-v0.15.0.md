# Audit-Befunde Release v0.15.0

Autonomer Audit des Release-Diffs `v0.14.0..HEAD` (201 Dateien) am 2026-07-09.
Methodik: 15-Experten-Cleanup-Workflow (Lesbarkeit, Prinzipien, Smells,
Architektur, Cross-Layer, Tests, Korrektheit Backend/Frontend/Infra, Security)
plus vollständige Testsuite. Die adversariale Verifikationsstufe des Workflows
lief nach 26 von 66 Befunden ins Credit-Limit; die restlichen hochseveren
Befunde wurden manuell nachverifiziert (Vermerk je Befund).

Dieses Dokument ist zum Klären und Planen in einer neuen Session gedacht, noch
kein Umsetzungsplan. Keine Fixes sind angewendet.

## Verifikationslegende

- `verifiziert`: selbst end-to-end am Code, an Compose-Dateien, git-Historie
  oder empirisch geprüft.
- `3 Stimmen`: im Workflow durch drei unabhängige Verifizierer bestätigt.
- `1 Stimme`: im Workflow durch einen Verifizierer bestätigt.
- `unverifiziert`: vom Find-Experten gemeldet, Workflow-Verifikation traf das
  Credit-Limit; Plausibilität am Fundort gesichtet, aber nicht voll bestätigt.

## Testsuite (alle grün)

| Stufe | Ergebnis |
| --- | --- |
| `make verify` (Lint, Race-Unit, Build, Integration) | exit 0 |
| Playwright-E2E (echter Prod-Image-Stack, Desktop + Mobile) | 23/23 |
| Fuzz (4 Targets je 90s) | keine Crasher |
| TSE-Live gegen fiskaly-TEST-TSS | alle PASS, keine TSS angelegt |

Die grüne Suite deckt die untenstehenden Befunde nicht ab: der Reset-Endpunkt
wird von den E2E-Tests bewusst genutzt, der Wiederanlauf-Race hat einen engen
Trigger, das Backfill betrifft nur Bestandsdaten, und die Test-Quality-Befunde
sind selbst die Lücke.

---

## C1 — CRITICAL: Unauthentifizierter DB-Reset-Endpunkt auf dem Demo-Host

Status: verifiziert (end-to-end).

v0.15.0 fügt `POST /test/reset-and-seed` hinzu (Commit 6901630, E2E-Seed-Fundament).
Der Endpunkt ist unauthentifiziert und ohne Rate-Limit und truncated das
append-only Kassenjournal (er löst den Append-only-Trigger bewusst über
`SET LOCAL session_replication_role = replica`, `backend/seed/writer.go:106`).
Registriert wird er, sobald `JOTTI_ALLOW_SEED=1` gesetzt ist
(`backend/app/app.go:66`, `backend/app/routes.go:120-137`).

Bisher war das Flag ungefährlich: es schaltete nur das CLI-Subkommando
`jotti seed` frei. Ab v0.15.0 exponiert dasselbe Flag zusätzlich den
HTTP-Endpunkt. Gesetzt ist es in:

- `docker-compose.rocks.yml:83` — öffentliche jotti.rocks-Demo. `nginx.rocks.conf:126`
  proxied `location /api/` auf `http://backend:3000/`. Damit erreicht
  `POST https://<demo>/api/test/reset-and-seed` den Endpunkt aus dem Internet
  und leert die Demo-DB ohne Anmeldung. (nginx.rocks hat ein `api_limit`
  burst=20, verhindert aber keinen einzelnen erfolgreichen Wipe.)
- `docker-compose.local.yml:92` — LAN-Vereinsstack ("smallest setup, cash only").
  Jedes Gerät im Vereins-WLAN könnte aufbewahrungspflichtiges Kassenjournal
  ohne Login löschen.
- `docker-compose.yml:64` — Dev-Stack.

`docker-compose.prod.yml` und `docker-compose.release.yml` sind clean. Der
dokumentierte Produktions-/Release-Installpfad (`make prod-*`, Release-ZIP) ist
also nicht betroffen. Das begrenzt den Radius, aber die Live-Demo ist real
exponiert.

Der Endpunkt war zu v0.14.0 nicht vorhanden (`handler.go` und `testResetArea`
neu im Range).

Zugehörige Teilbefunde im selben Cluster:

- Kein Rate-Limit: `testResetArea` setzt `RateLimited` nicht, obwohl der Aufruf
  ein volles Truncate + Reseed auslöst (DoS-Vektor auf jedem ALLOW_SEED-Host).
  `backend/app/routes.go:125-137`.
- IoC-Bruch: `SetupRoutes` liest `os.Getenv` direkt statt über `config.Config`.
  `backend/app/app.go:66`.
- E2E-Compose bindet Port auf `0.0.0.0` bei repo-öffentlichen Secrets
  (`JWT_SECRET`, `RELAY_AUTH_TOKEN` im Klartext). `docker-compose.e2e.yml:120`.

Klärung nötig (siehe Entscheidungen unten): Die Demo braucht vermutlich bewusst
einen Reset. Deshalb ist die Fix-Richtung eine Produktentscheidung, kein rein
mechanischer Fix.

---

## Major-Befunde

### M1 — Kassenabschluss-Wiederanlauf nutzt veralteten Ist-Bestand

Status: 3 Stimmen. `backend/api/kasse/kassenfuehrung/application/command.go:354-366`.

Nach einem Teilfehler (Kassensturz-Event geschrieben, Schritt 2/3 gescheitert)
setzt der `defer` die Sitzung wieder auf `offen`. Folgen danach neue Buchungen,
überschreibt der spätere Abschluss den frisch übergebenen `istBestandCents`
stumm mit dem alten Sturz-Wert, während `sollBestandCents` frisch berechnet
wird. Die Differenzbuchung enthält dann legitime neue Umsätze als
Soll-Ist-Differenz, und der Tagesabschluss dokumentiert einen falschen
Kassenbestand. Die bestehenden Wiederanlauf-Tests decken nur den sofortigen
Retry ohne Zwischenbuchungen ab.

Fix-Richtung: im Wiederanlauf prüfen, ob nach dem vorhandenen Kassensturz-Event
weitere Buchungs-Events im Stream liegen (Sturz-Version bzw. protokollierter
Soll gegen aktuellen Stand), und bei Konflikt mit klarem Fehler abbrechen statt
den alten Ist-Bestand stumm wiederzuverwenden.

### M2 — TSS-Seriennummer wird für Bestandsinstallationen nicht nachgezogen

Status: verifiziert (git-Historie + Persistenzpfad). `backend/repository/tse_repo/fiskaly_setup.go:65-72`.

v0.14.0 las die TSS-Seriennummer aus dem falschen JSON-Feld: auf der
TSS-Ressource `tss_serial_number` (das existiert dort nicht, nur auf
Transaktions-Responses). v0.15.0 korrigiert das auf `serial_number`. Jede
TSE-Einrichtung unter v0.14.0 hat dadurch eine leere `seriennummer` in
`tse_stammdaten` persistiert (`fiskaly_setup.go:194`). Die Seriennummer wird nur
im Setup-Flow geschrieben; ein `make prod-update` zieht sie nicht nach.
Betroffene Installationen (die Erstinstallation 2026-07-07 lief auf v0.14.0)
exportieren weiter ein leeres DSFinV-K-Pflichtfeld `TSE_SERIAL` — genau das, was
die neue Inhaltsprüfung `regelTSEStammdaten` als Verstoß definiert.

Fix-Richtung: einmaliger Backfill bei konfigurierter TSE und leerer
`tse_stammdaten.seriennummer` (Stammdaten per `RetrieveTSSStammdaten` nachziehen,
z. B. beim Start oder vor dem Export) oder mindestens ein dokumentierter
Upgrade-Schritt (Einrichtung erneut durchlaufen) in Release-Notes/Runbook.

Offene Frage: Hat die 2026-07-07-Instanz TSE tatsächlich eingerichtet und macht
sie DSFinV-K-Exporte? Davon hängt die Dringlichkeit ab.

### M3 — ops-smoke.sh konfiguriert nie den Betreiber

Status: verifiziert (grep: kein `update-betreiber` im Skript). `scripts/ops-smoke.sh:272-285`.

Auf einem frischen Host (Skript-Pflicht) ist die `betreiber`-Tabelle leer.
`kassensitzung-eroeffnen` liefert dann `400 betreiber_nicht_konfiguriert`, und
der gesamte release-Smoke-Lauf bricht am ersten fachlichen Schritt ab. Das
Skript ist der noch offene Ops-Smoke-Human-Gate und wurde noch nicht real
ausgeführt, würde also beim ersten Lauf scheitern.

Fix-Richtung: vor `kassensitzung-eroeffnen` einen Schritt ergänzen, der
`POST /api/admin/update-betreiber` mit Dummy-Stammdaten aufruft.

### M4 — ops-smoke.sh Bash-Default hängt `}` an jeden JSON-Body

Status: verifiziert (empirisch). `scripts/ops-smoke.sh:137-141`.

`local url="$1" data="${2:-{}}"` liefert bei gesetztem Body den Body plus ein
zusätzliches `}`. Empirisch: Eingabe `{"a":1}` wird zu `{"a":1}}`. set-password,
login und der Rate-Limit-Check senden dadurch syntaktisch ungültiges JSON. Es
funktioniert heute nur, weil `helper.ReadBody` den `json.Decoder` nutzt und
Trailing-Daten ignoriert; jede Decoder-Verschärfung (`decoder.More()`-Check)
lässt install/release/ops-Smoke sofort scheitern.

Fix-Richtung: Default separat setzen, z. B.
`local url="$1" data="${2:-}"; [[ -n "$data" ]] || data='{}'`.

### M5 — Vakuose Saldo-Assertions in E2E-Specs

Status: unverifiziert (Fundort gesichtet). `e2e/tests/tischservice-teilzahlung.mobile.spec.ts:45-50`.

`page.getByText('0,00 €').first()` / `'5,00 €'` verankert nicht am
Saldo-Element. Die StickyActionBar zeigt bei leerer Auswahl immer `0,00 €`, und
Positionszeilen zeigen Einzelpreise, also passieren die Assertions auch bei
falschem Saldo. Die Kernaussage der Specs (Restsaldo/Ausgleich) wird faktisch
nicht geprüft. Betroffen auch `bestellen-kassieren.spec.ts:69` und
`kassenabschluss.mobile.spec.ts:47`.

Fix-Richtung: Saldo-Assertion auf das Saldo-Element im Tisch-Header scopen (oder
einen Saldo-Helper in `support/servicekraft.ts`) statt ungescoptem
`getByText().first()`.

### M6 — Race-anfällige Conditionals in der E2E-Support-Schicht

Status: unverifiziert (Fundort gesichtet). `e2e/support/servicekraft.ts:193-228`.

`isVisible()`/`count()` warten nicht auf das Rendern; `TablePage` rendert
Tab-Inhalte erst nach `stateLoading`. Läuft der Fetch noch, wird der
Ausgabe- bzw. Kassieren-Zweig still übersprungen, und die kassenabschluss-Spec
scheitert erst Minuten später am `tische_saldo_offen`-Gate. Mit CI `retries: 0`
schwer diagnostizierbar. Gleiches Muster in Zeilen 57 und 154.

Fix-Richtung: vor den Conditionals auf ein deterministisches Ready-Signal warten
(z. B. verschwundener Ladehinweis oder sichtbare Positionsliste), erst danach
verzweigen.

---

## Minor-Befunde

Alle bestätigt (3, 1 Stimme oder unverifiziert je Vermerk); keiner mit
Laufzeit-Korrektheitswirkung. Nach Bereich gruppiert.

### Backend (Go)

- N1 `backend/app/routes.go:97-102` — Doc-Kommentar von `mountArea` verspricht
  einen Rückgabewert ("Gibt die absoluten Pfade des Bereichs zurück"), die
  Funktion hat keinen und verwirft die Pfade mit `_`. Letzten Satz streichen.
  (3 Stimmen)
- N2 `backend/dsfinvkpruefung/indexxml.go:30-34` — totes Feld
  `indexSpalte.DezimalKomma`, nirgends gesetzt oder gelesen. Entfernen. (1 Stimme)
- N3 `backend/dsfinvkpruefung/inhalt.go:203-214` — Kommentar behauptet Zugriff
  über feste Spaltenordnung, real über benannte Spalten-Map. Kommentar kürzen.
  (1 Stimme)
- N4 `backend/dsfinvkpruefung/pruefung.go:86-106` — Kommentar verspricht
  "Lesefehler als leerer Inhalt", aber `data, _ := io.ReadAll(...)` speichert
  partiell gelesene Bytes bei CRC-Fehler. `nil` bei ReadAll-Fehler setzen wie im
  Open-Pfad. (3 Stimmen)
- N5 `backend/dsfinvkpruefung/paket.go:84-101` — `istPfad`/`hatEndung` sind
  Handkopien von `strings.ContainsAny`/`strings.HasSuffix`. Durch stdlib
  ersetzen. (1 Stimme)
- N6 `backend/dsfinvkpruefung/csv.go:138-149` — `gleicheReihenfolge` ist ein
  handgeschriebenes `slices.Equal`. Durch `slices.Equal` ersetzen (Aufrufer in
  csv.go, inhalt.go, pruefung_test.go). (1 Stimme)
- N7 `backend/dsfinvkpruefung/paket.go:13` — Konstante `regelDateinameGrafik`
  deckt jede Fremdformat-Datei ab, nicht nur Grafik. In
  `regelDateinameFremdformat` umbenennen. (1 Stimme)
- N8 `backend/dsfinvkpruefung/pruefung.go:9-12` — Paket-Doku sagt "ausschließlich
  formale Strukturkonformität", aber das Paket prüft seit inhalt.go fachliche
  Inhaltsregeln. Package-/Befund-/Pruefen-Doku ergänzen. (unverifiziert)
- N9 `backend/dsfinvkpruefung/inhalt.go:307-334` — Kommentar verspricht eine
  Umkehrregel für den Tagesabschluss-Bon-Namen, die die Schleife nicht
  implementiert. Regel ergänzen oder Satz streichen. (unverifiziert)
- N10 `backend/seed/szenario.go:24-33` — Demo-Usernames doppelt kodiert
  (Konstanten vs. Literale in `demoSzenario`). Konstanten direkt in den
  Benutzer-Definitionen verwenden. (1 Stimme)
- N11 `backend/seed/writer.go:98-119` — `session_replication_role = replica` gilt
  für die ganze Reset-Transaktion und deaktiviert auch die FK-Prüfung für alle
  Seed-Inserts. Nach `SeedTruncateAll` auf `DEFAULT`/`origin` zurücksetzen;
  Kommentar präzisieren. (unverifiziert)
- N12 `windows/relay/main.go:311-313` — Kommentar sagt "markiert nach drei
  Versuchen als fehlgeschlagen", das Backend gibt erst nach
  `MaxDruckversuche = 6` auf. Kommentar angleichen. (unverifiziert)
- N13 `backend/api/druck/auftrag/http/handler.go:118-132` — `DiscardAlle...Handler`
  liefert `{"verworfen": n}`, der einzige Client verwirft die Antwort; der
  Schwester-Endpunkt antwortet leer. Auf `SendEmptyResponse` vereinheitlichen
  oder Zahl im Toast anzeigen. (unverifiziert)

### Frontend (TypeScript/React)

- N14 `frontend/src/components/ui/tabs.tsx:61-116` — deutsche Bezeichner
  (`aktualisiereAffordance`, `scrolle`, `richtung`, `kind`) in generischer
  shadcn-Primitive gemischt mit englischen; verstößt gegen Regel 6
  (Infrastruktur englisch). Interne Namen anglisieren, Export bleibt. (1 Stimme)
- N15 `frontend/src/service/components/PositionAuswahlListe.tsx:16-21` —
  `onAdd(id, maxMenge)` reicht Daten zurück, die der Aufrufer schon besitzt;
  2 von 3 Aufrufern verwerfen `maxMenge` per Wrapper. Auf `onAdd(id)`
  vereinfachen. (1 Stimme)
- N16 `frontend/src/service/TablePage.tsx:29-37` — `produktlistenFreiraum` (samt
  Kommentar) stylt auch den Kassieren-Tab, der keine Produktliste hat. Neutral
  benennen. (1 Stimme)
- N17 `frontend/src/admin/settings/DruckstationConfigPage.tsx:217-323` —
  `FehlgeschlageneDruckauftraege` ist ~106 Zeilen und bettet einen 40-zeiligen
  AlertDialog inline ein. "Alle verwerfen"-Dialog als Same-File-Komponente
  extrahieren (analog `FehlgeschlagenerDruckauftragRow`). (1 Stimme)
- N18 `frontend/src/service/components/table/HistorieStornierungDrawer.tsx:98-106` —
  identisches 9-zeiliges Position→AuswahlPosition-Mapping in drei Drawern
  (auch HistorieUmbuchungDrawer 143-151, DirektverkaufStornoDrawer 97-105).
  Helper in `drawerUtils.ts` extrahieren. (1 Stimme)
- N19 `frontend/src/service/TablePage.tsx:63-76` — Query-Ladefehler-Alert mit
  Retry-Button dupliziert in `KassensitzungPage.tsx:32-48`. Kleine
  `LadefehlerAlert`-Komponente in `components/common`. (1 Stimme)

### Docs

- N20 `docs/plans/plan-qa-automatisierung.md:237-240` — verortet den
  DSFinV-K-Export in der Finanzamt-Ansicht; real liegt er unter
  `/admin/auswertung`. Widerspricht dem abgehakten N5-Kriterium. Text
  korrigieren. (1 Stimme)
- N21 `docs/plans/plan-v1.0.0-release.md:3,51` und `audit-v1.0.0.md:5` — tote
  Links auf das gelöschte `plan-v0.14.0-breaking.md`. Entlinken / als
  historischen Verweis formulieren. (1 Stimme)
- N22 `docs/prds/prd-praxistest-fixes.md:221,287-289` — zwei übrig gebliebene
  "Sammel-Retry"-Erwähnungen nach dem Pivot zu "Sammel-Verwerfen". Angleichen.
  (1 Stimme)
- N23 `docs/plans/plan-befund-fixes-v1.0.0.md` — vollständig abgeschlossener Plan
  (42/42) liegt noch in `docs/plans/`; Git-Workflow verlangt Löschung nach
  Merge. Löschen. (1 Stimme)
- N24 `docs/plans/guide-manuelle-qa-v1.0.0.md:63` — offene Checkbox verweist auf
  den gelöschten Befund-Report. Text aktualisieren oder Halbsatz streichen.
  (1 Stimme)
- N25 `database/migrations/README.md:17` — nennt `01_initial` als letzte
  Migration, obwohl `02_druckauftrag_backoff` existiert. Aktualisieren oder
  Klammerzusatz streichen. (unverifiziert)

### CI / Skripte

- N26 `.github/workflows/ci.yml:506-513` — deutscher Job-Kommentar mit
  CAPS-Emphase und `ae`-Transliteration, einziger nicht-englischer Kommentar der
  Datei; zusätzlich ist die `if`-Bedingung des e2e-Jobs tautologisch (Workflow
  triggert nur auf push/pull_request), der e2e-Pfadfilter (Zeilen 21, 46-52)
  damit toter Code. Kommentar anglisieren; Bedingung auf
  `needs.changes.outputs.e2e == 'true'` reduzieren oder Filter entfernen.
  (1 Stimme + unverifiziert)
- N27 `scripts/setup-dev-tools.sh:27-37` — "CI reference"-Block nennt Go 1.26.0
  und pnpm 10, CI pinnt seit 63c0e0d Go 1.26.5 und pnpm/action-setup 11. Als
  D13-Paritätsskript selbst gedriftet. Auf 1.26.5 aktualisieren, pnpm-Major
  abgleichen. (2 Stimmen)
- N28 `.github/workflows/ci.yml:558-562` — `e2e/`-TypeScript wird nirgends
  typgeprüft (`pnpm typecheck` in keinem Gate); zusätzlich fehlt `helpers/**/*.ts`
  im `include` von `e2e/tsconfig.json:17`. typecheck-Schritt ergänzen. (unverifiziert)
- N29 `scripts/ops-smoke.sh:135-141` — Helper-Header nennt falschen Namen und
  Parameter (`http_status METHOD URL`), Funktion heißt `http_post_status` und ist
  POST-fix. Kommentar korrigieren. (unverifiziert)
- N30 `scripts/ops-smoke.sh:444-454` — `restore_jotti_version` stellt bei leerem
  vorherigem Wert nicht zurück, der Release-Pin bleibt in `.env`. Auch Leer-Fall
  zurückschreiben. (unverifiziert)
- N31 `docker-compose.e2e.yml:120` — Port-Mapping auf `0.0.0.0` bei
  repo-öffentlichen Secrets und `JOTTI_ALLOW_SEED=1`. Auf `127.0.0.1` einschränken.
  (unverifiziert; Teil des C1-Clusters)

### Tests (nur Lesbarkeit)

- N32 `backend/dsfinvkpruefung/pruefung_integration_test.go:22-49` — `cleanSeedDB`
  ist eine Kopie des Seeder-Helpers, dem seit c8a54ca die
  tse_konfiguration-Normalisierung fehlt. Nachziehen oder Spiegel-Anspruch aus
  dem Kommentar streichen. (unverifiziert)
- N33 `reverse-proxy/caddyfile_test.go:164-173` — `TestRenderHTTPOnlyCaddyfile`
  prüft nur den nackten CSP-Wert, die Schwester-Tests die vollen Header inkl.
  X-Content-Type-Options / X-Frame-Options. Angleichen. (unverifiziert)
- N34 `backend/dsfinvkpruefung/pruefung_test.go:160-325` — Fehler-Rückgabewert von
  `PruefenBytes` in ~25 Stellen verworfen (`befunde, _ :=`). Kleinen Helper
  `muessePruefen(t, ...)` einführen. (unverifiziert)
- N35 `backend/api/kasse/kassenfuehrung/application/command_test.go:273-277` —
  `ergebnis, err := ...` gefolgt von `_ = ergebnis`. Auf `_, err :=` reduzieren.
  (unverifiziert)
- N36 `backend/app/matrix_integration_test.go:29-33` — Feld `testUser.role` gesetzt,
  nie gelesen; Map ist bereits nach Rolle geschlüsselt. `map[user.Role]string`.
  (unverifiziert)
- N37 `backend/api/fiskal/tse_live/tse_live_suite_test.go:228-231,396-397,576-583` —
  `starteWorker` behauptet eine Stop-Funktion (läuft über `t.Cleanup`);
  verstümmelter Kommentar "j=letzten"; `pruefeStammdatenVollstaendigkeit` liest
  Credentials erneut per `os.Getenv` statt sie durchzureichen. Kommentare
  korrigieren, Credentials durchreichen. (unverifiziert)
- N38 `backend/api/druck/relay/relay_integration_test.go:250` — eine Stelle nutzt
  `defer resp.Body.Close() //nolint:errcheck`, fünf andere das
  `_ =`-Wrapper-Muster. Angleichen. (unverifiziert)
- N39 `e2e/tests/kassieren-fehlerpfade.spec.ts:61-96` und
  `bestellen-kassieren.spec.ts:34-42,75-90` — `oeffneTisch`/`bestellePosition`
  reimplementiert statt aus `support/servicekraft.ts` genutzt. Importieren.
  (unverifiziert)
- N40 `e2e/tests/direktverkauf-storno.mobile.spec.ts:39` — `.first()` auf einem
  bereits per `.last()` kollabierten Locator, No-op. Entfernen. (unverifiziert)
- N41 `e2e/tests/bestellen-kassieren.spec.ts:8-14` — Kommentar behauptet Lauf in
  beiden Viewports (Config beschränkt auf mobile-service) und nennt Tisch 1
  fälschlich unbenutzt (ist Frühschoppen-Stammtisch). Beide Kommentare
  korrigieren. (unverifiziert)
- N42 `e2e/tests/admin-dsfinvk-export.spec.ts:36-38` — dynamischer Import mit
  `.then()`-Kette statt statischem Top-Level-Import. Umstellen. (unverifiziert)
- N43 `e2e/playwright.config.ts:8-10` — Kommentar "CI ist strenger" beschreibt
  Einstellungen, die global gelten (retries/trace/screenshot). Korrigieren.
  (unverifiziert)
- N44 `e2e/tsconfig.json:17` — `helpers/**/*.ts` nicht im `include`; Utilities auf
  `support/` und `helpers/` verteilt. Zusammenführen oder include ergänzen.
  (unverifiziert)

---

## Coverage-Lücken (kein Experten-Fokus, verdienen Zweitblick)

- G1 `docker-compose.e2e.yml` (Repo-Root): Compose-Verkabelung von
  `JOTTI_ALLOW_SEED=1` + `PROXY_HTTP_ONLY=1` + `0.0.0.0`-Bind, von keinem
  pathspec-Experten abgedeckt. (Teil von C1)
- G2 `reverse-proxy` PROXY_HTTP_ONLY-Klartextmodus (caddyfile.go, main.go): kann
  der Modus in Prod versehentlich aktiv werden? Gilt die CSP-/Header-Parität zu
  den TLS-Modi wirklich?
- G3 `.claude/settings.json` PostToolUse-Hook: führt bei jedem Write/Edit eine
  Shell-Pipeline aus, die per Verzeichnis-Traversal das erste
  `node_modules/.bin/prettier` ausführt (Supply-Chain-Ausführungsvektor).
  Injection-Sicherheit der Dateipfade und Traversal-Grenzen prüfen.
- G4 Secret-Hygiene: `.env.fiskaly-test.example` + `.gitignore`-Negationen. Prüfen,
  dass `!.env.fiskaly-test.example` die echte `.env.fiskaly-test` nicht
  un-ignored und das Example keine echten Credentials enthält.
- G5 Dependency-/Toolchain-Diff: shadcn von dependencies nach devDependencies
  verschoben (Laufzeit-Impact?); neues `e2e/`-Lockfile (greift die
  24h-minimumReleaseAge-Policy?); Go-1.26.5-Bumps in vier Modulen vs.
  Dockerfile-Builder-Images.
- G6 Compliance-Dokus (verfahrensdokumentation.md, leitfaden/): read-docs prüfte
  nur Prosa/Slop, nicht die fachliche Korrektheit gegen Code und
  `docs/rechtsquellen/`. Verfahrensdoku ist zugleich Herstellerdoku nach BSI
  TR-03153-1.
- G7 `backend/sqlc/queries/seed.sql` (+34 Zeilen Reset-SQL): welche Tabellen
  truncated der Reset genau, kann er TSE-signierte Daten wegwerfen?
- G8 `backend/dsfinvkpruefung/` Spec-Konformität: kein Experte gleicht die
  Prüfregeln inhaltlich mit der DSFinV-K-Spezifikation ab (Pflichtfelder,
  Feldformate, index.xml-Struktur). Ein fachlich falscher Validator gibt
  trügerische Sicherheit.
- G9 Gating der TSE-Live-Tests: `//go:build integration`-Tests skippen nur bei
  fehlenden `FISKALY_TEST_*`-Vars; `scripts/test-integration.sh` läuft mit
  `-tags=integration` über `./...`. Mit gesourcten Credentials trifft ein
  normaler Integrationslauf die echte fiskaly-TEST-API.
- G10 `frontend/src/index.css` Farb-Token (WCAG-AA-Fix): tatsächliche
  Kontrastwerte / Dark-Mode-Regressionen nicht per Code-Lesen verifizierbar,
  a11y-kritisch für BYOD-Outdoor-Nutzung.

---

## Widerlegt (nicht erneut untersuchen)

- `frontend/src/components/common/FormFields.tsx:86-96` — `useId` nur in
  `UsernameField`: behebt ein echtes Defizit (Input hatte gar keine id,
  `htmlFor` zeigte ins Leere, brach `getByLabel`), keine Kollisionsgefahr bei den
  Geschwistern. Reine Geschmacksfrage.
- `backend/api/kasse/kassenfuehrung/application/command.go:443-451` — nil-Guard für
  `DruckauftragRepo`: dokumentierte, bewusste Best-Effort-Entscheidung; Entfernen
  würde nur Testkonstruktionen brechen. Kein Defekt.

---

## Zu klärende Entscheidungen

1. C1 Reset-Endpunkt: Soll die Demo (jotti.rocks) den Reset weiter über HTTP
   nutzen? Optionen:
   a. Eigenes Flag `JOTTI_ENABLE_TEST_API`, nur in `docker-compose.e2e.yml`
      gesetzt; `JOTTI_ALLOW_SEED` bleibt für das CLI-Subkommando.
      `JOTTI_ALLOW_SEED` aus rocks/local entfernen. (Empfehlung, wenn die Demo
      keinen HTTP-Reset braucht.)
   b. Endpunkt auf der Demo behalten, aber hinter Shared-Secret-Token (wie das
      Relay) plus Rate-Limit. (Wenn der Demo-Reset per HTTP gewünscht ist.)
   c. Zusätzlich in beiden Fällen: `docker-compose.e2e.yml` auf `127.0.0.1`
      binden und den E2E-Reset-Pfad rate-limiten.
2. M2 Backfill: einmaliger automatischer Backfill der TSS-Seriennummer oder nur
   dokumentierter Re-Setup-Schritt? Hat die 2026-07-07-Instanz TSE eingerichtet?
3. Umfang dieser Runde: die 24 Minor + 4 Test-Quality-Fixes (N32-N44) sind
   risikoarm und ohne Deployment-Bezug, sofort machbar. M1 (Race) und M3/M4
   (ops-smoke) sind eigenständige Code-Fixes. Reihenfolge/Batching festlegen.

---

## Referenzen

- Workflow-Ergebnis (vollständige Rohbefunde, 66 dedupliziert): Run
  `wf_ac95e1c8-b81`.
- Range: `git diff v0.14.0..HEAD` (HEAD = 39ec13b, Tag v0.15.0).
