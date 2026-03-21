---
description: "Use when working on event sourcing, domain events, table operations, event replay, snapshots, or tisch-aggregate state."
applyTo: "backend/domain/table/**,backend/repository/event_repo/**"
---

# Event-Sourcing-Referenz

Events für Tisch-Operationen. Subject-Format: `"tisch:<id>"`. State wird durch Replay aller Events rekonstruiert. Snapshots optimieren Lesezugriffe.

## Event-Typen

- `tisch.bestellung-aufgenommen:v1`
- `tisch.zahlung-kassiert:v1`
- `tisch.stornierung-erteilt:v1`
- `tisch.ausgabe-bestaetigt:v1`
- `tisch.auszahlung-geleistet:v1`

Alle Event-Typen und deren Datenstrukturen: siehe `backend/domain/table/events.go` und die zugehörigen `*Event.go`-Dateien im selben Verzeichnis.

## State-Rekonstruktion

- Events sind immutable (append-only). Nie Events updaten oder löschen.
- Saldo = Summe(Bestellungen) − Summe(Zahlungen) − Summe(Stornierungen) + Summe(Auszahlungen)
- UnbezahltePositionen = bestellt − bezahlt − storniert
- AusstehendePositionen = bestellt − ausgegeben − storniert

## Event-Store

Tabelle: `events` (append-only). Repository: `backend/repository/event_repo/`.

## Weiterführende Dokumentation

- **Invarianten, Event-Strukturen, Replay-Logik:** [docs/handbuch.md](../../docs/handbuch.md) Kap. 3 (Kassenbetrieb)
- **Namenskonventionen für Events und Felder:** [docs/language.md](../../docs/language.md)

## JSON-Tags in Event-Data-Structs

Event-Data-Structs (zum Beispiel `bestellungAufgenommenV1Data`, `zahlungKassiertV1Data`, `stornierungErteiltV1Data`, `ausgabeBestaetigtV1Data`, `auszahlungGeleistetV1Data`) behalten `json`-Tags fuer `json.Marshal` und `json.Unmarshal` im Event Store. Das ist Persistenz-Serialisierung und kein HTTP-Concern.

Dies ist die einzige erlaubte Ausnahme von der Regel "keine `json`-Tags in `domain/`".
