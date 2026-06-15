# Plan: VPS-Self-Hoster-Pfad auf Windows-Parität bringen

> Source PRD: n/a (aus Deployment-Review und Findings, siehe Konversation)

## Goal

Der Self-Hoster-VPS-Pfad (Weg B) soll mindestens so professionell, automatisiert,
sicher und benutzerfreundlich sein wie der Windows-Pfad. Konkret:

- Erststart ohne jede Dateiänderung lauffähig (heute effektiv kaputt: hartkodierte
  `jotti.rocks`-Domain im Init-Script plus `<domain>`-Platzhalter in der nginx-Config).
- Automatisches HTTPS ohne certbot-Choreografie.
- Versionierte, gepinnte Images statt Build-aus-Source auf dem VPS.
- Backup, Restore und sicheres Update als Tooling (nicht nur als Doku-Hinweis).
- Opt-in-Serverhärtung.

## Architectural decisions

Durchgängige Entscheidungen über alle Phasen:

- **TLS/Reverse-Proxy (Self-Hoster):** Caddy mit automatischem Public-HTTPS
  (HTTP-01/TLS-ALPN). Ersetzt nginx + certbot + `docker-compose.initial-cert.yml`
  - Renewal-Sidecar + `<domain>`-Templating. Wiederverwendet das bestehende
    `jotti-reverse-proxy`-Image (xcaddy) und das `(jotti_proxy)`-Snippet.
- **TLS (jotti.rocks):** bleibt nginx + certbot (Multi-Domain Landing/Demo/Auth +
  acme-dns). rocks wird dafür vom Self-Hoster-prod-File entkoppelt.
- **Konfigurationsquelle:** Domain, E-Mail und Version kommen aus `.env`
  (`JOTTI_DOMAIN`, `LETSENCRYPT_EMAIL`, `JOTTI_VERSION`). Kein Edit getrackter
  Dateien mehr.
- **Images:** Self-Hoster zieht gepinnte `ghcr.io/nicograef/jotti-*:${JOTTI_VERSION}`
  (Parität zu `docker-compose.release.yml`). rocks baut weiter aus Source (Staging).
- **Admin-Tooling:** Bash-Scripts unter `scripts/` + `make`-Targets + systemd/cron
  (passt zur bestehenden `scripts/`- und Makefile-Welt; kein Build-Schritt nötig).
- **Härtung:** opt-in Script plus Doku-Abschnitt, niemals verpflichtend im Deploy.
- **Projektname/Volumes:** überall `name: jotti` und unveränderte Volume-Namen
  (`postgres-data`, `letsencrypt`, `certbot-challenges`), damit Live-Daten und
  Zertifikate erhalten bleiben.

## Inventory

- `docker-compose.prod.yml:44-99` — backend/frontend bauen aus Source (`build:`),
  kein gepinntes Image.
- `docker-compose.prod.yml:101-143` — reverse-proxy (stock nginx, mountet
  `reverse-proxy/nginx.conf` read-only) + certbot-Renewal-Sidecar.
- `docker-compose.rocks.yml:14-23` — rocks override hängt am prod-File und tauscht
  nur die nginx-Config/fügt Landing/resolver/acme-dns hinzu.
- `docker-compose.initial-cert.yml` — Einmal-Cert-Stack (nur nginx + certbot),
  genutzt von `prod-init.sh` und `rocks-init.sh`.
- `reverse-proxy/nginx.conf` — Self-Hoster-Config mit `<domain>`-Platzhaltern
  (server_name, beide Cert-Pfade, Redirect); read-only gemountet, kein envsubst.
- `reverse-proxy/nginx.rocks.conf:6-126` — rocks-Multi-Domain-Config inkl.
  `limit_req` (10r/s, burst 20), Security-Header, CSP; bleibt für rocks.
- `reverse-proxy/caddyfile.go:41-117` — `renderCaddyfile` + `(jotti_proxy)`-Snippet
  (Security-Header, CSP, `handle_path /api/*`, SPA-Proxy). Basis für den Public-Mode.
- `reverse-proxy/main.go:43-53` — `loadConfig` liest Env (`LAN_IP`, `PROXY_ZONE`,
  `ACMEDNS_BASE_URL`, `PROXY_LE_STAGING` …). Hier kommt der Public-Mode-Schalter rein.
- `reverse-proxy/Dockerfile:13-16` — xcaddy-Build mit `caddy-dns/acmedns`. Hier
  kommt das Rate-Limit-Modul dazu.
