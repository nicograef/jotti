# Plan: Nacharbeit bis v1.0.0 (nicht breaking, nachziehbar)

> Übernimmt die Phasen 7–9 und 11–13 aus `plan-v0.14.0-vorabversion.md` sowie die restlichen C-Befunde. Nichts hiervon erzeugt Update-Aufwand für Betreiber: Backend, Frontend und Relay werden bei jedem Update komplett gemeinsam getauscht, persistente Daten (Schema, Event-JSON) werden nicht angefasst. Befund-Details im [Audit-Bericht](audit-v1.0.0.md); Status wird nur hier geführt. Arbeitsdokument, nach Abarbeitung aus `docs/plans/` entfernen.

## Prioritäten

Zwei Klassen:

1. **Vor dem v1.0.0-Tag** (friert mit dem Release ein, danach nur per Major-Bump änderbar): Block 1 (B4, B6). Für die installierte v0.14.0 ist das unkritisch, weil die einzigen API-Clients (eigenes Frontend, Print-Relay) mit ausgeliefert werden.
2. **Frei nachziehbar** (jederzeit als 0.14.x/1.0.x): alle übrigen Blöcke.

Die Blöcke sind unabhängig voneinander und einzeln ausführbar (eigener Implementierer, Review-Gate, `make verify`, Commit); eine feste Reihenfolge gibt es nicht. Block 6 (Release-Vorbereitung) kommt sinnvollerweise zuletzt.

## Autonome Ausführung (entschieden 2026-07-07)

Blöcke 1 bis 5 setzt ein Multi-Agent-Workflow in einem Lauf um. Block 6 ist ausdrücklich nicht Teil davon und folgt als eigene Session, sobald v0.14.0 getaggt ist und alle Änderungen vorliegen.

Vorbedingung: Start nur auf sauberem main. Der Breaking-Plan (`plan-v0.14.0-breaking.md`) ist vollständig committed, `git status` leer, `make verify` grün. Solange die parallele Session noch WIP im Tree hat, nicht starten.

Orchestrierung: Alle fünf Blöcke parallel implementieren und reviewen, jeder in einem eigenen isolierten git worktree (Implementierer plus Review-Gate mit Fix-Runden). Danach sequenzielle Integration auf main, ein Block nach dem anderen: rebase auf den aktuellen Stand, `make verify`, ein Commit pro Block (Conventional Commits, keine Co-Authored-By-Trailer). Im Integrations-Commit die Akzeptanz-Checkboxen des Blocks abhaken.

Fehler-Politik: Wird ein Block nach drei Fix-Runden nicht grün, wird er ausgelassen (Worktree zur Inspektion liegen lassen) und die übrigen Blöcke werden normal integriert. Der Abschlussbericht nennt ausgelassene Blöcke mit Diagnose.

Push: Nach der letzten Integration main nach origin pushen und das CI-Ergebnis abwarten; bei Rot nachbessern und erneut pushen. Das gilt insbesondere für das Block-4-Kriterium "neuer CI-Job läuft grün".

---

## Block 1: HTTP-API-Feinschliff (B4, B6, C3, D2) — vor dem v1.0.0-Tag

### Kontext

- `backend/api/middleware/middleware.go:206-244` — Auth-Fehler als 400, Rolle aus Token-Claim, `users.GetUser` wird bereits geladen
- `frontend/src/lib/Backend.ts:127` — Auto-Logout nur bei 401
- B6: `{"status":"ok"}`-Responses bei relay/beleg, gemischte Datumsformate, `details`-Freitexte
- Entschieden: 401 für `missing_authorization`, `invalid_authorization_format`, `invalid_jwt`, `user_inactive`; 403 für `insufficient_permissions`. Frontend-Auto-Logout bleibt an 401 gebunden.
- Entschieden (D2): Login-Fehlercodes `no_password_set`/`user_inactive` bleiben erhalten; die Abwägung (Verständlichkeit für nicht-technische Helfer wiegt schwerer als das Enumerationsrisiko, Login-Throttling existiert) wird in `docs/compliance.md` bzw. der Verfahrensdoku festgehalten.
- Entschieden: `details` bleibt als optionales Diagnosefeld erhalten. Strukturiert bei `validation_error` (zog-Issues) und beim Kassenabschluss-Gate (das Frontend liest beide aus, z. B. `KasseAbschliessenSection.tsx`), sonst kurze englische Diagnose. Vertrag als Godoc-Kommentar in `backend/api/helper/http.go` dokumentieren, Freitexte auf einheitlichen Stil bringen.

