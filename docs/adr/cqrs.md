# ADR: CQRS-Projektionen für Tisch-Zustand

## Status

**Entschieden** — Lazy Projection für operative Queries, Background Worker für analytische Queries.

## Kontext

jotti implementiert CQRS bereits auf der logischen Ebene: Die Application-Schicht trennt `Command`- und `Query`-Structs konsequent, HTTP-Handler sind aufgeteilt, und Repository-Interfaces sind nach Lese-/Schreibzugriff getrennt (Interface Segregation).

| Aspekt                                        | Status       | Anmerkung                               |
| --------------------------------------------- | ------------ | --------------------------------------- |
| Logische Command/Query-Trennung (Application) | ✅ Vorhanden | `Command`- und `Query`-Structs getrennt |
| Separate HTTP-Handler                         | ✅ Vorhanden | `CommandHandler` und `QueryHandler`     |
| Interface Segregation                         | ✅ Vorhanden | Separate Repo-Interfaces                |
| Separates Read Model                          | ❌ Fehlt     | Queries lesen denselben Event Store     |
| Event-getriebene Projektionen                 | ❌ Fehlt     | Kein automatisches Projektion-Update    |
| Separate Datenspeicher                        | ❌ Fehlt     | Single DB für Read und Write            |

jotti befindet sich auf CQRS-Stufe 1 (_Logische Trennung_) und hat klare Ansatzpunkte, um auf Stufe 2 (_Separate Read Models_) zu migrieren. Das [ADR: Event-Sourcing für Tisch-Operationen](event-sourcing.md) identifiziert leseseitige Nachteile, die durch Projektionen adressiert werden.

## Bewertete Alternativen

| Kriterium                              | Synchrone Projektion   | DB-Trigger             | Lazy Projection         | Background Worker         |
| -------------------------------------- | ---------------------- | ---------------------- | ----------------------- | ------------------------- |
| Command bleibt einfach (single INSERT) | ❌ Transaktion nötig   | ✅                     | ✅                      | ✅                        |
| Keine C→Q-Abhängigkeit                 | ❌ Command kennt RM    | ✅                     | ✅                      | ✅                        |
| Kein Backfill nötig                    | ❌ Einmaliger Backfill | ❌ Einmaliger Backfill | ✅ Self-healing         | ⚠️ Worker-Initialisierung |
| Starke Konsistenz                      | ✅ Gleiche Transaktion | ✅ Gleiche Transaktion | ✅ Read-Time-Konsistenz | ❌ Eventual Consistency¹  |
| Snapshot-Eliminierung                  | ✅                     | ✅                     | ✅                      | ✅                        |
| Wartbarkeit / Testbarkeit              | ✅ Go-Code             | ❌ PL/pgSQL            | ✅ Go-Code              | ✅ Go-Code                |
| JSONB-Varianten-Manipulation           | ✅ Go-Logik            | ❌ Komplex in SQL      | ✅ Go-Logik             | ✅ Go-Logik               |
| Infrastruktur-Overhead                 | Gering                 | Gering                 | Gering                  | Mittel (Worker-Lifecycle) |
| Implementierungsaufwand                | Mittel                 | Mittel                 | **Gering**              | Hoch                      |

> ¹ Eventual Consistency ist für **operative Queries** (Saldo, offene Positionen) nicht akzeptabel. Für **analytische Projektionen** (Umsatzauswertungen, Tagesabrechnung) ist sie hingegen ausreichend.

**DB-Trigger** verlagern Business-Logik (Event-Auswertung, JSONB-Varianten-Mapping) nach PL/pgSQL — schwer wartbar und testbar.

## Entscheidung

Zwei Projektionspfade, passend zur Konsistenz-Anforderung:

| Projektion                                                      | Konsistenz           | Ansatz            |
| --------------------------------------------------------------- | -------------------- | ----------------- |
| **Operativ** — Saldo, Unbezahlt, Ungeliefert je Tisch           | Strong Consistency   | Lazy Projection   |
| **Analytisch** — Tagesumsatz, Produktstatistiken, Stornierungen | Eventual Consistency | Background Worker |

### Operative Projektionen — Lazy Projection

Read-Through-Cache: Beim Lesezugriff prüft die Query-Seite, ob die Projektion aktuell ist (`last_event_id`), und replayed fehlende Events bei Bedarf.

**Begründung:**

- **Commands bleiben einfach** — Ein `INSERT INTO events`, kein transaktionaler Overhead. `command.go` bleibt unverändert.
- **Keine C→Q-Abhängigkeit** — Der Command-Service kennt weder `table_state` noch `TableStateRepo`. CQRS-Trennung bleibt rein.
- **Kein Backfill nötig** — Projektion befüllt sich beim ersten Lesezugriff selbst.
- **Selbstheilend** — Fehlerhafte Projektionen werden beim nächsten Read automatisch korrigiert.

Trade-off: Leicht erhöhte Latenz beim ersten Read nach mehreren Writes — für jottis Lastprofil (wenige Events pro Tisch) vernachlässigbar.

### Analytische Projektionen — Background Worker

Für retrospektive Auswertungen (Tagesumsatz, Produktstatistiken, Stornierungsraten). Worker pollt `events`-Tabelle und projiziert neue Events asynchron auf analytische Tabellen (z.B. `daily_revenue`, `variant_sales`). Eventual Consistency ist hier akzeptabel.

Konkret ergeben sich aus dem [Entwickler-Handbuch §7.2](../design/handbuch.md) folgende analytische Read Models:

