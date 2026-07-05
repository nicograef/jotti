# Plan: Ops-Härtung Runde 1

> Source PRD: docs/prds/prd-ops-haertung.md

## Goal

Das Backup-System nachweislich verlässlich machen (F1 bis F4 aus
`docs/plans/findings-phase4-backup-test.md`), die CI-Lücken schließen
(Windows-Module, website-Job, shellcheck, Dependabot) und den Betrieb auf
allen drei Wegen härten (Self-Hosting, Windows-Kasse, Demo-VPS):
Fail-fast-Versionspinning, Log-Rotation, non-root Images, Healthchecks,
Start-Retry, Versionsauskunft, nginx-Reload und Windows-Host-Spiegel.

## Architectural decisions

Durable Entscheidungen, die für alle Phasen gelten:

- **`read_env`-Fix**: das Muster `{ grep -E "^${key}=" .env 2>/dev/null || true; } | tail …`
  wird identisch in alle vier prod-Skripte kopiert. Die Skripte bleiben
  bewusst eigenständige Kopien (einzeln kopierbar), kein gemeinsames Sourcing.
- **Compose-Parametrisierung**: Backup- und Restore-Skript lesen die
  Compose-Datei aus der Umgebungsvariablen `COMPOSE_FILE` (Docker-üblicher
  Name), Default `docker-compose.prod.yml`. Die Skripte übergeben sie immer
  explizit per `-f`. Die Restore-Bestätigung nennt Compose-Datei und Dump,
  damit ein aus der Shell geerbtes `COMPOSE_FILE` auffällt.
- **Verify-Skript**: `scripts/prod-backup-verify.sh` plus Make-Target
  `prod-backup-verify`. Wegwerf-Postgres per `docker run --rm`; die
  Postgres-Version wird aus der Compose-Datei (`COMPOSE_FILE`) gegrept,
  damit sie nie vom Stack abweicht.
- **Erfolgs-Ping**: optionale Variable `BACKUP_PING_URL` (Umgebung oder
  `.env`), curl mit kurzem Timeout nur nach erfolgreichem, integritätsgeprüftem
  Dump. Fehlschlag des Pings ist eine Warnung, nie ein Skriptfehler
  (Dead-Man-Switch: Alarm gibt der externe Dienst beim Ausbleiben).
- **Log-Rotation**: pro Compose-Datei ein YAML-Anchor
  (`x-logging: &default-logging`, json-file, max-size 5m, max-file 3), je
  dauerhaft laufendem Service referenziert. Gilt für alle fünf Stacks
  (dev, local, release, prod, rocks); `docker-compose.initial-cert.yml` ist
  ein Einmal-Werkzeug und bleibt außen vor.
- **Versionspinning**: `prod-init` und `prod-update` validieren
  `JOTTI_VERSION` als `vMAJOR.MINOR.PATCH` (das vorhandene `parse_semver`
  wird wiederverwendet bzw. nach prod-init kopiert) und brechen sonst mit
  einer Meldung ab, die den Fix nennt. `.env.example` liefert den Schlüssel
  leer mit aufforderndem Kommentar. Der Compose-Default `:-latest` bleibt
  als (über den geführten Weg unerreichbarer) Fallback.
- **Backend-Version**: `var version = "dev"` in `backend/main.go`, per
  ldflags gesetzt (Vorbild `Makefile:110`). Das Backend-Dockerfile bekommt
  `ARG VERSION=dev`, der Release-Workflow übergibt `--build-arg`. Der
  Health-Endpoint liefert ein zusätzliches Feld `version`; Response bleibt
  abwärtskompatibel. rocks- und Dev-Builds melden `dev` (akzeptiert).
- **Windows-Host-Spiegel**: Ziel `<stateDir>\backups` (unter Windows
  `%PROGRAMDATA%\jotti\backups`), Kopie per `docker cp` aus dem
  postgres-Container, Rotation mit demselben `keptBackups`. Entscheidungslogik
  als reine Funktion in `windows/starter/core`, Seiteneffekte im Wrapper.
  Spiegel-Fehlschlag ist ein Hinweis, kein Startabbruch.
