---
title: Plan — Semantisches Changelog mit git-cliff
description: Release-Notes aus Conventional Commits per git-cliff generieren statt aus GitHubs PR-basierten Auto-Notes, die bei trunk-based Development ohne PRs leer bleiben. Gefiltert auf user-relevante Änderungen, deutsche Gruppen-Überschriften, manuelles SemVer-Tagging bleibt.
---

# Plan: Semantisches Changelog mit git-cliff

> Source PRD: n/a (aus der Deployment-/Release-Analyse in der Konversation abgeleitet)

## Goal

Die GitHub-Release-Notes von jotti sollen ein **sauberes, lesbares, nach Typ
gruppiertes Changelog** enthalten, das ausschließlich aus der Git-History
generiert wird — ohne PR-Abhängigkeit.

**Warum:** [release.yml](../../.github/workflows/release.yml) nutzt
`generate_release_notes: true`. Diese GitHub-API listet **nur gemergte PRs**
(plus neue Contributors). Da auf `main` trunk-based ohne PRs gearbeitet wird,
bleibt der Release-Body bis auf den `**Full Changelog**`-Compare-Link **leer**
(verifiziert an v0.10.0 und v0.11.0). Die Commits sind bereits sauber
Conventional Commits — daraus lässt sich direkt ein Changelog erzeugen.

## Architectural decisions

Durchgängige Entscheidungen (gelten für alle Phasen):

- **Tool: git-cliff** über die offizielle Action `orhun/git-cliff-action`
  (Major-Version beim Einbau auf den dann aktuellen Stand pinnen). Liest
  Conventional Commits direkt aus `git log`, keine PR-Abhängigkeit, ein Binary,
  kein Node-Runtime in CI nötig.
- **Verworfen:** release-please (PR-zentriert), semantic-release (übernimmt
  Versionierung + Tagging, schwergewichtig), GitHub-native `.github/release.yml`
  (PR-basiert).
- **Quelle = Conventional Commits.** Vorhandenes Format `typ(scope): subject`
  wird vorausgesetzt; unkonventionelle Commits werden verworfen
  (`filter_unconventional = true`).
- **Filter auf user-relevante Typen:** nur `feat` und `fix` erscheinen im
  Changelog. `docs`, `refactor`, `chore`, `ci`, `test`, `build`, `style`, `perf`
  werden ausgefiltert. Begründung: zwischen zwei Minor-Tags liegen ~150 Commits,
  davon der Großteil nicht user-relevant; Zielgruppe sind nicht-technische
  Vereine.
- **Scope `website` wird ausgeblendet** (Marketing-Seite, kein Produkt-Änderung
  für Anwender).
- **Deutsche Gruppen-Überschriften:** `✨ Neue Funktionen` (feat),
  `🐛 Fehlerbehebungen` (fix).
- **Breaking Changes als eigene Sektion oben.** Commits mit `typ!:` oder
  `BREAKING CHANGE`-Footer landen in `⚠️ Breaking Changes` vor allen anderen
  Gruppen. Bei einem 0.x-Produkt sollen inkompatible Änderungen vor dem Update
  sichtbar sein.
- **Eintrags-Format mit Scope-Präfix.** Jeder Eintrag zeigt seinen Scope fett
  vorangestellt (z.B. `**kasse:** …`), Commits ohne Scope ohne Präfix. Gibt die
  Bereichszuordnung; technische Scopes (`dsfinvk`) bleiben Jargon (akzeptiert).
- **`**Full Changelog**`-Compare-Link bleibt** im Footer (`vPrev...vCur`), wie
  bei den bisherigen Auto-Notes.
- **Nur Release-Body, kein eingechecktes `CHANGELOG.md`.** Vermeidet
  Commit-back-Loops auf einem trunk-based Repo; das Changelog lebt auf der
  GitHub-Release-Seite. (Eingechecktes `CHANGELOG.md` bleibt als spätere
  Erweiterung möglich, siehe Open questions.)
- **Manuelles SemVer-Tagging bleibt unverändert.** git-cliff erzeugt nur den
  Body, vergibt keine Versionen und legt keine Tags an. Lightweight-Tags
  (`v0.X.Y`) bleiben wie bisher.
- **Changelog wird in beiden Trigger-Pfaden erzeugt** (Tag-Push *und*
  `workflow_dispatch`-Dry-Run). Im Dry-Run dient es als Vorschau (Ausgabe ins
  Job-Summary), beim Tag-Push wird es an das Release gehängt. Das passt zur
  bestehenden Workflow-Philosophie „Dispatch validiert alles, publiziert nichts".

## Inventory

- [.github/workflows/release.yml](../../.github/workflows/release.yml):
  - `:26` — `actions/checkout@v6` **ohne `fetch-depth`** → flacher Klon. git-cliff
    braucht die volle History inkl. aller Tags (`fetch-depth: 0`).
  - `:35-44` — „Version bestimmen"-Step setzt `version` und `publish`
    (true bei Push, false bei Dispatch).
  - `:125-131` — Release-Step (`softprops/action-gh-release@v3`) mit
    `generate_release_notes: true` → der zu ersetzende Teil.