- `reverse-proxy/caddyfile_test.go:1-30` — Testmuster für den Renderer (string-Asserts).
- `scripts/prod-init.sh:19-25` — hartkodierte `DOMAIN="jotti.rocks"` / `EMAIL`.
- `scripts/rocks-init.sh:21-28` — rocks-Init (Multi-Domain), bleibt nginx+certbot.
- `scripts/init-env.sh` — erzeugt `.env` aus `.env.example`, füllt nur Secrets.
- `.env.example:13-16` — nur `VPS_PUBLIC_IP` (für rocks); keine Domain/E-Mail/Version.
- `Makefile:153-189` — prod-_ und rocks-_ Targets (`prod-up` macht `up -d --build`).
- `docker-compose.release.yml:54,70,102,113` — Vorlage für gepinnte GHCR-Images.
- `.github/workflows/release.yml:49-71` — baut/pusht u. a. `jotti-reverse-proxy`.
- `cmd/starter/core/backup.go` — `ShouldBackup`/`DumpsToDelete` (Rotation-Logik,
  Vorbild für die Bash-Variante).
- `cmd/starter/core/update.go` — `IsNewerVersion`/`parseSemver` (Downgrade-Guard-Vorbild).
- `packaging/windows/KURZANLEITUNG.md:76-124` — Windows-UX für Update/Restore/Repair,
  die der VPS-Pfad spiegeln soll.
- `docs/betrieb/leitfaden-hosting.md:168-231` — Weg B, wird neu gefasst.

## Resolved decisions

- Self-Hoster-VPS auf Caddy Auto-TLS umstellen (statt nginx+certbot env-templaten).
- Admin-Tooling als Bash + make + cron/systemd (nicht als Go-CLI).
- Härtung als optionales `scripts/prod-harden.sh` + Doku-Abschnitt (opt-in).
- rocks bleibt nginx; wird vom Self-Hoster-prod-File entkoppelt.
- Domain/E-Mail/Version aus `.env`; gepinnte GHCR-Images für Self-Hoster.

## Open questions / Risks

- **Rate-Limiting im Public-Mode** braucht ein Caddy-Modul (z. B.
  `github.com/mholt/caddy-ratelimit`) im xcaddy-Build, da Stock-Caddy kein
  Rate-Limit kennt. Alternative wäre, das Limit ins Backend zu ziehen; der Plan
  hält an Parität zu nginx (10r/s, burst 20) per Modul fest.
- **HSTS-Parität:** Der lokale `(jotti_proxy)`-Header nutzt `max-age=31536000`
  ohne `includeSubDomains`/`preload`; die prod-nginx nutzt
  `max-age=63072000; includeSubDomains; preload`. Im Public-Mode den stärkeren
  Wert setzen (bewusste Public-Domain-Zusage).
- **Live-Demo-Risiko:** Der prod.yml-Umbau darf `jotti.rocks` nicht brechen,
  daher rocks-Entkopplung als erste Phase und Volume-/Projektnamen erhalten.
- **Caddy-Smoke-Test in CI:** Echte Ausstellung braucht öffentliches DNS; in CI
  nur Renderer-Unit-Tests plus `caddy validate` der gerenderten Config.

---

## Phase 1: rocks entkoppeln & nginx-Pfad erhalten

### Context

- `docker-compose.rocks.yml:1-23` — heute Override auf `docker-compose.prod.yml`.
- `docker-compose.prod.yml:1-156` — gemeinsame Basis (postgres, migrate, backend,
  frontend, reverse-proxy[nginx], certbot), die Phase 2 umbaut.
- `docker-compose.initial-cert.yml` — von `rocks-init.sh` weiter genutzt.
- `scripts/rocks-init.sh:27-28` — `COMPOSE_PROD=(-f docker-compose.prod.yml -f docker-compose.rocks.yml)`.
- `Makefile:170-189` — rocks-\* Targets mit doppeltem `-f`.

### What to build

`docker-compose.rocks.yml` wird ein eigenständiges, vollständiges Compose-File
(kein Override mehr): es enthält postgres, migrate, backend (build), frontend
(build), reverse-proxy (nginx + `nginx.rocks.conf`), certbot, resolver und
acme-dns. `name: jotti` und alle Volume-Namen bleiben identisch, damit die
laufende jotti.rocks-Datenbank und die Zertifikate erhalten bleiben. Die
rocks-\* Makefile-Targets und `rocks-init.sh` werden auf das alleinige rocks-File
umgestellt. `docker-compose.initial-cert.yml` bleibt unverändert (rocks-Cert-Bootstrap).

