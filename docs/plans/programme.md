# Programme: Qualitäts-Durchgang über das ganze Repo

Base: `eb16138a`. Scope: der ganze Code, alle Pläne, PRDs und Dokumente; `docs/rechtsquellen/` bleibt Originaltext. Keine neuen Features, kein Release.

## Waves

1. **Review** (read-only, ein Workflow): sechs Bereiche × fünf Expertenlinsen, je ein `sonnet`-Verifier.
   - Bereiche: `be-kasse` (backend/api/kasse, backend/api/druck, backend/domain), `be-rest` (übriges backend, database), `fe-admin`, `fe-service` (inkl. components, lib, hooks), `edges` (website, e2e, windows, packaging, resolver, scripts, reverse-proxy, compose, Makefile, CI), `docs` (docs/ inkl. plans und prds, Root-Markdown).
   - Linsen: Korrektheit und Sicherheit; Fiskal-Compliance und Freeze-Disziplin; Architektur (Schichten, Cross-Layer-Konsistenz, Modularität, SOLID, Schwellen); Simplify (Reuse, KISS/YAGNI, Effizienz, Altitude, Lesbarkeit); Cleanup (veraltet, ungenutzt, Testkosten, Tests nach der Testregel).
2. **Fix**: Lanes nach Dateibesitz, höchstens zwei schreibende gleichzeitig (5,8 GB RAM).
   - `fix/q-backend` besitzt backend/ und database/; `fix/q-frontend` besitzt frontend/. Beide parallel.
   - `fix/q-edges` besitzt alles unter Wave-1-Bereich `edges`; läuft parallel als dritte Lane mit einem auf ihre Module begrenzten Gate, das volle Gate läuft beim Landen.
   - `fix/q-cross`: Befunde, die Backend und Frontend zugleich ändern; startet nach Landung von backend und frontend.
   - Doku-Zeilen, die ein Code-Befund falsch macht, ändert die Lane des Befunds; übrige Doku-Befunde gehen an Wave 3.
3. **Distill** (`fix/q-distill`): Doku und Kommentare des ganzen Repos auf den Endstand, mit den Doku-Befunden aus Wave 1. FLAGs im Commit-Body.
4. **Distill verify** (`fix/q-verify`): eigener Workflow mit frischen Agenten nach dem Landen von Wave 3. Er prüft jede falsifizierbare Behauptung der Doku gegen Code, Befehle und Upstream, berichtet die FLAGs aus Wave 3 und korrigiert FALSE und STALE.
5. **Abschluss**: CI grün auf main, Programme-Datei entfernt.

## Regeln

- Migrationen: `15` (TSE-Stammdaten je TSS) und `16` (Sitzungen gültig ab) gehören `fix/q-backend`; jede weitere vergibt der Lead in Landereihenfolge.
- Gates: `make check check-repo check-format`, `make lint`; backend zusätzlich `make check-integration` mit `TEST_PG_PORT=5447`; frontend und cross zusätzlich `make test-e2e` mit eigenem `E2E_STACK_PROJECT` und `E2E_STACK_PORT=8091`. Shell-Änderungen: `make check-shell`.
- Port 5434 gehört `gyva-db-1` (anderes Projekt) und bleibt unberührt.
- Entfernen mit externen Konsumenten oder persistierten Daten, Docker- oder Reverse-Proxy-Änderungen und neue Abhängigkeiten gehen vorher an Nico.

## Übergabe

- Nicos Arbeitsrechner fährt alles: Waves 2 bis 5 und danach die Feature-Spur. Eine andere Session beansprucht eine Wave nur, indem sie deren Status-Zeile auf `running` setzt und das pusht.
- `programme-befunde.json` hält die offenen, verifizierten Befunde mit ihrer Lane: `cross`, `docs` und `feature-windows`. Nico hat #135 freigegeben (Idempotenzschlüssel für Tisch- und Direktverkauf-Storno und Umbuchung); Migration 14 macht die `stornierungId` bereits eindeutig.
- Meldet `fix/q-backend` für #52 eine nötige Frontend-Änderung, gehört sie zu `fix/q-cross`.
- Nach dem Abschluss wendet die Session die Dependabot-PRs an. Den Resolver auf jotti.rocks deployt Nico, sobald `fix/q-edges` gelandet ist. Den Windows-Starter (#174, Verzeichnisrechte) testet Nico vor dem Release auf einem echten Windows-PC.

## Feature-Spur

Nach dem Abschluss, nacheinander, je ein Plan mit implement-plan-Lanes. PRDs und Pläne entstehen vorher; Migrationen der Features beginnen bei `17`.

1. **#145 Kommentar je Position**: freigegeben mit den Empfehlungen des PRD (additives Feld als Freeze-Ausnahme mit Zeile in `docs/decisions.md`, kein DSFinV-K-Export, Kommentar auf jedem Stückbon).
2. **Auto-Pfand je Variante**: PRD freigegeben. Der Abholbon lässt Pfand-Positionen in jedem Bonmodus weg, im selben Release; die Regel „gleicher Steuersatz“ gilt ohne Kombi-Sonderfall.
3. **Stationsmodell**: PRD als Variante Z freigegeben (heutige Druckerkonfiguration bleibt Standard, zusätzliche Drucker optional je Produkt); gebaut wird es hier.
4. **Windows-Einzeldatei**: Entscheidungen 2 bis 7 nach Empfehlung freigegeben; Plan für alle Stufen, gebaut wird nur Stufe 1, Messung über einen Windows-Runner der CI, danach entscheidet Nico. Das PRD samt Befund #222 gehört dieser Spur; Distill lässt es aus. Nicos Entscheidungen vom 7. November: Windows 10 22H2 und 11, nur x64; Go für Stufe 1 bei höchstens 120 MB, höchstens 60 s bis zur Anmeldeseite beim ersten und 15 s bei späteren Starts und ohne Defender-Sperre; die Übernahme wird gebaut; jotti läuft als Windows-Dienst. Nach dem Schreiben des Plans stellt die Session Nico die nächste Fragenrunde zum Plan.

## Status

| Phase | Lane | Status | Note |
| --- | --- | --- | --- |
| 1 Review | — | landed | 236 bestätigte Befunde |
| 2 Fix backend | fix/q-backend | landed | c9ee99ec, Migrationen 15 und 16 |
| 2 Fix frontend | fix/q-frontend | landed | fac74e5c |
| 2 Fix edges | fix/q-edges | landed | 07cda43f |
| 2 Fix cross | fix/q-cross | waiting | |
| 3 Distill | fix/q-distill | waiting | |
| 4 Distill verify | fix/q-verify | waiting | |
| 5 Abschluss | — | waiting | |
| F1 #145 Kommentar je Position | feat/kommentar-je-position | waiting | |
| F2 Auto-Pfand je Variante | feat/auto-pfand | waiting | |
| F3 Stationsmodell | feat/stationsmodell | waiting | |
| F4 Windows-Einzeldatei Stufe 1 | feat/windows-einzeldatei | waiting | |
