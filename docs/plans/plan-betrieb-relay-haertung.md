# Plan: Betriebs- und Relay-Härtung (Windows-unabhängig)

> Source PRD: `docs/prds/prd-betrieb-relay-haertung.md`
> Abgrenzung: Die klickbare Windows-EXE-Verpackung (`jotti-start.exe`,
> `.env`-Datei-Parsing im Relay) ist **bewusst nicht** Teil dieses Plans — sie
> ist in `docs/prds/prd-windows-verpackung.md` beschrieben und wird separat
> geplant.

## Goal

Die beim Erstellen der Windows-PRD gefundenen, **plattformunabhängigen** Mängel
beheben und damit den Betrieb (dev, local, prod) korrekter, sicherer und
einfacher machen — ohne neue Domänenfeatures:

- Der Bondruck (Relay) ist heute in **keiner** Umgebung funktionsfähig
  verdrahtet und gleichzeitig unsicher (leerer Token wird akzeptiert).
- Die Konfiguration ist inkonsistent (`JWT_SECRET` Pflicht, `RELAY_AUTH_TOKEN`
  still optional) und das Secret muss beim Relay über ein Prozess-Flag
  übergeben werden.
- Die Ersteinrichtung verlangt manuelle `.env`-Pflege inkl. `openssl`.
- Es gibt keinen nutzbaren Backend-Healthcheck für die Container-Orchestrierung.
- Die drei Compose-Dateien duplizieren denselben Service-Block mehrfach.

## Architectural decisions

Durable decisions, die für alle Phasen gelten:

- **Relay-Authentifizierung & -Konfiguration:** `RELAY_AUTH_TOKEN` ist die
  **einzige** Quelle des Relay-Secrets. Das Backend lädt ihn als **Pflichtwert**
  (fatal beim Start, wenn nicht gesetzt — analog `JWT_SECRET`); der Handler weist
  zusätzlich einen leeren **oder** abweichenden Token ab (Defense-in-Depth). Der
  Relay-Client wird **vollständig über Umgebungsvariablen** konfiguriert
  (`RELAY_AUTH_TOKEN`, `RELAY_BACKEND_URL` mit Default `http://localhost/api`,
  `RELAY_POLL_SECONDS` mit Default `2`) und **ohne Kommandozeilen-Flags**
  gestartet. Das deckt beide realen Szenarien mit **null Argumenten** ab:
  (1) alles auf einem Rechner und (2) Backend in Cloud/VPS, Relay lokal am
  Drucker (nur `RELAY_BACKEND_URL` zeigt dann auf die öffentliche Adresse).
- **`/health` ist ein Ops-Endpunkt, kein RPC-Endpunkt:** Er wird **bewusst von
  der harten „POST-only"-Guardrail ausgenommen** und per **GET** erreichbar
  gemacht. Alle übrigen Routen bleiben strikt POST-only. Diese Ausnahme wird im
  Handbuch als bewusste Entscheidung dokumentiert.
- **`.env`-Erzeugung:** Eine fehlende `.env` wird per `make init`
  (→ `scripts/init-env.sh`, bash + `openssl`) mit sicheren Zufallswerten
  erzeugt. Der Vorgang ist **idempotent** und überschreibt eine vorhandene
  `.env` **nie**.
- **Compose-Struktur:** Eine gemeinsame `docker-compose.base.yml` trägt die
  geteilten Services; `local` und `prod` werden **Overrides** (`-f base -f …`).
  Der **Dev-Stack bleibt eigenständig** (er nutzt Live-Reload-Images und
  unterscheidet sich grundlegend).

## Inventory

Relevante bestehende Dateien und Stellen (verifiziert):

- `backend/config/config.go:37` — `JWT_SECRET` via `parseEnvString` (Pflicht,
  fatal). `:43` — `RelayToken: os.Getenv("RELAY_AUTH_TOKEN")` (still optional).
  `:48` — `parseEnvString` ruft bei leerem Wert ohne Default `log.Fatalf`.
- `backend/api/relay/http/handler.go:46`, `:82` — `if body.Token != h.RelayToken`
  (leerer konfigurierter Token akzeptiert `{"token":""}`).
