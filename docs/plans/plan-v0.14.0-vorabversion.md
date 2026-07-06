# Plan: v0.14.0 Vorabversion (automatisierbare Audit- und Release-Fixes)

> Quellen: [Audit-Bericht v1.0.0](audit-v1.0.0.md) und [Release-Guide v1.0.0](plan-v1.0.0-release.md), Stand 2026-07-06.

## Ziel

Alle Befunde und Release-Vorbereitungen umsetzen, die von Coding-Agenten automatisiert korrigiert, implementiert und reviewt werden können. Ergebnis ist ein release-fertiger `main` für eine Vorabversion `v0.14.0`: alle A-Blocker gefixt, alle Vor-Freeze-Entscheidungen (B) umgesetzt, C-Empfehlungen abgearbeitet, freeze-relevante D-Punkte erledigt, Doku konsistent, Versionsstellen gebumpt, CHANGELOG angelegt.

Bewusst nicht Teil dieses Plans (bleibt manuell, für v1.0.0): Abschnitt F des Audits komplett, Gate 2 (Fiskal-E2E mit fiskaly-TEST-TSE), Gate 3 (Test-Server-Roundtrips), C7 (Latenzmessung braucht E2E), die Hardware-Verifikation zu A5 (Gate 2.4), Tag-Push und Veröffentlichung.

## Ausführungsmodell (autonome Multi-Agent-Session)

- Phasen strikt **sequenziell** abarbeiten (Phasen teilen sich Dateien: `01_initial.up.sql`, Golden-Tests, Frontend-Schemas); keine parallelen Implementierer auf demselben Working Tree.
- Pro Phase: frischer Implementierer-Subagent, danach unabhängiges Review-Gate, eigener `make verify`-Lauf, eigener Commit (Conventional Commits auf Englisch, keine AI-Trailer).
- Direkt-Commits auf `main` sind für diese Session ausdrücklich autorisiert (Ausnahme von der No-Auto-Commit-Regel); die globale Regel gilt danach unverändert weiter.
- Nach jeder Phase die erledigten Checkboxen hier **und** in `audit-v1.0.0.md` abhaken.
- Session endet mit release-fertigem `main`; das Tag `v0.14.0` setzt und pusht Nico selbst (vorbereiteter Befehl am Ende von Phase 13).

## Architektur-Entscheidungen

Gelten über alle Phasen hinweg:

- **Version:** `v0.14.0` (nächstes v0-Minor nach `v0.13.1`). Die v0-Semantik (Breaking Changes erlaubt) bleibt bis zum echten `v1.0.0`-Tag bestehen; Schema-Änderungen dieses Plans gehen noch direkt in `01_initial.up.sql`.
- **Idempotenz (B2):** buchende Vorgänge tragen eine clientseitig erzeugte ID (`crypto.randomUUID()`), die im Event persistiert wird; Deduplizierung über partiellen UNIQUE-Index; ein Duplikat antwortet idempotent erfolgreich (kein Fehler für den Client). Direktverkauf verpflichtend, Bestellung/Geldtransit über optionalen Idempotenz-Schlüssel.
- **Auth-Statuscodes (B4):** 401 für `missing_authorization`, `invalid_authorization_format`, `invalid_jwt`, `user_inactive`; 403 für `insufficient_permissions`. Frontend-Auto-Logout bleibt an 401 gebunden.
- **Seed-Schutz (A3):** `jotti seed` läuft nur mit explizitem Opt-in `JOTTI_ALLOW_SEED=1` (im Dev-Compose gesetzt, in Prod nie); der bestehende Kassenjournal-Guard bleibt als zweite Schicht.
- **Version-Single-Source:** die per ldflags eingebrannte Build-Version ist die einzige Quelle; sie fließt in `cashregister.csv` (C11), den Admin-Footer (C12) und `/health`. `frontend/package.json` bleibt tot und wird aus dem Bump-Set gestrichen (C13).

## Inventar

Die Datei- und Zeilenreferenzen stammen aus dem Audit vom 2026-07-06 (HEAD unverändert, stichprobenhaft verifiziert). Zentrale Orte:

