# Plan: v1.0-Release-Blocker (Code-Seite)

> Source PRD: n/a (aus Task-Beschreibung; hervorgegangen aus dem gelöschten
> `docs/plans/offene-punkte.md`, Abgleich gegen den Code-Stand am 2026-07-20)

## Goal

Die vier code-seitigen v1.0-Blocker beheben, die vor dem Tag `v1.0.0` fällig
sind und ein Coding-Agent umsetzen kann:

1. **zog** — `GTE(0).Required()` auf Geld-Summenfeldern korrigieren (legitimer
   0-Betrag wird sonst mit HTTP 500 statt sauberer Validierung abgelehnt).
2. **G7** — die lokale DB-Wipe-Fähigkeit entfernen (Reset/Wipe wird nur noch
   für die Demo-Instanz `rocks` gebraucht).
3. **G3** — den Prettier-PostToolUse-Hook auf die Repo-Wurzel begrenzen.
4. **CHANGELOG** — eine stehende `CHANGELOG.md` anlegen.

Die **manuelle** Release-Vorbereitung (Pin-Bump, Versions-Bump, Tag, QA mit
echter Hardware) liegt außerhalb dieses Plans und steht in
`docs/plans/guide-manuelle-qa-v1.0.0.md`.

## Architectural decisions

Durable-Entscheidungen für alle Phasen:

- **zog-Semantik.** Auf einem stets gesetzten `int`-Feld ist `GTE(0).Required()`
  ein Anti-Muster: zog wertet den Go-Zero-Value `0` als „fehlend", die
  Kombination bedeutet damit versehentlich „> 0". Kanonischer Referenz-Fix:
  Commit `531a17f` (`backend/domain/kasse/kassensitzung_events.go` —
  `BetragCents`, `IstBestandCents`: `Required()` gestrichen, `GTE(0)` bleibt,
  deutscher Kommentar; Vorbild `SollBestandCents` = bares `z.Int()`).
  Zielzustand pro Feld: **kein** `GTE(0).Required()` mehr — entweder
  `z.Int().GTE(0)` (0 ist gültig) oder `z.Int().GTE(1)` (0 ist ungültig,
  Absicht explizit; `GTE(1)` ist das im Repo etablierte „muss positiv"-Idiom,
  vgl. `UserID`/`TischID`).
- **Freeze-Disziplin.** Keine Änderung an Event-Feldnamen, Event-Versionen oder
  Event-JSON-Struktur. `Required()` zu streichen bzw. auf `GTE(1)` zu heben
  ändert nur die Validierungsstrenge, nicht das serialisierte JSON — der
  Contract-Guard `backend/domain/kasse/event_json_contract_test.go` bleibt grün.
- **G7-Ansatz.** Fähigkeit entfernen statt absichern: `scripts/reset-and-seed.sh`
  wird **rocks-only**. Kein Bestands-Guard, kein neues Flag — `--yes` bleibt die
  einzige Nicht-interaktiv-Logik.
- **CHANGELOG.** Stehende `CHANGELOG.md` in der Repo-Wurzel, generiert per
  git-cliff aus der vollen History (Keep-a-Changelog-Stil mit
  `## [x.y.z] - YYYY-MM-DD`-Überschriften). Die bestehende
  Release-Notes-Erzeugung in `release.yml` (`git-cliff --latest --strip header`)
  darf sich dabei **nicht** ändern.

## Inventory

- `backend/domain/kasse/kassensitzung_events.go` — Referenz-Muster für den
  zog-Fix (bereits korrekt: `BetragCents`, `IstBestandCents` ohne `Required()`).
- **Kategorie A — Domain-Entitäts-Schemas** (validieren intern gebaute Structs,
  keine HTTP-Eingabe):
  - `backend/domain/kasse/bestellung.go — bestellungSchema` → `GesamtPreisCents`
  - `backend/domain/kasse/zahlung.go — zahlungSchema` → `GesamtZahlungCents`
  - `backend/domain/kasse/stornierung.go — stornierungSchema` → `GesamtStornierungCents`
  - `backend/domain/kasse/umbuchung.go — umbuchungSchema` → `GesamtCents`
