# Domain-Driven Design (DDD) — Theorie und Anwendung in jotti

Dieses Dokument dient als theoretisches Nachschlagewerk für Domain-Driven Design im Kontext des jotti-Projekts. Es erklärt die zentralen DDD-Konzepte, bewertet deren Relevanz für ein Non-Profit-POS-System und zeigt, wo und wie jotti DDD-Prinzipien umsetzt.

> **Verwandte Dokumente:**
>
> - [Ubiquitous Language](../language.md) — Kanonische Fachbegriffe
> - [Event-Sourcing & CQRS Theorie](event-sourcing-cqrs.md) — Event-Sourcing und CQRS im Detail
> - [Go Backend Architektur](go-backend.md) — Schichtenarchitektur im Backend
> - [Architektur-Übersicht](README.md) — Index aller Theorie-Dokumente

---

## Inhaltsverzeichnis

1. [Was ist Domain-Driven Design?](#1-was-ist-domain-driven-design)
2. [Strategisches Design](#2-strategisches-design)
3. [Taktisches Design](#3-taktisches-design)
4. [DDD in jotti — Mapping](#4-ddd-in-jotti--mapping)
5. [Anti-Patterns und Fallstricke](#5-anti-patterns-und-fallstricke)
6. [Entscheidungshilfe: Wann lohnt sich DDD?](#6-entscheidungshilfe-wann-lohnt-sich-ddd)
7. [Referenzen](#7-referenzen)

---

## 1. Was ist Domain-Driven Design?

Domain-Driven Design ist ein Ansatz zur Softwareentwicklung, der die **Fachdomäne** in den Mittelpunkt stellt. Eric Evans formulierte DDD 2003 in seinem Buch _"Domain-Driven Design: Tackling Complexity in the Heart of Software"_. Die Kernthese:

> **Software sollte die Sprache und Struktur der Fachdomäne widerspiegeln, nicht die der Datenbank oder des Frameworks.**

DDD umfasst zwei Ebenen:

- **Strategisches Design** — Wie wird das System in Fachbereiche (Bounded Contexts) aufgeteilt?
- **Taktisches Design** — Wie werden Domain-Objekte innerhalb eines Bounded Context modelliert?

### Warum DDD für jotti?

jotti ist kein triviales CRUD-System. Die Tisch-Operationen (Bestellungen, Zahlungen, Stornierungen, Lieferungen) haben fachliche Regeln, die über einfaches Speichern hinausgehen:

- Saldo-Berechnung als Invariante
- Nur bestellte Positionen können bezahlt, geliefert oder storniert werden
- Stornierung erfordert eine erhöhte Berechtigung (Serviceleitung/Admin)
- Jede Aktion erzeugt ein unveränderliches Event (Audit Trail)

Diese Komplexität rechtfertigt ein bewusstes Domain-Modell — aber kein Enterprise-Grade DDD mit Dutzenden Aggregaten. jotti nutzt DDD **pragmatisch und selektiv**.

---

## 2. Strategisches Design

### 2.1 Ubiquitous Language

Die Ubiquitous Language ist das Fundament von DDD. Alle Beteiligten — Entwickler, Domänenexperten, Dokumentation — verwenden **dieselben Begriffe**. Es gibt keine Übersetzung zwischen „was der Fachbereich sagt" und „was der Code tut".

**Prinzipien:**

- **Ein Begriff = eine Bedeutung.** „Bestellung" heißt im Code `Bestellung`, nicht `Order`, `Request` oder `Transaction`.
- **Begriffe sind kontextgebunden.** „Tisch" im Kassenbetrieb ist eine Abrechnungseinheit; „Tisch" in den Stammdaten ist ein physischer Gegenstand mit Name und Status.
- **Die Sprache entwickelt sich weiter.** Wenn ein neues Konzept entsteht, wird es benannt und dokumentiert.

**jotti-Umsetzung:** Fachbegriffe der Domäne sind deutsch (Bestellung, Zahlung, Tisch, Stornierung). Infrastruktur-Code bleibt englisch (Auth, Config, DB). Die kanonischen Begriffe sind in [docs/language.md](../language.md) dokumentiert.

### 2.2 Bounded Contexts

Ein Bounded Context ist ein klar abgegrenzter Fachbereich mit eigener Ubiquitous Language und eigenem Modell. Innerhalb eines Bounded Context sind Begriffe eindeutig; über Kontextgrenzen hinweg können sie unterschiedliche Bedeutungen haben.

**Warum Bounded Contexts?**

In einem monolithischen Modell entsteht schnell ein „Big Ball of Mud" — ein einzelnes Modell, das alle Aspekte abdecken soll und dabei inkonsistent wird. Bounded Contexts schaffen klare Grenzen:

```
┌──────────────────────────────────────────────────────────────────────┐
│                         jotti (System)                               │
│                                                                      │
│  ┌─────────────────────────┐     ┌──────────────────────────────┐   │
│  │   Kassenbetrieb          │     │   Stammdaten                  │   │
│  │   (Bounded Context)      │     │   (Bounded Context)           │   │
│  │                          │     │                               │   │
│  │  Tisch = Abrechnungs-    │     │  Tisch = Physischer Tisch     │   │
│  │         einheit           │     │         mit Name + Status     │   │
│  │  Bestellung              │     │  Produkt + Variante           │   │
│  │  Zahlung                 │     │  Benutzer                     │   │
│  │  Lieferung               │     │                               │   │
│  │  Stornierung             │     │                               │   │
│  │                          │     │                               │   │
│  │  Persistenz: Events      │     │  Persistenz: CRUD             │   │
│  └─────────────────────────┘     └──────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────┐                                        │
│  │   Auth (Infrastruktur)   │  ← Kein eigenständiger                │
│  │   Login, JWT, Rollen     │    Bounded Context                    │
│  └─────────────────────────┘                                        │
└──────────────────────────────────────────────────────────────────────┘
```

**jotti-Umsetzung:**

| Bounded Context   | Backend-Pfad                                     | API-Prefix   | Persistenz     |
| ----------------- | ------------------------------------------------ | ------------ | -------------- |
| **Kassenbetrieb** | `api/table/`                                     | `/service/*` | Event-Sourcing |
| **Stammdaten**    | `api/product/`, `api/user/`, `api/table/` (CRUD) | `/admin/*`   | CRUD           |
| **Auth**          | `api/auth/`                                      | `/auth/*`    | CRUD (Users)   |

### 2.3 Context Mapping

Context Mapping beschreibt, wie Bounded Contexts miteinander interagieren. DDD definiert verschiedene Beziehungstypen:

| Beziehungstyp             | Beschreibung                                      | Relevanz für jotti                                                                     |
| ------------------------- | ------------------------------------------------- | -------------------------------------------------------------------------------------- |
| **Shared Kernel**         | Gemeinsamer Code zwischen Kontexten               | `Position`-Struct wird in Kassenbetrieb und Events geteilt                             |
| **Customer/Supplier**     | Ein Kontext liefert Daten, der andere konsumiert  | Stammdaten (Supplier) → Kassenbetrieb (Customer): Produkt-IDs, Varianten-Namen, Preise |
| **Conformist**            | Consumer übernimmt exakt das Modell des Suppliers | Kassenbetrieb übernimmt Varianten-IDs und Preise aus Stammdaten                        |
| **Anti-Corruption Layer** | Übersetzungsschicht zwischen Kontexten            | Nicht nötig — jotti ist ein Monolith mit geteilter DB                                  |
| **Published Language**    | Standardisiertes Format für Daten                 | Event-Format (CloudEvents-inspiriert) mit Versionierung                                |

**In jotti:** Die Beziehung zwischen Kassenbetrieb und Stammdaten ist ein **Customer/Supplier**-Verhältnis. Der Kassenbetrieb referenziert Produkt-IDs und kopiert Name + Preis **zum Zeitpunkt der Bestellung** in das Event. Dadurch sind Events historisch korrekt, selbst wenn der Produktpreis später geändert wird.

---

## 3. Taktisches Design

### 3.1 Entities

Eine Entity ist ein Objekt mit **eigener Identität**, die über die Zeit bestehen bleibt. Zwei Entities mit denselben Attributen sind nicht gleich, wenn ihre IDs unterschiedlich sind.

**Merkmale:**

- Eindeutige ID (z.B. Datenbank-PK, UUID)
- Veränderlicher Zustand (Attribute können sich ändern)
- Lebenszyklus (Erstellung, Änderung, Löschung)

**jotti-Beispiele:**

| Entity     | ID                    | Veränderlicher Zustand       |
| ---------- | --------------------- | ---------------------------- |
| `User`     | `users.id`            | Name, Username, Role, Status |
| `Tisch`    | `tables.id`           | Name, Status                 |
| `Produkt`  | `products.id`         | Name, Kategorie, Status      |
| `Variante` | `product_variants.id` | Name, Preis, Status          |

### 3.2 Value Objects

Ein Value Object hat **keine eigene Identität**. Es wird vollständig durch seine Attribute definiert. Zwei Value Objects mit denselben Attributen sind gleich.

**Merkmale:**

- Keine ID
- Immutable (unveränderlich nach Erstellung)
- Gleichheit über Attribute, nicht über Identität
- Austauschbar (wird ersetzt, nicht geändert)

**jotti-Beispiele:**

| Value Object  | Attribute                      | Kontext                                |
| ------------- | ------------------------------ | -------------------------------------- |
| `Position`    | ID, Name, PreisCents, Quantity | Eine Zeile in einer Bestellung/Zahlung |
| `Geldbeträge` | Wert in Cents (int)            | Immer als Cent-Integer, nie Float      |
| `Rolle`       | admin, senior_service, service | Enum, keine eigene Identität           |
| `Kategorie`   | food, beverage, other          | Enum, keine eigene Identität           |

**Warum sind Geldbeträge Value Objects?**

Geld ist das klassische Beispiel für Value Objects. Ein 10€-Schein ist austauschbar — es zählt nur der Wert, nicht welcher konkrete Schein es ist. In jotti werden Geldbeträge als `int` (Cents) modelliert, was Floating-Point-Rundungsfehler verhindert:

```go
// RICHTIG: Geldbetrag als Cent-Integer (Value Object)
type Position struct {
    PreisCents int  // 350 = 3,50€
    Quantity   int
}

// FALSCH: Geldbetrag als Float
type Position struct {
    Price    float64  // 3.50 — Rundungsfehler möglich!
    Quantity int
}
```

### 3.3 Aggregates

Ein Aggregate ist ein Cluster von Entities und Value Objects, die **als eine Einheit** behandelt werden. Das Aggregate definiert eine **Konsistenzgrenze**: Alle Invarianten innerhalb des Aggregates müssen bei jeder Transaktion erfüllt sein.

**Bestandteile:**

- **Aggregate Root** — Die einzige Entity, über die von außen auf das Aggregate zugegriffen wird
- **Interne Entities/Value Objects** — Nur über die Aggregate Root erreichbar
- **Invarianten** — Geschäftsregeln, die immer gelten müssen

**Das Tisch-Aggregat in jotti:**

```
┌─────────────────────────────────────────────────────┐
│              Tisch-Aggregat                          │
│              (Aggregate Root: Tisch)                 │
│                                                     │
│  Events (Value Objects, immutable):                 │
│  ┌──────────────────────────────────────────────┐   │
│  │ BestellungAufgegeben  ← Positionen           │   │
│  │ ZahlungRegistriert    ← Positionen           │   │
│  │ ProdukteGeliefert     ← Positionen           │   │
│  │ ProdukteStorniert     ← Positionen           │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  Rekonstruierter Zustand:                           │
│  ┌──────────────────────────────────────────────┐   │
│  │ SaldoCents            (Invariante: berechnet) │   │
│  │ UnbezahltePositionen  (abgeleitet)           │   │
│  │ UngeliefertePositionen (abgeleitet)          │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  Invarianten:                                       │
│  • Saldo = Σ Bestellungen − Σ Zahlungen − Σ Storno │
│  • Nur bestellte Positionen können bezahlt werden   │
│  • Nur bestellte Positionen können storniert werden │
│  • Event-Stream ist append-only (immutable)         │
└─────────────────────────────────────────────────────┘
```

**Warum ist der Tisch die Aggregate Root?**

- Alle Kassenbetrieb-Operationen beziehen sich auf **einen Tisch**
- Der Event-Stream wird über das Subject `"tisch:<id>"` identifiziert
- Alle Invarianten (Saldo, unbezahlte/ungelieferte Positionen) sind tisch-bezogen
- Konkurrierende Zugriffe auf denselben Tisch müssen konsistent aufgelöst werden

### 3.4 Domain Events

Ein Domain Event beschreibt etwas, das in der Domäne **passiert ist** — in der Vergangenheitsform. Domain Events sind immutable und bilden das Kassenjournal.

**Merkmale:**

- Vergangenheitsform: `BestellungAufgegeben`, nicht `BestellungAufgeben`
- Immutable (unveränderlich nach Erstellung)
- Tragen alle relevanten Daten (self-contained)
- Versioniert (`:v1`-Suffix) für Schema-Evolution

**jotti-Domain-Events:**

| Domain Event           | Typ-String                       | Bedeutung                          |
| ---------------------- | -------------------------------- | ---------------------------------- |
| `BestellungAufgegeben` | `tisch.bestellung-aufgegeben:v1` | Gast hat Positionen bestellt       |
| `ZahlungRegistriert`   | `tisch.zahlung-registriert:v1`   | Zahlung für Positionen eingegangen |
| `ProdukteStorniert`    | `tisch.produkte-storniert:v1`    | Positionen wurden storniert        |
| `ProdukteGeliefert`    | `tisch.produkte-geliefert:v1`    | Positionen wurden ausgeliefert     |

**Zusätzlich (technisch, kein Domain Event):**

| System Event | Typ-String          | Bedeutung                                      |
| ------------ | ------------------- | ---------------------------------------------- |
| `Snapshot`   | `tisch.snapshot:v1` | Technische Optimierung — komprimierter Zustand |

### 3.5 Domain Services

Ein Domain Service enthält Geschäftslogik, die **keiner einzelnen Entity zugeordnet** werden kann. Domain Services sind zustandslos und operieren auf Aggregaten oder Events.

**jotti-Beispiele:**

```go
// Domain Service: Zustandsrekonstruktion aus Events
// In: domain/table/events.go
func GetSaldoFromEvents(events []event.Event) int { ... }
func GetUnbezahltePositionenFromEvents(events []event.Event) []Position { ... }
func GetUngeliefertePositionenFromEvents(events []event.Event) []Position { ... }
```

Diese Funktionen sind **reine Funktionen** (keine Seiteneffekte, deterministisch) und gehören zur Domain-Schicht, nicht zur Application-Schicht.

### 3.6 Application Services

Application Services orchestrieren den Ablauf einer Anwendungsoperation. Sie koordinieren Domain-Objekte, Repositories und Infrastruktur — enthalten aber **keine Geschäftslogik** selbst.

**jotti-Beispiele:**

```go
// Application Service: Bestellung aufgeben
// In: api/table/application/command.go
func (s *CommandService) BestellungAufgeben(ctx, userID, tischID, positionen, comment) {
    // 1. Tisch aus DB laden (Repository)
    // 2. Event erstellen (Domain)
    // 3. Event speichern (Repository)
    // 4. Optional: Snapshot erstellen
}
```

### 3.7 Repositories

Ein Repository abstrahiert den Datenzugriff. Es bietet eine **Collection-artige Schnittstelle** für das Laden und Speichern von Aggregaten.

**Prinzipien:**

- Repository-Interface gehört zur Domain-Schicht
- Implementierung gehört zur Infrastruktur-Schicht
- Ein Repository pro Aggregate Root
- Keine SQL-Details in der Domain

**jotti-Repositories:**

| Repository     | Aggregat             | Muster                              |
| -------------- | -------------------- | ----------------------------------- |
| `event_repo`   | Events (Event Store) | Append-only: WriteEvent, ReadEvents |
| `table_repo`   | Tisch (CRUD)         | CRUD: Get, GetAll, Create, Update   |
| `product_repo` | Produkt + Varianten  | CRUD mit Nested Entities            |
| `user_repo`    | Benutzer             | CRUD + Password-Hash                |

---

## 4. DDD in jotti — Mapping

### Zusammenfassung: Was nutzt jotti von DDD?

| DDD-Konzept           | Genutzt?     | Umsetzung in jotti                                   |
| --------------------- | ------------ | ---------------------------------------------------- |
| Ubiquitous Language   | ✅ Ja        | Deutsche Fachbegriffe, dokumentiert in `language.md` |
| Bounded Contexts      | ✅ Ja        | Kassenbetrieb vs. Stammdaten vs. Auth                |
| Context Mapping       | ✅ Implizit  | Customer/Supplier (Stammdaten → Kassenbetrieb)       |
| Entities              | ✅ Ja        | User, Tisch, Produkt, Variante                       |
| Value Objects         | ✅ Ja        | Position, Geldbeträge (Cents), Rollen, Kategorien    |
| Aggregates            | ✅ Ja        | Tisch-Aggregat mit Event Stream                      |
| Domain Events         | ✅ Ja        | 4 Event-Typen + Snapshot                             |
| Domain Services       | ✅ Ja        | Zustandsrekonstruktion (reine Funktionen)            |
| Application Services  | ✅ Ja        | Command/Query Handler                                |
| Repositories          | ✅ Ja        | Repository-Pattern mit sqlc                          |
| Factories             | ⚠️ Teilweise | `New*`-Funktionen für Handler/Services               |
| Specifications        | ❌ Nein      | Nicht kompliziert genug                              |
| Sagas/Process Manager | ❌ Nein      | Keine verteilten Transaktionen                       |
| Module Mapping        | ✅ Ja        | Verzeichnisstruktur = Bounded Context                |

### Was jotti bewusst weglässt

**Specifications:** Das Specification-Pattern (komplexe Geschäftsregeln als Objekte) ist für jottis Regeln überdimensioniert. Die Validierung erfolgt direkt in `zog`-Schemas.

**Sagas:** Sagas koordinieren verteilte Transaktionen über mehrere Aggregaten hinweg. jotti ist ein Monolith — alle Operationen laufen in derselben Datenbank.

**CQRS mit separaten Datenbanken:** jotti nutzt eine einzelne PostgreSQL-Instanz für Write Store (Events) und Read Store (CRUD). Eine separate Read-DB ist für die aktuelle Last (Vereinsfeste) nicht nötig.

---

## 5. Anti-Patterns und Fallstricke

### 5.1 Anemic Domain Model

**Problem:** Domain-Objekte sind nur Datencontainer ohne Verhalten. Alle Logik steckt in Services.

**Symptome:**

- Domain-Structs haben nur Felder, keine Methoden
- Application Services enthalten Geschäftslogik
- Die Domain-Schicht könnte durch DTOs ersetzt werden

**In jotti:** Teilweise vorhanden — die Zustandsrekonstruktion erfolgt in Domain Services (`events.go`), nicht in den Event-Structs selbst. Das ist akzeptabel, weil Event-Sourcing naturgemäß den Zustand extern rekonstruiert. Die Events selbst sind Value Objects und **sollten** keine Methoden haben.

### 5.2 Big Ball of Mud

**Problem:** Kein erkennbares Modell, alles ist mit allem verbunden.

**Vermeidung in jotti:** Klare Verzeichnisstruktur (`domain/`, `repository/`, `api/`), Bounded Contexts als separate Pfade, Repository-Pattern als Abstraktion.

### 5.3 Premature Abstraction

**Problem:** Zu viele Schichten und Interfaces für ein einfaches Problem.

**In jotti:** Bewusst vermieden. CRUD-Entities (User, Produkt) haben keine Domain Events, keine Aggregates, kein Event-Sourcing. Nur der Kassenbetrieb rechtfertigt die zusätzliche Komplexität.

### 5.4 Shared Kernel Creep

**Problem:** Gemeinsam genutzte Typen wachsen unkontrolliert und koppeln Bounded Contexts.

**In jotti:** Der Shared Kernel ist minimal: `Position`-Struct und `event.Event`. Diese Typen sind stabil und ändern sich selten.

---

## 6. Entscheidungshilfe: Wann lohnt sich DDD?

### Entscheidungsmatrix

| Kriterium               | Einfaches CRUD         | DDD sinnvoll              |
| ----------------------- | ---------------------- | ------------------------- |
| Geschäftsregeln         | Wenige, offensichtlich | Komplex, sich entwickelnd |
| Datenmodell             | 1:1 mit DB-Schema      | Abweichend von DB-Schema  |
| Team-Kommunikation      | Technisch geprägt      | Fachlich geprägt          |
| Änderungsrate Fachlogik | Niedrig                | Hoch                      |
| Audit/Compliance        | Nicht relevant         | Kritisch                  |
| Persistenzmuster        | Standard CRUD          | Event-Sourcing, CQRS      |

### Anwendung auf jotti

- **Stammdaten** → Einfaches CRUD. Keine DDD-Patterns nötig (außer Repository).
- **Kassenbetrieb** → DDD sinnvoll. Geschäftsregeln, Event-Sourcing, Audit Trail, fachliche Sprache.

**Faustregel:** DDD dort einsetzen, wo die Geschäftslogik die Persistenzlogik an Komplexität übersteigt.

---

## 7. Referenzen

### Bücher

- **Eric Evans** (2003): _Domain-Driven Design: Tackling Complexity in the Heart of Software_ — Das Grundlagenwerk
- **Vaughn Vernon** (2013): _Implementing Domain-Driven Design_ — Praxisorientierter Nachfolger
- **Scott Millett & Nick Tune** (2015): _Patterns, Principles, and Practices of Domain-Driven Design_ — Muster-Katalog

### Online-Quellen

- [DDD Foundational Guide](https://spartner.software/kennisbank/domain-driven-design-ddd) — Strategisches & taktisches Design kompakt erklärt
- [Martin Fowler: DDD](https://martinfowler.com/tags/domain%20driven%20design.html) — Artikel-Sammlung zu DDD-Patterns
- [Event-Driven Architecture in Golang](https://github.com/PacktPublishing/Event-Driven-Architecture-in-Golang) — DDD + Event-Sourcing + CQRS in Go

### Projekt-intern

- [Ubiquitous Language](../language.md) — Kanonische Fachbegriffe
- [ADR: Event-Sourcing](../adr/event-sourcing.md) — Entscheidung für Event-Sourcing
- [ADR: sqlc](../adr/orm.md) — Entscheidung für sqlc
