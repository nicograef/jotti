# ADR: Persistenz-Strategie für Tisch-Operationen

## Status

**Entschieden** — Event Sourcing für Tisch-Operationen + Synchrone Projektion. CRUD für Stammdaten und Auth. Ersetzt vorherige ADR (2025-03-13).

## Kontext

Servicekräfte nehmen Bestellungen auf, liefern aus, kassieren und stornieren — alles pro Tisch. Diese vier Kernoperationen (K-01 bis K-04) bilden den Kassenbetrieb, jottis Core Domain. Die zentrale Architekturentscheidung: **Wie werden Tisch-Operationen persistiert?**

### Lastprofil

| Kennzahl                         | Wert           | Quelle                                              |
| -------------------------------- | -------------- | --------------------------------------------------- |
| Tische pro Veranstaltung         | 5–50           | [Produktbeschreibung §3](../produktbeschreibung.md) |
| Events pro Tisch (geschätzt)     | ~50–200        | [Handbuch §3.4](../design/handbuch.md)              |
| Gesamte Events pro Veranstaltung | < 10.000       | Berechnung: 50 × 200                                |
| Gleichzeitige Benutzer           | 5–30           | [Produktbeschreibung §3](../produktbeschreibung.md) |
| Veranstaltungen pro Jahr         | 2–3            | [Produktbeschreibung §3](../produktbeschreibung.md) |
| Betriebsdauer pro Veranstaltung  | Wenige Stunden | [Produktbeschreibung §4](../produktbeschreibung.md) |

**Konsequenz:** Performance ist kein Entscheidungskriterium. Bei < 10.000 Events ist Event Replay in Mikrosekunden erledigt. Bei 8 CRUD-Tabellen mit < 10.000 Zeilen ist jede SQL-Aggregation trivial. Die Entscheidung muss auf **fachliche Passung, Korrektheit und Wartbarkeit** gestützt werden — nicht auf Skalierung.

### Anforderungen als Entscheidungstreiber

**Direkt betroffen:**

| Anforderung                                                                                                             | Relevanz                                                                                                         |
| ----------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| **K-01–K-04** (Bestellen, Zahlen, Liefern, Stornieren) — „wird als unveränderliches Event im Kassenjournal gespeichert" | Alle vier Kernoperationen fordern unveränderliche Persistenz.                                                    |
| **K-06** (Kassenjournal) — „Events sind unveränderlich, Tischzustand wird aus Event Stream berechnet"                   | K-06 setzt Event Sourcing als Implementierung voraus (→ siehe Transparenzhinweis unten).                         |
| **Q-04** (Datenintegrität) — „Tischzustand ausschließlich aus Event Stream berechnet"                                   | Setzt ES als Implementierung voraus (→ siehe Transparenzhinweis unten).                                          |
| **Q-02** (Mehrbenutzerfähigkeit) — „Gleichzeitige Operationen am selben Tisch ohne Datenverlust"                        | Erfordert Concurrency Control. ES: OCC über `(subject, version)`. CRUD: Multi-Table-Locking oder Row-Versioning. |

**Indirekt beeinflusst:**

| Anforderung                                                                  | Auswirkung                                                                                                                |
| ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **K-05** (Tischübersicht) — Saldo und Zähler für alle aktiven Tische         | ES ohne Projektion: N separate Event-Replays. CRUD: 1 aggregierendes SELECT. ES + Projektion: 1 SELECT auf `table_state`. |
| **R-01–R-05** (Reporting) — Tagesabrechnung, Umsatz pro Produkt/Servicekraft | ES: JSONB-Parsing über alle Events. CRUD: Standard-SQL-Aggregation. ES + Projektion: Hybrid.                              |

