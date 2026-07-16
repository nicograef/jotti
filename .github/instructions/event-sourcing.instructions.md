---
description: "Use when working on event sourcing, domain events, kasse operations, tisch-session state, kassensitzung state, or kassenjournal."
applyTo: "backend/domain/kasse/**,backend/domain/tisch/**,backend/repository/kassenjournal_repo/**"
---

Repo-weite Regeln und Guardrails stehen kanonisch in `AGENTS.md`. Diese Datei ergänzt nur event-sourcing-spezifische Details für Kasse-Streams und Projektionen.

# Event-Sourcing-Referenz

Events für Kasse-Operationen (Tisch-Sessions, Direktverkäufe und Kassensitzungen). Drei Subject-Formate:

- Tisch-Session: `"kassensitzung-{nr}/tisch-{id}"` (z.B. `"kassensitzung-1/tisch-42"`)
- Direktverkauf: `"kassensitzung-{nr}/direktverkauf-{uuid}"` (ein Stream pro Barverkauf)
- Kassensitzung: `"kassensitzung-{nr}"` (z.B. `"kassensitzung-1"`)

## Event-Typen

**Tisch-Session Events** (Subject: `kassensitzung-{nr}/tisch-{id}`):

- `bestellung-aufgenommen:v1`
- `zahlung-kassiert:v1`
- `stornierung-erteilt:v1` (kassenwirksame Warenrücknahme bezahlter Positionen)
- `bestellung-korrigiert:v1` (geldneutrale Korrektur unbezahlter Positionen)
- `bestellung-umgebucht:v1` (geldneutrale Umbuchung zwischen Tischen)

**Direktverkauf Events** (Subject: `kassensitzung-{nr}/direktverkauf-{uuid}`):

- `direktverkauf-getaetigt:v1` (immer Version 1 des Streams)
- `direktverkauf-storniert:v1`

**Kassensitzung Events** (Subject: `kassensitzung-{nr}`):

- `kassensitzung-eroeffnet:v1` (enthält den Anfangsbestand — kein eigenes Event)
- `geldtransit-gebucht:v1`
- `kassensturz-durchgefuehrt:v1`
- `differenz-soll-ist-gebucht:v1`
- `tagesabschluss-erstellt:v1`

Alle Event-Typen und deren Datenstrukturen: `backend/domain/kasse/tisch_session_events.go` (Tisch-Session) und `backend/domain/kasse/kassensitzung_events.go` (Kassensitzung).

## State-Rekonstruktion

- Saldo = Summe(Bestellungen) − Summe(Zahlungen) − Summe(Korrekturen) ± Summe(Umbuchungen); stets ≥ 0. Die Warenrücknahme (`stornierung-erteilt`) verändert den Saldo nicht (wirkt nur auf den Kassenbestand).
- UnbezahltePositionen = bestellt − bezahlt − korrigiert − umgebucht (Abgang) + umgebucht (Zugang)

## Event-Store

Tabelle: `kassenjournal` (append-only). Projektion und Kassensitzungs-Entität werden in derselben Transaktion wie das Event-INSERT aktualisiert:

- `tisch_sessions` — session-scoped Tisch-Projektion (PK: `subject`)
- `kassensitzungen` — Kassensitzung-Entität (CRUD, PK: `z_nr`)

Routing über `StreamType`-Parameter: `"tisch-session"` | `"kassensitzung"` | `"direktverkauf"` (ohne Projektion).

## Weiterführende Dokumentation

- **Invarianten, Event-Strukturen, Projektionen:** [docs/handbuch.md](../../docs/handbuch.md) Kap. 3 (Kasse)
- **Namenskonventionen für Events und Felder:** [docs/language.md](../../docs/language.md)

## JSON-Tags in Event-Data-Structs

Event-Data-Structs (zum Beispiel `BestellungAufgenommenV1Data`, `ZahlungKassiertV1Data`, `StornierungErteiltV1Data`, `KassensitzungEroeffnetV1Data` u. a.) behalten `json`-Tags fuer `json.Marshal` und `json.Unmarshal` im Event Store. Das ist Persistenz-Serialisierung und kein HTTP-Concern.

Dies ist die einzige erlaubte Ausnahme von der Regel "keine `json`-Tags in `domain/`".