### Acceptance criteria

- [x] `docker compose -f docker-compose.rocks.yml config` ist valide und enthält
      keine Referenz mehr auf `docker-compose.prod.yml`.
- [x] `make rocks-up` startet Landing (`jotti.rocks`), Demo (`demo.jotti.rocks`)
      und acme-dns (`auth.jotti.rocks`) wie bisher.
- [x] Projektname bleibt `jotti`; `jotti_postgres-data`, `letsencrypt` und
      `certbot-challenges` werden weiterverwendet (kein Datenverlust, kein
      Neu-Ausstellen der Zertifikate).
- [x] `make rocks-down`, `rocks-logs`, `rocks-reset-db`, `rocks-reset-and-seed`
      funktionieren mit dem neuen File.
- [x] `make check` bleibt grün.

---

## Phase 2: Self-Hoster-VPS auf Caddy Auto-TLS + gepinnte Images

### Context

- `reverse-proxy/caddyfile.go:41-117` — `renderCaddyfile` + `(jotti_proxy)`-Snippet
  wiederverwenden; neuen Public-Site-Renderer ergänzen.
- `reverse-proxy/main.go:43-96` — `loadConfig`/`main` um Public-Mode erweitern
  (z. B. `JOTTI_DOMAIN` gesetzt ⇒ Public-Mode statt LAN-Mode: keine
  acme-dns-Registrierung, keine Statusseite, Public-Caddyfile).
- `reverse-proxy/Dockerfile:13-16` — Rate-Limit-Modul in den xcaddy-Build aufnehmen.
- `reverse-proxy/caddyfile_test.go:1-30` — Testmuster für den neuen Renderer.
- `docker-compose.release.yml:54,70,102,113` — Vorlage für gepinnte GHCR-Images.
- `docker-compose.prod.yml:44-156` — Ziel des Umbaus.
- `scripts/prod-init.sh:1-186` — Neufassung (Caddy statt certbot).
- `.env.example:13-16` — neue Keys ergänzen.
- `Makefile:153-164` — prod-up/prod-init/prod-down/prod-logs anpassen.

### What to build

Ein Public-Mode im `jotti-reverse-proxy`-Image: ist `JOTTI_DOMAIN` gesetzt,
rendert der Entrypoint einen Public-Caddyfile (eine Site `JOTTI_DOMAIN`, globale
`email`-Direktive aus `LETSENCRYPT_EMAIL`, automatisches HTTP-01-Zertifikat,
optionaler `www.`-Redirect, `PROXY_LE_STAGING` für Tests). Der Public-Mode
importiert das bestehende `(jotti_proxy)`-Snippet (Security-Header, CSP,
`handle_path /api/*`, SPA-Proxy), setzt aber HSTS auf den stärkeren Public-Wert
und ergänzt Rate-Limiting (Parität zu nginx: 10r/s, burst 20) per xcaddy-Modul.
LAN-Mode (lokal/release) bleibt unverändert.

`docker-compose.prod.yml` wird neu gefasst: postgres + migrate/backend/frontend/
reverse-proxy als gepinnte `ghcr.io/nicograef/jotti-*:${JOTTI_VERSION:-latest}`,
reverse-proxy im Public-Mode mit `JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL`/`PROXY_LE_STAGING`,
Ports 80/443 (+udp), Caddy-Daten-Volume. Kein certbot, kein initial-cert mehr.
`.env.example` bekommt `JOTTI_DOMAIN`, `LETSENCRYPT_EMAIL`, `JOTTI_VERSION`.
`scripts/prod-init.sh` wird neu geschrieben: liest Domain/E-Mail aus `.env`
(keine Hartkodierung), prüft Docker/Compose/`.env`/DNS (zeigt Domain auf diesen
Server?), `docker compose -f docker-compose.prod.yml pull`, `up -d`, pollt
`/api/health` (statt fester sleeps), verifiziert HTTPS und gibt eine Zusammenfassung
aus. `make prod-up` zieht Images und startet ohne `--build`.

### Acceptance criteria

- [x] Mit nur gesetztem `JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL` in `.env` startet
      `make prod-init` den Stack und liefert ein gültiges Zertifikat (grünes
      Schloss), ohne dass eine getrackte Datei editiert wurde. (Public-Mode +
      `.env`-getriebenes `prod-init.sh` umgesetzt; echte Ausstellung ist deploy-zeitig.)
