# Plan: Repository-Restrukturierung und Hygiene

> Source PRD: n/a (abgeleitet aus der Struktur-Analyse in der Konversation)

## Goal

Die Repo-Topographie aufräumen, ohne Verhalten zu ändern: Go-Module
vereinheitlichen, das konventionswidrige `cmd/`-Verzeichnis umbenennen,
verwaiste Dateien entfernen und tote Doku-Referenzen beseitigen. Jede Phase ist
ein eigenständig mergebarer Commit, der den Baum grün lässt (`make check`, bei
infrastrukturellen Phasen `make verify`).

## Architectural decisions

Durchgängige Festlegungen über alle Phasen:

- Modulpfade: ein einheitliches Präfix `github.com/nicograef/jotti/...` für alle
  fünf Go-Module. Die Module bleiben getrennt (verschiedene Deploy-Targets,
  minimale Abhängigkeiten), gebunden über ein `go.work` im Root.
- Verzeichnis der Windows-Host-Binaries: `cmd/` wird zu `windows/` (enthält
  `windows/starter` und `windows/relay`). Das löst die Kollision mit der
  Go-Konvention, nach der `cmd/` die Einstiegspunkte des Moduls hält.
- Quelle der Wahrheit für Marken-Assets: `assets/` (dokumentiert in
  `assets/assets-and-design.md`). `frontend/public/icons/` und `website/icons/`
  sind abgeleitete Laufzeitkopien.
- Kein Verhaltens- oder API-Change. Reine Struktur, Benennung und Aufräumen.

## Inventory

- `backend/go.mod:1` Modulpfad `github.com/nicograef/jotti/backend` (Referenzstil).
- `resolver/go.mod:1` `jotti-resolver`, `reverse-proxy/go.mod:1`
  `jotti-reverse-proxy`, `cmd/relay/go.mod:1` `jotti-relay`,
  `cmd/starter/go.mod:1` `jotti-starter` (nackte Namen, abweichend).
- Kein `go.work` im Root.
- `Makefile:101,110,113,116,121,122,266,269` referenzieren `cmd/relay` und
  `cmd/starter` (Build, Cross-Compile, Release, Lint-Targets).
- Weitere `cmd/`-Referenzen: `.gitignore`, `README.md`,
  `scripts/prod-backup.sh`, `scripts/prod-update.sh`, `docs/handbuch.md`,
  `docs/plans/plan-vps-paritaet.md`, `docs/language.md`,
  `packaging/windows/KURZANLEITUNG.md`.
- Interne Imports der nackten Modulnamen: `cmd/relay/env.go` (`jotti-relay`),
  `cmd/starter/{main,system,update,backup}.go` (`jotti-starter`);
  `resolver/Dockerfile`, `reverse-proxy/Dockerfile` (Modulname im Build).
- `frontend/src/components/ui/*.tsx`: 56 shadcn-Komponenten, eingeführt en bloc
  in Commit `d103c7f`; 25 davon werden nirgends importiert.
- `frontend/public/vite.svg`, `frontend/public/icons/jotti-icon-dark-64.png`:
  in `index.html`, `manifest.webmanifest` und `src/` nicht referenziert.
- Tote Doku-Referenzen auf gelöschte oder umbenannte Dateien:
  `docs/plans/plan-finanzamt.md` (`EinstellungenPage.tsx`, umbenannt zu
  `FinanzamtPage.tsx`), `docs/prds/prd-tse-setup-wizard.md:3,4,5,136`
  (`plan-tse-fiskaly-fixes.md`, `2026-06-11-tse-fiskaly-audit.md`,
  `prd-tse-integration.md`), `docs/prds/prd-windows-nativ-ohne-docker.md:3`
  (`prd-windows-verpackung.md`), `docs/betrieb/leitfaden-rocks-dns.md:8`
  (`plan-lokale-tls-vertrauenswuerdig.md`),
  `docs/plans/plan-vps-paritaet.md:44,49` (`reverse-proxy/nginx.conf`, in
  Commit `64bac61` durch die Caddy-Migration gelöscht).
- `test-integration.sh` liegt im Root, die übrigen Skripte in `scripts/`.
- `docs/test-01.md`: Scratch-Datei „Notizen aus dem Praxistest" mit
  Platzhalternamen.
- Konvention `AGENTS.md:112`: abgeschlossene Pläne werden nach dem Merge aus
  `docs/plans/` gelöscht. Querverweise auf gelöschte Pläne sind die systemische
  Ursache der toten Links.

## Resolved decisions

- Go-Module: `go.work` hinzufügen und alle Pfade auf
  `github.com/nicograef/jotti/...` vereinheitlichen; die fünf Module bleiben
  getrennt.
- `cmd/` wird zu `windows/` umbenannt.
- Die EN/DE-Umbenennung des Frontends (`products`/`tables`/`users` zu
  `produkte`/`tische`/`benutzer` inkl. Routen) wird aufgeschoben und bleibt als
  Schuld in `docs/language.md` getrackt (siehe Non-Goals).

