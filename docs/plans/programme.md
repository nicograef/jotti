# Programme: Qualitäts-Durchgang und Feature-Spur

Der Qualitäts-Durchgang über das ganze Repo (Waves 1 bis 5, Basis `eb16138a`) ist gelandet. Offen ist die Feature-Spur. Diese Datei bleibt, bis sie durch ist, weil ihre Status-Zeilen hier stehen.

## Regeln

- Nicos Arbeitsrechner fährt alles; keine andere Session arbeitet am Programm.
- Je ein Plan nach dem anderen mit `/nico:implement-plan`: Run-Worktree `../jotti-wt/plan/<slug>`, Branch `plan/<slug>`, Commit je Kriterium, Review und Tick je Phase.
- Migrationen 15 bis 17 sind gelandet. Die Features nehmen 18 (Auto-Pfand) und 19 (Zusatzdrucker); #145 braucht keine.
- Gates: `make check check-repo check-format`, `make lint`, `make check-integration` mit `TEST_PG_PORT=5447`, `make test-e2e` mit eigenem `E2E_STACK_PROJECT` und `E2E_STACK_PORT=8091`. Shell-Änderungen: `make check-shell`.
- Höchstens zwei schreibende Lanes gleichzeitig (5,8 GB RAM).
- Entfernen mit externen Konsumenten oder persistierten Daten, Docker- oder Reverse-Proxy-Änderungen und neue Abhängigkeiten gehen vorher an Nico, außer der Plan nennt seine Freigabe.
- Beim Landen eines Plans: Plan und PRD per `git rm`, eine Zeile im noch unveröffentlichten Abschnitt `[1.0.0]` von `CHANGELOG.md` wie bei der Pfandrückgabe, Status-Zeile auf `landed`.

## Feature-Spur

1. **#145 Kommentar je Position** (gelandet): freigegeben mit den Empfehlungen des PRD. Additives Feld als Freeze-Ausnahme mit D22, kein DSFinV-K-Export, Kommentar auf jedem Stückbon, AGENTS.md-Freeze-Satz umformuliert.
2. **Auto-Pfand je Variante** (`plan-auto-pfand-je-variante.md`): Der Abholbon lässt Pfand-Positionen in jedem Bonmodus weg, im selben Release; die Regel „gleicher Steuersatz“ gilt ohne Kombi-Sonderfall.
3. **Zusatzdrucker** (`plan-zusatzdrucker.md`, Stationsmodell-PRD als Variante Z): Die heutige Druckerkonfiguration bleibt Standard, zusätzliche Drucker sind optional je Produkt.
4. **Windows-Einzeldatei Stufe 1** (`plan-windows-einzeldatei.md`, Phasen 1 bis 6): Windows 10 22H2 und 11, nur x64; jotti läuft als Windows-Dienst; die Übernahme wird gebaut. Phase 6 entscheidet automatisch über die Messung auf dem Windows-Runner der CI: Go bei höchstens 120 MB, höchstens 60 s bis zur Anmeldeseite beim ersten und 15 s bei späteren Starts und ohne Defender-Sperre. Reißt ein Wert auch nach den Abhilfen des Plans, entscheidet Nico.

## Offen bei Nico

- Den Windows-Starter (#174, Verzeichnisrechte) vor dem Release auf einem echten Windows-PC testen.

## Status

| Phase | Lane | Status | Note |
| --- | --- | --- | --- |
| 1 Review | — | landed | 236 bestätigte Befunde |
| 2 Fix backend | fix/q-backend | landed | c9ee99ec, Migrationen 15 und 16 |
| 2 Fix frontend | fix/q-frontend | landed | fac74e5c |
| 2 Fix edges | fix/q-edges | landed | 07cda43f; Resolver auf jotti.rocks am 7. November deployt |
| 2 Fix cross | fix/q-cross | landed | 46a1fd0c, Migration 17 |
| 3 Distill | fix/q-distill | landed | b908ad0d, 42 FLAGs im Commit d5fd6dd0 |
| 4 Distill verify | fix/q-verify | landed | b0f6993c: 728 Behauptungen, 53 korrigiert, 69 UNREACHED |
| 5 Abschluss | — | landed | Dependabot 08811709, 19d2c26c und c4b5db24 (`@hookform/resolvers` 7), Folgebefunde a8eeede3 |
| F1 #145 Kommentar je Position | plan/kommentar-je-position | landed | fca378c1, D22 |
| F2 Auto-Pfand je Variante | plan/auto-pfand-je-variante | running | Migration 18 |
| F3 Zusatzdrucker | plan/zusatzdrucker | waiting | Migration 19 |
| F4 Windows-Einzeldatei Stufe 1 | plan/windows-einzeldatei | waiting | Phasen 1 bis 6 |
