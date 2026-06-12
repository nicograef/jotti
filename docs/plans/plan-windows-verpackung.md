# Plan: Klickbare Windows-Verpackung für den lokalen Betrieb

> Source PRD: `docs/prds/prd-windows-verpackung.md`
>
> Revision 2026-06-12 (2. Iteration): Ports fest auf 80/443 (KISS — Phase
> „Compose-Port-Parameterisierung" und Relay-Port-Kopplung gestrichen, Starter
> braucht keinen `.env`-Parser mehr); Release-/Versionierungskonzept ergänzt
> (produktweite SemVer-Tags, trunk-based, Tag-Push als einziges Publish-Gate).
>
> PRD am 2026-06-12 entsprechend aktualisiert (GHCR-Distribution, feste Ports
> 80/443, Release per Version-Tag) — Plan und PRD sind deckungsgleich.

## Goal

Ein Vereins-Admin kann jotti auf einem Windows-Rechner per Doppelklick starten (`jotti-start.exe`) und den Bondruck separat per Doppelklick aktivieren (`jotti-relay.exe`) — ohne `.env` von Hand anlegen, Geheimnisse erzeugen oder Kommandozeile benutzen zu müssen. Verteilt als GitHub-Release-ZIP mit vorgebauten Windows-Binaries und vorgebauten Container-Images, sodass das ZIP **ohne Quellcode** lauffähig ist. Releases entstehen ausschließlich durch produktweite Version-Tags (`v0.X.Y`) auf `main`.

## Architectural decisions

### Verpackung & Laufzeit

- **Quellort Starter:** `cmd/starter/` als eigenständiges Go-Modul (eigenes `go.mod`, module `jotti-starter`, go 1.26.0, keine Dependencies — analog `cmd/relay/`).
- **Distribution mit vorgebauten Images:** `docker-compose.local.yml` hat drei Build-Kontexte (`migrate`, `backend`, `frontend`) — ein ZIP ohne Quellbaum kann nicht `--build`en. Der Release-Workflow pusht deshalb `ghcr.io/nicograef/jotti-{backend,frontend,migrate}` (Repo ist public, Pull ohne Login) und das ZIP enthält ein `docker-compose.release.yml` mit beim Release fest gepinntem Image-Tag. `database/migrations/` liegt im ZIP und wird wie bisher in den migrate-Container gemountet (kein Dockerfile-Umbau).
- **Ports fest 80/443 (KISS):** keine Port-Variablen, keine Compose-Änderung, keine `.env`-Port-Keys. Ist ein Port belegt, nennt die Preflight-Diagnose die typischen Verursacher (VMware Workstation, IIS, Skype) und fordert auf, das Programm zu beenden — ohne Prozess-Lookup (kein netstat-Parsing). Bewusst akzeptiertes Restrisiko; sollte der Fall in der Praxis häufig auftreten, ist Port-Konfigurierbarkeit der dokumentierte Eskalationspfad.
- **Health-Check:** `GET https://localhost/api/health` (TLS-Verify aus, selbstsigniert). nginx proxyt nur `/api/*` und `/` — ein Check auf `/health` würde von der Frontend-SPA mit 200 beantwortet (False-Positive). Nur HTTP 200 gilt als bereit; das Backend liefert 503 „degraded", solange die DB nicht antwortet.
- **Secrets als Hex:** 32 Bytes aus `crypto/rand`, hex-kodiert (64 Zeichen) — identisch zu `openssl rand -hex 32` in `scripts/init-env.sh`. Base64-Zeichen (`+/=`) würden die Postgres-URL im migrate-CMD (`postgres://user:pass@…`) brechen.
- **Konsolen-UX:** Ein per Doppelklick gestartetes Konsolenprogramm schließt sein Fenster beim Exit sofort. Der Starter streamt deshalb die Compose-Ausgabe live (Pull sichtbar statt eingefrorenem Fenster) und wartet am Ende — bei Erfolg **und** Fehler — auf Enter, sonst verschwindet die angezeigte URL sofort wieder. Das Relay pausiert bei Konfigurationsfehlern vor dem Exit.
- **`.env`-Parsing nur im Relay:** Der Starter schreibt die `.env` nur (liest sie nie — Ports sind fest, Compose liest sie selbst). Das Relay bekommt einen minimalen Key=Value-Parser (~20 Zeilen, kein Shared-Package), der CRLF (Notepad-Edit), UTF-8-BOM und optionale Anführungszeichen toleriert; bereits gesetzte Env-Variablen haben Vorrang vor der Datei.
- **Firewall-Hinweis:** immer Teil der Erfolgsausgabe (kein programmatischer netsh-Check).
- **Kein Code-Signing** (out of scope); SmartScreen-/Defender-Hinweis gehört in die Kurzanleitung.