- **CI**: Windows-Job filtert auf `windows/**` und arbeitet in
  `windows/relay` und `windows/starter`. Neuer shellcheck-Job mit Pfadfilter
  `scripts/**` (shellcheck ist auf ubuntu-latest vorinstalliert).
  `dependabot.yml` monatlich und gruppiert: github-actions (`/`), gomod
  (backend, resolver, reverse-proxy, windows/relay, windows/starter), npm
  (frontend, website), docker (backend, frontend, database/migrate,
  reverse-proxy, website). Der postgres-Pin in den Compose-Dateien bleibt
  manuell gepflegt (bewusst synchron mit `ci.yml`).
- **Commit-Modus**: ein Commit pro Phase (Conventional Commits), Message wird
  vorgeschlagen, committet wird erst nach Freigabe (No-Auto-Commit).
- **Verifikation**: `make verify` deckt Go und Frontend; Shell-Skripte und
  Compose-Dateien werden pro Phase über die genannten Kommandos geprüft
  (shellcheck, `docker compose config`, manuelle Läufe). E2E-Gate für die
  Skript- und Image-Phasen ist der Release-Workflow als Dry-Run
  (`workflow_dispatch`).

## Inventory

- `scripts/prod-backup.sh:23,42-45,88-122` — hartes `COMPOSE_PROD`, buggy
  `read_env`, Dump ohne Integritätsprüfung, Rotation.
- `scripts/prod-restore.sh:23,42-45,110-136` — buggy `read_env`, alle
  Container-Operationen über die prod-Compose, kein Proxy-Recreate (F4).
- `scripts/prod-update.sh:43-46,52-57,103-104` — buggy `read_env`,
  `parse_semver` vorhanden, leeres `JOTTI_VERSION` fällt auf `latest` zurück.
- `scripts/prod-init.sh:36-39,75-85` — buggy `read_env`, validiert
  Domain/E-Mail, aber nicht die Version.
- `docker-compose.prod.yml:44,59,96,112,118-119` — `${JOTTI_VERSION:-latest}`,
  `JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL` ohne Default (F1-Warnungen).
- `docker-compose.rocks.yml:132-178` — stock nginx ohne Reload, certbot-Loop
  als Muster für ein Loop-Kommando.
- `docker-compose.release.yml` / `docker-compose.local.yml` /
  `docker-compose.yml` — dauerhaft laufende Services, nirgends `logging:`.
- `packaging/systemd/jotti-backup.service:9,20-23` — Header erwähnt nur
  `WorkingDirectory`, `ExecStart` fehlt; `Environment=` auskommentiert (F3).
- `packaging/cron/jotti-backup.cron:12` — gleiche Pfad-Anpassung nötig.
- `.env.example:19-27` — `JOTTI_VERSION=latest`, Backup-Block.
- `.github/workflows/ci.yml:37-41,183-225,263-286` — Filter `cmd/**` und
  Referenz auf gelöschtes `scripts/check-website.sh`; cmd-ci arbeitet in
  `./cmd/…`; website-ci installiert in `./frontend`.
- `.github/workflows/release.yml:69-91,98-123` — Image-Build ohne
  Version-Build-Arg; Smoke-Test endet nach dem Health-Check, `down -v` mit
  `always()`.
- `backend/main.go:29-45` — ein `db.Ping()`, sofortiges Fatal.
- `backend/app/app.go:48-49`, `backend/api/health/health.go` — Health-Wiring
  und `HealthResponse` (status, timestamp).
- `backend/Dockerfile:13`, `database/migrate/Dockerfile` — Build ohne
  ldflags/ARG, beide Images laufen als root.
- `frontend/Dockerfile`, `website/Dockerfile` — Runtime jeweils
  nginx:1.27-alpine (busybox-wget vorhanden, Healthcheck möglich);
  reverse-proxy-Image basiert auf caddy:2.11.4 (alpine).
- `windows/starter/backup.go:49-143`, `windows/starter/main.go:100-106,137-147`,
  `windows/starter/core/backup.go` — Pre-Update-Dump nur ins Volume,
  `stateDir` als Ankerpunkt, `ShouldBackup`/`DumpsToDelete` samt Tests als
  Vorbild für die Spiegel-Logik.
