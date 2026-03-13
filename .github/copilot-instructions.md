# jotti — Copilot-Anweisungen

Dieses Projekt ist ein Mobile-Kassensystem (mPOS) für Vereine. Vollständige Agent-Anweisungen: siehe `AGENTS.md` im Projekt-Root.

## Universelle Regeln

1. **Alle API-Endpunkte sind POST-only.** Keine GET/PUT/DELETE.
2. **Geldbeträge sind immer in Cent (int).** Niemals Floats für Geld.
3. **Events sind immutable.** Nie Events updaten oder löschen.
4. **Validierung auf beiden Seiten.** Backend: `zog`. Frontend: `Zod`.
5. **Deutsche Ubiquitous Language.** Fachbegriffe deutsch (Bestellung, Zahlung, Tisch, Position). Infrastruktur-Code englisch. Alle Benutzer-sichtbaren Strings auf Deutsch.
6. **Frontend API-Aufrufe nur über Backend-Klassen.** Nie direkt `fetch()`. Alle Domain-Backend-Klassen nutzen `BackendClient` aus `src/lib/Backend.ts`.
7. **`sqlc/dbgen/` nie editieren** — generierter Code.
8. **Keine Secrets oder Passwörter in Code committen.**
9. **Domain-Modelle ohne `json`-Tags.** Domain-Structs in `domain/` tragen keine `json`-Tags. HTTP-Responses verwenden dedizierte DTOs in `api/<domain>/http/` mit Mapper-Funktionen. Einzige Ausnahme: Event-Data-Structs fuer Event-Store-Serialisierung.

## Befehle

Alle Befehle über **Makefile** im Root: `make test`, `make lint`, `make build`, `make check`, `make check-full`, `make verify`, `make sqlc`, `make dev`. Siehe `make help`.

## Referenzdokumente

Die folgenden Dokumente beschreiben jotti vollständig. Sie werden **nicht automatisch geladen** (zu groß). Bevor du eine Aufgabe beginnst, prüfe ob du Kontext aus einem dieser Dokumente brauchst — und lies dann gezielt den relevanten Abschnitt, nicht das ganze Dokument.

| Dokument                | Pfad                          | Wann lesen?                                                                                 |
| ----------------------- | ----------------------------- | ------------------------------------------------------------------------------------------- |
| **Anforderungen**       | `docs/anforderungen.md`       | Neue Features implementieren, Akzeptanzkriterien prüfen, Rollen/Berechtigungen klären       |
| **Handbuch**            | `docs/design/handbuch.md`     | Architekturentscheidungen, Invarianten, Event-Strukturen, Schichtenarchitektur, Read Models |
| **Ubiquitous Language** | `docs/design/language.md`     | Benennungen, Namenskonventionen pro Schicht, Ist/Soll-Abweichungen, Begriffsklärung         |
| **Produktbeschreibung** | `docs/produktbeschreibung.md` | Positionierung, Zielgruppe, Abgrenzung, README/Marketing-Texte                              |
