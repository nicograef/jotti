# jotti — Copilot-Anweisungen

jotti ist ein quelloffenes Mobile-Kassensystem (mPOS) für Vereine. Backend: Go, Frontend: React/TypeScript. Vollständige Projektregeln und Konventionen: siehe `AGENTS.md`.

Diese Datei ist absichtlich kurz. `AGENTS.md` ist die kanonische repo-weite Quelle; diese Datei spiegelt nur wenige harte Guardrails und den modus-übergreifenden Bewertungsmaßstab als Sicherheitskopie.

## Bewertungsmaßstab (gilt in jedem Modus: Ask, Plan, Agent)

Jede Änderung wird an **Korrektheit, Einfachheit, Codequalität und Konsistenz** gemessen — und nur daran. **Aufwand, Zeit, Arbeitsumfang, Kosten und Breaking Changes sind nachrangig** und nie ein Gegenargument gegen eine korrekte, einfache, saubere und konsistente Lösung. „Arbeitsumfang“ meint Aufwand, nicht Feature-Scope: keine ungefragten Features, kein Gold-Plating (Scope Guard bleibt). Details: `AGENTS.md` → Qualitätsprinzipien.

## Harte Guardrails

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE.
2. **Geldbeträge immer in Cent (int).** Niemals Floats für Geld.
3. **Event-Sourcing für Kasse-Operationen bleibt append-only.** Einträge im Kassenjournal werden nie aktualisiert oder gelöscht.
4. **Domain-Modelle tragen keine `json`-Tags.** Ausnahme: Event-Data-Structs für Event-Store-Persistenz.
5. **Frontend API-Aufrufe laufen über Backend-Klassen auf Basis des `BackendClient`-Interfaces.** Nie direkt `fetch()` im Domain-Code.
6. **Niemals** `sqlc/dbgen/` editieren (generierter Code).

## Verweis

- Für Workflow, Self-Review, Reviewer-Zusammenfassung und weitere Repo-Regeln gilt `AGENTS.md`.