- **Kategorie B — Event-Data-Schemas** (validieren Event-JSON vor Persistenz):
  - `backend/domain/kasse/direktverkauf_events.go — direktverkaufGetaetigtV1DataSchema` → `GesamtbetragCents`
  - `backend/domain/kasse/direktverkauf_events.go — direktverkaufStorniertV1DataSchema` → `GesamtStornierungCents`
  - `backend/domain/kasse/tisch_session_events.go — bestellungAufgenommenV1DataSchema` → `GesamtPreisCents`
  - `backend/domain/kasse/tisch_session_events.go — zahlungKassiertV1DataSchema` → `GesamtZahlungCents`
  - `backend/domain/kasse/tisch_session_events.go — stornierungErteiltV1DataSchema` → `GesamtStornierungCents`
  - `backend/domain/kasse/tisch_session_events.go — bestellungKorrigiertV1DataSchema` → `GesamtCents`
  - `backend/domain/kasse/tisch_session_events.go — bestellungUmgebuchtV1DataSchema` → `GesamtCents`
- `scripts/reset-and-seed.sh` — Stack-Reset; heute `local|rocks`, entfernt den
  `local`-Zweig.
- `Makefile` — Targets `local-reset-and-seed` und `local-reset-db` (beide
  wipen das LAN-Volume `jotti-local_postgres-data`), `.PHONY`-Liste,
  `rocks-reset-and-seed` (bleibt).
- `.claude/settings.json` — PostToolUse-Hook (`matcher: "Write|Edit"`), die
  `node_modules`-Suche via `dirname`.
- `cliff.toml` — git-cliff-Config (`header = ""`, Body ohne Versions-Überschrift,
  `commit_parsers`, `tag_pattern`).
- `.github/workflows/release.yml` — Schritt „Changelog generieren"
  (`orhun/git-cliff-action@v4`, `args: --latest --strip header`).

## Resolved decisions

- **Roadmap-Lücke** (K-23, R-02, F-08, F-09): ersatzlos gestrichen, im Code
  bereits umgesetzt — **nicht Teil dieses Plans**.
- **G7-Scope** (Nutzer, 2026-07-20): Lokaler Reset/Wipe wird gar nicht
  gebraucht; Wipen nur für die Demo-Instanz `rocks`. Deshalb **beide** lokalen
  Reset-Targets entfernen (`local-reset-and-seed` **und** `local-reset-db` —
  identischer Footgun auf demselben Volume) und das Skript rocks-only machen.
  Kein Guard, kein zweites Flag (`--yes` + `--force-wipe` wären
  Doppel-Logik). Entwickler-Reset bleibt unberührt (`make clean` + `make dev` +
  `make seed` auf dem `jotti-*-dev`-Stack).
- **Session-Scope** (Nutzer, 2026-07-20): CHANGELOG anlegen ja; Pin-/Versions-
  Bump bleibt Release-Commit-Aktion (Guide), nicht dieser Plan.

## Open questions / Risks

