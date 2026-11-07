# Programme: Qualitäts-Durchgang über das ganze Repo

Base: `eb16138a`. Scope: der ganze Code, alle Pläne, PRDs und Dokumente; `docs/rechtsquellen/` bleibt Originaltext. Keine neuen Features, kein Release.

## Waves

1. **Review** (read-only, ein Workflow): sechs Bereiche × fünf Expertenlinsen, je ein `sonnet`-Verifier.
   - Bereiche: `be-kasse` (backend/api/kasse, backend/api/druck, backend/domain), `be-rest` (übriges backend, database), `fe-admin`, `fe-service` (inkl. components, lib, hooks), `edges` (website, e2e, windows, packaging, resolver, scripts, reverse-proxy, compose, Makefile, CI), `docs` (docs/ inkl. plans und prds, Root-Markdown).
   - Linsen: Korrektheit und Sicherheit; Fiskal-Compliance und Freeze-Disziplin; Architektur (Schichten, Cross-Layer-Konsistenz, Modularität, SOLID, Schwellen); Simplify (Reuse, KISS/YAGNI, Effizienz, Altitude, Lesbarkeit); Cleanup (veraltet, ungenutzt, Testkosten, Tests nach der Testregel).
2. **Fix**: Lanes nach Dateibesitz, höchstens zwei schreibende gleichzeitig (5,8 GB RAM).
   - `fix/q-backend` besitzt backend/ und database/; `fix/q-frontend` besitzt frontend/. Beide parallel.
   - `fix/q-edges` besitzt alles unter Wave-1-Bereich `edges`; startet, sobald eine der beiden ersten Lanes gelandet ist.
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

## Status

| Phase | Lane | Status | Note |
| --- | --- | --- | --- |
| 1 Review | — | landed | 236 bestätigte Befunde |
| 2 Fix backend | fix/q-backend | running | |
| 2 Fix frontend | fix/q-frontend | running | |
| 2 Fix edges | fix/q-edges | waiting | |
| 2 Fix cross | fix/q-cross | waiting | |
| 3 Distill | fix/q-distill | waiting | |
| 4 Distill verify | fix/q-verify | waiting | |
| 5 Abschluss | — | waiting | |