- [x] HTTP leitet auf HTTPS um; `/` liefert die SPA, `/api/health` antwortet `200`.
      (Caddy-Public-Site leitet HTTP automatisch um; `(jotti_proxy)` proxyt SPA +
      `/api/`; `prod-init.sh` pollt `/api/health`.)
- [x] Security-Header und CSP sind identisch zur bisherigen prod-Konfiguration;
      HSTS ist `max-age=63072000; includeSubDomains; preload`; `/api/` ist auf
      10r/s (burst 20) rate-limitiert.
- [x] `docker-compose.prod.yml` enthält kein `build:` und keinen certbot-Service;
      Images sind über `${JOTTI_VERSION}` gepinnt.
- [x] `prod-init.sh` enthält keine hartkodierte Domain/E-Mail; fehlende
      `JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL` brechen mit klarer Meldung ab.
- [x] Renderer-Unit-Tests für den Public-Mode (analog `caddyfile_test.go`) sind
      grün; `caddy validate` der gerenderten Config besteht (alle drei Varianten:
      default, www-Redirect, Staging — gegen das xcaddy-Image verifiziert).
- [x] `make check` inkl. `check-local-proxy` bleibt grün; LAN-/release-Pfad
      unverändert lauffähig.

---

## Phase 3: Backup & Restore

### Context

- `cmd/starter/core/backup.go:18-38` — `ShouldBackup`/`DumpsToDelete`
  (Rotations-Logik als Vorbild).
- `docker-compose.prod.yml` (nach Phase 2) — Service `postgres` für `pg_dump`/`psql`
  per `docker compose exec`.
- `Makefile:153-164` — prod-Targets, hier kommen Backup/Restore dazu.
- `docs/betrieb/leitfaden-betreiber.md` — 10-Jahre-Aufbewahrung (Begründung).

### What to build

`scripts/prod-backup.sh`: zieht `pg_dump` via `docker compose -f
docker-compose.prod.yml exec -T postgres` in eine zeitgestempelte Datei
(`jotti-YYYYMMDD-HHMMSS.sql[.gz]`) in ein konfigurierbares Host-Verzeichnis
(`BACKUP_DIR`, Default z. B. `./backups`) und rotiert auf die letzten N
(`BACKUP_KEEP`, Default z. B. 14). `scripts/prod-restore.sh`: listet vorhandene
Dumps, nimmt einen ausgewählten/zuletzten, fragt eine destruktive Bestätigung ab
und spielt ihn per `psql` zurück. `make prod-backup` / `make prod-restore`.
systemd-Timer- und cron-Snippet (im Repo unter `packaging/` oder dokumentiert)
für einen täglichen Dump, plus Hinweis, Backups vom Server wegzukopieren.

### Acceptance criteria

- [x] `make prod-backup` erzeugt einen zeitgestempelten Dump im `BACKUP_DIR` und
      hält nur die neuesten `BACKUP_KEEP` (ältere werden rotiert). (Rotation gegen
      17 Fixtures verifiziert: behält die neuesten 14, rotiert die 3 ältesten,
      Fremddateien unangetastet.)
- [x] `BACKUP_KEEP<=0` löscht defensiv nichts (kein versehentliches Leeren).
- [x] `make prod-restore` stellt einen Dump nach destruktiver Bestätigung wieder
      her; ein Backup→Restore-Round-Trip auf einer Testinstanz erhält die Daten.
      (Pipeline `pg_dump --clean --if-exists | gzip` → `gzip -dc | psql
  ON_ERROR_STOP` gegen ein wegwerfbares postgres:17.8 verifiziert: der
      gesicherte Stand wird exakt wiederhergestellt; echte prod-Instanz ist
      deploy-zeitig.)
- [x] systemd-Timer/cron-Snippet ist dokumentiert und idempotent installierbar.
- [x] Scripts sind `set -euo pipefail`, ändern nichts ohne Bestätigung und passen
      zum Stil von `scripts/prod-init.sh`.

---

## Phase 4: Sichere Updates (Pre-Update-Backup + Health-Verify + Rollback)

### Context

- `cmd/starter/main.go:100-131` — Windows-Ablauf (Pre-Update-Backup → up →
  waitForHealth) als Vorbild.
- `cmd/starter/core/update.go:32-69` — `IsNewerVersion`/`parseSemver`
  (Downgrade-Guard-Vorbild).