> **Transparenzhinweis — Zirkuläre Abhängigkeit:** Die Anforderungen K-06 und Q-04 formulieren Event Sourcing als _Anforderung_, nicht als _Lösung_. Diese ADR begründet die Entscheidung unter anderem mit diesen Anforderungen. Das ist zirkulär. Die Entscheidung stützt sich daher primär auf die fachliche Analyse (§ Begründung), nicht auf K-06/Q-04. Diese beiden Anforderungen werden als _konsistent mit_ der Entscheidung gewertet, nicht als _Treiber_ der Entscheidung.

### ACL-Problem (Anti-Corruption Layer)

Produktpreise und -namen können sich zwischen Bestellung und Abrechnung ändern. Der Kassenbetrieb muss sich gegen Stammdaten-Änderungen schützen:

- **ES-Lösung:** Fat Events — Produktdaten zum Bestellzeitpunkt im Event eingefroren. Explizit, sichtbar, Teil des Domänenmodells.
- **CRUD-Lösung:** Denormalisierte Spalten (`produkt_name`, `variante_name`, `einzelpreis`) in der Positions-Tabelle. Funktional identisch, aber weniger offensichtlich.

Beide Lösungen funktionieren. Der Unterschied ist konzeptuelle Klarheit, nicht Korrektheit.

---

## Entscheidung

**Event Sourcing für Tisch-Operationen, erweitert um synchrone Projektion (CQRS Stufe 2). CRUD für Stammdaten und Auth.**

Tisch-Operationen werden als immutable Events in einer `events`-Tabelle gespeichert. Eine synchrone `table_state`-Projektion wird in derselben Transaktion wie das Event-INSERT aktualisiert und dient als optimiertes Read-Modell.

---

## Bewertete Alternativen

### Option A: CRUD mit normalisierten Tabellen

**Schema-Skizze (8 Tabellen):**

```sql
-- Vorgangstabellen (je Header + Items)
bestellungen       (id, tisch_id, user_id, user_name, kommentar, created_at)
positionen         (id, bestellung_id, variante_id, produkt_name, variante_name,
                    kategorie, einzelpreis, menge)
                    -- produkt_name, variante_name, einzelpreis = denormalisierte Kopie (Fat Data)

zahlungen          (id, tisch_id, user_id, user_name, kommentar, gesamt_cents, created_at)
zahlung_positionen (zahlung_id, position_id, menge)

lieferungen        (id, tisch_id, user_id, user_name, kommentar, created_at)
lieferung_positionen (lieferung_id, position_id, menge)

stornierungen      (id, tisch_id, user_id, user_name, kommentar, gesamt_cents, created_at)
stornierung_positionen (stornierung_id, position_id, menge)
```

**Saldo-Berechnung:**

```sql
SELECT
    COALESCE(SUM(p.einzelpreis * p.menge), 0)
  - COALESCE((SELECT SUM(gesamt_cents) FROM zahlungen WHERE tisch_id = 42), 0)
  - COALESCE((SELECT SUM(gesamt_cents) FROM stornierungen WHERE tisch_id = 42), 0)
AS saldo_cents
FROM bestellungen b
JOIN positionen p ON p.bestellung_id = b.id
WHERE b.tisch_id = 42;
```

**Kassenjournal (K-06) als UNION:**

```sql
SELECT 'bestellung' AS typ, b.created_at, b.user_name, ...
FROM bestellungen b WHERE b.tisch_id = 42
UNION ALL
SELECT 'zahlung' AS typ, z.created_at, z.user_name, ...
FROM zahlungen z WHERE z.tisch_id = 42
UNION ALL
SELECT 'lieferung' AS typ, l.created_at, l.user_name, ...
FROM lieferungen l WHERE l.tisch_id = 42
UNION ALL
SELECT 'stornierung' AS typ, s.created_at, s.user_name, ...
FROM stornierungen s WHERE s.tisch_id = 42
ORDER BY created_at ASC;
```

**OCC-Strategie:** Kein einheitlicher Versionszähler pro Tisch. Mögliche Workarounds: separate `tisch_version`-Spalte in `tische` (UPDATE bei jedem Vorgang → Serialisierungspunkt), optimistisches Locking über Positionsstatus (komplex), oder UNIQUE-Constraints pro Tabelle.