- [Makefile](../../Makefile)`:90-94` — Version kommt ausschließlich aus
  `VERSION` (= Tag-Name); kein `VERSION`-File. Kontext, bleibt unangetastet.
- Tags: `v0.9.0`, `v0.9.1`, `v0.10.0`, `v0.11.0`, alle **lightweight**. Pattern
  für git-cliff: `tag_pattern = "v[0-9]*"`.
- Kein bestehendes `cliff.toml` / git-cliff im Repo.

## Resolved decisions

Aus drei Klärungsrunden (alle bestätigt):

- Nur `feat` + `fix` im Changelog; übrige Typen ausgefiltert.
- Scope `website` ausgeblendet.
- Deutsche Gruppen-Überschriften mit Emoji.
- Eintrags-Format mit fettem Scope-Präfix.
- Breaking Changes als eigene Sektion `⚠️ Breaking Changes` ganz oben.
- `**Full Changelog**`-Compare-Link im Footer behalten.
- Release-Body-only, kein eingechecktes `CHANGELOG.md`.
- Manuelles Tagging bleibt; git-cliff rendert nur.
- Changelog auch im Dry-Run erzeugen (Vorschau im Job-Summary).
- **Backfill** der bestehenden Releases **v0.9.1, v0.10.0, v0.11.0** (Phase 3).
  **v0.9.0 ausgelassen** (erster Tag → Genesis-Dump aller Frühcommits wäre
  unlesbar). Einmaliger, lokaler Vorgang per `gh release edit`; die Bodies
  werden vor dem Schreiben auf die Live-Releases bestätigt.

## Open questions / Risks

- **Gemischte Sprache der Subjects.** Commit-Betreffe sind teils Deutsch, teils
  Englisch; git-cliff übersetzt nicht. Das Changelog bleibt damit gemischt-
  sprachig/technisch. Akzeptiert für jetzt; eine kuratierte deutsche
  „Was ist neu"-Sektion wäre ein separates, manuelles Thema.
- **Eingechecktes `CHANGELOG.md`** als spätere Erweiterung möglich (für
  Repo-Browser); erfordert dann einen Commit-back-Step und ist bewusst
  außerhalb dieses Plans.
- **Breaking-Changes-Sektion ohne Doppelung.** Breaking-Commits dürfen nicht
  zusätzlich in ihrer feat/fix-Gruppe erscheinen. Die exakte Tera-/Parser-Lösung
  (Gruppen-Sortierpräfix `<!-- N -->`, Zuordnung über `commit_parsers`) wird in
  Phase 1 lokal verifiziert.
- **Eingechecktes `CHANGELOG.md`** als spätere Erweiterung möglich (für
  Repo-Browser); erfordert dann einen Commit-back-Step und ist bewusst
  außerhalb dieses Plans.

---

## Phase 1: cliff.toml + lokale Verifikation

### Context

- Keine bestehende Config — `cliff.toml` wird im Repo-Root neu angelegt.
- Tags sind lightweight, Pattern `v[0-9]*`.

### What to build

Ein `cliff.toml` im Repo-Root, das aus der vorhandenen Commit-History ein nach
Typ gruppiertes, deutsches Changelog erzeugt: `⚠️ Breaking Changes` zuerst, dann
`✨ Neue Funktionen` (feat), dann `🐛 Fehlerbehebungen` (fix). Jeder Eintrag mit
fettem Scope-Präfix; `scope = website` und alle übrigen Typen ausgefiltert;
Compare-Link im Footer. Startkonfiguration (exakte Tera-Details werden beim
lokalen Lauf finalisiert):

```toml
[changelog]
header = ""
body = """
{% for group, commits in commits | group_by(attribute="group") %}
### {{ group | upper_first }}
{% for commit in commits %}
- {% if commit.scope %}**{{ commit.scope }}:** {% endif %}\
{{ commit.message | split(pat="\n") | first | upper_first }}
{%- endfor %}
{% endfor %}
"""
footer = """
{% if previous and previous.version and version %}
**Full Changelog**: https://github.com/nicograef/jotti/compare/{{ previous.version }}...{{ version }}
{% endif %}
"""
trim = true

[git]
conventional_commits = true
filter_unconventional = true
filter_commits = true
# Reihenfolge zählt: Breaking zuerst, sonst landen feat!/fix! in feat/fix.
# Das <!-- N -->-Präfix steuert die Gruppen-Sortierung und wird im Output entfernt.
commit_parsers = [
  { message = "^.*!:",          group = "<!-- 0 -->⚠️ Breaking Changes" },
  { body = ".*BREAKING CHANGE", group = "<!-- 0 -->⚠️ Breaking Changes" },
  { message = "^feat",          group = "<!-- 1 -->✨ Neue Funktionen" },
  { message = "^fix",           group = "<!-- 2 -->🐛 Fehlerbehebungen" },
  { scope = "website", skip = true },
  { message = ".*", skip = true },
]
tag_pattern = "v[0-9]*"
```

