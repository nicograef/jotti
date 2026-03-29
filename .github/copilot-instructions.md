# jotti — Copilot-Anweisungen

jotti ist ein quelloffenes Mobile-Kassensystem (mPOS) für Vereine. Backend: Go, Frontend: React/TypeScript. Vollständige Projektregeln und Konventionen: siehe `AGENTS.md`.

## Wichtigste Regeln

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE.
2. **Geldbeträge immer in Cent (int).** Niemals Floats für Geld.
3. **Event-Sourcing für Kasse-Operationen.** Kassenjournal ist immutable (append-only).
4. **CRUD für Stammdaten** (Benutzer, Produkte, Tische). Soft-Deletes via `status = 'deleted'`.
5. **Validierung mit Schemas.** Backend: `zog`. Frontend: `Zod`.
6. **Deutsche Ubiquitous Language.** Fachbegriffe deutsch (Bestellung, Zahlung, Tisch), Infrastruktur englisch.
7. **Domain-Modelle tragen keine `json`-Tags.** Response-DTOs in der HTTP-Schicht (`api/<domain>/http/`).
8. **Frontend API-Aufrufe nur über Backend-Klassen.** Nie direkt `fetch()`.
9. **Backend ist Single Source of Truth für Daten-Filterung.**
10. **Niemals** `sqlc/dbgen/` editieren (generierter Code).

## Qualitätsprinzipien

- Qualität vor Quantität, Korrektheit vor Geschwindigkeit.
- Self-Review vor dem Präsentieren: korrekt, sauber, lesbar, wartbar, im Scope.
- Nach jeder Aufgabe einen narrativen Zusammenfassungsabsatz für den Reviewer mitliefern.
