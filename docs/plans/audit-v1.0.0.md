# Audit-Bericht v1.0.0 (Stand 2026-07-06)

Multi-Experten-Audit vor dem v1.0.0-Release, komplementär zum [Release-Guide](plan-v1.0.0-release.md). Methode: 16 Experten-Reviews plus 4 Nachreviews über alle Gates und Qualitätsachsen (Steuer, TSE/Signatur, DSFinV-K, Beleg, Events, Schema, Security, Robustheit, Performance, API, Ops, Frontend, UX, Tests, Release-Mechanik, Leitfäden, Bootstrap/Seed, LAN-TLS-Infra, Website-Konsistenz). Jeder High-Befund wurde von einem unabhängigen Skeptiker-Agenten am Code gegen­geprüft, Mediums im Batch; Rechtsaussagen sind an `docs/rechtsquellen/` belegt. Arbeitsdokument mit Abhak-Listen; nach Abarbeitung aus `docs/plans/` entfernen.

Ground Truth: `make verify` grün (inkl. Integrationstests gegen echte DB), `golangci-lint` 0 Befunde.

Gesamtbild: kein Critical-Befund. Die fiskalische Kernarithmetik, die Signatur-Outbox, der DSFinV-K-Export und das Append-only-Journal sind von mehreren Experten unabhängig als solide bestätigt (Abschnitt E). Die Befunde konzentrieren sich auf eine Beleg-Pflichtangabe, Erstinbetriebnahme-Härtung, Contract-Feinschliff vor dem Freeze und Doku-Lücken für nicht-technische Betreiber.

---

## A. Release-Blocker (vor dem Tag fixen)

- [x] **A1 Steuersatz auf dem Kassenbeleg ausweisen.** Der Beleg druckt nur Kennzeichen A/B/C ohne Prozentsatz, ohne Legende und ohne Steuerbefreiungshinweis; § 6 Satz 1 Nr. 5 KassenSichV verlangt den anzuwendenden Steuersatz bzw. den Befreiungshinweis. Fix: Steuermatrix-Zeilen als `A (19 %): ...` oder Legende `A = 19 %, B = 7 %, C = 0 % (umsatzsteuerfrei)` plus Befreiungshinweis; Golden-Tests anpassen. Reiner Formatter-Fix. (`backend/api/druck/bondruck/application/escpos/formatter.go:237,250-259,393-406`; unabhängig bestätigt von 3 Experten. Solange offen, versprechen Website `index.astro:1331` und `README.md:57` mehr als geliefert.)
- [x] **A2 Startup-Validierung der Secrets.** JWT_SECRET/RELAY_AUTH_TOKEN werden nur auf nicht-leer geprüft, der öffentlich im Repo stehende `.env.example`-Platzhalter wird akzeptiert (JWT-Forgery = kompletter Auth-Bypass); POSTGRES_PASSWORD defaultet still auf `admin`. Fix: Mindestlänge erzwingen, bekannte Platzhalter fatal ablehnen, POSTGRES_PASSWORD ohne Default; zusätzlich Secret-Check in `scripts/prod-init.sh` (prüft bisher nur Domain/Email/Version). (`backend/config/config.go:34-44`, `scripts/prod-init.sh:90-103`)
- [x] **A3 Seed-Guard verschärfen.** Das `seed`-Subkommando ist ins Prod-Binary einkompiliert und nur durch einen Kassenjournal-Count geschützt; auf frischer Prod-DB (nach Bootstrap, vor erstem Verkauf) injiziert `jotti seed` Demo-Admins mit öffentlich bekanntem Passwort `jotti123` und Fake-TSE-Events irreversibel ins append-only Journal. Fix: Guard zusätzlich auf CountUsers==0 und fehlende Betreiber-Stammdaten, oder Env-Flag `JOTTI_ALLOW_SEED=1`, oder Build-Tag-Ausschluss im Release-Binary. (`backend/seed/writer.go:60-66`, `backend/main.go:62`)
- [x] **A4 DSFinV-K: BON_NAME beim Tagesabschluss füllen.** Jeder Kassenabschluss erzeugt eine `AVSonstige`-Zeile mit leerem BON_NAME; DSFinV-K 2.4 (S. 46, 84) verlangt bei AVSonstige zwingend eine Beschreibung, das amtliche Prüftool bemängelt das. Fix: `bonName`-Feld oder konstanter Text "Tagesabschluss" in `buildTransactions`. (`backend/api/fiskal/dsfinvk/mapper.go:427-436,713-715`)
- [x] **A5 QR-Code-Modulgröße prüfen und dynamisch wählen.** Feste Modulgröße 6 ergibt bei realer fiskaly-Payload (~350-470 Byte, QR-Version ~18) über 534 Dots plus Ruhezone und übersteigt die druckbare Breite von 576 Dots; der Pflicht-QR wäre auf 80-mm-Papier nicht druckbar. Fix: Modulgröße 4-5 oder längenabhängig; auf Zielhardware verifizieren (Gate 2.4). (`backend/api/druck/bondruck/application/escpos/constants.go:30`, `formatter.go:313`)