### Versionierung & Release (trunk-based)

- **Produktweite SemVer-Tags `v0.X.Y`** auf `main` sind der einzige Release-Mechanismus und markieren zugleich den Prod-Stand in der Git-History. Minor = Feature-Release, Patch = wichtiger Bugfix. Die `0.x` spiegelt den Pre-Release-Status (AGENTS.md). Es existieren noch keine Tags; das erste Release wird `v0.1.0`.
- **Wann wird gebaut:** Der teure Build (3 Images + 2 Exes + ZIP + Smoke-Test) läuft **nur** bei Tag-Push `v*` — das Tag-Setzen ist die bewusste menschliche Entscheidung „das ist ein Release". Zusätzlich `workflow_dispatch` als **Dry-Run**: baut und smoke-testet alles, publiziert aber nichts (kein Image-Push, kein GitHub-Release); Versionsstring dann `dev-<shortsha>`.
- **Wo wird gebaut:** Images ausschließlich im Workflow (buildx auf dem Runner, `--load` für den Smoke-Test, Push **erst nach** bestandenem Smoke-Test). Exes im Workflow per Make-Target; dasselbe Target funktioniert lokal (Cross-Compile ist billig) — `make release-windows VERSION=…` baut lokal nur Exes + ZIP, nie Images.
- **Kein Test-Rerun im Release:** Bei trunk-based Development zeigt der Tag auf einen Commit, dessen CI auf `main` bereits grün war. Der Release-Workflow prüft nur, was die CI nicht abdeckt: den Stack aus der Release-Compose-Datei hochfahren und `/api/health` abfragen.
- **Image-Tags:** exakt `vX.Y.Z`, **kein `:latest`** (verhindert Drift bei späteren Server-Deploys). Images tragen das OCI-Label `org.opencontainers.image.version`; die Exes bekommen die Version per ldflags einkompiliert und geben sie beim Start aus.
- **Release-Notes:** auto-generiert (`generate_release_notes: true`) — Conventional Commits machen die Commit-Liste lesbar, Pflegeaufwand null.
- **CI-Struktur:** `ci.yml` bleibt eine Datei (der paths-filter modularisiert bereits pro Bereich) und erhält nur den neuen `cmd`-Job; das Release wird eine eigene Datei `release.yml`. Kein Aufsplitten, keine Reusable Workflows — CI (Tests) und Release (Image-Builds) teilen keine teure Logik.
- **Server-Deploys (rocks/prod)** können später auf dieselben GHCR-Images umstellen — die Image-Namen sind bewusst neutral. Die Umstellung selbst ist nicht Teil dieses Plans.

## Inventory

- `cmd/relay/main.go:36` — `defaultBackendURL = "https://localhost/api"` (passt unverändert, Ports bleiben fest).
- `cmd/relay/main.go:66-104` — `loadConfigFromEnv(getenv)`; deutsche Pflichtfeld-Meldung für `RELAY_AUTH_TOKEN` existiert bereits (Z. 69).
- `cmd/relay/main.go:93-96` — TLS-Verify wird für die lokale Default-URL bereits automatisch deaktiviert.
- `cmd/relay/go.mod` — module `jotti-relay`, go 1.26.0, keine Dependencies; Tests in `main_test.go` ohne Build-Tag.
- `scripts/init-env.sh:20-23` — Idempotenz-Verhalten (vorhandene `.env` → Exit ohne Änderung); `:33-39` — Secrets via `openssl rand -hex 32`. Wird im Release-Workflow für die Smoke-Test-`.env` wiederverwendet (`make init`).
- `.env.example:1-11` — die vier Keys (`POSTGRES_USER`, `POSTGRES_PASSWORD`, `JWT_SECRET`, `RELAY_AUTH_TOKEN`); bleibt unverändert.
- `docker-compose.local.yml:16` — `name: jotti-local` (stabiler Projektname → Volume überlebt Ordner-Umbenennung und ZIP-Upgrades); `:37-52, 54-57, 88-91` — die drei Build-Kontexte; `:106-108` — Ports `80:80`/`443:443` bleiben hartkodiert.
- `reverse-proxy/nginx.local.conf:44-54` — `/api/`-Proxy mit Prefix-Strip → `/api/health` erreicht den Backend-Handler `/health`.
- `reverse-proxy/local-entrypoint.sh:72-84` — Zertifikat wird bei IP-Wechsel neu erzeugt, aber nur beim Entrypoint-Lauf — deshalb force-recreate des Proxys beim Start (wie `make local-up`).
- `backend/api/health/health.go:28-58` — Health-Handler: GET, DB-Ping, 503 bei degraded.
- `Makefile:22-23` — `make init`; `:91-92` — `build-relay`; `:152-154` — `local-up` inkl. `--force-recreate reverse-proxy`.
- `.github/workflows/ci.yml:23-32` — paths-filter **ohne** `cmd/**` (für Relay/Starter läuft heute keine CI); `:42-45` — checkout@v6, setup-go@v6, go 1.26.0 als Versionsreferenz für `release.yml`.
- `database/migrate/Dockerfile` — migrate-Image ohne eingebackene Migrationen; der Mount bleibt auch im Release-Compose.
- `docs/betrieb/leitfaden-hosting.md:42-128` — Weg A (manueller CLI-Ablauf), wird um Starter-Abschnitt ergänzt.