- `backend/api/druck/bondruck/application/escpos/` — Beleg-Formatter und ESC/POS-Konstanten (A1, A5)
- `backend/config/config.go`, `scripts/prod-init.sh`, `backend/seed/writer.go` — Erstinbetriebnahme (A2, A3, C22)
- `backend/api/fiskal/dsfinvk/mapper.go`, `backend/api/fiskal/setup/application/setup.go` — DSFinV-K und TSE-Stammdaten (A4, C6, C11, D15)
- `backend/domain/kasse/` — Events und Contract-Guard (B1, B3, D10)
- `database/migrations/01_initial.up.sql` — Schema (B5, B2-Index, C9)
- `backend/api/middleware/middleware.go`, `frontend/src/lib/Backend.ts` — Auth-Fehlerpfad (B4, C3, D2)
- `backend/app/app.go`, `backend/api/kasse/kassenfuehrung/application/command.go` — Robustheit (C2, C8)
- `frontend/src/` — Query-Fehler, Banner, Beleg-Zugriff (C1, C14, C15)
- `Makefile`, `docker-compose.prod.yml`, `windows/starter/` — Update-Pfade (C4, C5)
- `.github/workflows/ci.yml`, `scripts/test-integration.sh` — CI (C10, D13)
- `docs/`, `README.md`, `AGENTS.md`, `website/` — Doku-Konsistenz (C12, C16–C23, Gate 5)

## Getroffene Entscheidungen

Aus der Klärungsrunde mit Nico (2026-07-06):

- Vorabversion heißt `v0.14.0`, nicht `v1.0.0-beta`/`-rc` (Freeze-Semantik greift erst mit v1.0.0).
- D-Scope: nur die freeze-relevanten D2, D10, D13, D15; D1, D3–D9, D11–D12, D14, D16 bleiben dokumentiert liegen.
- Session-Endpunkt: release-fertiger `main` ohne Tag-Push (Außenwirkung bleibt bei Nico).
- B2 wird komplett nach Audit-Empfehlung umgesetzt (nicht nur Direktverkauf).

Vom Planersteller nach Audit-Empfehlung aufgelöst (Veto vor Session-Start möglich):

- **A3:** Opt-in-Env-Flag statt Build-Tag-Ausschluss; hält den Dev-Flow (`make seed` nach Bootstrap) intakt, der ein reiner CountUsers==0-Guard brechen würde.
- **C12:** Variante „Footer zeigt `jotti <version>` via `/health`, Beleg-Aussage in der Verfahrensdoku streichen" (statt nur den Satz zu korrigieren).
- **C13:** `frontend/package.json` wird nicht verdrahtet, sondern als tot dokumentiert und aus dem Gate-6-Bump-Set gestrichen.
- **D2:** Login-Fehlercodes `no_password_set`/`user_inactive` bleiben erhalten; die Abwägung (Verständlichkeit für nicht-technische Helfer im Vereins-LAN wiegt schwerer als das Enumerationsrisiko, Login-Throttling existiert) wird in `docs/compliance.md` bzw. der Verfahrensdoku festgehalten. Damit ist die vom Audit geforderte bewusste Entscheidung getroffen.
- **Gate 1 `/code-audit`:** durch das Multi-Experten-Audit vom 2026-07-06 erfüllt; Phase 13 macht nur noch einen finalen Cleanup-/Review-Pass über die geänderten Bereiche.

## Offene Punkte / Risiken

- **C10-Harness:** Eine echte „Vorversions-DB" gibt es erst nach dem Freeze; solange `01_initial.up.sql` editiert wird, kann der Job nur gegen den aktuellen Stand mit Seed-Daten laufen (Details in Phase 11).
- **A5:** Der Codefix ist automatisierbar, der Beweis auf echter 80-mm-Hardware nicht; Gate 2.4 bleibt nach diesem Plan offen.
- **B2/C9-Indexe:** `kassenjournal` ist append-only mit Owner-festem Trigger; Indexe sind unkritisch (kein UPDATE/DELETE), aber der Implementierer muss die REVOKE/GRANT-Struktur unangetastet lassen.
- **Golden-Tests werden mehrfach angefasst** (Phase 1, 3, 4); jede Phase hinterlässt sie grün, die Reihenfolge ist deshalb unkritisch, aber nicht umsortierbar ohne Nachdenken.