- `scripts/prod-backup.sh` (Phase 3) — wird vom Update-Flow aufgerufen.
- `packaging/windows/KURZANLEITUNG.md:76-124` — Update/Restore/Downgrade-UX.
- `Makefile:153-164` — neues `prod-update`-Target.

### What to build

`scripts/prod-update.sh`: vergleicht die laufende mit der neuen `JOTTI_VERSION`,
verweigert Downgrades (klare Meldung), zieht zwingend ein Pre-Update-Backup
(ruft `prod-backup.sh`), `docker compose pull`, `up -d` (führt Migrationen aus),
pollt `/api/health`. Bei ungesundem Stack: klare, geführte Restore-Anleitung
(bzw. automatische Wiederherstellung des eben gezogenen Dumps) und Abbruch mit
Fehlercode. `make prod-update`. Update-Flow im Hosting-Leitfaden dokumentiert
(Version in `.env` bumpen → `make prod-update`), inkl. Downgrade-Warnung wie bei
Windows.

### Acceptance criteria

- [x] `make prod-update` zieht vor jedem Versionswechsel automatisch ein Backup,
      bevor Migrationen laufen. (Ruft `prod-backup.sh` in Schritt 3, vor `pull`/`up`.)
- [x] Ein Update über zwei Versionen landet gesund (`/api/health` = `200`); Daten
      bleiben erhalten. (Health-Gate über `jotti-backend`-Container + HTTPS-`/api/health`-Check;
      echte Mehrversions-Migration ist deploy-zeitig.)
- [x] Ein simuliert fehlschlagendes Update (z. B. unerreichbares Health) bricht
      mit klarer Restore-Anleitung/Rollback ab, ohne Datenverlust. (Unhealthy oder
      `up`-Fehler triggert `rollback_guidance` mit benanntem Pre-Update-Dump + Exit 1.)
- [x] Downgrade (`JOTTI_VERSION` kleiner als laufend) wird verweigert. (Semver-Vergleich
      `is_downgrade` gegen den Image-Tag des laufenden Containers; 12 Fälle unit-getestet,
      inkl. numerisch 0.2.0<0.10.0 und Nicht-Semver wie `latest`.)
- [x] Der Update-Weg ist im Leitfaden dokumentiert (Parität zur Windows-Kurzanleitung).
      (Abschnitt „jotti aktualisieren" in `leitfaden-hosting.md` Weg B inkl. Downgrade-Warnung.)

---

## Phase 5: Server-Härtung & Leitfaden-Neufassung

### Context

- `docs/betrieb/leitfaden-hosting.md:168-231` — Weg B, wird neu gefasst.
- `scripts/prod-init.sh` (Phase 2) — Stil-/Helper-Vorbild für das Härtungs-Script.
- `docs/betrieb/leitfaden-betreiber.md` — Querverweis (Backups/Aufbewahrung).

### What to build

`scripts/prod-harden.sh` (opt-in, idempotent): konfiguriert ufw (SSH zuerst
freigeben, dann 80/443 erlauben, Rest sperren; Postgres nie exponiert), optional
fail2ban (sshd-Jail), Hinweis auf unattended-upgrades. Schützt vor SSH-Aussperrung
(SSH-Regel vor `ufw enable`). Hosting-Leitfaden Weg B wird neu gefasst: `.env`-getriebener
Ablauf (`JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL`/`JOTTI_VERSION`), Caddy Auto-TLS (keine
certbot-Schritte), Update-Flow, Backup/Restore + cron, Härtungs-Abschnitt. Die
veralteten Anweisungen "Domain/E-Mail in prod-init.sh anpassen" und nginx-`<domain>`-Edit
werden entfernt. Doku-Stil minimal halten (keine em-dashes/liberales Bold).

### Acceptance criteria

- [ ] `scripts/prod-harden.sh` ist idempotent und sperrt SSH nicht aus
      (SSH-Freigabe vor `ufw enable`).
- [ ] Härtung ist opt-in (nicht Teil von `prod-init.sh`).
- [ ] `leitfaden-hosting.md` Weg B beschreibt den realen `.env`-/Caddy-/Update-/
      Backup-/Härtungs-Ablauf; keine Verweise mehr auf Datei-Handedits.
- [ ] `make website-check`/Doku-Linting (falls zutreffend) bleibt grün; keine
      toten Links.
- [ ] Doku-Stil entspricht der House-Style-Vorgabe (minimal, run-in-Labels ok).

```

```