Lokal verifizieren (Node ist vorhanden): `npx git-cliff --latest` rendert den
Bereich seit dem vorletzten Tag. Output gegen `git log v0.10.0..v0.11.0`
gegenprüfen — nur feat/fix, kein `website`, Scope-Präfixe korrekt, deutsche
Überschriften, Gruppen-Sortierpräfixe `<!-- N -->` im Output entfernt, Breaking
Changes (falls vorhanden) nur in der eigenen Sektion (keine Doppelung),
Compare-Link im Footer.

### Acceptance criteria

- [ ] `cliff.toml` liegt im Repo-Root.
- [ ] `npx git-cliff --latest` läuft fehlerfrei und gibt ein nach Typ
      gruppiertes Markdown-Changelog aus.
- [ ] Nur `feat`- und `fix`-Commits erscheinen; `docs`/`refactor`/`chore`/
      `ci`/`test`/`build` fehlen; `scope = website` fehlt.
- [ ] Gruppen-Reihenfolge ist Breaking Changes → Neue Funktionen →
      Fehlerbehebungen; die `<!-- N -->`-Sortierpräfixe sind im Output nicht
      sichtbar.
- [ ] Einträge zeigen den Scope als fettes Präfix; Commits ohne Scope erscheinen
      ohne Präfix.
- [ ] Breaking-Commits (`typ!:`/`BREAKING CHANGE`) erscheinen ausschließlich in
      `⚠️ Breaking Changes`, nicht zusätzlich in feat/fix.
- [ ] Footer enthält den `**Full Changelog**`-Compare-Link.
- [ ] Stichprobe gegen `git log v0.10.0..v0.11.0` bestätigt Vollständigkeit der
      feat/fix-Einträge.

---

## Phase 2: Integration in den Release-Workflow

### Context

- [.github/workflows/release.yml](../../.github/workflows/release.yml)`:26`
  (Checkout), `:35-44` (Version/publish-Flag), `:125-131` (Release-Step).

### What to build

git-cliff in [release.yml](../../.github/workflows/release.yml) einhängen:

1. **Checkout auf volle History:** `fetch-depth: 0` am `actions/checkout`-Step
   ergänzen, damit git-cliff alle Tags und Commits sieht.
2. **Changelog-Step** (vor dem Release, läuft in **beiden** Trigger-Pfaden) via
   `orhun/git-cliff-action` mit `config: cliff.toml` und `args: --latest
   --strip header`; Ergebnis in eine Step-Output-Variable. Den gerenderten
   Body zusätzlich ins `$GITHUB_STEP_SUMMARY` schreiben, damit der Dry-Run eine
   Vorschau liefert.
3. **Release-Step** (`if: publish == 'true'`): `generate_release_notes: true`
   entfernen und stattdessen `body:` aus dem Changelog-Step-Output setzen. ZIP-
   `files:` bleibt unverändert.

Verifizieren über `workflow_dispatch` (Dry-Run): Lauf baut/smoke-testet wie
bisher und zeigt das gerenderte Changelog im Job-Summary, ohne zu publizieren.
Der nächste echte Tag-Push erzeugt dann ein Release mit ausgefülltem Body.

### Acceptance criteria

- [ ] `actions/checkout` nutzt `fetch-depth: 0`.
- [ ] Neuer Changelog-Step läuft in Push- und Dispatch-Pfad und rendert per
      `cliff.toml`.
- [ ] Im `workflow_dispatch`-Dry-Run erscheint das gerenderte Changelog im
      Job-Summary; es wird **kein** Release/Image publiziert.
- [ ] Release-Step nutzt `body:` aus dem Changelog-Output statt
      `generate_release_notes: true`; das Windows-ZIP wird weiterhin angehängt.
- [ ] Der nächste Tag-Push (oder ein manueller Re-Run gegen einen Tag) erzeugt
      ein Release, dessen Body das gruppierte feat/fix-Changelog enthält.

---

## Phase 3: Backfill bestehender Releases

### Context

- Bestehende Releases: `v0.9.0`, `v0.9.1`, `v0.10.0`, `v0.11.0` — Bodies aktuell
  leer bis auf den Compare-Link.
- `v0.9.0` ist der erste Tag (kein Vorgänger) → wird **ausgelassen**.

### What to build

Einmalig, lokal: für `v0.9.1`, `v0.10.0`, `v0.11.0` je den Body mit dem in
Phase 1 verifizierten `cliff.toml` rendern und auf das jeweilige Live-Release
schreiben. Pro Tag den Range gegen den Vorgänger rendern
(`git-cliff <vPrev>..<vCur>`), das Ergebnis sichten und mit
`gh release edit <vCur> --notes-file <datei>` setzen.

Da dies **publizierte Releases verändert** (outward-facing), werden die drei
gerenderten Bodies erst vorgelegt und bestätigt, bevor `gh release edit` läuft.
`v0.9.0` bleibt unangetastet.

### Acceptance criteria

- [ ] Bodies für `v0.9.1`, `v0.10.0`, `v0.11.0` lokal gerendert und vor dem
      Schreiben bestätigt.
- [ ] Die drei Releases zeigen das gruppierte feat/fix-Changelog (gleiche Form
      wie künftige Releases aus Phase 2).
- [ ] `v0.9.0` bleibt unverändert.
