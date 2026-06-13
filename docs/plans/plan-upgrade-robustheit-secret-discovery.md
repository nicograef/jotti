# Plan: Upgrade-Robustheit & Secret-Discovery

> Source PRD: n/a (aus Vorfallanalyse v0.9.0 → v0.10.0)
> Nachfolger von `plan-windows-update-robustheit.md`. Jener Plan führte das
> `jotti-config`-Volume (Phase 1) und das `%PROGRAMDATA%\jotti`-Zustandsverzeichnis
> (Phase 2) ein — **dieser Plan schließt die Lücke, die deren Kombination beim
> Upgrade von einer Vor-Phase-2-Version aufgerissen hat.**

## Goal

Kein plausibler Upgrade-Pfad darf mehr Daten verlieren oder aus ihnen aussperren —
auch nicht das Upgrade von einer alten Version, deren Secret an einem anderen Ort
lag. Konkret: das Install-Secret wird **überall gefunden**, wo eine frühere
Version es abgelegt haben kann; wird **nichts** gefunden, obwohl Daten existieren,
**bricht der Start mit Anleitung ab** statt still ein neues (= falsches) Secret zu
erzeugen. Bereits feststeckende Installationen kommen über ein geführtes
Reparatur-Werkzeug datenerhaltend zurück.

## Grundproblem (Reframe)

> **Das Secret-Bootstrap suchte nur am neuen kanonischen Ort — nicht dort, wo
> ältere Versionen es tatsächlich hingeschrieben haben.**

Belegter Ablauf des Vorfalls:

- **v0.9.0** (vor Phase 1/2) schrieb die `.env` **ordnerlokal neben die Exe**
  (`filepath.Join(filepath.Dir(composePath), ".env")`), z. B.
  `C:\Program Files\jotti\.env`. Dort lagen `POSTGRES_PASSWORD` und `JWT_SECRET`.
