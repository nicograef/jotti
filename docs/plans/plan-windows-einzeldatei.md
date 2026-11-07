# Plan: Windows-Einzeldatei

> Source PRD: docs/prds/prd-windows-einzeldatei.md

## Goal

Ein Verein lädt `jotti.exe`, doppelklickt sie und kassiert ohne Docker Desktop.
Postgres 17, Migrationen, Backend, Caddy mit App, Status-Seite und Relay laufen in dieser einen Datei.
Bestehende Docker-Installationen ziehen per Übernahme verlustfrei um; der Linux-Betrieb bleibt bei Compose.

Stufe 1 (Phasen 1–6) ist ein Durchstich mit Messung auf einem Windows-Runner der CI.
Phase 6 ist das Go/Stop-Gate: Alle späteren Phasen hängen davon ab. Es entscheidet automatisch anhand der Messwerte; Nico entscheidet nur, wenn ein Wert auch nach den Abhilfen reißt.

## Architectural decisions

- **Modul.** Neues Go-Modul `windows/einzeldatei` (`github.com/nicograef/jotti/windows/einzeldatei`, go 1.27.3), Programmdatei `jotti.exe`.
  Es steht in `go.work`. `replace` auf `../../backend`, `../../reverse-proxy`, `../starter` und `../relay` hält `go mod tidy` und govulncheck mit `GOWORK=off` lauffähig.
  Jede `replace`-Zeile kommt mit dem ersten Import ihres Moduls.
- **D13.** Das Modul bekommt ein eigenes Ziel `check-einzeldatei`, einen eigenen CI-Job `einzeldatei-ci` und einen eigenen Pfadfilter `einzeldatei`.
  Der Matrix `windows-ci` tritt es nicht bei, weil sein Pfadfilter auch Backend, Proxy, Frontend und Migrationen umfasst.
- **Keine Migration, keine Event- oder API-Änderung.** Schema, sqlc, Abfragen, Trigger und REVOKE bleiben unverändert.
  Der Plan belegt keine Nummer; die nächste freie ab 17 bleibt für die anderen Pläne.
- **Wiederverwendung statt Kopie.** Das Backend startet über `backend/app — Serve()`, das Container-Binary ruft dieselbe Funktion.
  Die Caddyfile-Erzeugung kommt aus `reverse-proxy/caddyconf`, der LAN-Modus aus `reverse-proxy/lan`.
  Das Relay kommt aus `windows/relay/relay`, die Hostfunktionen aus `windows/starter/host`, die reine Logik weiter aus `windows/starter/core`.
- **Assets.** `scripts/einzeldatei-assets.sh` füllt das gitignorierte `windows/einzeldatei/assets/files/`.
  Hinein kommen die gekürzten Postgres-Binaries als `postgres.zip`, der App-Build als `app/` und die Migrationen als `migrations/`.
  `windows/einzeldatei/assets — FS` bettet `all:files` ein; das eingecheckte `files/.gitkeep` hält `make check-einzeldatei` ohne Assets lauffähig.
- **Pins.** `scripts/einzeldatei-assets.sh` pinnt `POSTGRES_WINDOWS_VERSION` (17.11) mit SHA-256; `scripts/check-pins.sh` gleicht sie mit dem `postgres:`-Tag ab (D15).
  golang-migrate v4.20.1 wie `database/migrate/Dockerfile`; Caddy v2.11.7, `caddy-dns/acmedns` v0.7.0 und `mholt/caddy-ratelimit` v0.1.0 wie `reverse-proxy/Dockerfile`.
  `check-pins.sh` gleicht auch die Caddy-Versionen ab.
- **Ablage.** Der Programmordner `%ProgramFiles%\jotti\<version>\` hält `jotti.exe`, `postgres\` und `app\`, behalten werden die aktuelle und die vorige Version.
  Das Datenverzeichnis kommt aus `core.StateDir()` (`%PROGRAMDATA%\jotti`).
  Es hält `.env`, `last-version`, `pgdata\`, `caddy\`, `install.json`, `logs\` und `backups\`.
- **Ports.** Postgres lauscht auf `127.0.0.1:15432`, das Backend auf `127.0.0.1:13000`, die Status-Seite auf `127.0.0.1:8484`, Caddy auf 80 und 443.
  Die festen Loopback-Ports liegen unter dem dynamischen Portbereich von Windows (ab 49152); die Port-Prüfung meldet Konflikte.
- **Postgres.** Gestartet und gestoppt wird über `pg_ctl`, das unter Administratorrechten einen eingeschränkten Token setzt.
  initdb läuft mit `-E UTF8 --locale=C --locale-provider=icu --icu-locale=en-US --auth=scram-sha-256` und dem Benutzer und Passwort aus `.env`.
- **Migrationen.** golang-migrate als Bibliothek mit `source/iofs` und `database/pgx/v5` nutzt dieselbe Tabelle `schema_migrations` wie das Migrate-Image.
- **Caddy im Prozess.** `windows/einzeldatei/internal/proxy — Run()` adaptiert die erzeugte Caddyfile und lädt sie mit `caddy.Load`.
  D18 gilt: Caddy terminiert TLS, leitet `/api/` weiter und liefert die App aus; nur Upstream, App-Wurzel und Speicherort unterscheiden sich vom Container.
  Die Einzeldatei nutzt nur die LAN-Konfiguration; Stufe 1 rendert sie ohne Install-State, also nur mit der Fallback-Site und interner CA.
- **Dienst.** Der Windows-Dienst `jotti` startet manuell und läuft unter dem virtuellen Konto `NT SERVICE\jotti`.
  Sein Binary ist `<Programmordner>\jotti.exe dienst`. `golang.org/x/sys/windows/svc` und `svc/mgr` liefern Dienst und Registrierung.
- **Konsole.** Der Doppelklick ohne Argument startet die Konsole: Einrichten, Update, Menü. Ihr Manifest verlangt Administratorrechte.
  Konsolentexte sind ASCII (`scripts/check-language.sh`, Regel 1a).
- **Testumgebung.** Die Konsole reicht `ACMEDNS_BASE_URL` und `PROXY_LE_STAGING` aus ihrer Umgebung an Prozess und Dienst weiter.
  Die CI zeigt `ACMEDNS_BASE_URL` auf `http://127.0.0.1:9`, damit kein Lauf sich bei `auth.jotti.rocks` registriert oder Let's-Encrypt-Limits verbraucht.
- **Relay im Prozess.** Das Relay spricht das Backend direkt auf `http://127.0.0.1:13000` an, mit `RELAY_AUTH_TOKEN` aus `.env`.
- **Aufbewahrung.** Kein Pfad löscht `pgdata`.
  Zurückspielen und eine abgebrochene Übernahme benennen den Cluster in `pgdata-verworfen-<zeit>` um; „Entfernen" lässt das Datenverzeichnis stehen.
- **CI-Messung.** `einzeldatei-build` (ubuntu) baut per Cross-Compile wie der Release, `einzeldatei-windows` (`windows-2025`) führt `scripts/einzeldatei-smoke.sh` in Git Bash aus.
  Die Schwellen aus Stufe 1 bleiben danach als dauerhaftes Gate im Job.
  Ein Kriterium mit Runner-Lauf prüft den CI-Lauf des gepushten HEAD-Commits; vorher wird der Branch gepusht.

## Inventory