- `packaging/windows/KURZANLEITUNG.md:101-104` — bestehender Backup-Absatz,
  Andockpunkt für das manuelle Sichern.
- `Makefile:109-113,153-176,265-269,303-304` — ldflags-Vorbild, prod-Targets,
  check-relay/check-starter, website-check.
- `docs/leitfaden/aktualisieren-backups.md`, `docs/leitfaden/self-hosting.md:34-39`
  — Backup-/Update-Doku; self-hosting zeigt bereits ein gepinntes Beispiel.
- `docs/plans/findings-phase4-backup-test.md:101-130,226-233` — verifizierter
  F2-Fix-Diff samt Repro-Snippets und rocks-sichere Restore-Sequenz
  (Zeilenreferenzen dort gelten für Commit 21d8c4b).

## Resolved decisions

- `COMPOSE_FILE` als Variablenname für die F4-Parametrisierung (Docker-üblich);
  Absicherung über die Compose-Datei-Anzeige in der Restore-Bestätigung.
- Log-Rotation in allen fünf Stacks inklusive dev und local; initial-cert nicht.
- Verify-Skript heißt `prod-backup-verify` (Skript und Make-Target).
- Phasenschnitt mit 11 Phasen bestätigt.
- Dependabot überwacht Docker-Basis-Images über die fünf Dockerfiles; der
  postgres-Pin in Compose/CI bleibt manuell (Drift zu `ci.yml` vermeiden).
- nginx-Reload-Loop als command-Override im rocks-Compose (Interims-Fix laut
  PRD; keine neuen Dateien, Caddy-Migration ist eine spätere PRD).

## Open questions / Risks

- Parallel läuft der Sprachkonsistenz-Umbau (`docs/plans/plan-sprachkonsistenz.md`,
  Backend-Bezeichner). Überschneidung ist minimal (dieser Plan berührt im Go-Code
  nur `backend/main.go`, `backend/api/health`, `backend/app`-Wiring und
  `windows/`); Phase 10 erst starten, wenn der Arbeitsbaum im Backend sauber ist.
- shellcheck kann in den übrigen Skripten (`rocks-init.sh`, `reset-and-seed.sh`,
  `init-env.sh`, …) Findings aufdecken. In Phase 7 werden nur echte Fehler
  gefixt, keine Umbauten; unlösbare Findings werden gezielt per Direktive
  dokumentiert.
- Non-root-Images: Backend lauscht auf 3000 (>1024) und schreibt nicht ins
  Dateisystem; Restrisiko sind versteckte Schreibpfade, die der
  Release-Smoke-Test aufdecken würde.
- Der rocks-Reload-Loop und die rocks-Healthchecks sind erst beim nächsten
  VPS-Deploy live verifizierbar; lokal bleibt `docker compose config` plus
  Review der Loop-Semantik (Signalverhalten von `nginx -g "daemon off;"` als
  Vordergrundprozess).

---

## Phase 1: F2/F1/F3 — stillen Backup-Totalausfall beseitigen

**User stories**: 1, 2, 3, 9

### Context

- `scripts/prod-backup.sh:42-45`, `scripts/prod-restore.sh:42-45`,
  `scripts/prod-update.sh:43-46`, `scripts/prod-init.sh:36-39` — identischer
  `read_env`-Helfer mit grep-Pipefail-Bug.
- `docker-compose.prod.yml:118-119` — `JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL`
  ohne Default (F1-Warnungen gegen fremde Stacks).
- `packaging/systemd/jotti-backup.service:9,20-23`,
  `packaging/cron/jotti-backup.cron:12` — unvollständige Pfad-Hinweise (F3).
- `docs/plans/findings-phase4-backup-test.md:101-130` — verifizierter Fix-Diff
  und Repro-Snippets.

### What to build

