# Plan: Orchestrierung — Praxis-Feedback, Vollaudit, autonome Fixes, v1.0.0

> Source PRD: n/a. Super-Plan für eine eigene Orchestrator-Session. Er führt
> `plan-praxis-feedback.md` aus, lässt danach das gesamte Repo reviewen, plant die Fixes
> aus dem Findings-Dokument und setzt sie autonom um. Erst danach folgt v1.0.0.

## Goal

Eine einzige Lead-Session orchestriert alles autonom: Plan 2 umsetzen, jede Datei in jotti
nach einem Fable-Sweep mit Opus-Reviewern und den Cleanup-Kriterien prüfen, die Befunde in
einem selbsttragenden Findings-Dokument konsolidieren, daraus einen Fix-Plan erzeugen, ihn
mit Opus und Sonnet umsetzen und jede Phase nach einem Fable-Sweep von Opus reviewen lassen.
Die Lead-Session reviewt und implementiert nichts selbst.

## Architectural decisions

- **Rollen und Modelle**: Lead-Session = Fable 5.1, nur Orchestrierung (Workflows starten,
  Ergebnisse lesen, Plan-Dateien pflegen, Commits landen). Sweep = Fable 5.1: je Bereich
  bzw. je Diff ein kurzer, flacher Durchgang, der Hotspots und Fragen an die Reviewer
  übergibt. Reviewer, Skeptiker und Plan-Kritik = Opus 5.
  Implementierer = Opus 5 (Implementierung, Debugging, Migrationen, Architektur) und
  Sonnet 5 (mechanische Fixes, Umbenennungen, Doku-Sweeps, Formatierung). Planer = Opus 5.
- **Fable-Ausnahme**: `CLAUDE.md` untersagt Fable für Subagenten aus Kostengründen. Für diesen
  Plan hat der Eigentümer Fable nur für den Sweep freigegeben; der Preis rechtfertigt keine
  Fable-Reviewer. Die Regel in `CLAUDE.md` bleibt unverändert; die Ausnahme gilt nur für
  den Sweep.
- **Werkzeuge**: `implement-plan` (Worktrees, Commit je Kriterium, Fold, Landen),
  `create-plan`, `cleanup` (Kriterien-Dateien), Workflow-Tool für Fan-out und
  adversariale Verifikation. Der Audit-Workflow liegt als benannter Workflow in
  `.claude/workflows/jotti-full-audit.js` und wird mit
  `Workflow({ name: 'jotti-full-audit', args: { date: '<YYYY-MM-DD>' } })` gestartet.
- **Git**: Alles landet auf dem Feature-Branch der Orchestrator-Session, nie auf `main`.
  Kein Force-Push, kein `--no-verify`, keine KI-Trailer. Migrationsnummern werden beim
  Landen vergeben.
- **Datenschutz**: Keine Vereins- oder Personendaten im Repo. Der Vereins-Antwortplan bleibt
  außerhalb.

## Inventory

- `docs/plans/plan-praxis-feedback.md` — Plan 2, 12 Phasen (0–11) mit „Depends on"-Zeilen
- `.claude/workflows/jotti-full-audit.js` — Audit-Workflow: 22 Einheiten × 3 Linsen
  (Cleanup-Skill, Korrektheit/Security, Konventionen/Doku), 8 Cross-Layer-Flüsse, Dedupe,
  2–3 Skeptiker je Blocker/Major (Kappung 400), Konsolidierung je Bereich, Assembler
- Handbook-Skills: `create-plan`, `implement-plan` (mit `orchestration.md`,
  `integration.md`, `recovery.md`), `cleanup` (`readability.md`, `readability-de.md`,
  `principles.md`, `code-smells.md`, `architecture.md`, `cross-layer.md`),
  `verification-depth.md`, `dispatching-parallel-agents`
- `Makefile` — `make check`, `make verify`, `make website-check`, `make test-e2e`,
  `make sqlc`, `make rebuild-projections`
- `.github/workflows/ci.yml` — Job `upgrade-path` als Pflicht-Gate für Schema-Änderungen

## Resolved decisions

- **Reihenfolge**: erst Plan 2 (Phasen 0–10), dann Vollaudit, dann Audit-Fixes, dann
  v1.0.0 (Plan 2 Phase 11). Das Audit prüft damit den Stand nach dem Feedback, nicht davor.
- **Audit-Umfang**: jede Datei außer `backend/sqlc/dbgen/`, Lockfiles, Binärdateien,
  `docs/rechtsquellen/`. Minor-Befunde bleiben ungeprüft und werden so gekennzeichnet.
