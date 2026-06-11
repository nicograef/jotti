# Plan: Agent-Setup-Bereinigung (Copilot, Claude Code, Skills)

> Source PRD: n/a — abgeleitet aus dem Code-Audit des Agent-Setups (Chat, 2026-06-11)

## Goal

Das AI-Agent-Setup (GitHub Copilot, Claude Code, Handbook-Skills) reduzieren,
vereinfachen und in der Qualität verbessern: eine gefährliche Auto-Approval
entfernen, die Permission-Posture kohärent machen, duplizierte und
widersprüchliche Instruktionen abbauen und die Skill-Familie konsolidieren.

Die Arbeit verteilt sich auf drei Orte:

| Ort | Inhalt | Versioniert |
| --- | --- | --- |
| jotti-Repo | `AGENTS.md`, `.github/copilot-instructions.md`, `.github/instructions/`, `.claude/settings.local.json` | ja (dieses Repo) |
| Maschine | `~/.claude/settings.json`, `~/.claude/CLAUDE.md`, `~/.claude/statusline.sh`, `~/.config/Code/User/settings.json` | nein — vor Änderung Kopie nach `~/.claude/backups/` |
| handbook-Repo | `skills/` (cleanup, deslop, code-audit, README) | ja (`~/r/handbook`) |

## Architectural decisions

Durable Entscheidungen, auf die sich alle Phasen beziehen:

- **Permission-Posture: Guarded.** Die globale Claude-Code-Allowlist enthält
  keine Befehle mit beliebiger Code-Ausführung oder freiem Netzwerkzugriff
  (`node`, `npx`, `pnpm exec`, `go run`, `curl`, `http`). Solche Befehle
  prompten. Die Deny-Liste bleibt bestehen und ist damit wieder wirksam.
- **Skill-Familie: zwei Entry-Points.** `cleanup` ist der einzige Entry-Point
  für Diff-/Bereichs-Reviews (inkl. Slop-Erkennung); `code-audit` ist
  ausschließlich repo-weit/cross-layer. `deslop` entfällt als eigenständiger
  Entry-Point; seine Referenzdateien wandern nach `cleanup/`.
- **Skill-Distribution: maschinenweit über `~/.claude/skills`.** Skills leben
  im handbook-Repo und sind als `~/.claude/skills` symlinkt. Sowohl Claude
  Code als auch GitHub Copilot (VS Code) konsumieren sie von dort. Kein
  Kopieren in Projekt-Repos.
- **Instruktionshierarchie bleibt.** `AGENTS.md` kanonisch, dünner
  Copilot-Spiegel, bereichsspezifische `*.instructions.md` — unverändert.
- **Lizenz-Terminologie:** In Agent-Instruktionen heißt jotti
  „Source-Available (nicht-kommerziell)", nicht „quelloffen"/„open-source".

## Inventory

**Maschine:**

- `~/.config/Code/User/settings.json:48-66` — `chat.tools.terminal.autoApprove`
  mit `"rm": true` (Z. 58), totem Pseudo-Regex `".*Load.*\""` (Z. 59) und zwei
  veralteten cloudevents-Einmal-Approvals (Z. 50, 54). Achtung: JSONC mit
  trailing commas — nicht mit striktem `jq` validieren.
- `~/.claude/settings.json:3-74` — Allowlist inkl. `go run`, `node`,
  `pnpm exec`, `npx`, `http`, `curl`; Deny-Liste Z. 75-81.
- `~/.claude/CLAUDE.md:42-53` — jotti-spezifische TS/React- und
  Go-Konventionen als globale Regeln; Z. 62 beschreibt jotti als „open-source".
- `~/.claude/statusline.sh:11-15` — Modellkürzung kennt nur
  `opus|sonnet|haiku`; Z. 18-21 Git-Check mit redundantem erstem Zweig und
  Präzedenz-Quirk (`&&` bindet stärker als `||`).
- `~/.claude/backups/` — existiert, Ziel für Sicherungskopien.

**jotti-Repo:**

