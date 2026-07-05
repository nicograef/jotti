# Findings — Phase 4 Backup-Test (jotti.rocks Live-Demo)

> Kontext: Backup-Test aus `plan-vps-improvements.md`, Phase 4, gegen den
> laufenden `docker-compose.rocks.yml`-Stack (Compose-Projekt `jotti`).
> Durchgeführt: **2026-07-05**, Host jotti.rocks (netcup-VPS), Checkout
> `/home/nico/jotti`.
>
> **Auditiert gegen jotti-Repo @ `21d8c4b` (branch `main`, gepusht zu
> `origin` = `github.com/nicograef/jotti`).** Alle `:NN`-Zeilenreferenzen
> gelten für diesen Commit. Auf dem Dev-Laptop zuerst sicherstellen:
> `git -C <jotti> rev-parse HEAD` == `21d8c4b…` (sonst `git fetch && git
> checkout 21d8c4b`, oder die Refs anhand des zitierten Codes neu verorten —
> die Diffs unten haben genug Kontext, um auch bei Zeilendrift zu greifen).
>
> Legende Status: **Confirmed** = im Test reproduziert · **By inspection** =
> aus Code/Config abgeleitet, (noch) nicht ausgeführt.

## Übersicht

| # | Finding | Schwere | Status |
|---|---------|---------|--------|
| F1 | `prod-backup.sh` gibt zwei env-Warnungen gegen den rocks-Stack aus | Kosmetisch | Confirmed |
| F2 | `prod-backup.sh` **stirbt lautlos** ohne `BACKUP_DIR`/`BACKUP_KEEP` in `.env` | **Hoch** | Confirmed |
| F3 | systemd-Units hardcoden `/opt/jotti` **und** lassen die Env-Vars aus → Timer doppelt kaputt auf diesem Host | Hoch | Confirmed (systemd) |
| F4 | `prod-restore.sh` ist gegen den rocks-Stack nicht sicher | Hoch (bei Fehlbedienung) | By inspection |

---

## F1 — env-Warnungen von `prod-backup.sh` gegen rocks (kosmetisch)

**Beobachtung.** `make prod-backup` (das intern `docker-compose.prod.yml`
verwendet) gibt beim Lesen der prod-Compose-Datei zwei Warnungen auf **stderr**:

```
level=warning msg="The \"JOTTI_DOMAIN\" variable is not set. Defaulting to a blank string."
level=warning msg="The \"LETSENCRYPT_EMAIL\" variable is not set. Defaulting to a blank string."
```

**Ursache.** `docker-compose.prod.yml` referenziert `JOTTI_DOMAIN` /
`LETSENCRYPT_EMAIL`, die in der rocks-`.env` nicht vorkommen (rocks nutzt
`docker-compose.rocks.yml` mit anderen Variablen). Compose interpoliert sie zu
Leerstrings.

**Impact.** Keiner für den Backup-Pfad. Die Warnungen gehen nach **stderr**,
der gzip-Dump nach **stdout** — der Dump wird nicht korrumpiert (im Test
verifiziert: Dump vollständig, 17/17 Tabellen deckungsgleich). Reines Rauschen.

**Empfehlung.** Akzeptieren oder unterdrücken (z. B. die beiden Variablen mit
Defaults in `docker-compose.prod.yml` versehen: `${JOTTI_DOMAIN:-}`). Niedrige
Priorität.

---

## F2 — `prod-backup.sh` stirbt lautlos ohne `BACKUP_DIR`/`BACKUP_KEEP` (HOCH)

**Beobachtung.** `make prod-backup` bricht mit **Exit 1 und komplett ohne
Ausgabe** ab (weder `[INFO]` noch `[ERROR]`), es wird **kein Dump** erzeugt.
Für ein Backup-System der gefährlichste Fehlermodus: stiller Totalausfall.

**Root Cause.** `scripts/prod-backup.sh` läuft unter `set -euo pipefail`
(`:2`). Die aktuelle `.env` enthält **keine** `BACKUP_DIR=`/`BACKUP_KEEP=`-Zeile.