Den `read_env`-Helfer in allen vier prod-Skripten identisch fixen
(grep-No-Match wird geschluckt, echte Fehler nicht), damit fehlende
`.env`-Schlüssel auf die eingebauten Defaults fallen statt das Skript unter
`set -e`/`pipefail` lautlos zu töten. In `docker-compose.prod.yml` bekommen
die nur dort genutzten Variablen leere Defaults (`${JOTTI_DOMAIN:-}`,
`${LETSENCRYPT_EMAIL:-}`), damit Skriptläufe gegen andere Stacks keine
Warnungen ausgeben. Die systemd- und cron-Vorlagen nennen in ihren
Kommentaren alle anzupassenden Pfade inklusive `ExecStart` und stellen klar,
dass die `Environment=`-Zeilen nach dem F2-Fix optional sind.

### Acceptance criteria

- [ ] Das Repro-Snippet aus den Findings (leere `.env`, `read_env BACKUP_DIR`)
      erreicht den `./backups`-Default mit Exit 0; der Fix steht wortgleich in
      allen vier Skripten (`grep -c` über `scripts/prod-*.sh`).
- [ ] `bash -x scripts/prod-backup.sh` mit einer `.env` ohne
      `BACKUP_DIR`/`BACKUP_KEEP` läuft bis zum Dump-Schritt weiter statt
      lautlos mit Exit 1 zu enden (der E2E-Beweis folgt in Phase 5).
- [ ] `docker compose -f docker-compose.prod.yml config` mit einer `.env` ohne
      Domain/E-Mail gibt keine Variablen-Warnungen mehr aus.
- [ ] shellcheck über die vier Skripte ist grün.
- [ ] Beide Vorlagen erwähnen `ExecStart` bzw. den cron-Pfad als anzupassend.

---

## Phase 2: F4 — Backup und Restore auf die Compose-Datei parametrisieren

**User stories**: 11

### Context

- `scripts/prod-backup.sh:23` — hartes `COMPOSE_PROD`.
- `scripts/prod-restore.sh:23,110-136` — `up`/`stop`/`up` über die
  prod-Compose, kein Reverse-Proxy-Recreate.
- `docs/plans/findings-phase4-backup-test.md:226-233` — getestete rocks-sichere
  Restore-Sequenz.
- `Makefile:169-173`, `docs/leitfaden/aktualisieren-backups.md` — Aufrufer
  und Doku.

### What to build

Beide Skripte lesen die Compose-Datei aus `COMPOSE_FILE` (Default
`docker-compose.prod.yml`) und verwenden sie in jedem `docker compose`-Aufruf.
Das Restore-Skript erzeugt nach dem abschließenden `up -d` den Reverse-Proxy
neu (`up -d --no-deps --force-recreate reverse-proxy`), was die nginx-502-Falle
des rocks-Stacks entschärft und für Caddy-Stacks ein harmloser No-op ist. Die
interaktive Bestätigung bleibt und nennt neben dem Dump auch die verwendete
Compose-Datei. `prod-update.sh` ruft `prod-backup.sh` unverändert auf (Default
greift). Die Restore-Doku in `aktualisieren-backups.md` erwähnt die
Parametrisierung.

### Acceptance criteria

- [ ] Kein `docker compose`-Aufruf in Backup/Restore referenziert die
      prod-Compose noch hart; `COMPOSE_FILE=docker-compose.local.yml` gegen
      einen laufenden Local-Stack erzeugt einen Dump.
- [ ] Restore endet mit dem Force-Recreate des Reverse-Proxy.
- [ ] Die Bestätigungsabfrage zeigt Compose-Datei und Dump; Abbruch ohne
      `yes` verändert nichts.
- [ ] shellcheck grün; Doku erwähnt `COMPOSE_FILE`.

---

## Phase 3: Dump-Integritätsprüfung und Erfolgs-Ping

**User stories**: 4, 5

### Context

- `scripts/prod-backup.sh:88-103` — Dump in `.partial`, `mv` ohne Prüfung.
- `.env.example:23-27` — Backup-Konfigurationsblock.
- `docs/leitfaden/aktualisieren-backups.md:18-33` — Backup-Abschnitt.

### What to build