| Read Model                  | Anforderung                               | Projizierte Daten                                                                                     |
| --------------------------- | ----------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| Tagesabrechnung             | [R-01](../anforderungen.md) (Should-have) | Gesamtumsatz, Umsatz pro Servicekraft, Stornierungsübersicht, offene Beträge                          |
| Abrechnung pro Tisch        | [R-03](../anforderungen.md) (Should-have) | Alle Operationen chronologisch (Bestellungen, Zahlungen, Lieferungen, Stornierungen) mit Gesamt-Saldo |
| Abrechnung pro Servicekraft | [R-04](../anforderungen.md) (Should-have) | Umsatz, Bestellanzahl, Anzahl und Betrag der Stornierungen pro Person                                 |
| Produktumsatz               | [R-05](../anforderungen.md) (Should-have) | Verkaufte Menge pro Variante (abzgl. Stornierungen), Ranking, Gesamteinnahmen                         |

Alle Reporting-Ansichten aggregieren Tisch-Events tischübergreifend und sind nur für Admins zugänglich (vgl. [Anforderungen R-01–R-05](../anforderungen.md)).

Laut [Bounded-Context-Map (Handbuch §2.1–2.2)](../design/handbuch.md) konsumieren zwei Downstream-Kontexte die Projektionen des Kassenbetriebs:

- **Abrechnung** — Read-only-Projektionen der oben genannten analytischen Read Models. Kassenbetrieb → Abrechnung über Published Language (Event-driven): Tisch-Events werden zu Auswertungen projiziert.
- **Ausgabe** — Event-getrieben für Bondruck und Küchendisplay (KDS). Braucht Echtzeit-Zugang zu Bestellungs-Events (Kassenbetrieb → Ausgabe über Published Language). Nicht Teil des MVP, aber architektonisch bereits als Downstream-Kontext vorgesehen.

### Architektur

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Frontend                                   │
└──────┬──────────────────────────────────────────┬───────────────────────┘
       │ Commands (Schreiben)                     │ Queries (Lesen)
       ▼                                          ▼
┌──────────────────┐         ┌────────────────────────────────────────────┐
│ Command Handler  │         │              Query Handler                 │
│ (schreibt Events)│         │  A) Operativ (Strong)  B) Analytisch (Ev.) │
└──────┬───────────┘         └───────────┬────────────────────┬───────────┘
       │ INSERT                          │ SELECT             │ SELECT
       ▼                                 ▼                    ▼
┌──────────────────┐   ┌────────────────────┐   ┌────────────────────────┐
│   events         │   │  table_state       │   │  daily_revenue         │
│   (append-only)  │   │  (operative        │   │  variant_sales         │
│                  │   │   Projektion,      │   │  (analytische          │
│                  │   │   Strong           │   │   Projektionen,        │
│                  │   │   Consistency)     │   │   Eventual             │
└──────────────────┘   └────────────────────┘   │   Consistency)         │
       │                        ▲               └────────────────────────┘
       │    Lazy Projection      │                          ▲
       │    (beim ersten Read)   │               ┌──────────────────────┐
       └────────────────────────┘               │  Background Worker   │
                                                │  (async, z.B. ~30s)  │
                                                └──────────────────────┘
```

### Nicht empfohlen

- **Trigger-basierte Projektion** — Business-Logik in PL/pgSQL schwer wartbar
- **Separate Read-Datenbank** — Overhead nicht gerechtfertigt bei jottis Größe
- **Auflösung des Event Store zugunsten von CRUD** — verliert Audit-Trail-Garantien (siehe [ADR: Event-Sourcing](event-sourcing.md))

## Konsequenzen

**Positiv:** Vereinfachte Query-Seite mit vorberechnetem Zustand, Snapshot-Mechanismus wird durch sauberere Projektion abgelöst, typisierte Projektionen als Enabler für Analytics.

**Negativ:** Zusätzliche Tabelle `table_state` mit Upsert-Logik, leicht erhöhte Read-Latenz beim ersten Zugriff nach Writes, zwei Projektionsmechanismen zu pflegen (Lazy + Worker).

### Priorität

**Mittel** — Kein unmittelbarer Performance-Engpass (Snapshots reichen aktuell), aber klarer Architektur-Gewinn: Snapshot-as-Event ist konzeptuell fragwürdig, eine echte Projektion ist sauberer. Enabler für Zusatzfeatures (Tagesabrechnung, Umsatz pro Produkt).

## Umsetzungsschritte

1. **Migration: `table_state`-Tabelle anlegen** — `database/migrations/XX_table_state.up.sql`
2. **sqlc-Queries definieren** — `UpsertTableState`, `GetTableState` in `backend/sqlc/queries/table_state.sql`, danach `make sqlc`
3. **Repository anlegen** — `backend/repository/table_state_repo/` (Interface + Implementierung, analog zu `event_repo`)
4. **Domain-Logik: `ApplyEventToState()`** — Reine Funktion in `backend/domain/table/projection.go`, kein DB-Zugriff
5. **Query-Umbau: `ensureProjectionUpToDate()`** — In `backend/api/table/application/query.go`: State laden → fehlende Events replayed → State persistieren → zurückgeben
6. **Snapshot ablösen** — `TischSnapshotErstellen()` aus Command-Service entfernen; Snapshot-Event-Typ bleibt für historische Kompatibilität
7. **Tests** — Unit-Tests für `ApplyEventToState()`, Integrationstests für Lazy-Projection-Pfad