- `backend/main.go — run()` — Startfolge (Projektionen, Initial-Admin, Server), die `app.Serve()` übernimmt.
- `backend/app/app.go — NewApp()`, `(*App).Run()` — `Addr` lauscht auf allen Schnittstellen; TSE-Worker und Watchdog starten hier.
- `backend/config/config.go — Load()`, `ValidateSecrets()`, `postgresConfig` — liest nur Umgebungsvariablen; die Einzeldatei setzt die Felder von `Config.Postgres` einzeln, der Typ bleibt unexportiert.
- `backend/db — ConnString()`, `PingWithRetry()` — Verbindungsaufbau, den `main.go` heute selbst verdrahtet.
- `backend/bootstrap — EnsureInitialAdmin()`, `Result.Log()` — Anmelde-Code im Log, den `core.ParseAdminOTP()` liest.
- `backend/seed — Run()` — Demo-Daten für den Übernahme-Integrationstest.
- `backend/sqlc/queries/relay.sql — GetOffeneDruckauftraege` — liest offene Aufträge ohne Sperre; zwei Relays drucken jeden Bon doppelt.
- `reverse-proxy/caddyfile.go — renderCaddyfile()`, `renderHTTPOnlyCaddyfile()`, `proxySnippet()`, `appHandler` — `backend:3000` und `/srv/app` fest verdrahtet.- `reverse-proxy/main.go — runLANMode()`, `runCaddy()`, `statePath` — LAN-Modus als `package main`, Caddy als Kindprozess.
- `reverse-proxy/statuspage.go — statusListenAddr`, `newStatusServer()` — lauscht auf `:8484`; im Container begrenzt erst Compose auf Loopback.
- `reverse-proxy/state.go — ensureState()`, `register.go — registerWithACMEDNS()`, `hostname.go — deriveHostname()` — Install-State und grüne Adresse.
- `reverse-proxy/Dockerfile` — Caddy- und Plugin-Versionen, die die Einzeldatei übernimmt; kopiert wie `Dockerfile.dockerignore` nur `reverse-proxy/*.go`, kein Unterpaket.
- `database/migrate/Dockerfile — MIGRATE_VERSION` — golang-migrate-Version.
- `windows/starter/core — StateDir()`, `MaterializeEnv()`, `IsDowngrade()`, `ShouldBackup()`, `DumpsToDelete()`, `ParseAdminOTP()`, `SelectLANIP()`, `StateDirSDDL()`, `PortBelegtDiagnose()` — reine Logik, die die Einzeldatei nutzt.
- `windows/starter/statedir_windows.go — prepareStateDir()`, `lanip.go — detectLANIP()`, `ports.go — checkPorts()`, `firewall.go — ensureFirewall()`, `update.go — notifyIfUpdateAvailable()` — Hostfunktionen in `package main`.
- `windows/starter/stack.go` — Container- und Volume-Namen, die die Übernahme liest.
- `windows/starter/backup.go — dumpDatabase()`, `configvolume.go — readConfigVolume()` — Vorbild für Dump und Geheimnis-Lesen per `docker exec`.
- `windows/relay/main.go — main()`, `loadConfigFromEnv()`, `fuehreZyklusAus()` — Relay als `package main`.
- `docker-compose.release.yml` — Volumes `postgres-data`, `caddy-data`, `proxy-state`, `jotti-config`, die die Übernahme liest.
- `Makefile — go-module-check`, `check`, `GO_MODULES`, `release-windows`, `starter-syso` — Gate- und Build-Ziele.
- `.github/workflows/ci.yml — changes`, `windows-ci`; `release.yml` — Windows-Release-ZIP und Smoke-Test; `security-scans.yml — vuln-scan-go`.
- `scripts/check-pins.sh — collect_pins()`, `scripts/check-language.sh`, `scripts/check-compose-twins.sh`, `scripts/lib.sh — start_throwaway_postgres()`, `scripts/test-integration.sh`, `scripts/ops-smoke.sh — step_sale()`.

## Open questions / Risks

Entschieden von Nico am 7. November:

- Caddy liefert die App auch in der Einzeldatei aus (D18 bleibt); neu bewertet wird das nur, wenn die Datei wegen Caddy über 120 MB kommt.
- Nico hat keinen Windows-PC. Was nur ein echter PC prüft, deckt der Windows-Runner der CI ab, so weit er reicht:
  - Defender misst Phase 5 mit eingeschaltetem Echtzeitschutz; lässt er sich nicht einschalten, gilt das Kriterium als nicht messbar und das Gate fragt Nico.
  - Die Verzeichnisrechte aus #174 prüft ein Test auf dem Windows-Runner mit echten Rechten (`icacls`).
  - Die Übernahme prüfen die Integrationstests unter Linux; das erste Release mit Übernahme heißt im Leitfaden Beta, bis ein Verein sie bestätigt.
- Das Go/Stop-Gate entscheidet automatisch (Phase 6); bei Stop landet nur die Entscheidungszeile.
- Frage 14 freigegeben: `reverse-proxy/Dockerfile` und sein `.dockerignore` übernehmen alle Go-Quellen unter `reverse-proxy/` ohne Tests.
- Für alle übrigen Fragen unten gilt die jeweilige Empfehlung.


1. **D18 oder Backend liefert die App.** Der Plan behält Caddy (D18) und damit eine Proxy-Konfiguration für Container und Einzeldatei.
   Alternative: Das Backend liefert App und TLS selbst (certmagic mit acme-dns), Caddy entfällt in der Einzeldatei.
   Gewinn: geschätzt 20–30 MB weniger Dateigröße und eine Schicht weniger im Prozess (Schätzung, nicht gemessen).
   Preis: Header, CSP, Rate-Limit und Fallback-CA gäbe es zweimal, in Caddyfile und Go, und sie liefen still auseinander; D18 bräche für Windows.
   Empfehlung: Caddy behalten; neu entscheiden nur, wenn Phase 5 die 120 MB wegen Caddy reißt.
2. **Runner ist kein Windows 10/11.** GitHub bietet für x64 nur Windows-Server-Images (`windows-2025`); die Messung läuft dort.
   Empfehlung: Nico prüft vor dem ersten Release mit `jotti.exe` einmal auf Windows 10 22H2 und Windows 11.
   Annahme aus dem Gedächtnis, nicht geprüft: Der Verbraucher-ESU für Windows 10 endete im Oktober 2026.
3. **Defender auf dem Runner.** Vermutlich schaltet das Runner-Image den Echtzeitschutz ab (Inferenz).
   Phase 5 schaltet ihn ein und hält `defender_echtzeitschutz` fest. Lässt er sich nicht einschalten, ist das Defender-Kriterium nicht messbar.
   SmartScreen und Mark-of-the-Web lassen sich in der CI nicht nachstellen.
   Empfehlung: Dann prüft Nico auf einem echten PC, und sein Ergebnis ersetzt das Defender-Kriterium in Phase 6.
4. **Quelle der Postgres-Binaries.** Der Plan nimmt das ZIP von EDB für 17.11 mit gepinnter Prüfsumme.
   Bietet EDB kein ZIP für 17.11, stoppt Phase 4. Alternative sind die gekürzten Binaries von zonky (`embedded-postgres-binaries`). Empfehlung: EDB.
5. **Sortierfolge.** Docker-Cluster sortieren nach glibc `en_US.utf8`, die Einzeldatei nach ICU `en-US`.
   Textsortierung kann sich nach der Übernahme in Randfällen ändern; Gleichheit und Eindeutigkeit bleiben byteweise gleich. Empfehlung: ICU `en-US` wie in den Architekturentscheidungen.
6. **`golang.org/x/sys` als direkte Abhängigkeit.** Dienst, Registrierung und Job-Steuerung brauchen es.
   Caddy zieht es ohnehin transitiv. Empfehlung: als Teil der freigegebenen Abhängigkeiten werten.
7. **Ab wann der Release `jotti.exe` anhängt.** Empfehlung: ab Phase 11, wenn Erstinstallationen Gleichstand haben.
   Docker-Vereine bekommen bis Phase 12 eine klare Ablehnung mit Verweis auf das ZIP. Alternative: erst ab Phase 12.
8. **Übernahme nicht in der CI testbar.** Windows-Runner von GitHub fahren keine Linux-Container.
   Die Integrationstests prüfen Dump, Restore und Abgleich unter Linux. Empfehlung: Nico zieht vor dem ersten Release mit Übernahme einmal eine echte Docker-Installation um.
9. **Doppeldruck.** `GetOffeneDruckauftraege` sperrt nicht. Läuft nach dem Umstieg noch `jotti-relay.exe`, druckt jeder Bon doppelt.
   Empfehlung: Die Konsole warnt, wenn ein Prozess `jotti-relay.exe` läuft; Leitfaden und Übernahme sagen es ausdrücklich. Eine Sperre in der Abfrage gehört nicht in diesen Plan.