## B. Nur jetzt möglich: vor dem Schema-/Contract-Freeze entscheiden

Nach dem Tag nur noch per Major-Bump korrigierbar. Alle Punkte sind heute (v0-Beta, keine Bestandsdaten) billig.

- [ ] **B1 Event-Key `einzelpreis` zu `einzelpreisCents` umbenennen.** Cent-Betrag ohne das laut `docs/language.md` verbindliche Suffix, im eingefrorenen Event-Contract und deckungsgleich in API-Responses und Zod-Schemas. Jetzt in Events, DTOs, Frontend-Schemas und language.md nachziehen. (`backend/domain/kasse/bestellung.go:45`)
- [ ] **B2 Idempotenz-Schlüssel für buchende Endpunkte entscheiden.** Doppel-Tap oder WLAN-Retry bei verlorener Antwort erzeugt heute zwei signierte Events (Direktverkauf: frische Server-UUID je Request; Bestellung/Geldtransit: GetMaxVersion erst beim Write). Client-generierte ID im Request/Event dedupliziert das; nachträglich wäre das ein Contract-Bruch. Empfehlung: verkaufID clientseitig erzeugen, optionaler Schlüssel für Bestellung/Geldtransit, partieller UNIQUE-Index. (`backend/api/kasse/direktverkauf/application/command.go:88`, `frontend/src/lib/Backend.ts:110`)
- [ ] **B3 Event-Contract-Guard vervollständigen und entkoppeln.** `bestellung-korrigiert:v1` fehlt komplett (Reporting liest gesamtCents/positionen/kommentar roh), Tagesabschluss/Kassensturz und `*Id`-Felder ungepinnt; bestehende Tests round-trippen über dieselben Structs und würden einen Tag-Rename nicht bemerken. Fix: fixe JSON-Literale je Event unmarshalen und Felder asserten, EventType-Konstanten gegen die SQL-Literale prüfen, Enumerations-Meta-Test. (`backend/domain/kasse/event_json_contract_test.go:40-57`)
- [ ] **B4 HTTP-Statuscodes für Auth-Fehler festlegen.** `invalid_jwt`, `missing_authorization`, `insufficient_permissions` liefern 400 statt 401/403; das Frontend macht Auto-Logout nur bei 401, ein ablaufendes Token strandet den Helfer. Der Fehlercontract friert mit 1.0 ein. (`backend/api/middleware/middleware.go:206-244`, `frontend/src/lib/Backend.ts:127`)
- [ ] **B5 Schema-Feinschliff in `01_initial.up.sql`.** (a) CHECKs auf Geldspalten (`saldo_cents >= 0`, `gesamt_zahlungen_cents >= 0`, `preis_cents >= 0`); (b) `kassenjournal.id` und `kassensitzungen.z_nr` auf GENERATED ALWAYS AS IDENTITY; (c) `tse_signaturauftraege.transaktion_nummer`/`signatur_zaehler` auf BIGINT (fiskaly typisiert bis 2^53); (d) ENUM-vs-TEXT+CHECK-Entscheidung bewusst dokumentieren (ADD VALUE erst nach COMMIT nutzbar, Zwei-Migrations-Muster in `database/migrations/README.md` festhalten).
- [ ] **B6 API-Kosmetik vor dem Freeze.** Einheitliche leere Erfolgs-Response (`{}` statt `{"status":"ok"}` bei relay/beleg), einheitliche Datumsformate (Kalendertag als YYYY-MM-DD, Zeitpunkte RFC3339), `details`-Freitexte entweder streichen oder als englisches Diagnosefeld dokumentieren.