- **Findings-Dokument** ist selbsttragend: Zahlen, Top 10, Defektklassen mit
  Gate-Vorschlägen, Befunde je Datei mit Status, verworfene Befunde. Es ist die einzige
  Eingabe des Fix-Plans.
- **Fix-Plan** folgt `create-plan`: vertikale Schnitte, `Depends on`, testbare
  Abnahmekriterien, Review-Tier je Phase nach `verification-depth.md`. Defektklassen
  werden zu Gates (Lint, CI-grep, Test), nicht nur zu Einzelfixes.
- **Cleanup-Fixes ändern kein Verhalten.** Große Refactorings aus dem Audit werden im
  Fix-Plan als eigene Phase mit Entscheidung „jetzt / v1.1 / nie" geführt.
- **Stopp-Bedingungen** (aus `implement-plan`): Merge-Konflikt, zweimal gleicher
  Testfehler nach Debugging, nötiger Force-Push, Löschen fremder Worktrees. Dann
  Übergabe an den Menschen, nicht raten.

- Phase A: Kriterium 1 bleibt offen, bis der Eigentümer nach dem Merge Plan 2 Phase 0
  Kriterium 5 (Dependabot-PRs) und Phase 10 Kriterium 4 (#111) abhakt; alle übrigen
  Kriterien der Phasen 0–10 sind abgehakt.

## Open questions / Risks

- Dauer und Kosten: das Audit hat sechs Fable-Sweeps, rund 80 Opus-Reviewer und Skeptiker
  je Befund; ein Vorlauf mit 268 Agenten brauchte 3,3 Stunden. Rechne mit einem Tag
  Laufzeit für Phase B und einem weiteren für Phase D.
- Zwei Vereine setzen jotti ab Ende September produktiv ein. Phase A muss vorher landen;
  Phasen B–D dürfen v1.0.0 nicht über diesen Termin hinaus verzögern, sonst v1.0.0 mit
  Stand nach Phase A taggen und den Rest in v1.1 planen.
- Das Parallelitätslimit je Workflow ist CPUs − 2 (in der Cloud-Session 2 Agenten). Der
  Audit läuft deshalb im Split-Modus des Workflows: sechs Bereichs-Läufe parallel
  (`area`, `sectionsDir`), danach ein Assemble-Lauf (`assembleFrom`).

---

## Phase A: Plan 2 umsetzen

**Depends on**: none

### Context

- `docs/plans/plan-praxis-feedback.md` — Phasen 0–10; Phase 11 (Release) bleibt hier
  ausgeklammert
- Handbook `implement-plan` — Run-Contract, Worktrees, Commit je Kriterium, Fold, Landen

### What to build

Die Lead-Session führt `implement-plan` für Plan 2 aus, Phasen 0–10. Parallelgruppen nach
den „Depends on"-Zeilen. Implementierer je Phase: Opus für 0, 4, 5, 7, 8, 10; Sonnet für
1, 2, 3, 6, 9. Nach jeder Phase ein Review des Diffs (Korrektheit, Konventionen,
Cleanup-Kriterien); Defekte gehen per `SendMessage` an den Phasen-Worker zurück, nie an
einen neuen Fixer. Gate je Phase: `make check`, für Schema-Phasen `make verify` und der
CI-Job `upgrade-path`, für Website-Phasen `make website-check`, für Phase 4
`make test-e2e`.

### Acceptance criteria

- [ ] Alle Kriterien der Phasen 0–10 in `plan-praxis-feedback.md` abgehakt, je Kriterium
      ein Commit mit Trailer `Plan: praxis-feedback phase <N> criterion <M>`
- [x] Je Phase ein Review-Protokoll mit „keine offenen Defekte" vor dem Fold
- [x] `plan/praxis-feedback` auf den Feature-Branch gelandet, alle Worktrees entfernt
- [x] `make verify`, `make website-check`, `make test-e2e` auf dem gelandeten Stand grün

---

## Phase B: Vollaudit und Findings-Dokument

**Depends on**: A

### Context

- `.claude/workflows/jotti-full-audit.js` — benannter Workflow
- Handbook `cleanup` — Kriterien-Dateien, die die Reviewer lesen

### What to build

`Workflow({ name: 'jotti-full-audit', args: { date: '<heute>' } })` auf dem gelandeten
Stand nach Phase A; auf kleinen Maschinen im Split-Modus (sechs Läufe mit `area` und
`sectionsDir`, dann ein Lauf mit `assembleFrom`). Ergebnis ist `docs/plans/findings-jotti-audit.md`. Die Lead-Session
liest nur die Kennzahlen und committet die Datei. Meldet der Workflow Reviewer ohne
Ergebnis oder eine Kappung, wird das im Dokument-Kopf sichtbar; ein Nachlauf nur für die
fehlenden Einheiten ist über `resumeFromRunId` möglich.

### Acceptance criteria

- [ ] `docs/plans/findings-jotti-audit.md` existiert, enthält Zahlen, Top 10,
      Defektklassen, Befunde je Bereich, verworfene Befunde
- [ ] Kein Befund ohne Datei und Zeilenbereich; jeder Blocker/Major trägt einen
      Verifikationsstatus
- [ ] „Reviewer ohne Ergebnis" ist 0 oder die fehlenden Einheiten sind im Kopf benannt
- [ ] Datei committet; Leak-Check auf Personendaten negativ

---

## Phase C: Fix-Plan aus dem Findings-Dokument

**Depends on**: B

### Context

- `docs/plans/findings-jotti-audit.md` — Eingabe
- Handbook `create-plan` — Template, Ask-Gate, Self-Review
- `docs/plans/plan-praxis-feedback.md` — bereits umgesetzt; Überschneidungen vermeiden

### What to build

Ein Opus-Planer erzeugt `docs/plans/plan-jotti-audit-fixes.md` nach `create-plan`:
Phasen als vertikale Schnitte je Defektklasse oder Bereich, Gates zuerst (Lint, CI-grep,
Test), dann Blocker, Major, Minor; große Refactorings als eigene Entscheidungsphase.
Jede Phase nennt Implementierer-Modell (Opus/Sonnet), Review-Tier und Gate-Befehl. Ein
Fable-Sweep übergibt an drei Opus-Kritiker, die den Plan gegen das Findings-Dokument
prüfen (Vollständigkeit, keine Verhaltensänderung bei Cleanup, Freeze-Disziplin, Rule 18);
der Planer arbeitet die Kritik ein. Das Ask-Gate wird durchlaufen; verbleibende Fragen
stehen als „Open questions" im Plan, die Lead-Session entscheidet sie nach `question-
rules.md` oder stoppt.

### Acceptance criteria

- [ ] `docs/plans/plan-jotti-audit-fixes.md` deckt jeden bestätigten Blocker/Major-Befund
      und jede Defektklasse ab; nicht übernommene Befunde stehen mit Begründung darin
- [ ] Jede Phase hat `Depends on`, Modell, Review-Tier, Gate-Befehl und testbare Kriterien
- [ ] Kritik (Fable-Sweep, Opus-Kritiker) dokumentiert und eingearbeitet
- [ ] Datei committet

---

## Phase D: Audit-Fixes umsetzen

**Depends on**: C

### Context

- `docs/plans/plan-jotti-audit-fixes.md`
- Handbook `implement-plan`, `verification-depth.md`, `systematic-debugging`

### What to build

`implement-plan` für den Fix-Plan, Modell je Phase wie im Plan festgelegt, je Phase ein
Fable-Sweep und ein Opus-Review mit Rückgabe an den Worker. Gates wie in Phase A. Neue
Gates aus den Defektklassen werden vor den Einzelfixes gelandet, damit sie die Fixes
prüfen. Stopp-Bedingungen aus `implement-plan` gelten unverändert.

### Acceptance criteria

- [ ] Alle Kriterien des Fix-Plans abgehakt und gelandet; Plan-Datei danach gelöscht
- [ ] Jede Defektklasse hat ein Gate in CI oder Lint, das auf dem gelandeten Stand grün ist
- [ ] `make verify`, `make website-check`, `make test-e2e` grün
- [ ] `docs/plans/findings-jotti-audit.md` gelöscht oder auf die offenen Reste reduziert

---

## Phase E: Release v1.0.0

**Depends on**: A, D

### Context

- `docs/plans/plan-praxis-feedback.md` — Phase 11
- `docs/plans/guide-manuelle-qa-v1.0.0.md`

### What to build

Plan 2 Phase 11 ausführen: QA-Guide, Tag, Release, `PREVIOUS_VERSION`-Bump. Manuelle
Schritte des QA-Guides (Hardware, Windows-Rechner, fiskaly-Konto) bleiben beim Menschen;
die Lead-Session bereitet alles vor und stoppt vor dem Tag mit einer Übergabe.

### Acceptance criteria

- [ ] Alle Kriterien von Plan 2 Phase 11 abgehakt
- [ ] `plan-praxis-feedback.md` und dieser Plan gelöscht, Git-Historie bewahrt sie