10. **Release-Rauchtest ohne Release-Compose.** `release.yml` testet die Images heute mit `docker-compose.release.yml`.
    Option (a): `docker-compose.prod.yml` bekommt `PROXY_HTTP_ONLY` durchgereicht, eine Footgun für Self-Hoster.
    Option (b): `release.yml` taggt die frisch gebauten Images auf die Namen von `docker-compose.local.yml` um und startet mit `--no-build`.
    Empfehlung: (b), weil keine Compose-Datei eines Vereins sich ändert.
11. **Kurzanleitung ohne ZIP.** Empfehlung: `packaging/windows/KURZANLEITUNG.md` entfällt mit dem ZIP; die Konsole verlinkt den Leitfaden auf jotti.rocks.
12. **Code-Signatur.** Phase 13 braucht Nicos Entscheidung und beschreibt beide Wege.
    Empfehlung: Signatur über Azure Trusted Signing, falls jotti die Zulassungsbedingungen erfüllt, sonst unsigniert. Die Bedingungen prüft vorher der `nico:web-researcher`.
13. **Wartezeit vor Phase 14.** Phase 14 wartet auf zwei veröffentlichte Releases nach dem ersten Release mit Übernahme; der Lauf stoppt davor.
    Empfehlung: Phase 14 startet als eigener Lauf, sobald ihr Tag-Kriterium erfüllt ist.
14. **Docker-Build des Proxys.** Phase 2 und 7 legen Pakete unter `reverse-proxy/` an. `reverse-proxy/Dockerfile` und `Dockerfile.dockerignore` lassen heute nur `reverse-proxy/*.go` durch.
    Phase 2 ändert beide auf alle Go-Quellen unter `reverse-proxy/` ohne Tests; AGENTS.md verlangt dafür Nicos Zustimmung.
    Empfehlung: zustimmen. Image-Inhalt und Laufzeit bleiben gleich, nur der Build-Kontext wächst um die Unterordner.
15. **#174 Verzeichnisrechte.** Phase 3 verschiebt `prepareStateDir()`, Phase 10 erweitert die SDDL um das Dienstkonto.
    Nicos Test von #174 auf einem echten PC steht aus. Empfehlung: Phase 3 erst nach diesem Test landen, sonst gilt sein Ergebnis für den verschobenen Code.
16. **Dateigröße.** Reißt die Datei 120 MB, kürzt Phase 5 zuerst die Postgres-Allowlist. Ein Packer wie UPX scheidet aus, weil er Virenscanner anschlägt.
17. **Was bei Stop landet.** Phasen 1–5 dienen nur der Einzeldatei; nach einem Stop sind Modul, Windows-Job und die Pakete aus Phasen 1–3 Ballast.
    Empfehlung: Bei Stop landet nur die Entscheidungszeile; der Branch `plan/windows-einzeldatei` bleibt ungemergt als Befund stehen.
    Alternative: Phasen 1–3 landen als Refactoring ohne Nutzerwirkung.

## Phase 1: Backend-Start als aufrufbare Funktion

**Depends on**: none

### Context

- `backend/main.go — main()`, `run()` — Startfolge heute im `package main`.
- `backend/app/app.go — NewApp()`, `(*App).Run()` — `Addr: ":%d"`.
- `backend/config/config.go — Config`, `postgresConfig`, `ValidateSecrets()`.
- `backend/db — ConnString()`, `PingWithRetry()`.

### What to build

- `backend/app/serve.go — Serve(ctx context.Context, cfg config.Config, db *sql.DB, version string) error`: Projektionen neu aufbauen, `bootstrap.EnsureInitialAdmin()`, `NewApp()`, `Run()`.
- `backend/main.go — run()` ruft `app.Serve()` mit dem Signal-Kontext; `seed` bleibt unverändert.
- `backend/config/config.go`: neues Feld `Config.ListenHost` (leer heißt alle Schnittstellen, wie heute); `Load()` lässt es leer.
- `backend/app/app.go — NewApp()` setzt `Addr` mit `net.JoinHostPort(cfg.ListenHost, port)`.
- `backend/db — Open(ctx context.Context, connString string) (*sql.DB, error)` übernimmt Pool-Einstellungen und `PingWithRetry()` aus `main.go`.

### Acceptance criteria

- [ ] Das Backend-Modul besteht sein Gate: `make check-backend`
- [ ] Repositories und Abläufe gegen echtes Postgres bleiben grün: `TEST_PG_CONTAINER=jotti-pg-einzeldatei TEST_PG_PORT=5447 make check-integration`
- [ ] Das Container-Binary startet über `app.Serve()` und bedient die App: `E2E_STACK_PORT=8093 E2E_STACK_PROJECT=jotti-e2e-einzeldatei-1 make test-e2e`

## Phase 2: Caddyfile-Erzeugung als importierbares Paket

**Depends on**: none

### Context

- `reverse-proxy/caddyfile.go — renderCaddyfile()`, `renderPublicCaddyfile()`, `renderHTTPOnlyCaddyfile()`, `proxySnippet()`, `wildcardSite()`, `appHandler`, `contentSecurityPolicy`.
- `reverse-proxy/caddyfile_test.go — TestRocksCaddyfileCarriesSameCSP`, `TestRocksCaddyfileServesSameApp` — lesen `Caddyfile.rocks`.
- `reverse-proxy/main.go — runLANMode()`, `runPublicMode()`, `runHTTPOnlyMode()`.
- `reverse-proxy/Dockerfile` (`COPY reverse-proxy/*.go`), `reverse-proxy/Dockerfile.dockerignore`.

### What to build

- Neues Paket `reverse-proxy/caddyconf` mit `RenderLAN(LANInput) string`, `RenderPublic(PublicInput) string`, `RenderHTTPOnly() string`.
- `LANInput` trägt die Felder von `caddyfileInput`, die acme-dns-Zugangsdaten als eigene Felder statt `InstallState`.
  Dazu kommen `Upstream` (Container: `backend:3000`), `AppRoot` (Container: `/srv/app`) und `StorageRoot` (leer heißt Caddys Vorgabe).
- Public und HTTP-only rendern weiter die Container-Werte; nur die Einzeldatei braucht andere.
- `StorageRoot` rendert `storage file_system <pfad>` in die globalen Optionen.
- Pfade rendern mit Schrägstrichen, ein Pfad mit Leerzeichen in Anführungszeichen (`%ProgramFiles%` enthält eines). Mit Container-Werten bleibt die Ausgabe byteweise gleich.
- `reverse-proxy/caddyfile.go` und sein Test ziehen nach `reverse-proxy/caddyconf/`; die Rocks-Tests lesen `../Caddyfile.rocks`. `main.go` ruft das Paket mit den Container-Werten.
- `reverse-proxy/Dockerfile` kopiert alle Go-Quellen unter `reverse-proxy/`; `Dockerfile.dockerignore` lässt `reverse-proxy/**/*.go` ohne `*_test.go` durch (offene Frage 14).
- Neuer Test `TestRenderLAN_WindowsWerte` in `reverse-proxy/caddyconf/caddyfile_test.go`.
  Upstream `127.0.0.1:13000`, App-Wurzel `C:/Program Files/jotti/v0.0.1-ci/app` in Anführungszeichen und Speicherort erscheinen; sonst gleicht die Ausgabe dem Container.

### Acceptance criteria

- [ ] Das Proxy-Modul besteht sein Gate: `make check-local-proxy`
- [ ] Loopback-Upstream, App-Wurzel mit Leerzeichen und Speicherort rendern wie erwartet: `cd reverse-proxy && go test -count=1 -run '^TestRenderLAN_WindowsWerte$' -v ./caddyconf/ | grep -- '--- PASS: TestRenderLAN_WindowsWerte '`
- [ ] Das Image baut mit dem Unterpaket, und der HTTP-only-Modus der E2E-Suite läuft unverändert: `E2E_STACK_PORT=8094 E2E_STACK_PROJECT=jotti-e2e-einzeldatei-2 make test-e2e`

## Phase 3: Hostfunktionen des Starters als Paket