## C. Vor 1.0 empfohlen (Robustheit, Ops, UX, Doku)

Robustheit und Betrieb:

- [ ] **C1 Query-Fehler im Frontend sichtbar machen.** Alle Daten-Hooks verwerfen `isError`; bei 500/Netzabbruch rendert z. B. die Tischseite den Leer-Default (Saldo 0,00 EUR, Badge "Alles ausgegeben!") statt eines Fehlers. Fix: globaler QueryCache.onError-Toast plus expliziter Fehlerzustand auf den kritischen Seiten. (`frontend/src/service/table/hooks.ts:34`, `frontend/src/main.tsx:12`)
- [ ] **C2 Panic-Recovery.** Signatur-Worker, Watchdog und Rate-Limiter-Cleanup laufen ohne recover(); ein Panic stoppt die Signierung still oder reißt den Prozess ab. Fix: defer/recover mit Neustart im Run-Loop plus Recovery-Middleware in der HTTP-Chain. (`backend/app/app.go:87-91`)
- [ ] **C3 Autorisierung gegen DB-Rolle statt Token-Claim.** Rollenänderungen wirken sonst erst nach Token-Ablauf (12 h); Status-Prüfung ist bereits DB-frisch, Rolle nicht. (`backend/api/middleware/middleware.go:224`)
- [ ] **C4 `make prod-up` entschärfen.** Wird als Update-Weg beworben, zieht aber Images und migriert ohne Pre-Update-Backup und ohne Downgrade-Guard; zudem `:-latest`-Fallback in `docker-compose.prod.yml` entfernen. Einziger beworbener Update-Weg: `prod-update`. (`Makefile:156-158`)
- [ ] **C5 Windows-Starter: Downgrade-Sperre.** Der Standardweg für Vereine startet eine ältere Exe klaglos gegen neuere Daten; `is_downgrade`-Logik aus `prod-update.sh` spiegeln. (`windows/starter/update.go:27`, `main.go:100-112`)
- [x] **C6 TSE-Stammdaten hart persistieren.** Abruf ist einmalig und best-effort (nur Warn-Log); schlägt er fehl, bleiben tse.csv-Pflichtfelder dauerhaft leer. Zusätzlich TSE_SERIAL aus den Stammdaten statt aus der ersten Signatur beziehen. (`backend/api/fiskal/setup/application/setup.go:346`, `dsfinvk/mapper.go:1235`)
- [ ] **C7 Signatur-Latenz messen.** Happy Path macht 3 serielle fiskaly-Roundtrips auf einem seriellen Worker (~1,6 Signaturen/s bei 200 ms RTT); im Fiskal-E2E (Gate 2.2) p95 unter Burst messen, sonst Retrieve-first zu Start-first optimieren oder Zusage anpassen. (`backend/api/fiskal/signatur/tse_signatur_worker.go:388`)
- [ ] **C8 Kassenabschluss-Wiederanlauf idempotent machen.** Teilfehler nach Schritt 1 hängt beim Retry ein zweites `kassensturz-durchgefuehrt:v1` an; vor Schritt 1 vorhandenen Kassensturz der Sitzung erkennen. (`backend/api/kasse/kassenfuehrung/application/command.go:334`)
- [ ] **C9 Index für TSE-Queue-Monitoring.** `GetTSESignaturQueueZustand` scannt die nie geleerte Tabelle voll; partieller Index auf `(erledigt_am) WHERE status='erledigt'`. (`backend/sqlc/queries/tse_signaturauftraege.sql:78-85`)
- [ ] **C10 Gate-4-CI-Job (b) anlegen.** Migration auf befüllter Vorversions-DB plus Boot plus `rebuild-projections`; als Harness ab der ersten `02_`-Migration Pflicht. (`.github/workflows/ci.yml`)

