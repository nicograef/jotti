# ADR 13: Go-Modul-Prüfungen und Compose-Stacks bleiben getrennt

- **Status:** akzeptiert (2026-09-09)
- **Kontext-Dokumente:** `docs/plans/plan-jotti-audit-fixes.md` Phase 14 (nach
  Merge gelöscht, siehe Git-Historie); `Makefile`,
  `.github/workflows/ci.yml`, die sieben `docker-compose*.yml`

## Kontext

Zwei Vorschläge zur Entdopplung stehen im Raum: die Go-Modul-Prüfungen in
`Makefile` und CI als Matrix, und die Compose-Dateien als Basisdatei mit
Overrides.

### Go-Modul-Prüfungen

`go.work` führt fünf Module: `backend`, `resolver`, `reverse-proxy`,
`windows/relay`, `windows/starter`. Das `Makefile` hat je Modul ein
`check-*`-Ziel (Zeilen 293–306), `check` ruft alle fünf (Zeile 324). Die vier
Nicht-Backend-Ziele unterscheiden sich untereinander nur im Verzeichnis — mit
einer Ausnahme:

| Ziel                | Zeile | Lint                                           | Build                   |
| ------------------- | ----- | ---------------------------------------------- | ----------------------- |
| `check-backend`     | 294   | zwei Läufe (`--build-tags=integration`/`unit`) | `go build ./...`        |
| `check-relay`       | 297   | ein Lauf                                       | `go build -o /dev/null` |
| `check-starter`     | 300   | ein Lauf                                       | `go build ./...`        |
| `check-resolver`    | 303   | ein Lauf                                       | `go build -o /dev/null` |
| `check-local-proxy` | 306   | ein Lauf                                       | `go build -o /dev/null` |

`check-backend` testet zusätzlich mit `-tags=unit`. `check-starter` legt als
einziges der vier Nicht-Backend-Ziele ein Binary im Arbeitsbaum ab.

In der CI stehen fünf Jobs für dieselben fünf Module: `backend-ci` (57),
`backend-golangci` (104), `resolver-ci` (138), `local-proxy-ci` (185) und
`windows-ci` (232). `windows-ci` ist bereits eine Matrix:
`module: [relay, starter]` (Zeile 238).

Jeder Job trägt eine eigene Auslösebedingung, gebunden an einen eigenen
Pfadfilter des `changes`-Jobs (Filter in den Zeilen 38–55):

| Job              | `if:` (Zeile) | Filter             |
| ---------------- | ------------- | ------------------ |
| `backend-ci`     | 59            | `backend/**`       |
| `resolver-ci`    | 140           | `resolver/**`      |
| `local-proxy-ci` | 187           | `reverse-proxy/**` |
| `windows-ci`     | 234           | `windows/**`       |

`windows-ci` kann eine Matrix sein, weil beide Module unter demselben
Pfadfilter liegen. `resolver` und `reverse-proxy` haben je einen eigenen.
GitHub Actions wertet `jobs.<id>.if` einmal für den ganzen Job aus, bevor die
Matrix expandiert — eine gemeinsame Matrix über die drei Filter hätte also
entweder keine Pfadfilter mehr oder bräuchte eine im `changes`-Job berechnete
Matrix per `fromJSON`.

Weitere Unterschiede zwischen den Jobs: `backend-ci` setzt
`cache-dependency-path: backend/go.sum` (69) und schreibt ein Coverage-Profil
(98–102); `windows-ci` setzt `cache: false` (244); der Lint des Backends liegt
in einem eigenen Job mit zwei `golangci-lint`-Läufen (125–136).

### Compose-Dateien

Sieben Dateien, drei Projektnamen. Der Projektname bestimmt den realen
Volume-Namen — Docker legt ein Volume als `<projekt>_<volume>` an, und keine der
Dateien überschreibt einen Volume-Namen per `name:` oder `external:`:

| Datei                             | Projektname (Zeile) | Volumes (Zeile)                                                                     |
| --------------------------------- | ------------------- | ----------------------------------------------------------------------------------- |
| `docker-compose.yml`              | `jotti-dev` (1)     | `postgres-data`, `frontend-node-modules`, `frontend-pnpm-store` (113)               |
| `docker-compose.e2e.yml`          | `jotti-e2e` (1)     | keine                                                                               |
| `docker-compose.local.yml`        | `jotti-local` (19)  | `postgres-data`, `caddy-data`, `proxy-state`, `jotti-config`, `jotti-backups` (189) |
| `docker-compose.release.yml`      | `jotti-local` (21)  | dieselben fünf (181)                                                                |
| `docker-compose.prod.yml`         | `jotti` (1)         | `postgres-data`, `caddy-data` (184)                                                 |
| `docker-compose.rocks.yml`        | `jotti` (1)         | `postgres-data`, `certbot-challenges`, `letsencrypt`, `acme-dns-data` (325)         |
| `docker-compose.initial-cert.yml` | `jotti` (1)         | `certbot-challenges`, `letsencrypt` (36)                                            |

Zwei Kopplungen laufen ausschließlich über den Projektnamen:

