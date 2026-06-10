# Plan: Klickbare Windows-Verpackung für den lokalen Betrieb

> Source PRD: `docs/prds/prd-windows-verpackung.md`

## Goal

Ein Vereins-Admin kann jotti auf einem Windows-Rechner per Doppelklick starten (`jotti-start.exe`) und den Bondruck separat per Doppelklick aktivieren (`jotti-relay.exe`) — ohne `.env` von Hand anlegen, Geheimnisse erzeugen oder Kommandozeile benutzen zu müssen. Verteilt als GitHub-Release-ZIP mit vorgebauten Windows-Binaries.

## Architectural decisions

- **Quellort Starter:** `cmd/starter/` als eigenständiges Go-Modul (separates `go.mod`, analog `cmd/relay/`).
- **Port-Konfiguration:** Env-Variable `HTTPS_PORT` (Default `443`). Compose-Datei referenziert `${HTTPS_PORT:-443}`. HTTP-Port bleibt fest 80 (nur Redirect).
- **.env-Parsing:** Jedes Binary (`starter`, `relay`) enthält seinen eigenen minimalen Key=Value-Parser (~20 Zeilen). Kein Shared-Package — hält die Module unabhängig.
- **Firewall-Hinweis:** Der Starter gibt bei Erfolg immer einen Hinweis zur Windows-Firewall aus (kein programmatischer Check).
- **Health-Check:** Starter ruft `GET https://localhost:{HTTPS_PORT}/health` auf (über den Reverse-Proxy, TLS-Verify aus wegen selbstsigniertem Zert).
- **Release:** Make-Target `release-windows` + GitHub Actions Workflow für automatisierte Release-Builds.

## Inventory

- `cmd/relay/main.go:26-63` — Relay-Config-Loading (`loadConfigFromEnv`), zeigt das Pattern für env-basierte Konfiguration.
- `cmd/relay/main.go:64-115` — Relay-Main-Loop, dient als Referenz für Startup-Ausgaben und Signal-Handling.
- `cmd/relay/go.mod` — Eigenständiges Modul `jotti-relay`, go 1.26.0, keine Dependencies.
- `scripts/init-env.sh:18-75` — Bestehende `.env`-Erzeugung (openssl-basiert, idempotent). Die Starter-Core-Logik reimplementiert dies in Go (kryptografisch sicher, ohne openssl-Abhängigkeit).
- `.env.example:1-11` — Template mit `POSTGRES_USER`, `POSTGRES_PASSWORD`, `JWT_SECRET`, `RELAY_AUTH_TOKEN`.
- `docker-compose.local.yml:86-105` — Reverse-Proxy-Service, Ports derzeit hardcodiert `80:80` und `443:443`.
- `reverse-proxy/local-entrypoint.sh:16-88` — LAN-IP-Erkennung + Zertifikat-Erzeugung (Referenz für IP-Logik im Starter).
- `backend/api/health/health.go:22-41` — Health-Handler (GET, JSON `{"status":"ok","timestamp":"…"}`).
- `Makefile:88-93` — Bestehende Build-Targets (`build-backend`, `build-relay`, `build-frontend`).
- `Makefile:153-161` — Lokaler Betrieb (`local-up`, `local-down`, `local-logs`).
- `docs/betrieb/leitfaden-hosting.md:57-112` — Weg A Anleitung, wird um Starter-Referenz ergänzt.

## Resolved decisions

- Starter lebt in `cmd/starter/` mit eigenem `go.mod` (kein Dependency-Sharing mit Backend).
- Portname: `HTTPS_PORT` (Default 443). Nur ein Port konfigurierbar; HTTP (80) dient nur dem Redirect.
- Jedes Binary parsed `.env` selbst — kein Shared-Package, Module bleiben entkoppelt.
- Release-Workflow via GitHub Actions (nicht nur lokales Make-Target).
- Firewall-Hinweis immer als Teil der Erfolgsausgabe (kein programmatischer netsh-Check).