Version-Single-Source (Gate 5/6):

- [x] **C11 KASSE_SW_VERSION an die Release-Version binden.** Hartkodierte Konstante `"1.0"` in der 10 Jahre archivierten cashregister.csv, entkoppelt vom ldflags-Tag; Test ist tautologisch. Fix: Version in den Snapshot durchreichen oder mindestens Konstante `1.0.0` plus Guard-Test gegen die Build-Version. (`backend/api/fiskal/dsfinvk/mapper.go:45,600`)
- [ ] **C12 Verfahrensdokumentation korrigieren oder Version anzeigen.** Behauptet Versions-Ausweisung in Admin-UI und auf dem Beleg; beides existiert nicht. Entweder Footer um `jotti <version>` (via /health) ergänzen und Beleg-Aussage streichen, oder Satz korrigieren. (`docs/verfahrensdokumentation.md:167`, `frontend/src/admin/AdminSidebar.tsx:160-172`)
- [ ] **C13 Versionsangaben aufräumen.** `frontend/package.json` "0.0.0" ist tot (verdrahten oder aus dem Gate-6-Bump-Set streichen); Beispiel-Versionen v0.1.0/v0.2.0 in self-hosting.md, verfahrensdokumentation.md, .env.example, docker-compose.release.yml beim Release anheben; Plan Gate 6 nennt das nicht existierende Image `ghcr.io/nicograef/jotti` (real: jotti-backend/-frontend/-migrate/-reverse-proxy).

UX (Vereinsfest-Tauglichkeit):

- [ ] **C14 Drucker-Ausfall proaktiv melden.** Fehlgeschlagene Druckaufträge sind nur auf /admin/druckstationen sichtbar; Banner auf dem Admin-Dashboard analog TSE-Warnung. (`frontend/src/admin/settings/DruckstationConfigPage.tsx:204`)
- [ ] **C15 Beleg-Schnellzugriff nach dem Kassieren.** Kassenbeleg erfordert heute 4 Interaktionen über die Historie; Aktions-Button im Erfolgs-Toast oder Drawer. (`frontend/src/service/components/table/ZahlungDrawer.tsx:66`)

Doku und Leitfäden (Gate 5):

- [ ] **C16 Kassenmeldepflicht-Frist ergänzen.** Ein-Monats-Frist nach § 146a Abs. 4 AO (BMF 28.06.2024) fehlt in `finanzamt-anmelden.md` und `checkliste.md`; compliance.md §7.1 führt sie korrekt.
- [ ] **C17 Laien-Anleitung TSE-/Internet-Ausfall.** Kernbotschaft "weiterverkaufen erlaubt, jotti signiert automatisch nach" fehlt in fehlersuche.md/haeufige-fragen.md, obwohl die Architektur es deckt.
- [ ] **C18 Checkliste nach Betriebspfad trennen.** "Tägliche Datenbank-Backups laufen" gilt nur für den Serverpfad; der Windows-Standardweg kann die Pflicht nicht erfüllen und liest sich fälschlich als nicht-konform. (`docs/leitfaden/checkliste.md:27`)
- [ ] **C19 README: entferntes Auszahlung-Feature streichen.** `README.md:21` bewirbt "Auszahlung leisten / negativen Saldo ausgleichen"; durch das Storno-Rework entfallen.
- [ ] **C20 AGENTS.md auf Freeze-Disziplin umstellen.** Instruiert noch die v0-Praxis (direkt in 01_initial editieren, Breaking Changes erwünscht) und referenziert die gelöschte `01_initial.down.sql`; auf additive `NN_*.up.sql` und `database/migrations/README.md` umschreiben. (`AGENTS.md:59-61`)
- [ ] **C21 CHANGELOG.md anlegen** (Gate 5; cliff erzeugt bisher nur flüchtige Release-Notes).
- [x] **C22 prod-init: OTP ausgeben.** installation.md verspricht den Einmalpasswort-Code in der prod-init-Ausgabe; das Skript druckt nur den grep-Hinweis. Skript greppt selbst und zeigt den Code (analog Windows-Starter), oder Doku angleichen. (`scripts/prod-init.sh:203-204`)
- [ ] **C23 fiskaly-Preisangabe entschärfen.** "ca. 8 EUR/Monat" ist unbelegte Marktannahme in drei Leitfäden; einheitlich auf "aktuellen Preis bei fiskaly erfragen" umstellen.

