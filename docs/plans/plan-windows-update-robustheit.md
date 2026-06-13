# Plan: Robuste Windows-Updates & Betriebssicherheit

> Source PRD: n/a (Task-Beschreibung)
> Nachfolger von `plan-windows-verpackung.md` (Phase 4/5 lieferten ZIP + Doku;
> dieser Plan härtet den **Update- und Datenlebenszyklus**).

## Goal

Updates des Windows-Release für **nicht-technische Vereinshelfer** so robust und
einfach machen, dass keine plausible Fehlbedienung Daten verliert oder
unzugänglich macht. Heutiger Zustand: Daten liegen in Docker-Volumes, der
**Schlüssel** dazu (zufälliges `POSTGRES_PASSWORD`) liegt als loses `.env` im
Programmordner — genau dem Ordner, den der Nutzer beim Update anfasst.

## Grundproblem (Reframe)

> **Installations-Zustand ist an den Programmordner gekoppelt.**

Alle Update-Risiken sind Symptome davon. Die Lösung trennt zwei Fragen sauber:

- **Wo lebt die Installation?** → Konventions-/Ortsfrage → `%PROGRAMDATA%\jotti`.
- **Was ist die Quelle der Wahrheit für das Secret?** → Lebenszyklusfrage →
  ein **Docker-Volume**, das dasselbe Schicksal wie die Daten teilt.

Beide ergänzen sich: das Volume garantiert Korrektheit (Schlüssel und Daten
können nicht auseinanderdriften), `%PROGRAMDATA%\jotti` liefert den stabilen,
konventionsgerechten **Host-Spiegel**, den die beiden Datei-Leser (Starter für
`compose --env-file`, Relay als nicht-elevierter reiner Datei-Leser) brauchen.

## Risiken (nach Schwere)

| # | Risiko | Schwere | Adressiert in |
|---|--------|---------|---------------|
| 1 | `.env` verloren/neu generiert → `POSTGRES_PASSWORD` ≠ Volume → Aussperrung aus den eigenen Daten | 🔴 | Phase 1 |
| 2 | Kein Backup + forward-only Migrationen → kaputte Migration/Update ohne Wiederherstellungspunkt | 🔴 | Phase 3 |
| 3 | Keine Update-Kenntnis → Helfer laufen ewig auf alter Version | 🟡 | Phase 5 |
| 4 | Kein Rollback / nicht-atomares Update | 🟡 | Phase 3 (Restore) |
| 5 | „Über denselben Ordner entpacken" ist fehleranfällig (Fresh-Folder → Risiko 1) | 🟡 | Phase 2 |
| 6 | Migrationen per Bind-Mount aus dem Ordner → Stack nicht selbst-enthalten | 🟢 | Phase 4 |

## Architectural decisions

- **Secret-Quelle der Wahrheit: `jotti-config`-Volume.** Wird beim Erststart
  einmal befüllt (frische Secrets), danach nur gelesen. Entfernt **nur**
  zusammen mit den Daten (`down -v`/Volume-Löschung) → Aussperrung strukturell
  unmöglich.