## Open questions / Risks

- **Docker-CLI-Pfad auf Windows:** `docker compose` muss im PATH liegen (Docker Desktop installiert es standardmäßig nach `C:\Program Files\Docker\Docker\resources\bin`). Falls nicht gefunden, ist der Preflight-Fehler klar — aber der Hinweis muss den genauen Pfad nennen.
- **Antivirus-False-Positives:** Selbst-kompilierte unsigned `.exe`-Dateien können Windows Defender triggern. Code-Signing ist out-of-scope, aber die Kurzanleitung sollte darauf hinweisen.

---

## Phase 1: Starter-Core (reine Logik + Unit-Tests)

**User stories**: 2, 3, 4, 5, 6, 8, 11, 12, 14

### Context

- `scripts/init-env.sh:34-40` — Secret-Erzeugung via `openssl rand -base64 32`; die Go-Implementierung nutzt `crypto/rand` direkt.
- `reverse-proxy/local-entrypoint.sh:16-34` — LAN-IP-Erkennung über Default-Gateway-Interface; Go-Äquivalent nutzt `net.Interfaces()`.
- `.env.example:1-11` — Die erwarteten Keys und Defaults.

### What to build

Ein Go-Package `cmd/starter/core` mit seiteneffektfreien, unit-testbaren Funktionen:

1. **Secret-Erzeugung** — `GenerateSecret() string`: 32 Bytes aus `crypto/rand`, base64-kodiert.
2. **`.env`-Materialisierung** — `MaterializeEnv(targetPath string, exists func(string) bool, write func(string, []byte) error) error`: Erzeugt die `.env` nur wenn sie fehlt; setzt `POSTGRES_USER=admin`, `POSTGRES_PASSWORD`, `JWT_SECRET`, `RELAY_AUTH_TOKEN` (generiert), `HTTPS_PORT=443`. Vorhandene Datei wird nie überschrieben.
3. **Preflight-Auswertung** — `EvaluatePreflights(checks []PreflightResult) []Diagnosis`: Bildet Ergebnisse (Docker vorhanden? Docker gestartet? Port frei?) auf deutsche Klartext-Diagnosen mit Handlungshinweis ab.
4. **LAN-IP-Auswahl** — `SelectLANIP(interfaces []NetInterface) (string, error)`: Wählt die erste private IPv4 (RFC 1918), ignoriert Loopback und Link-Local.
5. **Zugriffs-URL-Bau** — `BuildAccessURL(ip string, port int) string`: Gibt `https://{ip}` (Port 443) oder `https://{ip}:{port}` (anderer Port) zurück.

### Acceptance criteria

- [ ] `cmd/starter/` existiert mit eigenem `go.mod` (module `jotti-starter`, go 1.26).
- [ ] Alle fünf Funktionen sind implementiert und unit-getestet (`-tags=unit`).
- [ ] `.env`-Idempotenz: Test belegt, dass vorhandene Datei nie überschrieben wird.
- [ ] Secret-Entropie: Zwei Aufrufe liefern unterschiedliche Werte; Länge ≥ 32 Base64-Zeichen.
- [ ] LAN-IP: Loopback (127.x), Link-Local (169.254.x) werden ignoriert; private IPs (10.x, 172.16-31.x, 192.168.x) werden gewählt.
- [ ] URL-Bau: Default-Port → kein Port-Suffix; abweichender Port → mit Suffix.
- [ ] Preflight: Jede Fehlerbedingung bildet auf eine deutsche Diagnose mit Handlungshinweis ab.
- [ ] `go build ./...` und `go test -tags=unit ./...` in `cmd/starter/` erfolgreich.

---

## Phase 2: Compose-Port-Parameterisierung

**User stories**: 5, 12

### Context

- `docker-compose.local.yml:96-97` — Hardcodierte Ports `80:80` und `443:443`.
- `.env.example:1-11` — Bestehendes Template ohne Port-Variable.
- `scripts/init-env.sh:42-69` — AWK-Verarbeitung des Templates.