### Was zu bauen ist

Statuscodes gemäß Entscheidung (401/403); die Autorisierung prüft die Rolle aus dem bereits geladenen DB-User statt aus dem Token-Claim (C3), damit Rollenänderungen sofort wirken. API-Kosmetik: leere Erfolgs-Responses einheitlich `{}` (der Relay-Client wertet den Response-Body nicht aus, die Umstellung bei relay/beleg ist gefahrlos); Kalendertage als YYYY-MM-DD, Zeitpunkte als RFC3339; `details` gemäß Entscheidung dokumentieren und vereinheitlichen. Die D2-Abwägung dokumentieren.

### Akzeptanzkriterien

- [x] Middleware-Tests asserten 401 für fehlende/ungültige Tokens und inaktive User, 403 für fehlende Berechtigung
- [x] Abgelaufenes Token führt im Frontend zum Auto-Logout (Test)
- [x] Rollenwechsel wirkt beim nächsten Request, nicht erst nach Token-Ablauf (Test)
- [x] Erfolgs- und Datumsformate einheitlich; `details`-Entscheidung umgesetzt und dokumentiert
- [x] D2-Entscheidung dokumentiert; `make verify` grün

---

## Block 2: Backend-Robustheit (C2, C8)

### Kontext

- `backend/app/app.go:87-91` — Signatur-Worker, Watchdog, Rate-Limiter-Cleanup ohne recover()
- `backend/api/kasse/kassenfuehrung/application/command.go:334` — Teilfehler-Retry hängt zweites `kassensturz-durchgefuehrt:v1` an
- Der C9-Index ist in den Breaking-Plan (Phase 2) gewandert.

### Was zu bauen ist

Run-Loops bekommen defer/recover mit Log und Neustart (ein Panic stoppt die Signierung nicht dauerhaft); die HTTP-Chain bekommt eine Recovery-Middleware (Panic ergibt 500 statt Prozessabbruch). Der Kassenabschluss-Wiederanlauf erkennt einen bereits vorhandenen Kassensturz der Sitzung und überspringt Schritt 1 idempotent.

### Akzeptanzkriterien

- [x] Provozierter Panic im Worker: Signierung läuft nach Neustart des Loops weiter (Test)
- [x] Panic in einem Handler ergibt 500, Prozess lebt (Test)
- [x] Abschluss-Retry nach Teilfehler erzeugt keinen zweiten Kassensturz (Test gegen das Journal)
- [x] `make verify` grün

---

## Block 3: Frontend-Robustheit und UX (C1, C14)

### Kontext

- `frontend/src/service/table/hooks.ts:34` — Daten-Hooks verwerfen `isError`
- `frontend/src/main.tsx:12` — QueryClient ohne globales Error-Handling
- `frontend/src/admin/settings/DruckstationConfigPage.tsx:204` — Druckfehler nur auf der Unterseite sichtbar
- Der Backend-Endpunkt für das Dashboard-Banner existiert bereits (`POST /admin/get-fehlgeschlagene-druckauftraege`, `backend/api/druck/auftrag/http/handler.go`), das Banner ist reine Frontend-Arbeit.
- Entschieden: C15 (Kassenbeleg in einer Interaktion nach der Zahlung) entfällt ersatzlos. Belegdruck ist der seltene Ausnahmefall, der bestehende Weg über die Tisch-Historie bleibt (YAGNI).

### Was zu bauen ist