`read_env()` (`:42-45`):

```bash
read_env() {
  local key="$1"
  grep -E "^${key}=" .env 2>/dev/null | tail -n1 | cut -d= -f2- | sed '...'
}
```

Fehlt der Key, findet `grep` nichts und beendet sich mit Exit 1. Wegen
`pipefail` gibt die **gesamte Pipeline** Exit 1 zurück. Dieser Status schlägt
auf die Zuweisung durch (`:71`):

```bash
BACKUP_DIR="${BACKUP_DIR:-$(read_env BACKUP_DIR)}"
```

Der Exit-Status der Zuweisung ist der der letzten Kommando-Substitution (= 1).
Unter `set -e` **bricht die Shell hier sofort ab** — noch **vor** dem
`./backups`-Default (`:72`) und vor jeder Ausgabe.

**Evidenz** (`bash -x`, letzte ausgeführte Zeile):

```
+ [[ ! -f .env ]]
++ read_env BACKUP_DIR
++ grep -E '^BACKUP_DIR=' .env
+ BACKUP_DIR=            <-- danach sofort Exit 1, keine weitere Zeile
```

Minimal-Repro bestätigt: buggy `read_env` → Exit 1, Default-Zeile nie erreicht;
gefixte Variante → Default `./backups` erreicht, Exit 0.

**Pikant.** Der `[[ -n "$BACKUP_DIR" ]] || BACKUP_DIR="./backups"`-Fallback und
der Kommentar „default: ./backups" (`:17,72`) zeigen: das Skript ist explizit
dafür gebaut, **ohne** `.env`-Einträge zu funktionieren. Genau dieser
Default-Pfad ist durch `set -e` + `pipefail` **unerreichbar**.

**Fix (minimal, zielgerichtet).** `grep`-No-Match unter `pipefail` abfangen:

```diff
 read_env() {
   local key="$1"
-  grep -E "^${key}=" .env 2>/dev/null | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
+  { grep -E "^${key}=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2- | sed 's/^[[:space:]]*//; s/[[:space:]]*$//'
 }
```

Alternativen: pro Zuweisung `... $(read_env BACKUP_DIR || true) ...`, oder
`read_env` ein explizites `|| true` / `return 0` verpassen. Der `grep || true`
oben ist am DRY-sten (fixt beide Aufrufer auf einmal) und schluckt nur den
No-Match, keine echten Fehler.

**Fix ohne Prod/DB verifizierbar** (reine Shell-Logik — auf dem Laptop in einem
jotti-Checkout ausführen, dessen `.env` kein `BACKUP_DIR=` enthält; ein leeres
`.env` genügt auch):

```bash
# VORHER (buggy): stirbt lautlos, Exit 1, KEINE "erreicht"-Zeile
bash -c 'set -euo pipefail
  read_env(){ grep -E "^$1=" .env 2>/dev/null | tail -n1 | cut -d= -f2-; }
  D="${D:-$(read_env BACKUP_DIR)}"; echo "erreicht: D=[$D]"'; echo "exit=$?"

# NACHHER (fix: grep || true): Exit 0, "erreicht: D=[./backups]"
bash -c 'set -euo pipefail
  read_env(){ { grep -E "^$1=" .env 2>/dev/null || true; } | tail -n1 | cut -d= -f2-; }
  D="${D:-$(read_env BACKUP_DIR)}"; [ -n "$D" ] || D=./backups; echo "erreicht: D=[$D]"'; echo "exit=$?"
```

**Impact über den Handbetrieb hinaus.** Betrifft direkt den systemd-Timer
(→ F3): der läuft ohne Env-Vars und scheitert jede Nacht identisch.
**Confirmed in systemd (Step 7b):** der wie ausgeliefert installierte Service
(Env-Vars auskommentiert) endet mit `status=1/FAILURE`, **0 Dumps**, kein
`[INFO]`/`[ERROR]` im Journal. Mit gesetztem `Environment=BACKUP_DIR/BACKUP_KEEP`
(Step 7c) läuft er sauber durch (192K-Dump, nur die zwei F1-Warnungen).