## Non-Goals

- EN/DE-Frontend-Umbenennung. Dokumentierter SOLL (`docs/language.md:30`:
  Routen `/admin/produkte`, `/service/tische`), aber hohe Churn und
  user-sichtbar. Eigene spätere Phase.
- Geteilte Extraktion von `Produkt.ts` / `Tisch.ts` zwischen Admin und Service.
  Analyse hat ergeben: `Tisch.ts` ist keine Duplikation (Admin
  CRUD-Stammdaten, Service Laufzeit-Session-View, verschiedene Bounded
  Contexts); `Produkt.ts` überlappt nur teilweise (Admin trägt `steuersatz`,
  Service trägt `KategorieLabels`/`KategorieOrder`), die Divergenz ist
  gewollt. Eine erzwungene Shared-Lib würde Admin und Service koppeln.
- Konsolidierung der `docker-compose*.yml`. Läuft über
  `docs/plans/plan-vps-paritaet.md` (Caddy ersetzt `initial-cert.yml`).
- Top-Level-Dateien in `backend/api/` (`admin.go`, `service.go` usw.). Das sind
  die Routing-Kompositionsdateien, akzeptierte Konvention.
- `anforderungen.md:246,338` (`elster-meldung.md`,
  `verfahrensdokumentation.md`). Bewusste Vorausverweise in offenen
  Akzeptanzkriterien auf noch zu erstellende Compliance-Dokumente. Bleiben.

## Open questions / Risks

- `go.work` darf die Docker-Builds nicht stören. Die Dockerfiles bauen je Modul
  aus einem Sub-Kontext, der das Root-`go.work` nicht enthält; der Build nutzt
  dann die modul-eigene `go.mod`. In Phase 2 explizit verifizieren (auch dass
  keine `GOWORK`-Umgebung den per-Modul-Build umlenkt).
- Reihenfolge: das Verzeichnis-Rename (Phase 1) kommt vor der
  Modulpfad-Vereinheitlichung (Phase 2), damit die neuen Pfade die endgültigen
  Verzeichnisse spiegeln.

---

## Phase 1: `cmd/` zu `windows/` umbenennen

### Context

- `Makefile:101,110,113,116,121,122,266,269` Build-, Cross-Compile-, Release-
  und Lint-Targets mit `cmd/relay` und `cmd/starter`.
- `.gitignore` (Einträge `cmd/starter/jotti-start.exe`,
  `cmd/relay/jotti-relay.exe`), `README.md`, `scripts/prod-backup.sh`,
  `scripts/prod-update.sh`, `docs/handbuch.md`, `docs/language.md`,
  `docs/plans/plan-vps-paritaet.md`, `packaging/windows/KURZANLEITUNG.md`.

### What to build

`git mv cmd windows`, sodass `windows/starter` und `windows/relay` entstehen.
Alle Pfad-Referenzen in Makefile, Skripten, `.gitignore`, Doku und Packaging auf
`windows/...` aktualisieren. Modulpfade bleiben in dieser Phase unverändert
(folgen in Phase 2). Die Binärnamen (`jotti-start.exe`, `jotti-relay.exe`)
bleiben gleich.

### Acceptance criteria

- [x] `cmd/` existiert nicht mehr; `windows/starter` und `windows/relay` sind da.
- [x] `grep -rn 'cmd/relay\|cmd/starter'` (ohne `.git`) liefert keine Treffer.
- [x] `make build` und `make release-windows VERSION=test` laufen durch.
- [x] `make check` ist grün.

---

## Phase 2: `go.work` und einheitliche Modulpfade

### Context

- `backend/go.mod:1`, `resolver/go.mod:1`, `reverse-proxy/go.mod:1`,
  `windows/relay/go.mod:1`, `windows/starter/go.mod:1`.
- Interne Imports: `windows/relay/env.go` (`jotti-relay`),
  `windows/starter/{main,system,update,backup}.go` (`jotti-starter`).
- `resolver/Dockerfile`, `reverse-proxy/Dockerfile` (Modulname im Build-Schritt).

### What to build

Ein `go.work` im Root, das alle fünf Module einbindet
(`go work use ./backend ./resolver ./reverse-proxy ./windows/relay
./windows/starter`). Die vier nackten Modulpfade umbenennen auf
`github.com/nicograef/jotti/resolver`, `.../reverse-proxy`, `.../windows/relay`,
`.../windows/starter`. Interne Imports der umbenannten Module sowie die
Dockerfile-Build-Schritte entsprechend anpassen.

### Acceptance criteria

- [x] `go.work` bindet alle fünf Module; `go build ./...` je Modul läuft.
- [x] Alle Modulpfade teilen das Präfix `github.com/nicograef/jotti/`.
- [x] Docker-Builds je Modul (resolver, reverse-proxy, windows-Binaries) bauen
      unverändert; `go.work` beeinflusst sie nicht.
- [x] `make check` ist grün.

---

## Phase 3: Verwaiste Frontend-Dateien entfernen

### Context