- **zog-0-Erreichbarkeit ist die einzige Urteils-Risikostelle.** Eine falsche
  Einschätzung („0 nicht erreichbar") entfernt entweder einen sinnvollen Schutz
  oder lässt einen echten 500-Pfad stehen. Mitigation: pro Feld die
  Konstruktionspfade tracen (wer ruft den `New…Event`-Konstruktor / baut das
  Struct, kann die Positionssumme 0 werden — z. B. 0-Preis-Produkte?), die
  Entscheidung im Code-Kommentar begründen und je Feld einen Regressionstest
  ergänzen.
- **CHANGELOG-Templating.** Der Header lässt sich gefahrlos ergänzen
  (`--strip header` entfernt ihn aus den Release-Notes). Die **Versions-
  Überschrift** im Body ist der heikle Teil: sie darf in der `CHANGELOG.md`
  erscheinen, aber die `--latest --strip header`-Ausgabe der Release-Notes nicht
  verändern. Vor/Nach-Vergleich dieser Ausgabe ist Pflicht (siehe Phase 4).

---

## Phase 1: zog — Geld-Summenfelder korrekt validieren

### Context

- `backend/domain/kasse/kassensitzung_events.go` — Referenz-Fix `531a17f`
  (Muster + Kommentarstil übernehmen).
- Kategorie A (Domain-Entitäts-Schemas) und Kategorie B (Event-Data-Schemas)
  aus dem Inventory — insgesamt 11 Felder.
- `backend/domain/kasse/event_json_contract_test.go` — Contract-Guard, muss
  grün bleiben.
- HTTP-Präsenzprüfung: die `api/kasse/**/http`-Request-Schemas erzwingen
  Anwesenheit von Eingabefeldern via Ptr+NotNil; die 11 Felder hier sind jedoch
  **abgeleitet** (server-seitig aus Positionen summiert), nicht direkt aus dem
  Request übernommen — Präsenz ist strukturell garantiert.

### What to build

Für jedes der 11 Felder entscheiden, ob eine **0-Summe fachlich erreichbar**
ist, und `GTE(0).Required()` entsprechend auflösen:

- **0 gültig** → `z.Int().GTE(0)` (kein `Required()`), mit deutschem Kommentar
  im Stil von `531a17f`.
- **0 ungültig** → `z.Int().GTE(1)`, mit Kommentar, der die „muss positiv"-
  Absicht benennt (ersetzt die versehentliche `Required()`-Semantik durch eine
  ehrliche).

Kein Feld behält `GTE(0).Required()`. Die Entscheidung pro Feld wird durch das
Tracing der Konstruktionspfade begründet (kann die summierte Position 0 werden —
etwa durch ein 0-Preis-Produkt oder eine leere-Wert-Umbuchung?).

### Acceptance criteria

- [ ] Alle 11 Felder aus dem Inventory sind auf `GTE(0)` **oder** `GTE(1)`
      umgestellt; kein `GTE(0).Required()` verbleibt auf einem Geld-Summenfeld
      in `backend/domain/kasse/`.
- [ ] Jede Feld-Entscheidung (0 gültig vs. ungültig) ist mit einem knappen
      deutschen Kommentar am Feld begründet.
- [ ] Für jedes auf `GTE(0)` gesetzte Feld existiert ein Regressionstest, der
      belegt, dass ein 0-Wert die Validierung/Event-Erstellung besteht (Vorbild:
      `TestNewKassensitzungEroeffnetEvent_ErlaubtNullBetrag`).
- [ ] Für jedes auf `GTE(1)` gesetzte Feld belegt ein Test, dass 0 als
      Validierungsfehler (nicht als 500/„fehlend") abgelehnt wird.
- [ ] `event_json_contract_test.go` unverändert grün; keine Event-Feldnamen,
      -Versionen oder JSON-Strukturen berührt.
- [ ] `make test` grün.

---

## Phase 2: G7 — lokale DB-Wipe-Fähigkeit entfernen

### Context

- `scripts/reset-and-seed.sh` — `case "$STACK"`-Verzweigung mit `local`- und
  `rocks`-Zweig, Usage-Text, Kopfkommentar.
- `Makefile` — `local-reset-and-seed` (ruft `reset-and-seed.sh local --yes`),
  `local-reset-db` (direkter `docker volume rm jotti-local_postgres-data`),
  `.PHONY`-Liste, `rocks-reset-and-seed` (bleibt unverändert).

### What to build

`reset-and-seed.sh` wird rocks-only: `local`-Zweig, `local`-Erwähnungen in
Usage/Argumentprüfung/Kopfkommentar entfernen (Skript nimmt nur noch `rocks`).
Die beiden lokalen Wipe-Targets `local-reset-and-seed` und `local-reset-db`
werden aus dem `Makefile` entfernt (inkl. `.PHONY`-Eintrag). Die
nicht-destruktiven LAN-Targets `local-up`, `local-down`, `local-logs` bleiben.
`rocks-reset-and-seed` bleibt unverändert (`--yes`).

### Acceptance criteria

- [ ] `scripts/reset-and-seed.sh` akzeptiert nur noch `rocks`; ein `local`-
      Argument wird mit klarer Meldung abgewiesen; Usage/Kopfkommentar nennen
      `local` nicht mehr.
- [ ] `Makefile` enthält weder `local-reset-and-seed` noch `local-reset-db`
      (auch nicht in `.PHONY`); `local-up`/`local-down`/`local-logs` und
      `rocks-reset-and-seed` bleiben unverändert.
- [ ] Keine verwaisten Referenzen auf die entfernten Targets/den `local`-Zweig
      (grep über `Makefile`, `scripts/`, `docs/`, `.github/`).
- [ ] `make help` läuft ohne Fehler; `bash -n scripts/reset-and-seed.sh` ist
      syntaktisch sauber.

---

## Phase 3: G3 — Prettier-Hook auf die Repo-Wurzel begrenzen

### Context

- `.claude/settings.json` — PostToolUse-Hook (`matcher: "Write|Edit"`); die
  `while`-Schleife läuft per `dirname` bis `/` hoch und würde ein außerhalb des
  Repos liegendes `node_modules/.bin/prettier` ausführen (aktuell No-op).

### What to build

Vor dem Binary-Test in der `while`-Schleife einen Guard einfügen, der die Suche
auf `$CLAUDE_PROJECT_DIR` (und Unterverzeichnisse) begrenzt:

```
if [ -n "$CLAUDE_PROJECT_DIR" ]; then case "$d" in "$CLAUDE_PROJECT_DIR"|"$CLAUDE_PROJECT_DIR"/*) ;; *) break;; esac; fi;
```

### Acceptance criteria

- [ ] Der Guard ist in der `while`-Schleife vor dem `[ -x "$d/node_modules/.bin/prettier" ]`-
      Test eingefügt; oberhalb von `$CLAUDE_PROJECT_DIR` bricht die Suche ab.
- [ ] `.claude/settings.json` bleibt valides JSON (`jq . .claude/settings.json`).
- [ ] Verhalten innerhalb des Repos unverändert (eine `.ts`/`.md`-Edit wird
      weiterhin über das repo-eigene `node_modules/.bin/prettier` formatiert).

---

## Phase 4: CHANGELOG.md anlegen

### Context

- `cliff.toml` — `header = ""`, Body ohne `## {{ version }}`-Überschrift,
  `commit_parsers` (website-Scope + Nicht-feat/fix werden gefiltert),
  `tag_pattern = "v[0-9]*"`.
- `.github/workflows/release.yml` — Schritt „Changelog generieren"
  (`args: --latest --strip header`), Ausgabe geht nur in Release-Notes/Job-
  Summary.
- Release-Stand: `v0.14.0` ist der einzige bisherige Tag; CHANGELOG.md bekommt
  eine `## [0.14.0]`-Sektion plus `## [Unreleased]` für die Commits seither.

### What to build

Eine stehende `CHANGELOG.md` in der Repo-Wurzel, per git-cliff aus der vollen
History erzeugt, im Keep-a-Changelog-Stil (`# Changelog`-Header,
`## [x.y.z] - YYYY-MM-DD`-Überschriften, `## [Unreleased]`-Sektion). Die
git-cliff-Config so anpassen, dass Header und Versions-Überschriften in der
`CHANGELOG.md` erscheinen, **ohne** die `--latest --strip header`-Ausgabe der
Release-Notes zu verändern. Empfehlung: Header via `[changelog].header`
ergänzen (wird durch `--strip header` aus den Release-Notes entfernt) und die
Versions-Überschrift im Body konditional rendern; falls ein einzelnes Template
beide Ausgaben nicht sauber bedient, eine dedizierte
`cliff.changelog.toml` für die Datei anlegen. Den Regenerierungs-Befehl kurz im
Datei-Kopf oder in einem Kommentar dokumentieren.

### Acceptance criteria

- [ ] `CHANGELOG.md` existiert in der Repo-Wurzel mit `# Changelog`-Header, einer
      `## [0.14.0]`-Sektion und einer `## [Unreleased]`-Sektion; Einträge nach
      Typ gruppiert (Neue Funktionen / Fehlerbehebungen), website-Scope
      herausgefiltert wie in `cliff.toml`.
- [ ] Die Release-Notes-Ausgabe (`git-cliff --latest --strip header` mit der
      wirksamen Config) ist gegenüber vorher **byte-identisch** — per Vor/Nach-
      Vergleich belegt.
- [ ] Der Regenerierungs-Befehl für künftige Releases ist dokumentiert.
- [ ] `make check` grün (keine Backend-/Frontend-Regression durch die neue
      Datei/Config).