**Workaround für den laufenden Test.** `BACKUP_DIR=./backups BACKUP_KEEP=14
make prod-backup` — bei gesetzten Env-Vars wird der `:-`-Default gar nicht erst
ausgewertet, `read_env` nie aufgerufen, Skript bleibt unverändert. Damit wurde
der Test-Dump erfolgreich erzeugt (192K, 17/17 Tabellen deckungsgleich).

---

## F3 — systemd-Units: `/opt/jotti` hardcodiert + Env-Vars auskommentiert (HOCH)

**Beobachtung.** `packaging/systemd/jotti-backup.service`:

- `WorkingDirectory=/opt/jotti` und `ExecStart=/opt/jotti/scripts/prod-backup.sh`
  (`:20-21`) — der reale Checkout ist `/home/nico/jotti`. So wie ausgeliefert
  findet die Unit das Skript hier nicht.
- `Environment=BACKUP_DIR=…` / `BACKUP_KEEP=…` sind **auskommentiert**
  (`:22-23`). Der Timer startet das Skript also **ohne** diese Vars → er läuft
  in **F2** und scheitert bei jedem Lauf.

**Konsequenz.** Der Timer ist auf diesem Host **doppelt kaputt**: falscher Pfad
*und* Silent-Death-Bug. systemd markiert den oneshot zwar als `failed`
(sichtbar nur in `journalctl -u jotti-backup`), aber es entstehen **null
Dumps** — man glaubt, ein nächtliches Backup zu haben, hat aber keins.

**Fix.**

1. **Pfade** an den echten Checkout anpassen. Der Header dokumentiert nur
   `WorkingDirectory`, **vergisst `ExecStart`** — beide müssen angepasst werden:

   ```diff
   -WorkingDirectory=/opt/jotti
   -ExecStart=/opt/jotti/scripts/prod-backup.sh
   +WorkingDirectory=/srv/jotti                    # euer Checkout-Pfad
   +ExecStart=/srv/jotti/scripts/prod-backup.sh
   ```

2. **Env-Vars.** Ist **F2 gefixt**, trägt der `./backups`-Default und die
   `Environment=`-Zeilen dürfen auskommentiert bleiben. **Ohne** F2-Fix sind sie
   Pflicht — sonst stirbt der Timer wie in Step 7b (`status=1/FAILURE`, 0 Dumps):

   ```diff
   -# Environment=BACKUP_DIR=/var/backups/jotti
   -# Environment=BACKUP_KEEP=14
   +Environment=BACKUP_DIR=/srv/jotti/backups
   +Environment=BACKUP_KEEP=14
   ```

Empfehlung: den F2-Fix machen (Punkt 2 wird damit optional) **und** im
Unit-Kommentar `ExecStart` mit erwähnen.

**Getestet (Step 7).** Genau diese zwei Anpassungen (Pfade auf
`/home/nico/jotti` + `Environment=` gesetzt) ließen den oneshot sauber
durchlaufen; ohne die `Environment=`-Zeilen scheiterte er reproduzierbar. Siehe
Test-Ergebnisse unten.

---

## F4 — `prod-restore.sh` ist nicht rocks-sicher (By inspection)

**Beobachtung.** `scripts/prod-restore.sh` fährt alle Container-Operationen über
`docker-compose.prod.yml`:

- `:111` — `up -d --wait postgres` (startet/recreatet postgres per prod-Config)
- `:114` — `stop backend frontend reverse-proxy`
- `:128-129` — der eigentliche Restore
  (`decompress | docker compose … exec -T postgres sh -c 'psql … -v ON_ERROR_STOP=1'`)
  — **inhaltlich identisch** zur rocks-sicheren Prozedur
- `:136` — finales `up -d` (recreatet den **ganzen Stack** per prod-Config)