**Vorteile:**

- ✅ Referentielle Integrität über FK-Constraints (DB validiert Beziehungen)
- ✅ Typisierte Spalten (kein JSONB → `sqlc` generiert typsichere Structs)
- ✅ Triviale SQL-Aggregation für Reporting (R-01–R-05)
- ✅ Standard-CRUD-Wissen — niedrige Einstiegshürde für Contributors
- ✅ Ad-hoc-Analysen mit Standard-SQL möglich

**Nachteile:**

- ❌ **Kein natürlicher Audit Trail.** Vorgangstabellen sind INSERT-only, aber es fehlt ein einheitliches „wer hat wann was getan"-Log. Separate `audit_log`-Tabelle nötig → zwei Mechanismen.
- ❌ **Kassenjournal = UNION über 4 Tabellen.** Jeder neue Vorgangstyp erfordert Schema-Erweiterung UND Query-Erweiterung.
- ❌ **OCC über mehrere Tabellen.** Kein natürlicher Serialisierungspunkt pro Tisch.
- ❌ **8+ Tabellen** statt 1. Mehr Schema, mehr Migrations, mehr Repository-Code, mehr sqlc-Queries.
- ❌ **Denormalisierte Produktdaten** (Fat Data) funktional identisch zu Fat Events, aber weniger explizit als Architekturentscheidung.
- ❌ **Widerspricht K-06 und Q-04 direkt.** Die Anforderungen müssten umformuliert werden.

**Gesamtbewertung:** Technisch machbar. Für jottis Domäne (Kassensystem mit Buchführungs-Charakter) erzwingt CRUD ein zustandsbasiertes Modell auf eine vorgangsbasierte Domäne. Die Vereinfachung auf SQL-Ebene wird durch Audit-Komplexität, OCC-Probleme und Schema-Fragmentierung an anderer Stelle erkauft.

### Option B: Event Sourcing ohne Projektion (vorheriger Ist-Zustand)

**Schema:** 1 Events-Tabelle (`events`) mit JSONB-Payload. Event-Typen: `BestellungAufgegeben`, `ZahlungRegistriert`, `ProdukteGeliefert`, `ProdukteStorniert`. (Ehemals auch `tisch.snapshot` — abgelöst durch synchrone Projektion, siehe [ADR: CQRS](cqrs.md).)

| Aspekt               | Implementierung                                                                       |
| -------------------- | ------------------------------------------------------------------------------------- |
| State Reconstruction | Synchrone Projektion via `ApplyEvent()` in `table_state` (siehe [ADR: CQRS](cqrs.md)) |
| OCC                  | UNIQUE Constraint `(subject, version)` + `GetMaxVersion()` + Retry bei Conflict       |
| Snapshot             | `tisch.snapshot:v1` als Event-Typ im Event Stream                                     |
| Read-Optimization    | `ReadEventsWithSnapshot()` — lädt letzten Snapshot + nachfolgende Events              |
| Kassenjournal        | `ReadEventsBySubject()` — 1 Query, chronologisch sortiert                             |

**Vorteile:**

- ✅ Natürliches Domänenmodell — Geschäftsvorfälle als First-Class-Citizens
- ✅ Unveränderlichkeit strukturell garantiert (DB-Trigger: `prevent_event_mutation()`)
- ✅ Audit Trail = Event Stream — ein einziger Mechanismus
- ✅ Minimale Schema-Komplexität (1 Tabelle für den gesamten Kassenbetrieb)
- ✅ Fat Events lösen ACL-Problem explizit und sichtbar
- ✅ OCC trivial über eine Spalte und einen UNIQUE Constraint
- ✅ Kassenjournal = 1 Query

**Nachteile:**

