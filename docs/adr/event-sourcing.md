# ADR: Event-Sourcing für Tisch-Operationen

## Status

**Entschieden** — Event-Sourcing für Tisch-Operationen, CRUD für Stammdaten.

## Kontext

Servicekräfte nehmen Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Frage: Wie werden diese Tisch-Operationen persistiert?

**Verworfene Alternative:** Reines CRUD mit 8 Tabellen (4 Header + 4 Items), Transaktionen für Schreiboperationen, komplexe SQL-Aggregation für Leseoperationen.

## Entscheidung

Tisch-Operationen werden als immutable Events in einer `events`-Tabelle gespeichert. Stammdaten (Benutzer, Produkte, Tische) verbleiben in CRUD-Tabellen.

## Begründung

1. **Audit-Trail strukturell gegeben** — Append-Only-Events machen nachvollziehbar, wer wann was bestellt/bezahlt/storniert hat.
2. **Immutabilität schützt vor Manipulation** — DB-Trigger verhindern UPDATE/DELETE auf Events. Kritisch für ein System, das Geldbeträge verwaltet.
3. **Begrenzte Komplexität** — Event-Sourcing nur für Tisch-Operationen, nicht für das gesamte System.
4. **Schema-Flexibilität** — Neue Features (Rabatte, Trinkgeld) erfordern nur neue Event-Typen statt DB-Migrationen.

## Akzeptierte Nachteile

- Keine referenzielle Integrität auf DB-Ebene für Tisch-Operationen (Validierung in der Anwendung via zog)
- JSONB nicht typsicher auf DB-Ebene
- Höhere Einstiegshürde (Zustandsrekonstruktion, Snapshots, Event-Versionierung)
- Ad-hoc-Analysen erfordern JSONB-Parsing statt einfachem SQL

## Implementierung

### Event-Modell (angelehnt an [CloudEvents](https://cloudevents.io/))

```go
type Event struct {
    ID      int              `json:"id"`
    UserID  int              `json:"userId"`
    Type    string           `json:"type"`      // z.B. "tisch.bestellung-aufgegeben:v1"
    Time    time.Time        `json:"time"`
    Subject string           `json:"subject"`   // z.B. "tisch:42"
    Data    json.RawMessage  `json:"data"`
}
```

### Event-Typen

| Event-Typ                        | Beschreibung          |
| -------------------------------- | --------------------- |
| `tisch.bestellung-aufgegeben:v1` | Bestellung aufgegeben |
| `tisch.zahlung-registriert:v1`   | Zahlung registriert   |
| `tisch.produkte-storniert:v1`    | Positionen storniert  |
| `tisch.produkte-geliefert:v1`    | Positionen geliefert  |
| `tisch.snapshot:v1`              | Zustandssnapshot      |

### Append-Only-Garantie

- **Privilege Revocation**: Nur SELECT und INSERT erlaubt.
- **DB-Trigger** gegen UPDATE/DELETE/TRUNCATE.

### Snapshots

Snapshots sind selbst Events (`tisch.snapshot:v1`) mit berechnetem Zustand. `ReadEventsWithSnapshot()` lädt den letzten Snapshot + nachfolgende Events in einer SQL-Abfrage.

Diese Snapshot-Implementierung als Event ist eine bewusste Vereinfachung (vgl. [Handbuch §3.4](../design/handbuch.md#34-event-replay-und-snapshots)). Das [ADR: CQRS](cqrs.md) plant die Ablösung durch Lazy Projection, bei der Snapshots automatisch als materialisierte Sicht verwaltet werden.

### CQRS

- **Commands** erstellen Events: `BestellungAufgeben`, `ZahlungRegistrieren`, `ProdukteStornieren`, `ProdukteLiefern`, `TischSnapshotErstellen`
- **Queries** rekonstruieren Zustand: `GetTischSaldo`, `GetTischHistorie`, `GetTischUnbezahlt`, `GetTischUngeliefert`, `GetGesamtZahlungenFromEvents`
