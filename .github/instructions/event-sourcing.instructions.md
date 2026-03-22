---
description: "Use when working on event sourcing, domain events, kasse operations, tisch-session state, kassensitzung state, or kassenjournal."
applyTo: "backend/domain/kasse/**,backend/domain/table/**,backend/repository/event_repo/**,backend/repository/kassenjournal_repo/**"
---

# Event-Sourcing-Referenz

Events für Kasse-Operationen (Tisch-Sessions und Kassensitzungen). Zwei Subject-Formate:

- Tisch-Session: `"kassensitzung-{YYYYMMDD}-tisch-{id}"` (z.B. `"kassensitzung-20260501-tisch-42"`)
- Kassensitzung: `"kassensitzung-{YYYYMMDD}"` (z.B. `"kassensitzung-20260501"`)

State wird durch zwei synchrone Projektionen rekonstruiert (`tisch_session_state`, `kassensitzung_state`), die in derselben Transaktion wie das Event-INSERT aktualisiert werden. Routing über expliziten `StreamType`-Parameter.

## Event-Typen

**Tisch-Session Events** (Subject: `kassensitzung-{YYYYMMDD}-tisch-{id}`):

- `bestellung-aufgenommen:v1`
- `zahlung-kassiert:v1`
- `stornierung-erteilt:v1`
- `ausgabe-bestaetigt:v1`
- `auszahlung-geleistet:v1`

**Kassensitzung Events** (Subject: `kassensitzung-{YYYYMMDD}`):

- `kassensitzung-eroeffnet:v1`
- `anfangsbestand-gesetzt:v1`
- `kassenbewegung-gebucht:v1`
- `kassensturz-durchgefuehrt:v1`
- `differenz-soll-ist-gebucht:v1`
- `tagesabschluss-erstellt:v1`

Alle Event-Typen und deren Datenstrukturen: siehe `backend/domain/kasse/` (künftig) bzw. `backend/domain/table/events.go` und die zugehörigen `*Event.go`-Dateien.

## State-Rekonstruktion

- Das Kassenjournal (`kassenjournal`-Tabelle) ist immutable (append-only). Nie Einträge im Kassenjournal updaten oder löschen.
- Saldo = Summe(Bestellungen) − Summe(Zahlungen) − Summe(Stornierungen) + Summe(Auszahlungen)
- UnbezahltePositionen = bestellt − bezahlt − storniert
- AusstehendePositionen = bestellt − ausgegeben − storniert

## Event-Store

Tabelle: `kassenjournal` (append-only). Zwei synchrone Projektionen:

- `tisch_session_state` — session-scoped Tisch-Projektion (PK: `subject`)
- `kassensitzung_state` — Kassensitzung-Projektion (PK: `kassensitzung_nr`)

Routing über `StreamType`-Parameter: `"tisch-session"` | `"kassensitzung"`.

## Weiterführende Dokumentation

- **Invarianten, Event-Strukturen, Projektionen:** [docs/handbuch.md](../../docs/handbuch.md) Kap. 3 (Kasse)
- **Namenskonventionen für Events und Felder:** [docs/language.md](../../docs/language.md)
- **Persistenz-ADR:** [docs/adr/event-sourcing.md](../../docs/adr/event-sourcing.md)
- **CQRS-Projektionen-ADR:** [docs/adr/cqrs.md](../../docs/adr/cqrs.md)

## JSON-Tags in Event-Data-Structs

Event-Data-Structs (zum Beispiel `bestellungAufgenommenV1Data`, `zahlungKassiertV1Data`, `stornierungErteiltV1Data`, `ausgabeBestaetigtV1Data`, `auszahlungGeleistetV1Data`) behalten `json`-Tags fuer `json.Marshal` und `json.Unmarshal` im Event Store. Das ist Persistenz-Serialisierung und kein HTTP-Concern.

Dies ist die einzige erlaubte Ausnahme von der Regel "keine `json`-Tags in `domain/`".
