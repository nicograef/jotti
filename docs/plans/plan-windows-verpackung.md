# Plan: Klickbare Windows-Verpackung für den lokalen Betrieb

> Source PRD: `docs/prds/prd-windows-verpackung.md`
>
> Revision 2026-06-12 (2. Iteration): Ports fest auf 80/443 (KISS — Phase
> „Compose-Port-Parameterisierung" und Relay-Port-Kopplung gestrichen, Starter
> braucht keinen `.env`-Parser mehr); Release-/Versionierungskonzept ergänzt
> (produktweite SemVer-Tags, trunk-based, Tag-Push als einziges Publish-Gate).
>
> Revision 2026-06-12 (3. Iteration): **Admin-Rechte sind Voraussetzung** —
> `jotti-start.exe` läuft per `requireAdministrator`-Manifest elevated (eine
> UAC-Bestätigung pro Start, Nutzer-Entscheid). Damit werden drei
> „erklären"-Stellen zu „beheben"-Stellen: Die Firewall-Regel wird automatisch
> gesetzt (statt Hinweistext), Port-Konflikte nennen den exakten
> Verursacher-Prozess (statt typische Verursacher zu raten), Docker Desktop
> wird bei Bedarf selbst gestartet und der Container-Modus selbst umgeschaltet
> (statt Anleitung). **Kein Autostart** (Nutzer-Entscheid: Feste sind oft
> eintägig, danach wochen- bis monatelange Pause — der tägliche manuelle
> Doppelklick ist gewollt). Phase C (nativ ohne Docker) ist als eigene PRD
> ausgelagert: `docs/prds/prd-windows-nativ-ohne-docker.md`.
>
> PRD am 2026-06-12 entsprechend aktualisiert — Plan und PRD sind deckungsgleich.
>
> Revision 2026-06-13 (Anpassung an Option-3-TLS): Der lokale Stack läuft jetzt
> auf Caddy statt nginx (`plan-lokale-tls-vertrauenswuerdig.md`, als vollständig
> umgesetzt angenommen, bevor dieser Plan angegangen wird). Folgen: (1) der
> reverse-proxy ist ein **vierter** Build-Kontext / viertes GHCR-Image
> (Custom-xcaddy, kein offizielles nginx-Image mehr); (2) der Starter übergibt die
> host-seitig ermittelte LAN-IP als `LAN_IP`-Env an Compose (Caddy-Container —
> sonst rendert Caddy nur die Fallback-Site); (3) `nginx.local.conf` und
> `local-entrypoint.sh` entfallen im ZIP (Caddy rendert sein Caddyfile zur
> Laufzeit im Container); (4) die Erfolgsausgabe verweist auf die Status-Seite
> `http://localhost:8484` statt selbst eine `https://{ip}`-URL zu bauen — grünes
> Schloss ist der Normalfall, die Zertifikatswarnung betrifft nur noch den
> Fallback. Phase 6 des TLS-Plans ist entsprechend auf „bereits integriert"
> gesetzt, damit sie diese Edits nicht überschreibt.

## Goal

Ein Vereins-Admin kann jotti auf einem Windows-Rechner per Doppelklick starten (`jotti-start.exe`, läuft mit Administratorrechten — eine UAC-Bestätigung pro Start) und den Bondruck separat per Doppelklick aktivieren (`jotti-relay.exe`, unprivilegiert) — ohne `.env` von Hand anlegen, Geheimnisse erzeugen, Firewall konfigurieren, Docker Desktop vorab starten oder Kommandozeile benutzen zu müssen. Verteilt als GitHub-Release-ZIP mit vorgebauten Windows-Binaries und vorgebauten Container-Images, sodass das ZIP **ohne Quellcode** lauffähig ist. Releases entstehen ausschließlich durch produktweite Version-Tags (`v0.X.Y`) auf `main`.

## Architectural decisions

### Verpackung & Laufzeit