- `frontend/src/components/ui/*.tsx`: 25 nie importierte shadcn-Komponenten
  (accordion, aspect-ratio, avatar, breadcrumb, button-group, calendar,
  carousel, chart, collapsible, combobox, command, context-menu, direction,
  form, hover-card, kbd, menubar, navigation-menu, pagination, popover,
  radio-group, resizable, slider, table, toggle-group).
- `frontend/public/vite.svg`, `frontend/public/icons/jotti-icon-dark-64.png`.
- `frontend/components.json` (shadcn-Config) bleibt, damit Komponenten bei
  Bedarf per `pnpm dlx shadcn add <name>` zurückgeholt werden können.

### What to build

Die 25 ungenutzten shadcn-Komponenten, `vite.svg` und das verwaiste
`jotti-icon-dark-64.png` löschen. Vor dem Löschen je Datei quote-agnostisch
prüfen, dass kein Import (`'` oder `"`) darauf zeigt, auch nicht transitiv von
einer anderen `ui/`-Komponente.

### Acceptance criteria

- [x] Kein Import zeigt auf eine entfernte Komponente.
- [x] `make lint-frontend` und `make build-frontend` laufen durch.
- [x] `make test-frontend` ist grün.

---

## Phase 4: Tote Doku-Referenzen und Plan-Hygiene

### Context

- `docs/plans/plan-finanzamt.md` (verweist auf `EinstellungenPage.tsx`;
  Kriterium bei `:200` markiert dessen Entfernung bereits als erledigt).
- `docs/prds/prd-tse-setup-wizard.md:3,4,5,136`,
  `docs/prds/prd-windows-nativ-ohne-docker.md:3`,
  `docs/betrieb/leitfaden-rocks-dns.md:8`,
  `docs/plans/plan-vps-paritaet.md:44,49`.
- Konvention `AGENTS.md:112`.

### What to build

Pro Plan in `docs/plans/` den Checkbox-Status prüfen. Vollständig abgehakte
Pläne löschen (Konvention `AGENTS.md:112`); das beseitigt unter anderem die
`EinstellungenPage.tsx`-Verweise mit `plan-finanzamt.md`. In den überlebenden
dauerhaften Dokumenten (PRDs, Betreiber-Leitfäden) die toten Links auf gelöschte
oder umbenannte Dateien beseitigen: entweder auf den korrekten Nachfolger
umbiegen (`prd-windows-verpackung.md` zu `prd-windows-nativ-ohne-docker.md`,
`EinstellungenPage.tsx` zu `FinanzamtPage.tsx`) oder, wo das Ziel endgültig weg
ist (gelöschter Plan, gelöschtes Audit), zu reinem Fließtext ohne
Datei-Verweis umformulieren. `plan-vps-paritaet.md` Inventory: den bereits
gelöschten `reverse-proxy/nginx.conf` als entfernt markieren.

Die Vorausverweise in `docs/anforderungen.md:246,338` bleiben unangetastet.

### Acceptance criteria

- [x] Grep über `docs/` findet keine Referenz mehr auf eine nicht existierende
      Repo-Datei, mit Ausnahme der annotierten Vorausverweise in
      `anforderungen.md`.
- [x] Vollständig abgehakte Pläne sind aus `docs/plans/` entfernt.
- [x] Kein Betreiber-Leitfaden (`docs/betrieb/`) verlinkt auf einen internen
      Plan.

---

## Phase 5: Datei-Platzierung und Asset-Quelle

### Context

- `test-integration.sh` (Root) gegenüber `scripts/*.sh`; `Makefile`-Target
  `test-integration`.
- `assets/` mit `assets/assets-and-design.md`; `frontend/public/icons/`,
  `website/icons/` als Laufzeitkopien.
- `docs/test-01.md` (zwei offene Praxistest-Befunde).

### What to build

Drei kleine, voneinander unabhängige Aufräum-Schritte (je ein Commit möglich):

1. `test-integration.sh` nach `scripts/` verschieben und die Referenz im
   `Makefile` aktualisieren.
2. `assets/` in `assets-and-design.md` als Quelle der Wahrheit dokumentieren und
   `frontend/public/icons/` plus `website/icons/` als abgeleitete Kopien
   benennen. Optional, nur falls trivial: ein `make sync-assets`-Target, das aus
   `assets/` kopiert. Kein Build-Automatismus erzwingen (kein Gold-Plating).
3. `docs/test-01.md` auflösen: die beiden Befunde nach `docs/anforderungen.md`
   oder in das Issue-Tracking überführen und die Datei löschen, oder, falls sie
   bleibt, aussagekräftig benennen (z. B. `docs/praxistest-2026-06-13.md`).

### Acceptance criteria

- [x] `test-integration.sh` liegt in `scripts/`, `make test-integration` läuft.
- [x] Asset-Quelle ist in `assets-and-design.md` dokumentiert.
- [x] `docs/test-01.md` ist aufgelöst (verschoben/umbenannt/migriert).
- [x] `make verify` ist grün.
