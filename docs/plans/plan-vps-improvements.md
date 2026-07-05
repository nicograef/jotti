# Plan: VPS-Verbesserungen (jotti.rocks Host)

> Source PRD: n/a (Task-Beschreibung + VPS-Audit vom 2026-07-04)
>
> **Status: ✅ ABGESCHLOSSEN — 2026-07-05.** Alle 5 Phasen durchgeführt und
> verifiziert (28/28 Acceptance-Kriterien, 0 offen). Ergebnis-Zusammenfassung
> direkt unter „Goal"; Backup-Test-Findings in `findings-phase4-backup-test.md`.

## Goal

Den netcup-VPS (Debian, hostet jotti.rocks aus `~/jotti`) gemäß den
Handbook-Vorgaben (`~/handbook`) nachziehen: Komfort-Setup (Dotfiles/Tools),
Sicherheits-Härtung, Disk-/Log-Hygiene, Robustheit (Swap, Kernel-Reboot) —
und einmalig testen, ob das automatisierte jotti-Backup (`prod-backup.sh` +
systemd-Timer) gegen die Live-Demo funktioniert, inklusive Restore.

## Durchführung & Ergebnisse (2026-07-05)

Alle fünf Phasen an einem Tag durchgeführt und verifiziert (28/28 Kriterien,
Details je Phase in den `Acceptance criteria`-Abschnitten unten).

| Phase | Ergebnis |
|---|---|
| 1 — Dotfiles & Ops-Tools | ✅ `install.sh` gelaufen (Aliases, git-Defaults, `gh`), git-Identität, alle Tools (jq/htop/tmux/rsync/ncdu/tree/unzip), `EDITOR=vim`. |
| 2 — Updates & Härtung | ✅ `unattended-upgrades` aktiv, `PermitRootLogin no`, fail2ban `bantime 3600`. Passwort-Login als `nico` bewusst **unverändert** (weiterhin `yes`). |
| 3 — Disk & Logs | ✅ journald 500 MB-Cap (26 MB Ist), `daemon.json` Log-Rotation angelegt, Build-Cache 18,85 GB → 284 MB, stale `jotti-proxy_postgres-data` + 42 anonyme Volumes entfernt. |
| 4 — Backup-Test | ✅ Dump + destruktiver Restore-Roundtrip gegen die Live-Demo bestanden (Marker weg, 17/17 Tabellen == Referenz, kein 502). Timer-Test + Cleanup ok. **4 Findings** (s. u. + `findings-phase4-backup-test.md`). |
| 5 — Wartungsfenster | ✅ 2 GB Swap (persistent via fstab), `swappiness=10`, Kernel `6.1.0-50` (Reboot 11:44), Log-Opts `10m/3` auf allen Containern, Stack nach Reboot healthy, Seiten/TLS ok, `ufw/fail2ban/docker` active, 0 Boot-Fehler. |

**Wichtigster Befund (Phase 4, F2):** `prod-backup.sh` **stirbt lautlos** ohne
`BACKUP_DIR`/`BACKUP_KEEP` in `.env` — der ausgelieferte systemd-Timer produziert
dadurch **null Dumps**. Ein-Zeilen-Fix + Volldetail in
`findings-phase4-backup-test.md`.

**Offen (bewusst außerhalb dieses Plans):** die jotti-Repo-Fixes F1–F4 (Code-
Änderungen, fürs Dev-Setup vorgesehen). Kein dauerhaftes DB-Backup installiert
(Demo/Seed-Daten); die Test-Dumps wurden nach Abschluss entfernt.

## Architectural decisions

- **Passwort-Login bleibt**: `PasswordAuthentication yes` für User `nico`
  wird NICHT angefasst (explizite Entscheidung). Der defekte
  `~/.ssh/authorized_keys`-Ordner (Verzeichnis statt Datei) bleibt out of scope.