---

## Phase 1: Beleg-Blocker: Steuersatz ausweisen, QR-Modulgröße dynamisch (A1, A5)

### Kontext

- `backend/api/druck/bondruck/application/escpos/formatter.go:237,250-259,393-406` — Steuermatrix druckt nur Kennzeichen A/B/C ohne Prozentsatz
- `backend/api/druck/bondruck/application/escpos/constants.go:30` — feste `QRCodeModuleSize6`
- `backend/api/druck/bondruck/application/escpos/formatter.go:313` — QR-Aufbau
- `docs/rechtsquellen/` — § 6 Satz 1 Nr. 5 KassenSichV (Steuersatz bzw. Befreiungshinweis auf dem Beleg)

### Was zu bauen ist

Die Steuermatrix weist je Kennzeichen den Prozentsatz aus (`A (19 %): …` oder Legende `A = 19 %, B = 7 %, C = 0 % (umsatzsteuerfrei)`) inklusive Befreiungshinweis für den 0-%-Satz. Die QR-Modulgröße wird längenabhängig gewählt, sodass reale fiskaly-Payloads (~350–470 Byte) inklusive Ruhezone innerhalb der druckbaren 576 Dots bleiben. Golden-Tests anpassen; ein Test rechnet die QR-Breite für eine realistische Payloadlänge nach.

### Akzeptanzkriterien

- [x] Beleg zeigt je verwendetem Steuerkennzeichen den Prozentsatz; 0 % trägt den Befreiungshinweis (Formulierung gegen `docs/rechtsquellen/` bzw. `docs/steuerrecht.md` geprüft)
- [x] QR-Code bleibt für Payloads bis mindestens 500 Byte innerhalb 576 Dots (Test belegt die Rechnung)
- [x] Golden-Tests aktualisiert, `make verify` grün
- [x] Vermerk: Sichtprüfung auf echter Hardware (Gate 2.4) bleibt offen

---

## Phase 2: Erstinbetriebnahme-Härtung (A2, A3, C22)

### Kontext

- `backend/config/config.go:34-44` — Secrets nur auf nicht-leer geprüft, `POSTGRES_PASSWORD` defaultet auf `admin`
- `scripts/prod-init.sh:90-103` — prüft nur Domain/Email/Version, keine Secrets
- `scripts/prod-init.sh:203-204` — druckt nur den grep-Hinweis statt des OTP
- `backend/seed/writer.go:60-66`, `backend/main.go:62` — Seed-Guard nur über Kassenjournal-Count
- Marker-Präfix `ADMIN-EINMALPASSWORT` ist modulübergreifender Vertrag (bootstrap, starter/core, prod-init); Windows-Starter greppt bereits selbst

### Was zu bauen ist

Startup-Validierung: JWT_SECRET und RELAY_AUTH_TOKEN brauchen eine Mindestlänge, die bekannten `.env.example`-Platzhalter werden fatal abgelehnt, POSTGRES_PASSWORD hat keinen Default mehr. `prod-init.sh` prüft die Secrets ebenfalls und greppt das Admin-OTP selbst aus den Logs (analog Windows-Starter, ANSI-tolerant per Substring/Regex, nie bare Zeilen). Der Seed-Befehl verlangt zusätzlich `JOTTI_ALLOW_SEED=1`; das Dev-Compose setzt das Flag, Prod-Compose nicht.

### Akzeptanzkriterien

- [x] Backend startet nicht mit Platzhalter- oder Kurz-Secrets; Fehlermeldung nennt die betroffene Variable
- [x] Backend startet nicht ohne explizites POSTGRES_PASSWORD
- [x] `prod-init.sh` bricht bei schwachen Secrets ab und gibt am Ende den OTP-Code aus
- [x] `jotti seed` ohne `JOTTI_ALLOW_SEED=1` verweigert; `make seed` im Dev-Flow funktioniert weiter
- [x] Tests für die Validierungsregeln; `make verify` grün