- `docker-compose.release.yml` trägt denselben Namen wie
  `docker-compose.local.yml`, damit ein ZIP-Upgrade dieselben Volumes
  weiterbenutzt — der Dateikopf sagt das ausdrücklich („Same project name
  (jotti-local) and named volumes as the local file on purpose").
- `docker-compose.initial-cert.yml` heißt `jotti`, damit
  `jotti_certbot-challenges` und `jotti_letsencrypt` aus dem Bootstrap in den
  Rocks-Stack übergehen (`scripts/rocks-init.sh:27`).

Wie ähnlich sich die Dateien wirklich sind, sagt `diff`:

| Paar            | geänderte Zeilen |
| --------------- | ---------------- |
| local ↔ release | 54               |
| local ↔ prod    | 117              |
| local ↔ e2e     | 128              |
| prod ↔ rocks    | 208              |
| local ↔ rocks   | 243              |

Nur `local` und `release` sind Zwillinge; die 54 Zeilen sind größtenteils
Kommentarkopf. Sachlich unterscheiden sie sich in vier `build:`/`image:`-Paaren
(migrate, backend, frontend, reverse-proxy), dem Migrations-Bind-Mount und
`JOTTI_ALLOW_SEED`.

`docker-compose.release.yml` ist außerdem ein ausgeliefertes Artefakt: der
`Makefile` kopiert genau diese eine Datei ins Release-ZIP (147) und pinnt darin
den Image-Tag (152). Auf sie zeigen ein kompiliertes Binary
(`windows/starter/main.go:33`) und drei Skripte im Installationsordner
(`packaging/windows/jotti-stop.cmd:7`, `jotti-repair.cmd:18`,
`jotti-restore.cmd:15`).

### Erwogene Alternativen

1. **Eine Matrix über alle Nicht-Backend-Module in der CI.** Kostet entweder die
   Pfadfilter (jeder PR baut alle Module) oder eine im `changes`-Job berechnete
   `fromJSON`-Matrix.
2. **Ein generisches `check-%`-Ziel im `Makefile`.** Die vier Ziele sind nicht
   gleich: `check-backend` hat zwei Lint-Läufe und einen Test-Build-Tag,
   `check-starter` baut ohne `-o /dev/null`. Ein Muster mit vier
   Sonderfall-Variablen ist länger und schwerer zu lesen als vier Zeilen.
3. **Basisdatei plus Overrides für alle Stacks.** Nur ein Paar ist ein
   Zwillingspaar; die übrigen unterscheiden sich in 117 bis 243 Zeilen.
4. **Basisdatei nur für `local` und `release`.** Technisch volume-sicher: bleibt
   `name: jotti-local` in beiden Overrides stehen, heißen die Volumes weiter
   `jotti-local_postgres-data` und so fort. Aber das Release-ZIP müsste zwei
   Dateien ausliefern, und Binary plus drei `.cmd` müssten die zweite kennen.
5. **Alles lassen.**

## Entscheidung

**Die fünf `check-*`-Ziele, die fünf CI-Jobs und die sieben Compose-Dateien
bleiben getrennt** (Alternative 5).

- **Korrektheit:** Kein Vorschlag behebt einen Fehler. Alternative 1 und 3
  können welche einführen: ein verlorener Pfadfilter macht die CI langsamer,
  ein verlorener oder geänderter Projektname benennt jedes Volume um. Ein
  umbenanntes Volume ist leer — Postgres startet auf einer frischen
  Datenverzeichnis-Struktur, und die alten Daten liegen unerreichbar daneben.
- **Einfachheit:** Ein Compose-Aufruf ist heute ein `-f` und eine Datei, die
  vollständig lesbar ist. Basis plus Override heißt: zwei Dateien lesen und die
  Merge-Regeln kennen, um zu wissen, was läuft. Für Betreiber, die den Stack
  einmal im Jahr anfassen, ist das die teurere Form.
- **Konsistenz:** Sieben Dateien in einer Form schlagen fünf in einer und zwei
  in einer anderen. Die vier `check-*`-Ziele sind bereits gleich gebaut.
- **Produkt-Konservatismus:** Die betroffene Fläche ist die
  Installations-Infrastruktur echter Instanzen mit aufbewahrungspflichtigen
  Daten. Eine Entdopplung ohne Nutzerwirkung rechtfertigt dort kein Risiko.

**Empfehlung: nie.**

## Konsequenzen

- Ein sechstes Go-Modul bekommt ein sechstes `check-*`-Ziel und einen sechsten
  CI-Job — oder tritt einer bestehenden Matrix bei, wenn es sich den Pfadfilter
  eines vorhandenen Moduls teilt, wie `windows/relay` und `windows/starter`.
- Eine Änderung an der Prüfkette (etwa eine neue `goimports`-Version) trifft
  fünf `Makefile`-Zeilen und fünf CI-Jobs. Das ist der bewusst getragene Preis.
- Die Abweichung in `check-starter` (`go build ./...` statt
  `go build -o /dev/null ./...`) bleibt bestehen und ist ein eigener kleiner
  Fix, keine Folge dieser Entscheidung.
- Jede neue Compose-Datei trägt ein ausdrückliches `name:`. Ohne die Zeile
  leitet Docker den Projektnamen aus dem Verzeichnisnamen ab, und die Volumes
  eines Betreibers hängen dann daran, wie sein Ordner heißt.
- `docker compose -p …` bleibt Wegwerf-Stacks vorbehalten. Der einzige Gebrauch
  ist `-p jotti-screenshots` im Screenshot-Ziel (`Makefile:364-365`), das den
  Stack mit `down -v` wieder abräumt.
- **Wieder aufgreifen, wenn** ein dritter Stack als Zwilling eines bestehenden
  entsteht, oder wenn die Pfadfilter der CI aus einem anderen Grund entfallen.
  Für die Compose-Seite ist die Vorbedingung härter: ein Vorschlag muss je
  Datei den Projektnamen und die Liste der realen Volume-Namen vorher und
  nachher zeigen, und den Weg für Installationen benennen, die die einzelne
  Release-Datei bereits im Ordner liegen haben.