Globaler `QueryCache.onError`-Toast plus expliziter Fehlerzustand auf den kritischen Seiten (Tischseite, Kasse): bei 500/Netzabbruch erscheint ein Fehler statt der Leer-Defaults (Saldo 0,00, „Alles ausgegeben!"). Das Admin-Dashboard zeigt ein Banner bei fehlgeschlagenen Druckaufträgen (analog TSE-Warnung).

### Akzeptanzkriterien

- [ ] Fehlerzustand statt Leer-Default auf Tischseite/Kasse bei Query-Fehler (Tests)
- [ ] Banner auf dem Admin-Dashboard bei fehlgeschlagenen Druckaufträgen
- [ ] Frontend-Tests grün, Lint mit `--max-warnings=0`; `make verify` grün

---

## Block 4: CI-Härtung (C10, D13)

### Kontext

- `.github/workflows/ci.yml` — golangci-lint auf `latest`, kein Upgrade-Pfad-Job
- `scripts/test-integration.sh` — deckt `up` auf leerer DB bereits ab; Startup-Race beobachtet (v0.14.0-Session, 2026-07-07): `pg_isready` meldet grün, aber `migrate` scheitert mit „connection reset by peer" — das initdb-Restart-Fenster von `postgres:17`; ein Wiederholungslauf war sauber
- Release-Guide Gate 4 (b): Migration auf befüllter Vorversions-DB plus Boot plus `rebuild-projections`

### Was zu bauen ist

Neuer CI-Job als Upgrade-Harness: DB befüllen (Seed-Daten), App booten, `rebuild-projections` laufen lassen. Die Vorversion ist parametrisiert; nach dem Freeze durch die Erstinstallation wird sie auf das letzte Release gepinnt (ab der ersten `02_`-Migration Pflicht-Gate). Dazu D13: golangci-lint auf feste Version pinnen; `go mod tidy -diff` und `-count=1` zwischen CI und `make verify` angleichen, damit beide deckungsgleich grün sind. Außerdem `scripts/test-integration.sh` gegen das Postgres-Startup-Race härten: nicht nur `pg_isready` abwarten, sondern eine echte Verbindung (z. B. `SELECT 1` in Schleife) vor dem `migrate`-Aufruf, damit das initdb-Restart-Fenster keine Flakes erzeugt.

### Akzeptanzkriterien

- [ ] Neuer CI-Job läuft grün (Seed, Boot, rebuild-projections); Pinning-Plan im Job oder Migrations-README dokumentiert
- [ ] golangci-lint-Version gepinnt
- [ ] CI-Checks und `make verify` prüfen dieselben Dinge mit denselben Flags
- [ ] `test-integration.sh` wartet auf eine echte DB-Verbindung statt nur `pg_isready` (Startup-Race behoben)
- [ ] `make verify` grün

---

## Block 5: Doku-Konsistenz (C12, C16–C20, C23; Gate 5)

### Kontext

- `docs/verfahrensdokumentation.md:167` — behauptet Versions-Ausweisung in Admin-UI und auf dem Beleg
- `frontend/src/admin/AdminSidebar.tsx:160-172` — Footer ohne Version
- `docs/leitfaden/finanzamt-anmelden.md`, `docs/leitfaden/checkliste.md` — Kassenmeldepflicht-Frist fehlt (C16); Checkliste mischt Betriebspfade (C18, `checkliste.md:27`)
- `docs/leitfaden/fehlersuche.md`, `docs/leitfaden/haeufige-fragen.md` — TSE-/Internet-Ausfall-Botschaft fehlt (C17)
- `README.md:21` — bewirbt entfallenes Auszahlung-Feature (C19)
- `AGENTS.md:59-64` — v0-Praxis und gelöschte `01_initial.down.sql` (C20; durch die Erstinstallation ist die Freeze-Disziplin de facto schon in Kraft, das macht diesen Punkt dringlicher als die übrigen Doku-Punkte)
- fiskaly-Preisangabe „ca. 8 EUR/Monat" in drei Leitfäden (C23)
- Gate 5: `compliance.md`, `handbuch.md`, `anforderungen.md`, `language.md` nach F-12-Streichung querlesen

### Was zu bauen ist

Admin-Footer zeigt `jotti <version>` aus `/health`; der Beleg-Satz in der Verfahrensdoku wird gestrichen (C12). Ein-Monats-Frist nach § 146a Abs. 4 AO ergänzen (C16, Beleg: BMF 28.06.2024 in `docs/rechtsquellen/`); Laien-Absatz „weiterverkaufen erlaubt, jotti signiert nach" (C17); Checkliste nach Betriebspfad Server/Windows trennen (C18); Auszahlung aus dem README streichen (C19); AGENTS.md auf Freeze-Disziplin und additive `NN_*.up.sql` umschreiben, Verweis auf die gelöschte Down-Migration entfernen (C20); fiskaly-Preis durch „aktuellen Preis bei fiskaly erfragen" ersetzen (C23). Abschließend Gate-5-Querlesen der vier Kern-Dokumente gegen den ausgelieferten Funktionsumfang. Stil: minimal formatiert, keine Slop-Syntax.

### Akzeptanzkriterien

- [ ] Version im Admin-Footer sichtbar; Verfahrensdoku-Aussage deckt sich mit dem Code
- [ ] C16–C19, C23 umgesetzt; Rechtsaussagen mit Beleg aus `docs/rechtsquellen/`
- [ ] AGENTS.md beschreibt die geltende Migrations-Disziplin
- [ ] Kern-Dokumente konsistent mit dem Funktionsumfang (F-12 als Nicht-Ziel überall konsistent)
- [ ] `make verify` grün

---

## Block 6: Release-Vorbereitung v1.0.0 (C13, C21; Gate-1-Rest) — nicht Teil des autonomen Laufs

### Kontext

- Eigene Session nach dem v0.14.0-Tag (siehe Autonome Ausführung): Changelog-Basis und der finale Review-Pass brauchen den vollständigen Stand inklusive der Blöcke 1–5; TODO/FIXME-Entscheidungen bleiben Ermessenssache.

- `cliff.toml` — Changelog-Generierung vorhanden, bisher nur flüchtige Release-Notes (C21)
- `docs/leitfaden/self-hosting.md:39`, `.env.example:21`, `docker-compose.release.yml`, `docs/verfahrensdokumentation.md` — Beispiel-/Pin-Versionen (C13)
- Die Gate-6-Image-Namen im Release-Guide sind bereits korrigiert (Umstrukturierung 2026-07-06); von C13 bleibt: Beispiel-Versionen anheben, `frontend/package.json` als tot dokumentieren und aus dem Bump-Set streichen.
- Release-Guide Gate 1 — TODO/FIXME-Grep, voller Verify, Cleanup

### Was zu bauen ist

`CHANGELOG.md` anlegen (cliff-basiert) mit Einträgen ab v0.14.0 und dem Funktionsumfang-Stand für 1.0.0. Beispiel-Versionen in Doku und `.env.example`/Release-Compose anheben; `frontend/package.json` als tot dokumentieren. Gate-1-Rest: Grep auf `TODO`/`FIXME`/`XXX` in fiskalisch relevanten Modulen (Befunde beheben oder bewusst dokumentieren), voller `make verify` plus `make lint-backend-full`, finaler Cleanup-/Review-Pass über die seit dem Audit geänderten Bereiche.

### Akzeptanzkriterien

- [ ] `CHANGELOG.md` existiert und wird ab 1.0.0 gepflegt
- [ ] Versionsstellen konsistent; `frontend/package.json` als tot dokumentiert
- [ ] Kein unbegründetes TODO/FIXME/XXX in fiskalischen Modulen
- [ ] Voller `make verify` und `make lint-backend-full` grün

---

## Bewusst liegen gelassen (nach 1.0 ohne Bruch nachrüstbar)

Aus dem Audit, Abschnitt D — dokumentiert, bewusst kein Teil dieses Plans: D1 (api_secret at rest), D3 (Container-Härtung), D4 (Resolver-Einschränkung), D5 (JWT-localStorage-Abwägung dokumentieren), D6 (React-Query-Tuning), D7 (Export-Timeout/RAM), D8 (Storno-Beleg-Zeitstempel), D9 (Doppeldruck-Semantik dokumentieren), D11 (rebuild-projections-Lücke dokumentieren), D12 (Steuersatz-Konstanten-Warnung), D14 (Windows-Dump-Integrität), D16 (OTP in Container-Logs).