**Depends on**: none

### Context

- `windows/starter/statedir_windows.go — prepareStateDir()`, `statedir_other.go — prepareStateDir()`.
- `windows/starter/lanip.go — detectLANIP()`, `outboundIP()`, `localInterfaces()`.
- `windows/starter/ports.go — checkPorts()`, `portAvailable()`, `lookupPortOwners()`; `reverseProxyRunning()` bleibt Docker-spezifisch im Starter.
- `windows/starter/firewall.go — ensureFirewall()`, `windows/starter/update.go — notifyIfUpdateAvailable()`, `fetchLatestRelease()`.

### What to build

- Neues Paket `windows/starter/host` mit `PrepareStateDir(dir string) error`, `DetectLANIP() string`, `CheckPorts(ports []int) string`, `EnsureFirewallRule(rule FirewallRule) bool`, `NotifyIfUpdateAvailable(version string)`.
- `FirewallRule` trägt Name, Ports und optional den Programmpfad; der Starter übergibt `jotti`, `80,443` und keinen Pfad, wie heute.
- Windows-Dateien behalten die Endung `_windows.go`, die Gegenstücke `_other.go`.
- `windows/starter/main.go — run()` ruft das Paket; Verhalten und Konsolentexte des Starters bleiben gleich.

### Acceptance criteria

- [ ] Das Starter-Modul besteht sein Gate: `make check-starter`
- [ ] Der Windows-Zweig kompiliert und besteht vet: `cd windows/starter && GOOS=windows go vet ./...`
- [ ] Der Starter baut für Windows: `make build-starter-windows`

## Phase 4: Durchstich der Einzeldatei

**Depends on**: 1, 2, 3

### Context

- `backend/app — Serve()`, `backend/config — Config`, `ValidateSecrets()`, `backend/db — Open()` aus Phase 1.
- `reverse-proxy/caddyconf — RenderLAN()` aus Phase 2.
- `windows/starter/host — PrepareStateDir()` aus Phase 3; `windows/starter/core — StateDir()`, `MaterializeEnv()`, `IsDowngrade()`.
- `Makefile — go-module-check`, `build-starter-windows`, `starter-syso`, `check`, `GO_MODULES`.
- `.github/workflows/ci.yml — changes`, `windows-ci`; `.github/workflows/security-scans.yml — vuln-scan-go`.
- `scripts/check-pins.sh — collect_pins()`.

### What to build