- **Root-Login aus**: `PermitRootLogin no` wird explizit gesetzt
  (aktuell nur auskommentiert → effektiv `prohibit-password`).
- **Kein dauerhaftes DB-Backup**: Demo-Instanz mit Seed-Daten. Der
  systemd-Timer wird nur zum Test installiert und danach wieder entfernt.
- **Backup-Test gegen die Live-Demo**: Dump via `make prod-backup`
  (originalgetreu), Restore aber über eine rocks-sichere Prozedur mit
  `docker-compose.rocks.yml` — `prod-restore.sh` ist gegen den Live-Stack
  NICHT sicher (siehe Risiken). Kurze Demo-Downtime ist akzeptiert.
- **Keine Code-Änderungen am jotti-Repo** in diesem Plan. Erkenntnisse aus
  dem Backup-Test werden als Findings dokumentiert (mögliche Follow-ups).
- **Downtime-Bündelung**: Alles, was Container-Neustart/Reboot braucht
  (Kernel, `daemon.json`, Container-Recreate), läuft gesammelt in Phase 5.

## Inventory

Live-Stack:

- `~/jotti/docker-compose.rocks.yml` — laufender Stack (Compose-Projekt
  `jotti`, 8 Container). Services: `postgres`, `migrate`, `backend`,
  `frontend`, `website`, `reverse-proxy`, `certbot`, `resolver`, `acme-dns`
  (`docker-compose.rocks.yml:18-212`). DB-Name ist `jotti`
  (`docker-compose.rocks.yml:25`), User aus `.env` (`POSTGRES_USER`).
- `~/jotti/Makefile:169-170` — `prod-backup` ruft `./scripts/prod-backup.sh`.
- `~/jotti/Makefile:185-188` — `rocks-up` (up -d --build + force-recreate reverse-proxy).
- `~/jotti/Makefile:195-200` — `rocks-reset-db` / `rocks-reset-and-seed`
  (Notfall-Wiederherstellung der Demo-Seed-Daten).
- `~/jotti/.gitignore:37` — `/backups/` ist gitignored (Dumps landen sicher).

Backup-Maschinerie (Testgegenstand):

- `~/jotti/scripts/prod-backup.sh:23-24` — hartkodiert
  `COMPOSE_PROD="docker-compose.prod.yml"`, `PG_SERVICE="postgres"`.
- `~/jotti/scripts/prod-backup.sh:96-101` — Dump via
  `docker compose -f docker-compose.prod.yml exec -T postgres pg_dump --clean
  --if-exists -U "$POSTGRES_USER" -d jotti | gzip`. Compose findet Container
  über Labels (Projekt `jotti` + Service `postgres`) — beide Stacks matchen,
  daher Hypothese: der Dump trifft den Live-Postgres.
- `~/jotti/scripts/prod-backup.sh:111-122` — Rotation (behalte neueste
  `BACKUP_KEEP`, Default 14; Default-`BACKUP_DIR` ist `./backups`).
- `~/jotti/scripts/prod-restore.sh:111,114,136` — **rocks-unsicher**:
  `up -d --wait postgres` und finales `up -d` mit der Prod-Compose-Datei
  würden die laufenden rocks-Container nach Prod-Konfiguration umbauen.
- `~/jotti/packaging/systemd/jotti-backup.service` — oneshot,
  `WorkingDirectory=/opt/jotti` (muss für den Test auf `/home/nico/jotti`
  angepasst werden).
- `~/jotti/packaging/systemd/jotti-backup.timer` — `OnCalendar=*-*-* 03:30:00`,
  `Persistent=true`.

Handbook-Referenzen:

- `~/handbook/install.sh:2` → `scripts/install-dotfiles.sh` — symlinkt
  `templates/.bash_aliases`, setzt git-Defaults (`init.defaultBranch main`,
  `pull.rebase true`, `push.autoSetupRemote true`, `rerere.enabled true`),
  installiert `gh` nach `~/.local/bin` (`install-dotfiles.sh:25-68`).