- ❌ **JSONB ohne DB-Typsicherheit.** Event-Payloads sind unstrukturiert auf DB-Ebene. Validierung nur in der Anwendung (zog).
- ❌ **Event Replay bei jedem Read.** K-05 (Tischübersicht mit 50 Tischen) = 50 separate Replays pro Seitenaufruf.
- ❌ **Snapshot-as-Event ist ein Anti-Pattern.** Infrastruktur-Artefakt (`tisch.snapshot:v1`) im fachlichen Event Stream. Vermischt Domänen-Events mit Optimierungs-Artefakten.
- ❌ **Ad-hoc-SQL-Analysen** erfordern JSONB-Parsing. Nicht trivial für Reporting.
- ❌ **Höhere Einstiegshürde.** Event Replay, OCC-Versionierung — Konzepte, die über Standard-CRUD hinausgehen.
- ❌ **Keine referentielle Integrität** auf DB-Ebene.

**Gesamtbewertung:** Fachlich richtig, Read-Seite hat praktische Schwächen. Der Snapshot-as-Event-Mechanismus ist ein konzeptuelles Problem, das die Klarheit des Event Streams untergräbt.

### Option C: Event Sourcing + Synchrone Projektion (gewählt)

**Schema:** Events-Tabelle als Source of Truth + `table_state`-Projektionstabelle. UPSERT auf `table_state` in derselben Transaktion wie Event-INSERT.

**Architektur-Skizze:**

```
Command (Schreiben)                    Query (Lesen)
       │                                      │
       ▼                                      ▼
┌──────────────────┐              ┌─────────────────────────┐
│ EventRepo.       │              │ TableStateRepo.         │
│ WriteEvent()     │              │ GetState()              │
│                  │              │                         │
│ BEGIN TX         │              │ SELECT * FROM           │
│  INSERT event    │              │   table_state           │
│  UPSERT state ←──── synchron   │ WHERE tisch_id = ?      │
│ COMMIT TX        │              └─────────────────────────┘
└──────────────────┘
```

**table_state-Schema (Skizze):**