## Resolved decisions

- **Keine konfigurierbaren Ports** (Nutzer-Entscheid, KISS): Die frühere Phase „Compose-Port-Parameterisierung" und die Relay-Port-Kopplung entfallen ersatzlos. Dadurch: kein `.env`-Parser im Starter, `.env.example` unverändert, Relay-Default-URL unverändert.
- **GHCR statt Quellcode-im-ZIP:** Ein `docker build` auf dem Vereins-PC wäre der fragilste Schritt des ganzen Ablaufs (Go-/npm-Downloads während des Builds, zweistellige Minuten auf Altgeräten). Ein Image-Pull ist ein einzelner robuster Download.
- **SemVer `v0.X.Y` produktweit, Tag-Push + Dry-Run, nur Smoke-Test als Gate, keine `:latest`-Tags, Version in Exes (ldflags) und Image-Labels, Release-Notes auto-generiert, `ci.yml` + `release.yml`** — Klärungsrunden vom 2026-06-12.
- **Compose-Datei-Auflösung im Starter:** `docker-compose.release.yml` neben der Exe (ZIP-Fall), sonst `docker-compose.local.yml` neben Exe bzw. im Arbeitsverzeichnis (Dev im Repo).
- **Day-2/Autostart:** Die Container haben `restart: unless-stopped` — am zweiten Festtag läuft der Stack u. U. schon, sobald Docker Desktop startet. Hält der **eigene** reverse-proxy den Port, ist das kein Preflight-Fehler, sondern der Erfolgs-Pfad (Start ist idempotent).
- **Starter spiegelt `make local-up`** inklusive `--force-recreate reverse-proxy`, damit der Zertifikats-Check bei IP-Wechsel tatsächlich läuft.
- **LAN-IP primär über die Outbound-Route** (UDP-„Connect" zu einer externen Adresse, es wird kein Paket gesendet; `LocalAddr` liefert die IP des Default-Route-Interfaces). „Erste private IPv4" wäre auf Windows-Rechnern mit Docker Desktop falsch — vEthernet-/WSL-Adapter tragen private 172.x-Adressen, die Smartphones nicht erreichen. Interface-Heuristik nur als Fallback (192.168.x vor 10.x vor 172.16-31.x).
- **Tests in `cmd/starter` ohne Build-Tag** — wie `cmd/relay`. Bewusste Abweichung von der PRD-Notation `-tags=unit`: In den cmd-Modulen gibt es keine Integrationstests, der Tag hätte keinen Trennzweck. Neuer CI-Job für `cmd/**` (deckt damit auch das bisher CI-lose Relay ab).
- **Health-Timeout 120 s** nach Rückkehr von `up -d` (Erststart: Postgres-Init + Migrationen auf Altgeräten; der Image-Pull blockiert davor sichtbar im `up`-Aufruf selbst).
- **`jotti-stop.cmd`** (eine Zeile `docker compose … down` + `pause`) liegt im ZIP — Story 19 „sauber beenden" per Doppelklick; Docker-Desktop-GUI bleibt als Alternative dokumentiert.

## Open questions / Risks

- **Docker-CLI-Pfad auf Windows:** `docker` muss im PATH liegen (Docker Desktop installiert nach `C:\Program Files\Docker\Docker\resources\bin`). Die Preflight-Diagnose nennt den Pfad.
- **SmartScreen/Defender:** Unsignierte `.exe`s lösen beim ersten Start SmartScreen aus („Weitere Informationen → Trotzdem ausführen"). Code-Signing bleibt out of scope; die Kurzanleitung beschreibt den Klickweg.
- **Port 80/443 belegt = Dead-End ohne Programmbeendigung:** bewusst akzeptiert (KISS). Eskalationspfad, falls praxisrelevant: Port-Konfigurierbarkeit nachrüsten.
- **Compose-Drift local vs. release:** Zwei Compose-Dateien können auseinanderlaufen. Der Smoke-Test im Release-Workflow fängt funktionale Drift ab.
- **DHCP-IP-Wechsel:** Am nächsten Festtag kann der Rechner eine neue LAN-IP haben → neue URL für die Handys + erneute Zertifikatswarnung pro Gerät. Der Starter zeigt immer die aktuelle URL; Kurzanleitung erwähnt es (optionaler Tipp im Leitfaden: DHCP-Reservierung im Router).
- **Erststart braucht Internet** (Image-Pull). Folgetage laufen aus dem lokalen Image-Cache. Kurzanleitung: Erststart vorab zuhause durchführen, nicht erst auf dem Fest.

---

## Phase 1: Starter-Core (reine Logik + Unit-Tests + CI)

**User stories**: 2, 3, 4, 6, 8, 11, 12¹, 14 — ¹ Story 12 in der Fassung „verständliche Meldung bei belegtem Port" (PRD-Folgeänderung: ohne Port-Wechsel-Hinweis).

### Context

- `scripts/init-env.sh:20-39` — Idempotenz-Verhalten und Hex-Secret-Erzeugung als Referenz; die Go-Implementierung nutzt `crypto/rand` direkt.
- `.env.example:1-11` — die vier zu erzeugenden Keys.
- `cmd/relay/go.mod` + `cmd/relay/main_test.go` — Muster für Modulzuschnitt und ungetaggte Tests.
- `.github/workflows/ci.yml:23-32` — paths-filter, dem `cmd/**` fehlt.

### What to build

Ein Go-Package `cmd/starter/core` mit seiteneffektfreien, unit-testbaren Funktionen:

1. **Secret-Erzeugung** — `GenerateSecret() string`: 32 Bytes aus `crypto/rand`, hex-kodiert (64 Zeichen, kompatibel zur Postgres-URL im migrate-CMD).
2. **`.env`-Materialisierung** — erzeugt die `.env` nur, wenn sie fehlt (Dateizugriffe injiziert, testbar); Inhalt: deutscher Kommentar-Header („nichts hier muss geändert werden"; schützt nebenbei vor einem späteren Notepad-BOM auf der ersten Key-Zeile), `POSTGRES_USER=admin`, `POSTGRES_PASSWORD`/`JWT_SECRET`/`RELAY_AUTH_TOKEN` generiert. Vorhandene Datei wird nie überschrieben.
3. **Preflight-Auswertung** — bildet Prüfergebnisse auf deutsche Diagnosen mit Handlungshinweis ab: Docker-CLI fehlt (nennt den Docker-Desktop-Pfad), Daemon antwortet nicht („Docker Desktop starten und warten, bis der Wal grün ist"), Daemon im Windows-Container-Modus („auf Linux-Container umschalten"), Port 80/443 durch Fremdprozess belegt (typische Verursacher nennen, Programm beenden, jotti danach erneut starten).
4. **LAN-IP-Auswahl** — `SelectLANIP(outboundIP string, interfaces []NetInterface) (string, error)`: bevorzugt die Outbound-Route-IP, wenn sie privat (RFC 1918) ist; Fallback-Heuristik über die Interfaces mit Präferenz 192.168.x > 10.x > 172.16-31.x; ignoriert Loopback und Link-Local (169.254.x).
5. **Zugriffs-URL-Bau** — `BuildAccessURL(ip string) string`: `https://{ip}` (Port fest 443, kein Suffix).

Dazu CI: paths-filter in `ci.yml` um `cmd: 'cmd/**'` erweitern und einen Job `cmd-ci` ergänzen (setup-go 1.26.0; `go vet`, `go build ./...`, `go test ./...` jeweils in `cmd/relay` und `cmd/starter`).

### Acceptance criteria

- [ ] `cmd/starter/` existiert mit eigenem `go.mod` (module `jotti-starter`, go 1.26.0, keine Dependencies).
- [ ] Secrets: 64 Hex-Zeichen; zwei Aufrufe liefern unterschiedliche Werte.
- [ ] `.env`-Idempotenz: vorhandene Datei wird nie überschrieben; fehlt sie, enthält sie die vier Keys aus `.env.example` und den Kommentar-Header.
- [ ] LAN-IP: Outbound-IP wird bevorzugt; eine 172.x-Adresse (vEthernet/WSL-Muster) gewinnt nie gegen eine 192.168.x; Loopback/169.254.x werden ignoriert.
- [ ] Preflight: jede Bedingung bildet auf eine deutsche Diagnose mit Handlungshinweis ab.
- [ ] `go build ./...` und `go test ./...` in `cmd/starter/` erfolgreich.
- [ ] `ci.yml` führt den neuen `cmd-ci`-Job bei Änderungen unter `cmd/**` aus (greift damit auch für `cmd/relay`).

---

## Phase 2: Starter-Shell + Binary

**User stories**: 7, 8, 9, 13, 14

### Context

- `cmd/starter/core` (Phase 1) — liefert die reine Logik.
- `Makefile:152-154` — `local-up` inkl. `--force-recreate reverse-proxy` als zu spiegelndes Verhalten.
- `reverse-proxy/local-entrypoint.sh:72-84` — warum das force-recreate nötig ist (Zertifikat-Regenerierung bei IP-Wechsel läuft nur im Entrypoint).
- `backend/api/health/health.go:28-58` + `reverse-proxy/nginx.local.conf:44-54` — Health über `/api/health` (Prefix-Strip), nur 200 = bereit.

### What to build

Die `main.go` in `cmd/starter/` verbindet Core-Logik mit echten Seiteneffekten:

1. **Compose-Datei finden:** `docker-compose.release.yml` neben der Exe (ZIP), sonst `docker-compose.local.yml` neben Exe bzw. im Arbeitsverzeichnis (Repo-Dev). Nichts gefunden → deutsche Diagnose.
2. `.env` neben der Compose-Datei materialisieren (Core).
3. **Preflights:** `docker` via `exec.LookPath`; `docker info -f '{{.OSType}}'` prüft Daemon-Erreichbarkeit und Container-Modus in einem Aufruf; Port-Probe (`net.Listen`) auf 80 und 443 — **entfällt**, wenn der eigene reverse-proxy bereits läuft (`docker compose ps -q --status running reverse-proxy` nicht leer → Day-2-Pfad, kein Fehlalarm).
4. `docker compose -f <datei> up -d --build` mit live durchgereichter Ausgabe (stdout/stderr), danach `up -d --no-deps --force-recreate reverse-proxy` — wie `make local-up`. (`--build` ist bei der image-basierten Release-Datei ein No-op.)
5. **Health-Loop:** bis 120 s `GET https://localhost/api/health` (InsecureSkipVerify), Fortschrittsanzeige; nur HTTP 200 gilt. Timeout → Warnung + Exit-Code 1.
6. **Erfolgsausgabe:** Versionszeile (`jotti Starter vX.Y.Z`), Zugriffs-URL aus LAN-IP, Hinweis auf die einmalige Zertifikatswarnung pro Gerät, Firewall-Hinweis (Freigabe für private Netzwerke), Sicherheitswarnung (niemals ins Internet öffnen).
7. **Immer** — Erfolg wie Fehler — „Enter zum Schließen" vor dem Exit, damit das Doppelklick-Fenster lesbar bleibt. Exit-Code 0/1.

Make-Target `build-starter-windows` kompiliert `GOOS=windows GOARCH=amd64` mit `-ldflags "-X main.version=$(VERSION)"` nach `cmd/starter/jotti-start.exe` (`VERSION ?= dev`).

### Acceptance criteria

- [ ] `cmd/starter/main.go` implementiert den beschriebenen Ablauf; Exit-Code 0 bei Erfolg, 1 bei jedem Preflight-/Health-Fehler.
- [ ] Health-Check läuft gegen `/api/health` und akzeptiert nur HTTP 200.
- [ ] Ausgabe enthält Version, `https://{LAN-IP}`, Zertifikats-, Firewall- und Sicherheitshinweis.
- [ ] Fenster bleibt bei Erfolg und Fehler offen (Enter-Prompt).
- [ ] Läuft der eigene Stack bereits (Day 2 / Docker-Autostart), gibt es keinen Port-Fehlalarm; erneuter Start ist idempotent.
- [ ] Compose-/`.env`-Pfade funktionieren aus dem entpackten ZIP (neben der Exe) und im Repo (Dev).
- [ ] `make build-starter-windows` erzeugt `cmd/starter/jotti-start.exe`.

---

## Phase 3: Relay `.env`-Fallback für den Doppelklick

**User stories**: 15, 16, 17, 18

### Context

- `cmd/relay/main.go:66-104` — `loadConfigFromEnv(getenv)`; der Datei-Fallback bedient dieselbe Signatur.
- `cmd/relay/main.go:36, 93-96` — https-Default und TLS-Auto-Skip existieren bereits und bleiben unverändert (Ports fest); ebenso die deutsche Pflichtfeld-Meldung (Z. 69). **Nicht erneut bauen.**
- Das Relay läuft in keinem Compose-Stack (kein `relay`-Service in irgendeiner Compose-Datei); Env-Präzedenz bleibt trotzdem Pflicht für den Server-Betrieb (systemd o. Ä.).

### What to build

1. **`.env`-Parser** in `cmd/relay/` (eigene Datei, ~20 Zeilen): Key=Value, ignoriert Kommentare/Leerzeilen, trimmt Whitespace, CRLF und optionale Anführungszeichen, toleriert UTF-8-BOM.
2. **Fallback-Logik in `main()`:** Sind die `RELAY_*`-Variablen nicht gesetzt, wird die `.env` neben der Exe bzw. im Arbeitsverzeichnis geladen; echte Env-Variablen haben Vorrang.
3. **Fehler-UX:** Bei Konfigurationsfehlern deutsche Meldung + Enter-Pause statt sofort schließendem Fenster; Startup-Log um Versionszeile ergänzen (ldflags wie beim Starter).

Make-Target `build-relay-windows` erzeugt `cmd/relay/jotti-relay.exe` (`GOOS=windows GOARCH=amd64`, ldflags-Version).

### Acceptance criteria

- [ ] Parser unit-getestet: Key=Value, Kommentare, Anführungszeichen, Leerzeilen, CRLF, BOM.
- [ ] Doppelklick neben `.env`: Token und Backend-URL kommen aus der Datei, ohne manuelle Eingabe.
- [ ] Bereits gesetzte Env-Variablen haben Vorrang vor der Datei.
- [ ] Fehlt `RELAY_AUTH_TOKEN` in Datei und Env: deutsche Meldung, Fenster bleibt offen.
- [ ] `make build-relay-windows` erzeugt `cmd/relay/jotti-relay.exe`.

---

## Phase 4: Release-Pipeline (Versionierung + GHCR + ZIP + Workflow)

**User stories**: 1

### Context

- `docker-compose.local.yml:37-52, 54-57, 88-91` — die drei Build-Kontexte, die im ZIP durch Images ersetzt werden; `:16` — `name: jotti-local` muss erhalten bleiben.
- `database/migrate/Dockerfile` — migrate-Image; der Mount von `./database/migrations` bleibt auch im Release-Compose.
- `scripts/init-env.sh` — erzeugt die Smoke-Test-`.env` auf dem Runner (`make init`).
- `.github/workflows/ci.yml` — Versionsreferenz (checkout@v6, setup-go@v6, go 1.26.0).
- `Makefile:88-97` — bestehende Build-Targets.

### What to build

1. **`docker-compose.release.yml`:** inhaltlich wie die lokale Datei, aber `image: ghcr.io/nicograef/jotti-{backend,frontend,migrate}:<TAG>` statt `build:`-Sektionen; `name: jotti-local` bleibt (gleiches Compose-Projekt → gleiches Volume → ZIP-Upgrade ohne Datenverlust). Der Tag steht als Platzhalter im Repo und wird beim Release gepinnt.
2. **Make-Target `release-windows`:** nimmt `VERSION` entgegen, baut beide Exes (mit ldflags-Version), staged den ZIP-Inhalt und erzeugt `dist/jotti-windows-{VERSION}.zip`. Baut **keine** Images — die entstehen nur im Workflow.
3. **Workflow `.github/workflows/release.yml`:**
   - Trigger: Tag-Push `v*` (Publish) und `workflow_dispatch` (Dry-Run, Version `dev-<shortsha>`, publiziert nichts).
   - Ablauf: Images mit buildx bauen und `--load`en (Tag + OCI-Label `org.opencontainers.image.version`) → `make init` + Stack aus `docker-compose.release.yml` auf dem Runner hochfahren → **Smoke-Test** `curl -k https://localhost/api/health` bis 200 → erst danach: GHCR-Push (Login via `GITHUB_TOKEN`, `packages: write`), Tag in der Release-Compose-Datei pinnen, `make release-windows VERSION=<tag>`, GitHub-Release mit ZIP-Asset und `generate_release_notes: true`. Im Dry-Run enden die Schritte nach dem Smoke-Test.

**ZIP-Inhalt:** `jotti-start.exe`, `jotti-relay.exe`, `jotti-stop.cmd` (`docker compose -f docker-compose.release.yml down` + `pause`), `docker-compose.release.yml` (Tag gepinnt), `reverse-proxy/nginx.local.conf`, `reverse-proxy/local-entrypoint.sh`, `database/migrations/`, `.env.example`, `KURZANLEITUNG.md`.

### Acceptance criteria

- [ ] `make release-windows VERSION=v0.0.0-test` erzeugt lokal ein ZIP unter `dist/` mit genau dem beschriebenen Inhalt (ohne Image-Builds).
- [ ] Tag-Push `v0.X.Y` publiziert: drei Images public auf GHCR (exakter Tag, kein `:latest`, OCI-Version-Label) + GitHub-Release mit ZIP und auto-generierten Notes.
- [ ] Image-Push und Release passieren erst **nach** bestandenem Smoke-Test (Stack aus Release-Compose wird auf dem Runner healthy).
- [ ] `workflow_dispatch` (Dry-Run) baut und smoke-testet, publiziert aber nichts.
- [ ] Release-Compose im ZIP: gepinnter Tag, `name: jotti-local`, keine `build:`-Sektionen.
- [ ] Beide Exes im ZIP sind Windows-amd64-Binaries und geben beim Start `vX.Y.Z` aus.

---

## Phase 5: Dokumentation

**User stories**: 19, 20, 21, 22

### Context

- `docs/betrieb/leitfaden-hosting.md:42-128` — Weg A beschreibt derzeit den manuellen CLI-Ablauf.

### What to build

1. **`KURZANLEITUNG.md`** (im Release-ZIP, ≤ 1 Seite): Docker Desktop installieren; **Erststart mit Internet vorab zuhause** durchführen (Image-Download), nicht erst auf dem Fest; ZIP entpacken; SmartScreen-Dialog („Weitere Informationen → Trotzdem ausführen"); `jotti-start.exe` doppelklicken und warten; angezeigte URL an die Helfer-Handys geben (Zertifikatswarnung einmal pro Gerät bestätigen); `jotti-relay.exe` für den Bondruck; bei „Port belegt" das störende Programm beenden (typische Verursacher nennen); Beenden per `jotti-stop.cmd` oder Docker Desktop — Daten bleiben erhalten; am nächsten Tag dieselben zwei Doppelklicks (URL kann sich ändern, der Starter zeigt sie erneut an); **niemals** Port-Weiterleitung ins Internet.
2. **Update `docs/betrieb/leitfaden-hosting.md`:** Weg A erhält vor dem manuellen Weg einen Abschnitt „Starter verwenden (empfohlen)" mit Verweis auf Release-ZIP und Kurzanleitung; der manuelle Weg bleibt als Alternative. Optionaler Tipp: DHCP-Reservierung im Router für eine stabile Adresse.

### Acceptance criteria

- [ ] `KURZANLEITUNG.md` existiert und beschreibt den vollständigen Ablauf inkl. SmartScreen-, Zertifikats- und Beenden-Schritten in ≤ 1 Seite.
- [ ] Leitfaden Weg A verweist auf den Starter als empfohlenen Weg; der manuelle Weg bleibt dokumentiert.
- [ ] Sicherheitshinweis (nie ins Internet öffnen, Zertifikatswarnung) in beiden Dokumenten.
- [ ] Hinweis auf Erststart-mit-Internet und Datenpersistenz über Festtage in der Kurzanleitung.