### What to build

Die lokale Compose-Datei referenziert `${HTTPS_PORT:-443}` für den HTTPS-Port des Reverse-Proxy. `.env.example` erhält den neuen Key `HTTPS_PORT=443`. Der Starter schreibt den Key ebenfalls in die generierte `.env`. Die bestehende `init-env.sh` ignoriert den neuen Key (er hat bereits einen Default-Wert im Template und braucht keine Secret-Generierung).

### Acceptance criteria

- [ ] `docker-compose.local.yml` Reverse-Proxy-Port-Zeile nutzt `${HTTPS_PORT:-443}:443`.
- [ ] `.env.example` enthält `HTTPS_PORT=443` mit kurzem Kommentar.
- [ ] `make local-up` funktioniert weiterhin ohne gesetzte `HTTPS_PORT`-Variable (Default greift).
- [ ] Mit `HTTPS_PORT=8443` in `.env` bindet der Proxy auf Host-Port 8443.

---

## Phase 3: Starter-Shell + Binary

**User stories**: 7, 8, 9, 13, 14

### Context

- `cmd/starter/core` (Phase 1) — liefert die reine Logik.
- `docker-compose.local.yml` — der zu startende Stack.
- `backend/api/health/health.go:22-41` — Health-Endpoint für den Check.

### What to build

Die `main.go` in `cmd/starter/` verbindet Core-Logik mit echten Seiteneffekten:

1. `.env` materialisieren (Core-Funktion aufrufen, echtes Dateisystem).
2. Preflight-Checks ausführen: `docker` im PATH? `docker info` erfolgreich? HTTPS-Port frei?
3. Bei Fehler: deutsche Diagnose ausgeben, Exit-Code 1.
4. `docker compose -f docker-compose.local.yml up -d --build` ausführen.
5. Health-Check-Loop: bis zu 60 s `GET https://localhost:{port}/health` pollen (TLS-InsecureSkipVerify). Bei Timeout: Warnung + Exit-Code 1.
6. Erfolgsausgabe: Zugriffs-URL, Firewall-Hinweis, Sicherheitshinweis (nie ins Internet öffnen). Exit-Code 0.

Make-Target `build-starter-windows` kompiliert `GOOS=windows GOARCH=amd64` nach `cmd/starter/jotti-start.exe`.

### Acceptance criteria

- [ ] `cmd/starter/main.go` implementiert den beschriebenen Ablauf.
- [ ] Exit-Code 0 bei Erfolg, 1 bei jedem Preflight-/Health-Fehler.
- [ ] Ausgabe enthält `https://{LAN-IP}` (bzw. mit Port), Firewall-Hinweis, Sicherheitswarnung.
- [ ] `make build-starter-windows` erzeugt `cmd/starter/jotti-start.exe`.
- [ ] Compose-Pfad ist relativ zum Binary-Standort (funktioniert aus dem entpackten ZIP).

---

## Phase 4: Relay .env-Parser + Windows-Startup

**User stories**: 15, 16, 17, 18

### Context

- `cmd/relay/main.go:37-63` — `loadConfigFromEnv(getenv)` nimmt eine `func(string) string`; der Parser muss diese Signatur bedienen.
- `cmd/relay/main.go:64-70` — Startup-Log zeigt Backend-URL und Poll-Intervall.

### What to build

Ein minimaler `.env`-Parser in `cmd/relay/` (eigene Datei, ~20 Zeilen): liest Key=Value-Zeilen, ignoriert Kommentare (`#`) und Leerzeilen, trimmt Whitespace und optionale Anführungszeichen. Die `main()`-Funktion prüft, ob Env-Variablen bereits gesetzt sind (Docker-Injection); falls nicht, lädt sie `.env` aus dem aktuellen Verzeichnis und setzt die Werte als Fallback.