- `backend/app/app.go:45` — `/health` wird auf dem Haupt-Mux registriert. `:66` —
  Relay-Routen werden **immer** gemountet. `:72` — der gesamte Mux wird mit
  `PostMethodOnlyMiddleware` umhüllt.
- `backend/api/middleware/middleware.go:125` — `PostMethodOnlyMiddleware` weist
  jede Nicht-POST-Anfrage ab (auch `/health`).
- `backend/api/health/health.go:26` — `HealthCheck.Handler()` ist
  methodenunabhängig; die POST-Beschränkung kommt allein aus der Middleware.
- `cmd/relay/main.go:23` — `--backend` (Default `https://jotti.meinverein.de`).
  `:24` — `--token`. `:36` — `log.Fatal`, wenn `--token` leer. Eigenes Go-Modul
  `jotti-relay`, **ohne** bestehende Tests.
- `backend/config/config_test.go`, `backend/app/app_test.go:17,34,44,59`,
  `test-integration.sh:62` — rufen `config.Load()` und setzen heute **nur**
  `JWT_SECRET`.
- `docker-compose.yml` (Service `backend-dev`, env-Block),
  `docker-compose.local.yml` (Service `backend`),
  `docker-compose.prod.yml` (Service `backend`) — keiner setzt
  `RELAY_AUTH_TOKEN`.
- `.env.example` — enthält `POSTGRES_USER`, `POSTGRES_PASSWORD`, `JWT_SECRET`;
  **kein** `RELAY_AUTH_TOKEN`.
- `backend/Dockerfile:17` — Runtime-Image `alpine:3.22` (BusyBox-`wget`
  vorhanden → GET-Healthcheck trivial).
- `Makefile` — `local-up/down/logs`, `prod-up/down/logs`, `prod-reset-db`,
  `jotti-rocks-up/down/logs` referenzieren die Compose-Dateien direkt
  (`-f docker-compose.{prod,local}.yml`).
- `scripts/prod-init.sh:24-25,55`, `scripts/jotti-rocks-init.sh:25-26,55` —
  referenzieren Compose-Dateien direkt und brechen ab, wenn `.env` fehlt
  (erzeugen **keine** Secrets).
- `docker-compose.initial-cert.yml` — eigenständig (nur reverse-proxy +
  certbot), **referenziert keine Basis-Services** → von Phase 5 unberührt.
- `docs/betrieb/leitfaden-hosting.md:83-87` — manuelle `cp .env.example .env` +
  `openssl rand`. `README.md:60` — `cp .env.example .env`.
  `docs/handbuch.md:708,725,817` — Relay-Token-Dokumentation.

## Resolved decisions

- **Scope:** alle gefundenen Issues (a–g) sind Teil dieses Plans.
- **Token-Absicherung:** Pflicht beim Start (fatal, wie `JWT_SECRET`) **und**
  Ablehnung leerer/abweichender Tokens im Handler.
- **Relay-Konfiguration:** vollständig über Umgebungsvariablen
  (`RELAY_AUTH_TOKEN`, `RELAY_BACKEND_URL`, `RELAY_POLL_SECONDS`); **alle Flags
  (`--token`, `--backend`, `--poll`) entfallen**. Da der Token aus
  Sicherheitsgründen ohnehin aus der Umgebung kommt, würde ein verbleibendes
  `--backend`-Flag das „Doppelklick ohne Argumente"-Ziel in Szenario 2
  unterlaufen — eine einzige Konfigurationsquelle ist einfacher und konsistent.
- **`/health`:** GET-Ausnahme von der POST-only-Guardrail; Container-Healthcheck
  per GET.
- **`.env`-Generierung:** `make init` → `scripts/init-env.sh` (bash + `openssl`),
  idempotent, überschreibt nie.
- **Compose-Entdoppelung:** gemeinsame Basis nur für `local` + `prod`; Dev bleibt
  eigenständig.
- **Docs:** betroffene Dokumentation im selben Plan aktualisieren
  (Hosting-Leitfaden, Handbuch, README).

## Open questions / Risks