- `~/handbook/scripts/setup-server.sh:132-138` — fail2ban-`jail.local`-Vorlage
  mit `maxretry = 5`, `bantime = 3600`.
- `~/handbook/guides/docker-setup.md` (Abschnitt „Log rotation") —
  `/etc/docker/daemon.json` mit `max-size: 10m`, `max-file: 3`.
- `~/handbook/guides/provision-server.md` — SSH-Härtung, UFW, fail2ban.

Ist-Zustand (Audit 2026-07-04, alle Befunde einzeln nachverifiziert):

- `/etc/ssh/sshd_config:123` — `# PermitRootLogin yes` (nur auskommentiert).
- unattended-upgrades: nicht installiert; Updates liefen bisher nur manuell.
- Kernel: 6.1.0-44 läuft, 6.1.0-50 installiert (Uptime 112 Tage) → Reboot nötig.
- `/etc/fail2ban/jail.local` — nur `[sshd] backend=systemd enabled = true`;
  bantime fällt auf 10m-Default zurück (Handbook-Vorlage: 3600s).
- `/etc/docker/daemon.json` — existiert nicht; alle 8 Container loggen
  json-file ohne Limits.
- `/etc/systemd/journald.conf:25-33` — `SystemMaxUse` unset; Journal bereits
  2,5 GB (Default-Cap 4 GB).
- Swap: 0 B bei 3,8 GB RAM.
- Docker: 18,85 GB Build-Cache (98 % reclaimable), ~40 verwaiste anonyme
  Volumes, dazu das benannte Alt-Volume `jotti-proxy_postgres-data` aus
  einem früheren Projekt.
- Dotfiles: `install-dotfiles.sh` nie gelaufen — kein `~/.bash_aliases`,
  keine `~/.gitconfig` (auch keine git-Identität), kein `gh`. Fehlende Tools:
  unzip, jq, htop, tmux, rsync, ncdu, tree. `EDITOR` unset (Default nano).

## Resolved decisions

- Passwort-Login als `nico` bleibt bestehen; nur `PermitRootLogin no` wird gesetzt.
- Kein dauerhaftes DB-Backup (Demo/Seed-Daten); Timer nach dem Test entfernen.
- Backup-Test gegen die Live-Demo inkl. destruktivem Restore-Roundtrip
  (kurze Downtime ok, Wiederherstellung notfalls via `make rocks-reset-and-seed`).
- Alle vier Verbesserungsgruppen (Updates/Härtung, Disk/Logs, Swap/Reboot,
  Dotfiles/Tools) sind im Scope.
- Dev-Toolchain (go, node, pnpm aus `setup-dev-tools.sh`) wird NICHT
  installiert — alles läuft in Docker.

> **Assumption:** Plan liegt unter `~/docs/plans/` (kein Repo). Das Handbook
> ist als englischsprachige, host-unabhängige Knowledge Base ungeeignet, das
> öffentliche jotti-Repo für VPS-spezifische Ops-Details ebenfalls.

## Open questions / Risks

- **Label-Matching: verifiziert.** `docker compose -f docker-compose.prod.yml
  ps postgres` findet den laufenden `jotti-postgres` (Labels
  `com.docker.compose.project=jotti` + `service=postgres`). Die im Live-`.env`
  fehlenden Variablen `JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL` erzeugen bei
  Compose v5.3.0 nur stderr-Warnungen („Defaulting to a blank string"), kein
  Abort — und können den gzip-Dump auf stdout nicht korrumpieren. Die
  Warnungen sind als erwartetes Finding zu dokumentieren.
- **502-Falle nach Restore** (Phase 4): nginx im weiterlaufenden
  `jotti-reverse-proxy` löst `proxy_pass http://backend:3000` /
  `http://frontend:80` nur beim Start auf (`reverse-proxy/nginx.rocks.conf:130,140`).
  Nach Recreate von backend/frontend drohen hängende 502 — deshalb zwingend
  den Reverse-Proxy force-recreaten (macht `make rocks-up` aus genau diesem
  Grund, `Makefile:188`). Ein 502 wäre sonst leicht als „Restore fehlgeschlagen"
  fehlzudeuten und könnte den destruktiven Reset-Fallback provozieren.
- **Restore-Downtime**: Während des Restores sind `backend`/`frontend`
  gestoppt (demo.jotti.rocks down, jotti.rocks-Website bleibt oben).
- **`prod-restore.sh` niemals gegen den Live-Stack ausführen**
  (`scripts/prod-restore.sh:111,136`) — Restore nur über die in Phase 4
  beschriebene rocks-Prozedur.
- **Reboot-Risiko** (Phase 5): Container haben `restart: unless-stopped` und
  Docker-Service ist enabled (verifiziert). Fallback: netcup SCP VNC-Konsole.
- **SSH-Reload-Risiko** (Phase 2): Nach `PermitRootLogin no` bestehende
  Session offen halten und Passwort-Login als `nico` in zweitem Terminal
  testen, bevor die Session geschlossen wird.
- **Volume-Prune** (Phase 3): `jotti-proxy_postgres-data` vor dem Löschen
  inspizieren — könnte alte Daten enthalten.

---

## Phase 1: Komfort — Dotfiles & Ops-Tools

### Context

- `~/handbook/install.sh:2`, `~/handbook/scripts/install-dotfiles.sh:25-68` —
  Symlink `.bash_aliases`, git-Defaults, gh-Installation nach `~/.local/bin`.
- `~/handbook/templates/.bash_aliases` — die Aliases (ll, gfp, glo, …).

### What to build

Das vorhandene Dotfiles-Setup des Handbooks auf dem VPS ausführen und die
fehlenden Ops-Basistools installieren. Danach fühlt sich die Shell an wie in
den übrigen Umgebungen (Codespaces etc.).

1. `bash ~/handbook/install.sh` ausführen.
2. git-Identität setzen: `git config --global user.name "<Name>"` und
   `git config --global user.email graef.nico@gmail.com`.
3. `sudo apt install -y unzip jq htop tmux rsync ncdu tree`.
4. `echo 'export EDITOR=vim' >> ~/.bashrc` (Default-Editor ist derzeit nano).

### Acceptance criteria

- [x] Neue Shell: `ll`, `gfp`, `glo` funktionieren (Aliases aktiv).
- [x] `git config --global --list` zeigt die vier Handbook-Defaults plus
      `user.name`/`user.email`.
- [x] `gh --version` läuft (aus `~/.local/bin`).
- [x] `jq`, `htop`, `tmux`, `rsync`, `ncdu`, `tree`, `unzip` sind installiert.
- [x] `echo $EDITOR` → `vim` in neuer Shell.

---

## Phase 2: Updates & Härtung

### Context

- `/etc/ssh/sshd_config:123` — `# PermitRootLogin yes` (auskommentiert).
- `/etc/fail2ban/jail.local` — ohne `bantime`/`maxretry`.
- `~/handbook/scripts/setup-server.sh:132-138` — Ziel-Vorlage für jail.local.
- `~/handbook/guides/provision-server.md` — SSH-Härtungs-Referenz.

### What to build

Automatische Sicherheitsupdates aktivieren und die zwei verbliebenen
Härtungslücken schließen. **`PasswordAuthentication` bleibt unangetastet.**

1. `sudo apt install -y unattended-upgrades` — das Bookworm-Paket schreibt
   `/etc/apt/apt.conf.d/20auto-upgrades` bereits bei der Installation
   (debconf-Default `enable_auto_updates: true`). Idempotente Absicherung
   statt interaktivem `dpkg-reconfigure -plow`:
   `echo 'unattended-upgrades unattended-upgrades/enable_auto_updates boolean true' | sudo debconf-set-selections && sudo dpkg-reconfigure -f noninteractive unattended-upgrades`.
2. In `/etc/ssh/sshd_config` die Zeile `# PermitRootLogin yes` durch
   `PermitRootLogin no` ersetzen, dann `sudo sshd -t` und
   `sudo systemctl reload ssh`. Bestehende Session offen halten, Login in
   zweitem Terminal testen.
3. `/etc/fail2ban/jail.local` um `maxretry = 5` und `bantime = 3600`
   ergänzen (gemäß Handbook-Vorlage), `sudo systemctl restart fail2ban`.

### Acceptance criteria

- [x] `/etc/apt/apt.conf.d/20auto-upgrades` enthält
      `Update-Package-Lists "1"` und `Unattended-Upgrade "1"`.
- [x] `sudo unattended-upgrade --dry-run --debug` läuft fehlerfrei durch.
- [x] `sudo sshd -T | grep -i permitrootlogin` → `permitrootlogin no`.
- [x] `sudo sshd -T | grep -i passwordauth` → weiterhin `yes` (unverändert!).
- [x] Passwort-Login als `nico` in neuem Terminal erfolgreich getestet.
- [x] `sudo fail2ban-client get sshd bantime` → `3600`.

---

## Phase 3: Disk- & Log-Hygiene

### Context

- `/etc/systemd/journald.conf:25-33` — `SystemMaxUse` unset, Journal 2,5 GB.
- `~/handbook/guides/docker-setup.md` (Abschnitt „Log rotation") — daemon.json-Snippet.
- Audit: 18,85 GB Build-Cache (18,57 GB reclaimable), ~40 dangling anonyme
  Volumes, stale benanntes Volume `jotti-proxy_postgres-data`.

### What to build

Journald begrenzen, Docker-Log-Rotation vorbereiten (wirksam erst nach
Container-Recreate in Phase 5) und Docker-Altlasten aufräumen.

1. `SystemMaxUse=500M` in `/etc/systemd/journald.conf` setzen,
   `sudo systemctl restart systemd-journald`,
   `sudo journalctl --vacuum-size=500M`.
2. `/etc/docker/daemon.json` anlegen:
   `{"log-driver": "json-file", "log-opts": {"max-size": "10m", "max-file": "3"}}`.
   Kein Docker-Restart jetzt — übernimmt der Reboot in Phase 5.
3. `docker builder prune -f` (~18,6 GB) und `docker image prune -f`.
4. `jotti-proxy_postgres-data` inspizieren (z. B.
   `docker run --rm -v jotti-proxy_postgres-data:/d:ro alpine ls -la /d`);
   wenn verzichtbar → explizit `docker volume rm jotti-proxy_postgres-data`.
   **Wichtig:** `docker volume prune -f` entfernt auf Docker ≥ 23 nur
   ungenutzte ANONYME Volumes — das benannte Alt-Volume überlebt den Prune,
   der manuelle `rm` ist also Pflicht, kein Fallback. Die aktiven
   `jotti_*`-Volumes sind doppelt geschützt (benannt UND in Benutzung).

### Acceptance criteria

- [x] `sudo journalctl --disk-usage` ≤ ~500 MB. (26,3 MB; `SystemMaxUse=500M` gesetzt)
- [x] `sudo dockerd --validate --config-file /etc/docker/daemon.json` →
      `configuration OK` (reine JSON-Syntaxprüfung reicht NICHT — ein
      valides JSON mit falschem Key verhindert den dockerd-Start und würde
      erst beim Reboot in Phase 5 den ganzen Stack lahmlegen).
- [x] `docker system df` → Build-Cache < 1 GB. (18,85 GB → 284 MB)
- [x] Entscheidung zu `jotti-proxy_postgres-data` dokumentiert und umgesetzt
      (verwaister PG17-Cluster aus defunktem `jotti-proxy`-Projekt, seit 13.03.
      tot, nicht vom Live-Stack genutzt → gelöscht); 42 dangling anonyme
      Volumes entfernt.
- [x] Alle 8 jotti-Container laufen weiterhin (`docker ps` → 8x Up).

---

## Phase 4: Backup-Test gegen die Live-Demo

### Context

- `~/jotti/scripts/prod-backup.sh:23-24,96-101,111-122` — Testgegenstand.
- `~/jotti/scripts/prod-restore.sh:111,114,136` — NICHT verwenden (rocks-unsicher).
- `~/jotti/packaging/systemd/jotti-backup.{service,timer}` — Automation-Testgegenstand.
- `~/jotti/docker-compose.rocks.yml:18-115` — Services `postgres`, `backend`,
  `frontend` für die Restore-Prozedur; DB `jotti` (`:25`).
- `~/jotti/Makefile:200` — `rocks-reset-and-seed` als Notfall-Wiederherstellung.

### What to build

End-to-End-Test des automatisierten jotti-Backups gegen die laufende
Demo-Instanz: Dump → Manipulation → Restore → Verifikation → Timer-Test →
Aufräumen. Alle Erkenntnisse werden als Findings festgehalten.

1. **Pre-Check** (read-only, bereits bestätigt):
   `docker compose -f docker-compose.prod.yml ps postgres` findet den
   laufenden `jotti-postgres`.
2. **Referenzzustand festhalten** — UNMITTELBAR VOR dem Dump (die Demo ist
   öffentlich; jeder Write zwischen Dump und Referenz erzeugt sonst ein
   falsches Negativ): Zeilenzahlen relevanter Tabellen via
   `docker compose -f docker-compose.rocks.yml exec -T postgres
   sh -c 'psql -U "$POSTGRES_USER" -d jotti -c "..."'`.
3. **Dump-Test**: `make prod-backup` in `~/jotti`. Erwartung: Dump landet in
   `~/jotti/backups/jotti-<timestamp>.sql.gz`. Zwei stderr-Warnungen zu
   `JOTTI_DOMAIN`/`LETSENCRYPT_EMAIL` sind erwartet und harmlos (→ Finding).
   Dump verifizieren: `gzip -dc <dump> | grep -c 'CREATE TABLE'` > 0 und
   Stichprobe realer Seed-Daten.
4. **Daten manipulieren**: Marker-Datensatz einfügen (oder einen
   Seed-Datensatz ändern) — beweist später, dass der Restore wirklich
   zurücksetzt.
5. **Rocks-sicherer Restore**:
   `docker compose -f docker-compose.rocks.yml stop backend frontend`,
   dann `gzip -dc <dump> | docker compose -f docker-compose.rocks.yml exec -T
   postgres sh -c 'psql -U "$POSTGRES_USER" -d jotti -v ON_ERROR_STOP=1'`,
   dann `docker compose -f docker-compose.rocks.yml up -d` und **zwingend**
   `docker compose -f docker-compose.rocks.yml up -d --no-deps
   --force-recreate reverse-proxy` (sonst 502-Falle, siehe Risiken).
   Hinweis: `up -d` lässt den exited One-off `migrate` erneut laufen —
   gegen das restaurierte Schema ein harmloser No-op, aber erwartbare Ausgabe.
6. **Verifikation**: Marker ist verschwunden, Zeilenzahlen == Referenz
   (bei Abweichung erst die COPY-Blöcke im Dump zählen, bevor „Restore
   fehlgeschlagen" gefolgert wird), https://demo.jotti.rocks funktioniert
   (Login, Kernseiten).
7. **Timer-Test**: `jotti-backup.service` mit angepasstem
   `WorkingDirectory=/home/nico/jotti` + `ExecStart=/home/nico/jotti/scripts/prod-backup.sh`
   und `jotti-backup.timer` nach `/etc/systemd/system/` kopieren,
   `daemon-reload`, dann `sudo systemctl enable --now jotti-backup.timer`
   (ohne aktiven Timer erscheint er nicht in `systemctl list-timers`!) und
   `sudo systemctl start jotti-backup.service` (manueller Trigger des
   oneshot). Verify: neuer Dump (root-owned — kosmetisch, Timer läuft als
   root) + sauberes `journalctl -u jotti-backup`; `systemctl list-timers`
   zeigt den 03:30-Schedule.
8. **Aufräumen**: `sudo systemctl disable --now jotti-backup.timer`, Units
   aus `/etc/systemd/system/` entfernen, `daemon-reload`. Test-Dumps
   optional behalten (gitignored) oder löschen.
9. **Findings dokumentieren** (in diesem Plan-File, Abschnitt unten
   ergänzen): env-Warnungen von `prod-backup.sh` gegen rocks;
   `prod-restore.sh`-Inkompatibilität; `/opt/jotti`-Hardcoding der Units;
   ggf. Follow-up-Issues fürs jotti-Repo formulieren.

### Acceptance criteria

- [x] Dump der Live-DB erzeugt: nicht-leer, inhaltlich plausibel, 17 Tabellen
      == Referenz, nur die zwei erwarteten env-Warnungen auf stderr.
      **ABER:** bares `make prod-backup` scheitert lautlos (Finding F2) —
      Dump nur mit `BACKUP_DIR=./backups BACKUP_KEEP=14` erzeugbar.
- [x] Restore-Roundtrip: Marker weg, Datenstand == Referenz (17/17 Tabellen).
- [x] https://demo.jotti.rocks und https://jotti.rocks funktionieren nach dem
      Restore (Reverse-Proxy force-recreatet, Backend via Proxy `/api/health`
      → 200, kein 502); alle 8 Container Up.
- [x] Timer-Test: manueller Trigger erzeugt Dump (mit `Environment=`-Workaround;
      wie ausgeliefert → `status=1/FAILURE`, 0 Dumps — F2/F3), `list-timers`
      zeigt den aktivierten 03:30-Timer.
- [x] Timer/Service wieder deinstalliert (`list-timers` ohne jotti-backup;
      keine Units mehr in `/etc/systemd/system/`).
- [x] Findings-Abschnitt ergänzt (siehe unten + `findings-phase4-backup-test.md`).

---

## Phase 5: Wartungsfenster — Swap, Reboot, Log-Rotation aktivieren

### Context

- Audit: Kernel 6.1.0-50 installiert, 6.1.0-44 läuft; Swap 0 B;
  `daemon.json` aus Phase 3 wartet auf Docker-Neustart; Log-Opts gelten nur
  für neu erstellte Container.
- `~/jotti/Makefile:185-188` — `rocks-up` als Referenz für den Stack-Start.

### What to build

Ein gebündeltes Wartungsfenster (wenige Minuten Downtime), das Kernel,
Docker-Daemon-Konfiguration und Container-Log-Rotation gleichzeitig aktiviert
und die Reboot-Resilienz des Stacks beweist. Frischer Dump aus Phase 4
existiert als Sicherheitsnetz.

1. **Preflight**: `sudo dockerd --validate --config-file
   /etc/docker/daemon.json` → `configuration OK` erneut prüfen (eine
   fehlerhafte daemon.json würde dockerd nach dem Reboot nicht starten
   lassen → kompletter Stack down, Recovery nur via netcup-VNC).
2. Swap einrichten: `sudo fallocate -l 2G /swapfile && sudo chmod 600
   /swapfile && sudo mkswap /swapfile && sudo swapon /swapfile`, Eintrag
   `/swapfile none swap sw 0 0` in `/etc/fstab`, `vm.swappiness=10` via
   `/etc/sysctl.d/99-swappiness.conf`. (`/` ist ext4 — fallocate-Swapfile
   ist dort unproblematisch; verifiziert.)
3. `sudo reboot`.
4. Nach dem Reboot: Stack-Zustand prüfen (kommt per `unless-stopped` allein
   hoch), dann `docker compose -f docker-compose.rocks.yml up -d
   --force-recreate` um die neuen Log-Opts auf alle Container anzuwenden.

### Acceptance criteria

- [x] `uname -r` → `6.1.0-50-amd64` (Reboot 2026-07-05 11:44).
- [x] `free -h` → 2,0 GB Swap aktiv (`/swapfile`, 2097148 KB); `swappiness` → 10;
      übersteht den Reboot (fstab-Zeile 19 + `/etc/sysctl.d/99-swappiness.conf`).
- [x] Alle 8 jotti-Container Up (backend+postgres healthy), https://jotti.rocks
      und https://demo.jotti.rocks → 200, TLS gültig (`ssl_verify=0`),
      Backend via Proxy `/api/health` → 200.
- [x] `docker inspect … LogConfig` → `max-size:10m, max-file:3` (Container
      um 11:45 post-Boot recreatet; verifiziert an backend/postgres/reverse-proxy/
      frontend/website).
- [x] `systemctl is-active ufw fail2ban docker` → dreimal `active`.
- [x] Uptime neu (Boot 11:44); `journalctl -p err -b` → 0 Fehlerzeilen.

---

## Findings aus Phase 4 (durchgeführt 2026-07-05)

> Volldetail (Mechanismus, Evidenz, Fix-Diffs) in
> [`findings-phase4-backup-test.md`](findings-phase4-backup-test.md).

- **F1 — env-Warnungen (kosmetisch).** `prod-backup.sh` läuft über
  `docker-compose.prod.yml`; die dort referenzierten `JOTTI_DOMAIN` /
  `LETSENCRYPT_EMAIL` fehlen in der rocks-`.env` → zwei stderr-Warnungen.
  Harmlos: Warnungen auf stderr, Dump auf stdout, nicht korrumpiert (verifiziert).

- **F2 — `prod-backup.sh` stirbt lautlos (HOCH, echter Bug).** Unter
  `set -euo pipefail` liefert `read_env` bei fehlendem Key (`.env` hat kein
  `BACKUP_DIR=`/`BACKUP_KEEP=`) via `pipefail` Exit≠0; das schlägt auf die
  Zuweisung `BACKUP_DIR="${BACKUP_DIR:-$(read_env …)}"` durch → `set -e`
  bricht **vor** dem `./backups`-Default und **vor jeder Ausgabe** ab (Exit 1,
  kein Dump). Der eigentlich vorgesehene Default-Pfad ist unerreichbar.
  Fix verifiziert: `grep … || true` in `read_env`. Workaround: Env-Vars setzen.

- **F3 — systemd-Units doppelt kaputt (HOCH).** `jotti-backup.service`
  hardcodet `/opt/jotti` (`WorkingDirectory`+`ExecStart`) **und** lässt
  `Environment=BACKUP_DIR/BACKUP_KEEP` auskommentiert → der ausgelieferte Timer
  trifft F2 und produziert **null Dumps** (im Test bestätigt: `status=1/FAILURE`).
  systemd meldet zwar `failed`, aber nur in `journalctl` sichtbar.

- **F4 — `prod-restore.sh` nicht rocks-sicher (by inspection).** Nutzt
  `docker-compose.prod.yml` (`up -d`) → würde den rocks-Stack umkonfigurieren.
  Im Test daher die rocks-sichere Prozedur verwendet (`psql`-Restore in den
  laufenden `postgres` + Reverse-Proxy force-recreate).

**Empfohlene Follow-ups fürs jotti-Repo:** (1) F2-Fix in `prod-backup.sh`
[höchste Prio, Backup-Zuverlässigkeit]; (2) F3: `ExecStart`-Pfad im
Unit-Kommentar erwähnen + klarstellen, dass ohne F2-Fix `Environment=` Pflicht
ist; (3) rocks-sichere Restore-Prozedur dokumentieren/skripten; (4) optional
`${JOTTI_DOMAIN:-}`-Defaults in `docker-compose.prod.yml`.