Nach dem Dump prüft das Backup-Skript die `.partial`-Datei mit `gzip -t`;
schlägt die Prüfung fehl, wird die Datei verworfen (bestehender trap) und das
Skript endet mit Fehlerstatus und klarer Meldung — eine korrupte Datei zählt
nie als Backup. Danach, und nur nach vollem Erfolg, ruft das Skript eine
optional gesetzte `BACKUP_PING_URL` per curl auf (kurzer Timeout, wenige
Sekunden); ein Fehlschlag des Pings ist eine Warnung, kein Skriptfehler.
`.env.example` dokumentiert den neuen Schlüssel (leer). Der Leitfaden erklärt
den Dead-Man-Switch-Gedanken (Alarm beim Ausbleiben, Dienst frei wählbar).

### Acceptance criteria

- [ ] Ein absichtlich beschädigter gzip-Strom führt zu Exit ungleich 0 und
      hinterlässt keine `.sql.gz` im Backup-Verzeichnis.
- [ ] Mit `BACKUP_PING_URL` auf einen lokalen Test-Endpunkt kommt nach einem
      erfolgreichen Lauf genau ein Request an; ohne die Variable keiner.
- [ ] Eine unerreichbare Ping-URL erzeugt eine Warnung, der Exit-Code bleibt 0.
- [ ] shellcheck grün; `.env.example` und Leitfaden ergänzt.

---

## Phase 4: Verify-Skript prod-backup-verify

**User stories**: 6, 10

### Context

- `scripts/prod-restore.sh:119-133` — decompress-plus-psql-Kern mit
  `ON_ERROR_STOP` als Vorbild.
- `scripts/prod-backup.sh:29-52` — Helfer-/Stilvorlage (info/warn/fatal,
  Projektwurzel).
- `docker-compose.prod.yml:21` — gepinnte Postgres-Version des Stacks.
- `Makefile:169-176` — Einordnung neben prod-backup/prod-restore.

### What to build

Neues `scripts/prod-backup-verify.sh` plus Make-Target `prod-backup-verify`:
nimmt einen Dump als Argument oder den neuesten aus `BACKUP_DIR`, startet
einen Wegwerf-Postgres-Container (`docker run --rm`, Version aus der
Compose-Datei gegrept, kein Netz und keine Volumes des laufenden Stacks),
spielt den Dump per psql mit `ON_ERROR_STOP` ein und prüft, dass die
Tabellenanzahl größer null ist. Ausgabe ist eine kurze Zusammenfassung
(Dump, Tabellenzahl, Ergebnis) plus passender Exit-Code. Der Leitfaden
empfiehlt einen gelegentlichen manuellen Lauf und bekommt zusätzlich den
konkreten, kopierbaren Offsite-Kopierbefehl (Story 10, z. B. ein
rsync/scp-Einzeiler für `./backups`).

### Acceptance criteria

- [ ] Lauf gegen einen echten Dump (lokaler Stack, `make prod-backup` mit
      `COMPOSE_FILE`) endet mit Exit 0 und nennt die Tabellenzahl.
- [ ] Ein beschädigter Dump führt zu Exit ungleich 0.
- [ ] Laufender Stack und dessen Volumes bleiben unberührt (`docker ps` und
      `docker volume ls` vorher/nachher identisch); kein Wegwerf-Container
      bleibt zurück.
- [ ] shellcheck grün; Make-Target vorhanden; Leitfaden um Verify-Empfehlung
      und Offsite-Befehl ergänzt.

---

## Phase 5: Release-Smoke-Test fährt Backup plus Verify

**User stories**: 21

### Context

- `.github/workflows/release.yml:98-123` — `make init`, Stack-Start,
  Health-Wait, Logs bei Fehler, `down -v` mit `always()`.
- `docker-compose.release.yml` — Service `postgres` (Container
  `jotti-postgres-local`), migrierte, leere Datenbank nach dem Start.

### What to build

Nach dem bestehenden Health-Check-Step ruft der Workflow das Backup-Skript
mit `COMPOSE_FILE=docker-compose.release.yml` auf und danach
`prod-backup-verify.sh` auf den soeben erzeugten Dump. Beide Steps laufen vor
dem `down -v`; schlägt einer fehl, ist das Release rot. Damit sind
Backup-Pfad und Verify-Pfad bei jedem Release Ende-zu-Ende geprüft; der
destruktive Restore-Pfad bleibt bewusst unautomatisiert (sein psql-Kern ist
mit dem Verify-Pfad identisch).