- `AGENTS.md:3` — „kostenloses, quelloffenes Mobile-Kassensystem" neben
  „proprietäre Source-Available-Lizenz" (Z. 5).
- `AGENTS.md:64-78` (Wichtige Regeln) vs. `AGENTS.md:103-116` (Grenzen) —
  sechs Constraints doppelt (POST-only, Cent-Beträge, Kassenjournal
  append-only, keine `json`-Tags in `domain/`, kein direktes `fetch()`,
  `sqlc/dbgen/` nie editieren).
- `.github/copilot-instructions.md:3` — „quelloffenes".
- `.github/instructions/backend.instructions.md:57-106` — DTO-Beispiel zeigt
  dieselbe Lektion zweimal (`varianteDTO` Z. 60-78, `produktDTO` Z. 80-105).
- `.github/instructions/event-sourcing.instructions.md:15` und `:44-51` —
  Projektion/`StreamType`-Routing doppelt beschrieben.
- `.claude/settings.local.json:4` — Read-Permission auf `zog@v0.22.0`
  versionsgepinnt.

**handbook-Repo (`~/r/handbook`):**

- `skills/cleanup/SKILL.md` (157 Z.) — referenziert bereits Slop-Patterns;
  eigene Referenzen: `principles.md`, `code-smells.md`, `architecture.md`,
  `readability.md`, `readability-de.md`.
- `skills/deslop/SKILL.md` (98 Z.) — Referenzen: `code.md`, `text.md`,
  `config.md`, `text-de.md`.