- **v0.10.0** (mit Phase 1/2) liest den State aus `%PROGRAMDATA%\jotti\` — ein
  **anderer Ordner**. [ResolveEnv](../../cmd/starter/core/env.go) bekam einen
  leeren `localContent` (die alte `.env` lag woanders) und ein leeres Volume
  (gab es bei v0.9.0 noch nicht) → es erzeugte **frische Secrets**.
- Folge: neues `POSTGRES_PASSWORD` ≠ `postgres-data` → Migrate-Lockout; neues
  `JWT_SECRET` → alle Tokens ungültig. Beide Symptome aus **einem** Start. Das
  Daten-Volume selbst blieb durch den Code unangetastet (gleicher Projektname,
  gleiche Volume-Namen — verifiziert im Diff `v0.9.0..v0.10.0`).

Die „adoptiere lokale `.env`"-Verzweigung in `ResolveEnv` war genau für dieses
Upgrade gedacht, lief aber ins Leere, weil sie nur den **einen** neuen Pfad las.

## Architectural decisions

Durable Entscheidungen über alle Phasen:

- **Secret-Suchreihenfolge (Discovery):** `jotti-config`-Volume → Host-`.env`
  unter `%PROGRAMDATA%\jotti` → ordnerlokale `.env` neben Compose/Exe. Erster
  nicht-leerer Treffer gewinnt; ein adoptierter Treffer (nicht aus dem Volume)
  wird ins Volume geschrieben und damit zur Vorwärts-Quelle der Wahrheit. Pattern
  analog zu [relay `envSearchDirs`](../../cmd/relay/env.go) (Konsistenz zwischen
  den beiden Datei-Lesern).
- **Fail-Safe:** Wird **kein** Secret gefunden **und** existiert bereits ein
  `postgres-data`-Volume → Start **abbrechen** mit deutscher Anleitung (Suchorte
  nennen). **Nie** frische Secrets neben vorhandenen Daten erzeugen. Frische
  Secrets ausschließlich bei echter Erstinstallation (kein `postgres-data`).
- **Recovery-Mechanismus:** Das DB-Rollen-Passwort wird über den **lokalen
  Trust-Socket** (`docker exec … psql`, kein Passwort nötig) an das Install-Secret
  **angeglichen** (`ALTER USER admin PASSWORD …`) — ohne das alte Passwort zu
  kennen, datenerhaltend, idempotent. Ausgeliefert als `jotti-repair.cmd`
  (konsistent zur `jotti-stop.cmd`/`jotti-restore.cmd`-Konvention).
- **Backup-Trigger erweitern:** Sichern, wenn Daten existieren **und** (Version
  gewechselt **oder** kein `last-version`-Marker vorhanden). So ist auch das erste
  Upgrade von einer Vor-Phase-3-Version abgedeckt.
- **Volume-Labels:** Der Starter legt `jotti-config` selbst an (das Secret muss
  vor `compose up` hinein), versieht es aber mit den Compose-Labels
  (`com.docker.compose.project=jotti-local`, `com.docker.compose.volume=jotti-config`).
  **Kein** `external: true` — das würde `down -v` daran hindern, das Volume mit den
  Daten zu entfernen (und damit die Lockout-Garantie aus Phase 1 brechen).
- **Plattform:** OS-spezifische Schritte nur unter `runtime.GOOS == "windows"`;
  der Linux-Dev-Lauf (`go run`) bleibt inert und grün (Version `"dev"`, ordnerlokal).

## Inventory

- `cmd/starter/core/env.go:67-75` — `ResolveEnv(volumeContent, localContent)`:
  zwei-Quellen-Entscheidung; muss auf **mehrere** lokale Kandidaten erweitert
  werden und den Fail-Safe-Fall kennen.
- `cmd/starter/core/env.go:15-20` — `StateDir`: kanonischer Ort
  `%PROGRAMDATA%\jotti`; liefert den Ausgangspunkt der Suchreihenfolge.
- `cmd/starter/core/env.go:27-37` — `EnvContent`: erzeugt alle drei frischen
  Secrets; genau die Neugenerierung, die der Fail-Safe verhindern muss.
- `cmd/starter/system.go:267-281` — `materializeEnvFromVolume`: liest heute nur
  Volume + `envPath`; zentrale Stelle für Discovery + Fail-Safe.
- `cmd/starter/system.go:250-259` — Konstanten `configVolume`/`configVolumePath`/
  `configHelperImage`; Basis für das gelabelte Anlegen.
- `cmd/starter/main.go:55-70` — `resolveStateDir`/`envPath`/Aufruf von
  `materializeEnvFromVolume`; hier wird die Kandidatenliste gebaut.
- `cmd/relay/env.go:63-75` — `envSearchDirs(goos, programData, exeDir, wd)`:
  bestehende Mehr-Ort-Suche; Vorlage für die Starter-Discovery.
- `cmd/starter/core/backup.go:12-23` — `ShouldBackup(lastVersion, currentVersion,
  postgresDataExists)`: heute `false` bei leerem `lastVersion`; Phase 2 erweitert.
- `cmd/starter/backup.go:39-68` — `maybeBackupBeforeUpdate`; ruft `ShouldBackup`.
- `cmd/starter/backup.go:81-95` — `volumeExists`; wiederverwendbar für die
  `postgres-data`-Prüfung im Fail-Safe.
- `docker-compose.release.yml:149-156` — Volume-Block (`jotti-config` deklariert,
  read-only in postgres gemountet); Labels werden host-seitig beim Anlegen gesetzt.
- `packaging/windows/jotti-stop.cmd`, `.../jotti-restore.cmd` — `.cmd`-Konvention
  für Helfer-Werkzeuge; Vorlage für `jotti-repair.cmd`.
- `Makefile:118-125` — `release-windows`: packt die `.cmd`-Dateien ins ZIP;
  `jotti-repair.cmd` ergänzen.
- `packaging/windows/KURZANLEITUNG.md` — enthält bereits einen Update-Abschnitt
  (aus `plan-windows-update-robustheit.md` Phase 6); um Recovery + `down -v`-Warnung
  erweitern.

## Resolved decisions

- **Umfang:** Prävention + Recovery + Backup/Doku (volle Härtung).
- **Suchorte:** Volume + `%PROGRAMDATA%\jotti` + ordnerlokal neben der Exe.
- **Fail-Safe-Verhalten:** Abbrechen mit Anleitung (nie still neu erzeugen neben
  vorhandenen Daten).
- **Erst-Backup:** Ja — bei vorhandenem `postgres-data` ohne Marker vor den
  Migrationen sichern.

## Open questions / Risks

- **`jotti-repair.cmd` setzt das DB-Passwort auf das *aktuelle* Install-Secret.**
  Damit bleibt ein evtl. zwischenzeitlich neu erzeugtes `JWT_SECRET` aktiv → alte
  Tokens bleiben ungültig, einmal neu einloggen. Bewusst akzeptiert (Daten zählen,
  Tokens sind kurzlebig).
- **Bereits gelabelte/ungelabelte Bestands-Volumes:** Labels sind nach dem Anlegen
  unveränderlich. Der Code-Fix wirkt für **neue** Installationen; bestehende
  Volumes verlieren die Warnung erst nach einmaligem Neu-Anlegen (Doku-Hinweis).
- **Linux-Dev muss grün bleiben:** Der erweiterte Backup-Trigger darf im Dev-Lauf
  (Version `"dev"`) nicht feuern; Discovery/Fail-Safe sind Windows-gated.
- **Discovery nach UAC-Elevation:** Arbeitsverzeichnis ist dann `System32` — die
  ordnerlokale Suche muss über `filepath.Dir(composePath)` und das Exe-Verzeichnis
  laufen, nicht über das Arbeitsverzeichnis (wie `resolveComposeFile`).

---

## Phase 1: Robuster Secret-Bootstrap (Discovery + Fail-Safe)

**Risiken:** Lockout + JWT-Reset beim Upgrade von einer Vor-Phase-2-Version.

### Context

- `cmd/starter/core/env.go:67-75` — `ResolveEnv` von zwei Quellen auf eine
  geordnete Kandidatenliste + Fail-Safe-Signal erweitern.
- `cmd/starter/system.go:267-281` — `materializeEnvFromVolume`: Kandidaten lesen,
  Ergebnis interpretieren, ggf. abbrechen.
- `cmd/starter/main.go:55-70` — Kandidaten-Verzeichnisse bauen (Compose-Dir,
  Exe-Dir, `%PROGRAMDATA%\jotti`).
- `cmd/relay/env.go:63-75` — `envSearchDirs` als Vorlage für die Reihenfolge.
- `cmd/starter/backup.go:81-95` — `volumeExists` für die `postgres-data`-Prüfung.

### What to build

Der Starter ermittelt das Install-Secret aus der ersten nicht-leeren Quelle in
fester Reihenfolge: `jotti-config`-Volume → `%PROGRAMDATA%\jotti\.env` →
ordnerlokale `.env` neben Compose/Exe. Ein Treffer, der nicht aus dem Volume kam,
wird ins Volume geschrieben und damit zur Vorwärts-Quelle der Wahrheit; die
Host-`.env` wird gespiegelt. Wird **keine** Quelle fündig, entscheidet das
Vorhandensein von `postgres-data`: existiert es, **bricht der Start mit klarer
deutscher Anleitung ab** (welche `.env` wohin gehört) und verändert **nichts**;
existiert es nicht, ist es eine echte Erstinstallation und es werden frische
Secrets erzeugt (wie bisher). Die reine Auswahl-/Fail-Safe-Logik liegt testbar in
`core`, die Datei-/Docker-Zugriffe im Starter.

### Acceptance criteria

- [x] Upgrade-Layout mit `.env` **neben der Exe** (Vor-Phase-2-Fall) → Secret wird
      adoptiert, **keine** Neugenerierung, kein Lockout; Volume wird mit dem
      adoptierten Secret befüllt.
- [x] Secret nur im `%PROGRAMDATA%\jotti\.env` → adoptiert; Secret nur im Volume →
      übernommen (unverändert, kein erneutes Seeden).
- [x] `postgres-data` vorhanden, aber an **keinem** Ort ein Secret → Start bricht
      mit Anleitung ab, erzeugt/ändert nichts.
- [x] Echte Erstinstallation (kein `postgres-data`, kein Secret) → frische Secrets
      wie bisher.
- [x] Linux-Dev-Lauf (`go run`) unverändert grün; `core`-Logik unit-getestet
      (Reihenfolge, Adoption, Fail-Safe).

---

## Phase 2: Pre-Update-Backup auch beim Erst-Upgrade

**Risiken:** Kaputte Migration beim ersten Upgrade ohne Sicherungspunkt.

### Context

- `cmd/starter/core/backup.go:12-23` — `ShouldBackup`: Bedingung um „Daten
  vorhanden, aber kein Marker" erweitern.
- `cmd/starter/backup.go:39-68` — `maybeBackupBeforeUpdate`: nutzt das erweiterte
  Prädikat; restliche Sequenz (postgres-only up, Dump, Rotation) bleibt.

### What to build

`ShouldBackup` löst zusätzlich aus, wenn ein `postgres-data`-Volume existiert,
aber **kein** `last-version`-Marker vorliegt (unbekannte → bekannte Version =
Erst-Upgrade). Damit zieht der Starter auch beim ersten Upgrade von einer
Vor-Phase-3-Version vor den Migrationen einen Dump. Die echte Erstinstallation
(kein `postgres-data`) bleibt ohne Backup, der Linux-Dev-Lauf (Version `"dev"`)
bleibt inert.

### Acceptance criteria

- [x] `postgres-data` vorhanden + **kein** Marker → Dump vor den Migrationen.
- [x] `postgres-data` vorhanden + Marker == Version → **kein** Dump.
- [x] Echte Erstinstallation (kein `postgres-data`) → **kein** Dump.
- [x] Marker wird wie bisher nur nach gesundem Start fortgeschrieben.
- [x] `core`-Logik unit-getestet; Linux-Dev-Lauf grün.

---

## Phase 3: Geführte Wiederherstellung (`jotti-repair.cmd`)

**Risiken:** Bereits feststeckende Installationen (Volume-Secret ≠
`postgres-data`-Passwort) ohne datensicheren Weg zurück.

### Context

- `packaging/windows/jotti-restore.cmd` — `.cmd`-Konvention (postgres-only up,
  `docker exec`, voller Neustart) als Vorlage.
- `cmd/starter/system.go:250-259` — Container-/Volume-Namen; `jotti-postgres-local`,
  lokaler Trust-Socket.
- `Makefile:118-125` — Packaging der `.cmd`-Dateien.

### What to build

Ein `jotti-repair.cmd` für den Fall „Daten da, aber Passwort passt nicht": Es
fährt nur `postgres` hoch, liest das aktuelle Install-Secret (Host-`.env` =
Volume-Wert) und gleicht über `docker exec … psql` (lokaler Trust, kein Passwort
nötig) das Rollen-Passwort daran an (`ALTER USER admin PASSWORD …`), dann voller
Neustart. Datenerhaltend, idempotent, mehrfach ausführbar. Das Skript wird ins
Release-ZIP gepackt. Es kennt das alte Passwort nicht und braucht es nicht.

### Acceptance criteria

- [ ] Installation mit Volume-Secret ≠ `postgres-data`-Passwort → nach `jotti-repair`
      authentifiziert `migrate`/`backend`, Daten unverändert vorhanden, Login geht.
- [ ] Erneutes Ausführen ist gefahrlos (idempotent), auch wenn bereits konsistent.
- [ ] `jotti-repair.cmd` liegt im Release-ZIP (`release-windows`).
- [ ] Skript verändert **keine** Daten (nur das Rollen-Passwort) und fasst keine
      anderen Volumes an.

---

## Phase 4: `jotti-config`-Volume mit Compose-Labels anlegen

**Risiken:** Verwirrende „volume … not created by Docker Compose"-Warnung bei jedem
Start (schlechte UX für nicht-technische Helfer).

### Context

- `cmd/starter/system.go:306-314` — `writeConfigVolume`: legt das Volume heute
  implizit per `docker run -v` an (ohne Labels).
- `cmd/starter/backup.go:81-95` — `volumeExists` für die Idempotenz-Prüfung.
- `docker-compose.release.yml:149-156` — Volume bleibt **nicht** `external`.

### What to build

Bevor der Starter zum ersten Mal ins `jotti-config`-Volume schreibt, legt er es —
falls fehlend — explizit mit den Compose-Labels an
(`com.docker.compose.project=jotti-local`, `com.docker.compose.volume=jotti-config`).
Compose erkennt es dann als eigenes Volume (keine Warnung), und `down -v` entfernt
es weiterhin zusammen mit den Daten. Bestehende Volumes bleiben unangetastet
(Labels sind unveränderlich) — der Doku-Hinweis nennt den einmaligen Neu-Anlege-Weg.

### Acceptance criteria

- [ ] Frische Installation → **keine** „not created by Docker Compose"-Warnung.
- [ ] `down -v` entfernt `jotti-config` weiterhin (Lifecycle/Lockout-Garantie
      erhalten — gegen ein gelabeltes Volume verifiziert).
- [ ] `external: true` wird **nicht** verwendet.
- [ ] Anlegen ist idempotent; ein vorhandenes Volume wird nicht neu erzeugt.

---

## Phase 5: Doku „sicher aktualisieren & wiederherstellen"

**Risiken:** Bediensicherheit — Helfer wenden gefährliche Schritte an (`down -v`),
kennen den Recovery-Weg nicht.

### Context

- `packaging/windows/KURZANLEITUNG.md` — bestehender Update-Abschnitt; um Recovery
  und Warnungen erweitern.

### What to build

Der Update-/Betriebs-Abschnitt der KURZANLEITUNG wird ergänzt: irgendwohin
entpacken ist sicher (das Secret folgt den Daten via Volume); **niemals
`docker compose down -v`** ausführen (löscht Daten **und** Schlüssel); bei Lockout
nach einem Upgrade von einer sehr alten Version `jotti-repair.cmd` ausführen bzw.
die alte `.env` aus dem alten Programmordner nach `%PROGRAMDATA%\jotti\.env` legen;
explizite Garantie „Daten/Schlüssel/Zertifikate bleiben erhalten". Einmaliger
Hinweis zum Entfernen der Compose-Warnung bei Bestandsinstallationen.

### Acceptance criteria

- [ ] Abschnitt beschreibt sicheres Update (entpacken egal, Secret folgt Volume).
- [ ] Klare Warnung vor `down -v` mit Begründung.
- [ ] Recovery-Weg (`jotti-repair.cmd` / alte `.env` an den kanonischen Ort) erklärt.
- [ ] Garantie „Daten/Schlüssel/Zertifikate bleiben" benannt.