Gegen den laufenden rocks-Stack würden `:111` und `:136` die Container nach
**prod-Konfiguration** umbauen (andere Volumes/Netze/Env) statt nur die DB
zurückzuspielen; zusätzlich fehlt der Reverse-Proxy-Force-Recreate → 502-Falle.

**Impact.** Bei versehentlichem Einsatz gegen die Live-Demo: Rekonfiguration
des Stacks, potenziell Datenverlust/Downtime. Deshalb im Plan explizit
verboten.

**Handhabung im Test.** Restore erfolgt **nicht** über `prod-restore.sh`,
sondern über die rocks-sichere Prozedur (stop `backend`/`frontend` → Dump per
`psql` in den laufenden `postgres` → `up -d` → Reverse-Proxy force-recreate).

**Empfehlung.** Der Kern-Restore (`:128-129`) ist bereits rocks-tauglich; nur
die umgebenden `up`-Aufrufe hängen an `COMPOSE_PROD`. Eine rocks-sichere Variante
ist daher billig: Compose-Datei parametrisieren (z. B. `COMPOSE_FILE` per
Env/Flag statt hartem `docker-compose.prod.yml`) und nach dem finalen `up -d` den
Reverse-Proxy force-recreaten. Getestete, rocks-sichere Sequenz (manuell, ohne
das Skript):

```bash
DUMP=backups/jotti-<ts>.sql.gz
docker compose -f docker-compose.rocks.yml stop backend frontend
gzip -dc "$DUMP" | docker compose -f docker-compose.rocks.yml exec -T postgres \
  sh -c 'psql -U "$POSTGRES_USER" -d jotti -v ON_ERROR_STOP=1'
docker compose -f docker-compose.rocks.yml up -d
docker compose -f docker-compose.rocks.yml up -d --no-deps --force-recreate reverse-proxy
```

---

## Test-Ergebnisse (Steps 4–8, durchgeführt 2026-07-05)

- [x] **Dump (Step 3):** `jotti-20260705-110129.sql.gz`, 192K, 17 `CREATE TABLE`,
      Dump endet sauber; alle 17 COPY-Zeilenzahlen == Referenz-Snapshot.
- [x] **Restore-Roundtrip (Steps 4–6):** Marker (`produkte`-Row
      `PHASE4-RESTORE-MARKER`) eingefügt (Count 22→23), rocks-sicherer Restore
      (`backend`/`frontend` stop → `psql -v ON_ERROR_STOP=1`, exit 0, keine
      Fehler → `up -d` → Reverse-Proxy force-recreate). Danach: Marker weg
      (Count zurück auf 22), **alle 17 Tabellen == Referenz**.
- [x] **Kein 502 (Step 6):** `demo.jotti.rocks` & `jotti.rocks` HTTP 200,
      TLS gültig; Backend durch den Proxy erreichbar (`/api/health` → 200).
- [x] **Timer-Test (Step 7):** wie ausgeliefert → `status=1/FAILURE`, 0 Dumps
      (F2/F3 bestätigt); mit `Environment=`-Workaround → Exit 0, root-owned Dump,
      sauberes Journal; `list-timers` → `Mon 2026-07-06 03:30:00 CEST`.
- [x] **Cleanup (Step 8):** Timer disabled, Units entfernt (0 left),
      Test-Dump gelöscht, `list-timers` ohne jotti-backup. Nur der Step-3-Dump
      bleibt (`backups/`, nico-owned, gitignored).

## Empfohlene Follow-ups fürs jotti-Repo

1. **F2-Fix** in `scripts/prod-backup.sh` (`read_env` `grep … || true`).
   Höchste Priorität — betrifft die Zuverlässigkeit des gesamten Backups.
2. **F3** — `ExecStart`-Pfad im Unit-Kommentar mit erwähnen; klarstellen, dass
   ohne F2-Fix `Environment=BACKUP_DIR/BACKUP_KEEP` zwingend gesetzt sein muss.
3. **F4** — rocks-sichere Restore-Prozedur dokumentieren/skripten.
4. **F1** — optional `${JOTTI_DOMAIN:-}` Defaults in `docker-compose.prod.yml`.