Damit funktioniert `jotti-relay.exe` per Doppelklick neben der `.env` ohne manuelle Eingabe, aber auch weiterhin im Docker-Compose-Stack (wo Env schon gesetzt ist).

### Acceptance criteria

- [ ] `.env`-Parser unit-getestet: Key=Value, Kommentare ignoriert, Anführungszeichen gestrippt, leere Zeilen übersprungen.
- [ ] Relay startet per Doppelklick neben `.env` (Token + Backend-URL aus Datei).
- [ ] Im Docker-Compose-Stack (env-injected) ändert sich das Verhalten nicht (vorgesetzte Env hat Vorrang).
- [ ] Fehlt `RELAY_AUTH_TOKEN` in `.env` und Env: klare Fehlermeldung auf Deutsch.
- [ ] `RELAY_BACKEND_URL` fällt auf `https://localhost/api` zurück (nicht `http://` — lokaler Stack hat TLS).
- [ ] Make-Target `build-relay-windows` erzeugt `cmd/relay/jotti-relay.exe` (GOOS=windows GOARCH=amd64).

---

## Phase 5: Release-Packaging (Make + GitHub Actions)

**User stories**: 1

### Context

- `Makefile:88-93` — bestehende Build-Targets.
- `.github/` — Verzeichnis für Workflows (falls vorhanden).

### What to build

1. **Make-Target `release-windows`**: Cross-kompiliert beide Binaries, kopiert `docker-compose.local.yml`, `reverse-proxy/nginx.local.conf`, `reverse-proxy/local-entrypoint.sh`, `.env.example` und die Kurzanleitung (Phase 6) in ein Staging-Verzeichnis, erzeugt `dist/jotti-windows-{version}.zip`.
2. **GitHub Actions Workflow** (`.github/workflows/release-windows.yml`): Triggered auf Tag-Push (`v*`). Checkt aus, ruft `make release-windows` auf, erstellt GitHub Release mit dem ZIP als Asset.

### Acceptance criteria

- [ ] `make release-windows` erzeugt ein ZIP unter `dist/` mit korrektem Inhalt.
- [ ] ZIP enthält: `jotti-start.exe`, `jotti-relay.exe`, `docker-compose.local.yml`, `reverse-proxy/` (nginx.local.conf + local-entrypoint.sh), `.env.example`, `KURZANLEITUNG.md`.
- [ ] GitHub Actions Workflow baut auf Tag-Push und publiziert Release mit ZIP-Asset.
- [ ] Binaries im ZIP sind lauffähige Windows-amd64-Executables.

---

## Phase 6: Dokumentation

**User stories**: 19, 20, 21, 22

### Context

- `docs/betrieb/leitfaden-hosting.md:57-112` — Weg A beschreibt derzeit den manuellen CLI-Ablauf.

### What to build

1. **`KURZANLEITUNG.md`** (liegt im Release-ZIP): Schritt-für-Schritt-Anleitung für den Zwei-Doppelklick-Ablauf. Voraussetzung Docker Desktop, Entpacken, `jotti-start.exe` doppelklicken, warten, URL ans Handy geben, `jotti-relay.exe` für Bondruck. Beenden: `docker compose -f docker-compose.local.yml down` oder Docker Desktop stoppen.
2. **Update `docs/betrieb/leitfaden-hosting.md`**: Weg A erhält einen Abschnitt „Starter verwenden (empfohlen)" vor dem manuellen Weg, der auf das Release-ZIP und die Kurzanleitung verweist. Der manuelle Weg bleibt als Alternative erhalten.

### Acceptance criteria

- [ ] `KURZANLEITUNG.md` existiert, beschreibt den vollständigen Ablauf in ≤ 1 Seite.
- [ ] Leitfaden Weg A verweist auf Starter als empfohlenen Weg.
- [ ] Manueller Weg bleibt im Leitfaden als Alternative dokumentiert.
- [ ] Sicherheitshinweis (nie ins Internet öffnen, Zertifikatswarnung) in beiden Dokumenten.
