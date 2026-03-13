# Findings (nicht als Quick Win umgesetzt)

Diese Datei dokumentiert die übrigen Findings aus dem Repo-Quality-Audit, die **nicht** als Quick Win in diesem PR umgesetzt wurden.

## Stage 2 – KISS / Vereinfachung

- **Query-Layer-Indirection reduzieren**  
  Mehrere Query-Wrapper (Application + HTTP-Interfaces) sind weitgehend pass-through und könnten mittelfristig vereinfacht werden.

- **Status-Command-Pattern konsolidieren**  
  Wiederholte Aktivieren/Deaktivieren/Löschen-Logik in User-/Produkt-/Tisch-Command-Services kann in gemeinsame Helper zusammengeführt werden.

- **Test-Mocks in Produktionspfaden prüfen**  
  Einige Mock-Implementierungen liegen in Produktionspaketen; mögliche Verlagerung in `*_test.go` sollte als separater Refactor geprüft werden.

## Stage 3 – Verifikation / Environment

- `make verify` ist in der aktuellen Umgebung blockiert durch fehlende Tools:
  - `golangci-lint` nicht installiert
  - `goimports` nicht installiert
  - `pnpm` nicht installiert

- Backend-Unit-Tests konnten separat erfolgreich laufen:
  - `cd /home/runner/work/jotti/jotti/backend && go test -tags=unit -count=1 ./...`

## Stage 4 – Mobile UX (offene Punkte)

- **Allgemeine Fehlermeldungen in Drawern**  
  Mehrere Drawer verwenden `toast.error('Aktion fehlgeschlagen')`; kontextspezifische Fehlermeldungen wären hilfreicher.

- **Kleine Touch-Targets bei Mengensteuerung**  
  Plus/Minus-Buttons mit `size="icon-sm"` sind auf kleinen Geräten potentiell zu klein.

- **Ladezustands-Kommunikation bei Tabs**  
  Während Reloads in `TablePage` könnten Tabs klarere Rückmeldung geben (z. B. deaktiviert/Loading-Hinweis).

- **Login-Fehlerdarstellung**  
  Gleicher Backend-Fehler wird auf mehrere Felder gelegt; ein einheitlicher Form-Error wäre visuell ruhiger.

## Hinweise

- Diese Liste ist bewusst als Follow-up-Backlog formuliert und enthält keine strukturellen Großrefactorings.
- Alle umgesetzten Quick Wins sind in den regulären Codeänderungen dieses PR enthalten.