## D. Niedrig (nach 1.0 ohne Bruch nachrüstbar)

- [ ] D1 fiskaly `api_secret` at rest verschlüsseln oder mindestens Backup-Hinweis (Klartext in jedem pg_dump); API gibt das Secret korrekt nie heraus.
- [ ] D2 Login-Fehlercodes `no_password_set`/`user_inactive` verraten Kontenexistenz; vor dem Freeze bewusst entscheiden.
- [ ] D3 Container-Härtung: no-new-privileges, cap_drop, read_only fürs Backend; resolver/reverse-proxy laufen als root (CAP_NET_BIND_SERVICE), Caddyfile mit 0600 statt 0644 schreiben.
- [ ] D4 Resolver löst beliebige IPv4 auf (auch Loopback/Public); auf private Bereiche einschränken.
- [ ] D5 JWT im localStorage: bewusste Abwägung, CSP strikt halten; dokumentieren.
- [ ] D6 React Query: staleTime gegen Fokus-Refetch-Bursts, refetchInterval fürs "Live"-Dashboard.
- [ ] D7 DSFinV-K-Export: globaler WriteTimeout 10 s kann große Exporte abschneiden (SetWriteDeadline je Request); Export baut ZIP im RAM (bewusst pro Sitzung, dokumentieren).
- [ ] D8 Tisch-Storno-Beleg druckt den Erste-Bestellung-Zeitstempel nicht (§ 5.3-Konsistenz); zudem Doku-Code-Divergenz Klarschriftzeit (evt.Time) vs. compliance.md (TSE-logTime) auflösen.
- [ ] D9 Druck: At-least-once-Semantik (Doppeldruck bei verlorener Ergebnismeldung) in der Verfahrensdoku festhalten.
- [ ] D10 Contract-Test um Wertkonstanten `einlage`/`entnahme` ergänzen (SQL hängt an Literalen).
- [ ] D11 rebuild-projections deckt `kassensitzungen.status` nicht ab; dokumentieren oder ergänzen.
- [ ] D12 Steuersatz-Konstanten nicht datumsbewusst: an `Prozent()`/`Aufteilen()` festhalten, dass Satzänderungen nur datumsbewusst (event.Time) erfolgen dürfen, nie als Konstanten-Swap (10-Jahre-Replay).
- [ ] D13 CI: golangci-lint von `latest` auf feste Version pinnen; `go mod tidy -diff` und `-count=1` angleichen, damit CI-Grün und make-verify-Grün deckungsgleich sind; vor dem Tag einmal voller `make verify` auf dem Release-Commit.
- [ ] D14 Windows-Backup: Dump-Integritätsprüfung ergänzen (analog `gzip -t` im Docker-Pfad).
- [x] D15 DSFinV-K accuracy-Metadaten (2 vs. amtlich 5) angleichen oder totes Feld entfernen; rein dokumentarisch.
- [ ] D16 Betreiberdoku: Container-Logs enthalten bis zur Erstanmeldung ein gültiges Admin-OTP; Log-Zugriff einschränken.

## E. Bestätigt sauber (Auszug)