### Acceptance criteria

- [ ] Ein `workflow_dispatch`-Dry-Run läuft grün durch, inklusive der neuen
      Backup- und Verify-Steps.
- [ ] Die Steps stehen nach dem Health-Check und vor dem `down -v`; `down -v`
      behält `if: always()`.
- [ ] Bricht das Backup ab (kein Dump), schlägt der Workflow fehl statt den
      Verify-Step leer durchzuwinken.

---

## Phase 6: Versionspinning fail-fast

**User stories**: 7, 8

### Context

- `scripts/prod-update.sh:52-57,103-104` — `parse_semver` vorhanden, leeres
  `JOTTI_VERSION` fällt auf `latest` zurück.
- `scripts/prod-init.sh:75-85` — validiert Domain/E-Mail, nicht die Version.
- `.env.example:19-20` — liefert `JOTTI_VERSION=latest` aus.
- `docs/leitfaden/self-hosting.md:34-39` — zeigt bereits ein gepinntes
  Beispiel (`v0.2.0`).

### What to build

`prod-init` und `prod-update` validieren `JOTTI_VERSION` aus der `.env` früh
als Release-Tag im Format `vMAJOR.MINOR.PATCH` (via `parse_semver`; in
prod-init wird der Helfer kopiert, Skripte bleiben eigenständig). Leer,
`latest` oder anderes Format bricht mit einer Meldung ab, die den Fix nennt
(Release-Tag in `.env` eintragen, Verweis auf die GitHub-Releases-Seite).
`.env.example` liefert den Schlüssel leer mit aufforderndem Kommentar aus;
der Compose-Default `:-latest` bleibt bestehen, wird über den geführten Weg
aber nicht mehr erreicht.

### Acceptance criteria

- [ ] `prod-init`/`prod-update` mit `JOTTI_VERSION` leer, `latest` oder `0.3`
      brechen vor jeder Docker-Aktion mit einer Meldung ab, die den Fix nennt;
      mit `v0.3.1` laufen sie bis zum nächsten Prüfschritt weiter.
- [ ] `.env.example` enthält kein `latest` mehr; `make init` erzeugt weiterhin
      eine gültige `.env` (idempotent).
- [ ] Der Release-Smoke-Test bleibt grün (Release-Compose nutzt
      `JOTTI_VERSION` nicht).
- [ ] shellcheck grün.

---

## Phase 7: CI-Reparaturen, shellcheck und Dependabot

**User stories**: 20, 22, 23, 24

### Context

- `.github/workflows/ci.yml:37-41` — Filter `cmd/**` und Referenz auf das
  gelöschte `scripts/check-website.sh`.
- `.github/workflows/ci.yml:183-225` — cmd-ci arbeitet in `./cmd/relay|starter`
  (existiert nicht mehr).
- `.github/workflows/ci.yml:263-286` — website-ci installiert in `./frontend`
  und cached `frontend/pnpm-lock.yaml`.
- `Makefile:265-269,303-304` — check-relay/check-starter und website-check
  als inhaltliche Vorbilder.
- `scripts/` — zehn Shell-Skripte für den shellcheck-Scope.

### What to build

Der Windows-Job filtert auf `windows/**` und arbeitet in den tatsächlichen
Modulverzeichnissen (`windows/relay`, `windows/starter`); Schritte bleiben
Format, Vet, Build, Test. Der website-Job installiert die Abhängigkeiten im
website-Ordner, cached dessen Lockfile und der Pfadfilter verliert die
Referenz auf das gelöschte Skript. Neu: ein shellcheck-Job mit Pfadfilter
`scripts/**`, der alle `scripts/*.sh` prüft (dabei aufgedeckte echte Fehler
in bestehenden Skripten werden gefixt, keine Umbauten). Neu außerdem
`.github/dependabot.yml`: monatlich und gruppiert für GitHub Actions, alle
fünf Go-Module, beide npm-Projekte und die fünf Dockerfiles.

