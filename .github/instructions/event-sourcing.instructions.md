---
description: "Use when working on event sourcing, domain events, table operations, event replay, snapshots, or tisch-aggregate state."
applyTo: "backend/domain/table/**,backend/repository/event_repo/**"
---

# Event-Sourcing-Referenz

Events für Tisch-Operationen. Subject-Format: `"tisch:<id>"`. State wird durch Replay aller Events rekonstruiert. Snapshots optimieren Lesezugriffe.

## Event-Typen

- `tisch.bestellung-aufgegeben:v1`
- `tisch.zahlung-registriert:v1`
- `tisch.produkte-storniert:v1`
- `tisch.produkte-geliefert:v1`
- `tisch.snapshot:v1`

Alle Event-Typen und deren Datenstrukturen: siehe `backend/domain/table/events.go` und die zugehörigen `*Event.go`-Dateien im selben Verzeichnis.

## State-Rekonstruktion

- Events sind immutable (append-only). Nie Events updaten oder löschen.
- Saldo = Summe(Bestellungen) − Summe(Zahlungen) − Summe(Stornierungen)
- UnbezahltePositionen = bestellt − bezahlt − storniert
- UngeliefertePositionen = bestellt − geliefert − storniert

## Event-Store

Tabelle: `events` (append-only). Repository: `backend/repository/event_repo/`.

## Weiterführende Dokumentation

- **Invarianten, Event-Strukturen, Replay-Logik:** [docs/design/handbuch.md](../../docs/design/handbuch.md) Kap. 3 (Kassenbetrieb)
- **Namenskonventionen für Events und Felder:** [docs/design/language.md](../../docs/design/language.md)

## JSON-Tags in Event-Data-Structs

Event-Data-Structs (zum Beispiel `bestellungAufgegebenV1Data`, `zahlungRegistriertV1Data`, `produkteStorniertV1Data`, `produkteGeliefertV1Data`) behalten `json`-Tags fuer `json.Marshal` und `json.Unmarshal` im Event Store. Das ist Persistenz-Serialisierung und kein HTTP-Concern.

Dies ist die einzige erlaubte Ausnahme von der Regel "keine `json`-Tags in `domain/`".