- Steuerarithmetik Ende-zu-Ende konsistent: eine Basis (einzelpreis mal menge, `steuer.Aufteilen`) für Beleg, TSE-processData, DSFinV-K und Reporting; Rundung kaufmännisch, kombi 70/30 summenneutral, Storno-Vorzeichen exakt, Saldo nie negativ (FIFO plus strict reduce).
- Signatur-Outbox: transaktionale Outbox, stabile tx_id mit UNIQUE, RetrieveTransaction-Heilung ergibt at-least-once mit Dedupe; FinishTransaction strikt vor Druck (§ 5.6 Variante A); Ausfallpfade (Störungsprotokoll, Nachsignierung, TSE_TA_FEHLER, Abschluss-Gate) umgesetzt.
- DSFinV-K: index.xml und DTD byte-identisch zur amtlichen Vorlage, CSV-Regeln korrekt, Storno-Abbildung spec-konform, Anfangsbestand/Geldtransit/Differenz wie die amtlichen Beispiele, BEDIENER-Felder nach § 6.4.
- Append-only mehrschichtig erzwungen (REVOKE/GRANT plus Owner-fester Trigger); OCC über UNIQUE(subject, version) wasserdicht für zustandsvalidierende Commands; Event, Outbox, Projektion und CRUD atomar in einer Transaktion; Replay deterministisch und idempotent.
- Security-Fundament: HS256 erzwungen, Argon2id nach OWASP, OTP aus crypto/rand mit Sperre, doppeltes Login-Throttling, Security-Header über Caddy für beide Auslieferungswege, ausschließlich parametrisiertes SQL, POST-only, non-root Backend, Postgres nur im internen Netz.
- Ops: prod-update mit Downgrade-Sperre, Pre-Update-Backup und Rollback-Anleitung; backup-verify läuft Ende-zu-Ende in der Release-CI; Backup enthält die Kassen-Seriennummer (§ 3.7).
- Tests überdurchschnittlich verhaltensgetrieben (fiskalische Projektion über alle Event-Typen, processData gegen Anhang-I-Beispiele, 20 DSFinV-K-Tabellen mit Golden Rows); Frontend TS-strict ohne any, kein direktes fetch, Doppel-Submit-Schutz, Geld string-basiert in Cent.
- Leitfäden: alle geprüften Klickwege, Menünamen und make-Targets stimmen mit Code und Release-ZIP überein; Verfahrensdoku deckt sich bis auf C12 mit dem Code; Release-Pfad (Tag zu ldflags zu Images zu Compose-Pinning) ist single-source bis auf C11.

## F. Nur manuell prüfbar (aus dem Release-Plan, nicht automatisierbar)

- [ ] fiskaly-TEST-Konto: Setup-Wizard real durchlaufen (TSS und Client aus jotti), TEST-zu-LIVE, PUK/PIN-Verwahrung (Gate 2.1)
- [ ] Je ein realer Fall pro signaturpflichtigem Vorgangstyp; p95-Latenz unter Burst messen (Gate 2.2, siehe C7)
- [ ] TSE stören, Störungsprotokoll und Nachsignierung mit Beleg-Vermerk beobachten (Gate 2.3)
- [ ] Beleg inkl. ESC/POS-QR auf echtem 80-mm-Drucker, QR-Lesbarkeit prüfen (Gate 2.4, siehe A5)
- [ ] Export-ZIP mit IDEA oder fiskaly-Prüftooling gegenlesen (Gate 2.5, prüft auch A4)
- [ ] prod-update/-backup/-backup-verify/-restore-Roundtrip auf Test-Server (Gate 3)
- [ ] TLS/Let's Encrypt live grün (Regressionscheck, Gate 3)
- [ ] make release-windows auf echtem Windows-Rechner smoke-testen (Gate 3/6)
- [ ] prod-init auf leerem Server bis zum ersten Admin-Login (Gate 3, prüft auch C22)
- [ ] Parallelzugriffstest mit zwei Geräten am selben Tisch (Gate 3)
- [ ] Smoke-Test der gepinnten 1.0.0-Images auf frischem Server (Gate 6)