- **Korrigiert ein PRD-Detail (kein zweites Gerät als Sonderfall):** Es gibt
  **zwei** reale Betriebsszenarien — (1) alles auf einem Rechner, (2) Backend in
  Cloud/VPS + Relay lokal am Drucker. Beide nutzen dieselbe env-basierte
  Relay-Konfiguration (Szenario 2 setzt `RELAY_BACKEND_URL` auf die öffentliche
  Adresse); einen „Relay auf zweiter Station"-Sonderfall mit `--token`/`--backend`-
  Override gibt es nicht. Die zwei Szenarien sind im Source-PRD
  `docs/prds/prd-betrieb-relay-haertung.md` (Abschnitt „Further Notes")
  dokumentiert.
- **Pflicht-Token-Ripple:** `RELAY_AUTH_TOKEN` als Pflicht bricht jeden
  `config.Load()`-Aufrufer, der ihn nicht setzt — betrifft `config_test.go`,
  `app_test.go` und `test-integration.sh`. Diese müssen den Token setzen, sonst
  `log.Fatal`. (Der Fatal-Pfad selbst wird — konsistent mit `JWT_SECRET` — nicht
  unit-getestet.)
- **Guardrail-Ausnahme:** Die GET-Freigabe für `/health` ist eine bewusste
  Aufweichung der „POST-only"-Guardrail und muss im Handbuch klar als
  Ops-Ausnahme markiert werden, damit sie nicht als Regelbruch missverstanden
  wird.
- **Compose-Radius:** Die Base/Override-Umstellung berührt viele Makefile-Targets
  und beide Init-Skripte. Jeder Aufrufpfad muss nach der Umstellung dieselbe
  Service-Menge auflösen wie zuvor (`docker compose … config` zur Kontrolle).

---

## Phase 1: Relay-Token verpflichtend & abgesichert (Backend)

**Issues:** (a) Verdrahtung, (b) Sicherheit, (c) Konsistenz

### Context

- `backend/config/config.go:37,43,48` — `JWT_SECRET` ist Pflicht, `RELAY_AUTH_TOKEN`
  nicht; `parseEnvString` liefert den fatalen Pflicht-Mechanismus.
- `backend/api/relay/http/handler.go:46,82` — Token-Vergleich, der leeren Token
  durchlässt.
- `backend/app/app.go:66` — Relay-Routen werden immer gemountet.
- `docker-compose.yml`, `docker-compose.local.yml`, `docker-compose.prod.yml` —
  Backend-Service-Env ohne `RELAY_AUTH_TOKEN`.
- `.env.example` — ohne `RELAY_AUTH_TOKEN`.
- `backend/config/config_test.go`, `backend/app/app_test.go:17,34,44,59`,
  `test-integration.sh:62` — setzen heute nur `JWT_SECRET`.

### What to build

Den Relay-Token zu einem vollwertigen Pflicht-Secret machen: Das Backend lädt
`RELAY_AUTH_TOKEN` über denselben Pflicht-Mechanismus wie `JWT_SECRET` (fatal,
wenn nicht gesetzt). Der Relay-Handler lehnt Anfragen ab, sobald der
konfigurierte Token leer ist oder der übergebene Token nicht exakt passt. Der
Token wird in der Backend-Env aller drei Compose-Dateien gesetzt (aus der `.env`
übergeben) und in `.env.example` dokumentiert. Alle `config.Load()`-Aufrufer in
Tests/Skripten setzen den Token, damit die bestehende Suite grün bleibt.

### Acceptance criteria

- [x] Startet das Backend ohne gesetztes `RELAY_AUTH_TOKEN`, bricht es mit klarer
      Fehlermeldung ab (analog `JWT_SECRET`).
- [x] Eine Relay-Anfrage mit leerem oder falschem Token erhält `unauthorized`,
      auch wenn der konfigurierte Token leer wäre.
- [x] `RELAY_AUTH_TOKEN` wird in der Backend-Env von `docker-compose.yml`,
      `docker-compose.local.yml` und `docker-compose.prod.yml` aus der `.env`
      übergeben und ist in `.env.example` dokumentiert.
- [x] `config_test.go`, `app_test.go` und `test-integration.sh` setzen den Token;
      `make test` und die Integrationstests laufen unverändert grün.
- [x] Unit-Tests decken ab: Token vorhanden → akzeptiert; leerer/falscher Token →
      abgelehnt.

---

## Phase 2: Relay vollständig über Umgebungsvariablen konfigurieren (Relay-Binary)

**Issues:** (d) Secret nicht mehr im Prozess-Argument

### Context

- `cmd/relay/main.go:23,24,36` — `--backend`-Default, `--token`-Flag, Fatal bei
  leerem Token. Eigenes Modul `jotti-relay`, bisher ohne Tests.

### What to build

Der Relay-Client wird vollständig über Umgebungsvariablen konfiguriert und ohne
Kommandozeilen-Argumente gestartet: `RELAY_AUTH_TOKEN` (Pflicht; Abbruch mit
klarer Meldung, wenn leer), `RELAY_BACKEND_URL` (Default `http://localhost/api`
für den Alltagsfall „alles auf einem Rechner") und `RELAY_POLL_SECONDS`
(Default `2`). Die Flags `--token`, `--backend` und `--poll` entfallen ersatzlos;
der bisherige Backend-Platzhalter `https://jotti.meinverein.de` verschwindet
damit ebenfalls. Die Konfigurations-Auflösung wird in eine kleine, testbare
Funktion gezogen, sodass das `jotti-relay`-Modul seinen ersten Unit-Test erhält.
Der Dev-/Test-Workflow in der Bondruck-Doku wird auf Umgebungsvariablen
umgestellt (Szenario 2 setzt zusätzlich `RELAY_BACKEND_URL` auf die öffentliche
Adresse).

### Acceptance criteria

- [x] `jotti-relay` startet allein mit gesetzten Umgebungsvariablen und **ohne
      jedes Kommandozeilen-Argument** in beiden Szenarien.
- [x] Fehlt `RELAY_AUTH_TOKEN`, bricht das Relay mit einer verständlichen Meldung
      ab.
- [x] Ohne `RELAY_BACKEND_URL` nutzt das Relay den Default `http://localhost/api`;
      ein gesetzter Wert (z. B. VPS-Adresse) wird verwendet.
- [x] Die Flags `--token`, `--backend`, `--poll` existieren nicht mehr.
- [x] Ein Unit-Test im `jotti-relay`-Modul prüft die Konfigurations-Auflösung
      (Token gesetzt/fehlend, Backend-URL Default/überschrieben).
- [x] Die Bondruck-Test-Doku beschreibt den Start per Umgebungsvariablen.

---

## Phase 3: Ein-Kommando-`.env`-Erzeugung

**Issues:** (e) keine manuelle Secret-Pflege mehr

### Context

- `scripts/prod-init.sh:55`, `scripts/jotti-rocks-init.sh:55` — brechen ab, wenn
  `.env` fehlt; erzeugen keine Secrets.
- `docs/betrieb/leitfaden-hosting.md:83-87`, `README.md:60` — manuelle
  `cp .env.example .env` + `openssl rand`-Anleitung.
- `.env.example` — Vorlage der benötigten Schlüssel (nach Phase 1 inkl.
  `RELAY_AUTH_TOKEN`).

### What to build

Ein idempotentes Einrichtungs-Kommando: `make init` ruft `scripts/init-env.sh`,
das eine fehlende `.env` mit kryptografisch sicheren Zufallswerten für alle
Secrets (`POSTGRES_PASSWORD`, `JWT_SECRET`, `RELAY_AUTH_TOKEN`) erzeugt und
sinnvolle Defaults für die übrigen Schlüssel setzt. Eine vorhandene `.env` wird
nie überschrieben. Die Init-Skripte und die Doku verweisen auf `make init`
statt auf die manuelle `openssl`-Prozedur.

### Acceptance criteria

- [x] `make init` erzeugt aus einem frischen Checkout eine vollständige `.env`
      mit sicheren Zufalls-Secrets, ohne weitere manuelle Eingriffe.
- [x] Ein zweiter Aufruf von `make init` lässt eine vorhandene `.env` unverändert
      (idempotent, keine Überschreibung).
- [x] Die erzeugte `.env` enthält alle Schlüssel, die dev/local/prod benötigen
      (inkl. `RELAY_AUTH_TOKEN`).
- [x] `prod-init.sh`/`jotti-rocks-init.sh` verweisen bei fehlender `.env` auf
      `make init`.
- [x] Hosting-Leitfaden und README beschreiben `make init` statt der manuellen
      `openssl`-Schritte.

---

## Phase 4: Backend-Healthcheck für die Orchestrierung

**Issues:** (f) `/health` nutzbar machen

### Context

- `backend/app/app.go:45,72` — `/health` hängt hinter `PostMethodOnlyMiddleware`.
- `backend/api/middleware/middleware.go:125` — weist Nicht-POST ab.
- `backend/api/health/health.go:26` — Handler ist methodenunabhängig.
- `backend/Dockerfile:17` — `alpine:3.22` (BusyBox-`wget` vorhanden).
- `docker-compose.local.yml`, `docker-compose.prod.yml` — Backend ohne
  Healthcheck; reverse-proxy hängt an `service_started`.

### What to build

`/health` wird als Ops-Endpunkt von der POST-only-Regel ausgenommen und per GET
erreichbar gemacht, während alle übrigen Routen strikt POST-only bleiben. Der
Backend-Service erhält in `local` und `prod` einen Docker-Healthcheck (GET via
BusyBox-`wget` gegen `/health`), und der reverse-proxy wartet auf
`condition: service_healthy` des Backends. Die bewusste Guardrail-Ausnahme wird
im Handbuch dokumentiert.

### Acceptance criteria

- [x] `GET /health` liefert `200` (bzw. `503` bei DB-Problemen); andere Routen
      weisen GET weiterhin ab.
- [x] Der Backend-Service meldet in `local` und `prod` einen Docker-Health-Status
      (`healthy`/`unhealthy`).
- [x] Der reverse-proxy startet erst, wenn das Backend `healthy` ist.
- [x] Ein Test belegt: `GET /health` erlaubt, `GET` auf eine andere Route
      abgelehnt.
- [x] Das Handbuch dokumentiert die GET-Ausnahme für `/health` als bewusste
      Ops-Entscheidung.

---

## Phase 5: Compose-Entdoppelung (Basis + Override für local/prod)

**Issues:** (g) Duplikation reduzieren

### Context

- `docker-compose.yml`, `docker-compose.local.yml`, `docker-compose.prod.yml` —
  duplizieren `postgres`, `migrate`, `backend`, `frontend`.
- `Makefile` — `local-*`, `prod-*`, `jotti-rocks-*` referenzieren die
  Compose-Dateien direkt.
- `scripts/prod-init.sh:24-25`, `scripts/jotti-rocks-init.sh:25-26` —
  Compose-Datei-Referenzen.
- `docker-compose.initial-cert.yml` — eigenständig, unberührt.
- Dev-Stack (`docker-compose.yml`) — Live-Reload-Images, bleibt eigenständig.

### What to build

Die für `local` und `prod` gemeinsamen Services wandern in eine neue
`docker-compose.base.yml`; `local` und `prod` enthalten nur noch ihre
Unterschiede (reverse-proxy/TLS, certbot, Ports, Resource-Limits) als Override.
Alle Makefile-Targets und die beiden Init-Skripte werden auf `-f base -f override`
umgestellt. Der Dev-Stack bleibt unverändert eigenständig. Die `RELAY_AUTH_TOKEN`-
und Healthcheck-Ergänzungen aus Phase 1 und 4 werden dabei in die Basis
konsolidiert.

### Acceptance criteria

- [x] `local`- und `prod`-Stack starten über `-f base -f override` und verhalten
      sich identisch zu vorher (gleiche Services, Ports, Volumes, Env).
- [x] `docker compose … config` löst für jedes Target dieselbe effektive
      Konfiguration auf wie vor der Umstellung.
- [x] Alle betroffenen Makefile-Targets (`local-*`, `prod-*`, `jotti-rocks-*`)
      und beide Init-Skripte verwenden die Base/Override-Aufrufe.
- [x] Geteilte Service-Definitionen stehen genau einmal in der Basis; `local`/
      `prod` enthalten nur noch ihre Abweichungen.
- [x] Der Dev-Stack ist unverändert und weiterhin per `make dev` lauffähig.