### Acceptance criteria

- [ ] Lokal äquivalent zum Windows-Job: `make check-relay check-starter` grün;
      der Pfadfilter matcht `windows/**`.
- [ ] website-Job: Install und Cache zeigen auf `website/`; kein Verweis auf
      `scripts/check-website.sh` mehr in der Workflow-Datei.
- [ ] `shellcheck scripts/*.sh` lokal grün.
- [ ] `dependabot.yml` deckt alle vier Ökosysteme mit den genannten
      Verzeichnissen ab, Intervall monatlich, Updates gruppiert.
- [ ] Erster CI-Lauf auf dem Branch: alle betroffenen Jobs grün.

---

## Phase 8: Compose-Härtung (Log-Rotation, Healthchecks, nginx-Reload)

**User stories**: 12, 15, 17

### Context

- `docker-compose.yml`, `docker-compose.local.yml`, `docker-compose.release.yml`,
  `docker-compose.prod.yml`, `docker-compose.rocks.yml` — nirgends `logging:`.
- `docker-compose.rocks.yml:132-158` — stock-nginx-Proxy; `:160-178`
  certbot-Loop als Muster für das Loop-Kommando.
- `frontend/Dockerfile`, `website/Dockerfile` — Runtime nginx:1.27-alpine
  (busybox-wget vorhanden); reverse-proxy-Image auf caddy:2.11.4.

### What to build

Jede der fünf Compose-Dateien bekommt einen einheitlichen Logging-Anchor
(json-file, kleine Maximalgröße, wenige Dateien), den jeder dauerhaft
laufende Service referenziert — wirksam auf jedem Host ohne
Daemon-Konfiguration. Frontend, Website und Reverse-Proxy erhalten einfache
HTTP-Healthchecks in den Compose-Dateien (wget gegen den lokalen Port;
Redirect-Antworten gelten als gesund). Der rocks-Reverse-Proxy bekommt per
command-Override einen periodischen `nginx -s reload` (alle 12 Stunden) neben
dem Vordergrund-nginx: übernimmt erneuerte Zertifikate und löst die
Upstream-Adressen neu auf.

### Acceptance criteria

- [ ] `docker compose -f <Datei> config` ist für alle fünf Dateien grün; jeder
      dauerhafte Service referenziert den Logging-Anchor.
- [ ] Dev- oder Local-Stack hochgefahren: `docker inspect` zeigt die
      json-file-Limits, `docker ps` zeigt frontend und reverse-proxy als
      healthy.
- [ ] rocks-Compose: Reload-Loop im command, nginx bleibt PID-1-artig im
      Vordergrund (Signalverhalten geprüft); Verifikation auf dem VPS beim
      nächsten Deploy als dokumentierter Schritt.
- [ ] Release-Smoke-Test (Dry-Run) grün mit Healthchecks und Log-Optionen.

---

## Phase 9: Non-root Backend- und Migrate-Image

**User stories**: 16

### Context

- `backend/Dockerfile:15-20` — Alpine-Runtime als root.
- `database/migrate/Dockerfile` — migrate-Binary plus eingebackene
  Migrationen, läuft als root.
- `docker-compose.prod.yml:78-87` — Healthcheck via wget auf Port 3000
  (>1024, non-root-tauglich).

### What to build

Beide Images legen einen unprivilegierten Benutzer an und setzen `USER`;
keiner der beiden Prozesse braucht Schreibrechte im Dateisystem (der
Migrate-Ordner `/migrations` bleibt lesbar). Verhalten und Ports bleiben
identisch.

### Acceptance criteria

- [ ] Dev-/Local-Stack: migrate läuft erfolgreich durch, Backend wird healthy,
      `docker exec jotti-backend-* id` zeigt uid ungleich 0.
- [ ] Release-Smoke-Test (Dry-Run) grün — deckt beide Images im Zusammenspiel
      mit Healthchecks und Migrationen.

---

## Phase 10: Backend-Start-Retry und Versionsauskunft

**User stories**: 18, 19

### Context