- **Quellort Starter:** `cmd/starter/` als eigenständiges Go-Modul (eigenes `go.mod`, module `jotti-starter`, go 1.26.0, keine Dependencies — analog `cmd/relay/`).
- **Admin-Rechte als Voraussetzung (Nutzer-Entscheid):** `jotti-start.exe` trägt ein eingebettetes `requireAdministrator`-Manifest. Doppelklick löst die UAC-Abfrage aus („Unbekannter Herausgeber" mangels Code-Signing — die Kurzanleitung erklärt den Dialog). Technisch: Manifest-XML in `cmd/starter/`, daraus einmalig per `rsrc` ein eingechecktes `rsrc_windows_amd64.syso` erzeugt — keine Go-Dependency, der Go-Linker bindet `.syso` automatisch ein, der Plattform-Suffix im Dateinamen beschränkt es auf windows/amd64-Builds (Linux-Builds in CI bleiben unberührt). Das Relay bleibt unprivilegiert (nur ausgehende Verbindungen, kein Manifest).
- **Arbeitsverzeichnis nach UAC-Elevation ist `C:\Windows\System32`:** Compose-Datei- und `.env`-Auflösung laufen deshalb **immer relativ zur Exe** (`os.Executable()`); das Arbeitsverzeichnis ist nur der Fallback für den Repo-Dev-Lauf (`go run`).
- **Windows-Seiteneffekte hinter `runtime.GOOS == "windows"`:** Firewall-Regel, Docker-Desktop-Start, Engine-Switch und Port-Verursacher-Lookup laufen nur unter Windows; der Repo-Dev-Lauf unter Linux überspringt sie.
- **Distribution mit vorgebauten Images:** `docker-compose.local.yml` hat seit dem Caddy-Umbau **vier** Build-Kontexte (`migrate`, `backend`, `frontend`, `reverse-proxy` — letzterer ist jetzt ein Custom-xcaddy-Image, kein offizielles nginx-Image mit gemounteter Config mehr) — ein ZIP ohne Quellbaum kann nicht `--build`en. Der Release-Workflow pusht deshalb `ghcr.io/nicograef/jotti-{backend,frontend,migrate,reverse-proxy}` (Repo ist public, Pull ohne Login) und das ZIP enthält ein `docker-compose.release.yml` mit beim Release fest gepinntem Image-Tag. `database/migrations/` liegt im ZIP und wird wie bisher in den migrate-Container gemountet (kein Dockerfile-Umbau).
- **Ports fest 80/443 (KISS):** keine Port-Variablen, keine Compose-Änderung, keine `.env`-Port-Keys. Sollte Port-Konfigurierbarkeit doch nötig werden, bleibt sie der dokumentierte Eskalationspfad.
- **Port-Diagnose mit exaktem Verursacher:** Ist 80/443 belegt (und nicht durch den eigenen reverse-proxy), ermittelt der Starter den haltenden Prozess über `powershell -NoProfile -Command "Get-NetTCPConnection … | ConvertTo-Json"` inklusive Prozessname — Admin-Rechte machen auch System- und Fremdprozess-Namen lesbar (z. B. PID 4 = `System` → http.sys/IIS). Die Diagnose nennt Prozessname + PID und fordert auf, das Programm zu beenden. **Kein automatisches Beenden** fremder Prozesse oder Dienste. Kein netstat-Text-Parsing — die JSON-Ausgabe parst eine unit-getestete Core-Funktion; schlägt der Lookup fehl, fällt die Diagnose auf die generische Meldung mit typischen Verursachern zurück.
- **Docker-Reparatur statt Anleitung:** Antwortet der Daemon nicht und liegt `Docker Desktop.exe` am Standardpfad (`C:\Program Files\Docker\Docker\`), startet der Starter Docker Desktop selbst und pollt `docker info` mit Fortschrittsanzeige (bis 120 s — WSL2-/VM-Kaltstart auf Altgeräten). Meldet `docker info -f '{{.OSType}}'` den Windows-Container-Modus, schaltet der Starter per `DockerCli.exe -SwitchLinuxEngine` selbst um und prüft erneut. Nur „Docker Desktop ist nicht installiert" bleibt ein manueller Schritt (Kurzanleitung; automatische Installation via winget ist der dokumentierte Eskalationspfad).
- **Firewall-Regel statt Hinweis:** Der Starter setzt idempotent eine eingehende Regel — `netsh advfirewall firewall show rule name="jotti"`, falls fehlend: `add rule name="jotti" dir=in action=allow protocol=TCP localport=80,443 remoteip=localsubnet profile=any`. `remoteip=localsubnet` + `profile=any` macht die Freigabe unabhängig von der Netzwerkprofil-Einstufung — der häufigste reale Showstopper ist ein als „Öffentlich" eingestuftes Vereins-WLAN, das eingehende Verbindungen pauschal blockt — und beschränkt sie zugleich aufs lokale Subnetz (passt zum „niemals ins Internet"-Modell). Die Erfolgsausgabe **bestätigt** die Freigabe statt sie anzumahnen; schlägt `netsh` wider Erwarten fehl: Warnung mit manuellem Hinweis, kein Abbruch.
- **Health-Check:** `GET https://localhost/api/health` (TLS-Verify aus — `localhost` trifft Caddys Fallback-Catch-all mit interner CA). Caddy proxyt nur `/api/*` und `/` — ein Check auf `/health` würde von der Frontend-SPA mit 200 beantwortet (False-Positive). Nur HTTP 200 gilt als bereit; das Backend liefert 503 „degraded", solange die DB nicht antwortet.
- **LAN-IP-Übergabe an Compose (Folge des Caddy-Umbaus):** Der lokale Stack ermittelt die LAN-IP host-seitig und reicht sie via `LAN_IP`-Env an den Caddy-Container (`make local-up` macht das mit `ip route get`; ein Bridge-Container sähe nur die Docker-Bridge-IP). Ohne `LAN_IP` rendert Caddy nur die Fallback-Site. Der Starter spiegelt das: `SelectLANIP` (Core) liefert die IP, die beim `docker compose up` als `LAN_IP`-Env gesetzt wird. `ACMEDNS_BASE_URL`/`PROXY_LE_STAGING` bleiben auf ihren Compose-Defaults — der Starter setzt sie nicht.
- **Anzeige über die Status-Seite, nicht die Konsole:** Die primäre Adresse ist der grüne Hostname `<lan-ip>.<install-id>.lokal.jotti.rocks` (echtes LE-Zertifikat). Den kann der Starter nicht selbst bauen — er kennt die Install-ID nicht. Option 3 serviert deshalb eine Status-Seite unter `http://localhost:8484` (nur `127.0.0.1`) mit grüner Adresse, QR-Code und Fallback. Die Erfolgsausgabe des Starters verweist nur dorthin; eine eigene `https://{ip}`-URL baut der Starter nicht mehr. Die Zertifikatswarnung gilt nur noch für den Fallback `https://<LAN-IP>`, nicht den Normalfall.
- **Secrets als Hex:** 32 Bytes aus `crypto/rand`, hex-kodiert (64 Zeichen) — identisch zu `openssl rand -hex 32` in `scripts/init-env.sh`. Base64-Zeichen (`+/=`) würden die Postgres-URL im migrate-CMD (`postgres://user:pass@…`) brechen.
- **Konsolen-UX:** Ein per Doppelklick gestartetes Konsolenprogramm schließt sein Fenster beim Exit sofort. Der Starter streamt deshalb die Compose-Ausgabe live (Pull sichtbar statt eingefrorenem Fenster) und wartet am Ende — bei Erfolg **und** Fehler — auf Enter, sonst verschwindet die angezeigte URL sofort wieder. Das Relay pausiert bei Konfigurationsfehlern vor dem Exit.
- **`.env`-Parsing nur im Relay:** Der Starter schreibt die `.env` nur (liest sie nie — Ports sind fest, Compose liest sie selbst). Das Relay bekommt einen minimalen Key=Value-Parser (~20 Zeilen, kein Shared-Package), der CRLF (Notepad-Edit), UTF-8-BOM und optionale Anführungszeichen toleriert; bereits gesetzte Env-Variablen haben Vorrang vor der Datei.
- **Kein Code-Signing** (out of scope); SmartScreen-/Defender-Hinweis und der UAC-Dialog („Unbekannter Herausgeber") gehören in die Kurzanleitung.

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
- `docker-compose.local.yml:18` — `name: jotti-local` (stabiler Projektname → Volumes überleben Ordner-Umbenennung und ZIP-Upgrades); `:39-42, 56-59, 90-93, 103-106` — die vier Build-Kontexte (inkl. reverse-proxy/Caddy); `:117-120` — Ports `80:80`/`443:443` (+`443:443/udp` für HTTP/3) bleiben hartkodiert; `:139-143` — Volumes `postgres-data`, `caddy-data`, `proxy-state` (das frühere `tls-certs` entfällt).
- `reverse-proxy/caddyfile.go` — das zur Laufzeit gerenderte Caddyfile übernimmt den `/api/*`-Proxy mit Prefix-Strip → `/api/health` erreicht den Backend-Handler `/health` (das frühere `reverse-proxy/nginx.local.conf` existiert nicht mehr).
- `reverse-proxy/main.go` — der Go-Entrypoint des Caddy-Containers (ersetzt `local-entrypoint.sh`): liest `LAN_IP`, rendert das Caddyfile und startet Caddy. Das Rendering läuft nur beim Entrypoint-Lauf — deshalb force-recreate des Proxys beim Start (wie `make local-up`).
- `backend/api/health/health.go:28-58` — Health-Handler: GET, DB-Ping, 503 bei degraded.
- `Makefile:22-23` — `make init`; `:91-92` — `build-relay`; `:158-162` — `local-up` ermittelt die LAN-IP host-seitig (`ip route get`), übergibt sie als `LAN_IP`-Env und macht `--force-recreate reverse-proxy` — der Starter spiegelt genau diesen Ablauf.
- `.github/workflows/ci.yml:23-32` — paths-filter **ohne** `cmd/**` (für Relay/Starter läuft heute keine CI); die Jobs `resolver-ci`/`local-proxy-ci` + ihre paths-filter sind durch Option 3 bereits ergänzt — neu kommt nur der `cmd`-Job hinzu; checkout@v6, setup-go@v6, go 1.26.0 als Versionsreferenz für `release.yml`.
- `database/migrate/Dockerfile` — migrate-Image ohne eingebackene Migrationen; der Mount bleibt auch im Release-Compose.
- `docs/betrieb/leitfaden-hosting.md:42-128` — Weg A (manueller CLI-Ablauf), wird um Starter-Abschnitt ergänzt.

## Resolved decisions

- **Admin-Rechte immer vorausgesetzt (Nutzer-Entscheid 2026-06-12):** `requireAdministrator`-Manifest, UAC-Abfrage bei jedem Start ist akzeptiert. Keine Aufteilung in Setup-Programm und Tagesstart — alle privilegierten Schritte (Firewall, Docker-Start, Engine-Switch, Port-Lookup) laufen idempotent bei jedem Start. Das Relay bleibt unprivilegiert.
- **Kein Autostart (Nutzer-Entscheid 2026-06-12):** keine geplante Aufgabe, kein Windows-Dienst, kein Anmelde-Start. Vereinsfeste laufen oft nur einen Tag, danach wird jotti wochen- bis monatelang nicht gebraucht — jotti soll nicht dauerhaft mitlaufen. (Unabhängig davon bleibt der Start idempotent: startet Docker Desktop von sich aus und läuft der Stack via `restart: unless-stopped` schon, ist das der Erfolgs-Pfad, kein Fehler.)
- **Firewall automatisch statt Hinweistext (Folge des Admin-Entscheids):** Regel `jotti`, eingehend, TCP 80/443, `remoteip=localsubnet`, `profile=any` — profilunabhängig (löst den „öffentliches Netzwerk"-Showstopper) und zugleich aufs lokale Subnetz beschränkt. Der bisherige Firewall-Hinweis in der Erfolgsausgabe wird zur Bestätigung.
- **Exakte Port-Diagnose statt Raten (Folge des Admin-Entscheids):** Verursacher-Prozess wird per `Get-NetTCPConnection` (JSON) benannt; das frühere „Dead-End ohne Programmbeendigung"-Restrisiko entfällt weitgehend. Automatisches Beenden fremder Prozesse/Dienste bleibt bewusst ausgeschlossen.
- **Docker-Selbststart & Engine-Switch (Folge des Admin-Entscheids):** „Docker Desktop starten und warten, bis der Wal grün ist" und „auf Linux-Container umschalten" sind keine Nutzer-Anweisungen mehr, sondern Auto-Fixes. Manuell bleibt nur die Docker-Desktop-Installation.
- **Keine konfigurierbaren Ports** (Nutzer-Entscheid, KISS): Die frühere Phase „Compose-Port-Parameterisierung" und die Relay-Port-Kopplung entfallen ersatzlos. Dadurch: kein `.env`-Parser im Starter, `.env.example` unverändert, Relay-Default-URL unverändert.
- **GHCR statt Quellcode-im-ZIP:** Ein `docker build` auf dem Vereins-PC wäre der fragilste Schritt des ganzen Ablaufs (Go-/npm-Downloads während des Builds, zweistellige Minuten auf Altgeräten). Ein Image-Pull ist ein einzelner robuster Download.
- **SemVer `v0.X.Y` produktweit, Tag-Push + Dry-Run, nur Smoke-Test als Gate, keine `:latest`-Tags, Version in Exes (ldflags) und Image-Labels, Release-Notes auto-generiert, `ci.yml` + `release.yml`** — Klärungsrunden vom 2026-06-12.
- **Compose-Datei-Auflösung im Starter:** immer relativ zur Exe (`os.Executable()`): `docker-compose.release.yml` neben der Exe (ZIP-Fall), sonst `docker-compose.local.yml` neben der Exe; das Arbeitsverzeichnis ist nur Repo-Dev-Fallback — nach UAC-Elevation wäre es `System32`.
- **Day-2:** Die Container haben `restart: unless-stopped` — am zweiten Festtag läuft der Stack u. U. schon, sobald Docker Desktop startet. Hält der **eigene** reverse-proxy den Port, ist das kein Preflight-Fehler, sondern der Erfolgs-Pfad (Start ist idempotent).
- **Starter spiegelt `make local-up`** inklusive host-seitiger `LAN_IP`-Übergabe und `--force-recreate reverse-proxy`, damit das Caddyfile-Rendering bei IP-Wechsel tatsächlich läuft.
- **LAN-IP primär über die Outbound-Route** (UDP-„Connect" zu einer externen Adresse, es wird kein Paket gesendet; `LocalAddr` liefert die IP des Default-Route-Interfaces). „Erste private IPv4" wäre auf Windows-Rechnern mit Docker Desktop falsch — vEthernet-/WSL-Adapter tragen private 172.x-Adressen, die Smartphones nicht erreichen. Interface-Heuristik nur als Fallback (192.168.x vor 10.x vor 172.16-31.x). Das Ergebnis wird seit dem Caddy-Umbau als `LAN_IP`-Env an den Proxy-Container gereicht (deckt sich mit der host-seitigen Erkennung in `plan-lokale-tls-vertrauenswuerdig.md` Phase 4).
- **Tests in `cmd/starter` ohne Build-Tag** — wie `cmd/relay`. Bewusste Abweichung von der PRD-Notation `-tags=unit`: In den cmd-Modulen gibt es keine Integrationstests, der Tag hätte keinen Trennzweck. Neuer CI-Job für `cmd/**` (deckt damit auch das bisher CI-lose Relay ab).
- **Health-Timeout 120 s** nach Rückkehr von `up -d` (Erststart: Postgres-Init + Migrationen auf Altgeräten; der Image-Pull blockiert davor sichtbar im `up`-Aufruf selbst).
- **`jotti-stop.cmd`** (eine Zeile `docker compose … down` + `pause`) liegt im ZIP — Story 19 „sauber beenden" per Doppelklick; Docker-Desktop-GUI bleibt als Alternative dokumentiert.
- **Caddy/Option-3-Integration (2026-06-13):** Vierter Build-Kontext → viertes GHCR-Image `jotti-reverse-proxy`; ZIP ohne `nginx.local.conf`/`local-entrypoint.sh` (Caddy rendert sein Caddyfile zur Laufzeit); `LAN_IP`-Env an Compose; Erfolgsausgabe verweist auf die Status-Seite `http://localhost:8484`; grünes Schloss ist Normalfall, Zertifikatswarnung nur noch Fallback. Setzt voraus, dass `plan-lokale-tls-vertrauenswuerdig.md` vollständig umgesetzt ist (inkl. Status-Seite aus dessen Phase 5).

## Open questions / Risks

- **Docker-CLI-Pfad auf Windows:** `docker` muss im PATH liegen (Docker Desktop installiert nach `C:\Program Files\Docker\Docker\resources\bin`). Die Preflight-Diagnose nennt den Pfad.
- **SmartScreen/Defender + UAC:** Unsignierte `.exe`s lösen beim ersten Start SmartScreen aus („Weitere Informationen → Trotzdem ausführen"); der UAC-Dialog zeigt bei jedem Start „Unbekannter Herausgeber". Code-Signing bleibt out of scope; die Kurzanleitung beschreibt beide Dialoge.
- **UAC-Voraussetzung:** Der angemeldete Nutzer muss Administrator sein oder Admin-Zugangsdaten kennen — per Nutzer-Entscheid vorausgesetzt; die Kurzanleitung nennt es als Anforderung.
- **Port 80/443 belegt:** kein Rate-Dead-End mehr — der Verursacher wird exakt benannt. Restaufwand bleibt beim Admin: Programm selbst beenden, jotti erneut starten. Schlägt der PowerShell-Lookup fehl (sollte auf Windows 10/11 mit PowerShell 5.1 nicht vorkommen), greift der generische Fallback mit typischen Verursachern.
- **Compose-Drift local vs. release:** Zwei Compose-Dateien können auseinanderlaufen. Der Smoke-Test im Release-Workflow fängt funktionale Drift ab.
- **DHCP-IP-Wechsel:** Am nächsten Festtag kann der Rechner eine neue LAN-IP haben → neuer Hostname, aber **dasselbe** Wildcard-Zertifikat (`*.<install-id>.lokal.jotti.rocks`), also keine neue Warnung im Normalfall. Die Status-Seite zeigt die aktuelle grüne Adresse + QR-Code neu an; Kurzanleitung erwähnt es (optionaler Tipp im Leitfaden: DHCP-Reservierung im Router).
- **Erststart braucht Internet** — für den Image-Pull und für die erste Zertifikatsausstellung (acme-dns-Registrierung + Let's Encrypt über die jotti.rocks-Infra). Bis das grüne Zertifikat da ist (und generell offline) trägt der Fallback `https://<LAN-IP>`. Folgetage laufen aus dem lokalen Image-Cache. Kurzanleitung: Erststart vorab zuhause mit Internet durchführen, nicht erst auf dem Fest.
- **Firewall deckt nur TCP:** Caddy bindet zusätzlich `443/udp` (HTTP/3). Die Firewall-Regel `jotti` öffnet nur TCP 80/443; HTTP/3-Versuche aus dem WLAN werden geblockt und der Browser fällt transparent auf HTTP/2 über TCP zurück (funktional unkritisch). Bewusste Entscheidung: TCP-only lassen; eine UDP-443-Freigabe ist der dokumentierte Eskalationspfad, falls HTTP/3 erwünscht ist.

---

## Phase 1: Starter-Core (reine Logik + Unit-Tests + CI)

**User stories**: 2, 3, 4, 6, 8, 11, 12, 14

### Context

- `scripts/init-env.sh:20-39` — Idempotenz-Verhalten und Hex-Secret-Erzeugung als Referenz; die Go-Implementierung nutzt `crypto/rand` direkt.
- `.env.example:1-11` — die vier zu erzeugenden Keys.
- `cmd/relay/go.mod` + `cmd/relay/main_test.go` — Muster für Modulzuschnitt und ungetaggte Tests.
- `.github/workflows/ci.yml:23-32` — paths-filter, dem `cmd/**` fehlt.

### What to build

Ein Go-Package `cmd/starter/core` mit seiteneffektfreien, unit-testbaren Funktionen:

1. **Secret-Erzeugung** — `GenerateSecret() string`: 32 Bytes aus `crypto/rand`, hex-kodiert (64 Zeichen, kompatibel zur Postgres-URL im migrate-CMD).
2. **`.env`-Materialisierung** — erzeugt die `.env` nur, wenn sie fehlt (Dateizugriffe injiziert, testbar); Inhalt: deutscher Kommentar-Header („nichts hier muss geändert werden"; schützt nebenbei vor einem späteren Notepad-BOM auf der ersten Key-Zeile), `POSTGRES_USER=admin`, `POSTGRES_PASSWORD`/`JWT_SECRET`/`RELAY_AUTH_TOKEN` generiert. Vorhandene Datei wird nie überschrieben.
3. **Preflight-Auswertung** — bildet Prüfergebnisse auf deutsche Diagnosen mit Handlungshinweis ab: Docker-CLI fehlt (nennt den Docker-Desktop-Pfad); Docker Desktop nicht installiert (Exe am Standardpfad fehlt); Docker-Desktop-Start fehlgeschlagen bzw. Daemon nach Timeout nicht erreichbar; Engine-Umschaltung fehlgeschlagen; Port 80/443 belegt durch `<Prozessname> (PID …)` mit der Aufforderung, das Programm zu beenden und jotti erneut zu starten; Port belegt ohne ermittelbaren Verursacher (generischer Fallback: typische Verursacher wie VMware Workstation, IIS, Skype).
4. **Port-Verursacher-Parsing** — `ParsePortOwners(jsonOut []byte) ([]PortOwner, error)` für die `Get-NetTCPConnection`-JSON-Ausgabe: PowerShell liefert bei einem Treffer ein einzelnes Objekt, bei mehreren ein Array; fehlende Felder (Prozessname nicht auflösbar) werden toleriert. Ergebnis: Zuordnung Port → Prozessname/PID für die Diagnose.
5. **LAN-IP-Auswahl** — `SelectLANIP(outboundIP string, interfaces []NetInterface) (string, error)`: bevorzugt die Outbound-Route-IP, wenn sie privat (RFC 1918) ist; Fallback-Heuristik über die Interfaces mit Präferenz 192.168.x > 10.x > 172.16-31.x; ignoriert Loopback und Link-Local (169.254.x). Das Ergebnis wird in Phase 2 als `LAN_IP`-Env an den Caddy-Container gereicht (nicht mehr zum Bau einer Anzeige-URL — die grüne Adresse zeigt die Status-Seite `http://localhost:8484`, der Starter kennt die Install-ID nicht).

Dazu CI: paths-filter in `ci.yml` um `cmd: 'cmd/**'` erweitern und einen Job `cmd-ci` ergänzen (setup-go 1.26.0; `go vet`, `go build ./...`, `go test ./...` jeweils in `cmd/relay` und `cmd/starter`).

### Acceptance criteria

- [x] `cmd/starter/` existiert mit eigenem `go.mod` (module `jotti-starter`, go 1.26.0, keine Dependencies).
- [x] Secrets: 64 Hex-Zeichen; zwei Aufrufe liefern unterschiedliche Werte.
- [x] `.env`-Idempotenz: vorhandene Datei wird nie überschrieben; fehlt sie, enthält sie die vier Keys aus `.env.example` und den Kommentar-Header.
- [x] LAN-IP: Outbound-IP wird bevorzugt; eine 172.x-Adresse (vEthernet/WSL-Muster) gewinnt nie gegen eine 192.168.x; Loopback/169.254.x werden ignoriert.
- [x] Port-Verursacher-Parser: einzelnes JSON-Objekt, JSON-Array und fehlender Prozessname werden korrekt auf Port→Name/PID abgebildet; kaputtes JSON liefert einen Fehler (→ generischer Fallback).
- [x] Preflight: jede Bedingung bildet auf eine deutsche Diagnose mit Handlungshinweis ab — inklusive „Port 80 ist durch ‚X' (PID n) belegt".
- [x] `go build ./...` und `go test ./...` in `cmd/starter/` erfolgreich.
- [x] `ci.yml` führt den neuen `cmd-ci`-Job bei Änderungen unter `cmd/**` aus (greift damit auch für `cmd/relay`).

---

## Phase 2: Starter-Shell + Binary

**User stories**: 7, 8, 9, 11, 12, 13, 14

### Context

- `cmd/starter/core` (Phase 1) — liefert die reine Logik.
- `Makefile:158-162` — `local-up` inkl. host-seitiger `LAN_IP`-Übergabe und `--force-recreate reverse-proxy` als zu spiegelndes Verhalten.
- `reverse-proxy/main.go` — warum das force-recreate nötig ist (Caddyfile-Rendering aus `LAN_IP` läuft nur im Entrypoint).
- `backend/api/health/health.go:28-58` + `reverse-proxy/caddyfile.go` — Health über `/api/health` (Prefix-Strip durch das gerenderte Caddyfile), nur 200 = bereit.

### What to build

Die `main.go` in `cmd/starter/` verbindet Core-Logik mit echten Seiteneffekten. Alle Windows-spezifischen Schritte (Docker-Desktop-Start, Engine-Switch, Port-Verursacher-Lookup, Firewall) laufen nur bei `runtime.GOOS == "windows"` — der Repo-Dev-Lauf unter Linux überspringt sie.

1. **Compose-Datei finden — immer relativ zur Exe** (`os.Executable()`; nach UAC-Elevation ist das Arbeitsverzeichnis `System32`): `docker-compose.release.yml` neben der Exe (ZIP), sonst `docker-compose.local.yml` neben der Exe; Arbeitsverzeichnis nur als Repo-Dev-Fallback (`go run`). Nichts gefunden → deutsche Diagnose.
2. `.env` neben der Compose-Datei materialisieren (Core).
3. **Preflights mit Auto-Fix:**
   - `docker` via `exec.LookPath`; fehlt es → Diagnose (Docker-Desktop-Pfad).
   - `docker info -f '{{.OSType}}'`: Daemon antwortet nicht → liegt `Docker Desktop.exe` am Standardpfad, selbst starten und `docker info` bis 120 s pollen (Fortschrittsanzeige); fehlt die Exe oder läuft der Timeout ab → Diagnose.
   - `OSType=windows` → `DockerCli.exe -SwitchLinuxEngine`, danach Re-Check; Fehlschlag → Diagnose.
   - Port-Probe (`net.Listen`) auf 80 und 443 — **entfällt**, wenn der eigene reverse-proxy bereits läuft (`docker compose ps -q --status running reverse-proxy` nicht leer → Day-2-Pfad, kein Fehlalarm). Belegt → Verursacher-Lookup per `powershell -NoProfile -Command "Get-NetTCPConnection -State Listen -LocalPort <port> -ErrorAction SilentlyContinue | Select-Object LocalPort,OwningProcess,@{n='ProcessName';e={(Get-Process -Id $_.OwningProcess).ProcessName}} | ConvertTo-Json"` → Core-Parser → exakte Diagnose (generischer Fallback bei Lookup-/Parse-Fehler).
4. **Firewall-Regel idempotent setzen:** `netsh advfirewall firewall show rule name="jotti"`; fehlt sie: `add rule name="jotti" dir=in action=allow protocol=TCP localport=80,443 remoteip=localsubnet profile=any`. Fehlschlag → Warnung mit manuellem Hinweis, kein Abbruch.
5. `docker compose -f <datei> up -d --build` mit live durchgereichter Ausgabe (stdout/stderr) und gesetztem `LAN_IP`-Env (aus `SelectLANIP`), danach `up -d --no-deps --force-recreate reverse-proxy` — wie `make local-up` (`Makefile:158-162`). Ohne `LAN_IP` rendert Caddy nur die Fallback-Site. (`--build` ist bei der image-basierten Release-Datei ein No-op.)
6. **Health-Loop:** bis 120 s `GET https://localhost/api/health` (InsecureSkipVerify), Fortschrittsanzeige; nur HTTP 200 gilt. Timeout → Warnung + Exit-Code 1.
7. **Erfolgsausgabe:** Versionszeile (`jotti Starter vX.Y.Z`), Verweis auf die Status-Seite `http://localhost:8484` („dort stehen Zugangsadresse und QR-Code"), Bestätigung „Firewall-Freigabe für das lokale Netzwerk ist eingerichtet", Sicherheitswarnung (niemals ins Internet öffnen). Der Starter baut keine eigene Zugriffs-URL mehr — die grüne Adresse (und der Fallback mit einmaliger Warnung) kommen von der Status-Seite.
8. **Immer** — Erfolg wie Fehler — „Enter zum Schließen" vor dem Exit, damit das Doppelklick-Fenster lesbar bleibt. Exit-Code 0/1.

**Manifest:** `cmd/starter/` erhält die Manifest-XML (`requireAdministrator`) und das daraus einmalig per `rsrc` erzeugte, eingecheckte `rsrc_windows_amd64.syso` — keine Go-Dependency; der Go-Linker bindet es bei windows/amd64-Builds automatisch ein.

Make-Target `build-starter-windows` kompiliert `GOOS=windows GOARCH=amd64` mit `-ldflags "-X main.version=$(VERSION)"` nach `cmd/starter/jotti-start.exe` (`VERSION ?= dev`).

### Acceptance criteria

- [x] `cmd/starter/main.go` implementiert den beschriebenen Ablauf; Exit-Code 0 bei Erfolg, 1 bei jedem Preflight-/Health-Fehler.
- [x] Die Exe enthält das `requireAdministrator`-Manifest — Doppelklick löst die UAC-Abfrage aus.
- [x] Firewall-Regel `jotti` wird beim ersten Lauf angelegt (eingehend, TCP 80/443, `localsubnet`, alle Profile) und bei weiteren Läufen nicht dupliziert.
- [x] Docker Desktop installiert, aber nicht gestartet → der Starter startet es selbst und fährt ohne Nutzeraktion fort; Windows-Container-Modus → automatische Umschaltung.
- [x] Port belegt durch Fremdprozess → Diagnose nennt Prozessname + PID; läuft der eigene Stack bereits (Day 2), gibt es keinen Fehlalarm; erneuter Start ist idempotent.
- [x] Health-Check läuft gegen `/api/health` und akzeptiert nur HTTP 200.
- [x] Ausgabe enthält Version, den Verweis auf die Status-Seite (`http://localhost:8484`), Firewall-Bestätigung und Sicherheitswarnung.
- [x] Fenster bleibt bei Erfolg und Fehler offen (Enter-Prompt).
- [x] Compose-/`.env`-Pfade werden über `os.Executable()` aufgelöst (funktioniert trotz `System32`-Arbeitsverzeichnis nach UAC) und im Repo-Dev-Fall übers Arbeitsverzeichnis.
- [x] Der Starter setzt beim `docker compose up` das `LAN_IP`-Env aus `SelectLANIP`; ohne gesetztes `LAN_IP` würde Caddy nur die Fallback-Site rendern.
- [x] `make build-starter-windows` erzeugt `cmd/starter/jotti-start.exe`.

---

## Phase 3: Relay `.env`-Fallback für den Doppelklick

**User stories**: 15, 16, 17, 18

### Context

- `cmd/relay/main.go:66-104` — `loadConfigFromEnv(getenv)`; der Datei-Fallback bedient dieselbe Signatur.
- `cmd/relay/main.go:36, 93-96` — https-Default und TLS-Auto-Skip existieren bereits und bleiben unverändert (Ports fest); ebenso die deutsche Pflichtfeld-Meldung (Z. 69). **Nicht erneut bauen.**
- Das Relay läuft in keinem Compose-Stack (kein `relay`-Service in irgendeiner Compose-Datei); Env-Präzedenz bleibt trotzdem Pflicht für den Server-Betrieb (systemd o. Ä.).
- Das Relay bleibt **unprivilegiert** (kein Manifest): Es baut nur ausgehende Verbindungen auf (Backend-Polling, Drucker im LAN) — die Windows-Firewall blockt eingehend, nicht ausgehend.
- Der Caddy-Umbau berührt das Relay nicht: Die Default-URL `https://localhost/api` trifft Caddys Fallback-Catch-all, der `/api/*` unverändert proxyt; TLS-Auto-Skip greift weiterhin.

### What to build

1. **`.env`-Parser** in `cmd/relay/` (eigene Datei, ~20 Zeilen): Key=Value, ignoriert Kommentare/Leerzeilen, trimmt Whitespace, CRLF und optionale Anführungszeichen, toleriert UTF-8-BOM.
2. **Fallback-Logik in `main()`:** Sind die `RELAY_*`-Variablen nicht gesetzt, wird die `.env` neben der Exe bzw. im Arbeitsverzeichnis geladen; echte Env-Variablen haben Vorrang.
3. **Fehler-UX:** Bei Konfigurationsfehlern deutsche Meldung + Enter-Pause statt sofort schließendem Fenster; Startup-Log um Versionszeile ergänzen (ldflags wie beim Starter).

Make-Target `build-relay-windows` erzeugt `cmd/relay/jotti-relay.exe` (`GOOS=windows GOARCH=amd64`, ldflags-Version).

### Acceptance criteria

- [x] Parser unit-getestet: Key=Value, Kommentare, Anführungszeichen, Leerzeilen, CRLF, BOM.
- [x] Doppelklick neben `.env`: Token und Backend-URL kommen aus der Datei, ohne manuelle Eingabe.
- [x] Bereits gesetzte Env-Variablen haben Vorrang vor der Datei.
- [x] Fehlt `RELAY_AUTH_TOKEN` in Datei und Env: deutsche Meldung, Fenster bleibt offen.
- [x] `make build-relay-windows` erzeugt `cmd/relay/jotti-relay.exe`.

---

## Phase 4: Release-Pipeline (Versionierung + GHCR + ZIP + Workflow)

**User stories**: 1

### Context

- `docker-compose.local.yml:39-42, 56-59, 90-93, 103-106` — die vier Build-Kontexte (inkl. reverse-proxy/Caddy), die im ZIP durch Images ersetzt werden; `:18` — `name: jotti-local` muss erhalten bleiben; `:139-143` — Volumes `caddy-data`/`proxy-state` (kein `tls-certs` mehr).
- `database/migrate/Dockerfile` — migrate-Image; der Mount von `./database/migrations` bleibt auch im Release-Compose.
- `scripts/init-env.sh` — erzeugt die Smoke-Test-`.env` auf dem Runner (`make init`).
- `.github/workflows/ci.yml` — Versionsreferenz (checkout@v6, setup-go@v6, go 1.26.0).
- `Makefile:88-97` — bestehende Build-Targets.
- Das `rsrc_windows_amd64.syso` (Phase 2) wird vom Go-Linker nur bei windows/amd64-Builds eingebunden — der Smoke-Test auf dem Linux-Runner bleibt unberührt.

### What to build

1. **`docker-compose.release.yml`:** inhaltlich wie die lokale Datei, aber `image: ghcr.io/nicograef/jotti-{backend,frontend,migrate,reverse-proxy}:<TAG>` statt der vier `build:`-Sektionen; `name: jotti-local` und die Volumes (`postgres-data`, `caddy-data`, `proxy-state`) bleiben (gleiches Compose-Projekt → gleiche Volumes → ZIP-Upgrade ohne Datenverlust, Zertifikate inklusive). Der Tag steht als Platzhalter im Repo und wird beim Release gepinnt.
2. **Make-Target `release-windows`:** nimmt `VERSION` entgegen, baut beide Exes (mit ldflags-Version), staged den ZIP-Inhalt und erzeugt `dist/jotti-windows-{VERSION}.zip`. Baut **keine** Images — die entstehen nur im Workflow.
3. **Workflow `.github/workflows/release.yml`:**
   - Trigger: Tag-Push `v*` (Publish) und `workflow_dispatch` (Dry-Run, Version `dev-<shortsha>`, publiziert nichts).
   - Ablauf: alle vier Images (backend, frontend, migrate, reverse-proxy) mit buildx bauen und `--load`en (Tag + OCI-Label `org.opencontainers.image.version`) → `make init` + Stack aus `docker-compose.release.yml` auf dem Runner hochfahren → **Smoke-Test** `curl -k https://localhost/api/health` bis 200 → erst danach: GHCR-Push (Login via `GITHUB_TOKEN`, `packages: write`), Tag in der Release-Compose-Datei pinnen, `make release-windows VERSION=<tag>`, GitHub-Release mit ZIP-Asset und `generate_release_notes: true`. Im Dry-Run enden die Schritte nach dem Smoke-Test.

**ZIP-Inhalt:** `jotti-start.exe`, `jotti-relay.exe`, `jotti-stop.cmd` (`docker compose -f docker-compose.release.yml down` + `pause`), `docker-compose.release.yml` (Tag gepinnt), `database/migrations/`, `.env.example`, `KURZANLEITUNG.md`. **Keine** `reverse-proxy/nginx.local.conf` / `local-entrypoint.sh` mehr — Caddy rendert sein Caddyfile zur Laufzeit im Container, es gibt keine zu mountende Proxy-Config.

### Acceptance criteria

- [x] `make release-windows VERSION=v0.0.0-test` erzeugt lokal ein ZIP unter `dist/` mit genau dem beschriebenen Inhalt (ohne Image-Builds).
- [x] Tag-Push `v0.X.Y` publiziert: vier Images public auf GHCR (backend, frontend, migrate, reverse-proxy; exakter Tag, kein `:latest`, OCI-Version-Label) + GitHub-Release mit ZIP und auto-generierten Notes.
- [x] Image-Push und Release passieren erst **nach** bestandenem Smoke-Test (Stack aus Release-Compose wird auf dem Runner healthy).
- [x] `workflow_dispatch` (Dry-Run) baut und smoke-testet, publiziert aber nichts.
- [x] Release-Compose im ZIP: gepinnter Tag, `name: jotti-local`, Volumes `caddy-data`/`proxy-state` erhalten, keine `build:`-Sektionen (alle vier Services image-basiert).
- [x] Beide Exes im ZIP sind Windows-amd64-Binaries und geben beim Start `vX.Y.Z` aus; `jotti-start.exe` fordert Administratorrechte an (Manifest).

---

## Phase 5: Dokumentation

**User stories**: 19, 20, 21, 22

### Context

- `docs/betrieb/leitfaden-hosting.md:42-128` — Weg A beschreibt derzeit den manuellen CLI-Ablauf.

### What to build

1. **`KURZANLEITUNG.md`** (im Release-ZIP, ≤ 1 Seite): Voraussetzung: Windows-Benutzer mit Administratorrechten; Docker Desktop installieren (einmalig — **nicht** vorab starten, das erledigt jotti); **Erststart mit Internet vorab zuhause** durchführen (Image-Download **und** erste Zertifikatsausstellung), nicht erst auf dem Fest; ZIP entpacken; SmartScreen-Dialog („Weitere Informationen → Trotzdem ausführen"); UAC-Dialog mit „Ja" bestätigen („Unbekannter Herausgeber" ist normal — die Programme sind nicht signiert); `jotti-start.exe` doppelklicken und warten — Firewall-Freigabe und Docker-Start passieren automatisch; danach die **Status-Seite `http://localhost:8484`** öffnen und den dort gezeigten QR-Code / die grüne Adresse an die Helfer-Handys geben (grünes Schloss, keine Warnung; solange das Zertifikat noch aussteht oder ohne Internet trägt der Fallback `https://<LAN-IP>` mit einmaliger Warnung pro Gerät); öffnet ein Handy die grüne Adresse nicht, hilft die Router-Anleitung (DNS-Rebind-Schutz) aus der Betriebsdoku; `jotti-relay.exe` für den Bondruck; bei „Port belegt durch ‚X'" das genannte Programm beenden und jotti erneut starten; Beenden per `jotti-stop.cmd` oder Docker Desktop — Daten **und Zertifikate** bleiben erhalten; am nächsten Tag dieselben zwei Doppelklicks inkl. UAC (die Adresse kann sich ändern, die Status-Seite zeigt sie erneut — dasselbe Zertifikat, keine neue Warnung); **niemals** Port-Weiterleitung ins Internet.
2. **Update `docs/betrieb/leitfaden-hosting.md`:** Weg A wurde durch Option 3 bereits auf das grüne Schloss + Status-Seite umgestellt (TLS-Plan Phase 6); hier kommt **davor** der Abschnitt „Starter verwenden (empfohlen)" hinzu — Verweis auf Release-ZIP und Kurzanleitung sowie die Admin-Rechte-Voraussetzung; der manuelle Weg bleibt als Alternative. Optionaler Tipp: DHCP-Reservierung im Router für eine stabile Adresse. (Den QR-/grüne-Adresse-Ablauf selbst beschreibt bereits der durch Option 3 aktualisierte Weg A — hier nicht duplizieren, nur auf den Starter als Einstieg verweisen.)

### Acceptance criteria

- [ ] `KURZANLEITUNG.md` existiert und beschreibt den vollständigen Ablauf inkl. SmartScreen-, UAC-, Zertifikats- und Beenden-Schritten in ≤ 1 Seite.
- [ ] Kein manueller Firewall-Schritt mehr in der Anleitung — Firewall und Docker-Start sind als automatisch beschrieben.
- [ ] Leitfaden Weg A verweist auf den Starter als empfohlenen Weg (inkl. Admin-Voraussetzung); der manuelle Weg bleibt dokumentiert.
- [ ] Sicherheitshinweis (nie ins Internet öffnen) in beiden Dokumenten; Zertifikatswarnung ist als Fallback-Fall beschrieben, grünes Schloss als Normalfall.
- [ ] Hinweis auf Erststart-mit-Internet (Image-Pull + Zertifikatsausstellung) und Daten-/Zertifikatspersistenz über Festtage in der Kurzanleitung.