- **Host-Zustandsverzeichnis: `%PROGRAMDATA%\jotti\`** (nur Windows; unter
  Linux-Dev-Lauf Fallback auf ordnerlokal, analog zu den bestehenden
  `runtime.GOOS=="windows"`-Schritten). Enthält den **`.env`-Spiegel**, einen
  `last-version`-Marker und exportierte Backups. Beide Exes lesen `.env` von
  hier. **Keine Binaries** in ProgramData — Programmdateien bleiben dort, wo der
  Nutzer entpackt (Ort wird durch Volume+Spiegel sicherheitsirrelevant).
- **Backups: `jotti-backups`-Volume.** `pg_dump` zeitgestempelt, **vor** den
  Migrationen, ausgelöst durch erkannten Versionswechsel. Optionaler Export nach
  `%PROGRAMDATA%\jotti\backups\`.
- **Migrationen: ins `jotti-migrate`-Image gebacken** (`COPY`), Bind-Mount aus
  der Compose entfernt → Stack selbst-enthalten.
- **Update-Modell: manuell + benachrichtigend.** Download → entpacken → starten.
  Der Starter **meldet** neue Versionen (online), **wendet sie nie automatisch
  an** (Auto-Update einer laufenden Kasse mitten im Fest ist die falsche
  Voreinstellung). Selbst-aktualisierender Starter (Download+Entpacken+Neustart)
  ist **explizit out-of-scope** — das Überschreiben einer laufenden Exe unter
  Windows ist fragil und mit stateless Ordner unnötig.
- **Unverändert:** Projektname `jotti-local`; bestehende Volumes
  (`postgres-data`, `caddy-data`, `proxy-state`). Neu: `jotti-config`,
  `jotti-backups`.

## Inventory

- `cmd/starter/main.go:50-58` — `MaterializeEnv` schreibt `.env` ordnerlokal,
  idempotent; läuft aktuell **vor** `ensureDocker` (Phase 1 dreht das um).
- `cmd/starter/core/env.go:27-39` — idempotente Erzeugung; Basis für
  „aus Volume wiederherstellen statt neu generieren".
- `cmd/starter/main.go:91-112` — `resolveComposeFile` sucht neben der Exe / im
  Arbeitsverzeichnis; bleibt, ergänzt um ProgramData-Auflösung für `.env`.
- `cmd/starter/system.go:243-259` — `composeUp`/`runCompose`: zentrale Stelle für
  `--env-file`, Postgres-zuerst-Sequenz und `pg_dump`-Schritt.
- `cmd/relay/env.go:37-51` — Relay liest `.env` neben der Exe / im wd; muss
  zusätzlich `%PROGRAMDATA%\jotti\.env` lesen.
- `database/migrate/Dockerfile` — installiert nur das `migrate`-Binary,
  **enthält keine Migrationen**; `CMD … -path /migrations …`.
- `docker-compose.release.yml:50-51` — Bind-Mount `./database/migrations:/migrations`
  (Phase 4 entfernt).
- `docker-compose.release.yml:137-141` — Volume-Block (erweitern).
- `packaging/windows/KURZANLEITUNG.md` — Doku ohne Update-Abschnitt (Phase 5).
- `packaging/windows/jotti-stop.cmd` — nutzt `down` (kein `-v`); bleibt.

## Resolved decisions

- **Scope: Full Suite** (alle sechs Risiken).
- **Secret-Ablage: Volume-autoritativ + ProgramData-Spiegel** (begründet oben:
  Volume = Korrektheit, ProgramData = konventioneller Host-Spiegel für die
  Datei-Leser).
- **`pg_dump` vor Migration** durch „erst nur `postgres` hochfahren, dumpen, dann
  voller `up`" — Backup entsteht vor dem schemaverändernden `migrate`-Service.
- **Online-Check non-fatal & kurz getimt**; offline → still überspringen.

## Open questions / Risks

- **Reihenfolge Starter/Relay:** Relay läuft nicht-eleviert und kann vor dem
  Starter laufen. Fehlt der `.env`-Spiegel, meldet das Relay klar „zuerst
  jotti-start.exe ausführen" — kein stiller Fehler.
- **Compose-Interpolation braucht Host-Werte:** Reihenfolge im Starter wird
  `ensureDocker` → Secret aus Volume lesen/erzeugen → Host-`.env` schreiben →
  `compose --env-file`.
- **Backup-Aufbewahrung:** letzte N Dumps rotieren (Volume füllt sonst).
  Konkrete N in der Implementierung; Vorschlag 5.
- **Linux-Dev-Lauf** muss grün bleiben: ProgramData-Pfad nur unter Windows,
  sonst ordnerlokal (wie heute).

---

## Phase 1: Volume als Quelle der Wahrheit fürs Secret (Lockout unmöglich)

**Risiken:** 1

### Context

- `cmd/starter/main.go:50-70` — Reihenfolge `MaterializeEnv`↔`ensureDocker`
  umdrehen (der Volume-Read braucht einen laufenden Docker-Daemon).
- `cmd/starter/core/env.go:27-39` — „erzeugen" → „aus Volume lesen-oder-erzeugen".
- `cmd/starter/system.go:243-259` — Compose erhält `--env-file` auf den Host-Spiegel.

### What to build

Das `jotti-config`-Volume wird Quelle der Wahrheit fürs Install-Secret. Nach
`ensureDocker`: Volume vorhanden → Secret lesen und als Host-`.env`
materialisieren; leer → frische Secrets erzeugen, **ins Volume schreiben** und
spiegeln. Der Host-Spiegel bleibt in dieser Phase **ordnerlokal** (neben der
Compose) — die Relokation nach ProgramData folgt in Phase 2. Compose nutzt
`--env-file` auf den Spiegel. Unter Linux (Dev) bleibt alles ordnerlokal ohne
Volume-Zwang.

Damit ist der kritische Lockout strukturell ausgeschlossen — **unabhängig** davon,
wo der Spiegel liegt: Geht der Host-`.env` verloren oder wird im fremden Ordner
neu gestartet, stellt der Starter dasselbe Secret aus dem Volume wieder her.

### Acceptance criteria

- [x] Erststart erzeugt Secrets und legt sie im `jotti-config`-Volume ab
      (+ ordnerlokaler Spiegel).
- [x] Host-`.env` löschen → nächster Start restauriert **dasselbe** Secret aus
      dem Volume; Stack kommt hoch, **Daten intakt**.
- [x] Neues ZIP in **fremden Ordner** entpacken → Start zieht dasselbe Secret aus
      dem Volume → **kein Lockout**, Daten intakt.
- [x] Nur `down -v` / Volume-Löschung entfernt das Secret (zusammen mit den Daten).
- [x] Linux-Dev-Lauf (`go run`) unverändert grün (ordnerlokales `.env`).

---

## Phase 2: ProgramData-Zustandsverzeichnis (kanonischer Ort, Relay-unabhängig)

**Risiken:** 5

### Context

- `cmd/starter/main.go:50,91-112` — `.env`-/State-Pfad nach `%PROGRAMDATA%\jotti`
  (nur Windows) statt ordnerlokal.
- `cmd/relay/env.go:37-51` — Relay liest `.env` zusätzlich aus `%PROGRAMDATA%\jotti`.

### What to build

Den Host-Spiegel und kleinen Maschinen-State (`.env`-Spiegel, `last-version`-Marker,
Backup-Export-Anker) aus dem Programmordner nach `%PROGRAMDATA%\jotti\` verlegen
(nur Windows; Linux-Dev bleibt ordnerlokal). Beide Exes lesen `.env` von dort.
Programmdateien bleiben dort, wo der Nutzer entpackt — der Ort wird damit
sicherheits- **und** bedienungsirrelevant: „irgendwohin entpacken und starten"
genügt, die „über denselben Ordner"-Regel entfällt.

### Acceptance criteria

- [x] Starter schreibt/liest `.env` und `last-version` unter `%PROGRAMDATA%\jotti`.
- [x] Relay findet `.env` in `%PROGRAMDATA%\jotti`, unabhängig vom eigenen Ordner;
      fehlt der Spiegel, klare Meldung „zuerst jotti-start.exe ausführen".
- [x] ZIP in **beliebigen** Ordner entpacken und starten → funktioniert ohne
      „über denselben Ordner"-Regel.
- [x] Linux-Dev-Lauf unverändert grün (ordnerlokal).

---

## Phase 3: Automatisches Pre-Update-Backup + Restore-Pfad

**Risiken:** 2, 4

### Context

- `cmd/starter/system.go:243-259` — Sequenz „erst `postgres`, dumpen, dann voller `up`".
- `%PROGRAMDATA%\jotti\last-version` (Phase 2) — Versionswechsel erkennen.

### What to build

Der Starter vergleicht die eigene `version` mit `last-version`. Bei Wechsel
**und** existierendem `postgres-data`: nur `postgres` hochfahren, healthy
abwarten, `pg_dump` zeitgestempelt ins `jotti-backups`-Volume schreiben (letzte N
rotieren), dann voller `up` (inkl. `migrate`). Nach gesundem Stack `last-version`
fortschreiben. Dokumentierter Restore (z. B. `jotti-restore.cmd` oder Doku-Schritt)
spielt einen Dump zurück. Optionaler Export des jüngsten Dumps nach
`%PROGRAMDATA%\jotti\backups\`.

### Acceptance criteria

- [ ] Update vA→vB erzeugt **vor** den Migrationen einen Dump im
      `jotti-backups`-Volume.
- [ ] Gleiche Version erneut starten → **kein** Dump (nur bei Wechsel).
- [ ] Simulierte kaputte Migration → aus dem Dump wiederherstellbar, Daten zurück.
- [ ] Max. N Dumps; ältere werden rotiert.
- [ ] `last-version` wird nur nach gesundem Start fortgeschrieben.

---

## Phase 4: Migrationen ins Image backen (Stack selbst-enthalten)

**Risiken:** 6

### Context

- `database/migrate/Dockerfile` — `COPY` der Migrationen ergänzen.
- `docker-compose.release.yml:50-51` — Bind-Mount entfernen.
- `Makefile:127` (`release-windows`) — `database/migrations` nicht mehr ins ZIP kopieren.

### What to build

Das `jotti-migrate`-Image enthält die Migrationen per `COPY` nach `/migrations`.
Der Bind-Mount entfällt; der Stack braucht den Host-Ordner `database/migrations`
nicht mehr. Das Release-ZIP wird entsprechend schlanker.

### Acceptance criteria

- [ ] Stack startet ohne `database/migrations`-Ordner auf dem Host; Migrationen
      werden trotzdem angewandt.
- [ ] Release-Smoke-Test (Workflow) bleibt grün.
- [ ] ZIP enthält kein `database/migrations` mehr.

---

## Phase 5: Online-Update-Benachrichtigung

**Risiken:** 3

### Context

- `cmd/starter/main.go:40-85` / `printSuccess` — Hinweis nach gesundem Start.

### What to build

Beim Start fragt der Starter (kurzer Timeout, non-fatal) die GitHub-Releases-API
nach der neuesten Version. Ist sie neuer als `version`, erscheint ein deutlicher
Hinweis mit Download-Link. Offline oder Fehler → still überspringen, kein
Zeitverlust, kein Abbruch. **Kein** automatisches Anwenden.

### Acceptance criteria

- [ ] Existiert ein neueres Release, erscheint „Neue Version vX verfügbar: <link>".
- [ ] Offline/timeout → kein Hinweis, kein spürbarer Startverzug, kein Fehler.
- [ ] Aktuelle Version → kein Hinweis.

---

## Phase 6: Doku — „jotti aktualisieren"

**Risiken:** Bediensicherheit

### Context

- `packaging/windows/KURZANLEITUNG.md:58-68` — vor/nach „Beenden/Am nächsten Festtag".

### What to build

Abschnitt **„jotti aktualisieren"**: zuhause mit Internet, `jotti-stop.cmd`, neues
ZIP entpacken (Ort egal — Secret & Daten bleiben), `jotti-start.exe`. Kurz
benennen: automatisches Backup vor dem Update, Restore-Hinweis, und die Garantie
„Daten/Schlüssel/Zertifikate bleiben erhalten". Forward-only/kein Downgrade klar
benennen.

### Acceptance criteria

- [ ] Update-Abschnitt beschreibt Download→Entpacken→Start mit den
      Sicherheitsgarantien (Secret/Daten/Zertifikate bleiben).
- [ ] Automatisches Backup und Restore-Weg erwähnt.
- [ ] Hinweis „kein Downgrade auf ältere Version" enthalten.