```sql
CREATE TABLE table_state (
    tisch_id                  INT PRIMARY KEY REFERENCES tische(id),
    saldo_cents               INT NOT NULL DEFAULT 0,
    last_event_id             INT NOT NULL REFERENCES events(id),
    last_event_version        INT NOT NULL,
    unbezahlte_positionen     JSONB NOT NULL DEFAULT '[]',
    ungelieferte_positionen   JSONB NOT NULL DEFAULT '[]',
    gesamt_zahlungen_cents    INT NOT NULL DEFAULT 0,
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Wie die Projektion funktioniert:**

1. Command-Service ruft `EventRepo.WriteEvent(event)` auf (unverändert)
2. Repository-intern: `BEGIN TX → INSERT INTO events → ApplyEvent(state, event) → UPSERT INTO table_state → COMMIT TX`
3. Die Apply-Funktion (`ApplyEventToState()`) ist eine reine Funktion in der Domain-Schicht — kein DB-Zugriff
4. Der Command-Service kennt weder `table_state` noch den Projektor — CQRS-Trennung bleibt intakt
5. Query-Service liest direkt aus `table_state` — kein Event Replay mehr nötig

**Vorteile:**

- ✅ **Alle ES-Vorteile bleiben erhalten** (Audit Trail, Immutabilität, natürliches Domänenmodell, OCC, Fat Events)
- ✅ **K-05 (Tischübersicht) wird trivial:** `SELECT * FROM table_state` statt N Event-Replays
- ✅ **Snapshot-as-Event wird eliminiert** — `tisch.snapshot:v1` muss nicht mehr erzeugt werden
- ✅ **Reporting profitiert** — Saldo und Positionen sind vorberechnet, kombinierbar mit `events`-Tabelle
- ✅ **Strong Consistency** — Projektion in derselben TX wie Event-INSERT, kein Stale State
- ✅ **Selbstheilend** — Bei Inkonsistenz kann `table_state` jederzeit aus Events reberechnet werden

**Nachteile:**

- ❌ Minimal komplexerer Write-Pfad (1 INSERT + 1 UPSERT pro Transaktion)
- ❌ Neue Tabelle + Apply-Funktion als Code (~100–200 Zeilen)
- ❌ Projektion muss konsistent mit Event-Replay-Logik bleiben — Testabdeckung essenziell
- ❌ JSONB-Spalten in `table_state` für Positionslisten

**Gesamtbewertung:** Kombiniert fachliche Richtigkeit des Event Sourcing mit praktischer Lesbarkeit. Der Mehraufwand ist gering und der Architekturgewinn erheblich: Snapshot-Anti-Pattern wird eliminiert, Read-Performance wird optimal, Reporting wird ermöglicht.

### Option D: Insert-Only CRUD (verworfen)

Normalisierte Tabellen wie Option A, aber alle INSERT-only (nie UPDATE/DELETE). Status wird durch Existenz von Einträgen in Verknüpfungstabellen berechnet.

Konvergiert strukturell auf Event Sourcing — ohne dessen Vorteile (einheitlicher Event Stream, triviales OCC, Kassenjournal als 1 Query). 8+ INSERT-only-Tabellen statt 1 Events-Tabelle, Kassenjournal als UNION, ungelöstes OCC-Problem. Kein etabliertes Pattern — weder reines ES noch reines CRUD. **Nicht empfohlen.**

---

## Bewertungsmatrix

| Kriterium             |   Gewicht   |  A: CRUD   | B: ES (Ist) | C: ES + Projektion | D: Insert-Only |
| --------------------- | :---------: | :--------: | :---------: | :----------------: | :------------: |
| Fachliche Passung     |  **Hoch**   |    ⭐⭐    | ⭐⭐⭐⭐⭐  |     ⭐⭐⭐⭐⭐     |     ⭐⭐⭐     |
| Audit-Trail           |  **Hoch**   |    ⭐⭐    | ⭐⭐⭐⭐⭐  |     ⭐⭐⭐⭐⭐     |    ⭐⭐⭐⭐    |
| Unveränderlichkeit    |  **Hoch**   |   ⭐⭐⭐   | ⭐⭐⭐⭐⭐  |     ⭐⭐⭐⭐⭐     |    ⭐⭐⭐⭐    |
| Read-Performance K-05 | **Mittel**  | ⭐⭐⭐⭐⭐ |    ⭐⭐     |     ⭐⭐⭐⭐⭐     |    ⭐⭐⭐⭐    |
| Reporting R-01–R-05   | **Mittel**  | ⭐⭐⭐⭐⭐ |    ⭐⭐     |      ⭐⭐⭐⭐      |    ⭐⭐⭐⭐    |
| Schema-Einfachheit    | **Mittel**  |    ⭐⭐    | ⭐⭐⭐⭐⭐  |      ⭐⭐⭐⭐      |      ⭐⭐      |
| OCC-Einfachheit       | **Mittel**  |    ⭐⭐    | ⭐⭐⭐⭐⭐  |     ⭐⭐⭐⭐⭐     |      ⭐⭐      |
| Einstiegshürde        | **Mittel**  | ⭐⭐⭐⭐⭐ |   ⭐⭐⭐    |       ⭐⭐⭐       |     ⭐⭐⭐     |
| Ad-hoc-SQL            | **Niedrig** | ⭐⭐⭐⭐⭐ |    ⭐⭐     |       ⭐⭐⭐       |    ⭐⭐⭐⭐    |

**Zusammenfassung:** Option A gewinnt bei Einstiegshürde, SQL-Analyse und Reporting — verliert bei fachlicher Passung, Audit Trail und OCC. Option B gewinnt bei fachlicher Passung und Schema-Einfachheit — verliert bei Read-Performance und Reporting. **Option C erbt alle Stärken von B und behebt die Read-/Reporting-Schwächen.** Option D bietet keine klaren Vorteile.

---

## Begründung

### 1. Die Domäne ist narrativ — nicht zustandsbasiert

Bestellungen, Zahlungen, Lieferungen, Stornierungen sind diskrete Geschäftsvorfälle. Ein Tisch „hat" keinen Saldo — ein Saldo _entsteht_ aus der Summe der Vorfälle. Event Sourcing bildet das direkt ab. CRUD erzwingt eine zustandsbasierte Sicht auf eine inhärent vorgangsbasierte Domäne.

Die Buchführungs-Analogie ist direkt zutreffend:

| Buchführung                                            | jotti                                        |
| ------------------------------------------------------ | -------------------------------------------- |
| Buchung (immutable Eintrag im Hauptbuch)               | Event (immutable Eintrag im Event Stream)    |
| Korrekturbuchung (Gegenbuchung, nie Löschung)          | Stornierung (neues Event, kein DELETE)       |
| Kontostand (berechnet aus allen Buchungen)             | Tisch-Saldo (berechnet aus allen Events)     |
| Hauptbuch (chronologische, unveränderliche Auflistung) | Kassenjournal (chronologischer Event Stream) |

**Kernbeobachtung:** Wenn man CRUD sauber implementiert (mit Audit Trail, INSERT-only für Vorgangstabellen, denormalisierte Produktdaten), konvergiert man strukturell auf ein Event-Sourcing-ähnliches Modell — nur verteilt auf mehr Tabellen und ohne einheitlichen Event Stream.

### 2. Immutabilität ist nicht optional — und bei ES strukturell gegeben

Wenn Geldbeträge im Spiel sind, muss Manipulation auf DB-Ebene verhindert werden. ES + DB-Trigger (`prevent_event_mutation()`) bietet das als einheitlichen Mechanismus. CRUD mit Audit-Log bietet es als zwei getrennte Mechanismen, deren Konsistenz zueinander sichergestellt werden muss.

> **Ehrliche Einordnung:** Immutabilität ist _nicht_ ES-exklusiv. Ein CRUD-System kann denselben Schutz bieten — über INSERT-only-Tabellen mit DB-Trigger. Der Unterschied: Bei ES ist der Event Stream _gleichzeitig_ Datenspeicher und Audit Trail (ein Mechanismus). Bei CRUD braucht man Datentabellen + separates Audit-Log (zwei Mechanismen).

### 3. Audit Trail = Event Stream — ein Mechanismus

Der Event Stream IST das Kassenjournal (K-06). Bei CRUD müsste man einen separaten Audit-Mechanismus bauen, der funktional dasselbe leistet, aber ohne die Garantie, dass der Audit Trail die _einzige_ Wahrheitsquelle ist.

### 4. Fat Events lösen das ACL-Problem explizit

Das Einfrieren von Produktpreisen zum Bestellzeitpunkt ist bei ES ein sichtbares, dokumentiertes Domänenkonzept (Anti-Corruption Layer / Fat Events). Bei CRUD geschieht dasselbe über denormalisierte Spalten, aber ohne die konzeptuelle Rahmung.

### 5. OCC ist bei ES trivial

Ein UNIQUE Constraint `(subject, version)` auf einer Tabelle. Bei CRUD bräuchte man eine synthetische `tisch_version`-Spalte (UPDATE auf einer eigentlich stabilen Entity) oder Multi-Table-Locking.

### 6. Die Projektion behebt die Read-Schwächen

`table_state` als synchrone Projektion macht Read-Zugriffe trivial (1 SELECT), eliminiert das Snapshot-as-Event-Anti-Pattern und ermöglicht Reporting — bei minimalem Write-Overhead (1 zusätzlicher UPSERT pro Transaktion). Details zur Projektionsarchitektur: [ADR: CQRS](cqrs.md).

---

## Akzeptierte Nachteile

1. **JSONB ohne DB-Typsicherheit.** Event-Payloads sind JSONB — die Anwendung validiert per zog, nicht die Datenbank. _Mitigation:_ Strikte Schema-Validierung im Code, Testabdeckung für Event-Parsing.

2. **Keine referentielle Integrität für Tisch-Operationen.** Eine `PositionRef` im Zahlungs-Event kann theoretisch auf eine nicht-existierende Position verweisen. _Mitigation:_ Invarianten-Prüfung in der Application-Schicht (`validatePositionRefs()`).

3. **Höhere Einstiegshürde.** Contributors müssen Event Replay, OCC-Versionierung und die Apply-Funktion verstehen. _Mitigation:_ Gut dokumentierte Domain-Schicht, [Handbuch §3](../design/handbuch.md).

4. **Ad-hoc-SQL-Analyse erschwert.** JSONB-Parsing statt Standard-SQL. _Mitigation:_ `table_state` enthält vorberechnete Werte; für tiefergehende Analyse existieren die JSONB-Operatoren.

5. **Projektion muss konsistent mit Event-Replay-Logik sein.** Zwei Codepfade für dieselbe Berechnung. _Mitigation:_ Apply-Funktion als Single Source of Truth, Replay-Funktionen darauf aufbauen lassen.

---

## Scope

| Bereich                                     | Persistenz                    | Begründung                                                                   |
| ------------------------------------------- | ----------------------------- | ---------------------------------------------------------------------------- |
| **Kassenbetrieb** (Tisch-Operationen)       | Event Sourcing (+ Projektion) | Vorgangsbasierte Domäne, Audit Trail, Immutabilität                          |
| **Stammdaten** (Produkte, Tische, Benutzer) | CRUD                          | Einfache Entities, kein Audit Trail im Domänenmodell nötig, stabile Schemata |
| **Auth** (Login, Passwort)                  | CRUD                          | Infrastruktur, kein Domänen-Audit-Wert                                       |

---

## Implementierungshinweise

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

> **Entfernt:** `tisch.snapshot:v1` wurde durch die synchrone `table_state`-Projektion abgelöst (siehe [ADR: CQRS](cqrs.md)). Der Snapshot-Event-Typ wird nicht mehr erzeugt und der zugehörige Code wurde vollständig entfernt.

### Append-Only-Garantie

- **Privilege Revocation**: Nur SELECT und INSERT erlaubt.
- **DB-Trigger** (`prevent_event_mutation()`) gegen UPDATE/DELETE/TRUNCATE.

### Synchrone Projektion

Die `table_state`-Projektion wird in derselben Transaktion wie das Event-INSERT aktualisiert. Details zur Projektionsarchitektur, zum Apply-Mechanismus und zur CQRS-Trennung: [ADR: CQRS](cqrs.md).

### CQRS

- **Commands** erstellen Events: `BestellungAufgeben`, `ZahlungRegistrieren`, `ProdukteStornieren`, `ProdukteLiefern`
- **Queries** lesen aus `table_state`: `GetTischSaldo`, `GetTischUnbezahlt`, `GetTischUngeliefert`
- **Queries** lesen aus `events`: `GetTischHistorie` (Kassenjournal)

---

## Revisionsklausel

Die Entscheidung sollte revidiert werden, wenn:

- Die Anzahl der Event-Typen deutlich wächst (> 10) und die Apply-Funktion unkontrollierbar wird
- Neue Anforderungen referentielle Integrität auf DB-Ebene zwingend erfordern
- Die JSONB-Parsing-Komplexität für Reporting nicht mehr beherrschbar ist

Bei jottis aktuellem Scope (4 fachliche Event-Typen, < 10k Events, bewusster Feature-Freeze) sind diese Szenarien unwahrscheinlich.

---

## Referenzen

- [ADR: CQRS](cqrs.md) — Projektionsarchitektur, Stufen-Modell, `table_state`-Details
- [Handbuch §3](../design/handbuch.md) — Domain-Modell, Tisch-Aggregat, Invarianten, Event Replay
- [Anforderungen](../anforderungen.md) — K-01–K-06, Q-02, Q-04, R-01–R-05