- `skills/code-audit/SKILL.md` (59 Z.) — Description überlappt mit cleanup
  („readability review").
- `skills/README.md:3` — „copy individual skill directories into project
  repos as needed" (stimmt nicht mit der tatsächlichen Nutzung überein);
  Tabelle Z. 7-24 listet deslop als eigene Zeile.
- `skills/README.md:18-20, 24, 35` — tote Verweise auf die am 2026-06-11
  gelöschten Skills (design-interface, improve-architecture, handbook-sync,
  extract; handbook-Commit `2f1ccdb`).

## Resolved decisions

- **Scope:** Ein Plan für alle drei Orte (jotti, Maschine, handbook).
- **Permission-Posture:** Guarded (siehe Architectural decisions).
- **Skill-Konsolidierung:** deslop → cleanup mergen; Referenzdateien bleiben
  erhalten; code-audit nur in der Description rescopen.
- **Distribution:** Kein „copy into repos". Skills werden von beiden Agents
  aus `~/.claude/skills` gelesen (Symlink auf handbook). Nur Doku fixen.
- **Kein manuelles Aufräumen von Caches/Sessions/Workspaces/Memory.** Bewusst
  ausgeschlossen: Claude Code räumt eigenen State automatisch auf
  (`.last-cleanup` 2026-06-11, Transcript-Retention per Default 30 Tage über
  `cleanupPeriodDays` steuerbar), die `~/.claude`-Verzeichnisse sind winzig
  (< 2 MB), VS Codes `workspaceStorage` (324 MB) regeneriert sich nach dem
  Löschen. Stattdessen adressiert Phase 8 veraltete *Inhalte*, die Agents
  tatsächlich als Kontext lesen.

## Open questions / Risks

- **Maschinendateien sind unversioniert.** Vor jeder Änderung Kopie nach
  `~/.claude/backups/` bzw. für VS Code-Settings eine `.bak`-Datei daneben.
- **Guarded-Posture kann Reibung erzeugen.** Falls in den Folgetagen häufige
  Prompts für einen konkreten Befehl auftreten, gezielt einen engen
  Allow-Eintrag ergänzen (z. B. `Bash(npx vitest:*)`) statt den breiten
  zurückzuholen.
- **Copilot-Skill-Discovery verifizieren.** Annahme (vom User bestätigt):
  VS Code Copilot liest Skills aus `~/.claude/skills`. In Phase 7 einmal
  praktisch gegenprüfen, bevor das README es behauptet.
- **Deutsche Slop-Referenzen überlappen.** `cleanup/readability-de.md` und
  `deslop/text-de.md` decken ähnliches ab — beim Merge (Phase 6) auf
  Redundanz prüfen und ggf. zusammenführen.

---

## Phase 1: Copilot Terminal-Auto-Approve entschärfen (Maschine)

### Context

- `~/.config/Code/User/settings.json:48-66` — `chat.tools.terminal.autoApprove`

### What to build

Die Auto-Approve-Map von gefährlichen und toten Einträgen befreien:
`"rm": true` entfernen (Copilot darf nie unbestätigt löschen), den
wirkungslosen Pseudo-Regex-Eintrag `".*Load.*\""` entfernen, die beiden
veralteten cloudevents-Kommandozeilen-Approvals entfernen. Die produktiven
Einträge (`go`, `make`, `java`, `./mvnw`, `mvn clean`, `docker build`, `sdk`)
bleiben.

### Acceptance criteria

- [x] `"rm": true` ist entfernt
- [x] Pseudo-Regex-Eintrag und beide cloudevents-Einträge sind entfernt
- [x] Sicherungskopie der Datei existiert
- [x] VS Code lädt die Settings ohne Fehler (JSONC bleibt parsebar)

---

## Phase 2: Claude-Code-Permissions auf Guarded umstellen (Maschine)

### Context

- `~/.claude/settings.json:3-74` — Allowlist; `:75-81` — Deny-Liste

### What to build

Die sechs Einträge mit beliebiger Code-Ausführung bzw. freiem Netzwerkzugriff
aus der Allowlist entfernen: `Bash(go run:*)`, `Bash(node:*)`,
`Bash(npx:*)`, `Bash(pnpm exec:*)`, `Bash(http:*)`, `Bash(curl:*)`.
Deny-Liste unverändert lassen. Der tägliche Workflow (`make:*`,
`pnpm run:*`, `go test:*`, `go build:*`, Read-only-Git/Docker) bleibt
prompt-frei.

### Acceptance criteria

- [ ] Die sechs Einträge sind entfernt, alle übrigen unverändert
- [ ] `jq empty ~/.claude/settings.json` ist erfolgreich
- [ ] Sicherungskopie in `~/.claude/backups/` existiert
- [ ] Stichprobe in einer Claude-Code-Session: `make check` läuft ohne
      Permission-Prompt, `curl example.com` prompted

---

## Phase 3: Globalen Claude-Kontext und Statusline verschlanken (Maschine)

### Context

- `~/.claude/CLAUDE.md:42-53` — projektspezifische Konventionen; `:61-65` —
  Projektliste
- `~/.claude/statusline.sh:11-15, 18-21` — Modellkürzung und Git-Check

### What to build

`~/.claude/CLAUDE.md` auf echt projektübergreifenden Inhalt kürzen:
Identität, Tooling-/Infra-Präferenzen, Commit-Stil (Conventional Commits,
Englisch, kein Auto-Commit), Projektliste. Die TS/React-, Go- und
Java-Konventionsblöcke entfernen — sie sind in den jeweiligen Repo-AGENTS.md
kanonisch (jotti: Regeln 1, 7, 8). Die jotti-Beschreibung von „open-source"
auf „source-available (non-commercial)" korrigieren.

`statusline.sh` vereinfachen: Git-Check auf den `rev-parse`-Test reduzieren
(erster Zweig redundant, Präzedenz-Quirk entfällt); Modellkürzung entweder
streichen (Anzeigename ist kurz genug) oder um `fable` ergänzen.

### Acceptance criteria

- [ ] CLAUDE.md enthält keine framework-/architekturspezifischen Regeln mehr
      (kein `BackendClient`, kein „POST-only", keine sqlc/Flyway-Details)
- [ ] jotti ist als source-available beschrieben
- [ ] Statusline rendert korrekt für Fable- und Opus-Modelle, in und außerhalb
      von Git-Repos (Test per Pipe mit Beispiel-JSON)
- [ ] Sicherungskopien beider Dateien existieren

---

## Phase 4: Lizenz-Terminologie in jotti-Agent-Instruktionen fixen (jotti)

### Context

- `AGENTS.md:3` — „kostenloses, quelloffenes Mobile-Kassensystem"
- `.github/copilot-instructions.md:3` — „quelloffenes Mobile-Kassensystem"

### What to build

Beide Stellen auf die korrekte Terminologie umstellen, z. B. „kostenloses
Mobile-Kassensystem mit offenem Quellcode (Source-Available,
nicht-kommerziell)". Der Lizenz-Satz in `AGENTS.md:5` bleibt unverändert —
nur der Widerspruch im Eröffnungssatz wird beseitigt.

### Acceptance criteria

- [ ] `grep -ri quelloffen AGENTS.md .github/` liefert keine Treffer mehr
- [ ] Beide Dateien beschreiben die Lizenz konsistent als Source-Available
- [ ] Keine inhaltlichen Änderungen über die Terminologie hinaus

---

## Phase 5: jotti-Instruktionen deduplizieren und reduzieren (jotti)

### Context

- `AGENTS.md:64-78` (Wichtige Regeln) und `:103-116` (Grenzen) — sechs
  doppelte Constraints
- `.github/instructions/backend.instructions.md:57-106` — doppeltes
  DTO-Beispiel
- `.github/instructions/event-sourcing.instructions.md:15, 44-51` — doppelte
  Projektions-/Routing-Beschreibung
- `.claude/settings.local.json:4` — versionsgepinnte zog-Permission

### What to build

„Wichtige Regeln" und „Grenzen" zu einer Sektion zusammenführen: nummerierte
Regeln mit ✅/⚠️/🚫-Markern inline; die drei Nicht-Duplikate aus „Grenzen"
(Nachfragen vor neuen Dependencies/Docker-Änderungen, keine Secrets,
`make sqlc`/`make lint`-Reminder) bleiben erhalten. Im Backend-Instructions-
DTO-Beispiel nur das verschachtelte `produktDTO`-Beispiel behalten. In den
Event-Sourcing-Instructions die Projektions-/`StreamType`-Beschreibung nur
einmal führen. Die zog-Permission auf `Read(//home/nico/go/pkg/mod/**)`
verallgemeinern.

### Acceptance criteria

- [ ] Jeder Constraint steht genau einmal in `AGENTS.md`; keine Regel ist
      inhaltlich verloren gegangen
- [ ] `AGENTS.md` ist um ≥ 20 Zeilen kürzer
- [ ] `backend.instructions.md` zeigt genau ein DTO-Mapping-Beispiel
- [ ] `event-sourcing.instructions.md` beschreibt Projektion/Routing einmal
- [ ] `.claude/settings.local.json` ist nicht mehr versionsgepinnt

---

## Phase 6: Skill-Familie konsolidieren — deslop → cleanup (handbook)

### Context

- `~/r/handbook/skills/cleanup/SKILL.md` — künftiger einziger Entry-Point für
  Diff-/Bereichs-Reviews
- `~/r/handbook/skills/deslop/` — `SKILL.md` entfällt; `code.md`, `text.md`,
  `config.md`, `text-de.md` wandern nach `cleanup/`
- `~/r/handbook/skills/code-audit/SKILL.md` — Description rescopen
- `~/r/handbook/skills/README.md:7-24` — Skill-Tabelle

### What to build

Den deslop-Workflow (Scope-Ermittlung, Content-Typ-Erkennung, Slop-Entfernung)
in den cleanup-Workflow integrieren, sodass cleanup auch die bisherigen
deslop-Trigger („deslop", „remove slop", „clean up AI code") abdeckt. Die vier
deslop-Referenzdateien nach `cleanup/` verschieben und alle relativen Links
aktualisieren; dabei `text-de.md` gegen `readability-de.md` auf Redundanz
prüfen und ggf. zusammenführen. Das Verzeichnis `deslop/` entfernen. Die
code-audit-Description auf repo-weite/cross-layer Audits einengen (kein
„readability review"-Trigger mehr). Skill-Tabelle im README anpassen —
einschließlich der toten Zeilen für die bereits gelöschten Skills
(design-interface, improve-architecture, handbook-sync, extract) und des
„Improve Architecture"-Schritts in der Workflow-Sektion.

### Acceptance criteria

- [ ] `skills/deslop/` existiert nicht mehr; keine toten Links in `cleanup/`
- [ ] cleanup-Description nennt die übernommenen deslop-Trigger
- [ ] code-audit-Description überlappt nicht mehr mit cleanup
- [ ] Neue Claude-Code-Session listet cleanup (gemergt) und kein deslop
- [ ] README-Tabelle ist konsistent mit den vorhandenen Skill-Verzeichnissen

---

## Phase 7: Skill-Distributions-Doku korrigieren (handbook)

### Context

- `~/r/handbook/skills/README.md:3` — „copy individual skill directories into
  project repos as needed"

### What to build

Das README auf das tatsächliche Distributionsmodell umstellen: Skills leben im
handbook-Repo und sind als `~/.claude/skills` symlinkt; sowohl Claude Code als
auch GitHub Copilot (VS Code) lesen sie von dort; kein Kopieren in
Projekt-Repos. Vorher einmal praktisch verifizieren, dass Copilot die Skills
tatsächlich aus `~/.claude/skills` aufnimmt (z. B. Skill-Aufruf in einer
Copilot-Chat-Session), damit das README keine ungeprüfte Behauptung enthält.

### Acceptance criteria

- [ ] README beschreibt Symlink-Konsum durch beide Agents; „copy into project
      repos" ist entfernt
- [ ] Copilot-Discovery wurde praktisch verifiziert (oder die Einschränkung
      ist im README dokumentiert)
- [ ] „Adding a New Skill"-Verweis und Workflow-Sektion stimmen weiterhin

---

## Phase 8: Abgeschlossene Plan- und PRD-Artefakte entfernen (jotti)

### Context

- `docs/plans/` — fünf vollständig abgearbeitete Pläne (Stand 2026-06-11,
  alle Checkboxen erledigt): `plan-error-handling.md` (17/17),
  `plan-error-handling-simplification.md` (7/7),
  `plan-lokale-tls-selbstsigniert.md` (11/11),
  `plan-position-steuersatz.md` (15/15), `plan-tse-integration.md` (36/36).
  Offen bleiben: `plan-bondruck-test-escpresso.md`,
  `plan-windows-verpackung.md`, dieser Plan.
- `docs/prds/` — sechs PRDs + `notes.md`; ein Teil gehört zu bereits
  gelieferten Features, ein Teil zu künftiger Arbeit (z. B.
  `prd-betrieb-relay-haertung.md`, `prd-lokale-tls-vertrauenswuerdig.md`).

### What to build

Abgeschlossene Pläne aus `docs/plans/` löschen — die Git-Historie bewahrt
sie; im Arbeitsbaum sind sie veralteter Kontext, der Agents (z. B.
`/implement-plan`, Codebase-Suchen) in die Irre führen kann. PRDs einzeln
beurteilen: PRDs vollständig gelieferter Features löschen, sofern ihre
Anforderungen in `docs/anforderungen.md` bzw. `docs/handbuch.md` kanonisch
abgedeckt sind; PRDs offener oder künftiger Arbeit bleiben. Die Policy als
einen Satz festhalten (z. B. im Git-Workflow-Abschnitt von `AGENTS.md`):
abgeschlossene Pläne werden nach dem Merge gelöscht.

### Acceptance criteria

- [ ] `docs/plans/` enthält ausschließlich Pläne mit offenen Checkboxen
- [ ] Jedes verbleibende PRD gehört zu offener oder künftiger Arbeit
- [ ] Vor jeder Löschung geprüft: keine eingehenden Verweise aus `docs/`,
      `AGENTS.md` oder `.github/instructions/` auf die gelöschte Datei
- [ ] Die Lösch-Policy ist als ein Satz dokumentiert