- Modul `windows/einzeldatei` in `go.work` und `go.work.sum`, mit `replace` auf Backend, Proxy und Starter.
- `windows/einzeldatei/assets/assets.go — FS` mit `//go:embed all:files`; `files/.gitkeep` eingecheckt, `files/*` sonst in `.gitignore`.
- `scripts/einzeldatei-assets.sh` lädt das Postgres-ZIP, prüft die Prüfsumme und kürzt es auf eine Allowlist.
  Die Allowlist umfasst `postgres`, `pg_ctl`, `initdb`, `pg_dump`, `pg_restore`, deren DLLs, die DLLs aus `lib\`, die initdb braucht, und `share\`.
  Danach baut das Skript die App mit `VERSION` und kopiert `database/migrations/*.up.sql`.
- `windows/einzeldatei/internal/layout — Paths(programFiles, programData, version string) Paths`, `Extract(src fs.FS, dst string) error` (idempotent über eine Marker-Datei je Version).
- `windows/einzeldatei/internal/layout — DockerInstallation(dataDir string) (bool, error)`: `last-version` vorhanden und `pgdata` fehlt heißt Docker-Installation.
  Der Starter schreibt `last-version` erst nach einem gesunden Start; eine `.env` allein bleibt auch nach einem abgebrochenen ersten Start der Einzeldatei zurück.
  Der Start verweigert dann mit Hinweis auf das ZIP. Neuer Test `TestDockerInstallation` deckt alle vier Kombinationen ab.
- `windows/einzeldatei/internal/postgres — Cluster{Bin, Data string; Port int}` mit `Init(user, password string)`, `Start()`, `Stop()`, `Running()` über `initdb` und `pg_ctl`.
  `Start()` stoppt einen verwaisten Postmaster vorher mit `-m fast`.
- `windows/einzeldatei/internal/schema — Migrate(ctx context.Context, dsn string, src fs.FS) error` mit golang-migrate.
- `windows/einzeldatei/internal/proxy — Run(ctx context.Context, caddyfile string) error` mit Caddy als Bibliothek samt `acmedns` und `caddy-ratelimit`.
- `windows/einzeldatei/main.go — run()` in Stufe 1:
  - Datenverzeichnis vorbereiten, Docker-Installation ausschließen, `.env` vorbereiten, Downgrade prüfen, Assets in den Programmordner entpacken.
  - Cluster beim ersten Start anlegen und starten, migrieren.
  - `app.Serve()` mit `ListenHost 127.0.0.1` und Port 13000 starten; die Konfiguration kommt aus `.env`, geprüft mit `config.ValidateSecrets()`.
  - Caddy mit `RenderLAN()` ohne Install-State starten: Upstream `127.0.0.1:13000`, App-Wurzel `<Programmordner>\app`, Speicher `<Daten>\caddy`.
  - Auf `https://127.0.0.1/api/health` warten, `last-version` schreiben.
  - Enter beendet in der Reihenfolge Caddy, Backend, Postgres.
- `windows/einzeldatei/jotti.manifest` (requireAdministrator) und `rsrc_windows_amd64.syso` über das Ziel `einzeldatei-syso`.
- Makefile: `build-einzeldatei-windows` (Assets und Cross-Compile mit `-X main.version`), `check-einzeldatei` (`$(go-module-check)` plus `GOOS=windows go vet ./...`).
  `check-einzeldatei` kommt in `check`, das Modul in `GO_MODULES`.
- `scripts/check-pins.sh` gleicht `POSTGRES_WINDOWS_VERSION` mit dem `postgres:`-Tag ab.
- CI: Filter `einzeldatei` und drei Jobs, die alle an diesem Filter oder `ci` hängen.
  - Der Filter umfasst alles, was die Datei einbettet oder baut: `windows/**`, `backend/**`, `reverse-proxy/**`, `frontend/**`, `database/migrations/**`, `scripts/einzeldatei-*.sh`.
    Dazu kommen wie bei den übrigen Go-Filtern `go.work`, `go.work.sum`, `scripts/setup-dev-tools.sh`, `.golangci.yml` und `Makefile`.
  - `einzeldatei-ci` führt `make check-einzeldatei` aus.
  - `einzeldatei-build` baut mit `VERSION=v0.0.1-ci` und lädt das Artefakt `jotti-exe` hoch.
  - `einzeldatei-windows` auf `windows-2025` führt `scripts/einzeldatei-smoke.sh start jotti.exe` aus.
  - `security-scans.yml` bekommt einen Matrix-Eintrag für das Modul.
- `scripts/einzeldatei-smoke.sh start <exe>` startet die Datei mit einer Pipe auf stdin.
  Es wartet, bis `https://127.0.0.1/api/health` 200 liefert und `https://127.0.0.1/` die `index.html` der App (beide mit `-k`, interne CA).
  Danach sendet es Enter und prüft Exit 0 ohne verbliebenen `postgres.exe`.

### Acceptance criteria

- [ ] Das neue Modul besteht sein Gate: `make check-einzeldatei`
- [ ] Die Docker-Erkennung verweigert `last-version` ohne `pgdata` und lässt die übrigen Fälle durch: `cd windows/einzeldatei && go test -count=1 -run '^TestDockerInstallation$' -v ./internal/layout/ | grep -- '--- PASS: TestDockerInstallation '`
- [ ] Die Postgres-Pins von Compose und Einzeldatei stimmen überein: `bash scripts/check-pins.sh`
- [ ] Die Workflows bestehen actionlint: `make check-workflows`
- [ ] Die Datei baut per Cross-Compile mit allen Assets: `make build-einzeldatei-windows VERSION=v0.0.1-ci`
- [ ] Auf dem Windows-Runner startet sie, liefert Anmeldeseite und Health und endet sauber: `id="$(gh run list -w CI -e push -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" && gh run watch "$id" --exit-status >/dev/null && gh run view "$id" --json jobs -q '.jobs[] | select(.name=="einzeldatei-windows") | .conclusion' | grep -x success`
- [ ] Die Repo-Gates bleiben grün: `make check-repo check-format`

## Phase 5: Messung auf dem Windows-Runner

**Depends on**: 4

### Context

- `scripts/einzeldatei-smoke.sh` und `.github/workflows/ci.yml — einzeldatei-windows` aus Phase 4.
- `docs/prds/prd-windows-einzeldatei.md` — Entscheidungen des Owners, Punkte 5 und 6 (Schwellen).

### What to build

- `scripts/einzeldatei-smoke.sh messung <exe> <tsv>`: Der Lauf startet auf einem frischen Runner und misst `erststart_s` vom Prozessstart bis `/` und `/api/health` antworten.
  Nach Enter misst er auf demselben Runner `folgestart_s` und hält `groesse_bytes` der Datei fest.
- `einzeldatei-windows` schaltet vor dem Laden des Artefakts den Defender-Echtzeitschutz ein und aktualisiert die Signaturen (`pwsh`).
  Danach schreibt der Job `defender_echtzeitschutz` (aus `Get-MpComputerStatus`) und `defender_funde` (aus `Get-MpThreatDetection`) in die TSV.
- Der Job lädt `messung.tsv` immer als Artefakt `einzeldatei-messung` hoch und schreibt die Werte in die Job-Summary.
- Als letzter Schritt läuft `scripts/einzeldatei-smoke.sh auswerten <tsv>` und scheitert über einer Schwelle: 120000000 Bytes, 60 s, 15 s, Echtzeitschutz aus, Funde größer 0.
  Ein gerissener Wert färbt den Job rot, das Artefakt liegt trotzdem vor; Phase 6 bewertet es.

### Acceptance criteria

- [ ] Shell-Skripte und Workflows bestehen ihre Prüfungen: `make check-shell check-workflows`
- [ ] Der Messlauf des HEAD-Commits hat alle fünf Werte geliefert und liegt lokal vor: `id="$(gh run list -w CI -e push -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" && gh run watch "$id" >/dev/null; rm -rf /tmp/einzeldatei-messung && gh run download "$id" -n einzeldatei-messung -D /tmp/einzeldatei-messung && test "$(cut -f1 /tmp/einzeldatei-messung/messung.tsv | sort | paste -sd' ')" = 'defender_echtzeitschutz defender_funde erststart_s folgestart_s groesse_bytes'`
- [ ] Die Auswertung lässt Werte unter den Schwellen durch und scheitert an einem überschrittenen: `printf 'groesse_bytes\t90000000\nerststart_s\t10\nfolgestart_s\t5\ndefender_echtzeitschutz\ttrue\ndefender_funde\t0\n' > /tmp/em-ok.tsv && printf 'groesse_bytes\t130000000\nerststart_s\t10\nfolgestart_s\t5\ndefender_echtzeitschutz\ttrue\ndefender_funde\t0\n' > /tmp/em-zu-gross.tsv && bash scripts/einzeldatei-smoke.sh auswerten /tmp/em-ok.tsv && ! bash scripts/einzeldatei-smoke.sh auswerten /tmp/em-zu-gross.tsv`

## Phase 6: Go/Stop-Gate nach Stufe 1

**Depends on**: 5

### Context

- `/tmp/einzeldatei-messung/messung.tsv` aus dem Messlauf von Phase 5; Format `<schlüssel>\t<wert>` je Zeile.
- `docs/decisions.md` — die nächste freie D-Nummer bei der Landung.

### What to build

- Das Gate entscheidet automatisch: Halten alle vier Messwerte ihre Schwelle, ist das Go. So hat Nico es am 7. November festgelegt.
- Reißt ein Wert, versucht die Lane zuerst die Abhilfen des Plans, etwa die Postgres-Allowlist zu kürzen, und misst neu. Erst wenn danach noch ein Wert reißt, legt die Session Nico die Messwerte vor; Nico entscheidet dann Go oder Stop.
- Die Session hängt das Ergebnis als Zeile an `docs/decisions.md` an.
  Bei Go: Der Windows-Weg ist die Einzeldatei mit eingebettetem Postgres 17, Caddy im Prozess und Windows-Dienst.
  Bei Stop: kein Einzeldatei-Weg, begründet mit den gerissenen Schwellen. Dann landet nur diese Zeile; der Branch `plan/windows-einzeldatei` bleibt ungemergt stehen.

### Acceptance criteria

- [ ] Die Datei ist höchstens 120 MB groß: `awk -F'\t' '$1=="groesse_bytes"{f=1; exit !($2<=120000000)} END{if(!f) exit 1}' /tmp/einzeldatei-messung/messung.tsv`
- [ ] Der erste Start erreicht die Anmeldeseite in höchstens 60 s: `awk -F'\t' '$1=="erststart_s"{f=1; exit !($2<=60)} END{if(!f) exit 1}' /tmp/einzeldatei-messung/messung.tsv`
- [ ] Ein späterer Start erreicht sie in höchstens 15 s: `awk -F'\t' '$1=="folgestart_s"{f=1; exit !($2<=15)} END{if(!f) exit 1}' /tmp/einzeldatei-messung/messung.tsv`
- [ ] Defender lief mit Echtzeitschutz und hat nichts gemeldet: `awk -F'\t' '$1=="defender_echtzeitschutz"&&$2=="true"{e=1} $1=="defender_funde"&&$2==0{f=1} END{exit !(e&&f)}' /tmp/einzeldatei-messung/messung.tsv`
- [ ] Das Ergebnis steht als Zeile in den Entscheidungen: `grep -n 'Einzeldatei' docs/decisions.md`

## Phase 7: LAN-Modus als importierbares Paket

**Depends on**: 6

### Context

- `reverse-proxy/main.go — runLANMode()`, `runCaddy()`, `statePath`.
- `reverse-proxy/state.go — ensureState()`, `register.go — registerWithACMEDNS()`, `hostname.go — deriveHostname()`, `resolveLANIP()`, `probe.go — probeCert()`, `checkRebind()`.
- `reverse-proxy/status.go — decideStatus()`, `statuspage.go — statusListenAddr`, `newStatusServer()`.
- `reverse-proxy/caddyconf — RenderLAN()` aus Phase 2.

### What to build

- Neues Paket `reverse-proxy/lan` mit den Dateien aus dem Kontext samt Tests.
- Einstieg ist `lan.Run(ctx context.Context, opts Options, startCaddy func(ctx context.Context, caddyfile string) error) error`.
- `Options` trägt `StatePath`, `StatusAddr`, `LANIP`, `ACMEDNSURL`, `LEStaging`, `Upstream`, `AppRoot` und `StorageRoot`.
- `reverse-proxy/main.go — runLANMode()` ruft `lan.Run()` mit den Container-Werten (`/state/install.json`, `:8484`) und dem Kindprozess-Start über `runCaddy()`.
- Die Status-Seite lauscht auf `opts.StatusAddr`; Texte, Logik und QR-Code bleiben gleich.

### Acceptance criteria

- [ ] Das Proxy-Modul besteht sein Gate: `make check-local-proxy`
- [ ] Status-Seite, Install-State und Hostname verhalten sich unverändert: `cd reverse-proxy && go test -count=1 ./lan/`
- [ ] Die E2E-Suite bleibt grün: `E2E_STACK_PORT=8093 E2E_STACK_PROJECT=jotti-e2e-einzeldatei-7 make test-e2e`

## Phase 8: Grüne Adresse, Status-Seite und Firewall in der Einzeldatei

**Depends on**: 7

### Context

- `reverse-proxy/lan — Run()`, `Options` aus Phase 7.
- `windows/starter/host — DetectLANIP()`, `CheckPorts()`, `EnsureFirewallRule()` aus Phase 3.
- `windows/einzeldatei/main.go — run()`, `windows/einzeldatei/internal/proxy — Run()` aus Phase 4.
- `scripts/einzeldatei-smoke.sh`, `scripts/check-pins.sh`, `.github/workflows/ci.yml — einzeldatei-windows`.

### What to build

- `windows/einzeldatei/main.go — run()` ersetzt den direkten `RenderLAN()`-Aufruf durch `lan.Run()` mit `proxy.Run()` als Caddy-Start.
  Die Optionen: `StatePath <Daten>\install.json`, `StorageRoot <Daten>\caddy`, `StatusAddr 127.0.0.1:8484`, Upstream `127.0.0.1:13000` und `LANIP` aus `host.DetectLANIP()`.
  `ACMEDNS_BASE_URL` und `PROXY_LE_STAGING` kommen aus der Umgebung.
- Vor dem Start prüft `host.CheckPorts()` die Ports 80, 443, 8484, 13000 und 15432.
- `host.EnsureFirewallRule()` setzt die Regel `jotti-einzeldatei` für TCP 80/443 und das Programm im Programmordner.
- Die Konsole verweist auf `http://localhost:8484` wie der Starter.
- `scripts/check-pins.sh` gleicht die Caddy-, `acmedns`- und `caddy-ratelimit`-Versionen in `windows/einzeldatei/go.mod` mit `reverse-proxy/Dockerfile` ab.
- `scripts/einzeldatei-smoke.sh lan <exe>` läuft mit `ACMEDNS_BASE_URL=http://127.0.0.1:9`, also nur mit der Fallback-Adresse.
  - Die Status-Seite antwortet auf `127.0.0.1:8484`, nicht auf `<LAN-IP>:8484`.
  - `https://<LAN-IP>/api/health` antwortet mit der internen CA.
  - Port 13000 ist von der LAN-IP aus nicht erreichbar.
  - Die Firewall-Regel existiert.
- `einzeldatei-windows` ruft nach `start` und `messung` zusätzlich `lan` auf.

### Acceptance criteria

- [ ] Das Modul besteht sein Gate: `make check-einzeldatei`
- [ ] Postgres- und Caddy-Pins stimmen überein: `bash scripts/check-pins.sh`
- [ ] Shell-Skripte und Workflows bestehen ihre Prüfungen: `make check-shell check-workflows`
- [ ] Auf dem Windows-Runner bestehen `start`, `messung` mit Schwellen und `lan`: `id="$(gh run list -w CI -e push -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" && gh run watch "$id" --exit-status >/dev/null && gh run view "$id" --json jobs -q '.jobs[] | select(.name=="einzeldatei-windows") | .conclusion' | grep -x success`

## Phase 9: Relay als importierbares Paket

**Depends on**: 6

### Context

- `windows/relay/main.go — main()`, `loadConfigFromEnv()`, `poll()`, `meldeErgebnis()`, `fuehreZyklusAus()`, `druckeAuftrag()`.
- `windows/relay/env.go — envWithFileFallback()` — bleibt für das eigenständige Relay.

### What to build

- Neues Paket `windows/relay/relay` mit `LoadConfig(getenv func(string) string) (Config, error)` und `Run(ctx context.Context, cfg Config) error`.
- Poll-Zyklus, Druck und Rückmeldung ziehen samt Tests ins Paket; `Run()` endet mit dem Kontext, belegt durch den neuen Test `TestRun_EndetMitKontext`.
- `windows/relay/main.go` bleibt das eigenständige Relay: Env-Datei, Konsole, Enter zum Schließen.

### Acceptance criteria

- [ ] Das Relay-Modul besteht sein Gate: `make check-relay`
- [ ] Der Windows-Zweig kompiliert und besteht vet: `cd windows/relay && GOOS=windows go vet ./...`
- [ ] `Run()` endet, wenn der Kontext endet: `cd windows/relay && go test -count=1 -run '^TestRun_EndetMitKontext$' -v ./relay/ | grep -- '--- PASS: TestRun_EndetMitKontext '`

## Phase 10: Windows-Dienst mit eingebautem Relay

**Depends on**: 8, 9

### Context

- `windows/einzeldatei/main.go — run()` — Stufe-2-Ablauf, der in den Dienst wandert.
- `windows/einzeldatei/internal/postgres — Cluster`, `windows/einzeldatei/internal/schema — Migrate()`, `windows/einzeldatei/internal/proxy — Run()`.
- `windows/relay/relay — LoadConfig()`, `Run()` aus Phase 9.
- `windows/starter/core — StateDirSDDL()`, `ParseAdminOTP()`, `AdminCodeHinweis()`; `windows/starter/host — PrepareStateDir()`.
- `backend/bootstrap — Result.Log()` — Anmelde-Code im Dienst-Log.

### What to build

- `windows/einzeldatei/go.mod` bekommt `replace` auf `../relay` und `golang.org/x/sys` (offene Frage 6).
- `windows/einzeldatei/internal/service — Install(exePath string, env []string) error`, `Start()`, `Stop()`, `Status()`, `Remove()` über `svc/mgr`.
  Der Dienst `jotti` startet manuell, läuft unter `NT SERVICE\jotti` und erbt nur die weitergereichten Testvariablen im Registry-Wert `Environment`.
- `windows/einzeldatei/internal/service — Run()` ist der Handler unter `svc.Run`.
  Beim Start fährt er Postgres hoch, migriert und startet Backend, LAN-Modus und Relay mit `BackendURL http://127.0.0.1:13000`.
  Beim Stopp hält er Relay, Backend und Caddy an und fährt danach Postgres mit `pg_ctl stop -m fast` herunter.
- Die Aufsicht beobachtet den Postmaster über die PID aus `postmaster.pid` und startet ihn nach einem Absturz neu.
  Nach fünf Neustarts in zehn Minuten stoppt der Dienst mit Fehler im Log.
- `windows/einzeldatei/internal/logfile — New(dir string, maxBytes int64, keep int) io.WriteCloser` rotiert `logs\jotti.log` mit 5 MB und drei Dateien.
  Der Dienst setzt `slog` auf JSON in diese Datei und schreibt bei jedem Start eine Startmarke. Neuer Test `TestLogfile_Rotation`.
- `windows/starter/core — StateDirSDDL()` nimmt zusätzliche SIDs; `host.PrepareStateDir()` gibt dem Dienstkonto Änderungsrechte auf das Datenverzeichnis.
  Neuer Test `TestStateDirSDDL` in `windows/starter/core/statedir_test.go` prüft die SDDL mit und ohne Dienstkonto.
- Konsole beim ersten Doppelklick:
  - kopiert sich in den Programmordner, entpackt die Assets, registriert den Dienst, setzt Rechte und Firewall-Regel;
  - startet den Dienst und wartet auf `https://127.0.0.1/api/health`;
  - zeigt den Anmelde-Code aus dem Log seit der letzten Startmarke über `core.ParseAdminOTP()` und verweist auf die Status-Seite.
- Konsole bei laufendem Dienst: Status mit Version, dann `1` Beenden, Enter schließt nur das Fenster.
- Läuft ein Prozess `jotti-relay.exe`, warnt die Konsole vor doppeltem Druck.
- `scripts/einzeldatei-smoke.sh dienst <exe>` prüft auf dem Runner:
  - Der Dienst läuft (`sc query jotti`) und `/api/health` antwortet, nachdem das Konsolenfenster zu ist.
  - Anmelden mit dem Code aus dem Log, Passwort setzen und einen Direktverkauf buchen wie `ops-smoke.sh — step_sale()`.
  - Nach `taskkill //F //IM postgres.exe` ist `/api/health` binnen 30 s wieder grün.
  - Menüpunkt `1` stoppt den Dienst; danach läuft kein `postgres.exe` mehr.
- Die Modi `messung` und `lan` laufen gegen den Dienst weiter.

### Acceptance criteria

- [ ] Einzeldatei und Starter bestehen ihre Gates: `make check-einzeldatei check-starter`
- [ ] Die SDDL enthält das Dienstkonto mit Änderungsrechten und schützt das Verzeichnis: `cd windows/starter && go test -count=1 -run '^TestStateDirSDDL$' -v ./core/ | grep -- '--- PASS: TestStateDirSDDL '`
- [ ] Die Rotation hält höchstens drei Dateien und schneidet bei 5 MB: `cd windows/einzeldatei && go test -count=1 -run '^TestLogfile_Rotation$' -v ./internal/logfile/ | grep -- '--- PASS: TestLogfile_Rotation '`
- [ ] Shell-Skripte und Workflows bestehen ihre Prüfungen: `make check-shell check-workflows`
- [ ] Auf dem Windows-Runner bestehen `messung`, `lan` und `dienst`: `id="$(gh run list -w CI -e push -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" && gh run watch "$id" --exit-status >/dev/null && gh run view "$id" --json jobs -q '.jobs[] | select(.name=="einzeldatei-windows") | .conclusion' | grep -x success`

## Phase 11: Update, Backup, Zurückspielen und Entfernen

**Depends on**: 10

### Context

- `windows/starter/core — ShouldBackup()`, `DumpsToDelete()`, `IsDowngrade()`, `IsNewerVersion()`.
- `windows/einzeldatei/internal/service`, `windows/einzeldatei/internal/postgres — Cluster` aus Phase 10.
- `.github/workflows/release.yml` — Windows-Release-ZIP und `gh release create`.
- `Makefile — build-einzeldatei-windows`.

### What to build

- `windows/einzeldatei/internal/layout — PlanUpdate(installed, own string) UpdateAction` entscheidet zwischen Einrichten, Menü, Update und Ablehnung (Downgrade); neuer Test `TestPlanUpdate`.
- Update:
  - Eine neuere Datei erkennt die installierte Version und fragt einmal nach.
  - Sie stoppt den Dienst und richtet `<Programmordner>\<neu>\` ein.
  - Sie biegt Dienst-Binary und Firewall-Regel auf die neue Version um und startet den Dienst.
  - Ältere Versionen als die vorige löscht sie.
- Der Dienst zieht beim Start mit `core.ShouldBackup()` per `pg_dump -Fc` ein Backup `backups\jotti-<alte-version>-<zeit>.dump` und behält fünf.
  Scheitert das Backup, stoppt er; die Konsole meldet es, biegt den Dienst auf die vorige Version zurück und startet sie.
- `postgres.Cluster — Dump()`, `Restore()` über `pg_dump` und `pg_restore` aus dem Programmordner.
- Menü `2` Backup zurückspielen:
  - Die Konsole zeigt das jüngste Backup, fragt nach und stoppt den Dienst.
  - Sie schreibt die Marke `zurueckspielen` mit Dump und Version und biegt den Dienst auf die Version aus dem Dateinamen des Dumps.
  - Der Dienst benennt `pgdata` in `pgdata-verworfen-<zeit>` um, legt einen leeren Cluster an, spielt den Dump ein und setzt `last-version` auf dessen Version.
  - Liegt die Version nicht im Programmordner, verweigert die Konsole mit Download-Hinweis.
- Menü `3` Entfernen löscht Dienst, Firewall-Regel und Programmordner, nie das Datenverzeichnis, und sagt das ausdrücklich.
- Der Update-Hinweis aus `host.NotifyIfUpdateAvailable()` erscheint nach dem Start; aktualisiert wird nie selbst.
- `release.yml` baut `jotti-<version>.exe` mit `make build-einzeldatei-windows` auch im Probelauf; nur beim Tag-Push hängt es die Datei neben dem ZIP an (offene Frage 7).
  Nach dem letzten Push startet `gh workflow run release.yml --ref <Branch>` den Probelauf.
- `einzeldatei-build` baut zusätzlich `VERSION=v0.0.2-ci`.
  `scripts/einzeldatei-smoke.sh betrieb <alt.exe> <neu.exe>` prüft auf dem Runner:
  - Installation der alten Version und eine Buchung.
  - Update auf die neue Version: Backup im Ordner, `/api/health` meldet die neue Version.
  - Die alte Datei verweigert danach den Start.
  - Zurückspielen: Die alte Version läuft, die Buchung ist da, `pgdata-verworfen-*` existiert.
  - Entfernen: Dienst und Programmordner sind weg, das Datenverzeichnis bleibt.

### Acceptance criteria

- [ ] Einzeldatei und Starter bestehen ihre Gates: `make check-einzeldatei check-starter`
- [ ] Die Update-Entscheidung aus installierter und eigener Version ist richtig: `cd windows/einzeldatei && go test -count=1 -run '^TestPlanUpdate$' -v ./internal/layout/ | grep -- '--- PASS: TestPlanUpdate '`
- [ ] Workflows und Shell-Skripte bestehen ihre Prüfungen: `make check-workflows check-shell`
- [ ] Der Release-Probelauf baut die Datei: `gh run watch "$(gh run list -w release.yml -e workflow_dispatch -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" --exit-status`
- [ ] Auf dem Windows-Runner bestehen alle Modi bis `betrieb`: `id="$(gh run list -w CI -e push -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" && gh run watch "$id" --exit-status >/dev/null && gh run view "$id" --json jobs -q '.jobs[] | select(.name=="einzeldatei-windows") | .conclusion' | grep -x success`

## Phase 12: Übernahme einer Docker-Installation

**Depends on**: 11

### Context

- `windows/starter/stack.go` — `postgresContainer`, `reverseProxyContainer`, `postgresDataVolume`, `configVolume`.
- `windows/starter/backup.go — dumpDatabase()`, `configvolume.go — readConfigVolume()` — `docker exec` und `docker cp` als Vorbild.
- `windows/einzeldatei/internal/layout — DockerInstallation()` aus Phase 4 — wird vom Abbruch zum Angebot.
- `database/migrations/01_initial.up.sql` — `kassensitzungen.status`, `kassenidentitaet.seriennummer`, `kassenjournal`.
- `scripts/test-integration.sh`, `scripts/lib.sh — start_throwaway_postgres()`, `backend/seed — Run()`.

### What to build

- `windows/einzeldatei/internal/uebernahme`:
  - `Erkenne()` findet `jotti-local_postgres-data` per `docker volume inspect`; ohne laufenden Daemon bittet es, Docker Desktop zu starten.
  - `PruefeOffeneSitzung()` liest per `docker exec ... psql` die Kassensitzungen mit `status <> 'abgeschlossen'` und verweigert bei einer offenen.
  - `Kennzahlen(db) (Kennzahlen, error)` liest Schema-Version aus `schema_migrations`, Zeilenzahl je Tabelle im Schema `public`, höchste `kassenjournal`-ID und `kassenidentitaet.seriennummer`.
  - `Abgleich(quelle, ziel Kennzahlen) []Abweichung` ist eine reine Funktion; neuer Test `TestAbgleich`.
- Ablauf in der Konsole, wenn die Erkennung anschlägt:
  - Sie startet `jotti-postgres-local`, liest die Quell-Kennzahlen und zieht `pg_dump -Fc` nach `backups\uebernahme-<zeit>.dump`.
  - Sie kopiert `/data/caddy` und `/state/install.json` per `docker cp` aus `jotti-reverse-proxy-local` und gleicht `.env` mit `/jotti-config/.env` ab.
  - Sie stoppt die vier Container, warnt vor einem laufenden `jotti-relay.exe` und richtet den Dienst mit der Marke `uebernahme` ein.
- Der Dienst legt bei dieser Marke den Cluster an, spielt den Dump ein, liest die Ziel-Kennzahlen und gleicht ab.
  Bei einer Abweichung benennt er den eingebetteten Cluster in `pgdata-verworfen-<zeit>` um und stoppt; die Konsole startet die Docker-Container wieder und nennt die Abweichung.
  Docker-Volumes bleiben in jedem Fall stehen; der nächste Start bietet die Übernahme erneut an.
- `scripts/test-integration.sh` führt danach die Integrationstests von `windows/einzeldatei` aus, mit denselben Variablen, `TEST_RUN` und `TEST_PG_CONTAINER`.
- `windows/einzeldatei/internal/uebernahme/uebernahme_integration_test.go` (`//go:build integration`), Test `TestUebernahme_DemoSzenario`:
  - spielt `seed.Run()` in die Datenbank `jotti` ein, dumpt per `docker exec`, spielt nach `jotti_ziel` ein;
  - erwartet einen leeren Abgleich und dieselbe Seriennummer.

### Acceptance criteria

- [ ] Das Modul besteht sein Gate: `make check-einzeldatei`
- [ ] Abweichende Zeilenzahl, Journal-ID, Seriennummer oder Schema-Version ergeben je eine Abweichung, gleiche Werte keine: `cd windows/einzeldatei && go test -count=1 -run '^TestAbgleich$' -v ./internal/uebernahme/ | grep -- '--- PASS: TestAbgleich '`
- [ ] Demo-Szenario dumpen, einspielen, Abgleich leer: `TEST_PG_CONTAINER=jotti-pg-einzeldatei TEST_PG_PORT=5447 TEST_RUN='^TestUebernahme_DemoSzenario$' make check-integration | grep -- '--- PASS: TestUebernahme_DemoSzenario '`
- [ ] Die übrigen Integrationstests bleiben grün: `TEST_PG_CONTAINER=jotti-pg-einzeldatei TEST_PG_PORT=5447 make check-integration`
- [ ] Shell-Skripte bestehen ihre Prüfung: `make check-shell`
- [ ] Der Windows-Runner bleibt grün: `id="$(gh run list -w CI -e push -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" && gh run watch "$id" --exit-status >/dev/null && gh run view "$id" --json jobs -q '.jobs[] | select(.name=="einzeldatei-windows") | .conclusion' | grep -x success`

## Phase 13: Dokumentation und Code-Signatur

**Depends on**: 12

### Context

- `docs/leitfaden/installation.md`, `aktualisieren.md`, `aktualisieren-backups.md`, `datenaufbewahrung.md`, `fehlersuche.md`, `haeufige-fragen.md`, `veranstaltungstag.md` — nennen heute `jotti-start.exe` und Docker Desktop.
- `packaging/windows/KURZANLEITUNG.md`, `docs/verfahrensdokumentation.md`, `docs/handbuch.md`, `docs/language.md`.
- `docs/decisions.md` — D15 (Postgres-Pins), D13 (Compose-Dateien).
- `.github/workflows/release.yml` — Windows-Assets.

### What to build

- Der Leitfaden beschreibt `jotti.exe` als Windows-Weg.
  Installation, Update, Zurückspielen, Entfernen, Übernahme, Datenverzeichnis, Fehlersuche (Ports, Firewall, Defender, Dienst-Log) und der Hinweis auf `jotti-relay.exe` gehören dazu.
  Das Docker-ZIP steht dort als Übergangsweg bis zu seinem Wegfall.
- Die Kurzanleitung im ZIP verweist auf `jotti.exe` und die Übernahme.
- Die Verfahrensdokumentation beschreibt im Systemüberblick die Einzeldatei; das Handbuch beschreibt im Betriebsteil Dienst, Datenverzeichnis und Backups.
- `docs/language.md` bekommt Einzeldatei und Übernahme mit Definition und zu vermeidenden Wörtern aus dem PRD.
- D15 wird umformuliert: Die Pins gelten auch für die eingebettete Postgres-Version, und weder Linux noch Windows hat einen Major-Upgrade-Pfad.
  Die Zeile nennt keine Windows-Installationsart, damit Phase 14 sie nicht erneut ändern muss.
- Code-Signatur nach Nicos Entscheidung, festgehalten als neue Entscheidungszeile:
  - Bei Kauf signiert `release.yml` `jotti-<version>.exe` mit dem Zertifikat aus einem Repository-Secret, bevor es angehängt wird.
  - Bei „unsigniert" erklärt der Leitfaden den SmartScreen-Dialog „Weitere Informationen", dann „Trotzdem ausführen".
- Nach dem letzten Push startet `gh workflow run release.yml --ref <Branch>` den Probelauf.

### Acceptance criteria

- [ ] Die Signatur-Entscheidung steht in den Entscheidungen: `grep -n 'Code-Signatur' docs/decisions.md`
- [ ] Installation, Update, Aufbewahrung und Fehlersuche beschreiben `jotti.exe`: `for f in installation aktualisieren datenaufbewahrung fehlersuche; do grep -q 'jotti.exe' "docs/leitfaden/$f.md" || exit 1; done`
- [ ] Doku-Gates, Links, Sprache und Format sind grün: `make check-repo check-format`
- [ ] Die Website baut mit den geänderten Seiten: `make website-check`
- [ ] Der Release-Probelauf ist grün: `gh run watch "$(gh run list -w release.yml -e workflow_dispatch -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" --exit-status`

## Phase 14: Docker-ZIP entfernen

**Depends on**: 13

### Context

- `docker-compose.release.yml`, `packaging/windows/jotti-stop.cmd`, `jotti-restore.cmd`, `jotti-repair.cmd`, `KURZANLEITUNG.md`.
- `windows/starter` — ganzes Modul; `core` und `host` nutzt die Einzeldatei.
- `Makefile — release-windows`, `build-starter-windows`, `starter-syso`, `check-starter`, `check`, `GO_MODULES`.
- `.github/workflows/ci.yml — changes.windows`, `windows-ci`; `release.yml` — ZIP und Smoke-Test auf der Release-Compose; `security-scans.yml — vuln-scan-go`.
- `scripts/check-compose-twins.sh`, `scripts/check-language.sh` (Regel 1b), `scripts/check-pins.sh`, `docs/decisions.md — D13`.

### What to build

- `windows/starter/core` und `windows/starter/host` ziehen samt Tests nach `windows/einzeldatei/internal/core` und `internal/host`; das Modul `windows/starter` entfällt samt `go.work`-Eintrag und `replace`.
- `docker-compose.release.yml`, die drei `.cmd`-Skripte, die Kurzanleitung und `scripts/check-compose-twins.sh` entfallen.
  Regel 1b der Sprachprüfung entfällt mit den `.cmd`-Dateien.
- Makefile: `release-windows`, `build-starter-windows`, `starter-syso` und `check-starter` entfallen; `GO_MODULES` und `check` folgen.
- CI: `windows-ci` prüft nur noch `relay`, der Filter `windows` verliert Compose- und Packaging-Pfade, `security-scans.yml` verliert den Starter.
- `release.yml` hängt nur noch `jotti-<version>.exe` an und testet die Images mit `docker-compose.local.yml` und `--no-build` nach Umtaggen (offene Frage 10).
- D13 wird umformuliert: Das Beispiel für die Matrix nennt nur noch das Windows-Relay.
- Leitfaden und Handbuch verlieren den Übergangsweg über das ZIP; die Übernahme bleibt beschrieben.
- Nach dem letzten Push startet `gh workflow run release.yml --ref <Branch>` den Probelauf.

### Acceptance criteria

- [ ] Mindestens drei Release-Tags enthalten die Übernahme, also zwei nach ihrem ersten Release: `git fetch -q --tags && test "$(git tag --contains "$(git log --format=%H --diff-filter=A -- windows/einzeldatei/internal/uebernahme | tail -1)" | grep -c '^v')" -ge 3`
- [ ] Kein Verweis auf entfernte Dateien bleibt: `! git grep -n -e 'docker-compose.release.yml' -e 'jotti-start' -e 'jotti-repair' -e 'jotti-restore.cmd' -e 'jotti-stop.cmd' -- ':!CHANGELOG.md' ':!docs/plans/'`
- [ ] Die Einzeldatei besteht ihr Gate mit den übernommenen Paketen: `make check-einzeldatei check-relay`
- [ ] Repo-Gates und Workflows sind grün: `make check-repo check-format check-workflows`
- [ ] Der Release-Probelauf ist grün: `gh run watch "$(gh run list -w release.yml -e workflow_dispatch -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" --exit-status`
- [ ] Der Windows-Runner bleibt grün: `id="$(gh run list -w CI -e push -c "$(git rev-parse HEAD)" -L 1 --json databaseId -q '.[0].databaseId')" && gh run watch "$id" --exit-status >/dev/null && gh run view "$id" --json jobs -q '.jobs[] | select(.name=="einzeldatei-windows") | .conclusion' | grep -x success`