---

## Phase 3: DSFinV-K-Fixes (A4, C6, C11, D15)

### Kontext

- `backend/api/fiskal/dsfinvk/mapper.go:427-436,713-715` — AVSonstige mit leerem BON_NAME
- `backend/api/fiskal/setup/application/setup.go:346` — TSE-Stammdaten-Abruf einmalig und best-effort
- `backend/api/fiskal/dsfinvk/mapper.go:1235` — TSE_SERIAL aus der ersten Signatur statt aus Stammdaten
- `backend/api/fiskal/dsfinvk/mapper.go:45,600` — hartkodierte `KASSE_SW_VERSION`-Konstante `"1.0"`, tautologischer Test
- DSFinV-K 2.4 (S. 46, 84) unter `docs/rechtsquellen/technik-spezifikationen/DSFinV-K-2.4/`

### Was zu bauen ist

`buildTransactions` füllt BON_NAME beim Tagesabschluss (konstanter Text „Tagesabschluss"). Der TSE-Stammdaten-Abruf wird hart: Fehlschlag ist ein Setup-Fehler mit Retry-Pfad statt Warn-Log; TSE_SERIAL kommt aus den persistierten Stammdaten. Die Software-Version wird aus der ldflags-Build-Version in den DSFinV-K-Snapshot durchgereicht; ein Guard-Test vergleicht gegen die Build-Version statt gegen dieselbe Konstante. Die accuracy-Metadaten werden auf den amtlichen Wert angeglichen oder das tote Feld entfernt (D15).

### Akzeptanzkriterien

- [ ] Jede AVSonstige-Zeile trägt einen BON_NAME (Golden Rows angepasst)
- [ ] Fehlgeschlagener Stammdaten-Abruf lässt das TSE-Setup nicht stillschweigend unvollständig; `tse.csv`-Pflichtfelder kommen aus persistierten Stammdaten
- [ ] `cashregister.csv` enthält die Build-Version; Guard-Test schlägt bei Drift fehl
- [ ] D15 entschieden und umgesetzt (angleichen oder entfernen)
- [ ] `make verify` grün

---

## Phase 4: Event-Contract vor dem Freeze (B1, B3, D10)

### Kontext

- `backend/domain/kasse/bestellung.go:45` — JSON-Key `einzelpreis` ohne Cents-Suffix
- `backend/domain/kasse/event_json_contract_test.go:40-57` — Round-Trip über dieselben Structs, `bestellung-korrigiert:v1` fehlt
- `docs/language.md` — verbindliche Namenskonvention (Cents-Suffix)
- SQL-Konsumenten der Event-Keys (u. a. `kj_extract_umsatz_pro_steuersatz`) und Frontend-Zod-Schemas

### Was zu bauen ist

Rename `einzelpreis` zu `einzelpreisCents` durchgängig: Event-Data-Structs, SQL-Extraktoren, Response-DTOs, Frontend-Zod-Schemas, `docs/language.md` (kein Dual-Read, v0-Politik). Der Contract-Guard wird auf fixe JSON-Literale umgestellt: je Event-Typ ein eingefrorenes JSON-Beispiel, das unmarshalt und feldweise assertet wird; `bestellung-korrigiert:v1` sowie Tagesabschluss/Kassensturz und alle `*Id`-Felder werden gepinnt; EventType-Konstanten werden gegen die SQL-Literale geprüft; ein Meta-Test enumeriert alle Event-Typen. Wertkonstanten `einlage`/`entnahme` kommen in den Contract-Test (D10).

### Akzeptanzkriterien

- [ ] Kein JSON-Key `"einzelpreis"` mehr im Repo (nur `einzelpreisCents`), inklusive SQL und Frontend
- [ ] Contract-Test basiert auf fixen JSON-Literalen und würde einen Tag-Rename bemerken
- [ ] Alle Kassenjournal-Event-Typen sind abgedeckt, inklusive `bestellung-korrigiert:v1`; Meta-Test schlägt bei neuem, ungepinntem Event-Typ fehl
- [ ] `einlage`/`entnahme` als Wertkonstanten gepinnt
- [ ] `make verify` grün (inklusive Replay/rebuild-projections in den Integrationstests)

---

## Phase 5: Schema-Feinschliff (B5)

### Kontext

- `database/migrations/01_initial.up.sql` — letzter Edit vor dem Freeze
- `database/migrations/README.md` — Migrations-Konvention
- `backend/sqlc/` — nach Typänderungen `make sqlc`

### Was zu bauen ist

CHECK-Constraints auf Geldspalten (`saldo_cents >= 0`, `gesamt_zahlungen_cents >= 0`, `preis_cents >= 0`); `kassenjournal.id` und `kassensitzungen.z_nr` auf GENERATED ALWAYS AS IDENTITY; `tse_signaturauftraege.transaktion_nummer` und `signatur_zaehler` auf BIGINT; die ENUM-vs-TEXT+CHECK-Entscheidung wird in `database/migrations/README.md` begründet festgehalten (inklusive Zwei-Migrations-Muster für spätere ENUM-Erweiterungen).

### Akzeptanzkriterien

- [ ] Constraints und Typen wie oben in `01_initial.up.sql`; sqlc regeneriert, Go-Typen konsistent (int64 für BIGINT)
- [ ] Integrationstests auf frischer DB grün; `make rebuild-projections` läuft fehlerfrei
- [ ] ENUM-Entscheidung im Migrations-README dokumentiert
- [ ] `make verify` grün

---

## Phase 6: Idempotenz buchender Endpunkte (B2)

### Kontext

- `backend/api/kasse/direktverkauf/application/command.go:88` — frische Server-UUID je Request
- `frontend/src/lib/Backend.ts:110` — zentrale Request-Schicht (Retry-/Doppel-Tap-Pfad)
- `database/migrations/01_initial.up.sql` — Dedup-Index; Append-only-Trigger und REVOKE/GRANT unangetastet lassen

### Was zu bauen ist

Der Direktverkauf bekommt eine clientseitig erzeugte `verkaufId` (UUID) im Request, die validiert, ins Event geschrieben und über einen partiellen UNIQUE-Index dedupliziert wird; ein Duplikat liefert eine idempotente Erfolgs-Antwort. Bestellung und Geldtransit erhalten einen optionalen Idempotenz-Schlüssel mit derselben Mechanik. Frontend erzeugt die IDs pro logischem Vorgang (nicht pro Retry).

### Akzeptanzkriterien

- [ ] Zwei identische Direktverkauf-Requests erzeugen genau ein Event; die zweite Antwort ist erfolgreich und referenziert den ersten Vorgang (Integrationstest)
- [ ] Gleiche Semantik für Bestellung/Geldtransit mit gesetztem Schlüssel; ohne Schlüssel Verhalten wie heute
- [ ] Frontend sendet die IDs; Doppel-Submit-Schutz bleibt zusätzlich bestehen
- [ ] Event-Contract-Test aus Phase 4 um die neuen Felder ergänzt
- [ ] `make verify` grün

---

## Phase 7: HTTP-API-Feinschliff (B4, B6, C3, D2)

### Kontext

- `backend/api/middleware/middleware.go:206-244` — Auth-Fehler als 400, Rolle aus Token-Claim, `users.GetUser` wird bereits geladen
- `frontend/src/lib/Backend.ts:127` — Auto-Logout nur bei 401
- B6: `{"status":"ok"}`-Responses bei relay/beleg, gemischte Datumsformate, `details`-Freitexte

### Was zu bauen ist

Statuscodes gemäß Architektur-Entscheidung (401/403); die Autorisierung prüft die Rolle aus dem bereits geladenen DB-User statt aus dem Token-Claim (C3), damit Rollenänderungen sofort wirken. API-Kosmetik: leere Erfolgs-Responses einheitlich `{}`; Kalendertage als YYYY-MM-DD, Zeitpunkte als RFC3339; `details` entweder streichen oder als englisches Diagnosefeld dokumentieren. Die D2-Abwägung (Login-Fehlercodes bleiben) wird in der Compliance-/Verfahrensdoku festgehalten.

### Akzeptanzkriterien

- [ ] Middleware-Tests asserten 401 für fehlende/ungültige Tokens und inaktive User, 403 für fehlende Berechtigung
- [ ] Abgelaufenes Token führt im Frontend zum Auto-Logout (Test)
- [ ] Rollenwechsel wirkt beim nächsten Request, nicht erst nach Token-Ablauf (Test)
- [ ] Erfolgs- und Datumsformate einheitlich; `details`-Entscheidung umgesetzt und dokumentiert
- [ ] D2-Entscheidung dokumentiert; `make verify` grün

---

## Phase 8: Backend-Robustheit (C2, C8, C9)

### Kontext

- `backend/app/app.go:87-91` — Signatur-Worker, Watchdog, Rate-Limiter-Cleanup ohne recover()
- `backend/api/kasse/kassenfuehrung/application/command.go:334` — Teilfehler-Retry hängt zweites `kassensturz-durchgefuehrt:v1` an
- `backend/sqlc/queries/tse_signaturauftraege.sql:78-85` — Full Scan der nie geleerten Tabelle

### Was zu bauen ist

Run-Loops bekommen defer/recover mit Log und Neustart (ein Panic stoppt die Signierung nicht dauerhaft); die HTTP-Chain bekommt eine Recovery-Middleware (Panic ergibt 500 statt Prozessabbruch). Der Kassenabschluss-Wiederanlauf erkennt einen bereits vorhandenen Kassensturz der Sitzung und überspringt Schritt 1 idempotent. Partieller Index `(erledigt_am) WHERE status = 'erledigt'` für das Queue-Monitoring.

### Akzeptanzkriterien

- [ ] Provozierter Panic im Worker: Signierung läuft nach Neustart des Loops weiter (Test)
- [ ] Panic in einem Handler ergibt 500, Prozess lebt (Test)
- [ ] Abschluss-Retry nach Teilfehler erzeugt keinen zweiten Kassensturz (Test gegen das Journal)
- [ ] Index in `01_initial.up.sql`; `make verify` grün

---

## Phase 9: Frontend-Robustheit und UX (C1, C14, C15)

### Kontext

- `frontend/src/service/table/hooks.ts:34` — Daten-Hooks verwerfen `isError`
- `frontend/src/main.tsx:12` — QueryClient ohne globales Error-Handling
- `frontend/src/admin/settings/DruckstationConfigPage.tsx:204` — Druckfehler nur auf der Unterseite sichtbar
- `frontend/src/service/components/table/ZahlungDrawer.tsx:66` — Beleg erst über 4 Interaktionen erreichbar

### Was zu bauen ist

Globaler `QueryCache.onError`-Toast plus expliziter Fehlerzustand auf den kritischen Seiten (Tischseite, Kasse): bei 500/Netzabbruch erscheint ein Fehler statt der Leer-Defaults (Saldo 0,00, „Alles ausgegeben!"). Das Admin-Dashboard zeigt ein Banner bei fehlgeschlagenen Druckaufträgen (analog TSE-Warnung). Nach dem Kassieren ist der Kassenbeleg mit einer Interaktion erreichbar (Aktion im Erfolgs-Toast oder Drawer).

### Akzeptanzkriterien

- [ ] Fehlerzustand statt Leer-Default auf Tischseite/Kasse bei Query-Fehler (Tests)
- [ ] Banner auf dem Admin-Dashboard bei fehlgeschlagenen Druckaufträgen
- [ ] Beleg-Druck in einer Interaktion nach der Zahlung erreichbar
- [ ] Frontend-Tests grün, Lint mit `--max-warnings=0`; `make verify` grün

---

## Phase 10: Update-Pfade entschärfen (C4, C5)

### Kontext

- `Makefile:156-158` — `prod-up` zieht Images und migriert ohne Backup/Downgrade-Guard
- `docker-compose.prod.yml` — `:-latest`-Fallback
- `windows/starter/update.go:27`, `windows/starter/main.go:100-112` — Starter startet ältere Exe gegen neuere Daten
- `scripts/prod-update.sh` — Referenz-`is_downgrade`-Logik

### Was zu bauen ist

`prod-up` wird als reiner Start-/Neustart-Weg positioniert: kein stilles Update mehr; ohne gesetzte Version bricht es ab statt `latest` zu ziehen; Make-Help und Doku nennen `prod-update` als einzigen Update-Weg. Der Windows-Starter spiegelt die `is_downgrade`-Sperre aus `prod-update.sh` und verweigert den Start einer älteren Version gegen neuere Daten.

### Akzeptanzkriterien

- [ ] `prod-up` ohne gepinnte Version schlägt mit klarer Meldung fehl; kein `:-latest` mehr im Prod-Compose
- [ ] Doku/Help: `prod-update` ist der einzige beworbene Update-Weg
- [ ] Starter-Downgrade-Test (ältere Version gegen neuere Datenversion wird verweigert)
- [ ] `make verify` grün

---

## Phase 11: CI-Härtung (C10, D13)

### Kontext

- `.github/workflows/ci.yml` — golangci-lint auf `latest`, kein Upgrade-Pfad-Job
- `scripts/test-integration.sh` — deckt `up` auf leerer DB bereits ab
- Release-Guide Gate 4 (b): Migration auf befüllter Vorversions-DB plus Boot plus `rebuild-projections`

### Was zu bauen ist

Neuer CI-Job als Upgrade-Harness: DB befüllen (Seed-Daten), App booten, `rebuild-projections` laufen lassen. Die Vorversion ist parametrisiert; solange `01_initial.up.sql` noch editiert wird (v0), läuft der Job gegen den aktuellen Stand und dokumentiert, dass die Vorversion beim v1.0.0-Tag auf das letzte Release gepinnt wird (ab der ersten `02_`-Migration Pflicht-Gate). Dazu D13: golangci-lint auf feste Version pinnen; `go mod tidy -diff` und `-count=1` zwischen CI und `make verify` angleichen, damit beide deckungsgleich grün sind.

### Akzeptanzkriterien

- [ ] Neuer CI-Job läuft grün (Seed, Boot, rebuild-projections); Pinning-Plan im Job oder Migrations-README dokumentiert
- [ ] golangci-lint-Version gepinnt
- [ ] CI-Checks und `make verify` prüfen dieselben Dinge mit denselben Flags
- [ ] `make verify` grün

---

## Phase 12: Doku-Konsistenz (C12, C16–C20, C23; Gate 5)

### Kontext

- `docs/verfahrensdokumentation.md:167` — behauptet Versions-Ausweisung in Admin-UI und auf dem Beleg
- `frontend/src/admin/AdminSidebar.tsx:160-172` — Footer ohne Version
- `docs/leitfaden/finanzamt-anmelden.md`, `docs/leitfaden/checkliste.md` — Kassenmeldepflicht-Frist fehlt (C16); Checkliste mischt Betriebspfade (C18, `checkliste.md:27`)
- `docs/leitfaden/fehlersuche.md`, `docs/leitfaden/haeufige-fragen.md` — TSE-/Internet-Ausfall-Botschaft fehlt (C17)
- `README.md:21` — bewirbt entfallenes Auszahlung-Feature (C19)
- `AGENTS.md:59-64` — v0-Praxis und gelöschte `01_initial.down.sql` (C20)
- fiskaly-Preisangabe „ca. 8 EUR/Monat" in drei Leitfäden (C23)
- Gate 5: `compliance.md`, `handbuch.md`, `anforderungen.md`, `language.md` nach F-12-Streichung querlesen

### Was zu bauen ist

Admin-Footer zeigt `jotti <version>` aus `/health`; der Beleg-Satz in der Verfahrensdoku wird gestrichen (C12). Ein-Monats-Frist nach § 146a Abs. 4 AO ergänzen (C16, Beleg: BMF 28.06.2024 in `docs/rechtsquellen/`); Laien-Absatz „weiterverkaufen erlaubt, jotti signiert nach" (C17); Checkliste nach Betriebspfad Server/Windows trennen (C18); Auszahlung aus dem README streichen (C19); AGENTS.md auf Freeze-Disziplin und additive `NN_*.up.sql` umschreiben, Verweis auf die gelöschte Down-Migration entfernen (C20); fiskaly-Preis durch „aktuellen Preis bei fiskaly erfragen" ersetzen (C23). Abschließend Gate-5-Querlesen der vier Kern-Dokumente gegen den ausgelieferten Funktionsumfang. Stil: minimal formatiert, keine Slop-Syntax.

### Akzeptanzkriterien

- [ ] Version im Admin-Footer sichtbar; Verfahrensdoku-Aussage deckt sich mit dem Code
- [ ] C16–C19, C23 umgesetzt; Rechtsaussagen mit Beleg aus `docs/rechtsquellen/`
- [ ] AGENTS.md beschreibt die ab v1.0.0 geltende Migrations-Disziplin (Freeze de facto ab diesem Commit)
- [ ] Kern-Dokumente konsistent mit dem Funktionsumfang (F-12 als Nicht-Ziel überall konsistent)
- [ ] `make verify` grün

---

## Phase 13: Release-Vorbereitung v0.14.0 (C13, C21; Gate 1 und 6, soweit automatisierbar)

### Kontext

- `cliff.toml` — Changelog-Generierung vorhanden, bisher nur flüchtige Release-Notes (C21)
- `docs/leitfaden/self-hosting.md:39` (v0.2.0), `.env.example:21`, `docker-compose.release.yml`, `docs/verfahrensdokumentation.md` — Beispiel-/Pin-Versionen (C13)
- `docs/plans/plan-v1.0.0-release.md` Gate 6 — nennt das nicht existierende Image `ghcr.io/nicograef/jotti` (real: `jotti-backend`/`-frontend`/`-migrate`/`-reverse-proxy`)
- Release-Guide Gate 1 — TODO/FIXME-Grep, voller Verify, Cleanup

### Was zu bauen ist

`CHANGELOG.md` anlegen (cliff-basiert) mit v0.14.0-Eintrag und dem Funktionsumfang-Stand. Versionsstellen auf `v0.14.0` heben: Beispiel-Versionen in Doku und `.env.example`/Release-Compose; `frontend/package.json` als tot dokumentieren und aus dem Gate-6-Bump-Set des Release-Guides streichen; Gate-6-Image-Namen im Release-Guide korrigieren. Gate-1-Rest: Grep auf `TODO`/`FIXME`/`XXX` in fiskalisch relevanten Modulen (Befunde beheben oder bewusst dokumentieren), voller `make verify` plus `make lint-backend-full`, finaler Cleanup-/Review-Pass über alle in diesem Plan geänderten Bereiche. Abschließend die Checkboxen in `audit-v1.0.0.md` und diesem Plan abgleichen und den vorbereiteten Tag-Befehl ausgeben (`git tag -a v0.14.0 …` plus Push-Hinweis), ohne ihn auszuführen.

### Akzeptanzkriterien

- [ ] `CHANGELOG.md` existiert mit vollständigem v0.14.0-Eintrag
- [ ] Alle Versionsstellen konsistent auf v0.14.0; Release-Guide-Image-Namen korrekt
- [ ] Kein unbegründetes TODO/FIXME/XXX in fiskalischen Modulen
- [ ] Voller `make verify` und `make lint-backend-full` grün auf dem Release-Commit
- [ ] Alle umgesetzten Audit-Punkte in `audit-v1.0.0.md` abgehakt; offen bleiben nur F, C7, Gate 2/3 und die Nicht-Scope-D-Punkte
- [ ] Tag-Befehl im Chat ausgegeben, nicht ausgeführt