- `backend/main.go:29-45` — ein `db.Ping()`, sofortiges Fatal.
- `backend/api/health/health.go` — `HealthResponse` (status, timestamp),
  Handler-Test daneben.
- `backend/app/app.go:48-49` — Health-Wiring.
- `backend/Dockerfile:13` — Build ohne ldflags; `Makefile:110` als
  ldflags-Vorbild; `.github/workflows/release.yml:69-91` — Image-Build-Loop.

### What to build

Der Verbindungsaufbau zur Datenbank bekommt einen begrenzten Retry
(Größenordnung 30 Sekunden, kurze Abstände, sichtbares Log pro Versuch) statt
des sofortigen Fatals; danach gilt weiterhin: ohne Datenbank kein Start. Die
Retry-Entscheidung liegt in einem kleinen, rein testbaren Helfer (Unit-Test
für Abbruchbedingung und Erfolgsfall, Seiteneffekte gemockt). Die Version
wird per ldflags einkompiliert (`ARG VERSION=dev` im Backend-Dockerfile, vom
Release-Workflow per `--build-arg` mit dem Tag befüllt) und im
Health-Endpoint als zusätzliches Feld `version` ausgegeben; die Response
bleibt abwärtskompatibel. Der bestehende Handler-Test wird um das
Versionsfeld erweitert.

### Acceptance criteria

- [ ] Start ohne erreichbare Datenbank: sichtbare Retry-Logs, Fatal erst nach
      dem Zeitbudget; Datenbank kommt wenige Sekunden verspätet: Start
      erfolgreich (manuell mit pausiertem postgres-Container geprüft).
- [ ] `/health` liefert `version` (Default `dev`); Unit-Tests für
      Retry-Helfer und Handler grün.
- [ ] Docker-Build mit `--build-arg VERSION=v9.9.9`: `/health` meldet
      `v9.9.9`; Release-Workflow übergibt das Build-Arg.
- [ ] `go build ./backend/...` und `make verify` grün.

---

## Phase 11: Windows-Host-Spiegel und manuelles Sichern

**User stories**: 13, 14

### Context

- `windows/starter/backup.go:49-143` — Pre-Update-Dump ins Volume plus
  Rotation, Andockpunkt für den Spiegel.
- `windows/starter/main.go:137-147` — `stateDir` (`%PROGRAMDATA%\jotti`).
- `windows/starter/core/backup.go` — `ShouldBackup`/`DumpsToDelete` samt Tests
  als Vorbild für reine Logik.
- `packaging/windows/KURZANLEITUNG.md:101-104` — bestehender Backup-Absatz.

### What to build

Nach jedem erfolgreichen Pre-Update-Dump kopiert der Starter die Datei
zusätzlich per `docker cp` in `<stateDir>\backups` und rotiert dort mit
derselben Aufbewahrungszahl (`keptBackups`). Die Entscheidungslogik (welche
Datei kopieren, welche Host-Dateien löschen) liegt als reine Funktion im
core-Paket mit Unit-Tests; die Seiteneffekte im dünnen Wrapper in
`backup.go`. Ein Fehlschlag des Spiegels ist ein Hinweis, kein Startabbruch
(der Dump im Volume existiert bereits) — ein `docker compose down -v`
vernichtet damit nicht mehr Daten und Backups zugleich. Die KURZANLEITUNG
bekommt einen kurzen Absatz zum manuellen Sichern nach dem Fest: ein
kopierbarer Befehl, Ergebnis extern kopieren (z. B. USB-Stick), plus Hinweis
auf den gespiegelten Ordner.

### Acceptance criteria

- [ ] core-Unit-Tests für die Spiegel-Logik grün (Kopier- und
      Rotationsentscheidung, keep-Grenzfälle wie bei `DumpsToDelete`).
- [ ] Spiegel-Fehlschlag bricht den Start nicht ab (Code-Pfad liefert Hinweis
      statt Fehler); Linux-Dev-Lauf bleibt inert (Dev-Version sichert nie).
- [ ] KURZANLEITUNG-Absatz mit einem kopierbaren Befehl vorhanden.
- [ ] `make check-starter` und `make verify` grün.
